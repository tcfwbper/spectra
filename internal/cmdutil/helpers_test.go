package cmdutil

import (
	"sync"
)

// --- Interfaces for testability ---

// socketClientSender defines the interface for SocketClient's Send method.
// Production code should implement this interface to allow mock injection in tests.
type socketClientSender interface {
	Send(sessionID, projectRoot string, message []byte) (*Response, int, error)
}

// errorFormatterFunc defines the interface for error formatting.
// Production code should accept this as a dependency to allow mock injection in tests.
type errorFormatterFunc func(msg string) string

// --- Response type expected from SocketClient ---

// Response represents the parsed JSON response from the Runtime socket.
// This type must be defined in the production socket_client.go once it exists.
// Until then, it lives here to allow test compilation.
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// --- Mock: SocketClient ---

// mockSocketClient records calls to Send and returns configured results.
type mockSocketClient struct {
	mu sync.Mutex

	// Return values
	response *Response
	exitCode int
	err      error

	// Captured arguments
	calledWith []mockSocketClientCall
}

type mockSocketClientCall struct {
	sessionID   string
	projectRoot string
	message     []byte
}

func (m *mockSocketClient) Send(sessionID, projectRoot string, message []byte) (*Response, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calledWith = append(m.calledWith, mockSocketClientCall{
		sessionID:   sessionID,
		projectRoot: projectRoot,
		message:     message,
	})
	return m.response, m.exitCode, m.err
}

func (m *mockSocketClient) calls() []mockSocketClientCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]mockSocketClientCall(nil), m.calledWith...)
}

// --- Mock: ErrorFormatter ---

// mockErrorFormatter records calls to FormatError and returns configured results.
type mockErrorFormatter struct {
	mu sync.Mutex

	// Return value
	result string

	// Captured arguments
	calledWith []string
}

func (m *mockErrorFormatter) FormatError(msg string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calledWith = append(m.calledWith, msg)
	return m.result
}

func (m *mockErrorFormatter) calls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.calledWith...)
}

// --- Mock: StorageLayout ---

// mockStorageLayout records calls to GetRuntimeSocketPath and returns configured path.
type mockStorageLayout struct {
	mu sync.Mutex

	// Return value
	socketPath string

	// Captured arguments
	calledWith []mockStorageLayoutCall
}

type mockStorageLayoutCall struct {
	projectRoot string
	sessionID   string
}

func (m *mockStorageLayout) GetRuntimeSocketPath(projectRoot, sessionID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calledWith = append(m.calledWith, mockStorageLayoutCall{
		projectRoot: projectRoot,
		sessionID:   sessionID,
	})
	return m.socketPath
}

func (m *mockStorageLayout) calls() []mockStorageLayoutCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]mockStorageLayoutCall(nil), m.calledWith...)
}

// --- storageLayoutProvider interface ---

// storageLayoutProvider defines the interface for storage path resolution.
// Production code should implement this to allow test injection.
type storageLayoutProvider interface {
	GetRuntimeSocketPath(projectRoot, sessionID string) string
}

// --- Test message structs ---

// validStruct is a simple JSON-serializable struct for test input.
type validStruct struct {
	Type string `json:"type,omitempty"`
}

// unserializableStruct contains a channel field that cannot be serialized to JSON.
type unserializableStruct struct {
	Ch chan int `json:"ch"`
}

// testMsg is a struct matching the RuntimeMessage wire format for testing.
type testMsg struct {
	Type            string `json:"type"`
	ClaudeSessionID string `json:"claudeSessionID"`
}
