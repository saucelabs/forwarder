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
	"bytes"
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

	// MITMFilter specifies a function to determine whether a CONNECT request should be MITMed.
	MITMFilter func(*http.Request) bool

	// WithoutWarning disables the warning header added to requests and responses when modifier errors occur.
	WithoutWarning bool

	// ErrorResponse specifies a custom error HTTP response to send when a proxying error occurs.
	ErrorResponse func(req *http.Request, err error) *http.Response

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

	// CloseAfterReply closes the connection after the response has been sent.
	CloseAfterReply bool

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

		if tconn, ok := conn.(*net.TCPConn); ok {
			tconn.SetKeepAlive(true)
			tconn.SetKeepAlivePeriod(3 * time.Minute)
		}

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

	var (
		brw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		s   = newSession(conn, brw)
		ctx = withSession(s)
	)

	const maxConsecutiveErrors = 5
	errorsN := 0
	for {
		if err := p.handle(ctx, conn, brw); err != nil {
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

		if s.Hijacked() {
			log.Debugf(context.TODO(), "closing connection: %v", conn.RemoteAddr())
			return
		}
	}
}

func (p *Proxy) readHeaderTimeout() time.Duration {
	if p.ReadHeaderTimeout > 0 {
		return p.ReadHeaderTimeout
	}
	return p.ReadTimeout
}

func (p *Proxy) readRequest(ctx *Context, conn net.Conn, brw *bufio.ReadWriter) (*http.Request, error) {
	// Wait for the connection to become readable before trying to
	// read the next request. This prevents a ReadHeaderTimeout or
	// ReadTimeout from starting until the first bytes of the next request
	// have been received.
	if _, err := brw.Peek(1); err != nil {
		return nil, err
	}

	var (
		wholeReqDeadline time.Time // or zero if none
		hdrDeadline      time.Time // or zero if none
	)
	t0 := time.Now()
	if d := p.readHeaderTimeout(); d > 0 {
		hdrDeadline = t0.Add(d)
	}
	if d := p.ReadTimeout; d > 0 {
		wholeReqDeadline = t0.Add(d)
	}

	if deadlineErr := conn.SetReadDeadline(hdrDeadline); deadlineErr != nil {
		log.Errorf(context.TODO(), "can't set read header deadline: %v", deadlineErr)
	}

	req, err := http.ReadRequest(brw.Reader)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(p.requestContext(ctx, req))

	// Adjust the read deadline if necessary.
	if !hdrDeadline.Equal(wholeReqDeadline) {
		if deadlineErr := conn.SetReadDeadline(wholeReqDeadline); deadlineErr != nil {
			log.Errorf(context.TODO(), "can't set read deadline: %v", deadlineErr)
		}
	}

	return req, err
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

func (p *Proxy) handleMITM(ctx *Context, req *http.Request, session *Session, brw *bufio.ReadWriter, conn net.Conn) error {
	log.Debugf(req.Context(), "mitm: attempting MITM for connection %s", req.Host)

	res := proxyutil.NewResponse(200, nil, req)

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf(req.Context(), "mitm: error modifying CONNECT response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		log.Debugf(req.Context(), "mitm: connection hijacked by response modifier")
		return nil
	}

	if err := res.Write(brw); err != nil {
		log.Errorf(req.Context(), "mitm: got error while writing response back to client: %v", err)
	}
	if err := brw.Flush(); err != nil {
		log.Errorf(req.Context(), "mitm: got error while flushing response back to client: %v", err)
	}

	b, err := brw.Peek(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Debugf(req.Context(), "mitm: connection closed prematurely: %v", err)
		} else {
			log.Errorf(req.Context(), "mitm: failed to peek connection %s: %v", req.Host, err)
		}
		return errClose
	}

	// Drain all of the rest of the buffered data.
	buf := make([]byte, brw.Reader.Buffered())
	brw.Read(buf)

	// 22 is the TLS handshake.
	// https://tools.ietf.org/html/rfc5246#section-6.2.1
	if len(b) > 0 && b[0] == 22 {
		// Prepend the previously read data to be read again by http.ReadRequest.
		tlsconn := tls.Server(&peekedConn{
			conn,
			io.MultiReader(bytes.NewReader(buf), conn),
		}, p.mitm.TLSForHost(req.Host))

		if err := tlsconn.Handshake(); err != nil {
			p.mitm.HandshakeErrorCallback(req, err)
			if errors.Is(err, io.EOF) {
				log.Debugf(req.Context(), "mitm: connection closed prematurely: %v", err)
			} else {
				log.Errorf(req.Context(), "mitm: failed to handshake connection %s: %v", req.Host, err)
			}
			return errClose
		}

		cs := tlsconn.ConnectionState()
		log.Debugf(req.Context(), "mitm: negotiated %s for connection: %s", cs.NegotiatedProtocol, req.Host)

		if cs.NegotiatedProtocol == "h2" {
			return p.mitm.H2Config().Proxy(p.closing, tlsconn, req.URL)
		}

		brw.Writer.Reset(tlsconn)
		brw.Reader.Reset(tlsconn)
		return p.handle(ctx, tlsconn, brw)
	}

	// Prepend the previously read data to be read again by http.ReadRequest.
	brw.Reader.Reset(io.MultiReader(bytes.NewReader(buf), conn))
	return p.handle(ctx, conn, brw)
}

