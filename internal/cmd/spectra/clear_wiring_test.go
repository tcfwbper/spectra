package spectra

import (
	"bytes"
	"io"
)

// runClear is a test helper that wires test fakes into RunClearCommand.
// It executes the clear command with the provided dependencies and returns the exit code.
func runClear(finder ClearSpectraFinder, layout ClearStorageLayout, stdin io.Reader, stdout, stderr *bytes.Buffer, args []string) int {
	return RunClearCommand(ClearCommandOptions{
		Finder: finder,
		Layout: layout,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   args,
	})
}
