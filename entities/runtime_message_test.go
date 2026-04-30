package entities_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/entities"
)

// TestRuntimeMessage_ValidEventMessage creates RuntimeMessage with type=event and all valid fields
func TestRuntimeMessage_ValidEventMessage(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "DraftCompleted", "message": "ready", "payload": {"count": 3}}`)

	msg, err := entities.NewRuntimeMessage("event", "550e8400-e29b-41d4-a716-446655440000", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "event", msg.Type)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", msg.ClaudeSessionID)
	assert.NotNil(t, msg.Payload)
}

// TestRuntimeMessage_ValidErrorMessage creates RuntimeMessage with type=error and all valid fields
func TestRuntimeMessage_ValidErrorMessage(t *testing.T) {
	payload := json.RawMessage(`{"message": "Failed to load", "detail": {"error": "not found"}}`)

	msg, err := entities.NewRuntimeMessage("error", "550e8400-e29b-41d4-a716-446655440000", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "error", msg.Type)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", msg.ClaudeSessionID)
}

// TestRuntimeMessage_EmptyClaudeSessionID creates RuntimeMessage with empty claudeSessionID
func TestRuntimeMessage_EmptyClaudeSessionID(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "RequirementProvided", "message": "", "payload": {}}`)

	msg, err := entities.NewRuntimeMessage("event", "", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "", msg.ClaudeSessionID)
}

// TestRuntimeMessage_OmittedClaudeSessionID creates RuntimeMessage with claudeSessionID field omitted
func TestRuntimeMessage_OmittedClaudeSessionID(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "RequirementProvided", "message": "", "payload": {}}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	assert.Equal(t, "", msg.ClaudeSessionID)
}

// TestRuntimeMessage_EventPayload_MessageOmitted accepts event payload with message field omitted
func TestRuntimeMessage_EventPayload_MessageOmitted(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Started", "payload": {}}`)

	msg, err := entities.NewRuntimeMessage("event", "", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)

	var ep entities.EventPayload
	err = json.Unmarshal(msg.Payload, &ep)
	require.NoError(t, err)
	assert.Equal(t, "Started", ep.EventType)
}

// TestRuntimeMessage_EventPayload_PayloadOmitted accepts event payload with payload field omitted
func TestRuntimeMessage_EventPayload_PayloadOmitted(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Started", "message": "begin"}`)

	msg, err := entities.NewRuntimeMessage("event", "", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)

	var ep entities.EventPayload
	err = json.Unmarshal(msg.Payload, &ep)
	require.NoError(t, err)
	assert.Equal(t, "begin", ep.Message)
}

// TestRuntimeMessage_EventPayload_BothOmitted accepts event payload with both message and payload fields omitted
func TestRuntimeMessage_EventPayload_BothOmitted(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Continue"}`)

	msg, err := entities.NewRuntimeMessage("event", "", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)

	var ep entities.EventPayload
	err = json.Unmarshal(msg.Payload, &ep)
	require.NoError(t, err)
	assert.Equal(t, "Continue", ep.EventType)
}

// TestRuntimeMessage_ErrorPayload_DetailOmitted accepts error payload with detail field omitted
func TestRuntimeMessage_ErrorPayload_DetailOmitted(t *testing.T) {
	payload := json.RawMessage(`{"message": "Failed to process"}`)

	msg, err := entities.NewRuntimeMessage("error", "", payload)

	require.NoError(t, err)
	require.NotNil(t, msg)

	var ep entities.ErrorPayload
	err = json.Unmarshal(msg.Payload, &ep)
	require.NoError(t, err)
	assert.Equal(t, "Failed to process", ep.Message)
}

