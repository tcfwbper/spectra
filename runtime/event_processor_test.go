package runtime

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
)

// =============================================================================
// Test Helpers — EventProcessor
// =============================================================================
//
// Production surface expected in runtime/event_processor.go:
//   - type EventProcessor struct { ... }
//   - func NewEventProcessor(ps *PersistentSession, wfDef EventProcessorWorkflowDef, transitionToNode *TransitionToNode, terminationNotifier chan<- struct{}) *EventProcessor
//   - func (ep *EventProcessor) ProcessEvent(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse
//   - type EventProcessorWorkflowDef interface { Nodes() []*components.Node; Transitions() []*components.Transition; ExitTransitions() []*components.ExitTransition }
//
// TransitionEvaluator (EvaluateTransition) is a package-level function already
// existing in runtime/transition_evaluator.go.
// ValidateClaudeSessionID is a package-level function already existing in
// runtime/validate_claude_session_id.go.
// =============================================================================

// --- Mock: EventProcessor WorkflowDefinition interface ---

type mockEventProcessorWorkflowDef struct {
	nodes           []*components.Node
	transitions     []*components.Transition
	exitTransitions []*components.ExitTransition
}

func (m *mockEventProcessorWorkflowDef) Nodes() []*components.Node {
	return m.nodes
}

func (m *mockEventProcessorWorkflowDef) Transitions() []*components.Transition {
	return m.transitions
}

func (m *mockEventProcessorWorkflowDef) ExitTransitions() []*components.ExitTransition {
	return m.exitTransitions
}

// --- Mock: TransitionToNode for EventProcessor testing ---

type mockTransitionToNodeForEvent struct {
	mu              sync.Mutex
	executeCalled   int
	executeTarget   string
	executeMessage  string
	executeErr      error
	executeCallback func(target, message string)
}

func (m *mockTransitionToNodeForEvent) Execute(targetNodeName, message string) error {
	m.mu.Lock()
	m.executeCalled++
	m.executeTarget = targetNodeName
	m.executeMessage = message
	m.mu.Unlock()
	if m.executeCallback != nil {
		m.executeCallback(targetNodeName, message)
	}
	return m.executeErr
}

// --- Fixture Builder: EventProcessor ---

type eventProcessorFixture struct {
	session             *mockSession
	ps                  *PersistentSession
	wfDef               *mockEventProcessorWorkflowDef
	transitionToNode    *mockTransitionToNodeForEvent
	terminationNotifier chan struct{}
}

func newEventProcessorFixture(t *testing.T) *eventProcessorFixture {
	t.Helper()
	sess := newDefaultMockSession()
	sess.getStatusResult = "running"
	sess.getCurrentStateResult = "NodeA"
	sess.getSessionDataResultVal = "cs-789"
	sess.getSessionDataResultOK = true

	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	wfDef := &mockEventProcessorWorkflowDef{
		nodes:           []*components.Node{mustNewNode(t, "NodeA", "agent", "Coder")},
		transitions:     []*components.Transition{mustNewTransition(t, "NodeA", "MsgSent", "NodeB")},
		exitTransitions: []*components.ExitTransition{},
	}

	ttn := &mockTransitionToNodeForEvent{}

	return &eventProcessorFixture{
		session:             sess,
		ps:                  ps,
		wfDef:               wfDef,
		transitionToNode:    ttn,
		terminationNotifier: newTerminationChannel(),
	}
}

// mustNewEventRuntimeMessage creates a RuntimeMessage of type "event" with the given payload.
func mustNewEventRuntimeMessage(t *testing.T, claudeSessionID string, payload json.RawMessage) *entities.RuntimeMessage {
	t.Helper()
	msg, err := entities.NewRuntimeMessage("event", payload, claudeSessionID)
	require.NoError(t, err, "mustNewEventRuntimeMessage")
	return msg
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewEventProcessor_ValidDeps(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor constructor")

	// Setup
	f := newEventProcessorFixture(t)
	_ = f

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)

	// Assert: Returns non-nil *EventProcessor; no panic
	// require.NotNil(t, ep)
}

// =============================================================================
// Happy Path — ProcessEvent
// =============================================================================

