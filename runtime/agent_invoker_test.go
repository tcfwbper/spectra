package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for AgentInvoker tests ---

// mockSessionForInvoker provides a mock Session for AgentInvoker tests.
type mockSessionForInvoker struct {
	mock.Mock
	mu          sync.RWMutex
	sessionID   string
	sessionData map[string]any
	failCalled  bool
}

func newMockSessionForInvoker() *mockSessionForInvoker {
	return &mockSessionForInvoker{
		sessionID:   uuid.New().String(),
		sessionData: make(map[string]any),
	}
}

func (m *mockSessionForInvoker) GetID() string {
	return m.sessionID
}

func (m *mockSessionForInvoker) GetSessionDataSafe(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *mockSessionForInvoker) UpdateSessionDataSafe(key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(key, value)
	if args.Error(0) == nil {
		m.sessionData[key] = value
	}
	return args.Error(0)
}

func (m *mockSessionForInvoker) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCalled = true
	args := m.Called(err, terminationNotifier)
	return args.Error(0)
}

func (m *mockSessionForInvoker) getFailCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failCalled
}

// mockUUIDGenerator provides a mock for UUID generation.
type mockUUIDGenerator struct {
	mock.Mock
}

func (m *mockUUIDGenerator) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// mockCommandInterceptor captures command arguments for verification.
type mockCommandInterceptor struct {
	mu       sync.Mutex
	args     []string
	dir      string
	env      []string
	stdout   *os.File
	stderr   *os.File
	started  bool
	startErr error
}

func (m *mockCommandInterceptor) getArgs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.args...)
}

func (m *mockCommandInterceptor) getDir() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dir
}

func (m *mockCommandInterceptor) getEnv() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.env...)
}

// --- Test fixture helper ---

func createAgentInvokerFixture(t *testing.T) (string, *mockSessionForInvoker) {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "agents"), 0755))
	sess := newMockSessionForInvoker()
	return tmpDir, sess
}

func defaultAgentDefinition() storage.AgentDefinition {
	return storage.AgentDefinition{
		Role:         "TestRole",
		Model:        "sonnet",
		Effort:       "normal",
		SystemPrompt: "You are a test agent",
		AgentRoot:    "agents",
	}
}

// --- Happy Path — New Session ---

func TestAgentInvoker_NewSession_GeneratesSessionID(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "TestNode.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "TestNode.ClaudeSessionID", mock.MatchedBy(func(val any) bool {
		s, ok := val.(string)
		if !ok {
			return false
		}
		_, err := uuid.Parse(s)
		return err == nil
	})).Return(nil)

	agentDef := defaultAgentDefinition()
	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("TestNode", "test message", agentDef)

	require.NoError(t, err)
	sess.AssertCalled(t, "UpdateSessionDataSafe", "TestNode.ClaudeSessionID", mock.AnythingOfType("string"))

	// Verify the stored value is a valid UUID v4
	for _, call := range sess.Calls {
		if call.Method == "UpdateSessionDataSafe" && call.Arguments.String(0) == "TestNode.ClaudeSessionID" {
			storedValue := call.Arguments.Get(1).(string)
			_, parseErr := uuid.Parse(storedValue)
			assert.NoError(t, parseErr, "stored ClaudeSessionID should be valid UUID v4")
		}
	}
}

func TestAgentInvoker_NewSession_StartsProcess(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "TestNode.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "TestNode.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.Model = "sonnet"
	agentDef.Effort = "normal"

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("TestNode", "test message", agentDef)

	assert.NoError(t, err, "process should start successfully for new session")
}

// --- Happy Path — Existing Session ---

func TestAgentInvoker_ExistingSession_UsesStoredSessionID(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "TestNode.ClaudeSessionID").Return("existing-uuid-1234", true)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("TestNode", "resume message", agentDef)

	assert.NoError(t, err)
	sess.AssertNotCalled(t, "UpdateSessionDataSafe", mock.Anything, mock.Anything)
}

