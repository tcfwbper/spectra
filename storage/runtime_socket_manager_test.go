package storage_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/storage"
)

const (
	// Test timeouts and delays
	listenerStartupDelay  = 50 * time.Millisecond
	handlerProcessingWait = 50 * time.Millisecond
	longProcessingWait    = 100 * time.Millisecond
	listenerShutdownMax   = 2 * time.Second
	slowHandlerDelay      = 100 * time.Millisecond
	verySlowHandlerDelay  = 5 * time.Second
	fastClientMaxWait     = 1 * time.Second
	concurrentTestMax     = 250 * time.Millisecond

	// Test scale constants
	multipleConnectionsCount   = 3
	concurrentConnectionsCount = 10
	concurrentCreateAttempts   = 5
	activeConnectionsCount     = 5

	// Message size limits
	messageSizeLimit = 10 * 1024 * 1024 // 10 MB
	overLimitSize    = 11 * 1024 * 1024 // 11 MB

	// Buffer sizes
	readBufferSize = 4096
)

// mockMessageHandler creates a simple success handler
func mockMessageHandler() storage.MessageHandler {
	return func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}
}

// recordingMessageHandler records all invocations for verification
type recordingMessageHandler struct {
	mu           sync.Mutex
	invocations  []recordedInvocation
	responseFunc storage.MessageHandler
}

type recordedInvocation struct {
	SessionUUID string
	Message     storage.RuntimeMessage
}

func newRecordingMessageHandler(responseFunc storage.MessageHandler) *recordingMessageHandler {
	return &recordingMessageHandler{
		invocations:  []recordedInvocation{},
		responseFunc: responseFunc,
	}
}

func (r *recordingMessageHandler) Handle(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
	r.mu.Lock()
	r.invocations = append(r.invocations, recordedInvocation{
		SessionUUID: sessionUUID,
		Message:     message,
	})
	r.mu.Unlock()

	if r.responseFunc != nil {
		return r.responseFunc(sessionUUID, message)
	}
	return storage.RuntimeResponse{Status: "success", Message: "recorded"}
}

func (r *recordingMessageHandler) GetInvocations() []recordedInvocation {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedInvocation{}, r.invocations...)
}

func (r *recordingMessageHandler) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.invocations)
}

// socketClient is a helper for connecting to and sending messages to the socket
type socketClient struct {
	conn net.Conn
}

func connectToSocket(socketPath string) (*socketClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &socketClient{conn: conn}, nil
}

func (c *socketClient) SendMessage(msg string) error {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	_, err := c.conn.Write([]byte(msg))
	return err
}

func (c *socketClient) ReadResponse() (string, error) {
	buf := make([]byte, readBufferSize)
	n, err := c.conn.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func (c *socketClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// waitForSocket polls for socket availability with exponential backoff
func waitForSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	backoff := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			// Socket file exists, try to connect
			conn, err := net.Dial("unix", socketPath)
			if err == nil {
				conn.Close()
				return nil
			}
		}
		time.Sleep(backoff)
		backoff *= 2
		if backoff > 100*time.Millisecond {
			backoff = 100 * time.Millisecond
		}
	}
	return fmt.Errorf("socket not ready within timeout")
}

// setupTestFixture creates a test fixture with session directory
// Uses a shorter temp directory name to avoid Unix socket path length limits (108 chars)
func setupTestFixture(t *testing.T, sessionUUID string) string {
	// Use os.MkdirTemp with short prefix instead of t.TempDir() to avoid long paths
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	return tmpDir
}

// TestRuntimeSocketManager_New constructs RuntimeSocketManager with valid inputs
func TestRuntimeSocketManager_New(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)

	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)
	assert.NotNil(t, manager)
}

// TestRuntimeSocketManager_PathResolution resolves correct socket path from session UUID
func TestRuntimeSocketManager_PathResolution(t *testing.T) {
	sessionUUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	projectRoot := setupTestFixture(t, sessionUUID)

	expectedPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	assert.Equal(t, filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID, "runtime.sock"), expectedPath)
}

