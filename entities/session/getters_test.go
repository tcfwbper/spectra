package session

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Happy Path — GetStatusSafe

func TestGetStatusSafe_Initializing(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	assert.Equal(t, "initializing", session.GetStatusSafe())
}

func TestGetStatusSafe_Running(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	assert.Equal(t, "running", session.GetStatusSafe())
}

func TestGetStatusSafe_Completed(t *testing.T) {
	session := createTestSession(t, "completed", "exit")
	assert.Equal(t, "completed", session.GetStatusSafe())
}

func TestGetStatusSafe_Failed(t *testing.T) {
	session := createTestSessionWithError(t, "failed", "processing", &RuntimeError{Issuer: "test", Message: "error"})
	assert.Equal(t, "failed", session.GetStatusSafe())
}

// Happy Path — GetCurrentStateSafe

func TestGetCurrentStateSafe_EntryNode(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	assert.Equal(t, "start", session.GetCurrentStateSafe())
}

func TestGetCurrentStateSafe_IntermediateNode(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	assert.Equal(t, "processing", session.GetCurrentStateSafe())
}

func TestGetCurrentStateSafe_ExitNode(t *testing.T) {
	session := createTestSession(t, "completed", "exit")
	assert.Equal(t, "exit", session.GetCurrentStateSafe())
}

// Happy Path — GetErrorSafe

func TestGetErrorSafe_NilWhenNoError(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	assert.Nil(t, session.GetErrorSafe())
}

func TestGetErrorSafe_AgentError(t *testing.T) {
	agentError := &AgentError{NodeName: "agent", Message: "agent failed"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	err := session.GetErrorSafe()
	assert.NotNil(t, err)
	assert.Equal(t, agentError, err)
}

func TestGetErrorSafe_RuntimeError(t *testing.T) {
	runtimeError := &RuntimeError{Issuer: "runtime", Message: "runtime failed"}
	session := createTestSessionWithError(t, "failed", "processing", runtimeError)

	err := session.GetErrorSafe()
	assert.NotNil(t, err)
	assert.Equal(t, runtimeError, err)
}

// Idempotency

func TestGetStatusSafe_Idempotent(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	status1 := session.GetStatusSafe()
	status2 := session.GetStatusSafe()
	status3 := session.GetStatusSafe()

	assert.Equal(t, "running", status1)
	assert.Equal(t, status1, status2)
	assert.Equal(t, status2, status3)
}

func TestGetCurrentStateSafe_Idempotent(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	state1 := session.GetCurrentStateSafe()
	state2 := session.GetCurrentStateSafe()

	assert.Equal(t, "processing", state1)
	assert.Equal(t, state1, state2)
}

func TestGetErrorSafe_Idempotent(t *testing.T) {
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	err1 := session.GetErrorSafe()
	err2 := session.GetErrorSafe()

	assert.Equal(t, agentError, err1)
	assert.Equal(t, err1, err2)
	// Same pointer
	assert.True(t, err1 == err2)
}

// Concurrent Behaviour

func TestGetters_ConcurrentGetStatusSafe(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status := session.GetStatusSafe()
			assert.Equal(t, "running", status)
		}()
	}

	wg.Wait()
}

func TestGetters_ConcurrentGetCurrentStateSafe(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			state := session.GetCurrentStateSafe()
			assert.Equal(t, "processing", state)
		}()
	}

	wg.Wait()
}

func TestGetters_ConcurrentGetErrorSafe(t *testing.T) {
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := session.GetErrorSafe()
			assert.Equal(t, agentError, err)
		}()
	}

	wg.Wait()
}

// Invariants — No Mutation

func TestGetStatusSafe_NoMutation(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt

	_ = session.GetStatusSafe()

	assert.Equal(t, oldUpdatedAt, session.UpdatedAt)
}

func TestGetCurrentStateSafe_NoMutation(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt

	_ = session.GetCurrentStateSafe()

	assert.Equal(t, oldUpdatedAt, session.UpdatedAt)
}

func TestGetErrorSafe_NoMutation(t *testing.T) {
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)
	oldUpdatedAt := session.UpdatedAt

	_ = session.GetErrorSafe()

	assert.Equal(t, oldUpdatedAt, session.UpdatedAt)
}

// Edge Cases

func TestGetStatusSafe_BeforeRun(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	assert.Equal(t, "initializing", session.GetStatusSafe())
}

func TestGetCurrentStateSafe_NeverEmpty(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	state := session.GetCurrentStateSafe()
	assert.NotEmpty(t, state)
	assert.Equal(t, "start", state)
}

func TestGetErrorSafe_NilBeforeFailure(t *testing.T) {
	session := createTestSession(t, "initializing", "start")
	assert.Nil(t, session.GetErrorSafe())

	session2 := createTestSession(t, "running", "processing")
	assert.Nil(t, session2.GetErrorSafe())
}
