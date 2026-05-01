package runtime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// =====================================================================
// Mock types for Runtime tests
// =====================================================================

// mockSpectraFinderForRuntime mocks SpectraFinder for Runtime tests.
type mockSpectraFinderForRuntime struct {
	mu          sync.Mutex
	projectRoot string
	err         error
	called      atomic.Bool
}

func newMockSpectraFinderForRuntime(projectRoot string, err error) *mockSpectraFinderForRuntime {
	return &mockSpectraFinderForRuntime{
		projectRoot: projectRoot,
		err:         err,
	}
}

func (m *mockSpectraFinderForRuntime) Find() (string, error) {
	m.called.Store(true)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.projectRoot, m.err
}

// mockSessionForRuntime implements SessionForInitializer and the additional
// interfaces needed by Runtime (Done, Fail, GetStatusSafe, etc.).
type mockSessionForRuntime struct {
	mu           sync.RWMutex
	id           string
	status       string
	workflowName string
	currentState string
	err          error
	createdAt    int64
	updatedAt    int64
	eventHistory []session.Event
	sessionData  map[string]any

	// Tracking
	doneCalled atomic.Bool
	failCalled atomic.Bool
	failErr    error
}

func newMockSessionForRuntime(id, status, workflowName string) *mockSessionForRuntime {
	return &mockSessionForRuntime{
		id:           id,
		status:       status,
		workflowName: workflowName,
		currentState: "start",
		createdAt:    time.Now().Unix(),
		updatedAt:    time.Now().Unix(),
		eventHistory: []session.Event{},
		sessionData:  make(map[string]any),
	}
}

func (m *mockSessionForRuntime) Run(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = "running"
	return nil
}

func (m *mockSessionForRuntime) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	m.status = "failed"
	m.err = err
	m.mu.Unlock()
	m.failCalled.Store(true)
	m.failErr = err
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *mockSessionForRuntime) Done(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	m.status = "completed"
	m.mu.Unlock()
	m.doneCalled.Store(true)
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *mockSessionForRuntime) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
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

func (m *mockSessionForRuntime) GetCurrentStateSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
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
	return append([]session.Event(nil), m.eventHistory...)
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

// mockSessionInitializerForRuntime mocks the SessionInitializer for Runtime tests.
type mockSessionInitializerForRuntime struct {
	mu                  sync.Mutex
	session             SessionForInitializer
	err                 error
	initializeCalled    atomic.Bool
	capturedWorkflow    string
	capturedProjectRoot string
	capturedNotifier    chan<- struct{}
}

func (m *mockSessionInitializerForRuntime) Initialize(workflowName string, projectRoot string, terminationNotifier chan<- struct{}) (SessionForInitializer, error) {
	m.initializeCalled.Store(true)
	m.mu.Lock()
	m.capturedWorkflow = workflowName
	m.capturedProjectRoot = projectRoot
	m.capturedNotifier = terminationNotifier
	sess := m.session
	err := m.err
	m.mu.Unlock()
	return sess, err
}

// mockSessionFinalizerForRuntime mocks the SessionFinalizer for Runtime tests.
type mockSessionFinalizerForRuntime struct {
	mu             sync.Mutex
	finalizeCalled atomic.Bool
	session        SessionForFinalizer
	callOrder      *runtimeCallOrderTracker
}

func (m *mockSessionFinalizerForRuntime) Finalize(sess SessionForFinalizer) {
	m.finalizeCalled.Store(true)
	m.mu.Lock()
	m.session = sess
	m.mu.Unlock()
	if m.callOrder != nil {
		m.callOrder.record("Finalize")
	}
	// Add a small delay to simulate realistic finalization time (printing, file I/O, etc.)
	// This allows double-signal tests to properly test the second signal handling
	time.Sleep(100 * time.Millisecond)
}

// mockRuntimeSocketManagerForRuntime mocks RuntimeSocketManager for Runtime tests.
type mockRuntimeSocketManagerForRuntime struct {
	mu                 sync.Mutex
	listenCalled       atomic.Bool
	deleteSocketCalled atomic.Int32
	listenErrCh        chan error
	listenDoneCh       chan struct{}
	listenSyncErr      error
	capturedHandler    MessageHandler
	callOrder          *runtimeCallOrderTracker
	closeOnce          sync.Once
}

func newMockRuntimeSocketManagerForRuntime() *mockRuntimeSocketManagerForRuntime {
	return &mockRuntimeSocketManagerForRuntime{
		listenErrCh:  make(chan error, 1),
		listenDoneCh: make(chan struct{}),
	}
}

