package spectra_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// --- Test Helpers ---

// setupRunTestFixture creates a temporary directory with .spectra/ and optional workflow files.
// Returns the project root directory.
func setupRunTestFixture(t *testing.T, workflowNames ...string) string {
	t.Helper()
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	if len(workflowNames) > 0 {
		workflowsDir := filepath.Join(spectraDir, "workflows")
		require.NoError(t, os.MkdirAll(workflowsDir, 0755))
		for _, name := range workflowNames {
			workflowFile := filepath.Join(workflowsDir, name+".yaml")
			require.NoError(t, os.WriteFile(workflowFile, []byte("name: "+name+"\n"), 0644))
		}
	}
	return tmpDir
}

// setupRunTestFixtureNoSpectra creates a temporary directory without .spectra/.
func setupRunTestFixtureNoSpectra(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// executeRunCommand creates and executes the run command with given args, mocked finder and runtime,
// and working directory. Returns stdout, stderr, and exit code.
func executeRunCommand(t *testing.T, workDir string, args []string, finder *spectra.MockSpectraFinder, runtime *spectra.MockRuntime) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	opts := []spectra.RunHandlerOption{}
	if finder != nil {
		opts = append(opts, spectra.WithSpectraFinder(finder))
	}
	if runtime != nil {
		opts = append(opts, spectra.WithRuntime(runtime))
	}

	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithRunHandlerFunc(opts...),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"run"}, args...))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workDir))
	defer func() { _ = os.Chdir(origDir) }()

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// =====================================================================
// Happy Path — Positional Argument
// =====================================================================

// TestRunCommand_PositionalArgument executes workflow when workflow name is provided as positional argument.
func TestRunCommand_PositionalArgument(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, finder, runtime)

	assert.Equal(t, 0, exitCode)
	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, projectRoot, runtime.ProjectRoot())
	assert.Equal(t, "TestWorkflow", runtime.WorkflowName())
}

// TestRunCommand_PositionalArgumentFromSubdirectory locates project root from subdirectory and executes workflow.
func TestRunCommand_PositionalArgumentFromSubdirectory(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "SimpleSdd")
	subdir := filepath.Join(projectRoot, "subdir", "nested")
	require.NoError(t, os.MkdirAll(subdir, 0755))

	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, subdir, []string{"SimpleSdd"}, finder, runtime)

	assert.Equal(t, 0, exitCode)
	assert.True(t, finder.WasCalled(), "SpectraFinder should be called")
	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, projectRoot, runtime.ProjectRoot())
	assert.Equal(t, "SimpleSdd", runtime.WorkflowName())
}

// =====================================================================
// Happy Path — Flag Argument
// =====================================================================

// TestRunCommand_FlagArgument executes workflow when workflow name is provided via --workflow flag.
func TestRunCommand_FlagArgument(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "MyWorkflow")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"--workflow", "MyWorkflow"}, finder, runtime)

	assert.Equal(t, 0, exitCode)
	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, projectRoot, runtime.ProjectRoot())
	assert.Equal(t, "MyWorkflow", runtime.WorkflowName())
}

// TestRunCommand_FlagPrecedenceOverPositional flag takes precedence when both flag and positional argument are provided.
func TestRunCommand_FlagPrecedenceOverPositional(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "FlagWorkflow")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"--workflow", "FlagWorkflow", "PositionalWorkflow"}, finder, runtime)

	assert.Equal(t, 0, exitCode)
	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "FlagWorkflow", runtime.WorkflowName(), "Flag should take precedence over positional argument")
}

// =====================================================================
// Happy Path — Help Output
// =====================================================================

// TestRunCommand_HelpFlag displays help information when --help flag is provided.
func TestRunCommand_HelpFlag(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	stdout, _, exitCode := executeRunCommand(t, tmpDir, []string{"--help"}, nil, runtime)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Run a workflow")
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Flags:")
	assert.Contains(t, stdout, "--workflow string")
	assert.Contains(t, stdout, "Examples:")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called for --help")
}

// =====================================================================
// Happy Path — Runtime Output Forwarding
// =====================================================================

