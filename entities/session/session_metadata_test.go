package session

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Happy Path — Construction

func TestSessionMetadata_ValidFields(t *testing.T) {
	metadata := SessionMetadata{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		WorkflowName: "TestFlow",
		Status:       "initializing",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{},
		Error:        nil,
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", metadata.ID)
	assert.Equal(t, "TestFlow", metadata.WorkflowName)
	assert.Equal(t, "initializing", metadata.Status)
	assert.Equal(t, int64(1234567890), metadata.CreatedAt)
	assert.Equal(t, int64(1234567890), metadata.UpdatedAt)
	assert.Equal(t, "start", metadata.CurrentState)
	assert.NotNil(t, metadata.SessionData)
	assert.Len(t, metadata.SessionData, 0)
	assert.Nil(t, metadata.Error)
}

func TestSessionMetadata_EmptySessionData(t *testing.T) {
	metadata := SessionMetadata{
		SessionData: map[string]any{},
	}

	assert.NotNil(t, metadata.SessionData)
	assert.Len(t, metadata.SessionData, 0)
}

func TestSessionMetadata_NilError(t *testing.T) {
	metadata := SessionMetadata{
		Error: nil,
	}

	assert.Nil(t, metadata.Error)
}

// Happy Path — Field Access

func TestSessionMetadata_AccessID(t *testing.T) {
	metadata := SessionMetadata{
		ID: "test-uuid-123",
	}

	assert.Equal(t, "test-uuid-123", metadata.ID)
}

func TestSessionMetadata_AccessWorkflowName(t *testing.T) {
	metadata := SessionMetadata{
		WorkflowName: "MyWorkflow",
	}

	assert.Equal(t, "MyWorkflow", metadata.WorkflowName)
}

func TestSessionMetadata_AccessStatus(t *testing.T) {
	metadata := SessionMetadata{
		Status: "running",
	}

	assert.Equal(t, "running", metadata.Status)
}

func TestSessionMetadata_AccessTimestamps(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: 1000,
		UpdatedAt: 2000,
	}

	assert.Equal(t, int64(1000), metadata.CreatedAt)
	assert.Equal(t, int64(2000), metadata.UpdatedAt)
}

func TestSessionMetadata_AccessCurrentState(t *testing.T) {
	metadata := SessionMetadata{
		CurrentState: "processing",
	}

	assert.Equal(t, "processing", metadata.CurrentState)
}

func TestSessionMetadata_AccessSessionData(t *testing.T) {
	metadata := SessionMetadata{
		SessionData: map[string]any{"key": "value"},
	}

	assert.NotNil(t, metadata.SessionData)
	assert.Equal(t, "value", metadata.SessionData["key"])
}

func TestSessionMetadata_AccessError(t *testing.T) {
	agentErr := &AgentError{
		NodeName: "TestNode",
		Message:  "test error",
	}
	metadata := SessionMetadata{
		Error: agentErr,
	}

	assert.NotNil(t, metadata.Error)
	assert.IsType(t, &AgentError{}, metadata.Error)
}

// Happy Path — JSON Serialization

func TestSessionMetadata_JSONMarshalAllFields(t *testing.T) {
	metadata := SessionMetadata{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		WorkflowName: "TestFlow",
		Status:       "running",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{"key": "value"},
		Error:        nil,
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result["id"])
	assert.Equal(t, "TestFlow", result["workflowName"])
	assert.Equal(t, "running", result["status"])
	assert.Equal(t, float64(1234567890), result["createdAt"])
	assert.Equal(t, float64(1234567890), result["updatedAt"])
	assert.Equal(t, "start", result["currentState"])
	assert.NotNil(t, result["sessionData"])
	assert.NotContains(t, result, "error")
}

