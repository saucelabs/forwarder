// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpbin

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/internal/version"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/slog"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/httpbin"
	"github.com/saucelabs/forwarder/utils/httpx"
	"github.com/spf13/cobra"
)

type command struct {
	httpServerConfig *forwarder.HTTPServerConfig
	logConfig        *log.Config
}

func (c *command) runE(cmd *cobra.Command, _ []string) (cmdErr error) {
	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := slog.New(c.logConfig)

	defer func() {
		if err := logger.Close(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "close logger: %s\n", err)
		}
	}()
	defer func() {
		if cmdErr != nil {
			logger.Error("fatal error exiting", "error", cmdErr)
			cmd.SilenceErrors = true
		}
	}()

	{
		var cfg []byte

		//nolint:errcheck // OneLine never fails.
		cfg, _ = cobrautil.FlagsDescriber{
			Format:          cobrautil.OneLine,
			ShowChangedOnly: true,
			ShowHidden:      true,
		}.DescribeFlags(cmd.Flags())
		logger.Info("configuration: " + string(cfg))

		//nolint:errcheck // OneLine never fails.
		cfg, _ = cobrautil.FlagsDescriber{
			Format:          cobrautil.OneLine,
			ShowChangedOnly: false,
			ShowHidden:      true,
		}.DescribeFlags(cmd.Flags())
		logger.Debug("all configuration: " + string(cfg))
	}

	g := runctx.NewGroup()

	s, err := forwarder.NewHTTPServer(c.httpServerConfig, httpbin.Handler(), logger.With("name", "server"))
	if err != nil {
		return err
	}
	defer s.Close()
	g.Add(s.Run)

	g.Add(func(ctx context.Context) error {
		logger.With("name", "api").Info("HTTP server listen", "socket", forwarder.APIUnixSocket)
		r := prometheus.NewRegistry()
		h := forwarder.NewAPIHandler("HTTPBin "+version.Version, r, nil)
		return httpx.ServeUnixSocket(ctx, h, forwarder.APIUnixSocket)
	})

	return g.Run()
}

func Command() *cobra.Command {
	c := command{
		httpServerConfig: forwarder.DefaultHTTPServerConfig(),
		logConfig:        log.DefaultConfig(),
	}

	cmd := &cobra.Command{
		Use:   "httpbin [--protocol <http|https|h2>] [--address <host:port>] [flags]",
		Short: "Start HTTP(S) server that serves httpbin.org API",
		RunE:  c.runE,
	}

	fs := cmd.Flags()
	bind.HTTPServerConfig(fs, c.httpServerConfig, "")
	bind.HTTPLogConfig(fs, []bind.NamedParam[httplog.Mode]{
		{Name: "server", Param: &c.httpServerConfig.LogHTTPMode},
	})
	bind.LogConfig(fs, c.logConfig)

	bind.AutoMarkFlagFilename(cmd)

	return cmd
}
