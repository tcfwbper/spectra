package spectra_test

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// --- Test Helpers ---

// executeRunCommand creates and executes the run command with the given mock runtime and args,
// returning stdout, stderr, and exit code.
func executeRunCommand(t *testing.T, mockRuntime *spectra.MockRunRuntime, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithRunRuntime(mockRuntime),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"run"}, args...))

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// executeRunCommandNoRuntime creates and executes the run command without a mock runtime,
// returning stdout, stderr, and exit code. Used for validation-only tests.
func executeRunCommandNoRuntime(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := spectra.NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"run"}, args...))

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// =====================================================================
// Happy Path — Successful Execution
// =====================================================================

// TestRunCommand_SuccessfulWorkflow exits with code 0 when Runtime.Run returns nil.
func TestRunCommand_SuccessfulWorkflow(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	stdout, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 0, exitCode)
	assert.Empty(t, stderr)
	assert.Empty(t, stdout)
}

// TestRunCommand_RuntimeInvokedWithWorkflowName passes workflow name to Runtime.Run exactly as provided.
func TestRunCommand_RuntimeInvokedWithWorkflowName(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "TestWorkflow123"})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, "TestWorkflow123", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// TestRunCommand_SingleRuntimeInvocation invokes Runtime.Run exactly once per command execution.
func TestRunCommand_SingleRuntimeInvocation(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, mockRT.RunCallCount())
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Happy Path — Help and Usage
// =====================================================================

// TestRunCommand_HelpFlag displays usage information when invoked with --help.
func TestRunCommand_HelpFlag(t *testing.T) {
	stdout, _, exitCode := executeRunCommandNoRuntime(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "--workflow")
	assert.Contains(t, stdout, "Usage:")
}

// TestRunCommand_UsageIncludesWorkflowFlag usage information documents the required --workflow flag.
func TestRunCommand_UsageIncludesWorkflowFlag(t *testing.T) {
	stdout, _, exitCode := executeRunCommandNoRuntime(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "--workflow")
	// The flag description should indicate it specifies workflow name
	assert.Regexp(t, `(?i)workflow`, stdout)
}

// =====================================================================
// Validation Failures — Missing or Empty Workflow Flag
// =====================================================================

// TestRunCommand_MissingWorkflowFlag exits with code 1 when --workflow flag not provided.
func TestRunCommand_MissingWorkflowFlag(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: required flag --workflow not provided")
}

// TestRunCommand_EmptyWorkflowString exits with code 1 when --workflow value is empty string.
func TestRunCommand_EmptyWorkflowString(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{"--workflow", ""})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: --workflow flag cannot be empty")
}

// TestRunCommand_WorkflowFlagWithoutValue treats missing flag value same as missing flag.
func TestRunCommand_WorkflowFlagWithoutValue(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{"--workflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: required flag --workflow not provided")
}

// =====================================================================
// Validation Failures — Positional Arguments
// =====================================================================

// TestRunCommand_PositionalArgument rejects positional argument and suggests using --workflow flag.
func TestRunCommand_PositionalArgument(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{"MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: unexpected argument 'MyWorkflow'. Use --workflow flag to specify workflow name.")
}