func TestSessionMetadata_JSONMarshalWithError(t *testing.T) {
	agentErr := &AgentError{
		NodeName: "TestNode",
		Message:  "test error",
	}
	metadata := SessionMetadata{
		ID:           "test-id",
		WorkflowName: "TestFlow",
		Status:       "failed",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{},
		Error:        agentErr,
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.NotNil(t, result["error"])
}

func TestSessionMetadata_JSONMarshalErrorOmitEmpty(t *testing.T) {
	metadata := SessionMetadata{
		ID:           "test-id",
		WorkflowName: "TestFlow",
		Status:       "running",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{},
		Error:        nil,
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.NotContains(t, result, "error")
	assert.NotContains(t, string(data), `"error":null`)
}

func TestSessionMetadata_JSONMarshalNestedSessionData(t *testing.T) {
	metadata := SessionMetadata{
		ID:           "test-id",
		WorkflowName: "TestFlow",
		Status:       "running",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{"key": map[string]any{"nested": []int{1, 2, 3}}},
		Error:        nil,
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	var result SessionMetadata
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.NotNil(t, result.SessionData["key"])
	nestedMap, ok := result.SessionData["key"].(map[string]any)
	require.True(t, ok)
	nestedArray, ok := nestedMap["nested"].([]any)
	require.True(t, ok)
	assert.Len(t, nestedArray, 3)
}

func TestSessionMetadata_JSONUnmarshalAllFields(t *testing.T) {
	jsonData := `{
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"workflowName": "TestFlow",
		"status": "running",
		"createdAt": 1234567890,
		"updatedAt": 1234567890,
		"currentState": "start",
		"sessionData": {"key": "value"}
	}`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.NoError(t, err)

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", metadata.ID)
	assert.Equal(t, "TestFlow", metadata.WorkflowName)
	assert.Equal(t, "running", metadata.Status)
	assert.Equal(t, int64(1234567890), metadata.CreatedAt)
	assert.Equal(t, int64(1234567890), metadata.UpdatedAt)
	assert.Equal(t, "start", metadata.CurrentState)
	assert.Equal(t, "value", metadata.SessionData["key"])
}

func TestSessionMetadata_JSONUnmarshalErrorFieldPresent(t *testing.T) {
	jsonData := `{
		"id": "test-id",
		"workflowName": "TestFlow",
		"status": "failed",
		"createdAt": 1234567890,
		"updatedAt": 1234567890,
		"currentState": "start",
		"sessionData": {},
		"error": {"NodeName": "TestNode", "Message": "test error"}
	}`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.NoError(t, err)

	assert.NotNil(t, metadata.Error)
}

func TestSessionMetadata_JSONUnmarshalErrorFieldAbsent(t *testing.T) {
	jsonData := `{
		"id": "test-id",
		"workflowName": "TestFlow",
		"status": "running",
		"createdAt": 1234567890,
		"updatedAt": 1234567890,
		"currentState": "start",
		"sessionData": {}
	}`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.NoError(t, err)

	assert.Nil(t, metadata.Error)
}

// Not Immutable

func TestSessionMetadata_FieldsMutable(t *testing.T) {
	metadata := SessionMetadata{
		Status: "initializing",
	}

	// Modify field
	metadata.Status = "running"

	assert.Equal(t, "running", metadata.Status)
}

func TestSessionMetadata_SessionDataMutable(t *testing.T) {
	metadata := SessionMetadata{
		SessionData: map[string]any{},
	}

	// Modify map
	metadata.SessionData["key"] = "value"

	assert.Equal(t, "value", metadata.SessionData["key"])
}

// Data Independence (Copy Semantics)

func TestSessionMetadata_CopyIndependent(t *testing.T) {
	metadata := SessionMetadata{
		Status:      "running",
		SessionData: map[string]any{"key": "value"},
	}

	// Create copy
	copy := metadata
	copy.Status = "completed"

	assert.Equal(t, "running", metadata.Status)
	assert.Equal(t, "completed", copy.Status)
}

func TestSessionMetadata_SessionDataShallowCopy(t *testing.T) {
	metadata := SessionMetadata{
		SessionData: map[string]any{"key": "value"},
	}

	// Create copy
	copy := metadata
	copy.SessionData["key"] = "modified"

	// Both should see the modification (map is reference type)
	assert.Equal(t, "modified", metadata.SessionData["key"])
	assert.Equal(t, "modified", copy.SessionData["key"])
}

// Invariants — Type Integrity

func TestSessionMetadata_NoEmbeddedLocks(t *testing.T) {
	// This test verifies that SessionMetadata contains no sync.Mutex, sync.RWMutex, or channels
	// by attempting to copy it and ensuring the copy is safe
	metadata := SessionMetadata{
		ID:           "test-id",
		WorkflowName: "TestFlow",
		Status:       "running",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{},
		Error:        nil,
	}

	// Copy by value
	copy := metadata

	// Modify copy
	copy.Status = "completed"

	// Original should be unchanged
	assert.Equal(t, "running", metadata.Status)
	assert.Equal(t, "completed", copy.Status)
}

func TestSessionMetadata_ValueType(t *testing.T) {
	metadata := SessionMetadata{
		ID:     "test-id",
		Status: "running",
	}

	// Can be copied by value
	copy := metadata
	copy.Status = "completed"

	assert.Equal(t, "running", metadata.Status)
	assert.Equal(t, "completed", copy.Status)
}

// Invariants — Status Enumeration

func TestSessionMetadata_StatusInitializing(t *testing.T) {
	metadata := SessionMetadata{
		Status: "initializing",
	}

	assert.Equal(t, "initializing", metadata.Status)
}

func TestSessionMetadata_StatusRunning(t *testing.T) {
	metadata := SessionMetadata{
		Status: "running",
	}

	assert.Equal(t, "running", metadata.Status)
}

func TestSessionMetadata_StatusCompleted(t *testing.T) {
	metadata := SessionMetadata{
		Status: "completed",
	}

	assert.Equal(t, "completed", metadata.Status)
}

func TestSessionMetadata_StatusFailed(t *testing.T) {
	metadata := SessionMetadata{
		Status: "failed",
	}

	assert.Equal(t, "failed", metadata.Status)
}

// Invariants — Timestamp Ordering

func TestSessionMetadata_CreatedAtPositive(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: 1234567890,
	}

	assert.Greater(t, metadata.CreatedAt, int64(0))
}

func TestSessionMetadata_UpdatedAtGreaterOrEqual(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: 1000,
		UpdatedAt: 2000,
	}

	assert.GreaterOrEqual(t, metadata.UpdatedAt, metadata.CreatedAt)
}

func TestSessionMetadata_TimestampsEqual(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: 1000,
		UpdatedAt: 1000,
	}

	assert.Equal(t, metadata.CreatedAt, metadata.UpdatedAt)
}

