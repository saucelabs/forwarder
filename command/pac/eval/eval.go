// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package eval

import (
	"fmt"
	"net/url"
	"os"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/pac"
	"github.com/spf13/cobra"
)

type command struct {
	pac                 *url.URL
	dnsConfig           *forwarder.DNSConfig
	httpTransportConfig *forwarder.HTTPTransportConfig
}

func (c *command) runE(cmd *cobra.Command, args []string) error {
	if len(c.dnsConfig.Servers) > 0 {
		if err := c.dnsConfig.Apply(); err != nil {
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
	cfg := pac.ProxyResolverConfig{
		Script:    script,
		AlertSink: os.Stderr,
	}
	pr, err := pac.NewProxyResolver(&cfg, nil)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	for _, arg := range args {
		u, err := url.Parse(arg)
		if err != nil {
			return fmt.Errorf("parse URL: %w", err)
		}
		proxy, err := pr.FindProxyForURL(u, "")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, proxy)
	}

	return nil
}

func Command() *cobra.Command {
	c := command{
		pac:                 &url.URL{Scheme: "file", Path: "pac.js"},
		dnsConfig:           forwarder.DefaultDNSConfig(),
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
	}

	cmd := &cobra.Command{
		Use:     "eval --pac <file|url> [flags] <url>...",
		Short:   "Evaluate a PAC file for given URL (or URLs)",
		Long:    long,
		RunE:    c.runE,
		Example: example,
	}

	fs := cmd.Flags()
	bind.PAC(fs, &c.pac)
	bind.DNSConfig(fs, c.dnsConfig)
	bind.HTTPTransportConfig(fs, c.httpTransportConfig)

	bind.AutoMarkFlagFilename(cmd)

	return cmd
}

const long = `Evaluate a PAC file for given URL (or URLs).
The output is a list of proxy strings, one per URL.
The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
The PAC file must contain FindProxyForURL or FindProxyForURLEx and must be valid.
Alerts are written to stderr.
`

const example = `  # Evaluate PAC file for multiple URLs
  forwarder pac eval --pac pac.js https://www.google.com https://www.facebook.com
`
