// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.

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

var _ log.StructuredLogger = TraceIDAppendingLogger{}

type TraceIDAppendingLogger struct {
	sl log.StructuredLogger
}

func NewTraceIDAppendingLogger(sl log.StructuredLogger) TraceIDAppendingLogger {
	return TraceIDAppendingLogger{sl: sl}
}

func (l TraceIDAppendingLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.sl.ErrorContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.sl.WarnContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.sl.InfoContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.sl.DebugContext(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) With(args ...any) log.StructuredLogger {
	return TraceIDAppendingLogger{l.sl.With(args...)}
}

func (l TraceIDAppendingLogger) args(ctx context.Context, args ...any) []any {
	if id := ContextTraceID(ctx); id != "" {
		return append(args, "id", id)
	}
	return args
}
