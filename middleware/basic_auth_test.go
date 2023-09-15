// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	ba := NewBasicAuth()

	t.Run("basic auth", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r.SetBasicAuth("user", "pass")

		if user, pass, ok := ba.BasicAuth(r); !ok || user != "user" || pass != "pass" {
			t.Errorf("BasicAuth failed, got %v %v %v", user, pass, ok)
		}
		if !ba.AuthenticatedRequest(r, "user", "pass") {
			t.Errorf("AuthenticatedRequest failed")
		}
	})

	t.Run("url encoded", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r.SetBasicAuth("user%3f", "p%2fass")

		if user, pass, ok := ba.BasicAuth(r); !ok || user != "user?" || pass != "p/ass" {
			t.Errorf("BasicAuth failed, got %v %v %v", user, pass, ok)
		}
		if !ba.AuthenticatedRequest(r, "user?", "p/ass") {
			t.Errorf("AuthenticatedRequest failed")
		}
	})
}

func TestBasicAuthWrap(t *testing.T) {
	ba := NewBasicAuth()

	h := ba.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Foo") != "" {
			t.Errorf("auth header should not be forwarded")
		}
		w.WriteHeader(http.StatusOK)
	}), "user", "pass")

	t.Run("Authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r.SetBasicAuth("user", "pass")

		h.ServeHTTP(w, r)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("got %v", w.Result().StatusCode)
		}
	})

	t.Run("Not Authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

		h.ServeHTTP(w, r)
		if w.Result().StatusCode != http.StatusUnauthorized {
			t.Errorf("got %v", w.Result().StatusCode)
		}
	})
}
