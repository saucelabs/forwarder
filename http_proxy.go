// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/elazarl/goproxy"
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

type HTTPProxyConfig struct {
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
	// HTTPProxy will add basic auth headers for requests to these URLs.
	SiteCredentials []string `json:"site_credentials"`

	// Transport specifies http.Transport configuration used when making Transport requests.
	// If nil, DefaultHTTPTransportConfig will be used.
	Transport *HTTPTransportConfig `json:"http"`
}

func DefaultHTTPProxyConfig() *HTTPProxyConfig {
	return &HTTPProxyConfig{
		Transport: DefaultHTTPTransportConfig(),
	}
}

func (c *HTTPProxyConfig) Validate() error {
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

// HTTPProxy is a proxy that can be used to make Transport requests.
// It supports upstream proxy and PAC and can add basic auth headers for requests to specific URLs.
type HTTPProxy struct {
	config    HTTPProxyConfig
	transport *http.Transport
	userInfo  *userInfoMatcher
	pacParser *pacman.Parser
	proxy     *goproxy.ProxyHttpServer
	basicAuth *BasicAuthUtil
	log       Logger
}

func NewHTTPProxy(cfg *HTTPProxyConfig, r *net.Resolver, log Logger) (*HTTPProxy, error) {
	if cfg.Transport == nil {
		cfg.Transport = DefaultHTTPTransportConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Parse site credential list into map of host:port -> base64 encoded input.
	m, err := newUserInfoMatcher(cfg.SiteCredentials, log)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	p := &HTTPProxy{
		config:    *cfg,
		userInfo:  m,
		basicAuth: &BasicAuthUtil{Header: ProxyAuthorizationHeader},
		log:       log,
	}
	p.configureTransport(r)

	if p.config.PACURI != nil {
		pacParser, err := pacman.New(p.config.PACURI.String(), p.config.PACProxiesCredentials...)
		if err != nil {
			return nil, fmt.Errorf("pac parser: %w", err)
		}
		p.pacParser = pacParser
	}

	p.configureProxy()

	return p, nil
}

func (hp *HTTPProxy) configureTransport(r *net.Resolver) {
	hp.log.Infof("Using HTTP transport config: %+v", *hp.config.Transport)

	c := hp.config.Transport
	hp.transport = &http.Transport{
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

	hp.transport.Proxy = nil
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func (hp *HTTPProxy) configureProxy() {
	hp.proxy = goproxy.NewProxyHttpServer()
	hp.proxy.Logger = goproxyLogger{hp.log}
	hp.proxy.Verbose = true
	hp.proxy.KeepDestinationHeaders = true
	// This is required.
	//
	// See: https://maelvls.dev/go-ignores-proxy-localhost/
	// See: https://github.com/golang/go/issues/28866
	// See: https://github.com/elazarl/goproxy/issues/306
	hp.proxy.KeepHeader = true

	hp.proxy.Tr = hp.transport.Clone()
	hp.configureLocalhostProxy()

	switch hp.Mode() {
	case Direct:
		hp.configureDirect()
	case Upstream:
		hp.configureUpstreamProxy()
	case PAC:
		hp.configurePACProxy()
	default:
		panic(fmt.Errorf("unknown mode %q", hp.Mode()))
	}

	hp.configureSiteBasicAuth()
}

func (hp *HTTPProxy) configureLocalhostProxy() {
	if !hp.config.ProxyLocalhost {
		hp.log.Infof("Localhost proxy disabled")
		hp.proxy.OnRequest(goproxy.IsLocalHost).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return nil, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadGateway, "Can't use proxy for local addresses")
		})
	} else {
		hp.log.Infof("Localhost proxy enabled")
		hp.proxy.OnRequest(goproxy.IsLocalHost).HandleConnect(goproxy.AlwaysMitm)
	}
}

func (hp *HTTPProxy) configureDirect() {
	hp.proxy.Tr.Proxy = nil
	hp.proxy.ConnectDial = nil
}

func (hp *HTTPProxy) configureUpstreamProxy() {
	hp.log.Infof("Using upstream proxy %s", hp.config.UpstreamProxyURI)

	hp.proxy.OnRequest(goproxy.IsLocalHost).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		hp.basicAuth.SetBasicAuthFromUserInfo(r, hp.config.UpstreamProxyURI.User)
		return r, nil
	})
	hp.proxy.Tr.Proxy = http.ProxyURL(hp.config.UpstreamProxyURI)

	hp.proxy.ConnectDial = hp.proxy.NewConnectDialToProxyWithHandler(hp.config.UpstreamProxyURI.String(), func(r *http.Request) {
		hp.basicAuth.SetBasicAuthFromUserInfo(r, hp.config.UpstreamProxyURI.User)
	})
}

func (hp *HTTPProxy) configurePACProxy() {
	hp.log.Infof("Using PAC proxy %s", hp.config.PACURI)

	hp.proxy.Tr.Proxy = func(r *http.Request) (*url.URL, error) {
		return hp.pacFindProxy(r.URL)
	}
	hp.proxy.ConnectDialWithReq = func(req *http.Request, network string, addr string) (net.Conn, error) {
		proxy, err := hp.pacFindProxy(req.URL)
		if err != nil {
			return nil, err
		}
		if proxy != nil {
			return hp.proxy.NewConnectDialToProxy(proxy.String())(network, addr)
		}

		return net.Dial(network, addr)
	}
}

func (hp *HTTPProxy) pacFindProxy(u *url.URL) (*url.URL, error) {
	proxies, err := hp.pacParser.FindProxy(u.String())
	if err != nil {
		return nil, err
	}

	// No proxy found.
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxy found")
	}

	up := proxies[0].GetURI()
	hp.log.Debugf("Using proxy %s for %s", up.Redacted(), u.Redacted())

	return up, nil
}

func (hp *HTTPProxy) configureSiteBasicAuth() {
	if hp.userInfo == nil {
		return
	}

	hp.proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if u := hp.userInfo.MatchURL(r.URL); u != nil {
			p, _ := u.Password()
			r.SetBasicAuth(u.Username(), p)
		}
		return r, nil
	})
}

// Mode returns mode of operation of the proxy as specified in the config.
func (hp *HTTPProxy) Mode() Mode {
	switch {
	case hp.config.UpstreamProxyURI != nil:
		return Upstream
	case hp.config.PACURI != nil:
		return PAC
	default:
		return Direct
	}
}

func (hp *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hp.proxy.ServeHTTP(w, r)
}
