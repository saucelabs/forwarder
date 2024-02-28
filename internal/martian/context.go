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
	"time"
)

type contextKey int

const (
	traceIDContextKey contextKey = iota
)

func withTraceID(ctx context.Context, id traceID) context.Context {
	return context.WithValue(ctx, traceIDContextKey, id)
}

func TraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDContextKey); v != nil {
		return v.(traceID).id
	}
	return ""
}

func Duration(ctx context.Context) time.Duration {
	if v := ctx.Value(traceIDContextKey); v != nil {
		return v.(traceID).Duration()
	}
	return 0
}
