package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for EventProcessor tests ---

// mockSessionForEvent provides a mock Session for EventProcessor tests.
type mockSessionForEvent struct {
	mock.Mock
	mu           sync.RWMutex
	status       string
	currentState string
	sessionData  map[string]any
	sessionID    string
	workflowName string
	eventHistory []session.Event
	err          error
}

func newMockSessionForEvent(status, currentState string) *mockSessionForEvent {
	return &mockSessionForEvent{
		status:       status,
		currentState: currentState,
		sessionData:  make(map[string]any),
		sessionID:    uuid.New().String(),
		workflowName: "TestWorkflow",
		eventHistory: []session.Event{},
	}
}

func (m *mockSessionForEvent) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockSessionForEvent) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *mockSessionForEvent) GetSessionDataSafe(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Called(key)
	val, ok := m.sessionData[key]
	return val, ok
}

func (m *mockSessionForEvent) UpdateEventHistorySafe(event session.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(event)
	if args.Error(0) == nil {
		m.eventHistory = append(m.eventHistory, event)
	}
	return args.Error(0)
}

func (m *mockSessionForEvent) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(err, terminationNotifier)
	if args.Error(0) == nil {
		m.status = "failed"
		m.err = err
		select {
		case terminationNotifier <- struct{}{}:
		default:
		}
	}
	return args.Error(0)
}

func (m *mockSessionForEvent) UpdateCurrentStateSafe(newState string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Called(newState)
	m.currentState = newState
	return nil
}

func (m *mockSessionForEvent) GetID() string {
	return m.sessionID
}

func (m *mockSessionForEvent) GetWorkflowName() string {
	return m.workflowName
}

func (m *mockSessionForEvent) getEventHistory() []session.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]session.Event(nil), m.eventHistory...)
}

// mockWorkflowDefinitionLoaderForEvent provides a mock WorkflowDefinitionLoader.
type mockWorkflowDefinitionLoaderForEvent struct {
	mock.Mock
}

func (m *mockWorkflowDefinitionLoaderForEvent) Load(workflowName string) (*storage.WorkflowDefinition, error) {
	args := m.Called(workflowName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.WorkflowDefinition), args.Error(1)
}

// mockTransitionToNode provides a mock TransitionToNode for EventProcessor tests.
type mockTransitionToNode struct {
	mock.Mock
}

func (m *mockTransitionToNode) Transition(message string, targetNodeName string, isExitTransition bool) error {
	args := m.Called(message, targetNodeName, isExitTransition)
	return args.Error(0)
}

// --- Test fixture helper ---

func createEventProcessorFixture(t *testing.T, status, currentState string) (
	*EventProcessor,
	*mockSessionForEvent,
	*mockWorkflowDefinitionLoaderForEvent,
	*mockTransitionToNode,
	chan struct{},
) {
	t.Helper()
	sess := newMockSessionForEvent(status, currentState)
	loader := &mockWorkflowDefinitionLoaderForEvent{}
	transitioner := &mockTransitionToNode{}
	terminationNotifier := make(chan struct{}, 2)

	ep, err := NewEventProcessor(sess, loader, transitioner, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, ep)

	return ep, sess, loader, transitioner, terminationNotifier
}

func buildEventWorkflowWithTransition(fromNode, eventType, toNode, agentRole string) *storage.WorkflowDefinition {
	nodes := []storage.Node{
		{Name: "Entry", Type: "human"},
	}
	if fromNode != "Entry" {
		nodes = append(nodes, storage.Node{Name: fromNode, Type: "agent", AgentRole: agentRole})
	}
	if toNode != "Entry" && toNode != fromNode {
		nodes = append(nodes, storage.Node{Name: toNode, Type: "agent", AgentRole: "Worker"})
	}
	transitions := []storage.Transition{
		{FromNode: "Entry", EventType: "Start", ToNode: fromNode},
		{FromNode: fromNode, EventType: eventType, ToNode: toNode},
	}
	return &storage.WorkflowDefinition{
		Name:        "TestWorkflow",
		EntryNode:   "Entry",
		Nodes:       nodes,
		Transitions: transitions,
	}
}

func buildAgentNodeWorkflow(nodeName, agentRole string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: nodeName, Type: "agent", AgentRole: agentRole},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: nodeName},
			{FromNode: nodeName, EventType: "Approved", ToNode: "NodeB"},
		},
	}
}

