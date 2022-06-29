// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
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
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
)

const (
	ConstantBackoff = 300
	DNSTimeout      = 1 * time.Minute
	MaxRetry        = 3
)

// Possible ways to run Forwarder.
type Mode string

const (
	Direct   Mode = "DIRECT"
	Upstream Mode = "Upstream"
	PAC      Mode = "PAC"
)

// Valid proxy schemes.
const (
	HTTP   = "http"
	HTTPS  = "https"
	SOCKS5 = "socks5"
	SOCKS  = "socks"
	QUIC   = "quic"
)

// State helps the proxy to don't run the same state multiple times.
type State string

const (
	// Initializing means that a new proxy has been instantiated, but has not
	// yet finished setup.
	Initializing State = "Initializing"

	// Setup state means it's done setting it up, but not running yet.
	Setup State = "Setup"

	// Running means proxy is running.
	Running State = "Running"
)

var (
	ErrFailedToAllocatePort    = customerror.New("No available port to use", customerror.WithStatusCode(http.StatusInternalServerError))
	ErrFailedToDialToDNS       = customerror.NewFailedToError("dial to DNS")
	ErrFailedToStartProxy      = customerror.NewFailedToError("start proxy")
	ErrInvalidDNSURI           = customerror.NewInvalidError("dns URI")
	ErrInvalidLocalProxyURI    = customerror.NewInvalidError("local proxy URI")
	ErrInvalidOrParentOrPac    = customerror.NewInvalidError("params. Can't set upstream proxy, and PAC at the same time")
	ErrInvalidPACProxyURI      = customerror.NewInvalidError("PAC proxy URI")
	ErrInvalidPACURI           = customerror.NewInvalidError("PAC URI")
	ErrInvalidProxyParams      = customerror.NewInvalidError("params")
	ErrInvalidUpstreamProxyURI = customerror.NewInvalidError("upstream proxy URI")
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
//nolint:maligned
type Options struct {
	*LoggingOptions

	*RetryPortOptions

	// AutomaticallyRetryPort if `true`, and the specified port is in-use, will
	// try to automatically allocate a free port.
	AutomaticallyRetryPort bool

	// DNSURIs are DNS URIs:
	// - Known protocol: udp, tcp
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 53 - 65535.
	// Example: udp://10.0.0.3:53
	DNSURIs []string `json:"dns_uris" validate:"omitempty,dive,dnsURI"`

	// ProxyLocalhost if `true`, requests to `localhost`/`127.0.0.1` will be
	// forwarded to any upstream - if set.
	ProxyLocalhost bool
}

// Default sets `Options` default values.
func (o *Options) Default() {
	loggingOptions := &LoggingOptions{}
	loggingOptions.Default()

	o.LoggingOptions = loggingOptions

	retryPortOptions := &RetryPortOptions{}
	retryPortOptions.Default()

	o.AutomaticallyRetryPort = false
	o.ProxyLocalhost = false
	o.RetryPortOptions = retryPortOptions
}

// Proxy definition. Proxy can be protected, or not. It can forward connections
// to an upstream proxy protected, or not. The upstream proxy can be
// automatically setup via PAC. PAC content can be retrieved from multiple
// sources, e.g.: a HTTP server, also, protected or not.
//
// Protection means basic auth protection.
type Proxy struct {
	// LocalProxyURI is the local proxy URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 80 - 65535.
	// Example: http://127.0.0.1:8080
	LocalProxyURI string `json:"local_proxy_uri" validate:"required,proxyURI"`

	// UpstreamProxyURI is the upstream proxy URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 80 - 65535.
	// Example: http://u456:p456@127.0.0.1:8085
	UpstreamProxyURI string `json:"upstream_proxy_uri" validate:"omitempty,proxyURI"`

	// PACURI is the PAC URI:
	// - Known schemes: http, https, socks, socks5, or quic
	// - Some hostname (x.io - min 4 chars), or IP
	// - Port in a valid range: 80 - 65535.
	// Example: http://127.0.0.1:8087/data.pac
	PACURI string `json:"pac_uri" validate:"omitempty,gte=6"`

	// Mode the Proxy is running.
	Mode Mode

	// Current state of the proxy. Multiple calls to `Run`, if running, will do
	// nothing.
	State State

	// Options to setup proxy.
	*Options

	mutex *sync.RWMutex

	// Parsed local proxy URI.
	parsedLocalProxyURI *url.URL

	// Parsed upstream proxy URI.
	parsedUpstreamProxyURI *url.URL

	// PAC parser implementation.
	pacParser *pac.Parser

	// Credentials for proxies specified in PAC content.
	pacProxiesCredentials []string

	// Credentials for passing basic authentication to requests
	siteCredentials map[string]string

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

// Sets the default DNS.
//nolint:interfacer
func setupDNS(mutex *sync.RWMutex, dnsURIs []string) error {
	mutex.Lock()
	defer mutex.Unlock()

	parsedDNSURIs := []*url.URL{}

	for _, dnsURI := range dnsURIs {
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

					logger.Get().Tracelnf(customerror.NewFailedToError(errMsg, customerror.WithError(err)).Error())
				} else {
					logger.Get().Tracelnf("Request resolved by DNS @ %s", parsedDNSURI)

					break
				}
			}

			if finalError != nil {
				ErrAllDNSResolversFailed := customerror.New(
					"All DNS resolvers failed",
					customerror.WithStatusCode(http.StatusInternalServerError),
					customerror.WithError(finalError),
				)

				logger.Get().Traceln(ErrAllDNSResolversFailed)

				return finalConn, ErrAllDNSResolversFailed
			}

			return finalConn, nil
		},
	}

	return nil
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
			logger.Get().Traceln("Setting basic auth header from connection handler to upstream proxy.")

			setProxyBasicAuthHeader(uri, req)
		}
	}

	ctx.Proxy.ConnectDial = ctx.Proxy.NewConnectDialToProxyWithHandler(uri.String(), connectReqHandler)

	logger.Get().Tracelnf("Connection to the upstream proxy %s is set up", uri.Redacted())
}

