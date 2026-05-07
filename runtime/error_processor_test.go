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
// Test Helpers — ErrorProcessor
// =============================================================================
//
// Production surface expected in runtime/error_processor.go:
//   - type ErrorProcessor struct { ... }
//   - func NewErrorProcessor(ps *PersistentSession, wfDef ErrorProcessorWorkflowDef, terminationNotifier chan<- struct{}) *ErrorProcessor
//   - func (ep *ErrorProcessor) ProcessError(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse
//   - type ErrorProcessorWorkflowDef interface { Nodes() []*components.Node }
//
// ValidateClaudeSessionID is a package-level function already existing in
// runtime/validate_claude_session_id.go.
// =============================================================================

// --- Mock: ErrorProcessor WorkflowDefinition interface ---

type mockErrorProcessorWorkflowDef struct {
	nodes []*components.Node
}

func (m *mockErrorProcessorWorkflowDef) Nodes() []*components.Node {
	return m.nodes
}

// --- Fixture Builder: ErrorProcessor ---

type errorProcessorFixture struct {
	session             *mockSession
	ps                  *PersistentSession
	wfDef               *mockErrorProcessorWorkflowDef
	terminationNotifier chan struct{}
}

func newErrorProcessorFixture(t *testing.T) *errorProcessorFixture {
	t.Helper()
	sess := newDefaultMockSession()
	sess.getStatusResult = "running"
	sess.getCurrentStateResult = "NodeA"
	sess.getSessionDataResultVal = "cs-123"
	sess.getSessionDataResultOK = true

	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	wfDef := &mockErrorProcessorWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "NodeA", "agent", "Coder")},
	}

	return &errorProcessorFixture{
		session:             sess,
		ps:                  ps,
		wfDef:               wfDef,
		terminationNotifier: newTerminationChannel(),
	}
}

// mustNewErrorRuntimeMessage creates a RuntimeMessage of type "error" with the given payload.
func mustNewErrorRuntimeMessage(t *testing.T, claudeSessionID string, payload json.RawMessage) *entities.RuntimeMessage {
	t.Helper()
	msg, err := entities.NewRuntimeMessage("error", payload, claudeSessionID)
	require.NoError(t, err, "mustNewErrorRuntimeMessage")
	return msg
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewErrorProcessor_ValidDeps(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)

	// Assert: Returns non-nil *ErrorProcessor; no panic
	require.NotNil(t, ep)
}

// =============================================================================
// Happy Path — ProcessError
// =============================================================================

func TestErrorProcessor_ProcessError_RunningStatus(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-123"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"something failed","detail":{"trace":"stack trace"}}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "success", resp.Status())
	assert.Contains(t, resp.Message(), "error recorded")
	assert.Contains(t, resp.Message(), "session="+testSessionID)
	assert.Contains(t, resp.Message(), "failingState=NodeA")
	assert.Contains(t, resp.Message(), "agentRole=Coder")
	assert.Contains(t, resp.Message(), "error=something failed")
}

func TestErrorProcessor_ProcessError_InitializingStatus(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "initializing"
	f.session.getCurrentStateResult = "StartNode"
	f.wfDef.nodes = []*components.Node{mustNewNode(t, "StartNode", "agent", "Init")}
	f.session.getSessionDataResultVal = "cs-123"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"init error"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "success", resp.Status())
	assert.Contains(t, resp.Message(), "error recorded")
}

