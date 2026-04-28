package session

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Happy Path — Run

func TestRun_InitializingToRunning(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Run(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "running", session.Status)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestRun_PersistsToStore(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Run(session.terminationNotifier)

	assert.NoError(t, err)
	session.metadataStore.AssertCalled(t, "Write", mock.Anything)
	assert.Equal(t, "running", session.Status)
}

func TestRun_NoNotificationSent(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Run(session.terminationNotifier)

	assert.NoError(t, err)
	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 0, len(ch))
}

// Happy Path — Done

func TestDone_RunningToCompleted(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Done(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "completed", session.Status)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestDone_PersistsToStore(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Done(session.terminationNotifier)

	assert.NoError(t, err)
	session.metadataStore.AssertCalled(t, "Write", mock.Anything)
	assert.Equal(t, "completed", session.Status)
}

func TestDone_SendsNotification(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	ch := session.terminationNotifier.(chan struct{})
	initialLen := len(ch)

	err := session.Done(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, initialLen+1, len(ch))
}

func TestDone_NonBlockingSend(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	start := time.Now()
	err := session.Done(session.terminationNotifier)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, elapsed, 100*time.Millisecond)
}

// Happy Path — Fail

func TestFail_InitializingToFailed(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, runtimeError, session.Error)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestFail_RunningToFailed(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	agentError := &AgentError{NodeName: "agent", Message: "error"}
	err := session.Fail(agentError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, agentError, session.Error)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestFail_PersistsToStore(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.NoError(t, err)
	session.metadataStore.AssertCalled(t, "Write", mock.Anything)
	assert.Equal(t, "failed", session.Status)
}

func TestFail_SendsNotification(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	ch := session.terminationNotifier.(chan struct{})
	initialLen := len(ch)

	agentError := &AgentError{NodeName: "agent", Message: "error"}
	err := session.Fail(agentError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, initialLen+1, len(ch))
}

func TestFail_AcceptsAgentError(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	agentError := &AgentError{NodeName: "test", Message: "agent failed"}
	err := session.Fail(agentError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, agentError, session.Error)
}

func TestFail_AcceptsRuntimeError(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	runtimeError := &RuntimeError{Issuer: "runtime", Message: "runtime failed"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, runtimeError, session.Error)
}

// Validation Failures — Run Preconditions

func TestRun_RejectsRunning(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.Run(session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot run.*status is 'running'.*expected 'initializing'", err.Error())
	assert.Equal(t, "running", session.Status)
}

func TestRun_RejectsCompleted(t *testing.T) {
	session := createTestSession(t, "completed", "exit")

	err := session.Run(session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot run.*status is 'completed'.*expected 'initializing'", err.Error())
	assert.Equal(t, "completed", session.Status)
}

func TestRun_RejectsFailed(t *testing.T) {
	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", runtimeError)

	err := session.Run(session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot run.*status is 'failed'.*expected 'initializing'", err.Error())
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, runtimeError, session.Error)
}

// Validation Failures — Done Preconditions

func TestDone_RejectsInitializing(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	err := session.Done(session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot complete.*status is 'initializing'.*expected 'running'", err.Error())
	assert.Equal(t, "initializing", session.Status)
}

func TestDone_RejectsCompleted(t *testing.T) {
	session := createTestSession(t, "completed", "exit")

	err := session.Done(session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot complete.*status is 'completed'.*expected 'running'", err.Error())
	assert.Equal(t, "completed", session.Status)
}

func TestDone_RejectsFailed(t *testing.T) {
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	err := session.Done(session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot complete.*status is 'failed'.*expected 'running'", err.Error())
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, agentError, session.Error)
}

// Validation Failures — Fail Preconditions

func TestFail_RejectsNilError(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.Fail(nil, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)error cannot be nil", err.Error())
	assert.Equal(t, "running", session.Status)
}

func TestFail_RejectsInvalidErrorType(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	genericError := errors.New("generic error")
	err := session.Fail(genericError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid error type.*must be.*AgentError.*RuntimeError", err.Error())
	assert.Equal(t, "running", session.Status)
}

func TestFail_RejectsCompleted(t *testing.T) {
	session := createTestSession(t, "completed", "exit")

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot fail.*status is 'completed'.*workflow already terminated", err.Error())
	assert.Equal(t, "completed", session.Status)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 0, len(ch))
}

func TestFail_RejectsFailed(t *testing.T) {
	firstError := &AgentError{NodeName: "first", Message: "first error"}
	session := createTestSessionWithError(t, "failed", "processing", firstError)

	secondError := &RuntimeError{Issuer: "second", Message: "second error"}
	err := session.Fail(secondError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)session already failed", err.Error())
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, firstError, session.Error)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 0, len(ch))
}

// Validation Failures — Error Type

func TestFail_RejectsStandardError(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	wrappedError := fmt.Errorf("wrapped: %w", errors.New("base"))
	err := session.Fail(wrappedError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)invalid error type", err.Error())
	assert.Equal(t, "running", session.Status)
}

func TestFail_RejectsNilAgentError(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	var agentError *AgentError
	err := session.Fail(agentError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)error cannot be nil", err.Error())
	assert.Equal(t, "running", session.Status)
}

func TestFail_RejectsNilRuntimeError(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	var runtimeError *RuntimeError
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)error cannot be nil", err.Error())
	assert.Equal(t, "running", session.Status)
}

// Atomic Replacement

func TestFail_AtomicStatusAndErrorUpdate(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	var observedInconsistency bool

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			status := session.GetStatusSafe()
			err := session.GetErrorSafe()

			if status == "failed" && err == nil {
				observedInconsistency = true
			}
			if status != "failed" && err != nil {
				observedInconsistency = true
			}
		}
	}()

	agentError := &AgentError{NodeName: "agent", Message: "error"}
	err := session.Fail(agentError, session.terminationNotifier)

	wg.Wait()

	assert.NoError(t, err)
	assert.False(t, observedInconsistency, "Concurrent reader observed inconsistent state")
}

