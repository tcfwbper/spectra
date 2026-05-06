package session

import (
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"

	"github.com/tcfwbper/spectra/entities"
)

// --- Test Constants ---

const (
	testUUID          = "550e8400-e29b-41d4-a716-446655440000"
	testUUID2         = "660e8400-e29b-41d4-a716-446655440000"
	testWorkflow      = "my-workflow"
	testEntryNode     = "start"
	testCreatedAt     = int64(1700000000)
	testSessionID     = testUUID
	testEventID       = "770e8400-e29b-41d4-a716-446655440000"
	testEventID2      = "880e8400-e29b-41d4-a716-446655440000"
	testEventType     = "TaskCompleted"
	testEmittedBy     = "start"
	testEmittedAt     = int64(1700000001)
	testAgentRole     = "reviewer"
	testRuntimeIssuer = "orchestrator"
)

// --- Fixture Builders ---

// newTestSession constructs a valid session for testing using default test constants.
func newTestSession(t *testing.T) *Session {
	t.Helper()
	s, err := NewSession(testUUID, testWorkflow, testEntryNode, testCreatedAt)
	if err != nil {
		t.Fatalf("newTestSession: unexpected error: %v", err)
	}
	return s
}

// newRunningSession constructs a session and transitions it to "running".
func newRunningSession(t *testing.T) *Session {
	t.Helper()
	s := newTestSession(t)
	if err := s.Run(); err != nil {
		t.Fatalf("newRunningSession: unexpected error from Run(): %v", err)
	}
	return s
}

// newTerminationChannel creates a buffered channel with capacity 2 for termination notification.
func newTerminationChannel() chan struct{} {
	return make(chan struct{}, 2)
}

// newTestAgentError creates a valid *AgentError for testing.
func newTestAgentError(t *testing.T) *entities.AgentError {
	t.Helper()
	ae, err := entities.NewAgentError(
		testAgentRole,
		"something went wrong",
		json.RawMessage(`{"detail":"info"}`),
		testEmittedAt,
		testSessionID,
		testEntryNode,
	)
	if err != nil {
		t.Fatalf("newTestAgentError: unexpected error: %v", err)
	}
	return ae
}

// newTestAgentError2 creates a second distinct *AgentError for testing first-error-wins.
func newTestAgentError2(t *testing.T) *entities.AgentError {
	t.Helper()
	ae, err := entities.NewAgentError(
		testAgentRole,
		"second error",
		json.RawMessage(`{"detail":"second"}`),
		testEmittedAt+1,
		testSessionID,
		testEntryNode,
	)
	if err != nil {
		t.Fatalf("newTestAgentError2: unexpected error: %v", err)
	}
	return ae
}

// newTestRuntimeError creates a valid *RuntimeError for testing.
func newTestRuntimeError(t *testing.T) *entities.RuntimeError {
	t.Helper()
	re, err := entities.NewRuntimeError(
		testRuntimeIssuer,
		"runtime failure",
		json.RawMessage(`{"detail":"runtime"}`),
		testEmittedAt,
		testSessionID,
		testEntryNode,
	)
	if err != nil {
		t.Fatalf("newTestRuntimeError: unexpected error: %v", err)
	}
	return re
}

// newTestEvent creates a valid *Event for testing using the specified ID.
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

// newTestEventMinimal creates a valid *Event with a specified message (including empty).
func newTestEventMinimal(t *testing.T, id string, message string) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		id,
		testEventType,
		message,
		json.RawMessage(`{}`),
		testEmittedBy,
		testEmittedAt,
		testSessionID,
	)
	if err != nil {
		t.Fatalf("newTestEventMinimal: unexpected error: %v", err)
	}
	return ev
}

// newTestEventEmptyPayload creates a valid *Event with an empty JSON object payload.
func newTestEventEmptyPayload(t *testing.T, id string) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		id,
		testEventType,
		"msg",
		json.RawMessage(`{}`),
		testEmittedBy,
		testEmittedAt,
		testSessionID,
	)
	if err != nil {
		t.Fatalf("newTestEventEmptyPayload: unexpected error: %v", err)
	}
	return ev
}

// newTestEventUnchecked constructs an Event with arbitrary field values for
// validation tests that must exercise Session-owned checks with invalid inputs.
func newTestEventUnchecked(t *testing.T, id string, eventType string, message string, payload json.RawMessage, emittedBy string, emittedAt int64, sessionID string) entities.Event {
	t.Helper()

	var event entities.Event
	setUnexportedField(t, &event, "id", id)
	setUnexportedField(t, &event, "eventType", eventType)
	setUnexportedField(t, &event, "message", message)
	setUnexportedField(t, &event, "payload", payload)
	setUnexportedField(t, &event, "emittedBy", emittedBy)
	setUnexportedField(t, &event, "emittedAt", emittedAt)
	setUnexportedField(t, &event, "sessionID", sessionID)

	return event
}

func setUnexportedField(t *testing.T, target any, fieldName string, value any) {
	t.Helper()

	field := reflect.ValueOf(target).Elem().FieldByName(fieldName)
	if !field.IsValid() {
		t.Fatalf("setUnexportedField: unknown field %q", fieldName)
	}

	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}
