package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewSessionError_ValidInputs(t *testing.T) {
	detail := json.RawMessage(`{"code":500}`)
	se, err := NewSessionError("connection lost", detail, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.NoError(t, err)
	assert.Equal(t, "connection lost", se.Message())
	assert.JSONEq(t, `{"code":500}`, string(se.Detail()))
	assert.Equal(t, int64(1700000000), se.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", se.SessionID())
	assert.Equal(t, "Processing", se.FailingState())
}

func TestNewSessionError_NilDetail(t *testing.T) {
	se, err := NewSessionError("timeout", nil, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Waiting")

	require.NoError(t, err)
	assert.Equal(t, "timeout", se.Message())
	assert.Nil(t, se.Detail())
	assert.Equal(t, int64(1700000000), se.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", se.SessionID())
	assert.Equal(t, "Waiting", se.FailingState())
}

func TestNewSessionError_EmptyObjectDetail(t *testing.T) {
	detail := json.RawMessage(`{}`)
	se, err := NewSessionError("err", detail, 1, "550e8400-e29b-41d4-a716-446655440000", "Init")

	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(se.Detail()))
}

// --- Validation Failures — Message ---

func TestNewSessionError_EmptyMessage(t *testing.T) {
	_, err := NewSessionError("", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_WhitespaceOnlyMessage(t *testing.T) {
	_, err := NewSessionError("   ", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

// --- Validation Failures — Detail ---

func TestNewSessionError_DetailIsArray(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`[1,2,3]`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_DetailIsPrimitive(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`"hello"`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_DetailIsNumber(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`123`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_DetailIsBooleanTrue(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`true`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_DetailIsInvalidJSON(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`{broken`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

// --- Validation Failures — OccurredAt ---

func TestNewSessionError_OccurredAtZero(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`{}`), 0, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_OccurredAtNegative(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`{}`), -1, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

// --- Validation Failures — SessionID ---

func TestNewSessionError_InvalidSessionID(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`{}`), 1700000000, "not-a-uuid", "Processing")

	require.Error(t, err)
}

func TestNewSessionError_EmptySessionID(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`{}`), 1700000000, "", "Processing")

	require.Error(t, err)
}

// --- Validation Failures — FailingState ---

func TestNewSessionError_EmptyFailingState(t *testing.T) {
	_, err := NewSessionError("err", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "")

	require.Error(t, err)
}

// --- Immutability ---

func TestSessionError_Immutability(t *testing.T) {
	detail := json.RawMessage(`{"code":500}`)
	se, err := NewSessionError("connection lost", detail, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")
	require.NoError(t, err)

	// Verify all getters return construction values consistently
	assert.Equal(t, "connection lost", se.Message())
	assert.JSONEq(t, `{"code":500}`, string(se.Detail()))
	assert.Equal(t, int64(1700000000), se.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", se.SessionID())
	assert.Equal(t, "Processing", se.FailingState())

	// Call getters again to confirm stability
	assert.Equal(t, "connection lost", se.Message())
	assert.JSONEq(t, `{"code":500}`, string(se.Detail()))
	assert.Equal(t, int64(1700000000), se.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", se.SessionID())
	assert.Equal(t, "Processing", se.FailingState())
}