// setupUpstreamProxyConnection dynamically forwards connections to an upstream
// proxy setup via PAC.
func setupPACUpstreamProxyConnection(p *Proxy, ctx *goproxy.ProxyCtx) error {
	urlToFindProxyFor := ctx.Req.URL.String()

	logger.Get().Tracelnf("Finding proxy for %s", urlToFindProxyFor)

	pacProxies, err := p.pacParser.Find(urlToFindProxyFor)
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
			setupUpstreamProxyConnection(ctx, pacProxyURI)

			return nil
		}
	}

	logger.Get().Debugln("Found no proxy for", urlToFindProxyFor)
	// Clear upstream proxy settings (if any) for this request.
	resetUpstreamSettings(ctx)

	return nil
}

// encodeSiteCredentials converts credentials (strings of "user:pass") into
// base64 encoded strings to be used as basic authentication headers.
func encodeSiteCredentials(creds string) (string, string, error) {
	tokens := strings.Split(creds, ":")
	if len(tokens) != 4 { //nolint
		return "", "", fmt.Errorf("failed to parse %s as site auth", creds)
	}

	for _, token := range tokens {
		if len(token) == 0 {
			return "", "", fmt.Errorf("failed to find credentials in %s", creds)
		}
	}

	encodedCredential, err := credential.NewBasicAuth(tokens[2], tokens[3])
	if err != nil {
		return "", "", fmt.Errorf("failed to parse credentials from %s", creds)
	}

	return fmt.Sprintf("%s:%s", tokens[0], tokens[1]), encodedCredential.ToBase64(), nil
}

func parseSiteCredentials(creds []string) (map[string]string, error) {
	credMap := make(map[string]string, len(creds))

	for _, credentialText := range creds {
		hostport, creds, err := encodeSiteCredentials(credentialText)
		if err != nil {
			return nil, err
		}

		_, found := credMap[hostport]
		if found {
			return nil, fmt.Errorf("multiple credentials for %s", hostport)
		}

		credMap[hostport] = creds
	}

	return credMap, nil
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

// Run starts the proxy. it fails to start, it will exit with fatal. It's safe
// to call it multiple times - nothing will happen.
func (p *Proxy) Run() {
	// Should not panic, but exit with proper error if method is called without
	// Proxy is setup.
	if p == nil {
		logger.Get().Fatalln(ErrFailedToStartProxy, "Proxy isn't set up")
	}

	// Do nothing if already running.
	p.mutex.RLock()
	if p.State == Running {
		logger.Get().Traceln("Proxy is already running")

		return
	}
	p.mutex.RUnlock()

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

	logger.Get().Debuglnf("Listening on %s", p.parsedLocalProxyURI.Host)

	// Updates state.
	p.mutex.Lock()
	p.State = Running
	p.mutex.Unlock()

	if err := http.ListenAndServe(p.parsedLocalProxyURI.Host, p.proxy); err != nil {
		logger.Get().Fatalln(customerror.Wrap(ErrFailedToStartProxy, err))
	}
}

func (p *Proxy) setupProxyHandlers() {
	p.proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		logger.Get().Debuglnf("%s %s -> %s", ctx.Req.Method, ctx.Req.RemoteAddr, ctx.Req.Host)
		logger.Get().Debuglnf("%q", dumpHeaders(ctx.Req))

		if err := p.setupHandlers(ctx); err != nil {
			logger.Get().Errorlnf("Failed to setup handler (HTTPS) for request %s. %+v", ctx.Req.URL.Redacted(), err)

			return goproxy.RejectConnect, host
		}

		return nil, host
	})

	p.proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		logger.Get().Debuglnf("%s %s -> %s %s %s", req.Method, req.RemoteAddr, req.URL.Scheme, req.Host, req.URL.Port())
		logger.Get().Tracelnf("%q", dumpHeaders(ctx.Req))

		if err := p.setupHandlers(ctx); err != nil {
			logger.Get().Errorlnf("Failed to setup handler (HTTP) for request %s. %+v", ctx.Req.URL.Redacted(), err)

			return nil, goproxy.NewResponse(
				ctx.Req,
				goproxy.ContentTypeText,
				http.StatusInternalServerError,
				err.Error(),
			)
		}

		p.addAuthHeader(req)

		return ctx.Req, nil
	})

	p.proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp != nil {
			logger.Get().Debuglnf("%s <- %s %v (%v bytes)",
				resp.Request.RemoteAddr, resp.Request.Host, resp.Status, resp.ContentLength)
		} else {
			logger.Get().Tracelnf("%s <- %s response is empty", ctx.Req.Host, ctx.Req.RemoteAddr)
		}

		return resp
	})
}