// TestRunCommand_MultiplePositionalArguments reports error for first positional argument.
func TestRunCommand_MultiplePositionalArguments(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{"arg1", "arg2"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: unexpected argument 'arg1'. Use --workflow flag to specify workflow name.")
}

// TestRunCommand_PositionalWithWorkflowFlag rejects positional argument even when --workflow flag is valid.
func TestRunCommand_PositionalWithWorkflowFlag(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{"--workflow", "Valid", "extra-arg"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: unexpected argument 'extra-arg'. Use --workflow flag to specify workflow name.")
}

// =====================================================================
// Happy Path — No Workflow Validation
// =====================================================================

// TestRunCommand_NoWorkflowNameValidation does not validate workflow name format before invoking Runtime.
func TestRunCommand_NoWorkflowNameValidation(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", "../../../etc/passwd"})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, "../../../etc/passwd", mockRT.WorkflowName())
}

// TestRunCommand_SpecialCharactersInWorkflowName accepts workflow names with special characters.
func TestRunCommand_SpecialCharactersInWorkflowName(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", "workflow-with-special!@#$%"})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, "workflow-with-special!@#$%", mockRT.WorkflowName())
}

// =====================================================================
// Error Handling — Generic Runtime Errors
// =====================================================================

// TestRunCommand_RuntimeInitializationError exits with code 1 when Runtime returns initialization error.
func TestRunCommand_RuntimeInitializationError(t *testing.T) {
	errMsg := "failed to initialize session: failed to load workflow definition: workflow file not found: MyWorkflow"
	mockRT := spectra.NewMockRunRuntime(errors.New(errMsg))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: "+errMsg)
}

// TestRunCommand_RuntimeSessionFailureError exits with code 1 when Runtime returns session failure error.
func TestRunCommand_RuntimeSessionFailureError(t *testing.T) {
	errMsg := "session failed: agent execution error: ArchitectAgent failed to generate specifications"
	mockRT := spectra.NewMockRunRuntime(errors.New(errMsg))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: "+errMsg)
}

// TestRunCommand_RuntimeProjectNotInitializedError exits with code 1 when Runtime returns project not initialized error.
func TestRunCommand_RuntimeProjectNotInitializedError(t *testing.T) {
	errMsg := "failed to locate project root: spectra not initialized"
	mockRT := spectra.NewMockRunRuntime(errors.New(errMsg))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: "+errMsg)
}

// TestRunCommand_RuntimeUnexpectedError exits with code 1 for any non-signal Runtime error.
func TestRunCommand_RuntimeUnexpectedError(t *testing.T) {
	errMsg := "unexpected error from runtime"
	mockRT := spectra.NewMockRunRuntime(errors.New(errMsg))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: "+errMsg)
}

// =====================================================================
// Error Handling — SIGINT Termination
// =====================================================================

// TestRunCommand_SIGINTExactMatch exits with code 130 when Runtime error contains exact SIGINT message.
func TestRunCommand_SIGINTExactMatch(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("session terminated by signal SIGINT"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, exitCode)
	assert.Contains(t, stderr, "Error: session terminated by signal SIGINT")
}

// TestRunCommand_SIGINTSubstringMatch exits with code 130 when error contains SIGINT substring.
func TestRunCommand_SIGINTSubstringMatch(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("cleanup failed after session terminated by signal SIGINT"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, exitCode)
	assert.NotEmpty(t, stderr)
}

// TestRunCommand_SIGINTWithContextError exits with code 130 when error wraps SIGINT message.
func TestRunCommand_SIGINTWithContextError(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(fmt.Errorf("workflow execution interrupted: session terminated by signal SIGINT: context"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, exitCode)
	assert.NotEmpty(t, stderr)
}

// =====================================================================
// Error Handling — SIGTERM Termination
// =====================================================================

// TestRunCommand_SIGTERMExactMatch exits with code 143 when Runtime error contains exact SIGTERM message.
func TestRunCommand_SIGTERMExactMatch(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("session terminated by signal SIGTERM"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 143, exitCode)
	assert.Contains(t, stderr, "Error: session terminated by signal SIGTERM")
}

// TestRunCommand_SIGTERMSubstringMatch exits with code 143 when error contains SIGTERM substring.
func TestRunCommand_SIGTERMSubstringMatch(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("cleanup failed after session terminated by signal SIGTERM"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 143, exitCode)
	assert.NotEmpty(t, stderr)
}

// TestRunCommand_SIGTERMWithContextError exits with code 143 when error wraps SIGTERM message.
func TestRunCommand_SIGTERMWithContextError(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(fmt.Errorf("workflow execution interrupted: session terminated by signal SIGTERM: context"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 143, exitCode)
	assert.NotEmpty(t, stderr)
}

// =====================================================================
// Error Handling — Exit Code Priority
// =====================================================================

// TestRunCommand_SIGINTPriorityOverGenericError SIGINT exit code takes precedence when substring detected.
func TestRunCommand_SIGINTPriorityOverGenericError(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("multiple errors: session terminated by signal SIGINT"))

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, exitCode, "SIGINT exit code 130 should take precedence over generic error exit code 1")
}

// TestRunCommand_SIGTERMPriorityOverGenericError SIGTERM exit code takes precedence when substring detected.
func TestRunCommand_SIGTERMPriorityOverGenericError(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("multiple errors: session terminated by signal SIGTERM"))

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 143, exitCode, "SIGTERM exit code 143 should take precedence over generic error exit code 1")
}

// TestRunCommand_SIGINTBeforeSIGTERM SIGINT detected first when both substrings present.
func TestRunCommand_SIGINTBeforeSIGTERM(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("session terminated by signal SIGINT then session terminated by signal SIGTERM"))

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, exitCode, "SIGINT (130) should be checked first when both SIGINT and SIGTERM are present")
}

// =====================================================================
// Error Output Format
// =====================================================================