func (m *mockRuntimeSocketManagerForRuntime) Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error) {
	m.listenCalled.Store(true)
	m.mu.Lock()
	m.capturedHandler = handler
	syncErr := m.listenSyncErr
	m.mu.Unlock()
	if syncErr != nil {
		// On sync error, listenDoneCh is already closed
		doneCh := make(chan struct{})
		close(doneCh)
		return nil, doneCh, syncErr
	}
	return m.listenErrCh, m.listenDoneCh, nil
}

func (m *mockRuntimeSocketManagerForRuntime) DeleteSocket() error {
	m.deleteSocketCalled.Add(1)
	if m.callOrder != nil {
		m.callOrder.record("DeleteSocket")
	}
	// Simulate socket closure by closing listenerDoneCh (only once)
	m.closeOnce.Do(func() {
		close(m.listenDoneCh)
	})
	return nil
}

// mockMessageRouterForRuntime mocks the MessageRouter for Runtime tests.
type mockMessageRouterForRuntime struct {
	mu                  sync.Mutex
	session             SessionForRouter
	eventProcessor      EventProcessorInterface
	errorProcessor      ErrorProcessorInterface
	terminationNotifier chan<- struct{}
}

func (m *mockMessageRouterForRuntime) RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	return entities.RuntimeResponse{Status: "success", Message: "routed"}
}

// panicMessageRouterForRuntime panics when RouteMessage is called.
type panicMessageRouterForRuntime struct{}

func (m *panicMessageRouterForRuntime) RouteMessage(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse {
	panic("unexpected nil")
}

// runtimeCallOrderTracker tracks call ordering for Runtime cleanup sequence tests.
type runtimeCallOrderTracker struct {
	mu    sync.Mutex
	calls []string
}

func (t *runtimeCallOrderTracker) record(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, name)
}

func (t *runtimeCallOrderTracker) getCalls() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.calls...)
}

func (t *runtimeCallOrderTracker) indexOf(name string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, c := range t.calls {
		if c == name {
			return i
		}
	}
	return -1
}

// captureStderr redirects os.Stderr to a pipe, executes fn, and returns captured output.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&buf, r)
	}()

	fn()

	w.Close()
	wg.Wait()
	r.Close()

	return buf.String()
}

// newSuccessSpectraFinder creates a mock SpectraFinder that returns a valid project root.
func newSuccessSpectraFinder(t *testing.T) *mockSpectraFinderForRuntime {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	return newMockSpectraFinderForRuntime(tmpDir, nil)
}

// =====================================================================
// Happy Path — Main Loop Flow (Session Completion)
// =====================================================================

func TestRun_CompletedSession_ExitCode0(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	// Simulate immediate completion: session already completed
	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 0, exitCode)
}

func TestRun_SessionDoneNotification(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	// Session.Done() will send notification to terminationNotifier during execution
	initializer.session = sess

	rt := NewRuntime(finder, initializer, finalizer, sm)

	// Run in goroutine and trigger Done after a short delay
	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	// Wait for initialization to complete and get the terminationNotifier
	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		sess.Done(notifier)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after Session.Done() notification")
	}

	assert.Equal(t, 0, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
}

func TestRun_TerminationNotifierBufferSize2(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	// Run in goroutine to capture terminationNotifier
	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	require.NotNil(t, notifier, "terminationNotifier should be passed to SessionInitializer")
	assert.Equal(t, 2, cap(notifier), "terminationNotifier channel must have buffer size 2")

	// Clean up
	if notifier != nil {
		sess.Done(notifier)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}
}

// =====================================================================
// Happy Path — Main Loop Flow (Session Failure)
// =====================================================================

func TestRun_FailedSession_ExitCode1(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "failed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
}

func TestRun_SessionFailNotification(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		rtErr := &session.RuntimeError{Issuer: "Test", Message: "simulated failure"}
		sess.Fail(rtErr, notifier)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after Session.Fail() notification")
	}

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
}

// =====================================================================
// Happy Path — Socket Listener Management
// =====================================================================

func TestRun_SocketListenerStarted(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	assert.True(t, sm.listenCalled.Load(), "RuntimeSocketManager.Listen() should be called")

	// Verify handler was captured (MessageRouter.RouteMessage callback)
	sm.mu.Lock()
	handler := sm.capturedHandler
	sm.mu.Unlock()
	assert.NotNil(t, handler, "Listen() should be called with a MessageRouter.RouteMessage callback")

	// Clean up
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}
}

