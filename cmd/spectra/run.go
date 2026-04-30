package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/storage"
)

// SpectraFinderInterface defines the interface for locating the project root.
type SpectraFinderInterface interface {
	Find() (string, error)
}

// RuntimeInterface defines the interface for executing workflows.
type RuntimeInterface interface {
	Run(projectRoot, workflowName string, stdout, stderr io.Writer) (int, error)
}

// defaultSpectraFinder is a wrapper that implements SpectraFinderInterface.
type defaultSpectraFinder struct{}

func (d *defaultSpectraFinder) Find() (string, error) {
	return storage.SpectraFinder("")
}

// RunHandlerOption is a function that configures the run handler.
type RunHandlerOption func(*runHandlerConfig)

type runHandlerConfig struct {
	finder  SpectraFinderInterface
	runtime RuntimeInterface
}

// WithSpectraFinder sets a custom SpectraFinder for testing.
func WithSpectraFinder(finder SpectraFinderInterface) RunHandlerOption {
	return func(cfg *runHandlerConfig) {
		cfg.finder = finder
	}
}

// WithRuntime sets a custom Runtime for testing.
func WithRuntime(runtime RuntimeInterface) RunHandlerOption {
	return func(cfg *runHandlerConfig) {
		cfg.runtime = runtime
	}
}

// WithRunHandlerFunc creates a HandlerOption that configures the run command with options.
func WithRunHandlerFunc(opts ...RunHandlerOption) HandlerOption {
	return func(rootCfg *rootConfig) {
		rootCfg.runHandler = &runHandlerWrapper{opts: opts}
	}
}

// runHandlerWrapper wraps the run command execution for testing.
type runHandlerWrapper struct {
	opts []RunHandlerOption
}

// Execute implements SubcommandHandler interface (not used for run command).
func (w *runHandlerWrapper) Execute() int {
	// This should never be called - the run command uses RunE directly
	return 1
}

// newRunCommand creates the run subcommand with the given options.
func newRunCommand(opts []RunHandlerOption) *cobra.Command {
	var workflowName string

	runCmd := &cobra.Command{
		Use:   "run [WorkflowName]",
		Short: "Run a workflow",
		Long:  "Run a workflow",
		Example: `  spectra run SimpleSdd
  spectra run --workflow SimpleSdd`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Check for too many arguments
			if len(args) > 1 {
				return fmt.Errorf("too many arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Configure dependencies
			cfg := &runHandlerConfig{
				finder: &defaultSpectraFinder{},
			}
			for _, opt := range opts {
				opt(cfg)
			}

			// Determine workflow name (flag takes precedence over positional)
			var finalWorkflowName string
			if cmd.Flags().Changed("workflow") {
				finalWorkflowName = workflowName
			} else if len(args) > 0 {
				finalWorkflowName = args[0]
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: workflow name is required\n")
				return &exitError{code: 1}
			}

			// Validate workflow name is not empty after trimming
			trimmedName := strings.TrimSpace(finalWorkflowName)
			if trimmedName == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: workflow name cannot be empty\n")
				return &exitError{code: 1}
			}

			// Use the original (untrimmed) workflow name for execution
			finalWorkflowName = trimmedName

			// Find project root
			projectRoot, err := cfg.finder.Find()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: .spectra directory not found. Run 'spectra init' to initialize a project.\n")
				return &exitError{code: 1}
			}

			// Execute workflow via Runtime
			var exitCode int
			if cfg.runtime != nil {
				// Use injected runtime (for testing)
				exitCode, err = cfg.runtime.Run(projectRoot, finalWorkflowName, cmd.OutOrStdout(), cmd.ErrOrStderr())
				if err != nil {
					// Runtime returned an error - write to stderr if not already written
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
					if exitCode == 0 {
						exitCode = 1
					}
				}
			} else {
				// TODO: Use real runtime implementation when available
				// For now, just return success for minimal implementation
				exitCode = 0
			}

			if exitCode != 0 {
				return &exitError{code: exitCode}
			}
			return nil
		},
	}

	runCmd.Flags().StringVar(&workflowName, "workflow", "", "Workflow name to execute (alternative to positional argument)")

	return runCmd
}
