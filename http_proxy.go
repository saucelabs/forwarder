// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/saucelabs/forwarder/hostsfile"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/internal/martian"
	"github.com/saucelabs/forwarder/internal/martian/fifo"
	"github.com/saucelabs/forwarder/internal/martian/httpspec"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/forwarder/pac"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

type ProxyLocalhostMode string

const (
	DenyProxyLocalhost   ProxyLocalhostMode = "deny"
	AllowProxyLocalhost  ProxyLocalhostMode = "allow"
	DirectProxyLocalhost ProxyLocalhostMode = "direct"
)

func (m *ProxyLocalhostMode) UnmarshalText(text []byte) error {
	switch ProxyLocalhostMode(text) {
	case DenyProxyLocalhost, AllowProxyLocalhost, DirectProxyLocalhost:
		*m = ProxyLocalhostMode(text)
		return nil
	default:
		return fmt.Errorf("invalid mode: %s", text)
	}
}

func (m ProxyLocalhostMode) String() string {
	return string(m)
}

func (m ProxyLocalhostMode) isValid() bool {
	switch m {
	case DenyProxyLocalhost, AllowProxyLocalhost, DirectProxyLocalhost:
		return true
	default:
		return false
	}
}

type ProxyFunc func(*http.Request) (*url.URL, error)

// Alias all martian types to avoid exposing them.
type (
	RequestModifier         = martian.RequestModifier
	ResponseModifier        = martian.ResponseModifier
	RequestResponseModifier = martian.RequestResponseModifier
	RequestModifierFunc     = martian.RequestModifierFunc
	ResponseModifierFunc    = martian.ResponseModifierFunc

	ConnectFunc = martian.ConnectFunc
)

// ErrConnectFallback is returned by a ConnectFunc to indicate
// that the CONNECT request should be handled by martian.
var ErrConnectFallback = martian.ErrConnectFallback

type HTTPProxyConfig struct {
	HTTPServerConfig
	ExtraListeners    []NamedListenerConfig
	Name              string
	MITM              *MITMConfig
	MITMDomains       Matcher
	ProxyLocalhost    ProxyLocalhostMode
	UpstreamProxy     *url.URL
	UpstreamProxyFunc ProxyFunc
	DenyDomains       Matcher
	DirectDomains     Matcher
	RequestIDHeader   string
	RequestModifiers  []RequestModifier
	ResponseModifiers []ResponseModifier
	ConnectFunc       ConnectFunc
	ConnectTimeout    time.Duration
	PromHTTPOpts      []middleware.PrometheusOpt

	// TestingHTTPHandler uses Martian's [http.Handler] implementation
	// over [http.Server] instead of the default TCP server.
	TestingHTTPHandler bool
}

func DefaultHTTPProxyConfig() *HTTPProxyConfig {
	return &HTTPProxyConfig{
		HTTPServerConfig: HTTPServerConfig{
			ListenerConfig:    *DefaultListenerConfig(":3128"),
			Protocol:          HTTPScheme,
			IdleTimeout:       1 * time.Hour,
			ReadHeaderTimeout: 1 * time.Minute,
			shutdownConfig:    defaultShutdownConfig(),
			TLSServerConfig: TLSServerConfig{
				HandshakeTimeout: 10 * time.Second,
			},
		},
		Name:            "forwarder",
		ProxyLocalhost:  DenyProxyLocalhost,
		RequestIDHeader: "X-Request-Id",
		ConnectTimeout:  60 * time.Second, // http.Transport sets a constant 1m timeout for CONNECT requests.
	}
}

func (c *HTTPProxyConfig) Validate() error {
	if err := c.HTTPServerConfig.Validate(); err != nil {
		return err
	}
	for _, lc := range c.ExtraListeners {
		if lc.Name == "" {
			return errors.New("extra listener name is required")
		}
	}
	if c.Protocol != HTTPScheme && c.Protocol != HTTPSScheme {
		return fmt.Errorf("unsupported protocol: %s", c.Protocol)
	}
	if !c.ProxyLocalhost.isValid() {
		return fmt.Errorf("unsupported proxy_localhost: %s", c.ProxyLocalhost)
	}
	if err := validateProxyURL(c.UpstreamProxy); err != nil {
		return fmt.Errorf("upstream_proxy_uri: %w", err)
	}

	return nil
}

