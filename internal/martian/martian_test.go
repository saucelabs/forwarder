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
	"net/http"
	"testing"

	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
)

func TestModifierFuncs(t *testing.T) {
	const value = "true"

	reqmod := RequestModifierFunc(
		func(req *http.Request) error {
			req.Header.Set("Request-Modified", value)
			return nil
		})

	resmod := ResponseModifierFunc(
		func(res *http.Response) error {
			res.Header.Set("Response-Modified", value)
			return nil
		})

	req, err := http.NewRequest(http.MethodGet, "http://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("reqmod.ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("Request-Modified"), value; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", "Request-Modified", got, want)
	}

	res := proxyutil.NewResponse(200, nil, req)

	if err := resmod.ModifyResponse(res); err != nil {
		t.Fatalf("resmod.ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.Header.Get("Response-Modified"), value; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Response-Modified", got, want)
	}
}
