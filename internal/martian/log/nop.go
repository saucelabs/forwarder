// Copyright 2023 Sauce Labs Inc., all rights reserved.

package log

import (
	"context"
)

type nopLogger struct{}

var _ Logger = nopLogger{}

func (nopLogger) Infof(_ context.Context, _ string, _ ...any) {}

func (nopLogger) Debugf(_ context.Context, _ string, _ ...any) {}

func (nopLogger) Errorf(_ context.Context, _ string, _ ...any) {}
