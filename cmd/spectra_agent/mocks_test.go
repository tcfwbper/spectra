package spectra_agent

import (
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
