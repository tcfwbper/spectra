package spectra

import (
	"bytes"
)

// runInit is a test helper that wires test fakes into RunInitCommand.
// It executes the init command with the provided dependencies and returns the exit code.
func runInit(deps *fakeInitDeps, stdout, stderr *bytes.Buffer) int {
	getwdFunc := func() (string, error) {
		if deps.getwdErr != nil {
			return "", deps.getwdErr
		}
		return deps.cwd, nil
	}

	return RunInitCommand(InitCommandOptions{
		GetwdFunc:        getwdFunc,
		GitignoreEnsurer: deps.gitignoreEnsurer,
		DirectoryCreator: deps.directoryCreator,
		Copier:           deps.copier,
		Stdout:           stdout,
		Stderr:           stderr,
	})
}
