package forwarder

import "log"

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

// stdLogger is a logger that uses the standard log package.
type stdLogger struct {
	name string
}

func namedStdLogger(name string) stdLogger {
	if name != "" {
		name = "[" + name + "] "
	}
	return stdLogger{
		name: name,
	}
}

func (s stdLogger) Errorf(format string, args ...interface{}) {
	log.Printf(s.name+"ERROR: "+format, args...)
}

func (s stdLogger) Infof(format string, args ...interface{}) {
	log.Printf(s.name+"INFO: "+format, args...)
}

func (s stdLogger) Debugf(format string, args ...interface{}) {
	log.Printf(s.name+"DEBUG: "+format, args...)
}

// goproxyLogger is a logger that implements the goproxy.Logger interface.
type goproxyLogger struct {
	Logger
}

func (l goproxyLogger) Printf(format string, v ...interface{}) {
	l.Debugf(format, v...)
}
