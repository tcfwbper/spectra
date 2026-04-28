package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Happy Path — Construction

func TestSession_ValidConstruction(t *testing.T) {
	// Mock workflow with EntryNode="start", WorkflowName="TestWorkflow"
	entryNode := "start"
	workflowName := "TestWorkflow"

	// Session constructed via SessionInitializer
	now := time.Now().Unix()
	session := createTestSession(t, "initializing", entryNode)
	session.WorkflowName = workflowName

	// Returns valid Session; SessionMetadata fields embedded and accessible
	assert.NotEmpty(t, session.ID)
	_, err := uuid.Parse(session.ID)
	require.NoError(t, err, "ID should be a valid UUID")

	assert.Equal(t, "initializing", session.Status)
	assert.Equal(t, workflowName, session.WorkflowName)
	assert.Equal(t, entryNode, session.CurrentState)
	assert.GreaterOrEqual(t, session.CreatedAt, now)
	assert.GreaterOrEqual(t, session.UpdatedAt, now)
	assert.Equal(t, session.CreatedAt, session.UpdatedAt)
	assert.NotNil(t, session.SessionData)
	assert.Equal(t, 0, len(session.SessionData))
	assert.Nil(t, session.Error)
	assert.NotNil(t, session.EventHistory)
	assert.Equal(t, 0, len(session.EventHistory))
	assert.NotNil(t, session.terminationNotifier)
}

func TestSession_TimestampInitialization(t *testing.T) {
	// Session constructed
	session := createTestSession(t, "initializing", "start")

	// CreatedAt == UpdatedAt; both are POSIX timestamps > 0
	assert.Equal(t, session.CreatedAt, session.UpdatedAt)
	assert.Greater(t, session.CreatedAt, int64(0))
	assert.Greater(t, session.UpdatedAt, int64(0))
}

func TestSession_EmptyEventHistory(t *testing.T) {
	// Session constructed
	session := createTestSession(t, "initializing", "start")

	// EventHistory is empty slice (length 0); not nil
	assert.NotNil(t, session.EventHistory)
	assert.Equal(t, 0, len(session.EventHistory))
}

func TestSession_EmptySessionData(t *testing.T) {
	// Session constructed
	session := createTestSession(t, "initializing", "start")

	// SessionData is empty map (length 0); not nil
	assert.NotNil(t, session.SessionData)
	assert.Equal(t, 0, len(session.SessionData))
}

func TestSession_SessionMetadataEmbedded(t *testing.T) {
	// Session constructed with ID="test-uuid", WorkflowName="TestFlow"
	session := createTestSession(t, "initializing", "start")
	testID := uuid.New().String()
	session.ID = testID
	session.WorkflowName = "TestFlow"

	// Can access session.ID, session.WorkflowName, session.Status, etc. directly
	assert.Equal(t, testID, session.ID)
	assert.Equal(t, "TestFlow", session.WorkflowName)
	assert.Equal(t, "initializing", session.Status)
	assert.Greater(t, session.CreatedAt, int64(0))
	assert.Greater(t, session.UpdatedAt, int64(0))
	assert.Equal(t, "start", session.CurrentState)
	assert.NotNil(t, session.SessionData)
	assert.Nil(t, session.Error)
}

// Happy Path — Field Access via Getters

func TestSession_GetStatusSafe(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	// Call GetStatusSafe()
	status := session.GetStatusSafe()

	// Returns "running"
	assert.Equal(t, "running", status)
}

func TestSession_GetCurrentStateSafe(t *testing.T) {
	// Session with CurrentState="processing"
	session := createTestSession(t, "running", "processing")

	// Call GetCurrentStateSafe()
	state := session.GetCurrentStateSafe()

	// Returns "processing"
	assert.Equal(t, "processing", state)
}

func TestSession_GetErrorSafeNil(t *testing.T) {
	// Session with Status="running", Error=nil
	session := createTestSession(t, "running", "processing")

	// Call GetErrorSafe()
	err := session.GetErrorSafe()

	// Returns nil
	assert.Nil(t, err)
}

