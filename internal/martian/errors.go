// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package martian

import (
	"errors"
	"io"
	"net"
	"os"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	_ "unsafe" // for go:linkname
)

var errClose = errors.New("closing connection")

func errno(v error) uintptr {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Uintptr {
		return uintptr(rv.Uint())
	}
	return 0
}

//go:linkname h2ErrClosedBody golang.org/x/net/http2.errClosedBody
var h2ErrClosedBody error //nolint:errname // this is an exported variable from golang.org/x/net/http2

func init() {
	if h2ErrClosedBody == nil {
		panic("http2.errClosedBody not linked")
	}
}

// isClosedConnError reports whether err is an error from use of a closed network connection.
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, h2ErrClosedBody) {
		return true
	}

	// TODO(bradfitz): x/tools/cmd/bundle doesn't really support
	// build tags, so I can't make an http2_windows.go file with
	// Windows-specific stuff. Fix that and move this, once we
	// have a way to bundle this into std's net/http somehow.
	if runtime.GOOS == "windows" {
		var se *os.SyscallError
		if errors.As(err, &se) {
			if se.Syscall == "wsarecv" || se.Syscall == "wsasend" {
				const WSAECONNABORTED = 10053
				const WSAECONNRESET = 10054
				if n := errno(se.Err); n == WSAECONNRESET || n == WSAECONNABORTED {
					return true
				}
			}
		}
	}

	return strings.Contains(err.Error(), "use of closed network connection")
}

// isCloseable reports whether err is an error that indicates the client connection should be closed.
func isCloseable(err error) bool {
	if err == nil {
		return false
	}

	var neterr net.Error
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, io.ErrClosedPipe) ||
		(errors.As(err, &neterr) && !neterr.Timeout()) ||
		strings.Contains(err.Error(), "tls:")
}

type ErrorStatus struct { //nolint:errname // ErrorStatus is a type name not a variable.
	Err    error
	Status int
}

func (e ErrorStatus) Error() string {
	return e.Err.Error()
}

func (e ErrorStatus) Unwrap() error {
	return e.Err
}
