// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
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
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
	forwarderlog "github.com/saucelabs/forwarder/log"
)

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func addTrailerHeader(rw http.ResponseWriter, tr http.Header) int {
	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	announcedTrailers := len(tr)
	if announcedTrailers == 0 {
		return 0
	}

	trailerKeys := make([]string, 0, announcedTrailers)
	for k := range tr {
		trailerKeys = append(trailerKeys, k)
	}
	rw.Header().Add("Trailer", strings.Join(trailerKeys, ", "))

	return announcedTrailers
}

func copyBody(w io.Writer, body io.ReadCloser) error {
	bufp := copyBufPool.Get().(*[]byte) //nolint:forcetypeassert // It's *[]byte.
	buf := *bufp
	defer copyBufPool.Put(bufp)

	_, err := io.CopyBuffer(w, body, buf)
	return err
}

// proxyHandler wraps Proxy and implements http.Handler.
//
// Known limitations:
//   - MITM is not supported
//   - HTTP status code 100 is not supported, see [issue 2184]
//
// [issue 2184]: https://github.com/golang/go/issues/2184
type proxyHandler struct {
	*Proxy
}

// Handler returns proxy as http.Handler, see [proxyHandler] for details.
func (p *Proxy) Handler() http.Handler {
	return proxyHandler{p}
}

func (p proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	session := newSessionWithResponseWriter(rw)
	if req.TLS != nil {
		session.MarkSecure()
	}
	ctx := withSession(session)

	outreq := req.Clone(ctx.addToContext(req.Context()))
	if req.ContentLength == 0 {
		outreq.Body = http.NoBody
	}
	if outreq.Body != nil {
		defer outreq.Body.Close()
	}
	outreq.Close = false

	p.handleRequest(ctx, rw, outreq)
}

func (p proxyHandler) handleConnectRequest(ctx *Context, rw http.ResponseWriter, req *http.Request) {
	session := ctx.Session()
	tracedLog := log.WithTraceID(ctx.ID())

	if err := p.reqmod.ModifyRequest(req); err != nil {
		tracedLog.Errorf("martian: error modifying CONNECT request: %v", err)
		p.warning(req.Header, err)
	}
	if session.Hijacked() {
		tracedLog.Debugf("martian: connection hijacked by request modifier")
		return
	}

	tracedLog.Debugf("martian: attempting to establish CONNECT tunnel: %s", req.URL.Host)
	var (
		res  *http.Response
		cr   io.Reader
		cw   io.WriteCloser
		cerr error
	)
	if p.ConnectPassthrough {
		pr, pw := io.Pipe()
		req.Body = pr
		defer req.Body.Close()

		// perform the HTTP roundtrip
		res, cerr = p.roundTrip(ctx, req)
		if res != nil {
			cr = res.Body
			cw = pw

			if res.StatusCode/100 == 2 {
				res = proxyutil.NewResponse(200, http.NoBody, req)
			}
		}
	} else {
		var cconn net.Conn
		res, cconn, cerr = p.connect(ctx, req)

		if cconn != nil {
			defer cconn.Close()
			cr = cconn
			cw = cconn
		}
	}

	if cerr != nil {
		tracedLog.Errorf("martian: failed to CONNECT: %v", cerr)
		res = p.errorResponse(req, cerr)
		p.warning(res.Header, cerr)
	}
	defer res.Body.Close()

	if err := p.resmod.ModifyResponse(res); err != nil {
		tracedLog.Errorf("martian: error modifying CONNECT response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		tracedLog.Debugf("martian: connection hijacked by response modifier")
		return
	}

	if res.StatusCode != http.StatusOK {
		if cerr == nil {
			tracedLog.Errorf("martian: CONNECT rejected with status code: %d", res.StatusCode)
		}
		writeResponse(rw, res, tracedLog)
		return
	}

	if req.ProtoMajor == 1 {
		res.ContentLength = -1
	}

	if err := p.tunnel(ctx, "CONNECT", rw, req, res, cw, cr); err != nil {
		tracedLog.Errorf("martian: CONNECT tunnel: %v", err)
		panic(http.ErrAbortHandler)
	}
}

func (p proxyHandler) handleUpgradeResponse(ctx *Context, rw http.ResponseWriter, req *http.Request, res *http.Response) {
	resUpType := upgradeType(res.Header)

	uconn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		log.WithTraceID(ctx.ID()).Errorf("martian: %s tunnel: internal error: switching protocols response with non-ReadWriteCloser body", resUpType)
		panic(http.ErrAbortHandler)
	}

	res.Body = nil

	if err := p.tunnel(ctx, resUpType, rw, req, res, uconn, uconn); err != nil {
		log.WithTraceID(ctx.ID()).Errorf("martian: %s tunnel: %w", resUpType, err)
		panic(http.ErrAbortHandler)
	}
}

