package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectra-ai/spectra/entities"
)

// --- Mock MessageHandler ---

// mockHandleCall records a single invocation of MessageHandler.Handle.
type mockHandleCall struct {
	SessionUUID string
	Msg         entities.RuntimeMessage
}

// mockMessageHandler is a test double for the MessageHandler interface.
type mockMessageHandler struct {
	mu         sync.Mutex
	calls      []mockHandleCall
	response   entities.RuntimeResponse
	blockCh    chan struct{} // if non-nil, Handle blocks until closed
	panicMsg   string        // if non-empty, Handle panics with this message
	panicOnce  bool          // if true, panic only on first call
	panicCount int           // tracks how many times Handle has been called (for panicOnce)
}

func newMockMessageHandler(resp entities.RuntimeResponse) *mockMessageHandler {
	return &mockMessageHandler{response: resp}
}

func (m *mockMessageHandler) Handle(sessionUUID string, msg entities.RuntimeMessage) entities.RuntimeResponse {
	if m.blockCh != nil {
		<-m.blockCh
	}

	m.mu.Lock()
	shouldPanic := false
	if m.panicMsg != "" {
		if m.panicOnce {
			if m.panicCount == 0 {
				shouldPanic = true
			}
			m.panicCount++
		} else {
			shouldPanic = true
		}
	}
	if !shouldPanic {
		m.calls = append(m.calls, mockHandleCall{SessionUUID: sessionUUID, Msg: msg})
	}
	m.mu.Unlock()

	if shouldPanic {
		panic(m.panicMsg)
	}
	return m.response
}

func (m *mockMessageHandler) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *mockMessageHandler) lastCall() mockHandleCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[len(m.calls)-1]
}

func (m *mockMessageHandler) allCalls() []mockHandleCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]mockHandleCall, len(m.calls))
	copy(cp, m.calls)
	return cp
}

// --- Fixture Builders for RuntimeSocketManager tests ---

// makeSocketTestDir creates a temp dir with the session directory structure and returns
// the projectRoot and the expected socket path.
func makeSocketTestDir(t *testing.T) (projectRoot string, socketPath string) {
	t.Helper()
	projectRoot = makeTempDirWithSessionDir(t, testSessionUUID)
	socketPath = filepath.Join(projectRoot, ".spectra", "sessions", testSessionUUID, RuntimeSocketFile)
	return projectRoot, socketPath
}

// dialSocket connects a client to the given Unix domain socket path.
func dialSocket(t *testing.T, socketPath string) net.Conn {
	t.Helper()
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err, "failed to dial socket")
	t.Cleanup(func() { conn.Close() })
	return conn
}

// sendMessage writes a message string terminated by newline to the connection.
func sendMessage(t *testing.T, conn net.Conn, msg string) {
	t.Helper()
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	_, err := conn.Write([]byte(msg))
	require.NoError(t, err, "failed to send message")
}

// readResponse reads a newline-terminated JSON response from the connection.
func readResponse(t *testing.T, conn net.Conn) string {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	require.NoError(t, err, "failed to read response")
	return line
}

// parseResponse parses a JSON response string into a map.
func parseResponse(t *testing.T, raw string) map[string]string {
	t.Helper()
	var resp map[string]string
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err, "failed to parse response JSON")
	return resp
}

// --- Happy Path — Construction ---

func TestNewRuntimeSocketManager_StoresSocketPath(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager in storage/runtime_socket_manager.go")

	projectRoot := makeTempDirWithSessionDir(t, testSessionUUID)
	ml := newMockLogger()

	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)

	require.NotNil(t, mgr)
	expectedPath := GetRuntimeSocketPath(projectRoot, testSessionUUID)
	_ = expectedPath
	// Assert: internal socket path equals expectedPath
	// (needs access to unexported field or a getter; verify via CreateSocket behavior)
}

func TestNewRuntimeSocketManager_NoIO(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager in storage/runtime_socket_manager.go")

	ml := newMockLogger()

	// Non-existent path — constructor must not touch filesystem.
	assert.NotPanics(t, func() {
		mgr := NewRuntimeSocketManager("/nonexistent/path", "aaaa-bbbb", ml)
		require.NotNil(t, mgr)
	})
}

// --- Happy Path — CreateSocket ---