// addAuthHeader modifies the request and adds an authorization header if necessary.
func (p *Proxy) addAuthHeader(req *http.Request) {
	hostport := req.Host

	if req.URL.Port() == "" {
		var port string
		if req.URL.Scheme == "http" {
			port = "80"
		}

		hostport = fmt.Sprintf("%s:%s", req.Host, port)
	}

	// If req.Host is in the auth map, add the basic auth header
	// using the credentials. These credentials are already base64
	// encoded.
	creds, found := p.siteCredentials[hostport]
	if found {
		logger.Get().Tracelnf("Found site credentials for %s", req.Host)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", creds))
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
	siteCredentials []string,
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

	// Will not copy logger reference, so, storing a reference.
	var externalLogger sypl.Sypl

	if options.LoggingOptions.Logger != nil {
		externalLogger = *options.LoggingOptions.Logger
	}

	if err := deepCopy(options, finalOptions); err != nil {
		return nil, err
	}

	// Parse site credential list into map of host:port -> base64 encoded credentials.
	creds, err := parseSiteCredentials(siteCredentials)
	if err != nil {
		return nil, err
	}

	p := &Proxy{
		LocalProxyURI:         localProxyURI,
		Mode:                  Direct,
		Options:               finalOptions,
		PACURI:                pacURI,
		State:                 Initializing,
		UpstreamProxyURI:      upstreamProxyURI,
		pacProxiesCredentials: pacProxiesCredentials,
		mutex:                 &sync.RWMutex{},
		siteCredentials:       creds,
	}

	if err := validation.Get().Struct(p); err != nil {
		return nil, customerror.Wrap(ErrInvalidProxyParams, err)
	}

	// Should allow to set logger.
	if options.LoggingOptions.Logger != nil {
		finalOptions.LoggingOptions.Logger = &externalLogger
	}

	logger.Setup(finalOptions.LoggingOptions)

	// Can't have upstream proxy configuration, and PAC at the same time.
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
		proxyLogger := &logger.ProxyLogger{
			Logger: logger.Get().New("goproxy"),
		}

		proxy.Logger = proxyLogger
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
	// DNS.
	//////

	if p.Options.DNSURIs != nil {
		if err := setupDNS(p.mutex, p.Options.DNSURIs); err != nil {
			return nil, err
		}
	}

	//////
	// Local proxy setup.
	//////

	parsedLocalProxyURI, err := url.ParseRequestURI(p.LocalProxyURI)
	if err != nil {
		return nil, customerror.Wrap(ErrInvalidLocalProxyURI, err)
	}

	err = loadCredentialFromEnvVar("FORWARDER_LOCALPROXY_AUTH", parsedLocalProxyURI)
	if err != nil {
		return nil, err
	}

	p.parsedLocalProxyURI = parsedLocalProxyURI
	p.LocalProxyURI = parsedLocalProxyURI.String()

	if p.UpstreamProxyURI != "" {
		p.Mode = Upstream

		parsedUpstreamProxyURI, err := url.ParseRequestURI(p.UpstreamProxyURI)
		if err != nil {
			return nil, customerror.Wrap(ErrInvalidUpstreamProxyURI, err)
		}

		err = loadCredentialFromEnvVar("FORWARDER_UPSTREAMPROXY_AUTH", parsedUpstreamProxyURI)
		if err != nil {
			return nil, err
		}

		p.parsedUpstreamProxyURI = parsedUpstreamProxyURI
		p.UpstreamProxyURI = parsedUpstreamProxyURI.String()
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

	// Setup the request and response handlers
	p.setupProxyHandlers()

	// Local proxy authentication.
	if parsedLocalProxyURI.User.Username() != "" {
		if err := p.setupBasicAuth(parsedLocalProxyURI); err != nil {
			return nil, err
		}
	}

	// Updates state.
	p.State = Setup

	return p, nil
}
