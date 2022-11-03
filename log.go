package forwarder

import "strings"

// Logger is the logger used by the forwarder package.
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// NopLogger is a logger that does nothing.
var NopLogger = nopLogger{} //nolint:gochecknoglobals // nop implementation

type nopLogger struct{}

func (l nopLogger) Errorf(format string, args ...interface{}) {
}

func (l nopLogger) Infof(format string, args ...interface{}) {
}

func (l nopLogger) Debugf(format string, args ...interface{}) {
}

// goproxyLogger is a logger that implements the goproxy.Logger interface.
type goproxyLogger struct {
	Logger
}

func (l goproxyLogger) Printf(format string, v ...interface{}) {
	if strings.HasPrefix(format, "[%03d] WARN: ") {
		l.Logger.Infof(format[13:], v...)
		return
	}

	l.Logger.Debugf(strings.Replace(format, "INFO: ", "", 1), v...)
}
