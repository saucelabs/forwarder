// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ready

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/saucelabs/forwarder"
	"github.com/spf13/cobra"
)

type Config struct {
	APIAddress    string
	APIUnixSocket string
	Endpoint      string
	Timeout       time.Duration
}

func DefaultConfig() Config {
	return Config{
		APIAddress:    "localhost:10000",
		APIUnixSocket: forwarder.APIUnixSocket,
		Endpoint:      "/readyz",
		Timeout:       2 * time.Second,
	}
}

type command struct {
	Config
	apiAddr string
}

func (c *command) runE(cmd *cobra.Command, _ []string) error {
	var (
		addr string
		tr   http.Transport
	)
	if socketPath, ok := strings.CutPrefix(c.apiAddr, "unix:"); ok {
		addr = c.apiAddr
		tr.DialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}
	} else {
		host, port, err := net.SplitHostPort(c.apiAddr)
		if err != nil {
			return err
		}
		if host == "" {
			host = "localhost"
		}
		addr = net.JoinHostPort(host, port)
	}

	httpc := http.Client{
		Transport: &tr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, fmt.Sprintf("http://%s%s", addr, c.Endpoint), http.NoBody)
	if err != nil {
		return err
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
		if _, err := cmd.ErrOrStderr().Write(b); err != nil {
			return err
		}

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func Command() *cobra.Command {
	return CommandWithConfig(DefaultConfig())
}

func CommandWithConfig(cfg Config) *cobra.Command {
	c := command{
		Config:  cfg,
		apiAddr: cfg.APIAddress,
	}
	if os.Getenv("PLATFORM") == "container" {
		c.apiAddr = "unix:" + cfg.APIUnixSocket
	}

	cmd := &cobra.Command{
		Use:   "ready [--api-address <host:port>] [flags]",
		Short: "Readiness probe for the Forwarder",
		Long:  long,
		RunE:  c.runE,
	}

	fs := cmd.Flags()
	bindAPIAddr(fs, &c.apiAddr)

	return cmd
}

const long = `Readiness probe for the Forwarder.
This is equivalent to calling /readyz endpoint on the Forwarder API server.`