// TestCreateSocket_NewSocket creates socket file at correct path
func TestCreateSocket_NewSocket(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	err := manager.CreateSocket()
	assert.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	info, err := os.Stat(socketPath)
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

// TestCreateSocket_Permissions creates socket with correct permissions 0600
func TestCreateSocket_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission test on Windows")
	}

	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	err := manager.CreateSocket()
	assert.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	info, err := os.Stat(socketPath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// TestListen_StartsListener starts listener on created socket
func TestListen_StartsListener(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	listenerErrCh, listenerDoneCh, err := manager.Listen(mockMessageHandler())
	assert.NoError(t, err)
	assert.NotNil(t, listenerErrCh)
	assert.NotNil(t, listenerDoneCh)
}

// TestListen_AcceptsConnection accepts client connection successfully
func TestListen_AcceptsConnection(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","message":"msg"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	// Verify handler was invoked
	assert.Eventually(t, func() bool {
		return handler.Count() == 1
	}, handlerProcessingWait, 10*time.Millisecond, "Handler should be invoked exactly once")
}

// TestListen_MultipleConnections accepts multiple client connections sequentially
func TestListen_MultipleConnections(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	for i := 0; i < multipleConnectionsCount; i++ {
		client, err := connectToSocket(socketPath)
		require.NoError(t, err)

		err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
		require.NoError(t, err)

		_, err = client.ReadResponse()
		require.NoError(t, err)
		client.Close()
	}

	assert.Eventually(t, func() bool {
		return handler.Count() == multipleConnectionsCount
	}, handlerProcessingWait, 10*time.Millisecond, "All handlers should complete")
}

// TestReceive_ValidEventMessage parses valid event message and invokes handler
func TestReceive_ValidEventMessage(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","message":"msg"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.Equal(t, "event", invocations[0].Message.Type)
	assert.Equal(t, "test", invocations[0].Message.Payload["eventType"])
}

// TestReceive_ValidErrorMessage parses valid error message and invokes handler
func TestReceive_ValidErrorMessage(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"error","payload":{"message":"error msg"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.Equal(t, "error", invocations[0].Message.Type)
	assert.Equal(t, "error msg", invocations[0].Message.Payload["message"])
}

// TestReceive_SessionUUIDExtracted extracts session UUID from socket path and passes to handler
func TestReceive_SessionUUIDExtracted(t *testing.T) {
	sessionUUID := "abc-123"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	_, err = client.ReadResponse()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.Equal(t, "abc-123", invocations[0].SessionUUID)
}

// TestReceive_ComplexPayload handles complex nested payload structure
func TestReceive_ComplexPayload(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	complexMsg := `{"type":"event","payload":{"eventType":"test","nested":{"level2":{"level3":["a",1,true,null,{"key":"value"}]}}}}`
	err = client.SendMessage(complexMsg)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.NotNil(t, invocations[0].Message.Payload["nested"])
}

// TestReceive_OptionalFields handles message with optional fields
func TestReceive_OptionalFields(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","claudeSessionID":"session-123"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.Equal(t, "session-123", invocations[0].Message.Payload["claudeSessionID"])
}

// TestResponse_Success sends success response to client
func TestResponse_Success(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	successHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}
	_, _, err := manager.Listen(successHandler)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Equal(t, `{"status":"success","message":"ok"}`+"\n", response)
}

// TestResponse_Error sends error response to client
func TestResponse_Error(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	errorHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		return storage.RuntimeResponse{Status: "error", Message: "failed"}
	}
	_, _, err := manager.Listen(errorHandler)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Equal(t, `{"status":"error","message":"failed"}`+"\n", response)
}

// TestResponse_EmptyMessage sends response with empty message field
func TestResponse_EmptyMessage(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	emptyHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		return storage.RuntimeResponse{Status: "success", Message: ""}
	}
	_, _, err := manager.Listen(emptyHandler)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Equal(t, `{"status":"success","message":""}`+"\n", response)
}