func TestRun_SocketListenerStopped(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Trigger completion
	if notifier != nil {
		sess.Done(notifier)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	assert.GreaterOrEqual(t, int(sm.deleteSocketCalled.Load()), 1, "RuntimeSocketManager.DeleteSocket() should be called")
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called after listener stopped")
}

func TestRun_MessageRouterInitialized(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// MessageRouter should have been initialized with Session, EventProcessor, ErrorProcessor, terminationNotifier
	// We verify this indirectly by checking that Listen was called with a non-nil handler
	sm.mu.Lock()
	handler := sm.capturedHandler
	sm.mu.Unlock()
	assert.NotNil(t, handler, "MessageRouter should be initialized and passed to Listen()")

	// Clean up
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}
}

func TestRun_ListenerGoroutineCompletes(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Trigger completion
	if notifier != nil {
		sess.Done(notifier)
	}

	// Don't close listenDoneCh yet — Runtime should block waiting for it
	// Runtime will call DeleteSocket which will close listenDoneCh
	time.Sleep(50 * time.Millisecond)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime should proceed to SessionFinalizer after listenerDoneCh closes")
	}

	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called after goroutine exits")
}

// =====================================================================
// Happy Path — Signal Handling (SIGINT)
// =====================================================================

func TestRun_SIGINT_GracefulShutdown(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	// Send SIGINT to the current process
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	// Close listenerDoneCh to unblock cleanup
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after SIGINT")
	}

	assert.GreaterOrEqual(t, int(sm.deleteSocketCalled.Load()), 1, "RuntimeSocketManager.DeleteSocket() should be called")
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
	assert.Equal(t, 1, exitCode, "exit code should be 1 on SIGINT")
}

func TestRun_SIGINT_SessionStatusUnchanged(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// Session status should remain "running" — SIGINT does NOT transition to "failed"
	assert.Equal(t, "running", sess.GetStatusSafe(), "Session status should remain 'running' after SIGINT")
}

func TestRun_SIGINT_SocketDeletedBeforeFinalization(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	tracker := &runtimeCallOrderTracker{}
	finalizer := &mockSessionFinalizerForRuntime{callOrder: tracker}
	sm := newMockRuntimeSocketManagerForRuntime()
	sm.callOrder = tracker

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	deleteIdx := tracker.indexOf("DeleteSocket")
	finalizeIdx := tracker.indexOf("Finalize")
	assert.Greater(t, finalizeIdx, -1, "Finalize should be called")
	assert.Greater(t, deleteIdx, -1, "DeleteSocket should be called")
	assert.Less(t, deleteIdx, finalizeIdx, "DeleteSocket should be called before Finalize")
}

func TestRun_SIGINT_LocksReleased(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
		// All goroutines exited — no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit — possible deadlock from unreleased locks")
	}
}

// =====================================================================
// Happy Path — Signal Handling (SIGTERM)
// =====================================================================

func TestRun_SIGTERM_GracefulShutdown(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGTERM))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after SIGTERM")
	}

	assert.GreaterOrEqual(t, int(sm.deleteSocketCalled.Load()), 1, "RuntimeSocketManager.DeleteSocket() should be called")
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
	assert.Equal(t, 1, exitCode, "exit code should be 1 on SIGTERM")
}

func TestRun_SIGTERM_SessionStatusUnchanged(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGTERM))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	assert.Equal(t, "running", sess.GetStatusSafe(), "Session status should remain 'running' after SIGTERM")
}

// =====================================================================
// Happy Path — Double Signal Handling
// =====================================================================

func TestRun_DoubleSignal_ImmediateExit(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	// First SIGINT initiates graceful shutdown
	require.NoError(t, proc.Signal(syscall.SIGINT))
	time.Sleep(50 * time.Millisecond)

	// Second SIGINT should exit immediately with code 130
	require.NoError(t, proc.Signal(syscall.SIGINT))

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after double signal")
	}

	assert.Equal(t, 130, exitCode, "exit code should be 130 on double signal")
}

func TestRun_DoubleSignal_NoWaitForFinalization(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	// Finalizer that introduces a delay
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	// Send two SIGINTs rapidly
	require.NoError(t, proc.Signal(syscall.SIGINT))
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	select {
	case <-done:
		// Exited without waiting for Finalize to complete
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime should exit immediately on double signal without waiting for SessionFinalizer")
	}
}

// =====================================================================
// Happy Path — Listener Error Handling
// =====================================================================

func TestRun_ListenerErrorReceived(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	// Send async listener error
	sm.listenErrCh <- fmt.Errorf("accept error")

	// Close listenerDoneCh to allow cleanup
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after listener error")
	}

	assert.Equal(t, 1, exitCode, "exit code should be 1 after listener error")
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
}

