package cmdutil

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

// Response represents the parsed JSON response from the Runtime socket.
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// storageLayoutProvider defines the interface for storage path resolution.
type storageLayoutProvider interface {
	GetRuntimeSocketPath(projectRoot, sessionID string) string
}

// sendOption represents an optional configuration for the Send function.
type sendOption struct {
	timeout time.Duration
}

// defaultTimeout is the default timeout for the entire Send operation.
const defaultTimeout = 30 * time.Second

// Send connects to the Runtime Unix domain socket, sends a JSON message, and
// returns the parsed response. It enforces a timeout for the entire operation
// (connect + send + receive) and classifies errors into transport errors
// (exit code 2) or runtime errors (exit code 3).
func Send(layout storageLayoutProvider, sessionID, projectRoot string, message []byte, opts ...sendOption) (*Response, int, error) {
	// Determine timeout.
	timeout := defaultTimeout
	for _, opt := range opts {
		if opt.timeout > 0 {
			timeout = opt.timeout
		}
	}

	// Resolve socket path.
	socketPath := layout.GetRuntimeSocketPath(projectRoot, sessionID)

	// Check if socket file exists.
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, ExitTransportError, fmt.Errorf("socket file not found: %s", socketPath)
	}

	// Connect to the Unix domain socket with deadline.
	deadline := time.Now().Add(timeout)
	conn, err := net.DialTimeout("unix", socketPath, timeout)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, ExitTransportError, fmt.Errorf("connection timeout after %s", timeout)
		}
		return nil, ExitTransportError, fmt.Errorf("connection refused: Runtime is not running for session %s", sessionID)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close socket: %s\n", closeErr)
		}
	}()

	// Set deadline for the entire operation.
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, ExitTransportError, fmt.Errorf("failed to send message: %w", err)
	}

	// Write message followed by newline.
	data := append(message, '\n')
	if _, err := conn.Write(data); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ExitTransportError, fmt.Errorf("connection timeout after %s", timeout)
		}
		return nil, ExitTransportError, fmt.Errorf("failed to send message: %w", err)
	}

	// Read response (one newline-terminated JSON line).
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, ExitTransportError, fmt.Errorf("connection timeout after %s", timeout)
			}
			// If the connection was reset by peer, the message was never delivered.
			if isConnectionReset(err) {
				return nil, ExitTransportError, fmt.Errorf("failed to send message: %w", err)
			}
			return nil, ExitTransportError, fmt.Errorf("failed to read response: %w", err)
		}
		return nil, ExitTransportError, fmt.Errorf("failed to read response: connection closed")
	}

	responseData := scanner.Bytes()

	// Parse response JSON.
	var resp Response
	if err := json.Unmarshal(responseData, &resp); err != nil {
		return nil, ExitRuntimeError, fmt.Errorf("malformed response from Runtime: %w", err)
	}

	// Validate status field.
	if resp.Status == "" {
		return nil, ExitRuntimeError, fmt.Errorf("response missing 'status' field")
	}

	switch resp.Status {
	case "success":
		return &resp, ExitSuccess, nil
	case "error":
		return &resp, ExitRuntimeError, nil
	default:
		return nil, ExitRuntimeError, fmt.Errorf("invalid response status '%s'", resp.Status)
	}
}

// isConnectionReset returns true if the error indicates the connection was reset by peer.
func isConnectionReset(err error) bool {
	if errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	// Fallback for wrapped errors that may not unwrap to syscall.ECONNRESET.
	return strings.Contains(err.Error(), "connection reset by peer")
}
