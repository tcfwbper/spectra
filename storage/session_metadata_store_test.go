package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectra-ai/spectra/entities"
	"github.com/spectra-ai/spectra/entities/session"
)

// --- Fixture Builders for SessionMetadataStore tests ---

// makeValidMetadata creates a valid SessionMetadata snapshot with all fields populated and Error=nil.
func makeValidMetadata() session.SessionMetadata {
	return session.SessionMetadata{
		ID:           testSessionUUID,
		WorkflowName: testWorkflowName,
		Status:       "running",
		CreatedAt:    1700000000,
		UpdatedAt:    1700000100,
		CurrentState: "ReviewCode",
		SessionData:  map[string]any{"key": "value"},
		Error:        nil,
	}
}

// makeMetadataWithAgentError creates a SessionMetadata with an AgentError.
func makeMetadataWithAgentError(t *testing.T, agentRole string) session.SessionMetadata {
	t.Helper()
	ae, err := entities.NewAgentError(
		agentRole,
		"something went wrong",
		json.RawMessage(`{"context":"test"}`),
		1700000200,
		testSessionUUID,
		"ReviewCode",
	)
	require.NoError(t, err)
	meta := makeValidMetadata()
	meta.Error = ae
	return meta
}

// makeMetadataWithRuntimeError creates a SessionMetadata with a RuntimeError.
func makeMetadataWithRuntimeError(t *testing.T, issuer string) session.SessionMetadata {
	t.Helper()
	re, err := entities.NewRuntimeError(
		issuer,
		"runtime failure",
		json.RawMessage(`{"detail":"info"}`),
		1700000300,
		testSessionUUID,
		"ReviewCode",
	)
	require.NoError(t, err)
	meta := makeValidMetadata()
	meta.Error = re
	return meta
}

// writeSessionFile writes content to session.json in the session directory.
func writeSessionFile(t *testing.T, sessionDir string, content string) string {
	t.Helper()
	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}

// makeValidSessionJSON returns a valid pretty-printed session.json content string.
func makeValidSessionJSON() string {
	return `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "running",
  "createdAt": 1700000000,
  "updatedAt": 1700000100,
  "currentState": "ReviewCode",
  "sessionData": {
    "key": "value"
  }
}`
}

// makeSessionJSONWithAgentError returns session.json with an AgentError object.
func makeSessionJSONWithAgentError() string {
	return `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "failed",
  "createdAt": 1700000000,
  "updatedAt": 1700000200,
  "currentState": "ReviewCode",
  "sessionData": {},
  "error": {
    "agentRole": "Reviewer",
    "message": "something went wrong",
    "detail": {"context": "test"},
    "occurredAt": 1700000200,
    "sessionID": "550e8400-e29b-41d4-a716-446655440000",
    "failingState": "ReviewCode"
  }
}`
}

// makeSessionJSONWithRuntimeError returns session.json with a RuntimeError object.
func makeSessionJSONWithRuntimeError() string {
	return `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "failed",
  "createdAt": 1700000000,
  "updatedAt": 1700000300,
  "currentState": "ReviewCode",
  "sessionData": {},
  "error": {
    "issuer": "system",
    "message": "runtime failure",
    "detail": {"detail": "info"},
    "occurredAt": 1700000300,
    "sessionID": "550e8400-e29b-41d4-a716-446655440000",
    "failingState": "ReviewCode"
  }
}`
}

// --- Happy Path — Construction ---

func TestNewSessionMetadataStore_ValidInputs(t *testing.T) {
	store := NewSessionMetadataStore("/tmp/project", testSessionUUID)
	require.NotNil(t, store)
}

func TestNewSessionMetadataStore_NoFileSystemAccess(t *testing.T) {
	// Provide a non-existent projectRoot — constructor must not touch filesystem.
	store := NewSessionMetadataStore("/nonexistent", testSessionUUID)
	require.NotNil(t, store)
}