func (p proxyHandler) tunnel(ctx *Context, name string, rw http.ResponseWriter, req *http.Request, res *http.Response, cw io.WriteCloser, cr io.Reader) error {
	var (
		rc        = http.NewResponseController(rw)
		donec     = make(chan bool, 2)
		tracedLog = log.WithTraceID(ctx.ID())
	)
	switch req.ProtoMajor {
	case 1:
		conn, brw, err := rc.Hijack()
		if err != nil {
			return err
		}
		defer conn.Close()

		if err := res.Write(brw); err != nil {
			return fmt.Errorf("got error while writing response back to client: %w", err)
		}
		if err := brw.Flush(); err != nil {
			return fmt.Errorf("got error while flushing response back to client: %w", err)
		}
		if err := drainBuffer(cw, brw.Reader); err != nil {
			return fmt.Errorf("got error while draining buffer: %w", err)
		}

		go copySync(tracedLog, "outbound "+name, cw, conn, donec)
		go copySync(tracedLog, "inbound "+name, conn, cr, donec)
	case 2:
		copyHeader(rw.Header(), res.Header)
		rw.WriteHeader(res.StatusCode)

		if err := rc.Flush(); err != nil {
			return fmt.Errorf("got error while flushing response back to client: %w", err)
		}

		go copySync(tracedLog, "outbound "+name, cw, req.Body, donec)
		go copySync(tracedLog, "inbound "+name, tracedWriteFlusher{rw, rc, tracedLog}, cr, donec)
	default:
		return fmt.Errorf("unsupported protocol version: %d", req.ProtoMajor)
	}

	tracedLog.Debugf("martian: established %s tunnel, proxying traffic", name)
	<-donec
	<-donec
	tracedLog.Debugf("martian: closed %s tunnel", name)

	return nil
}

// handleRequest handles a request and writes the response to the given http.ResponseWriter.
// It returns an error if the request.
func (p proxyHandler) handleRequest(ctx *Context, rw http.ResponseWriter, req *http.Request) {
	session := ctx.Session()

	if req.Method == http.MethodConnect {
		p.handleConnectRequest(ctx, rw, req)
		return
	}

	tracedLog := log.WithTraceID(ctx.ID())

	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.RequestURI = ""

	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
		if session.IsSecure() {
			req.URL.Scheme = "https"
		}
	} else if req.URL.Scheme == "http" {
		if session.IsSecure() && !p.AllowHTTP {
			tracedLog.Infof("martian: forcing HTTPS inside secure session")
			req.URL.Scheme = "https"
		}
	}

	reqUpType := upgradeType(req.Header)
	if reqUpType != "" {
		tracedLog.Debugf("martian: upgrade request: %s", reqUpType)
	}
	if err := p.reqmod.ModifyRequest(req); err != nil {
		tracedLog.Errorf("martian: error modifying request: %v", err)
		p.warning(req.Header, err)
	}
	if session.Hijacked() {
		tracedLog.Debugf("martian: connection hijacked by request modifier")
		return
	}

	// After stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if reqUpType != "" {
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", reqUpType)
	}

	// perform the HTTP roundtrip
	res, err := p.roundTrip(ctx, req)
	if err != nil {
		tracedLog.Errorf("martian: failed to round trip: %v", err)
		res = p.errorResponse(req, err)
		p.warning(res.Header, err)
	}
	defer res.Body.Close()

	// set request to original request manually, res.Request may be changed in transport.
	// see https://github.com/google/martian/issues/298
	res.Request = req

	resUpType := upgradeType(res.Header)
	if resUpType != "" {
		tracedLog.Debugf("martian: upgrade response: %s", resUpType)
	}
	if err := p.resmod.ModifyResponse(res); err != nil {
		tracedLog.Errorf("martian: error modifying response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		tracedLog.Debugf("martian: connection hijacked by response modifier")
		return
	}

	// after stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if resUpType != "" {
		res.Header.Set("Connection", "Upgrade")
		res.Header.Set("Upgrade", resUpType)
	}

	if !req.ProtoAtLeast(1, 1) || req.Close || res.Close || p.Closing() {
		tracedLog.Debugf("martian: received close request: %v", req.RemoteAddr)
		res.Close = true
	}
	if p.CloseAfterReply {
		res.Close = true
	}

	// deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode == http.StatusSwitchingProtocols {
		p.handleUpgradeResponse(ctx, rw, req, res)
	} else {
		writeResponse(rw, res, tracedLog)
	}
}

func newTracedWriteFlusher(rw http.ResponseWriter, tracedLog forwarderlog.Logger) tracedWriteFlusher {
	return tracedWriteFlusher{
		rw:  rw,
		rc:  http.NewResponseController(rw),
		log: tracedLog,
	}
}

type tracedWriteFlusher struct {
	rw  io.Writer
	rc  *http.ResponseController
	log forwarderlog.Logger
}

func (w tracedWriteFlusher) Write(p []byte) (n int, err error) {
	n, err = w.rw.Write(p)

	if n > 0 {
		if err := w.rc.Flush(); err != nil {
			w.log.Errorf("martian: got error while flushing response back to client: %v", err)
		}
	}

	return
}

func (w tracedWriteFlusher) CloseWrite() error {
	// This is a nop implementation of closeWriter.
	// It avoids printing the error log "cannot close write side of inbound CONNECT tunnel".
	return nil
}

func writeResponse(rw http.ResponseWriter, res *http.Response, tracedLog forwarderlog.Logger) {
	copyHeader(rw.Header(), res.Header)
	if res.Close {
		res.Header.Set("Connection", "close")
	}
	announcedTrailers := addTrailerHeader(rw, res.Trailer)
	rw.WriteHeader(res.StatusCode)

	// This flush is needed for http/1 server to flush the status code and headers.
	// It prevents the server from buffering the response and trying to calculate the response size.
	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	var err error
	if shouldFlush(res) {
		err = copyBody(newTracedWriteFlusher(rw, tracedLog), res.Body)
	} else {
		err = copyBody(rw, res.Body)
	}
	if err != nil {
		tracedLog.Errorf("martian: got error while writing response back to client: %v", err)
		panic(http.ErrAbortHandler)
	}

	res.Body.Close() // close now, instead of defer, to populate res.Trailer
	if len(res.Trailer) == announcedTrailers {
		copyHeader(rw.Header(), res.Trailer)
	} else {
		h := rw.Header()
		for k, vv := range res.Trailer {
			for _, v := range vv {
				h.Add(http.TrailerPrefix+k, v)
			}
		}
	}
}
