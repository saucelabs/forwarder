// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package run

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/header"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/internal/version"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/martianlog"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/pac"
	"github.com/saucelabs/forwarder/ruleset"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/httphandler"
	"github.com/saucelabs/forwarder/utils/httpx"
	"github.com/spf13/cobra"
	"go.uber.org/goleak"
	"go.uber.org/multierr"
)

type command struct {
	promReg             *prometheus.Registry
	dnsConfig           *forwarder.DNSConfig
	httpTransportConfig *forwarder.HTTPTransportConfig
	connectTo           []forwarder.HostPortPair
	pac                 *url.URL
	credentials         []*forwarder.HostPortUser
	denyDomains         []ruleset.RegexpListItem
	directDomains       []ruleset.RegexpListItem
	connectHeaders      []header.Header
	requestHeaders      []header.Header
	responseHeaders     []header.Header
	httpProxyConfig     *forwarder.HTTPProxyConfig
	mitm                bool
	mitmConfig          *forwarder.MITMConfig
	mitmDomains         []ruleset.RegexpListItem
	proxyProtocol       bool
	proxyProtocolConfig *forwarder.ProxyProtocolConfig
	apiServerConfig     *forwarder.HTTPServerConfig
	logConfig           *log.Config

	dryRun bool
	goleak bool
}

