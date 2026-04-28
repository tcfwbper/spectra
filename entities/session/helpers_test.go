package session

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// mockSessionMetadataStore is a mock implementation of SessionMetadataStore for testing.
type mockSessionMetadataStore struct {
	mock.Mock
}

func (m *mockSessionMetadataStore) Write(metadata SessionMetadata) error {
	args := m.Called(metadata)
	return args.Error(0)
}

// mockEventStore is a mock implementation of EventStore for testing.
type mockEventStore struct {
	mock.Mock
}

func (m *mockEventStore) WriteEvent(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

// mockLogger is a mock implementation of a logger for testing.
type mockLogger struct {
	mock.Mock
	warnings []string
	mu       sync.Mutex
}

func (m *mockLogger) Warning(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnings = append(m.warnings, msg)
	m.Called(msg)
}

func (m *mockLogger) GetWarnings() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.warnings...)
}

// SessionMetadata represents the persistable state of a session.
type SessionMetadata struct {
	ID           string         `json:"id"`
	WorkflowName string         `json:"workflow_name"`
	Status       string         `json:"status"`
	CreatedAt    int64          `json:"created_at"`
	UpdatedAt    int64          `json:"updated_at"`
	CurrentState string         `json:"current_state"`
	SessionData  map[string]any `json:"session_data"`
	Error        error          `json:"error,omitempty"`
}

// Session represents a single execution instance of a workflow.
// It embeds SessionMetadata and adds runtime-only fields.
type Session struct {
	SessionMetadata
	EventHistory        []Event
	mu                  sync.RWMutex
	terminationNotifier chan<- struct{}
	metadataStore       *mockSessionMetadataStore
	eventStore          *mockEventStore
	logger              *mockLogger
}

// Event represents a workflow event.
type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	SessionID string         `json:"session_id"`
	EmittedAt int64          `json:"emitted_at"`
	EmittedBy string         `json:"emitted_by"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload"`
}

// AgentError represents an error from an agent node.
type AgentError struct {
	NodeName string
	Message  string
}

func (e *AgentError) Error() string {
	return e.Message
}

// RuntimeError represents an error from the runtime.
type RuntimeError struct {
	Issuer  string
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Message
}

// createTestSession creates a session for testing with the given initial state.
func createTestSession(t *testing.T, status string, currentState string) *Session {
	t.Helper()
	now := time.Now().Unix()
	sessionID := uuid.New().String()

	metadataStore := &mockSessionMetadataStore{}
	eventStore := &mockEventStore{}
	logger := &mockLogger{}

	terminationNotifier := make(chan struct{}, 2)

	session := &Session{
		SessionMetadata: SessionMetadata{
			ID:           sessionID,
			WorkflowName: "TestWorkflow",
			Status:       status,
			CreatedAt:    now,
			UpdatedAt:    now,
			CurrentState: currentState,
			SessionData:  make(map[string]any),
			Error:        nil,
		},
		EventHistory:        []Event{},
		terminationNotifier: terminationNotifier,
		metadataStore:       metadataStore,
		eventStore:          eventStore,
		logger:              logger,
	}

	return session
}

// createTestSessionWithUpdatedAt creates a session with a specific UpdatedAt timestamp.
func createTestSessionWithUpdatedAt(t *testing.T, status string, currentState string, updatedAt int64) *Session {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.UpdatedAt = updatedAt
	return session
}

// createTestSessionWithData creates a session with pre-populated SessionData.
func createTestSessionWithData(t *testing.T, status string, currentState string, data map[string]any) *Session {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.SessionData = data
	return session
}

// createTestSessionWithError creates a session with a pre-set error.
func createTestSessionWithError(t *testing.T, status string, currentState string, err error) *Session {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.Error = err
	return session
}

// createTestSessionWithEvents creates a session with pre-populated EventHistory.
func createTestSessionWithEvents(t *testing.T, status string, currentState string, events []Event) *Session {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.EventHistory = events
	return session
}
