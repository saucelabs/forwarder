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
	"runtime"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/pacman"
)

type HTTPTransportConfig struct {
	// DialTimeout is the maximum amount of time a dial will wait for
	// a connect to complete.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	DialTimeout time.Duration `json:"dial_timeout"`

	// KeepAlive specifies the interval between keep-alive
	// probes for an active network connection.
	// If zero, keep-alive probes are sent with a default value
	// (currently 15 seconds), if supported by the protocol and operating
	// system. Network protocols or operating systems that do
	// not support keep-alives ignore this field.
	// If negative, keep-alive probes are disabled.
	KeepAlive time.Duration `json:"keep_alive"`

	// TLSHandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration `json:"tls_handshake_timeout"`

	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int `json:"max_idle_conns"`

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `json:"max_idle_conns_per_host"`

	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	MaxConnsPerHost int `json:"max_conns_per_host"`

	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeout time.Duration `json:"idle_conn_timeout"`

	// ResponseHeaderTimeout, if non-zero, specifies the amount of
	// time to wait for a server's response headers after fully
	// writing the request (including its body, if any). This
	// time does not include the time to read the response body.
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout"`

	// ExpectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"`
}

func DefaultHTTPTransportConfig() *HTTPTransportConfig {
	// The default values are taken from [hashicorp/go-cleanhttp](https://github.com/hashicorp/go-cleanhttp/blob/a0807dd79fc1680a7b1f2d5a2081d92567aab97d/cleanhttp.go#L19.
	return &HTTPTransportConfig{
		DialTimeout:           30 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}

type ProxyConfig struct {
	// BasicAuth is the username and password for the proxy basic auth.
	BasicAuth *url.Userinfo `json:"basic_auth"`

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

	// HTTP specifies http.Transport configuration used when making HTTP requests.
	// If nil, DefaultHTTPTransportConfig will be used.
	HTTP *HTTPTransportConfig `json:"http"`
}

func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		HTTP: DefaultHTTPTransportConfig(),
	}
}

func (c *ProxyConfig) Validate() error {
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
	if cfg.HTTP == nil {
		cfg.HTTP = DefaultHTTPTransportConfig()
	}

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
	p.log.Infof("Using HTTP transport config: %+v", *p.config.HTTP)

	c := p.config.HTTP
	p.transport = &http.Transport{
		Proxy: nil,
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: c.KeepAlive,
			Resolver:  r,
		}),
		MaxIdleConns:          c.MaxIdleConns,
		IdleConnTimeout:       c.IdleConnTimeout,
		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
		ExpectContinueTimeout: c.ExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
	}

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
	u := p.config.BasicAuth

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

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