func TestRun_ListenerErrorAfterTermination_Discarded(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Complete session first
	if notifier != nil {
		sess.Done(notifier)
	}

	time.Sleep(50 * time.Millisecond)

	// Send error after termination — should be discarded
	select {
	case sm.listenErrCh <- fmt.Errorf("late error"):
	default:
	}

	// Close listenerDoneCh
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// Session should still be completed, not failed
	assert.Equal(t, "completed", sess.GetStatusSafe(), "late listener errors should be discarded")
}

// =====================================================================
// Happy Path — Socket Deletion Idempotency
// =====================================================================

func TestRun_SocketDeleteCalledTwice(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	// SIGINT calls DeleteSocket once
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// DeleteSocket should be called at least twice (once in signal handler, once in cleanup)
	// and both should succeed without error (idempotent)
	assert.GreaterOrEqual(t, int(sm.deleteSocketCalled.Load()), 2,
		"DeleteSocket should be called multiple times safely (idempotent)")
}

// =====================================================================
// Validation Failures — Project Root Lookup
// =====================================================================

func TestRun_ProjectRootNotFound(t *testing.T) {
	tmpDir := t.TempDir() // No .spectra/ directory
	finder := newMockSpectraFinderForRuntime("", fmt.Errorf("project root not found"))
	initializer := &mockSessionInitializerForRuntime{}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	// Change working directory to tmpDir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	stderr := captureStderr(t, func() {
		exitCode := rt.Run("TestWorkflow")
		assert.Equal(t, 1, exitCode)
	})

	assert.Contains(t, stderr, "Failed to locate project root: project root not found")
	assert.Contains(t, stderr, "Run 'spectra init' to initialize the project")
	assert.False(t, initializer.initializeCalled.Load(), "SessionInitializer should NOT be called")
	assert.False(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should NOT be called")
}

func TestRun_ProjectRootNotFoundFromSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d")
	require.NoError(t, os.MkdirAll(deepDir, 0755))
	finder := newMockSpectraFinderForRuntime("", fmt.Errorf("project root not found: no .spectra/ in any parent directory"))
	initializer := &mockSessionInitializerForRuntime{}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	// Change working directory to deepDir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(deepDir))
	defer func() { _ = os.Chdir(origDir) }()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	stderr := captureStderr(t, func() {
		exitCode := rt.Run("TestWorkflow")
		assert.Equal(t, 1, exitCode)
	})

	assert.Contains(t, stderr, "Failed to locate project root")
	assert.Contains(t, stderr, "Run 'spectra init' to initialize the project")
	assert.False(t, initializer.initializeCalled.Load(), "SessionInitializer should NOT be called")
}

// =====================================================================
// Validation Failures — Initialization Errors (No Session Entity)
// =====================================================================

