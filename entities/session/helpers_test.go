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

// testSession wraps Session with access to mock stores for testing.
type testSession struct {
	*Session
	metadataStore            *mockSessionMetadataStore
	eventStore               *mockEventStore
	logger                   *mockLogger
	terminationNotifierChan  chan struct{} // Bidirectional channel for testing
}

// createTestSession creates a session for testing with the given initial state.
func createTestSession(t *testing.T, status string, currentState string) *testSession {
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
		mu:                  sync.RWMutex{},
		terminationNotifier: terminationNotifier,
		metadataStore:       metadataStore,
		eventStore:          eventStore,
		logger:              logger,
	}

	return &testSession{
		Session:                 session,
		metadataStore:           metadataStore,
		eventStore:              eventStore,
		logger:                  logger,
		terminationNotifierChan: terminationNotifier,
	}
}

// createTestSessionWithUpdatedAt creates a session with a specific UpdatedAt timestamp.
func createTestSessionWithUpdatedAt(t *testing.T, status string, currentState string, updatedAt int64) *testSession {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.UpdatedAt = updatedAt
	return session
}

// createTestSessionWithData creates a session with pre-populated SessionData.
func createTestSessionWithData(t *testing.T, status string, currentState string, data map[string]any) *testSession {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.SessionData = data
	return session
}

// createTestSessionWithError creates a session with a pre-set error.
func createTestSessionWithError(t *testing.T, status string, currentState string, err error) *testSession {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.Error = err
	return session
}

// createTestSessionWithEvents creates a session with pre-populated EventHistory.
func createTestSessionWithEvents(t *testing.T, status string, currentState string, events []Event) *testSession {
	t.Helper()
	session := createTestSession(t, status, currentState)
	session.EventHistory = events
	return session
}
