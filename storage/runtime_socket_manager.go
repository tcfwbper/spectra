package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/tcfwbper/spectra/entities"
)

const maxMessageSize = 10 * 1024 * 1024 // 10 MB

// MessageHandler is the callback function for processing messages
type MessageHandler func(sessionUUID string, message entities.RuntimeMessage) entities.RuntimeResponse

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
	var msg entities.RuntimeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		// Malformed JSON
		return
	}

	// Validate message structure
	if err := msg.Validate(); err != nil {
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
