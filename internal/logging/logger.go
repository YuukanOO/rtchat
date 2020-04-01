package logging

import (
	"log"
	"os"
)

type (
	// Logger exposes a tiny interface to log stuff in the application.
	Logger interface {
		// Debug prints an info message but only in dev mode.
		Debug(string, ...interface{})
		// Info prints an info message.
		Info(string, ...interface{})
		// Error prints an error message.
		Error(string, ...interface{})
	}

	defaultLogger struct {
		debug bool
		info  *log.Logger
		err   *log.Logger
	}
)

// New instantiates a new basic logger using the log standard package.
func New(debug bool) Logger {
	return &defaultLogger{
		debug: debug,
		info:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		err:   log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime),
	}
}

func (l *defaultLogger) Debug(format string, v ...interface{}) {
	if l.debug {
		l.Info(format, v...)
	}
}

func (l *defaultLogger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

func (l *defaultLogger) Error(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}
