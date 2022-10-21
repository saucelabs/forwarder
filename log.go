package forwarder

// Logger is the logger used by the forwarder package.
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// nopLogger is a logger that does nothing.
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
	l.Debugf(format, v...)
}