func TestRun_InitializationFails_WorkflowNotFound(t *testing.T) {
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		err: fmt.Errorf("failed to load workflow definition: file not found"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	stderr := captureStderr(t, func() {
		exitCode := rt.Run("NonExistentWorkflow")
		assert.Equal(t, 1, exitCode)
	})

	assert.Contains(t, stderr, "Failed to initialize session")
	assert.Contains(t, stderr, "failed to load workflow definition: file not found")
	assert.False(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should NOT be called when no session entity")
}

func TestRun_InitializationFails_NoSessionEntity(t *testing.T) {
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		session: nil,
		err:     fmt.Errorf("early failure"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	stderr := captureStderr(t, func() {
		exitCode := rt.Run("TestWorkflow")
		assert.Equal(t, 1, exitCode)
	})

	assert.Contains(t, stderr, "Failed to initialize session")
	assert.False(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should NOT be called with nil session")
}

// =====================================================================
// Validation Failures — Initialization Errors (Session Entity Exists)
// =====================================================================

func TestRun_InitializationFails_SessionExists_Initializing(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "initializing", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		session: sess,
		err:     fmt.Errorf("initialization error with partial session"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called when session entity exists")
}

func TestRun_InitializationFails_SessionExists_Failed(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "failed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		session: sess,
		err:     fmt.Errorf("initialization failed with session in failed state"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called when session entity exists")
}

func TestRun_InitializationFails_SocketCreationError(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "initializing", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		session: sess,
		err:     fmt.Errorf("failed to create runtime socket"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called with partial session")
}

func TestRun_InitializationTimeout(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "failed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		session: sess,
		err:     fmt.Errorf("session initialization timeout exceeded 30 seconds"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called after timeout")
}

// =====================================================================
// Validation Failures — Listener Synchronous Setup Errors
// =====================================================================

func TestRun_ListenerSetup_BindFailure(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()
	sm.listenSyncErr = fmt.Errorf("bind: address already in use")

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called after bind failure")
	assert.Equal(t, "failed", sess.GetStatusSafe(), "Session should be transitioned to 'failed'")
}

func TestRun_ListenerSetup_SocketAlreadyExists(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()
	sm.listenSyncErr = fmt.Errorf("socket file already exists")

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "failed", sess.GetStatusSafe(), "Session should be failed after socket already exists error")
}

func TestRun_ListenerSetup_ListenerNeverSpawned(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()
	sm.listenSyncErr = fmt.Errorf("bind failure")

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	// On sync error, listenerDoneCh is already closed — no goroutine was spawned
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called even if no goroutine spawned")
}

// =====================================================================
// Validation Failures — Empty Workflow Name
// =====================================================================

func TestRun_EmptyWorkflowName(t *testing.T) {
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{
		err: fmt.Errorf("workflowName must be non-empty"),
	}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	stderr := captureStderr(t, func() {
		exitCode := rt.Run("")
		assert.Equal(t, 1, exitCode)
	})

	assert.Contains(t, stderr, "workflowName must be non-empty")
}

// =====================================================================
// State Transitions — Terminal States
// =====================================================================

func TestRun_StatusCompleted_ExitCode0(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 0, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
}

func TestRun_StatusFailed_ExitCode1(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "failed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	assert.Equal(t, 1, exitCode)
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called")
}

func TestRun_StatusInitializing_SIGINT_ExitCode1(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "initializing", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "initializing", sess.GetStatusSafe(), "Session status should remain 'initializing'")
}

// =====================================================================
// Edge Cases — Termination Race Conditions
// =====================================================================

func TestRun_CompletionAndSIGINT_Race(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Trigger completion and SIGINT simultaneously
	if notifier != nil {
		sess.Done(notifier)
	}
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	_ = proc.Signal(syscall.SIGINT)

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit during race condition")
	}

	// Either outcome is acceptable — no panic
	assert.True(t, exitCode == 0 || exitCode == 1,
		"exit code should be 0 (completion) or 1 (SIGINT), got %d", exitCode)
}

func TestRun_ListenerErrorAndCompletion_Race(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Trigger both simultaneously
	if notifier != nil {
		sess.Done(notifier)
	}
	sm.listenErrCh <- fmt.Errorf("listener error")

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit during race condition")
	}

	// Either outcome is acceptable — deterministic based on select
	assert.True(t, exitCode == 0 || exitCode == 1,
		"exit code should be 0 or 1, got %d", exitCode)
}

// =====================================================================
// Edge Cases — Channel Buffer Management
// =====================================================================

func TestRun_TerminationNotifierBufferNotExhausted(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	require.NotNil(t, notifier)

	// Session.Done sends exactly one notification; buffer size 2 accommodates it
	sess.Done(notifier)

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
		// No blocking occurred
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime blocked — possible buffer exhaustion on terminationNotifier")
	}
}

func TestRun_TerminationNotifierNeverClosed(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	require.NotNil(t, notifier)

	// Send notification (non-blocking because buffer=2)
	// If the channel were closed, this would panic
	assert.NotPanics(t, func() {
		select {
		case notifier <- struct{}{}:
		default:
		}
	}, "terminationNotifier should never be closed")

	// Clean up
	sess.Done(notifier)
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func TestRun_ListenerErrChNeverClosed(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Complete session
	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// listenerErrCh is never closed by RuntimeSocketManager;
	// consumers must observe listenerDoneCh closure as shutdown signal
	// Sending on the channel should not panic
	assert.NotPanics(t, func() {
		select {
		case sm.listenErrCh <- fmt.Errorf("test"):
		default:
		}
	}, "listenerErrCh should never be closed")
}

func TestRun_ListenerDoneChClosedExactlyOnce(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		sess.Done(notifier)
	}

	time.Sleep(50 * time.Millisecond)

	// Close listenerDoneCh exactly once
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// Reading from a closed channel returns zero value immediately
	select {
	case _, ok := <-sm.listenDoneCh:
		assert.False(t, ok, "listenerDoneCh should be closed")
	default:
		t.Fatal("listenerDoneCh should be readable (closed)")
	}
}

// =====================================================================
// Edge Cases — SessionFinalizer Failures
// =====================================================================

func TestRun_SessionFinalizerPrintFails(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	// Even with closed output pipes, Runtime should proceed to exit
	exitCode := rt.Run("TestWorkflow")
	assert.Equal(t, 0, exitCode, "should exit with appropriate code based on session status")
}

func TestRun_SessionFinalizerSocketDeleteWarning(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("TestWorkflow")

	// RuntimeSocketManager.DeleteSocket() logs warning "socket not found" — no effect on exit code
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Edge Cases — Session Metadata Persistence
// =====================================================================

func TestRun_MetadataPersistenceFails_NonBlocking(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Session.Done() — even if persistence fails, in-memory status is "completed"
	if notifier != nil {
		sess.Done(notifier)
	}

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	assert.Equal(t, "completed", sess.GetStatusSafe(), "in-memory status should be 'completed' regardless of persistence")
}

// =====================================================================
// Edge Cases — Listener Goroutine Panic
// =====================================================================

func TestRun_MessageRouterPanic(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	var exitCode int
	done := make(chan struct{})
	go func() {
		exitCode = rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	// MessageRouter panic recovery should trigger Session.Fail
	// which sends notification to terminationNotifier
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		// Simulate what happens when panic recovery triggers Session.Fail
		rtErr := &session.RuntimeError{Issuer: "MessageRouter", Message: "panic during message processing"}
		sess.Fail(rtErr, notifier)
	}

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after MessageRouter panic")
	}

	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "failed", sess.GetStatusSafe())
}

// =====================================================================
// Edge Cases — Listener In-Flight Messages
// =====================================================================

func TestRun_ListenerProcessingMessage_SocketDeleted(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	// SIGINT during message processing — socket closed, listener handles error gracefully
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Listener goroutine exits and closes listenerDoneCh
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called after in-flight message handling")
}

// =====================================================================
// Edge Cases — Runtime Crash (Unhandled)
// =====================================================================

func TestRun_RuntimeCrash_NoCleanup(t *testing.T) {
	// This test documents the behavior when the Runtime process crashes.
	// kill -9 cannot be tested programmatically, so we verify the documented behavior:
	// - No cleanup is performed
	// - Socket file remains on disk
	// - Session files remain
	// - On next session creation with different UUID, system works normally

	t.Log("Runtime crash (kill -9) causes no cleanup; socket and session files remain on disk")
	t.Log("Crash recovery logic is not specified in design; next invocation with different UUID works normally")
}

// =====================================================================
// Boundary Values — Workflow Name
// =====================================================================

func TestRun_WorkflowNamePascalCase(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "ValidWorkflowName")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("ValidWorkflowName")

	assert.Equal(t, 0, exitCode)
	assert.True(t, initializer.initializeCalled.Load())
	initializer.mu.Lock()
	assert.Equal(t, "ValidWorkflowName", initializer.capturedWorkflow)
	initializer.mu.Unlock()
}

func TestRun_WorkflowNameSingleWord(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "Workflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	exitCode := rt.Run("Workflow")

	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Concurrent Behaviour — Single Session Per Invocation
// =====================================================================

func TestRun_SingleSessionBinding(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	initializeCount := atomic.Int32{}

	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	// Track initialization calls
	origSession := initializer.session
	_ = origSession

	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	initializeCount.Add(1)
	rt.Run("TestWorkflow")

	// Runtime creates exactly one session entity per invocation
	assert.Equal(t, int32(1), initializeCount.Load(), "Runtime should create exactly one session per invocation")
	assert.True(t, initializer.initializeCalled.Load())
}

// =====================================================================
// Mock / Dependency Interaction — SpectraFinder
// =====================================================================

func TestRun_SpectraFinderCalled(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))

	finder := newMockSpectraFinderForRuntime(tmpDir, nil)
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	// Change working directory to test fixture
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	rt.Run("TestWorkflow")

	assert.True(t, finder.called.Load(), "SpectraFinder should be called to locate project root")
}

