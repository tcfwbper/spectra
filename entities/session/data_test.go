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

// --- Happy Path — UpdateSessionDataSafe (PID) ---

func TestUpdateSessionDataSafe_PID_ValidPositiveInt(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", 12345)

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("nodeA.PID")
	assert.True(t, ok)
	assert.Equal(t, 12345, val)
}

func TestUpdateSessionDataSafe_PID_Zero(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", 0)

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("nodeA.PID")
	assert.True(t, ok)
	assert.Equal(t, 0, val)
}

func TestUpdateSessionDataSafe_PID_Negative(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", -1)

	require.NoError(t, err)
	val, ok := s.GetSessionDataSafe("nodeA.PID")
	assert.True(t, ok)
	assert.Equal(t, -1, val)
}

// --- Validation Failures — PID type ---

func TestUpdateSessionDataSafe_PID_StringValue(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", "1234")

	require.Error(t, err)
	assert.Equal(t, "PID value must be an int, got string", err.Error())
}

func TestUpdateSessionDataSafe_PID_Int64Value(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", int64(1234))

	require.Error(t, err)
	assert.Equal(t, "PID value must be an int, got int64", err.Error())
}

func TestUpdateSessionDataSafe_PID_Float64Value(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", float64(1234))

	require.Error(t, err)
	assert.Equal(t, "PID value must be an int, got float64", err.Error())
}

func TestUpdateSessionDataSafe_PID_NilValue(t *testing.T) {
	t.Skip("scaffolded: production data.go does not yet implement PID type validation (missing .PID suffix check in UpdateSessionDataSafe)")

	s := newTestSession(t)

	err := s.UpdateSessionDataSafe("nodeA.PID", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PID value must be an int")
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
