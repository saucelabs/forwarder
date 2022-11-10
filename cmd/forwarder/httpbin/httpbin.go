package httpbin

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/httpbin"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type command struct {
	httpServerConfig *forwarder.HTTPServerConfig
	apiServerConfig  *forwarder.HTTPServerConfig
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

	r := prometheus.NewRegistry()
	a, err := forwarder.NewHTTPServer(c.apiServerConfig, forwarder.NewAPIHandler(r, s, ""), logger.Named("api"))
	if err != nil {
		return err
	}

	return c.runHTTPServers(s, a)
}

func (c *command) runHTTPServers(servers ...*forwarder.HTTPServer) error {
	var eg *errgroup.Group
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	eg, ctx = errgroup.WithContext(ctx)
	for _, s := range servers {
		s := s
		eg.Go(func() error { return s.Run(ctx) })
	}
	return eg.Wait()
}

func Command() (cmd *cobra.Command) {
	c := command{
		httpServerConfig: forwarder.DefaultHTTPServerConfig(),
		apiServerConfig:  forwarder.DefaultHTTPServerConfig(),
		logConfig:        log.DefaultConfig(),
	}
	c.apiServerConfig.Addr = ":10000"

	defer func() {
		fs := cmd.Flags()
		bind.HTTPServerConfig(fs, c.httpServerConfig, "")
		bind.HTTPServerConfig(fs, c.apiServerConfig, "api")
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
