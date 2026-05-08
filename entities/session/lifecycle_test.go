package session

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Run ---

func TestRun_FromInitializing(t *testing.T) {
	s := newTestSession(t)

	err := s.Run()

	require.NoError(t, err)
	assert.Equal(t, "running", s.GetStatusSafe())
	assert.GreaterOrEqual(t, s.GetMetadataSnapshotSafe().UpdatedAt, s.CreatedAt)
}

// --- State Transitions — Run ---

func TestRun_FromRunning_ReturnsError(t *testing.T) {
	s := newRunningSession(t)

	err := s.Run()

	require.Error(t, err)
	assert.Equal(t, "cannot run session: status is 'running', expected 'initializing'", err.Error())
}

func TestRun_FromCompleted_ReturnsError(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()
	_ = s.Done(ch)

	err := s.Run()

	require.Error(t, err)
	assert.Equal(t, "cannot run session: status is 'completed', expected 'initializing'", err.Error())
}

func TestRun_FromFailed_ReturnsError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr := newTestAgentError(t)
	_ = s.Fail(agentErr, ch)

	err := s.Run()

	require.Error(t, err)
	assert.Equal(t, "cannot run session: status is 'failed', expected 'initializing'", err.Error())
}

// --- Happy Path — Done ---

func TestDone_FromRunning(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()

	err := s.Done(ch)

	require.NoError(t, err)
	assert.Equal(t, "completed", s.GetStatusSafe())

	// Verify notification was sent
	select {
	case <-ch:
		// OK: received notification
	default:
		t.Fatal("expected notification on terminationNotifier channel")
	}

	// Verify no second notification
	select {
	case <-ch:
		t.Fatal("unexpected second notification on terminationNotifier channel")
	default:
		// OK: no extra notification
	}

	assert.GreaterOrEqual(t, s.GetMetadataSnapshotSafe().UpdatedAt, s.CreatedAt)
}

// --- State Transitions — Done ---

func TestDone_FromInitializing_ReturnsError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()

	err := s.Done(ch)

	require.Error(t, err)
	assert.Equal(t, "cannot complete session: status is 'initializing', expected 'running'", err.Error())
}

func TestDone_FromCompleted_ReturnsError(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()
	_ = s.Done(ch)

	err := s.Done(ch)

	require.Error(t, err)
	assert.Equal(t, "cannot complete session: status is 'completed', expected 'running'", err.Error())
}

func TestDone_FromFailed_ReturnsError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr := newTestAgentError(t)
	_ = s.Fail(agentErr, ch)

	err := s.Done(ch)

	require.Error(t, err)
	assert.Equal(t, "cannot complete session: status is 'failed', expected 'running'", err.Error())
}

// --- Happy Path — Fail ---

func TestFail_FromInitializing_WithAgentError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr := newTestAgentError(t)

	err := s.Fail(agentErr, ch)

	require.NoError(t, err)
	assert.Equal(t, "failed", s.GetStatusSafe())
	assert.Same(t, agentErr, s.GetErrorSafe())

	// Verify notification was sent
	select {
	case <-ch:
		// OK
	default:
		t.Fatal("expected notification on terminationNotifier channel")
	}
}

func TestFail_FromRunning_WithRuntimeError(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()
	runtimeErr := newTestRuntimeError(t)

	err := s.Fail(runtimeErr, ch)

	require.NoError(t, err)
	assert.Equal(t, "failed", s.GetStatusSafe())
	assert.Same(t, runtimeErr, s.GetErrorSafe())

	// Verify notification
	select {
	case <-ch:
		// OK
	default:
		t.Fatal("expected notification on terminationNotifier channel")
	}
}

// --- Validation Failures — Fail ---

func TestFail_NilError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()

	err := s.Fail(nil, ch)

	require.Error(t, err)
	assert.Equal(t, "error cannot be nil", err.Error())
	assert.Equal(t, "initializing", s.GetStatusSafe())
}

func TestFail_InvalidErrorType(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()

	err := s.Fail(errors.New("plain"), ch)

	require.Error(t, err)
	assert.Equal(t, "invalid error type: must be *AgentError or *RuntimeError", err.Error())
	assert.Equal(t, "initializing", s.GetStatusSafe())
}

// --- State Transitions — Fail ---

func TestFail_FromFailed_ReturnsError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr1 := newTestAgentError(t)
	agentErr2 := newTestAgentError2(t)
	_ = s.Fail(agentErr1, ch)

	err := s.Fail(agentErr2, ch)

	require.Error(t, err)
	assert.Equal(t, "session already failed", err.Error())
	assert.Same(t, agentErr1, s.GetErrorSafe())
}

func TestFail_FromCompleted_ReturnsError(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()
	_ = s.Done(ch)
	agentErr := newTestAgentError(t)

	err := s.Fail(agentErr, ch)

	require.Error(t, err)
	assert.Equal(t, "cannot fail session: status is 'completed', workflow already terminated", err.Error())
}
