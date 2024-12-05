package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/evanphx/wildcat"
	"github.com/panjf2000/gnet/v2"
)

type httpServer struct {
	gnet.BuiltinEventEngine

	addr      string
	multicore bool
	eng       gnet.Engine
}

type httpCodec struct {
	parser        *wildcat.HTTPParser
	contentLength int
	buf           []byte
}

var CRLF = []byte("\r\n\r\n")

func (hc *httpCodec) parse(data []byte) (int, error) {
	// Perform a legit HTTP request parsing.
	bodyOffset, err := hc.parser.Parse(data)
	if err != nil {
		return 0, err
	}

	// First check if the Content-Length header is present.
	contentLength := hc.getContentLength()
	if contentLength > -1 {
		return bodyOffset + contentLength, nil
	}

	// If the Content-Length header is not found,
	// we need to find the end of the body section.
	if idx := bytes.Index(data, CRLF); idx != -1 {
		return idx + 4, nil
	}

	return 0, errors.New("invalid http request")
}

var contentLengthKey = []byte("Content-Length")

func (hc *httpCodec) getContentLength() int {
	if hc.contentLength != -1 {
		return hc.contentLength
	}

	val := hc.parser.FindHeader(contentLengthKey)
	if val != nil {
		i, err := strconv.ParseInt(string(val), 10, 0)
		if err == nil {
			hc.contentLength = int(i)
		}
	}

	return hc.contentLength
}

func (hc *httpCodec) resetParser() {
	hc.contentLength = -1
}

func (hc *httpCodec) reset() {
	hc.resetParser()
	hc.buf = hc.buf[:0]
}

func (hc *httpCodec) appendResponse() {
	hc.buf = append(hc.buf, "HTTP/1.1 200 OK\r\nServer: gnet\r\nContent-Type: text/plain\r\nDate: "...)
	hc.buf = time.Now().AppendFormat(hc.buf, "Mon, 02 Jan 2006 15:04:05 GMT")
	hc.buf = append(hc.buf, "\r\nContent-Length: 12\r\n\r\nHello World!"...)
}

func (hs *httpServer) OnBoot(eng gnet.Engine) gnet.Action {
	hs.eng = eng
	log.Printf("echo server with multi-core=%t is listening on %s\n", hs.multicore, hs.addr)
	return gnet.None
}

func (hs *httpServer) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(&httpCodec{parser: wildcat.NewHTTPParser()})
	return nil, gnet.None
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	hc := c.Context().(*httpCodec)
	buf, _ := c.Next(-1)

pipeline:
	nextOffset, err := hc.parse(buf)
	if err != nil {
		goto response
	}
	hc.resetParser()
	hc.appendResponse()
	buf = buf[nextOffset:]
	if len(buf) > 0 {
		goto pipeline
	}
response:
	c.Write(hc.buf)
	hc.reset()
	return gnet.None
}

var (
	port = flag.Int("port", 8080, "server port")
)

func main() {
	multicore := false
	runtime.GOMAXPROCS(1)
	os.Setenv("GOMEMLIMIT", "1GiB")

	hs := &httpServer{
		addr:      fmt.Sprintf("tcp://127.0.0.1:%d", *port),
		multicore: multicore,
	}

	log.Println("server exits:", gnet.Run(hs, hs.addr, gnet.WithLockOSThread(true)))
}
