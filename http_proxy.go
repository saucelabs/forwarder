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

	"github.com/elazarl/goproxy"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/pacman"
)

type HTTPProxyConfig struct {
	// UpstreamProxy is the upstream proxy , ex. http://user:password@127.0.0.1:8080.
	// Only one of `UpstreamProxy` or `PAC` can be set.
	// Requirements:
	// - Known schemes: http, https, socks, socks5, or quic.
	// - Hostname or IP.
	// - Port in a valid range: 1 - 65535.
	// - Username and password are optional.
	UpstreamProxy *url.URL `json:"upstream_proxy_uri"`

	// UpstreamProxyCredentials allow to override the credentials from the `UpstreamProxy`.
	UpstreamProxyCredentials *url.Userinfo `json:"upstream_proxy_credentials"`

	// PAC is the PAC , which is used to determine the upstream proxy, ex. http://127.0.0.1:8087/data.pac.
	// Only one of `UpstreamProxy` or `PAC` can be set.
	PAC *url.URL `json:"pac_uri"`

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
}

func DefaultHTTPProxyConfig() *HTTPProxyConfig {
	return &HTTPProxyConfig{}
}

func (c *HTTPProxyConfig) Validate() error {
	if err := validateProxyURL(c.UpstreamProxy); err != nil {
		return fmt.Errorf("upstream_proxy_uri: %w", err)
	}
	if err := validatedUserInfo(c.UpstreamProxyCredentials); err != nil {
		return fmt.Errorf("upstream_proxy_credentials: %w", err)
	}
	if err := validateProxyURL(c.PAC); err != nil {
		return fmt.Errorf("pac_uri: %w", err)
	}
	if c.UpstreamProxy != nil && c.PAC != nil {
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
	basicAuth *middleware.BasicAuth
	log       Logger
}

func NewHTTPProxy(cfg *HTTPProxyConfig, t *http.Transport, log Logger) (*HTTPProxy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// If not set, use http.DefaultTransport.
	if t == nil {
		log.Infof("HTTP transport not configured, using standard library default")
		t = http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // we know it's a *http.Transport
	}

	// Parse site credential list into map of host:port -> base64 encoded input.
	m, err := newUserInfoMatcher(cfg.SiteCredentials, log)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	p := &HTTPProxy{
		config:    *cfg,
		transport: t,
		userInfo:  m,
		basicAuth: middleware.NewProxyBasicAuth(),
		log:       log,
	}

	if p.config.PAC != nil {
		pacParser, err := pacman.New(p.config.PAC.String(), p.config.PACProxiesCredentials...)
		if err != nil {
			return nil, fmt.Errorf("pac parser: %w", err)
		}
		p.pacParser = pacParser
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
	hp.log.Infof("Using upstream proxy %s", hp.config.UpstreamProxy)

	hp.proxy.OnRequest(goproxy.IsLocalHost).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if hp.config.UpstreamProxyCredentials != nil {
			hp.basicAuth.SetBasicAuthFromUserInfo(r, hp.config.UpstreamProxyCredentials)
		} else {
			hp.basicAuth.SetBasicAuthFromUserInfo(r, hp.config.UpstreamProxy.User)
		}
		return r, nil
	})
	hp.proxy.Tr.Proxy = http.ProxyURL(hp.config.UpstreamProxy)

	hp.proxy.ConnectDial = hp.proxy.NewConnectDialToProxyWithHandler(hp.config.UpstreamProxy.String(), func(r *http.Request) {
		if hp.config.UpstreamProxyCredentials != nil {
			hp.basicAuth.SetBasicAuthFromUserInfo(r, hp.config.UpstreamProxyCredentials)
		} else {
			hp.basicAuth.SetBasicAuthFromUserInfo(r, hp.config.UpstreamProxy.User)
		}
	})
}

func (hp *HTTPProxy) configurePACProxy() {
	hp.log.Infof("Using PAC proxy %s", hp.config.PAC)

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
	case hp.config.UpstreamProxy != nil:
		return Upstream
	case hp.config.PAC != nil:
		return PAC
	default:
		return Direct
	}
}

func (hp *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hp.proxy.ServeHTTP(w, r)
}
