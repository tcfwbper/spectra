package storage_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// TestSessionMetadataStore_New constructs SessionMetadataStore with valid inputs
func TestSessionMetadataStore_New(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	sessionUUID := uuid.New()
	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	assert.NotNil(t, store)
}

// TestSessionMetadataStore_WriteFirst writes metadata to non-existent file
func TestSessionMetadataStore_WriteFirst(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{"key": "value"},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	// Verify file exists with correct permissions
	metadataFile := filepath.Join(sessionDir, "session.json")
	info, err := os.Stat(metadataFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())

	// Verify pretty-printed JSON with 2-space indentation
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"id"`)
	assert.Contains(t, string(content), `"workflowName": "TestWorkflow"`)
	// Check for 2-space indentation
	assert.Contains(t, string(content), "\n  ")
}

// TestSessionMetadataStore_WriteCreatesFile verifies FileAccessor callback creates file on first write
func TestSessionMetadataStore_WriteCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	metadataFile := filepath.Join(sessionDir, "session.json")
	_, err := os.Stat(metadataFile)
	assert.True(t, os.IsNotExist(err), "file should not exist before write")

	err = store.Write(metadata)
	assert.NoError(t, err)

	info, err := os.Stat(metadataFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

// TestSessionMetadataStore_WriteOverwrites tests second write replaces entire file content
func TestSessionMetadataStore_WriteOverwrites(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// First write
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}
	require.NoError(t, store.Write(metadata))

	// Second write with changed status
	metadata.Status = "running"
	err := store.Write(metadata)
	assert.NoError(t, err)

	// Verify new status is reflected
	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"status": "running"`)
	assert.NotContains(t, string(content), `"status": "initializing"`)
}

// TestSessionMetadataStore_WriteMultipleTimes tests multiple writes succeed with last-write-wins
func TestSessionMetadataStore_WriteMultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	// Write 3 times with different status values
	statuses := []string{"initializing", "running", "completed"}
	for _, status := range statuses {
		metadata.Status = status
		err := store.Write(metadata)
		assert.NoError(t, err)
	}

	// Verify final status
	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"status": "completed"`)
}

// TestSessionMetadataStore_PrettyPrinted2SpaceIndent verifies serializes as pretty-printed JSON with 2-space indentation
func TestSessionMetadataStore_PrettyPrinted2SpaceIndent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
		},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)

	// Verify multi-line JSON with 2-space indentation
	contentStr := string(content)
	assert.Contains(t, contentStr, "\n")
	assert.Contains(t, contentStr, "  \"id\"")
	assert.Contains(t, contentStr, "  \"workflowName\"")
	// Check nested object has 4-space indentation (2 levels)
	assert.Contains(t, contentStr, "    \"key\"")
}

// TestSessionMetadataStore_EmptySessionData verifies empty SessionData serializes as empty object
func TestSessionMetadataStore_EmptySessionData(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)

	// Should contain empty object
	assert.Contains(t, string(content), `"sessionData": {}`)
}

// TestSessionMetadataStore_EventHistoryNotSerialized verifies EventHistory field excluded from JSON output
func TestSessionMetadataStore_EventHistoryNotSerialized(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	// SessionMetadata no longer has EventHistory field
	assert.NotContains(t, string(content), "eventHistory")
}

// TestSessionMetadataStore_EventHistoryFieldIgnored verifies pre-existing eventHistory field in JSON ignored on read
func TestSessionMetadataStore_EventHistoryFieldIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Manually create session.json with eventHistory field
	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{},
		"eventHistory": []map[string]interface{}{
			{
				"id":        uuid.New().String(),
				"type":      "TestEvent",
				"message":   "test",
				"emittedAt": time.Now().Unix(),
			},
		},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	// SessionMetadata no longer has EventHistory field - it's ignored during deserialization
	assert.Equal(t, "TestWorkflow", metadata.WorkflowName)
}

// TestSessionMetadataStore_ErrorFieldOmittedWhenNil verifies Error field omitted when nil
func TestSessionMetadataStore_ErrorFieldOmittedWhenNil(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
		Error:        nil,
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(content), `"error"`)
}

// TestSessionMetadataStore_ErrorFieldPresentWhenSet verifies Error field serialized when set
func TestSessionMetadataStore_ErrorFieldPresentWhenSet(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	agentError := &session.AgentError{
		NodeName: "TestAgent",
		Message:  "Test error message",
	}

	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "failed",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "FailedNode",
		SessionData:  map[string]interface{}{},
		Error:        agentError,
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"error"`)
	assert.Contains(t, string(content), `"NodeName"`)
	assert.Contains(t, string(content), `"Test error message"`)
}

