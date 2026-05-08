package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Construction ---

func TestNewRuntimeMessage_EventType(t *testing.T) {
	payload := json.RawMessage(`{"eventType":"ReviewNeeded"}`)
	msg, err := NewRuntimeMessage("event", payload, "sess-123")

	require.NoError(t, err)
	assert.Equal(t, "event", msg.Type())
	assert.JSONEq(t, `{"eventType":"ReviewNeeded"}`, string(msg.Payload()))
	assert.Equal(t, "sess-123", msg.ClaudeSessionID())
}

func TestNewRuntimeMessage_ErrorType(t *testing.T) {
	payload := json.RawMessage(`{"detail":"something failed"}`)
	msg, err := NewRuntimeMessage("error", payload, "sess-456")

	require.NoError(t, err)
	assert.Equal(t, "error", msg.Type())
	assert.JSONEq(t, `{"detail":"something failed"}`, string(msg.Payload()))
	assert.Equal(t, "sess-456", msg.ClaudeSessionID())
}

func TestNewRuntimeMessage_EmptyClaudeSessionID(t *testing.T) {
	payload := json.RawMessage(`{"key":"value"}`)
	msg, err := NewRuntimeMessage("event", payload, "")

	require.NoError(t, err)
	assert.Equal(t, "", msg.ClaudeSessionID())
}

func TestNewRuntimeMessage_EmptyPayloadObject(t *testing.T) {
	payload := json.RawMessage(`{}`)
	msg, err := NewRuntimeMessage("event", payload, "")

	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(msg.Payload()))
}

// --- Validation Failures — Type ---

func TestNewRuntimeMessage_EmptyType(t *testing.T) {
	_, err := NewRuntimeMessage("", json.RawMessage(`{}`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_UnrecognizedType(t *testing.T) {
	_, err := NewRuntimeMessage("unknown", json.RawMessage(`{}`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_WarningType(t *testing.T) {
	_, err := NewRuntimeMessage("warning", json.RawMessage(`{}`), "")

	require.Error(t, err)
}

// --- Validation Failures — Payload ---

func TestNewRuntimeMessage_NilPayload(t *testing.T) {
	_, err := NewRuntimeMessage("event", nil, "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsArray(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`[1,2,3]`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsEmptyArray(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`[]`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsString(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`"hello"`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsNumber(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`123`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsBoolean(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`true`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsNull(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`null`), "")

	require.Error(t, err)
}

func TestNewRuntimeMessage_PayloadIsInvalidJSON(t *testing.T) {
	_, err := NewRuntimeMessage("event", json.RawMessage(`{invalid`), "")

	require.Error(t, err)
}

// --- Immutability ---

func TestRuntimeMessage_PayloadImmutability(t *testing.T) {
	original := []byte(`{"key":"value"}`)
	payload := json.RawMessage(original)
	msg, err := NewRuntimeMessage("event", payload, "")
	require.NoError(t, err)

	// Mutate the original slice after construction
	original[2] = 'X'
	payload[3] = 'Y'

	assert.JSONEq(t, `{"key":"value"}`, string(msg.Payload()))
}

func TestRuntimeMessage_GetterReturnImmutability(t *testing.T) {
	payload := json.RawMessage(`{"key":"value"}`)
	msg, err := NewRuntimeMessage("event", payload, "")
	require.NoError(t, err)

	// Get the payload and mutate the returned slice
	returned := msg.Payload()
	returned[0] = 'X'

	// Second call should still return the original value
	assert.JSONEq(t, `{"key":"value"}`, string(msg.Payload()))
}
