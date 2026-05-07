package runtime_race

import (
	"testing"
)

// =============================================================================
// Concurrent Behaviour — EventProcessor
// =============================================================================
//
// Production surface required:
//   - runtime.NewEventProcessor(ps, wfDef, transitionToNode, terminationNotifier)
//   - runtime.EventProcessor.ProcessEvent(sessionUUID, msg) RuntimeResponse
//   - runtime.EventProcessorWorkflowDef interface
//
// This test verifies that concurrent ProcessEvent calls are serialized by
// PersistentSession's internal lock without data races.
// =============================================================================

func TestEventProcessor_ProcessEvent_ConcurrentEventsSerialize(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent; also requires exported interfaces or test-accessible constructors for race testing from external package")

	// Setup:
	// - Mock PersistentSession: GetStatusSafe() returns "running", GetCurrentStateSafe() returns "NodeA".
	// - Mock WorkflowDefinition with matching node and transitions.
	// - Stub ValidateClaudeSessionID returns nil.
	// - Mock TransitionToNode returns nil.
	// - Create two valid RuntimeMessages with different event types.
	//
	// Act:
	// - Launch two goroutines: both call ep.ProcessEvent("sess-uuid", msg) concurrently.
	//
	// Assert:
	// - Both calls complete without data race.
	// - Each returns a RuntimeResponse.
}