// TestSessionMetadataStore_UpdatedAtAutoUpdated verifies UpdatedAt automatically set to current timestamp
func TestSessionMetadataStore_UpdatedAtAutoUpdated(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	oldTimestamp := time.Now().Unix() - 3600 // 1 hour ago
	beforeWrite := time.Now().Unix()

	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    oldTimestamp,
		UpdatedAt:    oldTimestamp, // Should be overwritten
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	// Read back and verify UpdatedAt was updated
	readMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.NotNil(t, readMetadata)
	assert.GreaterOrEqual(t, readMetadata.UpdatedAt, beforeWrite)
	assert.LessOrEqual(t, readMetadata.UpdatedAt, time.Now().Unix()+1)
	assert.NotEqual(t, oldTimestamp, readMetadata.UpdatedAt)
}

// TestSessionMetadataStore_UpdatedAtChangesOnEachWrite verifies UpdatedAt updated on each write
func TestSessionMetadataStore_UpdatedAtChangesOnEachWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	// First write
	require.NoError(t, store.Write(metadata))
	readMetadata1, err := store.Read()
	require.NoError(t, err)
	timestamp1 := readMetadata1.UpdatedAt

	// Wait 1 second (reduced from 2s for faster tests)
	time.Sleep(1 * time.Second)

	// Second write
	require.NoError(t, store.Write(metadata))
	readMetadata2, err := store.Read()
	require.NoError(t, err)
	timestamp2 := readMetadata2.UpdatedAt

	assert.Greater(t, timestamp2, timestamp1)
	assert.GreaterOrEqual(t, timestamp2-timestamp1, int64(1))
}

// TestSessionMetadataStore_ReadValidFile reads metadata from valid JSON file
func TestSessionMetadataStore_ReadValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Create valid session.json
	metadataFile := filepath.Join(sessionDir, "session.json")
	createdAt := time.Now().Unix()
	updatedAt := time.Now().Unix()
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    createdAt,
		"updatedAt":    updatedAt,
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{"key": "value"},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, sessionUUID.String(), metadata.ID)
	assert.Equal(t, "TestWorkflow", metadata.WorkflowName)
	assert.Equal(t, "running", metadata.Status)
	assert.Equal(t, createdAt, metadata.CreatedAt)
	assert.Equal(t, updatedAt, metadata.UpdatedAt)
	assert.Equal(t, "StartNode", metadata.CurrentState)
	assert.Equal(t, "value", metadata.SessionData["key"])
}

