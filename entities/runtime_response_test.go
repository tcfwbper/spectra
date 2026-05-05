package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — Construction ---

func TestSuccessResponse_WithMessage(t *testing.T) {
	resp := SuccessResponse("operation completed")

	assert.Equal(t, "success", resp.Status())
	assert.Equal(t, "operation completed", resp.Message())
}

func TestSuccessResponse_EmptyMessage(t *testing.T) {
	resp := SuccessResponse("")

	assert.Equal(t, "success", resp.Status())
	assert.Equal(t, "", resp.Message())
}

func TestErrorResponse_WithMessage(t *testing.T) {
	resp := ErrorResponse("something failed")

	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "something failed", resp.Message())
}

func TestErrorResponse_EmptyMessage(t *testing.T) {
	resp := ErrorResponse("")

	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "", resp.Message())
}

func TestSuccessResponse_MessageWithNewlines(t *testing.T) {
	resp := SuccessResponse("line1\nline2\nline3")

	assert.Equal(t, "success", resp.Status())
	assert.Equal(t, "line1\nline2\nline3", resp.Message())
}

func TestErrorResponse_MessageWithNewlines(t *testing.T) {
	resp := ErrorResponse("error\ndetails")

	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "error\ndetails", resp.Message())
}

// --- Immutability ---

// TestRuntimeResponse_StatusImmutable verifies that the status field is
// unexported and cannot be modified after construction. Since fields are
// unexported, we verify the getter consistently returns the construction value.
func TestRuntimeResponse_StatusImmutable(t *testing.T) {
	resp := SuccessResponse("msg")

	// Getter must consistently return the constructed value
	assert.Equal(t, "success", resp.Status())
	assert.Equal(t, "success", resp.Status())
}

// TestRuntimeResponse_MessageImmutable verifies that the message field is
// unexported and cannot be modified after construction. Since fields are
// unexported, we verify the getter consistently returns the construction value.
func TestRuntimeResponse_MessageImmutable(t *testing.T) {
	resp := ErrorResponse("original")

	// Getter must consistently return the constructed value
	assert.Equal(t, "original", resp.Message())
	assert.Equal(t, "original", resp.Message())
}
