// Copyright 2023 Sauce Labs Inc. All rights reserved.
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
	"crypto/tls"
	"io"
	"net"
	"reflect"
)

type closeWriter interface {
	CloseWrite() error
}

var (
	_ closeWriter = (*net.TCPConn)(nil)
	_ closeWriter = (*tls.Conn)(nil)
)

// asCloseWriter returns a closeWriter for w if it implements closeWriter.
// If w is a pointer to a struct, it checks if any of the fields implement closeWriter.
func asCloseWriter(w io.Writer) (closeWriter, bool) {
	if cw, ok := w.(closeWriter); ok {
		return cw, true
	}

	v := reflect.Indirect(reflect.ValueOf(w))

	if v.Kind() != reflect.Struct {
		return nil, false
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.CanInterface() {
			if cw, ok := f.Interface().(closeWriter); ok {
				return cw, true
			}
		}
	}

	return nil, false
}
