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

package header

import (
	"net/http"
	"strings"
	"testing"
)

func TestViaModifier(t *testing.T) {
	m := NewViaModifier("martian")
	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := m.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("Via"), "1.1 martian"; !strings.HasPrefix(got, want) {
		t.Errorf("req.Header.Get(%q): got %q, want prefixed with %q", "Via", got, want)
	}

	req.Header.Set("Via", "1.0\talpha\t(martian)")
	if err := m.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("Via"), "1.0\talpha\t(martian), 1.1 martian"; !strings.HasPrefix(got, want) {
		t.Errorf("req.Header.Get(%q): got %q, want prefixed with %q", "Via", got, want)
	}

	m = NewViaModifierWithBoundary("martian", "boundary")
	req.Header.Set("Via", "1.0\talpha\t(martian), 1.1 martian-boundary, 1.1 beta")
	if err := m.ModifyRequest(req); err == nil {
		t.Fatal("ModifyRequest(): got nil, want request loop error")
	}
	if !req.Close {
		t.Fatalf("req.Close: got %v, want true", req.Close)
	}
}

func BenchmarkViaModifier(b *testing.B) {
	const via = "1.0\talpha\t(martian), 1.1 martian-boundary, 1.1 beta"

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	if err != nil {
		b.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	m := NewViaModifier("martian")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Header.Set("Via", via)
		if err := m.ModifyRequest(req); err != nil {
			b.Fatalf("ModifyRequest(): got %v, want no error", err)
		}
	}
}
