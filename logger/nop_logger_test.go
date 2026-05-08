package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Type Hierarchy (compile-time check) ---

// TestNopLogger_ImplementsLogger verifies NopLogger satisfies Logger at compile time.
func TestNopLogger_ImplementsLogger(t *testing.T) {
	var _ Logger = NewNopLogger()
}

// --- Happy Path — Construction ---

func TestNewNopLogger_ReturnsLogger(t *testing.T) {
	logger := NewNopLogger()
	assert.NotNil(t, logger, "NewNopLogger() must return a non-nil Logger")
}

// --- Happy Path — Debug ---

func TestNopLogger_Debug_DoesNotPanic(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Debug("msg", "key", "value")
	})
}

// --- Happy Path — Info ---

func TestNopLogger_Info_DoesNotPanic(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Info("msg", "key", "value")
	})
}

// --- Happy Path — Warn ---

func TestNopLogger_Warn_DoesNotPanic(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Warn("msg", "key", "value")
	})
}

// --- Happy Path — Error ---

func TestNopLogger_Error_DoesNotPanic(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Error("msg", "key", "value")
	})
}

// --- Null / Empty Input ---

func TestNopLogger_EmptyMessage(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Debug("")
		logger.Info("")
		logger.Warn("")
		logger.Error("")
	})
}

func TestNopLogger_NilArgs(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Debug("msg")
		logger.Info("msg")
		logger.Warn("msg")
		logger.Error("msg")
	})
}

func TestNopLogger_OddArgs(t *testing.T) {
	logger := NewNopLogger()
	assert.NotPanics(t, func() {
		logger.Debug("msg", "key")
		logger.Info("msg", "key")
		logger.Warn("msg", "key")
		logger.Error("msg", "key")
	})
}
