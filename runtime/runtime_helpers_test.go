package runtime

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/storage"
)

// =============================================================================
// Runtime Test Helpers
//
// This file contains mocks, fakes, and fixture builders specifically for testing
// the runtime.Run() function. These mocks represent the dependencies that Run()
// constructs internally, and will be wired via seams once the production surface
// is established.
// =============================================================================

// --- Interfaces representing Runtime's internal dependencies ---
// These mirror what the production Run function interacts with.
// They are used by the test fixture mocks.

// runtimeSpectraFinder abstracts SpectraFinder.Find() for testing.
type runtimeSpectraFinder interface {
	Find() (string, error)
}

// runtimeWorkflowDef abstracts WorkflowDefinition for testing.
type runtimeWorkflowDef interface {
	EntryNode() string
	Nodes() []*components.Node
}

// --- Mock Implementations ---

// mockSpectraFinder is a mock for SpectraFinder.Find().
type mockSpectraFinder struct {
	result string
	err    error
}

func (m *mockSpectraFinder) Find() (string, error) {
	return m.result, m.err
}

// mockRuntimeSessionInitializer is a mock for SessionInitializer.Initialize().
type mockRuntimeSessionInitializer struct {
	mu                   sync.Mutex
	initializeCalled     int
	capturedWorkflowName string
	capturedSessionID    string
	capturedTermNotifier chan<- struct{}
	result               InitResult
	initializeFunc       func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult
}

func (m *mockRuntimeSessionInitializer) Initialize(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initializeCalled++
	m.capturedWorkflowName = workflowName
	m.capturedSessionID = sessionID
	m.capturedTermNotifier = terminationNotifier
	if m.initializeFunc != nil {
		return m.initializeFunc(workflowName, sessionID, terminationNotifier)
	}
	return m.result
}

// mockRuntimeSocketManager is a mock for RuntimeSocketManager lifecycle methods.
// It satisfies the SocketManager interface.
type mockRuntimeSocketManager struct {
	mu sync.Mutex

	// CreateSocket
	createSocketCalled int
	createSocketErr    error

	// Listen
	listenCalled    int
	listenErrCh     chan error
	listenDoneCh    chan struct{}
	listenErr       error
	capturedHandler storage.MessageHandler

	// DeleteSocket
	deleteSocketCalled int
	deleteSocketFunc   func() // optional hook for custom behavior
}

func (m *mockRuntimeSocketManager) CreateSocket() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createSocketCalled++
	return m.createSocketErr
}

func (m *mockRuntimeSocketManager) Listen(handler storage.MessageHandler) (<-chan error, <-chan struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listenCalled++
	m.capturedHandler = handler
	if m.listenErr != nil {
		return nil, nil, m.listenErr
	}
	return m.listenErrCh, m.listenDoneCh, nil
}

func (m *mockRuntimeSocketManager) DeleteSocket() {
	m.mu.Lock()
	m.deleteSocketCalled++
	fn := m.deleteSocketFunc
	m.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// Ensure mockRuntimeSocketManager satisfies SocketManager.
var _ SocketManager = (*mockRuntimeSocketManager)(nil)

// mockProcessCleaner is a mock for ProcessCleaner.
type mockProcessCleaner struct {
	mu          sync.Mutex
	cleanCalled int
	cleanFunc   func()
}

func (m *mockProcessCleaner) Clean() {
	m.mu.Lock()
	m.cleanCalled++
	fn := m.cleanFunc
	m.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// Ensure mockProcessCleaner satisfies ProcessCleaner.
var _ ProcessCleaner = (*mockProcessCleaner)(nil)

// mockRuntimeTransitionToNode is a mock for TransitionToNode dispatch.
// It satisfies the TransitionToNodeExecutor interface.
type mockRuntimeTransitionToNode struct {
	mu               sync.Mutex
	transitionCalled int
	capturedNodeName string
	capturedMessage  string
	transitionErr    error
	transitionFunc   func(targetNodeName, message string) error
}

func (m *mockRuntimeTransitionToNode) Execute(targetNodeName, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transitionCalled++
	m.capturedNodeName = targetNodeName
	m.capturedMessage = message
	if m.transitionFunc != nil {
		return m.transitionFunc(targetNodeName, message)
	}
	return m.transitionErr
}

// Ensure mockRuntimeTransitionToNode satisfies TransitionToNodeExecutor.
var _ TransitionToNodeExecutor = (*mockRuntimeTransitionToNode)(nil)

// mockRuntimeSessionFinalizer is a mock for SessionFinalizer.Finalize().
type mockRuntimeSessionFinalizer struct {
	mu              sync.Mutex
	finalizeCalled  int
	capturedSession *PersistentSession
	result          int
}

func (m *mockRuntimeSessionFinalizer) Finalize(session *PersistentSession) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finalizeCalled++
	m.capturedSession = session
	return m.result
}

// mockRuntimeWorkflowDef is a mock for WorkflowDefinition used in runtime tests.
type mockRuntimeWorkflowDef struct {
	entryNode string
	nodes     []*components.Node
}

func (m *mockRuntimeWorkflowDef) EntryNode() string {
	return m.entryNode
}

func (m *mockRuntimeWorkflowDef) Nodes() []*components.Node {
	return m.nodes
}

// --- Call Sequence Tracker ---

// callSequenceTracker records the order in which named operations are invoked.
type callSequenceTracker struct {
	mu    sync.Mutex
	calls []string
}

func newCallSequenceTracker() *callSequenceTracker {
	return &callSequenceTracker{}
}

func (t *callSequenceTracker) Record(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, name)
}

