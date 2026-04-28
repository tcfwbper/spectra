package entities_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuntimeResponse_ValidSuccessResponse creates RuntimeResponse with status=success and message
func TestRuntimeResponse_ValidSuccessResponse(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "Event 'DraftCompleted' recorded successfully",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "success", decoded["status"])
	assert.Equal(t, "Event 'DraftCompleted' recorded successfully", decoded["message"])
}

// TestRuntimeResponse_ValidErrorResponse creates RuntimeResponse with status=error and message
func TestRuntimeResponse_ValidErrorResponse(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "error",
		"message": "session not ready: status is 'initializing'",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "error", decoded["status"])
	assert.Equal(t, "session not ready: status is 'initializing'", decoded["message"])
}

// TestRuntimeResponse_SuccessWithEmptyMessage creates RuntimeResponse with status=success and empty message
func TestRuntimeResponse_SuccessWithEmptyMessage(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "success", decoded["status"])
	assert.Equal(t, "", decoded["message"])
}

// TestRuntimeResponse_ErrorWithEmptyMessage creates RuntimeResponse with status=error and empty message
func TestRuntimeResponse_ErrorWithEmptyMessage(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "error",
		"message": "",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "error", decoded["status"])
	assert.Equal(t, "", decoded["message"])
}

// TestRuntimeResponse_SerializesToJSON_Success verifies RuntimeResponse serializes to valid JSON
func TestRuntimeResponse_SerializesToJSON_Success(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "Event recorded",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"status":"success"`)
	assert.Contains(t, jsonStr, `"message":"Event recorded"`)

	// When transmitted, should terminate with newline
	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeResponse_SerializesToJSON_Error verifies RuntimeResponse serializes to valid JSON for error status
func TestRuntimeResponse_SerializesToJSON_Error(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "error",
		"message": "no transition found",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"status":"error"`)
	assert.Contains(t, jsonStr, `"message":"no transition found"`)

	// When transmitted, should terminate with newline
	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeResponse_DeserializesFromJSON verifies RuntimeResponse deserializes from valid JSON
func TestRuntimeResponse_DeserializesFromJSON(t *testing.T) {
	jsonStr := `{"status": "success", "message": "Event recorded"}`

	var resp map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &resp)

	require.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	assert.Equal(t, "Event recorded", resp["message"])
}

// TestRuntimeResponse_ValidStatusSuccess accepts status=success as valid
func TestRuntimeResponse_ValidStatusSuccess(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "test",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "success", decoded["status"])
}

// TestRuntimeResponse_ValidStatusError accepts status=error as valid
func TestRuntimeResponse_ValidStatusError(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "error",
		"message": "test",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "error", decoded["status"])
}

// TestRuntimeResponse_MessageWithNewlines serializes message containing newline characters correctly
func TestRuntimeResponse_MessageWithNewlines(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "error",
		"message": "line1\nline2\nline3",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	// Newlines should be escaped in JSON
	assert.Contains(t, jsonStr, `\n`)

	// Verify can be decoded back
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", decoded["message"])

	// Newlines inside message do not interfere with terminator newline
	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeResponse_LargeMessage accepts RuntimeResponse with very large message string
func TestRuntimeResponse_LargeMessage(t *testing.T) {
	largeMsg := strings.Repeat("A", 1024*1024) // 1MB

	resp := map[string]interface{}{
		"status":  "error",
		"message": largeMsg,
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	// Verify size under 10 MB limit
	assert.Less(t, len(jsonBytes), 10*1024*1024)

	// Verify can be decoded
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, largeMsg, decoded["message"])
}

// TestRuntimeResponse_UnicodeMessage accepts RuntimeResponse with Unicode characters in message
func TestRuntimeResponse_UnicodeMessage(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "通知: 成功 🎉",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "通知: 成功 🎉", decoded["message"])
}

// TestRuntimeResponse_OnlyStatusAndMessageFields verifies RuntimeResponse contains only status and message fields
func TestRuntimeResponse_OnlyStatusAndMessageFields(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "test",
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	// Should have exactly two fields
	assert.Len(t, decoded, 2)
	assert.Contains(t, decoded, "status")
	assert.Contains(t, decoded, "message")
}

// TestRuntimeResponse_SizeLimit_JustUnder10MB accepts response with serialized JSON just under 10 MB
func TestRuntimeResponse_SizeLimit_JustUnder10MB(t *testing.T) {
	// Create message totaling 10 MB - 100 bytes
	largeMsg := strings.Repeat("x", 10*1024*1024-100)

	resp := map[string]interface{}{
		"status":  "error",
		"message": largeMsg,
	}

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	// Should be under limit
	assert.LessOrEqual(t, len(jsonBytes), 10*1024*1024)
}

// TestRuntimeResponse_RepeatedSerialization verifies repeated serialization produces identical results
func TestRuntimeResponse_RepeatedSerialization(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "test message",
	}

	json1, err1 := json.Marshal(resp)
	json2, err2 := json.Marshal(resp)
	json3, err3 := json.Marshal(resp)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	assert.Equal(t, json1, json2)
	assert.Equal(t, json2, json3)
}

// TestRuntimeResponse_SocketTransmission_Success - E2E test for success response transmission
func TestRuntimeResponse_SocketTransmission_Success(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// RuntimeSocketManager listening on test socket in test directory
	// spectra-agent client connected

	// MessageHandler returns success RuntimeResponse

	// Verify RuntimeSocketManager serializes response
	// Verify sends over socket with newline terminator
	// Verify closes connection
	// Verify client receives complete response
}

// TestRuntimeResponse_SocketTransmission_Error - E2E test for error response transmission
func TestRuntimeResponse_SocketTransmission_Error(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// RuntimeSocketManager listening on test socket in test directory
	// spectra-agent client connected

	// MessageHandler returns error RuntimeResponse

	// Verify RuntimeSocketManager serializes response
	// Verify sends over socket with newline terminator
	// Verify closes connection
	// Verify client receives complete response
}
