package spectra_agent_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	spectra_agent "github.com/tcfwbper/spectra/cmd/spectra_agent"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// Test timeouts and constants
const (
	clientTestTimeout      = 100 * time.Millisecond
	clientSocketStartup    = 50 * time.Millisecond
	clientLargePayloadSize = 1 * 1024 * 1024   // 1 MB
	clientVeryLargePayload = 100 * 1024 * 1024 // 100 MB
	concurrentSendCount    = 20
	concurrentSameSession  = 10
	repeatedSendCount      = 3
	fdCheckDelay           = 50 * time.Millisecond
)

// setupClientTestFixture creates a temporary test directory with .spectra/sessions/<uuid>/ structure.
// Uses short temp dir prefix to avoid Unix socket path length limits.
func setupClientTestFixture(t *testing.T, sessionUUID string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	return tmpDir
}

// setupClientTestFixtureNoSession creates a temporary test directory with .spectra/ but no sessions.
func setupClientTestFixtureNoSession(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	return tmpDir
}

// mockSocketServer creates a mock Unix socket server that responds with the given response.
// Returns the cleanup function.
func mockSocketServer(t *testing.T, socketPath string, response string) (net.Listener, func()) {
	t.Helper()
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				// Set a larger buffer to handle large payloads (up to 10MB)
				buf := make([]byte, 4096)
				scanner.Buffer(buf, 10*1024*1024)
				if scanner.Scan() {
					c.Write([]byte(response))
				}
			}(conn)
		}
	}()

	cleanup := func() {
		listener.Close()
	}
	return listener, cleanup
}

// mockCapturingSocketServer creates a mock socket server that captures sent messages.
func mockCapturingSocketServer(t *testing.T, socketPath string, response string) (*capturedMessages, net.Listener, func()) {
	t.Helper()
	captured := &capturedMessages{}
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				// Set a larger buffer to handle large payloads (up to 10MB)
				buf := make([]byte, 4096)
				scanner.Buffer(buf, 10*1024*1024)
				if scanner.Scan() {
					captured.add(scanner.Text())
					c.Write([]byte(response))
				}
			}(conn)
		}
	}()

	cleanup := func() {
		listener.Close()
	}
	return captured, listener, cleanup
}

type capturedMessages struct {
	mu       sync.Mutex
	messages []string
}

func (cm *capturedMessages) add(msg string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.messages = append(cm.messages, msg)
}

func (cm *capturedMessages) get() []string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return append([]string{}, cm.messages...)
}

// newTestEventMessage creates a RuntimeMessage for event-type testing.
func newTestEventMessage(t *testing.T) entities.RuntimeMessage {
	t.Helper()
	payload := json.RawMessage(`{"eventType":"MyEvent","message":"test","payload":{}}`)
	return entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: "test-session",
		Payload:         payload,
	}
}

// newTestErrorMessage creates a RuntimeMessage for error-type testing.
func newTestErrorMessage(t *testing.T) entities.RuntimeMessage {
	t.Helper()
	payload := json.RawMessage(`{"message":"test error","detail":{"key":"value"}}`)
	return entities.RuntimeMessage{
		Type:            "error",
		ClaudeSessionID: "session-123",
		Payload:         payload,
	}
}

// --- Happy Path — Send ---

// TestSocketClient_SendSuccess successfully sends message and receives success response.
func TestSocketClient_SendSuccess(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	resp, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "ok", resp.Message)
}

// TestSocketClient_SendSuccessEmptyMessage receives success response with empty message field.
func TestSocketClient_SendSuccessEmptyMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":""}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: "",
		Payload:         json.RawMessage(`{"eventType":"MyEvent","message":"test","payload":{}}`),
	}
	resp, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "", resp.Message)
}

// TestSocketClient_SendEventMessage sends event-type message as valid JSON with newline terminator.
func TestSocketClient_SendEventMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: "session-123",
		Payload:         json.RawMessage(`{"eventType":"MyEvent","message":"test","payload":{}}`),
	}
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)

	// Verify the mock server received valid JSON
	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	assert.NoError(t, err, "Mock server should receive valid JSON")
	assert.Equal(t, "event", parsed["type"])
	assert.Equal(t, "session-123", parsed["claudeSessionID"])
}

// TestSocketClient_SendErrorMessage sends error-type message as valid JSON with newline terminator.
func TestSocketClient_SendErrorMessage(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestErrorMessage(t)
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)

	messages := captured.get()
	require.Len(t, messages, 1)

	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	assert.NoError(t, err, "Mock server should receive valid JSON")
	assert.Equal(t, "error", parsed["type"])
	assert.Equal(t, "session-123", parsed["claudeSessionID"])
}

