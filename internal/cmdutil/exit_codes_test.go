package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — Constants ---

func TestExitSuccess_Value(t *testing.T) {
	assert.Equal(t, 0, ExitSuccess)
}

func TestExitInvocationError_Value(t *testing.T) {
	assert.Equal(t, 1, ExitInvocationError)
}

func TestExitTransportError_Value(t *testing.T) {
	assert.Equal(t, 2, ExitTransportError)
}

func TestExitRuntimeError_Value(t *testing.T) {
	assert.Equal(t, 3, ExitRuntimeError)
}

// --- Boundary Values — No Overlap ---

func TestExitCodes_Unique(t *testing.T) {
	codes := []int{ExitSuccess, ExitInvocationError, ExitTransportError, ExitRuntimeError}
	seen := make(map[int]bool)
	for _, c := range codes {
		assert.False(t, seen[c], "duplicate exit code value: %d", c)
		seen[c] = true
	}
}
