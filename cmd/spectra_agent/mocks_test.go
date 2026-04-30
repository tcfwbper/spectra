package main

import (
	"sync"
	"time"

	"github.com/tcfwbper/spectra/entities"
)

// MockStorageLayout is a mock implementation of StorageLayout for testing.
type MockStorageLayout struct {
	GetRuntimeSocketPathFunc func(projectRoot, sessionID string) (string, error)
}

// GetRuntimeSocketPath calls the mock function.
func (m *MockStorageLayout) GetRuntimeSocketPath(projectRoot, sessionID string) (string, error) {
	if m.GetRuntimeSocketPathFunc != nil {
		return m.GetRuntimeSocketPathFunc(projectRoot, sessionID)
	}
	return "", nil
}

// MockSocketClient is a mock implementation of SocketClient for testing.
type MockSocketClient struct {
	called bool
}

// NewMockSocketClient creates a new mock socket client.
func NewMockSocketClient() *MockSocketClient {
	return &MockSocketClient{}
}

// Send marks the client as called.
func (m *MockSocketClient) Send(sessionID, projectRoot string, msg entities.RuntimeMessage) (*RuntimeResponse, int, error) {
	m.called = true
	return &RuntimeResponse{Status: "success", Message: "ok"}, 0, nil
}

// WasCalled returns whether Send was called.
func (m *MockSocketClient) WasCalled() bool {
	return m.called
}

// CapturingMockSocketClient is a mock SocketClient that captures Send parameters.
type CapturingMockSocketClient struct {
	mu          sync.Mutex
	called      bool
	callCount   int
	sessionID   string
	projectRoot string
	message     entities.RuntimeMessage
	resp        *RuntimeResponse
	exitCode    int
	err         error
}

// NewCapturingMockSocketClient creates a CapturingMockSocketClient with a success response.
func NewCapturingMockSocketClient() *CapturingMockSocketClient {
	return &CapturingMockSocketClient{
		resp:     &RuntimeResponse{Status: "success", Message: "ok"},
		exitCode: 0,
		err:      nil,
	}
}

// NewCapturingMockSocketClientWithResponse creates a CapturingMockSocketClient with a custom response.
func NewCapturingMockSocketClientWithResponse(resp *RuntimeResponse, exitCode int, err error) *CapturingMockSocketClient {
	return &CapturingMockSocketClient{
		resp:     resp,
		exitCode: exitCode,
		err:      err,
	}
}

// Send captures call parameters and returns the configured response.
func (m *CapturingMockSocketClient) Send(sessionID, projectRoot string, msg entities.RuntimeMessage) (*RuntimeResponse, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.callCount++
	m.sessionID = sessionID
	m.projectRoot = projectRoot
	m.message = msg
	return m.resp, m.exitCode, m.err
}

// WasCalled returns whether Send was called.
func (m *CapturingMockSocketClient) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

// CallCount returns how many times Send was called.
func (m *CapturingMockSocketClient) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// SessionID returns the captured session ID.
func (m *CapturingMockSocketClient) SessionID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessionID
}

// ProjectRoot returns the captured project root.
func (m *CapturingMockSocketClient) ProjectRoot() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.projectRoot
}

// Message returns the captured RuntimeMessage.
func (m *CapturingMockSocketClient) Message() entities.RuntimeMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.message
}

// WithSocketClientTimeout returns a CommandOption that sets a custom SocketClient timeout.
func WithSocketClientTimeout(timeout time.Duration) CommandOption {
	return func(cfg *rootCommandConfig) {
		cfg.socketClientTimeout = timeout
	}
}

// WithMockSocketClient returns a CommandOption that injects a mock SocketClient.
func WithMockSocketClient(client *CapturingMockSocketClient) CommandOption {
	return func(cfg *rootCommandConfig) {
		cfg.mockSocketClient = client
	}
}

// MockSubcommandHandler is a mock implementation for testing subcommand handlers.
type MockSubcommandHandler struct {
	called   bool
	exitCode int
}

// NewMockSubcommandHandler creates a new mock handler with exit code 0.
func NewMockSubcommandHandler() *MockSubcommandHandler {
	return &MockSubcommandHandler{exitCode: 0}
}

// NewMockSubcommandHandlerWithExitCode creates a new mock handler with a specific exit code.
func NewMockSubcommandHandlerWithExitCode(exitCode int) *MockSubcommandHandler {
	return &MockSubcommandHandler{exitCode: exitCode}
}

// Execute marks the handler as called and returns the configured exit code.
func (m *MockSubcommandHandler) Execute() int {
	m.called = true
	return m.exitCode
}

// WasCalled returns whether the handler was executed.
func (m *MockSubcommandHandler) WasCalled() bool {
	return m.called
}

// ExitCode returns the configured exit code.
func (m *MockSubcommandHandler) ExitCode() int {
	return m.exitCode
}

// MockSpectraFinder is a mock implementation of SpectraFinder for testing.
type MockSpectraFinder func(startDir string) (string, error)

// NewMockSpectraFinder creates a new mock SpectraFinder.
func NewMockSpectraFinder(fn func(startDir string) (string, error)) MockSpectraFinder {
	return MockSpectraFinder(fn)
}

// Find calls the mock function.
func (m MockSpectraFinder) Find(startDir string) (string, error) {
	return m(startDir)
}
