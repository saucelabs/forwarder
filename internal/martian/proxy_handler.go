// Copyright 2023 Sauce Labs Inc., all rights reserved.
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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
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
	if body == http.NoBody {
		return nil
	}

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
	outreq := req.Clone(WithTraceID(p.BaseContex, newTraceID()))
	if req.ContentLength == 0 {
		outreq.Body = http.NoBody
	}
	if outreq.Body != nil {
		defer outreq.Body.Close()
	}
	outreq.Close = false

	p.handleRequest(rw, outreq)
}

func (p proxyHandler) handleConnectRequest(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Debugf(ctx, "error modifying CONNECT request: %v", err)
		p.writeErrorResponse(rw, req, err)
		return
	}

	log.Debugf(ctx, "attempting to establish CONNECT tunnel: %s", req.URL.Host)
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
				log.Debugf(ctx, "attempting to terminate TLS on CONNECT tunnel: %s", req.URL.Host)
				tconn := tls.Client(cconn, p.clientTLSConfig())
				if err := tconn.Handshake(); err == nil {
					crw = tconn
				} else {
					log.Errorf(ctx, "failed to terminate TLS on CONNECT tunnel: %v", err)
					cerr = err
				}
			}
		}
	}

	if res != nil {
		defer res.Body.Close()
	}
	if cerr != nil {
		log.Errorf(ctx, "failed to CONNECT: %v", cerr)
		p.writeErrorResponse(rw, req, cerr)
		return
	}

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Debugf(ctx, "error modifying CONNECT response: %v", err)
		p.writeErrorResponse(rw, req, err)
		return
	}

	if res.StatusCode != http.StatusOK {
		log.Errorf(ctx, "CONNECT rejected with status code: %d", res.StatusCode)
		p.writeResponse(rw, res)
		return
	}

	if req.ProtoMajor == 1 {
		res.ContentLength = -1
	}

	if err := p.tunnel("CONNECT", rw, req, res, crw); err != nil {
		log.Errorf(ctx, "CONNECT tunnel: %v", err)
		panic(http.ErrAbortHandler)
	}
}

func (p proxyHandler) handleUpgradeResponse(rw http.ResponseWriter, req *http.Request, res *http.Response) {
	ctx := req.Context()
	resUpType := upgradeType(res.Header)

	uconn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		log.Errorf(ctx, "%s tunnel: internal error: switching protocols response with non-ReadWriteCloser body", resUpType)
		panic(http.ErrAbortHandler)
	}

	res.Body = nil

	if err := p.tunnel(resUpType, rw, req, res, uconn); err != nil {
		log.Errorf(ctx, "%s tunnel: %w", resUpType, err)
		panic(http.ErrAbortHandler)
	}
}

func (p proxyHandler) tunnel(name string, rw http.ResponseWriter, req *http.Request, res *http.Response, crw io.ReadWriteCloser) error {
	ctx := req.Context()

	var (
		rc    = http.NewResponseController(rw)
		donec = make(chan bool, 2)
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
		if err := drainBuffer(crw, brw.Reader); err != nil {
			return fmt.Errorf("got error while draining buffer: %w", err)
		}

		go copySync(ctx, "outbound "+name, crw, conn, donec)
		go copySync(ctx, "inbound "+name, conn, crw, donec)
	case 2:
		copyHeader(rw.Header(), res.Header)
		rw.WriteHeader(res.StatusCode)

		if err := rc.Flush(); err != nil {
			return fmt.Errorf("got error while flushing response back to client: %w", err)
		}

		go copySync(ctx, "outbound "+name, crw, req.Body, donec)
		go copySync(ctx, "inbound "+name, writeFlusher{rw, rc}, crw, donec)
	default:
		return fmt.Errorf("unsupported protocol version: %d", req.ProtoMajor)
	}

	log.Debugf(ctx, "established %s tunnel, proxying traffic", name)
	<-donec
	<-donec
	log.Debugf(ctx, "closed %s tunnel", name)

	return nil
}

