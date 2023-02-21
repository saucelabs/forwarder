// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package paceval

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/pac"
	"github.com/spf13/cobra"
)

type command struct {
	pac                 *url.URL
	dnsConfig           *forwarder.DNSConfig
	httpTransportConfig *forwarder.HTTPTransportConfig
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	var resolver *net.Resolver
	if len(c.dnsConfig.Servers) > 0 {
		r, err := forwarder.NewResolver(c.dnsConfig, log.NopLogger)
		if err != nil {
			return err
		}
		resolver = r
	}
	t := forwarder.NewHTTPTransport(c.httpTransportConfig, resolver)

	script, err := forwarder.ReadURL(c.pac, t)
	if err != nil {
		return fmt.Errorf("read PAC file: %w", err)
	}
	cfg := pac.ProxyResolverConfig{
		Script:    script,
		AlertSink: os.Stderr,
	}
	pr, err := pac.NewProxyResolver(&cfg, resolver)
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

func Command() (cmd *cobra.Command) {
	c := command{
		pac:                 &url.URL{Scheme: "file", Path: "pac.js"},
		dnsConfig:           forwarder.DefaultDNSConfig(),
		httpTransportConfig: forwarder.DefaultHTTPTransportConfig(),
	}

	defer func() {
		fs := cmd.Flags()

		bind.PAC(fs, &c.pac)
		bind.DNSConfig(fs, c.dnsConfig)
		bind.HTTPTransportConfig(fs, c.httpTransportConfig)

		bind.MarkFlagFilename(cmd, "pac")

		fs.SortFlags = false
	}()
	return &cobra.Command{
		Use:     "pac-eval --pac <file|url> [flags] <url>...",
		Short:   "Evaluate a PAC file for given URLs",
		Long:    long,
		RunE:    c.RunE,
		Example: example,
	}
}

const long = `Evaluate a PAC file for given URLs.
The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
The URLs to evaluate are passed as arguments. The output is a list of proxy strings, one per URL.
The PAC file must contain FindProxyForURL or FindProxyForURLEx and must be valid.
All PAC util functions are supported (see below).
Alerts are written to stderr.
`

const example = `  # Evaluate a PAC file for a URL
  forwarder pac-eval --pac pac.js https://www.google.com

  # Evaluate a PAC file for multiple URLs
  forwarder pac-eval --pac pac.js https://www.google.com https://www.facebook.com

  # Evaluate a PAC file for multiple URLs using a PAC file from stdin
  cat pac.js | forwarder pac-eval --pac - https://www.google.com https://www.facebook.com

  # Evaluate a PAC file for multiple URLs using a PAC file from a URL
  forwarder pac-eval --pac https://example.com/pac.js https://www.google.com https://www.facebook.com
`
