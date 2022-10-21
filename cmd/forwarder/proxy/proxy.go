// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type command struct {
	promReg               *prometheus.Registry
	dnsConfig             *forwarder.DNSConfig
	httpProxyConfig       *forwarder.HTTPProxyConfig
	httpProxyServerConfig *forwarder.HTTPServerConfig
	apiServerConfig       *forwarder.HTTPServerConfig
	logConfig             *log.Config
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)

	var resolver *net.Resolver
	if len(c.dnsConfig.Servers) > 0 {
		r, err := forwarder.NewResolver(c.dnsConfig, logger.Named("dns"))
		if err != nil {
			return err
		}
		resolver = r
	}

	p, err := forwarder.NewHTTPProxy(c.httpProxyConfig, resolver, logger.Named("proxy"))
	if err != nil {
		return err
	}
	s, err := forwarder.NewHTTPServer(c.httpProxyServerConfig, p, logger.Named("server"))
	if err != nil {
		return err
	}

	a, err := forwarder.NewHTTPServer(c.apiServerConfig, forwarder.APIHandler(s, c.promReg), logger.Named("api"))
	if err != nil {
		return err
	}

	var eg *errgroup.Group
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	eg, ctx = errgroup.WithContext(ctx)
	eg.Go(func() error { return s.Run(ctx) })
	eg.Go(func() error { return a.Run(ctx) })
	return eg.Wait()
}

const long = `Start HTTP proxy. The proxy can listen to HTTP, HTTPS or HTTP2 traffic. 
It can be configured to use upstream proxy or PAC file.
It supports basic authentication for the proxy, the upstream proxy and backend servers. 
It supports custom DNS servers. 
`

const example = `Start HTTP proxy listening to localhost:8080:
  $ forwarder proxy --address localhost:8080

  Start a protected proxy protected with basic auth:
  $ forwarder proxy --address localhost:8080 --basic-auth user:pass

  Forward connections to an upstream proxy:
  $ forwarder proxy --address localhost:8080 --upstream-proxy-uri http://localhost:8089

  Forward connections to an upstream proxy protected with basic auth:
  $ forwarder proxy --address localhost:8080 --upstream-proxy-uri http://localhost:8089 --upstream-proxy-basic-auth user:pass

  Forward connections to an upstream proxy setup via PAC: 
  $ forwarder proxy --address localhost:8080 --pac-uri http://localhost:8090/pac

  Forward connections to an upstream proxy, setup via PAC protected with basic auth:
  $ forwarder proxy --address localhost:8080 --pac-uri http://user:pass@localhost:8090/pac -d http://user3:pwd4@localhost:8091 -d http://user2:pwd2@localhost:8092 

  Add basic auth header to requests to foo.bar:* and qux.baz:80.
  $ forwarder proxy --address localhost:8080 --site-credentials "foo.bar:0,qux.baz:80"
`

func Command() (cmd *cobra.Command) {
	c := command{
		promReg:               prometheus.NewRegistry(),
		dnsConfig:             forwarder.DefaultDNSConfig(),
		httpProxyConfig:       forwarder.DefaultHTTPProxyConfig(),
		httpProxyServerConfig: forwarder.DefaultHTTPServerConfig(),
		apiServerConfig:       forwarder.DefaultHTTPServerConfig(),
		logConfig:             log.DefaultConfig(),
	}
	c.httpProxyServerConfig.PromRegistry = c.promReg
	c.httpProxyServerConfig.BasicAuthHeader = middleware.ProxyAuthorizationHeader
	c.apiServerConfig.Addr = "localhost:0"

	defer func() {
		fs := cmd.Flags()
		bind.DNSConfig(fs, c.dnsConfig)
		bind.HTTPProxyConfig(fs, c.httpProxyConfig)
		bind.HTTPServerConfig(fs, c.httpProxyServerConfig, "")
		bind.HTTPServerConfig(fs, c.apiServerConfig, "api")
		bind.LogConfig(fs, c.logConfig)

		cmd.MarkFlagsMutuallyExclusive("upstream-proxy-uri", "pac-uri")
	}()
	return &cobra.Command{
		Use:     "proxy",
		Short:   "Start HTTP proxy",
		Long:    long,
		Example: example,
		RunE:    c.RunE,
	}
}