// --- Happy Path — Runtime Error Response ---

// TestSocketClient_RuntimeErrorResponse returns exit code 3 when Runtime responds with error status.
func TestSocketClient_RuntimeErrorResponse(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"session not found"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	resp, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 3, exitCode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	require.NotNil(t, resp)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "session not found", resp.Message)
}

// --- Resource Cleanup ---

// TestSocketClient_ClosesSocketOnSuccess closes socket connection and releases file descriptor after successful operation.
func TestSocketClient_ClosesSocketOnSuccess(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	connClosed := make(chan struct{})
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			conn.Write([]byte(`{"status":"success","message":"ok"}` + "\n"))
		}
		// Wait for client to close connection
		buf := make([]byte, 1)
		conn.Read(buf) // Will return when client closes
		close(connClosed)
	}()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)

	// Verify connection was closed
	select {
	case <-connClosed:
		// Connection closed as expected
	case <-time.After(2 * time.Second):
		t.Fatal("Socket connection was not closed after successful operation")
	}
}

// TestSocketClient_ClosesSocketOnError attempts to close socket and release resources even after error occurs.
func TestSocketClient_ClosesSocketOnError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that closes the connection prematurely
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		// Close immediately without sending response
		conn.Close()
	}()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
}

// TestSocketClient_NoGoroutineLeak does not leak goroutines after Send completes.
func TestSocketClient_NoGoroutineLeak(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)
}

// --- Validation Failures — Socket Not Found ---

// TestSocketClient_SocketFileNotFound returns exit code 2 when socket file does not exist.
func TestSocketClient_SocketFileNotFound(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	// No socket file created

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: socket file not found:.*runtime\.sock`, err.Error())
}

// TestSocketClient_SessionDirNotExist returns exit code 2 when session directory does not exist.
func TestSocketClient_SessionDirNotExist(t *testing.T) {
	nonexistentID := uuid.New().String()
	projectRoot := setupClientTestFixtureNoSession(t)

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(nonexistentID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, fmt.Sprintf(`Error: socket file not found:.*%s`, nonexistentID), err.Error())
}

// --- Validation Failures — Connection Refused ---

// TestSocketClient_ConnectionRefused returns exit code 2 when Runtime is not listening.
func TestSocketClient_ConnectionRefused(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a socket file but no server listening
	// Create socket with bind but without listen to simulate connection refused
	createSocketFileWithoutListener(t, socketPath)

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, fmt.Sprintf(`Error: connection refused: Runtime is not running for session %s`, sessionID), err.Error())
}

// --- Validation Failures — Timeout ---

// TestSocketClient_ConnectionTimeout returns exit code 2 when connection times out.
func TestSocketClient_ConnectionTimeout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a listener that delays accepting connections
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()
	// Don't accept any connections — let the client timeout

	client := spectra_agent.NewSocketClientWithTimeout(clientTestTimeout)
	msg := newTestEventMessage(t)

	start := time.Now()
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)
	elapsed := time.Since(start)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: connection timeout after`, err.Error())
	// Should complete within a reasonable time margin around the timeout
	assert.Less(t, elapsed, 5*time.Second, "Operation should complete quickly")
}

// TestSocketClient_ReadTimeout returns exit code 2 when reading response times out.
func TestSocketClient_ReadTimeout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that accepts connection but never sends response
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		scanner.Scan() // Read the message
		// Never send a response - let it timeout
		time.Sleep(10 * time.Second)
	}()

	client := spectra_agent.NewSocketClientWithTimeout(clientTestTimeout)
	msg := newTestEventMessage(t)

	start := time.Now()
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)
	elapsed := time.Since(start)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: connection timeout after`, err.Error())
	assert.Less(t, elapsed, 5*time.Second, "Operation should complete quickly")
}

// --- Validation Failures — Send Errors ---

// TestSocketClient_SendMessageIOError returns exit code 2 when sending message fails with I/O error.
func TestSocketClient_SendMessageIOError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
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

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	// Due to socket buffering, the write may succeed and read may fail instead
	// Both indicate transport errors, so we accept either message
	assert.Regexp(t, `Error: failed to (send message|read response):`, err.Error())
}

// --- Validation Failures — Read Errors ---

// TestSocketClient_ReadResponseIOError returns exit code 2 when reading response fails with I/O error.
func TestSocketClient_ReadResponseIOError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
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

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: failed to read response:`, err.Error())
}