func TestAgentInvoker_ExistingSession_NoSessionDataUpdate(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "NodeA.ClaudeSessionID").Return("uuid-5678", true)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("NodeA", "message", agentDef)

	assert.NoError(t, err)
	sess.AssertNotCalled(t, "UpdateSessionDataSafe", mock.Anything, mock.Anything)
}

// --- Happy Path — Command Construction ---

func TestAgentInvoker_CommandConstruction_AllFlags(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := storage.AgentDefinition{
		Role:            "Worker",
		Model:           "sonnet",
		Effort:          "high",
		SystemPrompt:    "You are X",
		AgentRoot:       "agents",
		AllowedTools:    []string{"Bash(*)", "Read(*)"},
		DisallowedTools: []string{"Write(*)"},
	}

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	assert.Contains(t, capturedArgs, "--permission-mode")
	assert.Contains(t, capturedArgs, "bypassPermission")
	assert.Contains(t, capturedArgs, "--model")
	assert.Contains(t, capturedArgs, "sonnet")
	assert.Contains(t, capturedArgs, "--effort")
	assert.Contains(t, capturedArgs, "high")
	assert.Contains(t, capturedArgs, "--system-prompt")
	assert.Contains(t, capturedArgs, "You are X")
	assert.Contains(t, capturedArgs, "--allowed-tools")
	assert.Contains(t, capturedArgs, "Bash(*)")
	assert.Contains(t, capturedArgs, "Read(*)")
	assert.Contains(t, capturedArgs, "--disallowed-tools")
	assert.Contains(t, capturedArgs, "Write(*)")
	assert.Contains(t, capturedArgs, "--session-id")
	assert.Contains(t, capturedArgs, "--print")
	assert.Contains(t, capturedArgs, "prompt")
}

func TestAgentInvoker_CommandConstruction_EmptyToolArrays(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AllowedTools = []string{}
	agentDef.DisallowedTools = []string{}

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	for _, arg := range capturedArgs {
		assert.NotEqual(t, "--allowed-tools", arg, "should omit --allowed-tools when empty")
		assert.NotEqual(t, "--disallowed-tools", arg, "should omit --disallowed-tools when empty")
	}
}

func TestAgentInvoker_CommandConstruction_WorkingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "agents", "logic"), 0755))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AgentRoot = "agents/logic"

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	capturedDir := invoker.CaptureCommandDir("Node", "prompt", agentDef)

	expectedDir := filepath.Join(tmpDir, "agents", "logic")
	assert.Equal(t, expectedDir, capturedDir, "working directory should be absolute path of ProjectRoot/AgentRoot")
}

func TestAgentInvoker_CommandConstruction_AgentRootDot(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AgentRoot = "."

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	capturedDir := invoker.CaptureCommandDir("Node", "prompt", agentDef)

	assert.Equal(t, tmpDir, capturedDir, "AgentRoot '.' should resolve to ProjectRoot")
}

// --- Happy Path — Environment Variables ---

func TestAgentInvoker_EnvVars_Injected(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)
	sess.sessionID = "session-123"

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedEnv := invoker.CaptureCommandEnv("Node", "prompt", agentDef)

	var foundSessionID, foundClaudeSessionID bool
	for _, env := range capturedEnv {
		if strings.HasPrefix(env, "SPECTRA_SESSION_ID=session-123") {
			foundSessionID = true
		}
		if strings.HasPrefix(env, "SPECTRA_CLAUDE_SESSION_ID=") {
			foundClaudeSessionID = true
		}
	}
	assert.True(t, foundSessionID, "should inject SPECTRA_SESSION_ID")
	assert.True(t, foundClaudeSessionID, "should inject SPECTRA_CLAUDE_SESSION_ID")
}