type HTTPProxy struct {
	config     HTTPProxyConfig
	pac        PACResolver
	creds      *CredentialsMatcher
	transport  http.RoundTripper
	log        log.Logger
	metrics    *httpProxyMetrics
	proxy      *martian.Proxy
	mitmCACert *x509.Certificate
	proxyFunc  ProxyFunc
	localhost  []string

	tlsConfig *tls.Config
	listeners []net.Listener
}

// NewHTTPProxy creates a new HTTP proxy.
// It is the caller's responsibility to call Close on the returned server.
func NewHTTPProxy(cfg *HTTPProxyConfig, pr PACResolver, cm *CredentialsMatcher, rt http.RoundTripper, log log.Logger) (*HTTPProxy, error) {
	hp, err := newHTTPProxy(cfg, pr, cm, rt, log)
	if err != nil {
		return nil, err
	}

	if hp.config.Protocol == HTTPSScheme {
		if err := hp.configureHTTPS(); err != nil {
			return nil, err
		}
	}

	lh, err := hostsfile.LocalhostAliases()
	if err != nil {
		return nil, fmt.Errorf("read localhost aliases: %w", err)
	}
	for i := range lh {
		lh[i] = strings.ToLower(lh[i])
	}
	hp.localhost = append(hp.localhost, lh...)

	ll, err := hp.listen()
	if err != nil {
		return nil, err
	}
	hp.listeners = ll

	for _, l := range hp.listeners {
		hp.log.Infof("PROXY server listen address=%s protocol=%s", l.Addr(), hp.config.Protocol)
	}

	return hp, nil
}

// NewHTTPProxyHandler is like NewHTTPProxy but returns http.Handler instead of *HTTPProxy.
func NewHTTPProxyHandler(cfg *HTTPProxyConfig, pr PACResolver, cm *CredentialsMatcher, rt http.RoundTripper, log log.Logger) (http.Handler, error) {
	hp, err := newHTTPProxy(cfg, pr, cm, rt, log)
	if err != nil {
		return nil, err
	}

	return hp.handler(), nil
}

func newHTTPProxy(cfg *HTTPProxyConfig, pr PACResolver, cm *CredentialsMatcher, rt http.RoundTripper, log log.Logger) (*HTTPProxy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.UpstreamProxy != nil && pr != nil {
		return nil, errors.New("cannot use both upstream proxy and PAC")
	}

	// If not set, use http.DefaultTransport.
	if rt == nil {
		log.Infof("HTTP transport not configured, using standard library default")
		rt = http.DefaultTransport.(*http.Transport).Clone()
	} else if tr, ok := rt.(*http.Transport); !ok {
		log.Debugf("using custom HTTP transport %T", rt)
	} else if tr.TLSClientConfig != nil && tr.TLSClientConfig.RootCAs != nil {
		log.Infof("using custom root CA certificates")
	}
	hp := &HTTPProxy{
		config:    *cfg,
		pac:       pr,
		creds:     cm,
		transport: rt,
		log:       log,
		metrics:   newHTTPProxyMetrics(cfg.PromRegistry, cfg.PromNamespace),
		localhost: []string{"localhost", "0.0.0.0", "::"},
	}

	if err := hp.configureProxy(); err != nil {
		return nil, err
	}

	return hp, nil
}

func (hp *HTTPProxy) configureHTTPS() error {
	if hp.config.CertFile == "" && hp.config.KeyFile == "" {
		hp.log.Infof("no TLS certificate provided, using self-signed certificate")
	} else {
		hp.log.Debugf("loading TLS certificate from %s and %s", hp.config.CertFile, hp.config.KeyFile)
	}

	hp.tlsConfig = httpsTLSConfigTemplate()

	return hp.config.ConfigureTLSConfig(hp.tlsConfig)
}

