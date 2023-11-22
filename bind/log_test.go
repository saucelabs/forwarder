// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/saucelabs/forwarder/httplog"
)

func TestHTTPLogUpdate(t *testing.T) {
	ptr := func(m httplog.Mode) *httplog.Mode {
		return &m
	}

	dst := []NamedParam[httplog.Mode]{
		{
			Name:  "foo",
			Param: new(httplog.Mode),
		},
		{
			Name:  "bar",
			Param: new(httplog.Mode),
		},
		{
			Name:  "baz",
			Param: new(httplog.Mode),
		},
	}

	src := []NamedParam[httplog.Mode]{
		{
			Name:  "",
			Param: ptr(httplog.None),
		},
		{
			Name:  "foo",
			Param: ptr(httplog.ShortURL),
		},
	}

	expected := []NamedParam[httplog.Mode]{
		{
			Name:  "foo",
			Param: ptr(httplog.ShortURL),
		},
		{
			Name:  "bar",
			Param: ptr(httplog.None),
		},
		{
			Name:  "baz",
			Param: ptr(httplog.None),
		},
	}

	httplogUpdate(dst, src)

	if diff := cmp.Diff(expected, dst); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n%s", diff)
	}
}

func TestHTTPLogUpdateWithoutDefault(t *testing.T) {
	ptr := func(m httplog.Mode) *httplog.Mode {
		return &m
	}

	dst := []NamedParam[httplog.Mode]{
		{
			Name:  "foo",
			Param: new(httplog.Mode),
		},
		{
			Name:  "bar",
			Param: new(httplog.Mode),
		},
		{
			Name:  "baz",
			Param: new(httplog.Mode),
		},
	}

	src := []NamedParam[httplog.Mode]{
		{
			Name:  "foo",
			Param: ptr(httplog.ShortURL),
		},
	}

	expected := []NamedParam[httplog.Mode]{
		{
			Name:  "foo",
			Param: ptr(httplog.ShortURL),
		},
		{
			Name:  "bar",
			Param: ptr(""),
		},
		{
			Name:  "baz",
			Param: ptr(""),
		},
	}

	httplogUpdate(dst, src)

	if diff := cmp.Diff(expected, dst); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n%s", diff)
	}
}
