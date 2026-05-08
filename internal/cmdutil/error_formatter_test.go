package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — FormatError ---

func TestFormatError_BasicMessage(t *testing.T) {
	result := FormatError("file not found")
	assert.Equal(t, "Error: file not found", result)
}

func TestFormatError_MessageWithSpaces(t *testing.T) {
	result := FormatError("could not connect to server")
	assert.Equal(t, "Error: could not connect to server", result)
}

// --- Null / Empty Input — FormatError ---

func TestFormatError_EmptyString(t *testing.T) {
	result := FormatError("")
	assert.Equal(t, "Error: ", result)
}

// --- Boundary Values — FormatError ---

func TestFormatError_AlreadyPrefixed(t *testing.T) {
	result := FormatError("Error: something")
	assert.Equal(t, "Error: Error: something", result)
}

// --- Happy Path — FormatWarning ---

func TestFormatWarning_BasicMessage(t *testing.T) {
	result := FormatWarning("deprecated flag")
	assert.Equal(t, "Warning: deprecated flag", result)
}

func TestFormatWarning_MessageWithSpaces(t *testing.T) {
	result := FormatWarning("config value is missing")
	assert.Equal(t, "Warning: config value is missing", result)
}

// --- Null / Empty Input — FormatWarning ---

func TestFormatWarning_EmptyString(t *testing.T) {
	result := FormatWarning("")
	assert.Equal(t, "Warning: ", result)
}

// --- Boundary Values — FormatWarning ---

func TestFormatWarning_AlreadyPrefixed(t *testing.T) {
	result := FormatWarning("Warning: something")
	assert.Equal(t, "Warning: Warning: something", result)
}