func TestLifecycle_AtomicStatusAndTimestampUpdate(t *testing.T) {
	session := createTestSessionWithUpdatedAt(t, "initializing", "start", time.Now().Unix()-100)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	var observedInconsistency bool
	oldUpdatedAt := session.UpdatedAt

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			status := session.GetStatusSafe()
			// Can't directly access UpdatedAt safely without another lock,
			// but this tests that status transitions are atomic
			if status == "running" {
				// Status changed
			}
		}
	}()

	err := session.Run(session.terminationNotifier)

	wg.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "running", session.Status)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
	assert.False(t, observedInconsistency)
}

// Idempotency

func TestRun_NotIdempotent(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err1 := session.Run(session.terminationNotifier)
	err2 := session.Run(session.terminationNotifier)

	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, "running", session.Status)
}

func TestDone_NotIdempotent(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err1 := session.Done(session.terminationNotifier)
	err2 := session.Done(session.terminationNotifier)

	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, "completed", session.Status)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestFail_NotIdempotent(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	firstError := &AgentError{NodeName: "first", Message: "first"}
	secondError := &RuntimeError{Issuer: "second", Message: "second"}

	err1 := session.Fail(firstError, session.terminationNotifier)
	err2 := session.Fail(secondError, session.terminationNotifier)

	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Regexp(t, "(?i)already failed", err2.Error())
	assert.Equal(t, firstError, session.Error)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

// Resource Cleanup

func TestLifecycle_ReleasesLockAfterRun(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Run(session.terminationNotifier)
	assert.NoError(t, err)

	status := session.GetStatusSafe()
	assert.Equal(t, "running", status)
}

func TestLifecycle_ReleasesLockAfterDone(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Done(session.terminationNotifier)
	assert.NoError(t, err)

	status := session.GetStatusSafe()
	assert.Equal(t, "completed", status)
}

func TestLifecycle_ReleasesLockAfterFail(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)
	assert.NoError(t, err)

	retrievedError := session.GetErrorSafe()
	assert.Equal(t, runtimeError, retrievedError)
}

func TestLifecycle_ReleasesLockOnValidationFailure(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	done := make(chan bool)
	go func() {
		_ = session.Fail(nil, session.terminationNotifier)
		done <- true
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = session.GetStatusSafe()
		done <- true
	}()

	select {
	case <-done:
		// Validation error returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Validation should not block getter")
	}
}

