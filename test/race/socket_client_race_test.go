package race_test

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra_agent "github.com/tcfwbper/spectra/cmd/spectra_agent"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

const (
	raceConcurrentSendCount        = 20
	raceConcurrentSameSessionCount = 10
)

// setupRaceClientTestFixture creates a temporary test directory with .spectra/sessions/<uuid>/ structure.
func setupRaceClientTestFixture(t *testing.T, sessionUUID string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "s")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	sessionDir := filepath.Join(tmpDir, ".spectra", "sessions", sessionUUID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	return tmpDir
}

// startRaceMockServer creates a mock socket server for race tests.
func startRaceMockServer(t *testing.T, socketPath string) (net.Listener, func()) {
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
				if scanner.Scan() {
					c.Write([]byte(`{"status":"success","message":"ok"}` + "\n"))
				}
			}(conn)
		}
	}()

	return listener, func() { listener.Close() }
}

func newRaceTestMessage() entities.RuntimeMessage {
	return entities.RuntimeMessage{
		Type:            "event",
		ClaudeSessionID: "test-session",
		Payload:         json.RawMessage(`{"eventType":"TestEvent","message":"race test","payload":{}}`),
	}
}

// TestSocketClient_ConcurrentSend multiple goroutines send messages concurrently with different sessions.
func TestSocketClient_ConcurrentSend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sessionID1 := uuid.New().String()
	sessionID2 := uuid.New().String()
	projectRoot1 := setupRaceClientTestFixture(t, sessionID1)
	projectRoot2 := setupRaceClientTestFixture(t, sessionID2)

	socketPath1 := storage.GetRuntimeSocketPath(projectRoot1, sessionID1)
	socketPath2 := storage.GetRuntimeSocketPath(projectRoot2, sessionID2)

	_, cleanup1 := startRaceMockServer(t, socketPath1)
	defer cleanup1()
	_, cleanup2 := startRaceMockServer(t, socketPath2)
	defer cleanup2()

	client := spectra_agent.NewSocketClient()

	var wg sync.WaitGroup
	exitCodes := make([]int, raceConcurrentSendCount)
	errors := make([]error, raceConcurrentSendCount)

	for i := 0; i < raceConcurrentSendCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			msg := newRaceTestMessage()
			var exitCode int
			var err error
			if idx%2 == 0 {
				_, exitCode, err = client.Send(sessionID1, projectRoot1, msg)
			} else {
				_, exitCode, err = client.Send(sessionID2, projectRoot2, msg)
			}
			exitCodes[idx] = exitCode
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	for i := 0; i < raceConcurrentSendCount; i++ {
		assert.Equal(t, 0, exitCodes[i], "goroutine %d should succeed", i)
		assert.NoError(t, errors[i], "goroutine %d should not error", i)
	}
}

// TestSocketClient_ConcurrentSendSameSession multiple goroutines send messages to same session concurrently.
func TestSocketClient_ConcurrentSendSameSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sessionID := uuid.New().String()
	projectRoot := setupRaceClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := startRaceMockServer(t, socketPath)
	defer cleanup()

	client := spectra_agent.NewSocketClient()

	var wg sync.WaitGroup
	exitCodes := make([]int, raceConcurrentSameSessionCount)
	errors := make([]error, raceConcurrentSameSessionCount)

	for i := 0; i < raceConcurrentSameSessionCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			msg := entities.RuntimeMessage{
				Type:            "event",
				ClaudeSessionID: "test-session",
				Payload:         json.RawMessage(`{"eventType":"TestEvent","message":"concurrent msg","payload":{}}`),
			}
			_, exitCode, err := client.Send(sessionID, projectRoot, msg)
			exitCodes[idx] = exitCode
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	for i := 0; i < raceConcurrentSameSessionCount; i++ {
		assert.Equal(t, 0, exitCodes[i], "goroutine %d should succeed", i)
		assert.NoError(t, errors[i], "goroutine %d should not error", i)
	}
}

// TestSocketClient_ConcurrentWithSocketDeletion send operations handle concurrent socket file deletion gracefully.
func TestSocketClient_ConcurrentWithSocketDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sessionID := uuid.New().String()
	projectRoot := setupRaceClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	listener, cleanup := startRaceMockServer(t, socketPath)
	_ = listener

	client := spectra_agent.NewSocketClient()

	var wg sync.WaitGroup

	// Start multiple goroutines sending messages
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := newRaceTestMessage()
			_, exitCode, _ := client.Send(sessionID, projectRoot, msg)
			// Either succeeds (0) or fails with transport error (2)
			assert.True(t, exitCode == 0 || exitCode == 2,
				"Expected exit code 0 or 2, got %d", exitCode)
		}()
	}

	// Delete the socket file midway
	time.Sleep(10 * time.Millisecond)
	cleanup()
	os.Remove(socketPath)

	wg.Wait()
	// No panics or data races expected
}

// TestSocketClient_ConcurrentConnectionAndCleanup socket creation, connection, and cleanup operations are race-free.
func TestSocketClient_ConcurrentConnectionAndCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	sessionID := uuid.New().String()
	projectRoot := setupRaceClientTestFixture(t, sessionID)
	socketPath := storage.GetRuntimeSocketPath(projectRoot, sessionID)

	_, cleanup := startRaceMockServer(t, socketPath)
	defer cleanup()

	client := spectra_agent.NewSocketClient()

	var wg sync.WaitGroup
	for i := 0; i < raceConcurrentSendCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := newRaceTestMessage()
			_, _, _ = client.Send(sessionID, projectRoot, msg)
		}()
	}

	wg.Wait()
	// No races detected during socket creation, connection, send, receive, or close operations
}
