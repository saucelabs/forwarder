// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package run

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/formatter"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
	"github.com/saucelabs/sypl/output"
)

const (
	defaultTimeFormat = "2006-01-02T15-04-05"
	infoLevel         = "info"
)

// logConfig is a configuration for the logger.
type logConfig struct {
	FileLevel string `validate:"required,gte=3"`
	FilePath  string `validate:"required"`
	Level     string `validate:"required,gte=3"`
}

func defaultLogConfig() logConfig {
	return logConfig{
		FileLevel: infoLevel,
		FilePath:  filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.log", "forwarder", time.Now().Format(defaultTimeFormat))),
		Level:     infoLevel,
	}
}

func newLogger(cfg logConfig, name string) forwarderLogger {
	l := sypl.NewDefault(name, level.MustFromString(cfg.Level))
	l.AddOutputs(
		output.File(cfg.FilePath, level.MustFromString(cfg.FileLevel)).SetFormatter(formatter.Text()),
	)
	l.PrintlnWithOptions(&options.Options{
		Fields: fields.Fields{
			"fileLevel": cfg.FileLevel,
			"filePath":  cfg.FilePath,
			"level":     cfg.Level,
		},
	}, level.Trace, "Logging is configured")
	return forwarderLogger{l}
}

// forwarderLogger is a wrapper around sypl.Logger that implements forwarder.Logger.
// It is needed because sypl.Logger methods return sypl.ISypl, which makes it impossible work with anything else...
type forwarderLogger struct {
	*sypl.Sypl
}

func (l forwarderLogger) Errorf(format string, args ...interface{}) {
	l.Sypl.Errorf(format+"\n", args...)
}

func (l forwarderLogger) Infof(format string, args ...interface{}) {
	l.Sypl.Infof(format+"\n", args...)
}

func (l forwarderLogger) Debugf(format string, args ...interface{}) {
	l.Sypl.Debugf(format+"\n", args...)
}
