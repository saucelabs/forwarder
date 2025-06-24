// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package martianlog

import (
	"github.com/saucelabs/forwarder/internal/martian"
	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/log"
)

func SetLogger(l log.Logger) {
	sl := newStructuredLoggerAdapter(l)
	tl := martian.NewTraceIDAppendingLogger(sl)
	martianlog.SetLogger(tl)
}
