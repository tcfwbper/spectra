package spectra

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers ---

// setupRunTestFixture creates a temporary directory with .spectra/workflows/ directory and
// an optional workflow file. Returns the project root directory.
func setupRunTestFixture(t *testing.T, workflowName string) string {
	t.Helper()
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))
	if workflowName != "" {
		workflowFile := filepath.Join(workflowsDir, workflowName+".yaml")
		require.NoError(t, os.WriteFile(workflowFile, []byte("name: "+workflowName+"\n"), 0644))
	}
	return tmpDir
}

// setupRunTestFixtureNoSpectra creates a temporary directory without .spectra/.
func setupRunTestFixtureNoSpectra(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// setupRunTestFixtureSpectraOnly creates a temporary directory with .spectra/ but no workflow files.
func setupRunTestFixtureSpectraOnly(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	return tmpDir
}

// executeRunCommand creates and executes the run command with a mocked Runtime injected
// via WithRuntime option. It changes the working directory to workDir and returns stdout,
// stderr, and exit code.
func executeRunCommand(t *testing.T, workDir string, args []string, mockRT *MockRuntime) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"run"}, args...))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// =====================================================================
// Happy Path — Positional Argument
// =====================================================================

// TestRunCommand_PositionalArgument executes workflow when workflow name is provided as positional argument.
func TestRunCommand_PositionalArgument(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "TestWorkflow", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// TestRunCommand_PositionalArgumentFromSubdirectory executes workflow from subdirectory
// (Runtime handles project root location).
func TestRunCommand_PositionalArgumentFromSubdirectory(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "SimpleSdd")
	subDir := filepath.Join(projectRoot, "subdir", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, subDir, []string{"SimpleSdd"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "SimpleSdd", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Happy Path — Flag Argument
// =====================================================================

// TestRunCommand_FlagArgument executes workflow when workflow name is provided via --workflow flag.
func TestRunCommand_FlagArgument(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "MyWorkflow")
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"--workflow", "MyWorkflow"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "MyWorkflow", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// TestRunCommand_FlagPrecedenceOverPositional flag takes precedence when both flag and
// positional argument are provided.
func TestRunCommand_FlagPrecedenceOverPositional(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "FlagWorkflow")
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"--workflow", "FlagWorkflow", "PositionalWorkflow"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "FlagWorkflow", mockRT.WorkflowName(), "Flag should take precedence over positional argument")
	assert.NotEqual(t, "PositionalWorkflow", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Happy Path — Help Output
// =====================================================================

// TestRunCommand_HelpFlag displays help information when --help flag is provided.
func TestRunCommand_HelpFlag(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "--help"})

	exitCode := cmd.Execute()

	output := stdout.String()
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, output, "Run a workflow")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Flags:")
	assert.Contains(t, output, "--workflow string")
	assert.Contains(t, output, "Examples:")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called for --help")
}

// =====================================================================
// Happy Path — Runtime Exit Code Propagation
// =====================================================================

// TestRunCommand_RuntimeExitCodeZero propagates exit code 0 from Runtime.
func TestRunCommand_RuntimeExitCodeZero(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Test")
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Test"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Test", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// TestRunCommand_RuntimeExitCodeOne propagates exit code 1 from Runtime.
func TestRunCommand_RuntimeExitCodeOne(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Test")
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Test"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Test", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// =====================================================================
// Validation Failures — Missing Workflow Name
// =====================================================================

// TestRunCommand_NoWorkflowName returns error when no workflow name is provided.
func TestRunCommand_NoWorkflowName(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run"})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	errOut := stderr.String()
	assert.Contains(t, errOut, "workflow name")
	assert.Contains(t, errOut, "required")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_NoWorkflowNameWithFlag returns error when --workflow flag is provided without value.
func TestRunCommand_NoWorkflowNameWithFlag(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "--workflow"})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr.String())
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Validation Failures — Empty Workflow Name
// =====================================================================

// TestRunCommand_EmptyWorkflowNamePositional returns error when positional workflow name is empty string.
func TestRunCommand_EmptyWorkflowNamePositional(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", ""})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	errOut := stderr.String()
	assert.Contains(t, errOut, "workflow name")
	assert.Contains(t, errOut, "empty")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_EmptyWorkflowNameFlag returns error when --workflow flag value is empty string.
func TestRunCommand_EmptyWorkflowNameFlag(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "--workflow", ""})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	errOut := stderr.String()
	assert.Contains(t, errOut, "workflow name")
	assert.Contains(t, errOut, "empty")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_WhitespaceWorkflowName returns error when workflow name contains only whitespace.
func TestRunCommand_WhitespaceWorkflowName(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "   "})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	errOut := stderr.String()
	assert.Contains(t, errOut, "workflow name")
	assert.Contains(t, errOut, "empty")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Validation Failures — Too Many Arguments
// =====================================================================

// TestRunCommand_TooManyArguments returns error when multiple positional arguments are provided.
func TestRunCommand_TooManyArguments(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "Workflow1", "Workflow2"})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "too many arguments")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_TooManyArgumentsWithThree returns error when three positional arguments are provided.
func TestRunCommand_TooManyArgumentsWithThree(t *testing.T) {
	mockRT := NewMockRuntime(0)
	var stdout, stderr bytes.Buffer

	cmd := NewRootCommandWithHandlers(
		WithRunHandlerFunc(WithRuntime(mockRT)),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "Workflow1", "Workflow2", "Workflow3"})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "too many arguments")
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Error Propagation — Project Root Not Found
// =====================================================================