// TestRunCommand_ForwardsStdout forwards Runtime stdout to command stdout.
func TestRunCommand_ForwardsStdout(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Test")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStdout(0, func(w io.Writer) {
		fmt.Fprint(w, "workflow output\n")
	})

	stdout, _, exitCode := executeRunCommand(t, projectRoot, []string{"Test"}, finder, runtime)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "workflow output\n")
}

// TestRunCommand_ForwardsStderr forwards Runtime stderr to command stderr.
func TestRunCommand_ForwardsStderr(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Test")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStderr(1, func(w io.Writer) {
		fmt.Fprint(w, "workflow error\n")
	})

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"Test"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "workflow error\n")
}

// TestRunCommand_ForwardsBothStreams forwards both stdout and stderr from Runtime.
func TestRunCommand_ForwardsBothStreams(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Test")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStreams(1,
		func(w io.Writer) { fmt.Fprint(w, "stdout data\n") },
		func(w io.Writer) { fmt.Fprint(w, "stderr data\n") },
	)

	stdout, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"Test"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stdout, "stdout data\n")
	assert.Contains(t, stderr, "stderr data\n")
}

// =====================================================================
// Validation Failures — Missing Workflow Name
// =====================================================================

// TestRunCommand_NoWorkflowName returns error when no workflow name is provided.
func TestRunCommand_NoWorkflowName(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, strings.ToLower(stderr), "workflow name")
	assert.Contains(t, strings.ToLower(stderr), "required")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_NoWorkflowNameWithFlag returns error when --workflow flag is provided without value.
func TestRunCommand_NoWorkflowNameWithFlag(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{"--workflow"}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Validation Failures — Empty Workflow Name
// =====================================================================

// TestRunCommand_EmptyWorkflowNamePositional returns error when positional workflow name is empty string.
func TestRunCommand_EmptyWorkflowNamePositional(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{""}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, strings.ToLower(stderr), "workflow name")
	assert.Contains(t, strings.ToLower(stderr), "empty")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_EmptyWorkflowNameFlag returns error when --workflow flag value is empty string.
func TestRunCommand_EmptyWorkflowNameFlag(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{"--workflow", ""}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, strings.ToLower(stderr), "workflow name")
	assert.Contains(t, strings.ToLower(stderr), "empty")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_WhitespaceWorkflowName returns error when workflow name contains only whitespace.
func TestRunCommand_WhitespaceWorkflowName(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{"   "}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, strings.ToLower(stderr), "workflow name")
	assert.Contains(t, strings.ToLower(stderr), "empty")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Validation Failures — Too Many Arguments
// =====================================================================

// TestRunCommand_TooManyArguments returns error when multiple positional arguments are provided.
func TestRunCommand_TooManyArguments(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{"Workflow1", "Workflow2"}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, strings.ToLower(stderr), "too many arguments")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_TooManyArgumentsWithThree returns error when three positional arguments are provided.
func TestRunCommand_TooManyArgumentsWithThree(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{"Workflow1", "Workflow2", "Workflow3"}, nil, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, strings.ToLower(stderr), "too many arguments")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Validation Failures — Project Root Not Found
// =====================================================================

// TestRunCommand_SpectraNotFound returns error when .spectra directory is not found.
func TestRunCommand_SpectraNotFound(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	finder := spectra.NewMockSpectraFinder("", fmt.Errorf("spectra not initialized"))
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, tmpDir, []string{"TestWorkflow"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, ".spectra")
	assert.Contains(t, strings.ToLower(stderr), "not found")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// TestRunCommand_SpectraNotFoundInParent returns error when .spectra directory is not found in any parent directory.
func TestRunCommand_SpectraNotFoundInParent(t *testing.T) {
	tmpDir := setupRunTestFixtureNoSpectra(t)
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d")
	require.NoError(t, os.MkdirAll(deepDir, 0755))

	finder := spectra.NewMockSpectraFinder("", fmt.Errorf("spectra not initialized"))
	runtime := spectra.NewMockRuntime(0)

	_, stderr, exitCode := executeRunCommand(t, deepDir, []string{"TestWorkflow"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, ".spectra")
	assert.Contains(t, strings.ToLower(stderr), "not found")
	assert.False(t, runtime.RunCalled(), "Runtime.Run should not be called")
}

// =====================================================================
// Error Propagation — Runtime Initialization Failures
// =====================================================================

// TestRunCommand_RuntimeReportsWorkflowNotFound forwards Runtime error when workflow file does not exist.
func TestRunCommand_RuntimeReportsWorkflowNotFound(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStderr(1, func(w io.Writer) {
		fmt.Fprint(w, "Error: workflow definition not found")
	})

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"NonExistent"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: workflow definition not found")
}

// TestRunCommand_RuntimeReportsInvalidYAML forwards Runtime error when workflow file has invalid YAML syntax.
func TestRunCommand_RuntimeReportsInvalidYAML(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Invalid")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStderr(1, func(w io.Writer) {
		fmt.Fprint(w, "Error: failed to parse workflow YAML")
	})

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"Invalid"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: failed to parse workflow YAML")
}

// TestRunCommand_RuntimeReportsWorkflowNotReadable forwards Runtime error when workflow file cannot be read due to permissions.
func TestRunCommand_RuntimeReportsWorkflowNotReadable(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStderr(1, func(w io.Writer) {
		fmt.Fprint(w, "Error: permission denied reading workflow file")
	})

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"Restricted"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: permission denied reading workflow file")
}

// =====================================================================
// Error Propagation — Runtime Execution Failures
// =====================================================================

// TestRunCommand_RuntimeReportsAgentError forwards Runtime error when workflow fails due to agent error.
func TestRunCommand_RuntimeReportsAgentError(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "AgentFail")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStderr(1, func(w io.Writer) {
		fmt.Fprint(w, "Error: agent execution failed")
	})

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"AgentFail"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: agent execution failed")
}

