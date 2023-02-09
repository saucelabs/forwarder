// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/fifo"
	"github.com/google/martian/v3/httpspec"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/forwarder/pac"
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

type HTTPProxyConfig struct {
	HTTPServerConfig
	ProxyLocalhost    ProxyLocalhostMode         `json:"proxy_localhost"`
	UpstreamProxy     *url.URL                   `json:"upstream_proxy_uri"`
	UpstreamProxyFunc ProxyFunc                  `json:"-"`
	RequestModifiers  []martian.RequestModifier  `json:"-"`
	ResponseModifiers []martian.ResponseModifier `json:"-"`
	CloseAfterReply   bool                       `json:"close_after_reply"`
	RemoveHeaders     []string                   `json:"remove_headers"`
}

func DefaultHTTPProxyConfig() *HTTPProxyConfig {
	return &HTTPProxyConfig{
		HTTPServerConfig: HTTPServerConfig{
			Protocol:          HTTPScheme,
			Addr:              ":3128",
			ReadHeaderTimeout: 1 * time.Minute,
			LogHTTPMode:       httplog.ErrOnlyLogMode,
		},
		ProxyLocalhost: DenyProxyLocalhost,
	}
}

func (c *HTTPProxyConfig) Validate() error {
	if err := c.HTTPServerConfig.Validate(); err != nil {
		return err
	}
	if c.Protocol != HTTPScheme && c.Protocol != HTTPSScheme && c.Protocol != TunnelScheme {
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
	config    HTTPProxyConfig
	pac       PACResolver
	creds     *CredentialsMatcher
	transport http.RoundTripper
	log       log.Logger
	proxy     *martian.Proxy
	addr      atomic.Pointer[string]

	TLSConfig *tls.Config
}

func NewHTTPProxy(cfg *HTTPProxyConfig, pr PACResolver, cm *CredentialsMatcher, rt http.RoundTripper, log log.Logger) (*HTTPProxy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.UpstreamProxy != nil && pr != nil {
		return nil, fmt.Errorf("cannot use both upstream proxy and PAC")
	}

	// If not set, use http.DefaultTransport.
	if rt == nil {
		log.Infof("HTTP transport not configured, using standard library default")
		rt = http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // we know it's a *http.Transport
	} else if _, ok := rt.(*http.Transport); !ok {
		log.Debugf("Using custom HTTP transport %T", rt)
	}

	hp := &HTTPProxy{
		config:    *cfg,
		pac:       pr,
		creds:     cm,
		transport: rt,
		log:       log,
	}

	if hp.config.Protocol == HTTPSScheme {
		if err := hp.configureHTTPS(); err != nil {
			return nil, err
		}
	}
	hp.configureProxy()

	return hp, nil
}

func (hp *HTTPProxy) configureHTTPS() error {
	if hp.config.CertFile == "" && hp.config.KeyFile == "" {
		hp.log.Infof("No SSL certificate provided, using self-signed certificate")
	}
	tlsCfg := httpsTLSConfigTemplate()
	err := hp.config.loadCertificate(tlsCfg)
	hp.TLSConfig = tlsCfg
	return err
}

func (hp *HTTPProxy) configureProxy() {
	hp.proxy = martian.NewProxy()
	hp.proxy.AllowHTTP = true
	hp.proxy.ConnectPassthrough = hp.config.Protocol == TunnelScheme
	hp.proxy.WithoutWarning = true
	hp.proxy.ErrorResponse = errorResponse
	hp.proxy.CloseAfterReply = hp.config.CloseAfterReply
	hp.proxy.ReadTimeout = hp.config.ReadTimeout
	hp.proxy.ReadHeaderTimeout = hp.config.ReadHeaderTimeout
	hp.proxy.WriteTimeout = hp.config.WriteTimeout
	// Martian has an intertwined logic for setting http.Transport and the dialer.
	// The dialer is wrapped, so that additional syscalls are made to the dialed connections.
	// As a result the dialer needs to be reset.
	if tr, ok := hp.transport.(*http.Transport); ok {
		hp.proxy.SetRoundTripper(tr)
		hp.proxy.SetDial(tr.Dial) //nolint:staticcheck // Martian does not use context
	} else {
		hp.proxy.SetRoundTripper(hp.transport)
	}

	var fn ProxyFunc
	switch {
	case hp.config.UpstreamProxyFunc != nil:
		hp.log.Infof("Using external proxy function")
		fn = hp.config.UpstreamProxyFunc
	case hp.config.UpstreamProxy != nil:
		u := hp.upstreamProxyURL()
		hp.log.Infof("Using upstream proxy: %s", u.Redacted())
		fn = http.ProxyURL(u)
	case hp.pac != nil:
		hp.log.Infof("Using PAC proxy")
		fn = hp.pacProxy
	default:
		hp.log.Infof("Using direct proxy")
	}

	hp.log.Infof("Localhost proxying mode: %s", hp.config.ProxyLocalhost)
	if hp.config.ProxyLocalhost == DirectProxyLocalhost {
		fn = hp.directLocalhost(fn)
	}
	hp.proxy.SetUpstreamProxyFunc(fn)

	mw := hp.middlewareStack()
	hp.proxy.SetRequestModifier(mw)
	hp.proxy.SetResponseModifier(mw)
}

func (hp *HTTPProxy) upstreamProxyURL() *url.URL {
	proxyURL := new(url.URL)
	*proxyURL = *hp.config.UpstreamProxy

	if u := hp.creds.MatchURL(proxyURL); u != nil {
		proxyURL.User = u
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

func (hp *HTTPProxy) middlewareStack() martian.RequestResponseModifier {
	// Wrap stack in a group so that we can run security checks before the httpspec modifiers.
	topg := fifo.NewGroup()
	if hp.config.BasicAuth != nil {
		hp.log.Infof("Basic auth enabled")
		topg.AddRequestModifier(hp.basicAuth(hp.config.BasicAuth))
	}
	if hp.config.ProxyLocalhost == DenyProxyLocalhost {
		topg.AddRequestModifier(hp.denyLocalhost())
	}

	// stack contains the request/response modifiers in the order they are applied.
	// fg is the inner stack that is executed after the core request modifiers and before the core response modifiers.
	stack, fg := httpspec.NewStack("forwarder")
	topg.AddRequestModifier(stack)
	topg.AddResponseModifier(stack)

	for _, hr := range hp.config.RemoveHeaders {
		fg.AddRequestModifier(newHeaderRemover(hr))
	}

	for _, m := range hp.config.RequestModifiers {
		fg.AddRequestModifier(m)
	}

	for _, m := range hp.config.ResponseModifiers {
		fg.AddResponseModifier(m)
	}

	logHTTP := httplog.NewLogger(hp.log.Infof, hp.config.LogHTTPMode).LogFunc()
	fg.AddRequestModifier(logHTTP)
	fg.AddResponseModifier(logHTTP)

	if hp.config.HTTPServerConfig.Addr != "" {
		p := middleware.NewPrometheus(hp.config.PromRegistry, hp.config.PromNamespace)
		stack.AddRequestModifier(p)
		stack.AddResponseModifier(p)
	}

	fg.AddRequestModifier(martian.RequestModifierFunc(hp.setBasicAuth))

	return topg.ToImmutable()
}

func (hp *HTTPProxy) abortIf(condition func(r *http.Request) bool, response func(*http.Request) *http.Response, returnErr error) martian.RequestModifier {
	return martian.RequestModifierFunc(func(req *http.Request) error {
		if !condition(req) {
			return nil
		}

		logHTTP := httplog.NewLogger(hp.log.Infof, hp.config.LogHTTPMode).LogFunc()
		logHTTP.ModifyRequest(req) //nolint:errcheck // This is only logging.

		ctx := martian.NewContext(req)
		_, brw, err := ctx.Session().Hijack()
		if err != nil {
			return err
		}

		resp := response(req)
		defer resp.Body.Close()
		resp.Close = true // hijacked connection is closed by Martian in handleLoop()
		resp.Write(brw)   //nolint:errcheck // it's a buffer
		if err := brw.Flush(); err != nil {
			return fmt.Errorf("got error while flushing response back to client: %w", err)
		}

		logHTTP.ModifyResponse(resp) //nolint:errcheck // this is only logging

		return returnErr
	})
}

func (hp *HTTPProxy) basicAuth(u *url.Userinfo) martian.RequestModifier {
	user := u.Username()
	pass, _ := u.Password()
	ba := middleware.NewProxyBasicAuth()

	return hp.abortIf(func(req *http.Request) bool {
		return !ba.AuthenticatedRequest(req, user, pass)
	}, unauthorizedResponse, errors.New("basic auth required"))
}

func (hp *HTTPProxy) denyLocalhost() martian.RequestModifier {
	return hp.abortIf(hp.isLocalhost, func(req *http.Request) *http.Response {
		return errorResponse(req, ErrProxyLocalhost)
	}, errors.New("localhost access denied"))
}

func (hp *HTTPProxy) directLocalhost(fn ProxyFunc) ProxyFunc {
	if fn == nil {
		return nil
	}

	return func(req *http.Request) (*url.URL, error) {
		if hp.isLocalhost(req) {
			return nil, nil
		}
		return fn(req)
	}
}

// nopResolver is a dns resolver that does not ever dial.
// It uses only the local resolver.
// It will look for the address in the /etc/hosts file.
var localhostResolver = nopResolver() //nolint:gochecknoglobals // This is a local resolver.

func (hp *HTTPProxy) isLocalhost(req *http.Request) bool {
	h := req.URL.Hostname()

	// Plain old localhost.
	if h == "localhost" {
		return true
	}

	if addrs, err := localhostResolver.LookupHost(context.Background(), h); err == nil {
		if ip := net.ParseIP(addrs[0]); ip != nil {
			return ip.IsLoopback()
		}
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

func (hp *HTTPProxy) Run(ctx context.Context) error {
	listener, err := hp.listener()
	if err != nil {
		return err
	}
	defer listener.Close()

	addr := listener.Addr().String()
	hp.addr.Store(&addr)
	hp.log.Infof("PROXY server listen address=%s protocol=%s", addr, hp.config.Protocol)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		hp.proxy.Close()
		listener.Close()
	}()

	srvErr := hp.proxy.Serve(listener)
	if srvErr != nil {
		if errors.Is(srvErr, net.ErrClosed) {
			srvErr = nil
		}
		return srvErr
	}

	wg.Wait()
	return nil
}

func (hp *HTTPProxy) listener() (net.Listener, error) {
	listener, err := net.Listen("tcp", hp.config.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to open listener on address %s: %w", hp.config.Addr, err)
	}

	switch hp.config.Protocol {
	case HTTPScheme:
		return listener, nil
	case HTTPSScheme, HTTP2Scheme:
		return tls.NewListener(listener, hp.TLSConfig), nil
	default:
		listener.Close()
		return nil, fmt.Errorf("invalid protocol %q", hp.config.Protocol)
	}
}

// Addr returns the address the server is listening on or an empty string if the server is not running.
func (hp *HTTPProxy) Addr() string {
	addr := hp.addr.Load()
	if addr == nil {
		return ""
	}
	return *addr
}

// headerRemover removes headers that match given prefix.
type headerRemover struct {
	prefix string
}

func newHeaderRemover(prefix string) martian.RequestModifier {
	return &headerRemover{prefix: http.CanonicalHeaderKey(prefix)}
}

func (m *headerRemover) ModifyRequest(req *http.Request) error {
	for k := range req.Header {
		kk := http.CanonicalHeaderKey(k)
		if strings.HasPrefix(kk, m.prefix) {
			req.Header.Del(k)
		}
	}
	return nil
}
