package pacserver

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
	t := forwarder.NewHTTPTransport(c.httpTransportConfig, nil)

	script, err := forwarder.ReadURL(c.pac, t)
	if err != nil {
		return fmt.Errorf("read PAC file: %w", err)
	}
	if _, err := pac.NewProxyResolver(&pac.ProxyResolverConfig{Script: script}, nil); err != nil {
		return err
	}

	if f := c.logConfig.File; f != nil {
		defer f.Close()
	}
	logger := stdlog.New(c.logConfig)

	s, err := forwarder.NewHTTPServer(c.httpServerConfig, servePAC(script), logger.Named("server"))
	if err != nil {
		return err
	}

	return runctx.Run(s.Run)
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
		bind.HTTPServerConfig(fs, c.httpServerConfig, "", true)
		bind.HTTPTransportConfig(fs, c.httpTransportConfig)
		bind.LogConfig(fs, c.logConfig)

		bind.MarkFlagFilename(cmd, "pac", "cert-file", "key-file", "log-file")

		fs.SortFlags = false
	}()
	return &cobra.Command{
		Use:     "pac-server --pac <file|url> [--protocol <http|https|h2>] [--address <host:port>] [flags]",
		Short:   "Start HTTP(S) server that serves a PAC file",
		Long:    long,
		RunE:    c.RunE,
		Example: example,
	}
}

const long = `Start HTTP(S) server that serves a PAC file.
The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
The PAC file must contain FindProxyForURL or FindProxyForURLEx and must be valid.
All PAC util functions are supported (see below).
Alerts are ignored.

You can start HTTP, HTTPS or H2 (HTTPS) server.
The server may be protected by basic authentication.
If you start an HTTPS server and you don't provide a certificate, the server will generate a self-signed certificate on startup.
`

const example = `  # Start a HTTP server serving a PAC file
  forwarder pac-server --pac pac.js --protocol http --address localhost:8080

  # Start a HTTPS server serving a PAC file
  forwarder pac-server --pac pac.js --protocol https --address localhost:80443

  # Start a HTTPS server serving a PAC file with custom certificate
  forwarder pac-server --pac pac.js --protocol https --address localhost:80443 --cert-file cert.pem --key-file key.pem

  # Start a HTTPS server serving a PAC file with basic authentication
  forwarder pac-server --pac pac.js --protocol https --address localhost:80443 --basic-auth user:pass
`
