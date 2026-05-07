package runtime

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities"
)

// =============================================================================
// Test Specification: runtime_test.go
// Source File Under Test: runtime/runtime.go
//
// All tests are scaffolded because runtime/runtime.go does not yet exist.
// Each test documents the exact production surface it requires via t.Skip().
// Once the Run() function and its internal dependency seams are implemented,
// these scaffolds can be converted to concrete assertions.
// =============================================================================

// --- Happy Path — Run ---

func TestRun_SuccessfulSessionCompletion(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and dependency injection seams (SpectraFinder, SessionInitializer, RuntimeSocketManager, TransitionToNode, SessionFinalizer)")

	// Setup: Stub all dependencies for successful flow.
	f := newRuntimeTestFixture(t)
	_ = f

	// Action: Call Run("my-workflow", mockLogger)

	// Assert: Returns (0, nil)
}

func TestRun_LogsSessionTerminationNotification(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and dependency injection seams")

	// Setup: Same as successful completion, capture Logger calls.
	f := newRuntimeTestFixture(t)
	_ = f

	// Action: Call Run("my-workflow", mockLogger)

	// Assert: Logger.Info called with "received session termination notification"
}

// --- Error Propagation ---

func TestRun_SpectraFinderFails(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and SpectraFinder injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SpectraFinder.err = errors.New("no .spectra dir")
	f.SpectraFinder.result = ""
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	_ = assert.Equal
	_ = require.Error
	// exitCode == 1
	// err.Error() == "failed to locate project root: no .spectra dir"
}

func TestRun_PreSessionDependencyConstructionFails(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and pre-session dependency constructor seam (e.g., WorkflowDefinitionLoader constructor)")

	// Setup: SpectraFinder succeeds, but pre-session dependency constructor fails.
	f := newRuntimeTestFixture(t)
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// exitCode == 1
	// err.Error() contains "failed to initialize runtime dependencies: "
}

func TestRun_SessionInitializerFailsBeforeSession(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and SessionInitializer injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionInitializer.result = InitResult{
		PersistentSession: nil,
		Error:             errors.New("workflow not found"),
	}
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// exitCode == 1
	// err.Error() == "failed to initialize session: workflow not found"
	// SessionFinalizer NOT invoked
}

func TestRun_SessionInitializerFailsAfterSession(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and SessionInitializer injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	ps := newTestPersistentSession(t, f.Session)
	f.SessionInitializer.result = InitResult{
		PersistentSession: ps,
		Error:             errors.New("timeout during initialization"),
	}
	f.SessionFinalizer.result = 1
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// exitCode == 1 (from SessionFinalizer)
	// err.Error() == "failed to initialize session: timeout during initialization"
	// SessionFinalizer invoked with PersistentSession
}

func TestRun_PostSessionDependencyConstructionFails(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and post-session dependency constructor seam")

	// Setup: Successful initialization, but post-session dep construction fails.
	f := newRuntimeTestFixture(t)
	f.SessionFinalizer.result = 1
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// PersistentSession.Fail() called with RuntimeError{Issuer:"Runtime", Message:"failed to initialize post-session dependencies"}
	// exitCode from SessionFinalizer
	// err.Error() contains "failed to initialize post-session dependencies: "
}

func TestRun_CreateSocketFails(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and RuntimeSocketManager injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SocketManager.createSocketErr = errors.New("permission denied")
	f.SessionFinalizer.result = 1
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// PersistentSession.Fail() called with RuntimeError{Issuer:"Runtime", Message:"failed to create runtime socket"}
	// exitCode from SessionFinalizer
	// err.Error() contains "failed to create runtime socket: "
}

func TestRun_ListenFails(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and RuntimeSocketManager injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SocketManager.listenErr = errors.New("bind error")
	f.SessionFinalizer.result = 1
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// PersistentSession.Fail() called with RuntimeError{Issuer:"Runtime", Message:"failed to start socket listener"}
	// exitCode from SessionFinalizer
	// err.Error() contains "failed to start socket listener: "
}

func TestRun_InitialDispatchFails(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and TransitionToNode injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.TransitionToNode.transitionErr = errors.New("agent def not found")
	f.WorkflowDef.entryNode = "start"
	f.SessionFinalizer.result = 1
	_ = f

	// Action: exitCode, err := Run("wf", mockLogger)

	// Assert
	// PersistentSession.Fail() called with RuntimeError{Issuer:"Runtime", Message:"failed to dispatch entry node", FailingState:"start"}
	// exitCode from SessionFinalizer
	// err.Error() contains "failed to dispatch entry node: "
}