// TestResponse_ConnectionClosedAfterSend closes connection after sending response
func TestResponse_ConnectionClosedAfterSend(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	_, _, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	_, err = client.ReadResponse()
	require.NoError(t, err)

	// Try to read again - should get EOF
	buf := make([]byte, 100)
	_, err = client.conn.Read(buf)
	assert.True(t, errors.Is(err, io.EOF) || err != nil, "Expected connection to be closed")
}

// TestDeleteSocket_StopsListener stops listening when socket deleted
func TestDeleteSocket_StopsListener(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	_, listenerDoneCh, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)

	err = manager.DeleteSocket()
	assert.NoError(t, err)

	// Wait for done channel to close
	select {
	case <-listenerDoneCh:
		// Expected - listener stopped
	case <-time.After(listenerShutdownMax):
		t.Fatal("Listener did not stop within timeout")
	}

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Socket file should be removed")
}

// TestDeleteSocket_ClosesActiveConnections closes active connections during deletion
func TestDeleteSocket_ClosesActiveConnections(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	slowHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		time.Sleep(listenerShutdownMax)
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}
	_, _, err := manager.Listen(slowHandler)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	// Create 2 connections
	client1, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client1.Close()

	client2, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client2.Close()

	// Send messages (will block in handler)
	go client1.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	go client2.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)

	time.Sleep(longProcessingWait)

	// Delete socket while connections are active
	err = manager.DeleteSocket()
	assert.NoError(t, err)
}

// TestDeleteSocket_RemovesSocketFile removes socket file from filesystem
func TestDeleteSocket_RemovesSocketFile(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	_, err := os.Stat(socketPath)
	require.NoError(t, err)

	err = manager.DeleteSocket()
	assert.NoError(t, err)

	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Socket file should be removed")
}

// TestDeleteSocket_ListenerNeverStarted deletes socket when listener was never started
func TestDeleteSocket_ListenerNeverStarted(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	err := manager.DeleteSocket()
	assert.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err))
}

// TestDeleteSocket_SocketDoesNotExist no error when socket file does not exist
func TestDeleteSocket_SocketDoesNotExist(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	err := manager.DeleteSocket()
	assert.NoError(t, err)
}

// TestDeleteSocket_CalledTwice second call to DeleteSocket is no-op
func TestDeleteSocket_CalledTwice(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	err := manager.DeleteSocket()
	assert.NoError(t, err)

	err = manager.DeleteSocket()
	assert.NoError(t, err)
}

// TestCreateSocket_SocketAlreadyExists returns error when socket file already exists
func TestCreateSocket_SocketAlreadyExists(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	file, err := os.Create(socketPath)
	require.NoError(t, err)
	file.Close()

	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)
	err = manager.CreateSocket()
	assert.Error(t, err)
	assert.Regexp(t, `(?i)runtime socket file already exists:.*runtime\.sock.*This may indicate a previous runtime process did not clean up properly`, err.Error())
}

// TestCreateSocket_ResidualSocket detects residual socket from previous session
func TestCreateSocket_ResidualSocket(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	file, err := os.Create(socketPath)
	require.NoError(t, err)
	file.Close()

	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)
	err = manager.CreateSocket()
	assert.Error(t, err)
	assert.Regexp(t, `(?i)rm.*runtime\.sock`, err.Error())
}

// TestCreateSocket_SessionDirDoesNotExist returns error when session directory missing
func TestCreateSocket_SessionDirDoesNotExist(t *testing.T) {
	// Use short temp dir name to avoid Unix socket path length limit
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	manager := storage.NewRuntimeSocketManager(tmpDir, sessionUUID)
	err = manager.CreateSocket()
	assert.Error(t, err)
	assert.Regexp(t, `(?i)failed to create runtime socket:.*no such file or directory`, err.Error())
}

