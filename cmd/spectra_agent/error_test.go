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

// setupErrorTestFixture creates a temporary test directory with .spectra/sessions/<uuid>/ structure.
func setupErrorTestFixture(t *testing.T, sessionUUID string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	return tmpDir
}

// setupErrorTestFixtureNoSession creates a temporary test directory with .spectra/ but no sessions.
func setupErrorTestFixtureNoSession(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	return tmpDir
}

// executeErrorCommand creates and executes the error subcommand with given args.
// The finder is configured to return projectRoot directly.
// Returns stdout, stderr, and exit code.
func executeErrorCommand(t *testing.T, projectRoot string, args []string) (string, string, int) {
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

// --- Happy Path — Report Error ---

// TestErrorCommand_ReportSuccessWithMessage successfully reports error with message and default detail.
func TestErrorCommand_ReportSuccessWithMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{"error", "test error message", "--session-id", sessionID})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")
}

// TestErrorCommand_ReportSuccessWithDetail successfully reports error with message and detail object.
func TestErrorCommand_ReportSuccessWithDetail(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
		"--detail", `{"stack":"...","code":500}`,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server received correct RuntimeMessage with detail object
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "error", parsed["type"])

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	detail, ok := payload["detail"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "...", detail["stack"])
	assert.Equal(t, float64(500), detail["code"])
}

// TestErrorCommand_ReportSuccessWithNullDetail successfully reports error with null detail.
func TestErrorCommand_ReportSuccessWithNullDetail(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
		"--detail", "null",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server receives RuntimeMessage with detail: null
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, payload["detail"])
}

// TestErrorCommand_ReportSuccessDefaultDetail successfully reports error with default empty object detail.
func TestErrorCommand_ReportSuccessDefaultDetail(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server receives RuntimeMessage with detail: {}
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	detail, ok := payload["detail"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, detail)
}

// TestErrorCommand_WithClaudeSessionID successfully reports error with Claude session ID.
func TestErrorCommand_WithClaudeSessionID(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	claudeSessionID := "550e8400-e29b-41d4-a716-446655440000"
	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
		"--claude-session-id", claudeSessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server receives RuntimeMessage with claudeSessionID
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)
	assert.Equal(t, claudeSessionID, parsed["claudeSessionID"])
}

// TestErrorCommand_DefaultClaudeSessionID successfully reports error with default empty Claude session ID.
func TestErrorCommand_DefaultClaudeSessionID(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server receives RuntimeMessage with claudeSessionID: ""
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)
	// When claudeSessionID is empty and omitempty is used, field may be absent
	csid, _ := parsed["claudeSessionID"].(string)
	assert.Equal(t, "", csid)
}

// TestErrorCommand_IgnoresRuntimeMessage uses default success message regardless of Runtime response message.
func TestErrorCommand_IgnoresRuntimeMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"Session marked as failed"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")
	assert.NotContains(t, stdout, "Session marked as failed")
}

// --- Happy Path — Message Format ---

// TestErrorCommand_ConstructsRuntimeMessage constructs correct RuntimeMessage JSON structure.
func TestErrorCommand_ConstructsRuntimeMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	_, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
		"--claude-session-id", "session-123",
		"--detail", `{"key":"value"}`,
	})

	assert.Equal(t, 0, exitCode)

	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	assert.Equal(t, "error", parsed["type"])
	assert.Equal(t, "session-123", parsed["claudeSessionID"])

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test error", payload["message"])

	detail, ok := payload["detail"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", detail["key"])
}

// TestErrorCommand_MessageWithWhitespace accepts message with whitespace and sends successfully.
func TestErrorCommand_MessageWithWhitespace(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "   test error   ", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify whitespace preserved
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "   test error   ", payload["message"])
}

// TestErrorCommand_MessageWithSpecialChars accepts message with special characters.
func TestErrorCommand_MessageWithSpecialChars(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", `error: "critical" <failure>`, "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify message is properly JSON-escaped
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err, "Message with special chars should be properly JSON-escaped")

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, `error: "critical" <failure>`, payload["message"])
}

// --- Validation Failures — Missing Message ---

// TestErrorCommand_MissingMessage returns exit code 1 when message argument is missing.
func TestErrorCommand_MissingMessage(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "--session-id", uuid.New().String(),
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: error message is required`, stderr)
}

// TestErrorCommand_EmptyMessage returns exit code 1 when message is empty string.
func TestErrorCommand_EmptyMessage(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "", "--session-id", uuid.New().String(),
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: error message is required`, stderr)
}

// --- Validation Failures — Invalid Detail ---

// TestErrorCommand_DetailNotJSON returns exit code 1 when detail is invalid JSON.
func TestErrorCommand_DetailNotJSON(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", uuid.New().String(),
		"--detail", "{invalid}",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --detail must be a JSON object or null`, stderr)
}

// TestErrorCommand_DetailPrimitiveString returns exit code 1 when detail is JSON string primitive.
func TestErrorCommand_DetailPrimitiveString(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", uuid.New().String(),
		"--detail", `"error detail string"`,
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --detail must be a JSON object or null`, stderr)
}