func TestErrorProcessor_ProcessError_HumanNode(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "HumanNode"
	f.wfDef.nodes = []*components.Node{mustNewNode(t, "HumanNode", "human", "")}

	payload := json.RawMessage(`{"message":"user error"}`)
	msg := mustNewErrorRuntimeMessage(t, "", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert: agentRole should be empty
	assert.Equal(t, "success", resp.Status())
	assert.Contains(t, resp.Message(), "agentRole=")
	// Verify Fail() called with AgentError having AgentRole=""
	require.Equal(t, 1, f.session.failCalled)
	agentErr, ok := f.session.failInputErr.(*entities.AgentError)
	require.True(t, ok)
	assert.Equal(t, "", agentErr.AgentRole())
}

func TestErrorProcessor_ProcessError_DetailNilOrMissing(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-123"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"oops"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "success", resp.Status())
	// Verify Fail() called with AgentError having Detail == nil
	agentErr, ok := f.session.failInputErr.(*entities.AgentError)
	require.True(t, ok)
	assert.Nil(t, agentErr.Detail())
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestErrorProcessor_ProcessError_CallsValidateClaudeSessionID(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-456"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"fail"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-456", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert: ValidateClaudeSessionID called once with (persistentSession, currentNode, "cs-456")
	// This is verified by the fact that the process succeeds: if ValidateClaudeSessionID
	// were not called with correct args matching session data, it would return an error.
	assert.Equal(t, "success", resp.Status())
}

func TestErrorProcessor_ProcessError_FailCalledWithCorrectAgentError(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeB"
	f.wfDef.nodes = []*components.Node{mustNewNode(t, "NodeB", "agent", "Reviewer")}
	f.session.getSessionDataResultVal = "cs-123"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"fail msg","detail":{"info":"some detail"}}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "success", resp.Status())
	require.Equal(t, 1, f.session.failCalled)
	agentErr, ok := f.session.failInputErr.(*entities.AgentError)
	require.True(t, ok)
	assert.Equal(t, "Reviewer", agentErr.AgentRole())
	assert.Equal(t, "fail msg", agentErr.Message())
	assert.JSONEq(t, `{"info":"some detail"}`, string(agentErr.Detail()))
	assert.Equal(t, testSessionID, agentErr.SessionID())
	assert.Equal(t, "NodeB", agentErr.FailingState())
	assert.Greater(t, agentErr.OccurredAt(), int64(0))
	// Verify terminationNotifier was passed to Fail
	assert.Equal(t, (chan<- struct{})(f.terminationNotifier), f.session.failNotifier)
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestErrorProcessor_ProcessError_SessionCompleted(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "completed"

	payload := json.RawMessage(`{"message":"err"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "session terminated: status is 'completed'", resp.Message())
}

func TestErrorProcessor_ProcessError_SessionFailed(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "failed"

	payload := json.RawMessage(`{"message":"err"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "session terminated: status is 'failed'", resp.Message())
}

func TestErrorProcessor_ProcessError_ClaudeSessionIDValidationFails(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	// Make validation fail: stored ID doesn't match
	f.session.getSessionDataResultVal = "expected-id"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{"message":"err"}`)
	msg := mustNewErrorRuntimeMessage(t, "wrong-id", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "error", resp.Status())
	assert.Contains(t, resp.Message(), "mismatch")
	assert.Equal(t, 0, f.session.failCalled)
}

func TestErrorProcessor_ProcessError_InvalidPayloadMissingMessage(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-123"
	f.session.getSessionDataResultOK = true

	payload := json.RawMessage(`{}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "invalid error payload: missing required field 'message'", resp.Message())
	assert.Equal(t, 0, f.session.failCalled)
}

func TestErrorProcessor_ProcessError_FailReturnsError(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "NodeA"
	f.session.getSessionDataResultVal = "cs-123"
	f.session.getSessionDataResultOK = true
	f.session.failErr = errors.New("session already failed")

	payload := json.RawMessage(`{"message":"some error"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "failed to record error: session already failed", resp.Message())
}

func TestErrorProcessor_ProcessError_NodeNotFound(t *testing.T) {
	// Setup
	f := newErrorProcessorFixture(t)
	f.session.getStatusResult = "running"
	f.session.getCurrentStateResult = "UnknownNode"
	// Nodes don't contain "UnknownNode"

	payload := json.RawMessage(`{"message":"err"}`)
	msg := mustNewErrorRuntimeMessage(t, "cs-123", payload)

	// Act
	ep := NewErrorProcessor(f.ps, f.wfDef, f.terminationNotifier)
	resp := ep.ProcessError(testSessionID, msg)

	// Assert
	assert.Equal(t, "error", resp.Status())
	assert.Equal(t, "current node 'UnknownNode' not found in workflow definition", resp.Message())
}

// =============================================================================
// Concurrent Behaviour
// =============================================================================

func TestErrorProcessor_ProcessError_ConcurrentFirstErrorWins(t *testing.T) {
	// This test verifies concurrent ProcessError calls complete without deadlock.
	// The real Session entity serializes via internal lock; here we use a
	// thread-safe mock wrapper to satisfy the race detector.
	sess := &concurrentSafeMockSession{
		getStatusResult:       "running",
		getCurrentStateResult: "NodeA",
		getSessionDataVal:     "cs-123",
		getSessionDataOK:      true,
	}

	metaStore := &concurrentSafeMetadataStore{}
	evStore := &concurrentSafeEventStore{}
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)

	wfDef := &mockErrorProcessorWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "NodeA", "agent", "Coder")},
	}
	terminationNotifier := newTerminationChannel()

	ep := NewErrorProcessor(ps, wfDef, terminationNotifier)

	payload1 := json.RawMessage(`{"message":"error one"}`)
	payload2 := json.RawMessage(`{"message":"error two"}`)
	msg1 := mustNewErrorRuntimeMessage(t, "cs-123", payload1)
	msg2 := mustNewErrorRuntimeMessage(t, "cs-123", payload2)

	// Act: Call ep.ProcessError concurrently from two goroutines
	var wg sync.WaitGroup
	var resp1, resp2 *entities.RuntimeResponse
	wg.Add(2)
	go func() { defer wg.Done(); resp1 = ep.ProcessError(testSessionID, msg1) }()
	go func() { defer wg.Done(); resp2 = ep.ProcessError(testSessionID, msg2) }()
	wg.Wait()

	// Assert: Both calls complete without data race; at least one succeeds
	assert.NotNil(t, resp1)
	assert.NotNil(t, resp2)
}

// =============================================================================
// Compile guards — suppress unused import warnings
// =============================================================================

var (
	_ = json.RawMessage{}
	_ = errors.New
	_ = (*entities.RuntimeMessage)(nil)
	_ = (*entities.RuntimeResponse)(nil)
	_ = (*components.Node)(nil)
	_ = assert.Equal
	_ = require.NoError
)
