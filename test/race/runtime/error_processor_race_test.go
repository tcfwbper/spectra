package runtime_race

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/runtime"
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
	// Setup:
	// - Mock PersistentSession: GetStatusSafe() returns "running", GetCurrentStateSafe() returns "NodeA".
	// - Mock WorkflowDefinition with matching node.
	// - Stub ValidateClaudeSessionID returns nil (via session data matching).
	// - Mock Fail() succeeds on first call and returns "session already failed" on second.
	// - Create two valid RuntimeMessages with different error payloads.
	sess := &raceSafeMockSession{
		statusResult:       "running",
		currentStateResult: "NodeA",
		sessionDataVal:     "cs-123",
		sessionDataOK:      true,
	}

	ps := runtime.NewPersistentSession(
		sess,
		&raceSafeMetadataStore{},
		&raceSafeEventStore{},
		logger.NewNopLogger(),
	)

	wfDef := &raceErrorProcessorWfDef{
		nodes: []*components.Node{mustNewNode(t, "NodeA", "agent", "Coder")},
	}
	terminationNotifier := newTerminationChannel()

	ep := runtime.NewErrorProcessor(ps, wfDef, terminationNotifier)

	payload1 := json.RawMessage(`{"message":"error one"}`)
	payload2 := json.RawMessage(`{"message":"error two"}`)
	msg1 := mustNewErrorRuntimeMessage(t, "cs-123", payload1)
	msg2 := mustNewErrorRuntimeMessage(t, "cs-123", payload2)

	// Act:
	// - Launch two goroutines: both call ep.ProcessError("sess-uuid", msg) concurrently.
	var wg sync.WaitGroup
	var resp1, resp2 *entities.RuntimeResponse
	wg.Add(2)
	go func() { defer wg.Done(); resp1 = ep.ProcessError(testSessionID, msg1) }()
	go func() { defer wg.Done(); resp2 = ep.ProcessError(testSessionID, msg2) }()
	wg.Wait()

	// Assert:
	// - One returns success response; the other returns error response.
	// - No data race detected by race detector (-race flag).
	assert.NotNil(t, resp1)
	assert.NotNil(t, resp2)

	// One should succeed (first to acquire lock) and one should fail (already failed).
	statuses := []string{resp1.Status(), resp2.Status()}
	assert.Contains(t, statuses, "success")
	assert.Contains(t, statuses, "error")
}
