// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package wait

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaiterWaitForServerReady(t *testing.T) {
	w := defaultWaiter
	w.MaxWait = 1 * time.Second
	w.Backoff = 250 * time.Millisecond

	t.Run("error", func(t *testing.T) {
		err := w.WaitForServerReady("http://localhost:1")
		if err == nil {
			t.Fatal("expected error")
		}
		t.Log(err)
	})

	t.Run("status 500", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer s.Close()

		err := w.WaitForServerReady(s.URL)
		if err == nil {
			t.Fatal("expected error")
		}
		t.Log(err)
	})

	t.Run("status 200", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()

		err := w.WaitForServerReady(s.URL)
		if err != nil {
			t.Fatal(err)
		}
	})
}