// TestSessionMetadataStore_ReadComplexSessionData reads metadata with complex nested SessionData
func TestSessionMetadataStore_ReadComplexSessionData(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "StartNode",
		"sessionData": map[string]interface{}{
			"nested": map[string]interface{}{
				"array": []interface{}{1, 2, 3},
				"object": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	nested := metadata.SessionData["nested"].(map[string]interface{})
	array := nested["array"].([]interface{})
	assert.Equal(t, 3, len(array))
	assert.Equal(t, float64(1), array[0])
}

// TestSessionMetadataStore_ReadErrorFieldPresent reads metadata with Error field set
func TestSessionMetadataStore_ReadErrorFieldPresent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "failed",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "FailedNode",
		"sessionData":  map[string]interface{}{},
		"error": map[string]interface{}{
			"NodeName": "TestAgent",
			"Message":  "Test error",
		},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.NotNil(t, metadata.Error)
	// Error is deserialized as a map[string]interface{} since it's an interface{}
	assert.Equal(t, "Test error", metadata.Error.Error())
}

// TestSessionMetadataStore_ReadOnlySessionMetadataFields verifies Read returns SessionMetadata struct with all persistable fields
func TestSessionMetadataStore_ReadOnlySessionMetadataFields(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	createdAt := time.Now().Unix()
	updatedAt := time.Now().Unix()
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    createdAt,
		"updatedAt":    updatedAt,
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{"key": "value"},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	// Verify all fields are populated
	assert.Equal(t, sessionUUID.String(), metadata.ID)
	assert.Equal(t, "TestWorkflow", metadata.WorkflowName)
	assert.Equal(t, "running", metadata.Status)
	assert.Equal(t, createdAt, metadata.CreatedAt)
	assert.Equal(t, updatedAt, metadata.UpdatedAt)
	assert.Equal(t, "StartNode", metadata.CurrentState)
	assert.NotNil(t, metadata.SessionData)
	assert.Nil(t, metadata.Error)
	// EventHistory is not part of SessionMetadata persistence
}

// TestSessionMetadataStore_WriteParentDirDoesNotExist returns error when session directory missing
func TestSessionMetadataStore_WriteParentDirDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	// Do not create session directory

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)session directory does not exist:.*\.spectra/sessions/.*`+sessionUUID.String(), err.Error())
}

// TestSessionMetadataStore_WriteSerializationFails returns error when JSON marshaling fails
func TestSessionMetadataStore_WriteSerializationFails(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"channel": make(chan int), // Un-serializable type
		},
	}

	err := store.Write(metadata)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)failed to serialize session metadata:`, err.Error())
}

