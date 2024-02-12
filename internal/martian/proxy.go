// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package martian

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/internal/martian/mitm"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
	"golang.org/x/net/http/httpguts"
)

// Proxy is an HTTP proxy with support for TLS MITM and customizable behavior.
type Proxy struct {
	RequestModifier
	ResponseModifier

	// RoundTripper specifies the round tripper to use for requests.
	RoundTripper http.RoundTripper

	// DialContext specifies the dial function for creating unencrypted TCP connections.
	// If not set and the RoundTripper is an *http.Transport, the Transport's DialContext is used.
	DialContext func(context.Context, string, string) (net.Conn, error)

	// ProxyURL specifies the upstream proxy to use for requests.
	// If not set and the RoundTripper is an *http.Transport, the Transport's ProxyURL is used.
	ProxyURL func(*http.Request) (*url.URL, error)

	// AllowHTTP disables automatic HTTP to HTTPS upgrades when the listener is TLS.
	AllowHTTP bool

	// RequestIDHeader specifies a special header name that the proxy will use to identify requests.
	// If the header is present in the request, the proxy will associate the value with the request in the logs.
	// If empty, no action is taken, and the proxy will generate a new request ID.
	RequestIDHeader string

	// ConnectRequestModifier modifies CONNECT requests to upstream proxy.
	// If ConnectPassthrough is enabled, this is ignored.
	ConnectRequestModifier func(*http.Request) error

	// ConnectFunc specifies a function to dial network connections for CONNECT requests.
	// Implementations can return ErrConnectFallback to indicate that the CONNECT request should be handled by martian.
	ConnectFunc ConnectFunc

	// ConnectTimeout specifies the maximum amount of time to connect to upstream before cancelling request.
	ConnectTimeout time.Duration

	// MITMConfig is config to use for MITMing of CONNECT requests.
	MITMConfig *mitm.Config

	// MITMFilter specifies a function to determine whether a CONNECT request should be MITMed.
	MITMFilter func(*http.Request) bool

	// MITMTLSHandshakeTimeout specifies the maximum amount of time to wait for a TLS handshake for a MITMed connection.
	// Zero means no timeout.
	MITMTLSHandshakeTimeout time.Duration

	// WithoutWarning disables the warning header added to requests and responses when modifier errors occur.
	WithoutWarning bool

	// ErrorResponse specifies a custom error HTTP response to send when a proxying error occurs.
	ErrorResponse func(req *http.Request, err error) *http.Response

	// IdleTimeout is the maximum amount of time to wait for the
	// next request. If IdleTimeout is zero, the value of ReadTimeout is used.
	// If both are zero, there is no timeout.
	IdleTimeout time.Duration

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout time.Duration

	// BaseContex is the base context for all requests.
	BaseContex context.Context //nolint:containedctx // It's intended to be used as a base context.

	// TestingSkipRoundTrip skips the round trip for requests and returns a 200 OK response.
	TestingSkipRoundTrip bool

	initOnce sync.Once

	conns     sync.WaitGroup
	connsMu   sync.Mutex // protects conns.Add/Wait from concurrent access
	closing   chan bool
	closeOnce sync.Once
}

