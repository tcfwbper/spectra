package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewAgentError_ValidInputs(t *testing.T) {
	detail := json.RawMessage(`{"reason":"timeout"}`)
	ae, err := NewAgentError("Reviewer", "agent failed", detail, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Review")

	require.NoError(t, err)
	assert.Equal(t, "Reviewer", ae.AgentRole())
	assert.Equal(t, "agent failed", ae.Message())
	assert.JSONEq(t, `{"reason":"timeout"}`, string(ae.Detail()))
	assert.Equal(t, int64(1700000000), ae.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ae.SessionID())
	assert.Equal(t, "Review", ae.FailingState())
}

func TestNewAgentError_EmptyAgentRole(t *testing.T) {
	ae, err := NewAgentError("", "human error", nil, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Approval")

	require.NoError(t, err)
	assert.Equal(t, "", ae.AgentRole())
	assert.Equal(t, "human error", ae.Message())
	assert.Nil(t, ae.Detail())
	assert.Equal(t, int64(1700000000), ae.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ae.SessionID())
	assert.Equal(t, "Approval", ae.FailingState())
}

func TestNewAgentError_SpecialCharsAgentRole(t *testing.T) {
	ae, err := NewAgentError("my agent/v2 (test)", "err", nil, 1, "550e8400-e29b-41d4-a716-446655440000", "Init")

	require.NoError(t, err)
	assert.Equal(t, "my agent/v2 (test)", ae.AgentRole())
}

// --- Error Propagation ---

func TestNewAgentError_PropagatesMessageError(t *testing.T) {
	_, err := NewAgentError("Agent", "", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewAgentError_PropagatesSessionIDError(t *testing.T) {
	_, err := NewAgentError("Agent", "msg", json.RawMessage(`{}`), 1700000000, "bad", "Processing")

	require.Error(t, err)
}

func TestNewAgentError_PropagatesOccurredAtError(t *testing.T) {
	_, err := NewAgentError("Agent", "msg", json.RawMessage(`{}`), 0, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewAgentError_PropagatesDetailError(t *testing.T) {
	_, err := NewAgentError("Agent", "msg", json.RawMessage(`[]`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Processing")

	require.Error(t, err)
}

func TestNewAgentError_PropagatesFailingStateError(t *testing.T) {
	_, err := NewAgentError("Agent", "msg", json.RawMessage(`{}`), 1700000000, "550e8400-e29b-41d4-a716-446655440000", "")

	require.Error(t, err)
}

// --- Immutability ---

func TestAgentError_Immutability(t *testing.T) {
	detail := json.RawMessage(`{"reason":"timeout"}`)
	ae, err := NewAgentError("Reviewer", "agent failed", detail, 1700000000, "550e8400-e29b-41d4-a716-446655440000", "Review")
	require.NoError(t, err)

	// Verify all getters return construction values consistently
	assert.Equal(t, "Reviewer", ae.AgentRole())
	assert.Equal(t, "agent failed", ae.Message())
	assert.JSONEq(t, `{"reason":"timeout"}`, string(ae.Detail()))
	assert.Equal(t, int64(1700000000), ae.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ae.SessionID())
	assert.Equal(t, "Review", ae.FailingState())

	// Call getters again to confirm stability
	assert.Equal(t, "Reviewer", ae.AgentRole())
	assert.Equal(t, "agent failed", ae.Message())
	assert.JSONEq(t, `{"reason":"timeout"}`, string(ae.Detail()))
	assert.Equal(t, int64(1700000000), ae.OccurredAt())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ae.SessionID())
	assert.Equal(t, "Review", ae.FailingState())
}