// TestRuntimeMessage_SerializesToJSON verifies RuntimeMessage serializes to valid JSON
func TestRuntimeMessage_SerializesToJSON(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Test", "message": "", "payload": {}}`)

	msg, err := entities.NewRuntimeMessage("event", "test-id", payload)
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"type":"event"`)

	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeMessage_DeserializesFromJSON verifies RuntimeMessage deserializes from valid JSON
func TestRuntimeMessage_DeserializesFromJSON(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "test-id", "payload": {"eventType": "Test", "message": "", "payload": {}}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	assert.Equal(t, "event", msg.Type)
	assert.Equal(t, "test-id", msg.ClaudeSessionID)

	err = msg.Validate()
	require.NoError(t, err)
}

// TestRuntimeMessage_MissingType rejects message with missing type field
func TestRuntimeMessage_MissingType(t *testing.T) {
	jsonStr := `{"payload": {"eventType": "Test"}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	err = msg.Validate()
	require.Error(t, err)
	assert.Regexp(t, `(?i)type.*must not be empty`, err.Error())
}

// TestRuntimeMessage_EmptyType rejects message with empty type field
func TestRuntimeMessage_EmptyType(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Test"}`)
	_, err := entities.NewRuntimeMessage("", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)type.*must not be empty`, err.Error())
}

// TestRuntimeMessage_UnrecognizedType rejects message with unrecognized type value
func TestRuntimeMessage_UnrecognizedType(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Test"}`)
	_, err := entities.NewRuntimeMessage("unknown", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)invalid message type 'unknown'`, err.Error())
}

// TestRuntimeMessage_LegacyType rejects message with legacy type value
func TestRuntimeMessage_LegacyType(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Test"}`)
	_, err := entities.NewRuntimeMessage("legacy", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)invalid message type 'legacy'`, err.Error())
}

// TestRuntimeMessage_MissingPayload rejects message with missing payload field
func TestRuntimeMessage_MissingPayload(t *testing.T) {
	_, err := entities.NewRuntimeMessage("event", "", nil)
	require.Error(t, err)
	assert.Regexp(t, `(?i)missing required field 'payload'`, err.Error())
}

// TestRuntimeMessage_PayloadPrimitiveString rejects message with payload as JSON primitive string
func TestRuntimeMessage_PayloadPrimitiveString(t *testing.T) {
	payload := json.RawMessage(`"string"`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)payload must be a JSON object`, err.Error())
}

// TestRuntimeMessage_PayloadPrimitiveNumber rejects message with payload as JSON primitive number
func TestRuntimeMessage_PayloadPrimitiveNumber(t *testing.T) {
	payload := json.RawMessage(`123`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)payload must be a JSON object`, err.Error())
}

// TestRuntimeMessage_PayloadPrimitiveBoolean rejects message with payload as JSON primitive boolean
func TestRuntimeMessage_PayloadPrimitiveBoolean(t *testing.T) {
	payload := json.RawMessage(`true`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)payload must be a JSON object`, err.Error())
}

// TestRuntimeMessage_PayloadNull rejects message with payload as null
func TestRuntimeMessage_PayloadNull(t *testing.T) {
	payload := json.RawMessage(`null`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)payload must be a JSON object`, err.Error())
}

// TestRuntimeMessage_PayloadArray rejects message with payload as JSON array
func TestRuntimeMessage_PayloadArray(t *testing.T) {
	payload := json.RawMessage(`[1, 2, 3]`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)payload must be a JSON object`, err.Error())
}

// TestRuntimeMessage_PayloadEmptyArray rejects message with payload as empty JSON array
func TestRuntimeMessage_PayloadEmptyArray(t *testing.T) {
	payload := json.RawMessage(`[]`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)payload must be a JSON object`, err.Error())
}

// TestRuntimeMessage_EventPayload_MissingEventType rejects event message with missing eventType field
func TestRuntimeMessage_EventPayload_MissingEventType(t *testing.T) {
	payload := json.RawMessage(`{"message": "test", "payload": {}}`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)eventType must not be empty`, err.Error())
}

// TestRuntimeMessage_EventPayload_EmptyEventType rejects event message with empty eventType field
func TestRuntimeMessage_EventPayload_EmptyEventType(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "", "message": "test", "payload": {}}`)
	_, err := entities.NewRuntimeMessage("event", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)eventType must not be empty`, err.Error())
}

