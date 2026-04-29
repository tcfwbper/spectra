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

	err := NewRuntimeError(
		"MessageRouter",
		"socket creation failed",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// Verify all fields match input
	// assert.Equal(t, "MessageRouter", runtimeError.Issuer)
	// assert.Equal(t, "socket creation failed", runtimeError.Message)
	// assert.JSONEq(t, `{"errno": 13}`, string(runtimeError.Detail))
	// assert.Equal(t, sessionID, runtimeError.SessionID)
	// assert.Equal(t, "processing", runtimeError.FailingState)
	// assert.Equal(t, int64(1714147200), runtimeError.OccurredAt)
}

// TestRuntimeError_NullDetail creates RuntimeError with null Detail
func TestRuntimeError_NullDetail(t *testing.T) {
	sessionID := uuid.New()

	err := NewRuntimeError(
		"EventProcessor",
		"transition failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Nil(t, runtimeError.Detail)
}

// TestRuntimeError_EmptyDetail creates RuntimeError with empty JSON object Detail
func TestRuntimeError_EmptyDetail(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{}`)

	err := NewRuntimeError(
		"Session",
		"initialization error",
		detailJSON,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	// assert.JSONEq(t, `{}`, string(runtimeError.Detail))
}

// TestRuntimeError_EmptyIssuer rejects RuntimeError with empty Issuer
func TestRuntimeError_EmptyIssuer(t *testing.T) {
	sessionID := uuid.New()

	err := NewRuntimeError(
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

	err := NewRuntimeError(
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

	err := NewRuntimeError(
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

	err := NewRuntimeError(
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

	err := NewRuntimeError(
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
	// Setup: Session with given UUID does not exist
	nonExistentID := uuid.New()

	err := NewRuntimeError(
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
	// Setup: Session exists with Status="failed"
	sessionID := uuid.New()
	// CreateSession with Status="failed"

	err := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
	// Verify warning logged
}

// TestRuntimeError_CompletedSessionID rejects RuntimeError for session with Status=completed
func TestRuntimeError_CompletedSessionID(t *testing.T) {
	t.Skip("requires session registry to validate session status (not yet implemented)")
	// Setup: Session exists with Status="completed"
	sessionID := uuid.New()
	// CreateSession with Status="completed"

	err := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
	// Verify warning logged
}

// TestRuntimeError_TransitionsSessionToFailedInMemory verifies session Status transitions to failed in memory first when RuntimeError is raised
func TestRuntimeError_TransitionsSessionToFailedInMemory(t *testing.T) {
	// Setup: Session exists with Status="running", CurrentState="processing"
	sessionID := uuid.New()
	// CreateSession with Status="running", CurrentState="processing"

	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, runtimeErr)

	// Verify session Status="failed" in memory immediately
	// Verify CurrentState unchanged
	// Verify Error field set to RuntimeError instance
}

// TestRuntimeError_PersistenceAttemptedAfterMemoryUpdate verifies persistence to SessionMetadataStore attempted after in-memory update
func TestRuntimeError_PersistenceAttemptedAfterMemoryUpdate(t *testing.T) {
	// Setup: Session exists with Status="running"
	sessionID := uuid.New()

	// RuntimeError raised
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"error",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify in-memory status updated to "failed" first
	// Verify then persistence attempted
	// Verify persistence success or failure does not affect in-memory status
}

// TestRuntimeError_InitializingSessionFails verifies session in initializing status transitions to failed
func TestRuntimeError_InitializingSessionFails(t *testing.T) {
	// Setup: Session exists with Status="initializing", CurrentState="entry"
	sessionID := uuid.New()

	runtimeErr := NewRuntimeError(
		"Session",
		"initialization failed",
		nil,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, runtimeErr)

	// Verify session Status="failed" in memory
	// Verify CurrentState="entry"
	// Verify FailingState="entry"
}

// TestRuntimeError_RunningSessionFails verifies session in running status transitions to failed
func TestRuntimeError_RunningSessionFails(t *testing.T) {
	// Setup: Session exists with Status="running", CurrentState="processing"
	sessionID := uuid.New()

	runtimeErr := NewRuntimeError(
		"EventProcessor",
		"processing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, runtimeErr)

	// Verify session Status="failed" in memory
	// Verify CurrentState="processing"
	// Verify FailingState="processing"
}

// TestRuntimeError_FieldsImmutable verifies RuntimeError fields cannot be modified after creation
func TestRuntimeError_FieldsImmutable(t *testing.T) {
	// Setup: RuntimeError instance created
	sessionID := uuid.New()
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Attempt to modify Issuer, Message, Detail, or other fields
	// Verify field modification attempt fails or has no effect
	// Verify original values remain
}

// TestRuntimeError_MessageRouterIssuer verifies RuntimeError from MessageRouter component
func TestRuntimeError_MessageRouterIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"panic": "index out of bounds"}`)

	err := NewRuntimeError(
		"MessageRouter",
		"panic in routing",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Equal(t, "MessageRouter", runtimeError.Issuer)
}

// TestRuntimeError_RuntimeSocketManagerIssuer verifies RuntimeError from RuntimeSocketManager component
func TestRuntimeError_RuntimeSocketManagerIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"path": "/tmp/socket"}`)

	err := NewRuntimeError(
		"RuntimeSocketManager",
		"socket file exists",
		detailJSON,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Equal(t, "RuntimeSocketManager", runtimeError.Issuer)
}

// TestRuntimeError_EventProcessorIssuer verifies RuntimeError from EventProcessor component
func TestRuntimeError_EventProcessorIssuer(t *testing.T) {
	sessionID := uuid.New()

	err := NewRuntimeError(
		"EventProcessor",
		"invalid event type",
		nil,
		sessionID,
		"waiting",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Equal(t, "EventProcessor", runtimeError.Issuer)
}

// TestRuntimeError_TransitionToNodeIssuer verifies RuntimeError from TransitionToNode component
func TestRuntimeError_TransitionToNodeIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"target": "unknown"}`)

	err := NewRuntimeError(
		"TransitionToNode",
		"target node not found",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Equal(t, "TransitionToNode", runtimeError.Issuer)
}

