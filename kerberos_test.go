// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"testing"

	"github.com/saucelabs/forwarder/log/slog"
)

func TestKerberosAdapterFailsWithoutConfig(t *testing.T) {
	cnf := KerberosConfig{}

	_, err := NewKerberosAdapter(cnf, slog.Default())
	// TODO: match expected error message
	if err != nil {
		t.Fatal(err)
	}
}
