// Copyright 2023 Sauce Labs Inc., all rights reserved.

package martian

import (
	"fmt"
	"sync/atomic"
	"time"
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