func TestSession_GetErrorSafeAgentError(t *testing.T) {
	// Session with Status="failed", Error=*AgentError
	agentError := &AgentError{NodeName: "TestNode", Message: "test error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	// Call GetErrorSafe()
	err := session.GetErrorSafe()

	// Returns *AgentError matching stored error
	assert.NotNil(t, err)
	agentErr, ok := err.(*AgentError)
	require.True(t, ok)
	assert.Equal(t, "TestNode", agentErr.NodeName)
	assert.Equal(t, "test error", agentErr.Message)
}

func TestSession_GetErrorSafeRuntimeError(t *testing.T) {
	// Session with Status="failed", Error=*RuntimeError
	runtimeError := &RuntimeError{Issuer: "Runtime", Message: "runtime error"}
	session := createTestSessionWithError(t, "failed", "processing", runtimeError)

	// Call GetErrorSafe()
	err := session.GetErrorSafe()

	// Returns *RuntimeError matching stored error
	assert.NotNil(t, err)
	runtimeErr, ok := err.(*RuntimeError)
	require.True(t, ok)
	assert.Equal(t, "Runtime", runtimeErr.Issuer)
	assert.Equal(t, "runtime error", runtimeErr.Message)
}

// State Transitions

func TestSession_InitializingToRunning(t *testing.T) {
	// Session with Status="initializing"
	session := createTestSession(t, "initializing", "start")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Run(terminationNotifier)
	err := session.Run(session.terminationNotifier)

	// Returns nil; Status="running"; UpdatedAt refreshed
	assert.NoError(t, err)
	assert.Equal(t, "running", session.Status)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
}

func TestSession_RunningToCompleted(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Done(terminationNotifier)
	err := session.Done(session.terminationNotifier)

	// Returns nil; Status="completed"; UpdatedAt refreshed; notification sent
	assert.NoError(t, err)
	assert.Equal(t, "completed", session.Status)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)

	// Check notification
	select {
	case <-session.terminationNotifier.(chan struct{}):
		// Notification received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected notification on terminationNotifier")
	}
}

func TestSession_InitializingToFailed(t *testing.T) {
	// Session with Status="initializing"
	session := createTestSession(t, "initializing", "start")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Fail(*RuntimeError, terminationNotifier)
	runtimeError := &RuntimeError{Issuer: "Test", Message: "test error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	// Returns nil; Status="failed"; Error set; UpdatedAt refreshed; notification sent
	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.NotNil(t, session.Error)
	assert.Equal(t, runtimeError, session.Error)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)

	// Check notification
	select {
	case <-session.terminationNotifier.(chan struct{}):
		// Notification received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected notification on terminationNotifier")
	}
}

func TestSession_RunningToFailed(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")
	oldUpdatedAt := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Fail(*AgentError, terminationNotifier)
	agentError := &AgentError{NodeName: "agent", Message: "agent error"}
	err := session.Fail(agentError, session.terminationNotifier)

	// Returns nil; Status="failed"; Error set; UpdatedAt refreshed; notification sent
	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.NotNil(t, session.Error)
	assert.Equal(t, agentError, session.Error)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)

	// Check notification
	select {
	case <-session.terminationNotifier.(chan struct{}):
		// Notification received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected notification on terminationNotifier")
	}
}

// Validation Failures — Status Preconditions

func TestSession_RunOnRunning(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	// Call Run(terminationNotifier)
	err := session.Run(session.terminationNotifier)

	// Returns error matching /cannot run.*status is 'running'/i; status unchanged
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot run.*status is 'running'", err.Error())
	assert.Equal(t, "running", session.Status)
}

func TestSession_RunOnCompleted(t *testing.T) {
	// Session with Status="completed"
	session := createTestSession(t, "completed", "exit")

	// Call Run(terminationNotifier)
	err := session.Run(session.terminationNotifier)

	// Returns error matching /cannot run.*status is 'completed'/i; status unchanged
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot run.*status is 'completed'", err.Error())
	assert.Equal(t, "completed", session.Status)
}

func TestSession_RunOnFailed(t *testing.T) {
	// Session with Status="failed"
	session := createTestSessionWithError(t, "failed", "processing", &RuntimeError{Issuer: "Test", Message: "error"})

	// Call Run(terminationNotifier)
	err := session.Run(session.terminationNotifier)

	// Returns error matching /cannot run.*status is 'failed'/i; status unchanged
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot run.*status is 'failed'", err.Error())
	assert.Equal(t, "failed", session.Status)
}

func TestSession_DoneOnInitializing(t *testing.T) {
	// Session with Status="initializing"
	session := createTestSession(t, "initializing", "start")

	// Call Done(terminationNotifier)
	err := session.Done(session.terminationNotifier)

	// Returns error matching /cannot complete.*status is 'initializing'/i; status unchanged
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot complete.*status is 'initializing'", err.Error())
	assert.Equal(t, "initializing", session.Status)
}