func buildHumanNodeWorkflow(nodeName string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: nodeName,
		Nodes: []storage.Node{
			{Name: nodeName, Type: "human"},
			{Name: "AgentA", Type: "agent", AgentRole: "Worker"},
		},
		Transitions: []storage.Transition{
			{FromNode: nodeName, EventType: "Continue", ToNode: "AgentA"},
		},
	}
}

func buildExitWorkflow(fromNode, eventType, exitToNode string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: fromNode, Type: "agent", AgentRole: "Worker"},
			{Name: exitToNode, Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: fromNode},
			{FromNode: fromNode, EventType: eventType, ToNode: exitToNode},
		},
		ExitTransitions: []storage.ExitTransition{
			{FromNode: fromNode, EventType: eventType, ToNode: exitToNode},
		},
	}
}

func buildEventPayload(t *testing.T, eventType, message string, payload json.RawMessage) json.RawMessage {
	t.Helper()
	ep := entities.EventPayload{
		EventType: eventType,
		Message:   message,
		Payload:   payload,
	}
	data, err := json.Marshal(ep)
	require.NoError(t, err)
	return data
}

func buildEventRuntimeMessage(t *testing.T, claudeSessionID, eventType, message string, payload json.RawMessage) entities.RuntimeMessage {
	t.Helper()
	return entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: claudeSessionID,
		Payload:         buildEventPayload(t, eventType, message, payload),
	}
}

// --- Happy Path — Construction ---

func TestEventProcessor_New(t *testing.T) {
	sess := newMockSessionForEvent("running", "AgentNode")
	loader := &mockWorkflowDefinitionLoaderForEvent{}
	transitioner := &mockTransitionToNode{}
	terminationNotifier := make(chan struct{}, 1)

	ep, err := NewEventProcessor(sess, loader, transitioner, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, ep)
}

// --- Happy Path — ProcessEvent (Agent Node) ---

