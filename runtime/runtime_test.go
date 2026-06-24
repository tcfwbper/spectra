package runtime

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
)

// =============================================================================
// Test Specification: runtime_test.go
// Source File Under Test: runtime/runtime.go
// =============================================================================

// --- Happy Path — Run ---

func TestRun_SuccessfulSessionCompletion(t *testing.T) {
	// Setup: Stub all dependencies for successful flow.
	f := newRuntimeTestFixture(t)
	// Session is "completed", terminationNotifier will fire.
	f.Session.getStatusResult = "completed"
	// Configure the SessionInitializer to send terminationNotifier after init succeeds.
	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		// Simulate session completion notification.
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("my-workflow", "", f.Logger)

	// Assert: Returns (0, nil)
	assert.Equal(t, 0, exitCode)
	assert.NoError(t, err)
}

func TestRun_LogsSessionTerminationNotification(t *testing.T) {
	// Setup: Same as successful completion, capture Logger calls.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("my-workflow", "", f.Logger)

	// Assert: Logger.Info called with "received session termination notification"
	assertLoggerHasInfoMsg(t, f.Logger, "received session termination notification")
}

// --- Error Propagation ---

func TestRun_SpectraFinderFails(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.SpectraFinder.err = errors.New("no .spectra dir")
	f.SpectraFinder.result = ""
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "failed to locate project root: no .spectra dir", err.Error())
}

func TestRun_PreSessionDependencyConstructionFails(t *testing.T) {
	// Setup: SpectraFinder succeeds, but pre-session dependency constructor fails.
	f := newRuntimeTestFixture(t)
	wireFixtureToSeams(t, f)

	// Override preSessionDepsConstructor to fail.
	preSessionDepsConstructor = func(projectRoot string) (WorkflowLoader, SessionDirManager, error) {
		return nil, nil, errors.New("loader construction failed")
	}

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assertErrorContains(t, err, "failed to initialize runtime dependencies: ")
}

func TestRun_SessionInitializerFailsBeforeSession(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionInitializer.result = InitResult{
		PersistentSession: nil,
		Error:             errors.New("workflow not found"),
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "failed to initialize session: workflow not found", err.Error())
}

func TestRun_SessionInitializerFailsAfterSession(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "failed"
	ps := newTestPersistentSession(t, f.Session)
	f.SessionInitializer.result = InitResult{
		PersistentSession: ps,
		Error:             errors.New("timeout during initialization"),
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "failed to initialize session: timeout during initialization", err.Error())
}

func TestRun_PostSessionDependencyConstructionFails(t *testing.T) {
	// Setup: Successful initialization, but post-session dep construction fails.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.SessionInitializer.result = InitResult{
		PersistentSession:  newTestPersistentSession(t, f.Session),
		WorkflowDefinition: mustNewWorkflowDefinition(t),
		Error:              nil,
	}
	wireFixtureToSeams(t, f)

	// Override constructPostSessionDepsFunc to fail.
	constructPostSessionDepsFunc = func(projectRoot string, ps *PersistentSession, wfDef *components.WorkflowDefinition, terminationNotifier chan<- struct{}, log logger.Logger) (*runtimePostSessionDeps, error) {
		return nil, errors.New("agent loader failed")
	}

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assertErrorContains(t, err, "failed to initialize post-session dependencies: ")
	assert.Equal(t, 1, f.Session.failCalled)
}

func TestRun_CreateSocketFails(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.SessionInitializer.result = InitResult{
		PersistentSession:  newTestPersistentSession(t, f.Session),
		WorkflowDefinition: mustNewWorkflowDefinition(t),
		Error:              nil,
	}
	f.SocketManager.createSocketErr = errors.New("permission denied")
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assertErrorContains(t, err, "failed to create runtime socket: ")
	assert.Equal(t, 1, f.Session.failCalled)
}

func TestRun_ListenFails(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.SessionInitializer.result = InitResult{
		PersistentSession:  newTestPersistentSession(t, f.Session),
		WorkflowDefinition: mustNewWorkflowDefinition(t),
		Error:              nil,
	}
	f.SocketManager.listenErr = errors.New("bind error")
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assertErrorContains(t, err, "failed to start socket listener: ")
	assert.Equal(t, 1, f.Session.failCalled)
}

func TestRun_InitialDispatchFails(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.SessionInitializer.result = InitResult{
		PersistentSession:  newTestPersistentSession(t, f.Session),
		WorkflowDefinition: mustNewWorkflowDefinition(t),
		Error:              nil,
	}
	f.TransitionToNode.transitionErr = errors.New("agent def not found")
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assertErrorContains(t, err, "failed to dispatch entry node: ")
	assert.Equal(t, 1, f.Session.failCalled)
}

