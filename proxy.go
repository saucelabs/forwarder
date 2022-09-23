// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/pacman"
)

const (
	DNSTimeout      = 1 * time.Minute
	httpPort        = 80
	httpsPort       = 443
	proxyAuthHeader = "Proxy-Authorization"
	authHeader      = "Authorization"
)

// Mode specifies mode of operation of the proxy.
type Mode string

const (
	Direct   Mode = "DIRECT"
	Upstream Mode = "Upstream"
	PAC      Mode = "PAC"
)

// Proxy definition. Proxy can be protected, or not.
// It can forward connections to an upstream proxy protected, or not.
// The upstream proxy can be automatically setup via PAC.
// PAC content can be retrieved from multiple sources, e.g.: a HTTP server, also, protected or not.
// Protection means basic auth.
type Proxy struct {
	config    ProxyConfig
	userInfo  *userInfoMatcher
	pacParser *pacman.Parser
	proxy     *goproxy.ProxyHttpServer
	log       Logger
}

func NewProxy(cfg ProxyConfig, log Logger) (*Proxy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Parse site credential list into map of host:port -> base64 encoded input.
	m, err := newUserInfoMatcher(cfg.SiteCredentials, log)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	p := &Proxy{
		config:   cfg.Clone(),
		userInfo: m,
		proxy:    goproxy.NewProxyHttpServer(),
		log:      log,
	}
	if p.config.PACURI != nil {
		pacParser, err := pacman.New(p.config.PACURI.String(), p.config.PACProxiesCredentials...)
		if err != nil {
			return nil, fmt.Errorf("pac parser: %w", err)
		}
		p.pacParser = pacParser
	}
	p.setupDNS()
	p.setupProxy()

	return p, nil
}

func (p *Proxy) setupDNS() {
	d := net.Dialer{
		Timeout: DNSTimeout,
	}
	setupDNS(p.config.DNSURIs, &d, p.log)
}

func (p *Proxy) setupProxy() {
	p.proxy.Logger = goproxyLogger{p.log}
	p.proxy.Verbose = true
	p.proxy.KeepDestinationHeaders = true
	// This is required.
	//
	// See: https://maelvls.dev/go-ignores-proxy-localhost/
	// See: https://github.com/golang/go/issues/28866
	// See: https://github.com/elazarl/goproxy/issues/306
	p.proxy.KeepHeader = true
	p.proxy.Tr = &http.Transport{}

	// Local proxy authentication.
	if u := p.config.LocalProxyURI.User; u.Username() != "" {
		p.setupBasicAuth(u)
	}

	p.setupProxyHandlers()
}

// Config returns a copy of the proxy configuration.
func (p *Proxy) Config() ProxyConfig {
	return p.config.Clone()
}

// Mode returns mode of operation of the proxy as specified in the config.
func (p *Proxy) Mode() Mode {
	switch {
	case p.config.UpstreamProxyURI != nil:
		return Upstream
	case p.config.PACURI != nil:
		return PAC
	default:
		return Direct
	}
}

// Sets the `Proxy-Authorization` header based on `uri` user info.
func (p *Proxy) setProxyBasicAuthHeader(uri *url.URL, req *http.Request) {
	req.Header.Set(
		proxyAuthHeader,
		fmt.Sprintf("Basic %s", userInfoBase64(uri.User)),
	)

	p.log.Debugf(
		"%s header set with %s:*** for %s",
		proxyAuthHeader,
		uri.User.Username(),
		req.URL.String(),
	)
}

// Removes any upstream proxy settings.
func resetUpstreamSettings(ctx *goproxy.ProxyCtx) {
	ctx.Proxy.ConnectDial = nil

	ctx.Proxy.Tr = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // FIXME https://github.com/saucelabs/forwarder/issues/47
		},
		Proxy: nil,
	}
}

// Returns `true` if should NOT proxy connections to any upstream proxy.
func (p *Proxy) shouldNotProxyLocalhost(ctx *goproxy.ProxyCtx) bool {
	if !p.config.ProxyLocalhost && isLocalhost(ctx.Req.URL.Hostname()) {
		resetUpstreamSettings(ctx)

		return true
	}

	return false
}

// setupUpstreamProxyConnection forwards connections to an upstream proxy.
func (p *Proxy) setupUpstreamProxyConnection(ctx *goproxy.ProxyCtx, uri *url.URL) {
	ctx.Proxy.Tr.Proxy = http.ProxyURL(uri)

	var connectReqHandler func(req *http.Request)

	if uri.User.Username() != "" {
		connectReqHandler = func(req *http.Request) {
			p.log.Debugf("Setting basic auth header from connection handler to upstream proxy.")
			p.setProxyBasicAuthHeader(uri, req)
		}
	}

	ctx.Proxy.ConnectDial = ctx.Proxy.NewConnectDialToProxyWithHandler(uri.String(), connectReqHandler)

	p.log.Debugf("Connection to the upstream proxy %s is set up", uri.Redacted())
}

