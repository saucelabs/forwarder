// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package proxy

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	martianlog "github.com/google/martian/v3/log"
	"github.com/mmatczuk/anyflag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/pac"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/spf13/cobra"
	"go.uber.org/goleak"
)

type command struct {
	promReg             *prometheus.Registry
	dnsConfig           *forwarder.DNSConfig
	httpTransportConfig *forwarder.HTTPTransportConfig
	pac                 *url.URL
	credentials         []*forwarder.HostPortUser
	httpProxyConfig     *forwarder.HTTPProxyConfig
	apiServerConfig     *forwarder.HTTPServerConfig
	logConfig           *log.Config
	goleak              bool
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	config := bind.DescribeFlags(cmd.Flags())

	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)
	logger.Debugf("Configuration\n%s", config)

	var resolver *net.Resolver
	if len(c.dnsConfig.Servers) > 0 {
		r, err := forwarder.NewResolver(c.dnsConfig, logger.Named("dns"))
		if err != nil {
			return err
		}
		resolver = r
	}
	t := forwarder.NewHTTPTransport(c.httpTransportConfig, resolver)

	var (
		script string
		pr     forwarder.PACResolver
	)
	if c.pac != nil {
		var err error
		script, err = forwarder.ReadURL(c.pac, t)
		if err != nil {
			return fmt.Errorf("read PAC file: %w", err)
		}
		pr, err = pac.NewProxyResolverPool(&pac.ProxyResolverConfig{Script: script}, nil)
		if err != nil {
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

	// Google Martian uses a global logger package.
	ml := logger.Named("proxy")
	ml.Decorate = func(format string) string {
		return strings.TrimPrefix(format, "martian: ")
	}
	martianlog.SetLogger(ml)

	p, err := forwarder.NewHTTPProxy(c.httpProxyConfig, pr, cm, t, logger.Named("proxy"))
	if err != nil {
		return err
	}
	f := runctx.Funcs{p.Run}

	if c.apiServerConfig.Addr != "" {
		h := forwarder.NewAPIHandler(c.promReg, p, script)
		a, err := forwarder.NewHTTPServer(c.apiServerConfig, h, logger.Named("api"))
		if err != nil {
			return err
		}
		f = append(f, a.Run)
	}

	if c.goleak {
		defer func() {
			if err := goleak.Find(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "goleak: %s", err)
			}
		}()
	}

	return f.Run()
}

const long = `Start HTTP(S) proxy server.
You can start HTTP or HTTPS server.
The server may be protected by basic authentication.
If you start an HTTPS server and you don't provide a certificate, the server will generate a self-signed certificate on startup.

You can start an API server that exposes metrics, health checks and other information about the proxy server.
You need to explicitly enable it by providing --api-address. 
To run on a random port, use --api-address=:0. 
It's recommended use basic authentication for the API server, especially if --proxy-localhost is enabled.

The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
All PAC util functions are supported (see below).
The supported upstream proxy types are "http", "https" and "socks5".
You can specify custom DNS servers.
They are used to resolve hostnames in PAC scripts and proxy server.

Basic authentication credentials for the upstream proxies and backend servers can be specified with --credentials flag.
`

const example = `  # Start a HTTP proxy server
  forwarder proxy --address localhost:3128

  # Start HTTP proxy with upstream proxy
  forwarder proxy --address localhost:3128 --upstream-proxy http://localhost:8081

  # Start HTTP proxy with local PAC script
  forwarder proxy --address localhost:3128 --pac ./pac.js 
  
  # Start HTTP proxy with remote PAC script
  forwarder proxy --address localhost:3128 --pac https://example.com/pac.js

  # Start HTTP proxy with custom DNS servers
  forwarder proxy --address localhost:3128 --dns-server 4.4.4.4 --dns-server 8.8.8.8

  # Start HTTP proxy and API server 
  forwarder proxy --address localhost:3128 --api-address localhost:8081 --api-basic-auth user:password

  # Start a HTTPS proxy server
  forwarder proxy --protocol https --address localhost:8443

  # Start a HTTPS server with custom certificate
  forwarder proxy --protocol https --address localhost:8443 --cert-file ./cert.pem --key-file ./cert.key

  # Start a HTTPS proxy server and require basic authentication
  forwarder proxy --protocol https --address localhost:8443 --basic-auth user:password

  # Add basic authentication header to requests to example.com and example.org on ports 80 and 443
  forwarder proxy --address localhost:3128 -c bob:bp@example.com:* -c alice:ap@example.org:80,alice:ap@example.org:443
`

const apiEndpoints = `API endpoints:
  /metrics - Prometheus metrics
  /healthz - health check
  /readyz - readiness check
  /configz - proxy configuration
  /pac - PAC file
  /version - version information
  /debug/pprof/ - pprof endpoints
`

func Command() (cmd *cobra.Command) {
	c := command{
		promReg:             prometheus.NewRegistry(),
		dnsConfig:           forwarder.DefaultDNSConfig(),
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
		httpProxyConfig:     forwarder.DefaultHTTPProxyConfig(),
		apiServerConfig:     forwarder.DefaultHTTPServerConfig(),
		logConfig:           log.DefaultConfig(),
	}
	c.httpProxyConfig.PromRegistry = c.promReg
	c.apiServerConfig.Addr = "localhost:10000"

	defer func() {
		fs := cmd.Flags()
		bind.HTTPProxyConfig(fs, c.httpProxyConfig, c.logConfig)
		fs.VarP(anyflag.NewSliceValueWithRedact[*forwarder.HostPortUser](c.credentials, &c.credentials, forwarder.ParseHostPortUser, forwarder.RedactHostPortUser),
			"credentials", "c",
			"site or upstream proxy basic authentication credentials in the form of `username:password@host:port`, "+
				"host and port can be set to \"*\" to match all (can be specified multiple times)")
		bind.PAC(fs, &c.pac)
		bind.DNSConfig(fs, c.dnsConfig)
		bind.HTTPServerConfig(fs, c.apiServerConfig, "api")
		bind.HTTPTransportConfig(fs, c.httpTransportConfig)

		bind.MarkFlagFilename(cmd, "cert-file", "key-file", "pac")
		cmd.MarkFlagsMutuallyExclusive("upstream-proxy", "pac")

		fs.BoolVar(&c.goleak, "goleak", false, "enable goleak")
		bind.MarkFlagHidden(cmd, "goleak")

		fs.SortFlags = false
	}()
	return &cobra.Command{
		Use:     "proxy [--protocol <http|https|h2>] [--address <host:port>] [--upstream-proxy <url>] [--pac <file|url>] [--credentials <username:password@host:port>]... [flags]",
		Short:   "Start HTTP(S) proxy",
		Long:    long,
		Example: example + "\n" + apiEndpoints,
		RunE:    c.RunE,
	}
}
