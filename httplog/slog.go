// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httplog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/saucelabs/forwarder/internal/martian"
	"github.com/saucelabs/forwarder/middleware"
	"github.com/saucelabs/forwarder/utils/maps"
)

// request represents an HTTP request for structured logging.
type request struct {
	Method           string              `json:"method,omitempty"`
	URL              string              `json:"url,omitempty"`
	Protocol         string              `json:"protocol,omitempty"`
	Host             string              `json:"host,omitempty"`
	Headers          map[string][]string `json:"headers,omitempty"`
	TransferEncoding []string            `json:"transfer_encoding,omitempty"`
	ContentLength    int64               `json:"content_length,omitempty"`
	Body             string              `json:"body,omitempty"`
	BodyError        string              `json:"body_error,omitempty"`
	Trailers         map[string][]string `json:"trailers,omitempty"`
}

func (r request) String() string {
	var b strings.Builder

	b.WriteString(r.Method)
	b.WriteRune(' ')
	b.WriteString(r.URL)

	add := func(k, v string) {
		if v == "" {
			return
		}
		b.WriteString(", ")
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}

	add("protocol", r.Protocol)
	add("host", r.Host)
	add("headers", formatMap(r.Headers))
	add("transfer_encoding", formatSlice(r.TransferEncoding))
	if r.ContentLength > 0 {
		add("content_length", strconv.FormatInt(r.ContentLength, 10))
	}
	add("body", r.Body)
	add("body_error", r.BodyError)
	add("trailers", formatMap(r.Trailers))

	return b.String()
}

// response represents an HTTP response for structured logging.
type response struct {
	StatusCode       int                 `json:"status_code,omitempty"`
	StatusText       string              `json:"status_text,omitempty"`
	Protocol         string              `json:"protocol,omitempty"`
	Headers          map[string][]string `json:"headers,omitempty"`
	TransferEncoding []string            `json:"transfer_encoding,omitempty"`
	ContentLength    int64               `json:"content_length,omitempty"`
	Body             string              `json:"body,omitempty"`
	BodyError        string              `json:"body_error,omitempty"`
	Trailers         map[string][]string `json:"trailers,omitempty"`
}

func (r response) String() string {
	var b strings.Builder

	b.WriteString(strconv.Itoa(r.StatusCode))

	add := func(k, v string) {
		if v == "" {
			return
		}
		b.WriteString(", ")
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}

	add("status_text", r.StatusText)
	add("protocol", r.Protocol)
	add("headers", formatMap(r.Headers))
	add("transfer_encoding", formatSlice(r.TransferEncoding))
	if r.ContentLength > 0 {
		add("content_length", strconv.FormatInt(r.ContentLength, 10))
	}
	add("body", r.Body)
	add("body_error", r.BodyError)
	add("trailers", formatMap(r.Trailers))

	return b.String()
}

func formatMap(m map[string][]string) string {
	if len(m) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteByte('[')

	keys := maps.Keys(m)
	sort.Strings(keys) // Stable order.
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(formatSlice(m[k]))
	}

	b.WriteByte(']')
	return b.String()
}

func formatSlice(s []string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) == 1 {
		return s[0]
	}

	return "[" + strings.Join(s, ",") + "]"
}

type structuredLogBuilder struct {
	req      request
	res      response
	duration string
	id       string
}

// WithShortURL sets the URL using a short form along with basic fields.
func (b *structuredLogBuilder) WithShortURL(e middleware.LogEntry) {
	if e.Request == nil {
		return
	}
	shortURL := buildShortURL(e.Request.URL)
	b.initBasicFields(e, shortURL)
}

func buildShortURL(u *url.URL) string {
	scheme, host, path := u.Scheme, u.Host, u.Path
	if scheme != "" {
		scheme += "://"
	}
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	return scheme + host + path
}

// WithURL sets the URL using the redacted form along with basic fields.
func (b *structuredLogBuilder) WithURL(e middleware.LogEntry) {
	if e.Request == nil {
		return
	}
	redactedURL := e.Request.URL.Redacted()
	b.initBasicFields(e, redactedURL)
}

func (b *structuredLogBuilder) initBasicFields(e middleware.LogEntry, urlStr string) {
	req := e.Request
	b.req.Method = req.Method
	b.req.URL = urlStr

	b.res.StatusCode = e.Status

	b.duration = e.Duration.String()
	b.id = martian.ContextTraceID(req.Context())
}

// WithHeaders copies headers, trailers, and other metadata from the request and response.
func (b *structuredLogBuilder) WithHeaders(e middleware.LogEntry) {
	req := e.Request
	if req == nil {
		return
	}

	b.req.Protocol = fmt.Sprintf("HTTP/%d.%d", req.ProtoMajor, req.ProtoMinor)
	b.req.Host = req.Host
	b.req.Headers = req.Header.Clone()
	if len(req.TransferEncoding) > 0 {
		b.req.TransferEncoding = req.TransferEncoding
	}
	if req.ContentLength >= 0 {
		b.req.ContentLength = req.ContentLength
	}
	if len(req.Trailer) > 0 {
		b.req.Trailers = req.Trailer.Clone()
	}

	res := e.Response
	if res == nil {
		return
	}

	b.res.Protocol = fmt.Sprintf("HTTP/%d.%d", res.ProtoMajor, res.ProtoMinor)
	parts := strings.SplitN(res.Status, " ", 2)
	if len(parts) == 2 {
		b.res.StatusText = parts[1]
	}
	b.res.Headers = res.Header.Clone()
	if len(res.TransferEncoding) > 0 {
		b.res.TransferEncoding = res.TransferEncoding
	}
	if res.ContentLength >= 0 {
		b.res.ContentLength = res.ContentLength
	}
	if len(res.Trailer) > 0 {
		b.res.Trailers = res.Trailer.Clone()
	}
}

// WithBody reads and logs the body of the request and response.
// Note: For CONNECT requests with successful responses, the body is skipped.
func (b *structuredLogBuilder) WithBody(e middleware.LogEntry) {
	req := e.Request
	if req == nil || req.Method == http.MethodConnect && e.Status/100 == 2 {
		return
	}

	if req.Body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			b.req.BodyError = err.Error()
		} else {
			b.req.Body = string(data)
			// Restore the body for further processing.
			req.Body = io.NopCloser(bytes.NewReader(data))
		}
	}

	res := e.Response
	if res != nil && res.Body != nil {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			b.res.BodyError = err.Error()
		} else {
			b.res.Body = string(data)
			// Restore the body.
			res.Body = io.NopCloser(bytes.NewReader(data))
		}
	}
}

// Args returns a slice of key-value pairs for logging purposes.
func (b *structuredLogBuilder) Args() []any {
	return []any{"request", b.req, "response", b.res, "duration", b.duration, "id", b.id}
}