// TestRunCommand_ErrorPrefixConsistency all error messages prefixed with "Error: ".
func TestRunCommand_ErrorPrefixConsistency(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("test error"))

	_, stderr, _ := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.True(t, strings.HasPrefix(stderr, "Error: test error"), "stderr should start with 'Error: test error', got: %q", stderr)
}

// TestRunCommand_ErrorToStderr error messages printed to stderr, not stdout.
func TestRunCommand_ErrorToStderr(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("test error"))

	stdout, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
	assert.Empty(t, stdout)
}

// TestRunCommand_NoAdditionalContextAdded does not add additional context to Runtime error message.
func TestRunCommand_NoAdditionalContextAdded(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("original error message"))

	_, stderr, _ := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	// stderr should contain exactly "Error: original error message" without additional wrapping
	assert.Contains(t, stderr, "Error: original error message")
	// Ensure no additional wrapping like "run failed: Error: ..." or similar
	trimmed := strings.TrimSpace(stderr)
	assert.Equal(t, "Error: original error message", trimmed)
}

// =====================================================================
// Output and Logging — Success Case
// =====================================================================

// TestRunCommand_NoOutputOnSuccess does not print to stdout or stderr on successful execution.
func TestRunCommand_NoOutputOnSuccess(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	stdout, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 0, exitCode)
	assert.Empty(t, stderr)
	assert.Empty(t, stdout)
}

// TestRunCommand_SessionFinalizerHandlesSuccessMessage run command does not duplicate SessionFinalizer success output.
func TestRunCommand_SessionFinalizerHandlesSuccessMessage(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	var stdout bytes.Buffer

	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithRunRuntime(mockRT),
	)
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"run", "--workflow", "MyWorkflow"})

	// Set the stdout on the mock so it can simulate SessionFinalizer writing
	mockRT.SetStdout(&stdout)
	mockRT.SetFinalizerOutput("Session completed successfully\n")

	exitCode := cmd.Execute()

	assert.Equal(t, 0, exitCode)
	// Only SessionFinalizer output should appear; run command adds nothing
	assert.Equal(t, "Session completed successfully\n", stdout.String())
}

// =====================================================================
// Blocking Behavior
// =====================================================================

// TestRunCommand_BlocksOnRuntimeRun does not return until Runtime.Run completes.
func TestRunCommand_BlocksOnRuntimeRun(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)
	mockRT.SetBlockDuration(100 * time.Millisecond)

	start := time.Now()
	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})
	elapsed := time.Since(start)

	assert.Equal(t, 0, exitCode)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "Command should block for at least 100ms")
}

// TestRunCommand_ReturnsImmediatelyOnRuntimeError returns as soon as Runtime.Run returns error.
func TestRunCommand_ReturnsImmediatelyOnRuntimeError(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("immediate error"))

	start := time.Now()
	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})
	elapsed := time.Since(start)

	assert.Equal(t, 1, exitCode)
	assert.Less(t, elapsed, 50*time.Millisecond, "Command should return immediately on error")
}

// =====================================================================
// No Retry Logic
// =====================================================================

// TestRunCommand_NoRetryOnFailure does not retry Runtime.Run on failure.
func TestRunCommand_NoRetryOnFailure(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(errors.New("runtime failure"))

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Equal(t, 1, mockRT.RunCallCount(), "Runtime.Run should be called exactly once, no retries")
}

// =====================================================================
// Edge Cases — Panic Recovery
// =====================================================================

// TestRunCommand_RuntimePanicWithRecovery recovers from Runtime panic and exits with code 1 if panic recovery implemented.
func TestRunCommand_RuntimePanicWithRecovery(t *testing.T) {
	mockRT := spectra.NewMockRunRuntimeWithPanic("nil pointer dereference")

	defer func() {
		if r := recover(); r != nil {
			// Panic propagated — this is the "without recovery" behavior.
			// Test passes: panic was not recovered by run command.
			return
		}
	}()

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	// If we reach here, panic recovery is implemented.
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: runtime panic: nil pointer dereference")
}

// TestRunCommand_RuntimePanicWithoutRecovery allows panic to propagate if panic recovery not implemented.
func TestRunCommand_RuntimePanicWithoutRecovery(t *testing.T) {
	mockRT := spectra.NewMockRunRuntimeWithPanic("nil pointer dereference")

	recovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})
	}()

	// If panic was recovered by our deferred func, then run command did not recover it
	// (i.e., "without recovery" behavior — panic propagated).
	// If panic was NOT recovered here, then run command caught it (handled in TestRunCommand_RuntimePanicWithRecovery).
	if recovered {
		// Panic propagated — this is the expected "without recovery" behavior.
		// Go default exit code for unrecovered panics is 2.
		t.Log("Panic propagated as expected when panic recovery is not implemented")
	}
}

