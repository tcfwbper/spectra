package main_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra_agent "github.com/tcfwbper/spectra/cmd/spectra_agent"
	"github.com/tcfwbper/spectra/storage"
)

// --- Test Helpers ---

// setupEventEmitTestFixture creates a temporary test directory with .spectra/sessions/<uuid>/ structure.
func setupEventEmitTestFixture(t *testing.T, sessionUUID string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	return tmpDir
}

// setupEventEmitTestFixtureNoSession creates a temporary test directory with .spectra/ but no sessions.
func setupEventEmitTestFixtureNoSession(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	return tmpDir
}

// executeEventEmitCommand creates and executes the event emit subcommand with given args.
// The finder is configured to return projectRoot directly.
// Returns stdout, stderr, and exit code.
func executeEventEmitCommand(t *testing.T, projectRoot string, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	finder := spectra_agent.NewMockSpectraFinder(func(startDir string) (string, error) {
		return projectRoot, nil
	})
	cmd := spectra_agent.NewRootCommandWithFinder(finder)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()
	return stdout.String(), stderr.String(), exitCode
}

// --- Happy Path — Emit Event ---

// TestEventEmitCommand_EmitSuccessMinimal successfully emits event with only required type argument.
func TestEventEmitCommand_EmitSuccessMinimal(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "SpecCompleted", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
}

// TestEventEmitCommand_EmitSuccessWithMessage successfully emits event with optional message.
func TestEventEmitCommand_EmitSuccessWithMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "ReviewApproved", "--session-id", sessionID,
		"--message", "All checks passed",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives RuntimeMessage with payload.message
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "All checks passed", payload["message"])
}

// TestEventEmitCommand_EmitSuccessWithPayload successfully emits event with optional payload.
func TestEventEmitCommand_EmitSuccessWithPayload(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "DataProcessed", "--session-id", sessionID,
		"--payload", `{"count":42,"status":"ok"}`,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives RuntimeMessage with correct payload object
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	outerPayload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	innerPayload, ok := outerPayload["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(42), innerPayload["count"])
	assert.Equal(t, "ok", innerPayload["status"])
}

// TestEventEmitCommand_EmitSuccessWithAll successfully emits event with all optional parameters.
func TestEventEmitCommand_EmitSuccessWithAll(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	claudeSessionID := "550e8400-e29b-41d4-a716-446655440000"
	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TaskCompleted", "--session-id", sessionID,
		"--message", "Task finished",
		"--payload", `{"duration":123}`,
		"--claude-session-id", claudeSessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify all parameters transmitted correctly
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	assert.Equal(t, "event", parsed["type"])
	assert.Equal(t, claudeSessionID, parsed["claudeSessionID"])

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "TaskCompleted", payload["eventType"])
	assert.Equal(t, "Task finished", payload["message"])

	innerPayload, ok := payload["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(123), innerPayload["duration"])
}

// TestEventEmitCommand_DefaultMessage successfully emits event with default empty message.
func TestEventEmitCommand_DefaultMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives RuntimeMessage with payload.message=""
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	// Message defaults to "" — may be absent due to omitempty or present as ""
	msg, _ := payload["message"].(string)
	assert.Equal(t, "", msg)
}

// TestEventEmitCommand_DefaultPayload successfully emits event with default empty payload object.
func TestEventEmitCommand_DefaultPayload(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives RuntimeMessage with payload.payload={}
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	outerPayload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)

	innerPayload, ok := outerPayload["payload"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, innerPayload)
}

// TestEventEmitCommand_DefaultClaudeSessionID successfully emits event with default empty Claude session ID.
func TestEventEmitCommand_DefaultClaudeSessionID(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives RuntimeMessage with claudeSessionID=""
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	csid, _ := parsed["claudeSessionID"].(string)
	assert.Equal(t, "", csid)
}

// TestEventEmitCommand_IgnoresRuntimeMessage uses default success message regardless of Runtime response message.
func TestEventEmitCommand_IgnoresRuntimeMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"Custom success message"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
	assert.NotContains(t, stdout, "Custom success message")
}

// --- Happy Path — Message Format ---

// TestEventEmitCommand_ConstructsRuntimeMessage constructs correct RuntimeMessage JSON structure.
func TestEventEmitCommand_ConstructsRuntimeMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	_, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "MyEvent", "--session-id", sessionID,
		"--message", "test message",
		"--payload", `{"key":"value"}`,
		"--claude-session-id", "session-123",
	})

	assert.Equal(t, 0, exitCode)

	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	assert.Equal(t, "event", parsed["type"])
	assert.Equal(t, "session-123", parsed["claudeSessionID"])

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "MyEvent", payload["eventType"])
	assert.Equal(t, "test message", payload["message"])

	innerPayload, ok := payload["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", innerPayload["key"])
}