func (t *callSequenceTracker) Calls() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, len(t.calls))
	copy(out, t.calls)
	return out
}

// --- Fake Signal Channel ---

// fakeSignalSource provides a controllable channel for OS signal injection.
type fakeSignalSource struct {
	ch chan os.Signal
}

func newFakeSignalSource() *fakeSignalSource {
	return &fakeSignalSource{
		ch: make(chan os.Signal, 2),
	}
}

func (f *fakeSignalSource) Send(sig os.Signal) {
	f.ch <- sig
}

func (f *fakeSignalSource) Chan() <-chan os.Signal {
	return f.ch
}

// --- Fake Timer/Clock ---

// fakeTimer provides a controllable timer for grace period and sub-timeout testing.
type fakeTimer struct {
	ch      chan struct{}
	stopped bool
	mu      sync.Mutex
}

func newFakeTimer() *fakeTimer {
	return &fakeTimer{
		ch: make(chan struct{}, 1),
	}
}

func (ft *fakeTimer) Fire() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if !ft.stopped {
		select {
		case ft.ch <- struct{}{}:
		default:
		}
	}
}

func (ft *fakeTimer) Stop() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.stopped = true
}

func (ft *fakeTimer) Chan() <-chan struct{} {
	return ft.ch
}

// --- Fixture Builder ---

// runtimeTestFixture holds all mock dependencies for a runtime test scenario.
type runtimeTestFixture struct {
	Logger               *mockLogger
	SpectraFinder        *mockSpectraFinder
	SessionInitializer   *mockRuntimeSessionInitializer
	SocketManager        *mockRuntimeSocketManager
	TransitionToNode     *mockRuntimeTransitionToNode
	SessionFinalizer     *mockRuntimeSessionFinalizer
	WorkflowDef          *mockRuntimeWorkflowDef
	Session              *mockSession
	SignalSource         *fakeSignalSource
	GraceTimer           *fakeTimer
	ListenerTimer        *fakeTimer
	SequenceTracker      *callSequenceTracker
	ClaudeProcessCleaner *mockProcessCleaner
}