func TestAgentInvoker_EnvVars_OverrideParent(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)
	sess.sessionID = "new-session"

	t.Setenv("SPECTRA_SESSION_ID", "old-value")

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedEnv := invoker.CaptureCommandEnv("Node", "prompt", agentDef)

	// The last occurrence of SPECTRA_SESSION_ID should be the new value
	var lastSessionID string
	for _, env := range capturedEnv {
		if strings.HasPrefix(env, "SPECTRA_SESSION_ID=") {
			lastSessionID = env
		}
	}
	assert.Equal(t, "SPECTRA_SESSION_ID=new-session", lastSessionID, "injected value should override parent environment")
}

// --- Happy Path — Message Handling ---

func TestAgentInvoker_Message_WithQuotes(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	message := `He said "hello"`
	capturedArgs := invoker.CaptureCommandArgs("Node", message, agentDef)

	assert.Contains(t, capturedArgs, message, "message with double quotes should be passed as-is without manual escaping")
}

func TestAgentInvoker_Message_WithNewlines(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	message := "Line1\nLine2\t$VAR"
	capturedArgs := invoker.CaptureCommandArgs("Node", message, agentDef)

	assert.Contains(t, capturedArgs, message, "message with newlines and special characters should be preserved exactly")
}

func TestAgentInvoker_Message_MultilineSystemPrompt(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.SystemPrompt = "Line 1\nLine 2\n\"quoted\""

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	assert.Contains(t, capturedArgs, agentDef.SystemPrompt, "multi-line system prompt should be passed as-is")
}

// --- Happy Path — Asynchronous Execution ---

func TestAgentInvoker_ReturnsImmediately(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	start := time.Now()
	err = invoker.InvokeAgent("Node", "test", agentDef)
	elapsed := time.Since(start)

	// Should return in under 500ms (process runs in background)
	assert.NoError(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond, "InvokeAgent should return immediately without waiting for process")
}

func TestAgentInvoker_NoOutputCapture(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	stdout, stderr := invoker.CaptureCommandOutputConfig("Node", "prompt", agentDef)

	assert.Nil(t, stdout, "stdout should not be redirected (inherit from parent)")
	assert.Nil(t, stderr, "stderr should not be redirected (inherit from parent)")
}

// --- Validation Failures — Session Data ---

func TestAgentInvoker_SessionIDInvalidType(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	// Return int instead of string
	sess.On("GetSessionDataSafe", "BadNode.ClaudeSessionID").Return(123, true)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("BadNode", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)invalid Claude session ID type for node 'BadNode': expected string`, err.Error())
}

func TestAgentInvoker_UpdateSessionDataFails(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(fmt.Errorf("write error"))

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to update session with new Claude session ID`, err.Error())
}

func TestAgentInvoker_UpdateSessionDataFails_ErrorWrapping(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(fmt.Errorf("validation failed: invalid key"))

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to update session with new Claude session ID:.*validation failed: invalid key`, err.Error())
}

// --- Validation Failures — Working Directory ---

func TestAgentInvoker_WorkingDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AgentRoot = "nonexistent"

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)agent working directory not found or invalid:.*nonexistent`, err.Error())
}

func TestAgentInvoker_WorkingDirIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	// Create a regular file (not a directory)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("data"), 0644))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AgentRoot = "file.txt"

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)agent working directory not found or invalid:.*file\.txt`, err.Error())
}

// --- Validation Failures — Process Start ---

func TestAgentInvoker_ClaudeCommandNotFound(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	// Set PATH to empty to ensure claude executable is not found
	t.Setenv("PATH", "")

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to start Claude CLI process:.*executable file not found`, err.Error())
}

func TestAgentInvoker_WorkingDirNoExecutePermission(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	agentsDir := filepath.Join(tmpDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	// Remove execute permissions
	require.NoError(t, os.Chmod(agentsDir, 0000))
	t.Cleanup(func() {
		os.Chmod(agentsDir, 0755) // Restore for cleanup
	})

	sess := newMockSessionForInvoker()
	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to start Claude CLI process:.*permission denied`, err.Error())
}

// --- Validation Failures — UUID Generation ---

func TestAgentInvoker_UUIDGenerationFails(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)

	agentDef := defaultAgentDefinition()

	uuidGen := &mockUUIDGenerator{}
	uuidGen.On("Generate").Return("", fmt.Errorf("entropy source unavailable"))

	invoker, err := NewAgentInvokerWithUUIDGenerator(sess, projectRoot, uuidGen)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to generate Claude session ID`, err.Error())
}