// handleRequest handles a request and writes the response to the given http.ResponseWriter.
// It returns an error if the request.
func (p proxyHandler) handleRequest(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == http.MethodConnect {
		p.handleConnectRequest(rw, req)
		return
	}

	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.RequestURI = ""

	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
		if req.TLS != nil {
			req.URL.Scheme = "https"
		}
	} else if req.URL.Scheme == "http" {
		if req.TLS != nil && !p.AllowHTTP {
			log.Infof(ctx, "forcing HTTPS inside secure session")
			req.URL.Scheme = "https"
		}
	}

	reqUpType := upgradeType(req.Header)
	if reqUpType != "" {
		log.Debugf(ctx, "upgrade request: %s", reqUpType)
	}

	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Debugf(ctx, "error modifying request: %v", err)
		p.writeErrorResponse(rw, req, err)
		return
	}

	// After stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if reqUpType != "" {
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", reqUpType)
	}

	// perform the HTTP roundtrip
	res, err := p.roundTrip(req)
	if err != nil {
		if isClosedConnError(err) {
			log.Debugf(ctx, "connection closed prematurely: %v", err)
		} else {
			log.Errorf(ctx, "failed to round trip: %v", err)
		}
		p.writeErrorResponse(rw, req, err)
		return
	}

	defer res.Body.Close()

	// set request to original request manually, res.Request may be changed in transport.
	// see https://github.com/google/martian/issues/298
	res.Request = req

	resUpType := upgradeType(res.Header)
	if resUpType != "" {
		log.Debugf(ctx, "upgrade response: %s", resUpType)
	}

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Debugf(ctx, "error modifying response: %v", err)
		p.writeErrorResponse(rw, req, err)
		return
	}

	// after stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if resUpType != "" {
		res.Header.Set("Connection", "Upgrade")
		res.Header.Set("Upgrade", resUpType)
	}

	if !req.ProtoAtLeast(1, 1) || req.Close || res.Close || p.Closing() {
		log.Debugf(ctx, "received close request: %v", req.RemoteAddr)
		res.Close = true
	}

	// deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode == http.StatusSwitchingProtocols {
		p.handleUpgradeResponse(rw, req, res)
	} else {
		p.writeResponse(rw, res)
	}
}

func newWriteFlusher(rw http.ResponseWriter) writeFlusher {
	return writeFlusher{
		rw: rw,
		rc: http.NewResponseController(rw),
	}
}

type writeFlusher struct {
	rw io.Writer
	rc *http.ResponseController
}

func (w writeFlusher) Write(p []byte) (n int, err error) {
	n, err = w.rw.Write(p)

	if n > 0 {
		if err := w.rc.Flush(); err != nil {
			log.Errorf(context.TODO(), "got error while flushing response back to client: %v", err)
		}
	}

	return
}

func (w writeFlusher) CloseWrite() error {
	// This is a nop implementation of closeWriter.
	// It avoids printing the error log "cannot close write side of inbound CONNECT tunnel".
	return nil
}

func (p proxyHandler) writeErrorResponse(rw http.ResponseWriter, req *http.Request, err error) {
	res := p.errorResponse(req, err)
	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf(req.Context(), "error modifying error response: %v", err)
		if !p.WithoutWarning {
			proxyutil.Warning(res.Header, err)
		}
	}
	p.writeResponse(rw, res)
}

func (p proxyHandler) writeResponse(rw http.ResponseWriter, res *http.Response) {
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
		err = copyBody(newWriteFlusher(rw), res.Body)
	} else {
		err = copyBody(rw, res.Body)
	}

	if err != nil {
		if isClosedConnError(err) {
			log.Debugf(res.Request.Context(), "connection closed prematurely while writing response: %v", err)
		} else {
			log.Errorf(res.Request.Context(), "got error while writing response: %v", err)
		}
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