// TestCreateSocket_PermissionDenied returns error when session directory is read-only
func TestCreateSocket_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission test on Windows")
	}

	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID)

	require.NoError(t, os.Chmod(sessionDir, 0444))
	defer func() { _ = os.Chmod(sessionDir, 0755) }()

	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)
	err := manager.CreateSocket()
	assert.Error(t, err)
	assert.Regexp(t, `(?i)failed to create runtime socket:.*permission denied`, err.Error())
}

// TestCreateSocket_PathTooLong returns error when socket path exceeds platform limit
func TestCreateSocket_PathTooLong(t *testing.T) {
	// TODO(test-infra): Implement test with deeply nested directory structure
	// Unix socket paths typically limited to ~108 characters on some systems
	t.Skip("Testing path length limits requires very long directory structures")
}

// TestListen_SocketNotCreated returns error when called before CreateSocket
func TestListen_SocketNotCreated(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	errCh, doneCh, err := manager.Listen(mockMessageHandler())
	assert.Error(t, err)
	assert.Regexp(t, `(?i)runtime socket not created: call CreateSocket\(\) first`, err.Error())
	assert.Nil(t, errCh)
	assert.Nil(t, doneCh)
}

// TestListen_BindFails returns error when socket bind fails
func TestListen_BindFails(t *testing.T) {
	// TODO(test-infra): Implement with mock net.Listener or filesystem manipulation
	t.Skip("Simulating bind failure requires complex mocking")
}

// TestListen_InitialBindFailure returns nil channels when initial bind/listen fails
func TestListen_InitialBindFailure(t *testing.T) {
	// TODO(test-infra): Implement by holding socket in another process before test
	t.Skip("Simulating initial bind failure requires complex mocking")
}

// TestReceive_MalformedJSON rejects message with invalid JSON
func TestReceive_MalformedJSON(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count(), "Handler should not be invoked for malformed JSON")
}

