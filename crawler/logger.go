package crawler

import "log"

// Logger is the interface of Crawler loggers.
type Logger interface {
	Debug(format string, args ...interface{})
	Warn(format string, args ...interface{})
}

// BasicLogger is a Logger that logs to stdout.
type BasicLogger struct{}

func (l *BasicLogger) log(prefix, format string, args ...interface{}) {
	log.Printf(prefix+" "+format, args...)
}

// Debug logs a debug message to stdout.
func (l *BasicLogger) Debug(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}

// Warn logs a warning to stdout.
func (l *BasicLogger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

// NoopLogger is a Logger that does not output anything.
type NoopLogger struct{}

// Debug does nothing.
func (l *NoopLogger) Debug(format string, args ...interface{}) {}

// Warn does nothing.
func (l *NoopLogger) Warn(format string, args ...interface{}) {}
