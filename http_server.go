// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/forwarder/utils/certutil"
)

type Scheme string

const (
	HTTPScheme   Scheme = "http"
	HTTPSScheme  Scheme = "https"
	HTTP2Scheme  Scheme = "h2"
	TunnelScheme Scheme = "tunnel"
)

func (s Scheme) String() string {
	return string(s)
}

func httpsTLSConfigTemplate() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, //nolint:gosec // allow weak ciphers
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
}

func h2TLSConfigTemplate() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		NextProtos: []string{"h2", "http/1.1"},
	}
}

type HTTPServerConfig struct {
	Protocol          Scheme
	Addr              string
	CertFile          string
	KeyFile           string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	LogHTTPMode       httplog.Mode

	PromNamespace string
	PromRegistry  prometheus.Registerer
	BasicAuth     *url.Userinfo
}

func DefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Protocol:          HTTPScheme,
		Addr:              ":8080",
		ReadHeaderTimeout: 1 * time.Minute,
		LogHTTPMode:       httplog.Errors,
	}
}

func (c *HTTPServerConfig) Validate() error {
	if err := validatedUserInfo(c.BasicAuth); err != nil {
		return fmt.Errorf("basic_auth: %w", err)
	}
	return nil
}

func (c *HTTPServerConfig) loadCertificate(tlsCfg *tls.Config) error {
	var (
		cert tls.Certificate
		err  error
	)

	if c.CertFile == "" && c.KeyFile == "" {
		cert, err = certutil.RSASelfSignedCert().Gen()
	} else {
		cert, err = tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	}

	if err == nil {
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return err
}

type HTTPServer struct {
	config HTTPServerConfig
	log    log.Logger
	srv    *http.Server
	addr   atomic.Pointer[string]

	Listener net.Listener
}

func NewHTTPServer(cfg *HTTPServerConfig, h http.Handler, log log.Logger) (*HTTPServer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	hs := &HTTPServer{
		config: *cfg,
		log:    log,
		srv: &http.Server{
			Addr:              cfg.Addr,
			Handler:           withMiddleware(cfg, log, h),
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
		},
	}

	switch hs.config.Protocol {
	case HTTPScheme:
		// do nothing
	case HTTPSScheme:
		if err := hs.configureHTTPS(); err != nil {
			return nil, err
		}
	case HTTP2Scheme:
		if err := hs.configureHTTP2(); err != nil {
			return nil, err
		}
	}

	return hs, nil
}

func withMiddleware(cfg *HTTPServerConfig, log log.Logger, h http.Handler) http.Handler {
	// Note that the order of execution is reversed.
	if cfg.BasicAuth != nil {
		p, _ := cfg.BasicAuth.Password()
		h = middleware.NewBasicAuth().Wrap(h, cfg.BasicAuth.Username(), p)
	}

	// Logger middleware must immediately follow the Prometheus middleware because it uses the Prometheus delegator.
	if cfg.LogHTTPMode != httplog.None {
		h = httplog.NewLogger(log.Infof, cfg.LogHTTPMode).LogFunc().Wrap(h)
	}

	// Prometheus middleware must be the first one to be executed to collect metrics for all other middlewares.
	h = middleware.NewPrometheus(cfg.PromRegistry, cfg.PromNamespace).Wrap(h)

	return h
}

func (hs *HTTPServer) configureHTTPS() error {
	if hs.config.CertFile == "" && hs.config.KeyFile == "" {
		hs.log.Infof("no SSL certificate provided, using self-signed certificate")
	}
	tlsCfg := httpsTLSConfigTemplate()
	err := hs.config.loadCertificate(tlsCfg)
	hs.srv.TLSConfig = tlsCfg
	hs.srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	return err
}

func (hs *HTTPServer) configureHTTP2() error {
	if hs.config.CertFile == "" && hs.config.KeyFile == "" {
		hs.log.Infof("no SSL certificate provided, using self-signed certificate")
	}
	tlsCfg := h2TLSConfigTemplate()
	err := hs.config.loadCertificate(tlsCfg)
	hs.srv.TLSConfig = tlsCfg
	return err
}

func (hs *HTTPServer) Run(ctx context.Context) error {
	listener, err := hs.listener()
	if err != nil {
		return err
	}
	defer listener.Close()

	addr := listener.Addr().String()
	hs.addr.Store(&addr)
	hs.log.Infof("HTTP server listen address=%s protocol=%s", addr, hs.config.Protocol)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		if err := hs.srv.Shutdown(context.Background()); err != nil {
			hs.log.Errorf("failed to shutdown server error=%s", err)
		}
	}()

	var srvErr error
	switch hs.config.Protocol {
	case HTTPScheme:
		srvErr = hs.srv.Serve(listener)
	case HTTP2Scheme, HTTPSScheme:
		srvErr = hs.srv.ServeTLS(listener, "", "")
	default:
		return fmt.Errorf("invalid protocol %q", hs.config.Protocol)
	}
	if srvErr != nil {
		if errors.Is(srvErr, http.ErrServerClosed) {
			hs.log.Debugf("server was shutdown gracefully")
			srvErr = nil
		}
		return srvErr
	}

	wg.Wait()
	return nil
}

func (hs *HTTPServer) listener() (net.Listener, error) {
	if hs.Listener != nil {
		return hs.Listener, nil
	}

	switch hs.config.Protocol {
	case HTTPScheme, HTTPSScheme, HTTP2Scheme:
		listener, err := net.Listen("tcp", hs.srv.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to open listener on address %s: %w", hs.srv.Addr, err)
		}
		return listener, nil
	default:
		return nil, fmt.Errorf("invalid protocol %q", hs.config.Protocol)
	}
}

// Addr returns the address the server is listening on or an empty string if the server is not running.
func (hs *HTTPServer) Addr() string {
	addr := hs.addr.Load()
	if addr == nil {
		return ""
	}
	return *addr
}
