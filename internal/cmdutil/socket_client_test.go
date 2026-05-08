package cmdutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Socket test helpers ---

// testSocketServer creates a temporary Unix domain socket listener and returns
// the socket path and a cleanup function. The handler is called for each accepted connection.
func testSocketServer(t *testing.T, handler func(conn net.Conn)) (socketPath string, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	socketPath = filepath.Join(dir, "runtime.sock")

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	var mu sync.Mutex
	var acceptedConn net.Conn

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return // listener closed
		}
		mu.Lock()
		acceptedConn = conn
		mu.Unlock()
		defer conn.Close()
		handler(conn)
	}()

	cleanup = func() {
		listener.Close()
		// Close accepted connection to unblock handlers doing I/O.
		mu.Lock()
		if acceptedConn != nil {
			acceptedConn.Close()
		}
		mu.Unlock()
		// Bounded wait: do not block forever if handler never returns.
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
		}
	}
	t.Cleanup(cleanup)
	return socketPath, cleanup
}

// testSocketServerRespond creates a socket that reads one line and responds with the given response string.
func testSocketServerRespond(t *testing.T, response string) string {
	t.Helper()
	socketPath, _ := testSocketServer(t, func(conn net.Conn) {
		// Read the incoming message (one line)
		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		// Write response
		fmt.Fprint(conn, response)
	})
	return socketPath
}

// testSocketServerCapture creates a socket that captures the received bytes and responds with success.
func testSocketServerCapture(t *testing.T) (socketPath string, captured *[]byte) {
	t.Helper()
	var mu sync.Mutex
	var buf []byte
	captured = &buf

	socketPath, _ = testSocketServer(t, func(conn net.Conn) {
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			mu.Lock()
			// Include the newline that was consumed by scanner
			buf = append([]byte(scanner.Text()), '\n')
			mu.Unlock()
		}
		// Respond with success
		fmt.Fprint(conn, `{"status":"success","message":"ok"}`+"\n")
	})
	return socketPath, captured
}

// testSocketServerEOFDetect creates a socket that responds with success and then
// detects whether the client closed the connection (EOF).
func testSocketServerEOFDetect(t *testing.T) (socketPath string, closed *bool) {
	t.Helper()
	var clientClosed bool
	closed = &clientClosed

	socketPath, _ = testSocketServer(t, func(conn net.Conn) {
		// Read incoming message
		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		// Respond with success
		fmt.Fprint(conn, `{"status":"success","message":"ok"}`+"\n")
		// Wait briefly for the client to close
		time.Sleep(50 * time.Millisecond)
		// Try to read — should get EOF if client closed
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			clientClosed = true
		}
	})
	return socketPath, closed
}

// testSocketServerEOFDetectOnMalformed creates a socket that responds with malformed data
// and detects whether the client closed the connection.
func testSocketServerEOFDetectOnMalformed(t *testing.T) (socketPath string, closed *bool) {
	t.Helper()
	var clientClosed bool
	closed = &clientClosed

	socketPath, _ = testSocketServer(t, func(conn net.Conn) {
		// Read incoming message
		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		// Respond with malformed data
		fmt.Fprint(conn, "not json\n")
		// Wait briefly for the client to close
		time.Sleep(50 * time.Millisecond)
		// Try to read — should get EOF if client closed
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			clientClosed = true
		}
	})
	return socketPath, closed
}

// callSend is a test helper that calls the production Send function.
func callSend(t *testing.T, layout storageLayoutProvider, sessionID, projectRoot string, message []byte, opts ...sendOption) (*Response, int, error) {
	t.Helper()
	return Send(layout, sessionID, projectRoot, message, opts...)
}

// sendOption is now defined in socket_client.go (production code).

// --- Happy Path — Send ---

func TestSend_SuccessResponse(t *testing.T) {
	socketPath := testSocketServerRespond(t, `{"status":"success","message":"ok"}`+"\n")
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{"type":"event"}`))

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	require.NotNil(t, response)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "ok", response.Message)
}

func TestSend_ErrorStatusResponse(t *testing.T) {
	socketPath := testSocketServerRespond(t, `{"status":"error","message":"session not found"}`+"\n")
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-2", "/tmp/project", []byte(`{"type":"event"}`))

	require.NoError(t, err)
	assert.Equal(t, ExitRuntimeError, exitCode)
	require.NotNil(t, response)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "session not found", response.Message)
}

func TestSend_SendsMessageWithNewline(t *testing.T) {
	socketPath, captured := testSocketServerCapture(t)
	layout := &mockStorageLayout{socketPath: socketPath}

	_, exitCode, _ := callSend(t, layout, "sess-3", "/tmp/project", []byte(`{"type":"event"}`))

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, []byte("{\"type\":\"event\"}\n"), *captured)
}

// --- Error Propagation ---

func TestSend_SocketFileNotFound(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent", "runtime.sock")
	layout := &mockStorageLayout{socketPath: nonExistentPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitTransportError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "socket file not found")
}

func TestSend_ConnectionRefused(t *testing.T) {
	// Create a regular file (not a socket) at the expected path
	dir := t.TempDir()
	fakePath := filepath.Join(dir, "runtime.sock")
	err := os.WriteFile(fakePath, []byte("not a socket"), 0644)
	require.NoError(t, err)
	layout := &mockStorageLayout{socketPath: fakePath}

	response, exitCode, connErr := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitTransportError, exitCode)
	require.Error(t, connErr)
	assert.Contains(t, connErr.Error(), "connection refused")
}

func TestSend_ConnectionTimeout(t *testing.T) {
	// Create a socket listener that accepts but never responds
	socketPath, _ := testSocketServer(t, func(conn net.Conn) {
		// Block forever (until test cleanup closes the listener)
		select {}
	})
	layout := &mockStorageLayout{socketPath: socketPath}

	// Use a short timeout to avoid real waiting
	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`), sendOption{timeout: 50 * time.Millisecond})

	assert.Nil(t, response)
	assert.Equal(t, ExitTransportError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection timeout")
}

