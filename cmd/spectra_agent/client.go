package spectra_agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// RuntimeResponse represents the response from the Runtime socket server.
type RuntimeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StorageLayout defines the interface for getting storage paths.
type StorageLayout interface {
	GetRuntimeSocketPath(projectRoot, sessionID string) (string, error)
}

// DefaultStorageLayout implements the default storage layout.
type DefaultStorageLayout struct{}

// GetRuntimeSocketPath returns the path to the runtime socket.
func (d *DefaultStorageLayout) GetRuntimeSocketPath(projectRoot, sessionID string) (string, error) {
	return storage.GetRuntimeSocketPath(projectRoot, sessionID), nil
}

// SocketClient handles socket communication with the Runtime.
type SocketClient struct {
	layout  StorageLayout
	timeout time.Duration
}

// NewSocketClient creates a new SocketClient with default settings.
func NewSocketClient() *SocketClient {
	return &SocketClient{
		layout:  &DefaultStorageLayout{},
		timeout: 30 * time.Second,
	}
}

// NewSocketClientWithTimeout creates a new SocketClient with a custom timeout.
func NewSocketClientWithTimeout(timeout time.Duration) *SocketClient {
	return &SocketClient{
		layout:  &DefaultStorageLayout{},
		timeout: timeout,
	}
}

// NewSocketClientWithLayout creates a new SocketClient with a custom storage layout.
func NewSocketClientWithLayout(layout StorageLayout) *SocketClient {
	return &SocketClient{
		layout:  layout,
		timeout: 30 * time.Second,
	}
}

// Send sends a RuntimeMessage to the socket and returns the response.
// Returns (response, exitCode, error).
// Exit codes: 0 = success, 2 = transport error, 3 = runtime error
func (sc *SocketClient) Send(sessionID, projectRoot string, msg entities.RuntimeMessage) (*RuntimeResponse, int, error) {
	// Get socket path
	socketPath, err := sc.layout.GetRuntimeSocketPath(projectRoot, sessionID)
	if err != nil {
		return nil, 2, fmt.Errorf("Error: failed to get socket path: %w", err)
	}

	// Check if socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, 2, fmt.Errorf("Error: socket file not found: %s", socketPath)
	}

	// Connect to socket with timeout
	conn, err := net.DialTimeout("unix", socketPath, sc.timeout)
	if err != nil {
		// Check if it's a connection refused error
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Timeout() {
				return nil, 2, fmt.Errorf("Error: connection timeout after %v", sc.timeout)
			}
		}
		// Check if socket file still exists (might have been deleted)
		if _, statErr := os.Stat(socketPath); os.IsNotExist(statErr) {
			return nil, 2, fmt.Errorf("Error: socket file not found: %s", socketPath)
		}
		// Connection refused
		return nil, 2, fmt.Errorf("Error: connection refused: Runtime is not running for session %s", sessionID)
	}

	// Ensure connection is closed
	var closeErr error
	defer func() {
		closeErr = conn.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close socket: %v\n", closeErr)
		}
	}()

	// Set deadline for the entire operation
	if err := conn.SetDeadline(time.Now().Add(sc.timeout)); err != nil {
		return nil, 2, fmt.Errorf("Error: failed to set deadline: %w", err)
	}

	// Serialize message to JSON
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, 2, fmt.Errorf("Error: failed to serialize message: %w", err)
	}

	// Send message with newline terminator
	msgBytes = append(msgBytes, '\n')
	if _, err := conn.Write(msgBytes); err != nil {
		return nil, 2, fmt.Errorf("Error: failed to send message: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			// Check if it's a timeout
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, 2, fmt.Errorf("Error: connection timeout after %v", sc.timeout)
			}
			return nil, 2, fmt.Errorf("Error: failed to read response: %w", err)
		}
		// Connection closed without data
		return nil, 2, fmt.Errorf("Error: failed to read response: connection closed by Runtime")
	}

	responseText := scanner.Text()

	// Parse response JSON
	var response RuntimeResponse
	if err := json.Unmarshal([]byte(responseText), &response); err != nil {
		return nil, 3, fmt.Errorf("Error: malformed response from Runtime: %w", err)
	}

	// Validate response structure
	if response.Status == "" {
		return nil, 3, fmt.Errorf("Error: response missing 'status' field")
	}

	if response.Status != "success" && response.Status != "error" {
		return nil, 3, fmt.Errorf("Error: invalid response status '%s'", response.Status)
	}

	// Check response status
	if response.Status == "error" {
		return &response, 3, fmt.Errorf("Error: %s", response.Message)
	}

	return &response, 0, nil
}
