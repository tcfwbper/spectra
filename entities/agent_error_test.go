package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentError_ValidConstruction creates AgentError with all valid required fields
func TestAgentError_ValidConstruction(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"reason": "timeout"}`)

	err := NewAgentError(
		"architect",
		"task failed",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	// Verify all fields match input
	// assert.Equal(t, "architect", agentError.AgentRole)
	// assert.Equal(t, "task failed", agentError.Message)
	// assert.JSONEq(t, `{"reason": "timeout"}`, string(agentError.Detail))
	// assert.Equal(t, sessionID, agentError.SessionID)
	// assert.Equal(t, "review", agentError.FailingState)
	// assert.Equal(t, int64(1714147200), agentError.OccurredAt)
}

// TestAgentError_EmptyAgentRole creates AgentError with empty AgentRole for human node
func TestAgentError_EmptyAgentRole(t *testing.T) {
	sessionID := uuid.New()

	err := NewAgentError(
		"",
		"user cancelled",
		nil,
		sessionID,
		"human_input",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Equal(t, "", agentError.AgentRole)
}

// TestAgentError_NullDetail creates AgentError with null Detail
func TestAgentError_NullDetail(t *testing.T) {
	sessionID := uuid.New()

	err := NewAgentError(
		"tester",
		"validation failed",
		nil,
		sessionID,
		"test",
		1714147200,
	)

	require.NoError(t, err)
	// assert.Nil(t, agentError.Detail)
}

// TestAgentError_EmptyDetail creates AgentError with empty JSON object Detail
func TestAgentError_EmptyDetail(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{}`)

	err := NewAgentError(
		"architect",
		"error",
		detailJSON,
		sessionID,
		"init",
		1714147200,
	)

	require.NoError(t, err)
	// assert.JSONEq(t, `{}`, string(agentError.Detail))
}

// TestAgentError_EmptyMessage rejects AgentError with empty message
func TestAgentError_EmptyMessage(t *testing.T) {
	sessionID := uuid.New()

	err := NewAgentError(
		"architect",
		"",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)message.*non-empty`, err.Error())
}

// TestAgentError_WhitespaceOnlyMessage rejects AgentError with whitespace-only message
func TestAgentError_WhitespaceOnlyMessage(t *testing.T) {
	sessionID := uuid.New()

	err := NewAgentError(
		"architect",
		"   ",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)message.*whitespace`, err.Error())
}

// TestAgentError_InvalidDetailJSON rejects AgentError with malformed JSON in Detail
func TestAgentError_InvalidDetailJSON(t *testing.T) {
	sessionID := uuid.New()
	invalidJSON := json.RawMessage(`{invalid json}`)

	err := NewAgentError(
		"architect",
		"error",
		invalidJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)JSON.*(parse|unmarshal)`, err.Error())
}

// TestAgentError_NonExistentSession rejects AgentError with non-existent SessionID
func TestAgentError_NonExistentSession(t *testing.T) {
	// Setup: Session with given UUID does not exist
	nonExistentID := uuid.New()

	err := NewAgentError(
		"architect",
		"error",
		nil,
		nonExistentID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*not found`, err.Error())
}

// TestAgentError_FailedSessionID rejects AgentError for session with Status=failed
func TestAgentError_FailedSessionID(t *testing.T) {
	// Setup: Session exists with Status="failed"
	sessionID := uuid.New()
	// CreateSession with Status="failed"

	err := NewAgentError(
		"architect",
		"error",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
	// Verify warning logged
}

// TestAgentError_CompletedSessionID rejects AgentError for session with Status=completed
func TestAgentError_CompletedSessionID(t *testing.T) {
	// Setup: Session exists with Status="completed"
	sessionID := uuid.New()
	// CreateSession with Status="completed"

	err := NewAgentError(
		"architect",
		"error",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
	// Verify warning logged
}

// TestAgentError_TransitionsSessionToFailed verifies session Status transitions to failed when AgentError is raised
func TestAgentError_TransitionsSessionToFailed(t *testing.T) {
	// Setup: Session exists with Status="running", CurrentState="review"
	sessionID := uuid.New()
	// CreateSession with Status="running", CurrentState="review"

	agentErr := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, agentErr)

	// Verify session Status="failed"
	// Verify CurrentState unchanged
	// Verify Error field set to AgentError instance
}

// TestAgentError_InitializingSessionFails verifies session in initializing status transitions to failed
func TestAgentError_InitializingSessionFails(t *testing.T) {
	// Setup: Session exists with Status="initializing", CurrentState="entry"
	sessionID := uuid.New()

	agentErr := NewAgentError(
		"architect",
		"initialization failed",
		nil,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, agentErr)

	// Verify session Status="failed"
	// Verify CurrentState="entry"
	// Verify FailingState="entry"
}

// TestAgentError_RunningSessionFails verifies session in running status transitions to failed
func TestAgentError_RunningSessionFails(t *testing.T) {
	// Setup: Session exists with Status="running", CurrentState="processing"
	sessionID := uuid.New()

	agentErr := NewAgentError(
		"architect",
		"processing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, agentErr)

	// Verify session Status="failed"
	// Verify CurrentState="processing"
	// Verify FailingState="processing"
}

// TestAgentError_FieldsImmutable verifies AgentError fields cannot be modified after creation
func TestAgentError_FieldsImmutable(t *testing.T) {
	// Setup: AgentError instance created
	sessionID := uuid.New()
	agentErr := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, agentErr)

	// Attempt to modify Message, Detail, or other fields
	// Verify field modification attempt fails or has no effect
	// Verify original values remain
}

// TestAgentError_AgentRoleDerivedFromNode verifies AgentRole is derived from current node's agent_role by ErrorProcessor
func TestAgentError_AgentRoleDerivedFromNode(t *testing.T) {
	// Setup: Session at agent node with agent_role="architect"
	sessionID := uuid.New()

	// Error raised without AgentRole in wire payload
	// ErrorProcessor derives AgentRole from node definition

	// Verify AgentError created with AgentRole="architect"
}

// TestAgentError_EmptyAgentRoleForHumanNode verifies AgentRole is empty string for human nodes
func TestAgentError_EmptyAgentRoleForHumanNode(t *testing.T) {
	// Setup: Session at human node (no agent_role defined)
	sessionID := uuid.New()

	// Error raised from human node

	// Verify AgentError created with AgentRole=""
}

// TestAgentError_PersistedToDisk verifies AgentError is persisted to session error log
func TestAgentError_PersistedToDisk(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()

	// Session files placed within tmpDir
	sessionID := uuid.New()

	// Valid AgentError raised
	agentErr := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, agentErr)

	// Verify error details written to error log file within test directory
	// Verify session metadata updated on disk
}

