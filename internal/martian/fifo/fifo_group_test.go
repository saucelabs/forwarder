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

package fifo

import (
	"errors"
	"net/http"
	"testing"

	"github.com/saucelabs/forwarder/internal/martian"
	_ "github.com/saucelabs/forwarder/internal/martian/header"
	"github.com/saucelabs/forwarder/internal/martian/martiantest"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
)

func TestModifyRequest(t *testing.T) {
	fg := NewGroup()
	tm := martiantest.NewModifier()

	fg.AddRequestModifier(tm)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := fg.ModifyRequest(req); err != nil {
		t.Fatalf("fg.ModifyRequest(): got %v, want no error", err)
	}
	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
}

func TestModifyRequestHaltsOnError(t *testing.T) {
	fg := NewGroup()

	reqerr := errors.New("request error")
	tm := martiantest.NewModifier()
	tm.RequestError(reqerr)
	fg.AddRequestModifier(tm)

	tm2 := martiantest.NewModifier()
	fg.AddRequestModifier(tm2)

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := fg.ModifyRequest(req); err != reqerr {
		t.Fatalf("fg.ModifyRequest(): got %v, want %v", err, reqerr)
	}

	if tm2.RequestModified() {
		t.Error("tm2.RequestModified(): got true, want false")
	}
}

func TestModifyRequestAggregatesErrors(t *testing.T) {
	fg := NewGroup()
	fg.SetAggregateErrors(true)

	reqerr1 := errors.New("1. request error")
	tm := martiantest.NewModifier()
	tm.RequestError(reqerr1)
	fg.AddRequestModifier(tm)

	tm2 := martiantest.NewModifier()
	reqerr2 := errors.New("2. request error")
	tm2.RequestError(reqerr2)
	fg.AddRequestModifier(tm2)

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	merr := martian.NewMultiError()
	merr.Add(reqerr1)
	merr.Add(reqerr2)

	if err := fg.ModifyRequest(req); err == nil {
		t.Fatalf("fg.ModifyRequest(): got %v, want not nil", err)
	}
	if err := fg.ModifyRequest(req); err.Error() != merr.Error() {
		t.Fatalf("fg.ModifyRequest(): got %v, want %v", err, merr)
	}

	if err, want := fg.ModifyRequest(req), "1. request error\n2. request error"; err.Error() != want {
		t.Fatalf("fg.ModifyRequest(): got %v, want %v", err, want)
	}
}

func TestModifyResponse(t *testing.T) {
	fg := NewGroup()
	tm := martiantest.NewModifier()

	fg.AddResponseModifier(tm)

	res := proxyutil.NewResponse(200, nil, nil)
	if err := fg.ModifyResponse(res); err != nil {
		t.Fatalf("fg.ModifyResponse(): got %v, want no error", err)
	}
	if !tm.ResponseModified() {
		t.Error("tm.ResponseModified(): got false, want true")
	}
}

func TestModifyResponseHaltsOnError(t *testing.T) {
	fg := NewGroup()

	reserr := errors.New("request error")
	tm := martiantest.NewModifier()
	tm.ResponseError(reserr)
	fg.AddResponseModifier(tm)

	tm2 := martiantest.NewModifier()
	fg.AddResponseModifier(tm2)

	res := proxyutil.NewResponse(200, nil, nil)
	if err := fg.ModifyResponse(res); err != reserr {
		t.Fatalf("fg.ModifyResponse(): got %v, want %v", err, reserr)
	}

	if tm2.ResponseModified() {
		t.Error("tm2.ResponseModified(): got true, want false")
	}
}

