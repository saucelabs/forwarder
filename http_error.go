package forwarder

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/elazarl/goproxy"
)

type denyError struct {
	error
}

// ErrorHeader is the header that is set on error responses with the error message.
const ErrorHeader = "X-Forwarder-Error"

var ErrProxyLocalhost = denyError{errors.New("localhost proxying is disabled")}

func httpErrorHandler(w io.WriteCloser, ctx *goproxy.ProxyCtx, err error) {
	resp := errorResponse(ctx, err)
	defer resp.Body.Close()

	if err := resp.Write(w); err != nil {
		ctx.Warnf("Failed to write HTTP error response: %s", err)
	}
	if err := w.Close(); err != nil {
		ctx.Warnf("Failed to close proxy client connection: %s", err)
	}
}

//nolint:errorlint // net.Error is an interface
func errorResponse(ctx *goproxy.ProxyCtx, err error) *http.Response {
	var (
		msg  string
		code int
	)

	if e, ok := err.(net.Error); ok { //nolint // switch is not an option for type asserts
		// net.Dial timeout
		if e.Timeout() {
			code = http.StatusGatewayTimeout
			msg = "Timed out connecting to remote host"
		} else {
			code = http.StatusBadGateway
			msg = "Failed to connect to remote host"
		}
	} else if _, ok := err.(denyError); ok {
		code = http.StatusBadGateway
		msg = fmt.Sprintf("Proxying is denied to host %q", ctx.Req.Host)
	} else {
		code = http.StatusInternalServerError
		msg = "An unexpected error occurred"
		ctx.Warnf("Unexpected error: %s", err)
	}

	resp := goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, code, msg+"\n")
	resp.ProtoMajor = ctx.Req.ProtoMajor
	resp.ProtoMinor = ctx.Req.ProtoMinor
	resp.Header.Set(ErrorHeader, err.Error())
	resp.Header.Set("Connection", "close")
	return resp
}
