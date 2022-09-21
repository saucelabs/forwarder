// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/forwarder/validation"
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

var (
	ErrInvalidDNSURI           = customerror.NewInvalidError("dns URI")
	ErrInvalidLocalProxyURI    = customerror.NewInvalidError("local proxy URI")
	ErrInvalidPACProxyURI      = customerror.NewInvalidError("PAC proxy URI")
	ErrInvalidPACURI           = customerror.NewInvalidError("PAC URI")
	ErrInvalidUpstreamProxyURI = customerror.NewInvalidError("upstream proxy URI")
)

// ProxyConfig definition.
type ProxyConfig struct {
	// LocalProxyURI is the local proxy URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 80 - 65535.
	// Example: http://127.0.0.1:8080
	LocalProxyURI string `json:"local_proxy_uri" validate:"required,proxyURI"`

	// LocalProxyAuth is the local proxy basic auth in the form of username:password.
	LocalProxyAuth string `json:"local_proxy_auth" validate:"omitempty,basicAuth"`

	// UpstreamProxyURI is the upstream proxy URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 80 - 65535.
	// Example: http://u456:p456@127.0.0.1:8085
	UpstreamProxyURI string `json:"upstream_proxy_uri" validate:"omitempty,proxyURI,excluded_with=PACURI"`

	// UpstreamProxyAuth is the upstream proxy basic auth in the form of username:password.
	UpstreamProxyAuth string `json:"upstream_proxy_auth" validate:"omitempty,basicAuth"`

	// PACURI is the PAC URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 80 - 65535.
	// Example: http://127.0.0.1:8087/data.pac
	PACURI string `json:"pac_uri" validate:"omitempty,proxyURI,excluded_with=UpstreamProxyURI"`

	// Credentials for proxies specified in PAC content.
	PACProxiesCredentials []string

	// DNSURIs are DNS URIs:
	// - Known protocol: udp, tcp
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 53 - 65535.
	// Example: udp://10.0.0.3:53
	DNSURIs []string `json:"dns_uris" validate:"omitempty,dive,dnsURI"`

	// ProxyLocalhost if `true`, requests to `localhost`/`127.0.0.1` will be
	// forwarded to any upstream - if set.
	ProxyLocalhost bool

	// SiteCredentials contains URLs with the credentials, for example:
	// - https://usr1:pwd1@foo.bar:4443
	// - http://usr2:pwd2@bar.foo:8080
	// - usr3:pwd3@bar.foo:8080
	// Proxy will add basic auth headers for requests to these URLs.
	SiteCredentials []string `json:"site_credentials" validate:"omitempty"`
}

func (c *ProxyConfig) Clone() ProxyConfig {
	var v ProxyConfig
	deepCopy(&v, c)
	return v
}

func (c *ProxyConfig) Validate() error {
	v := validation.Validator()
	return v.Struct(c)
}

// Proxy definition. Proxy can be protected, or not. It can forward connections
// to an upstream proxy protected, or not. The upstream proxy can be
// automatically setup via PAC. PAC content can be retrieved from multiple
// sources, e.g.: a HTTP server, also, protected or not.
//
// Protection means basic auth protection.
type Proxy struct {
	config ProxyConfig

	// Parsed local proxy URI.
	parsedLocalProxyURI *url.URL

	// Parsed upstream proxy URI.
	parsedUpstreamProxyURI *url.URL

	// PAC parser implementation.
	pacParser *pacman.Parser

	// credentials for passing basic authentication to requests
	creds *userInfoMatcher

	// Underlying proxy implementation.
	proxy *goproxy.ProxyHttpServer

	log Logger
}