func TestRun_ListenerErrorDuringSession(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function, RuntimeSocketManager, and event loop seam")

	// Setup: Full successful startup.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.SessionFinalizer.result = 1
	// listenerDoneCh will be closed after DeleteSocket
	listenerDoneCh := make(chan struct{})
	f.SocketManager.listenDoneCh = listenerDoneCh
	f.SocketManager.deleteSocketFunc = func() {
		close(listenerDoneCh)
	}
	_ = f

	// Action: send error to listenerErrCh to trigger termination

	// Assert
	// PersistentSession.Fail() called with RuntimeError{Message:"listener error"}
	// Logger logs "listener error: <error>"
	// Proceeds to cleanup and SessionFinalizer
}

func TestRun_ListenerErrorWhenSessionAlreadyCompleted(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and event loop seam")

	// Setup: Full successful startup, session already completed.
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "completed"
	f.SessionFinalizer.result = 0
	_ = f

	// Action: send error to listenerErrCh

	// Assert
	// PersistentSession.Fail() NOT called
	// Proceeds to cleanup and SessionFinalizer
}

func TestRun_SessionFailed(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and event loop seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "failed"
	f.Session.getErrorResult = errors.New("agent crashed")
	f.SessionFinalizer.result = 1
	_ = f

	// Action: send to terminationNotifier to simulate session failure

	// Assert
	// exitCode == 1
	// err.Error() == "session failed: agent crashed"
}

func TestRun_SessionFinalizerNonTerminalStatus(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and event loop seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.getErrorResult = nil
	f.SessionFinalizer.result = 1
	_ = f

	// Action: send to terminationNotifier

	// Assert
	// exitCode == 1
	// err.Error() == "session terminated with non-terminal status"
}

// --- Mock / Dependency Interaction ---

func TestRun_TerminationNotifierCapacity(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function; need seam to capture terminationNotifier channel passed to SessionInitializer.Initialize()")

	// Setup: Capture the terminationNotifier channel.
	f := newRuntimeTestFixture(t)
	var capturedNotifier chan<- struct{}
	f.SessionInitializer.initializeFunc = func(wfName string, notifier chan<- struct{}) InitResult {
		capturedNotifier = notifier
		return InitResult{Error: errors.New("short-circuit")}
	}
	_ = capturedNotifier

	// Action: Run("wf", mockLogger)

	// Assert: cap(capturedNotifier) == 2
}

func TestRun_SessionInitializerReceivesWorkflowName(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and SessionInitializer injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionInitializer.result = InitResult{Error: errors.New("short-circuit")}

	// Action: Run("my-workflow", mockLogger)

	// Assert: f.SessionInitializer.capturedWorkflowName == "my-workflow"
	_ = f
}

func TestRun_InitialDispatchMessage(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and TransitionToNode injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.id = "uuid-123"
	f.WorkflowDef.entryNode = "start"
	f.SessionFinalizer.result = 0
	_ = f

	// Action: Run("wf", mockLogger)

	// Assert: TransitionToNode.Transition() called with ("start", msg)
	// where msg contains "uuid-123" and "spectra-agent event emit"
}

func TestRun_DeleteSocketCalledDuringCleanup(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and RuntimeSocketManager injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionFinalizer.result = 0
	_ = f

	// Action: Run("wf", mockLogger) (trigger terminationNotifier)

	// Assert: f.SocketManager.deleteSocketCalled == 1
}

func TestRun_SessionFinalizerCalledAfterCleanup(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and call ordering seam")

	// Setup
	f := newRuntimeTestFixture(t)
	tracker := f.SequenceTracker
	f.SocketManager.deleteSocketFunc = func() { tracker.Record("DeleteSocket") }
	// SessionFinalizer would record: tracker.Record("Finalize")
	_ = f

	// Action: Run("wf", mockLogger)

	// Assert: "Finalize" appears after "DeleteSocket" in tracker.Calls()
}

func TestRun_SessionFinalizerNotInvokedWhenNoSession(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and SpectraFinder injection seam")

	// Setup: SpectraFinder fails => no PersistentSession created
	f := newRuntimeTestFixture(t)
	f.SpectraFinder.err = errors.New("no .spectra dir")
	f.SpectraFinder.result = ""
	_ = f

	// Action: Run("wf", mockLogger)

	// Assert: f.SessionFinalizer.finalizeCalled == 0
}

func TestRun_CleanupOrder(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and call ordering seam with signal.Stop injection")

	// Setup
	f := newRuntimeTestFixture(t)
	tracker := f.SequenceTracker
	_ = tracker
	// Record: signal.Stop, DeleteSocket, listenerDoneCh wait, SessionFinalizer

	// Action: Run("wf", mockLogger)

	// Assert call order:
	// signal.Stop before DeleteSocket
	// DeleteSocket before listenerDoneCh wait
	// listenerDoneCh wait before SessionFinalizer.Finalize
}

