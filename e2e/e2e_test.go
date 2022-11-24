//go:build e2e

package e2e

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
)

func TestStatusCodes(t *testing.T) {
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

	e := expect(t, *httpbin)
	for i := range validStatusCodes {
		code := validStatusCodes[i]
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			t.Parallel()
			e.GET(fmt.Sprintf("/status/%d", code)).Expect().Status(code)
		})
	}
}

// There is a difference in client behaviour between sending HTTP and HTTPS requests here
// For HTTPS client issues a CONNECT request to the proxy and then sends the original request.
// In case of error the client receives a 502 Bad Gateway and interprets it as URL error.
// For HTTP client which receives a 502 Bad Gateway interprets it as a response.
func TestBadGateway(t *testing.T) {
	hosts := []string{
		"wronghost",
		"httpbin:1",
	}

	t.Run("http", func(t *testing.T) {
		for _, h := range hosts {
			expect(t, "http://"+h).GET("/status/200").Expect().Status(http.StatusBadGateway)
		}
	})

	t.Run("https", func(t *testing.T) {
		c := newHTTPClient(t)
		for _, h := range hosts {
			expectError(t, c, http.MethodGet, "https://"+h+"/status/200", nil, errorMatches(t, "Bad Gateway"))
		}
	})
}

func TestProxyLocalhost(t *testing.T) {
	hosts := []string{
		"localhost",
		"127.0.0.1",
	}

	for _, h := range hosts {
		if os.Getenv("FORWARDER_PROXY_LOCALHOST") == "true" {
			expect(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").Expect().Status(http.StatusOK)
		} else {
			expect(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").Expect().Status(http.StatusBadGateway)
		}
	}
}
