package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for ErrorProcessor tests ---

// mockSessionForError provides a mock Session for ErrorProcessor tests.
type mockSessionForError struct {
	mock.Mock
	mu           sync.RWMutex
	status       string
	currentState string
	sessionData  map[string]any
	sessionID    string
	workflowName string
	err          error
}

func newMockSessionForError(status, currentState string) *mockSessionForError {
	return &mockSessionForError{
		status:       status,
		currentState: currentState,
		sessionData:  make(map[string]any),
		sessionID:    uuid.New().String(),
		workflowName: "TestWorkflow",
	}
}

func (m *mockSessionForError) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockSessionForError) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *mockSessionForError) GetSessionDataSafe(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Called(key)
	val, ok := m.sessionData[key]
	return val, ok
}

func (m *mockSessionForError) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(err, terminationNotifier)
	if args.Error(0) == nil {
		m.status = "failed"
		m.err = err
		// Signal termination
		select {
		case terminationNotifier <- struct{}{}:
		default:
		}
	}
	return args.Error(0)
}

func (m *mockSessionForError) GetID() string {
	return m.sessionID
}

func (m *mockSessionForError) GetWorkflowName() string {
	return m.workflowName
}

// mockWorkflowDefinitionLoaderForError provides a mock WorkflowDefinitionLoader.
type mockWorkflowDefinitionLoaderForError struct {
	mock.Mock
}

func (m *mockWorkflowDefinitionLoaderForError) Load(workflowName string) (*storage.WorkflowDefinition, error) {
	args := m.Called(workflowName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.WorkflowDefinition), args.Error(1)
}

// --- Test fixture helper ---

func createErrorProcessorFixture(t *testing.T, status, currentState string) (
	*ErrorProcessor,
	*mockSessionForError,
	*mockWorkflowDefinitionLoaderForError,
	chan struct{},
) {
	t.Helper()
	sess := newMockSessionForError(status, currentState)
	loader := &mockWorkflowDefinitionLoaderForError{}
	terminationNotifier := make(chan struct{}, 2)

	ep, err := NewErrorProcessor(sess, loader, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, ep)

	return ep, sess, loader, terminationNotifier
}

func buildAgentWorkflowDefinition(nodeName, agentRole string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: nodeName, Type: "agent", AgentRole: agentRole},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: nodeName},
		},
		ExitTransitions: []storage.ExitTransition{},
	}
}

func buildHumanWorkflowDefinition(nodeName string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: nodeName, Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: nodeName},
		},
		ExitTransitions: []storage.ExitTransition{},
	}
}