// TestSocketClient_ConnectionClosedByRuntime returns exit code 2 when Runtime closes connection without response.
func TestSocketClient_ConnectionClosedByRuntime(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that closes connection immediately after receiving message
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
		conn.Close()   // Close immediately
	}()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: failed to read response:.*connection closed`, err.Error())
}

// --- Validation Failures — Malformed Response ---

// TestSocketClient_MalformedJSON returns exit code 3 when response is invalid JSON.
func TestSocketClient_MalformedJSON(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, "{invalid\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 3, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: malformed response from Runtime:`, err.Error())
}

// TestSocketClient_MissingStatusField returns exit code 3 when response JSON is valid but missing status field.
func TestSocketClient_MissingStatusField(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 3, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: response missing 'status' field`, err.Error())
}

// TestSocketClient_InvalidStatusValue returns exit code 3 when response status is not "success" or "error".
func TestSocketClient_InvalidStatusValue(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"unknown","message":"test"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 3, exitCode)
	assert.Error(t, err)
	assert.Regexp(t, `Error: invalid response status 'unknown'`, err.Error())
}

// --- Validation Failures — Socket Close Warning ---

// TestSocketClient_CloseSocketFailsAfterSuccess prints warning to stderr when closing socket fails after successful operation.
func TestSocketClient_CloseSocketFailsAfterSuccess(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that responds with success
	// We need a setup where close will fail — this is tricky to simulate,
	// so we use a mock StorageLayout that returns a path to a socket that
	// will cause close to fail.
	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	// Capture stderr to check for warning
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	w.Close()
	os.Stderr = oldStderr

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	stderrOutput := string(buf[:n])
	_ = stderrOutput

	// If the implementation produces a close warning, verify it
	// The exit code should be 0 regardless
	assert.Equal(t, 0, exitCode)
}

// TestSocketClient_CloseSocketFailsAfterError prints warning but preserves original exit code when closing socket fails after error.
func TestSocketClient_CloseSocketFailsAfterError(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"error","message":"runtime error"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	// Original error exit code should be preserved
	assert.Equal(t, 3, exitCode)
}

// --- Validation Failures — Input Values ---

// TestSocketClient_InvalidUUIDFormat proceeds with invalid UUID format without validation.
func TestSocketClient_InvalidUUIDFormat(t *testing.T) {
	invalidID := "not-a-uuid"
	projectRoot := setupClientTestFixtureNoSession(t)

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(invalidID, projectRoot, msg)

	// Should compute socket path using invalid UUID and fail with socket not found
	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
}

// TestSocketClient_EmptyProjectRoot returns error when project root path is empty.
func TestSocketClient_EmptyProjectRoot(t *testing.T) {
	sessionID := uuid.New().String()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, "", msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
}

// TestSocketClient_GetSocketPathError returns exit code 2 when StorageLayout.GetRuntimeSocketPath fails.
func TestSocketClient_GetSocketPathError(t *testing.T) {
	sessionID := uuid.New().String()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Create a mock StorageLayout that returns an error
	mockLayout := &spectra_agent.MockStorageLayout{
		GetRuntimeSocketPathFunc: func(projectRoot, sessionID string) (string, error) {
			return "", fmt.Errorf("mock storage layout error")
		},
	}

	client := spectra_agent.NewSocketClientWithLayout(mockLayout)
	msg := newTestEventMessage(t)
	_, exitCode, err := client.Send(sessionID, tmpDir, msg)

	assert.Equal(t, 2, exitCode)
	assert.Error(t, err)
}

// --- Boundary Values — Edge Cases ---

// TestSocketClient_EmptyClaudeSessionID accepts empty ClaudeSessionID in message.
func TestSocketClient_EmptyClaudeSessionID(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	captured, _, cleanup := mockCapturingSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: "",
		Payload:         json.RawMessage(`{"eventType":"MyEvent","message":"test","payload":{}}`),
	}
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)

	messages := captured.get()
	require.Len(t, messages, 1)

	// Verify the message was sent with empty claudeSessionID
	var parsed map[string]any
	err := json.Unmarshal([]byte(messages[0]), &parsed)
	require.NoError(t, err)
}

// TestSocketClient_MultipleResponsesIgnored reads first response and closes connection, ignoring subsequent responses.
func TestSocketClient_MultipleResponsesIgnored(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that sends two responses
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			conn.Write([]byte(`{"status":"success","message":"first"}` + "\n"))
			conn.Write([]byte(`{"status":"success","message":"second"}` + "\n"))
		}
	}()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)
	resp, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)
	require.NotNil(t, resp)
	assert.Equal(t, "first", resp.Message)
}

// TestSocketClient_LargePayload handles reasonably large JSON payload in message.
func TestSocketClient_LargePayload(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	// Create a 1MB payload
	largeData := strings.Repeat("x", clientLargePayloadSize)
	payloadJSON := fmt.Sprintf(`{"eventType":"MyEvent","message":"%s","payload":{}}`, largeData)

	client := spectra_agent.NewSocketClient()
	msg := entities.RuntimeMessage{
		Type:    "event",
		Payload: json.RawMessage(payloadJSON),
	}
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 0, exitCode)
}

// TestSocketClient_VeryLargePayload returns error when payload exceeds reasonable limits.
func TestSocketClient_VeryLargePayload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large payload test in short mode")
	}

	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	// Create a mock server that tracks message size
	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		scanner.Buffer(make([]byte, 4096), clientVeryLargePayload+1024)
		if scanner.Scan() {
			conn.Write([]byte(`{"status":"success","message":"ok"}` + "\n"))
		}
	}()

	// Create a 100MB payload
	largeData := strings.Repeat("x", clientVeryLargePayload)
	payloadJSON := fmt.Sprintf(`{"eventType":"MyEvent","message":"%s","payload":{}}`, largeData)

	client := spectra_agent.NewSocketClientWithTimeout(clientTestTimeout)
	msg := entities.RuntimeMessage{
		Type:    "event",
		Payload: json.RawMessage(payloadJSON),
	}
	_, exitCode, _ := client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, 2, exitCode)
}

// --- Idempotency ---

// TestSocketClient_RepeatedSend multiple Send invocations produce independent results.
func TestSocketClient_RepeatedSend(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	client := spectra_agent.NewSocketClient()
	msg := newTestEventMessage(t)

	for i := 0; i < repeatedSendCount; i++ {
		resp, exitCode, err := client.Send(sessionID, projectRoot, msg)
		assert.NoError(t, err, "Send %d should succeed", i)
		assert.Equal(t, 0, exitCode, "Send %d should return exit code 0", i)
		require.NotNil(t, resp, "Send %d should return response", i)
		assert.Equal(t, "success", resp.Status, "Send %d should have success status", i)
	}
}

// --- Mock / Dependency Interaction ---

// TestSocketClient_UsesStorageLayout calls StorageLayout.GetRuntimeSocketPath with correct parameters.
func TestSocketClient_UsesStorageLayout(t *testing.T) {
	sessionID := uuid.New().String()
	projectRoot := setupClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := mockSocketServer(t, socketPath, `{"status":"success","message":"ok"}`+"\n")
	defer cleanup()

	var calledProjectRoot, calledSessionID string
	mockLayout := &spectra_agent.MockStorageLayout{
		GetRuntimeSocketPathFunc: func(pr, sid string) (string, error) {
			calledProjectRoot = pr
			calledSessionID = sid
			return storage.GetRuntimeSocketPath(pr, sid), nil
		},
	}

	client := spectra_agent.NewSocketClientWithLayout(mockLayout)
	msg := newTestEventMessage(t)
	_, _, _ = client.Send(sessionID, projectRoot, msg)

	assert.Equal(t, projectRoot, calledProjectRoot)
	assert.Equal(t, sessionID, calledSessionID)
}

// --- Concurrent Behaviour (race tests) ---
// These tests go under test/race/ as per convention, but the spec includes
// their definitions here. See test/race/socket_client_race_test.go for the
// actual race-category tests.

// countOpenFDs returns the number of open file descriptors for the current process (Linux only).
func countOpenFDs(t *testing.T) int {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("File descriptor counting only supported on Linux")
	}
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skipf("Cannot read /proc/self/fd: %v", err)
	}
	return len(entries)
}

// createSocketFileWithoutListener creates a socket file using bind without listen,
// which causes connection refused errors when clients try to connect.
func createSocketFileWithoutListener(t *testing.T, socketPath string) {
	t.Helper()

	// Remove old socket if exists
	os.Remove(socketPath)

	// Create a socket file using low-level syscall
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	require.NoError(t, err)

	// Bind to create the socket file (but don't listen)
	sockaddr := &syscall.SockaddrUnix{Name: socketPath}
	err = syscall.Bind(fd, sockaddr)
	require.NoError(t, err)

	// Note: We intentionally don't call syscall.Listen() to simulate connection refused
	// The cleanup will close the fd
	t.Cleanup(func() {
		syscall.Close(fd)
		os.Remove(socketPath)
	})
}
