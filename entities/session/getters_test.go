package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — GetStatusSafe ---

func TestGetStatusSafe_Initializing(t *testing.T) {
	s := newTestSession(t)

	assert.Equal(t, "initializing", s.GetStatusSafe())
}

func TestGetStatusSafe_Running(t *testing.T) {
	s := newRunningSession(t)

	assert.Equal(t, "running", s.GetStatusSafe())
}

func TestGetStatusSafe_Completed(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()
	err := s.Done(ch)
	require.NoError(t, err)

	assert.Equal(t, "completed", s.GetStatusSafe())
}

func TestGetStatusSafe_Failed(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr := newTestAgentError(t)
	err := s.Fail(agentErr, ch)
	require.NoError(t, err)

	assert.Equal(t, "failed", s.GetStatusSafe())
}

// --- Happy Path — GetCurrentStateSafe ---

func TestGetCurrentStateSafe_Initial(t *testing.T) {
	s := newTestSession(t)

	assert.Equal(t, "start", s.GetCurrentStateSafe())
}

func TestGetCurrentStateSafe_AfterUpdate(t *testing.T) {
	s := newTestSession(t)
	_ = s.UpdateCurrentStateSafe("processing")

	assert.Equal(t, "processing", s.GetCurrentStateSafe())
}

// --- Happy Path — GetErrorSafe ---

func TestGetErrorSafe_NoError(t *testing.T) {
	s := newTestSession(t)

	assert.Nil(t, s.GetErrorSafe())
}

func TestGetErrorSafe_AgentError(t *testing.T) {
	s := newTestSession(t)
	ch := newTerminationChannel()
	agentErr := newTestAgentError(t)
	_ = s.Fail(agentErr, ch)

	result := s.GetErrorSafe()

	assert.Same(t, agentErr, result)
}

func TestGetErrorSafe_RuntimeError(t *testing.T) {
	s := newRunningSession(t)
	ch := newTerminationChannel()
	runtimeErr := newTestRuntimeError(t)
	_ = s.Fail(runtimeErr, ch)

	result := s.GetErrorSafe()

	assert.Same(t, runtimeErr, result)
}

// --- Happy Path — GetMetadataSnapshotSafe ---

func TestGetMetadataSnapshotSafe_ReturnsAllFields(t *testing.T) {
	s := newRunningSession(t)

	snapshot := s.GetMetadataSnapshotSafe()

	assert.Equal(t, testUUID, snapshot.ID)
	assert.Equal(t, testWorkflow, snapshot.WorkflowName)
	assert.Equal(t, "running", snapshot.Status)
	assert.Equal(t, testCreatedAt, snapshot.CreatedAt)
	assert.GreaterOrEqual(t, snapshot.UpdatedAt, snapshot.CreatedAt)
	assert.Equal(t, testEntryNode, snapshot.CurrentState)
	assert.NotNil(t, snapshot.SessionData)
	assert.Nil(t, snapshot.Error)
}

func TestGetMetadataSnapshotSafe_EmptySessionData(t *testing.T) {
	s := newTestSession(t)

	snapshot := s.GetMetadataSnapshotSafe()

	assert.NotNil(t, snapshot.SessionData)
	assert.Len(t, snapshot.SessionData, 0)
}

func TestGetMetadataSnapshotSafe_WithSessionData(t *testing.T) {
	s := newTestSession(t)
	_ = s.UpdateSessionDataSafe("k", "v")

	snapshot := s.GetMetadataSnapshotSafe()

	assert.Equal(t, "v", snapshot.SessionData["k"])
}

// --- Data Independence (Copy Semantics) ---

func TestGetMetadataSnapshotSafe_MapIsolation(t *testing.T) {
	s := newTestSession(t)
	_ = s.UpdateSessionDataSafe("k", "v")
	snapshot := s.GetMetadataSnapshotSafe()

	// Mutate the returned map
	snapshot.SessionData["k"] = "modified"

	// Original session data is unaffected
	val, ok := s.GetSessionDataSafe("k")
	assert.True(t, ok)
	assert.Equal(t, "v", val)
}

func TestGetMetadataSnapshotSafe_InsertionIsolation(t *testing.T) {
	s := newTestSession(t)
	snapshot := s.GetMetadataSnapshotSafe()

	// Insert into the returned map
	snapshot.SessionData["new"] = "x"

	// Original session does not have the new key
	val, ok := s.GetSessionDataSafe("new")
	assert.False(t, ok)
	assert.Nil(t, val)
}
