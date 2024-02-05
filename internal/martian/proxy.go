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
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/saucelabs/forwarder/dialvia"
	"github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/internal/martian/mitm"
	"github.com/saucelabs/forwarder/internal/martian/nosigpipe"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
	"golang.org/x/net/http/httpguts"
)

var (
	errClose = errors.New("closing connection")
	noop     = Noop("martian")
)

func errno(v error) uintptr {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Uintptr {
		return uintptr(rv.Uint())
	}
	return 0
}

// isClosedConnError reports whether err is an error from use of a closed network connection.
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	// TODO(bradfitz): x/tools/cmd/bundle doesn't really support
	// build tags, so I can't make an http2_windows.go file with
	// Windows-specific stuff. Fix that and move this, once we
	// have a way to bundle this into std's net/http somehow.
	if runtime.GOOS == "windows" {
		var se *os.SyscallError
		if errors.As(err, &se) {
			if se.Syscall == "wsarecv" || se.Syscall == "wsasend" {
				const WSAECONNABORTED = 10053
				const WSAECONNRESET = 10054
				if n := errno(se.Err); n == WSAECONNRESET || n == WSAECONNABORTED {
					return true
				}
			}
		}
	}

	return strings.Contains(err.Error(), "use of closed network connection")
}

// isCloseable reports whether err is an error that indicates the client connection should be closed.
func isCloseable(err error) bool {
	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, io.ErrClosedPipe) {
		return true
	}

	var neterr net.Error
	if ok := errors.As(err, &neterr); ok && neterr.Timeout() {
		return true
	}

	return strings.Contains(err.Error(), "tls:")
}

// Proxy is an HTTP proxy with support for TLS MITM and customizable behavior.
type Proxy struct {
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

	roundTripper http.RoundTripper
	dial         func(context.Context, string, string) (net.Conn, error)
	mitm         *mitm.Config
	proxyURL     func(*http.Request) (*url.URL, error)
	conns        sync.WaitGroup
	connsMu      sync.Mutex // protects conns.Add/Wait from concurrent access
	closing      chan bool
	closeOnce    sync.Once

	reqmod RequestModifier
	resmod ResponseModifier
}