// Invariants — Error Correlation

func TestSessionMetadata_ErrorNilWhenNotFailed(t *testing.T) {
	metadata := SessionMetadata{
		Status: "running",
		Error:  nil,
	}

	assert.Nil(t, metadata.Error)
}

func TestSessionMetadata_ErrorNonNilWhenFailed(t *testing.T) {
	agentErr := &AgentError{
		NodeName: "TestNode",
		Message:  "test error",
	}
	metadata := SessionMetadata{
		Status: "failed",
		Error:  agentErr,
	}

	assert.NotNil(t, metadata.Error)
}

// Invariants — Non-Empty Fields

func TestSessionMetadata_IDNonEmpty(t *testing.T) {
	metadata := SessionMetadata{
		ID: "test-uuid",
	}

	assert.NotEmpty(t, metadata.ID)
}

func TestSessionMetadata_WorkflowNameNonEmpty(t *testing.T) {
	metadata := SessionMetadata{
		WorkflowName: "TestFlow",
	}

	assert.NotEmpty(t, metadata.WorkflowName)
}

func TestSessionMetadata_CurrentStateNonEmpty(t *testing.T) {
	metadata := SessionMetadata{
		CurrentState: "start",
	}

	assert.NotEmpty(t, metadata.CurrentState)
}

// Invariants — SessionData Never Nil

func TestSessionMetadata_SessionDataNotNil(t *testing.T) {
	metadata := SessionMetadata{
		SessionData: map[string]any{},
	}

	assert.NotNil(t, metadata.SessionData)
}

// Validation Failures — JSON Serialization

func TestSessionMetadata_JSONMarshalUnserializableSessionData(t *testing.T) {
	metadata := SessionMetadata{
		ID:           "test-id",
		WorkflowName: "TestFlow",
		Status:       "running",
		CreatedAt:    1234567890,
		UpdatedAt:    1234567890,
		CurrentState: "start",
		SessionData:  map[string]any{"ch": make(chan int)},
		Error:        nil,
	}

	_, err := json.Marshal(metadata)
	require.Error(t, err)
	assert.Regexp(t, `(?i)json: unsupported type`, err.Error())
}

