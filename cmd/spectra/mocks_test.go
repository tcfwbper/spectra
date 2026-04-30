package main

import "io"

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
type MockRuntime struct {
	runCalled    bool
	projectRoot  string
	workflowName string
	exitCode     int
	err          error
	stdoutFunc   func(w io.Writer)
	stderrFunc   func(w io.Writer)
	signalCh     chan struct{}
}

// NewMockRuntime creates a new mock Runtime that returns the given exit code.
func NewMockRuntime(exitCode int) *MockRuntime {
	return &MockRuntime{exitCode: exitCode}
}

// NewMockRuntimeWithError creates a new mock Runtime that returns an error.
func NewMockRuntimeWithError(err error) *MockRuntime {
	return &MockRuntime{exitCode: 1, err: err}
}

// NewMockRuntimeWithStdout creates a new mock Runtime that writes to stdout.
func NewMockRuntimeWithStdout(exitCode int, stdoutFunc func(w io.Writer)) *MockRuntime {
	return &MockRuntime{exitCode: exitCode, stdoutFunc: stdoutFunc}
}

// NewMockRuntimeWithStderr creates a new mock Runtime that writes to stderr.
func NewMockRuntimeWithStderr(exitCode int, stderrFunc func(w io.Writer)) *MockRuntime {
	return &MockRuntime{exitCode: exitCode, stderrFunc: stderrFunc}
}

// NewMockRuntimeWithStreams creates a new mock Runtime that writes to both stdout and stderr.
func NewMockRuntimeWithStreams(exitCode int, stdoutFunc func(w io.Writer), stderrFunc func(w io.Writer)) *MockRuntime {
	return &MockRuntime{exitCode: exitCode, stdoutFunc: stdoutFunc, stderrFunc: stderrFunc}
}

// Run records the call and returns the configured exit code.
func (m *MockRuntime) Run(projectRoot, workflowName string, stdout, stderr io.Writer) (int, error) {
	m.runCalled = true
	m.projectRoot = projectRoot
	m.workflowName = workflowName
	if m.stdoutFunc != nil {
		m.stdoutFunc(stdout)
	}
	if m.stderrFunc != nil {
		m.stderrFunc(stderr)
	}
	if m.signalCh != nil {
		<-m.signalCh
	}
	return m.exitCode, m.err
}

// RunCalled returns whether Run was called.
func (m *MockRuntime) RunCalled() bool {
	return m.runCalled
}

// ProjectRoot returns the project root passed to Run.
func (m *MockRuntime) ProjectRoot() string {
	return m.projectRoot
}

// WorkflowName returns the workflow name passed to Run.
func (m *MockRuntime) WorkflowName() string {
	return m.workflowName
}

// SetSignalCh sets a channel that Run will block on before returning.
func (m *MockRuntime) SetSignalCh(ch chan struct{}) {
	m.signalCh = ch
}
