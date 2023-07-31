// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net/url"
	"testing"
)

var base64Tests = []struct {
	decoded, encoded string
}{
	// RFC 3548 examples
	{"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l+"},
	{"\x14\xfb\x9c\x03\xd9", "FPucA9k="},
	{"\x14\xfb\x9c\x03", "FPucAw=="},

	// RFC 4648 examples
	{"", ""},
	{"f", "Zg=="},
	{"fo", "Zm8="},
	{"foo", "Zm9v"},
	{"foob", "Zm9vYg=="},
	{"fooba", "Zm9vYmE="},
	{"foobar", "Zm9vYmFy"},

	// Wikipedia examples
	{"sure.", "c3VyZS4="},
	{"sure", "c3VyZQ=="},
	{"sur", "c3Vy"},
	{"su", "c3U="},
	{"leasure.", "bGVhc3VyZS4="},
	{"easure.", "ZWFzdXJlLg=="},
	{"asure.", "YXN1cmUu"},
	{"sure.", "c3VyZS4="},
}

func TestReadURLData(t *testing.T) {
	for i := range base64Tests {
		tc := base64Tests[i]
		t.Run(tc.encoded, func(t *testing.T) {
			u := url.URL{
				Scheme: "data",
				Opaque: "//base64," + tc.encoded,
			}
			b, err := ReadURLString(&u, nil)
			if err != nil {
				t.Fatal(err)
			}
			if b != tc.decoded {
				t.Fatalf("expected %q, got %q", tc.decoded, b)
			}
		})
	}
}

func TestReadFileOrBase64(t *testing.T) {
	for i := range base64Tests {
		tc := base64Tests[i]
		t.Run(tc.encoded, func(t *testing.T) {
			b, err := ReadFileOrBase64("data:" + tc.encoded)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.decoded {
				t.Fatalf("expected %q, got %q", tc.decoded, b)
			}
		})
	}
}
