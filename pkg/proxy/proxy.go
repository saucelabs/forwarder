// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/forwarder/internal/credential"
	"github.com/saucelabs/forwarder/internal/logger"
	"github.com/saucelabs/forwarder/internal/pac"
	"github.com/saucelabs/forwarder/internal/validation"
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
)

var (
	ErrFailedToStartProxy      = customerror.NewFailedToError("start proxy", "", nil)
	ErrInvalidLocalProxyURI    = customerror.NewInvalidError("local proxy URI", "", nil)
	ErrInvalidOrParentOrPac    = customerror.NewInvalidError("params. Can't set upstream proxy, and PAC at the same time", "", nil)
	ErrInvalidPACProxyURI      = customerror.NewInvalidError("PAC proxy URI", "", nil)
	ErrInvalidPACURI           = customerror.NewInvalidError("PAC URI", "", nil)
	ErrInvalidProxyParams      = customerror.NewInvalidError("params", "", nil)
	ErrInvalidUpstreamProxyURI = customerror.NewInvalidError("upstream proxy URI", "", nil)
)

// Type aliasing.
type LoggingOptions = logger.Options

// Proxy connections. Proxy can be protected with basic auth. It can also
// forward connections to a parent proxy, and authorize connections against
// that.
//
// TODO: Add name to `Proxy`.
type Proxy struct {
	// URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars) or IP
	// - Port in a valid range: 80 - 65535.
	LocalProxyURI string `json:"uri" validate:"required,proxyURI"`

	parsedLocalProxyURI *url.URL

	// UpstreamProxyURI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars) or IP
	// - Port in a valid range: 80 - 65535.
	UpstreamProxyURI string `json:"upstream_proxy_uri" validate:"omitempty,proxyURI"`

	parsedUpstreamProxyURI *url.URL

	// PACURI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars) or IP
	// - Port in a valid range: 80 - 65535.
	PACURI string `json:"pac_uri" validate:"omitempty,gte=6"`

	// PAC parser implementation.
	pacParser *pac.Parser

	// Underlying proxy implementation.
	proxy *goproxy.ProxyHttpServer
}

// setupBasicAuth protects proxy with basic auth.
func (p *Proxy) setupBasicAuth(uri *url.URL) error {
	// Should be a valid credential.
	// TODO: Add to Proxy.
	c, err := credential.NewBasicAuthFromText(uri.User.String())
	if err != nil {
		return err
	}

	// TODO: Allows to set `realm`.
	auth.ProxyBasic(p.proxy, "localhost", func(user, pwd string) bool {
		ok := user == c.Username && pwd == c.Password

		logger.Get().PrintlnfWithOptions(&options.Options{
			Fields: fields.Fields{
				"authorized": ok,
			},
		}, level.Trace, "Incoming request. This proxy (%s) is protected", p.parsedLocalProxyURI.Redacted())

		return ok
	})

	logger.Get().Debuglnf("Basic auth setup for proxy @ %s", p.parsedLocalProxyURI.Redacted())

	return nil
}

// setupUpstreamProxyConnection forwards connections to an upstream proxy.
func (p *Proxy) setupUpstreamProxyConnection(uri *url.URL) {
	authRequired := false

	p.proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return uri, nil
	}

	var connectReqHandler func(req *http.Request)

	if uri.User.Username() != "" {
		authRequired = true

		connectReqHandler = func(req *http.Request) {
			logger.Get().Traceln("Setting basic auth header from connection handler to parent proxy.")

			p.setProxyBasicAuthHeader(uri, req)
		}
	}

	p.proxy.ConnectDial = p.proxy.NewConnectDialToProxyWithHandler(uri.String(), connectReqHandler)

	p.proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if authRequired {
				logger.Get().Traceln("Setting basic auth header from OnRequest handler to parent proxy.")

				p.setProxyBasicAuthHeader(uri, req)
			}

			return req, nil
		},
	)

	logger.Get().Debuglnf("Setup up forwarding connections to %s", uri.Redacted())
}

func (p *Proxy) pacHandler(l *sypl.Sypl, req *http.Request) (*http.Request, *http.Response) {
	urlToFindProxyFor := req.URL.String()

	l.Debuglnf("Finding proxy for %s", urlToFindProxyFor)

	pacProxies, err := p.pacParser.Find(urlToFindProxyFor)
	if err != nil {
		// TODO: Log and move on, or log and write error to req?
		return req, nil
	}

	// Should only do something if there's any proxy
	if len(pacProxies) > 0 {
		// TODO: Should find the best proxy from a list of possible
		// proxies?
		pacProxy := pacProxies[0]
		pacProxyURI := pacProxy.GetURI()

		// Should only do something if there's a proxy for the given
		// URL, not `DIRECT`.
		if pacProxyURI != nil {
			p.setupUpstreamProxyConnection(pacProxyURI)
		}
	} else {
		l.Debugln("Found no proxy for", urlToFindProxyFor)
	}

	return req, nil
}