func (p *Proxy) handleConnectRequest(ctx *Context, req *http.Request, session *Session, brw *bufio.ReadWriter, conn net.Conn) error {
	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Errorf(req.Context(), "error modifying CONNECT request: %v", err)
		p.warning(req.Header, err)
	}
	if session.Hijacked() {
		log.Debugf(req.Context(), "connection hijacked by request modifier")
		return nil
	}

	if p.shouldMITM(req) {
		return p.handleMITM(ctx, req, session, brw, conn)
	}

	log.Debugf(req.Context(), "attempting to establish CONNECT tunnel: %s", req.URL.Host)
	var (
		res  *http.Response
		crw  io.ReadWriteCloser
		cerr error
	)
	if p.ConnectFunc != nil {
		res, crw, cerr = p.ConnectFunc(req)
	}
	if p.ConnectFunc == nil || errors.Is(cerr, ErrConnectFallback) {
		var cconn net.Conn
		res, cconn, cerr = p.connect(req)

		if cconn != nil {
			defer cconn.Close()
			crw = cconn

			if shouldTerminateTLS(req) {
				log.Debugf(req.Context(), "attempting to terminate TLS on CONNECT tunnel: %s", req.URL.Host)
				tconn := tls.Client(cconn, p.clientTLSConfig())
				if err := tconn.Handshake(); err == nil {
					crw = tconn
				} else {
					log.Errorf(req.Context(), "failed to terminate TLS on CONNECT tunnel: %v", err)
					cerr = err
				}
			}
		}
	}

	if cerr != nil {
		log.Errorf(req.Context(), "failed to CONNECT: %v", cerr)
		res = p.errorResponse(req, cerr)
		p.warning(res.Header, cerr)
	}
	defer res.Body.Close()

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf(req.Context(), "error modifying CONNECT response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		log.Debugf(req.Context(), "connection hijacked by response modifier")
		return nil
	}

	if res.StatusCode != http.StatusOK {
		if cerr == nil {
			log.Errorf(req.Context(), "CONNECT rejected with status code: %d", res.StatusCode)
		}
		if err := res.Write(brw); err != nil {
			log.Errorf(req.Context(), "got error while writing response back to client: %v", err)
		}
		err := brw.Flush()
		if err != nil {
			log.Errorf(req.Context(), "got error while flushing response back to client: %v", err)
		}
		return err
	}

	res.ContentLength = -1

	if err := p.tunnel("CONNECT", res, brw, conn, crw); err != nil {
		log.Errorf(req.Context(), "CONNECT tunnel: %w", err)
	}

	return errClose
}