// setupPACUpstreamProxyConnection dynamically forwards connections to an upstream
// proxy setup via PAC.
func setupPACUpstreamProxyConnection(p *Proxy, ctx *goproxy.ProxyCtx) error {
	urlToFindProxyFor := ctx.Req.URL.String()
	hostToFindProxyFor := ctx.Req.URL.Hostname()

	p.log.Debugf("Finding proxy for %s", hostToFindProxyFor)

	pacProxies, err := p.pacParser.FindProxy(urlToFindProxyFor)
	if err != nil {
		return err
	}

	// Should only do something if there's any proxy
	if len(pacProxies) > 0 {
		// TODO: Should find the best proxy from a list of possible proxies?
		pacProxy := pacProxies[0]
		pacProxyURI := pacProxy.GetURI()

		// Should only set up upstream if there's a proxy and not `DIRECT`.
		if pacProxyURI != nil {
			p.log.Debugf("Using proxy %s for %s", pacProxyURI.Redacted(), hostToFindProxyFor)
			p.setupUpstreamProxyConnection(ctx, pacProxyURI)
			return nil
		}
	}

	p.log.Debugf("Using no proxy for %s", hostToFindProxyFor)
	// Clear upstream proxy settings (if any) for this request.
	resetUpstreamSettings(ctx)

	return nil
}

// DRY on handler's code.
func (p *Proxy) setupHandlers(ctx *goproxy.ProxyCtx) error {
	if p.shouldNotProxyLocalhost(ctx) {
		p.log.Debugf("Not proxifying request to localhost URL: %s", ctx.Req.URL.String())

		return nil
	}

	switch p.Mode() {
	case Direct:
		// Do nothing
	case Upstream:
		p.setupUpstreamProxyConnection(ctx, p.config.UpstreamProxyURI)
	case PAC:
		if err := setupPACUpstreamProxyConnection(p, ctx); err != nil {
			return err
		}
	}

	return nil
}

// setupBasicAuth protects proxy with basic auth.
func (p *Proxy) setupBasicAuth(u *url.Userinfo) {
	// TODO: Allows to set `realm`.
	auth.ProxyBasic(p.proxy, "localhost", func(username, password string) (ok bool) {
		defer func() {
			p.log.Debugf("Incoming request. This proxy (%s) is protected authorized=%v", p.config.LocalProxyURI.Redacted(), ok)
		}()

		pwd, _ := u.Password()
		// Securely compare passwords.
		return subtle.ConstantTimeCompare([]byte(u.Username()), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(pwd), []byte(password)) == 1
	})

	p.log.Debugf("Basic auth setup for proxy @ %s", p.config.LocalProxyURI.Redacted())
}

// Run starts the proxy.
// It's safe to call it multiple times - nothing will happen.
func (p *Proxy) Run() error {
	p.log.Infof("Listening on %s", p.config.LocalProxyURI.Host)
	if err := http.ListenAndServe(p.config.LocalProxyURI.Host, p.proxy); err != nil { //nolint:gosec // FIXME https://github.com/saucelabs/forwarder/issues/45
		return fmt.Errorf("start proxy: %w", err)
	}

	return nil
}

func (p *Proxy) logRequest(r *http.Request) {
	p.log.Debugf("%s %s -> %s %s %s", r.Method, r.RemoteAddr, r.URL.Scheme, r.Host, r.URL.Port())

	b, err := httputil.DumpRequest(r, false)
	if err != nil {
		p.log.Errorf("failed to dump request: %w", err)
	}
	p.log.Debugf(string(b))
}

func (p *Proxy) setupProxyHandlers() {
	p.proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		p.logRequest(ctx.Req)
		if err := p.setupHandlers(ctx); err != nil {
			p.log.Errorf("Failed to setup handler (HTTPS) for request %s. %+v", ctx.Req.URL.Redacted(), err)

			return goproxy.RejectConnect, host
		}

		return nil, host
	})

	p.proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		p.logRequest(ctx.Req)
		if err := p.setupHandlers(ctx); err != nil {
			p.log.Errorf("Failed to setup handler (HTTP) for request %s. %+v", ctx.Req.URL.Redacted(), err)

			return nil, goproxy.NewResponse(
				ctx.Req,
				goproxy.ContentTypeText,
				http.StatusInternalServerError,
				err.Error(),
			)
		}

		return ctx.Req, nil
	})

	p.proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		p.maybeAddAuthHeader(req)
		return ctx.Req, nil
	})

	p.proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp != nil {
			p.log.Debugf("%s <- %s %v (%v bytes)",
				resp.Request.RemoteAddr, resp.Request.Host, resp.Status, resp.ContentLength)
		} else {
			p.log.Debugf("%s <- %s response is empty", ctx.Req.Host, ctx.Req.RemoteAddr)
		}

		return resp
	})
}

// maybeAddAuthHeader modifies the request and adds an authorization header if necessary.
func (p *Proxy) maybeAddAuthHeader(req *http.Request) {
	hostport := req.Host

	if req.URL.Port() == "" {
		// When the destination URL doesn't contain an explicit port, Go http-parsed
		// URL Port() returns an empty string.
		switch req.URL.Scheme {
		case "http":
			hostport = fmt.Sprintf("%s:%d", req.Host, httpPort)
		case "https":
			hostport = fmt.Sprintf("%s:%d", req.Host, httpsPort)
		default:
			p.log.Errorf("Failed to determine port for %s.", req.URL.Redacted())
		}
	}

	if u := p.userInfo.Match(hostport); u != nil {
		req.Header.Set(authHeader, fmt.Sprintf("Basic %s", userInfoBase64(u)))
	}
}
