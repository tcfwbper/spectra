package spectra

import (
	"bytes"
	"testing"

	"github.com/tcfwbper/spectra/logger"
)

// runResult holds the output of a runRun invocation.
type runResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runRun is a test wiring helper that constructs the run command with a fake
// Runtime, sets the given args, captures stdout/stderr, and returns the result.
func runRun(t *testing.T, rt *fakeRuntime, args []string) runResult {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// Parse args to extract --workflow flag and positional args
	workflow, workflowProvided, positionalArgs := parseRunArgs(args)

	code := RunRunCommand(RunCommandOptions{
		Runtime:          rt,
		Workflow:         workflow,
		WorkflowProvided: workflowProvided,
		Args:             positionalArgs,
		Stdout:           stdout,
		Stderr:           stderr,
		Logger:           logger.NewNopLogger(),
	})
	return runResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: code,
	}
}

// parseRunArgs parses command-line args to extract --workflow value and positional args.
// Returns (workflow, workflowProvided, positionalArgs).
func parseRunArgs(args []string) (string, bool, []string) {
	var workflow string
	var workflowProvided bool
	var positionalArgs []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--workflow" {
			workflowProvided = true
			if i+1 < len(args) {
				workflow = args[i+1]
				i++ // skip next arg (the value)
			}
		} else {
			positionalArgs = append(positionalArgs, args[i])
		}
	}

	return workflow, workflowProvided, positionalArgs
}
