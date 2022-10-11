// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package run

import (
	"context"
	"net"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type command struct {
	dnsConfig              *forwarder.DNSConfig
	httpProxyConfig        *forwarder.HTTPProxyConfig
	upstreamProxyBasicAuth *url.Userinfo
	httpServerConfig       *forwarder.HTTPServerConfig
	logConfig              logConfig
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	if c.upstreamProxyBasicAuth != nil && c.httpProxyConfig.UpstreamProxyURI != nil {
		c.httpProxyConfig.UpstreamProxyURI.User = c.upstreamProxyBasicAuth
	}

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

const long = `Start the proxy. HTTPProxy can be protected with basic auth.
It can forward connections to an upstream proxy (protected, or not).
The upstream proxy can be automatically setup via PAC (protected, or not).
Also, credentials for proxies specified in PAC can be set.
Note: Can't setup upstream, and PAC at the same time.
`

const example = `Start a proxy listening to http://localhost:8080:
  $ forwarder run

  Start a proxy listening to http://0.0.0.0:8085:
  $ forwarder run -l "http://0.0.0.0:8085"

  Start a protected proxy:
  $ forwarder run -l "http://user:pwd@localhost:8085"

  Start a protected proxy, forwarding connection to an upstream proxy running at
  http://localhost:8089:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -u "http://localhost:8089"

  Start a protected proxy, forwarding connection to a protected upstream proxy
  running at http://user1:pwd1@localhost:8089:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -u "http://user1:pwd1@localhost:8089"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - server running at http://localhost:8090:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -p "http://localhost:8090"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - protected server running at http://user2:pwd2@localhost:8090:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -p "http://user2:pwd2@localhost:8090"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - protected server running at http://user2:pwd2@localhost:8090, specifying
  credential for protected proxies specified in PAC:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -p "http://user2:pwd2@localhost:8090" \
    -d "http://user3:pwd4@localhost:8091,http://user4:pwd5@localhost:8092"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - protected server running at http://user2:pwd2@localhost:8090, specifying
  credential for protected proxies specified in PAC, also forwarding "localhost"
  requests thru the upstream proxy:
  $ forwarder run \
    -t \
    -l "http://user:pwd@localhost:8085" \
    -p "http://user2:pwd2@localhost:8090" \
    -d "http://user3:pwd4@localhost:8091,http://user4:pwd5@localhost:8092"

  Start a protected proxy that adds basic auth header to requests to foo.bar:8090
  and qux.baz:80.
  $ forwarder run \
    -t \
    -l "http://user:pwd@localhost:8085" \
    --site-credentials "user1:pwd1@foo.bar:8090,user2:pwd2@qux:baz:80"
`

func Command() (cmd *cobra.Command) {
	c := command{
		dnsConfig:        forwarder.DefaultDNSConfig(),
		httpProxyConfig:  forwarder.DefaultHTTPProxyConfig(),
		httpServerConfig: forwarder.DefaultHTTPServerConfig(),
		logConfig:        defaultLogConfig(),
	}
	c.httpServerConfig.BasicAuthHeader = forwarder.ProxyAuthorizationHeader

	defer func() {
		fs := cmd.Flags()
		c.bindDNSConfig(fs)
		c.bindHTTPProxyConfig(fs)
		c.bindHTTPServerConfig(fs)
		c.bindLogConfig(fs)

		cmd.MarkFlagsMutuallyExclusive("upstream-proxy-uri", "pac-uri")
	}()
	return &cobra.Command{
		Use:     "run",
		Short:   "Start the proxy",
		Long:    long,
		Example: example,
		RunE:    c.RunE,
	}
}

func (c *command) bindDNSConfig(fs *pflag.FlagSet) {
	fs.VarP(anyflag.NewSliceValue[*url.URL](nil, &c.dnsConfig.Servers, forwarder.ParseDNSURI),
		"dns-server", "n", "DNS server, ex. -n udp://1.1.1.1:53 (can be specified multiple times)")
	fs.DurationVar(&c.dnsConfig.Timeout, "dns-timeout", c.dnsConfig.Timeout, "timeout for DNS queries if DNS server is specified")
}

func (c *command) bindHTTPProxyConfig(fs *pflag.FlagSet) {
	fs.VarP(anyflag.NewValue[*url.URL](c.httpProxyConfig.UpstreamProxyURI, &c.httpProxyConfig.UpstreamProxyURI, forwarder.ParseProxyURI),
		"upstream-proxy-uri", "u", "upstream proxy URI")
	fs.VarP(anyflag.NewValue[*url.Userinfo](c.upstreamProxyBasicAuth, &c.upstreamProxyBasicAuth, forwarder.ParseUserInfo),
		"upstream-proxy-basic-auth", "", "upstream proxy basic auth in the form of `username:password`")
	fs.VarP(anyflag.NewValue[*url.URL](c.httpProxyConfig.PACURI, &c.httpProxyConfig.PACURI, url.ParseRequestURI),
		"pac-uri", "p", "URI to PAC content, or directly, the PAC content")
	fs.StringSliceVarP(&c.httpProxyConfig.PACProxiesCredentials, "pac-proxies-credentials", "d", c.httpProxyConfig.PACProxiesCredentials,
		"PAC proxies credentials using standard URI format")
	fs.StringSliceVar(&c.httpProxyConfig.SiteCredentials, "site-credentials", c.httpProxyConfig.SiteCredentials,
		"target site credentials")
	fs.BoolVarP(&c.httpProxyConfig.ProxyLocalhost, "proxy-localhost", "t", c.httpProxyConfig.ProxyLocalhost,
		"if set, will proxy localhost requests to an upstream proxy")

	fs.DurationVar(&c.httpProxyConfig.Transport.DialTimeout, "http-dial-timeout", c.httpProxyConfig.Transport.DialTimeout,
		"dial timeout for HTTP connections")
	fs.DurationVar(&c.httpProxyConfig.Transport.KeepAlive, "http-keep-alive", c.httpProxyConfig.Transport.KeepAlive,
		"keep alive interval for HTTP connections")
	fs.DurationVar(&c.httpProxyConfig.Transport.TLSHandshakeTimeout, "http-tls-handshake-timeout", c.httpProxyConfig.Transport.TLSHandshakeTimeout,
		"TLS handshake timeout for HTTP connections")
	fs.IntVar(&c.httpProxyConfig.Transport.MaxIdleConns, "http-max-idle-conns", c.httpProxyConfig.Transport.MaxIdleConns,
		"maximum number of idle connections for HTTP connections")
	fs.IntVar(&c.httpProxyConfig.Transport.MaxIdleConnsPerHost, "http-max-idle-conns-per-host", c.httpProxyConfig.Transport.MaxIdleConnsPerHost,
		"maximum number of idle connections per host for HTTP connections")
	fs.IntVar(&c.httpProxyConfig.Transport.MaxConnsPerHost, "http-max-conns-per-host", c.httpProxyConfig.Transport.MaxConnsPerHost,
		"maximum number of connections per host for HTTP connections")
	fs.DurationVar(&c.httpProxyConfig.Transport.IdleConnTimeout, "http-idle-conn-timeout", c.httpProxyConfig.Transport.IdleConnTimeout,
		"idle connection timeout for HTTP connections")
	fs.DurationVar(&c.httpProxyConfig.Transport.ResponseHeaderTimeout, "http-response-header-timeout", c.httpProxyConfig.Transport.ResponseHeaderTimeout,
		"response header timeout for HTTP connections")
	fs.DurationVar(&c.httpProxyConfig.Transport.ExpectContinueTimeout, "http-expect-continue-timeout", c.httpProxyConfig.Transport.ExpectContinueTimeout,
		"expect continue timeout for HTTP connections")
}

func (c *command) bindHTTPServerConfig(fs *pflag.FlagSet) {
	fs.VarP(anyflag.NewValue[forwarder.Scheme](c.httpServerConfig.Protocol, &c.httpServerConfig.Protocol,
		anyflag.EnumParser[forwarder.Scheme](forwarder.HTTPScheme, forwarder.HTTPSScheme, forwarder.HTTP2Scheme)),
		"protocol", "", "HTTP server protocol, one of http, https, h2")
	fs.StringVarP(&c.httpServerConfig.Addr, "addr", "", c.httpServerConfig.Addr, "HTTP server listen address")
	fs.StringVar(&c.httpServerConfig.CertFile, "cert-file", c.httpServerConfig.CertFile, "HTTP server TLS certificate file")
	fs.StringVar(&c.httpServerConfig.KeyFile, "key-file", c.httpServerConfig.KeyFile, "HTTP server TLS key file")
	fs.DurationVar(&c.httpServerConfig.ReadTimeout, "read-timeout", c.httpServerConfig.ReadTimeout, "HTTP server read timeout")
	fs.VarP(anyflag.NewValue[*url.Userinfo](c.httpServerConfig.BasicAuth, &c.httpServerConfig.BasicAuth, forwarder.ParseUserInfo),
		"basic-auth", "", "basic-auth in the form of `username:password`")
}

func (c *command) bindLogConfig(fs *pflag.FlagSet) {
	fs.StringVar(&c.logConfig.Level, "log-level", c.logConfig.Level, "the log level")
	fs.StringVar(&c.logConfig.FileLevel, "log-file-level", c.logConfig.FileLevel, "the log file level")
	fs.StringVar(&c.logConfig.FilePath, "log-file-path", c.logConfig.FilePath, "the log file path")
}
