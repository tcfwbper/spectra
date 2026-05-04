package runtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock types for SessionFinalizer tests ---

// mockSessionForFinalizer is a mock Session for SessionFinalizer tests.
type mockSessionForFinalizer struct {
	id           string
	status       string
	workflowName string
	err          error
}

func (m *mockSessionForFinalizer) GetID() string           { return m.id }
func (m *mockSessionForFinalizer) GetStatusSafe() string   { return m.status }
func (m *mockSessionForFinalizer) GetWorkflowName() string { return m.workflowName }
func (m *mockSessionForFinalizer) GetErrorSafe() error     { return m.err }

// mockLoggerForFinalizer captures log messages for SessionFinalizer tests.
type mockLoggerForFinalizer struct {
	mu       sync.Mutex
	warnings []string
}

func (m *mockLoggerForFinalizer) Warning(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnings = append(m.warnings, msg)
}

func (m *mockLoggerForFinalizer) getWarnings() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.warnings...)
}

// --- Mock error types for SessionFinalizer tests ---

// finalizerAgentError represents an agent error for SessionFinalizer tests.
// SessionFinalizer is expected to recognize this type via interface assertion.
type finalizerAgentError struct {
	message      string
	agentRole    string
	failingState string
	detail       map[string]any
}

func (e *finalizerAgentError) Error() string             { return e.message }
func (e *finalizerAgentError) GetAgentRole() string      { return e.agentRole }
func (e *finalizerAgentError) GetFailingState() string   { return e.failingState }
func (e *finalizerAgentError) GetDetail() map[string]any { return e.detail }
func (e *finalizerAgentError) IsAgentError() bool        { return true }

// finalizerRuntimeError represents a runtime error for SessionFinalizer tests.
// SessionFinalizer is expected to recognize this type via interface assertion.
type finalizerRuntimeError struct {
	message      string
	issuer       string
	failingState string
	detail       map[string]any
}

func (e *finalizerRuntimeError) Error() string             { return e.message }
func (e *finalizerRuntimeError) GetIssuer() string         { return e.issuer }
func (e *finalizerRuntimeError) GetFailingState() string   { return e.failingState }
func (e *finalizerRuntimeError) GetDetail() map[string]any { return e.detail }
func (e *finalizerRuntimeError) IsRuntimeError() bool      { return true }

// --- Output capture helper ---

// captureStdoutStderr redirects os.Stdout and os.Stderr to pipes, executes fn,
// and returns the captured output as strings. Not thread-safe; do not use in parallel tests.
func captureStdoutStderr(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, err := os.Pipe()
	require.NoError(t, err)
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = wOut
	os.Stderr = wErr

	var outBuf, errBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&outBuf, rOut)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&errBuf, rErr)
	}()

	fn()

	wOut.Close()
	wErr.Close()
	wg.Wait()
	rOut.Close()
	rErr.Close()

	return outBuf.String(), errBuf.String()
}

// --- Test fixture ---

func createSessionFinalizerFixture(t *testing.T) *mockLoggerForFinalizer {
	t.Helper()
	logger := &mockLoggerForFinalizer{}
	return logger
}

// =====================================================================
// Happy Path — Construction
// =====================================================================

func TestSessionFinalizer_New(t *testing.T) {
	logger := createSessionFinalizerFixture(t)

	sf, err := NewSessionFinalizer(logger)

	require.NoError(t, err)
	assert.NotNil(t, sf)
}

// =====================================================================
// Happy Path — Finalize (Completed Session)
// =====================================================================

func TestFinalize_CompletedSession_StdoutOutput(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:           "abc-123",
		status:       "completed",
		workflowName: "TestWorkflow",
	}

	stdout, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stdout, "Session abc-123 completed successfully. Workflow: TestWorkflow")
	assert.Empty(t, stderr, "no stderr output expected for completed session")
}

// =====================================================================
// Happy Path — Finalize (Failed Session with AgentError)
// =====================================================================