// newRuntimeTestFixture creates a fully-wired fixture with defaults for a
// successful session completion scenario.
func newRuntimeTestFixture(t *testing.T) *runtimeTestFixture {
	t.Helper()

	sess := newDefaultMockSession()
	sess.getStatusResult = "completed"

	listenerDoneCh := make(chan struct{})
	close(listenerDoneCh) // already closed by default

	f := &runtimeTestFixture{
		Logger: newDefaultMockLogger(),
		SpectraFinder: &mockSpectraFinder{
			result: "/home/user/project",
		},
		SessionInitializer: &mockRuntimeSessionInitializer{},
		SocketManager: &mockRuntimeSocketManager{
			listenErrCh:  make(chan error, 1),
			listenDoneCh: listenerDoneCh,
		},
		TransitionToNode: &mockRuntimeTransitionToNode{},
		SessionFinalizer: &mockRuntimeSessionFinalizer{
			result: 0,
		},
		WorkflowDef: &mockRuntimeWorkflowDef{
			entryNode: testEntryNode,
		},
		Session:              sess,
		SignalSource:         newFakeSignalSource(),
		GraceTimer:          newFakeTimer(),
		ListenerTimer:       newFakeTimer(),
		SequenceTracker:     newCallSequenceTracker(),
		ClaudeProcessCleaner: &mockProcessCleaner{},
	}

	// Wire the session initializer to return a successful InitResult.
	// Note: We cannot construct a real PersistentSession with mockSession here
	// without the production Run function wiring it. The fixture expresses intent.
	f.SessionInitializer.result = InitResult{
		PersistentSession:  newTestPersistentSession(t, sess),
		WorkflowDefinition: nil, // will be set by the production code from loader
		Error:              nil,
	}

	return f
}

// newTestPersistentSession creates a PersistentSession backed by a mockSession
// for test scenarios that need a real PersistentSession reference.
func newTestPersistentSession(t *testing.T, sess *mockSession) *PersistentSession {
	t.Helper()
	return NewPersistentSession(
		sess,
		newDefaultMockMetadataStore(),
		newDefaultMockEventStore(),
		newDefaultMockLogger(),
	)
}

// --- Logger Assertion Helpers ---

// assertLoggerHasInfoMsg checks that mockLogger.infoCalls contains a message.
func assertLoggerHasInfoMsg(t *testing.T, l *mockLogger, msg string) {
	t.Helper()
	for _, call := range l.infoCalls {
		if call.msg == msg {
			return
		}
	}
	t.Errorf("expected Logger.Info with msg=%q, got: %+v", msg, l.infoCalls)
}

// assertLoggerHasInfoMsgContaining checks that mockLogger.infoCalls has a message
// containing the given substring.
func assertLoggerHasInfoMsgContaining(t *testing.T, l *mockLogger, substr string) {
	t.Helper()
	for _, call := range l.infoCalls {
		if contains(call.msg, substr) {
			return
		}
	}
	t.Errorf("expected Logger.Info with msg containing %q, got: %+v", substr, l.infoCalls)
}

// assertLoggerHasWarnMsg checks that mockLogger.warnCalls contains a message.
func assertLoggerHasWarnMsg(t *testing.T, l *mockLogger, msg string) {
	t.Helper()
	for _, call := range l.warnCalls {
		if call.msg == msg {
			return
		}
	}
	t.Errorf("expected Logger.Warn with msg=%q, got: %+v", msg, l.warnCalls)
}

// assertLoggerHasWarnMsgContaining checks that mockLogger.warnCalls has a message
// containing the given substring.
func assertLoggerHasWarnMsgContaining(t *testing.T, l *mockLogger, substr string) {
	t.Helper()
	for _, call := range l.warnCalls {
		if contains(call.msg, substr) {
			return
		}
	}
	t.Errorf("expected Logger.Warn with msg containing %q, got: %+v", substr, l.warnCalls)
}

// assertLoggerNoWarnMsg checks that no warn message matches exactly.
func assertLoggerNoWarnMsg(t *testing.T, l *mockLogger, msg string) {
	t.Helper()
	for _, call := range l.warnCalls {
		if call.msg == msg {
			t.Errorf("expected Logger.Warn NOT to contain msg=%q, but it does", msg)
			return
		}
	}
}

// contains is a simple string containment check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- Mock PersistentSession Wrappers for Runtime Tests ---

// runtimeMockPersistentSession wraps a mockSession with tracking for Fail calls
// specific to runtime error validation (checking RuntimeError fields).
type runtimeMockPersistentSession struct {
	*mockSession
	failCalls []failCallRecord
	mu        sync.Mutex
}

type failCallRecord struct {
	err      error
	notifier chan<- struct{}
}

func newRuntimeMockPersistentSession() *runtimeMockPersistentSession {
	return &runtimeMockPersistentSession{
		mockSession: newDefaultMockSession(),
	}
}