func TestProcessEvent_AgentNode_ValidClaudeSessionID(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	sess.On("UpdateCurrentStateSafe", "NodeB").Return(nil)
	transitioner.On("Transition", "ok", "NodeB", false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "ok", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	assert.Regexp(t, `(?i)event 'Approved' processed successfully.*session=.*currentState=.*sessionStatus=`, resp.Message)
}

func TestProcessEvent_ComplexPayload(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	complexPayload := json.RawMessage(`{"nested":{"array":[1,2,3],"bool":true,"null":null}}`)
	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", complexPayload)
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

func TestProcessEvent_OptionalFieldsPresent(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "full message", json.RawMessage(`{"key":"value"}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Happy Path — ProcessEvent (Human Node) ---

func TestProcessEvent_HumanNode_EmptyClaudeSessionID(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "HumanNode")

	wf := buildHumanNodeWorkflow("HumanNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", "proceed", "AgentA", false).Return(nil)

	msg := buildEventRuntimeMessage(t, "", "Continue", "proceed", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Happy Path — Event Recording ---

func TestProcessEvent_EventRecordedToStore(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	var capturedEvent session.Event
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Run(func(args mock.Arguments) {
		capturedEvent = args.Get(0).(session.Event)
	}).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "test message", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	sess.AssertCalled(t, "UpdateEventHistorySafe", mock.AnythingOfType("session.Event"))
	assert.NotEmpty(t, capturedEvent.ID, "event should have a generated UUID")
	assert.Equal(t, "Approved", capturedEvent.Type)
	assert.Equal(t, "test message", capturedEvent.Message)
	assert.Equal(t, "AgentNode", capturedEvent.EmittedBy)
	assert.Equal(t, sess.GetID(), capturedEvent.SessionID)
	assert.Greater(t, capturedEvent.EmittedAt, int64(0))
}

func TestProcessEvent_EventEmittedByAutoAssigned(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	var capturedEvent session.Event
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Run(func(args mock.Arguments) {
		capturedEvent = args.Get(0).(session.Event)
	}).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "AgentNode", capturedEvent.EmittedBy)
}

func TestProcessEvent_EventUUIDGenerated(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	var events []session.Event
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Run(func(args mock.Arguments) {
		events = append(events, args.Get(0).(session.Event))
	}).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg1 := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg1", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg1)

	msg2 := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg2", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg2)

	require.Len(t, events, 2)
	assert.NotEmpty(t, events[0].ID)
	assert.NotEmpty(t, events[1].ID)
	assert.NotEqual(t, events[0].ID, events[1].ID, "each event should have a unique UUID")

	// Validate UUID v4 format
	_, err1 := uuid.Parse(events[0].ID)
	_, err2 := uuid.Parse(events[1].ID)
	assert.NoError(t, err1, "event 1 ID should be valid UUID")
	assert.NoError(t, err2, "event 2 ID should be valid UUID")
}

// --- Happy Path — Transition Execution ---

func TestProcessEvent_TransitionToNodeCalled(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "NodeA")
	sess.currentState = "NodeA"
	claudeSessionID := uuid.New().String()
	sess.sessionData["NodeA.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "NodeA", Type: "agent", AgentRole: "Worker"},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "NodeA"},
			{FromNode: "NodeA", EventType: "Done", ToNode: "NodeB"},
		},
	}
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "NodeA.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", "event message", "NodeB", false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Done", "event message", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	transitioner.AssertCalled(t, "Transition", "event message", "NodeB", false)
}

func TestProcessEvent_ExitTransition(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildExitWorkflow("AgentNode", "Complete", "ExitNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "ExitNode", true).Run(func(args mock.Arguments) {
		sess.mu.Lock()
		sess.status = "completed"
		sess.mu.Unlock()
	}).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Complete", "done", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	assert.Regexp(t, `(?i)sessionStatus=completed`, resp.Message)
	transitioner.AssertCalled(t, "Transition", mock.Anything, "ExitNode", true)
}

func TestProcessEvent_CurrentStateUpdated(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "NodeA")
	claudeSessionID := uuid.New().String()
	sess.sessionData["NodeA.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "NodeA", Type: "agent", AgentRole: "Worker"},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "NodeA"},
			{FromNode: "NodeA", EventType: "Done", ToNode: "NodeB"},
		},
	}
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "NodeA.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "NodeB", false).Run(func(args mock.Arguments) {
		sess.mu.Lock()
		sess.currentState = "NodeB"
		sess.mu.Unlock()
	}).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Done", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "NodeB", sess.GetCurrentStateSafe())
	assert.Contains(t, resp.Message, "currentState=NodeB")
}

// --- Validation Failures — Session Status ---

func TestProcessEvent_StatusInitializing(t *testing.T) {
	ep, sess, _, _, _ := createEventProcessorFixture(t, "initializing", "EntryNode")

	msg := buildEventRuntimeMessage(t, "", "Test", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session not ready: status is 'initializing'", resp.Message)
}

func TestProcessEvent_StatusCompleted(t *testing.T) {
	ep, sess, _, _, _ := createEventProcessorFixture(t, "completed", "EndNode")

	msg := buildEventRuntimeMessage(t, "", "Test", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session not ready: status is 'completed'", resp.Message)
}

func TestProcessEvent_StatusFailed(t *testing.T) {
	ep, sess, _, _, _ := createEventProcessorFixture(t, "failed", "FailedNode")

	msg := buildEventRuntimeMessage(t, "", "Test", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session not ready: status is 'failed'", resp.Message)
}

// --- Validation Failures — Claude Session ID (Agent Node) ---

func TestProcessEvent_AgentNode_ClaudeSessionIDNotFound(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	// SessionData does NOT contain AgentNode.ClaudeSessionID
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(nil, false)

	msg := buildEventRuntimeMessage(t, uuid.New().String(), "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "claude session ID not found for node 'AgentNode'", resp.Message)
	// No event should be recorded
	sess.AssertNotCalled(t, "UpdateEventHistorySafe", mock.Anything)
}

func TestProcessEvent_AgentNode_ClaudeSessionIDMismatch(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")
	storedUUID := uuid.New().String()
	providedUUID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = storedUUID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(storedUUID, true)

	msg := buildEventRuntimeMessage(t, providedUUID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)claude session ID mismatch: expected `+storedUUID+` but got `+providedUUID, resp.Message)
	// No event should be recorded
	sess.AssertNotCalled(t, "UpdateEventHistorySafe", mock.Anything)
}

// --- Validation Failures — Claude Session ID (Human Node) ---

func TestProcessEvent_HumanNode_NonEmptyClaudeSessionID(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "HumanNode")

	wf := buildHumanNodeWorkflow("HumanNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)

	msg := buildEventRuntimeMessage(t, uuid.New().String(), "Continue", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "invalid claude session ID for human node: must be empty", resp.Message)
	// No event should be recorded
	sess.AssertNotCalled(t, "UpdateEventHistorySafe", mock.Anything)
}

// --- Validation Failures — Workflow Definition ---

func TestProcessEvent_WorkflowDefinitionNotFound(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")

	loader.On("Load", "TestWorkflow").Return(nil, fmt.Errorf("workflow definition not found: TestWorkflow"))

	msg := buildEventRuntimeMessage(t, uuid.New().String(), "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, resp.Message)
}

func TestProcessEvent_WorkflowDefinitionParseError(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")

	loader.On("Load", "TestWorkflow").Return(nil, fmt.Errorf("failed to parse workflow definition 'TestWorkflow': invalid YAML"))

	msg := buildEventRuntimeMessage(t, uuid.New().String(), "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, resp.Message)
}

// --- Validation Failures — Message Payload ---

func TestProcessEvent_MissingEventTypeField(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)

	// Payload without eventType field
	payload := json.RawMessage(`{"message":"test"}`)
	msg := entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: claudeSessionID,
		Payload:         payload,
	}
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)invalid event payload: missing eventType`, resp.Message)
}

// --- Error Propagation — Event Recording Failure ---

func TestProcessEvent_EventStoreWriteFailure(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(fmt.Errorf("disk full"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "EventProcessor"
	}), mock.Anything).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)failed to record event:`, resp.Message)
	// Transition should NOT be attempted
	sess.AssertNotCalled(t, "UpdateCurrentStateSafe", mock.Anything)
}