func TestFinalize_FailedSession_AgentError_FullDetail(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	agentErr := &finalizerAgentError{
		message:      "validation failed",
		agentRole:    "reviewer",
		failingState: "review_node",
		detail:       map[string]any{"code": float64(400), "context": "missing field"},
	}
	sess := &mockSessionForFinalizer{
		id:           "def-456",
		status:       "failed",
		workflowName: "ReviewWorkflow",
		err:          agentErr,
	}

	stdout, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Session def-456 failed. Workflow: ReviewWorkflow")
	assert.Contains(t, stderr, "Error: validation failed")
	assert.Contains(t, stderr, "Agent: reviewer")
	assert.Contains(t, stderr, "State: review_node")

	// Verify compact JSON detail
	detailJSON, marshalErr := json.Marshal(agentErr.detail)
	require.NoError(t, marshalErr)
	assert.Contains(t, stderr, "Detail: "+string(detailJSON))

	assert.Empty(t, stdout, "no stdout output expected for failed session")
}

func TestFinalize_FailedSession_AgentError_EmptyDetail(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	agentErr := &finalizerAgentError{
		message:      "agent error",
		agentRole:    "architect",
		failingState: "design_node",
		detail:       map[string]any{},
	}
	sess := &mockSessionForFinalizer{
		id:           "test-id",
		status:       "failed",
		workflowName: "TestWorkflow",
		err:          agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Session test-id failed. Workflow: TestWorkflow")
	assert.Contains(t, stderr, "Error: agent error")
	assert.Contains(t, stderr, "Agent: architect")
	assert.Contains(t, stderr, "State: design_node")
	assert.NotContains(t, stderr, "Detail:", "Detail line should NOT be present for empty detail")
}

func TestFinalize_FailedSession_AgentError_NullDetail(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	agentErr := &finalizerAgentError{
		message:      "error",
		agentRole:    "reviewer",
		failingState: "node1",
		detail:       nil,
	}
	sess := &mockSessionForFinalizer{
		id:           "test-id",
		status:       "failed",
		workflowName: "TestWorkflow",
		err:          agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Error: error")
	assert.Contains(t, stderr, "Agent: reviewer")
	assert.Contains(t, stderr, "State: node1")
	assert.NotContains(t, stderr, "Detail:", "Detail line should NOT be present for nil detail")
}

// =====================================================================
// Happy Path — Finalize (Failed Session with RuntimeError)
// =====================================================================

func TestFinalize_FailedSession_RuntimeError_FullDetail(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	rtErr := &finalizerRuntimeError{
		message:      "timeout exceeded",
		issuer:       "SessionInitializer",
		failingState: "start",
		detail:       map[string]any{"duration": "30s"},
	}
	sess := &mockSessionForFinalizer{
		id:           "ghi-789",
		status:       "failed",
		workflowName: "TestWorkflow",
		err:          rtErr,
	}

	stdout, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Session ghi-789 failed. Workflow: TestWorkflow")
	assert.Contains(t, stderr, "Error: timeout exceeded")
	assert.Contains(t, stderr, "Issuer: SessionInitializer")
	assert.Contains(t, stderr, "State: start")

	detailJSON, marshalErr := json.Marshal(rtErr.detail)
	require.NoError(t, marshalErr)
	assert.Contains(t, stderr, "Detail: "+string(detailJSON))

	assert.Empty(t, stdout, "no stdout output expected for failed session")
}

func TestFinalize_FailedSession_RuntimeError_EmptyDetail(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	rtErr := &finalizerRuntimeError{
		message:      "runtime error",
		issuer:       "MessageRouter",
		failingState: "node2",
		detail:       map[string]any{},
	}
	sess := &mockSessionForFinalizer{
		status:       "failed",
		workflowName: "TestWorkflow",
		err:          rtErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Error: runtime error")
	assert.Contains(t, stderr, "Issuer: MessageRouter")
	assert.Contains(t, stderr, "State: node2")
	assert.NotContains(t, stderr, "Detail:", "Detail line should NOT be present for empty detail")
}

// =====================================================================
// Validation Failures — Session Status
// =====================================================================

func TestFinalize_NonTerminalStatus_Initializing(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:           "test-id",
		status:       "initializing",
		workflowName: "TestWorkflow",
	}

	stdout, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	// Should log warning about non-terminal status
	warnings := logger.getWarnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "SessionFinalizer called with non-terminal session status 'initializing'. This may indicate a programming error or signal interruption.") {
			found = true
			break
		}
	}
	assert.True(t, found, "should log warning about non-terminal status 'initializing'")

	// Non-terminal status should print to stderr with "terminated with status" format
	assert.Contains(t, stderr, "Session test-id terminated with status 'initializing'. Workflow: TestWorkflow")
	assert.Empty(t, stdout, "non-terminal status should not print to stdout")
}