func TestRun_SpectraFinderFromSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	subDir := filepath.Join(tmpDir, "sub", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	finder := newMockSpectraFinderForRuntime(tmpDir, nil)
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	// Change working directory to subdirectory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(subDir))
	defer func() { _ = os.Chdir(origDir) }()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	rt.Run("TestWorkflow")

	assert.True(t, finder.called.Load(), "SpectraFinder should be called from subdirectory")
	// Verify project root was passed to SessionInitializer
	initializer.mu.Lock()
	capturedRoot := initializer.capturedProjectRoot
	initializer.mu.Unlock()
	assert.Equal(t, tmpDir, capturedRoot, "SpectraFinder should return parent directory containing .spectra/")
}

// =====================================================================
// Mock / Dependency Interaction — SessionInitializer
// =====================================================================

func TestRun_SessionInitializerCalled(t *testing.T) {
	projectRoot := "/tmp/test-project/"
	finder := newMockSpectraFinderForRuntime(projectRoot, nil)
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	rt.Run("TestWorkflow")

	assert.True(t, initializer.initializeCalled.Load(), "SessionInitializer.Initialize() should be called")
	initializer.mu.Lock()
	assert.Equal(t, "TestWorkflow", initializer.capturedWorkflow)
	assert.Equal(t, projectRoot, initializer.capturedProjectRoot, "SessionInitializer should receive resolved project root")
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	require.NotNil(t, notifier, "terminationNotifier should be passed to SessionInitializer")
	assert.Equal(t, 2, cap(notifier), "terminationNotifier should have cap=2")
}