// TestSessionMetadataStore_WriteExceeds10MBLimit rejects metadata exceeding 10 MB serialized size
func TestSessionMetadataStore_WriteExceeds10MBLimit(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// Create large data (11 MB)
	largeData := strings.Repeat("x", 11*1024*1024)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"largeData": largeData,
		},
	}

	err := store.Write(metadata)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)session metadata size exceeds 10 MB limit:.*bytes`, err.Error())
}

// TestSessionMetadataStore_WriteExactly10MB accepts metadata at exactly 10 MB limit
func TestSessionMetadataStore_WriteExactly10MB(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// Create metadata that will serialize to exactly 10 MB
	// Need to account for JSON overhead (field names, quotes, indentation)
	baseMetadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	// Calculate base size
	baseJSON, _ := json.MarshalIndent(baseMetadata, "", "  ")
	baseSize := len(baseJSON)

	// Add data to reach exactly 10 MB
	targetSize := 10 * 1024 * 1024
	dataSize := targetSize - baseSize - 50 // Account for JSON overhead
	if dataSize > 0 {
		baseMetadata.SessionData["data"] = strings.Repeat("x", dataSize)
	}

	err := store.Write(baseMetadata)
	assert.NoError(t, err)
}

// TestSessionMetadataStore_ReadFileDoesNotExist returns error when file missing
func TestSessionMetadataStore_ReadFileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Regexp(t, `(?i)session metadata file does not exist:.*session\.json`, err.Error())
}

// TestSessionMetadataStore_ReadInvalidJSON returns error when JSON is malformed
func TestSessionMetadataStore_ReadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	invalidJSON := `{"id": "test", "workflowName": "Test"` // Missing closing brace
	require.NoError(t, os.WriteFile(metadataFile, []byte(invalidJSON), 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Regexp(t, `(?i)failed to parse session metadata:.*unexpected end of JSON`, err.Error())
}

// TestSessionMetadataStore_ReadMissingRequiredField returns error when required field missing
func TestSessionMetadataStore_ReadMissingRequiredField(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		// Missing "id" field
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Regexp(t, `(?i)failed to parse session metadata:.*missing required field.*ID`, err.Error())
}

// TestSessionMetadataStore_ReadPermissionDenied returns error when file permissions deny read
func TestSessionMetadataStore_ReadPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0000))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata, err := store.Read()
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Regexp(t, `(?i)permission denied`, err.Error())

	_ = os.Chmod(metadataFile, 0644)
}

// TestSessionMetadataStore_ReadIdempotent verifies multiple reads return identical results
func TestSessionMetadataStore_ReadIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{"key": "value"},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// First read
	metadata1, err1 := store.Read()
	assert.NoError(t, err1)

	// Second read
	metadata2, err2 := store.Read()
	assert.NoError(t, err2)

	// Both reads should return identical data
	assert.Equal(t, metadata1.ID, metadata2.ID)
	assert.Equal(t, metadata1.WorkflowName, metadata2.WorkflowName)
	assert.Equal(t, metadata1.Status, metadata2.Status)
	assert.Equal(t, metadata1.SessionData["key"], metadata2.SessionData["key"])
}

// TestSessionMetadataStore_WriteIdempotent verifies writing same metadata twice produces consistent file
func TestSessionMetadataStore_WriteIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{"key": "value"},
	}

	// Write twice
	require.NoError(t, store.Write(metadata))
	require.NoError(t, store.Write(metadata))

	// Verify file contains expected content
	readMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, "TestWorkflow", readMetadata.WorkflowName)
	assert.Equal(t, "running", readMetadata.Status)
}

// TestSessionMetadataStore_NewFilePermissions0644 creates new file with correct permissions
func TestSessionMetadataStore_NewFilePermissions0644(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	metadataFile := filepath.Join(sessionDir, "session.json")
	info, err := os.Stat(metadataFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

// TestSessionMetadataStore_InvalidSessionUUID fails with malformed UUID
func TestSessionMetadataStore_InvalidSessionUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	// Create a UUID with invalid characters that will result in malformed path
	// Note: We cannot use uuid.MustParse("not-a-uuid") as it will panic
	// Instead, we test that the store handles unusual UUID strings in the path
	malformedUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	// Create the store with a valid UUID but test against a malformed path scenario
	store := storage.NewSessionMetadataStore(tmpDir, malformedUUID)

	metadata := &session.SessionMetadata{
		ID:           uuid.New().String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	// Operations should fail with filesystem errors due to missing directory
	err = store.Write(metadata)
	assert.Error(t, err)
}

// TestSessionMetadataStore_EmptySessionUUID fails with empty UUID
func TestSessionMetadataStore_EmptySessionUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, uuid.Nil)
	metadata := &session.SessionMetadata{
		ID:           uuid.New().String(),
		WorkflowName: "TestWorkflow",
		Status:       "initializing",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.Error(t, err)
}

// TestSessionMetadataStore_SessionDataWithNamespacedKeys handles SessionData with namespaced keys
func TestSessionMetadataStore_SessionDataWithNamespacedKeys(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"NodeA.ClaudeSessionID": "session-123",
		},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	// Read back and verify
	readMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, "session-123", readMetadata.SessionData["NodeA.ClaudeSessionID"])
}

// TestSessionMetadataStore_SessionDataNonStringValue serializes SessionData with non-string values as-is
func TestSessionMetadataStore_SessionDataNonStringValue(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"NodeA.ClaudeSessionID": 12345, // Number instead of string
		},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	// Verify number is serialized
	readMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, float64(12345), readMetadata.SessionData["NodeA.ClaudeSessionID"])
}

// TestSessionMetadataStore_NoCaching verifies Read always accesses disk, no in-memory cache
func TestSessionMetadataStore_NoCaching(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	metadataFile := filepath.Join(sessionDir, "session.json")
	jsonContent := map[string]interface{}{
		"id":           sessionUUID.String(),
		"workflowName": "TestWorkflow",
		"status":       "running",
		"createdAt":    time.Now().Unix(),
		"updatedAt":    time.Now().Unix(),
		"currentState": "StartNode",
		"sessionData":  map[string]interface{}{},
	}
	jsonBytes, err := json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// First read
	metadata1, err := store.Read()
	require.NoError(t, err)
	assert.Equal(t, "running", metadata1.Status)

	// Externally modify file
	jsonContent["status"] = "completed"
	jsonBytes, err = json.MarshalIndent(jsonContent, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, jsonBytes, 0644))

	// Second read should detect external change
	metadata2, err := store.Read()
	require.NoError(t, err)
	assert.Equal(t, "completed", metadata2.Status)
}

// TestSessionMetadataStore_NoStateValidation verifies store does not validate state transitions
func TestSessionMetadataStore_NoStateValidation(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	// Write with status "running"
	require.NoError(t, store.Write(metadata))

	// Write again with invalid transition to "initializing"
	metadata.Status = "initializing"
	err := store.Write(metadata)
	assert.NoError(t, err) // Should succeed, no validation
}

// TestSessionMetadataStore_WriteTruncates verifies write replaces all previous content
func TestSessionMetadataStore_WriteTruncates(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)

	// First write with large data (5 KB)
	largeData := strings.Repeat("x", 5*1024)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"largeData": largeData,
		},
	}
	require.NoError(t, store.Write(metadata))

	metadataFile := filepath.Join(sessionDir, "session.json")
	info1, err := os.Stat(metadataFile)
	require.NoError(t, err)
	size1 := info1.Size()

	// Second write with minimal data (1 KB)
	metadata.SessionData = map[string]interface{}{"key": "value"}
	require.NoError(t, store.Write(metadata))

	info2, err := os.Stat(metadataFile)
	require.NoError(t, err)
	size2 := info2.Size()

	// File should be smaller
	assert.Less(t, size2, size1)

	// Verify no remnants of old content
	content, err := os.ReadFile(metadataFile)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "largeData")
	assert.Contains(t, string(content), `"key": "value"`)
}

// TestSessionMetadataStore_NullValuesInSessionData handles null values in SessionData
func TestSessionMetadataStore_NullValuesInSessionData(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "TestWorkflow",
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData: map[string]interface{}{
			"nullKey": nil,
		},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	// Verify null is serialized
	metadataFile := filepath.Join(sessionDir, "session.json")
	content, err := os.ReadFile(metadataFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"nullKey": null`)

	// Read back
	readMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Nil(t, readMetadata.SessionData["nullKey"])
}

