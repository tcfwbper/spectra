package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
)

// socketFileCheckInterval controls how often the accept loop checks for
// socket file removal. Injectable for testing.
var socketFileCheckInterval = 100 * time.Millisecond

// newTicker is a seam for time.NewTicker, allowing tests to inject custom tickers.
var newTicker = func(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

// MessageHandler is the interface for processing validated runtime messages.
// It is injected at Listen() call time and invoked once per validated connection.
type MessageHandler interface {
	Handle(sessionUUID string, msg entities.RuntimeMessage) entities.RuntimeResponse
}

// RuntimeSocketManager manages the lifecycle of a Unix domain socket for a
// single session. It creates, listens on, and deletes the runtime socket file.
type RuntimeSocketManager struct {
	socketPath  string
	sessionUUID string
	logger      logger.Logger

	mu       sync.Mutex
	created  bool
	listener net.Listener
	conns    map[net.Conn]struct{}
	closed   bool
}

// NewRuntimeSocketManager creates a new RuntimeSocketManager. It composes the
// socket path via StorageLayout and stores it internally. No I/O is performed.
func NewRuntimeSocketManager(projectRoot, sessionUUID string, l logger.Logger) *RuntimeSocketManager {
	return &RuntimeSocketManager{
		socketPath:  GetRuntimeSocketPath(projectRoot, sessionUUID),
		sessionUUID: sessionUUID,
		logger:      l,
		conns:       make(map[net.Conn]struct{}),
	}
}

// CreateSocket checks for an existing socket file and creates a new Unix domain
// socket file at the session-specific path with permissions 0600.
func (m *RuntimeSocketManager) CreateSocket() error {
	if _, err := os.Stat(m.socketPath); err == nil {
		return fmt.Errorf("runtime socket file already exists: %s. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm %s", m.socketPath, m.socketPath)
	}

	// Create the Unix domain socket by binding a listener.
	ln, err := net.Listen("unix", m.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create runtime socket: %w", err)
	}

	// Set permissions to 0600.
	if chErr := os.Chmod(m.socketPath, 0600); chErr != nil {
		ln.Close()
		os.Remove(m.socketPath)
		return fmt.Errorf("failed to create runtime socket: %w", chErr)
	}

	m.mu.Lock()
	m.created = true
	m.listener = ln
	m.mu.Unlock()

	return nil
}

// Listen starts accepting connections on the created socket. It returns channels
// for error reporting and completion signaling.
func (m *RuntimeSocketManager) Listen(ctx context.Context, handler MessageHandler) (<-chan error, <-chan struct{}, error) {
	m.mu.Lock()
	if !m.created {
		m.mu.Unlock()
		return nil, nil, fmt.Errorf("runtime socket not created: call CreateSocket() first")
	}
	ln := m.listener
	m.mu.Unlock()

	// Verify the listener is still valid (e.g., socket file wasn't removed externally).
	if ln == nil {
		return nil, nil, fmt.Errorf("failed to listen on runtime socket: listener not available")
	}

	// Probe the listener by checking socket file still exists.
	if _, err := os.Stat(m.socketPath); err != nil {
		ln.Close()
		m.mu.Lock()
		m.listener = nil
		m.mu.Unlock()
		return nil, nil, fmt.Errorf("failed to listen on runtime socket: %w", err)
	}

	m.mu.Lock()
	m.closed = false
	m.mu.Unlock()

	listenerErrCh := make(chan error, 1)
	listenerDoneCh := make(chan struct{})

	go m.acceptLoop(ctx, ln, handler, listenerErrCh, listenerDoneCh)

	return listenerErrCh, listenerDoneCh, nil
}

// DeleteSocket stops the listener, closes all active connections, and removes
// the socket file. It is idempotent.
func (m *RuntimeSocketManager) DeleteSocket(ctx context.Context) {
	m.mu.Lock()
	ln := m.listener
	m.listener = nil
	m.closed = true
	// Snapshot and clear active connections
	conns := make([]net.Conn, 0, len(m.conns))
	for c := range m.conns {
		conns = append(conns, c)
	}
	m.conns = make(map[net.Conn]struct{})
	m.mu.Unlock()

	// Close the listener to stop accept loop
	if ln != nil {
		ln.Close()
	}

	// Close all active connections
	for _, c := range conns {
		c.Close()
	}

	// Remove socket file
	if err := os.Remove(m.socketPath); err != nil {
		if !os.IsNotExist(err) {
			m.logger.Warn(fmt.Sprintf("failed to delete runtime socket: %v. The socket file may need to be manually removed.", err))
		}
	}
}

func (m *RuntimeSocketManager) acceptLoop(ctx context.Context, ln net.Listener, handler MessageHandler, errCh chan<- error, doneCh chan struct{}) {
	defer close(doneCh)

	// Monitor context cancellation to close the listener
	go func() {
		<-ctx.Done()
		m.mu.Lock()
		if m.listener == ln {
			m.listener = nil
			m.closed = true
		}
		m.mu.Unlock()
		ln.Close()
	}()

	// Monitor socket file existence. If removed externally, close the listener.
	go func() {
		ticker := newTicker(socketFileCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.mu.Lock()
				closed := m.closed
				m.mu.Unlock()
				if closed {
					return
				}
				if _, err := os.Stat(m.socketPath); err != nil && os.IsNotExist(err) {
					m.mu.Lock()
					if m.listener == ln {
						m.listener = nil
					}
					m.mu.Unlock()
					ln.Close()
					return
				}
			}
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			m.mu.Lock()
			closed := m.closed
			m.mu.Unlock()
			if closed {
				return
			}
			// Check if context was cancelled
			select {
			case <-ctx.Done():
				return
			default:
			}
			// Check if socket file was removed
			if _, statErr := os.Stat(m.socketPath); statErr != nil && os.IsNotExist(statErr) {
				errCh <- fmt.Errorf("listener accept loop failed: socket file removed externally")
				return
			}
			errCh <- fmt.Errorf("listener accept loop failed: %w", err)
			return
		}

		m.mu.Lock()
		m.conns[conn] = struct{}{}
		m.mu.Unlock()

		go m.handleConnection(conn, handler)
	}
}