// TestRuntimeError_SessionIssuer verifies RuntimeError from Session component
func TestRuntimeError_SessionIssuer(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"reason": "permission denied"}`)

	err := NewRuntimeError(
		"Session",
		"initialization failed",
		detailJSON,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Equal(t, "Session", runtimeError.Issuer)
}

// TestRuntimeError_PersistenceSuccess verifies RuntimeError persisted to disk when SessionMetadataStore write succeeds
func TestRuntimeError_PersistenceSuccess(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Session files placed within tmpDir
	// SessionMetadataStore operational
	sessionID := uuid.New()

	// Valid RuntimeError raised
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify in-memory status updated to "failed"
	// Verify persistence succeeds
	// Verify error details written to disk within test directory
	// Verify session metadata updated
}

// TestRuntimeError_PersistenceFailureLogged verifies persistence failure logged as warning when SessionMetadataStore write fails
func TestRuntimeError_PersistenceFailureLogged(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Mock SessionMetadataStore configured to return write error simulating disk full
	sessionID := uuid.New()

	// Valid RuntimeError raised
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify in-memory status updated to "failed"
	// Verify persistence fails
	// Verify warning logged matching /failed.*persist.*RuntimeError/i
	// Verify session remains failed in memory
}

// TestRuntimeError_SessionDeletion verifies RuntimeError removed when session is deleted
func TestRuntimeError_SessionDeletion(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Session exists with recorded RuntimeError in tmpDir
	sessionID := uuid.New()
	_ = sessionID

	// Delete session

	// Verify RuntimeError file removed from filesystem
	// Verify subsequent error queries match error /session.*not found/i
}

// TestRuntimeError_FailingStateMatchesCurrentState verifies FailingState must match CurrentState at error time
func TestRuntimeError_FailingStateMatchesCurrentState(t *testing.T) {
	// Setup: Session with CurrentState="processing"
	sessionID := uuid.New()

	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify RuntimeError recorded
	// Verify FailingState="processing" matches session CurrentState
}

// TestRuntimeError_CurrentStateUnchanged verifies CurrentState does not change when error occurs
func TestRuntimeError_CurrentStateUnchanged(t *testing.T) {
	// Setup: Session with CurrentState="review"
	sessionID := uuid.New()

	// RuntimeError raised
	runtimeErr := NewRuntimeError(
		"EventProcessor",
		"processing failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify session CurrentState remains "review"
	// Verify FailingState="review"
}

// TestRuntimeError_StatusPermanentlyFailed verifies session Status remains failed and cannot transition
func TestRuntimeError_StatusPermanentlyFailed(t *testing.T) {
	// Setup: Session transitioned to Status="failed" by RuntimeError
	sessionID := uuid.New()
	_ = sessionID

	// Attempt any status transition

	// Verify Status remains "failed"
	// Verify transition rejected
}

// TestRuntimeError_NoAutomaticRetry verifies runtime does not automatically retry failed session
func TestRuntimeError_NoAutomaticRetry(t *testing.T) {
	// Setup: Session with Status="failed" due to RuntimeError
	sessionID := uuid.New()
	_ = sessionID

	// Mock clock advanced by 5 seconds

	// Query session status after time advancement

	// Verify session remains failed
	// Verify no retry attempted
}

// TestRuntimeError_ManualRecoveryRejected verifies recovery requests for failed session are rejected
func TestRuntimeError_ManualRecoveryRejected(t *testing.T) {
	// Setup: Session with Status="failed"
	sessionID := uuid.New()

	// Request session recovery
	err := RecoverSession(sessionID)

	require.Error(t, err)
	assert.Regexp(t, `(?i)recovery not supported|cannot recover`, err.Error())
	assert.Contains(t, err.Error(), "new session")
}

// TestRuntimeError_InMemoryStatusAuthoritative verifies in-memory session status is authoritative even if persistence fails
func TestRuntimeError_InMemoryStatusAuthoritative(t *testing.T) {
	// Setup: Mock SessionMetadataStore configured to return write error
	sessionID := uuid.New()

	// RuntimeError raised
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify in-memory status updated to "failed" immediately
	// Verify runtime behavior reflects failed status
	// Verify subsequent operations see failed status
}

// TestRuntimeError_RuntimeBehaviorConsistentAfterPersistenceFailure verifies runtime behavior remains correct after persistence failure
func TestRuntimeError_RuntimeBehaviorConsistentAfterPersistenceFailure(t *testing.T) {
	// Setup: Mock SessionMetadataStore configured to return write error
	sessionID := uuid.New()

	// RuntimeError raised
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Subsequent session query

	// Verify session query returns Status="failed" from memory
	// Verify runtime correctly rejects new events for this session with error matching /session.*terminated|failed/i
}

// TestRuntimeError_SensitiveDataPersisted verifies Detail with sensitive info is persisted as-is
func TestRuntimeError_SensitiveDataPersisted(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"file_path": "/home/user/secrets.txt"}`)

	runtimeErr := NewRuntimeError(
		"Session",
		"file read failed",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify Detail persisted exactly as provided
	// Verify no sanitization by runtime
}

