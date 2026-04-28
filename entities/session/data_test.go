package session

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Happy Path — UpdateSessionDataSafe

func TestUpdateSessionDataSafe_NewKey(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("key1", "value1")

	assert.NoError(t, err)
	assert.Equal(t, "value1", session.SessionData["key1"])
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestUpdateSessionDataSafe_OverwriteExisting(t *testing.T) {
	data := map[string]any{"key1": "old"}
	session := createTestSessionWithData(t, "running", "processing", data)
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("key1", "new")

	assert.NoError(t, err)
	assert.Equal(t, "new", session.SessionData["key1"])
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestUpdateSessionDataSafe_MultipleKeys(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err1 := session.UpdateSessionDataSafe("key1", "val1")
	err2 := session.UpdateSessionDataSafe("key2", "val2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "val1", session.SessionData["key1"])
	assert.Equal(t, "val2", session.SessionData["key2"])
}

func TestUpdateSessionDataSafe_NilValueForNonClaudeSessionID(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("logicSpec.result", nil)

	assert.NoError(t, err)
	assert.Nil(t, session.SessionData["logicSpec.result"])

	val, ok := session.GetSessionDataSafe("logicSpec.result")
	assert.Nil(t, val)
	assert.True(t, ok)
}

func TestUpdateSessionDataSafe_PersistsToStore(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("key", "value")

	assert.NoError(t, err)
	session.metadataStore.AssertCalled(t, "Write", mock.Anything)
}

// Happy Path — GetSessionDataSafe

func TestGetSessionDataSafe_ExistingKey(t *testing.T) {
	data := map[string]any{"key1": "value1"}
	session := createTestSessionWithData(t, "running", "processing", data)

	val, ok := session.GetSessionDataSafe("key1")

	assert.True(t, ok)
	assert.Equal(t, "value1", val)
}

func TestGetSessionDataSafe_MissingKey(t *testing.T) {
	data := map[string]any{"key1": "value1"}
	session := createTestSessionWithData(t, "running", "processing", data)

	val, ok := session.GetSessionDataSafe("key2")

	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestGetSessionDataSafe_NilValue(t *testing.T) {
	data := map[string]any{"key1": nil}
	session := createTestSessionWithData(t, "running", "processing", data)

	val, ok := session.GetSessionDataSafe("key1")

	assert.True(t, ok)
	assert.Nil(t, val)
}

func TestGetSessionDataSafe_EmptyKey(t *testing.T) {
	data := map[string]any{"key1": "value1"}
	session := createTestSessionWithData(t, "running", "processing", data)

	val, ok := session.GetSessionDataSafe("")

	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestGetSessionDataSafe_DoesNotModifyUpdatedAt(t *testing.T) {
	data := map[string]any{"key": "val"}
	session := createTestSessionWithData(t, "running", "processing", data)
	oldUpdatedAt := session.UpdatedAt

	val, ok := session.GetSessionDataSafe("key")

	assert.True(t, ok)
	assert.Equal(t, "val", val)
	assert.Equal(t, oldUpdatedAt, session.UpdatedAt)
}

// Happy Path — ClaudeSessionID

func TestUpdateSessionDataSafe_ClaudeSessionIDString(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("ProcessNode.ClaudeSessionID", "session-abc-123")

	assert.NoError(t, err)
	assert.Equal(t, "session-abc-123", session.SessionData["ProcessNode.ClaudeSessionID"])
}

func TestUpdateSessionDataSafe_ClaudeSessionIDEmptyString(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("ReviewNode.ClaudeSessionID", "")

	assert.NoError(t, err)
	assert.Equal(t, "", session.SessionData["ReviewNode.ClaudeSessionID"])
}

func TestUpdateSessionDataSafe_ClaudeSessionIDOverwrite(t *testing.T) {
	data := map[string]any{"Node1.ClaudeSessionID": "old-session"}
	session := createTestSessionWithData(t, "running", "processing", data)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("Node1.ClaudeSessionID", "new-session")

	assert.NoError(t, err)
	assert.Equal(t, "new-session", session.SessionData["Node1.ClaudeSessionID"])
}

func TestGetSessionDataSafe_ClaudeSessionID(t *testing.T) {
	data := map[string]any{"AgentNode.ClaudeSessionID": "sess-456"}
	session := createTestSessionWithData(t, "running", "processing", data)

	val, ok := session.GetSessionDataSafe("AgentNode.ClaudeSessionID")

	assert.True(t, ok)
	assert.Equal(t, "sess-456", val)
}

// Happy Path — Namespace Conventions

func TestUpdateSessionDataSafe_LogicSpecNamespace(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	output := map[string]any{"result": "done"}
	err := session.UpdateSessionDataSafe("logicSpec.output", output)

	assert.NoError(t, err)
	assert.Equal(t, output, session.SessionData["logicSpec.output"])
}

func TestUpdateSessionDataSafe_ArbitraryNamespace(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("custom.key", 123)

	assert.NoError(t, err)
	assert.Equal(t, 123, session.SessionData["custom.key"])
}

// Validation Failures — UpdateSessionDataSafe

func TestUpdateSessionDataSafe_EmptyKey(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.UpdateSessionDataSafe("", "value")

	assert.Error(t, err)
	assert.Regexp(t, "(?i)session data key cannot be empty", err.Error())
	assert.Empty(t, session.SessionData)
}

func TestUpdateSessionDataSafe_ClaudeSessionIDNonString(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.UpdateSessionDataSafe("Node.ClaudeSessionID", 123)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)ClaudeSessionID value must be a string.*got.*int", err.Error())
	assert.Empty(t, session.SessionData)
}

func TestUpdateSessionDataSafe_ClaudeSessionIDNil(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.UpdateSessionDataSafe("AgentNode.ClaudeSessionID", nil)

	assert.Error(t, err)
	assert.Regexp(t, "(?i)ClaudeSessionID value must be a string", err.Error())
	assert.Empty(t, session.SessionData)
}

func TestUpdateSessionDataSafe_ClaudeSessionIDMap(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.UpdateSessionDataSafe("ProcessNode.ClaudeSessionID", map[string]any{"id": "sess"})

	assert.Error(t, err)
	assert.Regexp(t, "(?i)ClaudeSessionID value must be a string.*got.*map", err.Error())
	assert.Empty(t, session.SessionData)
}

func TestUpdateSessionDataSafe_ClaudeSessionIDSlice(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	err := session.UpdateSessionDataSafe("Node.ClaudeSessionID", []string{"sess1", "sess2"})

	assert.Error(t, err)
	assert.Regexp(t, "(?i)ClaudeSessionID value must be a string.*got.*slice", err.Error())
	assert.Empty(t, session.SessionData)
}

// Idempotency

func TestUpdateSessionDataSafe_Idempotent(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err1 := session.UpdateSessionDataSafe("key", "val")
	err2 := session.UpdateSessionDataSafe("key", "val")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "val", session.SessionData["key"])
}

func TestGetSessionDataSafe_Idempotent(t *testing.T) {
	data := map[string]any{"key": "val"}
	session := createTestSessionWithData(t, "running", "processing", data)

	val1, ok1 := session.GetSessionDataSafe("key")
	val2, ok2 := session.GetSessionDataSafe("key")

	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Equal(t, "val", val1)
	assert.Equal(t, val1, val2)
}

// Concurrent Behaviour

func TestSessionData_ConcurrentWrites(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := string(rune('a' + index))
			_ = session.UpdateSessionDataSafe(key, index)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 10, len(session.SessionData))
}

func TestSessionData_ConcurrentWritesSameKey(t *testing.T) {
	data := map[string]any{"key": "initial"}
	session := createTestSessionWithData(t, "running", "processing", data)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_ = session.UpdateSessionDataSafe("key", index)
		}(i)
	}

	wg.Wait()

	// Last write wins
	assert.NotEqual(t, "initial", session.SessionData["key"])
}

