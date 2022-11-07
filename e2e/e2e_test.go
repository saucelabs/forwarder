//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestStatusCodes(t *testing.T) {
	if *httpbin == "" {
		t.Fatal("httpbin URL not set")
	}

	// List of all valid status codes plus some non-standard ones.
	// See https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
	validStatusCodes := []int{
		// FIXME: proxy wrongly supports 1xx, see #113
		// 100, 101, 102, 103, 122,
		200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
		300, 301, 302, 303, 304, 305, 306, 307, 308,
		400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423, 424, 425, 426, 428, 429, 431, 451,
		500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511, 599,
	}

	client := newHTTPClient(t)

	for i := range validStatusCodes {
		code := validStatusCodes[i]
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			assertResponse(t, client, http.MethodGet, *httpbin+"/status/"+fmt.Sprint(code), nil, statusCodeIs(t, code))
		})
	}
}

func TestBadGateway(t *testing.T) {
	client := newHTTPClient(t)

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "DNS error",
			url:  "https://wronghost",
		},
		{
			name: "connection refused",
			url:  "https://httpbin" + ":1",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			assertError(t, client, http.MethodGet, tc.url, nil, errorMatches(t, "Bad Gateway"))
		})
	}
}

func TestProxyLocalhost(t *testing.T) {
	code := http.StatusBadGateway
	if os.Getenv("FORWARDER_PROXY_LOCALHOST") == "true" {
		code = http.StatusOK
	}

	client := newHTTPClient(t)

	hosts := []string{
		"localhost",
		"127.0.0.1",
	}
	for _, h := range hosts {
		assertResponse(t, client, http.MethodGet, "http://"+h+":10000/version", nil, statusCodeIs(t, code))
	}
}
