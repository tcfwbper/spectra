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
	t.Skip("scaffolded: NewSessionMetadataStore not yet implemented in storage/session_metadata_store.go")

	// store := NewSessionMetadataStore("/tmp/project", testSessionUUID)
	// require.NotNil(t, store)
}

func TestNewSessionMetadataStore_NoFileSystemAccess(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore not yet implemented in storage/session_metadata_store.go")

	// Provide a non-existent projectRoot — constructor must not touch filesystem.
	// store := NewSessionMetadataStore("/nonexistent", testSessionUUID)
	// require.NotNil(t, store)
}

// --- Happy Path — Write ---

func TestSessionMetadataStore_Write_ValidMetadata(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// err := store.Write(meta)
	// require.NoError(t, err)
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, readErr := os.ReadFile(filePath)
	// require.NoError(t, readErr)
	// assert.True(t, json.Valid(data))
}

func TestSessionMetadataStore_Write_PrettyPrintedJSON(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// content := string(data)
	// assert.Contains(t, content, "\n")
	// assert.Contains(t, content, "  ") // 2-space indentation
}

func TestSessionMetadataStore_Write_OmitsErrorWhenNil(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata() // Error is nil
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// var parsed map[string]any
	// json.Unmarshal(data, &parsed)
	// _, hasError := parsed["error"]
	// assert.False(t, hasError)
}

func TestSessionMetadataStore_Write_WithAgentError(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithAgentError(t, "Reviewer")
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// var parsed map[string]any
	// json.Unmarshal(data, &parsed)
	// errObj := parsed["error"].(map[string]any)
	// assert.Equal(t, "Reviewer", errObj["agentRole"])
	// _, hasErrorType := errObj["errorType"]
	// assert.False(t, hasErrorType)
}

func TestSessionMetadataStore_Write_WithRuntimeError(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithRuntimeError(t, "system")
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// var parsed map[string]any
	// json.Unmarshal(data, &parsed)
	// errObj := parsed["error"].(map[string]any)
	// assert.Equal(t, "system", errObj["issuer"])
	// _, hasErrorType := errObj["errorType"]
	// assert.False(t, hasErrorType)
}

func TestSessionMetadataStore_Write_AgentErrorEmptyRole(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithAgentError(t, "") // empty agentRole
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// var parsed map[string]any
	// json.Unmarshal(data, &parsed)
	// errObj := parsed["error"].(map[string]any)
	// assert.Equal(t, "", errObj["agentRole"])
}

func TestSessionMetadataStore_Write_UpdatedAtPassThrough(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	meta.UpdatedAt = 1700000000
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// assert.Contains(t, string(data), `"updatedAt": 1700000000`)
}

func TestSessionMetadataStore_Write_ExcludesEventHistory(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// var parsed map[string]any
	// json.Unmarshal(data, &parsed)
	// _, hasEventHistory := parsed["eventHistory"]
	// assert.False(t, hasEventHistory)
}

func TestSessionMetadataStore_Write_TruncatesExistingFile(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	// Write old content to session.json first.
	writeSessionFile(t, sessionDir, `{"id":"old-data","status":"initializing"}`)
	meta := makeValidMetadata()
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// assert.NotContains(t, string(data), "old-data")
}

// --- Happy Path — Read ---

func TestSessionMetadataStore_Read_ValidFile(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeValidSessionJSON())
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// meta, err := store.Read()
	// require.NoError(t, err)
	// assert.Equal(t, testSessionUUID, meta.ID)
	// assert.Equal(t, "CodeReview", meta.WorkflowName)
	// assert.Equal(t, "running", meta.Status)
}

func TestSessionMetadataStore_Read_WithAgentError(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeSessionJSONWithAgentError())
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// meta, err := store.Read()
	// require.NoError(t, err)
	// ae, ok := meta.Error.(*entities.AgentError)
	// require.True(t, ok)
	// assert.Equal(t, "Reviewer", ae.AgentRole())
}

func TestSessionMetadataStore_Read_WithRuntimeError(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeSessionJSONWithRuntimeError())
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// meta, err := store.Read()
	// require.NoError(t, err)
	// re, ok := meta.Error.(*entities.RuntimeError)
	// require.True(t, ok)
	// assert.Equal(t, "system", re.Issuer())
}

func TestSessionMetadataStore_Read_IgnoresEventHistoryField(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

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
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// meta, err := store.Read()
	// require.NoError(t, err)
	// EventHistory is not part of SessionMetadata struct, so it should be ignored.
}

// --- Error Propagation ---

func TestSessionMetadataStore_Write_SessionDirNotExists(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot := makeTempDirWithSessions(t)
	meta := makeValidMetadata()
	_ = projectRoot
	_ = meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// err := store.Write(meta)
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "session directory does not exist:")
}

func TestSessionMetadataStore_Write_FileAccessorError(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	meta := makeValidMetadata()
	_ = meta

	// Stub FileAccessor to return an error from the preparation callback.
	// err should contain "failed to prepare file"
}

func TestSessionMetadataStore_Write_ExceedsMaxPayloadSize(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore, Write, and MaxPayloadSize not yet implemented in storage/session_metadata_store.go")

	projectRoot, _ := makeSessionDirFixture(t)
	_ = projectRoot

	// Construct metadata with very large SessionData (> 10 MB).
	// err should contain "session metadata size exceeds limit:" and "bytes (max"
}