func TestFinalize_NonTerminalStatus_Running(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:           "test-id",
		status:       "running",
		workflowName: "TestWorkflow",
	}

	stdout, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	// Should log warning about non-terminal status
	warnings := logger.getWarnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "SessionFinalizer called with non-terminal session status 'running'. This may indicate a programming error or signal interruption.") {
			found = true
			break
		}
	}
	assert.True(t, found, "should log warning about non-terminal status 'running'")

	// Non-terminal status should print to stderr with "terminated with status" format
	assert.Contains(t, stderr, "Session test-id terminated with status 'running'. Workflow: TestWorkflow")
	assert.Empty(t, stdout, "non-terminal status should not print to stdout")
}

// =====================================================================
// Error Propagation — Nil Error
// =====================================================================

func TestFinalize_FailedSession_NilError(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:           "xyz-999",
		status:       "failed",
		workflowName: "TestWorkflow",
		err:          nil,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Session xyz-999 failed. Workflow: TestWorkflow")
	assert.Contains(t, stderr, "Error: <unknown error>")
	assert.NotContains(t, stderr, "Agent:")
	assert.NotContains(t, stderr, "Issuer:")
	assert.NotContains(t, stderr, "State:")
	assert.NotContains(t, stderr, "Detail:")
}

// =====================================================================
// Error Propagation — Unknown Error Type
// =====================================================================

func TestFinalize_FailedSession_UnknownErrorType(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:           "test-id",
		status:       "failed",
		workflowName: "TestWorkflow",
		err:          errors.New("generic error"),
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Session test-id failed. Workflow: TestWorkflow")
	assert.Contains(t, stderr, "Error: generic error")
	assert.NotContains(t, stderr, "Agent:")
	assert.NotContains(t, stderr, "Issuer:")
	assert.NotContains(t, stderr, "State:")
	assert.NotContains(t, stderr, "Detail:")
}

// =====================================================================
// Error Propagation — Detail Serialization Failure
// =====================================================================

func TestFinalize_DetailSerializationFails(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	// func() is a Go function which is not JSON-serializable
	agentErr := &finalizerAgentError{
		message: "test error",
		detail:  map[string]any{"func": func() {}},
	}
	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Detail: <failed to serialize detail>")

	// Logger should contain serialization error
	warnings := logger.getWarnings()
	foundSerializationErr := false
	for _, w := range warnings {
		if strings.Contains(strings.ToLower(w), "serial") ||
			strings.Contains(strings.ToLower(w), "json") ||
			strings.Contains(strings.ToLower(w), "marshal") {
			foundSerializationErr = true
			break
		}
	}
	assert.True(t, foundSerializationErr, "should log serialization error")
}

// =====================================================================
// Boundary Values — Large Content
// =====================================================================

func TestFinalize_VeryLargeErrorMessage(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	largeMessage := strings.Repeat("x", 10*1024) // 10 KB
	agentErr := &finalizerAgentError{
		message: largeMessage,
	}
	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Error: "+largeMessage, "entire 10 KB message should be printed without truncation")
}

func TestFinalize_VeryLargeDetail(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	largeTrace := strings.Repeat("a", 1024*1024) // 1 MB
	rtErr := &finalizerRuntimeError{
		message: "error",
		detail:  map[string]any{"trace": largeTrace},
	}
	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    rtErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	detailJSON, marshalErr := json.Marshal(rtErr.detail)
	require.NoError(t, marshalErr)
	assert.Contains(t, stderr, "Detail: "+string(detailJSON), "entire 1 MB detail JSON should be printed")
}

// =====================================================================
// Boundary Values — Special Characters
// =====================================================================

func TestFinalize_SessionIDWithSpecialChars(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:     "abc\n123",
		status: "completed",
	}

	stdout, _ := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	// Session ID with newline should be printed as-is (no escaping)
	assert.Contains(t, stdout, "Session abc\n123 completed successfully")
}