func TestRun_PersistentSessionFailReturnsError(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function, event loop seam, and PersistentSession.Fail error handling")

	// Setup
	f := newRuntimeTestFixture(t)
	f.Session.getStatusResult = "running"
	f.Session.failErr = errors.New("session already in terminal state")
	f.SessionFinalizer.result = 1
	_ = f

	// Action: send error to listenerErrCh

	// Assert: Logger.Warn called with message containing
	// "attempted to fail session but session already in terminal state"
}

func TestRun_MessageRouterPassedToListen(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and RuntimeSocketManager injection seam")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionFinalizer.result = 0
	_ = f

	// Action: Run("wf", mockLogger)

	// Assert: f.SocketManager.capturedHandler is a MessageRouter instance
}

// --- State Transitions ---

func TestRun_OSSignalSIGINT(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and OS signal injection seam (signal channel)")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionFinalizer.result = 1
	_ = f

	// Action: inject SIGINT via mock signal channel

	// Assert:
	// PersistentSession.Fail() NOT called
	// Logger logs "received signal interrupt, initiating graceful shutdown"
	// Returns (1, error) where error is "session terminated by signal interrupt"
}

func TestRun_OSSignalSIGTERM(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and OS signal injection seam (signal channel)")

	// Setup
	f := newRuntimeTestFixture(t)
	f.SessionFinalizer.result = 1
	_ = f

	// Action: inject SIGTERM via mock signal channel

	// Assert:
	// PersistentSession.Fail() NOT called
	// Logger logs "received signal terminated, initiating graceful shutdown"
	// Returns (1, error) where error is "session terminated by signal terminated"
}

func TestRun_SecondSignalForcesExit(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function, OS signal injection seam, and slow DeleteSocket/blocked listenerDoneCh")

	// Setup
	f := newRuntimeTestFixture(t)
	// Block listenerDoneCh so cleanup doesn't complete
	f.SocketManager.listenDoneCh = make(chan struct{}) // never closed
	_ = f

	// Action: send first SIGINT, then second SIGINT during cleanup

	// Assert:
	// Logger logs "received second signal, forcing exit"
	// Returns (1, error) where error is "forced exit by second signal"
}

func TestRun_GracePeriodTimeout(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and fake timer/clock seam for 5-second grace period")

	// Setup
	f := newRuntimeTestFixture(t)
	// Block listenerDoneCh so cleanup hangs
	f.SocketManager.listenDoneCh = make(chan struct{}) // never closed
	_ = f

	// Action: fire fake 5-second timer during cleanup

	// Assert:
	// Logger logs "cleanup exceeded 5 second grace period, forcing exit"
	// Returns (1, error) where error is "cleanup timeout"
}

func TestRun_ListenerDoneChTimeout(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and fake timer/clock seam for 2-second sub-timeout")

	// Setup
	f := newRuntimeTestFixture(t)
	// listenerDoneCh never closes within 2 seconds (sub-timeout fires)
	f.SocketManager.listenDoneCh = make(chan struct{}) // never closed
	f.SessionFinalizer.result = 0
	_ = f

	// Action: fire fake 2-second sub-timer, but not 5-second grace timer

	// Assert:
	// Logger logs "listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer"
	// SessionFinalizer.Finalize() still invoked
}

func TestRun_ListenerDoneChAlreadyClosed(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function")

	// Setup: listenerDoneCh already closed
	f := newRuntimeTestFixture(t)
	alreadyClosed := make(chan struct{})
	close(alreadyClosed)
	f.SocketManager.listenDoneCh = alreadyClosed
	f.SessionFinalizer.result = 0
	_ = f

	// Action: Run completes without delay

	// Assert:
	// SessionFinalizer.Finalize() invoked without delay
	// No timeout warning logged
}

// --- Idempotency ---

func TestRun_DeleteSocketIdempotent(t *testing.T) {
	t.Skip("SCAFFOLDED: requires runtime.Run() function and RuntimeSocketManager injection seam")

	// Setup: CreateSocket returns error (socket never created)
	f := newRuntimeTestFixture(t)
	f.SocketManager.createSocketErr = errors.New("permission denied")
	// DeleteSocket is a no-op
	f.SessionFinalizer.result = 1
	_ = f

	// Action: Run("wf", mockLogger)

	// Assert:
	// No panic
	// Cleanup proceeds
	// SessionFinalizer.Finalize() invoked
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