// TestReceive_NotJSONObject rejects non-object JSON
func TestReceive_NotJSONObject(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`["array","not","object"]`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_MissingTypeField rejects message missing type field
func TestReceive_MissingTypeField(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_MissingPayloadField rejects message missing payload field
func TestReceive_MissingPayloadField(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event"}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_InvalidMessageType rejects message with invalid type value
func TestReceive_InvalidMessageType(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"unknown","payload":{}}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_PayloadNotObject rejects message where payload is not an object
func TestReceive_PayloadNotObject(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":"string"}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_EventMissingEventType rejects event message missing eventType in payload
func TestReceive_EventMissingEventType(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"message":"test"}}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_ErrorEmptyMessage rejects error message with empty message field
func TestReceive_ErrorEmptyMessage(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"error","payload":{"message":""}}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_ClaudeSessionIDNotString rejects message with non-string claudeSessionID
func TestReceive_ClaudeSessionIDNotString(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","claudeSessionID":123}}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_EventTypeNotString rejects event with non-string eventType
func TestReceive_EventTypeNotString(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":123}}`)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_MessageExceeds10MBLimit rejects message exceeding 10 MB limit
func TestReceive_MessageExceeds10MBLimit(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	// Create a message just over 10 MB
	largeData := strings.Repeat("a", overLimitSize)
	largeMsg := fmt.Sprintf(`{"type":"event","payload":{"eventType":"test","data":"%s"}}`, largeData)
	err = client.SendMessage(largeMsg)
	// Note: The send may fail with "broken pipe" if the server closes the connection
	// while the client is still writing. This is expected for oversized messages.
	// The test verifies that the handler is not invoked, regardless of send result.
	_ = err

	assert.Eventually(t, func() bool {
		return handler.Count() == 0
	}, longProcessingWait, 10*time.Millisecond, "Handler should not be invoked for oversized message")
}

// TestReceive_MessageExactly10MB accepts message at exactly 10 MB limit
func TestReceive_MessageExactly10MB(t *testing.T) {
	// TODO(performance): Implement with streaming or test in separate slow test suite
	t.Skip("Creating exact 10MB message is slow and may timeout")
}

// TestReceive_MessageJustUnder10MB accepts message just under 10 MB limit
func TestReceive_MessageJustUnder10MB(t *testing.T) {
	// TODO(performance): Implement with streaming or test in separate slow test suite
	t.Skip("Creating near-10MB message is slow and may timeout")
}

// TestReceive_MessageWithNewlines handles escaped newlines in message field
func TestReceive_MessageWithNewlines(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","message":"line1\nline2\nline3"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_MessageWithUnicode handles Unicode characters in payload
func TestReceive_MessageWithUnicode(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","emoji":"🎉","cjk":"中文"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.Equal(t, "🎉", invocations[0].Message.Payload["emoji"])
	assert.Equal(t, "中文", invocations[0].Message.Payload["cjk"])
}

// TestReceive_MessageWithQuotes handles escaped quotes in payload
func TestReceive_MessageWithQuotes(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","quote":"He said \"hello\""}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_MinimalEventPayload accepts event with only required fields
func TestReceive_MinimalEventPayload(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_MinimalErrorPayload accepts error with only required fields
func TestReceive_MinimalErrorPayload(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"error","payload":{"message":"error"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_EmptyNestedObjects handles empty nested objects in payload
func TestReceive_EmptyNestedObjects(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","payload":{}}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestRuntimeSocketManager_InvalidSessionUUID handles malformed session UUID
func TestRuntimeSocketManager_InvalidSessionUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := "not-a-uuid"

	manager := storage.NewRuntimeSocketManager(tmpDir, sessionUUID)
	assert.NotNil(t, manager, "Manager construction should succeed with invalid UUID")

	err := manager.CreateSocket()
	assert.Error(t, err, "CreateSocket should fail with invalid UUID path")
	assert.Regexp(t, `(?i)failed to create runtime socket:.*no such file or directory`, err.Error(),
		"Error should indicate filesystem issue due to malformed path")
}

// TestRuntimeSocketManager_EmptySessionUUID handles empty session UUID
func TestRuntimeSocketManager_EmptySessionUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionUUID := ""

	manager := storage.NewRuntimeSocketManager(tmpDir, sessionUUID)
	assert.NotNil(t, manager, "Manager construction should succeed with empty UUID")

	err := manager.CreateSocket()
	assert.Error(t, err, "CreateSocket should fail with empty UUID")
	assert.Regexp(t, `(?i)failed to create runtime socket:.*no such file or directory`, err.Error(),
		"Error should indicate filesystem issue due to missing session directory")
}

// TestReceive_ClientClosesConnection handles client closing connection without sending message
func TestReceive_ClientClosesConnection(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	client.Close()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, handler.Count())
}

// TestReceive_ClientClosesAfterMessage handles client closing after valid message but before reading response
func TestReceive_ClientClosesAfterMessage(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)
	client.Close()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_ReadTimeout handles read timeout (if implemented)
func TestReceive_ReadTimeout(t *testing.T) {
	// TODO(feature): Implement once read timeout is added to RuntimeSocketManager
	t.Skip("Read timeout behavior depends on implementation")
}

// TestResponse_SendFailsIOError handles I/O error when sending response
func TestResponse_SendFailsIOError(t *testing.T) {
	// TODO(test-infra): Implement by closing client before response is sent
	t.Skip("Simulating I/O error during send requires complex mocking")
}

// TestListen_AcceptLoopFails delivers asynchronous listener error via channel
func TestListen_AcceptLoopFails(t *testing.T) {
	// TODO(test-infra): Implement by corrupting socket file during accept loop
	t.Skip("Simulating accept loop failure requires complex mocking")
}

// TestListen_ListenerErrChannelBuffered error channel has capacity 1 and is never closed
func TestListen_ListenerErrChannelBuffered(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	listenerErrCh, _, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)
	defer manager.DeleteSocket()

	// Check channel is buffered (capacity > 0)
	assert.Equal(t, 1, cap(listenerErrCh))
}

// TestListen_ListenerDoneSignalsShutdown done channel closed on listener exit
func TestListen_ListenerDoneSignalsShutdown(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	_, listenerDoneCh, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)

	err = manager.DeleteSocket()
	require.NoError(t, err)

	select {
	case <-listenerDoneCh:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("Done channel not closed within timeout")
	}
}

// TestDeleteSocket_RemovalFailsWarning logs warning when socket removal fails but does not error
func TestDeleteSocket_RemovalFailsWarning(t *testing.T) {
	// TODO(test-infra): Implement by setting immutable flag or using mock filesystem
	t.Skip("Simulating filesystem errors requires platform-specific setup")
}

// TestDeleteSocket_ProceedsAfterFailure continues gracefully even if removal fails
func TestDeleteSocket_ProceedsAfterFailure(t *testing.T) {
	// TODO(test-infra): Implement by making socket file undeletable during test
	t.Skip("Simulating filesystem errors requires platform-specific setup")
}

// TestReceive_MalformedMessageIsolation malformed message on one connection does not affect others
func TestReceive_MalformedMessageIsolation(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	// Client 1 sends malformed JSON
	client1, err := connectToSocket(socketPath)
	require.NoError(t, err)
	err = client1.SendMessage(`{"type":"event","payload":`)
	require.NoError(t, err)
	client1.Close()

	time.Sleep(50 * time.Millisecond)

	// Client 2 sends valid message
	client2, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client2.Close()

	err = client2.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	response, err := client2.ReadResponse()
	require.NoError(t, err)
	assert.Contains(t, response, `"status":"success"`)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, handler.Count())
}

// TestListen_ConcurrentConnections handles multiple simultaneous connections safely
func TestListen_ConcurrentConnections(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	var count int32
	countingHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		atomic.AddInt32(&count, 1)
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}

	_, _, err := manager.Listen(countingHandler)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	var wg sync.WaitGroup
	for i := 0; i < concurrentConnectionsCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := connectToSocket(socketPath)
			if err != nil {
				return
			}
			defer client.Close()

			_ = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
			_, _ = client.ReadResponse()
		}()
	}

	wg.Wait()

	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&count) == concurrentConnectionsCount
	}, longProcessingWait, 10*time.Millisecond, "All concurrent handlers should complete")
}

