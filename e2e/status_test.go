//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
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

	var (
		ctx    = context.Background()
		client = newHTTPClient(t)
	)

	for i := range validStatusCodes {
		code := validStatusCodes[i]
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/status/%d", *httpbin, code), http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != code {
				t.Fatalf("expected status code %d, got %d", code, resp.StatusCode)
			}
		})
	}
}
