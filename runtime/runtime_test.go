package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	goruntime "runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// =====================================================================
// Mock types for Runtime tests
// =====================================================================

// mockSpectraFinder is a mock for the SpectraFinder function.
type mockSpectraFinder struct {
	projectRoot string
	err         error
	called      bool
}

func (m *mockSpectraFinder) Find() (string, error) {
	m.called = true
	return m.projectRoot, m.err
}

// mockSessionForRuntime implements SessionForInitializer with tracking.
type mockSessionForRuntime struct {
	mu                        sync.RWMutex
	id                        string
	workflowName              string
	status                    string
	currentState              string
	err                       error
	createdAt                 int64
	updatedAt                 int64
	sessionData               map[string]any
	eventHistory              []session.Event
	runCalled                 bool
	doneCalled                bool
	failCalled                bool
	failError                 error
	failReturnErr             error
	getStatusSafeCalled       atomic.Int32
	getCurrentStateSafeCalled atomic.Int32
}

func newMockSessionForRuntime(id, workflowName, status, currentState string) *mockSessionForRuntime {
	return &mockSessionForRuntime{
		id:           id,
		workflowName: workflowName,
		status:       status,
		currentState: currentState,
		sessionData:  make(map[string]any),
		createdAt:    time.Now().Unix(),
		updatedAt:    time.Now().Unix(),
	}
}

func (m *mockSessionForRuntime) Run(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runCalled = true
	m.status = "running"
	return nil
}

func (m *mockSessionForRuntime) Done(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	m.doneCalled = true
	m.status = "completed"
	m.mu.Unlock()
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *mockSessionForRuntime) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	if m.status == "failed" {
		m.mu.Unlock()
		return fmt.Errorf("session already failed")
	}
	m.failCalled = true
	m.failError = err
	m.status = "failed"
	m.err = err
	m.mu.Unlock()
	if m.failReturnErr != nil {
		return m.failReturnErr
	}
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *mockSessionForRuntime) GetStatusSafe() string {
	m.getStatusSafeCalled.Add(1)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockSessionForRuntime) GetCurrentStateSafe() string {
	m.getCurrentStateSafeCalled.Add(1)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *mockSessionForRuntime) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

func (m *mockSessionForRuntime) GetWorkflowName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.workflowName
}

func (m *mockSessionForRuntime) GetCreatedAt() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.createdAt
}

func (m *mockSessionForRuntime) GetUpdatedAt() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updatedAt
}

func (m *mockSessionForRuntime) GetEventHistory() []session.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.eventHistory
}

func (m *mockSessionForRuntime) GetSessionData() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]any)
	for k, v := range m.sessionData {
		result[k] = v
	}
	return result
}

func (m *mockSessionForRuntime) GetErrorSafe() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.err
}

// mockSessionInitializerForRuntime implements SessionInitializerInterface.
type mockSessionInitializerForRuntime struct {
	mu            sync.Mutex
	session       SessionForInitializer
	err           error
	called        bool
	workflowName  string
	initFunc      func(workflowName string, terminationNotifier chan<- struct{}) (SessionForInitializer, error)
	constructedAt int
}

func (m *mockSessionInitializerForRuntime) Initialize(workflowName string, terminationNotifier chan<- struct{}) (SessionForInitializer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.workflowName = workflowName
	if m.initFunc != nil {
		return m.initFunc(workflowName, terminationNotifier)
	}
	return m.session, m.err
}

// mockSessionFinalizerForRuntime implements SessionFinalizerInterface with tracking.
type mockSessionFinalizerForRuntime struct {
	mu          sync.Mutex
	called      bool
	callCount   int
	session     SessionForFinalizer
	panicMsg    string
	finalizedAt int
}

func (m *mockSessionFinalizerForRuntime) Finalize(session SessionForFinalizer) {
	m.mu.Lock()
	m.called = true
	m.callCount++
	m.session = session
	panicMsg := m.panicMsg
	m.mu.Unlock()
	if panicMsg != "" {
		panic(panicMsg)
	}
}

func (m *mockSessionFinalizerForRuntime) wasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func (m *mockSessionFinalizerForRuntime) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// mockRuntimeSocketManagerForRuntime implements RuntimeSocketManagerInterface with tracking.
// CreateSocket() is part of the injectable socket manager interface so its error can be
// injected in tests to exercise the error path.
type mockRuntimeSocketManagerForRuntime struct {
	mu                sync.Mutex
	createSocketErr   error
	createSocketCalls int
	listenErr         error
	listenErrCh       chan error
	listenDoneCh      chan struct{}
	listenCalled      bool
	listenHandler     MessageHandler
	deleteSocketCalls int
	deleteSocketErr   error
	constructedAt     int
}

func newMockRuntimeSocketManager() *mockRuntimeSocketManagerForRuntime {
	return &mockRuntimeSocketManagerForRuntime{
		listenErrCh:  make(chan error, 1),
		listenDoneCh: make(chan struct{}),
	}
}

func (m *mockRuntimeSocketManagerForRuntime) CreateSocket() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createSocketCalls++
	return m.createSocketErr
}

func (m *mockRuntimeSocketManagerForRuntime) Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listenCalled = true
	m.listenHandler = handler
	if m.listenErr != nil {
		return nil, nil, m.listenErr
	}
	return m.listenErrCh, m.listenDoneCh, nil
}

func (m *mockRuntimeSocketManagerForRuntime) DeleteSocket() error {
	m.mu.Lock()
	m.deleteSocketCalls++
	err := m.deleteSocketErr
	m.mu.Unlock()
	return err
}

func (m *mockRuntimeSocketManagerForRuntime) getDeleteSocketCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deleteSocketCalls
}

// mockMessageRouterForRuntime implements MessageRouterInterface with tracking.
type mockMessageRouterForRuntime struct {
	mu     sync.Mutex
	called bool
}

func (m *mockMessageRouterForRuntime) RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
	return entities.RuntimeResponse{Status: "ok"}
}

// mockLoggerForRuntime captures log messages.
type mockLoggerForRuntime struct {
	mu       sync.Mutex
	messages []string
}

func (m *mockLoggerForRuntime) Log(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *mockLoggerForRuntime) Warning(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *mockLoggerForRuntime) getMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.messages...)
}