// TestRunCommand_RuntimeReportsProjectRootNotFound propagates exit code 1 when Runtime
// fails to locate project root.
func TestRunCommand_RuntimeReportsProjectRootNotFound(t *testing.T) {
	projectRoot := setupRunTestFixtureNoSpectra(t)
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "TestWorkflow", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// =====================================================================
// Error Propagation — Runtime Initialization Failures
// =====================================================================

// TestRunCommand_RuntimeReportsWorkflowNotFound propagates exit code 1 when Runtime
// cannot find workflow file.
func TestRunCommand_RuntimeReportsWorkflowNotFound(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"NonExistent"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "NonExistent", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// TestRunCommand_RuntimeReportsInvalidYAML propagates exit code 1 when Runtime
// encounters invalid YAML.
func TestRunCommand_RuntimeReportsInvalidYAML(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "")
	// Create malformed YAML file
	workflowFile := filepath.Join(projectRoot, ".spectra", "workflows", "Invalid.yaml")
	require.NoError(t, os.WriteFile(workflowFile, []byte("{{invalid yaml"), 0644))
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Invalid"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Invalid", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// TestRunCommand_RuntimeReportsWorkflowNotReadable propagates exit code 1 when Runtime
// cannot read workflow file.
func TestRunCommand_RuntimeReportsWorkflowNotReadable(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Restricted"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Restricted", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// =====================================================================
// Error Propagation — Runtime Execution Failures
// =====================================================================

// TestRunCommand_RuntimeReportsAgentError propagates exit code 1 when Runtime fails
// due to agent error.
func TestRunCommand_RuntimeReportsAgentError(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "AgentFail")
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"AgentFail"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "AgentFail", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// TestRunCommand_RuntimeReportsSessionLockError propagates exit code 1 when Runtime
// detects another session is running.
func TestRunCommand_RuntimeReportsSessionLockError(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(1)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Concurrent"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Concurrent", mockRT.WorkflowName())
	assert.Equal(t, 1, exitCode)
}

// =====================================================================
// Edge Cases — Special Workflow Names
// =====================================================================

// TestRunCommand_WorkflowNameWithPathSeparators passes workflow name with path separators
// to Runtime without validation.
func TestRunCommand_WorkflowNameWithPathSeparators(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(0)

	_, _, _ = executeRunCommand(t, projectRoot, []string{"../malicious/workflow"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "../malicious/workflow", mockRT.WorkflowName(),
		"Command should not validate or sanitize the workflow name")
}

// TestRunCommand_WorkflowNameWithSpecialCharacters passes workflow name with special
// characters to Runtime.
func TestRunCommand_WorkflowNameWithSpecialCharacters(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Work@flow#123"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Work@flow#123", mockRT.WorkflowName())
	assert.Equal(t, mockRT.exitCode, exitCode)
}

// TestRunCommand_WorkflowNameWithUnicode passes workflow name with Unicode characters to Runtime.
func TestRunCommand_WorkflowNameWithUnicode(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"工作流程"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "工作流程", mockRT.WorkflowName())
	assert.Equal(t, mockRT.exitCode, exitCode)
}

// =====================================================================
// Edge Cases — Signal Handling
// =====================================================================

// TestRunCommand_PropagatesSIGINT propagates SIGINT signal to Runtime subprocess.
func TestRunCommand_PropagatesSIGINT(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")
	mockRT := NewMockRuntime(1)
	signalCh := make(chan struct{})
	mockRT.SetSignalCh(signalCh)

	done := make(chan int, 1)
	go func() {
		_, _, exitCode := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, mockRT)
		done <- exitCode
	}()

	// Allow time for the command to start and Runtime.Run to block
	time.Sleep(50 * time.Millisecond)

	// Send signal to mock Runtime (close the channel to unblock Run)
	close(signalCh)

	exitCode := <-done
	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.True(t, mockRT.SignalReceived(), "Mock Runtime should record signal delivery")
	assert.Equal(t, mockRT.exitCode, exitCode)
}

// =====================================================================
// Edge Cases — Blocking Runtime
// =====================================================================

// TestRunCommand_WaitsForRuntimeCompletion command waits for Runtime.Run to return before exiting.
func TestRunCommand_WaitsForRuntimeCompletion(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")
	mockRT := NewMockRuntime(0)
	mockRT.SetBlockDuration(150 * time.Millisecond)

	start := time.Now()
	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, mockRT)
	elapsed := time.Since(start)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.True(t, elapsed >= 100*time.Millisecond,
		"Command should wait for Runtime.Run to complete; elapsed: %v", elapsed)
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Idempotency
// =====================================================================

// TestRunCommand_MultipleInvocationsIndependent multiple sequential invocations are independent.
func TestRunCommand_MultipleInvocationsIndependent(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")

	// First invocation
	mockRT1 := NewMockRuntime(0)
	_, _, exitCode1 := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, mockRT1)

	assert.True(t, mockRT1.RunCalled(), "First Runtime.Run should be called")
	assert.Equal(t, "TestWorkflow", mockRT1.WorkflowName())
	assert.Equal(t, 0, exitCode1)

	// Second invocation
	mockRT2 := NewMockRuntime(0)
	_, _, exitCode2 := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, mockRT2)

	assert.True(t, mockRT2.RunCalled(), "Second Runtime.Run should be called")
	assert.Equal(t, "TestWorkflow", mockRT2.WorkflowName())
	assert.Equal(t, 0, exitCode2)
}

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

// TestRunCommand_PassesCorrectWorkflowName passes workflow name exactly as provided to Runtime.
func TestRunCommand_PassesCorrectWorkflowName(t *testing.T) {
	projectRoot := setupRunTestFixtureSpectraOnly(t)
	mockRT := NewMockRuntime(0)

	_, _, _ = executeRunCommand(t, projectRoot, []string{"My-Workflow_123"}, mockRT)

	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "My-Workflow_123", mockRT.WorkflowName(),
		"Workflow name should be passed exactly as provided, with no transformation")
}
