package session

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — UpdateSessionDataSafe ---

func TestUpdateSessionDataSafe_StringValue(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("logicSpec.output", "hello")

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("logicSpec.output")
	assert.True(t, ok)
	assert.Equal(t, "hello", val)
}

func TestUpdateSessionDataSafe_NilValue(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("someKey", nil)

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("someKey")
	assert.True(t, ok)
	assert.Nil(t, val)
}

func TestUpdateSessionDataSafe_OverwriteExisting(t *testing.T) {
	s := newTestSession(t)
	_ = s.UpdateSessionDataSafe("k", "v1")

	err := s.UpdateSessionDataSafe("k", "v2")

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("k")
	assert.True(t, ok)
	assert.Equal(t, "v2", val)
}

func TestUpdateSessionDataSafe_ClaudeSessionID_ValidString(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.ClaudeSessionID", "sess-123")

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("nodeA.ClaudeSessionID")
	assert.True(t, ok)
	assert.Equal(t, "sess-123", val)
}

func TestUpdateSessionDataSafe_ClaudeSessionID_EmptyString(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.ClaudeSessionID", "")

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("nodeA.ClaudeSessionID")
	assert.True(t, ok)
	assert.Equal(t, "", val)
}

func TestUpdateSessionDataSafe_UpdatesUpdatedAt(t *testing.T) {
	s := newTestSession(t)
	initialUpdatedAt := s.GetMetadataSnapshotSafe().UpdatedAt

	err := s.UpdateSessionDataSafe("k", "v")

	require.NoError(t, err)
	assert.GreaterOrEqual(t, s.GetMetadataSnapshotSafe().UpdatedAt, initialUpdatedAt)
}

// --- Validation Failures — key ---

func TestUpdateSessionDataSafe_EmptyKey(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("", "x")

	require.Error(t, err)
	assert.Equal(t, "session data key cannot be empty", err.Error())
}

// --- Validation Failures — ClaudeSessionID type ---

func TestUpdateSessionDataSafe_ClaudeSessionID_IntValue(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.ClaudeSessionID", 42)

	require.Error(t, err)
	assert.Equal(t, "ClaudeSessionID value must be a string, got int", err.Error())
}

func TestUpdateSessionDataSafe_ClaudeSessionID_NilValue(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.ClaudeSessionID", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClaudeSessionID value must be a string")
}

func TestUpdateSessionDataSafe_ClaudeSessionID_Stringer(t *testing.T) {
	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.ClaudeSessionID", bytes.NewBufferString("x"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClaudeSessionID value must be a string")
}

// --- Happy Path — GetSessionDataSafe ---

func TestGetSessionDataSafe_ExistingKey(t *testing.T) {
	s := newTestSession(t)
	_ = s.UpdateSessionDataSafe("k", "v")

	val, ok := s.GetSessionDataSafe("k")

	assert.True(t, ok)
	assert.Equal(t, "v", val)
}

func TestGetSessionDataSafe_MissingKey(t *testing.T) {
	s := newTestSession(t)

	val, ok := s.GetSessionDataSafe("nonexistent")

	assert.False(t, ok)
	assert.Nil(t, val)
}

// --- Null / Empty Input ---

func TestGetSessionDataSafe_EmptyKey(t *testing.T) {
	s := newTestSession(t)

	val, ok := s.GetSessionDataSafe("")

	assert.False(t, ok)
	assert.Nil(t, val)
}