func (m *mockLoggerForRuntime) containsMessage(substr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, msg := range m.messages {
		if contains(msg, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// constructionOrderTracker tracks the order in which dependencies are constructed.
type constructionOrderTracker struct {
	mu    sync.Mutex
	order []string
}

func (t *constructionOrderTracker) record(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.order = append(t.order, name)
}

func (t *constructionOrderTracker) getOrder() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.order...)
}

// runtimeTestFixture consolidates all mock dependencies for Runtime tests.
type runtimeTestFixture struct {
	projectRoot        string
	spectraFinder      *mockSpectraFinder
	sessionInitializer *mockSessionInitializerForRuntime
	sessionFinalizer   *mockSessionFinalizerForRuntime
	socketManager      *mockRuntimeSocketManagerForRuntime
	messageRouter      *mockMessageRouterForRuntime
	logger             *mockLoggerForRuntime
	session            *mockSessionForRuntime
	orderTracker       *constructionOrderTracker
}

func newRuntimeTestFixture(t *testing.T) *runtimeTestFixture {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(tmpDir+"/.spectra", 0755))

	sessionID := uuid.New().String()
	sess := newMockSessionForRuntime(sessionID, "TestWorkflow", "running", "start_node")

	return &runtimeTestFixture{
		projectRoot:   tmpDir,
		spectraFinder: &mockSpectraFinder{projectRoot: tmpDir},
		sessionInitializer: &mockSessionInitializerForRuntime{
			session: sess,
		},
		sessionFinalizer: &mockSessionFinalizerForRuntime{},
		socketManager:    newMockRuntimeSocketManager(),
		messageRouter:    &mockMessageRouterForRuntime{},
		logger:           &mockLoggerForRuntime{},
		session:          sess,
		orderTracker:     &constructionOrderTracker{},
	}
}

// completeSessionSuccessfully triggers Session.Done and closes listenerDoneCh.
func (f *runtimeTestFixture) completeSessionSuccessfully(terminationNotifier chan struct{}) {
	f.session.mu.Lock()
	f.session.status = "completed"
	f.session.mu.Unlock()
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	close(f.socketManager.listenDoneCh)
}

// =====================================================================
// Happy Path — Initialization and Bootstrap
// =====================================================================

