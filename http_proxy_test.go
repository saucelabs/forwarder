// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHeaderByPrefixRemoverModifyRequest(t *testing.T) {
	withHeader := func(header http.Header) *http.Request {
		req, err := http.NewRequest(http.MethodGet, "http://example", nil) //nolint:gocritic // This is header test.
		if err != nil {
			t.Fatal(err)
		}
		req.Header = header
		return req
	}

	tests := []struct {
		name     string
		prefix   string
		req      *http.Request
		expected http.Header
	}{
		{
			name:   "smoke",
			prefix: http.CanonicalHeaderKey("RemoveMe"),
			req: withHeader(http.Header{
				http.CanonicalHeaderKey("RemoveMeByPrefix"): nil,
				http.CanonicalHeaderKey("RemoveMeBy"):       nil,
				http.CanonicalHeaderKey("RemoveMe"):         nil,
				http.CanonicalHeaderKey("DontRemoveMe"):     nil,
			}),
			expected: http.Header{
				http.CanonicalHeaderKey("DontRemoveMe"): nil,
			},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			mod := newHeaderRemover(tc.prefix)
			req := tc.req
			err := mod.ModifyRequest(req)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(req.Header, tc.expected); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