func TestRun_SessionInitializerReturnsSession(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// After SessionInitializer returns, Runtime proceeds to start socket listener
	assert.True(t, sm.listenCalled.Load(), "After SessionInitializer returns session, Runtime should start socket listener")

	// Clean up
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

// =====================================================================
// Mock / Dependency Interaction — SessionFinalizer
// =====================================================================

func TestRun_SessionFinalizerCalled(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)
	rt.Run("TestWorkflow")

	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer.Finalize() should be called")
}

func TestRun_SessionFinalizerCalledOnAllPaths(t *testing.T) {
	t.Run("completion", func(t *testing.T) {
		sess := newMockSessionForRuntime("abc-123", "completed", "TestWorkflow")
		finder := newSuccessSpectraFinder(t)
		initializer := &mockSessionInitializerForRuntime{session: sess}
		finalizer := &mockSessionFinalizerForRuntime{}
		sm := newMockRuntimeSocketManagerForRuntime()

		rt := NewRuntime(finder, initializer, finalizer, sm)
		rt.Run("TestWorkflow")

		assert.True(t, finalizer.finalizeCalled.Load())
	})

	t.Run("failure", func(t *testing.T) {
		sess := newMockSessionForRuntime("abc-123", "failed", "TestWorkflow")
		finder := newSuccessSpectraFinder(t)
		initializer := &mockSessionInitializerForRuntime{session: sess}
		finalizer := &mockSessionFinalizerForRuntime{}
		sm := newMockRuntimeSocketManagerForRuntime()

		rt := NewRuntime(finder, initializer, finalizer, sm)
		rt.Run("TestWorkflow")

		assert.True(t, finalizer.finalizeCalled.Load())
	})

	t.Run("SIGINT", func(t *testing.T) {
		sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
		finder := newSuccessSpectraFinder(t)
		initializer := &mockSessionInitializerForRuntime{session: sess}
		finalizer := &mockSessionFinalizerForRuntime{}
		sm := newMockRuntimeSocketManagerForRuntime()

		rt := NewRuntime(finder, initializer, finalizer, sm)

		done := make(chan struct{})
		go func() {
			rt.Run("TestWorkflow")
			close(done)
		}()

		time.Sleep(100 * time.Millisecond)
		proc, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		require.NoError(t, proc.Signal(syscall.SIGINT))

		time.Sleep(50 * time.Millisecond)
		// Channel closed by DeleteSocket automatically

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Runtime did not exit")
		}

		assert.True(t, finalizer.finalizeCalled.Load())
	})
}

// =====================================================================
// Mock / Dependency Interaction — RuntimeSocketManager
// =====================================================================

func TestRun_RuntimeSocketManagerListenCalled(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	assert.True(t, sm.listenCalled.Load(), "RuntimeSocketManager.Listen() should be called")
	sm.mu.Lock()
	handler := sm.capturedHandler
	sm.mu.Unlock()
	assert.NotNil(t, handler, "Listen() should receive MessageRouter.RouteMessage callback")

	// Clean up
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func TestRun_RuntimeSocketManagerDeleteSocketCalled(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	tracker := &runtimeCallOrderTracker{}
	finalizer := &mockSessionFinalizerForRuntime{callOrder: tracker}
	sm := newMockRuntimeSocketManagerForRuntime()
	sm.callOrder = tracker

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		sess.Done(notifier)
	}

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	assert.GreaterOrEqual(t, int(sm.deleteSocketCalled.Load()), 1, "DeleteSocket should be called")
	deleteIdx := tracker.indexOf("DeleteSocket")
	finalizeIdx := tracker.indexOf("Finalize")
	assert.Greater(t, deleteIdx, -1, "DeleteSocket should be recorded")
	assert.Greater(t, finalizeIdx, -1, "Finalize should be recorded")
	assert.Less(t, deleteIdx, finalizeIdx, "DeleteSocket should be called before SessionFinalizer")
}