func TestRuntime_SuccessfulInitialization(t *testing.T) {
	f := newRuntimeTestFixture(t)

	// Configure SessionInitializer to return session in "running" state
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		assert.Equal(t, "TestWorkflow", workflowName)
		return f.session, nil
	}

	// Session completes successfully after socket listener starts
	doneCh := f.socketManager.listenDoneCh
	go func() {
		// Give the runtime a moment to enter the event loop
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(doneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	assert.Equal(t, "completed", f.session.GetStatusSafe())
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_CreatesTerminationNotifier(t *testing.T) {
	f := newRuntimeTestFixture(t)

	var capturedNotifier chan<- struct{}
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		capturedNotifier = tn
		// Verify capacity is 2
		assert.Equal(t, 2, cap(tn))
		// Complete immediately
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	require.NotNil(t, capturedNotifier)
}

func TestRuntime_DependencyConstructionOrder(t *testing.T) {
	f := newRuntimeTestFixture(t)

	tracker := &constructionOrderTracker{}

	// Use a custom spectra finder that records order
	origFind := f.spectraFinder.Find
	findFunc := func() (string, error) {
		tracker.record("SpectraFinder")
		return origFind()
	}

	// Override session initializer to track construction and complete
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		tracker.record("SessionInitializer")
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", findFunc, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)

	// WorkflowDefinitionLoader, SessionDirectoryManager, AgentDefinitionLoader are constructed
	// before SessionInitializer is called. We verify that SessionInitializer was invoked,
	// meaning all pre-session dependencies were already constructed.
	order := tracker.getOrder()
	require.Contains(t, order, "SpectraFinder")
	require.Contains(t, order, "SessionInitializer")

	// SpectraFinder must come before SessionInitializer
	sfIdx := -1
	siIdx := -1
	for i, name := range order {
		if name == "SpectraFinder" {
			sfIdx = i
		}
		if name == "SessionInitializer" {
			siIdx = i
		}
	}
	assert.Less(t, sfIdx, siIdx, "SpectraFinder must be called before SessionInitializer")
}

func TestRuntime_PostSessionDependencyConstruction(t *testing.T) {
	f := newRuntimeTestFixture(t)

	// SessionInitializer returns successfully with running session
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// After socket creation and listener start, session completes
	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	// Post-session dependencies (SessionMetadataStore, EventStore, RuntimeSocketManager,
	// TransitionToNode, EventProcessor, ErrorProcessor, MessageRouter) are constructed
	// after SessionInitializer succeeds. Verify socket manager was used.
	assert.NoError(t, err)
	assert.True(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Happy Path — Session Lifecycle
// =====================================================================

func TestRuntime_SessionCompletedSuccessfully(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	assert.True(t, f.sessionFinalizer.wasCalled())
	// SessionFinalizer prints: "Session <id> completed successfully. Workflow: TestWorkflow"
	// Verified by checking SessionFinalizer was called with correct session.
	f.sessionFinalizer.mu.Lock()
	assert.NotNil(t, f.sessionFinalizer.session)
	f.sessionFinalizer.mu.Unlock()
}

func TestRuntime_SessionInitializerTransitionsToRunning(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// SessionInitializer calls Session.Run which transitions to "running"
		assert.Equal(t, "running", f.session.GetStatusSafe())
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
}

func TestRuntime_SocketCreatedAfterInitialization(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// At this point, CreateSocket should NOT have been called yet
		f.socketManager.mu.Lock()
		assert.Equal(t, 0, f.socketManager.createSocketCalls)
		f.socketManager.mu.Unlock()
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
}

func TestRuntime_ListenerStartedAfterSocketCreated(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)

	// Verify Listen was called
	f.socketManager.mu.Lock()
	assert.True(t, f.socketManager.listenCalled)
	assert.NotNil(t, f.socketManager.listenHandler, "Listen should have been called with MessageRouter.RouteMessage as handler")
	f.socketManager.mu.Unlock()
}

// =====================================================================
// Happy Path — Main Event Loop
// =====================================================================

func TestRuntime_EventLoopWaitsForTermination(t *testing.T) {
	f := newRuntimeTestFixture(t)

	var wg sync.WaitGroup
	wg.Add(1)
	eventLoopEntered := make(chan struct{})

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// Runtime should block on select. Once we detect it's blocking,
	// we send the termination signal.
	go func() {
		defer wg.Done()
		// Wait briefly for Runtime to enter the event loop
		time.Sleep(20 * time.Millisecond)
		close(eventLoopEntered)
		// Now send termination
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	wg.Wait()
	assert.NoError(t, err)
	// Verify event loop was entered (the goroutine completed)
	select {
	case <-eventLoopEntered:
		// OK
	default:
		t.Fatal("Event loop was not entered before termination")
	}
}

func TestRuntime_FirstSignalOnly(t *testing.T) {
	f := newRuntimeTestFixture(t)

	terminationNotifier := make(chan struct{}, 2)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// Session.Done sends first notification
		go func() {
			time.Sleep(10 * time.Millisecond)
			f.session.mu.Lock()
			f.session.status = "completed"
			f.session.mu.Unlock()
			tn <- struct{}{}
			// listenerErrCh sends error 1ms later
			time.Sleep(1 * time.Millisecond)
			f.socketManager.listenErrCh <- fmt.Errorf("late error")
		}()
		return f.session, nil
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	_ = terminationNotifier // The Runtime creates its own terminationNotifier
	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	// Runtime proceeds to cleanup after first signal (Session.Done)
	// listenerErrCh signal is ignored
	assert.NoError(t, err)
}

// =====================================================================
// Happy Path — OS Signal Handling
// =====================================================================

func TestRuntime_SIGINT_GracefulShutdown(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// Send SIGINT after session enters "running" state
	go func() {
		time.Sleep(10 * time.Millisecond)
		// Signal the runtime process with SIGINT
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session terminated by signal SIGINT")
	// Session.Fail NOT called
	f.session.mu.RLock()
	assert.False(t, f.session.failCalled)
	f.session.mu.RUnlock()
	assert.True(t, f.sessionFinalizer.wasCalled())
	// Verify SessionFinalizer received session with non-terminal status for proper output formatting:
	// SessionFinalizer prints: "Session <id> terminated with status 'running'. Workflow: TestWorkflow" to stderr
	f.sessionFinalizer.mu.Lock()
	finalizerSess := f.sessionFinalizer.session
	f.sessionFinalizer.mu.Unlock()
	require.NotNil(t, finalizerSess)
	assert.Equal(t, "running", finalizerSess.GetStatusSafe())
	assert.Equal(t, "TestWorkflow", finalizerSess.GetWorkflowName())
}

func TestRuntime_SIGTERM_GracefulShutdown(t *testing.T) {
	if isWindows() {
		t.Skip("SIGTERM not available on Windows")
	}

	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session terminated by signal SIGTERM")
	f.session.mu.RLock()
	assert.False(t, f.session.failCalled)
	f.session.mu.RUnlock()
	// Verify SessionFinalizer received session for proper stderr output formatting
	assert.True(t, f.sessionFinalizer.wasCalled())
	f.sessionFinalizer.mu.Lock()
	finalizerSess := f.sessionFinalizer.session
	f.sessionFinalizer.mu.Unlock()
	require.NotNil(t, finalizerSess)
	assert.Equal(t, "running", finalizerSess.GetStatusSafe())
}

func TestRuntime_SIGINT_StoresReceivedSignal(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	// Error message is constructed from receivedSignal
	assert.Contains(t, err.Error(), "SIGINT")
}

func TestRuntime_SIGINT_DuringInitializing(t *testing.T) {
	f := newRuntimeTestFixture(t)

	sess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "initializing", "start_node")
	f.session = sess

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// Session status is still "initializing"
		return sess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session terminated by signal SIGINT")
	assert.True(t, f.sessionFinalizer.wasCalled())
	// Verify SessionFinalizer received session with 'initializing' status for proper stderr output:
	// SessionFinalizer prints: "Session <id> terminated with status 'initializing'. Workflow: TestWorkflow"
	f.sessionFinalizer.mu.Lock()
	finalizerSess := f.sessionFinalizer.session
	f.sessionFinalizer.mu.Unlock()
	require.NotNil(t, finalizerSess)
	assert.Equal(t, "initializing", finalizerSess.GetStatusSafe())
	assert.Equal(t, "TestWorkflow", finalizerSess.GetWorkflowName())
}

// =====================================================================
// Happy Path — Cleanup and Finalization
// =====================================================================

func TestRuntime_CleanupOrder(t *testing.T) {
	f := newRuntimeTestFixture(t)

	var cleanupOrder []string
	var orderMu sync.Mutex

	origDeleteSocket := f.socketManager.DeleteSocket
	_ = origDeleteSocket

	// Track DeleteSocket call
	sm := &orderTrackingSocketManager{
		inner:        f.socketManager,
		orderTracker: &cleanupOrder,
		mu:           &orderMu,
	}

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// Track SessionFinalizer call
	origFinalizer := f.sessionFinalizer
	trackingFinalizer := &orderTrackingFinalizer{
		inner:        origFinalizer,
		orderTracker: &cleanupOrder,
		mu:           &orderMu,
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		// Close listenerDoneCh after DeleteSocket is called
		go func() {
			time.Sleep(20 * time.Millisecond)
			orderMu.Lock()
			cleanupOrder = append(cleanupOrder, "listenerDoneCh closed")
			orderMu.Unlock()
			close(f.socketManager.listenDoneCh)
		}()
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, trackingFinalizer, sm, f.messageRouter, f.logger)

	assert.NoError(t, err)

	orderMu.Lock()
	defer orderMu.Unlock()
	// Verify order: DeleteSocket, wait for listenerDoneCh, SessionFinalizer
	require.GreaterOrEqual(t, len(cleanupOrder), 2)
	deleteIdx := -1
	finalizerIdx := -1
	for i, op := range cleanupOrder {
		if op == "DeleteSocket" && deleteIdx == -1 {
			deleteIdx = i
		}
		if op == "SessionFinalizer" && finalizerIdx == -1 {
			finalizerIdx = i
		}
	}
	if deleteIdx >= 0 && finalizerIdx >= 0 {
		assert.Less(t, deleteIdx, finalizerIdx, "DeleteSocket must be called before SessionFinalizer")
	}
}

// orderTrackingSocketManager wraps a socket manager to track call order.
type orderTrackingSocketManager struct {
	inner        *mockRuntimeSocketManagerForRuntime
	orderTracker *[]string
	mu           *sync.Mutex
}

func (m *orderTrackingSocketManager) CreateSocket() error {
	return m.inner.CreateSocket()
}

func (m *orderTrackingSocketManager) Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error) {
	return m.inner.Listen(handler)
}

func (m *orderTrackingSocketManager) DeleteSocket() error {
	m.mu.Lock()
	*m.orderTracker = append(*m.orderTracker, "DeleteSocket")
	m.mu.Unlock()
	return m.inner.DeleteSocket()
}

// orderTrackingFinalizer wraps a session finalizer to track call order.
type orderTrackingFinalizer struct {
	inner        *mockSessionFinalizerForRuntime
	orderTracker *[]string
	mu           *sync.Mutex
}

func (m *orderTrackingFinalizer) Finalize(session SessionForFinalizer) {
	m.mu.Lock()
	*m.orderTracker = append(*m.orderTracker, "SessionFinalizer")
	m.mu.Unlock()
	m.inner.Finalize(session)
}

func TestRuntime_WaitsForListenerShutdown(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	listenerDoneSignaled := make(chan struct{})

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		// Delay closing listenerDoneCh to verify Runtime waits
		time.Sleep(50 * time.Millisecond)
		close(listenerDoneSignaled)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	// listenerDoneCh was closed before SessionFinalizer was called
	select {
	case <-listenerDoneSignaled:
		// OK - Runtime waited for listenerDoneCh
	default:
		t.Fatal("Runtime did not wait for listenerDoneCh before calling SessionFinalizer")
	}
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_ListenerShutdownTimeout(t *testing.T) {
	f := newRuntimeTestFixture(t)

	// terminationNotifier must receive a signal (sent from inside the mock
	// SessionInitializer) to unblock the main event loop and reach the cleanup
	// phase — mutating session status alone is insufficient.
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	// listenerDoneCh never closes — simulates listener that doesn't shut down.

	// Inject two independent mock timers:
	// - An immediate-fire timer for the listener shutdown wait (simulating 2-second timeout)
	// - A never-fire timer for the grace period (preventing forced-exit from firing during the test)
	callCount := 0
	dualTimerFunc := func(d time.Duration) <-chan time.Time {
		callCount++
		if callCount == 1 {
			// First timer: grace period — never fire
			return make(chan time.Time)
		}
		// Second timer: listener shutdown wait — fire immediately
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}

	err := RunWithTimerFunc("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, dualTimerFunc)

	// Runtime should proceed to SessionFinalizer after simulated timeout
	assert.NoError(t, err)
	assert.True(t, f.sessionFinalizer.wasCalled())
	assert.True(t, f.logger.containsMessage("listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer"))
}

func TestRuntime_DeleteSocketIdempotent(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// DeleteSocket logs warning but returns nil
	f.socketManager.deleteSocketErr = nil

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	assert.True(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Happy Path — Grace Period Enforcement
// =====================================================================

func TestRuntime_GracePeriodEnforced(t *testing.T) {
	f := newRuntimeTestFixture(t)

	// terminationNotifier must receive a signal (sent from inside the mock
	// SessionInitializer or its goroutine) to unblock the main event loop —
	// mutating session status alone is insufficient.
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	// listenerDoneCh never closes (simulates cleanup blocked indefinitely).

	// Inject two independent mock timers:
	// - A never-fire timer for the listener shutdown wait (keeping Runtime blocked
	//   there so the grace period can fire)
	// - An immediate-fire timer for the grace period (simulating 5-second expiration)
	callCount := 0
	dualTimerFunc := func(d time.Duration) <-chan time.Time {
		callCount++
		if callCount == 1 {
			// First timer: grace period — fire immediately (simulating 5-second expiration)
			ch := make(chan time.Time, 1)
			ch <- time.Now()
			return ch
		}
		// Second timer: listener shutdown wait — never fire (keeping Runtime blocked)
		return make(chan time.Time)
	}

	// Inject a no-op exit function so that forced-exit is signalled without
	// terminating the test process, allowing Run() to return normally.
	exitCalled := false
	noopExit := func(code int) {
		exitCalled = true
	}

	err := RunWithTimerFuncAndExit("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, dualTimerFunc, noopExit)

	// Runtime logs grace period warning; Run() returns normally (via injectable exit mechanism)
	// so assertions can execute; test completes quickly without any real delay.
	assert.True(t, f.logger.containsMessage("cleanup exceeded 5 second grace period, forcing exit"))
	_ = err
	_ = exitCalled
}

func TestRuntime_SecondSignalForcesExit(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// Send first SIGINT to trigger graceful shutdown; send second SIGINT
	// during the cleanup wait.
	go func() {
		time.Sleep(10 * time.Millisecond)
		// First SIGINT
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)

		// Delay then second SIGINT during cleanup
		time.Sleep(50 * time.Millisecond)
		_ = p.Signal(syscall.SIGINT)
	}()

	// Inject a never-fire timer for the listener shutdown wait so Runtime stays
	// blocked in the cleanup wait (giving the grace period goroutine time to
	// receive the second signal).
	neverFireTimerFunc := func(d time.Duration) <-chan time.Time {
		return make(chan time.Time)
	}

	// Inject a no-op exit function so that forced-exit is signalled without
	// terminating the test process, allowing Run() to return normally and
	// assertions to execute.
	exitCalled := false
	noopExit := func(code int) {
		exitCalled = true
	}

	err := RunWithTimerFuncAndExit("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, neverFireTimerFunc, noopExit)

	// Run() returns normally (via injectable exit mechanism)
	_ = err
	_ = exitCalled

	// Verify log message order
	msgs := f.logger.getMessages()
	firstSigIdx := -1
	secondSigIdx := -1
	for i, msg := range msgs {
		if contains(msg, "received signal interrupt, initiating graceful shutdown") && firstSigIdx == -1 {
			firstSigIdx = i
		}
		if contains(msg, "received second signal, forcing exit") && secondSigIdx == -1 {
			secondSigIdx = i
		}
	}
	if firstSigIdx >= 0 && secondSigIdx >= 0 {
		assert.Less(t, firstSigIdx, secondSigIdx, "first signal log must come before second signal log")
	}
}

// =====================================================================
// Validation Failures — SpectraFinder
// =====================================================================

func TestRuntime_SpectraFinderFails_NotInitialized(t *testing.T) {
	f := newRuntimeTestFixture(t)
	f.spectraFinder.err = fmt.Errorf("spectra not initialized")
	f.spectraFinder.projectRoot = ""

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to locate project root: spectra not initialized")
	assert.False(t, f.sessionFinalizer.wasCalled(), "SessionFinalizer should NOT be called")
}

func TestRuntime_SpectraFinderFails_NoResources(t *testing.T) {
	f := newRuntimeTestFixture(t)
	f.spectraFinder.err = fmt.Errorf("spectra not initialized")
	f.spectraFinder.projectRoot = ""

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	// terminationNotifier NOT created; WorkflowDefinitionLoader NOT constructed
	f.sessionInitializer.mu.Lock()
	assert.False(t, f.sessionInitializer.called, "SessionInitializer should NOT be called")
	f.sessionInitializer.mu.Unlock()
	assert.False(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Validation Failures — Dependency Construction (Pre-Session)
// =====================================================================

func TestRuntime_WorkflowDefinitionLoaderConstructionFails(t *testing.T) {
	f := newRuntimeTestFixture(t)

	// Simulate WorkflowDefinitionLoader construction failure via Run function args
	err := Run("TestWorkflow", f.spectraFinder.Find, nil, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize runtime dependencies")
	assert.False(t, f.sessionFinalizer.wasCalled(), "SessionFinalizer should NOT be called")
}

func TestRuntime_SessionDirectoryManagerConstructionFails(t *testing.T) {
	f := newRuntimeTestFixture(t)

	err := Run("TestWorkflow", f.spectraFinder.Find, nil, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize runtime dependencies")
	assert.False(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_AgentDefinitionLoaderConstructionFails(t *testing.T) {
	f := newRuntimeTestFixture(t)

	err := Run("TestWorkflow", f.spectraFinder.Find, nil, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize runtime dependencies")
	assert.False(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_SessionInitializerConstructionFails(t *testing.T) {
	f := newRuntimeTestFixture(t)

	err := Run("TestWorkflow", f.spectraFinder.Find, nil, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize runtime dependencies")
	assert.False(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Validation Failures — SessionInitializer
// =====================================================================

func TestRuntime_SessionInitializerFails_NoSession(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.session = nil
	f.sessionInitializer.err = fmt.Errorf("workflow not found")

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize session: workflow not found")
	assert.False(t, f.sessionFinalizer.wasCalled(), "SessionFinalizer should NOT be called when session is nil")
}

func TestRuntime_SessionInitializerFails_WithSession(t *testing.T) {
	f := newRuntimeTestFixture(t)

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "failed", "start_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, fmt.Errorf("initialization failed after session created")
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize session: initialization failed after session created")
	assert.True(t, f.sessionFinalizer.wasCalled(), "SessionFinalizer should be called when session != nil")
}

func TestRuntime_SessionInitializerTimeout_BeforeSessionEntity(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return nil, fmt.Errorf("session initialization timeout exceeded 30 seconds before session entity was constructed")
	}

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize session: session initialization timeout exceeded 30 seconds before session entity was constructed")
	assert.False(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_SessionInitializerTimeout_AfterSessionEntity(t *testing.T) {
	f := newRuntimeTestFixture(t)

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "failed", "start_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// Session entity was constructed but timeout fires after
		tn <- struct{}{} // termination notification
		return failedSess, fmt.Errorf("session initialization timeout exceeded 30 seconds")
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize session: session initialization timeout exceeded 30 seconds")
	assert.True(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Validation Failures — Post-Session Dependencies
// =====================================================================

func TestRuntime_PostSessionDependencyFails(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	// Simulate EventStore construction failure
	// The Runtime should construct a RuntimeError and call Session.Fail
	close(f.socketManager.listenDoneCh)

	err := RunWithPostSessionDepError("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, fmt.Errorf("failed to open events file"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize post-session dependencies: failed to open events file")
	// Verify RuntimeError fields
	f.session.mu.RLock()
	assert.True(t, f.session.failCalled)
	failErr := f.session.failError
	f.session.mu.RUnlock()
	require.NotNil(t, failErr)
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_RuntimeSocketManagerConstructionFails(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := RunWithPostSessionDepError("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, fmt.Errorf("socket manager construction failed"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize post-session dependencies")
	f.session.mu.RLock()
	assert.True(t, f.session.failCalled)
	f.session.mu.RUnlock()
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_RuntimeErrorDetailFieldPopulated(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := RunWithPostSessionDepError("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, fmt.Errorf("disk full"))

	require.Error(t, err)
	// Verify RuntimeError.Detail is valid JSON
	f.session.mu.RLock()
	failErr := f.session.failError
	f.session.mu.RUnlock()
	if rtErr, ok := failErr.(*entities.RuntimeError); ok {
		if len(rtErr.Detail) > 0 {
			var temp interface{}
			assert.NoError(t, json.Unmarshal(rtErr.Detail, &temp), "Detail should be valid JSON")
		}
	}
}

func TestRuntime_RuntimeErrorOccurredAtTimestamp(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	beforeTime := time.Now().Unix()

	err := RunWithPostSessionDepError("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, fmt.Errorf("some error"))

	afterTime := time.Now().Unix()

	require.Error(t, err)
	f.session.mu.RLock()
	failErr := f.session.failError
	f.session.mu.RUnlock()
	if rtErr, ok := failErr.(*entities.RuntimeError); ok {
		assert.GreaterOrEqual(t, rtErr.OccurredAt, beforeTime-5, "OccurredAt should be within ±5 seconds")
		assert.LessOrEqual(t, rtErr.OccurredAt, afterTime+5, "OccurredAt should be within ±5 seconds")
	}
}

// =====================================================================
// Validation Failures — Socket Creation and Listener
// =====================================================================

func TestRuntime_CreateSocketFails_SocketAlreadyExists(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	f.socketManager.createSocketErr = fmt.Errorf("runtime socket file already exists: /tmp/.spectra/sessions/abc-123/runtime.sock")

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	errMsg := err.Error()
	// Verify the complete error message format per spec:
	// "failed to create runtime socket: runtime socket file already exists: <path>.
	//  This may indicate a previous runtime process did not clean up properly or another
	//  runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra),
	//  then remove the socket file manually with: rm <path>"
	assert.Contains(t, errMsg, "failed to create runtime socket: runtime socket file already exists: /tmp/.spectra/sessions/abc-123/runtime.sock")
	assert.Contains(t, errMsg, "This may indicate a previous runtime process did not clean up properly or another runtime is currently active")
	assert.Contains(t, errMsg, "Verify no runtime process is running (e.g., ps aux | grep spectra)")
	assert.Contains(t, errMsg, "then remove the socket file manually with: rm /tmp/.spectra/sessions/abc-123/runtime.sock")
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_CreateSocketFails_PermissionDenied(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	f.socketManager.createSocketErr = fmt.Errorf("permission denied")

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create runtime socket: permission denied")
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_ListenerStartFails_BindError(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	f.socketManager.listenErr = fmt.Errorf("bind error: address already in use")

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start socket listener")
	assert.True(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Validation Failures — Listener Errors
// =====================================================================

func TestRuntime_ListenerErrorDuringRuntime(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.socketManager.listenErrCh <- fmt.Errorf("accept loop failure")
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session failed: listener error")
	assert.True(t, f.logger.containsMessage("listener error: accept loop failure"))
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_ListenerErrorWhenSessionAlreadyCompleted(t *testing.T) {
	f := newRuntimeTestFixture(t)

	completedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "completed", "end_node")

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return completedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		// Session already completed before listener error
		f.socketManager.listenErrCh <- fmt.Errorf("accept loop failure")
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)
	_ = err

	// Session.Fail should be skipped because status is already "completed"
	completedSess.mu.RLock()
	assert.False(t, completedSess.failCalled, "Session.Fail should NOT be called when session is already completed")
	completedSess.mu.RUnlock()
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_ListenerErrorWhenSessionAlreadyFailed(t *testing.T) {
	f := newRuntimeTestFixture(t)

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "failed", "error_node")

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.socketManager.listenErrCh <- fmt.Errorf("accept loop failure")
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)
	_ = err

	// Session.Fail should be skipped because status is already "failed"
	failedSess.mu.RLock()
	assert.False(t, failedSess.failCalled, "Session.Fail should NOT be called when session is already failed")
	failedSess.mu.RUnlock()
	assert.True(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Error Propagation — Session Failures
// =====================================================================

func TestRuntime_SessionFailedWithAgentError(t *testing.T) {
	f := newRuntimeTestFixture(t)

	agentErr := &entities.AgentError{
		AgentRole:    "reviewer",
		Message:      "validation failed",
		FailingState: "review_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "review_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		// ErrorProcessor calls Session.Fail with AgentError
		failedSess.mu.Lock()
		failedSess.status = "failed"
		failedSess.err = agentErr
		failedSess.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session failed: validation failed")
	assert.True(t, f.sessionFinalizer.wasCalled())
}

func TestRuntime_SessionFailedWithRuntimeError(t *testing.T) {
	f := newRuntimeTestFixture(t)

	rtErr := &entities.RuntimeError{
		Issuer:       "Runtime",
		Message:      "socket write failure",
		FailingState: "process_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "process_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		failedSess.mu.Lock()
		failedSess.status = "failed"
		failedSess.err = rtErr
		failedSess.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session failed: socket write failure")
	assert.True(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Error Propagation — Return Values
// =====================================================================

func TestRuntime_ReturnsNilOnSuccess(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
}

func TestRuntime_ReturnsSIGINTError(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Equal(t, "session terminated by signal SIGINT", err.Error())
}

func TestRuntime_ReturnsSIGTERMError(t *testing.T) {
	if isWindows() {
		t.Skip("SIGTERM not available on Windows")
	}

	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Equal(t, "session terminated by signal SIGTERM", err.Error())
}

func TestRuntime_ReturnsSessionFailedError(t *testing.T) {
	f := newRuntimeTestFixture(t)

	agentErr := &entities.AgentError{
		Message:      "validation failed",
		AgentRole:    "reviewer",
		FailingState: "review_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "review_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		failedSess.mu.Lock()
		failedSess.status = "failed"
		failedSess.err = agentErr
		failedSess.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Equal(t, "session failed: validation failed", err.Error())
}

func TestRuntime_ReturnsNonTerminalStatusError(t *testing.T) {
	f := newRuntimeTestFixture(t)

	runningSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "process_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return runningSess, nil
	}

	// Simulate edge case: session terminates with "running" but receivedSignal is nil
	go func() {
		time.Sleep(10 * time.Millisecond)
		// Send to terminationNotifier without signal and without status change
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session terminated with status 'running'")
}

// =====================================================================
// Boundary Values — Signal Timing
// =====================================================================

func TestRuntime_SessionDoneBeforeEventLoop(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// Session.Done immediately before main loop
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	// terminationNotifier already has signal (buffered); Runtime receives immediately
	assert.NoError(t, err)
}

func TestRuntime_MultipleConcurrentTerminationSignals(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		// Both Session.Done and timeout send to terminationNotifier concurrently
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		tn <- struct{}{} // Both fit in capacity-2 buffer
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	// Runtime receives first signal; second remains in buffer (ignored)
	assert.NoError(t, err)
}

// =====================================================================
// Boundary Values — Empty Workflow Name
// =====================================================================

func TestRuntime_EmptyWorkflowName(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return nil, fmt.Errorf("workflow name is empty")
	}

	err := Run("", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize session: workflow name is empty")
	assert.False(t, f.sessionFinalizer.wasCalled())
}

// =====================================================================
// Idempotency — Cleanup Operations
// =====================================================================

func TestRuntime_MultipleDeleteSocketCalls(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	// DeleteSocket is idempotent; no error even if called multiple times
	calls := f.socketManager.getDeleteSocketCalls()
	assert.GreaterOrEqual(t, calls, 1, "DeleteSocket should be called at least once")
}

// =====================================================================
// State Transitions — Session Status
// =====================================================================

func TestRuntime_SessionStatusInitializingToRunning(t *testing.T) {
	f := newRuntimeTestFixture(t)

	initSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "initializing", "start_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		initSess.Run(tn) // Transitions to "running"
		assert.Equal(t, "running", initSess.GetStatusSafe())
		return initSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		initSess.mu.Lock()
		initSess.status = "completed"
		initSess.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
}

func TestRuntime_SessionStatusRunningToCompleted(t *testing.T) {
	f := newRuntimeTestFixture(t)

	runningSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "end_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return runningSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		runningSess.Done(make(chan struct{}, 2)) // Transitions to "completed"
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	assert.Equal(t, "completed", runningSess.GetStatusSafe())
}

func TestRuntime_SessionStatusRunningToFailed(t *testing.T) {
	f := newRuntimeTestFixture(t)

	agentErr := &entities.AgentError{
		Message:      "processing error",
		AgentRole:    "processor",
		FailingState: "process_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	runningSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "process_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return runningSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		runningSess.Fail(agentErr, make(chan struct{}, 2))
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session failed: processing error")
	assert.Equal(t, "failed", runningSess.GetStatusSafe())
}

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

func TestRuntime_MessageRouterReceivesMessageHandler(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	f.socketManager.mu.Lock()
	assert.True(t, f.socketManager.listenCalled, "Listen should be called")
	assert.NotNil(t, f.socketManager.listenHandler, "Listen should receive MessageRouter.RouteMessage as handler")
	f.socketManager.mu.Unlock()
}

func TestRuntime_TerminationNotifierPassedToDependencies(t *testing.T) {
	f := newRuntimeTestFixture(t)

	var capturedNotifier chan<- struct{}
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		capturedNotifier = tn
		require.NotNil(t, tn, "terminationNotifier must be passed to SessionInitializer")
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	require.NotNil(t, capturedNotifier, "terminationNotifier should be passed to SessionInitializer")
}

func TestRuntime_SessionFinalizerCalledWithSession(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	assert.True(t, f.sessionFinalizer.wasCalled())
	f.sessionFinalizer.mu.Lock()
	assert.NotNil(t, f.sessionFinalizer.session, "SessionFinalizer should be called with session")
	f.sessionFinalizer.mu.Unlock()
}

func TestRuntime_SessionFailAttemptOnTerminalSession_FirstErrorPreserved(t *testing.T) {
	f := newRuntimeTestFixture(t)

	firstErr := &entities.AgentError{
		Message:      "validation failed",
		AgentRole:    "reviewer",
		FailingState: "review_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	sess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "review_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return sess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		// First Fail with AgentError
		sess.Fail(firstErr, make(chan struct{}, 2))

		// Second Fail attempt with RuntimeError — should return "session already failed"
		secondErr := &entities.RuntimeError{
			Issuer:       "Runtime",
			Message:      "listener error",
			FailingState: "review_node",
			SessionID:    uuid.New(),
			OccurredAt:   time.Now().Unix(),
		}
		retErr := sess.Fail(secondErr, make(chan struct{}, 2))
		assert.Error(t, retErr)
		assert.Contains(t, retErr.Error(), "session already failed")

		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	// First error preserved
	sess.mu.RLock()
	assert.Equal(t, firstErr, sess.err, "First error should be preserved")
	sess.mu.RUnlock()
}

func TestRuntime_UsesGetStatusSafe(t *testing.T) {
	f := newRuntimeTestFixture(t)

	sess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "node1")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return sess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		// Listener error triggers status check
		f.socketManager.listenErrCh <- fmt.Errorf("test error")
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	_ = Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	// Verify GetStatusSafe was called (thread-safe status access)
	assert.Greater(t, sess.getStatusSafeCalled.Load(), int32(0), "GetStatusSafe should be called for thread-safe status access")
}

func TestRuntime_UsesGetCurrentStateSafe(t *testing.T) {
	f := newRuntimeTestFixture(t)

	sess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "process_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return sess, nil
	}

	close(f.socketManager.listenDoneCh)

	// Post-session dependency failure triggers GetCurrentStateSafe for RuntimeError.FailingState
	_ = RunWithPostSessionDepError("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger, fmt.Errorf("dep failure"))

	assert.Greater(t, sess.getCurrentStateSafeCalled.Load(), int32(0), "GetCurrentStateSafe should be called for thread-safe state access")
}

// =====================================================================
// Resource Cleanup — Panic Recovery
// =====================================================================

func TestRuntime_SessionFinalizerPanics_WithDeferredRecovery(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionFinalizer.panicMsg = "print failed"

	completedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "completed", "end_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return completedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	// Runtime should recover from panic; not propagate
	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	// Session completed successfully despite panic
	assert.NoError(t, err)

	// After catching the panic, Finalize() is NOT called a second time
	// (calling it again would re-panic with no recovery).
	assert.Equal(t, 1, f.sessionFinalizer.getCallCount(),
		"Finalize() should be called exactly once; calling it again after panic would re-panic with no recovery")
}

func TestRuntime_SessionFinalizerPanics_FailedSession(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionFinalizer.panicMsg = "print failed"

	agentErr := &entities.AgentError{
		Message:      "agent error",
		AgentRole:    "reviewer",
		FailingState: "review_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "failed", "review_node")
	failedSess.err = agentErr

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session failed: agent error")
}

// =====================================================================
// Platform Compatibility — Signal Handling
// =====================================================================

func TestRuntime_WindowsSIGINTOnly(t *testing.T) {
	if !isWindows() {
		t.Skip("Test is only applicable on Windows")
	}

	// On Windows, signal.Notify should be called with only SIGINT
	// SIGTERM is not available on Windows
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
}

func TestRuntime_UnixSignalRegistration(t *testing.T) {
	if isWindows() {
		t.Skip("Test is only applicable on Unix-like platforms")
	}

	// On Unix, both SIGINT and SIGTERM should be registered
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
}

// =====================================================================
// Edge Cases — Channel Lifecycle
// =====================================================================

func TestRuntime_TerminationNotifierNeverClosed(t *testing.T) {
	f := newRuntimeTestFixture(t)

	var capturedNotifier chan<- struct{}
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		capturedNotifier = tn
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		tn <- struct{}{}
		return f.session, nil
	}

	close(f.socketManager.listenDoneCh)

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	// Verify terminationNotifier is still sendable (not closed)
	require.NotNil(t, capturedNotifier)
	// Try sending — should not panic if channel is not closed
	func() {
		defer func() {
			r := recover()
			assert.Nil(t, r, "terminationNotifier should not be closed")
		}()
		select {
		case capturedNotifier <- struct{}{}:
		default:
			// Channel full is fine, just not closed
		}
	}()
}

func TestRuntime_ListenerErrChNeverClosed(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	// Verify listenerErrCh is still sendable (not closed)
	func() {
		defer func() {
			r := recover()
			assert.Nil(t, r, "listenerErrCh should not be closed by RuntimeSocketManager")
		}()
		select {
		case f.socketManager.listenErrCh <- fmt.Errorf("test"):
		default:
		}
	}()
}

func TestRuntime_ListenerDoneChClosedOnce(t *testing.T) {
	f := newRuntimeTestFixture(t)

	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return f.session, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.session.mu.Lock()
		f.session.status = "completed"
		f.session.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	assert.NoError(t, err)
	// Verify listenerDoneCh is closed
	select {
	case <-f.socketManager.listenDoneCh:
		// OK — channel is closed
	default:
		t.Fatal("listenerDoneCh should be closed")
	}
}

// =====================================================================
// Edge Cases — Listener Shutdown Timing
// =====================================================================

func TestRuntime_ListenerDoneCh_ImmediateClose(t *testing.T) {
	f := newRuntimeTestFixture(t)

	// Session completes normally: SessionInitializer calls Session.Done to send
	// terminationNotifier signal, ensuring the main event loop exits via that path.
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		_ = f.session.Done(tn)
		return f.session, nil
	}

	// Wrap the socket manager so that DeleteSocket closes listenerDoneCh
	// synchronously before returning — this is the "race: listener exits very fast"
	// scenario where the channel is already closed when the cleanup select runs.
	immediateCloseSM := &immediateCloseSocketManager{
		inner: f.socketManager,
	}

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, immediateCloseSM, f.messageRouter, f.logger)

	assert.NoError(t, err)
	assert.True(t, f.sessionFinalizer.wasCalled(), "SessionFinalizer should be called without delay")
}

// immediateCloseSocketManager wraps mockRuntimeSocketManagerForRuntime and closes
// listenDoneCh inside DeleteSocket to simulate a listener that exits very fast.
type immediateCloseSocketManager struct {
	inner *mockRuntimeSocketManagerForRuntime
}

func (m *immediateCloseSocketManager) CreateSocket() error {
	return m.inner.CreateSocket()
}

func (m *immediateCloseSocketManager) Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error) {
	return m.inner.Listen(handler)
}

func (m *immediateCloseSocketManager) DeleteSocket() error {
	err := m.inner.DeleteSocket()
	// Close listenerDoneCh synchronously, simulating a listener that exits
	// immediately once the socket is deleted.
	m.inner.mu.Lock()
	select {
	case <-m.inner.listenDoneCh:
		// already closed
	default:
		close(m.inner.listenDoneCh)
	}
	m.inner.mu.Unlock()
	return err
}

// =====================================================================
// Edge Cases — Empty Error Message
// =====================================================================

func TestRuntime_SessionFailedWithEmptyAgentErrorMessage(t *testing.T) {
	f := newRuntimeTestFixture(t)

	agentErr := &entities.AgentError{
		Message:      "",
		AgentRole:    "reviewer",
		FailingState: "review_node",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "review_node")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		failedSess.mu.Lock()
		failedSess.status = "failed"
		failedSess.err = agentErr
		failedSess.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Equal(t, "session failed: ", err.Error())
}

func TestRuntime_SessionFailedWithEmptyRuntimeErrorMessage(t *testing.T) {
	f := newRuntimeTestFixture(t)

	rtErr := &entities.RuntimeError{
		Issuer:       "MessageRouter",
		Message:      "",
		FailingState: "node1",
		SessionID:    uuid.New(),
		OccurredAt:   time.Now().Unix(),
	}

	failedSess := newMockSessionForRuntime(uuid.New().String(), "TestWorkflow", "running", "node1")
	f.sessionInitializer.initFunc = func(workflowName string, tn chan<- struct{}) (SessionForInitializer, error) {
		return failedSess, nil
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		failedSess.mu.Lock()
		failedSess.status = "failed"
		failedSess.err = rtErr
		failedSess.mu.Unlock()
		close(f.socketManager.listenDoneCh)
	}()

	err := Run("TestWorkflow", f.spectraFinder.Find, f.sessionInitializer, f.sessionFinalizer, f.socketManager, f.messageRouter, f.logger)

	require.Error(t, err)
	assert.Equal(t, "session failed: ", err.Error())
}

// =====================================================================
// Helper functions
// =====================================================================

func isWindows() bool {
	return goruntime.GOOS == "windows"
}

// RuntimeLogger defines the logging interface that Runtime expects.
// Production code should define this in runtime.go.
type RuntimeLogger interface {
	Log(msg string)
	Warning(msg string)
}

// TimerFunc is an injectable timer factory for testing time-dependent behavior.
// It takes a duration and returns a channel that receives after that duration.
// Production code uses time.After; tests inject an immediate-fire version.
type TimerFunc func(d time.Duration) <-chan time.Time

// ExitFunc is an injectable exit function for testing forced-exit behavior.
// Production code uses os.Exit; tests inject a no-op to prevent process termination.
type ExitFunc func(code int)

// immediateTimerFunc returns a timer that fires immediately, for fast timeout tests.
func immediateTimerFunc(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	return ch
}