func (p *Proxy) init() {
	p.initOnce.Do(func() {
		if p.RoundTripper == nil {
			p.RoundTripper = &http.Transport{
				// TODO(adamtanner): This forces the http.Transport to not upgrade requests
				// to HTTP/2 in Go 1.6+. Remove this once Martian can support HTTP/2.
				TLSNextProto:          make(map[string]func(string, *tls.Conn) http.RoundTripper),
				Proxy:                 http.ProxyFromEnvironment,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: time.Second,
			}
		}

		if t, ok := p.RoundTripper.(*http.Transport); ok {
			if p.DialContext == nil {
				p.DialContext = t.DialContext
			} else {
				t.DialContext = p.DialContext
			}
			if p.ProxyURL == nil {
				p.ProxyURL = t.Proxy
			} else {
				t.Proxy = p.ProxyURL
			}
		}

		if p.DialContext == nil {
			p.DialContext = (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext
		}

		if p.BaseContex == nil {
			p.BaseContex = context.Background()
		}

		p.closing = make(chan bool)
	})
}

// Close sets the proxy to the closing state so it stops receiving new connections,
// finishes processing any inflight requests, and closes existing connections without
// reading anymore requests from them.
func (p *Proxy) Close() {
	p.init()

	p.closeOnce.Do(func() {
		log.Infof(context.TODO(), "closing down proxy")

		close(p.closing)

		log.Infof(context.TODO(), "waiting for connections to close")
		p.connsMu.Lock()
		p.conns.Wait()
		p.connsMu.Unlock()
		log.Infof(context.TODO(), "all connections closed")
	})
}

// Closing returns whether the proxy is in the closing state.
func (p *Proxy) Closing() bool {
	select {
	case <-p.closing:
		return true
	default:
		return false
	}
}

// Serve accepts connections from the listener and handles the requests.
func (p *Proxy) Serve(l net.Listener) error {
	defer l.Close()

	p.init()

	var delay time.Duration
	for {
		if p.Closing() {
			return nil
		}

		conn, err := l.Accept()
		if err != nil {
			var nerr net.Error
			if ok := errors.As(err, &nerr); ok && nerr.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if max := time.Second; delay > max {
					delay = max
				}

				log.Debugf(context.TODO(), "temporary error on accept: %v", err)
				time.Sleep(delay)
				continue
			}

			if errors.Is(err, net.ErrClosed) {
				log.Debugf(context.TODO(), "listener closed, returning")
				return err
			}

			log.Errorf(context.TODO(), "failed to accept: %v", err)
			return err
		}
		delay = 0
		log.Debugf(context.TODO(), "accepted connection from %s", conn.RemoteAddr())

		go p.handleLoop(conn)
	}
}

func (p *Proxy) handleLoop(conn net.Conn) {
	p.connsMu.Lock()
	p.conns.Add(1)
	p.connsMu.Unlock()
	defer p.conns.Done()
	defer conn.Close()
	if p.Closing() {
		return
	}

	pc := newProxyConn(p, conn)

	const maxConsecutiveErrors = 5
	errorsN := 0
	for {
		if err := pc.handle(); err != nil {
			if errors.Is(err, errClose) || isCloseable(err) {
				log.Debugf(context.TODO(), "closing connection: %v", conn.RemoteAddr())
				return
			}

			errorsN++
			if errorsN >= maxConsecutiveErrors {
				log.Errorf(context.TODO(), "closing connection after %d consecutive errors: %v", errorsN, err)
				return
			}
		} else {
			errorsN = 0
		}
	}
}

func (p *Proxy) idleTimeout() time.Duration {
	if p.IdleTimeout > 0 {
		return p.IdleTimeout
	}
	return p.ReadTimeout
}

func (p *Proxy) readHeaderTimeout() time.Duration {
	if p.ReadHeaderTimeout > 0 {
		return p.ReadHeaderTimeout
	}
	return p.ReadTimeout
}

func (p *Proxy) modifyRequest(req *http.Request) error {
	if p.RequestModifier == nil {
		return nil
	}
	return p.RequestModifier.ModifyRequest(req)
}

func (p *Proxy) modifyResponse(res *http.Response) error {
	if p.ResponseModifier == nil {
		return nil
	}
	return p.ResponseModifier.ModifyResponse(res)
}

func (p *Proxy) shouldMITM(req *http.Request) bool {
	if p.MITMConfig == nil {
		return false
	}

	if p.MITMFilter != nil {
		return p.MITMFilter(req)
	}

	return true
}

func shouldTerminateTLS(req *http.Request) bool {
	h := req.Header.Get("X-Martian-Terminate-TLS")
	if h == "" {
		return false
	}
	b, _ := strconv.ParseBool(h)
	return b
}

func (p *Proxy) roundTrip(req *http.Request) (*http.Response, error) {
	if p.TestingSkipRoundTrip {
		log.Debugf(req.Context(), "skipping round trip")
		return proxyutil.NewResponse(200, http.NoBody, req), nil
	}

	return p.RoundTripper.RoundTrip(req)
}

func (p *Proxy) errorResponse(req *http.Request, err error) *http.Response {
	var res *http.Response
	if p.ErrorResponse != nil {
		res = p.ErrorResponse(req, err)
	} else {
		res = proxyutil.NewResponse(502, http.NoBody, req)
	}

	if !p.WithoutWarning {
		proxyutil.Warning(res.Header, err)
	}

	return res
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}
	return h.Get("Upgrade")
}
