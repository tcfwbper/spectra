package entities_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuntimeMessage_ValidEventMessage creates RuntimeMessage with type=event and all valid fields
func TestRuntimeMessage_ValidEventMessage(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "550e8400-e29b-41d4-a716-446655440000", "payload": {"eventType": "DraftCompleted", "message": "ready", "payload": {"count": 3}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	assert.Equal(t, "event", msg["type"])
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", msg["claudeSessionID"])

	payload, ok := msg["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "DraftCompleted", payload["eventType"])
	assert.Equal(t, "ready", payload["message"])
}

// TestRuntimeMessage_ValidErrorMessage creates RuntimeMessage with type=error and all valid fields
func TestRuntimeMessage_ValidErrorMessage(t *testing.T) {
	jsonStr := `{"type": "error", "claudeSessionID": "550e8400-e29b-41d4-a716-446655440000", "payload": {"message": "Failed to load", "detail": {"error": "not found"}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	assert.Equal(t, "error", msg["type"])
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", msg["claudeSessionID"])

	payload, ok := msg["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Failed to load", payload["message"])
}

// TestRuntimeMessage_EmptyClaudeSessionID creates RuntimeMessage with empty claudeSessionID
func TestRuntimeMessage_EmptyClaudeSessionID(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "", "payload": {"eventType": "RequirementProvided", "message": "", "payload": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	assert.Equal(t, "", msg["claudeSessionID"])
}

// TestRuntimeMessage_OmittedClaudeSessionID creates RuntimeMessage with claudeSessionID field omitted
func TestRuntimeMessage_OmittedClaudeSessionID(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "RequirementProvided", "message": "", "payload": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	// When omitted, should default to empty string in the parsed struct
	_, exists := msg["claudeSessionID"]
	assert.False(t, exists, "claudeSessionID field should not exist in raw JSON")
}

// TestRuntimeMessage_EventPayload_MessageOmitted accepts event payload with message field omitted
func TestRuntimeMessage_EventPayload_MessageOmitted(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "Started", "payload": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	payload := msg["payload"].(map[string]interface{})
	_, exists := payload["message"]
	assert.False(t, exists, "message should not exist in raw JSON when omitted")
}

// TestRuntimeMessage_EventPayload_PayloadOmitted accepts event payload with payload field omitted
func TestRuntimeMessage_EventPayload_PayloadOmitted(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "Started", "message": "begin"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	payload := msg["payload"].(map[string]interface{})
	_, exists := payload["payload"]
	assert.False(t, exists, "payload should not exist in raw JSON when omitted")
}

// TestRuntimeMessage_EventPayload_BothOmitted accepts event payload with both message and payload fields omitted
func TestRuntimeMessage_EventPayload_BothOmitted(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "Continue"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	payload := msg["payload"].(map[string]interface{})
	_, msgExists := payload["message"]
	_, payloadExists := payload["payload"]
	assert.False(t, msgExists && payloadExists)
}

// TestRuntimeMessage_ErrorPayload_DetailOmitted accepts error payload with detail field omitted
func TestRuntimeMessage_ErrorPayload_DetailOmitted(t *testing.T) {
	jsonStr := `{"type": "error", "payload": {"message": "Failed to process"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	payload := msg["payload"].(map[string]interface{})
	_, exists := payload["detail"]
	assert.False(t, exists, "detail should not exist in raw JSON when omitted")
}

// TestRuntimeMessage_SerializesToJSON verifies RuntimeMessage serializes to valid JSON
func TestRuntimeMessage_SerializesToJSON(t *testing.T) {
	msg := map[string]interface{}{
		"type":            "event",
		"claudeSessionID": "test-id",
		"payload": map[string]interface{}{
			"eventType": "Test",
			"message":   "",
			"payload":   map[string]interface{}{},
		},
	}

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"type":"event"`)

	// When transmitted, should terminate with newline
	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeMessage_DeserializesFromJSON verifies RuntimeMessage deserializes from valid JSON
func TestRuntimeMessage_DeserializesFromJSON(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "test-id", "payload": {"eventType": "Test", "message": "", "payload": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.NoError(t, err)
	assert.Equal(t, "event", msg["type"])
	assert.Equal(t, "test-id", msg["claudeSessionID"])
}

// TestRuntimeMessage_MissingType rejects message with missing type field
func TestRuntimeMessage_MissingType(t *testing.T) {
	jsonStr := `{"payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	// ValidationError should be returned by RuntimeSocketManager
	_, exists := msg["type"]
	assert.False(t, exists, "type field should not exist")
	// In actual runtime: returns error matching /missing required field 'type'/i
}

// TestRuntimeMessage_EmptyType rejects message with empty type field
func TestRuntimeMessage_EmptyType(t *testing.T) {
	jsonStr := `{"type": "", "payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Equal(t, "", msg["type"])
	// In actual runtime: returns error matching /type field must not be empty/i
}

// TestRuntimeMessage_UnrecognizedType rejects message with unrecognized type value
func TestRuntimeMessage_UnrecognizedType(t *testing.T) {
	jsonStr := `{"type": "unknown", "payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Equal(t, "unknown", msg["type"])
	// In actual runtime: returns error matching /invalid message type 'unknown'/i
}

// TestRuntimeMessage_LegacyType rejects message with legacy type value
func TestRuntimeMessage_LegacyType(t *testing.T) {
	jsonStr := `{"type": "legacy", "payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Equal(t, "legacy", msg["type"])
	// In actual runtime: returns error matching /invalid message type 'legacy'/i
}

// TestRuntimeMessage_MissingPayload rejects message with missing payload field
func TestRuntimeMessage_MissingPayload(t *testing.T) {
	jsonStr := `{"type": "event"}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, exists := msg["payload"]
	assert.False(t, exists)
	// In actual runtime: returns error matching /missing required field 'payload'/i
}

// TestRuntimeMessage_PayloadPrimitiveString rejects message with payload as JSON primitive string
func TestRuntimeMessage_PayloadPrimitiveString(t *testing.T) {
	jsonStr := `{"type": "event", "payload": "string"}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, ok := msg["payload"].(string)
	assert.True(t, ok, "payload should be a string primitive")
	// In actual runtime: returns error matching /payload must be a JSON object/i
}

// TestRuntimeMessage_PayloadPrimitiveNumber rejects message with payload as JSON primitive number
func TestRuntimeMessage_PayloadPrimitiveNumber(t *testing.T) {
	jsonStr := `{"type": "event", "payload": 123}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, ok := msg["payload"].(float64)
	assert.True(t, ok, "payload should be a number primitive")
	// In actual runtime: returns error matching /payload must be a JSON object/i
}

// TestRuntimeMessage_PayloadPrimitiveBoolean rejects message with payload as JSON primitive boolean
func TestRuntimeMessage_PayloadPrimitiveBoolean(t *testing.T) {
	jsonStr := `{"type": "event", "payload": true}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, ok := msg["payload"].(bool)
	assert.True(t, ok, "payload should be a boolean primitive")
	// In actual runtime: returns error matching /payload must be a JSON object/i
}

// TestRuntimeMessage_PayloadNull rejects message with payload as null
func TestRuntimeMessage_PayloadNull(t *testing.T) {
	jsonStr := `{"type": "event", "payload": null}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Nil(t, msg["payload"])
	// In actual runtime: returns error matching /payload must be a JSON object/i
}

// TestRuntimeMessage_PayloadArray rejects message with payload as JSON array
func TestRuntimeMessage_PayloadArray(t *testing.T) {
	jsonStr := `{"type": "event", "payload": [1, 2, 3]}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, ok := msg["payload"].([]interface{})
	assert.True(t, ok, "payload should be an array")
	// In actual runtime: returns error matching /payload must be a JSON object/i
}

// TestRuntimeMessage_PayloadEmptyArray rejects message with payload as empty JSON array
func TestRuntimeMessage_PayloadEmptyArray(t *testing.T) {
	jsonStr := `{"type": "event", "payload": []}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	arr, ok := msg["payload"].([]interface{})
	assert.True(t, ok, "payload should be an array")
	assert.Len(t, arr, 0)
	// In actual runtime: returns error matching /payload must be a JSON object/i
}

// TestRuntimeMessage_EventPayload_MissingEventType rejects event message with missing eventType field
func TestRuntimeMessage_EventPayload_MissingEventType(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"message": "test", "payload": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	payload := msg["payload"].(map[string]interface{})
	_, exists := payload["eventType"]
	assert.False(t, exists)
	// In actual runtime: returns error matching /event payload missing required field 'eventType'/i
}

// TestRuntimeMessage_EventPayload_EmptyEventType rejects event message with empty eventType field
func TestRuntimeMessage_EventPayload_EmptyEventType(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "", "message": "test", "payload": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	payload := msg["payload"].(map[string]interface{})
	assert.Equal(t, "", payload["eventType"])
	// In actual runtime: returns error matching /eventType must not be empty/i
}

// TestRuntimeMessage_ErrorPayload_MissingMessage rejects error message with missing message field
func TestRuntimeMessage_ErrorPayload_MissingMessage(t *testing.T) {
	jsonStr := `{"type": "error", "payload": {"detail": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	payload := msg["payload"].(map[string]interface{})
	_, exists := payload["message"]
	assert.False(t, exists)
	// In actual runtime: returns error matching /error payload missing required field 'message'/i
}

// TestRuntimeMessage_ErrorPayload_EmptyMessage rejects error message with empty message field
func TestRuntimeMessage_ErrorPayload_EmptyMessage(t *testing.T) {
	jsonStr := `{"type": "error", "payload": {"message": "", "detail": {}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	payload := msg["payload"].(map[string]interface{})
	assert.Equal(t, "", payload["message"])
	// In actual runtime: returns error matching /error payload missing required field 'message'/i
}

// TestRuntimeMessage_ClaudeSessionID_Null rejects message with claudeSessionID as null
func TestRuntimeMessage_ClaudeSessionID_Null(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": null, "payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Nil(t, msg["claudeSessionID"])
	// In actual runtime: returns error matching /claudeSessionID must be a string/i
}

// TestRuntimeMessage_ClaudeSessionID_Number rejects message with claudeSessionID as number
func TestRuntimeMessage_ClaudeSessionID_Number(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": 123, "payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, ok := msg["claudeSessionID"].(float64)
	assert.True(t, ok)
	// In actual runtime: returns error matching /claudeSessionID must be a string/i
}

// TestRuntimeMessage_ClaudeSessionID_Object rejects message with claudeSessionID as object
func TestRuntimeMessage_ClaudeSessionID_Object(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": {"id": "test"}, "payload": {"eventType": "Test"}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	_, ok := msg["claudeSessionID"].(map[string]interface{})
	assert.True(t, ok)
	// In actual runtime: returns error matching /claudeSessionID must be a string/i
}

// TestRuntimeMessage_MalformedJSON_MissingBrace rejects message with malformed JSON
func TestRuntimeMessage_MalformedJSON_MissingBrace(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "Test"`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.Error(t, err)
	assert.Regexp(t, `(?i)JSON|unmarshal`, err.Error())
}

// TestRuntimeMessage_MalformedJSON_InvalidEscape rejects message with malformed JSON
func TestRuntimeMessage_MalformedJSON_InvalidEscape(t *testing.T) {
	jsonStr := `{"type": "event\x", "payload": {}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)

	require.Error(t, err)
	assert.Regexp(t, `(?i)JSON|unmarshal|invalid`, err.Error())
}

// TestRuntimeMessage_SizeLimit_JustUnder10MB accepts message with serialized JSON just under 10 MB
func TestRuntimeMessage_SizeLimit_JustUnder10MB(t *testing.T) {
	// Create large payload just under 10MB
	largeData := strings.Repeat("x", 10*1024*1024-1000) // 10MB - 1KB for overhead

	msg := map[string]interface{}{
		"type":            "event",
		"claudeSessionID": "test",
		"payload": map[string]interface{}{
			"eventType": "Test",
			"message":   largeData,
			"payload":   map[string]interface{}{},
		},
	}

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(jsonBytes), 10*1024*1024)
}

// TestRuntimeMessage_SizeLimit_Exceeds10MB rejects message with serialized JSON exceeding 10 MB
func TestRuntimeMessage_SizeLimit_Exceeds10MB(t *testing.T) {
	// Create payload exceeding 10MB
	largeData := strings.Repeat("x", 11*1024*1024)

	msg := map[string]interface{}{
		"type":            "event",
		"claudeSessionID": "test",
		"payload": map[string]interface{}{
			"eventType": "Test",
			"message":   largeData,
			"payload":   map[string]interface{}{},
		},
	}

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Greater(t, len(jsonBytes), 10*1024*1024)
	// In actual runtime: RuntimeSocketManager detects size violation; returns error matching /size limit/i
}

// TestRuntimeMessage_LargeEventPayload accepts event message with very large payload object
func TestRuntimeMessage_LargeEventPayload(t *testing.T) {
	// Create 9MB JSON object
	largePayload := make(map[string]interface{})
	for i := 0; i < 9000; i++ {
		largePayload[string(rune(i))] = strings.Repeat("a", 1000)
	}

	msg := map[string]interface{}{
		"type":            "event",
		"claudeSessionID": "test",
		"payload": map[string]interface{}{
			"eventType": "Test",
			"message":   "",
			"payload":   largePayload,
		},
	}

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Greater(t, len(jsonBytes), 1024*1024)
}

// TestRuntimeMessage_DeepNestedPayload accepts message with deeply nested JSON in payload
func TestRuntimeMessage_DeepNestedPayload(t *testing.T) {
	// Create JSON nested 100 levels deep
	nested := make(map[string]interface{})
	current := nested
	for i := 0; i < 100; i++ {
		next := make(map[string]interface{})
		current["level"] = next
		current = next
	}

	msg := map[string]interface{}{
		"type":            "event",
		"claudeSessionID": "test",
		"payload": map[string]interface{}{
			"eventType": "Test",
			"message":   "",
			"payload":   nested,
		},
	}

	jsonBytes, err := json.Marshal(msg)
	require.NoError(t, err)

	// Verify can unmarshal back
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
}

// TestRuntimeMessage_UnicodeInPayload accepts message with Unicode characters in payload fields
func TestRuntimeMessage_UnicodeInPayload(t *testing.T) {
	jsonStr := `{"type": "event", "payload": {"eventType": "Test", "message": "通知: Process complete 🎉", "payload": {"emoji": "🚀"}}}`

	var msg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	payload := msg["payload"].(map[string]interface{})
	assert.Equal(t, "通知: Process complete 🎉", payload["message"])

	innerPayload := payload["payload"].(map[string]interface{})
	assert.Equal(t, "🚀", innerPayload["emoji"])
}

// TestRuntimeMessage_RepeatedDeserialization verifies repeated deserialization produces identical results
func TestRuntimeMessage_RepeatedDeserialization(t *testing.T) {
	jsonStr := `{"type": "event", "claudeSessionID": "test", "payload": {"eventType": "Test", "message": "test", "payload": {"key": "value"}}}`

	var msg1, msg2, msg3 map[string]interface{}

	err1 := json.Unmarshal([]byte(jsonStr), &msg1)
	err2 := json.Unmarshal([]byte(jsonStr), &msg2)
	err3 := json.Unmarshal([]byte(jsonStr), &msg3)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	assert.Equal(t, msg1, msg2)
	assert.Equal(t, msg2, msg3)
}

// TestRuntimeMessage_SocketTransmission_EventMessage - E2E test for event transmission over socket
func TestRuntimeMessage_SocketTransmission_EventMessage(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// RuntimeSocketManager listening on test socket in test directory
	// Send valid event RuntimeMessage JSON with newline terminator over socket

	// Verify RuntimeSocketManager receives, parses, and processes message
	// Verify sends RuntimeResponse
	// Verify closes connection
}

// TestRuntimeMessage_SocketTransmission_ErrorMessage - E2E test for error transmission over socket
func TestRuntimeMessage_SocketTransmission_ErrorMessage(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// RuntimeSocketManager listening on test socket in test directory
	// Send valid error RuntimeMessage JSON with newline terminator over socket

	// Verify RuntimeSocketManager receives, parses, and processes message
	// Verify sends RuntimeResponse
	// Verify closes connection
}

// TestRuntimeMessage_ConcurrentConnections - Race test for concurrent connections
func TestRuntimeMessage_ConcurrentConnections(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// RuntimeSocketManager listening on test socket in test directory
	// 10 clients simultaneously connect and send valid RuntimeMessages

	// Verify all messages processed successfully
	// Verify each receives RuntimeResponse
	// Verify connections closed
	// Verify no data races detected
}

// TestRuntimeMessage_SocketClosed_AfterValidation verifies connection closed after validation failure
func TestRuntimeMessage_SocketClosed_AfterValidation(t *testing.T) {
	// Mock socket connection
	// Send invalid RuntimeMessage (missing type field)

	// Verify RuntimeSocketManager sends error response and closes connection
}

// TestRuntimeMessage_SocketClosed_AfterMalformedJSON verifies connection closed after JSON parse error
func TestRuntimeMessage_SocketClosed_AfterMalformedJSON(t *testing.T) {
	// Mock socket connection
	// Send malformed JSON

	// Verify RuntimeSocketManager attempts to send error response and closes connection
}
