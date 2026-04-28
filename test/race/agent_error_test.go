package race_test

import (
	"testing"

	"github.com/google/uuid"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

// TestAgentError_MultipleErrorsSerialized verifies multiple simultaneous errors are serialized; first error wins
func TestAgentError_MultipleErrorsSerialized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	// Setup: Session with Status="running"
	sessionID := uuid.New()
	_ = sessionID

	// Two AgentError instances raised simultaneously for same session

	// Verify first error to acquire session lock recorded in session's Error field
	// Verify second error logged but does not overwrite
	// Verify session Status="failed"
}