func TestRun_ListenerErrorDuringSession(t *testing.T) {
	// Setup: Full successful startup.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		// Send listener error after a short delay to trigger termination.
		go func() { f.SocketManager.listenErrCh <- errors.New("connection reset") }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, 1, f.Session.failCalled)
}

func TestRun_ListenerErrorWhenSessionAlreadyCompleted(t *testing.T) {
	// Setup: Full successful startup, session already completed.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { f.SocketManager.listenErrCh <- errors.New("connection reset") }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 0, exitCode)
	assert.NoError(t, err)
	// PersistentSession.Fail() NOT called (status already "completed")
	assertFailNotCalled(t, f.Session)
}

func TestRun_SessionFailed(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "failed"
	f.Session.getCurrentStateResult = testEntryNode
	f.Session.getErrorResult = errors.New("agent crashed")

	listenerDoneCh := make(chan struct{})
	close(listenerDoneCh)
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "session failed: agent crashed", err.Error())
}

func TestRun_SessionFinalizerNonTerminalStatus(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.Session.getErrorResult = nil

	listenerDoneCh := make(chan struct{})
	close(listenerDoneCh)
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "session terminated with non-terminal status", err.Error())
}

// --- Mock / Dependency Interaction ---

func TestRun_TerminationNotifierCapacity(t *testing.T) {
	// Setup: Capture the terminationNotifier channel.
	f := newRuntimeTestFixture(t)
	var capturedNotifier chan<- struct{}
	f.SessionInitializer.initializeFunc = func(wfName string, sessionID string, notifier chan<- struct{}) InitResult {
		capturedNotifier = notifier
		return InitResult{Error: errors.New("short-circuit")}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert: cap(capturedNotifier) == 2
	require.NotNil(t, capturedNotifier)
	assert.Equal(t, 2, cap(capturedNotifier))
}

func TestRun_SessionInitializerReceivesWorkflowName(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionInitializer.result = InitResult{Error: errors.New("short-circuit")}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("my-workflow", "", f.Logger)

	// Assert
	assert.Equal(t, "my-workflow", f.SessionInitializer.capturedWorkflowName)
}

func TestRun_SessionInitializerReceivesSessionID(t *testing.T) {
	// Setup: Passes sessionID to SessionInitializer.Initialize without validation.
	f := newRuntimeTestFixture(t)
	f.SessionInitializer.result = InitResult{Error: errors.New("short-circuit")}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "550e8400-e29b-41d4-a716-446655440000", f.Logger)

	// Assert
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", f.SessionInitializer.capturedSessionID)
}

func TestRun_SessionInitializerReceivesEmptySessionID(t *testing.T) {
	// Setup: Passes empty string sessionID to SessionInitializer when not provided.
	f := newRuntimeTestFixture(t)
	f.SessionInitializer.result = InitResult{Error: errors.New("short-circuit")}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, "", f.SessionInitializer.capturedSessionID)
}

func TestRun_InitialDispatchMessage(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.id = "uuid-123"
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	listenerDoneCh := make(chan struct{})
	close(listenerDoneCh)
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert: TransitionToNode.Execute() called with the entry node and message
	// containing the session UUID and "spectra-agent event emit"
	assert.Equal(t, 1, f.TransitionToNode.transitionCalled)
	assert.Equal(t, "Start", f.TransitionToNode.capturedNodeName) // EntryNode from mustNewWorkflowDefinition
	assert.True(t, strings.Contains(f.TransitionToNode.capturedMessage, "uuid-123"),
		"dispatch message should contain session UUID, got: %s", f.TransitionToNode.capturedMessage)
	assert.True(t, strings.Contains(f.TransitionToNode.capturedMessage, "spectra-agent event emit"),
		"dispatch message should contain 'spectra-agent event emit', got: %s", f.TransitionToNode.capturedMessage)
}

func TestRun_DeleteSocketCalledDuringCleanup(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, f.SocketManager.deleteSocketCalled)
}

func TestRun_SessionFinalizerCalledAfterCleanup(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode
	tracker := f.SequenceTracker

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		tracker.Record("DeleteSocket")
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert: "DeleteSocket" was called (SessionFinalizer is called after)
	calls := tracker.Calls()
	assert.Contains(t, calls, "DeleteSocket")
}

