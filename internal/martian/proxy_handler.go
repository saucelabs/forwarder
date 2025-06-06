// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
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
	"errors"
	"fmt"
	"io"
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
	p.init()
	return proxyHandler{p}
}

func (p proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	outreq := req.Clone(withTraceID(p.BaseContext, newTraceID(req.Header.Get(p.RequestIDHeader))))
	if req.ContentLength == 0 {
		outreq.Body = http.NoBody
	}
	if outreq.Body != nil {
		defer outreq.Body.Close()
	}
	outreq.Close = false

	fixConnectReqContentLength(outreq)

	p.handleRequest(rw, outreq)
}

func (p proxyHandler) handleConnectRequest(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	terminateTLS := shouldTerminateTLS(req)
	req.Header.Del(terminateTLSHeader)

	if err := p.modifyRequest(req); err != nil {
		log.Debugf(ctx, "error modifying CONNECT request: %v", err)
		p.writeErrorResponse(rw, req, err)
		return
	}

	log.Debugf(ctx, "attempting to establish CONNECT tunnel: %s", req.URL.Host)
	res, crw, cerr := p.Connect(ctx, req, terminateTLS)
	if res != nil {
		defer res.Body.Close()
	}
	if crw != nil {
		defer crw.Close()
	}
	if cerr != nil {
		log.Errorf(ctx, "failed to CONNECT: %v", cerr)
		p.writeErrorResponse(rw, req, cerr)
		return
	}

	if err := p.modifyResponse(res); err != nil {
		log.Debugf(ctx, "error modifying CONNECT response: %v", err)
		p.writeErrorResponse(rw, req, err)
		return
	}

	if res.StatusCode != http.StatusOK {
		log.Infof(ctx, "CONNECT rejected with status code: %d", res.StatusCode)
		p.writeResponse(rw, res)
		return
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
		p.traceWroteResponse(res, errors.New("switching protocols response with non-writable body"))
		panic(http.ErrAbortHandler)
	}
	res.Body = panicBody

	if err := p.tunnel(resUpType, rw, req, res, uconn); err != nil {
		log.Errorf(ctx, "%s tunnel: %v", resUpType, err)
		panic(http.ErrAbortHandler)
	}
}

func (p proxyHandler) tunnel(name string, rw http.ResponseWriter, req *http.Request, res *http.Response, crw io.ReadWriteCloser) (ferr error) {
	rc := http.NewResponseController(rw)

	defer func() {
		p.traceWroteResponse(res, ferr)
	}()

	var cc []copier
	switch req.ProtoMajor {
	case 1:
		conn, brw, err := rc.Hijack()
		if err != nil {
			return err
		}
		defer conn.Close()

		pc := proxyConn{
			Proxy: p.Proxy,
			brw:   brw,
			conn:  conn,
		}
		if err := pc.writeResponse(res); err != nil {
			return err
		}

		if err := drainBuffer(crw, brw.Reader); err != nil {
			return fmt.Errorf("got error while draining buffer: %w", err)
		}

		cc = []copier{
			{"upstream " + name, crw, conn},
			{"downstream " + name, conn, crw},
		}
	case 2:
		copyHeader(rw.Header(), res.Header)
		rw.WriteHeader(res.StatusCode)

		if err := rc.Flush(); err != nil {
			return fmt.Errorf("got error while flushing response back to client: %w", err)
		}

		cc = []copier{
			{"upstream " + name, crw, req.Body},
			{"downstream " + name, makeH2Writer(rw, rc, req), crw},
		}
	default:
		return fmt.Errorf("unsupported protocol version: %d", req.ProtoMajor)
	}

	ctx := req.Context()

	log.Debugf(ctx, "established %s tunnel, proxying traffic", name)
	bicopy(ctx, cc...)
	log.Debugf(ctx, "closed %s tunnel duration=%s", name, ContextDuration(ctx))

	return nil
}

// handleRequest handles a request and writes the response to the given http.ResponseWriter.
// It returns an error if the request.
func (p proxyHandler) handleRequest(rw http.ResponseWriter, req *http.Request) {
	p.traceReadRequest(req, nil)

	ctx := req.Context()

	if req.Method == http.MethodConnect {
		p.handleConnectRequest(rw, req)
		return
	}

	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.RequestURI = ""

	p.fixRequestScheme(req)

	reqUpType := upgradeType(req.Header)
	if reqUpType != "" {
		log.Debugf(ctx, "upgrade request: %s", reqUpType)
	}

	if err := p.modifyRequest(req); err != nil {
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
			log.Errorf(ctx, "failed to round trip host=%s method=%s path=%s: %v",
				req.Host, req.Method, req.URL.Path, err)
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

	if err := p.modifyResponse(res); err != nil {
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

	// deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode == http.StatusSwitchingProtocols {
		p.handleUpgradeResponse(rw, req, res)
	} else {
		p.writeResponse(rw, res)
	}
}

type h2Writer struct {
	req   *http.Request
	w     io.Writer
	flush func() error
	close func() error
}

func makeH2Writer(rw http.ResponseWriter, rc *http.ResponseController, req *http.Request) h2Writer {
	return h2Writer{
		req:   req,
		w:     rw,
		flush: rc.Flush,
		close: req.Body.Close,
	}
}

func (w h2Writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)

	if n > 0 {
		if err := w.flush(); err != nil {
			log.Errorf(w.req.Context(), "got error while flushing response back to client: %v", err)
		}
	}

	return
}

func (w h2Writer) CloseWrite() error {
	// Send any DATA frames buffered in the transport.
	if err := w.flush(); err != nil {
		log.Errorf(w.req.Context(), "got error while flushing response back to client: %v", err)
	}
	// Close request body to signal the end of the request.
	// This results RST_STREAM frame with error code NO_ERROR to be sent to the other side.
	return w.close()
}

func (p proxyHandler) writeErrorResponse(rw http.ResponseWriter, req *http.Request, err error) {
	res := maybeConnectErrorResponse(err)
	if res == nil {
		res = p.errorResponse(req, err)
	}
	if err := p.modifyResponse(res); err != nil {
		log.Errorf(req.Context(), "error modifying error response: %v", err)
		if !p.WithoutWarning {
			proxyutil.Warning(res.Header, err)
		}
	}
	p.writeResponse(rw, res)
}

func (p proxyHandler) writeResponse(rw http.ResponseWriter, res *http.Response) {
	copyHeader(rw.Header(), res.Header)
	announcedTrailers := addTrailerHeader(rw, res.Trailer)
	rw.WriteHeader(res.StatusCode)

	// This flush is needed for http/1 server to flush the status code and headers.
	// It prevents the server from buffering the response and trying to calculate the response size.
	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	var err error
	switch {
	case isTextEventStream(res):
		w := newPatternFlushWriter(rw, http.NewResponseController(rw), sseFlushPattern)
		err = copyBody(w, res.Body)
	case shouldChunk(res):
		w := newPatternFlushWriter(rw, http.NewResponseController(rw), chunkFlushPattern)
		err = copyBody(w, res.Body)
	default:
		err = copyBody(rw, res.Body)
	}

	if err != nil {
		p.traceWroteResponse(res, err)
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

	if !skipTraceWroteResponse(res, err) {
		p.traceWroteResponse(res, err)
	}
}