// =====================================================================
// Edge Cases — Parallel Execution
// =====================================================================

// TestRunCommand_ParallelInvocationsIndependent multiple parallel invocations execute independently.
func TestRunCommand_ParallelInvocationsIndependent(t *testing.T) {
	mockRT1 := spectra.NewMockRunRuntime(nil)
	mockRT2 := spectra.NewMockRunRuntime(nil)

	var wg sync.WaitGroup
	wg.Add(2)

	var exitCode1, exitCode2 int

	go func() {
		defer wg.Done()
		_, _, exitCode1 = executeRunCommand(t, mockRT1, []string{"--workflow", "WorkflowA"})
	}()

	go func() {
		defer wg.Done()
		_, _, exitCode2 = executeRunCommand(t, mockRT2, []string{"--workflow", "WorkflowB"})
	}()

	wg.Wait()

	assert.Equal(t, 0, exitCode1)
	assert.Equal(t, 0, exitCode2)
	assert.Equal(t, "WorkflowA", mockRT1.WorkflowName())
	assert.Equal(t, "WorkflowB", mockRT2.WorkflowName())
}

// =====================================================================
// Edge Cases — Unknown Flags
// =====================================================================

// TestRunCommand_UnknownFlag returns error for unknown flag.
func TestRunCommand_UnknownFlag(t *testing.T) {
	_, stderr, exitCode := executeRunCommandNoRuntime(t, []string{"--unknown-flag", "value"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "unknown flag")
}

// TestRunCommand_ValidAndInvalidFlagsCombined returns error when unknown flag provided alongside valid flag.
func TestRunCommand_ValidAndInvalidFlagsCombined(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow", "--unknown-flag"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "unknown flag")
	assert.False(t, mockRT.RunCalled(), "Runtime should not be invoked when flag parsing fails")
}

// =====================================================================
// Platform Compatibility
// =====================================================================

// TestRunCommand_ExitCodeRangeCompatibility exit codes 0, 1, 130, 143 are valid on all platforms (0-255 range).
func TestRunCommand_ExitCodeRangeCompatibility(t *testing.T) {
	exitCodes := map[string]struct {
		err      error
		expected int
	}{
		"success": {nil, 0},
		"generic": {errors.New("error"), 1},
		"sigint":  {errors.New("session terminated by signal SIGINT"), 130},
		"sigterm": {errors.New("session terminated by signal SIGTERM"), 143},
	}

	for name, tc := range exitCodes {
		t.Run(name, func(t *testing.T) {
			mockRT := spectra.NewMockRunRuntime(tc.err)

			_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

			assert.Equal(t, tc.expected, exitCode)
			assert.GreaterOrEqual(t, exitCode, 0, "Exit code should be >= 0")
			assert.LessOrEqual(t, exitCode, 255, "Exit code should be <= 255")
		})
	}
}

// TestRunCommand_SIGTERMNotAvailableOnWindows SIGTERM error never returned by Runtime on Windows.
func TestRunCommand_SIGTERMNotAvailableOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Test documents Windows limitation: SIGTERM not supported on Windows")
	}

	// Simulate the behavior if Runtime incorrectly returned SIGTERM error on Windows.
	mockRT := spectra.NewMockRunRuntime(errors.New("session terminated by signal SIGTERM"))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	// Even on Windows, the exit code mapping logic still works identically.
	assert.Equal(t, 143, exitCode)
	assert.NotEmpty(t, stderr)
}

// =====================================================================
// Cobra Framework Integration
// =====================================================================

// TestRunCommand_UsesCobraFramework run subcommand implemented using Cobra library.
func TestRunCommand_UsesCobraFramework(t *testing.T) {
	cmd := spectra.NewRootCommand()

	// Verify that the run subcommand is registered and uses Cobra patterns
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"run", "--help"})

	exitCode := cmd.Execute()

	assert.Equal(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "--workflow")
}

// TestRunCommand_RegisteredAsSubcommand run command registered with root command.
func TestRunCommand_RegisteredAsSubcommand(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithRunRuntime(mockRT),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "--workflow", "MyWorkflow"})

	exitCode := cmd.Execute()

	assert.Equal(t, 0, exitCode)
	assert.True(t, mockRT.RunCalled(), "Runtime.Run should be invoked via root command delegation")
}

// =====================================================================
// Stateless Execution
// =====================================================================