func TestSessionData_ConcurrentReads(t *testing.T) {
	data := map[string]any{"key": "value"}
	session := createTestSessionWithData(t, "running", "processing", data)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, ok := session.GetSessionDataSafe("key")
			assert.True(t, ok)
			assert.Equal(t, "value", val)
		}()
	}

	wg.Wait()
}

func TestSessionData_ConcurrentReadWrite(t *testing.T) {
	data := map[string]any{"key": "old"}
	session := createTestSessionWithData(t, "running", "processing", data)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, ok := session.GetSessionDataSafe("key")
			assert.True(t, ok)
			assert.True(t, val == "old" || val == "new")
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = session.UpdateSessionDataSafe("key", "new")
	}()

	wg.Wait()
}

// Error Propagation

func TestUpdateSessionDataSafe_PersistenceFailureLogged(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateSessionDataSafe("key", "value")

	assert.NoError(t, err)
	assert.Equal(t, "value", session.SessionData["key"])
	session.logger.AssertCalled(t, "Warning", mock.MatchedBy(func(msg string) bool {
		return strings.Contains(strings.ToLower(msg), "persistence failed")
	}))
}

func TestUpdateSessionDataSafe_PersistenceFailureDoesNotRevert(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	err := session.UpdateSessionDataSafe("key", "value")

	assert.NoError(t, err)
	assert.Equal(t, "value", session.SessionData["key"])
}

