// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/saucelabs/forwarder/internal/credential"
	"github.com/saucelabs/forwarder/internal/customerror"
	"github.com/saucelabs/forwarder/internal/logger"
	"github.com/saucelabs/forwarder/internal/validation"
)

var (
	ErrFailedToStartProxy = customerror.NewFailedToError("start proxy", "", nil)
	ErrInvalidProxyHost   = customerror.NewInvalidError("proxy host", "", nil)
	ErrInvalidProxyURL    = customerror.NewInvalidError("proxy url", "", nil)
)

// LoggingOptions logging options.
type LoggingOptions = logger.Options

//////
// Helpers
//////

// Loads credentials from environment variables.
func loadCredentialsFromEnvVar(cred, parentProxyCredential *string) {
	credentialEnvVar := os.Getenv("PROXY_CREDENTIAL")
	if credentialEnvVar != "" {
		*cred = credentialEnvVar
	}

	parentProxyCredentialEnvVar := os.Getenv("PROXY_PARENT_CREDENTIAL")
	if parentProxyCredentialEnvVar != "" {
		*parentProxyCredential = parentProxyCredentialEnvVar
	}
}

// Proxy connections. Proxy can be protected with basic auth. It can also
// forward connections to a parent proxy, and authorize connections against
// that.
type Proxy struct {
	// Credential is the basic authentication credential.
	//
	// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#basic_authentication_scheme
	Credential *credential.BasicAuth

	// Host is the combination of the `Hostname`, and `Port`.
	//
	// See: https://developer.mozilla.org/en-US/docs/Web/API/URL/host
	Host string `json:"hostname" validate:"required,gte=10,hostname_port"`

	// ParentProxyURL is a valid URL to reach the parent proxy. Valid means:
	// - Known scheme: http, https, socks, socks5, or quic
	// - Some hostname: min 4 chars (x.io)
	// - Port in a valid range: 80 - 65535.
	ParentProxyURL string

	// ParentProxyCredential is the basic authentication credential to authorize
	// against the parent proxy.
	//
	// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#basic_authentication_scheme
	ParentProxyCredential *credential.BasicAuth

	// Underlying proxy implementation.
	proxy *goproxy.ProxyHttpServer
}

// setupBasicAuth protects proxy with basic auth.
func (p *Proxy) setupBasicAuth(cred string) error {
	// Should be a valid credential.
	c, err := credential.NewBasicAuthFromText(cred)
	if err != nil {
		return err
	}

	p.Credential = c

	// TODO: Allows to set `realm`.
	auth.ProxyBasic(p.proxy, "localhost", func(user, pwd string) bool {
		return user == p.Credential.Username && pwd == p.Credential.Password
	})

	logger.Get().Debugln("Basic auth setup")

	return nil
}

// setupParentProxyConnection forwards connections to the parent proxy.
func (p *Proxy) setupParentProxyConnection(parentProxyURL string) error {
	p.ParentProxyURL = parentProxyURL

	// Should be a valid parent proxy URL.
	if err := validation.Get().Var(p.ParentProxyURL, "proxyURL"); err != nil {
		return customerror.Wrap(ErrInvalidProxyURL, err)
	}

	p.proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return url.Parse(p.ParentProxyURL)
	}

	p.proxy.ConnectDial = p.proxy.NewConnectDialToProxy(p.ParentProxyURL)

	logger.Get().Debuglnf("Forwarding connections to parent proxy at %s", p.ParentProxyURL)

	return nil
}

// Sets proxy basic auth header.
func (p *Proxy) setProxyBasicAuthHeader(req *http.Request) {
	req.Header.Set(
		"Proxy-Authorization",
		fmt.Sprintf("Basic %s", p.ParentProxyCredential.ToBase64()),
	)
}

// setupParentProxyBasicAuth authorizes forwarded connections against the parent
// proxy.
func (p *Proxy) setupParentProxyBasicAuth(parentProxyCredential string) error {
	// Should be a valid credential.
	pPC, err := credential.NewBasicAuthFromText(parentProxyCredential)
	if err != nil {
		return err
	}

	p.ParentProxyCredential = pPC

	connectReqHandler := func(req *http.Request) {
		p.setProxyBasicAuthHeader(req)
	}

	p.proxy.ConnectDial = p.proxy.NewConnectDialToProxyWithHandler(p.ParentProxyURL, connectReqHandler)

	p.proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			p.setProxyBasicAuthHeader(req)

			return req, nil
		},
	)

	logger.Get().Debugln("Basic auth setup for parent proxy")

	return nil
}

// Run starts the proxy. If it fails to start, it will exit with fatal.
func (p *Proxy) Run() {
	logger.Get().Infolnf("Proxy started at %s", p.Host)

	// TODO: Allows to pass an error channel.
	if err := http.ListenAndServe(p.Host, p.proxy); err != nil {
		logger.Get().Fatalln(customerror.Wrap(ErrFailedToStartProxy, err))
	}
}

//////
// Factory
//////

// New is the Proxy factory. Errors can be introspected, and provide contextual
// information.
func New(
	host, cred, parentProxyURL, parentProxyCredential string,
	loggingOptions *LoggingOptions,
) (*Proxy, error) {
	// Setup components.
	validation.Setup()

	logger.Setup(loggingOptions)

	loadCredentialsFromEnvVar(&cred, &parentProxyCredential)

	// Instantiate proxy with minimum requirement.
	p := &Proxy{
		Host: host,
	}

	// Should be a valid host.
	if err := validation.Get().Struct(p); err != nil {
		return nil, customerror.Wrap(ErrInvalidProxyHost, err)
	}

	// Instantiate underlying proxy implementation. It can be abstracted in the
	// future to allow easy swapping.
	proxy := goproxy.NewProxyHttpServer()

	// TODO: Setup logger.
	if loggingOptions != nil &&
		(loggingOptions.Level != "info" ||
			loggingOptions.FileLevel != "info") {
		proxy.Verbose = true
	}

	// TODO: Do we need this?
	// proxy.KeepDestinationHeaders = true

	// TODO: This is required, otherwise calls to localhost breaks.
	// TODO: See: https://maelvls.dev/go-ignores-proxy-localhost/
	// TODO: See: https://github.com/golang/go/issues/28866
	// TODO: See: https://github.com/elazarl/goproxy/issues/306
	proxy.KeepHeader = true

	p.proxy = proxy

	// Should be able to protect proxy with Basic Auth.
	// TODO: Don't auth if localhost.
	if cred != "" {
		if err := p.setupBasicAuth(cred); err != nil {
			return nil, err
		}
	}

	// Should be able to forward connection to the parent proxy.
	if parentProxyURL != "" {
		if err := p.setupParentProxyConnection(parentProxyURL); err != nil {
			return nil, err
		}

		// Should be able to authorize against a protected parent proxy.
		if parentProxyCredential != "" {
			if err := p.setupParentProxyBasicAuth(parentProxyCredential); err != nil {
				return nil, err
			}
		}
	}

	return p, nil
}
