// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/forwarder/internal/credential"
	"github.com/saucelabs/forwarder/internal/logger"
	"github.com/saucelabs/forwarder/internal/pac"
	"github.com/saucelabs/forwarder/internal/validation"
	"github.com/saucelabs/randomness"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
)

const (
	MaxRetry        = 3
	ConstantBackoff = 300
)

// Possible ways to run Forwarder.
type Mode string

const (
	Direct   Mode = "DIRECT"
	Upstream Mode = "Upstream"
	PAC      Mode = "PAC"
)

// Valid proxy schemes.
type Schemes string

const (
	HTTP   Schemes = "http"
	HTTPS  Schemes = "https"
	SOCKS5 Schemes = "socks5"
	SOCKS  Schemes = "socks"
	QUIC   Schemes = "quic"
)

var (
	ErrFailedToAllocatePort    = customerror.New("No available port to use", "", http.StatusInternalServerError, nil)
	ErrFailedToStartProxy      = customerror.NewFailedToError("start proxy", "", nil)
	ErrInvalidLocalProxyURI    = customerror.NewInvalidError("local proxy URI", "", nil)
	ErrInvalidOrParentOrPac    = customerror.NewInvalidError("params. Can't set upstream proxy, and PAC at the same time", "", nil)
	ErrInvalidPACProxyURI      = customerror.NewInvalidError("PAC proxy URI", "", nil)
	ErrInvalidPACURI           = customerror.NewInvalidError("PAC URI", "", nil)
	ErrInvalidProxyParams      = customerror.NewInvalidError("params", "", nil)
	ErrInvalidUpstreamProxyURI = customerror.NewInvalidError("upstream proxy URI", "", nil)
)

// LoggingOptions defines logging options.
type LoggingOptions = logger.Options

// RetryPortOptions defines port's retry options.
type RetryPortOptions struct {
	// MaxRange defines the max port number. Default value is `65535`.
	MaxRange int

	// MaxRetry defines how many times to retry, until fail.
	MaxRetry int
}

// Default sets `RetryPortOptions` default values.
func (r *RetryPortOptions) Default() *RetryPortOptions {
	if r == nil {
		r = &RetryPortOptions{}
	}

	if r.MaxRange < 80 || r.MaxRange > 65535 {
		r.MaxRange = 65535
	}

	if r.MaxRetry == 0 {
		r.MaxRetry = 3
	}

	return r
}

// Options definition.
type Options struct {
	*LoggingOptions

	*RetryPortOptions

	// AutomaticallyRetryPort if `true`, and the specified port is in-use, will
	// try to automatically allocate a free port.
	AutomaticallyRetryPort bool

	// ProxyLocalhost if `true`, requests to `localhost`/`127.0.0.1` will be
	// forwarded to any upstream - if set.
	ProxyLocalhost bool
}

// Default sets `Options` default values.
func (o *Options) Default() {
	loggingOptions := &LoggingOptions{}
	loggingOptions.Default()

	retryPortOptions := &RetryPortOptions{}
	retryPortOptions.Default()

	o.AutomaticallyRetryPort = false
	o.ProxyLocalhost = false

	o.LoggingOptions = loggingOptions
	o.RetryPortOptions = retryPortOptions
}

// Proxy definition. Proxy can be protected, or not. It can forward connections
// to an upstream proxy protected, or not. The upstream proxy can be
// automatically setup via PAC. PAC content can be retrieved from multiple
// sources, e.g.: a HTTP server, also, protected or not.
//
// Protection means basic auth protection.
//
// TODO: Add name to `Proxy`.
type Proxy struct {
	// LocalProxyURI is local proxy URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars) or IP
	// - Port in a valid range: 80 - 65535.
	LocalProxyURI string `json:"uri" validate:"required,proxyURI"`

	// UpstreamProxyURI is the upstream proxy URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars) or IP
	// - Port in a valid range: 80 - 65535.
	UpstreamProxyURI string `json:"upstream_proxy_uri" validate:"omitempty,proxyURI"`

	// PACURI is PAC URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars) or IP
	// - Port in a valid range: 80 - 65535.
	PACURI string `json:"pac_uri" validate:"omitempty,gte=6"`

	// Options to setup proxy.
	*Options

	// Mode the Proxy is running.
	Mode Mode

	// Parsed local proxy URI.
	parsedLocalProxyURI *url.URL

	// Parsed upstream proxy URI.
	parsedUpstreamProxyURI *url.URL

	// PAC parser implementation.
	pacParser *pac.Parser

	// Credentials for proxies specified in PAC content.
	pacProxiesCredentials []string

	// Underlying proxy implementation.
	proxy *goproxy.ProxyHttpServer
}

