package session

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Happy Path — UpdateCurrentStateSafe

func TestUpdateCurrentStateSafe_UpdatesState(t *testing.T) {
	session := createTestSession(t, "running", "node1")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err)
	assert.Equal(t, "node2", session.CurrentState)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestUpdateCurrentStateSafe_MultipleUpdates(t *testing.T) {
	session := createTestSession(t, "running", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err1 := session.UpdateCurrentStateSafe("processing")
	err2 := session.UpdateCurrentStateSafe("review")
	err3 := session.UpdateCurrentStateSafe("complete")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, "complete", session.CurrentState)
}

func TestUpdateCurrentStateSafe_PersistsToStore(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err)
	session.metadataStore.AssertCalled(t, "Write", mock.Anything)
	assert.Equal(t, "node2", session.CurrentState)
}

func TestUpdateCurrentStateSafe_AcceptsAnyNonEmptyString(t *testing.T) {
	session := createTestSession(t, "running", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("UnknownNodeName")

	assert.NoError(t, err)
	assert.Equal(t, "UnknownNodeName", session.CurrentState)
}

// Happy Path — Self-Transition

func TestUpdateCurrentStateSafe_SelfTransition(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("processing")

	assert.NoError(t, err)
	assert.Equal(t, "processing", session.CurrentState)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

// Validation Failures — Empty Input

func TestUpdateCurrentStateSafe_EmptyNewState(t *testing.T) {
	session := createTestSession(t, "running", "node1")
	oldUpdatedAt := session.UpdatedAt

	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateCurrentStateSafe("")

	assert.NoError(t, err)
	assert.Equal(t, "node1", session.CurrentState)
	assert.Equal(t, oldUpdatedAt, session.UpdatedAt)

	session.logger.AssertCalled(t, "Warning", mock.MatchedBy(func(msg string) bool {
		return strings.Contains(strings.ToLower(msg), "empty newstate") ||
			strings.Contains(strings.ToLower(msg), "state unchanged")
	}))
}

// Idempotency

func TestUpdateCurrentStateSafe_Idempotent(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err1 := session.UpdateCurrentStateSafe("node2")
	oldUpdatedAt1 := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err2 := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "node2", session.CurrentState)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt1)
}

// Concurrent Behaviour

func TestCurrentState_ConcurrentUpdates(t *testing.T) {
	session := createTestSession(t, "running", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_ = session.UpdateCurrentStateSafe(string(rune('a' + index)))
		}(i)
	}

	wg.Wait()

	// Final state is one of the written values
	assert.NotEqual(t, "start", session.CurrentState)
}

func TestCurrentState_ConcurrentReadWrite(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			state := session.GetCurrentStateSafe()
			assert.True(t, state == "node1" || state == "node2")
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = session.UpdateCurrentStateSafe("node2")
	}()

	wg.Wait()
}

// Error Propagation

func TestUpdateCurrentStateSafe_PersistenceFailureLogged(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err)
	assert.Equal(t, "node2", session.CurrentState)
	session.logger.AssertCalled(t, "Warning", mock.MatchedBy(func(msg string) bool {
		return strings.Contains(strings.ToLower(msg), "persistence failed")
	}))
}

func TestUpdateCurrentStateSafe_PersistenceFailureDoesNotRevert(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err)
	assert.Equal(t, "node2", session.CurrentState)
}

// Invariants — Always Returns Nil

func TestUpdateCurrentStateSafe_AlwaysReturnsNil(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err)
}

func TestUpdateCurrentStateSafe_AlwaysReturnsNilOnEmpty(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateCurrentStateSafe("")

	assert.NoError(t, err)
}

// Invariants — UpdatedAt Refresh

func TestUpdateCurrentStateSafe_RefreshesUpdatedAt(t *testing.T) {
	session := createTestSession(t, "running", "node1")
	oldUpdatedAt := session.UpdatedAt

	time.Sleep(1 * time.Second)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("node2")

	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestUpdateCurrentStateSafe_EmptyInputDoesNotRefreshUpdatedAt(t *testing.T) {
	session := createTestSession(t, "running", "node1")
	oldUpdatedAt := session.UpdatedAt

	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateCurrentStateSafe("")

	assert.NoError(t, err)
	assert.Equal(t, oldUpdatedAt, session.UpdatedAt)
}

// Invariants — No Workflow Validation

func TestUpdateCurrentStateSafe_NoWorkflowValidation(t *testing.T) {
	session := createTestSession(t, "running", "validNode")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("NonExistentNode")

	assert.NoError(t, err)
	assert.Equal(t, "NonExistentNode", session.CurrentState)
}

func TestUpdateCurrentStateSafe_AcceptsInvalidNodeName(t *testing.T) {
	session := createTestSession(t, "running", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("!!!invalid!!!")

	assert.NoError(t, err)
	assert.Equal(t, "!!!invalid!!!", session.CurrentState)
}

// Edge Cases

func TestUpdateCurrentStateSafe_WhitespaceOnlyState(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("   ")

	assert.NoError(t, err)
	assert.Equal(t, "   ", session.CurrentState)
}

func TestUpdateCurrentStateSafe_VeryLongStateName(t *testing.T) {
	session := createTestSession(t, "running", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	longString := strings.Repeat("a", 10*1024)
	err := session.UpdateCurrentStateSafe(longString)

	assert.NoError(t, err)
	assert.Equal(t, longString, session.CurrentState)
}

func TestUpdateCurrentStateSafe_UnicodeStateName(t *testing.T) {
	session := createTestSession(t, "running", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("处理节点")

	assert.NoError(t, err)
	assert.Equal(t, "处理节点", session.CurrentState)
}

func TestUpdateCurrentStateSafe_SpecialCharactersInStateName(t *testing.T) {
	session := createTestSession(t, "running", "node1")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateCurrentStateSafe("node-2_final.state")

	assert.NoError(t, err)
	assert.Equal(t, "node-2_final.state", session.CurrentState)
}
