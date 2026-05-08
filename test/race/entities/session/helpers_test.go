package session_race

import (
	"encoding/json"
	"testing"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// --- Test Constants ---

const (
	testUUID          = "550e8400-e29b-41d4-a716-446655440000"
	testWorkflow      = "my-workflow"
	testEntryNode     = "start"
	testCreatedAt     = int64(1700000000)
	testSessionID     = testUUID
	testEventType     = "TaskCompleted"
	testEmittedBy     = "start"
	testEmittedAt     = int64(1700000001)
	testAgentRole     = "reviewer"
	testRuntimeIssuer = "orchestrator"
)

// --- Fixture Builders ---

// newTestSession constructs a valid session for race testing.
func newTestSession(t *testing.T) *session.Session {
	t.Helper()
	s, err := session.NewSession(testUUID, testWorkflow, testEntryNode, testCreatedAt)
	if err != nil {
		t.Fatalf("newTestSession: unexpected error: %v", err)
	}
	return s
}

// newTerminationChannel creates a buffered channel with capacity 2 for termination notification.
func newTerminationChannel() chan struct{} {
	return make(chan struct{}, 2)
}

// newTestAgentError creates a valid *AgentError for race testing.
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

// newTestEvent creates a valid *Event for race testing using the specified ID.
func newTestEvent(t *testing.T, id string) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		id,
		testEventType,
		"race test message",
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
