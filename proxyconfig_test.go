package forwarder

import (
	"net/url"
	"strings"
	"testing"
)

func TestProxyConfigValidate(t *testing.T) {
	var (
		validURL = newProxyURL(80, localProxyCredentialUsername, localProxyCredentialPassword)
		emptyURL = &url.URL{}
	)

	tests := []struct {
		name   string
		config ProxyConfig
		err    string
	}{
		{
			name: "normal",
			config: ProxyConfig{
				LocalProxyURI: validURL,
			},
		},
		{
			name: "both upstream and PAC are set",
			config: ProxyConfig{
				LocalProxyURI:    validURL,
				UpstreamProxyURI: newProxyURL(80, upstreamProxyCredentialUsername, upstreamProxyCredentialPassword),
				PACURI:           newProxyURL(80, "", ""),
			},
			err: "only one of upstream_proxy_uri or pac_uri can be set",
		},
		{
			name: "missing local proxy URI",
			err:  "local_proxy_uri is required",
		},
		{
			name: "invalid local proxy URI",
			config: ProxyConfig{
				LocalProxyURI: emptyURL,
			},
			err: "local_proxy_uri: invalid scheme",
		},
		{
			name: "invalid upstream proxy URI",
			config: ProxyConfig{
				LocalProxyURI:    validURL,
				UpstreamProxyURI: emptyURL,
			},
			err: "upstream_proxy_uri: invalid scheme",
		},
		{
			name: "invalid PAC server URI",
			config: ProxyConfig{
				LocalProxyURI: validURL,
				PACURI:        emptyURL,
			},
			err: "pac_uri: invalid scheme",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}
		})
	}
}

func TestParseUserInfo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "user:pass",
		},
		{
			name:  "no password",
			input: "user",
			err:   "expected username:password",
		},
		{
			name:  "empty password",
			input: "user:",
			err:   "password cannot be empty",
		},
		{
			name:  "no user",
			input: ":pass",
			err:   "username cannot be empty",
		},
		{
			name:  "empty",
			input: "",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			ui, err := ParseUserInfo(tc.input)
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}

			if ui.String() != tc.input {
				t.Errorf("expected %q, got %q", tc.input, ui.String())
			}
		})
	}
}

func TestParseProxyURI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "http://192.188.1.100:8080",
		},
		{
			name:  "invalid scheme",
			input: "tcp://192.188.1.100:8080",
			err:   "invalid scheme",
		},
		{
			name:  "no port",
			input: "http://192.188.1.100",
			err:   "port is required",
		},
		{
			name:  "port 0",
			input: "http://192.188.1.100:0",
			err:   "invalid port: 0",
		},
		{
			name:  "host too short",
			input: "http://foo:8080",
			err:   "invalid host",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProxyURI(tc.input)
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
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