func (c *command) runE(cmd *cobra.Command, _ []string) (cmdErr error) {
	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	onError, err := c.registerErrorsMetric()
	if err != nil {
		return fmt.Errorf("register errors metric: %w", err)
	}
	logger := stdlog.New(c.logConfig, stdlog.WithOnError(onError))

	defer func() {
		if err := logger.Close(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "close logger: %s\n", err)
		}
	}()

	defer func() {
		if cmdErr != nil {
			logger.Errorf("fatal error exiting: %s", cmdErr)
			cmd.SilenceErrors = true
		}
	}()

	logger.Infof("Forwarder %s (%s)", version.Version, version.Commit)
	logger.Debugf("resource limits: GOMAXPROCS=%d GOMEMLIMIT=%s", runtime.GOMAXPROCS(0), os.Getenv("GOMEMLIMIT"))

	var ep []forwarder.APIEndpoint

	{
		var (
			cfg []byte
			err error
		)

		cfg, err = cobrautil.FlagsDescriber{
			Format:          cobrautil.Plain,
			ShowChangedOnly: true,
			ShowHidden:      true,
		}.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		if len(cfg) > 0 {
			logger.Infof("configuration\n%s", cfg)
		} else {
			logger.Infof("using default configuration")
		}

		cfg, err = cobrautil.FlagsDescriber{
			Format:          cobrautil.Plain,
			ShowChangedOnly: false,
			ShowHidden:      true,
		}.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Debugf("all configuration\n%s\n\n", cfg)

		ep = append(ep, forwarder.APIEndpoint{
			Path:    "/configz",
			Handler: httphandler.SendFile("text/plain", cfg),
		})
	}

	martianlog.SetLogger(logger.Named("proxy"))

	if len(c.dnsConfig.Servers) > 0 {
		s := strings.ReplaceAll(fmt.Sprintf("%s", c.dnsConfig.Servers), " ", ", ")
		logger.Named("dns").Infof("using DNS servers %v", s)
		if err := c.dnsConfig.Apply(); err != nil {
			return fmt.Errorf("configure dns: %w", err)
		}
	}

	if c.httpTransportConfig.TLSClientConfig.KeyLogFile != "" {
		logger.Infof("using TLS key logging, writing to %s", c.httpTransportConfig.TLSClientConfig.KeyLogFile)
	}

	if len(c.connectTo) > 0 {
		c.httpTransportConfig.RedirectFunc = forwarder.DialRedirectFromHostPortPairs(c.connectTo)
	}

	var pr forwarder.PACResolver
	if c.pac != nil {
		// Disable metrics for receiving PAC file.
		cfg := *c.httpTransportConfig
		cfg.PromRegistry = nil
		rt, err := forwarder.NewHTTPTransport(&cfg)
		if err != nil {
			return err
		}

		script, err := forwarder.ReadURLString(c.pac, rt)
		if err != nil {
			return fmt.Errorf("read PAC file: %w", err)
		}
		pr, err = pac.NewProxyResolverPool(&pac.ProxyResolverConfig{Script: script}, nil)
		if err != nil {
			return err
		}
		if _, err := pr.FindProxyForURL(&url.URL{Scheme: "https", Host: "saucelabs.com"}, ""); err != nil {
			return err
		}
		pr = &forwarder.LoggingPACResolver{
			Resolver: pr,
			Logger:   logger.Named("pac"),
		}

		ep = append(ep, forwarder.APIEndpoint{
			Path:    "/pac",
			Handler: httphandler.SendFileString("application/x-ns-proxy-autoconfig", script),
		})
	}

	cm, err := forwarder.NewCredentialsMatcher(c.credentials, logger.Named("credentials"))
	if err != nil {
		return fmt.Errorf("credentials: %w", err)
	}

	if len(c.denyDomains) > 0 {
		dd, err := ruleset.NewRegexpMatcherFromList(c.denyDomains)
		if err != nil {
			return fmt.Errorf("deny domains: %w", err)
		}
		c.httpProxyConfig.DenyDomains = dd
	}

	if len(c.directDomains) > 0 {
		dd, err := ruleset.NewRegexpMatcherFromList(c.directDomains)
		if err != nil {
			return fmt.Errorf("direct domains: %w", err)
		}
		c.httpProxyConfig.DirectDomains = dd
	}

	c.configureHeadersModifiers()

	if c.mitm || c.mitmConfig.CACertFile != "" || len(c.mitmDomains) > 0 {
		c.httpProxyConfig.MITM = c.mitmConfig

		if len(c.mitmDomains) > 0 {
			dd, err := ruleset.NewRegexpMatcherFromList(c.mitmDomains)
			if err != nil {
				return fmt.Errorf("mitm domains: %w", err)
			}
			c.httpProxyConfig.MITMDomains = dd
		}
	}

	if c.proxyProtocol {
		c.httpProxyConfig.ProxyProtocolConfig = c.proxyProtocolConfig
	}

	g := runctx.NewGroup()
	{
		rt, err := forwarder.NewHTTPTransport(c.httpTransportConfig)
		if err != nil {
			return err
		}
		rt.DialContext = martianlog.LoggingDialContext(rt.DialContext)
		c.transportWithProxyConnectHeader(rt)

		p, err := forwarder.NewHTTPProxy(c.httpProxyConfig, pr, cm, rt, logger.Named("proxy"))
		if err != nil {
			return err
		}
		defer p.Close()
		g.Add(p.Run)

		if ca := p.MITMCACert(); ca != nil {
			ep = append(ep, forwarder.APIEndpoint{
				Path:    "/cacert",
				Handler: httphandler.SendCACert(ca),
			})
		}
	}

	{
		if err := c.registerGoMemLimitMetric(); err != nil {
			return fmt.Errorf("register GOMEMLIMIT metric: %w", err)
		}
		if err := c.registerGoMaxProcsMetric(); err != nil {
			return fmt.Errorf("register GOMAXPROCS metrics: %w", err)
		}
		if err := c.registerProcMetrics(); err != nil {
			return fmt.Errorf("register process metrics: %w", err)
		}
		if err := c.registerVersionMetric(); err != nil {
			return fmt.Errorf("register version metric: %w", err)
		}

		ep := append([]forwarder.APIEndpoint{
			{
				Path:    "/version",
				Handler: httphandler.Version(version.Version, version.Time, version.Commit),
			},
		}, ep...)
		h := forwarder.NewAPIHandler("Forwarder "+version.Version, c.promReg, nil, ep...)

		if os.Getenv("PLATFORM") == "container" {
			g.Add(func(ctx context.Context) error {
				logger.Named("api").Infof("HTTP server listen socket path=%s", forwarder.APIUnixSocket)
				return httpx.ServeUnixSocket(ctx, h, forwarder.APIUnixSocket)
			})
		}

		if c.apiServerConfig.Addr != "" {
			a, err := forwarder.NewHTTPServer(c.apiServerConfig, h, logger.Named("api"))
			if err != nil {
				return err
			}
			defer a.Close()
			g.Add(a.Run)
		}
	}

	if c.goleak {
		defer func() {
			if err := goleak.Find(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "goleak: %s", err)
				os.Exit(1)
			}
		}()
	}

	if c.dryRun {
		return nil
	}

	return g.Run()
}