// TestEventEmitCommand_EmptyMessageString accepts explicitly empty message string.
func TestEventEmitCommand_EmptyMessageString(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--message", "",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	msg, _ := payload["message"].(string)
	assert.Equal(t, "", msg)
}

// TestEventEmitCommand_MessageWithSpecialChars accepts message with special characters.
func TestEventEmitCommand_MessageWithSpecialChars(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--message", `message: "quoted" <value>`,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify message is properly JSON-escaped
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err, "Message with special chars should be properly JSON-escaped")

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, `message: "quoted" <value>`, payload["message"])
}

// --- Validation Failures — Missing Event Type ---

// TestEventEmitCommand_MissingEventType returns exit code 1 when event type argument is missing.
func TestEventEmitCommand_MissingEventType(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "--session-id", uuid.New().String(),
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: event type is required`, stderr)
}

// TestEventEmitCommand_EmptyEventType returns exit code 1 when event type is empty string.
func TestEventEmitCommand_EmptyEventType(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "", "--session-id", uuid.New().String(),
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: event type is required`, stderr)
}

// --- Validation Failures — Invalid Payload ---

// TestEventEmitCommand_PayloadNotJSON returns exit code 1 when payload is invalid JSON.
func TestEventEmitCommand_PayloadNotJSON(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", uuid.New().String(),
		"--payload", "{invalid}",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --payload must be a valid JSON object, e\.g\., \{\}`, stderr)
}

// TestEventEmitCommand_PayloadPrimitiveString returns exit code 1 when payload is JSON string primitive.
func TestEventEmitCommand_PayloadPrimitiveString(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", uuid.New().String(),
		"--payload", `"string"`,
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --payload must be a valid JSON object, e\.g\., \{\}`, stderr)
}

// TestEventEmitCommand_PayloadPrimitiveNumber returns exit code 1 when payload is JSON number primitive.
func TestEventEmitCommand_PayloadPrimitiveNumber(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", uuid.New().String(),
		"--payload", "42",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --payload must be a valid JSON object, e\.g\., \{\}`, stderr)
}

// TestEventEmitCommand_PayloadPrimitiveBoolean returns exit code 1 when payload is JSON boolean primitive.
func TestEventEmitCommand_PayloadPrimitiveBoolean(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", uuid.New().String(),
		"--payload", "true",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --payload must be a valid JSON object, e\.g\., \{\}`, stderr)
}

// TestEventEmitCommand_PayloadNull returns exit code 1 when payload is JSON null.
func TestEventEmitCommand_PayloadNull(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", uuid.New().String(),
		"--payload", "null",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --payload must be a valid JSON object, e\.g\., \{\}`, stderr)
}

// TestEventEmitCommand_PayloadArray returns exit code 1 when payload is JSON array.
func TestEventEmitCommand_PayloadArray(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", uuid.New().String(),
		"--payload", "[1,2,3]",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --payload must be a valid JSON object, e\.g\., \{\}`, stderr)
}

// --- Validation Failures — Socket Errors ---

// TestEventEmitCommand_SocketNotFound returns exit code 2 when socket file does not exist.
func TestEventEmitCommand_SocketNotFound(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	// No socket file created

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: socket file not found:.*runtime\.sock`, stderr)
}

// TestEventEmitCommand_ConnectionRefused returns exit code 2 when Runtime is not listening.
func TestEventEmitCommand_ConnectionRefused(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create socket file but no server listening
	createSocketFileWithoutListener(t, socketPath)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: connection refused: Runtime is not running for session`, stderr)
}

// TestEventEmitCommand_ConnectionTimeout returns exit code 2 when connection times out.
func TestEventEmitCommand_ConnectionTimeout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a listener that delays accepting connections
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()
	// Don't accept any connections — let the client timeout

	// Use a short timeout override for testing
	_, stderr, exitCode := executeEventEmitCommandWithTimeout(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	}, 100*time.Millisecond)

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: connection timeout after`, stderr)
}

// TestEventEmitCommand_SendIOError returns exit code 2 when sending message fails.
func TestEventEmitCommand_SendIOError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that closes connection immediately after accepting
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close() // Close immediately
	}()

	time.Sleep(clientSocketStartup)

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: failed to (send message|read response):`, stderr)
}

