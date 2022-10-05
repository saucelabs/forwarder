// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package run

import (
	"net"
	"net/url"

	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder"
	"github.com/spf13/cobra"
)

type command struct {
	dnsConfig         *forwarder.DNSConfig
	proxyConfig       *forwarder.ProxyConfig
	localProxyAuth    *url.Userinfo
	upstreamProxyAuth *url.Userinfo
	logConfig         logConfig
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	if c.localProxyAuth != nil && c.proxyConfig.LocalProxyURI != nil {
		c.proxyConfig.LocalProxyURI.User = c.localProxyAuth
	}
	if c.upstreamProxyAuth != nil && c.proxyConfig.UpstreamProxyURI != nil {
		c.proxyConfig.UpstreamProxyURI.User = c.upstreamProxyAuth
	}

	var resolver *net.Resolver
	if len(c.dnsConfig.Servers) > 0 {
		r, err := forwarder.NewResolver(c.dnsConfig, newLogger(c.logConfig, "dns"))
		if err != nil {
			return err
		}
		resolver = r
	}

	p, err := forwarder.NewProxy(c.proxyConfig, resolver, newLogger(c.logConfig, "proxy"))
	if err != nil {
		return err
	}

	return p.Run()
}

const long = `Start the proxy. Proxy can be protected with basic auth.
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
		dnsConfig:   forwarder.DefaultDNSConfig(),
		proxyConfig: forwarder.DefaultProxyConfig(),
		logConfig:   defaultLogConfig(),
	}
	defer func() {
		fs := cmd.Flags()

		fs.VarP(anyflag.NewSliceValue[*url.URL](nil, &c.dnsConfig.Servers, forwarder.ParseDNSURI),
			"dns-server", "n", "DNS server, ex. -n udp://1.1.1.1:53 (can be specified multiple times)")
		fs.DurationVar(&c.dnsConfig.Timeout, "dns-timeout", c.dnsConfig.Timeout, "timeout for DNS queries if DNS server is specified")

		fs.VarP(anyflag.NewValue[*url.URL](c.proxyConfig.LocalProxyURI, &c.proxyConfig.LocalProxyURI, forwarder.ParseProxyURI),
			"local-proxy-uri", "l", "local proxy URI")
		fs.VarP(anyflag.NewValue[*url.Userinfo](c.localProxyAuth, &c.localProxyAuth, forwarder.ParseUserInfo),
			"local-proxy-auth", "", "local proxy basic auth in the form of username:password")
		fs.VarP(anyflag.NewValue[*url.URL](c.proxyConfig.UpstreamProxyURI, &c.proxyConfig.UpstreamProxyURI, forwarder.ParseProxyURI),
			"upstream-proxy-uri", "u", "upstream proxy URI")
		fs.VarP(anyflag.NewValue[*url.Userinfo](c.upstreamProxyAuth, &c.upstreamProxyAuth, forwarder.ParseUserInfo),
			"upstream-proxy-auth", "", "upstream proxy basic auth in the form of username:password")
		fs.VarP(anyflag.NewValue[*url.URL](c.proxyConfig.PACURI, &c.proxyConfig.PACURI, url.ParseRequestURI),
			"pac-uri", "p", "URI to PAC content, or directly, the PAC content")
		fs.StringSliceVarP(&c.proxyConfig.PACProxiesCredentials, "pac-proxies-credentials", "d", c.proxyConfig.PACProxiesCredentials,
			"PAC proxies credentials using standard URI format")
		fs.StringSliceVar(&c.proxyConfig.SiteCredentials, "site-credentials", c.proxyConfig.SiteCredentials,
			"target site credentials")
		fs.BoolVarP(&c.proxyConfig.ProxyLocalhost, "proxy-localhost", "t", c.proxyConfig.ProxyLocalhost,
			"if set, will proxy localhost requests to an upstream proxy")

		fs.DurationVar(&c.proxyConfig.HTTP.DialTimeout, "http-dial-timeout", c.proxyConfig.HTTP.DialTimeout,
			"dial timeout for HTTP connections")
		fs.DurationVar(&c.proxyConfig.HTTP.KeepAlive, "http-keep-alive", c.proxyConfig.HTTP.KeepAlive,
			"keep alive interval for HTTP connections")
		fs.DurationVar(&c.proxyConfig.HTTP.TLSHandshakeTimeout, "http-tls-handshake-timeout", c.proxyConfig.HTTP.TLSHandshakeTimeout,
			"TLS handshake timeout for HTTP connections")
		fs.IntVar(&c.proxyConfig.HTTP.MaxIdleConns, "http-max-idle-conns", c.proxyConfig.HTTP.MaxIdleConns,
			"maximum number of idle connections for HTTP connections")
		fs.IntVar(&c.proxyConfig.HTTP.MaxIdleConnsPerHost, "http-max-idle-conns-per-host", c.proxyConfig.HTTP.MaxIdleConnsPerHost,
			"maximum number of idle connections per host for HTTP connections")
		fs.IntVar(&c.proxyConfig.HTTP.MaxConnsPerHost, "http-max-conns-per-host", c.proxyConfig.HTTP.MaxConnsPerHost,
			"maximum number of connections per host for HTTP connections")
		fs.DurationVar(&c.proxyConfig.HTTP.IdleConnTimeout, "http-idle-conn-timeout", c.proxyConfig.HTTP.IdleConnTimeout,
			"idle connection timeout for HTTP connections")
		fs.DurationVar(&c.proxyConfig.HTTP.ResponseHeaderTimeout, "http-response-header-timeout", c.proxyConfig.HTTP.ResponseHeaderTimeout,
			"response header timeout for HTTP connections")
		fs.DurationVar(&c.proxyConfig.HTTP.ExpectContinueTimeout, "http-expect-continue-timeout", c.proxyConfig.HTTP.ExpectContinueTimeout,
			"expect continue timeout for HTTP connections")

		fs.StringVar(&c.logConfig.Level, "log-level", c.logConfig.Level, "the log level")
		fs.StringVar(&c.logConfig.FileLevel, "log-file-level", c.logConfig.FileLevel, "the log file level")
		fs.StringVar(&c.logConfig.FilePath, "log-file-path", c.logConfig.FilePath, "the log file path")

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
