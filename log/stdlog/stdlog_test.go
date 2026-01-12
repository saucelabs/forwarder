// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package stdlog

import (
	"testing"

	flog "github.com/saucelabs/forwarder/log"
	"github.com/stretchr/testify/assert"
)

func TestLoggerNamedAllowsToPassCustomLevel(t *testing.T) {
	l := New(flog.DefaultConfig())
	f := l.Named("foo", WithLevel(0))
	assert.Equal(t, flog.Level(0), f.level)
}
