package spectra

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// RunRuntime defines the interface for runtime that returns an error
type RunRuntime interface {
	Run(workflowName string) error
}

// runHandlerWrapper wraps runtime options for injection
type runHandlerWrapper struct {
	opts []runOption
}

func (w *runHandlerWrapper) Execute() int {
	// This should not be called directly; the wrapper is only used
	// to pass options to newRunCommand
	return 0
}

// runOption configures the run command
type runOption func(*runConfig)

type runConfig struct {
	runtime RunRuntime
}

// WithRunRuntime sets the runtime for the run command
func WithRunRuntime(runtime RunRuntime) HandlerOption {
	return func(cfg *rootConfig) {
		cfg.runHandler = &runHandlerWrapper{
			opts: []runOption{
				func(rc *runConfig) {
					rc.runtime = runtime
				},
			},
		}
	}
}

// newRunCommand creates the run subcommand
func newRunCommand(opts []runOption) *cobra.Command {
	var workflow string

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a workflow",
		Long:  "Run a workflow by specifying its name with the --workflow flag",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply options to get runtime
			cfg := &runConfig{}
			for _, opt := range opts {
				opt(cfg)
			}

			// Execute the run logic
			exitCode := executeRun(cmd, workflow, cfg.runtime)
			if exitCode != 0 {
				return &exitError{code: exitCode}
			}
			return nil
		},
	}

	// Override Args validator to provide custom error message
	runCmd.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			//nolint:staticcheck // Error message format is specified by the test requirements
			return fmt.Errorf("unexpected argument '%s'. Use --workflow flag to specify workflow name.", args[0])
		}
		return nil
	}

	runCmd.Flags().StringVar(&workflow, "workflow", "", "Name of the workflow to execute (required)")

	// Customize flag parsing error messages
	runCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		// Convert "flag needs an argument" error to our custom message
		errMsg := err.Error()
		if strings.Contains(errMsg, "flag needs an argument") && strings.Contains(errMsg, "--workflow") {
			return fmt.Errorf("required flag --workflow not provided")
		}
		return err
	})

	return runCmd
}

// executeRun implements the run command logic
func executeRun(cmd *cobra.Command, workflow string, runtime RunRuntime) int {
	// Check if --workflow flag was provided
	if !cmd.Flags().Changed("workflow") {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: required flag --workflow not provided\n")
		return 1
	}

	// Check if workflow value is empty
	if workflow == "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: --workflow flag cannot be empty\n")
		return 1
	}

	// Runtime is required for actual execution
	if runtime == nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: runtime not configured\n")
		return 1
	}

	// Invoke runtime
	err := runtime.Run(workflow)

	// Handle success case
	if err == nil {
		return 0
	}

	// Determine exit code based on error message
	errMsg := err.Error()

	// Check for SIGINT first (priority over SIGTERM)
	if strings.Contains(errMsg, "session terminated by signal SIGINT") {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", errMsg)
		return 130
	}

	// Check for SIGTERM
	if strings.Contains(errMsg, "session terminated by signal SIGTERM") {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", errMsg)
		return 143
	}

	// Generic error
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", errMsg)
	return 1
}
