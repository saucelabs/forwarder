// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
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
	"sync"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/fifo"
	"github.com/google/martian/v3/httpspec"
	"github.com/google/martian/v3/martianlog"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/forwarder/pac"
	"go.uber.org/atomic"
)

type HTTPProxyConfig struct {
	HTTPServerConfig
	UpstreamProxy  *url.URL `json:"upstream_proxy_uri"`
	ProxyLocalhost bool     `json:"proxy_localhost"`
}

func DefaultHTTPProxyConfig() *HTTPProxyConfig {
	return &HTTPProxyConfig{
		HTTPServerConfig: HTTPServerConfig{
			Protocol:    HTTPScheme,
			Addr:        ":3128",
			ReadTimeout: 0, // TODO(mmatczuk): we need a more fine-grained timeouts see #129
		},
	}
}

func (c *HTTPProxyConfig) Validate() error {
	if err := c.HTTPServerConfig.Validate(); err != nil {
		return err
	}
	if c.Protocol != HTTPScheme && c.Protocol != HTTPSScheme {
		return fmt.Errorf("unsupported protocol: %s", c.Protocol)
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
	transport *http.Transport
	log       Logger
	hosts     map[string]net.IP
	proxy     *martian.Proxy
	addr      atomic.String

	TLSConfig *tls.Config
	Listener  net.Listener
}

func NewHTTPProxy(cfg *HTTPProxyConfig, pr PACResolver, cm *CredentialsMatcher, t *http.Transport, log Logger) (*HTTPProxy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.UpstreamProxy != nil && pr != nil {
		return nil, fmt.Errorf("cannot use both upstream proxy and PAC")
	}

	// If not set, use http.DefaultTransport.
	if t == nil {
		log.Infof("HTTP transport not configured, using standard library default")
		t = http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // we know it's a *http.Transport
	}

	// Read /etc/hosts and store it in memory.
	hosts, err := ReadHostsFile()
	if err != nil {
		return nil, err
	}

	hp := &HTTPProxy{
		config:    *cfg,
		pac:       pr,
		creds:     cm,
		transport: t,
		hosts:     hosts,
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
	hp.proxy.WithoutWarning = true
	hp.proxy.ErrorResponse = errorResponse

	// Martian has an intertwined logic for setting http.Transport and the dialer.
	// The dialer is wrapped, so that additional syscalls are made to the dialed connections.
	// As a result the dialer needs to be reset.
	hp.proxy.SetRoundTripper(hp.transport)
	hp.proxy.SetDial(hp.transport.Dial) //nolint:staticcheck // Martian does not use context
	hp.proxy.SetTimeout(hp.config.ReadTimeout)

	switch {
	case hp.config.UpstreamProxy != nil:
		u := hp.upstreamProxyURL()
		hp.log.Infof("Using upstream proxy: %s", u.Redacted())
		hp.proxy.SetDownstreamProxyFunc(http.ProxyURL(u))
	case hp.pac != nil:
		hp.log.Infof("Using PAC proxy")
		hp.proxy.SetDownstreamProxyFunc(hp.pacProxy)
	default:
		hp.log.Infof("Using direct proxy")
		hp.proxy.SetDownstreamProxyFunc(nil)
	}

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
	s, err := hp.pac.FindProxyForURL(r.URL, r.Host)
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
	if hp.config.ProxyLocalhost {
		hp.log.Infof("Localhost proxying enabled")
	} else {
		hp.log.Infof("Localhost proxying disabled")
		topg.AddRequestModifier(hp.denyLocalhost())
	}

	// stack contains the request/response modifiers in the order they are applied.
	// fg is the inner stack that is executed after the core request modifiers and before the core response modifiers.
	stack, fg := httpspec.NewStack("forwarder")
	topg.AddRequestModifier(stack)
	topg.AddResponseModifier(stack)

	if hp.config.LogHTTPRequests {
		logger := martianlog.NewLogger()
		logger.SetHeadersOnly(true)
		logger.SetDecode(false)
		fg.AddRequestModifier(logger)
		fg.AddResponseModifier(logger)
	}
	fg.AddRequestModifier(martian.RequestModifierFunc(hp.setBasicAuth))

	return topg.ToImmutable()
}

func abortIf(condition func(r *http.Request) bool, response func(*http.Request) *http.Response, returnErr error) martian.RequestModifier {
	return martian.RequestModifierFunc(func(req *http.Request) error {
		if !condition(req) {
			return nil
		}

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

		return returnErr
	})
}

func (hp *HTTPProxy) basicAuth(u *url.Userinfo) martian.RequestModifier {
	user := u.Username()
	pass, _ := u.Password()
	ba := middleware.NewProxyBasicAuth()

	return abortIf(func(req *http.Request) bool {
		return !ba.AuthenticatedRequest(req, user, pass)
	}, unauthorizedResponse, errors.New("basic auth required"))
}

func (hp *HTTPProxy) denyLocalhost() martian.RequestModifier {
	return abortIf(hp.isLocalhost, func(req *http.Request) *http.Response {
		return errorResponse(req, ErrProxyLocalhost)
	}, errors.New("localhost access denied"))
}

// The goproxy.IsLocalHost is broken.
// In addition, it doesn't work with hostnames that are not in /etc/hosts.
// See https://github.com/elazarl/goproxy/issues/487
func (hp *HTTPProxy) isLocalhost(req *http.Request) bool {
	h := req.URL.Hostname()

	// Plain old localhost.
	if h == "localhost" {
		return true
	}
	// Hostname from /etc/hosts file.
	if ip := hp.hosts[h]; ip != nil {
		return ip.IsLoopback()
	}
	// IP addresses.
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback()
	}
	// IPv6 without a port number - Hostname returns the invalid value.
	if ip := net.ParseIP(req.URL.Host); ip != nil {
		return ip.IsLoopback()
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

	hp.addr.Store(listener.Addr().String())
	hp.log.Infof("PROXY server listen address=%s protocol=%s", listener.Addr(), hp.config.Protocol)

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
	if hp.Listener != nil {
		return hp.Listener, nil
	}

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
	return hp.addr.Load()
}
