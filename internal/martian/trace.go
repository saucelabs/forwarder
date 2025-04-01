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

var _ log.StructuredLogger = TraceIDAppendingLogger{}

type TraceIDAppendingLogger struct {
	log.StructuredLogger
}

func (l TraceIDAppendingLogger) Error(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.Error(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.Warn(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) Info(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.Info(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.StructuredLogger.Debug(ctx, msg, l.args(ctx, args...)...)
}

func (l TraceIDAppendingLogger) With(args ...any) log.StructuredLogger {
	return TraceIDAppendingLogger{l.StructuredLogger.With(args...)}
}

func (l TraceIDAppendingLogger) args(ctx context.Context, args ...any) []any {
	if id := ContextTraceID(ctx); id != "" {
		return append(args, "id", id)
	}
	return args
}
