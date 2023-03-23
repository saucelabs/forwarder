// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/pac"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/spf13/cobra"
)

type command struct {
	pac                 *url.URL
	httpTransportConfig *forwarder.HTTPTransportConfig
	httpServerConfig    *forwarder.HTTPServerConfig
	logConfig           *log.Config
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	config := bind.DescribeFlags(cmd.Flags())

	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)
	logger.Debugf("configuration\n%s", config)

	t := forwarder.NewHTTPTransport(c.httpTransportConfig, nil)

	script, err := forwarder.ReadURL(c.pac, t)
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

	return runctx.Run(s.Run)
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

func Command() (cmd *cobra.Command) {
	c := command{
		pac:                 &url.URL{Scheme: "file", Path: "pac.js"},
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
		httpServerConfig:    forwarder.DefaultHTTPServerConfig(),
		logConfig:           log.DefaultConfig(),
	}

	defer func() {
		fs := cmd.Flags()

		bind.PAC(fs, &c.pac)
		bind.HTTPServerConfig(fs, c.httpServerConfig, "")
		bind.LogConfig(fs, c.logConfig)
		bind.HTTPTransportConfig(fs, c.httpTransportConfig)

		bind.MarkFlagFilename(cmd, "pac", "tls-cert-file", "tls-key-file", "log-file")
	}()
	return &cobra.Command{
		Use:     "server --pac <file|url> [--protocol <http|https|h2>] [--address <host:port>] [flags]",
		Short:   "Start HTTP server that serves a PAC file",
		Long:    long,
		RunE:    c.RunE,
		Example: example,
	}
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