// TestErrorCommand_DetailPrimitiveNumber returns exit code 1 when detail is JSON number primitive.
func TestErrorCommand_DetailPrimitiveNumber(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", uuid.New().String(),
		"--detail", "42",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --detail must be a JSON object or null`, stderr)
}

// TestErrorCommand_DetailPrimitiveBoolean returns exit code 1 when detail is JSON boolean primitive.
func TestErrorCommand_DetailPrimitiveBoolean(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", uuid.New().String(),
		"--detail", "true",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --detail must be a JSON object or null`, stderr)
}

// TestErrorCommand_DetailArray returns exit code 1 when detail is JSON array.
func TestErrorCommand_DetailArray(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", uuid.New().String(),
		"--detail", "[1,2,3]",
	})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --detail must be a JSON object or null`, stderr)
}

// --- Validation Failures — Socket Errors ---

// TestErrorCommand_SocketNotFound returns exit code 2 when socket file does not exist.
func TestErrorCommand_SocketNotFound(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	// No socket file created

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: socket file not found:.*runtime\.sock`, stderr)
}

// TestErrorCommand_ConnectionRefused returns exit code 2 when Runtime is not listening.
func TestErrorCommand_ConnectionRefused(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create socket file but no server listening
	createSocketFileWithoutListener(t, socketPath)

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: connection refused: Runtime is not running for session`, stderr)
}

// TestErrorCommand_ConnectionTimeout returns exit code 2 when connection times out.
func TestErrorCommand_ConnectionTimeout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a listener that delays accepting connections
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()
	// Don't accept any connections — let the client timeout

	// Use a short timeout override for testing
	_, stderr, exitCode := executeErrorCommandWithTimeout(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	}, 100*time.Millisecond)

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: connection timeout after`, stderr)
}

// TestErrorCommand_SendIOError returns exit code 2 when sending message fails.
func TestErrorCommand_SendIOError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
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

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: failed to (send message|read response):`, stderr)
}

// TestErrorCommand_ReadIOError returns exit code 2 when reading response fails.
func TestErrorCommand_ReadIOError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
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

	_, stderr, exitCode := executeErrorCommandWithTimeout(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	}, 100*time.Millisecond)

	assert.Equal(t, 2, exitCode)
	assert.Regexp(t, `Error: failed to read response:`, stderr)
}

// --- Validation Failures — Runtime Errors ---

// TestErrorCommand_RuntimeError returns exit code 3 when Runtime responds with error status.
func TestErrorCommand_RuntimeError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"session not found: abc-123"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: session not found: abc-123`, stderr)
}