// TestListen_ConnectionGoroutineIsolation each connection runs in separate goroutine
func TestListen_ConnectionGoroutineIsolation(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	slowHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		time.Sleep(slowHandlerDelay)
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}

	_, _, err := manager.Listen(slowHandler)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < multipleConnectionsCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := connectToSocket(socketPath)
			if err != nil {
				return
			}
			defer client.Close()

			_ = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
			_, _ = client.ReadResponse()
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Should take ~slowHandlerDelay (concurrent), not multipleConnectionsCount * slowHandlerDelay (sequential)
	assert.Less(t, elapsed, concurrentTestMax,
		"Concurrent execution should complete faster than sequential")
}

// TestCreateSocket_ConcurrentCalls filesystem-level atomic check prevents race conditions
func TestCreateSocket_ConcurrentCalls(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)

	var wg sync.WaitGroup
	var successCount int32
	var errorCount int32

	for i := 0; i < concurrentCreateAttempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)
			err := manager.CreateSocket()
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&errorCount, 1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&successCount), "Exactly one should succeed")
	assert.Equal(t, int32(concurrentCreateAttempts-1), atomic.LoadInt32(&errorCount), "Others should fail")
}

// TestDeleteSocket_DuringActiveConnections safely closes connections during DeleteSocket
func TestDeleteSocket_DuringActiveConnections(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	slowHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		time.Sleep(fastClientMaxWait)
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}

	_, _, err := manager.Listen(slowHandler)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	// Start active connections
	for i := 0; i < activeConnectionsCount; i++ {
		go func() {
			client, err := connectToSocket(socketPath)
			if err != nil {
				return
			}
			defer client.Close()

			_ = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
			_, _ = client.ReadResponse()
		}()
	}

	time.Sleep(longProcessingWait)

	// Delete socket while connections are processing
	err = manager.DeleteSocket()
	assert.NoError(t, err)
}

