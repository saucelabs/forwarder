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
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
)

type proxyConn struct {
	*Proxy
	session *Session
	brw     *bufio.ReadWriter
	conn    net.Conn
}

func (p *proxyConn) readRequest(ctx *Context) (*http.Request, error) {
	var idleDeadline time.Time // or zero if none
	if d := p.idleTimeout(); d > 0 {
		idleDeadline = time.Now().Add(d)
	}
	if deadlineErr := p.conn.SetReadDeadline(idleDeadline); deadlineErr != nil {
		log.Errorf(context.TODO(), "can't set idle deadline: %v", deadlineErr)
	}

	// Wait for the connection to become readable before trying to
	// read the next request. This prevents a ReadHeaderTimeout or
	// ReadTimeout from starting until the first bytes of the next request
	// have been received.
	if _, err := p.brw.Peek(1); err != nil {
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

	if deadlineErr := p.conn.SetReadDeadline(hdrDeadline); deadlineErr != nil {
		log.Errorf(context.TODO(), "can't set read header deadline: %v", deadlineErr)
	}

	req, err := http.ReadRequest(p.brw.Reader)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(p.requestContext(ctx, req))

	// Adjust the read deadline if necessary.
	if !hdrDeadline.Equal(wholeReqDeadline) {
		if deadlineErr := p.conn.SetReadDeadline(wholeReqDeadline); deadlineErr != nil {
			log.Errorf(context.TODO(), "can't set read deadline: %v", deadlineErr)
		}
	}

	return req, err
}

func (p *proxyConn) handleMITM(ctx *Context, req *http.Request) error {
	log.Debugf(req.Context(), "mitm: attempting MITM for connection %s", req.Host)

	res := proxyutil.NewResponse(200, nil, req)

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf(req.Context(), "mitm: error modifying CONNECT response: %v", err)
		p.warning(res.Header, err)
	}

	if err := p.writeResponse(res); err != nil {
		return err
	}

	b, err := p.brw.Peek(1)
	if err != nil {
		if isClosedConnError(err) {
			log.Debugf(req.Context(), "mitm: connection closed prematurely: %v", err)
		} else {
			log.Errorf(req.Context(), "mitm: failed to peek connection %s: %v", req.Host, err)
		}
		return errClose
	}

	// Drain all of the rest of the buffered data.
	buf := make([]byte, p.brw.Reader.Buffered())
	p.brw.Read(buf)

	// 22 is the TLS handshake.
	// https://tools.ietf.org/html/rfc5246#section-6.2.1
	if len(b) > 0 && b[0] == 22 {
		// Prepend the previously read data to be read again by http.ReadRequest.
		tlsconn := tls.Server(&peekedConn{
			p.conn,
			io.MultiReader(bytes.NewReader(buf), p.conn),
		}, p.mitm.TLSForHost(req.Host))

		var hctx context.Context
		if p.MITMTLSHandshakeTimeout > 0 {
			var hcancel context.CancelFunc
			hctx, hcancel = context.WithTimeout(req.Context(), p.MITMTLSHandshakeTimeout)
			defer hcancel()
		} else {
			hctx = req.Context()
		}
		if err = tlsconn.HandshakeContext(hctx); err != nil {
			p.mitm.HandshakeErrorCallback(req, err)
			if isClosedConnError(err) {
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

		p.brw.Writer.Reset(tlsconn)
		p.brw.Reader.Reset(tlsconn)
		p.conn = tlsconn
		return p.handle(ctx)
	}

	// Prepend the previously read data to be read again by http.ReadRequest.
	p.brw.Reader.Reset(io.MultiReader(bytes.NewReader(buf), p.conn))
	return p.handle(ctx)
}

func (p *proxyConn) handleConnectRequest(ctx *Context, req *http.Request) error {
	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Errorf(req.Context(), "error modifying CONNECT request: %v", err)
		p.warning(req.Header, err)
	}

	if p.shouldMITM(req) {
		return p.handleMITM(ctx, req)
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

	if res.StatusCode != http.StatusOK {
		if cerr == nil {
			log.Errorf(req.Context(), "CONNECT rejected with status code: %d", res.StatusCode)
		}
		return p.writeResponse(res)
	}

	res.ContentLength = -1

	if err := p.tunnel("CONNECT", res, crw); err != nil {
		log.Errorf(req.Context(), "CONNECT tunnel: %w", err)
	}

	return errClose
}

func (p *proxyConn) handleUpgradeResponse(res *http.Response) error {
	resUpType := upgradeType(res.Header)

	uconn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		log.Errorf(res.Request.Context(), "internal error: switching protocols response with non-writable body")
		return errClose
	}

	res.Body = nil

	if err := p.tunnel(resUpType, res, uconn); err != nil {
		log.Errorf(res.Request.Context(), "%s tunnel: %w", resUpType, err)
	}

	return errClose
}

func (p *proxyConn) tunnel(name string, res *http.Response, crw io.ReadWriteCloser) error {
	if err := res.Write(p.brw); err != nil {
		return fmt.Errorf("got error while writing response back to client: %w", err)
	}
	if err := p.brw.Flush(); err != nil {
		return fmt.Errorf("got error while flushing response back to client: %w", err)
	}
	if err := drainBuffer(crw, p.brw.Reader); err != nil {
		return fmt.Errorf("got error while draining read buffer: %w", err)
	}

	ctx := res.Request.Context()
	donec := make(chan bool, 2)
	go copySync(ctx, "outbound "+name, crw, p.conn, donec)
	go copySync(ctx, "inbound "+name, p.conn, crw, donec)

	log.Debugf(ctx, "switched protocols, proxying %s traffic", name)
	<-donec
	<-donec
	log.Debugf(ctx, "closed %s tunnel", name)

	return nil
}

func (p *proxyConn) handle(ctx *Context) error {
	log.Debugf(context.TODO(), "waiting for request: %v", p.conn.RemoteAddr())

	session := ctx.Session()
	ctx = withSession(session)

	req, err := p.readRequest(ctx)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errClose
		}

		if isClosedConnError(err) {
			log.Debugf(context.TODO(), "connection closed prematurely while reading request: %v", err)
		} else {
			log.Errorf(context.TODO(), "got error while reading request: %v", err)
		}
		return errClose
	}
	defer req.Body.Close()

	if p.Closing() {
		return errClose
	}

	if tconn, ok := p.conn.(*tls.Conn); ok {
		session.MarkSecure()

		cs := tconn.ConnectionState()
		req.TLS = &cs
	}

	req.RemoteAddr = p.conn.RemoteAddr().String()
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	if req.Method == http.MethodConnect {
		return p.handleConnectRequest(ctx, req)
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

	// after stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if reqUpType != "" {
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", reqUpType)
	}

	// perform the HTTP roundtrip
	res, err := p.roundTrip(ctx, req)
	if err != nil {
		if isClosedConnError(err) {
			log.Debugf(req.Context(), "connection closed prematurely: %v", err)
		} else {
			log.Errorf(req.Context(), "failed to round trip: %v", err)
		}

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

	// after stripping all the hop-by-hop connection headers above, add back any
	// necessary for protocol upgrades, such as for websockets.
	if resUpType != "" {
		res.Header.Set("Connection", "Upgrade")
		res.Header.Set("Upgrade", resUpType)
	}

	// deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode == http.StatusSwitchingProtocols {
		return p.handleUpgradeResponse(res)
	}

	return p.writeResponse(res)
}

