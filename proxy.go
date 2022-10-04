// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/pacman"
)

// ProxyConfig definition.
type ProxyConfig struct {
	// LocalProxyURI is the local proxy URI, ex. http://user:password@127.0.0.1:8080.
	// Requirements:
	// - Known schemes: http, https, socks, socks5, or quic.
	// - Hostname or IP.
	// - Port in a valid range: 1 - 65535.
	// - Username and password are optional.
	LocalProxyURI *url.URL `json:"local_proxy_uri"`

	// UpstreamProxyURI is the upstream proxy URI, ex. http://user:password@127.0.0.1:8080.
	// Only one of `UpstreamProxyURI` or `PACURI` can be set.
	// Requirements:
	// - Known schemes: http, https, socks, socks5, or quic.
	// - Hostname or IP.
	// - Port in a valid range: 1 - 65535.
	// - Username and password are optional.
	UpstreamProxyURI *url.URL `json:"upstream_proxy_uri"`

	// PACURI is the PAC URI, which is used to determine the upstream proxy, ex. http://127.0.0.1:8087/data.pac.
	// Only one of `UpstreamProxyURI` or `PACURI` can be set.
	PACURI *url.URL `json:"pac_uri"`

	// Credentials for proxies specified in PAC content.
	PACProxiesCredentials []string `json:"pac_proxies_credentials"`

	// ProxyLocalhost if `true`, requests to `localhost`, `127.0.0.*`, `0:0:0:0:0:0:0:1` will be forwarded to upstream.
	ProxyLocalhost bool `json:"proxy_localhost"`

	// SiteCredentials contains URLs with the credentials, ex.:
	// - https://usr1:pwd1@foo.bar:4443
	// - http://usr2:pwd2@bar.foo:8080
	// - usr3:pwd3@bar.foo:8080
	// Proxy will add basic auth headers for requests to these URLs.
	SiteCredentials []string `json:"site_credentials"`
}

func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		LocalProxyURI: &url.URL{Scheme: "http", Host: "localhost:8080"},
	}
}

func (c *ProxyConfig) Clone() *ProxyConfig {
	v := new(ProxyConfig)
	deepCopy(v, c)
	return v
}

func (c *ProxyConfig) Validate() error {
	if c.LocalProxyURI == nil {
		return fmt.Errorf("local_proxy_uri is required")
	}
	if err := validateProxyURI(c.LocalProxyURI); err != nil {
		return fmt.Errorf("local_proxy_uri: %w", err)
	}
	if err := validateProxyURI(c.UpstreamProxyURI); err != nil {
		return fmt.Errorf("upstream_proxy_uri: %w", err)
	}
	if err := validateProxyURI(c.PACURI); err != nil {
		return fmt.Errorf("pac_uri: %w", err)
	}
	if c.UpstreamProxyURI != nil && c.PACURI != nil {
		return fmt.Errorf("only one of upstream_proxy_uri or pac_uri can be set")
	}

	return nil
}

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
	transport *http.Transport
	userInfo  *userInfoMatcher
	pacParser *pacman.Parser
	proxy     *goproxy.ProxyHttpServer
	log       Logger
}

func NewProxy(cfg *ProxyConfig, r *net.Resolver, log Logger) (*Proxy, error) {
	cfg = cfg.Clone()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Parse site credential list into map of host:port -> base64 encoded input.
	m, err := newUserInfoMatcher(cfg.SiteCredentials, log)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	p := &Proxy{
		config:   *cfg,
		userInfo: m,
		log:      log,
	}
	p.setupTransport(r)

	if p.config.PACURI != nil {
		pacParser, err := pacman.New(p.config.PACURI.String(), p.config.PACProxiesCredentials...)
		if err != nil {
			return nil, fmt.Errorf("pac parser: %w", err)
		}
		p.pacParser = pacParser
	}

	p.setupProxy()

	return p, nil
}

func (p *Proxy) setupTransport(r *net.Resolver) {
	p.transport = http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // We know it's a *http.Transport.
	p.transport.DialContext = defaultTransportDialContext(&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  r,
	})
	p.transport.Proxy = nil
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func (p *Proxy) setupProxy() {
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Logger = goproxyLogger{p.log}
	p.proxy.Verbose = true
	p.proxy.KeepDestinationHeaders = true
	// This is required.
	//
	// See: https://maelvls.dev/go-ignores-proxy-localhost/
	// See: https://github.com/golang/go/issues/28866
	// See: https://github.com/elazarl/goproxy/issues/306
	p.proxy.KeepHeader = true

	p.proxy.Tr = p.transport.Clone()
	p.setupLocalhostProxy()

	switch p.Mode() {
	case Direct:
		p.setupDirect()
	case Upstream:
		p.setupUpstreamProxy()
	case PAC:
		p.setupPACProxy()
	default:
		panic(fmt.Errorf("unknown mode %q", p.Mode()))
	}

	p.setupProxyBasicAuth()
	p.setupSiteBasicAuth()
}

