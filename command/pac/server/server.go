// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/pac"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/osdns"
	"github.com/spf13/cobra"
)

type command struct {
	pac                 *url.URL
	dnsConfig           *osdns.Config
	httpTransportConfig *forwarder.HTTPTransportConfig
	httpServerConfig    *forwarder.HTTPServerConfig
	logConfig           *log.Config
}

func (c *command) runE(cmd *cobra.Command, _ []string) (cmdErr error) {
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

	{
		var (
			cfg []byte
			err error
		)

		d := cobrautil.FlagsDescriber{
			Format: cobrautil.Plain,
		}
		cfg, err = d.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Infof("configuration\n%s", cfg)

		d.ShowNotChanged = true
		cfg, err = d.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Debugf("all configuration\n%s\n\n", cfg)
	}

	if len(c.dnsConfig.Servers) > 0 {
		s := strings.ReplaceAll(fmt.Sprintf("%s", c.dnsConfig.Servers), " ", ", ")
		logger.Named("dns").Infof("using DNS servers %v", s)
		if err := osdns.Configure(c.dnsConfig); err != nil {
			return fmt.Errorf("configure DNS: %w", err)
		}
	}

	t, err := forwarder.NewHTTPTransport(c.httpTransportConfig)
	if err != nil {
		return err
	}

	script, err := forwarder.ReadURLString(c.pac, t)
	if err != nil {
		return fmt.Errorf("read PAC file: %w", err)
	}
	if err := validatePACScript(script); err != nil {
		return err
	}

	s, err := forwarder.NewHTTPServer(c.httpServerConfig, servePAC(script), logger.Named("server"))
	if err != nil {
		return err
	}
	defer s.Close()

	return runctx.NewGroup(s.Run).Run()
}

func validatePACScript(script string) error {
	pr, err := pac.NewProxyResolver(&pac.ProxyResolverConfig{Script: script}, nil)
	if err != nil {
		return err
	}
	_, err = pr.FindProxyForURL(&url.URL{Scheme: "https", Host: "saucelabs.com"}, "")
	return err
}

func servePAC(script string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.Write([]byte(script))
	})
}

func Command() *cobra.Command {
	c := command{
		pac:                 &url.URL{Scheme: "file", Path: "pac.js"},
		dnsConfig:           osdns.DefaultConfig(),
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
		httpServerConfig:    forwarder.DefaultHTTPServerConfig(),
		logConfig:           log.DefaultConfig(),
	}

	cmd := &cobra.Command{
		Use:     "server --pac <file|url> [--protocol <http|https|h2>] [--address <host:port>] [flags]",
		Short:   "Start HTTP server that serves a PAC file",
		Long:    long,
		RunE:    c.runE,
		Example: example,
	}

	fs := cmd.Flags()
	bind.PAC(fs, &c.pac)
	bind.DNSConfig(fs, c.dnsConfig)
	bind.HTTPServerConfig(fs, c.httpServerConfig, "")
	bind.HTTPTransportConfig(fs, c.httpTransportConfig)
	bind.HTTPLogConfig(fs, []bind.NamedParam[httplog.Mode]{
		{Name: "server", Param: &c.httpServerConfig.LogHTTPMode},
	})
	bind.LogConfig(fs, c.logConfig)

	bind.AutoMarkFlagFilename(cmd)

	return cmd
}

const long = `Start HTTP server that serves a PAC file.
You can start HTTP, HTTPS or H2 (HTTPS) server.
The server may be protected by basic authentication.
If you start an HTTPS server and you don't provide a certificate,
the server will generate a self-signed certificate on startup.

The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
The PAC file must contain FindProxyForURL or FindProxyForURLEx and must be valid.
Alerts are ignored.
`

const example = `  # HTTP server with basic authentication
  forwarder pac server --pac pac.js --basic-auth user:pass

  # HTTPS server with self-signed certificate
  forwarder pac server --pac pac.js --protocol https --address localhost:80443

  # HTTPS server with custom certificate
  forwarder pac server --pac pac.js --protocol https --address localhost:80443 --tls-cert-file cert.pem --tls-key-file key.pem
`