// TestEventEmitCommand_ReadIOError returns exit code 2 when reading response fails.
func TestEventEmitCommand_ReadIOError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that closes connection after receiving message but before sending response
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		scanner := bufio.NewScanner(conn)
		scanner.Scan() // Read the message
		conn.Close()   // Close without sending response
	}()

	_, stderr, exitCode := executeEventEmitCommandWithTimeout(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	}, 100*time.Millisecond)

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: failed to read response:`, stderr)
}

// --- Validation Failures — Runtime Errors ---

// TestEventEmitCommand_RuntimeError returns exit code 3 when Runtime responds with error status.
func TestEventEmitCommand_RuntimeError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"session not found: abc-123"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: session not found: abc-123`, stderr)
}

// TestEventEmitCommand_SessionTerminated returns exit code 3 when session is already terminated.
func TestEventEmitCommand_SessionTerminated(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"session terminated: session is in 'completed' status"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: session terminated:`, stderr)
}

// TestEventEmitCommand_InvalidEventType returns exit code 3 when Runtime rejects invalid event type.
func TestEventEmitCommand_InvalidEventType(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"invalid event type 'Foo' for current workflow"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "Foo", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: invalid event type 'Foo'`, stderr)
}

// TestEventEmitCommand_ClaudeSessionIDMismatch returns exit code 3 when Claude session ID does not match.
func TestEventEmitCommand_ClaudeSessionIDMismatch(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath,
		`{"status":"error","message":"claude session ID mismatch: expected 550e8400-e29b-41d4-a716-446655440000 but got 660e8400-e29b-41d4-a716-446655440001"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--claude-session-id", "660e8400-e29b-41d4-a716-446655440001",
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: claude session ID mismatch:`, stderr)
}

// TestEventEmitCommand_InvalidClaudeSessionIDForHumanNode returns exit code 3 when Claude session ID provided for human node.
func TestEventEmitCommand_InvalidClaudeSessionIDForHumanNode(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath,
		`{"status":"error","message":"invalid claude session ID for human node: must be empty"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--claude-session-id", "550e8400-e29b-41d4-a716-446655440000",
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: invalid claude session ID for human node: must be empty`, stderr)
}

// TestEventEmitCommand_MalformedRuntimeResponse returns exit code 3 when Runtime response is malformed JSON.
func TestEventEmitCommand_MalformedRuntimeResponse(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, "{invalid\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: malformed response from Runtime:`, stderr)
}

