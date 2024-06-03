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
	"io"
	"testing"
)

type nopCloseWriterImpl struct{}

func (nopCloseWriterImpl) Write([]byte) (int, error) {
	return 0, nil
}

func (nopCloseWriterImpl) CloseWrite() error {
	return nil
}

type nopCloseWriterPtrImpl struct{}

func (*nopCloseWriterPtrImpl) Write([]byte) (int, error) {
	return 0, nil
}

func (*nopCloseWriterPtrImpl) CloseWrite() error {
	return nil
}

type nopReadCloserImpl struct {
	io.Reader
	io.Closer
	nopCloseWriterImpl
}

type nopReadCloserPtrImpl struct {
	io.Reader
	io.Closer
	nopCloseWriterPtrImpl
}

type body struct {
	io.ReadCloser
	n int64
}

func (b *body) Count() int64 {
	return b.n
}

func (b *body) Read(p []byte) (n int, err error) {
	n, err = b.ReadCloser.Read(p)
	b.n += int64(n)
	return
}

type rwcBody struct {
	body
}

func (b *rwcBody) Write(p []byte) (int, error) {
	return b.ReadCloser.(io.ReadWriteCloser).Write(p) //nolint:forcetypeassert // We know it's a ReadWriteCloser.
}

func TestAsCloseWriter(t *testing.T) {
	tests := []struct {
		name string
		w    io.Writer
	}{
		{
			name: "nopCloseWriterImpl",
			w:    nopCloseWriterImpl{},
		},
		{
			name: "nopCloseWriterImpl ptr",
			w:    &nopCloseWriterImpl{},
		},
		{
			name: "nopCloseWriterPtrImpl",
			w:    &nopCloseWriterPtrImpl{},
		},
		{
			name: "struct nopCloseWriterImpl",
			w: struct {
				nopCloseWriterImpl
			}{
				nopCloseWriterImpl{},
			},
		},
		{
			name: "struct nopCloseWriterImpl ptr",
			w: struct {
				*nopCloseWriterImpl
			}{
				&nopCloseWriterImpl{},
			},
		},
		{
			name: "struct nopCloseWriterPtrImpl",
			w: struct {
				*nopCloseWriterPtrImpl
			}{
				&nopCloseWriterPtrImpl{},
			},
		},
		{
			name: "struct interface nopCloseWriterImpl",
			w: struct {
				io.Writer
			}{
				nopCloseWriterImpl{},
			},
		},
		{
			name: "struct interface nopCloseWriterImpl ptr",
			w: struct {
				io.Writer
			}{
				&nopCloseWriterImpl{},
			},
		},
		{
			name: "struct interface nopCloseWriterPtrImpl",
			w: struct {
				io.Writer
			}{
				&nopCloseWriterPtrImpl{},
			},
		},
		{
			name: "embedded nopCloseWriterImpl",
			w: struct {
				io.Writer
			}{
				struct {
					io.Writer
				}{
					nopCloseWriterImpl{},
				},
			},
		},
		{
			name: "embedded ptr nopCloseWriterImpl",
			w: struct {
				io.Writer
			}{
				&struct {
					io.Writer
				}{
					nopCloseWriterImpl{},
				},
			},
		},
		{
			"rwcBody",
			&rwcBody{body{ReadCloser: nopReadCloserImpl{}}},
		},
		{
			"rwcBody ptr",
			&rwcBody{body{ReadCloser: &nopReadCloserPtrImpl{}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := asCloseWriter(tt.w)
			if !ok {
				t.Errorf("asCloseWriter(%#v) = _, false; want true", tt.w)
			}
			if got == nil {
				t.Errorf("asCloseWriter(%#v) = nil; want non-nil", tt.w)
			}
		})
	}
}
