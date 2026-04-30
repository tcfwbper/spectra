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

// MockBuiltinResourceCopier is a mock implementation for testing BuiltinResourceCopier.
type MockBuiltinResourceCopier struct {
	copyWorkflowsCalled    bool
	copyAgentsCalled       bool
	copySpecFilesCalled    bool
	copyWorkflowsRoot      string
	copyAgentsRoot         string
	copySpecFilesRoot      string
	copyWorkflowsWarnings  []string
	copyAgentsWarnings     []string
	copySpecFilesWarnings  []string
	copyWorkflowsErr       error
	copyAgentsErr          error
	copySpecFilesErr       error
}

// NewMockBuiltinResourceCopier creates a new mock BuiltinResourceCopier.
func NewMockBuiltinResourceCopier() *MockBuiltinResourceCopier {
	return &MockBuiltinResourceCopier{}
}

// CopyWorkflows marks CopyWorkflows as called and records the project root.
func (m *MockBuiltinResourceCopier) CopyWorkflows(projectRoot string) ([]string, error) {
	m.copyWorkflowsCalled = true
	m.copyWorkflowsRoot = projectRoot
	return m.copyWorkflowsWarnings, m.copyWorkflowsErr
}

// CopyAgents marks CopyAgents as called and records the project root.
func (m *MockBuiltinResourceCopier) CopyAgents(projectRoot string) ([]string, error) {
	m.copyAgentsCalled = true
	m.copyAgentsRoot = projectRoot
	return m.copyAgentsWarnings, m.copyAgentsErr
}

// CopySpecFiles marks CopySpecFiles as called and records the project root.
func (m *MockBuiltinResourceCopier) CopySpecFiles(projectRoot string) ([]string, error) {
	m.copySpecFilesCalled = true
	m.copySpecFilesRoot = projectRoot
	return m.copySpecFilesWarnings, m.copySpecFilesErr
}

// CopyWorkflowsCalled returns whether CopyWorkflows was called.
func (m *MockBuiltinResourceCopier) CopyWorkflowsCalled() bool {
	return m.copyWorkflowsCalled
}

// CopyAgentsCalled returns whether CopyAgents was called.
func (m *MockBuiltinResourceCopier) CopyAgentsCalled() bool {
	return m.copyAgentsCalled
}

// CopySpecFilesCalled returns whether CopySpecFiles was called.
func (m *MockBuiltinResourceCopier) CopySpecFilesCalled() bool {
	return m.copySpecFilesCalled
}

// CopyWorkflowsProjectRoot returns the project root passed to CopyWorkflows.
func (m *MockBuiltinResourceCopier) CopyWorkflowsProjectRoot() string {
	return m.copyWorkflowsRoot
}

// CopyAgentsProjectRoot returns the project root passed to CopyAgents.
func (m *MockBuiltinResourceCopier) CopyAgentsProjectRoot() string {
	return m.copyAgentsRoot
}

// CopySpecFilesProjectRoot returns the project root passed to CopySpecFiles.
func (m *MockBuiltinResourceCopier) CopySpecFilesProjectRoot() string {
	return m.copySpecFilesRoot
}