// TestEventEmitCommand_MissingStatusField returns exit code 3 when Runtime response missing status field.
func TestEventEmitCommand_MissingStatusField(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"message":"ok"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: response missing 'status' field`, stderr)
}

// --- Boundary Values — Edge Cases ---

// TestEventEmitCommand_VeryLongEventType accepts very long event type string.
func TestEventEmitCommand_VeryLongEventType(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	longEventType := strings.Repeat("A", 1000)
	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", longEventType, "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
}

// TestEventEmitCommand_VeryLongMessage accepts very long message string.
func TestEventEmitCommand_VeryLongMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	longMessage := strings.Repeat("x", 10000)
	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--message", longMessage,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
}

// TestEventEmitCommand_EmptyPayloadObject accepts empty payload object explicitly.
func TestEventEmitCommand_EmptyPayloadObject(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--payload", "{}",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives payload.payload={}
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	outerPayload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	innerPayload, ok := outerPayload["payload"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, innerPayload)
}

// TestEventEmitCommand_ComplexPayloadObject accepts complex nested payload object.
func TestEventEmitCommand_ComplexPayloadObject(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	complexPayload := `{"data":[{"id":1,"values":[1,2,3]}],"meta":{"version":"1.0"}}`
	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--payload", complexPayload,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify complex payload transmitted correctly
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	outerPayload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	innerPayload, ok := outerPayload["payload"].(map[string]any)
	require.True(t, ok)

	data, ok := innerPayload["data"].([]any)
	require.True(t, ok)
	require.Len(t, data, 1)

	meta, ok := innerPayload["meta"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "1.0", meta["version"])
}

// TestEventEmitCommand_EmptyClaudeSessionID accepts empty Claude session ID explicitly.
func TestEventEmitCommand_EmptyClaudeSessionID(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--claude-session-id", "",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify mock server receives claudeSessionID=""
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	csid, _ := parsed["claudeSessionID"].(string)
	assert.Equal(t, "", csid)
}

// TestEventEmitCommand_EventTypeWithSpecialChars accepts event type with special characters.
func TestEventEmitCommand_EventTypeWithSpecialChars(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "Event-Type_123", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
}

// TestEventEmitCommand_WhitespaceInMessage accepts message with leading/trailing whitespace.
func TestEventEmitCommand_WhitespaceInMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
		"--message", "  test message  ",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify whitespace preserved
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "  test message  ", payload["message"])
}

// --- Mock / Dependency Interaction ---

// TestEventEmitCommand_UsesSocketClient delegates socket communication to SocketClient.Send.
func TestEventEmitCommand_UsesSocketClient(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)

	mockClient := spectra_agent.NewCapturingMockSocketClient()

	// Execute with injected mock SocketClient
	stdout, _, exitCode := executeEventEmitCommandWithMockClient(t, projectRoot, []string{
		"event", "emit", "MyEvent", "--session-id", sessionID,
		"--message", "test",
		"--payload", `{"key":"value"}`,
	}, mockClient)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")

	// Verify SocketClient.Send called once with correct parameters
	assert.True(t, mockClient.WasCalled())
	assert.Equal(t, 1, mockClient.CallCount())
	assert.Equal(t, sessionID, mockClient.SessionID())
	assert.Equal(t, projectRoot, mockClient.ProjectRoot())

	msg := mockClient.Message()
	assert.Equal(t, "event", msg.Type)

	var payload map[string]any
	err := json.Unmarshal(msg.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "MyEvent", payload["eventType"])
	assert.Equal(t, "test", payload["message"])

	innerPayload, ok := payload["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", innerPayload["key"])
}

// TestEventEmitCommand_ReceivesRootCommandContext receives session ID and project root from root command.
func TestEventEmitCommand_ReceivesRootCommandContext(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	// Event emit subcommand receives session ID and project root from root command initialization
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
}

// --- Idempotency ---

// TestEventEmitCommand_RepeatedInvocation multiple invocations with same arguments produce consistent results.
func TestEventEmitCommand_RepeatedInvocation(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	for i := 0; i < 3; i++ {
		stdout, _, exitCode := executeEventEmitCommand(t, projectRoot, []string{
			"event", "emit", "TestEvent", "--session-id", sessionID,
		})

		assert.Equal(t, 0, exitCode, "Invocation %d should return exit code 0", i)
		assert.Contains(t, stdout, "Event emitted successfully", "Invocation %d should produce same output", i)
	}
}

// --- Error Output Format ---

// TestEventEmitCommand_ErrorPrefixFormat all error messages are prefixed with "Error: ".
func TestEventEmitCommand_ErrorPrefixFormat(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	_, stderr, _ := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "--session-id", uuid.New().String(),
	})

	assert.Regexp(t, `^Error: `, stderr)
}

// TestEventEmitCommand_ErrorOutputToStderr error messages printed to stderr, not stdout.
func TestEventEmitCommand_ErrorOutputToStderr(t *testing.T) {
	projectRoot := setupEventEmitTestFixtureNoSession(t)

	stdout, stderr, _ := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "--session-id", uuid.New().String(),
	})

	assert.NotEmpty(t, stderr)
	assert.Empty(t, stdout)
}

// TestEventEmitCommand_SuccessOutputToStdout success message printed to stdout, not stderr.
func TestEventEmitCommand_SuccessOutputToStdout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupEventEmitTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, stderr, exitCode := executeEventEmitCommand(t, projectRoot, []string{
		"event", "emit", "TestEvent", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Event emitted successfully")
	assert.Empty(t, stderr)
}

// --- Helper: executeEventEmitCommandWithTimeout ---

// executeEventEmitCommandWithTimeout executes the event emit command with a custom SocketClient timeout.
func executeEventEmitCommandWithTimeout(t *testing.T, projectRoot string, args []string, timeout time.Duration) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	finder := spectra_agent.NewMockSpectraFinder(func(startDir string) (string, error) {
		return projectRoot, nil
	})
	cmd := spectra_agent.NewRootCommandWithFinderAndHandlers(
		finder,
		spectra_agent.WithSocketClientTimeout(timeout),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()
	return stdout.String(), stderr.String(), exitCode
}

// --- Helper: executeEventEmitCommandWithMockClient ---

// executeEventEmitCommandWithMockClient executes the event emit command with an injected mock SocketClient.
func executeEventEmitCommandWithMockClient(t *testing.T, projectRoot string, args []string, mockClient *spectra_agent.CapturingMockSocketClient) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	finder := spectra_agent.NewMockSpectraFinder(func(startDir string) (string, error) {
		return projectRoot, nil
	})
	cmd := spectra_agent.NewRootCommandWithFinderAndHandlers(
		finder,
		spectra_agent.WithMockSocketClient(mockClient),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()
	return stdout.String(), stderr.String(), exitCode
}
