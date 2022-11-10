package middleware

import (
	"net/http"
	"time"
)

type LogEntry struct {
	*http.Request
	Status   int
	Written  int64
	Duration time.Duration
}

type Logger func(e LogEntry)

func (l Logger) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		d, ok := w.(delegator)
		if !ok {
			panic("logger middleware requires delegator")
		}
		h.ServeHTTP(w, r)
		l(LogEntry{
			Request:  r,
			Status:   d.Status(),
			Written:  d.Written(),
			Duration: time.Since(start),
		})
	})
}