func TestFinalize_WorkflowNameWithSpecialChars(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		workflowName: "Test\tWorkflow",
		status:       "completed",
	}

	stdout, _ := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	// Workflow name with tab should be printed as-is (no escaping)
	assert.Contains(t, stdout, "Workflow: Test\tWorkflow")
}

func TestFinalize_ErrorMessageWithUnicode(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	agentErr := &finalizerAgentError{
		message: "错误: emoji 🚨",
	}
	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Error: 错误: emoji 🚨", "Unicode characters should be preserved")
}

// =====================================================================
// Boundary Values — Empty Fields
// =====================================================================

func TestFinalize_EmptySessionID(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		id:           "",
		status:       "completed",
		workflowName: "TestWorkflow",
	}

	stdout, _ := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stdout, "Session  completed successfully. Workflow: TestWorkflow")
}

func TestFinalize_EmptyWorkflowName(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		status:       "completed",
		workflowName: "",
	}

	stdout, _ := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stdout, "Workflow: ")
}

func TestFinalize_EmptyAgentRole(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	agentErr := &finalizerAgentError{
		agentRole: "",
	}
	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Agent: ", "empty agent role should be printed as-is")
}

// =====================================================================
// Idempotency
// =====================================================================

func TestFinalize_CalledMultipleTimes(t *testing.T) {
	logger := &mockLoggerForFinalizer{}

	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		status:       "completed",
		workflowName: "TestWorkflow",
	}

	for i := 0; i < 3; i++ {
		stdout, _ := captureStdoutStderr(t, func() {
			sf.Finalize(sess)
		})
		assert.Contains(t, stdout, "completed successfully", "call %d should print success message", i)
	}
}

// =====================================================================
// Resource Cleanup — Session Files Retained
// =====================================================================

func TestFinalize_SessionFilesNotDeleted(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	// Create temporary session directory with session.json and events.jsonl
	tmpDir := t.TempDir()
	sessionDir := tmpDir + "/abc-session"
	require.NoError(t, os.MkdirAll(sessionDir, 0775))
	require.NoError(t, os.WriteFile(sessionDir+"/session.json", []byte("{}"), 0644))
	require.NoError(t, os.WriteFile(sessionDir+"/events.jsonl", []byte(""), 0644))

	sess := &mockSessionForFinalizer{
		status: "completed",
	}

	captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	// Verify session directory and files still exist (SessionFinalizer does not perform resource cleanup)
	_, statErr := os.Stat(sessionDir)
	assert.NoError(t, statErr, "session directory should still exist")
	_, statErr = os.Stat(sessionDir + "/session.json")
	assert.NoError(t, statErr, "session.json should still exist")
	_, statErr = os.Stat(sessionDir + "/events.jsonl")
	assert.NoError(t, statErr, "events.jsonl should still exist")
}

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

func TestFinalize_NoReturnError(t *testing.T) {
	logger := &mockLoggerForFinalizer{}

	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    nil,
	}

	// Finalize should complete without panic. SessionFinalizer never returns error.
	captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})
}

func TestFinalize_OutputStreamClosed(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		status: "completed",
	}

	// Redirect stdout and stderr to closed pipes
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	_, wOut, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	_, wErr, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)

	// Close the write ends immediately to simulate closed streams
	wOut.Close()
	wErr.Close()

	os.Stdout = wOut
	os.Stderr = wErr

	// SessionFinalizer does not check for print errors — should not panic
	assert.NotPanics(t, func() {
		sf.Finalize(sess)
	})
}

// =====================================================================
// State Transitions — Terminal Status
// =====================================================================

func TestFinalize_CompletedStatus_NoErrorField(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	sess := &mockSessionForFinalizer{
		status: "completed",
		err:    nil,
	}

	stdout, _ := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stdout, "completed successfully")
}

func TestFinalize_FailedStatus_ErrorFieldPresent(t *testing.T) {
	logger := createSessionFinalizerFixture(t)
	sf, err := NewSessionFinalizer(logger)
	require.NoError(t, err)

	agentErr := &finalizerAgentError{
		message: "test",
	}
	sess := &mockSessionForFinalizer{
		status: "failed",
		err:    agentErr,
	}

	_, stderr := captureStdoutStderr(t, func() {
		sf.Finalize(sess)
	})

	assert.Contains(t, stderr, "Error: test")
}