// Sets proxy basic auth header.
func (p *Proxy) setProxyBasicAuthHeader(uri *url.URL, req *http.Request) {
	encodedCredential := base64.
		StdEncoding.
		EncodeToString([]byte(uri.User.String()))

	req.Header.Set(
		"Proxy-Authorization",
		fmt.Sprintf("Basic %s", encodedCredential),
	)

	logger.Get().Debugln("Proxy-Authorization header set")
}

// Run starts the proxy. If it fails to start, it will exit with fatal.
func (p *Proxy) Run() {
	localProxyHost := p.parsedLocalProxyURI.Host

	logger.Get().Infolnf("Proxy started at %s", localProxyHost)

	// TODO: Allows to pass an error channel.
	if err := http.ListenAndServe(localProxyHost, p.proxy); err != nil {
		logger.Get().Fatalln(customerror.Wrap(ErrFailedToStartProxy, err))
	}
}

//////
// Factory
//////

// New is the Proxy factory. Errors can be introspected, and provide contextual
// information.
func New(
	localProxyURI string,
	upstreamProxyURI string,
	pacURI string, pacProxiesCredentials []string,
	loggingOptions *LoggingOptions,
) (*Proxy, error) {
	// Components setup.
	validation.Setup()

	logger.Setup(loggingOptions)

	//////
	// Proxy setup.
	//////

	p := &Proxy{
		LocalProxyURI:    localProxyURI,
		UpstreamProxyURI: upstreamProxyURI,
		PACURI:           pacURI,
	}

	if err := validation.Get().Struct(p); err != nil {
		return nil, customerror.Wrap(ErrInvalidProxyParams, err)
	}

	//////
	// Underlying proxy implementation setup.
	//////

	// Instantiate underlying proxy implementation. It can be abstracted in the
	// future to allow easy swapping.
	proxy := goproxy.NewProxyHttpServer()

	if loggingOptions != nil && level.MustFromString(loggingOptions.Level) > level.Info {
		// TODO: Wrap logger, and implement goproxy's `Printf` interface.
		proxy.Verbose = true
	}

	proxy.KeepDestinationHeaders = true

	// This is required.
	//
	// See: https://maelvls.dev/go-ignores-proxy-localhost/
	// See: https://github.com/golang/go/issues/28866
	// See: https://github.com/elazarl/goproxy/issues/306
	proxy.KeepHeader = true

	p.proxy = proxy

	//////
	// Local proxy setup.
	//////

	parsedLocalProxyURI, err := url.ParseRequestURI(p.LocalProxyURI)
	if err != nil {
		return nil, customerror.Wrap(ErrInvalidLocalProxyURI, err)
	}

	err = loadCredentialFromEnvVar("FORWARDER_LOCALPROXY_CREDENTIAL", parsedLocalProxyURI)
	if err != nil {
		return nil, err
	}

	p.parsedLocalProxyURI = parsedLocalProxyURI

	// Local proxy authentication.
	if parsedLocalProxyURI.User.Username() != "" {
		if err := p.setupBasicAuth(parsedLocalProxyURI); err != nil {
			return nil, err
		}
	}

	//////
	// Upstream proxy setup.
	//////

	// Can't have parent proxy configuration, and PAC at the same time.
	if upstreamProxyURI != "" && pacURI != "" {
		return nil, ErrInvalidOrParentOrPac
	}

	// Should be able to forward connections to an upstream proxy.
	if upstreamProxyURI != "" {
		parsedUpstreamProxyURI, err := url.ParseRequestURI(p.UpstreamProxyURI)
		if err != nil {
			return nil, customerror.Wrap(ErrInvalidUpstreamProxyURI, err)
		}

		err = loadCredentialFromEnvVar("FORWARDER_UPSTREAMPROXY_CREDENTIAL", parsedUpstreamProxyURI)
		if err != nil {
			return nil, err
		}

		p.parsedUpstreamProxyURI = parsedUpstreamProxyURI

		p.setupUpstreamProxyConnection(parsedUpstreamProxyURI)
	} else if pacURI != "" {
		// `uri` doesn't need to be validated, this is already done by `pac.New`.
		// Also, there's no need to wrap `err`, pac is powered by `customerror`.
		pacParser, err := pac.New(pacURI, pacProxiesCredentials...)
		if err != nil {
			return nil, err
		}

		p.pacParser = pacParser

		// Register handler which will call `Find` for every request.
		p.proxy.OnRequest().DoFunc(
			func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				l := logger.Get().New("PAC request handler")

				return p.pacHandler(l, req)
			})
	}

	return p, nil
}