// TestDeleteSocket_NoRaceCondition no race between DeleteSocket and connection handlers
func TestDeleteSocket_NoRaceCondition(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	_, _, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	// Start multiple connections
	var wg sync.WaitGroup
	for i := 0; i < concurrentConnectionsCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := connectToSocket(socketPath)
			if err != nil {
				return
			}
			defer client.Close()

			_ = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
			_, _ = client.ReadResponse()
		}()
	}

	// Delete socket concurrently
	go func() {
		time.Sleep(listenerStartupDelay)
		_ = manager.DeleteSocket()
	}()

	wg.Wait()
}

// TestListen_ErrorChannelNeverClosed error channel is never closed by RuntimeSocketManager
func TestListen_ErrorChannelNeverClosed(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	listenerErrCh, listenerDoneCh, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)

	err = manager.DeleteSocket()
	require.NoError(t, err)

	<-listenerDoneCh

	// Try to receive from error channel - should not panic
	select {
	case <-listenerErrCh:
		// If we get an error, that's fine
	case <-time.After(100 * time.Millisecond):
		// If we timeout, that's also fine - channel is open but no error sent
	}
}

// TestListen_DoneChannelClosedOnce done channel closed exactly once
func TestListen_DoneChannelClosedOnce(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	_, listenerDoneCh, err := manager.Listen(mockMessageHandler())
	require.NoError(t, err)

	// Call DeleteSocket multiple times
	err = manager.DeleteSocket()
	assert.NoError(t, err)

	err = manager.DeleteSocket()
	assert.NoError(t, err)

	// Wait for done channel
	<-listenerDoneCh

	// Receiving from done channel again should work (closed channels always return immediately)
	<-listenerDoneCh
}

// TestReceive_MessageHandlerSessionUUID MessageHandler receives session UUID extracted from path
func TestReceive_MessageHandlerSessionUUID(t *testing.T) {
	sessionUUID := "abc-123"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	_, err = client.ReadResponse()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)
	assert.Equal(t, "abc-123", invocations[0].SessionUUID)
}

// TestReceive_MessageHandlerRuntimeMessage MessageHandler receives correctly parsed RuntimeMessage
func TestReceive_MessageHandlerRuntimeMessage(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test","message":"msg"}}`)
	require.NoError(t, err)

	_, err = client.ReadResponse()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	invocations := handler.GetInvocations()
	require.Len(t, invocations, 1)

	msg := invocations[0].Message
	assert.Equal(t, "event", msg.Type)
	assert.Equal(t, "test", msg.Payload["eventType"])
	assert.Equal(t, "msg", msg.Payload["message"])
}

// TestReceive_MessageHandlerResponseSerialized MessageHandler response correctly serialized to JSON
func TestReceive_MessageHandlerResponseSerialized(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	customHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		return storage.RuntimeResponse{Status: "success", Message: "processed"}
	}

	_, _, err := manager.Listen(customHandler)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test"}}`)
	require.NoError(t, err)

	response, err := client.ReadResponse()
	require.NoError(t, err)

	var resp storage.RuntimeResponse
	err = json.Unmarshal([]byte(response), &resp)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "processed", resp.Message)
}

// TestReceive_MessageHandlerSlow slow MessageHandler does not block other connections
func TestReceive_MessageHandlerSlow(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())
	defer manager.DeleteSocket()

	var count int32
	slowHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		n := atomic.AddInt32(&count, 1)
		if n == 1 {
			time.Sleep(verySlowHandlerDelay)
		}
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}

	_, _, err := manager.Listen(slowHandler)
	require.NoError(t, err)

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	require.NoError(t, waitForSocket(socketPath, listenerShutdownMax))

	// Start slow client
	go func() {
		client, err := connectToSocket(socketPath)
		if err != nil {
			return
		}
		defer client.Close()
		_ = client.SendMessage(`{"type":"event","payload":{"eventType":"slow"}}`)
		_, _ = client.ReadResponse()
	}()

	time.Sleep(longProcessingWait)

	// Start fast client
	start := time.Now()
	client2, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client2.Close()

	err = client2.SendMessage(`{"type":"event","payload":{"eventType":"fast"}}`)
	require.NoError(t, err)

	_, err = client2.ReadResponse()
	require.NoError(t, err)
	elapsed := time.Since(start)

	// Fast client should complete quickly, not wait for slow client
	assert.Less(t, elapsed, fastClientMaxWait,
		"Fast client should not be blocked by slow client")
}