// Sets the `Proxy-Authorization` header based on `uri` user info.
func setProxyBasicAuthHeader(uri *url.URL, req *http.Request) {
	encodedCredential := base64.
		StdEncoding.
		EncodeToString([]byte(uri.User.String()))

	req.Header.Set(
		"Proxy-Authorization",
		fmt.Sprintf("Basic %s", encodedCredential),
	)

	logger.Get().Debuglnf(
		"Proxy-Authorization header set with %s:*** for url %s",
		uri.User.Username(),
		req.URL.String(),
	)
}

// Removes any upstream proxy settings.
//nolint:gosec
func resetUpstreamSettings(ctx *goproxy.ProxyCtx) {
	ctx.Proxy.ConnectDial = nil
	ctx.Proxy.Tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, Proxy: nil}
}

// Returns `true` if should NOT proxy connections to any upstream proxy.
func (p *Proxy) shouldNotProxyLocalhost(ctx *goproxy.ProxyCtx) bool {
	if (strings.Contains(ctx.Req.URL.Hostname(), "127.0.0.1") ||
		strings.Contains(ctx.Req.URL.Hostname(), "localhost")) &&
		!p.ProxyLocalhost {
		resetUpstreamSettings(ctx)

		return true
	}

	return false
}

// setupUpstreamProxyConnection forwards connections to an upstream proxy.
func setupUpstreamProxyConnection(ctx *goproxy.ProxyCtx, uri *url.URL) {
	ctx.Proxy.Tr.Proxy = http.ProxyURL(uri)

	var connectReqHandler func(req *http.Request)

	if uri.User.Username() != "" {
		connectReqHandler = func(req *http.Request) {
			logger.Get().Traceln("Setting basic auth header from connection handler to parent proxy.")

			setProxyBasicAuthHeader(uri, req)
		}
	}

	ctx.Proxy.ConnectDial = ctx.Proxy.NewConnectDialToProxyWithHandler(uri.String(), connectReqHandler)

	logger.Get().Debuglnf("Setup up forwarding connections to %s", uri.Redacted())
}

// setupUpstreamProxyConnection dynamically forwards connections to an upstream
// proxy setup via PAC.
func setupPACUpstreamProxyConnection(p *Proxy, ctx *goproxy.ProxyCtx) error {
	urlToFindProxyFor := ctx.Req.URL.String()

	logger.Get().Debuglnf("Finding proxy for %s", urlToFindProxyFor)

	pacProxies, err := p.pacParser.Find(urlToFindProxyFor)
	if err != nil {
		return err
	}

	// Should only do something if there's any proxy
	if len(pacProxies) > 0 {
		// TODO: Should find the best proxy from a list of possible proxies?
		pacProxy := pacProxies[0]
		pacProxyURI := pacProxy.GetURI()

		// Should only do something if there's a proxy for the given URL, not
		// `DIRECT`.
		if pacProxyURI != nil {
			setupUpstreamProxyConnection(ctx, pacProxyURI)
		}
	} else {
		logger.Get().Debugln("Found no proxy for", urlToFindProxyFor)
	}

	return nil
}

// DRY on handler's code.
// nolint:exhaustive
func (p *Proxy) setupHandlers(ctx *goproxy.ProxyCtx) error {
	if p.shouldNotProxyLocalhost(ctx) {
		logger.Get().Tracelnf("ProxyLocalhost option disabled. Not proxifying request to %s", ctx.Req.URL.String())

		return nil
	}

	switch p.Mode {
	case Upstream:
		setupUpstreamProxyConnection(ctx, p.parsedUpstreamProxyURI)
	case PAC:
		if err := setupPACUpstreamProxyConnection(p, ctx); err != nil {
			return err
		}
	}

	return nil
}