func (p *Proxy) handleUpgradeResponse(res *http.Response, brw *bufio.ReadWriter, conn net.Conn) error {
	resUpType := upgradeType(res.Header)

	uconn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		log.Errorf(res.Request.Context(), "internal error: switching protocols response with non-writable body")
		return errClose
	}

	res.Body = nil

	if err := p.tunnel(resUpType, res, brw, conn, uconn); err != nil {
		log.Errorf(res.Request.Context(), "%s tunnel: %w", resUpType, err)
	}

	return errClose
}

func (p *Proxy) tunnel(name string, res *http.Response, brw *bufio.ReadWriter, conn net.Conn, crw io.ReadWriteCloser) error {
	if err := res.Write(brw); err != nil {
		return fmt.Errorf("got error while writing response back to client: %w", err)
	}
	if err := brw.Flush(); err != nil {
		return fmt.Errorf("got error while flushing response back to client: %w", err)
	}
	if err := drainBuffer(crw, brw.Reader); err != nil {
		return fmt.Errorf("got error while draining read buffer: %w", err)
	}

	ctx := res.Request.Context()
	donec := make(chan bool, 2)
	go copySync(ctx, "outbound "+name, crw, conn, donec)
	go copySync(ctx, "inbound "+name, conn, crw, donec)

	log.Debugf(ctx, "switched protocols, proxying %s traffic", name)
	<-donec
	<-donec
	log.Debugf(ctx, "closed %s tunnel", name)

	return nil
}

func drainBuffer(w io.Writer, r *bufio.Reader) error {
	if n := r.Buffered(); n > 0 {
		rbuf, err := r.Peek(n)
		if err != nil {
			return err
		}
		w.Write(rbuf)
	}
	return nil
}

var copyBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 32*1024)
		return &b
	},
}

func copySync(ctx context.Context, name string, w io.Writer, r io.Reader, donec chan<- bool) {
	bufp := copyBufPool.Get().(*[]byte) //nolint:forcetypeassert // It's *[]byte.
	buf := *bufp
	defer copyBufPool.Put(bufp)

	if _, err := io.CopyBuffer(w, r, buf); err != nil && !isClosedConnError(err) {
		log.Errorf(ctx, "failed to copy %s tunnel: %v", name, err)
	}
	if cw, ok := asCloseWriter(w); ok {
		cw.CloseWrite()
	} else if pw, ok := w.(*io.PipeWriter); ok {
		pw.Close()
	} else {
		log.Errorf(ctx, "cannot close write side of %s tunnel (%T)", name, w)
	}

	log.Debugf(ctx, "%s tunnel finished copying", name)
	donec <- true
}

