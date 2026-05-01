package main

import "time"

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
	copyWorkflowsCalled   bool
	copyAgentsCalled      bool
	copySpecFilesCalled   bool
	copyWorkflowsRoot     string
	copyAgentsRoot        string
	copySpecFilesRoot     string
	copyWorkflowsWarnings []string
	copyAgentsWarnings    []string
	copySpecFilesWarnings []string
	copyWorkflowsErr      error
	copyAgentsErr         error
	copySpecFilesErr      error
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

// MockSpectraFinder is a mock implementation of SpectraFinder for command tests.
type MockSpectraFinder struct {
	projectRoot string
	err         error
	called      bool
}

// NewMockSpectraFinder creates a new mock SpectraFinder that returns the given root.
func NewMockSpectraFinder(projectRoot string, err error) *MockSpectraFinder {
	return &MockSpectraFinder{projectRoot: projectRoot, err: err}
}

// Find returns the configured project root and error.
func (m *MockSpectraFinder) Find() (string, error) {
	m.called = true
	return m.projectRoot, m.err
}

// WasCalled returns whether Find was called.
func (m *MockSpectraFinder) WasCalled() bool {
	return m.called
}

// MockRuntime is a mock implementation of Runtime for command tests.
// The Runtime interface returns only an integer exit code from Run(workflowName).
type MockRuntime struct {
	runCalled      bool
	workflowName   string
	exitCode       int
	signalCh       chan struct{}
	signalReceived bool
	blockDuration  time.Duration
}

// NewMockRuntime creates a new mock Runtime that returns the given exit code.
func NewMockRuntime(exitCode int) *MockRuntime {
	return &MockRuntime{exitCode: exitCode}
}

// Run records the call and returns the configured exit code.
// If blockDuration is set, Run sleeps before returning.
// If signalCh is set, Run blocks until the channel is closed, then records signal receipt.
func (m *MockRuntime) Run(workflowName string) int {
	m.runCalled = true
	m.workflowName = workflowName
	if m.blockDuration > 0 {
		time.Sleep(m.blockDuration)
	}
	if m.signalCh != nil {
		<-m.signalCh
		m.signalReceived = true
	}
	return m.exitCode
}

// RunCalled returns whether Run was called.
func (m *MockRuntime) RunCalled() bool {
	return m.runCalled
}

// WorkflowName returns the workflow name passed to Run.
func (m *MockRuntime) WorkflowName() string {
	return m.workflowName
}

// SignalReceived returns whether the mock received a signal via its signal channel.
func (m *MockRuntime) SignalReceived() bool {
	return m.signalReceived
}

// SetSignalCh sets a channel that Run will block on before returning.
// When the channel is closed, Run records signal receipt and returns.
func (m *MockRuntime) SetSignalCh(ch chan struct{}) {
	m.signalCh = ch
}

// SetBlockDuration sets a duration that Run will sleep before returning.
func (m *MockRuntime) SetBlockDuration(d time.Duration) {
	m.blockDuration = d
}
