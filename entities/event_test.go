package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewEvent_ValidInputs(t *testing.T) {
	payload := json.RawMessage(`{"score":95}`)
	ev, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"ReviewCompleted",
		"done",
		payload,
		"ReviewNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ev.ID())
	assert.Equal(t, "ReviewCompleted", ev.Type())
	assert.Equal(t, "done", ev.Message())
	assert.JSONEq(t, `{"score":95}`, string(ev.Payload()))
	assert.Equal(t, "ReviewNode", ev.EmittedBy())
	assert.Equal(t, int64(1700000000), ev.EmittedAt())
	assert.Equal(t, "660e8400-e29b-41d4-a716-446655440000", ev.SessionID())
}

func TestNewEvent_EmptyMessage(t *testing.T) {
	payload := json.RawMessage(`{"key":"value"}`)
	ev, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"",
		payload,
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.NoError(t, err)
	assert.Equal(t, "", ev.Message())
}

func TestNewEvent_EmptyObjectPayload(t *testing.T) {
	payload := json.RawMessage(`{}`)
	ev, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		payload,
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(ev.Payload()))
}

// --- Validation Failures — ID ---

func TestNewEvent_InvalidID(t *testing.T) {
	_, err := NewEvent(
		"not-a-uuid",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_EmptyID(t *testing.T) {
	_, err := NewEvent(
		"",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

// --- Validation Failures — Type ---

func TestNewEvent_EmptyType(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_TypeStartsLowercase(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"reviewNeeded",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_TypeContainsHyphen(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"Review-Needed",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_TypeContainsUnderscore(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"Review_Needed",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_TypeContainsSpace(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"Review Needed",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

// --- Validation Failures — Payload ---

func TestNewEvent_NilPayload(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		nil,
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_PayloadIsArray(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`[1,2,3]`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_PayloadIsPrimitiveString(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`"hello"`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_PayloadIsPrimitiveNumber(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`42`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_PayloadIsPrimitiveBoolean(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`true`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_PayloadIsPrimitiveNull(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`null`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_PayloadIsInvalidJSON(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`{broken`),
		"StartNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

// --- Validation Failures — EmittedBy ---

func TestNewEvent_EmptyEmittedBy(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

// --- Validation Failures — EmittedAt ---

func TestNewEvent_EmittedAtZero(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		0,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

func TestNewEvent_EmittedAtNegative(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		-1,
		"660e8400-e29b-41d4-a716-446655440000",
	)

	require.Error(t, err)
}

// --- Validation Failures — SessionID ---

func TestNewEvent_InvalidSessionID(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"invalid",
	)

	require.Error(t, err)
}

func TestNewEvent_EmptySessionID(t *testing.T) {
	_, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskStarted",
		"msg",
		json.RawMessage(`{}`),
		"StartNode",
		1700000000,
		"",
	)

	require.Error(t, err)
}

// --- Immutability ---

func TestEvent_Immutability(t *testing.T) {
	payload := json.RawMessage(`{"score":95}`)
	ev, err := NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"ReviewCompleted",
		"done",
		payload,
		"ReviewNode",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)
	require.NoError(t, err)

	// Verify all getters return construction values consistently
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ev.ID())
	assert.Equal(t, "ReviewCompleted", ev.Type())
	assert.Equal(t, "done", ev.Message())
	assert.JSONEq(t, `{"score":95}`, string(ev.Payload()))
	assert.Equal(t, "ReviewNode", ev.EmittedBy())
	assert.Equal(t, int64(1700000000), ev.EmittedAt())
	assert.Equal(t, "660e8400-e29b-41d4-a716-446655440000", ev.SessionID())

	// Call getters again to confirm stability
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ev.ID())
	assert.Equal(t, "ReviewCompleted", ev.Type())
	assert.Equal(t, "done", ev.Message())
	assert.JSONEq(t, `{"score":95}`, string(ev.Payload()))
	assert.Equal(t, "ReviewNode", ev.EmittedBy())
	assert.Equal(t, int64(1700000000), ev.EmittedAt())
	assert.Equal(t, "660e8400-e29b-41d4-a716-446655440000", ev.SessionID())
}