// NewProxy returns a new HTTP proxy.
func NewProxy() *Proxy {
	proxy := &Proxy{
		roundTripper: &http.Transport{
			// TODO(adamtanner): This forces the http.Transport to not upgrade requests
			// to HTTP/2 in Go 1.6+. Remove this once Martian can support HTTP/2.
			TLSNextProto:          make(map[string]func(string, *tls.Conn) http.RoundTripper),
			Proxy:                 http.ProxyFromEnvironment,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
		closing: make(chan bool),
		reqmod:  noop,
		resmod:  noop,
	}
	proxy.SetDialContext((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext)
	return proxy
}

// GetRoundTripper gets the http.RoundTripper of the proxy.
func (p *Proxy) GetRoundTripper() http.RoundTripper {
	return p.roundTripper
}

// SetRoundTripper sets the http.RoundTripper of the proxy.
func (p *Proxy) SetRoundTripper(rt http.RoundTripper) {
	p.roundTripper = rt

	if tr, ok := p.roundTripper.(*http.Transport); ok {
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
		tr.Proxy = p.proxyURL
		tr.DialContext = p.dial
	}
}

// SetUpstreamProxy sets the proxy that receives requests from this proxy.
func (p *Proxy) SetUpstreamProxy(proxyURL *url.URL) {
	p.SetUpstreamProxyFunc(http.ProxyURL(proxyURL))
}

// SetUpstreamProxyFunc sets proxy function as in http.Transport.Proxy.
func (p *Proxy) SetUpstreamProxyFunc(f func(*http.Request) (*url.URL, error)) {
	p.proxyURL = f

	if tr, ok := p.roundTripper.(*http.Transport); ok {
		tr.Proxy = f
	}
}

// SetMITM sets the config to use for MITMing of CONNECT requests.
func (p *Proxy) SetMITM(config *mitm.Config) {
	p.mitm = config
}

// SetDialContext sets the dial func used to establish a connection.
func (p *Proxy) SetDialContext(dial func(context.Context, string, string) (net.Conn, error)) {
	p.dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, e := dial(ctx, network, addr)
		nosigpipe.IgnoreSIGPIPE(c)
		return c, e
	}

	if tr, ok := p.roundTripper.(*http.Transport); ok {
		tr.DialContext = p.dial
	}
}

// Close sets the proxy to the closing state so it stops receiving new connections,
// finishes processing any inflight requests, and closes existing connections without
// reading anymore requests from them.
func (p *Proxy) Close() {
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

// SetRequestModifier sets the request modifier.
func (p *Proxy) SetRequestModifier(reqmod RequestModifier) {
	if reqmod == nil {
		reqmod = noop
	}

	p.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (p *Proxy) SetResponseModifier(resmod ResponseModifier) {
	if resmod == nil {
		resmod = noop
	}

	p.resmod = resmod
}

// Serve accepts connections from the listener and handles the requests.
func (p *Proxy) Serve(l net.Listener) error {
	defer l.Close()

	var delay time.Duration
	for {
		if p.Closing() {
			return nil
		}

		conn, err := l.Accept()
		nosigpipe.IgnoreSIGPIPE(conn)
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

	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	pc := &proxyConn{
		Proxy:   p,
		session: newSession(conn, brw),
		brw:     brw,
		conn:    conn,
	}

	const maxConsecutiveErrors = 5
	errorsN := 0
	for {
		if err := pc.handle(withSession(pc.session)); err != nil {
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

		if pc.session.Hijacked() {
			log.Debugf(context.TODO(), "closing connection: %v", conn.RemoteAddr())
			return
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

func (p *Proxy) requestContext(mctx *Context, req *http.Request) context.Context {
	ctx := req.Context()
	ctx = mctx.addToContext(ctx)

	if h := p.RequestIDHeader; h != "" {
		if id := req.Header.Get(h); id != "" {
			ctx = context.WithValue(ctx, log.TraceContextKey, id)
		}
	}

	return ctx
}

func (p *Proxy) shouldMITM(req *http.Request) bool {
	if p.mitm == nil {
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

func (p *Proxy) roundTrip(ctx *Context, req *http.Request) (*http.Response, error) {
	if ctx.SkippingRoundTrip() {
		log.Debugf(req.Context(), "skipping round trip")
		return proxyutil.NewResponse(200, http.NoBody, req), nil
	}

	return p.roundTripper.RoundTrip(req)
}

func (p *Proxy) warning(h http.Header, err error) {
	if p.WithoutWarning {
		return
	}
	proxyutil.Warning(h, err)
}

func (p *Proxy) errorResponse(req *http.Request, err error) *http.Response {
	if p.ErrorResponse != nil {
		return p.ErrorResponse(req, err)
	}
	return proxyutil.NewResponse(502, http.NoBody, req)
}

func (p *Proxy) connect(req *http.Request) (*http.Response, net.Conn, error) {
	var proxyURL *url.URL
	if p.proxyURL != nil {
		u, err := p.proxyURL(req)
		if err != nil {
			return nil, nil, err
		}
		proxyURL = u
	}

	if proxyURL == nil {
		log.Debugf(req.Context(), "CONNECT to host directly: %s", req.URL.Host)

		conn, err := p.dial(req.Context(), "tcp", req.URL.Host)
		if err != nil {
			return nil, nil, err
		}

		return proxyutil.NewResponse(200, http.NoBody, req), conn, nil
	}

	switch proxyURL.Scheme {
	case "http", "https":
		return p.connectHTTP(req, proxyURL)
	case "socks5":
		return p.connectSOCKS5(req, proxyURL)
	default:
		return nil, nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}
}

func (p *Proxy) connectHTTP(req *http.Request, proxyURL *url.URL) (res *http.Response, conn net.Conn, err error) {
	log.Debugf(req.Context(), "CONNECT with upstream HTTP proxy: %s", proxyURL.Host)

	var d *dialvia.HTTPProxyDialer
	if proxyURL.Scheme == "https" {
		d = dialvia.HTTPSProxy(p.dial, proxyURL, p.clientTLSConfig())
	} else {
		d = dialvia.HTTPProxy(p.dial, proxyURL)
	}
	d.ConnectRequestModifier = p.ConnectRequestModifier

	var ctx context.Context
	if p.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(req.Context(), p.ConnectTimeout)
		defer cancel()
	} else {
		ctx = req.Context()
	}
	res, conn, err = d.DialContextR(ctx, "tcp", req.URL.Host)

	if res != nil {
		if res.StatusCode/100 == 2 {
			res.Body.Close()
			return proxyutil.NewResponse(200, http.NoBody, req), conn, nil
		}

		// If the proxy returns a non-2xx response, return it to the client.
		// But first, replace the Request with the original request.
		res.Request = req
	}

	return res, conn, err
}

func (p *Proxy) clientTLSConfig() *tls.Config {
	if tr, ok := p.roundTripper.(*http.Transport); ok && tr.TLSClientConfig != nil {
		return tr.TLSClientConfig.Clone()
	}

	return &tls.Config{}
}

func (p *Proxy) connectSOCKS5(req *http.Request, proxyURL *url.URL) (*http.Response, net.Conn, error) {
	log.Debugf(req.Context(), "CONNECT with upstream SOCKS5 proxy: %s", proxyURL.Host)

	d := dialvia.SOCKS5Proxy(p.dial, proxyURL)

	conn, err := d.DialContext(req.Context(), "tcp", req.URL.Host)
	if err != nil {
		return nil, nil, err
	}

	return proxyutil.NewResponse(200, http.NoBody, req), conn, nil
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}
	return h.Get("Upgrade")
}