func TestSend_WriteFails(t *testing.T) {
	// Create a socket listener that accepts and immediately closes the connection
	socketPath, _ := testSocketServer(t, func(conn net.Conn) {
		conn.Close()
	})
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitTransportError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message")
}

func TestSend_ReadFails(t *testing.T) {
	// Create a socket listener that accepts, reads the message, then closes without responding
	socketPath, _ := testSocketServer(t, func(conn net.Conn) {
		scanner := bufio.NewScanner(conn)
		scanner.Scan() // read the message
		conn.Close()   // close without responding
	})
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitTransportError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read response")
}

func TestSend_MalformedJSON(t *testing.T) {
	socketPath := testSocketServerRespond(t, "not json\n")
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitRuntimeError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed response from Runtime")
}

func TestSend_MissingStatusField(t *testing.T) {
	socketPath := testSocketServerRespond(t, `{"message":"hello"}`+"\n")
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitRuntimeError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "response missing 'status' field")
}

func TestSend_InvalidStatusValue(t *testing.T) {
	socketPath := testSocketServerRespond(t, `{"status":"unknown"}`+"\n")
	layout := &mockStorageLayout{socketPath: socketPath}

	response, exitCode, err := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitRuntimeError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response status 'unknown'")
}

// --- Resource Cleanup ---

func TestSend_ClosesConnectionOnSuccess(t *testing.T) {
	socketPath, closed := testSocketServerEOFDetect(t)
	layout := &mockStorageLayout{socketPath: socketPath}

	_, exitCode, _ := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	// Allow time for the server goroutine to detect EOF
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, exitCode)
	assert.True(t, *closed, "expected server to detect client connection closure (EOF)")
}

func TestSend_ClosesConnectionOnError(t *testing.T) {
	socketPath, closed := testSocketServerEOFDetectOnMalformed(t)
	layout := &mockStorageLayout{socketPath: socketPath}

	_, exitCode, _ := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	// Allow time for the server goroutine to detect EOF
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, ExitRuntimeError, exitCode)
	assert.True(t, *closed, "expected server to detect client connection closure after error")
}

func TestSend_CloseFailureWarning(t *testing.T) {
	// This test requires a connection wrapper seam where Close() returns an error.
	// The production code needs to support injecting a connection wrapper for this test.
	// Missing seam: connection wrapper injection in socket_client.go (e.g., connDialer option or connWrapper interface)
	t.Skip("scaffolded: requires connection wrapper seam in socket_client.go to inject Close() error")

	socketPath := testSocketServerRespond(t, `{"status":"success","message":"ok"}`+"\n")
	layout := &mockStorageLayout{socketPath: socketPath}

	// Once the seam exists, inject a connection wrapper where Close() returns an error
	// and capture stderr output to verify the warning message.
	_, exitCode, _ := callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	assert.Equal(t, 0, exitCode)
	// assert stderr contains "Warning: failed to close socket"
}

// --- Boundary Values — sessionID ---

func TestSend_MalformedSessionID(t *testing.T) {
	dir := t.TempDir()
	// Stub returns a path derived from the malformed ID (which does not exist)
	malformedPath := filepath.Join(dir, ".spectra", "sessions", "not-a-uuid", "runtime.sock")
	layout := &mockStorageLayout{socketPath: malformedPath}

	response, exitCode, err := callSend(t, layout, "not-a-uuid", "/tmp/project", []byte(`{}`))

	assert.Nil(t, response)
	assert.Equal(t, ExitTransportError, exitCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "socket file not found")
}

// --- Mock / Dependency Interaction ---

func TestSend_CallsGetRuntimeSocketPath(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent.sock")
	layout := &mockStorageLayout{socketPath: nonExistentPath}

	callSend(t, layout, "sess-1", "/tmp/project", []byte(`{}`))

	calls := layout.calls()
	require.Len(t, calls, 1)
	assert.Equal(t, "/tmp/project", calls[0].projectRoot)
	assert.Equal(t, "sess-1", calls[0].sessionID)
}

// --- Suppress unused import warnings ---
var (
	_ = json.Marshal
	_ = os.WriteFile
	_ = net.Listen
	_ = time.Sleep
	_ = fmt.Fprint
)
