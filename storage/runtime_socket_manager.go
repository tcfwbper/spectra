package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

const maxMessageSize = 10 * 1024 * 1024 // 10 MB

// RuntimeMessage represents a message received from spectra-agent
type RuntimeMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// RuntimeResponse represents the response sent back to spectra-agent
type RuntimeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MessageHandler is the callback function for processing messages
type MessageHandler func(sessionUUID string, message RuntimeMessage) RuntimeResponse

// RuntimeSocketManager manages the runtime socket lifecycle for a session
type RuntimeSocketManager struct {
	projectRoot string
	sessionUUID string
	listener    net.Listener
	mu          sync.Mutex
	done        chan struct{}
	closeOnce   sync.Once
}

// NewRuntimeSocketManager creates a new RuntimeSocketManager
func NewRuntimeSocketManager(projectRoot string, sessionUUID string) *RuntimeSocketManager {
	return &RuntimeSocketManager{
		projectRoot: projectRoot,
		sessionUUID: sessionUUID,
	}
}

// CreateSocket creates the runtime socket file
func (m *RuntimeSocketManager) CreateSocket() error {
	socketPath := GetRuntimeSocketPath(m.projectRoot, m.sessionUUID)

	// Check if socket file already exists
	if _, err := os.Stat(socketPath); err == nil {
		return fmt.Errorf("runtime socket file already exists: %s. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm %s", socketPath, socketPath)
	}

	// Create the Unix socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create runtime socket: %w", err)
	}

	// Set socket file permissions to 0600
	if err := os.Chmod(socketPath, 0600); err != nil {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	m.mu.Lock()
	m.listener = listener
	m.done = make(chan struct{})
	m.mu.Unlock()

	return nil
}

// Listen starts listening for connections on the socket
func (m *RuntimeSocketManager) Listen(handler MessageHandler) (<-chan error, <-chan struct{}, error) {
	m.mu.Lock()
	if m.listener == nil {
		m.mu.Unlock()
		return nil, nil, fmt.Errorf("runtime socket not created: call CreateSocket() first")
	}
	listener := m.listener
	done := m.done
	m.mu.Unlock()

	// Create buffered error channel (capacity 1)
	errCh := make(chan error, 1)

	// Start accept loop in goroutine
	go m.acceptLoop(listener, handler, errCh, done)

	return errCh, done, nil
}

// acceptLoop handles incoming connections
func (m *RuntimeSocketManager) acceptLoop(listener net.Listener, handler MessageHandler, errCh chan error, done chan struct{}) {
	defer close(done)

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Listener was closed, exit gracefully
			return
		}

		// Handle each connection in a separate goroutine
		go m.handleConnection(conn, handler)
	}
}

// handleConnection processes a single client connection
func (m *RuntimeSocketManager) handleConnection(conn net.Conn, handler MessageHandler) {
	defer func() { _ = conn.Close() }()

	// Read message with size limit
	reader := io.LimitReader(conn, maxMessageSize+1)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), maxMessageSize+1)

	if !scanner.Scan() {
		if scanner.Err() != nil {
			// Check if error is due to size limit
			if scanner.Err() == bufio.ErrTooLong {
				// Message exceeds size limit
				return
			}
		}
		// Connection closed or EOF
		return
	}

	data := scanner.Bytes()
	if len(data) > maxMessageSize {
		// Message exceeds 10 MB limit
		return
	}

	// Parse JSON message
	var msg RuntimeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		// Malformed JSON
		return
	}

	// Validate message structure
	if !m.validateMessage(&msg) {
		return
	}

	// Invoke message handler
	response := handler(m.sessionUUID, msg)

	// Serialize and send response
	responseData, err := json.Marshal(response)
	if err != nil {
		return
	}

	// Send response with newline
	responseData = append(responseData, '\n')
	_, _ = conn.Write(responseData)
}

// validateMessage validates the message structure and required fields
func (m *RuntimeSocketManager) validateMessage(msg *RuntimeMessage) bool {
	// Check required fields
	if msg.Type == "" {
		return false
	}

	if msg.Payload == nil {
		return false
	}

	// Validate message type
	if msg.Type != "event" && msg.Type != "error" {
		return false
	}

	// Validate type-specific requirements
	if msg.Type == "event" {
		eventType, ok := msg.Payload["eventType"]
		if !ok {
			return false
		}
		// eventType must be a string
		if _, ok := eventType.(string); !ok {
			return false
		}
	}

	if msg.Type == "error" {
		message, ok := msg.Payload["message"]
		if !ok {
			return false
		}
		// message must be a non-empty string
		messageStr, ok := message.(string)
		if !ok || messageStr == "" {
			return false
		}
	}

	// Validate claudeSessionID if present
	if claudeSessionID, ok := msg.Payload["claudeSessionID"]; ok {
		if _, ok := claudeSessionID.(string); !ok {
			return false
		}
	}

	return true
}

// DeleteSocket stops listening and removes the socket file
func (m *RuntimeSocketManager) DeleteSocket() error {
	m.mu.Lock()
	listener := m.listener
	m.mu.Unlock()

	// Close listener if it exists
	if listener != nil {
		m.closeOnce.Do(func() {
			// Closing the listener causes Accept() to fail
			// which makes acceptLoop exit and close the done channel
			_ = listener.Close()
		})
	}

	// Remove socket file
	socketPath := GetRuntimeSocketPath(m.projectRoot, m.sessionUUID)
	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		// Log warning but don't return error
		// In production, this would log to a logger
		_ = err
	}

	return nil
}
