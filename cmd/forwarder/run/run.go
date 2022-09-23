// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package run

import (
	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder"
	"github.com/spf13/cobra"
	"net/url"
)

type command struct {
	proxyConfig       forwarder.ProxyConfig
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

	p, err := forwarder.NewProxy(&c.proxyConfig, newLogger(c.logConfig))
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
		logConfig: defaultLogConfig(),
	}
	defer func() {
		fs := cmd.Flags()

		fs.VarP(anyflag.NewValue[*url.URL](&url.URL{Scheme: "http", Host: "localhost:8080"}, &c.proxyConfig.LocalProxyURI, parseProxyURI), "local-proxy-uri", "l", "sets local proxy URI")
		fs.VarP(anyflag.NewValue[*url.Userinfo](nil, &c.localProxyAuth, parseBasicAuth), "local-proxy-auth", "", "sets local proxy basic auth in the form of username:password")
		fs.VarP(anyflag.NewValue[*url.URL](nil, &c.proxyConfig.UpstreamProxyURI, parseProxyURI), "upstream-proxy-uri", "u", "sets upstream proxy URI")
		fs.VarP(anyflag.NewValue[*url.Userinfo](nil, &c.upstreamProxyAuth, parseBasicAuth), "upstream-proxy-auth", "", "sets upstream proxy basic auth in the form of username:password")
		fs.VarP(anyflag.NewSliceValue[*url.URL](nil, &c.proxyConfig.DNSURIs, parseDNSURI), "dns-uri", "n", "sets dns URI")
		fs.VarP(anyflag.NewValue[*url.URL](nil, &c.proxyConfig.PACURI, url.ParseRequestURI), "pac-uri", "p", "sets URI to PAC content, or directly, the PAC content")
		fs.StringSliceVarP(&c.proxyConfig.PACProxiesCredentials, "pac-proxies-credentials", "d", nil, "sets PAC proxies credentials using standard URI format")
		fs.StringSliceVar(&c.proxyConfig.SiteCredentials, "site-credentials", nil, "sets target site credentials")
		fs.BoolVarP(&c.proxyConfig.ProxyLocalhost, "proxy-localhost", "t", false, "if set, will proxy localhost requests to an upstream proxy - if any")

		fs.StringVar(&c.logConfig.Level, "log-level", c.logConfig.Level, "sets the log level")
		fs.StringVar(&c.logConfig.FileLevel, "log-file-level", c.logConfig.FileLevel, "sets the log file level")
		fs.StringVar(&c.logConfig.FilePath, "log-file-path", c.logConfig.FilePath, "sets the log file path")
	}()
	return &cobra.Command{
		Use:     "run",
		Short:   "Start the proxy",
		Long:    long,
		Example: example,
		RunE:    c.RunE,
	}
}
