package spectra

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — Run ---

func TestRun_Success_ExitsZero(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	result := runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 0, result.exitCode)
	assert.Empty(t, result.stderr)
}

func TestRun_PassesWorkflowNameToRuntime(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	_ = runRun(t, rt, []string{"--workflow", "deploy-prod"})

	assert.Equal(t, "deploy-prod", rt.workflowName)
}

func TestRun_PassesLoggerToRuntime(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	_ = runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.False(t, rt.loggerWasNil, "expected a non-nil logger to be passed to Runtime.Run")
}

// --- Validation Failures ---

func TestRun_MissingWorkflowFlag_ExitsOne(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	result := runRun(t, rt, nil)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "Error: required flag --workflow not provided")
}

func TestRun_EmptyWorkflowFlag_ExitsOne(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	result := runRun(t, rt, []string{"--workflow", ""})

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "Error: --workflow flag cannot be empty")
}

func TestRun_PositionalArgs_ExitsOne(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	result := runRun(t, rt, []string{"MyWorkflow"})

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "Error: unexpected argument 'MyWorkflow'. Use --workflow flag to specify workflow name.")
}

// --- Happy Path — Exit Code Mapping ---

func TestRun_SignalInterrupt_ExitsCode130(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(1, errors.New("session terminated by signal interrupt"))
	result := runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, result.exitCode)
	assert.Contains(t, result.stderr, "Error: session terminated by signal interrupt")
}

func TestRun_SignalTerminated_ExitsCode143(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(1, errors.New("session terminated by signal terminated"))
	result := runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 143, result.exitCode)
	assert.Contains(t, result.stderr, "Error: session terminated by signal terminated")
}

func TestRun_RuntimeError_ExitsWithRuntimeCode(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(1, errors.New("failed to initialize session: workflow file not found: MyWorkflow"))
	result := runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "Error: failed to initialize session: workflow file not found: MyWorkflow")
}

func TestRun_SignalSubstring_ExitsCode130(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(1, errors.New("runtime failure: session terminated by signal interrupt during cleanup"))
	result := runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 130, result.exitCode)
}

// --- Error Propagation ---

func TestRun_RuntimeCleanupTimeout_ExitsOne(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(1, errors.New("cleanup timeout"))
	result := runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "Error: cleanup timeout")
}

// --- Mock / Dependency Interaction ---

func TestRun_InvokesRuntimeExactlyOnce(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	_ = runRun(t, rt, []string{"--workflow", "MyWorkflow"})

	assert.Equal(t, 1, rt.calledCount)
}

func TestRun_DoesNotInvokeRuntimeOnValidationFailure(t *testing.T) {
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")

	rt := newFakeRuntime(0, nil)
	_ = runRun(t, rt, nil)

	assert.Equal(t, 0, rt.calledCount)
}