func (hp *HTTPProxy) configureProxy() error {
	hp.proxy = new(martian.Proxy)
	hp.proxy.AllowHTTP = true
	hp.proxy.RequestIDHeader = hp.config.RequestIDHeader
	hp.proxy.ConnectFunc = hp.config.ConnectFunc
	hp.proxy.ConnectTimeout = hp.config.ConnectTimeout
	hp.proxy.WithoutWarning = true
	hp.proxy.ErrorResponse = hp.errorResponse
	hp.proxy.IdleTimeout = hp.config.IdleTimeout
	hp.proxy.TLSHandshakeTimeout = hp.config.TLSServerConfig.HandshakeTimeout
	hp.proxy.ReadTimeout = hp.config.ReadTimeout
	hp.proxy.ReadHeaderTimeout = hp.config.ReadHeaderTimeout
	hp.proxy.WriteTimeout = hp.config.WriteTimeout

	if hp.config.MITM != nil {
		mc, err := newMartianMITMConfig(hp.config.MITM)
		if err != nil {
			return fmt.Errorf("mitm: %w", err)
		}
		if hp.config.MITM.CACertFile == "" {
			hp.log.Infof("using MITM with self-signed CA certificate, sha256 fingerprint=%x", sha256.Sum256(mc.CACert().Raw))
		} else {
			hp.log.Infof("using MITM")
		}
		registerMITMCacheMetrics(hp.config.PromRegistry, hp.config.PromNamespace+"_mitm_", mc.CacheMetrics)
		hp.mitmCACert = mc.CACert()

		hp.proxy.MITMConfig = mc

		if hp.config.MITMDomains != nil {
			hp.proxy.MITMFilter = func(req *http.Request) bool {
				return hp.config.MITMDomains.Match(req.URL.Hostname())
			}
		}
		hp.proxy.MITMTLSHandshakeTimeout = hp.config.TLSServerConfig.HandshakeTimeout
	}

	hp.proxy.RoundTripper = hp.transport
	switch {
	case hp.config.UpstreamProxyFunc != nil:
		hp.log.Infof("using external proxy function")
		hp.proxyFunc = hp.config.UpstreamProxyFunc
	case hp.config.UpstreamProxy != nil:
		u := hp.upstreamProxyURL()
		hp.log.Infof("using upstream proxy: %s", u.Redacted())
		hp.proxyFunc = http.ProxyURL(u)
	case hp.pac != nil:
		hp.log.Infof("using PAC proxy")
		hp.proxyFunc = hp.pacProxy
	default:
		hp.log.Infof("no upstream proxy specified")
	}

	if hp.config.DirectDomains != nil {
		hp.proxyFunc = hp.directDomains(hp.proxyFunc)
	}

	hp.log.Infof("localhost proxying mode=%s", hp.config.ProxyLocalhost)
	if hp.config.ProxyLocalhost == DirectProxyLocalhost {
		hp.proxyFunc = hp.directLocalhost(hp.proxyFunc)
	}
	hp.proxy.ProxyURL = hp.proxyFunc

	mw, trace := hp.middlewareStack()
	hp.proxy.RequestModifier = mw
	hp.proxy.ResponseModifier = mw
	hp.proxy.Trace = trace

	return nil
}

func (hp *HTTPProxy) upstreamProxyURL() *url.URL {
	proxyURL := new(url.URL)
	*proxyURL = *hp.config.UpstreamProxy

	if proxyURL.User == nil {
		if u := hp.creds.MatchURL(proxyURL); u != nil {
			proxyURL.User = u
		}
	}

	return proxyURL
}

func (hp *HTTPProxy) pacProxy(r *http.Request) (*url.URL, error) {
	s, err := hp.pac.FindProxyForURL(r.URL, "")
	if err != nil {
		return nil, err
	}

	p, err := pac.Proxies(s).First()
	if err != nil {
		return nil, err
	}

	proxyURL := p.URL()
	if u := hp.creds.MatchURL(proxyURL); u != nil {
		proxyURL.User = u
	}

	return proxyURL, nil
}

