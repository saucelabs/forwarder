// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
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

const (
	DNSTimeout = 1 * time.Minute
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

func NewProxy(cfg *ProxyConfig, log Logger) (*Proxy, error) {
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
	p.setupDNS()

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

func (p *Proxy) setupDNS() {
	d := net.Dialer{
		Timeout: DNSTimeout,
	}
	setupDNS(p.config.DNSURIs, &d, p.log)
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
	p.proxy.Tr.Proxy = func(request *http.Request) (*url.URL, error) {
		return nil, nil //nolint:nilnil // nil url means direct connection
	}
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
