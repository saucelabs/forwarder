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
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/forwarder/pac"
)

type HTTPProxyConfig struct {
	// UpstreamProxy is the URL of the upstream proxy to use.
	// UpstreamProxy and PAC are mutually exclusive.
	// If not set, no upstream proxy is used.
	UpstreamProxy *url.URL `json:"upstream_proxy_uri"`

	// ProxyLocalhost if `true`, requests to `localhost`, `127.0.0.*`, `0:0:0:0:0:0:0:1` will be forwarded to upstream.
	ProxyLocalhost bool `json:"proxy_localhost"`
}

func DefaultHTTPProxyConfig() *HTTPProxyConfig {
	return &HTTPProxyConfig{}
}

func (c *HTTPProxyConfig) Validate() error {
	if err := validateProxyURL(c.UpstreamProxy); err != nil {
		return fmt.Errorf("upstream_proxy_uri: %w", err)
	}

	return nil
}

// HTTPProxy is a proxy that can be used to make Transport requests.
// It supports upstream proxy and PAC and can add basic auth headers for requests to specific URLs.
type HTTPProxy struct {
	config    HTTPProxyConfig
	pac       PACResolver
	creds     *CredentialsMatcher
	transport *http.Transport
	proxy     *goproxy.ProxyHttpServer
	basicAuth *middleware.BasicAuth
	log       Logger
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

	p := &HTTPProxy{
		config:    *cfg,
		pac:       pr,
		creds:     cm,
		transport: t,
		basicAuth: middleware.NewProxyBasicAuth(),
		log:       log,
	}

	p.configureProxy()

	return p, nil
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func (hp *HTTPProxy) configureProxy() {
	hp.proxy = goproxy.NewProxyHttpServer()
	hp.proxy.Logger = goproxyLogger{hp.log}
	hp.proxy.Verbose = true
	hp.proxy.KeepDestinationHeaders = true

	hp.proxy.Tr = hp.transport.Clone()
	// Use the same dialer as the transport.
	hp.proxy.ConnectDial = nil
	hp.proxy.ConnectDialWithReq = nil
	// Set transport proxy function.
	switch {
	case hp.config.UpstreamProxy != nil:
		u := hp.upstreamProxyURL()
		hp.log.Infof("Using upstream proxy: %s", u.Redacted())
		hp.proxy.Tr.Proxy = http.ProxyURL(u)
	case hp.pac != nil:
		hp.log.Infof("Using PAC proxy")
		hp.proxy.Tr.Proxy = hp.pacProxy
	default:
		hp.log.Infof("Using direct proxy")
		hp.proxy.Tr.Proxy = nil
	}

	hp.configureProxyLocalhost()
	hp.setBasicAuthIfNeeded()
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
	var proxyURL *url.URL
	switch p.Mode {
	case pac.DIRECT:
		proxyURL = nil
	case pac.PROXY:
		proxyURL = &url.URL{
			Scheme: r.URL.Scheme,
			Host:   net.JoinHostPort(p.Host, p.Port),
		}
	case pac.HTTP, pac.HTTPS, pac.SOCKS, pac.SOCKS4, pac.SOCKS5:
		proxyURL = &url.URL{
			Scheme: strings.ToLower(p.Mode.String()),
			Host:   net.JoinHostPort(p.Host, p.Port),
		}
	}

	if u := hp.creds.MatchURL(proxyURL); u != nil {
		proxyURL.User = u
	}

	return proxyURL, nil
}

func (hp *HTTPProxy) configureProxyLocalhost() {
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

func (hp *HTTPProxy) setBasicAuthIfNeeded() {
	if hp.creds == nil {
		return
	}

	hp.proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if u := hp.creds.MatchURL(r.URL); u != nil {
			p, _ := u.Password()
			r.SetBasicAuth(u.Username(), p)
		}
		return r, nil
	})
}

func (hp *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hp.proxy.ServeHTTP(w, r)
}
