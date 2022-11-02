package httpbin

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/httpbin"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/spf13/cobra"
)

type command struct {
	httpServerConfig *forwarder.HTTPServerConfig
	logConfig        *log.Config
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)

	s, err := forwarder.NewHTTPServer(c.httpServerConfig, httpbin.Handler(), logger.Named("server"))
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return s.Run(ctx)
}

func Command() (cmd *cobra.Command) {
	c := command{
		httpServerConfig: forwarder.DefaultHTTPServerConfig(),
		logConfig:        log.DefaultConfig(),
	}

	defer func() {
		fs := cmd.Flags()
		bind.HTTPServerConfig(fs, c.httpServerConfig, "")
		bind.LogConfig(fs, c.logConfig)
		bind.MarkFlagFilename(cmd, "cert-file", "key-file", "log-file")
	}()
	return &cobra.Command{
		Use:    "httpbin [--protocol <http|https|h2>] [--address <host:port>] [flags]",
		Short:  "Start HTTP(S) server that serves httpbin.org API",
		RunE:   c.RunE,
		Hidden: true,
	}
}