// --- Error Propagation ---

func TestAgentInvoker_ErrorsDontCallSessionFail(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AgentRoot = "missing"

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	err = invoker.InvokeAgent("Node", "test", agentDef)

	require.Error(t, err)
	assert.False(t, sess.getFailCalled(), "Session.Fail() should not be called by AgentInvoker")
	sess.AssertNotCalled(t, "Fail", mock.Anything, mock.Anything)
}

// --- Resource Cleanup — Post-Start Failure ---

func TestAgentInvoker_PostStartFailure_TerminatesProcess_Unix(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	// Simulate post-start failure with process cleanup verification
	processTerminated, terminationMethod, err := invoker.SimulatePostStartFailure("Node", "test", agentDef, "unix")

	require.Error(t, err)
	assert.Regexp(t, `(?i)post-start validation failed`, err.Error())
	assert.True(t, processTerminated, "process should be terminated after post-start failure")
	assert.Equal(t, "sigterm_then_sigkill", terminationMethod, "Unix: should send SIGTERM, wait, then SIGKILL if needed")
}

func TestAgentInvoker_PostStartFailure_TerminatesProcess_Windows(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	// Simulate post-start failure with process cleanup verification
	processTerminated, terminationMethod, err := invoker.SimulatePostStartFailure("Node", "test", agentDef, "windows")

	require.Error(t, err)
	assert.Regexp(t, `(?i)post-start validation failed`, err.Error())
	assert.True(t, processTerminated, "process should be terminated after post-start failure")
	assert.Equal(t, "sigkill_immediate", terminationMethod, "Windows: should send SIGKILL immediately")
}

// --- Idempotency ---

func TestAgentInvoker_RepeatedInvocationExistingSession(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	existingUUID := uuid.New().String()
	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(existingUUID, true)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	// Call three times
	for i := 0; i < 3; i++ {
		err = invoker.InvokeAgent("Node", fmt.Sprintf("message %d", i), agentDef)
		assert.NoError(t, err, "invocation %d should succeed", i)
	}

	// UpdateSessionDataSafe should never be called (existing session)
	sess.AssertNotCalled(t, "UpdateSessionDataSafe", mock.Anything, mock.Anything)
}

// --- Boundary Values — Tools Arrays ---

func TestAgentInvoker_AllowedToolsSingleItem(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AllowedTools = []string{"Bash(*)"}

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	assert.Contains(t, capturedArgs, "--allowed-tools")
	assert.Contains(t, capturedArgs, "Bash(*)")
}

func TestAgentInvoker_AllowedToolsMultipleItems(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AllowedTools = []string{"Bash(*)", "Read(*)", "Edit(*)"}

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	assert.Contains(t, capturedArgs, "--allowed-tools")
	assert.Contains(t, capturedArgs, "Bash(*)")
	assert.Contains(t, capturedArgs, "Read(*)")
	assert.Contains(t, capturedArgs, "Edit(*)")
}

func TestAgentInvoker_AllowedToolsNilVsEmpty(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	// Test nil AllowedTools
	agentDefNil := defaultAgentDefinition()
	agentDefNil.AllowedTools = nil
	argsNil := invoker.CaptureCommandArgs("Node", "prompt", agentDefNil)

	// Test empty AllowedTools
	agentDefEmpty := defaultAgentDefinition()
	agentDefEmpty.AllowedTools = []string{}
	argsEmpty := invoker.CaptureCommandArgs("Node", "prompt", agentDefEmpty)

	// Both should omit --allowed-tools flag
	for _, arg := range argsNil {
		assert.NotEqual(t, "--allowed-tools", arg, "nil AllowedTools should omit --allowed-tools flag")
	}
	for _, arg := range argsEmpty {
		assert.NotEqual(t, "--allowed-tools", arg, "empty AllowedTools should omit --allowed-tools flag")
	}
}

