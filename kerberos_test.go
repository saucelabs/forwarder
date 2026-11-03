// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jcmturner/gokrb5/v8/iana/etypeID"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/saucelabs/forwarder/log/slog"
	"github.com/stretchr/testify/require"
)

const testKrb5Conf = `[libdefaults]
    default_realm = TEST.LOCAL
    dns_lookup_kdc = false

[realms]
    TEST.LOCAL = {
        kdc = 127.0.0.1:88
    }
`

func writeKerberosFixtures(t *testing.T) (cfgPath, ktPath string) {
	t.Helper()
	tmp := t.TempDir()

	cfgPath = filepath.Join(tmp, "krb5.conf")
	require.NoError(t, os.WriteFile(cfgPath, []byte(testKrb5Conf), 0o600))

	kt := keytab.New()
	require.NoError(t, kt.AddEntry("testuser", "TEST.LOCAL", "unused", time.Now(), 1, etypeID.AES256_CTS_HMAC_SHA1_96))
	ktBytes, err := kt.Marshal()
	require.NoError(t, err)

	ktPath = filepath.Join(tmp, "test.keytab")
	require.NoError(t, os.WriteFile(ktPath, ktBytes, 0o600))

	return cfgPath, ktPath
}

func TestKerberosAdapterFailsWithoutConfig(t *testing.T) {
	cnf := KerberosConfig{}

	_, err := NewKerberosAdapter(cnf, slog.Default())
	require.Error(t, err)
	require.ErrorContains(t, err, "kerberos config file (krb5.conf) not specified")

	cnf.CfgFilePath = "/tmp/test.cfg"

	_, err = NewKerberosAdapter(cnf, slog.Default())
	require.Error(t, err)
	require.ErrorContains(t, err, "kerberos keytab file not specified")

	cnf.KeyTabFilePath = "/tmp/keytab"

	_, err = NewKerberosAdapter(cnf, slog.Default())
	require.Error(t, err)
	require.ErrorContains(t, err, "kerberos username not specified")

	cnf.UserName = "user1"

	_, err = NewKerberosAdapter(cnf, slog.Default())
	require.Error(t, err)
	require.ErrorContains(t, err, "kerberos user realm not specified")
}

func TestKerberosAdapterDisablePAFXFAST(t *testing.T) {
	cfgPath, ktPath := writeKerberosFixtures(t)

	tests := []struct {
		name   string
		enable bool
		want   string
	}{
		{"off by default", false, `"DisablePAFXFast": false`},
		{"on when enabled", true, `"DisablePAFXFast": true`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adapter, err := NewKerberosAdapter(KerberosConfig{
				CfgFilePath:     cfgPath,
				KeyTabFilePath:  ktPath,
				UserName:        "testuser",
				UserRealm:       "TEST.LOCAL",
				DisablePAFXFAST: tc.enable,
			}, slog.Default())
			require.NoError(t, err)
			require.NotNil(t, adapter)

			var buf bytes.Buffer
			adapter.krb5client.Print(&buf)
			require.True(t, strings.Contains(buf.String(), tc.want),
				"expected gokrb5 Settings dump to contain %q, got:\n%s", tc.want, buf.String())
		})
	}
}
