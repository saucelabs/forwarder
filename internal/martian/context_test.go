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
)

func TestContexts(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx := TestContext(req)
	ctx.Set("key", "value")
	got, ok := ctx.Get("key")
	if !ok {
		t.Errorf("ctx.Get(%q): got !ok, want ok", "key")
	}
	if want := "value"; got != want {
		t.Errorf("ctx.Get(%q): got %q, want %q", "key", got, want)
	}

	ctx.SkipRoundTrip()
	if !ctx.SkippingRoundTrip() {
		t.Error("ctx.SkippingRoundTrip(): got false, want true")
	}

	ctx2 := TestContext(req)
	if ctx != ctx2 {
		t.Error("TestContext(): got new context, want existing context")
	}
}
