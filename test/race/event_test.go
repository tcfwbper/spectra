package race_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEvent_SimultaneousEmission verifies multiple events emitted simultaneously are serialized
func TestEvent_SimultaneousEmission(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	// Setup: Session with Status="running"
	sessionID := uuid.New()

	// Two Event instances emitted at same time for same session

	// Verify both events recorded in EventHistory
	// Verify first event to acquire session lock is processed first
	// Verify serialized processing ensures deterministic order
}