// TestErrorCommand_SessionTerminated returns exit code 3 when session is already terminated.
func TestErrorCommand_SessionTerminated(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"session terminated: session is in 'completed' status"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: session terminated:`, stderr)
}

// TestErrorCommand_ClaudeSessionIDMismatch returns exit code 3 when Claude session ID does not match.
func TestErrorCommand_ClaudeSessionIDMismatch(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath,
		`{"status":"error","message":"claude session ID mismatch: expected 550e8400-e29b-41d4-a716-446655440000 but got 660e8400-e29b-41d4-a716-446655440001"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
		"--claude-session-id", "660e8400-e29b-41d4-a716-446655440001",
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: claude session ID mismatch:`, stderr)
}

// TestErrorCommand_InvalidClaudeSessionIDForHumanNode returns exit code 3 when Claude session ID provided for human node.
func TestErrorCommand_InvalidClaudeSessionIDForHumanNode(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath,
		`{"status":"error","message":"invalid claude session ID for human node: must be empty"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
		"--claude-session-id", "550e8400-e29b-41d4-a716-446655440000",
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: invalid claude session ID for human node: must be empty`, stderr)
}

// TestErrorCommand_MalformedRuntimeResponse returns exit code 3 when Runtime response is malformed JSON.
func TestErrorCommand_MalformedRuntimeResponse(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, "{invalid\n")
	defer cleanup()

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: malformed response from Runtime:`, stderr)
}

// TestErrorCommand_MissingStatusField returns exit code 3 when Runtime response missing status field.
func TestErrorCommand_MissingStatusField(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"message":"ok"}`+"\n")
	defer cleanup()

	_, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 3, exitCode)
	assert.Regexp(t, `Error: response missing 'status' field`, stderr)
}

// --- Boundary Values — Edge Cases ---

// TestErrorCommand_WhitespaceOnlyMessage accepts whitespace-only message and sends successfully.
func TestErrorCommand_WhitespaceOnlyMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "   ", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify whitespace-only message sent
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "   ", payload["message"])
}

// TestErrorCommand_VeryLongMessage accepts very long error message.
func TestErrorCommand_VeryLongMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	longMessage := strings.Repeat("x", 10000)
	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", longMessage, "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")
}

// TestErrorCommand_EmptyDetailObject accepts empty detail object explicitly.
func TestErrorCommand_EmptyDetailObject(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
		"--detail", "{}",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server receives detail: {}
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	detail, ok := payload["detail"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, detail)
}

// TestErrorCommand_ComplexDetailObject accepts complex nested detail object.
func TestErrorCommand_ComplexDetailObject(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	complexDetail := `{"stack":[{"file":"a.go","line":10}],"context":{"user":"test"}}`
	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
		"--detail", complexDetail,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify complex detail transmitted correctly
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	payload, ok := parsed["payload"].(map[string]any)
	require.True(t, ok)
	detail, ok := payload["detail"].(map[string]any)
	require.True(t, ok)

	stack, ok := detail["stack"].([]any)
	require.True(t, ok)
	require.Len(t, stack, 1)

	context, ok := detail["context"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test", context["user"])
}

// TestErrorCommand_EmptyClaudeSessionID accepts empty Claude session ID explicitly.
func TestErrorCommand_EmptyClaudeSessionID(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
		"--claude-session-id", "",
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify mock server receives claudeSessionID: ""
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)

	csid, _ := parsed["claudeSessionID"].(string)
	assert.Equal(t, "", csid)
}

// --- Mock / Dependency Interaction ---

// TestErrorCommand_UsesSocketClient delegates socket communication to SocketClient.Send.
func TestErrorCommand_UsesSocketClient(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)

	mockClient := spectra_agent.NewCapturingMockSocketClient()

	// Execute with injected mock SocketClient
	stdout, _, exitCode := executeErrorCommandWithMockClient(t, projectRoot, []string{
		"error", "test error", "--session-id", sessionID,
		"--detail", `{"key":"value"}`,
	}, mockClient)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")

	// Verify SocketClient.Send called once with correct parameters
	assert.True(t, mockClient.WasCalled())
	assert.Equal(t, 1, mockClient.CallCount())
	assert.Equal(t, sessionID, mockClient.SessionID())
	assert.Equal(t, projectRoot, mockClient.ProjectRoot())

	msg := mockClient.Message()
	assert.Equal(t, "error", msg.Type)

	var payload map[string]any
	err := json.Unmarshal(msg.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "test error", payload["message"])

	detail, ok := payload["detail"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", detail["key"])
}

// TestErrorCommand_ReceivesRootCommandContext receives session ID and project root from root command.
func TestErrorCommand_ReceivesRootCommandContext(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	// Error subcommand receives session ID and project root from root command initialization
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")
}

// --- Idempotency ---

// TestErrorCommand_RepeatedInvocation multiple invocations with same arguments produce consistent results.
func TestErrorCommand_RepeatedInvocation(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	for i := 0; i < 3; i++ {
		stdout, _, exitCode := executeErrorCommand(t, projectRoot, []string{
			"error", "test error", "--session-id", sessionID,
		})

		assert.Equal(t, 0, exitCode, "Invocation %d should return exit code 0", i)
		assert.Contains(t, stdout, "Error reported successfully", "Invocation %d should produce same output", i)
	}
}

// --- Error Output Format ---

// TestErrorCommand_ErrorPrefixFormat all error messages are prefixed with "Error: ".
func TestErrorCommand_ErrorPrefixFormat(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	_, stderr, _ := executeErrorCommand(t, projectRoot, []string{
		"error", "--session-id", uuid.New().String(),
	})

	assert.Regexp(t, `^Error: `, stderr)
}

// TestErrorCommand_ErrorOutputToStderr error messages printed to stderr, not stdout.
func TestErrorCommand_ErrorOutputToStderr(t *testing.T) {
	projectRoot := setupErrorTestFixtureNoSession(t)

	stdout, stderr, _ := executeErrorCommand(t, projectRoot, []string{
		"error", "--session-id", uuid.New().String(),
	})

	assert.NotEmpty(t, stderr)
	assert.Empty(t, stdout)
}

// TestErrorCommand_SuccessOutputToStdout success message printed to stdout, not stderr.
func TestErrorCommand_SuccessOutputToStdout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupErrorTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	stdout, stderr, exitCode := executeErrorCommand(t, projectRoot, []string{
		"error", "test", "--session-id", sessionID,
	})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Error reported successfully")
	assert.Empty(t, stderr)
}

// --- Helper: executeErrorCommandWithTimeout ---

// executeErrorCommandWithTimeout executes the error command with a custom SocketClient timeout.
func executeErrorCommandWithTimeout(t *testing.T, projectRoot string, args []string, timeout time.Duration) (string, string, int) {
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

// --- Helper: executeErrorCommandWithMockClient ---

// executeErrorCommandWithMockClient executes the error command with an injected mock SocketClient.
func executeErrorCommandWithMockClient(t *testing.T, projectRoot string, args []string, mockClient *spectra_agent.CapturingMockSocketClient) (string, string, int) {
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
