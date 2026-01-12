// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package reflectx

import (
	"io"
	"reflect"
	"testing"
)

type closeWriter interface {
	CloseWrite() error
}

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

func TestLookupImplCloseWriter(t *testing.T) {
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
			got, ok := LookupImpl[closeWriter](reflect.ValueOf(tt.w))
			if !ok {
				t.Errorf("asCloseWriter(%#v) = _, false; want true", tt.w)
			}
			if got == nil {
				t.Errorf("asCloseWriter(%#v) = nil; want non-nil", tt.w)
			}
		})
	}
}
