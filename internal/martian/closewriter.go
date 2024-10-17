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
		return cw, ok
	}

	return valueAsCloseWriter(reflect.ValueOf(w))
}

// valueAsCloseWriter does BFS on v to find a first closeWriter in v or its fields.
func valueAsCloseWriter(v reflect.Value) (closeWriter, bool) {
	if v.CanInterface() {
		if cw, ok := v.Interface().(closeWriter); ok {
			return cw, true
		}

		// This works around issues with embedded interfaces.
		v = reflect.ValueOf(v.Interface())
	}

	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	for i := range v.NumField() {
		f := v.Field(i)

		if f.CanInterface() {
			if cw, ok := f.Interface().(closeWriter); ok {
				return cw, true
			}
		}
	}

	for i := range v.NumField() {
		if cw, ok := valueAsCloseWriter(v.Field(i)); ok {
			return cw, true
		}
	}

	return nil, false
}
