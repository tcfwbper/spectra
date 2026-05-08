package logger

// NopLogger is a concrete implementation of the Logger interface that discards
// all log output. It is intended for use in tests and situations where logging
// is not desired. All methods are no-ops that return immediately.
type NopLogger struct{}

// NewNopLogger returns a NopLogger instance satisfying the Logger interface.
func NewNopLogger() Logger {
	return &NopLogger{}
}

func (*NopLogger) Debug(_ string, _ ...any) {}
func (*NopLogger) Info(_ string, _ ...any)  {}
func (*NopLogger) Warn(_ string, _ ...any)  {}
func (*NopLogger) Error(_ string, _ ...any) {}