func (hp *HTTPProxy) middlewareStack() (martian.RequestResponseModifier, *martian.ProxyTrace) {
	var trace *martian.ProxyTrace

	// Wrap stack in a group so that we can run security checks before the httpspec modifiers.
	topg := fifo.NewGroup()
	if hp.config.BasicAuth != nil {
		hp.log.Infof("basic auth enabled")
		topg.AddRequestModifier(hp.basicAuth(hp.config.BasicAuth))
	}
	if hp.config.ProxyLocalhost == DenyProxyLocalhost {
		topg.AddRequestModifier(hp.denyLocalhost())
	}
	if hp.config.DenyDomains != nil {
		topg.AddRequestModifier(hp.denyDomains(hp.config.DenyDomains))
	}

	// stack contains the request/response modifiers in the order they are applied.
	// fg is the inner stack that is executed after the core request modifiers and before the core response modifiers.
	stack, fg := httpspec.NewStack(hp.config.Name)
	topg.AddRequestModifier(stack)
	topg.AddResponseModifier(stack)

	for _, m := range hp.config.RequestModifiers {
		fg.AddRequestModifier(m)
	}

	for _, m := range hp.config.ResponseModifiers {
		fg.AddResponseModifier(m)
	}

	if hp.config.LogHTTPMode != httplog.None {
		lf := httplog.NewLogger(hp.log.Infof, hp.config.LogHTTPMode).LogFunc()
		fg.AddResponseModifier(lf)
	}

	if hp.config.PromRegistry != nil {
		p := middleware.NewPrometheus(hp.config.PromRegistry, hp.config.PromNamespace, hp.config.PromHTTPOpts...)
		stack.AddRequestModifier(p)
		stack.AddResponseModifier(p)

		trace = new(martian.ProxyTrace)
		trace.ReadRequest = func(info martian.ReadRequestInfo) {
			if info.Req != nil {
				p.ReadRequest(info.Req)
			}
		}
		trace.WroteResponse = func(info martian.WroteResponseInfo) {
			if info.Res != nil {
				p.WroteResponse(info.Res)
			}
		}
	}

	fg.AddRequestModifier(martian.RequestModifierFunc(hp.setBasicAuth))
	fg.AddRequestModifier(martian.RequestModifierFunc(setEmptyUserAgent))

	return topg.ToImmutable(), trace
}

func (hp *HTTPProxy) basicAuth(u *url.Userinfo) martian.RequestModifier {
	user := u.Username()
	pass, _ := u.Password()
	ba := middleware.NewProxyBasicAuth()

	return martian.RequestModifierFunc(func(req *http.Request) error {
		if !ba.AuthenticatedRequest(req, user, pass) {
			return ErrProxyAuthentication
		}
		return nil
	})
}

func (hp *HTTPProxy) denyLocalhost() martian.RequestModifier {
	return martian.RequestModifierFunc(func(req *http.Request) error {
		if hp.isLocalhost(req.URL.Hostname()) {
			return ErrProxyLocalhost
		}
		return nil
	})
}

func (hp *HTTPProxy) denyDomains(r Matcher) martian.RequestModifier {
	return martian.RequestModifierFunc(func(req *http.Request) error {
		if r.Match(req.URL.Hostname()) {
			return ErrProxyDenied
		}
		return nil
	})
}

func (hp *HTTPProxy) directDomains(fn ProxyFunc) ProxyFunc {
	if fn == nil {
		return nil
	}

	return func(req *http.Request) (*url.URL, error) {
		if hp.config.DirectDomains.Match(req.URL.Hostname()) {
			return nil, nil
		}
		return fn(req)
	}
}

func (hp *HTTPProxy) directLocalhost(fn ProxyFunc) ProxyFunc {
	if fn == nil {
		return nil
	}

	return func(req *http.Request) (*url.URL, error) {
		if hp.isLocalhost(req.URL.Hostname()) {
			return nil, nil
		}
		return fn(req)
	}
}

func (hp *HTTPProxy) isLocalhost(host string) bool {
	host = strings.ToLower(host)

	if slices.Contains(hp.localhost, host) {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}

	return false
}

func (hp *HTTPProxy) setBasicAuth(req *http.Request) error {
	if req.Header.Get("Authorization") == "" {
		if u := hp.creds.MatchURL(req.URL); u != nil {
			p, _ := u.Password()
			req.SetBasicAuth(u.Username(), p)
		}
	}

	return nil
}

func setEmptyUserAgent(req *http.Request) error {
	if _, ok := req.Header["User-Agent"]; !ok {
		// If the outbound request doesn't have a User-Agent header set,
		// don't send the default Go HTTP client User-Agent.
		req.Header.Set("User-Agent", "")
	}
	return nil
}

func (hp *HTTPProxy) MITMCACert() *x509.Certificate {
	return hp.mitmCACert
}

func (hp *HTTPProxy) ProxyFunc() ProxyFunc {
	return hp.proxyFunc
}