// --- Boundary Values — Model and Effort ---

func TestAgentInvoker_ModelWithSpaces(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.Model = "sonnet 4.0"

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	assert.Contains(t, capturedArgs, "sonnet 4.0", "model with spaces should be passed safely as separate argument")
}

func TestAgentInvoker_EffortSpecialChars(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.Effort = "high-priority"

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	capturedArgs := invoker.CaptureCommandArgs("Node", "prompt", agentDef)

	assert.Contains(t, capturedArgs, "high-priority", "effort with special characters should be passed safely")
}

// --- Boundary Values — Paths ---

func TestAgentInvoker_AgentRootMultiLevel(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "logic"), 0755))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()
	agentDef.AgentRoot = "spec/logic"

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	capturedDir := invoker.CaptureCommandDir("Node", "prompt", agentDef)

	expectedDir := filepath.Join(tmpDir, "spec", "logic")
	assert.Equal(t, expectedDir, capturedDir, "multi-level relative AgentRoot should be joined correctly")
}

func TestAgentInvoker_ProjectRootAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "agents"), 0755))
	sess := newMockSessionForInvoker()

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	// Verify tmpDir is absolute
	absPath, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, tmpDir, absPath, "t.TempDir should return absolute path")

	invoker, err := NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	capturedDir := invoker.CaptureCommandDir("Node", "prompt", agentDef)

	expectedDir := filepath.Join(tmpDir, "agents")
	assert.Equal(t, expectedDir, capturedDir, "working directory should be absolute path")
	assert.True(t, filepath.IsAbs(capturedDir), "working directory must be absolute")
}

// --- Not Immutable — Session Data ---

func TestAgentInvoker_SessionDataMutated(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "NewNode.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "NewNode.ClaudeSessionID", mock.MatchedBy(func(val any) bool {
		s, ok := val.(string)
		if !ok {
			return false
		}
		_, err := uuid.Parse(s)
		return err == nil
	})).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	err = invoker.InvokeAgent("NewNode", "test", agentDef)

	require.NoError(t, err)
	sess.AssertCalled(t, "UpdateSessionDataSafe", "NewNode.ClaudeSessionID", mock.AnythingOfType("string"))

	// Verify the value is a UUID
	for _, call := range sess.Calls {
		if call.Method == "UpdateSessionDataSafe" {
			storedValue := call.Arguments.Get(1).(string)
			matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, storedValue)
			assert.True(t, matched, "stored value should be a valid UUID v4")
		}
	}
}

// --- Mock / Dependency Interaction ---

func TestAgentInvoker_CallsGetSessionDataSafe(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "TestNode.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "TestNode.ClaudeSessionID", mock.AnythingOfType("string")).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	_ = invoker.InvokeAgent("TestNode", "test", agentDef)

	sess.AssertCalled(t, "GetSessionDataSafe", "TestNode.ClaudeSessionID")
}

func TestAgentInvoker_CallsUpdateSessionDataSafeOnNewSession(t *testing.T) {
	projectRoot, sess := createAgentInvokerFixture(t)

	sess.On("GetSessionDataSafe", "Node.ClaudeSessionID").Return(nil, false)
	sess.On("UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.MatchedBy(func(val any) bool {
		s, ok := val.(string)
		if !ok {
			return false
		}
		_, err := uuid.Parse(s)
		return err == nil
	})).Return(nil)

	agentDef := defaultAgentDefinition()

	invoker, err := NewAgentInvoker(sess, projectRoot)
	require.NoError(t, err)

	_ = invoker.InvokeAgent("Node", "test", agentDef)

	sess.AssertNumberOfCalls(t, "UpdateSessionDataSafe", 1)
	sess.AssertCalled(t, "UpdateSessionDataSafe", "Node.ClaudeSessionID", mock.AnythingOfType("string"))
}
