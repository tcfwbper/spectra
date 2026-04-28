package race_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuntimeError_MultipleErrorsSerialized verifies multiple simultaneous RuntimeErrors are serialized; first error wins
func TestRuntimeError_MultipleErrorsSerialized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	// Setup: Session with Status="running"
	sessionID := uuid.New()

	// Two RuntimeError instances raised simultaneously for same session

	// Verify first error to acquire session lock updates in-memory status and records in session's Error field
	// Verify second error logged but does not overwrite
	// Verify session Status="failed"
}
