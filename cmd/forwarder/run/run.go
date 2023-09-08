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
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/header"
	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/pac"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/osdns"
	"github.com/spf13/cobra"
	"go.uber.org/goleak"
)

type command struct {
	promReg             *prometheus.Registry
	dnsConfig           *osdns.Config
	httpTransportConfig *forwarder.HTTPTransportConfig
	pac                 *url.URL
	credentials         []*forwarder.HostPortUser
	proxyHeaders        []header.Header
	requestHeaders      []header.Header
	responseHeaders     []header.Header
	httpProxyConfig     *forwarder.HTTPProxyConfig
	mitm                bool
	mitmConfig          *forwarder.MITMConfig
	apiServerConfig     *forwarder.HTTPServerConfig
	logConfig           *log.Config
	goleak              bool
}

func (c *command) runE(cmd *cobra.Command, _ []string) error {
	config, err := cobrautil.DescribeFlags(cmd.Flags(), false, cobrautil.Plain)
	if err != nil {
		return err
	}

	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)
	logger.Debugf("configuration\n%s", config)

	// Google Martian uses a global logger package.
	ml := logger.Named("proxy")
	martianlog.SetLogger(ml)

	if len(c.dnsConfig.Servers) > 0 {
		s := strings.ReplaceAll(fmt.Sprintf("%s", c.dnsConfig.Servers), " ", ", ")
		logger.Named("dns").Infof("using DNS servers %v", s)
		if err := osdns.Configure(c.dnsConfig); err != nil {
			return fmt.Errorf("configure dns: %w", err)
		}
	}

	var (
		script string
		pr     forwarder.PACResolver
		rt     http.RoundTripper
	)

	{
		var err error
		rt, err = forwarder.NewHTTPTransport(c.httpTransportConfig)
		if err != nil {
			return err
		}
	}

	if c.pac != nil {
		var err error
		script, err = forwarder.ReadURLString(c.pac, rt)
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
	}

	cm, err := forwarder.NewCredentialsMatcher(c.credentials, logger.Named("credentials"))
	if err != nil {
		return fmt.Errorf("credentials: %w", err)
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

	if c.mitmConfig.CACertFile != "" {
		c.mitm = true
	}
	if c.mitm {
		c.httpProxyConfig.MITM = c.mitmConfig
	}

	var g runctx.Group
	p, err := forwarder.NewHTTPProxy(c.httpProxyConfig, pr, cm, rt, logger.Named("proxy"))
	if err != nil {
		return err
	}
	g.Add(p.Run)

	if c.apiServerConfig.Addr != "" {
		h := forwarder.NewAPIHandler(c.promReg, p.Ready, config, script)
		a, err := forwarder.NewHTTPServer(c.apiServerConfig, h, logger.Named("api"))
		if err != nil {
			return err
		}
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

func Command() (cmd *cobra.Command) {
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

	defer func() {
		fs := cmd.Flags()
		bind.DNSConfig(fs, c.dnsConfig)
		bind.HTTPTransportConfig(fs, c.httpTransportConfig)
		bind.PAC(fs, &c.pac)
		bind.Credentials(fs, &c.credentials)
		bind.RequestHeaders(fs, &c.requestHeaders)
		bind.ResponseHeaders(fs, &c.responseHeaders)
		bind.HTTPProxyConfig(fs, c.httpProxyConfig, c.logConfig)
		bind.MITMConfig(fs, &c.mitm, c.mitmConfig)
		bind.HTTPServerConfig(fs, c.apiServerConfig, "api", forwarder.HTTPScheme)
		bind.PromNamespace(fs, &c.httpProxyConfig.PromNamespace)
		bind.AutoMarkFlagFilename(cmd)
		cmd.MarkFlagsMutuallyExclusive("proxy", "pac")

		fs.BoolVar(&c.goleak, "goleak", false, "enable goleak")
		bind.MarkFlagHidden(cmd, "goleak")
	}()

	return &cobra.Command{
		Use:     "run [--address <host:port>] [--pac <path or url>] [--credentials <username:password@host:port>]...",
		Short:   "Start HTTP (forward) proxy server",
		Long:    long,
		Example: example,
		RunE:    c.runE,
	}
}

const long = `Start HTTP (forward) proxy server.
You can start HTTP or HTTPS server.
If you start an HTTPS server and you don't provide a certificate, the server will generate a self-signed certificate on startup.

The server may be protected by basic authentication.
Whenever applicable, username and password are URL decoded.
This allows you to pass in special characters such as @ by using %%40 or pass in a colon with %%3a.
`

const example = `  # HTTP proxy with upstream proxy
  forwarder run --proxy http://localhost:8081

  # Start HTTP proxy with PAC script
  forwarder run --address localhost:3128 --pac https://example.com/pac.js

  # HTTPS proxy server with basic authentication
  forwarder run --protocol https --address localhost:8443 --basic-auth user:password
`