// --- Happy Path — Write ---

func TestSessionMetadataStore_Write_ValidMetadata(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.NoError(t, err)

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, readErr := os.ReadFile(filePath)
	require.NoError(t, readErr)
	assert.True(t, json.Valid(data))
}

func TestSessionMetadataStore_Write_PrettyPrintedJSON(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	content := string(data)
	assert.Contains(t, content, "\n")
	assert.Contains(t, content, "  ") // 2-space indentation
}

func TestSessionMetadataStore_Write_OmitsErrorWhenNil(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata() // Error is nil

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	_, hasError := parsed["error"]
	assert.False(t, hasError)
}

func TestSessionMetadataStore_Write_WithAgentError(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithAgentError(t, "Reviewer")

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	errObj := parsed["error"].(map[string]any)
	assert.Equal(t, "Reviewer", errObj["agentRole"])
	_, hasErrorType := errObj["errorType"]
	assert.False(t, hasErrorType)
}

func TestSessionMetadataStore_Write_WithRuntimeError(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithRuntimeError(t, "system")

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	errObj := parsed["error"].(map[string]any)
	assert.Equal(t, "system", errObj["issuer"])
	_, hasErrorType := errObj["errorType"]
	assert.False(t, hasErrorType)
}

func TestSessionMetadataStore_Write_AgentErrorEmptyRole(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithAgentError(t, "") // empty agentRole

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	errObj := parsed["error"].(map[string]any)
	assert.Equal(t, "", errObj["agentRole"])
}

func TestSessionMetadataStore_Write_UpdatedAtPassThrough(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	meta.UpdatedAt = 1700000000

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	assert.Contains(t, string(data), `"updatedAt": 1700000000`)
}

func TestSessionMetadataStore_Write_ExcludesEventHistory(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	_, hasEventHistory := parsed["eventHistory"]
	assert.False(t, hasEventHistory)
}

func TestSessionMetadataStore_Write_TruncatesExistingFile(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	// Write old content to session.json first.
	writeSessionFile(t, sessionDir, `{"id":"old-data","status":"initializing"}`)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	assert.NotContains(t, string(data), "old-data")
}

// --- Happy Path — Read ---

func TestSessionMetadataStore_Read_ValidFile(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeValidSessionJSON())

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	meta, err := store.Read()
	require.NoError(t, err)
	assert.Equal(t, testSessionUUID, meta.ID)
	assert.Equal(t, "CodeReview", meta.WorkflowName)
	assert.Equal(t, "running", meta.Status)
}

func TestSessionMetadataStore_Read_WithAgentError(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeSessionJSONWithAgentError())

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	meta, err := store.Read()
	require.NoError(t, err)
	ae, ok := meta.Error.(*entities.AgentError)
	require.True(t, ok)
	assert.Equal(t, "Reviewer", ae.AgentRole())
}

func TestSessionMetadataStore_Read_WithRuntimeError(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeSessionJSONWithRuntimeError())

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	meta, err := store.Read()
	require.NoError(t, err)
	re, ok := meta.Error.(*entities.RuntimeError)
	require.True(t, ok)
	assert.Equal(t, "system", re.Issuer())
}

func TestSessionMetadataStore_Read_IgnoresEventHistoryField(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	jsonWithHistory := `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "running",
  "createdAt": 1700000000,
  "updatedAt": 1700000100,
  "currentState": "ReviewCode",
  "sessionData": {},
  "eventHistory": [{"fake": "event"}]
}`
	writeSessionFile(t, sessionDir, jsonWithHistory)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	meta, err := store.Read()
	require.NoError(t, err)
	// EventHistory is not part of SessionMetadata struct, so it should be ignored.
	assert.Equal(t, testSessionUUID, meta.ID)
}

// --- Error Propagation ---