func TestCreateSocket_CreatesSocketFile(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)

	err := mgr.CreateSocket()

	require.NoError(t, err)
	_, statErr := os.Stat(socketPath)
	assert.NoError(t, statErr, "socket file should exist")
}

func TestCreateSocket_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix socket permission test not applicable on Windows")
	}
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)

	err := mgr.CreateSocket()

	require.NoError(t, err)
	info, statErr := os.Stat(socketPath)
	require.NoError(t, statErr)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// --- Happy Path — Listen ---

func TestListen_ReturnsChannels(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, _ := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))

	listenerErrCh, listenerDoneCh, err := mgr.Listen(ctx, handler)

	require.NoError(t, err)
	assert.NotNil(t, listenerErrCh)
	assert.NotNil(t, listenerDoneCh)
}

func TestListen_AcceptsConnection(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn, dialErr := net.Dial("unix", socketPath)
	require.NoError(t, dialErr, "client connection should succeed")
	conn.Close()
}

func TestListen_DispatchesToHandler(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":{"key":"val"}}`)
	_ = readResponse(t, conn)

	require.Eventually(t, func() bool { return handler.callCount() == 1 }, 5*time.Second, 10*time.Millisecond)
	call := handler.lastCall()
	assert.Equal(t, testSessionUUID, call.SessionUUID)
	assert.Equal(t, "event", call.Msg.Type())
}

func TestListen_SendsResponseToClient(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":{"key":"val"}}`)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "success", resp["status"])
	assert.Equal(t, "ok", resp["message"])
}

func TestListen_ClosesConnectionAfterResponse(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn, dialErr := net.Dial("unix", socketPath)
	require.NoError(t, dialErr)
	sendMessage(t, conn, `{"type":"event","payload":{"key":"val"}}`)
	_ = readResponse(t, conn)

	// Second read should return EOF or error
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	_, readErr := conn.Read(buf)
	assert.Error(t, readErr, "connection should be closed after response")
}

// --- Error Propagation ---

func TestCreateSocket_SocketAlreadyExists(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()

	// Pre-create a file at the socket path
	makeTempFile(t, socketPath)

	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	err := mgr.CreateSocket()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "runtime socket file already exists:")
	assert.Contains(t, err.Error(), socketPath)
}

func TestCreateSocket_DirectoryMissing(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket in storage/runtime_socket_manager.go")

	ml := newMockLogger()
	// Point to a non-existent directory
	mgr := NewRuntimeSocketManager("/nonexistent/path", testSessionUUID, ml)

	err := mgr.CreateSocket()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create runtime socket:")
}

func TestCreateSocket_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket in storage/runtime_socket_manager.go")

	projectRoot := makeTempDirWithSessionDir(t, testSessionUUID)
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", testSessionUUID)
	require.NoError(t, os.Chmod(sessionDir, 0555))
	t.Cleanup(func() { os.Chmod(sessionDir, 0755) })

	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)

	err := mgr.CreateSocket()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create runtime socket:")
}

func TestListen_BeforeCreateSocket(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, Listen in storage/runtime_socket_manager.go")

	projectRoot := makeTempDirWithSessionDir(t, testSessionUUID)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)

	ctx := context.Background()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	listenerErrCh, listenerDoneCh, err := mgr.Listen(ctx, handler)

	require.Error(t, err)
	assert.Equal(t, "runtime socket not created: call CreateSocket() first", err.Error())
	assert.Nil(t, listenerErrCh)
	assert.Nil(t, listenerDoneCh)
}

func TestListen_BindFailure(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	// Delete the socket file externally before calling Listen
	require.NoError(t, os.Remove(socketPath))

	ctx := context.Background()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to listen on runtime socket:")
}

// --- Happy Path — DeleteSocket ---

func TestDeleteSocket_RemovesFile(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	mgr.DeleteSocket(ctx)

	_, statErr := os.Stat(socketPath)
	assert.True(t, os.IsNotExist(statErr), "socket file should be removed")
}

func TestDeleteSocket_ClosesListener(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, _ := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, listenerDoneCh, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	mgr.DeleteSocket(ctx)

	select {
	case <-listenerDoneCh:
		// OK - done channel is closed
	case <-time.After(5 * time.Second):
		t.Fatal("listenerDoneCh was not closed after DeleteSocket")
	}
}

// --- Idempotency ---

func TestDeleteSocket_Idempotent(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, _ := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	mgr.DeleteSocket(ctx)

	// Second call should not panic or error
	assert.NotPanics(t, func() {
		mgr.DeleteSocket(ctx)
	})
}

