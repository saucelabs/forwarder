// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package run

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/header"
	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/internal/version"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/pac"
	"github.com/saucelabs/forwarder/ruleset"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/httphandler"
	"github.com/saucelabs/forwarder/utils/osdns"
	"github.com/spf13/cobra"
	"go.uber.org/goleak"
	"go.uber.org/multierr"
)

type command struct {
	promReg             *prometheus.Registry
	dnsConfig           *osdns.Config
	httpTransportConfig *forwarder.HTTPTransportConfig
	pac                 *url.URL
	credentials         []*forwarder.HostPortUser
	denyDomains         []ruleset.RegexpListItem
	directDomains       []ruleset.RegexpListItem
	proxyHeaders        []header.Header
	requestHeaders      []header.Header
	responseHeaders     []header.Header
	httpProxyConfig     *forwarder.HTTPProxyConfig
	mitm                bool
	mitmConfig          *forwarder.MITMConfig
	mitmDomains         []ruleset.RegexpListItem
	apiServerConfig     *forwarder.HTTPServerConfig
	logConfig           *log.Config
	goleak              bool
}

func (c *command) runE(cmd *cobra.Command, _ []string) (cmdErr error) { //nolint:maintidx // glue code
	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)

	defer func() {
		if cmdErr != nil {
			logger.Errorf("fatal error exiting: %s", cmdErr)
			cmd.SilenceErrors = true
		}
	}()

	var ep []forwarder.APIEndpoint

	{
		var (
			cfgStr string
			err    error
		)

		d := cobrautil.FlagsDescriber{
			Format: cobrautil.Plain,
		}
		cfgStr, err = d.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Infof("configuration\n%s", cfgStr)

		d.ShowNotChanged = true
		cfgStr, err = d.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Debugf("all configuration\n%s\n\n", cfgStr)

		ep = append(ep, forwarder.APIEndpoint{
			Path:    "/configz",
			Handler: httphandler.SendFileString("text/plain", cfgStr),
		})
	}

	martianlog.SetLogger(logger.Named("proxy"))

	if len(c.dnsConfig.Servers) > 0 {
		s := strings.ReplaceAll(fmt.Sprintf("%s", c.dnsConfig.Servers), " ", ", ")
		logger.Named("dns").Infof("using DNS servers %v", s)
		if err := osdns.Configure(c.dnsConfig); err != nil {
			return fmt.Errorf("configure dns: %w", err)
		}
	}

	var (
		pr forwarder.PACResolver
		rt http.RoundTripper
	)

	{
		var err error
		rt, err = forwarder.NewHTTPTransport(c.httpTransportConfig)
		if err != nil {
			return err
		}
	}

	if c.pac != nil {
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

	if len(c.proxyHeaders) > 0 {
		c.httpProxyConfig.ConnectRequestModifier = func(req *http.Request) error {
			if req.Header == nil {
				req.Header = http.Header{}
			}
			for _, h := range c.proxyHeaders {
				h.Apply(req.Header)
			}
			return nil
		}
	}

	if len(c.requestHeaders) > 0 {
		c.httpProxyConfig.RequestModifiers = append(c.httpProxyConfig.RequestModifiers, header.Headers(c.requestHeaders))
	}
	if len(c.responseHeaders) > 0 {
		c.httpProxyConfig.ResponseModifiers = append(c.httpProxyConfig.ResponseModifiers, header.Headers(c.responseHeaders))
	}

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

	g := runctx.NewGroup()
	{
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

	if c.apiServerConfig.Addr != "" {
		if err := c.registerProcMetrics(); err != nil {
			return fmt.Errorf("register process metrics: %w", err)
		}

		ep := append([]forwarder.APIEndpoint{
			{
				Path:    "/version",
				Handler: httphandler.Version(version.Version, version.Time, version.Commit),
			},
		}, ep...)

		h := forwarder.NewAPIHandler("Forwarder "+version.Version, c.promReg, nil, ep...)
		a, err := forwarder.NewHTTPServer(c.apiServerConfig, h, logger.Named("api"))
		if err != nil {
			return err
		}
		defer a.Close()
		g.Add(a.Run)
	}

	if c.goleak {
		defer func() {
			if err := goleak.Find(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "goleak: %s", err)
			}
		}()
	}

	return g.Run()
}

func (c *command) registerProcMetrics() error {
	return multierr.Combine(
		// Note that ProcessCollector is only available in Linux and Windows.
		c.promReg.Register(collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{Namespace: c.httpProxyConfig.PromNamespace})),
		c.promReg.Register(collectors.NewGoCollector()),
	)
}

func Command() *cobra.Command {
	c := command{
		promReg:             prometheus.NewRegistry(),
		dnsConfig:           osdns.DefaultConfig(),
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
		httpProxyConfig:     forwarder.DefaultHTTPProxyConfig(),
		mitmConfig:          forwarder.DefaultMITMConfig(),
		apiServerConfig:     forwarder.DefaultHTTPServerConfig(),
		logConfig:           log.DefaultConfig(),
	}
	c.httpProxyConfig.PromRegistry = c.promReg
	c.apiServerConfig.Addr = "localhost:10000"

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
	bind.PAC(fs, &c.pac)
	bind.Credentials(fs, &c.credentials)
	bind.DenyDomains(fs, &c.denyDomains)
	bind.DirectDomains(fs, &c.directDomains)
	bind.ProxyHeaders(fs, &c.proxyHeaders)
	bind.RequestHeaders(fs, &c.requestHeaders)
	bind.ResponseHeaders(fs, &c.responseHeaders)
	bind.HTTPProxyConfig(fs, c.httpProxyConfig, c.logConfig)
	bind.MITMConfig(fs, &c.mitm, c.mitmConfig)
	bind.MITMDomains(fs, &c.mitmDomains)
	bind.HTTPServerConfig(fs, c.apiServerConfig, "api", forwarder.HTTPScheme)
	bind.PromNamespace(fs, &c.httpProxyConfig.PromNamespace)
	bind.AutoMarkFlagFilename(cmd)
	cmd.MarkFlagsMutuallyExclusive("proxy", "pac")

	fs.BoolVar(&c.goleak, "goleak", false, "enable goleak")
	bind.MarkFlagHidden(cmd, "goleak")

	return cmd
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