func TestSessionMetadataStore_Write_SessionDirNotExists(t *testing.T) {
	projectRoot := makeTempDirWithSessions(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session directory does not exist:")
}

func TestSessionMetadataStore_Write_FileAccessorError(t *testing.T) {
	// Use a projectRoot that doesn't have session dir — FileAccessor callback returns error.
	projectRoot := makeTempDirWithSessions(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session directory does not exist:")
}

func TestSessionMetadataStore_Write_ExceedsMaxPayloadSize(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)

	// Construct metadata with very large SessionData (> 10 MB).
	meta := makeValidMetadata()
	meta.SessionData = map[string]any{"large": strings.Repeat("x", MaxPayloadSize+1)}

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session metadata size exceeds limit:")
	assert.Contains(t, err.Error(), "bytes (max")
}

func TestSessionMetadataStore_Write_ExceedsMaxPayloadSize_NoWrite(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	oldContent := `{"id":"old","status":"running"}`
	writeSessionFile(t, sessionDir, oldContent)

	// Construct oversized metadata. Attempt write.
	meta := makeValidMetadata()
	meta.SessionData = map[string]any{"large": strings.Repeat("x", MaxPayloadSize+1)}

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.Error(t, err)

	// File content should remain the old content.
	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	assert.Equal(t, oldContent, string(data))
}

func TestSessionMetadataStore_Write_UnserializableSessionData(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	meta.SessionData = map[string]any{"ch": make(chan int)}

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to serialize session metadata:")
	assert.Contains(t, err.Error(), "unsupported type")
}

func TestSessionMetadataStore_Read_FileNotExists(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	_, err := store.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session metadata file does not exist:")
}

func TestSessionMetadataStore_Read_InvalidJSON(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, `{"id":"550e8400-e29b-41d4-a716-446655440000"`)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	_, err := store.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse session metadata:")
}

func TestSessionMetadataStore_Read_MissingRequiredField(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	// Missing "id" field.
	jsonMissingID := `{
  "workflowName": "CodeReview",
  "status": "running",
  "createdAt": 1700000000,
  "updatedAt": 1700000100,
  "currentState": "ReviewCode",
  "sessionData": {}
}`
	writeSessionFile(t, sessionDir, jsonMissingID)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	_, err := store.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse session metadata:")
	assert.Contains(t, err.Error(), "missing required field")
}

func TestSessionMetadataStore_Read_ErrorBothFields(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	// Error object with both agentRole and issuer — ambiguous.
	jsonBoth := `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "failed",
  "createdAt": 1700000000,
  "updatedAt": 1700000200,
  "currentState": "ReviewCode",
  "sessionData": {},
  "error": {
    "agentRole": "Reviewer",
    "issuer": "system",
    "message": "conflict",
    "occurredAt": 1700000200,
    "sessionID": "550e8400-e29b-41d4-a716-446655440000",
    "failingState": "ReviewCode"
  }
}`
	writeSessionFile(t, sessionDir, jsonBoth)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	_, err := store.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reconstruct error: ambiguous error object contains both 'agentRole' and 'issuer'")
}

func TestSessionMetadataStore_Read_ErrorNeitherField(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	// Error object with neither agentRole nor issuer.
	jsonNeither := `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "failed",
  "createdAt": 1700000000,
  "updatedAt": 1700000200,
  "currentState": "ReviewCode",
  "sessionData": {},
  "error": {
    "message": "unknown error",
    "occurredAt": 1700000200,
    "sessionID": "550e8400-e29b-41d4-a716-446655440000",
    "failingState": "ReviewCode"
  }
}`
	writeSessionFile(t, sessionDir, jsonNeither)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	_, err := store.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reconstruct error: cannot determine error type")
}

