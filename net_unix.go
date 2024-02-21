// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build unix

package forwarder

import (
	"fmt"
	"os"
	"syscall"
)

func enableTCPKeepAlive(fd uintptr) {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set SO_KEEPALIVE: %v\n", err)
	}
}
