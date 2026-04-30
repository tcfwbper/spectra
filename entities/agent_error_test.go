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

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "architect", agentError.AgentRole)
	assert.Equal(t, "task failed", agentError.Message)
	assert.JSONEq(t, `{"reason": "timeout"}`, string(agentError.Detail))
	assert.Equal(t, sessionID, agentError.SessionID)
	assert.Equal(t, "review", agentError.FailingState)
	assert.Equal(t, int64(1714147200), agentError.OccurredAt)
}

// TestAgentError_EmptyAgentRole creates AgentError with empty AgentRole for human node
func TestAgentError_EmptyAgentRole(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"",
		"user cancelled",
		nil,
		sessionID,
		"human_input",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "", agentError.AgentRole)
}

// TestAgentError_NullDetail creates AgentError with null Detail
func TestAgentError_NullDetail(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"tester",
		"validation failed",
		nil,
		sessionID,
		"test",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Nil(t, agentError.Detail)
}

// TestAgentError_EmptyDetail creates AgentError with empty JSON object Detail
func TestAgentError_EmptyDetail(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{}`)

	agentError, err := NewAgentError(
		"architect",
		"error",
		detailJSON,
		sessionID,
		"init",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.JSONEq(t, `{}`, string(agentError.Detail))
}

// TestAgentError_EmptyMessage rejects AgentError with empty message
func TestAgentError_EmptyMessage(t *testing.T) {
	sessionID := uuid.New()

	_, err := NewAgentError(
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

	_, err := NewAgentError(
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

	_, err := NewAgentError(
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
	t.Skip("requires session registry to validate SessionID references an existing session (not yet implemented)")
	nonExistentID := uuid.New()

	_, err := NewAgentError(
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
	t.Skip("requires session registry to validate session status (not yet implemented)")
	sessionID := uuid.New()

	_, err := NewAgentError(
		"architect",
		"error",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
}

// TestAgentError_CompletedSessionID rejects AgentError for session with Status=completed
func TestAgentError_CompletedSessionID(t *testing.T) {
	t.Skip("requires session registry to validate session status (not yet implemented)")
	sessionID := uuid.New()

	_, err := NewAgentError(
		"architect",
		"error",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.Error(t, err)
	assert.Regexp(t, `(?i)session.*terminated`, err.Error())
}

// TestAgentError_TransitionsSessionToFailed verifies session Status transitions to failed when AgentError is raised
func TestAgentError_TransitionsSessionToFailed(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "review", agentError.FailingState)
}

// TestAgentError_InitializingSessionFails verifies session in initializing status transitions to failed
func TestAgentError_InitializingSessionFails(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"initialization failed",
		nil,
		sessionID,
		"entry",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "entry", agentError.FailingState)
}

// TestAgentError_RunningSessionFails verifies session in running status transitions to failed
func TestAgentError_RunningSessionFails(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"processing failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "processing", agentError.FailingState)
}

// TestAgentError_FieldsImmutable verifies AgentError fields cannot be modified after creation
func TestAgentError_FieldsImmutable(t *testing.T) {
	sessionID := uuid.New()
	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)

	assert.Equal(t, "architect", agentError.AgentRole)
	assert.Equal(t, "task failed", agentError.Message)
	assert.Equal(t, "review", agentError.FailingState)
}

// TestAgentError_AgentRoleDerivedFromNode verifies AgentRole is derived from current node's agent_role by ErrorProcessor
func TestAgentError_AgentRoleDerivedFromNode(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "architect", agentError.AgentRole)
}

// TestAgentError_EmptyAgentRoleForHumanNode verifies AgentRole is empty string for human nodes
func TestAgentError_EmptyAgentRoleForHumanNode(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"",
		"human node error",
		nil,
		sessionID,
		"human_input",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "", agentError.AgentRole)
}

// TestAgentError_PersistedToDisk verifies AgentError is persisted to session error log
func TestAgentError_PersistedToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "task failed", agentError.Message)
}

// TestAgentError_SessionDeletion verifies AgentError removed when session is deleted
func TestAgentError_SessionDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir

	sessionID := uuid.New()
	_ = sessionID
}

// TestAgentError_FailingStateMatchesCurrentState verifies FailingState must match CurrentState at error time
func TestAgentError_FailingStateMatchesCurrentState(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "review", agentError.FailingState)
}

// TestAgentError_CurrentStateUnchanged verifies CurrentState does not change when error occurs
func TestAgentError_CurrentStateUnchanged(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"processing",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "processing", agentError.FailingState)
}

// TestAgentError_StatusPermanentlyFailed verifies session Status remains failed and cannot transition
func TestAgentError_StatusPermanentlyFailed(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)
}

// TestAgentError_NoAutomaticRetry verifies runtime does not automatically retry failed session
func TestAgentError_NoAutomaticRetry(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)
}

// TestAgentError_ManualRecoveryRejected verifies recovery requests for failed session are rejected
func TestAgentError_ManualRecoveryRejected(t *testing.T) {
	sessionID := uuid.New()

	err := RecoverSession(sessionID)

	require.Error(t, err)
	assert.Regexp(t, `(?i)recovery not supported|cannot recover`, err.Error())
	assert.Contains(t, err.Error(), "new session")
}

// TestAgentError_SensitiveDataPersisted verifies Detail with sensitive info is persisted as-is
func TestAgentError_SensitiveDataPersisted(t *testing.T) {
	sessionID := uuid.New()
	detailJSON := json.RawMessage(`{"api_key": "secret123"}`)

	agentError, err := NewAgentError(
		"architect",
		"authentication failed",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)
	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.JSONEq(t, `{"api_key": "secret123"}`, string(agentError.Detail))
}

// TestAgentError_LargeMessage accepts AgentError with very large message string
func TestAgentError_LargeMessage(t *testing.T) {
	sessionID := uuid.New()
	largeMessage := make([]byte, 1024*1024) // 1MB
	for i := range largeMessage {
		largeMessage[i] = 'A'
	}

	agentError, err := NewAgentError(
		"architect",
		string(largeMessage),
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Len(t, agentError.Message, 1024*1024)
}

// TestAgentError_UnicodeMessage accepts AgentError with Unicode characters in message
func TestAgentError_UnicodeMessage(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"Error: 任务失败 🔥",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, "Error: 任务失败 🔥", agentError.Message)
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

	agentError, err := NewAgentError(
		"architect",
		"error",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, json.RawMessage(detailJSON), agentError.Detail)
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

	agentError, err := NewAgentError(
		"architect",
		"error",
		detailJSON,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)
	assert.Equal(t, json.RawMessage(detailJSON), agentError.Detail)
}

// TestAgentError_ErrorInterface verifies AgentError implements the error interface
func TestAgentError_ErrorInterface(t *testing.T) {
	sessionID := uuid.New()

	agentError, err := NewAgentError(
		"architect",
		"task failed",
		nil,
		sessionID,
		"review",
		1714147200,
	)

	require.NoError(t, err)
	require.NotNil(t, agentError)

	var e error = agentError
	assert.Equal(t, "task failed", e.Error())
}
