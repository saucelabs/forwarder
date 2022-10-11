package forwarder

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthUtil(t *testing.T) {
	ba := &BasicAuthUtil{Header: "Foo"}
	r := httptest.NewRequest("GET", "/", nil)
	ba.SetBasicAuth(r, "user", "pass")

	if user, pass, ok := ba.BasicAuth(r); !ok || user != "user" || pass != "pass" {
		t.Errorf("BasicAuth failed, got %v %v %v", user, pass, ok)
	}
	if !ba.AuthenticatedRequest(r, "user", "pass") {
		t.Errorf("AuthenticatedRequest failed")
	}
}

func TestBasicAuthUtilWrap(t *testing.T) {
	ba := &BasicAuthUtil{Header: "Foo"}

	h := ba.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), "user", "pass")

	t.Run("Authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ba.SetBasicAuth(r, "user", "pass")

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