// setupBasicAuth protects proxy with basic auth.
func (p *Proxy) setupBasicAuth(uri *url.URL) error {
	// Should be a valid credential.
	// TODO: Use URI instead of credential.
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

// Verifies if the port in `address` is in-use.
func (p *Proxy) isPortInUse(host string) bool {
	ln, err := net.Listen("tcp", host)
	if err != nil {
		return true
	}

	if ln != nil {
		ln.Close()
	}

	return false
}

// Finds available port.
func (p *Proxy) findAvailablePort(uri *url.URL) error {
	portInt, err := strconv.Atoi(uri.Port())
	if err != nil {
		return err
	}

	possiblePorts := []int64{int64(portInt)}

	r, err := randomness.New(portInt, p.Options.RetryPortOptions.MaxRange, MaxRetry, true)
	if err != nil {
		return err
	}

	// N times to retry option
	randomPorts, err := r.GenerateMany(p.Options.RetryPortOptions.MaxRetry)
	if err != nil {
		return err
	}

	possiblePorts = append(possiblePorts, randomPorts...)

	availablePorts := []int64{}

	// Find some available port.
	for _, port := range possiblePorts {
		isPortInUse := p.isPortInUse(net.JoinHostPort(uri.Hostname(), fmt.Sprintf("%d", port)))

		logger.Get().Tracelnf("Is %d available? %v", port, !isPortInUse)

		if !isPortInUse {
			availablePorts = append(availablePorts, port)

			logger.Get().Tracelnf("Added %d to available ports", port)
		}
	}

	// Any available port?
	if len(availablePorts) < 1 {
		return ErrFailedToAllocatePort
	}

	// Updates data struct.
	uri.Host = net.JoinHostPort(uri.Hostname(), fmt.Sprintf("%d", availablePorts[0]))

	p.parsedLocalProxyURI = uri

	logger.Get().PrintlnfWithOptions(&options.Options{
		Fields: fields.Fields{
			"availablePorts": availablePorts,
		},
	}, level.Debug, "Updated URI with new available port: %s", uri.String())

	return nil
}

// Run starts the proxy. If it fails to start, it will exit with fatal.
func (p *Proxy) Run() {
	// TODO: Allows to pass an error channel.
	if p.Options.AutomaticallyRetryPort && p.isPortInUse(p.parsedLocalProxyURI.Host) {
		r := retrier.New(retrier.ConstantBackoff(
			p.Options.RetryPortOptions.MaxRetry,
			ConstantBackoff*time.Millisecond,
		), nil)

		err := r.Run(func() error {
			return p.findAvailablePort(p.parsedLocalProxyURI)
		})
		if err != nil {
			logger.Get().Fatalln(customerror.Wrap(ErrFailedToStartProxy, err))
		}
	}

	logger.Get().Infolnf("Proxy to start at %s", p.parsedLocalProxyURI.Host)

	if err := http.ListenAndServe(p.parsedLocalProxyURI.Host, p.proxy); err != nil {
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
	options *Options,
) (*Proxy, error) {
	// Instantiate validator.
	validation.Setup()

	//////
	// Proxy setup.
	//////

	finalOptions := &Options{}
	finalOptions.Default()

	if options == nil {
		options = &Options{}
	}

	if err := deepCopy(options, finalOptions); err != nil {
		return nil, err
	}

	p := &Proxy{
		LocalProxyURI:         localProxyURI,
		UpstreamProxyURI:      upstreamProxyURI,
		PACURI:                pacURI,
		pacProxiesCredentials: pacProxiesCredentials,
		Mode:                  Direct,
		Options:               finalOptions,
	}

	if err := validation.Get().Struct(p); err != nil {
		return nil, customerror.Wrap(ErrInvalidProxyParams, err)
	}

	logger.Setup(finalOptions.LoggingOptions)

	// Can't have parent proxy configuration, and PAC at the same time.
	if p.UpstreamProxyURI != "" && p.PACURI != "" {
		return nil, ErrInvalidOrParentOrPac
	}

	//////
	// Underlying proxy implementation setup.
	//////

	// Instantiate underlying proxy implementation. It can be abstracted in the
	// future to allow easy swapping.
	proxy := goproxy.NewProxyHttpServer()

	if p.Options.LoggingOptions != nil && level.MustFromString(p.Options.LoggingOptions.Level) > level.Info {
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

	if p.UpstreamProxyURI != "" {
		p.Mode = Upstream

		parsedUpstreamProxyURI, err := url.ParseRequestURI(p.UpstreamProxyURI)
		if err != nil {
			return nil, customerror.Wrap(ErrInvalidUpstreamProxyURI, err)
		}

		err = loadCredentialFromEnvVar("FORWARDER_UPSTREAMPROXY_CREDENTIAL", parsedUpstreamProxyURI)
		if err != nil {
			return nil, err
		}

		p.parsedUpstreamProxyURI = parsedUpstreamProxyURI
	}

	if p.PACURI != "" {
		p.Mode = PAC

		// `uri` doesn't need to be validated, this is already done by `pac.New`.
		// Also, there's no need to wrap `err`, pac is powered by `customerror`.
		pacParser, err := pac.New(p.PACURI, p.pacProxiesCredentials...)
		if err != nil {
			return nil, err
		}

		p.pacParser = pacParser
	}

	// HTTPS handler.
	p.proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		logger.Get().Debugln("Request handled by the HTTPS handler")

		if err := p.setupHandlers(ctx); err != nil {
			logger.Get().Errorlnf("Failed to setup handler (HTTPS) for request %s. %+v", ctx.Req.URL.String(), err)

			return goproxy.RejectConnect, host
		}

		return nil, host
	})

	// HTTP handler.
	p.proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		logger.Get().Debugln("Request handled by the HTTP handler")

		if err := p.setupHandlers(ctx); err != nil {
			logger.Get().Errorlnf("Failed to setup handler (HTTP) for request %s. %+v", ctx.Req.URL.String(), err)

			return nil, goproxy.NewResponse(
				ctx.Req,
				goproxy.ContentTypeText,
				http.StatusInternalServerError,
				err.Error(),
			)
		}

		return ctx.Req, nil
	})

	// Local proxy authentication.
	if parsedLocalProxyURI.User.Username() != "" {
		if err := p.setupBasicAuth(parsedLocalProxyURI); err != nil {
			return nil, err
		}
	}

	return p, nil
}