func TestSession_DoneOnCompleted(t *testing.T) {
	// Session with Status="completed"
	session := createTestSession(t, "completed", "exit")

	// Call Done(terminationNotifier)
	err := session.Done(session.terminationNotifier)

	// Returns error matching /cannot complete.*status is 'completed'/i; status unchanged
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot complete.*status is 'completed'", err.Error())
	assert.Equal(t, "completed", session.Status)
}

func TestSession_DoneOnFailed(t *testing.T) {
	// Session with Status="failed"
	session := createTestSessionWithError(t, "failed", "processing", &AgentError{NodeName: "agent", Message: "error"})

	// Call Done(terminationNotifier)
	err := session.Done(session.terminationNotifier)

	// Returns error matching /cannot complete.*status is 'failed'/i; status unchanged
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot complete.*status is 'failed'", err.Error())
	assert.Equal(t, "failed", session.Status)
}

// Validation Failures — Terminal State Finality

func TestSession_FailOnCompleted(t *testing.T) {
	// Session with Status="completed"
	session := createTestSession(t, "completed", "exit")

	// Call Fail(*RuntimeError, terminationNotifier)
	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	// Returns error matching /cannot fail.*status is 'completed'/i; status unchanged; Error remains nil
	assert.Error(t, err)
	assert.Regexp(t, "(?i)cannot fail.*status is 'completed'", err.Error())
	assert.Equal(t, "completed", session.Status)
	assert.Nil(t, session.Error)
}

func TestSession_FailOnFailed(t *testing.T) {
	// Session with Status="failed", Error=*AgentError("first error")
	firstError := &AgentError{NodeName: "first", Message: "first error"}
	session := createTestSessionWithError(t, "failed", "processing", firstError)

	// Call Fail(*RuntimeError("second error"), terminationNotifier)
	secondError := &RuntimeError{Issuer: "second", Message: "second error"}
	err := session.Fail(secondError, session.terminationNotifier)

	// Returns error matching /session already failed/i; Error remains first error
	assert.Error(t, err)
	assert.Regexp(t, "(?i)session already failed", err.Error())
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, firstError, session.Error)
}

// Read-Only Convention

func TestSession_SessionMetadataFieldsExported(t *testing.T) {
	// Session constructed with ID=<uuid>, WorkflowName="TestFlow", CreatedAt=<timestamp>
	session := createTestSession(t, "initializing", "start")
	session.WorkflowName = "TestFlow"

	// Access session.ID, session.WorkflowName, session.CreatedAt
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "TestFlow", session.WorkflowName)
	assert.Greater(t, session.CreatedAt, int64(0))

	// All fields accessible; modification not prevented by compiler (read-only by convention only)
	// This test just verifies access is possible
}

func TestSession_MutationThroughMethodsOnly(t *testing.T) {
	// Session with Status="initializing"
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call session.Run(terminationNotifier) instead of session.Status = "running"
	err := session.Run(session.terminationNotifier)

	// Status updated correctly via method
	assert.NoError(t, err)
	assert.Equal(t, "running", session.Status)

	// Direct assignment possible but not recommended (just verify it's possible)
	session.Status = "completed"
	assert.Equal(t, "completed", session.Status)
}

// Atomic Replacement

func TestSession_ErrorReplacementAtomic(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Fail(*AgentError, terminationNotifier) under write lock
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	err := session.Fail(agentError, session.terminationNotifier)

	// Both Status and Error updated in same critical section
	assert.NoError(t, err)
	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, agentError, session.Error)
}

// Invariants — Status-Error Correlation

func TestSession_ErrorNilWhenNotFailed(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	// Check Error field via GetErrorSafe()
	err := session.GetErrorSafe()

	// Returns nil
	assert.Nil(t, err)
}

func TestSession_ErrorNonNilWhenFailed(t *testing.T) {
	// Session with Status="failed", Error=*AgentError
	agentError := &AgentError{NodeName: "agent", Message: "error"}
	session := createTestSessionWithError(t, "failed", "processing", agentError)

	// Check Error field via GetErrorSafe()
	err := session.GetErrorSafe()

	// Returns non-nil *AgentError
	assert.NotNil(t, err)
	assert.Equal(t, agentError, err)
}

// Invariants — Timestamp Ordering

func TestSession_UpdatedAtRefreshedOnMutation(t *testing.T) {
	// Session with Status="initializing", UpdatedAt=T0
	session := createTestSession(t, "initializing", "start")
	oldUpdatedAt := session.UpdatedAt

	// Wait 1 second; call Run(terminationNotifier)
	time.Sleep(1 * time.Second)

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	err := session.Run(session.terminationNotifier)

	// UpdatedAt > T0; UpdatedAt >= CreatedAt
	assert.NoError(t, err)
	assert.Greater(t, session.UpdatedAt, oldUpdatedAt)
	assert.GreaterOrEqual(t, session.UpdatedAt, session.CreatedAt)
}