func (m *RuntimeSocketManager) handleConnection(conn net.Conn, handler MessageHandler) {
	defer func() {
		// Recover from panics in MessageHandler to prevent crashing the accept loop.
		// No response is sent, no warning logged — the connection is simply closed.
		recover()
		conn.Close()
		m.mu.Lock()
		delete(m.conns, conn)
		m.mu.Unlock()
	}()

	// Read a single line (message terminated by newline)
	reader := bufio.NewReaderSize(conn, MaxPayloadSize+1)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		// If the error is because of line size exceeding buffer...
		if len(line) > MaxPayloadSize {
			m.logger.Warn("dropping connection: message size exceeds 10 MB limit")
			m.sendProtocolError(conn, "message size exceeds 10 MB limit")
			return
		}
		// Client closed without sending (or read error)
		// EOF is normal - no warning needed
		return
	}

	// Check message size
	if len(line) > MaxPayloadSize {
		m.logger.Warn("dropping connection: message size exceeds 10 MB limit")
		m.sendProtocolError(conn, "message size exceeds 10 MB limit")
		return
	}

	// Parse JSON
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		m.logger.Warn(fmt.Sprintf("dropping connection: malformed JSON: %v", err))
		m.sendProtocolError(conn, "malformed JSON")
		return
	}

	// Validate "type" field
	typeRaw, ok := raw["type"]
	if !ok {
		m.logger.Warn("dropping connection: missing required field \"type\"")
		m.sendProtocolError(conn, "missing required field \"type\"")
		return
	}

	var msgType string
	if err := json.Unmarshal(typeRaw, &msgType); err != nil {
		m.logger.Warn("dropping connection: invalid \"type\" field")
		m.sendProtocolError(conn, "invalid \"type\" field")
		return
	}

	if msgType != "event" && msgType != "error" {
		m.logger.Warn(fmt.Sprintf("dropping connection: unrecognized type %q", msgType))
		m.sendProtocolError(conn, fmt.Sprintf("unrecognized type %q", msgType))
		return
	}

	// Validate "payload" field
	payloadRaw, ok := raw["payload"]
	if !ok {
		m.logger.Warn("dropping connection: missing required field \"payload\"")
		m.sendProtocolError(conn, "missing required field \"payload\"")
		return
	}

	// Validate payload is a JSON object
	if !isJSONObject(payloadRaw) {
		m.logger.Warn("dropping connection: payload must be a JSON object")
		m.sendProtocolError(conn, "payload must be a JSON object")
		return
	}

	// Extract optional claudeSessionID
	var claudeSessionID string
	if csRaw, ok := raw["claudeSessionID"]; ok {
		json.Unmarshal(csRaw, &claudeSessionID)
	}

	// Construct RuntimeMessage via the entity constructor
	rtMsg, err := entities.NewRuntimeMessage(msgType, json.RawMessage(payloadRaw), claudeSessionID)
	if err != nil {
		m.logger.Warn(fmt.Sprintf("dropping connection: failed to construct RuntimeMessage: %v", err))
		m.sendProtocolError(conn, "internal validation error")
		return
	}

	// Invoke handler (no panic recovery per spec)
	resp := handler.Handle(m.sessionUUID, *rtMsg)

	// Check if socket was closed while handler was executing
	m.mu.Lock()
	closed := m.closed
	m.mu.Unlock()
	if closed {
		return
	}

	// Serialize and send response
	m.sendBusinessResponse(conn, &resp)
}

func (m *RuntimeSocketManager) sendProtocolError(conn net.Conn, message string) {
	resp := fmt.Sprintf(`{"status":"error","message":"%s"}`+"\n", escapeJSON(message))
	if _, err := conn.Write([]byte(resp)); err != nil {
		m.logger.Warn(fmt.Sprintf("failed to send response to client: %v", err))
	}
}

func (m *RuntimeSocketManager) sendBusinessResponse(conn net.Conn, resp *entities.RuntimeResponse) {
	data := fmt.Sprintf(`{"status":"%s","message":"%s"}`+"\n", escapeJSON(resp.Status()), escapeJSON(resp.Message()))
	if _, err := conn.Write([]byte(data)); err != nil {
		m.logger.Warn(fmt.Sprintf("failed to send response to client: %v", err))
	}
}

// escapeJSON escapes special characters for embedding in a JSON string value.
func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	// json.Marshal wraps in quotes, remove them
	return string(b[1 : len(b)-1])
}

// isJSONObject checks if the raw JSON is a JSON object (starts with '{').
func isJSONObject(raw json.RawMessage) bool {
	for _, b := range raw {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return true
		default:
			return false
		}
	}
	return false
}
