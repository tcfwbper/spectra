package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Type Hierarchy (compile-time check) ---

// TestSlogLogger_ImplementsLogger verifies SlogLogger satisfies Logger at compile time.
func TestSlogLogger_ImplementsLogger(t *testing.T) {
	var _ Logger = NewSlogLogger(nil)
}

// --- Helper: newBufferSlogLogger ---

// newBufferSlogLogger creates a *slog.Logger writing to a buffer at the given level,
// returning both the Logger and the buffer for assertion.
func newBufferSlogLogger(t *testing.T, level slog.Level) (Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})
	slogger := slog.New(handler)
	return NewSlogLogger(slogger), &buf
}

// --- Happy Path — Construction ---

func TestNewSlogLogger_WithValidLogger(t *testing.T) {
	logger, _ := newBufferSlogLogger(t, slog.LevelDebug)
	assert.NotNil(t, logger, "NewSlogLogger with valid *slog.Logger must return non-nil Logger")
}

func TestNewSlogLogger_WithNilFallsBackToDefault(t *testing.T) {
	logger := NewSlogLogger(nil)
	require.NotNil(t, logger, "NewSlogLogger(nil) must return non-nil Logger")
	// Subsequent calls must not panic.
	assert.NotPanics(t, func() {
		logger.Info("test")
	})
}

// --- Happy Path — Debug ---

func TestSlogLogger_Debug_DelegatesToSlog(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	logger.Debug("test event", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "level=DEBUG")
	assert.Contains(t, output, "msg=\"test event\"")
	assert.Contains(t, output, "key=value")
}

// --- Happy Path — Info ---

func TestSlogLogger_Info_DelegatesToSlog(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	logger.Info("info event", "count", 42)

	output := buf.String()
	assert.Contains(t, output, "level=INFO")
	assert.Contains(t, output, "msg=\"info event\"")
	assert.Contains(t, output, "count=42")
}

// --- Happy Path — Warn ---

func TestSlogLogger_Warn_DelegatesToSlog(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	logger.Warn("warn event", "detail", "something")

	output := buf.String()
	assert.Contains(t, output, "level=WARN")
	assert.Contains(t, output, "msg=\"warn event\"")
	assert.Contains(t, output, "detail=something")
}

// --- Happy Path — Error ---

func TestSlogLogger_Error_DelegatesToSlog(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	logger.Error("error event", "err", "timeout")

	output := buf.String()
	assert.Contains(t, output, "level=ERROR")
	assert.Contains(t, output, "msg=\"error event\"")
	assert.Contains(t, output, "err=timeout")
}

// --- Null / Empty Input ---

func TestSlogLogger_EmptyMessage(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	assert.NotPanics(t, func() {
		logger.Info("")
	})
	output := buf.String()
	assert.Contains(t, output, "msg=")
}

func TestSlogLogger_NoArgs(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	assert.NotPanics(t, func() {
		logger.Info("msg")
	})
	output := buf.String()
	assert.Contains(t, output, "msg=msg")
}

func TestSlogLogger_OddArgs(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	assert.NotPanics(t, func() {
		logger.Info("msg", "orphan_key")
	})
	output := buf.String()
	// slog handles odd args gracefully with !BADKEY marker
	assert.NotEmpty(t, output)
}

// --- Mock / Dependency Interaction ---

func TestSlogLogger_PassThroughNoTransformation(t *testing.T) {
	logger, buf := newBufferSlogLogger(t, slog.LevelDebug)

	logger.Info("evt", "a", 1, "b", "two")

	output := buf.String()
	assert.Contains(t, output, "a=1")
	assert.Contains(t, output, "b=two")
}
