// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
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
	"time"

	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/middleware"
)

type Scheme string

const (
	HTTPScheme  Scheme = "http"
	HTTPSScheme Scheme = "https"
	HTTP2Scheme Scheme = "h2"
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
	ListenerConfig
	TLSServerConfig
	shutdownConfig
	PromConfig

	Protocol          Scheme
	IdleTimeout       time.Duration
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	LogHTTPMode       httplog.Mode
	BasicAuth         *url.Userinfo
}

func DefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		ListenerConfig:    *DefaultListenerConfig(":8080"),
		Protocol:          HTTPScheme,
		IdleTimeout:       1 * time.Hour,
		ReadHeaderTimeout: 1 * time.Minute,
		shutdownConfig:    defaultShutdownConfig(),
	}
}

func (c *HTTPServerConfig) Validate() error {
	if err := validatedUserInfo(c.BasicAuth); err != nil {
		return fmt.Errorf("basic_auth: %w", err)
	}
	return nil
}

type HTTPServer struct {
	config   HTTPServerConfig
	log      log.StructuredLogger
	srv      *http.Server
	listener net.Listener
}

// NewHTTPServer creates a new HTTP server.
// It is the caller's responsibility to call Close on the returned server.
func NewHTTPServer(cfg *HTTPServerConfig, h http.Handler, log log.StructuredLogger) (*HTTPServer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	hs := &HTTPServer{
		config: *cfg,
		log:    log,
		srv: &http.Server{
			Handler:           withMiddleware(cfg, log, h),
			IdleTimeout:       cfg.IdleTimeout,
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

	l, err := hs.listen()
	if err != nil {
		return nil, err
	}
	hs.listener = l

	hs.log.Info("HTTP server listen", "address", l.Addr().String(), "protocol", hs.config.Protocol)

	return hs, nil
}

func withMiddleware(cfg *HTTPServerConfig, log log.StructuredLogger, h http.Handler) http.Handler {
	// Note that the order of execution is reversed.
	if cfg.BasicAuth != nil {
		p, _ := cfg.BasicAuth.Password()
		h = middleware.NewBasicAuth().Wrap(h, cfg.BasicAuth.Username(), p)
	}

	// Logger middleware must immediately follow the Prometheus middleware because it uses the Prometheus delegator.
	if cfg.LogHTTPMode != httplog.None {
		h = httplog.NewStructuredLogger(log.Info, cfg.LogHTTPMode).LogFunc().Wrap(h)
	}

	// Prometheus middleware must be the first one to be executed to collect metrics for all other middlewares.
	h = middleware.NewPrometheus(cfg.PromRegistry, cfg.PromNamespace).Wrap(h)

	return h
}

func (hs *HTTPServer) configureHTTPS() error {
	if hs.config.CertFile == "" && hs.config.KeyFile == "" {
		hs.log.Info("no TLS certificate provided, using self-signed certificate")
	} else {
		hs.log.Debug("loading TLS certificate from %s and %s", hs.config.CertFile, hs.config.KeyFile)
	}

	hs.srv.TLSConfig = httpsTLSConfigTemplate()
	hs.srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

	return hs.config.ConfigureTLSConfig(hs.srv.TLSConfig)
}

func (hs *HTTPServer) configureHTTP2() error {
	if hs.config.CertFile == "" && hs.config.KeyFile == "" {
		hs.log.Info("no TLS certificate provided, using self-signed certificate")
	} else {
		hs.log.Debug("loading TLS certificate", "cert", hs.config.CertFile, "key", hs.config.KeyFile)
	}

	hs.srv.TLSConfig = h2TLSConfigTemplate()

	return hs.config.ConfigureTLSConfig(hs.srv.TLSConfig)
}

func (hs *HTTPServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		ctx, cancel := shutdownContext(hs.config.shutdownConfig)
		defer cancel()

		if err := hs.srv.Shutdown(ctx); err != nil {
			hs.log.Debug("failed to gracefully shutdown server", "error", err)
			if err := hs.srv.Close(); err != nil {
				hs.log.Debug("failed to close server", "error", err)
			}
		}
	}()

	var srvErr error
	switch hs.config.Protocol {
	case HTTPScheme:
		srvErr = hs.srv.Serve(hs.listener)
	case HTTP2Scheme, HTTPSScheme:
		srvErr = hs.srv.ServeTLS(hs.listener, "", "")
	default:
		return fmt.Errorf("invalid protocol %q", hs.config.Protocol)
	}
	if srvErr != nil {
		if errors.Is(srvErr, http.ErrServerClosed) {
			hs.log.Debug("server was shutdown gracefully")
			srvErr = nil
		}
		return srvErr
	}

	wg.Wait()
	return nil
}

func (hs *HTTPServer) listen() (net.Listener, error) {
	switch hs.config.Protocol {
	case HTTPScheme, HTTPSScheme, HTTP2Scheme:
		l := Listener{
			ListenerConfig: hs.config.ListenerConfig,
			PromConfig:     hs.config.PromConfig,
		}
		if err := l.Listen(); err != nil {
			return nil, fmt.Errorf("failed to open listener on address %s: %w", hs.srv.Addr, err)
		}
		return &l, nil
	default:
		return nil, fmt.Errorf("invalid protocol %q", hs.config.Protocol)
	}
}

// Addr returns the address the server is listening on.
func (hs *HTTPServer) Addr() string {
	return hs.listener.Addr().String()
}

func (hs *HTTPServer) Close() error {
	return hs.listener.Close()
}
