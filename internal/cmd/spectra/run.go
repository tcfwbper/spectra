package spectra

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/tcfwbper/spectra/internal/cmdutil"
	"github.com/tcfwbper/spectra/logger"
)

// RunRuntime defines the interface for workflow execution used by the run command.
// The runtime executes the named workflow, using the provided logger for structured output.
type RunRuntime interface {
	Run(workflowName string, log logger.Logger) (int, error)
}

// RunCommandOptions holds injectable dependencies for the run command.
type RunCommandOptions struct {
	Runtime          RunRuntime
	Workflow         string
	WorkflowProvided bool
	Args             []string
	Stdout           io.Writer
	Stderr           io.Writer
	Logger           logger.Logger
}

// RunRunCommand executes the run command logic with the given options.
// Returns the exit code.
func RunRunCommand(opts RunCommandOptions) int {
	// Validate: check for positional arguments
	if len(opts.Args) > 0 {
		_, _ = fmt.Fprintf(opts.Stderr, "%s\n",
			cmdutil.FormatError(fmt.Sprintf("unexpected argument '%s'. Use --workflow flag to specify workflow name.", opts.Args[0])))
		return 1
	}

	// Validate: --workflow must be provided
	if !opts.WorkflowProvided {
		_, _ = fmt.Fprintf(opts.Stderr, "%s\n",
			cmdutil.FormatError("required flag --workflow not provided"))
		return 1
	}

	// Validate: --workflow must not be empty
	if opts.Workflow == "" {
		_, _ = fmt.Fprintf(opts.Stderr, "%s\n",
			cmdutil.FormatError("--workflow flag cannot be empty"))
		return 1
	}

	// Construct logger for runtime if not injected
	log := opts.Logger
	if log == nil {
		slogger := slog.New(slog.NewTextHandler(opts.Stderr, nil))
		log = logger.NewSlogLogger(slogger)
	}

	// Invoke runtime
	exitCode, err := opts.Runtime.Run(opts.Workflow, log)
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "%s\n", cmdutil.FormatError(err.Error()))
		return mapSignalExitCode(exitCode, err)
	}

	return exitCode
}