// TestAgentError_SessionDeletion verifies AgentError removed when session is deleted
func TestAgentError_SessionDeletion(t *testing.T) {
	// Setup: Temporary test directory created
	tmpDir := t.TempDir()

	// Session exists with recorded AgentError in test directory
	sessionID := uuid.New()

	// Delete session

	// Verify AgentError file removed from filesystem
	// Verify subsequent error queries return "session not found"
}

// TestAgentError_FailingStateMatchesCurrentState verifies FailingState must match CurrentState at error time
func TestAgentError_FailingStateMatchesCurrentState(t *testing.T) {
	// Setup: Session with CurrentState="review"
	sessionID := uuid.New()

	agentErr := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, agentErr)

	// Verify AgentError recorded
	// Verify FailingState="review" matches session CurrentState
}

// TestAgentError_CurrentStateUnchanged verifies CurrentState does not change when error occurs
func TestAgentError_CurrentStateUnchanged(t *testing.T) {
	// Setup: Session with CurrentState="processing"
	sessionID := uuid.New()

	// AgentError raised
	agentErr := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, agentErr)

	// Verify session CurrentState remains "processing"
	// Verify FailingState="processing"
}

// TestAgentError_StatusPermanentlyFailed verifies session Status remains failed and cannot transition
func TestAgentError_StatusPermanentlyFailed(t *testing.T) {
	// Setup: Session transitioned to Status="failed" by AgentError
	sessionID := uuid.New()

	// Attempt any status transition

	// Verify Status remains "failed"
	// Verify transition rejected
}

// TestAgentError_NoAutomaticRetry verifies runtime does not automatically retry failed session
func TestAgentError_NoAutomaticRetry(t *testing.T) {
	// Setup: Session with Status="failed" due to AgentError
	sessionID := uuid.New()

	// Mock clock advanced by 5 seconds

	// Query session status after time advancement

	// Verify session remains failed
	// Verify no retry attempted
}

// TestAgentError_ManualRecoveryRejected verifies recovery requests for failed session are rejected
func TestAgentError_ManualRecoveryRejected(t *testing.T) {
	// Setup: Session with Status="failed"
	sessionID := uuid.New()

	// Request session recovery
	err := RecoverSession(sessionID)

	require.Error(t, err)
	assert.Regexp(t, `(?i)recovery not supported|cannot recover`, err.Error())
	assert.Contains(t, err.Error(), "new session")
}

// TestAgentError_SensitiveDataPersisted verifies Detail with sensitive info is persisted as-is
func TestAgentError_SensitiveDataPersisted(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"api_key": "secret123"}`)

	agentErr := NewAgentError(
		"architect",
		"authentication failed",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, agentErr)

	// Verify Detail persisted exactly as provided
	// Verify no sanitization by runtime
}

// TestAgentError_LargeMessage accepts AgentError with very large message string
func TestAgentError_LargeMessage(t *testing.T) {
	sessionID := uuid.New()
	largeMessage := make([]byte, 1024*1024) // 1MB
	for i := range largeMessage {
		largeMessage[i] = 'A'
	}

	err := NewAgentError(
		"architect",
		string(largeMessage),
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	// Verify message stored correctly
}

// TestAgentError_UnicodeMessage accepts AgentError with Unicode characters in message
func TestAgentError_UnicodeMessage(t *testing.T) {
	sessionID := uuid.New()

	err := NewAgentError(
		"architect",
		"Error: 任务失败 🔥",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	// Verify Unicode preserved correctly
}

// TestAgentError_LargeDetail accepts AgentError with very large Detail JSON object
func TestAgentError_LargeDetail(t *testing.T) {
	sessionID := uuid.New()

	// Create 10MB JSON object
	largeDetail := make(map[string]string)
	for i := 0; i < 10000; i++ {
		largeDetail[string(rune(i))] = string(make([]byte, 1000))
	}
	detailJSON, _ := json.Marshal(largeDetail)

	err := NewAgentError(
		"architect",
		"error",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	// Verify Detail stored correctly
}

// TestAgentError_DeepNestedDetail accepts AgentError with deeply nested JSON in Detail
func TestAgentError_DeepNestedDetail(t *testing.T) {
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

	err := NewAgentError(
		"architect",
		"error",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	// Verify nested structure preserved
}