// --- Error Propagation — No Matching Transition ---

func TestProcessEvent_NoMatchingTransition(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)

	// Use an eventType that has no matching transition
	msg := buildEventRuntimeMessage(t, claudeSessionID, "Unknown", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)no transition found for event 'Unknown' from node 'AgentNode'`, resp.Message)
	// Event was already recorded
	sess.AssertCalled(t, "UpdateEventHistorySafe", mock.AnythingOfType("session.Event"))
}

func TestProcessEvent_SessionRemainsRunningAfterNoTransition(t *testing.T) {
	ep, sess, loader, _, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Unknown", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "running", sess.GetStatusSafe(), "session should remain running")
	// Session.Fail should NOT be called
	sess.AssertNotCalled(t, "Fail", mock.Anything, mock.Anything)
}

// --- Error Propagation — Transition Execution Failure ---

func TestProcessEvent_TransitionToNodeFails(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "NodeB", false).Return(fmt.Errorf("target node not found"))

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	assert.Regexp(t, `(?i)transition failed:`, resp.Message)
}

func TestProcessEvent_TransitionToNodeSessionFailCalledInternally(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	// TransitionToNode fails (and calls Session.Fail internally)
	transitioner.On("Transition", mock.Anything, "NodeB", false).Return(fmt.Errorf("agent invocation failed"))

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "error", resp.Status)
	// EventProcessor does NOT call Session.Fail when TransitionToNode fails
	sess.AssertNotCalled(t, "Fail", mock.Anything, mock.Anything)
}

// --- Boundary Values — Event Message Content ---

func TestProcessEvent_VeryLargeEventMessage(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	largeMessage := strings.Repeat("B", 5*1024*1024) // 5 MB
	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", largeMessage, json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

func TestProcessEvent_VeryLargePayload(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	largeData := strings.Repeat("Z", 5*1024*1024)
	payload := json.RawMessage(fmt.Sprintf(`{"data":"%s"}`, largeData))
	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", payload)
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

func TestProcessEvent_UnicodeInEvent(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "emoji 🎉", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Boundary Values — Field Values ---

func TestProcessEvent_MinimalEventPayload(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	// Minimal payload: only eventType
	payload := json.RawMessage(`{"eventType":"Approved"}`)
	msg := entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: claudeSessionID,
		Payload:         payload,
	}
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

func TestProcessEvent_EmptyNestedPayload(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
}

// --- Mock / Dependency Interaction — Session Methods ---

func TestProcessEvent_SessionGetSessionDataSafeCalled(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	sess.AssertCalled(t, "GetSessionDataSafe", "AgentNode.ClaudeSessionID")
}

func TestProcessEvent_SessionUpdateEventHistorySafeCalled(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	sess.AssertNumberOfCalls(t, "UpdateEventHistorySafe", 1)
}

func TestProcessEvent_SessionMethodsNotDirectlyModified(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	// All interactions should be via methods, verified by mock expectations being satisfied
	sess.AssertCalled(t, "GetSessionDataSafe", "AgentNode.ClaudeSessionID")
	sess.AssertCalled(t, "UpdateEventHistorySafe", mock.AnythingOfType("session.Event"))
}

// --- Mock / Dependency Interaction — TransitionEvaluator ---

func TestProcessEvent_TransitionEvaluatorCalled(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "NodeB", false).Return(nil)

	// Use "Approved" which matches the transition in buildAgentNodeWorkflow
	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	// Transition found and called indicates TransitionEvaluator was invoked
	assert.Equal(t, "success", resp.Status)
	transitioner.AssertCalled(t, "Transition", mock.Anything, "NodeB", false)
}

// --- Mock / Dependency Interaction — TransitionToNode ---

func TestProcessEvent_TransitionToNodeParameters(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", "proceed", "NodeB", false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "proceed", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	transitioner.AssertCalled(t, "Transition", "proceed", "NodeB", false)
}

func TestProcessEvent_TransitionToNodeExitTransition(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildExitWorkflow("AgentNode", "Complete", "ExitNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "ExitNode", true).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Complete", "done", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	transitioner.AssertCalled(t, "Transition", mock.Anything, "ExitNode", true)
}

// --- Mock / Dependency Interaction — WorkflowDefinitionLoader ---

func TestProcessEvent_WorkflowDefinitionLoaderCalled(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	loader.AssertCalled(t, "Load", "TestWorkflow")
}

// --- State Transitions ---

func TestProcessEvent_CurrentStateTransitionedByTransitionToNode(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "NodeA")
	claudeSessionID := uuid.New().String()
	sess.sessionData["NodeA.ClaudeSessionID"] = claudeSessionID

	wf := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "NodeA", Type: "agent", AgentRole: "Worker"},
			{Name: "NodeB", Type: "human"},
		},
		Transitions: []storage.Transition{
			{FromNode: "Entry", EventType: "Start", ToNode: "NodeA"},
			{FromNode: "NodeA", EventType: "Done", ToNode: "NodeB"},
		},
	}
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "NodeA.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "NodeB", false).Run(func(args mock.Arguments) {
		sess.mu.Lock()
		sess.currentState = "NodeB"
		sess.mu.Unlock()
	}).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Done", "msg", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	// EventProcessor does NOT directly modify CurrentState; TransitionToNode does
	assert.Equal(t, "NodeB", sess.GetCurrentStateSafe())
}

func TestProcessEvent_StatusTransitionedByTransitionToNode(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildExitWorkflow("AgentNode", "Complete", "ExitNode")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, "ExitNode", true).Run(func(args mock.Arguments) {
		sess.mu.Lock()
		sess.status = "completed"
		sess.mu.Unlock()
	}).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Complete", "done", json.RawMessage(`{}`))
	ep.ProcessEvent(sess.GetID(), msg)

	// EventProcessor does NOT directly modify Status; TransitionToNode does
	assert.Equal(t, "completed", sess.GetStatusSafe())
}

// --- Resource Cleanup ---

func TestProcessEvent_NoCleanupPerformed(t *testing.T) {
	ep, sess, loader, transitioner, _ := createEventProcessorFixture(t, "running", "AgentNode")
	claudeSessionID := uuid.New().String()
	sess.sessionData["AgentNode.ClaudeSessionID"] = claudeSessionID

	wf := buildAgentNodeWorkflow("AgentNode", "Reviewer")
	loader.On("Load", "TestWorkflow").Return(wf, nil)
	sess.On("GetSessionDataSafe", "AgentNode.ClaudeSessionID").Return(claudeSessionID, true)
	sess.On("UpdateEventHistorySafe", mock.AnythingOfType("session.Event")).Return(nil)
	transitioner.On("Transition", mock.Anything, mock.Anything, false).Return(nil)

	msg := buildEventRuntimeMessage(t, claudeSessionID, "Approved", "msg", json.RawMessage(`{}`))
	resp := ep.ProcessEvent(sess.GetID(), msg)

	assert.Equal(t, "success", resp.Status)
	// EventProcessor should NOT invoke socket deletion or SessionFinalizer
	// Verified by checking only expected methods were called
	sess.AssertNotCalled(t, "Fail", mock.Anything, mock.Anything)
}
