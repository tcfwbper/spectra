package spectra

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Runtime interface for workflow execution
type Runtime interface {
	Run(workflowName string) int
}

// defaultRuntimeFactory is a function that creates a default Runtime instance
// This can be overridden in tests
var defaultRuntimeFactory func() (Runtime, error)

// runCommandOption is a function that configures the run command
type runCommandOption func(*runCommandConfig)

type runCommandConfig struct {
	runtime Runtime
}

// WithRuntime sets a custom runtime for the run command
func WithRuntime(rt Runtime) runCommandOption {
	return func(cfg *runCommandConfig) {
		cfg.runtime = rt
	}
}

// runHandlerWrapper wraps run command options for dependency injection in tests
type runHandlerWrapper struct {
	opts []runCommandOption
}

// Execute is a no-op implementation for SubcommandHandler compatibility
func (w *runHandlerWrapper) Execute() int {
	return 0
}

// WithRunHandlerFunc creates a HandlerOption that carries run command options
func WithRunHandlerFunc(opts ...runCommandOption) HandlerOption {
	return WithRunHandler(&runHandlerWrapper{opts: opts})
}

// newRunCommand creates the run subcommand
func newRunCommand(opts []runCommandOption) *cobra.Command {
	cfg := &runCommandConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var workflowFlag string

	cmd := &cobra.Command{
		Use:   "run [flags] <WorkflowName>",
		Short: "Run a workflow",
		Long:  "Run a workflow",
		Example: `  spectra run DefaultLogicSpec
  spectra run --workflow DefaultLogicSpec`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine workflow name: flag takes precedence over positional argument
			flagChanged := cmd.Flags().Changed("workflow")

			var workflowName string
			if flagChanged {
				// Flag was explicitly provided, use its value (even if empty)
				workflowName = workflowFlag
			} else {
				// No flag, use positional argument
				if len(args) == 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: workflow name is required\n")
					return &exitError{code: 1}
				}
				if len(args) > 1 {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: too many arguments\n")
					return &exitError{code: 1}
				}
				workflowName = args[0]
			}

			// Validate workflow name is non-empty after trimming
			if strings.TrimSpace(workflowName) == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: workflow name cannot be empty\n")
				return &exitError{code: 1}
			}

			// Get runtime
			rt := cfg.runtime
			if rt == nil {
				// Use default runtime factory
				if defaultRuntimeFactory != nil {
					var err error
					rt, err = defaultRuntimeFactory()
					if err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to create runtime: %v\n", err)
						return &exitError{code: 1}
					}
				} else {
					// No runtime available - this should not happen in production
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: runtime not configured\n")
					return &exitError{code: 1}
				}
			}

			// Run workflow
			exitCode := rt.Run(workflowName)

			if exitCode != 0 {
				return &exitError{code: exitCode}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&workflowFlag, "workflow", "", "Workflow name to execute (alternative to positional argument)")

	return cmd
}
