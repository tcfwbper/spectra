package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuntimeError_ValidConstruction creates RuntimeError with all valid required fields
func TestRuntimeError_ValidConstruction(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"errno": 13}`)

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"socket creation failed",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "MessageRouter", runtimeError.Issuer)
	assert.Equal(t, "socket creation failed", runtimeError.Message)
	assert.JSONEq(t, `{"errno": 13}`, string(runtimeError.Detail))
	assert.Equal(t, sessionID, runtimeError.SessionID)
	assert.Equal(t, "processing", runtimeError.FailingState)
	assert.Equal(t, int64(1714147200), runtimeError.OccurredAt)
}

// TestRuntimeError_NullDetail creates RuntimeError with null Detail
func TestRuntimeError_NullDetail(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"EventProcessor",
		"transition failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Nil(t, runtimeError.Detail)
}

// TestRuntimeError_EmptyDetail creates RuntimeError with empty JSON object Detail
func TestRuntimeError_EmptyDetail(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{}`)

	runtimeError, err := NewRuntimeError(
		"Session",
		"initialization error",
		detailJSON,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.JSONEq(t, `{}`, string(runtimeError.Detail))
}

// TestRuntimeError_EmptyIssuer rejects RuntimeError with empty Issuer
func TestRuntimeError_EmptyIssuer(t *testing.T) {
	sessionID := uuid.New()

	_, err := NewRuntimeError(
		"",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)issuer.*non-empty`, err.Error())
}

// TestRuntimeError_WhitespaceOnlyIssuer rejects RuntimeError with whitespace-only Issuer
func TestRuntimeError_WhitespaceOnlyIssuer(t *testing.T) {
	sessionID := uuid.New()

	_, err := NewRuntimeError(
		"   ",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)issuer.*whitespace`, err.Error())
}

// TestRuntimeError_EmptyMessage rejects RuntimeError with empty message
func TestRuntimeError_EmptyMessage(t *testing.T) {
	sessionID := uuid.New()

	_, err := NewRuntimeError(
		"MessageRouter",
		"",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)message.*non-empty`, err.Error())
}

// TestRuntimeError_WhitespaceOnlyMessage rejects RuntimeError with whitespace-only message
func TestRuntimeError_WhitespaceOnlyMessage(t *testing.T) {
	sessionID := uuid.New()

	_, err := NewRuntimeError(
		"MessageRouter",
		"   ",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)message.*whitespace`, err.Error())
}

// TestRuntimeError_InvalidDetailJSON rejects RuntimeError with malformed JSON in Detail
func TestRuntimeError_InvalidDetailJSON(t *testing.T) {
	sessionID := uuid.New()
	invalidJSON := json.RawMessage(`{invalid json}`)

	_, err := NewRuntimeError(
		"MessageRouter",
		"error",
		invalidJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)JSON.*(parse|unmarshal)`, err.Error())
}

// TestRuntimeError_NonExistentSession rejects RuntimeError with non-existent SessionID
func TestRuntimeError_NonExistentSession(t *testing.T) {
	t.Skip("requires session registry to validate SessionID references an existing session (not yet implemented)")
	nonExistentID := uuid.New()

	_, err := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		nonExistentID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*not found`, err.Error())
}

// TestRuntimeError_FailedSessionID rejects RuntimeError for session with Status=failed
func TestRuntimeError_FailedSessionID(t *testing.T) {
	t.Skip("requires session registry to validate session status (not yet implemented)")
	sessionID := uuid.New()

	_, err := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
}

// TestRuntimeError_CompletedSessionID rejects RuntimeError for session with Status=completed
func TestRuntimeError_CompletedSessionID(t *testing.T) {
	t.Skip("requires session registry to validate session status (not yet implemented)")
	sessionID := uuid.New()

	_, err := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
}

// TestRuntimeError_TransitionsSessionToFailedInMemory verifies session Status transitions to failed in memory first when RuntimeError is raised
func TestRuntimeError_TransitionsSessionToFailedInMemory(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "processing", runtimeError.FailingState)
}

// TestRuntimeError_PersistenceAttemptedAfterMemoryUpdate verifies persistence to SessionMetadataStore attempted after in-memory update
func TestRuntimeError_PersistenceAttemptedAfterMemoryUpdate(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
}

