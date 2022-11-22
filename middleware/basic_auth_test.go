package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	ba := NewBasicAuth()
	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("user", "pass")

	if user, pass, ok := ba.BasicAuth(r); !ok || user != "user" || pass != "pass" {
		t.Errorf("BasicAuth failed, got %v %v %v", user, pass, ok)
	}
	if !ba.AuthenticatedRequest(r, "user", "pass") {
		t.Errorf("AuthenticatedRequest failed")
	}
}

func TestBasicAuthWrap(t *testing.T) {
	ba := NewBasicAuth()

	h := ba.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Foo") != "" {
			t.Errorf("Auth header should not be forwarded")
		}
		w.WriteHeader(http.StatusOK)
	}), "user", "pass")

	t.Run("Authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetBasicAuth("user", "pass")

		h.ServeHTTP(w, r)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("Authenticated failed, got %v", w.Result().StatusCode)
		}
	})

	t.Run("Not Authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		h.ServeHTTP(w, r)
		if w.Result().StatusCode != http.StatusUnauthorized {
			t.Errorf("Authenticated failed, got %v", w.Result().StatusCode)
		}
	})
}