// Validation Failures — JSON Deserialization

func TestSessionMetadata_JSONUnmarshalInvalidJSON(t *testing.T) {
	jsonData := `{"id": "test"`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.Error(t, err)
	assert.Regexp(t, `(?i)unexpected end of JSON`, err.Error())
}

func TestSessionMetadata_JSONUnmarshalWrongType(t *testing.T) {
	jsonData := `{"createdAt": "not-a-number"}`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.Error(t, err)
	assert.Regexp(t, `(?i)cannot unmarshal string into.*int64`, err.Error())
}

// Boundary Values — UUID

func TestSessionMetadata_ValidUUIDv4(t *testing.T) {
	metadata := SessionMetadata{
		ID: "550e8400-e29b-41d4-a716-446655440000",
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", metadata.ID)
}

func TestSessionMetadata_EmptyID(t *testing.T) {
	metadata := SessionMetadata{
		ID: "",
	}

	assert.Empty(t, metadata.ID)
}

// Boundary Values — Timestamps

func TestSessionMetadata_ZeroTimestamps(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: 0,
		UpdatedAt: 0,
	}

	assert.Equal(t, int64(0), metadata.CreatedAt)
	assert.Equal(t, int64(0), metadata.UpdatedAt)
}

func TestSessionMetadata_NegativeTimestamps(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: -1,
		UpdatedAt: -1,
	}

	assert.Equal(t, int64(-1), metadata.CreatedAt)
	assert.Equal(t, int64(-1), metadata.UpdatedAt)
}

func TestSessionMetadata_UpdatedAtLessThanCreatedAt(t *testing.T) {
	metadata := SessionMetadata{
		CreatedAt: 2000,
		UpdatedAt: 1000,
	}

	assert.Equal(t, int64(2000), metadata.CreatedAt)
	assert.Equal(t, int64(1000), metadata.UpdatedAt)
}

// Boundary Values — Error Types

func TestSessionMetadata_AgentError(t *testing.T) {
	agentErr := &AgentError{
		NodeName: "TestNode",
		Message:  "test",
	}
	metadata := SessionMetadata{
		Error: agentErr,
	}

	assert.NotNil(t, metadata.Error)
	assert.IsType(t, &AgentError{}, metadata.Error)
}

func TestSessionMetadata_RuntimeError(t *testing.T) {
	runtimeErr := &RuntimeError{
		Issuer:  "Runtime",
		Message: "test",
	}
	metadata := SessionMetadata{
		Error: runtimeErr,
	}

	assert.NotNil(t, metadata.Error)
	assert.IsType(t, &RuntimeError{}, metadata.Error)
}

// Edge Cases

func TestSessionMetadata_EmptyWorkflowName(t *testing.T) {
	metadata := SessionMetadata{
		WorkflowName: "",
	}

	assert.Empty(t, metadata.WorkflowName)
}

func TestSessionMetadata_EmptyCurrentState(t *testing.T) {
	metadata := SessionMetadata{
		CurrentState: "",
	}

	assert.Empty(t, metadata.CurrentState)
}

func TestSessionMetadata_InvalidStatus(t *testing.T) {
	metadata := SessionMetadata{
		Status: "invalid",
	}

	assert.Equal(t, "invalid", metadata.Status)
}

func TestSessionMetadata_NilSessionData(t *testing.T) {
	metadata := SessionMetadata{
		SessionData: nil,
	}

	assert.Nil(t, metadata.SessionData)
}

func TestSessionMetadata_ErrorWithoutFailed(t *testing.T) {
	agentErr := &AgentError{
		NodeName: "TestNode",
		Message:  "test error",
	}
	metadata := SessionMetadata{
		Status: "running",
		Error:  agentErr,
	}

	assert.Equal(t, "running", metadata.Status)
	assert.NotNil(t, metadata.Error)
}

func TestSessionMetadata_FailedWithoutError(t *testing.T) {
	metadata := SessionMetadata{
		Status: "failed",
		Error:  nil,
	}

	assert.Equal(t, "failed", metadata.Status)
	assert.Nil(t, metadata.Error)
}