// Invariants — UpdatedAt Refresh

func TestUpdateSessionDataSafe_RefreshesUpdatedAt(t *testing.T) {
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt

	time.Sleep(1 * time.Second)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("key", "value")

	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

// Invariants — ClaudeSessionID Validation

func TestSessionData_ClaudeSessionIDSuffixCaseSensitive(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Lowercase suffix should not match ClaudeSessionID validation
	err := session.UpdateSessionDataSafe("Node.claudesessionid", 123)

	// Should succeed because suffix doesn't match exactly
	assert.NoError(t, err)
	assert.Equal(t, 123, session.SessionData["Node.claudesessionid"])
}

func TestSessionData_ClaudeSessionIDNoNodeNameValidation(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("NonExistentNode.ClaudeSessionID", "session-id")

	assert.NoError(t, err)
	assert.Equal(t, "session-id", session.SessionData["NonExistentNode.ClaudeSessionID"])
}

// Edge Cases

func TestUpdateSessionDataSafe_LargeValue(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	largeValue := make([]byte, 10*1024*1024)
	err := session.UpdateSessionDataSafe("key", largeValue)

	assert.NoError(t, err)
	assert.Equal(t, largeValue, session.SessionData["key"])
}

func TestUpdateSessionDataSafe_ComplexNestedValue(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	nested := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep",
			},
		},
	}
	err := session.UpdateSessionDataSafe("key", nested)

	assert.NoError(t, err)
	assert.Equal(t, nested, session.SessionData["key"])
}

func TestGetSessionDataSafe_TypeAssertion(t *testing.T) {
	data := map[string]any{"key": 123}
	session := createTestSessionWithData(t, "running", "processing", data)

	val, ok := session.GetSessionDataSafe("key")

	assert.True(t, ok)
	assert.Equal(t, 123, val)

	// Caller can assert to int
	intVal, ok := val.(int)
	assert.True(t, ok)
	assert.Equal(t, 123, intVal)
}

func TestSessionData_KeyWithDotNotation(t *testing.T) {
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("a.b.c", "value")

	assert.NoError(t, err)
	assert.Equal(t, "value", session.SessionData["a.b.c"])
}

func TestUpdateSessionDataSafe_OverwriteDifferentType(t *testing.T) {
	data := map[string]any{"key": "string"}
	session := createTestSessionWithData(t, "running", "processing", data)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.UpdateSessionDataSafe("key", 123)

	assert.NoError(t, err)
	assert.Equal(t, 123, session.SessionData["key"])
}