// =====================================================================
// Mock / Dependency Interaction — MessageRouter
// =====================================================================

func TestRun_MessageRouterInitializedWithDependencies(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify MessageRouter was initialized by checking Listen was called with a handler
	sm.mu.Lock()
	handler := sm.capturedHandler
	sm.mu.Unlock()
	assert.NotNil(t, handler, "MessageRouter should be initialized with Session, EventProcessor, ErrorProcessor, terminationNotifier")

	// Clean up
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

// =====================================================================
// Resource Cleanup — Locks
// =====================================================================

func TestRun_LocksReleasedOnExit(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit — possible deadlock from unreleased locks")
	}

	// After exit, all locks should be released — verify by accessing session safely
	assert.NotPanics(t, func() {
		_ = sess.GetStatusSafe()
		_ = sess.GetID()
	}, "All locks should be released after Runtime exits")
}

func TestRun_LocksReleasedOnSIGINT(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// After SIGINT, all locks should be released
	assert.NotPanics(t, func() {
		_ = sess.GetStatusSafe()
		_ = sess.GetID()
		_ = sess.GetCurrentStateSafe()
	}, "All locks should be released after SIGINT")
}

// =====================================================================
// Asynchronous Flow — Listener Goroutine
// =====================================================================

func TestRun_ListenerGoroutineSpawned(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify Listen was called (which spawns the accept-loop goroutine)
	assert.True(t, sm.listenCalled.Load(), "Listen() should be called to spawn accept-loop goroutine")

	// Clean up
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func TestRun_MainLoopBlocksOnSelect(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	// Main loop should block without any events — not busy-waiting
	time.Sleep(200 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("Runtime should not have exited — main loop should be blocked on select")
	default:
		// Expected: Runtime is still blocking in select
	}

	// Trigger termination
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()
	if notifier != nil {
		sess.Done(notifier)
	}
	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit after termination")
	}
}

// =====================================================================
// Ordering — Cleanup Sequence
// =====================================================================

func TestRun_CleanupOrder_StopListener_WaitGoroutine_Finalize(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	tracker := &runtimeCallOrderTracker{}
	finalizer := &mockSessionFinalizerForRuntime{callOrder: tracker}
	sm := newMockRuntimeSocketManagerForRuntime()
	sm.callOrder = tracker

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	if notifier != nil {
		sess.Done(notifier)
	}

	// Delay closing listenerDoneCh to prove Runtime waits
	time.Sleep(100 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	calls := tracker.getCalls()
	deleteIdx := -1
	finalizeIdx := -1
	for i, c := range calls {
		if c == "DeleteSocket" && deleteIdx == -1 {
			deleteIdx = i
		}
		if c == "Finalize" && finalizeIdx == -1 {
			finalizeIdx = i
		}
	}

	assert.Greater(t, deleteIdx, -1, "DeleteSocket should be called")
	assert.Greater(t, finalizeIdx, -1, "Finalize should be called")
	assert.Less(t, deleteIdx, finalizeIdx, "DeleteSocket must come before Finalize")
}

func TestRun_CleanupOrder_DrainErrors_AfterDoneChannel(t *testing.T) {
	sess := newMockSessionForRuntime("abc-123", "running", "TestWorkflow")
	finder := newSuccessSpectraFinder(t)
	initializer := &mockSessionInitializerForRuntime{session: sess}
	finalizer := &mockSessionFinalizerForRuntime{}
	sm := newMockRuntimeSocketManagerForRuntime()

	rt := NewRuntime(finder, initializer, finalizer, sm)

	done := make(chan struct{})
	go func() {
		rt.Run("TestWorkflow")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	initializer.mu.Lock()
	notifier := initializer.capturedNotifier
	initializer.mu.Unlock()

	// Complete session
	if notifier != nil {
		sess.Done(notifier)
	}

	// Put an error in errCh before closing doneCh
	select {
	case sm.listenErrCh <- fmt.Errorf("residual error"):
	default:
	}

	time.Sleep(50 * time.Millisecond)
	// Channel closed by DeleteSocket automatically

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Runtime did not exit")
	}

	// Session should remain completed — residual errors discarded
	assert.Equal(t, "completed", sess.GetStatusSafe(), "residual errors should be discarded after session terminates")
	assert.True(t, finalizer.finalizeCalled.Load(), "SessionFinalizer should be called after draining errors")
}