func TestDeleteSocket_FileAlreadyGone(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Manually remove socket file before calling DeleteSocket
	os.Remove(socketPath)

	assert.NotPanics(t, func() {
		mgr.DeleteSocket(ctx)
	})
	assert.Equal(t, 0, ml.warnCallCount(), "no warning should be logged")
}

// --- Validation Failures ---

func TestPerConnection_MalformedJSON(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, "not valid json")
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "error", resp["status"])
	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
	assert.Contains(t, ml.warnMsgs[0], "malformed JSON")
}

func TestPerConnection_MissingTypeField(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"payload":{"k":"v"}}`)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "error", resp["status"])
	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
}

func TestPerConnection_InvalidTypeValue(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"unknown","payload":{"k":"v"}}`)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "error", resp["status"])
	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
}

func TestPerConnection_MissingPayload(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event"}`)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "error", resp["status"])
	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
}

func TestPerConnection_PayloadNotObject(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":"string"}`)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "error", resp["status"])
	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
}

func TestPerConnection_PayloadArray(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":[1,2]}`)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "error", resp["status"])
	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
}

func TestPerConnection_ClientClosesWithoutSending(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Connect and immediately close
	conn, dialErr := net.Dial("unix", socketPath)
	require.NoError(t, dialErr)
	conn.Close()

	// Verify no warning is logged within a reasonable window
	require.Never(t, func() bool { return ml.warnCallCount() > 0 }, 500*time.Millisecond, 50*time.Millisecond,
		"no warning should be logged for clean close")
}

// --- Boundary Values — Message Size ---

func TestPerConnection_ExceedsMaxSize(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Build a message larger than 10 MB
	bigPayload := strings.Repeat("x", MaxPayloadSize+1)
	msg := fmt.Sprintf(`{"type":"event","payload":{"data":"%s"}}`, bigPayload)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, msg)

	// May receive error response, depending on whether the server can still write
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	line, _ := reader.ReadString('\n')
	if line != "" {
		resp := parseResponse(t, line)
		assert.Equal(t, "error", resp["status"])
	}

	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
	assert.Contains(t, ml.warnMsgs[0], "message size exceeds 10 MB limit")
	assert.Equal(t, 0, handler.callCount(), "MessageHandler should not be invoked")
}

func TestPerConnection_AtMaxSize(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Build a valid JSON message that is exactly at the limit (10 MB including newline).
	// The exact size calculation is best-effort; production implementation determines parsing.
	prefix := `{"type":"event","payload":{"data":"`
	suffix := `"}}`
	padLen := MaxPayloadSize - len(prefix) - len(suffix) - 1 // -1 for newline
	if padLen < 0 {
		padLen = 0
	}
	msg := prefix + strings.Repeat("a", padLen) + suffix

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, msg)
	raw := readResponse(t, conn)

	resp := parseResponse(t, raw)
	assert.Equal(t, "success", resp["status"])
	require.Eventually(t, func() bool { return handler.callCount() == 1 }, 5*time.Second, 10*time.Millisecond)
}

// --- Null / Empty Input ---

func TestPerConnection_EmptyClaudeSessionID(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":{"k":"v"}}`)
	_ = readResponse(t, conn)

	require.Eventually(t, func() bool { return handler.callCount() == 1 }, 5*time.Second, 10*time.Millisecond)
	call := handler.lastCall()
	assert.Equal(t, "", call.Msg.ClaudeSessionID())
}

func TestPerConnection_WithClaudeSessionID(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":{"k":"v"},"claudeSessionID":"sess-123"}`)
	_ = readResponse(t, conn)

	require.Eventually(t, func() bool { return handler.callCount() == 1 }, 5*time.Second, 10*time.Millisecond)
	call := handler.lastCall()
	assert.Equal(t, "sess-123", call.Msg.ClaudeSessionID())
}

