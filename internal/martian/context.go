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
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/saucelabs/forwarder/internal/martian/log"
)

// Context provides information and storage for a single request/response pair.
// Contexts are linked to shared session that is used for multiple requests on
// a single connection.
type Context struct {
	session *Session
	n       uint64
	hash    uint32

	mu            sync.RWMutex
	vals          map[string]any
	skipRoundTrip bool
}

// Session provides information and storage about a connection.
type Session struct {
	mu     sync.RWMutex
	secure bool
}

type contextKey string

const martianKey contextKey = "martian.Context"

func FromContext(ctx context.Context) *Context {
	v := ctx.Value(martianKey)
	if v == nil {
		return nil
	}

	return v.(*Context) //nolint:forcetypeassert // We know the type.
}

// NewContext returns a context for the in-flight HTTP request.
func NewContext(req *http.Request) *Context {
	return FromContext(req.Context())
}

// TestContext builds a new session and associated context and returns the context.
// Intended for tests only.
func TestContext(req *http.Request) *Context {
	ctx := NewContext(req)
	if ctx != nil {
		return ctx
	}

	ctx = withSession(new(Session))
	*req = *req.Clone(ctx.addToContext(req.Context()))

	return ctx
}

// IsSecure returns whether the current session is from a secure connection,
// such as when receiving requests from a TLS connection that has been MITM'd.
func (s *Session) IsSecure() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.secure
}

// MarkSecure marks the session as secure.
func (s *Session) MarkSecure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.secure = true
}

// MarkInsecure marks the session as insecure.
func (s *Session) MarkInsecure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.secure = false
}

// addToContext returns context.Context with the current context to the passed context.
func (ctx *Context) addToContext(rctx context.Context) context.Context {
	if rctx == nil {
		rctx = context.Background()
	}
	mctx := context.WithValue(rctx, martianKey, ctx)
	return context.WithValue(mctx, log.TraceContextKey, ctx.ID())
}

// Session returns the session for the context.
func (ctx *Context) Session() *Session {
	return ctx.session
}

// ID returns the context ID.
func (ctx *Context) ID() string {
	return fmt.Sprintf("%d-%08x", ctx.n, ctx.hash)
}

// Get takes key and returns the associated value from the context.
func (ctx *Context) Get(key string) (any, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	val, ok := ctx.vals[key]

	return val, ok
}

// Set takes a key and associates it with val in the context. The value is
// persisted for the duration of the request and is removed on the following
// request.
func (ctx *Context) Set(key string, val any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.vals == nil {
		ctx.vals = make(map[string]any)
	}

	ctx.vals[key] = val
}

// SkipRoundTrip skips the round trip for the current request.
func (ctx *Context) SkipRoundTrip() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.skipRoundTrip = true
}

// SkippingRoundTrip returns whether the current round trip will be skipped.
func (ctx *Context) SkippingRoundTrip() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.skipRoundTrip
}

var nextID atomic.Uint64

// withSession builds a new context from an existing session.
// Session must be non-nil.
func withSession(s *Session) *Context {
	return &Context{
		session: s,
		n:       nextID.Add(1),
		hash:    uint32(time.Now().UnixNano()),
	}
}