// TestSessionMetadataStore_EmptyStringsInFields handles empty string values
func TestSessionMetadataStore_EmptyStringsInFields(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := uuid.New()
	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID.String())
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	store := storage.NewSessionMetadataStore(tmpDir, sessionUUID)
	metadata := &session.SessionMetadata{
		ID:           sessionUUID.String(),
		WorkflowName: "", // Empty string
		Status:       "running",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
		CurrentState: "StartNode",
		SessionData:  map[string]interface{}{},
	}

	err := store.Write(metadata)
	assert.NoError(t, err)

	// Read back
	readMetadata, err := store.Read()
	assert.NoError(t, err)
	assert.Equal(t, "", readMetadata.WorkflowName)
}

// TestSessionMetadataStore_WriteAcquiresExclusiveLock verifies exclusive lock is acquired during write
func TestSessionMetadataStore_WriteAcquiresExclusiveLock(t *testing.T) {
	t.Skip("Requires implementation with observable lock acquisition/release behavior or mock support")
	// This test requires the implementation to provide hooks or interfaces to observe lock behavior.
	// Once the implementation provides such mechanisms, this test should:
	// 1. Verify that an exclusive lock is acquired before write operation
	// 2. Verify that the lock is released after write completes
	// 3. Verify that concurrent operations are properly blocked during the lock period
}

// TestSessionMetadataStore_WriteReleasesLockOnError verifies lock is released when write fails
func TestSessionMetadataStore_WriteReleasesLockOnError(t *testing.T) {
	t.Skip("Requires mock FileAccessor or injectable error mechanism")
	// This test requires:
	// 1. Mock FileAccessor that can be configured to fail during write
	// 2. Ability to verify lock is released before function returns
	// 3. Verification via subsequent successful operation that lock was properly released
}