func (p *proxyConn) writeResponse(res *http.Response) error {
	req := res.Request

	if p.WriteTimeout > 0 {
		if deadlineErr := p.conn.SetWriteDeadline(time.Now().Add(p.WriteTimeout)); deadlineErr != nil {
			log.Errorf(req.Context(), "can't set write deadline: %v", deadlineErr)
		}
		defer p.conn.SetWriteDeadline(time.Time{})
	}

	if !req.ProtoAtLeast(1, 1) || req.Close || res.Close || p.Closing() {
		log.Debugf(req.Context(), "received close request: %v", req.RemoteAddr)
		res.Close = true
	}

	var err error
	if req.Method == "HEAD" && res.Body == http.NoBody {
		// The http package is misbehaving when writing a HEAD response.
		// See https://github.com/golang/go/issues/62015 for details.
		// This works around the issue by writing the response manually.
		err = writeHeadResponse(p.brw.Writer, res)
	} else {
		// Add support for Server Sent Events - relay HTTP chunks and flush after each chunk.
		// This is safe for events that are smaller than the buffer io.Copy uses (32KB).
		// If the event is larger than the buffer, the event will be split into multiple chunks.
		if shouldFlush(res) {
			err = res.Write(flushAfterChunkWriter{p.brw.Writer})
		} else {
			err = res.Write(p.brw)
		}
	}
	if err != nil {
		p.brw.Flush() // flush any remaining data
	} else {
		err = p.brw.Flush()
	}

	if err != nil {
		if isClosedConnError(err) {
			log.Debugf(req.Context(), "connection closed prematurely while writing response: %v", err)
		} else {
			log.Errorf(req.Context(), "got error while writing response: %v", err)
		}
		return errClose
	}

	if res.Close {
		return errClose
	}

	return nil
}