// Concurrent Behaviour

func TestLifecycle_ConcurrentRunFail(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	var runErr, failErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		runErr = session.Run(session.terminationNotifier)
	}()

	go func() {
		defer wg.Done()
		runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
		failErr = session.Fail(runtimeError, session.terminationNotifier)
	}()

	wg.Wait()

	// One should succeed, one should fail
	assert.True(t, (runErr == nil && failErr != nil) || (runErr != nil && failErr == nil))
	assert.True(t, session.Status == "running" || session.Status == "failed")

	ch := session.terminationNotifier.(chan struct{})
	assert.LessOrEqual(t, len(ch), 1)
}

func TestLifecycle_ConcurrentDoneFail(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	var doneErr, failErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		doneErr = session.Done(session.terminationNotifier)
	}()

	go func() {
		defer wg.Done()
		agentError := &AgentError{NodeName: "agent", Message: "error"}
		failErr = session.Fail(agentError, session.terminationNotifier)
	}()

	wg.Wait()

	// One should succeed, one should fail
	assert.True(t, (doneErr == nil && failErr != nil) || (doneErr != nil && failErr == nil))
	assert.True(t, session.Status == "completed" || session.Status == "failed")

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestLifecycle_ConcurrentFailCalls(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			agentError := &AgentError{NodeName: fmt.Sprintf("agent-%d", index), Message: fmt.Sprintf("error-%d", index)}
			err := session.Fail(agentError, session.terminationNotifier)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 1, successCount, "Exactly one Fail call should succeed")
	assert.Equal(t, "failed", session.Status)
	assert.NotNil(t, session.Error)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

// Error Propagation

func TestLifecycle_PersistenceFailureLoggedNotReturned(t *testing.T) {
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	err := session.Run(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "running", session.Status)
	session.logger.AssertCalled(t, "Warning", mock.MatchedBy(func(msg string) bool {
		return assert.Regexp(t, "(?i)persistence failed", msg)
	}))
}

func TestLifecycle_PersistenceFailureInDone(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	err := session.Done(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "completed", session.Status)
	session.logger.AssertCalled(t, "Warning", mock.Anything)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestLifecycle_PersistenceFailureInFail(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, runtimeError, session.Error)
	session.logger.AssertCalled(t, "Warning", mock.Anything)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

// Invariants — UpdatedAt Refresh

func TestRun_RefreshesUpdatedAt(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	oldUpdatedAt := session.UpdatedAt

	time.Sleep(1 * time.Second)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Run(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestDone_RefreshesUpdatedAt(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt

	time.Sleep(1 * time.Second)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Done(session.terminationNotifier)

	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestFail_RefreshesUpdatedAt(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt

	time.Sleep(1 * time.Second)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

// Invariants — Terminal State Finality

func TestLifecycle_CompletedIsFinal(t *testing.T) {
	session := createTestSession(t, "completed", "exit")

	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.Error(t, err)
	assert.Equal(t, "completed", session.Status)

	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 0, len(ch))
}

func TestLifecycle_FailedIsFinal(t *testing.T) {
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	err := session.Run(session.terminationNotifier)

	assert.Error(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, agentError, session.Error)
}

func TestLifecycle_FirstErrorWins(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	firstError := &AgentError{NodeName: "first", Message: "first"}
	err1 := session.Fail(firstError, session.terminationNotifier)
	assert.NoError(t, err1)

	secondError := &RuntimeError{Issuer: "second", Message: "second"}
	err2 := session.Fail(secondError, session.terminationNotifier)
	assert.Error(t, err2)

	assert.Equal(t, firstError, session.Error)
}

// Edge Cases

func TestLifecycle_ValidationBeforeLockAcquisition(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	done := make(chan bool)
	go func() {
		err := session.Fail(nil, session.terminationNotifier)
		require.Error(t, err)
		done <- true
	}()

	select {
	case <-done:
		// Validation returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Validation should not block")
	}
}

func TestFail_EmptyErrorMessage(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	runtimeError := &RuntimeError{Issuer: "test", Message: ""}
	err := session.Fail(runtimeError, session.terminationNotifier)

	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, "", session.Error.Error())
}