func (c *command) configureHeadersModifiers() {
	if len(c.connectHeaders) > 0 || len(c.requestHeaders) > 0 {
		connectHeaders := header.Headers(c.connectHeaders)
		requestHeaders := header.Headers(c.requestHeaders)
		m := forwarder.RequestModifierFunc(func(req *http.Request) error {
			if req.Method == http.MethodConnect {
				return connectHeaders.ModifyRequest(req)
			}
			return requestHeaders.ModifyRequest(req)
		})
		c.httpProxyConfig.RequestModifiers = append(c.httpProxyConfig.RequestModifiers, m)
	}
	if len(c.responseHeaders) > 0 {
		headers := header.Headers(c.responseHeaders)
		m := forwarder.ResponseModifierFunc(func(resp *http.Response) error {
			if req := resp.Request; req != nil && req.Method == http.MethodConnect {
				return nil
			}
			return headers.ModifyResponse(resp)
		})
		c.httpProxyConfig.ResponseModifiers = append(c.httpProxyConfig.ResponseModifiers, m)
	}
}

func (c *command) transportWithProxyConnectHeader(tr *http.Transport) {
	if len(c.connectHeaders) > 0 {
		tr.GetProxyConnectHeader = func(_ context.Context, _ *url.URL, _ string) (http.Header, error) {
			h := make(http.Header, len(c.connectHeaders))
			for _, ch := range c.connectHeaders {
				ch.Apply(h)
			}
			return h, nil
		}
	}
}

func (c *command) registerErrorsMetric() (func(name string), error) {
	m := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: c.httpProxyConfig.PromNamespace,
		Name:      "errors_total",
		Help:      "Number of errors",
	}, []string{"name"})

	if err := c.promReg.Register(m); err != nil {
		return nil, err
	}

	return func(name string) {
		m.WithLabelValues(name).Inc()
	}, nil
}

func (c *command) registerGoMemLimitMetric() error {
	return c.promReg.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "go_env",
		Name:      "gomemlimit",
		Help:      "Memory limit for the process",
	}, func() float64 {
		e := os.Getenv("GOMEMLIMIT")
		if e == "" {
			return 0
		}

		var v forwarder.SizeSuffix
		if err := v.Set(e); err != nil {
			return -1
		}

		return float64(v)
	}))
}

func (c *command) registerGoMaxProcsMetric() error {
	return c.promReg.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "go_env",
		Name:      "gomaxprocs",
		Help:      "Number of maximum goroutines that can be executed simultaneously",
	}, func() float64 {
		return float64(runtime.GOMAXPROCS(0))
	}))
}

func (c *command) registerProcMetrics() error {
	return multierr.Combine(
		// Note that ProcessCollector is only available in Linux and Windows.
		c.promReg.Register(collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{Namespace: c.httpProxyConfig.PromNamespace})),
		c.promReg.Register(collectors.NewGoCollector()),
	)
}

func (c *command) registerVersionMetric() error {
	return c.promReg.Register(c.constMetric("version", "Forwarder version, value is always 1", prometheus.Labels{
		"version": version.Version,
		"commit":  version.Commit,
		"time":    version.Time,
	}))
}

func (c *command) constMetric(name, help string, labels prometheus.Labels) prometheus.GaugeFunc {
	return prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   c.httpProxyConfig.PromNamespace,
		Name:        name,
		Help:        help,
		ConstLabels: labels,
	}, func() float64 {
		return 1
	})
}

const promNs = "forwarder"

