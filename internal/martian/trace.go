// Copyright 2023 Sauce Labs Inc., all rights reserved.

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

func newTraceID() traceID {
	t := time.Now()

	return traceID{
		id:        fmt.Sprintf("%d-%08x", idSeq.Add(1), uint32(t.UnixNano())),
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
	if id := TraceID(ctx); id != "" {
		return fmt.Sprintf("[%s] %s", id, format)
	}
	return format
}