func (hp *HTTPProxy) handler() http.Handler {
	return hp.proxy.Handler()
}

func (hp *HTTPProxy) Run(ctx context.Context) error {
	if hp.config.TestingHTTPHandler {
		hp.log.Infof("using http handler")
		return hp.runHTTPHandler(ctx)
	}
	return hp.run(ctx)
}

func (hp *HTTPProxy) runHTTPHandler(ctx context.Context) error {
	srv := http.Server{
		Handler:           hp.handler(),
		IdleTimeout:       hp.config.IdleTimeout,
		ReadTimeout:       hp.config.ReadTimeout,
		ReadHeaderTimeout: hp.config.ReadHeaderTimeout,
		WriteTimeout:      hp.config.WriteTimeout,
	}

	var g errgroup.Group
	g.Go(func() error {
		<-ctx.Done()
		ctxErr := ctx.Err()

		var cancel context.CancelFunc
		ctx, cancel = shutdownContext(hp.config.shutdownConfig)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			hp.log.Debugf("failed to gracefully shutdown server error=%s", err)
			if err := srv.Close(); err != nil {
				hp.log.Debugf("failed to close server error=%s", err)
			}
		}

		return ctxErr
	})
	for i := range hp.listeners {
		l := hp.listeners[i]
		g.Go(func() error {
			err := srv.Serve(l)
			if errors.Is(err, http.ErrServerClosed) {
				err = nil
			}
			return err
		})
	}
	return g.Wait()
}

func (hp *HTTPProxy) run(ctx context.Context) error {
	var g errgroup.Group
	g.Go(func() error {
		<-ctx.Done()
		ctxErr := ctx.Err()

		// Close listeners first to prevent new connections.
		if err := hp.Close(); err != nil {
			hp.log.Debugf("failed to close listeners error=%s", err)
		}

		var cancel context.CancelFunc
		ctx, cancel = shutdownContext(hp.config.shutdownConfig)
		defer cancel()

		if err := hp.proxy.Shutdown(ctx); err != nil {
			hp.log.Debugf("failed to gracefully shutdown server error=%s", err)
			if err := hp.proxy.Close(); err != nil {
				hp.log.Debugf("failed to close server error=%s", err)
			}
		}

		if tr, ok := hp.transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}

		return ctxErr
	})
	for i := range hp.listeners {
		l := hp.listeners[i]
		g.Go(func() error {
			err := hp.proxy.Serve(l)
			if errors.Is(err, net.ErrClosed) {
				err = nil
			}
			return err
		})
	}
	return g.Wait()
}

func (hp *HTTPProxy) listen() ([]net.Listener, error) {
	switch hp.config.Protocol {
	case HTTPScheme, HTTPSScheme, HTTP2Scheme:
	default:
		return nil, fmt.Errorf("invalid protocol %q", hp.config.Protocol)
	}

	if len(hp.config.ExtraListeners) == 0 {
		l := &Listener{
			ListenerConfig: hp.config.ListenerConfig,
			TLSConfig:      hp.tlsConfig,
			PromConfig: PromConfig{
				PromNamespace: hp.config.PromNamespace,
				PromRegistry:  hp.config.PromRegistry,
			},
		}
		if err := l.Listen(); err != nil {
			return nil, err
		}
		return []net.Listener{l}, nil
	}

	return MultiListener{
		ListenerConfigs: append([]NamedListenerConfig{{ListenerConfig: hp.config.ListenerConfig}}, hp.config.ExtraListeners...),
		TLSConfig: func(lc NamedListenerConfig) *tls.Config {
			return hp.tlsConfig
		},
		PromConfig: hp.config.PromConfig,
	}.Listen()
}

// Addr returns the address the server is listening on.
func (hp *HTTPProxy) Addr() (addrs []string, ok bool) {
	addrs = make([]string, len(hp.listeners))
	ok = true
	for i, l := range hp.listeners {
		addrs[i] = l.Addr().String()
		if addrs[i] == "" {
			ok = false
		}
	}
	return
}

func (hp *HTTPProxy) Close() error {
	var err error
	for _, l := range hp.listeners {
		if e := l.Close(); e != nil {
			err = multierr.Append(err, e)
		}
	}
	return err
}