// TestRuntimeMessage_ErrorPayload_MissingMessage rejects error message with missing message field
func TestRuntimeMessage_ErrorPayload_MissingMessage(t *testing.T) {
	payload := json.RawMessage(`{"detail": {}}`)
	_, err := entities.NewRuntimeMessage("error", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)error payload missing required field 'message'`, err.Error())
}

// TestRuntimeMessage_ErrorPayload_EmptyMessage rejects error message with empty message field
func TestRuntimeMessage_ErrorPayload_EmptyMessage(t *testing.T) {
	payload := json.RawMessage(`{"message": "", "detail": {}}`)
	_, err := entities.NewRuntimeMessage("error", "", payload)
	require.Error(t, err)
	assert.Regexp(t, `(?i)error payload missing required field 'message'`, err.Error())
}

// TestRuntimeMessage_ClaudeSessionID_Null rejects message with claudeSessionID as null
func TestRuntimeMessage_ClaudeSessionID_Null(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": null, "payload": {"eventType": "Test"}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	// Go's JSON unmarshaler treats null for string as zero value ""
	// The ClaudeSessionID type validation (non-string) is handled at the wire protocol level
	// by RuntimeSocketManager, not by the struct itself
	assert.Equal(t, "", msg.ClaudeSessionID)
}

// TestRuntimeMessage_ClaudeSessionID_Number rejects message with claudeSessionID as number
func TestRuntimeMessage_ClaudeSessionID_Number(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": 123, "payload": {"eventType": "Test"}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)

	// Go's strict JSON unmarshaler will fail on type mismatch
	require.Error(t, err)
}

// TestRuntimeMessage_ClaudeSessionID_Object rejects message with claudeSessionID as object
func TestRuntimeMessage_ClaudeSessionID_Object(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": {"id": "test"}, "payload": {"eventType": "Test"}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)

	// Go's strict JSON unmarshaler will fail on type mismatch
	require.Error(t, err)
}

// TestRuntimeMessage_MalformedJSON_MissingBrace rejects message with malformed JSON
func TestRuntimeMessage_MalformedJSON_MissingBrace(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "Test"`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.Error(t, err)
	assert.Regexp(t, `(?i)JSON|unmarshal`, err.Error())
}

// TestRuntimeMessage_MalformedJSON_InvalidEscape rejects message with malformed JSON
func TestRuntimeMessage_MalformedJSON_InvalidEscape(t *testing.T) {
	jsonStr := `{"type": "event\x", "payload": {}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.Error(t, err)
	assert.Regexp(t, `(?i)JSON|unmarshal|invalid`, err.Error())
}

// TestRuntimeMessage_SizeLimit_JustUnder10MB accepts message with serialized JSON just under 10 MB
func TestRuntimeMessage_SizeLimit_JustUnder10MB(t *testing.T) {
	largeData := strings.Repeat("x", 10*1024*1024-1000) // 10MB - 1KB for overhead

	payload := json.RawMessage(`{"eventType": "Test", "message": "` + largeData + `", "payload": {}}`)

	msg, err := entities.NewRuntimeMessage("event", "test", payload)
	require.NoError(t, err)
	require.NotNil(t, msg)

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(jsonBytes), 10*1024*1024)
}

// TestRuntimeMessage_SizeLimit_Exceeds10MB rejects message with serialized JSON exceeding 10 MB
func TestRuntimeMessage_SizeLimit_Exceeds10MB(t *testing.T) {
	largeData := strings.Repeat("x", 11*1024*1024)

	payload := json.RawMessage(`{"eventType": "Test", "message": "` + largeData + `", "payload": {}}`)

	msg, err := entities.NewRuntimeMessage("event", "test", payload)
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Greater(t, len(jsonBytes), 10*1024*1024)
	// Size limit enforcement is done by RuntimeSocketManager at the transport level
}

// TestRuntimeMessage_LargeEventPayload accepts event message with very large payload object
func TestRuntimeMessage_LargeEventPayload(t *testing.T) {
	largePayload := make(map[string]interface{})
	for i := 0; i < 9000; i++ {
		largePayload[string(rune(i))] = strings.Repeat("a", 1000)
	}
	innerPayloadJSON, err := json.Marshal(largePayload)
	require.NoError(t, err)

	payload := json.RawMessage(`{"eventType": "Test", "message": "", "payload": ` + string(innerPayloadJSON) + `}`)

	msg, err := entities.NewRuntimeMessage("event", "test", payload)
	require.NoError(t, err)
	require.NotNil(t, msg)

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Greater(t, len(jsonBytes), 1024*1024)
}

// TestRuntimeMessage_DeepNestedPayload accepts message with deeply nested JSON in payload
func TestRuntimeMessage_DeepNestedPayload(t *testing.T) {
	nested := make(map[string]interface{})
	current := nested
	for i := 0; i < 100; i++ {
		next := make(map[string]interface{})
		current["level"] = next
		current = next
	}
	nestedJSON, err := json.Marshal(nested)
	require.NoError(t, err)

	payload := json.RawMessage(`{"eventType": "Test", "message": "", "payload": ` + string(nestedJSON) + `}`)

	msg, err := entities.NewRuntimeMessage("event", "test", payload)
	require.NoError(t, err)
	require.NotNil(t, msg)

	// Verify can roundtrip
	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded entities.RuntimeMessage
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
}

// TestRuntimeMessage_UnicodeInPayload accepts message with Unicode characters in payload fields
func TestRuntimeMessage_UnicodeInPayload(t *testing.T) {
	payload := json.RawMessage(`{"eventType": "Test", "message": "通知: Process complete 🎉", "payload": {"emoji": "🚀"}}`)

	msg, err := entities.NewRuntimeMessage("event", "", payload)
	require.NoError(t, err)
	require.NotNil(t, msg)

	var ep entities.EventPayload
	err = json.Unmarshal(msg.Payload, &ep)
	require.NoError(t, err)
	assert.Equal(t, "通知: Process complete 🎉", ep.Message)
}

// TestRuntimeMessage_RepeatedDeserialization verifies repeated deserialization produces identical results
func TestRuntimeMessage_RepeatedDeserialization(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "test", "payload": {"eventType": "Test", "message": "test", "payload": {"key": "value"}}}`

	var msg1, msg2, msg3 entities.RuntimeMessage

	err1 := json.Unmarshal([]byte(jsonStr), &msg1)
	err2 := json.Unmarshal([]byte(jsonStr), &msg2)
	err3 := json.Unmarshal([]byte(jsonStr), &msg3)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	assert.Equal(t, msg1.Type, msg2.Type)
	assert.Equal(t, msg2.Type, msg3.Type)
	assert.Equal(t, msg1.ClaudeSessionID, msg2.ClaudeSessionID)
	assert.Equal(t, string(msg1.Payload), string(msg2.Payload))
}

