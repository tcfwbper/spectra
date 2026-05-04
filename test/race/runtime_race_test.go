package race_test

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/runtime"
)

// =====================================================================
// Mock types for Runtime race tests
// =====================================================================

// raceMockSessionForRuntime is a thread-safe mock Session for race tests.
type raceMockSessionForRuntime struct {
	mu           sync.RWMutex
	id           string
	workflowName string
	status       string
	currentState string
	err          error
	createdAt    int64
	updatedAt    int64
	sessionData  map[string]any
	eventHistory []session.Event
}

func newRaceMockSession(id, workflowName, status, currentState string) *raceMockSessionForRuntime {
	return &raceMockSessionForRuntime{
		id:           id,
		workflowName: workflowName,
		status:       status,
		currentState: currentState,
		sessionData:  make(map[string]any),
		createdAt:    time.Now().Unix(),
		updatedAt:    time.Now().Unix(),
	}
}

func (m *raceMockSessionForRuntime) Run(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = "running"
	return nil
}

func (m *raceMockSessionForRuntime) Done(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	m.status = "completed"
	m.mu.Unlock()
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *raceMockSessionForRuntime) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	if m.status == "failed" {
		m.mu.Unlock()
		return fmt.Errorf("session already failed")
	}
	if m.status == "completed" {
		m.mu.Unlock()
		return fmt.Errorf("cannot fail session: status is 'completed', workflow already terminated")
	}
	m.status = "failed"
	m.err = err
	m.mu.Unlock()
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *raceMockSessionForRuntime) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *raceMockSessionForRuntime) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *raceMockSessionForRuntime) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

func (m *raceMockSessionForRuntime) GetWorkflowName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.workflowName
}

func (m *raceMockSessionForRuntime) GetCreatedAt() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.createdAt
}

func (m *raceMockSessionForRuntime) GetUpdatedAt() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updatedAt
}

func (m *raceMockSessionForRuntime) GetEventHistory() []session.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.eventHistory
}

func (m *raceMockSessionForRuntime) GetSessionData() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]any)
	for k, v := range m.sessionData {
		result[k] = v
	}
	return result
}

func (m *raceMockSessionForRuntime) GetErrorSafe() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.err
}

// raceMockSessionInitializer implements runtime.SessionInitializerInterface for race tests.
type raceMockSessionInitializer struct {
	mu       sync.Mutex
	initFunc func(workflowName string, terminationNotifier chan<- struct{}) (runtime.SessionForInitializer, error)
}

func (m *raceMockSessionInitializer) Initialize(workflowName string, terminationNotifier chan<- struct{}) (runtime.SessionForInitializer, error) {
	m.mu.Lock()
	fn := m.initFunc
	m.mu.Unlock()
	if fn != nil {
		return fn(workflowName, terminationNotifier)
	}
	return nil, fmt.Errorf("not configured")
}

// raceMockSessionFinalizer implements runtime.SessionFinalizerInterface for race tests.
type raceMockSessionFinalizer struct {
	mu     sync.Mutex
	called bool
}

func (m *raceMockSessionFinalizer) Finalize(session runtime.SessionForFinalizer) {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
}

// raceMockRuntimeSocketManager implements runtime.RuntimeSocketManagerInterface for race tests.
type raceMockRuntimeSocketManager struct {
	mu           sync.Mutex
	listenErrCh  chan error
	listenDoneCh chan struct{}
	listenErr    error
	listenCalled bool
}

func newRaceMockSocketManager() *raceMockRuntimeSocketManager {
	return &raceMockRuntimeSocketManager{
		listenErrCh:  make(chan error, 1),
		listenDoneCh: make(chan struct{}),
	}
}

func (m *raceMockRuntimeSocketManager) CreateSocket() error {
	return nil
}

func (m *raceMockRuntimeSocketManager) Listen(handler runtime.MessageHandler) (<-chan error, <-chan struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listenCalled = true
	if m.listenErr != nil {
		return nil, nil, m.listenErr
	}
	return m.listenErrCh, m.listenDoneCh, nil
}

func (m *raceMockRuntimeSocketManager) DeleteSocket() error {
	return nil
}

// raceMockMessageRouter implements runtime.MessageRouterInterface for race tests.
type raceMockMessageRouter struct {
	mu sync.Mutex
}

