// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package forwarder provides a simple forward proxy server.
// The proxy can be protected with HTTP basic authentication.
// It can also forward connections to a parent proxy, and authorize connections against that.
// Both local, and parent credentials can be set via environment variables.
package forwarder