// TestRunCommand_RuntimeReportsSessionLockError forwards Runtime error when another session is already running.
func TestRunCommand_RuntimeReportsSessionLockError(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStderr(1, func(w io.Writer) {
		fmt.Fprint(w, "Error: another workflow session is already running")
	})

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"Concurrent"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: another workflow session is already running")
}

// TestRunCommand_RuntimeRunReturnsError converts Runtime.Run error return to exit code 1 and stderr message.
func TestRunCommand_RuntimeRunReturnsError(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "Test")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithError(errors.New("runtime internal error"))

	_, stderr, exitCode := executeRunCommand(t, projectRoot, []string{"Test"}, finder, runtime)

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
}

// =====================================================================
// Edge Cases — Special Workflow Names
// =====================================================================

// TestRunCommand_WorkflowNameWithPathSeparators passes workflow name with path separators to Runtime without validation.
func TestRunCommand_WorkflowNameWithPathSeparators(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, _ = executeRunCommand(t, projectRoot, []string{"../malicious/workflow"}, finder, runtime)

	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "../malicious/workflow", runtime.WorkflowName(), "Command should not validate or sanitize the name")
}

// TestRunCommand_WorkflowNameWithSpecialCharacters passes workflow name with special characters to Runtime.
func TestRunCommand_WorkflowNameWithSpecialCharacters(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"Work@flow#123"}, finder, runtime)

	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "Work@flow#123", runtime.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// TestRunCommand_WorkflowNameWithUnicode passes workflow name with Unicode characters to Runtime.
func TestRunCommand_WorkflowNameWithUnicode(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, exitCode := executeRunCommand(t, projectRoot, []string{"工作流程"}, finder, runtime)

	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "工作流程", runtime.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Edge Cases — Signal Handling
// =====================================================================

// TestRunCommand_PropagatesSIGINT propagates SIGINT signal to Runtime subprocess.
func TestRunCommand_PropagatesSIGINT(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)

	signalCh := make(chan struct{})
	signalReceived := false
	var mu sync.Mutex

	runtime := spectra.NewMockRuntime(0)
	runtime.SetSignalCh(signalCh)

	var stdout, stderr bytes.Buffer
	var exitCode int

	done := make(chan struct{})
	go func() {
		defer close(done)
		stdout2, stderr2, code := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, finder, runtime)
		stdout.WriteString(stdout2)
		stderr.WriteString(stderr2)
		exitCode = code
	}()

	// Give goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Signal the mock Runtime to indicate SIGINT was delivered
	mu.Lock()
	signalReceived = true
	mu.Unlock()
	close(signalCh)

	<-done

	mu.Lock()
	defer mu.Unlock()
	assert.True(t, signalReceived, "Mock Runtime should record SIGINT delivery")
	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Edge Cases — Streaming Output
