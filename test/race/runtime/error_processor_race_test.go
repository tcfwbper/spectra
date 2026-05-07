package runtime_race

import (
	"testing"
)

// =============================================================================
// Concurrent Behaviour — ErrorProcessor
// =============================================================================
//
// Production surface required:
//   - runtime.NewErrorProcessor(ps, wfDef, terminationNotifier)
//   - runtime.ErrorProcessor.ProcessError(sessionUUID, msg) RuntimeResponse
//   - runtime.ErrorProcessorWorkflowDef interface
//
// This test verifies that concurrent ProcessError calls for the same session
// serialize via PersistentSession's internal lock without data races.
// First-error-wins semantics are validated.
// =============================================================================

func TestErrorProcessor_ProcessError_ConcurrentFirstErrorWins(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/error_processor.go — NewErrorProcessor, ErrorProcessor.ProcessError; also requires exported interfaces or test-accessible constructors for race testing from external package")

	// Setup:
	// - Mock PersistentSession: GetStatusSafe() returns "running", GetCurrentStateSafe() returns "NodeA".
	// - Mock WorkflowDefinition with matching node.
	// - Stub ValidateClaudeSessionID returns nil.
	// - Mock Fail() succeeds on first call and returns "session already failed" on second.
	// - Create two valid RuntimeMessages with different error payloads.
	//
	// Act:
	// - Launch two goroutines: both call ep.ProcessError("sess-uuid", msg) concurrently.
	//
	// Assert:
	// - One returns success response; the other returns error response.
	// - No data race detected by race detector (-race flag).
}
