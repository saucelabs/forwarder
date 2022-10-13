// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type command struct {
	dnsConfig        *forwarder.DNSConfig
	httpProxyConfig  *forwarder.HTTPProxyConfig
	httpServerConfig *forwarder.HTTPServerConfig
	logConfig        logConfig
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	var resolver *net.Resolver
	if len(c.dnsConfig.Servers) > 0 {
		r, err := forwarder.NewResolver(c.dnsConfig, newLogger(c.logConfig, "dns"))
		if err != nil {
			return err
		}
		resolver = r
	}

	p, err := forwarder.NewHTTPProxy(c.httpProxyConfig, resolver, newLogger(c.logConfig, "proxy"))
	if err != nil {
		return err
	}

	s, err := forwarder.NewHTTPServer(c.httpServerConfig, p, newLogger(c.logConfig, "server"))
	if err != nil {
		return err
	}

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return s.Run(ctx)
}

const long = `Start HTTP proxy. The proxy can listen to HTTP, HTTPS or HTTP2 traffic. 
It can be configured to use upstream proxy or PAC file.
It supports basic authentication for the proxy, the upstream proxy and backend servers. 
It supports custom DNS servers. 
`

const example = `Start HTTP proxy listening to localhost:8080:
  $ forwarder proxy --addr localhost:8080

  Start a protected proxy protected with basic auth:
  $ forwarder proxy --addr localhost:8080 --basic-auth user:pass

  Forward connections to an upstream proxy:
  $ forwarder proxy --addr localhost:8080 --upstream-proxy-uri http://localhost:8089

  Forward connections to an upstream proxy protected with basic auth:
  $ forwarder proxy --addr localhost:8080 --upstream-proxy-uri http://localhost:8089 --upstream-proxy-basic-auth user:pass

  Forward connections to an upstream proxy setup via PAC: 
  $ forwarder proxy --addr localhost:8080 --pac-uri http://localhost:8090/pac

  Forward connections to an upstream proxy, setup via PAC protected with basic auth:
  $ forwarder proxy --addr localhost:8080 --pac-uri http://user:pass@localhost:8090/pac -d http://user3:pwd4@localhost:8091 -d http://user2:pwd2@localhost:8092 

  Add basic auth header to requests to foo.bar:* and qux.baz:80.
  $ forwarder proxy --addr localhost:8080 --site-credentials "foo.bar:0,qux.baz:80"
`

func Command() (cmd *cobra.Command) {
	c := command{
		dnsConfig:        forwarder.DefaultDNSConfig(),
		httpProxyConfig:  forwarder.DefaultHTTPProxyConfig(),
		httpServerConfig: forwarder.DefaultHTTPServerConfig(),
		logConfig:        defaultLogConfig(),
	}
	c.httpServerConfig.BasicAuthHeader = middleware.ProxyAuthorizationHeader

	defer func() {
		fs := cmd.Flags()
		bind.DNSConfig(fs, c.dnsConfig)
		bind.HTTPProxyConfig(fs, c.httpProxyConfig)
		bind.HTTPServerConfig(fs, c.httpServerConfig, "")
		c.bindLogConfig(fs)

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

func (c *command) bindLogConfig(fs *pflag.FlagSet) {
	fs.StringVar(&c.logConfig.Level, "log-level", c.logConfig.Level, "the log level")
	fs.StringVar(&c.logConfig.FileLevel, "log-file-level", c.logConfig.FileLevel, "the log file level")
	fs.StringVar(&c.logConfig.FilePath, "log-file-path", c.logConfig.FilePath, "the log file path")
}
