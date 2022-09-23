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
