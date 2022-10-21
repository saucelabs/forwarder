package log

import (
	"os"
)

// Config is a configuration for the loggers.
type Config struct {
	File    *os.File
	Verbose bool
}

func DefaultConfig() *Config {
	return &Config{
		File: nil,
	}
}