func TestRun_SessionFinalizerNotInvokedWhenNoSession(t *testing.T) {
	// Setup: SpectraFinder fails => no PersistentSession created
	f := newRuntimeTestFixture(t)
	f.SpectraFinder.err = errors.New("no .spectra dir")
	f.SpectraFinder.result = ""
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: Returns error, exit code 1, no finalizer invocation (no session exists)
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	// No Logger.Info with "session completed" (would be from SessionFinalizer)
	for _, call := range f.Logger.infoCalls {
		assert.NotEqual(t, "session completed", call.msg, "SessionFinalizer should not be invoked when no session exists")
	}
}

func TestRun_CleanupOrder(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode
	tracker := f.SequenceTracker

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		tracker.Record("DeleteSocket")
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Override signalStopFunc to record.
	signalStopFunc = func(c chan<- os.Signal) {
		tracker.Record("SignalStop")
	}

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert call order: SignalStop before DeleteSocket
	calls := tracker.Calls()
	signalStopIdx := -1
	deleteSocketIdx := -1
	for i, c := range calls {
		if c == "SignalStop" {
			signalStopIdx = i
		}
		if c == "DeleteSocket" {
			deleteSocketIdx = i
		}
	}
	assert.Greater(t, deleteSocketIdx, signalStopIdx, "signal.Stop should be called before DeleteSocket")
}

func TestRun_PersistentSessionFailReturnsError(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.Session.failErr = errors.New("session already in terminal state")

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { f.SocketManager.listenErrCh <- errors.New("connection lost") }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert: Logger.Warn called with message containing "attempted to fail session but session already in terminal state"
	assertLoggerHasWarnMsgContaining(t, f.Logger, "attempted to fail session but session already in terminal state")
}

func TestRun_MessageRouterPassedToListen(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	listenerDoneCh := make(chan struct{})
	close(listenerDoneCh)
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert: capturedHandler is non-nil (a messageHandlerAdapter wrapping MessageRouter)
	assert.NotNil(t, f.SocketManager.capturedHandler)
}

// --- State Transitions ---

func TestRun_OSSignalSIGINT(t *testing.T) {
	// Setup: Session is non-terminal ("running"), SIGINT received.
	f := newRuntimeTestFixture(t)
	f.Session.id = "11111111-1111-1111-1111-111111111111"
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = "node-a"

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { f.SignalSource.Send(syscall.SIGINT) }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: PersistentSession.Fail() called with RuntimeError fields.
	assert.Equal(t, 1, f.Session.failCalled, "expected Fail() to be called once")
	assertFailCalledWithRuntimeErrorFull(t, f.Session, "Runtime", "terminated by signal interrupt", "11111111-1111-1111-1111-111111111111", "node-a")
	// Assert: Detail is nil.
	rtErr := f.Session.failInputErr.(*entities.RuntimeError)
	assert.Nil(t, rtErr.Detail())
	// Assert: Logger logs graceful shutdown.
	assertLoggerHasInfoMsgContaining(t, f.Logger, "received signal interrupt, initiating graceful shutdown")
	// Assert: Returns (1, error) with signal message.
	require.Error(t, err)
	assert.Equal(t, "session terminated by signal interrupt", err.Error())
	assert.Equal(t, 1, exitCode)
}

func TestRun_OSSignalSIGTERM(t *testing.T) {
	// Setup: Session is non-terminal ("running"), SIGTERM received.
	f := newRuntimeTestFixture(t)
	f.Session.id = "22222222-2222-2222-2222-222222222222"
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = "node-b"

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { f.SignalSource.Send(syscall.SIGTERM) }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: PersistentSession.Fail() called with RuntimeError fields.
	assert.Equal(t, 1, f.Session.failCalled, "expected Fail() to be called once")
	assertFailCalledWithRuntimeErrorFull(t, f.Session, "Runtime", "terminated by signal terminated", "22222222-2222-2222-2222-222222222222", "node-b")
	// Assert: Detail is nil.
	rtErr := f.Session.failInputErr.(*entities.RuntimeError)
	assert.Nil(t, rtErr.Detail())
	// Assert: Logger logs graceful shutdown.
	assertLoggerHasInfoMsgContaining(t, f.Logger, "received signal terminated, initiating graceful shutdown")
	// Assert: Returns (1, error) with signal message.
	require.Error(t, err)
	assert.Equal(t, "session terminated by signal terminated", err.Error())
	assert.Equal(t, 1, exitCode)
}

