package forwarder

import (
	"net/url"
	"strings"
	"testing"
)

func TestProxyConfigValidate(t *testing.T) {
	emptyURI := &url.URL{}

	tests := []struct {
		name   string
		config ProxyConfig
		err    string
	}{
		{
			name: "Valid",
			config: ProxyConfig{
				LocalProxyURI: newProxyURL(80, localProxyCredentialUsername, localProxyCredentialPassword),
			},
		},
		{
			name: "Both upstream and PAC are set",
			config: ProxyConfig{
				LocalProxyURI:    newProxyURL(80, localProxyCredentialUsername, localProxyCredentialPassword),
				UpstreamProxyURI: newProxyURL(80, upstreamProxyCredentialUsername, upstreamProxyCredentialPassword),
				PACURI:           newProxyURL(80, "", ""),
			},
			err: "excluded_with",
		},
		{
			name: "Missing local proxy URI",
			err:  "required",
		},
		{
			name: "Invalid local proxy URI",
			config: ProxyConfig{
				LocalProxyURI: emptyURI,
			},
			err: "proxyURI",
		},
		{
			name: "Invalid upstream proxy URI",
			config: ProxyConfig{
				LocalProxyURI:    newProxyURL(80, "", ""),
				UpstreamProxyURI: emptyURI,
			},
			err: "proxyURI",
		},
		{
			name: "Invalid PAC server URI",
			config: ProxyConfig{
				PACURI: emptyURI,
			},
			err: "proxyURI",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err != nil && tc.err == "" {
				t.Fatalf("Expected no error, got %s", err)
			}

			if err == nil && tc.err != "" {
				t.Fatal("Expected error, got none")
			}

			if err != nil && tc.err != "" && !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("Expected error to contain %s, got %s", tc.err, err)
			}
		})
	}
}

func TestParseDNSURIDefaults(t *testing.T) {
	u, err := ParseDNSURI("1.1.1.1")
	if err != nil {
		t.Fatalf("expected success, got %q", err)
	}
	if expected := "udp://1.1.1.1:53"; u.String() != expected {
		t.Errorf("expected %q, got %q", expected, u.String())
	}
}

func TestParseDNSURI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "udp://1.1.1.1:53",
		},
		{
			name:  "custom scheme",
			input: "tcp://1.1.1.1:53",
		},
		{
			name:  "custom port",
			input: "udp://1.1.1.1:153",
		},
		{
			name:  "custom host",
			input: "udp://8.8.8.8:53",
		},
		{
			name:  "hostname",
			input: "udp://saucelabs.com:53",
			err:   "invalid hostname",
		},
		{
			name:  "unsupported scheme",
			input: "https://1.1.1.1:53",
			err:   "invalid protocol: https",
		},
		{
			name:  "port 0",
			input: "udp://1.1.1.1:0",
			err:   "invalid port: 0",
		},
		{
			name:  "URL path",
			input: "udp://1.1.1.1:53/path",
			err:   "path, query, and fragment are not allowed in DNS URI",
		},
		{
			name:  "URL query",
			input: "udp://1.1.1.1:53/?query=1",
			err:   "path, query, and fragment are not allowed in DNS URI",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			u, err := ParseDNSURI(tc.input)
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}

			if u.String() != tc.input {
				t.Errorf("expected %q, got %q", tc.input, u.String())
			}
		})
	}
}
