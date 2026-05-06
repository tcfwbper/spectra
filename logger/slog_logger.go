package logger

import "log/slog"

// SlogLogger is a concrete implementation of the Logger interface that delegates
// all logging to Go's standard library log/slog. It wraps a *slog.Logger instance
// and maps each Logger method to the corresponding slog level.
type SlogLogger struct {
	inner *slog.Logger
}

// NewSlogLogger returns a SlogLogger instance satisfying the Logger interface.
// If slogger is nil, slog.Default() is used as fallback.
func NewSlogLogger(slogger *slog.Logger) Logger {
	if slogger == nil {
		slogger = slog.Default()
	}
	return &SlogLogger{inner: slogger}
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.inner.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
}
