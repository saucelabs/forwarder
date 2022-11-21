package stdlog

import (
	"io"
	"log"
	"os"

	flog "github.com/saucelabs/forwarder/log"
)

func Default() StdLogger {
	return StdLogger{
		log:     log.Default(),
		verbose: true,
	}
}

func New(cfg *flog.Config) StdLogger {
	var w io.Writer = os.Stdout
	if cfg.File != nil {
		w = cfg.File
	}
	return StdLogger{
		log:     log.New(w, "", log.Ldate|log.Ltime|log.LUTC),
		verbose: cfg.Verbose,
	}
}

// StdLogger implements the forwarder.Logger interface using the standard log package.
type StdLogger struct {
	log     *log.Logger
	name    string
	verbose bool

	// Decorate allows to modify the log message before it is written.
	Decorate func(string) string
}

func (sl StdLogger) Named(name string) StdLogger {
	if name != "" {
		name = "[" + name + "] "
	}
	sl.name = name
	return sl
}

func (sl StdLogger) Errorf(format string, args ...interface{}) {
	if sl.Decorate != nil {
		format = sl.Decorate(format)
	}
	sl.log.Printf(sl.name+"ERROR: "+format, args...)
}

func (sl StdLogger) Infof(format string, args ...interface{}) {
	if sl.Decorate != nil {
		format = sl.Decorate(format)
	}
	sl.log.Printf(sl.name+"INFO: "+format, args...)
}

func (sl StdLogger) Debugf(format string, args ...interface{}) {
	if !sl.verbose {
		return
	}
	if sl.Decorate != nil {
		format = sl.Decorate(format)
	}
	sl.log.Printf(sl.name+"DEBUG: "+format, args...)
}
