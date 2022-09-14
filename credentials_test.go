// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"testing"

	"github.com/saucelabs/forwarder/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestSiteCredentialsMatcher(t *testing.T) {
	tests := []struct {
		name        string
		hostPort    string
		hostPortMap map[string]string
		portMap     map[string]string
		hostMap     map[string]string
		global      string
		isSet       bool
		expected    string
	}{
		{
			name:        "Matcher is not initialized",
			hostPortMap: map[string]string{},
			expected:    "",
			hostPort:    "abc:80",
			portMap:     map[string]string{},
			hostMap:     map[string]string{},
			isSet:       false,
			global:      "",
		},
		{
			name:        "Matches hostPort",
			hostPortMap: map[string]string{"abc:80": "user:pass"},
			expected:    "user:pass",
			hostPort:    "abc:80",
			portMap:     map[string]string{"*:80": "foo"},
			hostMap:     map[string]string{"abc:0": "bar"},
			isSet:       true,
			global:      "baz",
		},
		{
			name:        "Matches host wildcard",
			hostPortMap: map[string]string{"qux:80": "foo"},
			expected:    "user:pass",
			hostPort:    "abc:80",
			portMap:     map[string]string{"80": "user:pass"},
			hostMap:     map[string]string{"abc": "bar"},
			isSet:       true,
			global:      "baz",
		},
		{
			name:        "Matches port wildcard",
			hostPortMap: map[string]string{"qux:80": "foo"},
			expected:    "user:pass",
			hostPort:    "abc:80",
			portMap:     map[string]string{"90": "bar"},
			hostMap:     map[string]string{"abc": "user:pass"},
			isSet:       true,
			global:      "baz",
		},
		{
			name:        "Matches global wildcard",
			hostPortMap: map[string]string{"qux:80": "foo"},
			expected:    "user:pass",
			hostPort:    "abc:80",
			portMap:     map[string]string{"90": "bar"},
			hostMap:     map[string]string{"qux": "baz"},
			isSet:       true,
			global:      "user:pass",
		},
		{
			name:        "No match",
			hostPortMap: map[string]string{"qux:80": "foo"},
			expected:    "",
			hostPort:    "foobar:8080",
			portMap:     map[string]string{"80": "bar"},
			hostMap:     map[string]string{"qux": "baz"},
			isSet:       true,
			global:      "",
		},
	}

	logger.Setup(&LoggingOptions{
		Level: defaultProxyLoggingLevel,
	},
	)

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			matcher := siteCredentialsMatcher{
				siteCredentials:         tc.hostPortMap,
				siteCredentialsHost:     tc.hostMap,
				siteCredentialsPort:     tc.portMap,
				siteCredentialsWildcard: tc.global,
			}
			assert.Equalf(t, tc.isSet, matcher.isSet(), "Unexpected isSet: %v", matcher)
			creds := matcher.match(tc.hostPort)
			assert.Equalf(t, tc.expected, creds, "Unexpected result: %v", creds)
		})
	}
}
