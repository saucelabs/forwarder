// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package connfix

import (
	"io"
	"net"
)

type (
	readFromMixin   struct{ net.Conn }
	writeToMixin    struct{ net.Conn }
	closeWriteMixin struct{ net.Conn }
)

func (rf readFromMixin) ReadFrom(r io.Reader) (int64, error) {
	return rf.Conn.(io.ReaderFrom).ReadFrom(r) //nolint:forcetypeassert // we know the type is correct
}

var _ io.ReaderFrom = readFromMixin{}

func (wt writeToMixin) WriteTo(w io.Writer) (int64, error) {
	return wt.Conn.(io.WriterTo).WriteTo(w) //nolint:forcetypeassert // we know the type is correct
}

var _ io.WriterTo = writeToMixin{}

type _closeWriter interface {
	CloseWrite() error
}

func (cw closeWriteMixin) CloseWrite() error {
	return cw.Conn.(_closeWriter).CloseWrite() //nolint:forcetypeassert // we know the type is correct
}

var _ _closeWriter = closeWriteMixin{}

const (
	readerFrom = 1 << iota
	writerTo
	closeWriter
)

func flags(conn net.Conn) uint8 {
	var f uint8
	if _, ok := conn.(io.ReaderFrom); ok {
		f |= readerFrom
	}
	if _, ok := conn.(io.WriterTo); ok {
		f |= writerTo
	}
	if _, ok := conn.(_closeWriter); ok {
		f |= closeWriter
	}
	return f
}

// Combine returns a net.Conn that combines the functionality of the outer and inner net.Conn.
// It detects if the inner net.Conn provides any of the following interfaces:
//
//   - io.ReaderFrom
//   - io.WriterTo
//   - func CloseWrite()
//
// and returns a net.Conn that implements the same interfaces.
//
// The outer net.Conn may also provide these functions,
// they are used only if the inner net.Conn also provides them.
// This allows the implementors of the outer net.Conn to provide implementations that are used when possible.
func Combine(outer, inner net.Conn) net.Conn {
	readFromMixin := func() readFromMixin {
		if _, ok := outer.(io.ReaderFrom); ok {
			return readFromMixin{outer}
		}
		return readFromMixin{inner}
	}
	writeToMixin := func() writeToMixin {
		if _, ok := outer.(io.WriterTo); ok {
			return writeToMixin{outer}
		}
		return writeToMixin{inner}
	}
	closeWriteMixin := func() closeWriteMixin {
		if _, ok := outer.(_closeWriter); ok {
			return closeWriteMixin{outer}
		}
		return closeWriteMixin{inner}
	}

	switch flags(inner) {
	case 0:
		return struct {
			net.Conn
		}{outer}
	case readerFrom:
		return struct {
			net.Conn
			io.ReaderFrom
		}{outer, readFromMixin()}
	case writerTo:
		return struct {
			net.Conn
			io.WriterTo
		}{outer, writeToMixin()}
	case closeWriter:
		return struct {
			net.Conn
			_closeWriter
		}{outer, closeWriteMixin()}
	case readerFrom | writerTo:
		return struct {
			net.Conn
			io.ReaderFrom
			io.WriterTo
		}{outer, readFromMixin(), writeToMixin()}
	case readerFrom | closeWriter:
		return struct {
			net.Conn
			io.ReaderFrom
			_closeWriter
		}{outer, readFromMixin(), closeWriteMixin()}
	case writerTo | closeWriter:
		return struct {
			net.Conn
			io.WriterTo
			_closeWriter
		}{outer, writeToMixin(), closeWriteMixin()}
	case readerFrom | writerTo | closeWriter:
		return struct {
			net.Conn
			io.ReaderFrom
			io.WriterTo
			_closeWriter
		}{outer, readFromMixin(), writeToMixin(), closeWriteMixin()}
	default:
		panic("unreachable")
	}
}