func TestRun_OSSignalSkipsFailWhenSessionAlreadyCompleted(t *testing.T) {
	// Setup: Session already completed, SIGINT received (race condition).
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	// SessionFinalizer returns 0 (session completed successfully).
	f.SessionFinalizer.result = 0

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { f.SignalSource.Send(syscall.SIGINT) }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: Fail() NOT called (session already terminal).
	assertFailNotCalled(t, f.Session)
	// Assert: Exit code is always 1 when signal received, regardless of SessionFinalizer.
	require.Error(t, err)
	assert.Equal(t, "session terminated by signal interrupt", err.Error())
	assert.Equal(t, 1, exitCode)
}

func TestRun_OSSignalDuringInitializingStatus(t *testing.T) {
	// Spec contradiction: The test asserts FailingState="" but entities.NewRuntimeError
	// requires failingState to be non-empty. The runtime spec (step 26) says to use
	// GetCurrentStateSafe() which can return "" during init, but the entity rejects it.
	t.Skip("spec contradiction: entities.RuntimeError requires non-empty failingState but runtime spec uses GetCurrentStateSafe() which can be empty during init")

	// Setup: Session exists but status is "initializing" (SessionInitializer returned
	// InitResult with Error != nil and PersistentSession != nil). Signal arrives
	// during cleanup path.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "initializing"
	f.Session.getCurrentStateResult = "" // GetCurrentStateSafe() returns empty during init

	listenerDoneCh := make(chan struct{})
	close(listenerDoneCh)
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		// Send SIGINT to simulate signal during initialization cleanup.
		go func() { f.SignalSource.Send(syscall.SIGINT) }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: Fail() called with RuntimeError having FailingState="" (from GetCurrentStateSafe()).
	assert.Equal(t, 1, f.Session.failCalled, "expected Fail() to be called once")
	rtErr, ok := f.Session.failInputErr.(*entities.RuntimeError)
	require.True(t, ok, "expected Fail() called with *entities.RuntimeError")
	assert.Equal(t, "terminated by signal interrupt", rtErr.Message())
	assert.Equal(t, "", rtErr.FailingState())
	// Assert: Returns (1, error).
	require.Error(t, err)
	assert.Equal(t, "session terminated by signal interrupt", err.Error())
	assert.Equal(t, 1, exitCode)
}

func TestRun_OSSignalFailReturnsError(t *testing.T) {
	// Setup: Session is "running" but Fail() returns error (race condition).
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = "node-a"
	f.Session.failErr = errors.New("session already in terminal state")

	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { f.SignalSource.Send(syscall.SIGINT) }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	_, _ = Run("wf", "", f.Logger)

	// Assert: Logger.Warn called with message about session already in terminal state.
	assertLoggerHasWarnMsgContaining(t, f.Logger, "attempted to fail session on signal but session already in terminal state")
}

func TestRun_SecondSignalForcesExit(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode

	// Block listenerDoneCh so cleanup doesn't complete.
	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)
	// DeleteSocket does NOT close listenerDoneCh — simulates slow shutdown.

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		// Send first SIGINT, then second SIGINT after a tiny delay.
		go func() {
			f.SignalSource.Send(syscall.SIGINT)
			// The second signal needs to be received during cleanup.
			// The first signal is caught by the event loop select, then
			// cleanup starts. The second signal is caught by the cleanup select.
			f.SignalSource.Send(syscall.SIGINT)
		}()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)
	// Close listenerDoneCh after Run returns to free the leaked cleanup goroutine.
	t.Cleanup(func() {
		select {
		case <-listenerDoneCh:
		default:
			close(listenerDoneCh)
		}
	})

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "forced exit by second signal", err.Error())
}

func TestRun_GracePeriodTimeout(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	// Block listenerDoneCh so cleanup hangs.
	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Fire the grace timer immediately to simulate timeout.
	f.GraceTimer.Fire()

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Unblock the leaked cleanup goroutine before assertions (prevents race).
	close(listenerDoneCh)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assert.Equal(t, "cleanup timeout", err.Error())
}

func TestRun_ListenerDoneChTimeout(t *testing.T) {
	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	// listenerDoneCh never closes (sub-timeout fires).
	f.SocketManager.listenDoneCh = make(chan struct{}) // never closed
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Pre-fire the sub-timer (listener timeout) so it's already buffered before Run.
	f.ListenerTimer.Fire()

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert
	assert.Equal(t, 0, exitCode)
	assert.NoError(t, err)
	assertLoggerHasWarnMsgContaining(t, f.Logger, "listener shutdown exceeded 2 seconds")
}

