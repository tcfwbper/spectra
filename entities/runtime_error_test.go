package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewRuntimeError_ValidInputs(t *testing.T) {
	detail := json.RawMessage(`{"err":"ECONNREFUSED"}`)
	re, err := NewRuntimeError("MessageRouter", "socket creation failed", detail, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Connecting")

	require.NoError(t, err)
	assert.Equal(t, "MessageRouter", re.Issuer())
	assert.Equal(t, "socket creation failed", re.Message())
	assert.JSONEq(t, `{"err":"ECONNREFUSED"}`, string(re.Detail()))
	assert.Equal(t, int64(1700000000), re.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", re.SessionID())
	assert.Equal(t, "Connecting", re.FailingState())
}

func TestNewRuntimeError_ArbitraryIssuerName(t *testing.T) {
	re, err := NewRuntimeError("UnknownComponent99", "err", nil, 1, "550e8400-e29b-41d4-a716-446655440000", "Init")

	require.NoError(t, err)
	assert.Equal(t, "UnknownComponent99", re.Issuer())
}

// --- Validation Failures — Issuer ---

func TestNewRuntimeError_EmptyIssuer(t *testing.T) {
	_, err := NewRuntimeError("", "msg", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewRuntimeError_WhitespaceOnlyIssuer(t *testing.T) {
	_, err := NewRuntimeError("   ", "msg", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

// --- Error Propagation ---

func TestNewRuntimeError_PropagatesMessageError(t *testing.T) {
	_, err := NewRuntimeError("Router", "", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewRuntimeError_PropagatesSessionIDError(t *testing.T) {
	_, err := NewRuntimeError("Router", "msg", json.RawMessage(`{}`), 1700000000, "bad", "Processing")

	require.Error(t, err)
}

func TestNewRuntimeError_PropagatesOccurredAtError(t *testing.T) {
	_, err := NewRuntimeError("Router", "msg", json.RawMessage(`{}`), -1, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewRuntimeError_PropagatesDetailError(t *testing.T) {
	_, err := NewRuntimeError("Router", "msg", json.RawMessage(`[1]`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewRuntimeError_PropagatesFailingStateError(t *testing.T) {
	_, err := NewRuntimeError("Router", "msg", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "")

	require.Error(t, err)
}

// --- Immutability ---

func TestRuntimeError_Immutability(t *testing.T) {
	detail := json.RawMessage(`{"err":"ECONNREFUSED"}`)
	re, err := NewRuntimeError("MessageRouter", "socket creation failed", detail, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Connecting")
	require.NoError(t, err)

	// Verify all getters return construction values consistently
	assert.Equal(t, "MessageRouter", re.Issuer())
	assert.Equal(t, "socket creation failed", re.Message())
	assert.JSONEq(t, `{"err":"ECONNREFUSED"}`, string(re.Detail()))
	assert.Equal(t, int64(1700000000), re.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", re.SessionID())
	assert.Equal(t, "Connecting", re.FailingState())

	// Call getters again to confirm stability
	assert.Equal(t, "MessageRouter", re.Issuer())
	assert.Equal(t, "socket creation failed", re.Message())
	assert.JSONEq(t, `{"err":"ECONNREFUSED"}`, string(re.Detail()))
	assert.Equal(t, int64(1700000000), re.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", re.SessionID())
	assert.Equal(t, "Connecting", re.FailingState())
}