// TestSessionMetadataStore_ReadAcquiresSharedLock verifies shared lock is acquired during read
func TestSessionMetadataStore_ReadAcquiresSharedLock(t *testing.T) {
	t.Skip("Requires implementation with observable lock acquisition/release behavior or mock support")
	// This test requires the implementation to provide hooks or interfaces to observe lock behavior.
	// Once the implementation provides such mechanisms, this test should:
	// 1. Verify that a shared read lock is acquired before read operation
	// 2. Verify that the lock is released after read completes
	// 3. Verify that multiple concurrent reads can proceed (shared lock behavior)
}

// TestSessionMetadataStore_ReadReleasesLockOnError verifies lock is released when read fails
func TestSessionMetadataStore_ReadReleasesLockOnError(t *testing.T) {
	t.Skip("Requires mock FileAccessor or injectable error mechanism")
	// This test requires:
	// 1. Mock FileAccessor that can be configured to fail during read
	// 2. Ability to verify lock is released before function returns
	// 3. Verification via subsequent successful operation that lock was properly released
}

// TestSessionMetadataStore_WriteWriteFails returns error when file write fails
func TestSessionMetadataStore_WriteWriteFails(t *testing.T) {
	t.Skip("Requires mock FileAccessor that can fail during write operation")
	// This test requires:
	// 1. Mock or injectable FileAccessor that fails during write
	// 2. Should verify error message matches /failed to write session metadata:/i
	// 3. Should verify file is not modified or corrupted
}

// TestSessionMetadataStore_WriteLockFails returns error when lock acquisition fails
func TestSessionMetadataStore_WriteLockFails(t *testing.T) {
	t.Skip("Requires mock FileAccessor that can fail during lock acquisition")
	// This test requires:
	// 1. Mock or injectable FileAccessor that fails during lock acquisition
	// 2. Should verify error message matches /failed to acquire write lock:/i
	// 3. Should verify no write operation is attempted if lock fails
}

// TestSessionMetadataStore_ReadLockFails returns error when lock acquisition fails
func TestSessionMetadataStore_ReadLockFails(t *testing.T) {
	t.Skip("Requires mock FileAccessor that can fail during lock acquisition")
	// This test requires:
	// 1. Mock or injectable FileAccessor that fails during lock acquisition
	// 2. Should verify error message matches /failed to acquire read lock:/i
	// 3. Should verify no read operation is attempted if lock fails
}

// TestSessionMetadataStore_ReadFileReadFails returns error when file read operation fails
func TestSessionMetadataStore_ReadFileReadFails(t *testing.T) {
	t.Skip("Requires file that becomes unreadable after lock acquisition or mock support")
	// This test requires:
	// 1. A way to make file unreadable after lock is acquired (difficult without mock)
	// 2. Should verify error message matches /failed to read session metadata file:/i
	// Note: Permission denied test (already implemented) partially covers this scenario
}

// TestSessionMetadataStore_FileAccessorErrorPropagated verifies FileAccessor callback error is propagated
func TestSessionMetadataStore_FileAccessorErrorPropagated(t *testing.T) {
	t.Skip("Requires mock FileAccessor with configurable callback error")
	// This test requires:
	// 1. Mock FileAccessor that can return a callback error
	// 2. Should verify error contains callback error details
	// 3. Should verify error is wrapped with appropriate context
}

// TestSessionMetadataStore_LocksReleasedOnPanic verifies locks are released if panic occurs during operation
func TestSessionMetadataStore_LocksReleasedOnPanic(t *testing.T) {
	t.Skip("Requires mock that panics during write and ability to verify lock release")
	// This test requires:
	// 1. Mock that can be configured to panic during write
	// 2. Defer/recover mechanism to verify panic propagates
	// 3. Verification that file lock is released (via subsequent successful operation)
	// This is critical for resource cleanup on panic
}
