// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.

package martian

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/saucelabs/forwarder/internal/martian/log"
)

// traceID is a unique identifier for a request.
type traceID struct {
	id        string
	createdAt time.Time
}

var idSeq atomic.Uint64

func newTraceID(id string) traceID {
	t := time.Now()
	n := idSeq.Add(1)

	if id == "" {
		id = fmt.Sprintf("%d-%08x", n, uint32(t.UnixNano())) //nolint:gosec // no overflow
	}

	return traceID{
		id:        id,
		createdAt: t,
	}
}

func (t traceID) String() string {
	return t.id
}

func (t traceID) Duration() time.Duration {
	return time.Since(t.createdAt)
}

type TraceIDPrependingLogger struct {
	log.Logger
}

func (l TraceIDPrependingLogger) Infof(ctx context.Context, format string, args ...any) {
	l.Logger.Infof(ctx, l.format(ctx, format), args...)
}

func (l TraceIDPrependingLogger) Debugf(ctx context.Context, format string, args ...any) {
	l.Logger.Debugf(ctx, l.format(ctx, format), args...)
}

func (l TraceIDPrependingLogger) Errorf(ctx context.Context, format string, args ...any) {
	l.Logger.Errorf(ctx, l.format(ctx, format), args...)
}

var _ log.Logger = TraceIDPrependingLogger{}

func (l TraceIDPrependingLogger) format(ctx context.Context, format string) string {
	if id := ContextTraceID(ctx); id != "" {
		return fmt.Sprintf("[%s] %s", id, format)
	}
	return format
}