func (m *runtimeMockPersistentSession) Fail(err error, notifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCalls = append(m.failCalls, failCallRecord{err: err, notifier: notifier})
	m.mockSession.failCalled++
	m.mockSession.failInputErr = err
	m.mockSession.failNotifier = notifier
	return m.mockSession.failErr
}

func (m *runtimeMockPersistentSession) getFailCalls() []failCallRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]failCallRecord, len(m.failCalls))
	copy(out, m.failCalls)
	return out
}

// --- Seam Wiring ---

// wireFixtureToSeams replaces the package-level seam variables with the fixture's
// mock behaviors and registers a cleanup to restore originals. It enables
// integration-style testing of the Run function using controlled mocks.
func wireFixtureToSeams(t *testing.T, f *runtimeTestFixture) {
	t.Helper()

	// Save originals.
	origSpectraFinder := spectraFinderFunc
	origPreSession := preSessionDepsConstructor
	origSessionInit := sessionInitializeFunc
	origPostSession := constructPostSessionDepsFunc
	origSignalNotify := signalNotifyFunc
	origSignalStop := signalStopFunc
	origGraceTimer := newGraceTimerFunc
	origSubTimer := newSubTimerFunc

	// Restore on cleanup.
	t.Cleanup(func() {
		spectraFinderFunc = origSpectraFinder
		preSessionDepsConstructor = origPreSession
		sessionInitializeFunc = origSessionInit
		constructPostSessionDepsFunc = origPostSession
		signalNotifyFunc = origSignalNotify
		signalStopFunc = origSignalStop
		newGraceTimerFunc = origGraceTimer
		newSubTimerFunc = origSubTimer
	})

	// Wire spectraFinder.
	spectraFinderFunc = func() (string, error) {
		return f.SpectraFinder.result, f.SpectraFinder.err
	}

	// Wire preSessionDepsConstructor (not typically needed since sessionInitializeFunc
	// bypasses it, but wire it for completeness).
	preSessionDepsConstructor = func(projectRoot string) (WorkflowLoader, SessionDirManager, error) {
		return nil, nil, nil
	}

	// Wire sessionInitializeFunc.
	sessionInitializeFunc = func(projectRoot string, wfLoader WorkflowLoader, dirMgr SessionDirManager, log logger.Logger, workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		return f.SessionInitializer.Initialize(workflowName, sessionID, terminationNotifier)
	}

	// Wire constructPostSessionDepsFunc.
	constructPostSessionDepsFunc = func(projectRoot string, ps *PersistentSession, wfDef *components.WorkflowDefinition, terminationNotifier chan<- struct{}, log logger.Logger) (*runtimePostSessionDeps, error) {
		return &runtimePostSessionDeps{
			socketManager:        f.SocketManager,
			transitionNode:       f.TransitionToNode,
			messageRouter:        NewMessageRouter(ps, nil, nil, terminationNotifier, log),
			finalizer:            NewSessionFinalizer(log),
			claudeProcessCleaner: f.ClaudeProcessCleaner,
		}, nil
	}

	// Wire signalNotifyFunc.
	// We use a done channel to allow the relay goroutine to exit when test cleanup runs.
	signalDone := make(chan struct{})
	t.Cleanup(func() { close(signalDone) })
	signalNotifyFunc = func(c chan<- os.Signal, sig ...os.Signal) {
		// Relay from the fake signal source to the provided channel.
		go func() {
			for {
				select {
				case s, ok := <-f.SignalSource.ch:
					if !ok {
						return
					}
					select {
					case c <- s:
					case <-signalDone:
						return
					}
				case <-signalDone:
					return
				}
			}
		}()
	}

	// Wire signalStopFunc.
	signalStopFunc = func(c chan<- os.Signal) {
		// No-op in tests.
	}

	// Wire grace timer.
	newGraceTimerFunc = func(d time.Duration) (<-chan struct{}, func()) {
		return f.GraceTimer.Chan(), f.GraceTimer.Stop
	}

	// Wire sub timer.
	newSubTimerFunc = func(d time.Duration) (<-chan struct{}, func()) {
		return f.ListenerTimer.Chan(), f.ListenerTimer.Stop
	}
}

// --- Unused import guard (logger is used for interface typing) ---

var _ logger.Logger = (*mockLogger)(nil)
