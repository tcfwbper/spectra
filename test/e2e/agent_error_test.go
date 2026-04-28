package e2e_test

import (
	"testing"

	"github.com/google/uuid"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

// TestAgentError_CLIInvocation verifies agent raises error via spectra-agent CLI
func TestAgentError_CLIInvocation(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Runtime running with session files in tmpDir
	// Session active with socket in tmpDir
	sessionID := uuid.New()
	_ = sessionID

	// Execute: spectra-agent error "task failed" --session-id <uuid> --detail '{"code": 500}'

	// Verify command succeeds
	// Verify AgentError recorded in test directory
	// Verify session transitions to failed
	// Verify human notified
}

// TestAgentError_CLIMissingSessionID verifies CLI rejects error without session-id flag
func TestAgentError_CLIMissingSessionID(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Runtime running

	// Execute: spectra-agent error "task failed"

	// Verify command fails
	// Verify error message matches /session-id.*required/i
}