func TestSession_TimestampOrderingMaintained(t *testing.T) {
	// Session constructed at T0
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Perform multiple mutations (Run, UpdateSessionDataSafe, etc.)
	err := session.Run(session.terminationNotifier)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, session.UpdatedAt, session.CreatedAt)

	time.Sleep(10 * time.Millisecond)
	err = session.UpdateSessionDataSafe("key", "value")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, session.UpdatedAt, session.CreatedAt)

	time.Sleep(10 * time.Millisecond)
	err = session.UpdateCurrentStateSafe("newState")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, session.UpdatedAt, session.CreatedAt)
}

// Invariants — In-Memory Authority

func TestSession_InMemoryStateAuthoritative(t *testing.T) {
	// Mock SessionMetadataStore that returns error on write
	session := createTestSession(t, "initializing", "start")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	// Call Run(terminationNotifier)
	err := session.Run(session.terminationNotifier)

	// Returns nil; in-memory Status="running"; persistence failure logged as warning
	assert.NoError(t, err)
	assert.Equal(t, "running", session.Status)
	session.logger.AssertCalled(t, "Warning", mock.Anything)
}

func TestSession_PersistenceFailureLogged(t *testing.T) {
	// Mock SessionMetadataStore that fails; session with Status="running"; mock logger
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(assert.AnError)
	session.logger.On("Warning", mock.Anything).Return()

	// Call UpdateSessionDataSafe("key", "value")
	err := session.UpdateSessionDataSafe("key", "value")

	// Returns nil; warning logged matching /persistence failed/i; in-memory state updated
	assert.NoError(t, err)
	assert.Equal(t, "value", session.SessionData["key"])
	session.logger.AssertCalled(t, "Warning", mock.MatchedBy(func(msg string) bool {
		return assert.Regexp(t, "(?i)persistence failed", msg)
	}))
}

// Happy Path — SessionMetadata Persistence Integration

func TestSession_SessionMetadataExtraction(t *testing.T) {
	// Session with populated fields
	session := createTestSession(t, "running", "processing")
	session.SessionData["key"] = "value"
	session.EventHistory = append(session.EventHistory, Event{ID: "evt-1"})

	// Access embedded SessionMetadata fields directly or copy struct
	// SessionMetadata fields accessible
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "TestWorkflow", session.WorkflowName)
	assert.Equal(t, "running", session.Status)
	assert.Greater(t, session.CreatedAt, int64(0))
	assert.Greater(t, session.UpdatedAt, int64(0))
	assert.Equal(t, "processing", session.CurrentState)
	assert.Equal(t, "value", session.SessionData["key"])
	assert.Nil(t, session.Error)

	// Runtime-only fields (EventHistory, mu, terminationNotifier) not included in SessionMetadata
	assert.NotEmpty(t, session.EventHistory)
}

// Edge Cases

func TestSession_TerminationNotifierCapacity(t *testing.T) {
	// Session constructed with buffered channel
	session := createTestSession(t, "initializing", "start")

	// Check channel capacity
	ch := session.terminationNotifier.(chan struct{})
	assert.GreaterOrEqual(t, cap(ch), 2)
}

func TestSession_SingleNotificationOnDone(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Done(terminationNotifier)
	err := session.Done(session.terminationNotifier)

	// Exactly one value sent on channel; channel length increases by 1
	assert.NoError(t, err)
	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestSession_SingleNotificationOnFail(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Fail(*RuntimeError, terminationNotifier)
	runtimeError := &RuntimeError{Issuer: "Test", Message: "error"}
	err := session.Fail(runtimeError, session.terminationNotifier)

	// Exactly one value sent on channel; channel length increases by 1
	assert.NoError(t, err)
	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}

func TestSession_NoDuplicateNotifications(t *testing.T) {
	// Session with Status="running"
	session := createTestSession(t, "running", "processing")

	session.metadataStore.On("Write", mock.Anything).Return(nil)

	// Call Done(terminationNotifier); attempt second Done(terminationNotifier)
	err := session.Done(session.terminationNotifier)
	assert.NoError(t, err)

	err = session.Done(session.terminationNotifier)

	// First call sends notification; second call returns error without sending
	assert.Error(t, err)
	ch := session.terminationNotifier.(chan struct{})
	assert.Equal(t, 1, len(ch))
}
