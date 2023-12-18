// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build windows

package forwarder

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func enableTCPKeepAlive(fd uintptr) {
	if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_KEEPALIVE, 1); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set SO_KEEPALIVE: %v\n", err)
	}
}