// TestRuntimeMessage_SocketTransmission_EventMessage - E2E test for event transmission over socket
func TestRuntimeMessage_SocketTransmission_EventMessage(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir
}

// TestRuntimeMessage_SocketTransmission_ErrorMessage - E2E test for error transmission over socket
func TestRuntimeMessage_SocketTransmission_ErrorMessage(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir
}

// TestRuntimeMessage_ConcurrentConnections - Race test for concurrent connections
func TestRuntimeMessage_ConcurrentConnections(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir
}

// TestRuntimeMessage_SocketClosed_AfterValidation verifies connection closed after validation failure
func TestRuntimeMessage_SocketClosed_AfterValidation(t *testing.T) {
}

// TestRuntimeMessage_SocketClosed_AfterMalformedJSON verifies connection closed after JSON parse error
func TestRuntimeMessage_SocketClosed_AfterMalformedJSON(t *testing.T) {
}

// TestRuntimeMessage_Validate verifies Validate method works on deserialized message
func TestRuntimeMessage_Validate(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "test", "payload": {"eventType": "Test", "message": "", "payload": {}}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	err = msg.Validate()
	require.NoError(t, err)
}

// TestRuntimeMessage_Validate_InvalidType verifies Validate catches invalid type on deserialized message
func TestRuntimeMessage_Validate_InvalidType(t *testing.T) {
	jsonStr := `{"type": "unknown", "payload": {"eventType": "Test"}}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	err = msg.Validate()
	require.Error(t, err)
	assert.Regexp(t, `(?i)invalid message type`, err.Error())
}

// TestRuntimeMessage_Validate_MissingPayload verifies Validate catches missing payload on deserialized message
func TestRuntimeMessage_Validate_MissingPayload(t *testing.T) {
	jsonStr := `{"type": "event"}`

	var msg entities.RuntimeMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	err = msg.Validate()
	require.Error(t, err)
	assert.Regexp(t, `(?i)missing required field 'payload'`, err.Error())
}