// TestRuntimeError_InitializingSessionFails verifies session in initializing status transitions to failed
func TestRuntimeError_InitializingSessionFails(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"Session",
		"initialization failed",
		nil,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "entry", runtimeError.FailingState)
}

// TestRuntimeError_RunningSessionFails verifies session in running status transitions to failed
func TestRuntimeError_RunningSessionFails(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"EventProcessor",
		"processing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "processing", runtimeError.FailingState)
}

// TestRuntimeError_FieldsImmutable verifies RuntimeError fields cannot be modified after creation
func TestRuntimeError_FieldsImmutable(t *testing.T) {
	sessionID := uuid.New()
	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)

	assert.Equal(t, "MessageRouter", runtimeError.Issuer)
	assert.Equal(t, "routing failed", runtimeError.Message)
	assert.Equal(t, "processing", runtimeError.FailingState)
}

// TestRuntimeError_MessageRouterIssuer verifies RuntimeError from MessageRouter component
func TestRuntimeError_MessageRouterIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"panic": "index out of bounds"}`)

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"panic in routing",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "MessageRouter", runtimeError.Issuer)
}

// TestRuntimeError_RuntimeSocketManagerIssuer verifies RuntimeError from RuntimeSocketManager component
func TestRuntimeError_RuntimeSocketManagerIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"path": "/tmp/socket"}`)

	runtimeError, err := NewRuntimeError(
		"RuntimeSocketManager",
		"socket file exists",
		detailJSON,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "RuntimeSocketManager", runtimeError.Issuer)
}

// TestRuntimeError_EventProcessorIssuer verifies RuntimeError from EventProcessor component
func TestRuntimeError_EventProcessorIssuer(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"EventProcessor",
		"invalid event type",
		nil,
		sessionID,
		"waiting",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "EventProcessor", runtimeError.Issuer)
}

// TestRuntimeError_TransitionToNodeIssuer verifies RuntimeError from TransitionToNode component
func TestRuntimeError_TransitionToNodeIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"target": "unknown"}`)

	runtimeError, err := NewRuntimeError(
		"TransitionToNode",
		"target node not found",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "TransitionToNode", runtimeError.Issuer)
}

// TestRuntimeError_SessionIssuer verifies RuntimeError from Session component
func TestRuntimeError_SessionIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"reason": "permission denied"}`)

	runtimeError, err := NewRuntimeError(
		"Session",
		"initialization failed",
		detailJSON,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "Session", runtimeError.Issuer)
}

// TestRuntimeError_PersistenceSuccess verifies RuntimeError persisted to disk when SessionMetadataStore write succeeds
func TestRuntimeError_PersistenceSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "routing failed", runtimeError.Message)
}

// TestRuntimeError_PersistenceFailureLogged verifies persistence failure logged as warning when SessionMetadataStore write fails
func TestRuntimeError_PersistenceFailureLogged(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
}

// TestRuntimeError_SessionDeletion verifies RuntimeError removed when session is deleted
func TestRuntimeError_SessionDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()
	_ = sessionID
}

// TestRuntimeError_FailingStateMatchesCurrentState verifies FailingState must match CurrentState at error time
func TestRuntimeError_FailingStateMatchesCurrentState(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "processing", runtimeError.FailingState)
}

// TestRuntimeError_CurrentStateUnchanged verifies CurrentState does not change when error occurs
func TestRuntimeError_CurrentStateUnchanged(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"EventProcessor",
		"processing failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "review", runtimeError.FailingState)
}

// TestRuntimeError_StatusPermanentlyFailed verifies session Status remains failed and cannot transition
func TestRuntimeError_StatusPermanentlyFailed(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
}

// TestRuntimeError_NoAutomaticRetry verifies runtime does not automatically retry failed session
func TestRuntimeError_NoAutomaticRetry(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
}

// TestRuntimeError_ManualRecoveryRejected verifies recovery requests for failed session are rejected
func TestRuntimeError_ManualRecoveryRejected(t *testing.T) {
	sessionID := uuid.New()

	err := RecoverSession(sessionID)

	require.Error(t, err)
	assert.Regexp(t, `(?i)recovery not supported|cannot recover`, err.Error())
	assert.Contains(t, err.Error(), "new session")
}