// TestRunCommand_StatelessInvocations multiple sequential invocations are independent with no shared state.
func TestRunCommand_StatelessInvocations(t *testing.T) {
	// First invocation: returns error
	mockRT1 := spectra.NewMockRunRuntime(errors.New("first invocation error"))
	_, _, exitCode1 := executeRunCommand(t, mockRT1, []string{"--workflow", "W1"})

	// Second invocation: returns nil (success)
	mockRT2 := spectra.NewMockRunRuntime(nil)
	_, _, exitCode2 := executeRunCommand(t, mockRT2, []string{"--workflow", "W2"})

	assert.Equal(t, 1, exitCode1, "First invocation should fail")
	assert.Equal(t, 0, exitCode2, "Second invocation should succeed independently")
	assert.Equal(t, "W1", mockRT1.WorkflowName())
	assert.Equal(t, "W2", mockRT2.WorkflowName())
}

// =====================================================================
// Boundary Values — Workflow Name Length
// =====================================================================

// TestRunCommand_SingleCharacterWorkflowName accepts single-character workflow name.
func TestRunCommand_SingleCharacterWorkflowName(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "a"})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, "a", mockRT.WorkflowName())
	assert.Equal(t, 0, exitCode)
}

// TestRunCommand_VeryLongWorkflowName accepts very long workflow name without length restriction.
func TestRunCommand_VeryLongWorkflowName(t *testing.T) {
	longName := strings.Repeat("x", 10000)
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", longName})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, longName, mockRT.WorkflowName(), "10000-character string should not be truncated")
	assert.Equal(t, 0, exitCode)
}

// =====================================================================
// Boundary Values — Whitespace in Workflow Name
// =====================================================================

// TestRunCommand_WorkflowNameWithLeadingWhitespace passes workflow name with leading whitespace as-is.
func TestRunCommand_WorkflowNameWithLeadingWhitespace(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", " LeadingSpace"})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, " LeadingSpace", mockRT.WorkflowName(), "Leading space should be preserved")
}

// TestRunCommand_WorkflowNameWithTrailingWhitespace passes workflow name with trailing whitespace as-is.
func TestRunCommand_WorkflowNameWithTrailingWhitespace(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", "TrailingSpace "})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, "TrailingSpace ", mockRT.WorkflowName(), "Trailing space should be preserved")
}

// TestRunCommand_WorkflowNameAllWhitespace passes workflow name that is only whitespace (not empty string).
func TestRunCommand_WorkflowNameAllWhitespace(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", "   "})

	assert.True(t, mockRT.RunCalled(), "Whitespace-only string is not empty, validation should pass")
	assert.Equal(t, "   ", mockRT.WorkflowName())
}

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

// TestRunCommand_RuntimeRunCalledWithCorrectParameter Runtime.Run receives workflow name exactly as specified.
func TestRunCommand_RuntimeRunCalledWithCorrectParameter(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", "MyTestWorkflow"})

	assert.True(t, mockRT.RunCalled())
	assert.Equal(t, "MyTestWorkflow", mockRT.WorkflowName())
}

// TestRunCommand_RuntimeNotInvokedOnValidationFailure Runtime.Run not called when command-line validation fails.
func TestRunCommand_RuntimeNotInvokedOnValidationFailure(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", ""})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
	assert.False(t, mockRT.RunCalled(), "Runtime.Run should not be called when validation fails")
}

// =====================================================================
// Error Propagation
// =====================================================================

// TestRunCommand_PropagatesRuntimeErrorMessage prints Runtime error message to stderr without modification.
func TestRunCommand_PropagatesRuntimeErrorMessage(t *testing.T) {
	errMsg := "detailed runtime error with session ID abc-123 and context"
	mockRT := spectra.NewMockRunRuntime(errors.New(errMsg))

	_, stderr, exitCode := executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: "+errMsg)
}

// =====================================================================
// Integration — Runtime Context
// =====================================================================

// TestRunCommand_RuntimeReceivesNoAdditionalContext run command does not pass additional context beyond workflow name.
func TestRunCommand_RuntimeReceivesNoAdditionalContext(t *testing.T) {
	mockRT := spectra.NewMockRunRuntime(nil)

	_, _, _ = executeRunCommand(t, mockRT, []string{"--workflow", "MyWorkflow"})

	// Runtime.Run is called with single parameter: workflow name string.
	// The mock captures only the workflow name, confirming no additional parameters.
	require.True(t, mockRT.RunCalled())
	assert.Equal(t, "MyWorkflow", mockRT.WorkflowName())
}