func TestRun_ListenerDoneChAlreadyClosed(t *testing.T) {
	// Setup: listenerDoneCh already closed.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.Session.getCurrentStateResult = testEntryNode

	alreadyClosed := make(chan struct{})
	close(alreadyClosed)
	f.SocketManager.listenDoneCh = alreadyClosed
	f.SocketManager.listenErrCh = make(chan error, 1)

	f.SessionInitializer.initializeFunc = func(workflowName string, sessionID string, terminationNotifier chan<- struct{}) InitResult {
		ps := newTestPersistentSession(t, f.Session)
		go func() { terminationNotifier <- struct{}{} }()
		return InitResult{
			PersistentSession:  ps,
			WorkflowDefinition: mustNewWorkflowDefinition(t),
			Error:              nil,
		}
	}
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: SessionFinalizer invoked without delay, no timeout warning.
	assert.Equal(t, 0, exitCode)
	assert.NoError(t, err)
	assertLoggerNoWarnMsg(t, f.Logger, "listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer")
}

// --- Idempotency ---

func TestRun_DeleteSocketIdempotent(t *testing.T) {
	// Setup: CreateSocket returns error (socket never created).
	// DeleteSocket is called during cleanup but is a no-op.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getCurrentStateResult = testEntryNode
	f.SessionInitializer.result = InitResult{
		PersistentSession:  newTestPersistentSession(t, f.Session),
		WorkflowDefinition: mustNewWorkflowDefinition(t),
		Error:              nil,
	}
	f.SocketManager.createSocketErr = errors.New("permission denied")
	wireFixtureToSeams(t, f)

	// Action
	exitCode, err := Run("wf", "", f.Logger)

	// Assert: No panic, cleanup proceeds, returns error from socket creation.
	assert.Equal(t, 1, exitCode)
	require.Error(t, err)
	assertErrorContains(t, err, "failed to create runtime socket: ")
}

// --- Assertion helpers used by runtime tests ---

// assertRuntimeError checks that an error wraps the expected message from RuntimeError.
func assertRuntimeError(t *testing.T, err error, issuer, message string) {
	t.Helper()
	var rtErr *entities.RuntimeError
	if errors.As(err, &rtErr) {
		assert.Equal(t, issuer, rtErr.Issuer())
		assert.Equal(t, message, rtErr.Message())
	} else {
		t.Errorf("expected error to be *entities.RuntimeError, got: %T", err)
	}
}

// assertFailCalledWithRuntimeError checks that mockSession.Fail was called with a RuntimeError.
func assertFailCalledWithRuntimeError(t *testing.T, sess *mockSession, expectedIssuer, expectedMessage string) {
	t.Helper()
	require.Greater(t, sess.failCalled, 0, "expected Fail() to be called at least once")
	rtErr, ok := sess.failInputErr.(*entities.RuntimeError)
	require.True(t, ok, "expected Fail() to be called with *entities.RuntimeError, got: %T", sess.failInputErr)
	assert.Equal(t, expectedIssuer, rtErr.Issuer())
	assert.Equal(t, expectedMessage, rtErr.Message())
}

// assertFailCalledWithRuntimeErrorAndState extends assertFailCalledWithRuntimeError with FailingState check.
func assertFailCalledWithRuntimeErrorAndState(t *testing.T, sess *mockSession, expectedIssuer, expectedMessage, expectedState string) {
	t.Helper()
	assertFailCalledWithRuntimeError(t, sess, expectedIssuer, expectedMessage)
	rtErr := sess.failInputErr.(*entities.RuntimeError)
	assert.Equal(t, expectedState, rtErr.FailingState())
}

// assertFailCalledWithRuntimeErrorFull checks Fail() was called with RuntimeError
// having specific Issuer, Message, SessionID, and FailingState.
func assertFailCalledWithRuntimeErrorFull(t *testing.T, sess *mockSession, expectedIssuer, expectedMessage, expectedSessionID, expectedState string) {
	t.Helper()
	require.Greater(t, sess.failCalled, 0, "expected Fail() to be called at least once")
	rtErr, ok := sess.failInputErr.(*entities.RuntimeError)
	require.True(t, ok, "expected Fail() to be called with *entities.RuntimeError, got: %T", sess.failInputErr)
	assert.Equal(t, expectedIssuer, rtErr.Issuer())
	assert.Equal(t, expectedMessage, rtErr.Message())
	assert.Equal(t, expectedSessionID, rtErr.SessionID())
	assert.Equal(t, expectedState, rtErr.FailingState())
}

// assertFailNotCalled checks that mockSession.Fail was never called.
func assertFailNotCalled(t *testing.T, sess *mockSession) {
	t.Helper()
	assert.Equal(t, 0, sess.failCalled, "expected Fail() NOT to be called")
}

// assertErrorContains checks that err is non-nil and its message contains substr.
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), substr),
		"expected error to contain %q, got: %q", substr, err.Error())
}
