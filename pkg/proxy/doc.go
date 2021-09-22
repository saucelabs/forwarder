// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package proxy provides a simple proxy. The proxy can be protected with basic
// auth. It can also forward connections to a parent proxy, and authorize
// connections against that. Both local, and parent credentials can be set via
// environment variables. For local proxy credential, set `PROXY_CREDENTIAL`.
// For remote proxy credential, set `PROXY_PARENT_CREDENTIAL`.
package proxy
