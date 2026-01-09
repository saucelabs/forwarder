// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	ts "github.com/saucelabs/forwarder/internal/martian/h2/testing"
	tspb "github.com/saucelabs/forwarder/internal/martian/h2/testservice"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/slog"
	"github.com/saucelabs/forwarder/runctx"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // register gzip encoding
)

type command struct {
	addr            string
	plainText       bool
	tlsServerConfig *forwarder.TLSServerConfig
	logConfig       *log.Config
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

	{
		var gs *grpc.Server
		if c.plainText {
			gs = grpc.NewServer()
		} else {
			tlsCfg := new(tls.Config)
			if err := c.tlsServerConfig.ConfigureTLSConfig(tlsCfg); err != nil {
				return err
			}
			gs = grpc.NewServer(
				grpc.Creds(credentials.NewServerTLSFromCert(&tlsCfg.Certificates[0])),
			)
		}

		l, err := net.Listen("tcp", c.addr)
		if err != nil {
			return err
		}
		defer l.Close()

		tspb.RegisterTestServiceServer(gs, &ts.Server{})
		defer gs.Stop()

		g.Add(func(ctx context.Context) error {
			logger.Named("grpc").Info("server listen", "address", l.Addr())
			go func() {
				<-ctx.Done()
				gs.GracefulStop()
			}()
			return gs.Serve(l)
		})
	}

	return g.Run()
}

func Command() *cobra.Command {
	c := command{
		addr:            "localhost:1443",
		tlsServerConfig: new(forwarder.TLSServerConfig),
		logConfig:       log.DefaultConfig(),
	}

	cmd := &cobra.Command{
		Use:   "grpc [--address <host:port>] [flags]",
		Short: "Start gRPC server for testing",
		RunE:  c.runE,
	}

	fs := cmd.Flags()
	fs.StringVar(&c.addr, "address", c.addr, "<host:port>"+
		"Address to listen on. "+
		"If the host is empty, the server will listen on all available interfaces. ")
	fs.BoolVar(&c.plainText, "plain-text", c.plainText, "Run in plain-text mode i.e. without TLS.")
	bind.TLSServerConfig(fs, c.tlsServerConfig, "")
	bind.LogConfig(fs, c.logConfig)
	bind.AutoMarkFlagFilename(cmd)

	return cmd
}
