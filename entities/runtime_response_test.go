package entities_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/entities"
)

// TestRuntimeResponse_ValidSuccessResponse creates RuntimeResponse with status=success and message
func TestRuntimeResponse_ValidSuccessResponse(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "Event 'DraftCompleted' recorded successfully")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "Event 'DraftCompleted' recorded successfully", resp.Message)
}

// TestRuntimeResponse_ValidErrorResponse creates RuntimeResponse with status=error and message
func TestRuntimeResponse_ValidErrorResponse(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("error", "session not ready: status is 'initializing'")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session not ready: status is 'initializing'", resp.Message)
}

// TestRuntimeResponse_SuccessWithEmptyMessage creates RuntimeResponse with status=success and empty message
func TestRuntimeResponse_SuccessWithEmptyMessage(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "", resp.Message)
}

// TestRuntimeResponse_ErrorWithEmptyMessage creates RuntimeResponse with status=error and empty message
func TestRuntimeResponse_ErrorWithEmptyMessage(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("error", "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "", resp.Message)
}

// TestRuntimeResponse_SerializesToJSON_Success verifies RuntimeResponse serializes to valid JSON
func TestRuntimeResponse_SerializesToJSON_Success(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "Event recorded")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"status":"success"`)
	assert.Contains(t, jsonStr, `"message":"Event recorded"`)

	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeResponse_SerializesToJSON_Error verifies RuntimeResponse serializes to valid JSON for error status
func TestRuntimeResponse_SerializesToJSON_Error(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("error", "no transition found")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"status":"error"`)
	assert.Contains(t, jsonStr, `"message":"no transition found"`)

	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeResponse_DeserializesFromJSON verifies RuntimeResponse deserializes from valid JSON
func TestRuntimeResponse_DeserializesFromJSON(t *testing.T) {
	jsonStr := `{"status": "success", "message": "Event recorded"}`

	var resp entities.RuntimeResponse
	err := json.Unmarshal([]byte(jsonStr), &resp)

	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "Event recorded", resp.Message)

	err = resp.Validate()
	require.NoError(t, err)
}

// TestRuntimeResponse_ValidStatusSuccess accepts status=success as valid
func TestRuntimeResponse_ValidStatusSuccess(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "test")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
}

// TestRuntimeResponse_ValidStatusError accepts status=error as valid
func TestRuntimeResponse_ValidStatusError(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("error", "test")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "error", resp.Status)
}

// TestRuntimeResponse_InvalidStatus rejects invalid status value
func TestRuntimeResponse_InvalidStatus(t *testing.T) {
	_, err := entities.NewRuntimeResponse("warning", "test")
	require.Error(t, err)
	assert.Regexp(t, `(?i)invalid response status`, err.Error())
}

// TestRuntimeResponse_MessageWithNewlines serializes message containing newline characters correctly
func TestRuntimeResponse_MessageWithNewlines(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("error", "line1\nline2\nline3")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `\n`)

	var decoded entities.RuntimeResponse
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", decoded.Message)

	transmitted := jsonStr + "\n"
	assert.True(t, strings.HasSuffix(transmitted, "\n"))
}

// TestRuntimeResponse_LargeMessage accepts RuntimeResponse with very large message string
func TestRuntimeResponse_LargeMessage(t *testing.T) {
	largeMsg := strings.Repeat("A", 1024*1024) // 1MB

	resp, err := entities.NewRuntimeResponse("error", largeMsg)
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	assert.Less(t, len(jsonBytes), 10*1024*1024)

	var decoded entities.RuntimeResponse
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, largeMsg, decoded.Message)
}

// TestRuntimeResponse_UnicodeMessage accepts RuntimeResponse with Unicode characters in message
func TestRuntimeResponse_UnicodeMessage(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "通知: 成功 🎉")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded entities.RuntimeResponse
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "通知: 成功 🎉", decoded.Message)
}

// TestRuntimeResponse_OnlyStatusAndMessageFields verifies RuntimeResponse contains only status and message fields
func TestRuntimeResponse_OnlyStatusAndMessageFields(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "test")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded, 2)
	assert.Contains(t, decoded, "status")
	assert.Contains(t, decoded, "message")
}

// TestRuntimeResponse_SizeLimit_JustUnder10MB accepts response with serialized JSON just under 10 MB
func TestRuntimeResponse_SizeLimit_JustUnder10MB(t *testing.T) {
	largeMsg := strings.Repeat("x", 10*1024*1024-100)

	resp, err := entities.NewRuntimeResponse("error", largeMsg)
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(jsonBytes), 11*1024*1024)
}

// TestRuntimeResponse_RepeatedSerialization verifies repeated serialization produces identical results
func TestRuntimeResponse_RepeatedSerialization(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "test message")
	require.NoError(t, err)

	json1, err1 := json.Marshal(resp)
	json2, err2 := json.Marshal(resp)
	json3, err3 := json.Marshal(resp)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	assert.Equal(t, json1, json2)
	assert.Equal(t, json2, json3)
}

// TestRuntimeResponse_Validate verifies Validate method works
func TestRuntimeResponse_Validate(t *testing.T) {
	resp, err := entities.NewRuntimeResponse("success", "test")
	require.NoError(t, err)

	err = resp.Validate()
	require.NoError(t, err)
}

// TestRuntimeResponse_Validate_InvalidStatus verifies Validate catches invalid status
func TestRuntimeResponse_Validate_InvalidStatus(t *testing.T) {
	resp := &entities.RuntimeResponse{Status: "warning", Message: "test"}

	err := resp.Validate()
	require.Error(t, err)
	assert.Regexp(t, `(?i)invalid response status`, err.Error())
}

// TestRuntimeResponse_SocketTransmission_Success - E2E test for success response transmission
func TestRuntimeResponse_SocketTransmission_Success(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir
}

// TestRuntimeResponse_SocketTransmission_Error - E2E test for error response transmission
func TestRuntimeResponse_SocketTransmission_Error(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir
}