func TestSessionMetadataStore_Write_ExceedsMaxPayloadSize_NoWrite(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore, Write, and MaxPayloadSize not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	oldContent := `{"id":"old","status":"running"}`
	writeSessionFile(t, sessionDir, oldContent)
	_ = projectRoot

	// Construct oversized metadata. Attempt write.
	// File content should remain the old content.
}

func TestSessionMetadataStore_Write_UnserializableSessionData(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, _ := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	meta.SessionData = map[string]any{"ch": make(chan int)}
	_ = projectRoot
	_ = meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// err := store.Write(meta)
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to serialize session metadata:")
	// assert.Contains(t, err.Error(), "unsupported type")
}

func TestSessionMetadataStore_Read_FileNotExists(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, _ := makeSessionDirFixture(t)
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// _, err := store.Read()
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "session metadata file does not exist:")
}

func TestSessionMetadataStore_Read_InvalidJSON(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, `{"id":"550e8400-e29b-41d4-a716-446655440000"`)
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// _, err := store.Read()
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to parse session metadata:")
}

func TestSessionMetadataStore_Read_MissingRequiredField(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

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
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// _, err := store.Read()
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to parse session metadata:")
	// assert.Contains(t, err.Error(), "missing required field")
}

func TestSessionMetadataStore_Read_ErrorBothFields(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

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
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// _, err := store.Read()
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to reconstruct error: ambiguous error object contains both 'agentRole' and 'issuer'")
}

func TestSessionMetadataStore_Read_ErrorNeitherField(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

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
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// _, err := store.Read()
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to reconstruct error: cannot determine error type")
}

func TestSessionMetadataStore_Read_ErrorInvalidConstructorFields(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

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
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// _, err := store.Read()
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to reconstruct error:")
}

// --- Mock / Dependency Interaction ---

func TestSessionMetadataStore_Write_CallsFileAccessor(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	_, _, _ = projectRoot, sessionDir, meta

	// Verify FileAccessor is called exactly once with the session.json path.
}

func TestSessionMetadataStore_Write_ReadsErrorViaGetters(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeMetadataWithAgentError(t, "Reviewer")
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	//
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data, _ := os.ReadFile(filePath)
	// var parsed map[string]any
	// json.Unmarshal(data, &parsed)
	// errObj := parsed["error"].(map[string]any)
	// ae := meta.Error.(*entities.AgentError)
	// assert.Equal(t, ae.AgentRole(), errObj["agentRole"])
	// assert.Equal(t, ae.Message(), errObj["message"])
}

func TestSessionMetadataStore_Read_ReconstructsViaConstructor(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeSessionJSONWithAgentError())
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// meta, err := store.Read()
	// require.NoError(t, err)
	// ae, ok := meta.Error.(*entities.AgentError)
	// require.True(t, ok)
	// assert.Equal(t, "Reviewer", ae.AgentRole())
	// assert.Equal(t, "something went wrong", ae.Message())
}

// --- Idempotency ---

func TestSessionMetadataStore_Read_IdempotentReads(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Read not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	writeSessionFile(t, sessionDir, makeValidSessionJSON())
	_ = projectRoot

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// meta1, err1 := store.Read()
	// require.NoError(t, err1)
	// meta2, err2 := store.Read()
	// require.NoError(t, err2)
	// assert.Equal(t, meta1, meta2)
}

func TestSessionMetadataStore_Write_IdempotentWrites(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore and Write not yet implemented in storage/session_metadata_store.go")

	projectRoot, sessionDir := makeSessionDirFixture(t)
	meta := makeValidMetadata()
	_, _, _ = projectRoot, sessionDir, meta

	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// require.NoError(t, store.Write(meta))
	// filePath := filepath.Join(sessionDir, SessionMetadataFile)
	// data1, _ := os.ReadFile(filePath)
	//
	// require.NoError(t, store.Write(meta))
	// data2, _ := os.ReadFile(filePath)
	//
	// assert.Equal(t, string(data1), string(data2))
}

// --- Boundary Values — MaxPayloadSize ---

func TestSessionMetadataStore_Write_ExactlyAtMaxPayloadSize(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore, Write, and MaxPayloadSize not yet implemented in storage/session_metadata_store.go")

	projectRoot, _ := makeSessionDirFixture(t)
	_ = projectRoot

	// Construct metadata whose pretty-printed JSON is exactly MaxPayloadSize bytes.
	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// err := store.Write(meta)
	// require.NoError(t, err)
}

func TestSessionMetadataStore_Write_OneByteOverMaxPayloadSize(t *testing.T) {
	t.Skip("scaffolded: NewSessionMetadataStore, Write, and MaxPayloadSize not yet implemented in storage/session_metadata_store.go")

	projectRoot, _ := makeSessionDirFixture(t)
	_ = projectRoot

	// Construct metadata whose pretty-printed JSON is MaxPayloadSize + 1 bytes.
	// store := NewSessionMetadataStore(projectRoot, testSessionUUID)
	// err := store.Write(meta)
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "session metadata size exceeds limit:")
}

// Ensure imports are used (prevent compile errors from unused imports).
var (
	_ = assert.Equal
	_ = require.NoError
	_ = json.Marshal
	_ = os.ReadFile
	_ = filepath.Join
	_ = strings.Contains
	_ entities.AgentError
	_ session.SessionMetadata
)
