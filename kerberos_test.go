// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"testing"

	"github.com/saucelabs/forwarder/log/slog"
	"github.com/stretchr/testify/require"
)

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