func (p *Proxy) handle(ctx *Context, conn net.Conn, brw *bufio.ReadWriter) error {
	log.Debugf(context.TODO(), "waiting for request: %v", conn.RemoteAddr())

	session := ctx.Session()
	ctx = withSession(session)

	req, err := p.readRequest(ctx, conn, brw)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errClose
		}

		if isClosedConnError(err) {
			log.Debugf(context.TODO(), "connection closed prematurely: %v", err)
		} else {
			log.Errorf(context.TODO(), "failed to read request: %v", err)
		}
		return errClose
	}
	defer req.Body.Close()

	if p.Closing() {
		return errClose
	}

	if tconn, ok := conn.(*tls.Conn); ok {
		session.MarkSecure()

		cs := tconn.ConnectionState()
		req.TLS = &cs
	}

	req.RemoteAddr = conn.RemoteAddr().String()
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	if req.Method == http.MethodConnect {
		return p.handleConnectRequest(ctx, req, session, brw, conn)
	}

	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
		if session.IsSecure() {
			req.URL.Scheme = "https"
		}
	} else if req.URL.Scheme == "http" {
		if session.IsSecure() && !p.AllowHTTP {
			log.Infof(req.Context(), "forcing HTTPS inside secure session")
			req.URL.Scheme = "https"
		}
	}

	reqUpType := upgradeType(req.Header)
	if reqUpType != "" {
		log.Debugf(req.Context(), "upgrade request: %s", reqUpType)
	}
	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Errorf(req.Context(), "error modifying request: %v", err)
		p.warning(req.Header, err)
	}
	if session.Hijacked() {
		log.Debugf(req.Context(), "connection hijacked by request modifier")
		return nil
	}

	// after stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if reqUpType != "" {
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", reqUpType)
	}

	// perform the HTTP roundtrip
	res, err := p.roundTrip(ctx, req)
	if err != nil {
		log.Errorf(req.Context(), "failed to round trip: %v", err)
		res = p.errorResponse(req, err)
		p.warning(res.Header, err)
	}
	defer res.Body.Close()

	// set request to original request manually, res.Request may be changed in transport.
	// see https://github.com/google/martian/issues/298
	res.Request = req

	resUpType := upgradeType(res.Header)
	if resUpType != "" {
		log.Debugf(req.Context(), "upgrade response: %s", resUpType)
	}
	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf(req.Context(), "error modifying response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		log.Debugf(req.Context(), "connection hijacked by response modifier")
		return nil
	}

	// after stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if resUpType != "" {
		res.Header.Set("Connection", "Upgrade")
		res.Header.Set("Upgrade", resUpType)
	}

	var closing error
	if !req.ProtoAtLeast(1, 1) || req.Close || res.Close || p.Closing() {
		log.Debugf(req.Context(), "received close request: %v", req.RemoteAddr)
		res.Close = true
		closing = errClose
	}

	// deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode == http.StatusSwitchingProtocols {
		return p.handleUpgradeResponse(res, brw, conn)
	}

	if p.WriteTimeout > 0 {
		if deadlineErr := conn.SetWriteDeadline(time.Now().Add(p.WriteTimeout)); deadlineErr != nil {
			log.Errorf(req.Context(), "can't set write deadline: %v", deadlineErr)
		}
	}

	if req.Method == "HEAD" && res.Body == http.NoBody {
		// The http package is misbehaving when writing a HEAD response.
		// See https://github.com/golang/go/issues/62015 for details.
		// This works around the issue by writing the response manually.
		err = writeHeadResponse(brw.Writer, res)
	} else {
		// Add support for Server Sent Events - relay HTTP chunks and flush after each chunk.
		// This is safe for events that are smaller than the buffer io.Copy uses (32KB).
		// If the event is larger than the buffer, the event will be split into multiple chunks.
		if shouldFlush(res) {
			err = res.Write(flushAfterChunkWriter{brw.Writer})
		} else {
			err = res.Write(brw)
		}
	}
	if err != nil {
		log.Errorf(req.Context(), "got error while writing response back to client: %v", err)
		if errors.Is(err, io.ErrUnexpectedEOF) {
			closing = errClose
		}
	}
	err = brw.Flush()
	if err != nil {
		log.Errorf(req.Context(), "got error while flushing response back to client: %v", err)
	}

	if err := req.Body.Close(); err != nil {
		log.Errorf(req.Context(), "failed to close request body: %v", err)
		closing = errClose
	}
	if err := res.Body.Close(); err != nil {
		log.Errorf(req.Context(), "failed to close response body: %v", err)
		closing = errClose
	}

	if p.CloseAfterReply {
		closing = errClose
	}
	return closing
}

// A peekedConn subverts the net.Conn.Read implementation, primarily so that
// sniffed bytes can be transparently prepended.
type peekedConn struct {
	net.Conn
	r io.Reader
}

// Read allows control over the embedded net.Conn's read data. By using an
// io.MultiReader one can read from a conn, and then replace what they read, to
// be read again.
func (c *peekedConn) Read(buf []byte) (int, error) { return c.r.Read(buf) }

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
	res, conn, err = d.DialContextR(req.Context(), "tcp", req.URL.Host)

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