func TestModifyResponseAggregatesErrors(t *testing.T) {
	fg := NewGroup()
	fg.SetAggregateErrors(true)

	reserr1 := errors.New("1. response error")
	tm := martiantest.NewModifier()
	tm.ResponseError(reserr1)
	fg.AddResponseModifier(tm)

	tm2 := martiantest.NewModifier()
	reserr2 := errors.New("2. response error")
	tm2.ResponseError(reserr2)
	fg.AddResponseModifier(tm2)

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	martian.TestContext(req, nil, nil)

	res := proxyutil.NewResponse(200, nil, req)

	merr := martian.NewMultiError()
	merr.Add(reserr1)
	merr.Add(reserr2)

	if err := fg.ModifyResponse(res); err == nil {
		t.Fatalf("fg.ModifyResponse(): got %v, want %v", err, merr)
	}

	if err := fg.ModifyResponse(res); err.Error() != merr.Error() {
		t.Fatalf("fg.ModifyResponse(): got %v, want %v", err, merr)
	}
}

func TestModifyResponseInlineGroupsAggregateErrors(t *testing.T) {
	fg1 := NewGroup()
	fg1.SetAggregateErrors(true)
	reserr1 := errors.New("1. response error")
	tm1 := martiantest.NewModifier()
	tm1.ResponseError(reserr1)
	fg1.AddResponseModifier(tm1)

	fg2 := NewGroup()
	fg2.SetAggregateErrors(true)
	reserr2 := errors.New("2. response error")
	tm2 := martiantest.NewModifier()
	tm2.ResponseError(reserr2)
	fg2.AddResponseModifier(tm2)

	fg3 := NewGroup()
	fg3.SetAggregateErrors(true)
	reserr3 := errors.New("3. response error")
	tm3 := martiantest.NewModifier()
	tm3.ResponseError(reserr3)
	fg3.AddResponseModifier(tm3)

	fg2.AddResponseModifier(fg3)
	fg1.AddResponseModifier(fg2)
	ig := fg1.ToImmutable()

	if len(ig.resmods) != 3 {
		t.Fatalf("inner groups should be inlined")
	}

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	martian.TestContext(req, nil, nil)

	res := proxyutil.NewResponse(200, nil, req)

	merr := martian.NewMultiError()
	merr.Add(reserr1)
	merr.Add(reserr2)
	merr.Add(reserr3)

	if err := ig.ModifyResponse(res); err == nil {
		t.Fatalf("ig.ModifyResponse(): got %v, want %v", err, merr)
	}

	if err := ig.ModifyResponse(res); err.Error() != merr.Error() {
		t.Fatalf("ig.ModifyResponse(): got %v, want %v", err, merr)
	}
}

func TestModifyRequestInlineGroupsAggregateErrors(t *testing.T) {
	fg1 := NewGroup()
	fg1.SetAggregateErrors(true)
	reqerr1 := errors.New("1. request error")
	tm1 := martiantest.NewModifier()
	tm1.RequestError(reqerr1)
	fg1.AddRequestModifier(tm1)

	fg2 := NewGroup()
	fg2.SetAggregateErrors(true)
	reqerr2 := errors.New("2. request error")
	tm2 := martiantest.NewModifier()
	tm2.RequestError(reqerr2)
	fg2.AddRequestModifier(tm2)

	fg3 := NewGroup()
	fg3.SetAggregateErrors(true)
	reqerr3 := errors.New("3. request error")
	tm3 := martiantest.NewModifier()
	tm3.RequestError(reqerr3)
	fg3.AddRequestModifier(tm3)

	fg2.AddRequestModifier(fg3)
	fg1.AddRequestModifier(fg2)
	ig := fg1.ToImmutable()

	if len(ig.reqmods) != 3 {
		t.Fatalf("inner groups should be inlined")
	}

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	merr := martian.NewMultiError()
	merr.Add(reqerr1)
	merr.Add(reqerr2)
	merr.Add(reqerr3)

	if err := ig.ModifyRequest(req); err == nil {
		t.Fatalf("ig.ModifyRequest(): got %v, want not nil", err)
	}
	if err := ig.ModifyRequest(req); err.Error() != merr.Error() {
		t.Fatalf("ig.ModifyRequest(): got %v, want %v", err, merr)
	}

	if err, want := ig.ModifyRequest(req), "1. request error\n2. request error\n3. request error"; err.Error() != want {
		t.Fatalf("ig.ModifyRequest(): got %v, want %v", err, want)
	}
}
