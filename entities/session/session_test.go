package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewSession_ValidInputs(t *testing.T) {
	// Scaffolded: NewSession signature will add pid parameter (int, position 4).
	// Once production adds pid param, change call to: NewSession(testUUID, "my-workflow", "start", 1234, int64(1700000000))
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update to NewSession(id, workflowName, entryNode, pid int, createdAt int64)")

	s, err := NewSession(testUUID, "my-workflow", "start", int64(1700000000))

	require.NoError(t, err)
	require.NotNil(t, s)

	assert.Equal(t, testUUID, s.ID)
	assert.Equal(t, "my-workflow", s.WorkflowName)
	// assert.Equal(t, 1234, s.Pid) // TODO: uncomment when Pid field exists
	assert.Equal(t, "initializing", s.Status)
	assert.Equal(t, int64(1700000000), s.CreatedAt)
	assert.Equal(t, s.CreatedAt, s.UpdatedAt)
	assert.Equal(t, "start", s.CurrentState)
	assert.NotNil(t, s.SessionData)
	assert.Empty(t, s.SessionData)
	assert.Nil(t, s.Error)
	assert.NotNil(t, s.EventHistory)
	assert.Empty(t, s.EventHistory)
}

func TestNewSession_MinimalCreatedAt(t *testing.T) {
	// Scaffolded: NewSession signature will add pid parameter.
	// Once production adds pid param, change call to: NewSession(testUUID, "w", "n", 1, 1)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update to NewSession(id, workflowName, entryNode, pid int, createdAt int64)")

	s, err := NewSession(testUUID, "w", "n", 1)

	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, int64(1), s.CreatedAt)
}

// --- Validation Failures — id ---

func TestNewSession_InvalidUUID_Empty(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession("", "w", "start", 1, testCreatedAt)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession("", "w", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "invalid session ID: must be a valid UUID", err.Error())
}

func TestNewSession_InvalidUUID_Malformed(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession("not-a-uuid", "w", "start", 1, testCreatedAt)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession("not-a-uuid", "w", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "invalid session ID: must be a valid UUID", err.Error())
}

func TestNewSession_InvalidUUID_TooShort(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession("550e8400-e29b-41d4", "w", "start", 1, testCreatedAt)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession("550e8400-e29b-41d4", "w", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "invalid session ID: must be a valid UUID", err.Error())
}

// --- Validation Failures — workflowName ---

func TestNewSession_EmptyWorkflowName(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession(testUUID, "", "start", 1, testCreatedAt)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession(testUUID, "", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "workflow name cannot be empty", err.Error())
}

// --- Validation Failures — entryNode ---

func TestNewSession_EmptyEntryNode(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession(testUUID, "w", "", 1, testCreatedAt)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession(testUUID, "w", "", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "entry node cannot be empty", err.Error())
}

// --- Validation Failures — pid ---

func TestNewSession_PidZero(t *testing.T) {
	// Scaffolded: NewSession(testUUID, "w", "n", 0, testCreatedAt) — pid=0 rejected.
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update to NewSession(id, workflowName, entryNode, pid int, createdAt int64)")

	// Once production surface exists:
	// s, err := NewSession(testUUID, "w", "n", 0, testCreatedAt)
	// assert.Nil(t, s)
	// require.Error(t, err)
	// assert.Equal(t, "pid must be a positive integer", err.Error())
}

func TestNewSession_PidNegative(t *testing.T) {
	// Scaffolded: NewSession(testUUID, "w", "n", -1, testCreatedAt) — pid=-1 rejected.
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update to NewSession(id, workflowName, entryNode, pid int, createdAt int64)")

	// Once production surface exists:
	// s, err := NewSession(testUUID, "w", "n", -1, testCreatedAt)
	// assert.Nil(t, s)
	// require.Error(t, err)
	// assert.Equal(t, "pid must be a positive integer", err.Error())
}

// --- Validation Failures — createdAt ---

func TestNewSession_CreatedAtZero(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession(testUUID, "w", "n", 1, 0)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession(testUUID, "w", "n", 0)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "createdAt must be a positive POSIX timestamp", err.Error())
}

func TestNewSession_CreatedAtNegative(t *testing.T) {
	// Scaffolded: Once pid param is added, update to: NewSession(testUUID, "w", "n", 1, -1)
	t.Skip("blocked: NewSession does not yet accept pid parameter — awaiting production surface update")

	s, err := NewSession(testUUID, "w", "n", -1)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "createdAt must be a positive POSIX timestamp", err.Error())
}