func TestSessionMetadataStore_Read_ErrorInvalidConstructorFields(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	// AgentError with invalid fields (empty message which fails NewAgentError).
	jsonInvalidErr := `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflowName": "CodeReview",
  "status": "failed",
  "createdAt": 1700000000,
  "updatedAt": 1700000200,
  "currentState": "ReviewCode",
  "sessionData": {},
  "error": {
    "agentRole": "Reviewer",
    "message": "",
    "occurredAt": 1700000200,
    "sessionID": "550e8400-e29b-41d4-a716-446655440000",
    "failingState": "ReviewCode"
  }
}`
	writeSessionFile(t, sessionDir, jsonInvalidErr)

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	_, err := store.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reconstruct error:")
}

// --- Mock / Dependency Interaction ---

func TestSessionMetadataStore_Write_CallsFileAccessor(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	// Verify FileAccessor was called — the file should exist.
	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	_, err := os.Stat(filePath)
	assert.NoError(t, err)
}

func TestSessionMetadataStore_Write_ReadsErrorViaGetters(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithAgentError(t, "Reviewer")

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))

	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	errObj := parsed["error"].(map[string]any)
	ae := meta.Error.(*entities.AgentError)
	assert.Equal(t, ae.AgentRole(), errObj["agentRole"])
	assert.Equal(t, ae.Message(), errObj["message"])
}

func TestSessionMetadataStore_Read_ReconstructsViaConstructor(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeSessionJSONWithAgentError())

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	meta, err := store.Read()
	require.NoError(t, err)
	ae, ok := meta.Error.(*entities.AgentError)
	require.True(t, ok)
	assert.Equal(t, "Reviewer", ae.AgentRole())
	assert.Equal(t, "something went wrong", ae.Message())
}

// --- Idempotency ---

func TestSessionMetadataStore_Read_IdempotentReads(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeValidSessionJSON())

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	meta1, err1 := store.Read()
	require.NoError(t, err1)
	meta2, err2 := store.Read()
	require.NoError(t, err2)
	assert.Equal(t, meta1, meta2)
}

func TestSessionMetadataStore_Write_IdempotentWrites(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	require.NoError(t, store.Write(meta))
	filePath := filepath.Join(sessionDir, SessionMetadataFile)
	data1, _ := os.ReadFile(filePath)

	require.NoError(t, store.Write(meta))
	data2, _ := os.ReadFile(filePath)

	assert.Equal(t, string(data1), string(data2))
}

// --- Boundary Values — MaxPayloadSize ---

func TestSessionMetadataStore_Write_ExactlyAtMaxPayloadSize(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)

	// Construct metadata, serialize to measure overhead, then build one with exact size.
	meta := makeValidMetadata()
	meta.SessionData = map[string]any{}
	store := NewSessionMetadataStore(projectRoot, testSessionUUID)

	// Serialize the baseline to measure the overhead.
	baseData, err := json.MarshalIndent(map[string]any{
		"id": meta.ID, "workflowName": meta.WorkflowName,
		"status": meta.Status, "createdAt": meta.CreatedAt,
		"updatedAt": meta.UpdatedAt, "currentState": meta.CurrentState,
		"sessionData": map[string]any{"pad": ""},
	}, "", "  ")
	require.NoError(t, err)

	// Calculate how much padding is needed in the "pad" field to reach MaxPayloadSize.
	// The "pad" field value is a string, which adds 2 bytes for quotes in compact form.
	// In pretty-printed JSON, we need to account for the structure.
	overhead := len(baseData) // This includes `"pad": ""`
	padLen := MaxPayloadSize - overhead
	if padLen < 0 {
		t.Skip("overhead alone exceeds MaxPayloadSize")
	}
	meta.SessionData = map[string]any{"pad": strings.Repeat("a", padLen)}

	err = store.Write(meta)
	require.NoError(t, err)
}

func TestSessionMetadataStore_Write_OneByteOverMaxPayloadSize(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)

	// Construct metadata with data that's definitely over MaxPayloadSize.
	meta := makeValidMetadata()
	meta.SessionData = map[string]any{"large": strings.Repeat("x", MaxPayloadSize)}

	store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	err := store.Write(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session metadata size exceeds limit:")
}

