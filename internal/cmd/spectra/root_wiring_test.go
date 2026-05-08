package spectra

import (
	"bytes"
	"testing"
)

// rootResult holds the output of an executeRoot invocation.
type rootResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// executeRoot is a test wiring helper that constructs the root command,
// sets the given args, captures stdout/stderr, and returns the result.
func executeRoot(t *testing.T, args []string) rootResult {
	t.Helper()
	stdout, stderr := newCapturedBuffers()
	code := ExecuteWithOptions(RootCommandOptions{
		Stdout: stdout,
		Stderr: stderr,
		Args:   args,
	})
	return rootResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: code,
	}
}

// executeRootWithStubSubcommand is a test wiring helper that registers a stub
// subcommand with a configurable exit code, then executes the root command.
func executeRootWithStubSubcommand(t *testing.T, stubName string, stubExitCode int, args []string) rootResult {
	t.Helper()
	stdout, stderr := newCapturedBuffers()

	cmd := newRootCommandForTest()
	cmd.AddCommand(newStubSubcommand(stubName, stubExitCode))

	code := executeCommand(cmd, RootCommandOptions{
		Stdout: stdout,
		Stderr: stderr,
		Args:   args,
	})
	return rootResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: code,
	}
}

// newCapturedBuffers returns a pair of bytes.Buffer for stdout and stderr capture.
func newCapturedBuffers() (*bytes.Buffer, *bytes.Buffer) {
	return &bytes.Buffer{}, &bytes.Buffer{}
}