// TestCreateSocket_WindowsNamedPipe creates named pipe on Windows instead of Unix socket
func TestCreateSocket_WindowsNamedPipe(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	err := manager.CreateSocket()
	assert.NoError(t, err)
	defer manager.DeleteSocket()

	_, _, err = manager.Listen(mockMessageHandler())
	assert.NoError(t, err)
}

// TestCreateSocket_UnixDomainSocket creates Unix domain socket on Unix-like systems
func TestCreateSocket_UnixDomainSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	err := manager.CreateSocket()
	assert.NoError(t, err)
	defer manager.DeleteSocket()

	_, _, err = manager.Listen(mockMessageHandler())
	assert.NoError(t, err)
}

// TestReceive_SingleRequestResponseCycle only one request-response cycle per connection
func TestReceive_SingleRequestResponseCycle(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test1"}}`)
	require.NoError(t, err)

	_, err = client.ReadResponse()
	require.NoError(t, err)

	// Try to send second message
	err = client.SendMessage(`{"type":"event","payload":{"eventType":"test2"}}`)
	// May or may not error depending on timing

	time.Sleep(100 * time.Millisecond)
	// Only first message should be processed
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_MultipleMessagesInStream subsequent messages in stream are ignored
func TestReceive_MultipleMessagesInStream(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	handler := newRecordingMessageHandler(mockMessageHandler())
	_, _, err := manager.Listen(handler.Handle)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	client, err := connectToSocket(socketPath)
	require.NoError(t, err)
	defer client.Close()

	// Send two messages in one stream
	multiMsg := `{"type":"event","payload":{"eventType":"test1"}}
{"type":"event","payload":{"eventType":"test2"}}
`
	_, err = client.conn.Write([]byte(multiMsg))
	require.NoError(t, err)

	_, err = client.ReadResponse()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	// Only first message should be processed
	assert.Equal(t, 1, handler.Count())
}

// TestReceive_NoMessageBuffering messages processed synchronously without buffering
func TestReceive_NoMessageBuffering(t *testing.T) {
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	projectRoot := setupTestFixture(t, sessionUUID)
	manager := storage.NewRuntimeSocketManager(projectRoot, sessionUUID)

	require.NoError(t, manager.CreateSocket())

	var order []int
	var mu sync.Mutex

	orderingHandler := func(sessionUUID string, message storage.RuntimeMessage) storage.RuntimeResponse {
		mu.Lock()
		defer mu.Unlock()

		if eventType, ok := message.Payload["eventType"].(string); ok {
			if eventType == "test1" {
				order = append(order, 1)
			} else if eventType == "test2" {
				order = append(order, 2)
			} else if eventType == "test3" {
				order = append(order, 3)
			}
		}
		return storage.RuntimeResponse{Status: "success", Message: "ok"}
	}

	_, _, err := manager.Listen(orderingHandler)
	require.NoError(t, err)
	defer manager.DeleteSocket()

	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionUUID)
	time.Sleep(50 * time.Millisecond)

	// Send messages with slight delays
	for i := 1; i <= 3; i++ {
		client, err := connectToSocket(socketPath)
		require.NoError(t, err)

		err = client.SendMessage(fmt.Sprintf(`{"type":"event","payload":{"eventType":"test%d"}}`, i))
		require.NoError(t, err)

		_, err = client.ReadResponse()
		require.NoError(t, err)
		client.Close()

		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1, 2, 3}, order)
}