func Command() *cobra.Command {
	c := makeCommand()

	cmd := &cobra.Command{
		Use:     "run [--address <host:port>] [--pac <path or url>] [--credentials <username:password@host:port>]...",
		Short:   "Start HTTP (forward) proxy server",
		Long:    long,
		Example: example,
		RunE:    c.runE,
	}

	fs := cmd.Flags()
	bind.DNSConfig(fs, c.dnsConfig)
	bind.HTTPTransportConfig(fs, c.httpTransportConfig)
	bind.ConnectTo(fs, &c.connectTo)
	bind.PAC(fs, &c.pac)
	bind.Credentials(fs, &c.credentials)
	bind.DenyDomains(fs, &c.denyDomains)
	bind.DirectDomains(fs, &c.directDomains)
	bind.ConnectHeaders(fs, &c.connectHeaders)
	bind.RequestHeaders(fs, &c.requestHeaders)
	bind.ResponseHeaders(fs, &c.responseHeaders)
	bind.HTTPProxyConfig(fs, c.httpProxyConfig, c.logConfig)
	bind.MITMConfig(fs, &c.mitm, c.mitmConfig)
	bind.MITMDomains(fs, &c.mitmDomains)
	bind.ProxyProtocol(fs, &c.proxyProtocol, c.proxyProtocolConfig)
	bind.HTTPServerConfig(fs, c.apiServerConfig, "api", forwarder.HTTPScheme)
	bind.HTTPLogConfig(fs, []bind.NamedParam[httplog.Mode]{
		{Name: "api", Param: &c.apiServerConfig.LogHTTPMode},
		{Name: "proxy", Param: &c.httpProxyConfig.LogHTTPMode},
	})

	bind.ProxyHeaders(fs, &c.connectHeaders)
	fs.Lookup("proxy-header").Deprecated = "use --connect-header flag instead"
	cmd.MarkFlagsMutuallyExclusive("proxy-header", "connect-header")

	bind.AutoMarkFlagFilename(cmd)
	cmd.MarkFlagsMutuallyExclusive("proxy", "pac")

	fs.BoolVar(&c.goleak, "goleak", false, "enable goleak")

	bind.MarkFlagHidden(cmd,
		"goleak",
	)

	return cmd
}

func Metrics() (*prometheus.Registry, error) {
	c := makeCommand()
	c.logConfig = &log.Config{
		Level: log.ErrorLevel,
		File:  os.NewFile(10, os.DevNull),
	}
	c.dryRun = true

	cmd := &cobra.Command{
		Use:                "run",
		RunE:               c.runE,
		DisableFlagParsing: true,
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		return nil, err
	}

	return c.promReg, nil
}

func makeCommand() command {
	c := command{
		promReg:             prometheus.NewRegistry(),
		dnsConfig:           forwarder.DefaultDNSConfig(),
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
		httpProxyConfig:     forwarder.DefaultHTTPProxyConfig(),
		mitmConfig:          forwarder.DefaultMITMConfig(),
		proxyProtocolConfig: forwarder.DefaultProxyProtocolConfig(),
		apiServerConfig:     forwarder.DefaultHTTPServerConfig(),
		logConfig:           log.DefaultConfig(),
	}
	c.httpTransportConfig.PromRegistry = c.promReg
	c.httpTransportConfig.PromNamespace = promNs
	c.httpProxyConfig.PromRegistry = c.promReg
	c.httpProxyConfig.PromNamespace = promNs
	c.apiServerConfig.Addr = "localhost:10000"

	return c
}

const long = `Start HTTP (forward) proxy server.
You can start HTTP or HTTPS server.
If you start an HTTPS server and you don't provide a certificate, the server will generate a self-signed certificate on startup.
The server may be protected by basic authentication.
`

const example = `  # HTTP proxy with upstream proxy
  forwarder run --proxy http://localhost:8081

  # Start HTTP proxy with PAC script
  forwarder run --address localhost:3128 --pac https://example.com/pac.js

  # HTTPS proxy server with basic authentication
  forwarder run --protocol https --address localhost:8443 --basic-auth user:password
`
