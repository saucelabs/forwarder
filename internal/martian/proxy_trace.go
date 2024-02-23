// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.

package martian

import (
	"net/http"
)

// ProxyTrace is a set of hooks to run at various stages of a request.
// Any particular hook may be nil. Functions may be called concurrently
// from different goroutines and some may be called after the request has
// completed or failed.
type ProxyTrace struct {
	// ReadRequest is called with the result of reading the request.
	// It is called after the request has been read.
	ReadRequest func(ReadRequestInfo)

	// WroteResponse is called with the result of writing the response.
	// It is called after the response has been written.
	WroteResponse func(WroteResponseInfo)
}

type ReadRequestInfo struct {
	// Req is the request that was read.
	Req *http.Request
	// Err is any error encountered while reading the Request.
	Err error
}

func (p *Proxy) traceReadRequest(req *http.Request, err error) {
	if p.Trace != nil && p.Trace.ReadRequest != nil {
		p.Trace.ReadRequest(ReadRequestInfo{
			Req: req,
			Err: err,
		})
	}
}

type WroteResponseInfo struct {
	// Res is the response that was written.
	Res *http.Response
	// Err is any error encountered while writing the Request.
	Err error
}

func (p *Proxy) traceWroteResponse(res *http.Response, err error) {
	if p.Trace != nil && p.Trace.WroteResponse != nil {
		p.Trace.WroteResponse(WroteResponseInfo{
			Res: res,
			Err: err,
		})
	}
}