func (p *Proxy) setupLocalhostProxy() {
	if !p.config.ProxyLocalhost {
		p.log.Infof("Localhost proxy disabled")
		p.proxy.OnRequest(goproxy.IsLocalHost).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return nil, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadGateway, "Can't use proxy for local addresses")
		})
	} else {
		p.log.Infof("Localhost proxy enabled")
		p.proxy.OnRequest(goproxy.IsLocalHost).HandleConnect(goproxy.AlwaysMitm)
	}
}

func (p *Proxy) setupDirect() {
	p.proxy.Tr.Proxy = nil
	p.proxy.ConnectDial = nil
}

func (p *Proxy) setupUpstreamProxy() {
	p.log.Infof("Using upstream proxy %s", p.config.UpstreamProxyURI)

	p.proxy.OnRequest(goproxy.IsLocalHost).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		addProxyBasicAuthHeader(r, p.config.UpstreamProxyURI.User)
		return r, nil
	})
	p.proxy.Tr.Proxy = http.ProxyURL(p.config.UpstreamProxyURI)

	p.proxy.ConnectDial = p.proxy.NewConnectDialToProxyWithHandler(p.config.UpstreamProxyURI.String(), func(r *http.Request) {
		addProxyBasicAuthHeader(r, p.config.UpstreamProxyURI.User)
	})
}

func (p *Proxy) setupPACProxy() {
	p.log.Infof("Using PAC proxy %s", p.config.PACURI)

	p.proxy.Tr.Proxy = func(r *http.Request) (*url.URL, error) {
		return p.pacFindProxy(r.URL)
	}
	p.proxy.ConnectDialWithReq = func(req *http.Request, network string, addr string) (net.Conn, error) {
		proxy, err := p.pacFindProxy(req.URL)
		if err != nil {
			return nil, err
		}
		if proxy != nil {
			return p.proxy.NewConnectDialToProxy(proxy.String())(network, addr)
		}

		return net.Dial(network, addr)
	}
}

func (p *Proxy) pacFindProxy(u *url.URL) (*url.URL, error) {
	proxies, err := p.pacParser.FindProxy(u.String())
	if err != nil {
		return nil, err
	}

	// No proxy found.
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxy found")
	}

	up := proxies[0].GetURI()
	p.log.Debugf("Using proxy %s for %s", up.Redacted(), u.Redacted())

	return up, nil
}

// setupProxyBasicAuth enables basic auth for the proxy.
func (p *Proxy) setupProxyBasicAuth() {
	u := p.config.LocalProxyURI.User

	if u == nil || u.Username() == "" {
		return
	}

	const realm = "Forwarder"

	p.log.Infof("Basic auth enabled for realm %q and user %q", realm, u.Username())

	auth.ProxyBasic(p.proxy, realm, func(username, password string) bool {
		pwd, _ := u.Password()
		// Securely compare passwords.
		ok := subtle.ConstantTimeCompare([]byte(u.Username()), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(pwd), []byte(password)) == 1
		if !ok {
			p.log.Infof("invalid credentials for %s", username)
		}
		return ok
	})
}

func (p *Proxy) setupSiteBasicAuth() {
	if p.userInfo == nil {
		return
	}

	p.proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		addBasicAuthHeader(r, p.userInfo.MatchURL(r.URL))
		return r, nil
	})
}

// Config returns a copy of the proxy configuration.
func (p *Proxy) Config() *ProxyConfig {
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

// Run starts the proxy.
// It's safe to call it multiple times - nothing will happen.
func (p *Proxy) Run() error {
	p.log.Infof("Listening on %s", p.config.LocalProxyURI.Host)
	if err := http.ListenAndServe(p.config.LocalProxyURI.Host, p.proxy); err != nil { //nolint:gosec // FIXME https://github.com/saucelabs/forwarder/issues/45
		return fmt.Errorf("start proxy: %w", err)
	}

	return nil
}
