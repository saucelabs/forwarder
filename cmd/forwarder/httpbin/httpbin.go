// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpbin

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/internal/version"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/httpbin"
	"github.com/spf13/cobra"
)

type command struct {
	httpServerConfig *forwarder.HTTPServerConfig
	apiServerConfig  *forwarder.HTTPServerConfig
	logConfig        *log.Config
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
			cfgStr string
			err    error
		)

		d := cobrautil.FlagsDescriber{
			Format: cobrautil.Plain,
		}
		cfgStr, err = d.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Infof("configuration\n%s", cfgStr)

		d.ShowNotChanged = true
		cfgStr, err = d.DescribeFlags(cmd.Flags())
		if err != nil {
			return err
		}
		logger.Debugf("all configuration\n%s\n\n", cfgStr)
	}

	s, err := forwarder.NewHTTPServer(c.httpServerConfig, httpbin.Handler(), logger.Named("server"))
	if err != nil {
		return err
	}
	defer s.Close()

	r := prometheus.NewRegistry()
	a, err := forwarder.NewHTTPServer(c.apiServerConfig, forwarder.NewAPIHandler("HTTPBin "+version.Version, r, nil), logger.Named("api"))
	if err != nil {
		return err
	}
	defer a.Close()

	return runctx.NewGroup(s.Run, a.Run).Run()
}

func Command() *cobra.Command {
	c := command{
		httpServerConfig: forwarder.DefaultHTTPServerConfig(),
		apiServerConfig:  forwarder.DefaultHTTPServerConfig(),
		logConfig:        log.DefaultConfig(),
	}
	c.apiServerConfig.Addr = "localhost:10000"

	cmd := &cobra.Command{
		Use:    "httpbin [--protocol <http|https|h2>] [--address <host:port>] [flags]",
		Short:  "Start HTTP(S) server that serves httpbin.org API",
		RunE:   c.runE,
		Hidden: true,
	}

	fs := cmd.Flags()
	bind.HTTPServerConfig(fs, c.httpServerConfig, "")
	bind.HTTPServerConfig(fs, c.apiServerConfig, "api", forwarder.HTTPScheme)
	bind.HTTPLogConfig(fs, []bind.NamedParam[httplog.Mode]{
		{Name: "api", Param: &c.apiServerConfig.LogHTTPMode},
		{Name: "server", Param: &c.httpServerConfig.LogHTTPMode},
	})
	bind.LogConfig(fs, c.logConfig)

	bind.AutoMarkFlagFilename(cmd)

	return cmd
}
