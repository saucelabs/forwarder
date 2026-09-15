// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httplog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saucelabs/forwarder/middleware"
)

func TestStructuredLoggerRequestHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	req.Header.Set("X-Sl-Job-Id", "abc123")

	res := &http.Response{StatusCode: 200, Request: req}

	tests := []struct {
		name    string
		headers []HeaderField
		wantKV  map[string]string
	}{
		{
			name:    "no headers configured",
			headers: nil,
			wantKV:  nil,
		},
		{
			name:    "renamed header emitted",
			headers: []HeaderField{{Header: "X-Sl-Job-Id", Field: "job_id"}},
			wantKV:  map[string]string{"job_id": "abc123"},
		},
		{
			name:    "missing header skipped",
			headers: []HeaderField{{Header: "X-Missing", Field: "missing"}},
			wantKV:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []any
			l := NewStructuredLogger(func(_ string, args ...any) {
				got = args
			}, ShortURL).WithRequestHeaders(tc.headers)

			l.LogFunc()(middleware.LogEntry{Request: req, Response: res, Status: res.StatusCode})

			kv := argsToMap(t, got)
			for k, v := range tc.wantKV {
				if kv[k] != v {
					t.Errorf("field %q: want %q, got %q (all args=%v)", k, v, kv[k], got)
				}
			}
			for _, h := range tc.headers {
				if req.Header.Get(h.Header) == "" {
					if _, ok := kv[h.Field]; ok {
						t.Errorf("field %q should be omitted when header is empty", h.Field)
					}
				}
			}
		})
	}
}

func argsToMap(t *testing.T, args []any) map[string]string {
	t.Helper()
	m := map[string]string{}
	for i := 0; i+1 < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			continue
		}
		if v, ok := args[i+1].(string); ok {
			m[k] = v
		}
	}
	return m
}
