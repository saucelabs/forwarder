package forwarder

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder/middleware"
	"go.uber.org/atomic"
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

type HTTPServerConfig struct {
	Protocol    Scheme        `json:"protocol"`
	Addr        string        `json:"addr"`
	CertFile    string        `json:"cert_file"`
	KeyFile     string        `json:"key_file"`
	ReadTimeout time.Duration `json:"read_timeout"`

	PromNamespace   string                `json:"prom_namespace"`
	PromRegistry    prometheus.Registerer `json:"prom_registry"`
	BasicAuthHeader string                `json:"basic_auth_header"`
	BasicAuth       *url.Userinfo         `json:"basic_auth"`
}

func DefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Protocol:        HTTPScheme,
		Addr:            "0.0.0.0:8080",
		ReadTimeout:     5 * time.Second,
		BasicAuthHeader: middleware.AuthorizationHeader,
	}
}

type HTTPServer struct {
	config *HTTPServerConfig
	log    Logger
	srv    *http.Server
	addr   atomic.String

	Listener net.Listener
}

func NewHTTPServer(cfg *HTTPServerConfig, h http.Handler, log Logger) (*HTTPServer, error) {
	hs := &HTTPServer{
		config: cfg,
		log:    log,
		srv: &http.Server{
			Addr:        cfg.Addr,
			Handler:     withMiddleware(cfg, h),
			ReadTimeout: cfg.ReadTimeout,
		},
	}

	switch hs.config.Protocol {
	case HTTP2Scheme:
		if err := hs.configureHTTP2(); err != nil {
			return nil, err
		}
	case HTTPSScheme:
		if err := hs.configureHTTPS(); err != nil {
			return nil, err
		}
	case HTTPScheme:
		// do nothing
	}

	return hs, nil
}

func withMiddleware(cfg *HTTPServerConfig, h http.Handler) http.Handler {
	// Prometheus middleware must be the first one to be executed to collect metrics for all other middlewares.
	if cfg.PromRegistry != nil {
		h = middleware.NewPrometheus(cfg.PromRegistry, cfg.PromNamespace).Wrap(h)
	}

	if cfg.BasicAuth != nil {
		p, _ := cfg.BasicAuth.Password()
		h = middleware.NewBasicAuth(cfg.BasicAuthHeader).Wrap(h, cfg.BasicAuth.Username(), p)
	}
	return h
}

//nolint:gosec // allow RSA keys
func (hs *HTTPServer) configureHTTPS() error {
	if hs.config.CertFile != "" {
		if _, err := os.Stat(hs.config.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("cannot find SSL cert_file at %q", hs.config.CertFile)
		}
	}
	if hs.config.KeyFile != "" {
		if _, err := os.Stat(hs.config.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("cannot find SSL key_file at %q", hs.config.KeyFile)
		}
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	hs.srv.TLSConfig = tlsCfg
	hs.srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

	return nil
}

func (hs *HTTPServer) configureHTTP2() error {
	if hs.config.CertFile != "" {
		if _, err := os.Stat(hs.config.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("cannot find SSL cert_file at %q", hs.config.CertFile)
		}
	}
	if hs.config.KeyFile != "" {
		if _, err := os.Stat(hs.config.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("cannot find SSL key_file at %q", hs.config.KeyFile)
		}
	}

	tlsCfg := &tls.Config{
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

	hs.srv.TLSConfig = tlsCfg

	return nil
}

func (hs *HTTPServer) Run(ctx context.Context) error {
	listener, err := hs.listener()
	if err != nil {
		return err
	}
	hs.addr.Store(listener.Addr().String())
	hs.log.Infof("HTTP server listen address=%s protocol=%s", listener.Addr(), hs.config.Protocol)

	var wg sync.WaitGroup
	wg.Add(1)

	// handle http shutdown on server context done
	go func() {
		defer wg.Done()

		<-ctx.Done()
		if err := hs.srv.Shutdown(context.Background()); err != nil {
			hs.log.Errorf("Failed to shutdown server error=%s", err)
		}
	}()

	switch hs.config.Protocol {
	case HTTPScheme:
		if err := hs.srv.Serve(listener); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				hs.log.Debugf("Server was shutdown gracefully")
				return nil
			}
			return err
		}
	case HTTP2Scheme, HTTPSScheme:
		if hs.config.CertFile == "" {
			hs.log.Infof("No SSL certificate provided, using self-signed certificate")
			cert, err := RSASelfSignedCert().Gen()
			if err != nil {
				return err
			}
			hs.srv.TLSConfig.Certificates = []tls.Certificate{cert}
		}
		if err := hs.srv.ServeTLS(listener, hs.config.CertFile, hs.config.KeyFile); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				hs.log.Debugf("Server was shutdown gracefully")
				return nil
			}
			return err
		}
	default:
		return fmt.Errorf("unknown protocol %q", hs.config.Protocol)
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
		hs.log.Errorf("Invalid protocol", "protocol=%s", hs.config.Protocol)
		return nil, fmt.Errorf("invalid protocol %q", hs.config.Protocol)
	}
}

// Addr returns the address the server is listening on or an empty string if the server is not running.
func (hs *HTTPServer) Addr() string {
	return hs.addr.Load()
}
