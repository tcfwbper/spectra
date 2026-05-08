package logger

// Logger is a structured logging interface that provides four standard log levels.
// It follows the log/slog key-value pair convention for structured fields.
// All methods are safe for concurrent use from multiple goroutines.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}
