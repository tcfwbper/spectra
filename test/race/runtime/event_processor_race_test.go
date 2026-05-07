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
	// Setup:
	// - Mock PersistentSession: GetStatusSafe() returns "running", GetCurrentStateSafe() returns "NodeA".
	// - Mock WorkflowDefinition with matching node and transitions.
	// - Stub ValidateClaudeSessionID returns nil (via session data matching).
	// - Mock TransitionToNode returns nil.
	// - Create two valid RuntimeMessages with different event types.
	sess := &raceSafeMockSession{
		statusResult:       "running",
		currentStateResult: "NodeA",
		sessionDataVal:     "cs-789",
		sessionDataOK:      true,
	}

	ps := runtime.NewPersistentSession(
		sess,
		&raceSafeMetadataStore{},
		&raceSafeEventStore{},
		logger.NewNopLogger(),
	)

	wfDef := &raceEventProcessorWfDef{
		nodes: []*components.Node{
			mustNewNode(t, "NodeA", "agent", "Coder"),
			mustNewNode(t, "NodeB", "human", ""),
		},
		transitions: []*components.Transition{
			mustNewTransition(t, "NodeA", "MsgSent", "NodeB"),
		},
		exitTransitions: []*components.ExitTransition{},
	}

	ttn := &raceSafeTransitionToNode{sess: sess}
	terminationNotifier := newTerminationChannel()

	ep := runtime.NewEventProcessor(ps, wfDef, ttn, terminationNotifier)

	payload1 := json.RawMessage(`{"eventType":"MsgSent","message":"a","payload":{}}`)
	payload2 := json.RawMessage(`{"eventType":"MsgSent","message":"b","payload":{}}`)
	msg1 := mustNewEventRuntimeMessage(t, "cs-789", payload1)
	msg2 := mustNewEventRuntimeMessage(t, "cs-789", payload2)

	// Act:
	// - Launch two goroutines: both call ep.ProcessEvent("sess-uuid", msg) concurrently.
	var wg sync.WaitGroup
	var resp1, resp2 *entities.RuntimeResponse
	wg.Add(2)
	go func() { defer wg.Done(); resp1 = ep.ProcessEvent(testSessionID, msg1) }()
	go func() { defer wg.Done(); resp2 = ep.ProcessEvent(testSessionID, msg2) }()
	wg.Wait()

	// Assert:
	// - Both calls complete without data race.
	// - Each returns a RuntimeResponse.
	assert.NotNil(t, resp1)
	assert.NotNil(t, resp2)
}
