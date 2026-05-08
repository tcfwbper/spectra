package runtime

import (
	"sync"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// concurrentSafeMockSession is a thread-safe mock of the Session interface,
// used by concurrent tests to satisfy the race detector. Unlike mockSession
// (which is not synchronized), this mock protects all fields with a mutex.
type concurrentSafeMockSession struct {
	mu sync.Mutex

	// Configurable return values
	getStatusResult       string
	getCurrentStateResult string
	getSessionDataVal     any
	getSessionDataOK      bool
	failErr               error
	doneErr               error
	updateEventHistoryErr error

	// Call tracking
	failCalled int
}

func (m *concurrentSafeMockSession) Run() error { return nil }

func (m *concurrentSafeMockSession) Done(notifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.doneErr
}

func (m *concurrentSafeMockSession) Fail(err error, notifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCalled++
	return m.failErr
}

func (m *concurrentSafeMockSession) UpdateCurrentStateSafe(newState string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCurrentStateResult = newState
	return nil
}

func (m *concurrentSafeMockSession) UpdateSessionDataSafe(key string, value any) error {
	return nil
}

func (m *concurrentSafeMockSession) UpdateEventHistorySafe(event entities.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateEventHistoryErr
}

func (m *concurrentSafeMockSession) GetStatusSafe() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getStatusResult
}

func (m *concurrentSafeMockSession) GetCurrentStateSafe() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getCurrentStateResult
}

func (m *concurrentSafeMockSession) GetErrorSafe() error {
	return nil
}

func (m *concurrentSafeMockSession) GetMetadataSnapshotSafe() session.SessionMetadata {
	m.mu.Lock()
	defer m.mu.Unlock()
	return session.SessionMetadata{
		ID:           testSessionID,
		WorkflowName: testWorkflowName,
		Status:       m.getStatusResult,
		CurrentState: m.getCurrentStateResult,
	}
}

func (m *concurrentSafeMockSession) GetSessionDataSafe(key string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getSessionDataVal, m.getSessionDataOK
}

// concurrentSafeTransitionToNode is a thread-safe mock of the
// TransitionToNodeExecutor interface used by concurrent event processor tests.
type concurrentSafeTransitionToNode struct {
	mu   sync.Mutex
	sess *concurrentSafeMockSession
}

func (m *concurrentSafeTransitionToNode) Execute(targetNodeName, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simulate updating current state (as the real TransitionToNode does)
	m.sess.mu.Lock()
	m.sess.getCurrentStateResult = targetNodeName
	m.sess.mu.Unlock()
	return nil
}

// concurrentSafeMetadataStore is a thread-safe mock of SessionMetadataStore.
type concurrentSafeMetadataStore struct {
	mu sync.Mutex
}

func (m *concurrentSafeMetadataStore) Write(meta session.SessionMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

// concurrentSafeEventStore is a thread-safe mock of EventStore.
type concurrentSafeEventStore struct {
	mu sync.Mutex
}

func (m *concurrentSafeEventStore) Append(event *entities.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}