func TestEventProcessor_ProcessEvent_RegularTransition(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	// Add NodeB to workflow so transition target is valid
	f.wfDef.nodes = append(f.wfDef.nodes, mustNewNode(t, "NodeB", "agent", "Reviewer"))

	payload := json.RawMessage(`{"eventType":"MsgSent","message":"hello","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "success", resp.Status())
	// assert.Contains(t, resp.Message(), "event 'MsgSent' processed successfully")
	// assert.Contains(t, resp.Message(), "session="+testSessionID)
	// assert.Contains(t, resp.Message(), "currentState=NodeB")
	// assert.Contains(t, resp.Message(), "sessionStatus=running")
}

func TestEventProcessor_ProcessEvent_ExitTransition(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	f.wfDef.transitions = []*components.Transition{mustNewTransition(t, "NodeA", "TaskDone", "ExitNode")}
	f.wfDef.exitTransitions = []*components.ExitTransition{mustNewExitTransition(t, "NodeA", "TaskDone", "ExitNode")}
	f.wfDef.nodes = append(f.wfDef.nodes, mustNewNode(t, "ExitNode", "human", ""))

	payload := json.RawMessage(`{"eventType":"TaskDone","message":"done","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "success", resp.Status())
	// assert.Contains(t, resp.Message(), "event 'TaskDone' processed successfully")
	// assert.Contains(t, resp.Message(), "sessionStatus=completed")
	// assert.Equal(t, 1, f.session.doneCalled)
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestEventProcessor_ProcessEvent_CallsValidateClaudeSessionID(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"eventType":"MsgSent","message":"hi","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert: ValidateClaudeSessionID succeeds because session data matches
	// assert.Equal(t, "success", resp.Status())
}

func TestEventProcessor_ProcessEvent_EventEntityConstruction(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"eventType":"DataReady","message":"msg1","payload":{"key":"val"}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert: UpdateEventHistorySafe called with Event having correct fields
	// require.Equal(t, 1, f.session.updateEventHistoryCalled)
	// ev := f.session.updateEventHistoryInput
	// assert.NotEmpty(t, ev.ID())
	// assert.Equal(t, "DataReady", ev.Type())
	// assert.Equal(t, "msg1", ev.Message())
	// assert.JSONEq(t, `{"key":"val"}`, string(ev.Payload()))
	// assert.Equal(t, "NodeA", ev.EmittedBy())
	// assert.Greater(t, ev.EmittedAt(), int64(0))
	// assert.Equal(t, testSessionID, ev.SessionID())
}

func TestEventProcessor_ProcessEvent_TransitionEvaluatorCalledCorrectly(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent; TransitionEvaluator is a package-level function; validation via transition result")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeX"
	f.wfDef.nodes = []*components.Node{mustNewNode(t, "NodeX", "agent", "Coder")}
	f.wfDef.transitions = []*components.Transition{mustNewTransition(t, "NodeX", "EvType", "NodeY")}
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"eventType":"EvType","message":"m","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert: TransitionToNode.Execute was called with the correct target from the transition
	// assert.Equal(t, 1, f.transitionToNode.executeCalled)
	// assert.Equal(t, "NodeY", f.transitionToNode.executeTarget)
}

func TestEventProcessor_ProcessEvent_TransitionToNodeCalledCorrectly(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true
	f.wfDef.transitions = []*components.Transition{mustNewTransition(t, "NodeA", "Go", "NodeB")}

	payload := json.RawMessage(`{"eventType":"Go","message":"forward","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, 1, f.transitionToNode.executeCalled)
	// assert.Equal(t, "NodeB", f.transitionToNode.executeTarget)
	// assert.Equal(t, "forward", f.transitionToNode.executeMessage)
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestEventProcessor_ProcessEvent_SessionNotRunning_Initializing(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "initializing"

	payload := json.RawMessage(`{"eventType":"Ev","message":"x","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "session not ready: status is 'initializing'", resp.Message())
}

func TestEventProcessor_ProcessEvent_SessionNotRunning_Completed(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "completed"

	payload := json.RawMessage(`{"eventType":"Ev","message":"x","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "session not ready: status is 'completed'", resp.Message())
}

func TestEventProcessor_ProcessEvent_SessionNotRunning_Failed(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "failed"

	payload := json.RawMessage(`{"eventType":"Ev","message":"x","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "session not ready: status is 'failed'", resp.Message())
}

func TestEventProcessor_ProcessEvent_ClaudeSessionIDValidationFails(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	// Make validation fail: key not found
	f.session.getSessionDataResultVal = nil
	f.session.getSessionDataResultOK = false

	payload := json.RawMessage(`{"eventType":"Ev","message":"x","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Contains(t, resp.Message(), "claude session ID not found")
	// assert.Equal(t, 0, f.session.updateEventHistoryCalled)
}

func TestEventProcessor_ProcessEvent_InvalidPayloadMissingEventType(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"hi"}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "invalid event payload: missing eventType", resp.Message())
}

func TestEventProcessor_ProcessEvent_EventRecordingFails(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true
	f.session.updateEventHistoryErr = errors.New("validation error")

	payload := json.RawMessage(`{"eventType":"MsgSent","message":"hi","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "failed to record event: validation error", resp.Message())
	// Verify Fail() called with RuntimeError
	// require.Equal(t, 1, f.session.failCalled)
	// rtErr, ok := f.session.failInputErr.(*entities.RuntimeError)
	// require.True(t, ok)
	// assert.Equal(t, "EventProcessor", rtErr.Issuer())
	// assert.Equal(t, "failed to record event", rtErr.Message())
	// assert.Equal(t, "NodeA", rtErr.FailingState())
}

func TestEventProcessor_ProcessEvent_NoMatchingTransition(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true
	// No transition matching "UnknownEv" from "NodeA"
	f.wfDef.transitions = []*components.Transition{mustNewTransition(t, "NodeA", "Other", "NodeB")}

	payload := json.RawMessage(`{"eventType":"UnknownEv","message":"x","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "no transition found for event 'UnknownEv' from node 'NodeA'", resp.Message())
	// assert.Equal(t, 0, f.session.failCalled)
}

func TestEventProcessor_ProcessEvent_TransitionToNodeFails(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true
	f.transitionToNode.executeErr = errors.New("agent invoke failed")

	payload := json.RawMessage(`{"eventType":"MsgSent","message":"hi","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "transition failed: agent invoke failed", resp.Message())
	// require.Equal(t, 1, f.session.failCalled)
	// rtErr, ok := f.session.failInputErr.(*entities.RuntimeError)
	// require.True(t, ok)
	// assert.Equal(t, "EventProcessor", rtErr.Issuer())
	// assert.Equal(t, "transition failed", rtErr.Message())
	// assert.Equal(t, "NodeA", rtErr.FailingState())
}

func TestEventProcessor_ProcessEvent_DoneFailsAfterExitTransition(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true
	f.session.doneErr = errors.New("already completed")

	f.wfDef.transitions = []*components.Transition{mustNewTransition(t, "NodeA", "TaskDone", "ExitNode")}
	f.wfDef.exitTransitions = []*components.ExitTransition{mustNewExitTransition(t, "NodeA", "TaskDone", "ExitNode")}
	f.wfDef.nodes = append(f.wfDef.nodes, mustNewNode(t, "ExitNode", "human", ""))

	payload := json.RawMessage(`{"eventType":"TaskDone","message":"done","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "failed to complete session: already completed", resp.Message())
	// require.Equal(t, 1, f.session.failCalled)
	// rtErr, ok := f.session.failInputErr.(*entities.RuntimeError)
	// require.True(t, ok)
	// assert.Equal(t, "EventProcessor", rtErr.Issuer())
	// assert.Equal(t, "failed to complete session", rtErr.Message())
	// assert.Equal(t, "ExitNode", rtErr.FailingState())
}

func TestEventProcessor_ProcessEvent_NodeNotFound(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "Ghost"
	// wfDef.nodes doesn't contain "Ghost"

	payload := json.RawMessage(`{"eventType":"Ev","message":"x","payload":{}}`)
	msg := mustNewEventRuntimeMessage(t, "cs-789", payload)

	// Act
	// ep := NewEventProcessor(f.ps, f.wfDef, f.transitionToNode, f.terminationNotifier)
	// resp := ep.ProcessEvent(testSessionID, msg)
	_ = msg

	// Assert
	// assert.Equal(t, "error", resp.Status())
	// assert.Equal(t, "current node 'Ghost' not found in workflow definition", resp.Message())
}

// =============================================================================
// Concurrent Behaviour
// =============================================================================

func TestEventProcessor_ProcessEvent_ConcurrentEventsSerialize(t *testing.T) {
	t.Skip("scaffolded: awaiting runtime/event_processor.go — NewEventProcessor, EventProcessor.ProcessEvent")

	// Setup
	f := newEventProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-789"
	f.session.getSessionDataResultOK = true

	payload1 := json.RawMessage(`{"eventType":"MsgSent","message":"a","payload":{}}`)
	payload2 := json.RawMessage(`{"eventType":"MsgSent","message":"b","payload":{}}`)
	msg1 := mustNewEventRuntimeMessage(t, "cs-789", payload1)
	msg2 := mustNewEventRuntimeMessage(t, "cs-789", payload2)
	_ = msg1
	_ = msg2

	// Act: Call ep.ProcessEvent concurrently from two goroutines
	// var wg sync.WaitGroup
	// var resp1, resp2 *entities.RuntimeResponse
	// wg.Add(2)
	// go func() { defer wg.Done(); resp1 = ep.ProcessEvent(testSessionID, msg1) }()
	// go func() { defer wg.Done(); resp2 = ep.ProcessEvent(testSessionID, msg2) }()
	// wg.Wait()

	// Assert: Both calls complete without data race; each returns a RuntimeResponse
	// assert.NotNil(t, resp1)
	// assert.NotNil(t, resp2)
}

// =============================================================================
// Compile guards — suppress unused import warnings
// =============================================================================

var (
	_ = json.RawMessage{}
	_ = errors.New
	_ = (*sync.Mutex)(nil)
	_ = (*entities.RuntimeMessage)(nil)
	_ = (*entities.RuntimeResponse)(nil)
	_ = (*components.Node)(nil)
	_ = (*components.Transition)(nil)
	_ = (*components.ExitTransition)(nil)
	_ = assert.Equal
	_ = require.NoError
)
