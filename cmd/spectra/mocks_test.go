package spectra

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