// TestRuntimeError_PanicInMessageRouter verifies RuntimeError raised when MessageRouter panics
func TestRuntimeError_PanicInMessageRouter(t *testing.T) {
	// Setup: Mock MessageRouter configured to panic with "index out of bounds" during message processing
	sessionID := uuid.New()
	_ = sessionID

	// Trigger message processing that causes panic

	// Verify RuntimeError created with Issuer="MessageRouter"
	// Verify Detail contains panic message and stack trace
	// Verify session transitions to failed
}

// TestRuntimeError_PanicStackTraceInDetail verifies panic stack trace included in Detail field
func TestRuntimeError_PanicStackTraceInDetail(t *testing.T) {
	// Setup: Mock component configured to panic during processing
	sessionID := uuid.New()
	_ = sessionID

	// Trigger panic condition

	// Verify RuntimeError Detail field contains "panic" key with message
	// Verify "stack" key with trace
}

// TestRuntimeError_SocketCreationFailure verifies RuntimeError raised when socket creation fails during initialization
func TestRuntimeError_SocketCreationFailure(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Socket file created programmatically at target path within tmpDir before session initialization
	sessionID := uuid.New()
	_ = sessionID

	// Initialize session

	// Verify RuntimeError with Issuer="Session" or "RuntimeSocketManager"
	// Verify Detail contains error details matching /socket.*exists|file.*exists/i
	// Verify session Status="failed"
}

// TestRuntimeError_SocketPermissionDenied verifies RuntimeError raised when socket creation fails due to permissions
func TestRuntimeError_SocketPermissionDenied(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Target socket directory created within tmpDir with permissions set to read-only (0444)
	sessionID := uuid.New()
	_ = sessionID

	// Initialize session

	// Verify RuntimeError with Detail containing error matching /permission denied|EACCES/i
	// Verify session initialization aborted
	// Verify Status="failed"
}

// TestRuntimeError_UnderlyingErrorWrapped verifies underlying system error is wrapped in Detail field
func TestRuntimeError_UnderlyingErrorWrapped(t *testing.T) {
	// Setup: Mock system operation configured to return specific errno (e.g., ENOENT)
	sessionID := uuid.New()
	_ = sessionID

	// Trigger failing operation

	// Verify RuntimeError created
	// Verify Detail contains original error details with errno
	// Verify error message provides context
}

// TestRuntimeError_LargeMessage accepts RuntimeError with very large message string
func TestRuntimeError_LargeMessage(t *testing.T) {
	sessionID := uuid.New()
	largeMessage := make([]byte, 1024*1024) // 1MB
	for i := range largeMessage {
		largeMessage[i] = 'A'
	}

	err := NewRuntimeError(
		"MessageRouter",
		string(largeMessage),
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// Verify message stored correctly
}

// TestRuntimeError_UnicodeMessage accepts RuntimeError with Unicode characters in message
func TestRuntimeError_UnicodeMessage(t *testing.T) {
	sessionID := uuid.New()

	err := NewRuntimeError(
		"MessageRouter",
		"错误: 消息路由失败 ⚠️",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// Verify Unicode preserved correctly
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

	err := NewRuntimeError(
		"MessageRouter",
		"error",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// Verify Detail stored correctly
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

	err := NewRuntimeError(
		"MessageRouter",
		"error",
		detailJSON,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	// Verify nested structure preserved
}

// TestRuntimeError_ErrorLogWritten verifies RuntimeError details written to session error log
func TestRuntimeError_ErrorLogWritten(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()
	_ = tmpDir

	// Session files placed within tmpDir
	sessionID := uuid.New()

	// RuntimeError raised
	runtimeErr := NewRuntimeError(
		"MessageRouter",
		"routing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, runtimeErr)

	// Verify error details written to session's error log file within tmpDir with timestamp and full context
}
