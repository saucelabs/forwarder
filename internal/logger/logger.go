// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
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
	name              = "proxy"
)

var proxyLogger *sypl.Sypl

// Options for logging .
type Options struct {
	// Allows to set the internal Logger. For now, it will be straightforward
	// Sypl logger. In the future, could be the `sypl.IBasicPrinter` interface.
	Logger *sypl.Sypl `json:"-"`

	FileLevel string `validate:"required,gte=3"`
	FilePath  string `validate:"required"`
	Level     string `validate:"required,gte=3"`
}

// Default sets `Options` default values.
func (o *Options) Default() {
	if o.FileLevel == "" {
		o.FileLevel = infoLevel
	}

	if o.FilePath == "" {
		o.FilePath = fmt.Sprintf("%s-%s.log",
			filepath.Join(os.TempDir(), name),
			time.Now().Format(defaultTimeFormat),
		)
	}

	if o.Level == "" {
		o.Level = infoLevel
	}
}

// Get returns the logger. If the logger isn't configured, it will exit with fatal.
func Get() *sypl.Sypl {
	if proxyLogger == nil {
		log.Fatalln("Logger is not configired")
	}

	return proxyLogger
}

// Setup logger. If it fails to set up, it will exit with fatal.
func Setup(o *Options) *sypl.Sypl {
	// Do nothing, if the logger is already set up.
	if proxyLogger != nil {
		return proxyLogger
	}

	// Should allow to specify a logger.
	if o.Logger != nil {
		proxyLogger = o.Logger
	} else {
		if o == nil {
			o = &Options{}
		}

		o.Default()

		if err := validator.New().Struct(o); err != nil {
			log.Fatalln(err)
		}

		proxyLogger = sypl.NewDefault(name, level.MustFromString(o.Level))
		proxyLogger.AddOutputs(
			output.File(o.FilePath, level.MustFromString(o.FileLevel)).SetFormatter(formatter.Text()),
		)
	}

	proxyLogger.PrintlnWithOptions(&options.Options{
		Fields: fields.Fields{
			"fileLevel": o.FileLevel,
			"filePath":  o.FilePath,
			"level":     o.Level,
		},
	}, level.Trace, "Logging is configured")

	return proxyLogger
}

// ProxyLogger exist to satisfy `goproxy` logging interface.
type ProxyLogger struct {
	Logger *sypl.Sypl
}

// Printf satisfies `goproxy` logging interface. Default logging level will be
// `Debug`.
func (pL *ProxyLogger) Printf(format string, v ...interface{}) {
	format = strings.TrimSpace(format)
	pL.Logger.Debuglnf(format, v...)
}

type StandardLogger struct {
	Logger *sypl.Sypl
}

// Write satisfies the standard logging interface. Default logging level will be
// `Debug`.
func (pL *StandardLogger) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	pL.Logger.Debug(string(p))
	return len(p), nil
}

// RedirectStandardLogs redirects logs created with the standard library global logger
// to the ProxyLogger.
func RedirectStandardLogs() {
	log.SetFlags(0)
	log.SetPrefix("")
	standardLogger := &StandardLogger{
		Logger: Get().New("go"),
	}
	log.SetOutput(standardLogger)
}

// DisableStandardLogs disables logs created with the standard library global logger
func DisableStandardLogs() {
	log.SetOutput(ioutil.Discard)
}