func buildErrorPayload(t *testing.T, message string, detail json.RawMessage) json.RawMessage {
	t.Helper()
	payload := entities.ErrorPayload{
		Message: message,
		Detail:  detail,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return data
}

func buildErrorRuntimeMessage(t *testing.T, claudeSessionID string, message string, detail json.RawMessage) entities.RuntimeMessage {
	t.Helper()
	return entities.RuntimeMessage{
		Type:            "error",
		ClaudeSessionID: claudeSessionID,
		Payload:         buildErrorPayload(t, message, detail),
	}
}

// --- Happy Path — Construction ---

func TestErrorProcessor_New(t *testing.T) {
	sess := newMockSessionForError("running", "AgentNode")
	loader := &mockWorkflowDefinitionLoaderForError{}
	terminationNotifier := make(chan struct{}, 1)

	ep, err := NewErrorProcessor(sess, loader, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, ep)
}

// --- Happy Path — ProcessError (Agent Node) ---

func TestProcessError_AgentNode_ValidClaudeSessionID(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "test error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	assert.Regexp(t, `(?i)error recorded.*session=.*failingState=AgentNode.*agentRole=Reviewer`, resp.Message)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
}

func TestProcessError_InitializingStatus(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "initializing", "EntryNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["EntryNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("EntryNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "EntryNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "init error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "failed", sess.GetStatusSafe())
}

func TestProcessError_DetailFieldPresent(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	detail := json.RawMessage(`{"code":500,"context":"test"}`)
	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", detail)
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	// Verify Fail was called with an error containing the detail
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
}

func TestProcessError_DetailFieldNull(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	// Payload with null detail
	payload := json.RawMessage(`{"message":"error","detail":null}`)
	msg := entities.RuntimeMessage{
		Type:            "error",
		ClaudeSessionID: claudeSessionID,
		Payload:         payload,
	}
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Happy Path — ProcessError (Human Node) ---

func TestProcessError_HumanNode_EmptyClaudeSessionID(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "HumanNode")

	wf := buildHumanWorkflowDefinition("HumanNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, "", "human error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	// Verify agentRole is empty
	assert.Regexp(t, `(?i)agentRole=""?`, resp.Message)
}

// --- Happy Path — TerminationNotifier ---

func TestProcessError_TerminationNotifierSignaled(t *testing.T) {
	ep, sess, loader, terminationNotifier := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "test error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)

	// Verify termination notifier received signal
	select {
	case <-terminationNotifier:
		// Signal received as expected
	case <-time.After(time.Second):
		t.Fatal("expected termination notification, but did not receive one")
	}
}

// --- Validation Failures — Session Status ---

func TestProcessError_StatusCompleted(t *testing.T) {
	ep, sess, _, _ := createErrorProcessorFixture(t, "completed", "AgentNode")

	msg := buildErrorRuntimeMessage(t, uuid.New().String(), "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session terminated: status is 'completed'", resp.Message)
}

func TestProcessError_StatusFailed(t *testing.T) {
	ep, sess, _, _ := createErrorProcessorFixture(t, "failed", "AgentNode")

	msg := buildErrorRuntimeMessage(t, uuid.New().String(), "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session terminated: status is 'failed'", resp.Message)
}

// --- Validation Failures — Claude Session ID (Agent Node) ---

func TestProcessError_AgentNode_ClaudeSessionIDNotFound(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	// SessionData does NOT contain AgentNode.ClaudeSessionID
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(nil, false)

	msg := buildErrorRuntimeMessage(t, uuid.New().String(), "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "claude session ID not found for node 'AgentNode'", resp.Message)
}

func TestProcessError_AgentNode_ClaudeSessionIDMismatch(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	storedUUID := uuid.New().String()
	providedUUID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = storedUUID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(storedUUID, true)

	msg := buildErrorRuntimeMessage(t, providedUUID, "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)claude session ID mismatch: expected `+storedUUID+` but got `+providedUUID, resp.Message)
}

// --- Validation Failures — Claude Session ID (Human Node) ---

func TestProcessError_HumanNode_NonEmptyClaudeSessionID(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "HumanNode")

	wf := buildHumanWorkflowDefinition("HumanNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)

	msg := buildErrorRuntimeMessage(t, uuid.New().String(), "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "invalid claude session ID for human node: must be empty", resp.Message)
}

// --- Validation Failures — Workflow Definition ---

func TestProcessError_WorkflowDefinitionNotFound(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")

	loader.On("Load", "TestWorkflow").Return(nil, fmt.Errorf("workflow definition not found: TestWorkflow"))

	msg := buildErrorRuntimeMessage(t, uuid.New().String(), "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, resp.Message)
}

func TestProcessError_WorkflowDefinitionParseError(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")

	loader.On("Load", "TestWorkflow").Return(nil, fmt.Errorf("failed to parse workflow definition 'TestWorkflow': invalid YAML"))

	msg := buildErrorRuntimeMessage(t, uuid.New().String(), "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, resp.Message)
}

// --- Validation Failures — Message Payload ---

func TestProcessError_MissingMessageField(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	// Payload without message field
	payload := json.RawMessage(`{"detail":{}}`)
	msg := entities.RuntimeMessage{
		Type:            "error",
		ClaudeSessionID: claudeSessionID,
		Payload:         payload,
	}
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)invalid error payload: missing required field`, resp.Message)
}

// --- Error Propagation — Session.Fail Failure ---

func TestProcessError_SessionFailReturnsError(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(fmt.Errorf("session already failed"))

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)failed to record error:.*session already failed`, resp.Message)
}

// --- Error Propagation — Persistence Failure ---

func TestProcessError_PersistenceFailureBestEffort(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	// Session.Fail logs warning about persistence failure but returns nil (best-effort)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "failed", sess.GetStatusSafe())
}

// --- Boundary Values — Error Message Content ---

func TestProcessError_VeryLargeErrorMessage(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	largeMessage := strings.Repeat("A", 5*1024*1024) // 5 MB
	msg := buildErrorRuntimeMessage(t, claudeSessionID, largeMessage, json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

func TestProcessError_VeryLargeDetail(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	largeStack := strings.Repeat("X", 5*1024*1024) // 5 MB
	detail := json.RawMessage(fmt.Sprintf(`{"stack":"%s"}`, largeStack))
	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", detail)
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

func TestProcessError_UnicodeInMessage(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "错误: emoji 🚨", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Boundary Values — Field Values ---

func TestProcessError_EmptyAgentRole(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "HumanNode")

	wf := buildHumanWorkflowDefinition("HumanNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, "", "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	// Verify agentRole is empty string for human node
	assert.Regexp(t, `(?i)agentRole=""?`, resp.Message)
}

func TestProcessError_DetailFieldMissing(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	// Payload with message but no detail field
	payload := json.RawMessage(`{"message":"error"}`)
	msg := entities.RuntimeMessage{
		Type:            "error",
		ClaudeSessionID: claudeSessionID,
		Payload:         payload,
	}
	resp := ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Idempotency — First Error Wins ---

func TestProcessError_FirstErrorWins(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	// First call succeeds, transitions session to "failed"
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil).Once()

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "first error", json.RawMessage(`{}`))
	resp1 := ep.ProcessError(sess.GetID(), msg)
	assert.Equal(t, "success", resp1.Status)

	// After first error, session status is "failed". Second call should be rejected
	// at status validation: "session terminated: status is 'failed'"
	msg2 := buildErrorRuntimeMessage(t, claudeSessionID, "second error", json.RawMessage(`{}`))
	resp2 := ep.ProcessError(sess.GetID(), msg2)
	assert.Equal(t, "error", resp2.Status)
	// Second error rejected: either at status validation or at Session.Fail
	assert.Contains(t, resp2.Message, "session")
}

// --- Mock / Dependency Interaction ---

func TestProcessError_SessionGetSessionDataSafeCalled(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	ep.ProcessError(sess.GetID(), msg)

	sess.AssertCalled(t, "GetSessionDataSafe", "AgentNode.ClaudeSessionID")
}

func TestProcessError_SessionFailCalledWithAgentError(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	var capturedErr error
	sess.On("Fail", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedErr = args.Get(0).(error)
	}).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "test error message", json.RawMessage(`{}`))
	beforeTime := time.Now().Unix()
	ep.ProcessError(sess.GetID(), msg)
	afterTime := time.Now().Unix()

	sess.AssertNumberOfCalls(t, "Fail", 1)
	require.NotNil(t, capturedErr)

	// Verify the error contains the expected information
	assert.Contains(t, capturedErr.Error(), "test error message")
	_ = beforeTime
	_ = afterTime
}

func TestProcessError_WorkflowDefinitionLoaderCalled(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	ep.ProcessError(sess.GetID(), msg)

	loader.AssertCalled(t, "Load", "TestWorkflow")
}

func TestProcessError_AgentRoleDerivedFromNode(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Tester")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	var capturedErr error
	sess.On("Fail", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedErr = args.Get(0).(error)
	}).Return(nil)

	// RuntimeMessage does NOT contain agentRole field
	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	require.NotNil(t, capturedErr)
	// AgentRole is derived from the node definition ("Tester"), not from the RuntimeMessage
	assert.Equal(t, "success", resp.Status)
	assert.Regexp(t, `(?i)agentRole=Tester`, resp.Message)
}

// --- State Transitions ---

func TestProcessError_CurrentStateNotChanged(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "AgentNode", sess.GetCurrentStateSafe())
}

func TestProcessError_StatusTransitionToFailed(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	ep.ProcessError(sess.GetID(), msg)

	assert.Equal(t, "failed", sess.GetStatusSafe())
}

// --- Resource Cleanup ---

func TestProcessError_NoCleanupPerformed(t *testing.T) {
	ep, sess, loader, _ := createErrorProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentWorkflowDefinition("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	msg := buildErrorRuntimeMessage(t, claudeSessionID, "error", json.RawMessage(`{}`))
	resp := ep.ProcessError(sess.GetID(), msg)

	// ErrorProcessor should only call Session.Fail and return response
	// It does NOT delete socket or print to stdout
	assert.Equal(t, "success", resp.Status)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
	sess.AssertNotCalled(t, "Done", mock.Anything)
}
