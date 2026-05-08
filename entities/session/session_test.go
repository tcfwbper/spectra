package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewSession_ValidInputs(t *testing.T) {
	s, err := NewSession(testUUID, "my-workflow", "start", int64(1700000000))

	require.NoError(t, err)
	require.NotNil(t, s)

	assert.Equal(t, testUUID, s.ID)
	assert.Equal(t, "my-workflow", s.WorkflowName)
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
	s, err := NewSession(testUUID, "w", "n", 1)

	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, int64(1), s.CreatedAt)
}

// --- Validation Failures — id ---

func TestNewSession_InvalidUUID_Empty(t *testing.T) {
	s, err := NewSession("", "w", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "invalid session ID: must be a valid UUID", err.Error())
}

func TestNewSession_InvalidUUID_Malformed(t *testing.T) {
	s, err := NewSession("not-a-uuid", "w", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "invalid session ID: must be a valid UUID", err.Error())
}

func TestNewSession_InvalidUUID_TooShort(t *testing.T) {
	s, err := NewSession("550e8400-e29b-41d4", "w", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "invalid session ID: must be a valid UUID", err.Error())
}

// --- Validation Failures — workflowName ---

func TestNewSession_EmptyWorkflowName(t *testing.T) {
	s, err := NewSession(testUUID, "", "start", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "workflow name cannot be empty", err.Error())
}

// --- Validation Failures — entryNode ---

func TestNewSession_EmptyEntryNode(t *testing.T) {
	s, err := NewSession(testUUID, "w", "", testCreatedAt)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "entry node cannot be empty", err.Error())
}

// --- Validation Failures — createdAt ---

func TestNewSession_CreatedAtZero(t *testing.T) {
	s, err := NewSession(testUUID, "w", "n", 0)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "createdAt must be a positive POSIX timestamp", err.Error())
}

func TestNewSession_CreatedAtNegative(t *testing.T) {
	s, err := NewSession(testUUID, "w", "n", -1)

	assert.Nil(t, s)
	require.Error(t, err)
	assert.Equal(t, "createdAt must be a positive POSIX timestamp", err.Error())
}
