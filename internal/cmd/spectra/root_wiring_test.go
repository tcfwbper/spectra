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
//
// Scaffolded: requires root.go to define:
//   - NewRootCommand() *cobra.Command (or equivalent constructor)
//   - An execution path that returns (exitCode int)
func executeRoot(t *testing.T, args []string) rootResult {
	t.Helper()
	t.Skip("scaffolded: requires root.go to define NewRootCommand() and Execute() int")
	return rootResult{}
}

// executeRootWithStubSubcommand is a test wiring helper that registers a stub
// subcommand with a configurable exit code, then executes the root command.
//
// Scaffolded: requires root.go to expose subcommand registration or provide
// a constructor that accepts additional subcommands for testing.
func executeRootWithStubSubcommand(t *testing.T, stubName string, stubExitCode int, args []string) rootResult {
	t.Helper()
	t.Skip("scaffolded: requires root.go to define NewRootCommand() with subcommand registration seam")
	return rootResult{}
}

// newCapturedBuffers returns a pair of bytes.Buffer for stdout and stderr capture.
func newCapturedBuffers() (*bytes.Buffer, *bytes.Buffer) {
	return &bytes.Buffer{}, &bytes.Buffer{}
}
