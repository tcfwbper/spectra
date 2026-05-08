package spectra

import (
	"fmt"
	"io"
)

// InitGitignoreEnsurer defines the interface for gitignore handling in the init command.
type InitGitignoreEnsurer interface {
	Ensure(projectRoot string) error
}

// InitDirectoryCreator defines the interface for directory creation in the init command.
type InitDirectoryCreator interface {
	CreateAll(projectRoot string) error
}

// InitBuiltinResourceCopier defines the interface for built-in resource copying in the init command.
type InitBuiltinResourceCopier interface {
	CopyWorkflows(projectRoot string) ([]string, error)
	CopyAgents(projectRoot string) ([]string, error)
	CopySpecFiles(projectRoot string) ([]string, error)
}

// InitCommandOptions holds injectable dependencies for the init command.
type InitCommandOptions struct {
	GetwdFunc        func() (string, error)
	GitignoreEnsurer InitGitignoreEnsurer
	DirectoryCreator InitDirectoryCreator
	Copier           InitBuiltinResourceCopier
	Stdout           io.Writer
	Stderr           io.Writer
}

// RunInitCommand executes the init command logic with the given options.
// Returns the exit code.
func RunInitCommand(opts InitCommandOptions) int {
	// Step 1: Determine project root as CWD.
	projectRoot, err := opts.GetwdFunc()
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: failed to determine working directory: %s\n", err)
		return 1
	}

	// Phase 0: Ensure .gitignore.
	if err := opts.GitignoreEnsurer.Ensure(projectRoot); err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", err)
		return 1
	}

	// Phase 1: Create directories.
	if err := opts.DirectoryCreator.CreateAll(projectRoot); err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", err)
		return 1
	}

	// Phase 2a: Copy workflows.
	warnings, err := opts.Copier.CopyWorkflows(projectRoot)
	for _, w := range warnings {
		_, _ = fmt.Fprintln(opts.Stdout, w)
	}
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", err)
		return 1
	}

	// Phase 2b: Copy agents.
	warnings, err = opts.Copier.CopyAgents(projectRoot)
	for _, w := range warnings {
		_, _ = fmt.Fprintln(opts.Stdout, w)
	}
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", err)
		return 1
	}

	// Phase 2c: Copy spec files.
	warnings, err = opts.Copier.CopySpecFiles(projectRoot)
	for _, w := range warnings {
		_, _ = fmt.Fprintln(opts.Stdout, w)
	}
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", err)
		return 1
	}

	// All phases succeeded.
	_, _ = fmt.Fprintln(opts.Stdout, "Spectra project initialized successfully")
	return 0
}