func (m *raceMockMessageRouter) RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	return entities.RuntimeResponse{Status: "ok"}
}

// raceMockLogger captures log messages for race tests.
type raceMockLogger struct {
	mu       sync.Mutex
	messages []string
}

func (m *raceMockLogger) Log(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *raceMockLogger) Warning(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

// =====================================================================
// Concurrent Behaviour — Race Conditions
// =====================================================================

func TestRuntime_SessionDoneAndListenerErrorConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(tmpDir+"/.spectra", 0755))

	sess := newRaceMockSession(uuid.New().String(), "TestWorkflow", "running", "start_node")
	sm := newRaceMockSocketManager()
	logger := &raceMockLogger{}

	si := &raceMockSessionInitializer{
		initFunc: func(workflowName string, tn chan<- struct{}) (runtime.SessionForInitializer, error) {
			return sess, nil
		},
	}
	sf := &raceMockSessionFinalizer{}
	mr := &raceMockMessageRouter{}

	findFunc := func() (string, error) { return tmpDir, nil }

	// Both Session.Done and listenerErrCh send signals concurrently
	go func() {
		time.Sleep(10 * time.Millisecond)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			sess.Done(make(chan struct{}, 2))
		}()

		go func() {
			defer wg.Done()
			sm.listenErrCh <- fmt.Errorf("accept loop failure")
		}()

		wg.Wait()
		time.Sleep(20 * time.Millisecond)
		close(sm.listenDoneCh)
	}()

	err := runtime.Run("TestWorkflow", findFunc, si, sf, sm, mr, logger)

	// Runtime receives first signal (non-deterministic); proceeds to cleanup
	// No race condition on session state access (GetStatusSafe used)
	_ = err // Error depends on which signal wins
}

func TestRuntime_SessionDoneAndSIGINTConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(tmpDir+"/.spectra", 0755))

	sess := newRaceMockSession(uuid.New().String(), "TestWorkflow", "running", "start_node")
	sm := newRaceMockSocketManager()
	logger := &raceMockLogger{}

	si := &raceMockSessionInitializer{
		initFunc: func(workflowName string, tn chan<- struct{}) (runtime.SessionForInitializer, error) {
			return sess, nil
		},
	}
	sf := &raceMockSessionFinalizer{}
	mr := &raceMockMessageRouter{}

	findFunc := func() (string, error) { return tmpDir, nil }

	go func() {
		time.Sleep(10 * time.Millisecond)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			sess.Done(make(chan struct{}, 2))
		}()

		go func() {
			defer wg.Done()
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(syscall.SIGINT)
		}()

		wg.Wait()
		time.Sleep(20 * time.Millisecond)
		close(sm.listenDoneCh)
	}()

	err := runtime.Run("TestWorkflow", findFunc, si, sf, sm, mr, logger)

	// Appropriate error message based on which signal won
	_ = err
}

func TestRuntime_ListenerDoneChAlreadyClosed(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(tmpDir+"/.spectra", 0755))

	sess := newRaceMockSession(uuid.New().String(), "TestWorkflow", "running", "start_node")
	sm := newRaceMockSocketManager()
	logger := &raceMockLogger{}

	si := &raceMockSessionInitializer{
		initFunc: func(workflowName string, tn chan<- struct{}) (runtime.SessionForInitializer, error) {
			return sess, nil
		},
	}
	sf := &raceMockSessionFinalizer{}
	mr := &raceMockMessageRouter{}

	findFunc := func() (string, error) { return tmpDir, nil }

	// Close listenerDoneCh before Runtime waits
	close(sm.listenDoneCh)

	go func() {
		time.Sleep(10 * time.Millisecond)
		sess.mu.Lock()
		sess.status = "completed"
		sess.mu.Unlock()
	}()

	start := time.Now()
	err := runtime.Run("TestWorkflow", findFunc, si, sf, sm, mr, logger)

	elapsed := time.Since(start)
	// <-listenerDoneCh returns immediately (closed channel)
	assert.Less(t, elapsed, 5*time.Second, "Runtime should not delay when listenerDoneCh is already closed")
	_ = err

	sf.mu.Lock()
	assert.True(t, sf.called, "SessionFinalizer should be called without delay")
	sf.mu.Unlock()
}
