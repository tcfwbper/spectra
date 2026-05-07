package runtime

import (
	"encoding/json"
	"testing"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// --- Test Constants ---

const (
	testSessionID    = "550e8400-e29b-41d4-a716-446655440000"
	testWorkflowName = "my-workflow"
	testEntryNode    = "start"
	testCreatedAt    = int64(1700000000)
	testEventID      = "770e8400-e29b-41d4-a716-446655440000"
	testEventID2     = "880e8400-e29b-41d4-a716-446655440000"
	testEventType    = "TaskCompleted"
	testEmittedBy    = "start"
	testEmittedAt    = int64(1700000001)
)

// --- Mock Session ---

// mockSession implements the Session interface expected by PersistentSession.
// It records calls for assertion and returns configured values.
type mockSession struct {
	id           string
	workflowName string

	// Run
	runCalled int
	runErr    error

	// Done
	doneCalled   int
	doneNotifier chan<- struct{}
	doneErr      error

	// Fail
	failCalled   int
	failErr      error
	failInputErr error
	failNotifier chan<- struct{}

	// UpdateCurrentStateSafe
	updateCurrentStateCalled int
	updateCurrentStateInput  string
	updateCurrentStateErr    error

	// UpdateSessionDataSafe
	updateSessionDataCalled   int
	updateSessionDataInputKey string
	updateSessionDataInputVal any
	updateSessionDataErr      error

	// UpdateEventHistorySafe
	updateEventHistoryCalled int
	updateEventHistoryInput  entities.Event
	updateEventHistoryErr    error

	// GetStatusSafe
	getStatusResult string

	// GetCurrentStateSafe
	getCurrentStateResult string

	// GetErrorSafe
	getErrorResult error

	// GetMetadataSnapshotSafe
	getMetadataSnapshotResult session.SessionMetadata

	// GetSessionDataSafe
	getSessionDataResultVal any
	getSessionDataResultOK  bool
}

func (m *mockSession) Run() error {
	m.runCalled++
	return m.runErr
}

func (m *mockSession) Done(notifier chan<- struct{}) error {
	m.doneCalled++
	m.doneNotifier = notifier
	return m.doneErr
}

func (m *mockSession) Fail(err error, notifier chan<- struct{}) error {
	m.failCalled++
	m.failInputErr = err
	m.failNotifier = notifier
	return m.failErr
}

func (m *mockSession) UpdateCurrentStateSafe(newState string) error {
	m.updateCurrentStateCalled++
	m.updateCurrentStateInput = newState
	return m.updateCurrentStateErr
}

func (m *mockSession) UpdateSessionDataSafe(key string, value any) error {
	m.updateSessionDataCalled++
	m.updateSessionDataInputKey = key
	m.updateSessionDataInputVal = value
	return m.updateSessionDataErr
}

func (m *mockSession) UpdateEventHistorySafe(event entities.Event) error {
	m.updateEventHistoryCalled++
	m.updateEventHistoryInput = event
	return m.updateEventHistoryErr
}

func (m *mockSession) GetStatusSafe() string {
	return m.getStatusResult
}

func (m *mockSession) GetCurrentStateSafe() string {
	return m.getCurrentStateResult
}

func (m *mockSession) GetErrorSafe() error {
	return m.getErrorResult
}

func (m *mockSession) GetMetadataSnapshotSafe() session.SessionMetadata {
	result := m.getMetadataSnapshotResult
	result.ID = m.id
	result.WorkflowName = m.workflowName
	return result
}

func (m *mockSession) GetSessionDataSafe(key string) (any, bool) {
	return m.getSessionDataResultVal, m.getSessionDataResultOK
}

// --- Mock Node ---

// mockNode implements a Node interface with configurable Type() and Name()
// for ValidateClaudeSessionID testing.
type mockNode struct {
	nodeType string
	nodeName string
}

func (m *mockNode) Type() string { return m.nodeType }
func (m *mockNode) Name() string { return m.nodeName }

// --- Mock SessionMetadataStore ---

// mockSessionMetadataStore records Write calls for assertion.
type mockSessionMetadataStore struct {
	writeCalled int
	writeInput  session.SessionMetadata
	writeErr    error
}

func (m *mockSessionMetadataStore) Write(meta session.SessionMetadata) error {
	m.writeCalled++
	m.writeInput = meta
	return m.writeErr
}

// --- Mock EventStore ---

// mockEventStore records Append calls for assertion.
type mockEventStore struct {
	appendCalled int
	appendInput  *entities.Event
	appendErr    error
}

func (m *mockEventStore) Append(event *entities.Event) error {
	m.appendCalled++
	m.appendInput = event
	return m.appendErr
}

// --- Mock Logger ---

// mockLogger records Error calls for assertion on persistence failure logging.
type mockLogger struct {
	errorCalls []logCall
	warnCalls  []logCall
	infoCalls  []logCall
	debugCalls []logCall
}

type logCall struct {
	msg  string
	args []any
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.errorCalls = append(m.errorCalls, logCall{msg: msg, args: args})
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.warnCalls = append(m.warnCalls, logCall{msg: msg, args: args})
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.infoCalls = append(m.infoCalls, logCall{msg: msg, args: args})
}

func (m *mockLogger) Debug(msg string, args ...any) {
	m.debugCalls = append(m.debugCalls, logCall{msg: msg, args: args})
}

// --- Fixture Builders ---

// newDefaultMockSession returns a mockSession with valid default configuration.
func newDefaultMockSession() *mockSession {
	return &mockSession{
		id:           testSessionID,
		workflowName: testWorkflowName,
		getMetadataSnapshotResult: session.SessionMetadata{
			ID:           testSessionID,
			WorkflowName: testWorkflowName,
			Status:       "running",
			CreatedAt:    testCreatedAt,
			UpdatedAt:    testCreatedAt + 1,
			CurrentState: testEntryNode,
			SessionData:  map[string]any{},
		},
		getStatusResult:       "running",
		getCurrentStateResult: testEntryNode,
	}
}

// newDefaultMockMetadataStore returns a mockSessionMetadataStore with no errors.
func newDefaultMockMetadataStore() *mockSessionMetadataStore {
	return &mockSessionMetadataStore{}
}

// newDefaultMockEventStore returns a mockEventStore with no errors.
func newDefaultMockEventStore() *mockEventStore {
	return &mockEventStore{}
}

// newDefaultMockLogger returns a mockLogger that records calls.
func newDefaultMockLogger() *mockLogger {
	return &mockLogger{}
}

// newTestEvent creates a valid Event for testing.
func newTestEvent(t *testing.T, id string) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		id,
		testEventType,
		"test message",
		json.RawMessage(`{"key":"value"}`),
		testEmittedBy,
		testEmittedAt,
		testSessionID,
	)
	if err != nil {
		t.Fatalf("newTestEvent: unexpected error: %v", err)
	}
	return ev
}

// newTerminationChannel creates a buffered channel for termination notification.
func newTerminationChannel() chan struct{} {
	return make(chan struct{}, 2)
}

// assertLogContainsArg checks that at least one logCall contains the given key-value pair.
func assertLogContainsArg(t *testing.T, calls []logCall, expectedMsg string, key string, value any) {
	t.Helper()
	for _, call := range calls {
		if call.msg != expectedMsg {
			continue
		}
		for i := 0; i+1 < len(call.args); i += 2 {
			if call.args[i] == key && call.args[i+1] == value {
				return
			}
		}
	}
	t.Errorf("expected log call with msg=%q containing %s=%v, got calls: %+v", expectedMsg, key, value, calls)
}

// assertLogHasMessage checks that at least one logCall has the exact message.
func assertLogHasMessage(t *testing.T, calls []logCall, expectedMsg string) {
	t.Helper()
	for _, call := range calls {
		if call.msg == expectedMsg {
			return
		}
	}
	t.Errorf("expected log call with msg=%q, got calls: %+v", expectedMsg, calls)
}