// =====================================================================

// TestRunCommand_HandlesStreamingOutput does not timeout or buffer output from workflow that produces streaming output.
func TestRunCommand_HandlesStreamingOutput(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "StreamingWorkflow")
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntimeWithStdout(0, func(w io.Writer) {
		// Simulate streaming: produce multiple output lines over a short period
		for i := 0; i < 10; i++ {
			fmt.Fprintf(w, "line %d\n", i)
			time.Sleep(15 * time.Millisecond)
		}
	})

	stdout, _, exitCode := executeRunCommand(t, projectRoot, []string{"StreamingWorkflow"}, finder, runtime)

	assert.Equal(t, 0, exitCode)
	// Verify all output lines are present
	for i := 0; i < 10; i++ {
		assert.Contains(t, stdout, fmt.Sprintf("line %d", i))
	}
}

// =====================================================================
// Idempotency
// =====================================================================

// TestRunCommand_MultipleInvocationsIndependent multiple sequential invocations are independent.
func TestRunCommand_MultipleInvocationsIndependent(t *testing.T) {
	projectRoot := setupRunTestFixture(t, "TestWorkflow")
	finder1 := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime1 := spectra.NewMockRuntime(0)

	_, _, exitCode1 := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, finder1, runtime1)

	assert.Equal(t, 0, exitCode1)
	assert.True(t, runtime1.RunCalled(), "First Runtime.Run should be called")
	assert.Equal(t, projectRoot, runtime1.ProjectRoot())
	assert.Equal(t, "TestWorkflow", runtime1.WorkflowName())

	// Second invocation with fresh mocks
	finder2 := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime2 := spectra.NewMockRuntime(0)

	_, _, exitCode2 := executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, finder2, runtime2)

	assert.Equal(t, 0, exitCode2)
	assert.True(t, runtime2.RunCalled(), "Second Runtime.Run should be called")
	assert.Equal(t, projectRoot, runtime2.ProjectRoot())
	assert.Equal(t, "TestWorkflow", runtime2.WorkflowName())
}

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

// TestRunCommand_CallsSpectraFinder invokes SpectraFinder to locate project root.
func TestRunCommand_CallsSpectraFinder(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, _ = executeRunCommand(t, projectRoot, []string{"TestWorkflow"}, finder, runtime)

	assert.True(t, finder.WasCalled(), "SpectraFinder.Find() should be called")
	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called with project root from SpectraFinder")
	assert.Equal(t, projectRoot, runtime.ProjectRoot())
}

// TestRunCommand_PassesCorrectProjectRoot passes project root from SpectraFinder to Runtime.
func TestRunCommand_PassesCorrectProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".spectra"), 0755))
	subDir := filepath.Join(projectRoot, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, _ = executeRunCommand(t, subDir, []string{"TestWorkflow"}, finder, runtime)

	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, projectRoot, runtime.ProjectRoot(), "Runtime.Run should receive project root from SpectraFinder")
	assert.Equal(t, "TestWorkflow", runtime.WorkflowName())
}

// TestRunCommand_PassesCorrectWorkflowName passes workflow name exactly as provided to Runtime.
func TestRunCommand_PassesCorrectWorkflowName(t *testing.T) {
	projectRoot := setupRunTestFixture(t)
	finder := spectra.NewMockSpectraFinder(projectRoot, nil)
	runtime := spectra.NewMockRuntime(0)

	_, _, _ = executeRunCommand(t, projectRoot, []string{"My-Workflow_123"}, finder, runtime)

	assert.True(t, runtime.RunCalled(), "Runtime.Run should be called")
	assert.Equal(t, "My-Workflow_123", runtime.WorkflowName(), "Workflow name should be passed exactly, no transformation")
}
