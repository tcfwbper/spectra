package spectra

import (
	"testing"
)

// runResult holds the output of a runRun invocation.
type runResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runRun is a test wiring helper that constructs the run command with a fake
// Runtime, sets the given args, captures stdout/stderr, and returns the result.
//
// Scaffolded: requires run.go to define:
//   - RunRuntime interface with Run(workflowName string, log logger.Logger) (int, error)
//   - RunCommandOptions struct with injectable dependencies
//   - RunRunCommand(opts RunCommandOptions) int
func runRun(t *testing.T, rt *fakeRuntime, args []string) runResult {
	t.Helper()
	t.Skip("scaffolded: requires run.go to define RunRuntime interface and RunRunCommand(RunCommandOptions) int")
	return runResult{}
}