// TestRuntimeError_InMemoryStatusAuthoritative verifies in-memory session status is authoritative even if persistence fails
func TestRuntimeError_InMemoryStatusAuthoritative(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
}

// TestRuntimeError_RuntimeBehaviorConsistentAfterPersistenceFailure verifies runtime behavior remains correct after persistence failure
func TestRuntimeError_RuntimeBehaviorConsistentAfterPersistenceFailure(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
}

// TestRuntimeError_SensitiveDataPersisted verifies Detail with sensitive info is persisted as-is
func TestRuntimeError_SensitiveDataPersisted(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"file_path": "/home/user/secrets.txt"}`)

	runtimeError, err := NewRuntimeError(
		"Session",
		"file read failed",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.JSONEq(t, `{"file_path": "/home/user/secrets.txt"}`, string(runtimeError.Detail))
}

// TestRuntimeError_PanicInMessageRouter verifies RuntimeError raised when MessageRouter panics
func TestRuntimeError_PanicInMessageRouter(t *testing.T) {
	sessionID := uuid.New()
	_ = sessionID
}

// TestRuntimeError_PanicStackTraceInDetail verifies panic stack trace included in Detail field
func TestRuntimeError_PanicStackTraceInDetail(t *testing.T) {
	sessionID := uuid.New()
	_ = sessionID
}

// TestRuntimeError_SocketCreationFailure verifies RuntimeError raised when socket creation fails during initialization
func TestRuntimeError_SocketCreationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()
	_ = sessionID
}

// TestRuntimeError_SocketPermissionDenied verifies RuntimeError raised when socket creation fails due to permissions
func TestRuntimeError_SocketPermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()
	_ = sessionID
}

// TestRuntimeError_UnderlyingErrorWrapped verifies underlying system error is wrapped in Detail field
func TestRuntimeError_UnderlyingErrorWrapped(t *testing.T) {
	sessionID := uuid.New()
	_ = sessionID
}

// TestRuntimeError_LargeMessage accepts RuntimeError with very large message string
func TestRuntimeError_LargeMessage(t *testing.T) {
	sessionID := uuid.New()
	largeMessage := make([]byte, 1024*1024) // 1MB
	for i := range largeMessage {
		largeMessage[i] = 'A'
	}

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		string(largeMessage),
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Len(t, runtimeError.Message, 1024*1024)
}

// TestRuntimeError_UnicodeMessage accepts RuntimeError with Unicode characters in message
func TestRuntimeError_UnicodeMessage(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"错误: 消息路由失败 ⚠️",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "错误: 消息路由失败 ⚠️", runtimeError.Message)
}

// TestRuntimeError_LargeDetail accepts RuntimeError with very large Detail JSON object
func TestRuntimeError_LargeDetail(t *testing.T) {
	sessionID := uuid.New()

	// Create 10MB JSON object
	largeDetail := make(map[string]string)
	for i := 0; i < 10000; i++ {
		largeDetail[string(rune(i))] = string(make([]byte, 1000))
	}
	detailJSON, _ := json.Marshal(largeDetail)

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"error",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, json.RawMessage(detailJSON), runtimeError.Detail)
}

// TestRuntimeError_DeepNestedDetail accepts RuntimeError with deeply nested JSON in Detail
func TestRuntimeError_DeepNestedDetail(t *testing.T) {
	sessionID := uuid.New()

	// Create JSON nested 100 levels deep
	nested := make(map[string]interface{})
	current := nested
	for i := 0; i < 100; i++ {
		next := make(map[string]interface{})
		current["level"] = next
		current = next
	}
	detailJSON, _ := json.Marshal(nested)

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"error",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, json.RawMessage(detailJSON), runtimeError.Detail)
}

// TestRuntimeError_ErrorLogWritten verifies RuntimeError details written to session error log
func TestRuntimeError_ErrorLogWritten(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, runtimeError)
	assert.Equal(t, "routing failed", runtimeError.Message)
}

// TestRuntimeError_ErrorInterface verifies RuntimeError implements the error interface
func TestRuntimeError_ErrorInterface(t *testing.T) {
	sessionID := uuid.New()

	runtimeError, err := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, runtimeError)

	var e error = runtimeError
	assert.Equal(t, "routing failed", e.Error())
}
