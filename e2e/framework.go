package e2e

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func statusCodeIs(t *testing.T, code int) func(*http.Response) {
	t.Helper()
	return func(resp *http.Response) {
		if resp.StatusCode != code {
			t.Errorf("expected status code %d, got %d", code, resp.StatusCode)
		}
	}
}

func errorMatches(t *testing.T, want string) func(err error) {
	t.Helper()
	return func(err error) {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error %q, got %q", want, err.Error())
		}
	}
}

func assertResponse(t *testing.T, client *http.Client, method, url string, body io.Reader, cks ...func(*http.Response)) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("+%v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if t.Failed() {
			t.Log("response body:", string(respBody))
		}
	}()

	for _, f := range cks {
		f(resp)
	}
}

func assertError(t *testing.T, client *http.Client, method, url string, body io.Reader, ck func(err error)) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err == nil {
		t.Fatal("expected error")
	}
	ck(err)
}