func TestPerConnection_EmptyMessageInResponse(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse(""))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"event","payload":{"k":"v"}}`)
	raw := readResponse(t, conn)

	assert.Equal(t, `{"status":"success","message":""}`+"\n", raw)
}

// --- Mock / Dependency Interaction ---

func TestConstruction_CallsStorageLayout(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager in storage/runtime_socket_manager.go")

	// GetRuntimeSocketPath is a package-level function; constructor should compose via it.
	// Verify by observing that CreateSocket creates the file at the correct path.
	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	_, statErr := os.Stat(socketPath)
	assert.NoError(t, statErr, "socket should be at path from GetRuntimeSocketPath")
}

func TestPerConnection_InvokesNewRuntimeMessage(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn := dialSocket(t, socketPath)
	sendMessage(t, conn, `{"type":"error","payload":{"detail":"x"},"claudeSessionID":"c1"}`)
	_ = readResponse(t, conn)

	require.Eventually(t, func() bool { return handler.callCount() == 1 }, 5*time.Second, 10*time.Millisecond)
	call := handler.lastCall()
	assert.Equal(t, "error", call.Msg.Type())
	assert.JSONEq(t, `{"detail":"x"}`, string(call.Msg.Payload()))
	assert.Equal(t, "c1", call.Msg.ClaudeSessionID())
}

func TestDeleteSocket_LogsOnFileDeletionFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, _ := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Make parent dir read-only so socket file cannot be deleted
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", testSessionUUID)
	require.NoError(t, os.Chmod(sessionDir, 0555))
	t.Cleanup(func() { os.Chmod(sessionDir, 0755) })

	mgr.DeleteSocket(ctx)

	assert.Greater(t, ml.warnCallCount(), 0)
	assert.Contains(t, ml.warnMsgs[0], "failed to delete runtime socket:")
}

func TestPerConnection_LogsOnSendFailure(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockCh := make(chan struct{})
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	handler.blockCh = blockCh
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn, dialErr := net.Dial("unix", socketPath)
	require.NoError(t, dialErr)
	sendMessage(t, conn, `{"type":"event","payload":{"k":"v"}}`)

	// Close client before handler returns
	conn.Close()

	// Unblock handler
	close(blockCh)

	require.Eventually(t, func() bool { return ml.warnCallCount() > 0 }, 5*time.Second, 10*time.Millisecond)
	assert.Contains(t, ml.warnMsgs[0], "failed to send response to client:")
}

// --- Concurrent Behaviour ---

func TestListen_MultipleSimultaneousConnections(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	const numClients = 3
	var wg sync.WaitGroup
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(idx int) {
			defer wg.Done()
			conn, dialErr := net.Dial("unix", socketPath)
			if dialErr != nil {
				return
			}
			defer conn.Close()
			msg := fmt.Sprintf(`{"type":"event","payload":{"idx":%d}}`, idx)
			conn.Write([]byte(msg + "\n"))
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			reader := bufio.NewReader(conn)
			reader.ReadString('\n')
		}(i)
	}
	wg.Wait()

	require.Eventually(t, func() bool { return handler.callCount() == numClients }, 5*time.Second, 10*time.Millisecond)
}

func TestPerConnection_IsolationOnError(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Client A sends malformed JSON
	connA := dialSocket(t, socketPath)
	sendMessage(t, connA, "not valid json")
	rawA := readResponse(t, connA)
	respA := parseResponse(t, rawA)
	assert.Equal(t, "error", respA["status"])

	// Client B sends valid message
	connB := dialSocket(t, socketPath)
	sendMessage(t, connB, `{"type":"event","payload":{"k":"v"}}`)
	rawB := readResponse(t, connB)
	respB := parseResponse(t, rawB)
	assert.Equal(t, "success", respB["status"])
}

func TestPerConnection_HandlerPanicCrashesGoroutine(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handler that panics on first call only (uses panicOnce for race-free behavior)
	panicHandler := &mockMessageHandler{
		response:  *entities.SuccessResponse("ok"),
		panicMsg:  "handler panic",
		panicOnce: true,
	}
	_, _, err := mgr.Listen(ctx, panicHandler)
	require.NoError(t, err)

	// Client A triggers the panic
	connA, dialErrA := net.Dial("unix", socketPath)
	require.NoError(t, dialErrA)
	connA.Write([]byte(`{"type":"event","payload":{"k":"v"}}` + "\n"))
	// Connection should be closed without a graceful response
	connA.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 512)
	_, readErr := connA.Read(buf)
	// Either EOF or error expected (panic crashes the goroutine)
	_ = readErr
	connA.Close()

	// Client B should still work (panicOnce ensures no panic on second call)
	connB := dialSocket(t, socketPath)
	sendMessage(t, connB, `{"type":"event","payload":{"k":"v"}}`)
	rawB := readResponse(t, connB)
	respB := parseResponse(t, rawB)
	assert.Equal(t, "success", respB["status"])
}

// --- Resource Cleanup ---

func TestDeleteSocket_ClosesActiveConnections(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockCh := make(chan struct{})
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	handler.blockCh = blockCh
	_, listenerDoneCh, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Connect client and send a message (handler blocks)
	conn, dialErr := net.Dial("unix", socketPath)
	require.NoError(t, dialErr)
	conn.Write([]byte(`{"type":"event","payload":{"k":"v"}}` + "\n"))

	// Wait until the handler is actively blocked (connection was accepted and dispatched)
	require.Eventually(t, func() bool {
		// The handler goroutine is blocking on blockCh, meaning the connection was accepted.
		// We verify indirectly: if we can establish another connection, the accept loop is running.
		probe, err := net.DialTimeout("unix", socketPath, 100*time.Millisecond)
		if err == nil {
			probe.Close()
			return true
		}
		return false
	}, 5*time.Second, 10*time.Millisecond)

	// Delete socket while handler is blocked
	mgr.DeleteSocket(ctx)
	close(blockCh) // unblock handler so goroutine can exit

	// Client connection should be closed
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	_, readErr := conn.Read(buf)
	assert.Error(t, readErr, "client connection should be closed")
	conn.Close()

	// listenerDoneCh should be closed
	select {
	case <-listenerDoneCh:
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("listenerDoneCh was not closed")
	}
}

func TestListen_ContextCancellation(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, listenerDoneCh, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	cancel()

	select {
	case <-listenerDoneCh:
		// OK - accept loop exited
	case <-time.After(5 * time.Second):
		t.Fatal("listenerDoneCh was not closed after context cancellation")
	}

	// New connections should be refused
	_, dialErr := net.Dial("unix", socketPath)
	assert.Error(t, dialErr, "new connections should be refused after context cancellation")
}

// --- State Transitions ---

func TestPerConnection_SingleRequestResponse(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	conn, dialErr := net.Dial("unix", socketPath)
	require.NoError(t, dialErr)

	// Send first message
	conn.Write([]byte(`{"type":"event","payload":{"first":true}}` + "\n"))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	_, err1 := reader.ReadString('\n')
	require.NoError(t, err1)

	// Attempt second message
	conn.Write([]byte(`{"type":"event","payload":{"second":true}}` + "\n"))

	// Verify that only one handler call occurs within a reasonable window
	require.Never(t, func() bool { return handler.callCount() > 1 }, 500*time.Millisecond, 10*time.Millisecond,
		"only first message should be processed")
	assert.Equal(t, 1, handler.callCount())
	conn.Close()
}

// --- Asynchronous Flow ---

func TestListen_ListenerErrChannel(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen in storage/runtime_socket_manager.go")

	projectRoot, socketPath := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	listenerErrCh, _, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	// Simulate accept failure by closing the underlying listener externally.
	// This requires closing the socket file which forces the accept loop to fail.
	// The exact mechanism depends on production internals — removing the socket file
	// may trigger the accept loop to return an error.
	os.Remove(socketPath)

	// Force an accept attempt by trying to connect
	net.Dial("unix", socketPath)

	select {
	case err := <-listenerErrCh:
		assert.Contains(t, err.Error(), "listener accept loop failed:")
	case <-time.After(5 * time.Second):
		t.Fatal("expected error on listenerErrCh")
	}
}

func TestListen_DoneChannelClosedAfterDelete(t *testing.T) {
	t.Skip("scaffolded: awaiting production surface NewRuntimeSocketManager, CreateSocket, Listen, DeleteSocket in storage/runtime_socket_manager.go")

	projectRoot, _ := makeSocketTestDir(t)
	ml := newMockLogger()
	mgr := NewRuntimeSocketManager(projectRoot, testSessionUUID, ml)
	require.NoError(t, mgr.CreateSocket())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := newMockMessageHandler(*entities.SuccessResponse("ok"))
	_, listenerDoneCh, err := mgr.Listen(ctx, handler)
	require.NoError(t, err)

	mgr.DeleteSocket(ctx)

	select {
	case <-listenerDoneCh:
		// OK - done channel is closed
	case <-time.After(5 * time.Second):
		t.Fatal("listenerDoneCh was not closed after DeleteSocket")
	}
}
