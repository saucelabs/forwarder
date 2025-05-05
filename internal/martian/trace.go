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

var _ log.StructuredLogger = TraceIDPrependingLogger{}

type TraceIDPrependingLogger struct {
	log.StructuredLogger
}

func (l TraceIDPrependingLogger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.FatalContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDPrependingLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.ErrorContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDPrependingLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.WarnContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDPrependingLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.InfoContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDPrependingLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.DebugContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDPrependingLogger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.TraceContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDPrependingLogger) args(ctx context.Context, args ...any) []any {
	if id := ContextTraceID(ctx); id != "" {
		return append(args, "id", id)
	}
	return args
}