func NewProxy(cfg ProxyConfig, log Logger) (*Proxy, error) { //nolint // FIXME Function 'NewProxy' has too many statements (67 > 40) (funlen)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	p := &Proxy{
		config: cfg.Clone(),
		log:    log,
	}

	// Parse site credential list into map of host:port -> base64 encoded input.
	creds, err := newUserInfoMatcher(cfg.SiteCredentials, log)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	p.creds = creds

	//////
	// Underlying proxy implementation setup.
	//////

	// Instantiate underlying proxy implementation. It can be abstracted in the
	// future to allow easy swapping.
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Logger = goproxyLogger{log}
	p.proxy.Verbose = true
	p.proxy.KeepDestinationHeaders = true
	// This is required.
	//
	// See: https://maelvls.dev/go-ignores-proxy-localhost/
	// See: https://github.com/golang/go/issues/28866
	// See: https://github.com/elazarl/goproxy/issues/306
	p.proxy.KeepHeader = true

	//////
	// DNS.
	//////

	if p.config.DNSURIs != nil {
		if err := p.setupDNS(); err != nil {
			return nil, err
		}
	}

	//////
	// Local proxy setup.
	//////

	parsedLocalProxyURI, err := url.ParseRequestURI(p.config.LocalProxyURI)
	if err != nil {
		return nil, customerror.Wrap(ErrInvalidLocalProxyURI, err)
	}
	if p.config.LocalProxyAuth != "" {
		u, p, _ := strings.Cut(p.config.LocalProxyAuth, ":") // Data is already validated to contain password.
		parsedLocalProxyURI.User = url.UserPassword(u, p)
	}

	p.parsedLocalProxyURI = parsedLocalProxyURI
	p.config.LocalProxyURI = parsedLocalProxyURI.String()

	if p.config.UpstreamProxyURI != "" {
		parsedUpstreamProxyURI, err := url.ParseRequestURI(p.config.UpstreamProxyURI)
		if err != nil {
			return nil, customerror.Wrap(ErrInvalidUpstreamProxyURI, err)
		}
		if p.config.UpstreamProxyAuth != "" {
			u, p, _ := strings.Cut(p.config.UpstreamProxyAuth, ":") // Data is already validated to contain password.
			parsedUpstreamProxyURI.User = url.UserPassword(u, p)
		}

		p.parsedUpstreamProxyURI = parsedUpstreamProxyURI
		p.config.UpstreamProxyURI = parsedUpstreamProxyURI.String()
	}

	if p.config.PACURI != "" {
		pacParser, err := pacman.New(p.config.PACURI, p.config.PACProxiesCredentials...)
		if err != nil {
			return nil, fmt.Errorf("pac parser: %w", err)
		}
		p.pacParser = pacParser
	}

	// Setup the request and response handlers
	p.setupProxyHandlers()

	// Local proxy authentication.
	if u := parsedLocalProxyURI.User; u.Username() != "" {
		p.setupBasicAuth(u)
	}

	return p, nil
}

// Config returns a copy of the proxy configuration.
func (p *Proxy) Config() ProxyConfig {
	return p.config.Clone()
}

// Mode returns mode of operation of the proxy as specified in the config.
func (p *Proxy) Mode() Mode {
	switch {
	case p.config.UpstreamProxyURI != "":
		return Upstream
	case p.config.PACURI != "":
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

// Sets the default DNS.
func (p *Proxy) setupDNS() error {
	parsedDNSURIs := make([]*url.URL, 0, len(p.config.DNSURIs))
	for _, dnsURI := range p.config.DNSURIs {
		parsedDNSURI, err := url.ParseRequestURI(dnsURI)
		if err != nil {
			return customerror.Wrap(ErrInvalidDNSURI, err)
		}

		parsedDNSURIs = append(parsedDNSURIs, parsedDNSURI)
	}

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: DNSTimeout}

			var finalConn net.Conn
			var finalError error

			for i := 0; i < len(parsedDNSURIs); i++ {
				parsedDNSURI := parsedDNSURIs[i]

				c, err := d.DialContext(ctx, parsedDNSURI.Scheme, parsedDNSURI.Host)

				finalConn = c
				finalError = err

				if err != nil {
					errMsg := fmt.Sprintf("dial to DNS @ %s", parsedDNSURI.String())

					p.log.Debugf(customerror.NewFailedToError(errMsg, customerror.WithError(err)).Error())
				} else {
					p.log.Debugf("Request resolved by DNS @ %s", parsedDNSURI)

					break
				}
			}

			if finalError != nil {
				ErrAllDNSResolversFailed := customerror.New(
					"All DNS resolvers failed",
					customerror.WithStatusCode(http.StatusInternalServerError),
					customerror.WithError(finalError),
				)

				p.log.Debugf("error %s", ErrAllDNSResolversFailed)

				return finalConn, ErrAllDNSResolversFailed
			}

			return finalConn, nil
		},
	}

	return nil
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
		p.setupUpstreamProxyConnection(ctx, p.parsedUpstreamProxyURI)
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
			p.log.Debugf("Incoming request. This proxy (%s) is protected authorized=%v", p.parsedLocalProxyURI.Redacted(), ok)
		}()

		pwd, _ := u.Password()
		// Securely compare passwords.
		return subtle.ConstantTimeCompare([]byte(u.Username()), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(pwd), []byte(password)) == 1
	})

	p.log.Debugf("Basic auth setup for proxy @ %s", p.parsedLocalProxyURI.Redacted())
}

// Run starts the proxy.
// It's safe to call it multiple times - nothing will happen.
func (p *Proxy) Run() error {
	p.log.Infof("Listening on %s", p.parsedLocalProxyURI.Host)
	if err := http.ListenAndServe(p.parsedLocalProxyURI.Host, p.proxy); err != nil { //nolint:gosec // FIXME https://github.com/saucelabs/forwarder/issues/45
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

	if u := p.creds.Match(hostport); u != nil {
		req.Header.Set(authHeader, fmt.Sprintf("Basic %s", userInfoBase64(u)))
	}
}
