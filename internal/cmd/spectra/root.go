package spectra

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tcfwbper/spectra/internal/cmdutil"
	"github.com/tcfwbper/spectra/logger"
	runtimex "github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// version is the current version of the spectra CLI.
const version = "1.1.1"

// NewRootCommand creates and returns the root cobra.Command for the spectra CLI.
// It registers global flags (--version) and all subcommands (init, run, clear).
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "spectra [command]",
		Short:         "Spectra CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// No subcommand: print usage
			return cmd.Help()
		},
	}

	// Override version template to match spec format
	cmd.SetVersionTemplate("spectra version {{.Version}}\n")

	// Register subcommands
	cmd.AddCommand(newInitCobraCommand())
	cmd.AddCommand(newRunCobraCommand())
	cmd.AddCommand(newClearCobraCommand())

	// Initialize annotations map for exit code storage
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	return cmd
}

// RootCommandOptions holds configuration for executing the root command.
type RootCommandOptions struct {
	Stdout io.Writer
	Stderr io.Writer
	Args   []string
}

// ExecuteWithOptions runs the root command with the given options and returns the exit code.
func ExecuteWithOptions(opts RootCommandOptions) int {
	cmd := NewRootCommand()
	return executeCommand(cmd, opts)
}

// Execute builds and runs the Cobra command tree, returning the process exit code.
func Execute() int {
	return ExecuteWithOptions(RootCommandOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Args:   os.Args[1:],
	})
}

// executeCommand runs a cobra command with the given options and returns the exit code.
func executeCommand(cmd *cobra.Command, opts RootCommandOptions) int {
	stdout := opts.Stdout
	stderr := opts.Stderr
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(opts.Args)

	err := cmd.Execute()
	if err != nil {
		// Print error to stderr
		_, _ = fmt.Fprintln(stderr, cmdutil.FormatError(err.Error()))
		return 1
	}

	// Check if the command stored an exit code
	if code, ok := getExitCode(cmd); ok {
		return code
	}

	return 0
}

// exitCodeKey is used to store exit codes in cobra command annotations.
const exitCodeKey = "_spectra_exit_code"

// setExitCode stores an exit code in the command's context for retrieval after execution.
func setExitCode(cmd *cobra.Command, code int) {
	// Walk up to the root command to store the exit code
	root := cmd.Root()
	if root.Annotations == nil {
		root.Annotations = make(map[string]string)
	}
	root.Annotations[exitCodeKey] = fmt.Sprintf("%d", code)
}

// getExitCode retrieves the stored exit code from the root command annotations.
func getExitCode(cmd *cobra.Command) (int, bool) {
	root := cmd.Root()
	if root.Annotations == nil {
		return 0, false
	}
	val, ok := root.Annotations[exitCodeKey]
	if !ok {
		return 0, false
	}
	var code int
	_, err := fmt.Sscanf(val, "%d", &code)
	if err != nil {
		return 0, false
	}
	return code, true
}

// newInitCobraCommand creates the cobra.Command for the "init" subcommand.
func newInitCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Spectra project",
		RunE: func(cmd *cobra.Command, args []string) error {
			layout := &productionStorageLayout{}
			copier := NewBuiltinResourceCopier(
				builtinWorkflowsFS,
				builtinAgentsFS,
				builtinSpecFilesFS,
				layout,
			)
			opts := InitCommandOptions{
				GetwdFunc:        os.Getwd,
				GitignoreEnsurer: NewGitignoreEnsurer(),
				DirectoryCreator: NewDirectoryCreator(),
				Copier:           copier,
				Stdout:           cmd.OutOrStdout(),
				Stderr:           cmd.ErrOrStderr(),
			}
			code := RunInitCommand(opts)
			if code != 0 {
				setExitCode(cmd, code)
			}
			return nil
		},
	}
	return cmd
}

// newRunCobraCommand creates the cobra.Command for the "run" subcommand.
func newRunCobraCommand() *cobra.Command {
	var workflow string

	var sessionID string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a Spectra workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowProvided := cmd.Flags().Changed("workflow")

			opts := RunCommandOptions{
				Runtime:          &productionRuntime{},
				Workflow:         workflow,
				WorkflowProvided: workflowProvided,
				SessionID:        sessionID,
				Args:             args,
				Stdout:           cmd.OutOrStdout(),
				Stderr:           cmd.ErrOrStderr(),
			}
			code := RunRunCommand(opts)
			if code != 0 {
				setExitCode(cmd, code)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&workflow, "workflow", "", "Name of the workflow to execute")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "User-specified session UUID")
	return cmd
}

// newClearCobraCommand creates the cobra.Command for the "clear" subcommand.
func newClearCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear session data",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ClearCommandOptions{
				Finder: &productionClearSpectraFinder{},
				Layout: &productionClearStorageLayout{},
				Stdin:  os.Stdin,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
				Args:   args,
			}
			code := RunClearCommand(opts)
			if code != 0 {
				setExitCode(cmd, code)
			}
			return nil
		},
	}
	return cmd
}

// newRootCommandForTest creates a root command without real subcommands.
// Used for testing with stub subcommands.
func newRootCommandForTest() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "spectra [command]",
		Short:         "Spectra CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.SetVersionTemplate("spectra version {{.Version}}\n")
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	return cmd
}

// newStubSubcommand creates a stub subcommand that sets a specific exit code.
func newStubSubcommand(name string, exitCode int) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Stub command for testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if exitCode != 0 {
				setExitCode(cmd, exitCode)
			}
			return nil
		},
	}
}

// productionRuntime is the production adapter that delegates to the runtime package.
type productionRuntime struct{}

func (p *productionRuntime) Run(workflowName string, sessionID string, log logger.Logger) (int, error) {
	return runtimex.Run(workflowName, sessionID, log)
}

// productionStorageLayout is the production adapter for StorageLayoutInterface used by BuiltinResourceCopier.
type productionStorageLayout struct{}

func (p *productionStorageLayout) GetWorkflowPath(projectRoot, name string) string {
	return storage.GetWorkflowPath(projectRoot, name)
}

func (p *productionStorageLayout) GetAgentPath(projectRoot, name string) string {
	return storage.GetAgentPath(projectRoot, name)
}

// productionClearSpectraFinder is the production adapter for ClearSpectraFinder.
type productionClearSpectraFinder struct{}

func (p *productionClearSpectraFinder) Find() (string, error) {
	return storage.FindSpectraRoot("")
}

// productionClearStorageLayout is the production adapter for ClearStorageLayout.
type productionClearStorageLayout struct{}

func (p *productionClearStorageLayout) GetSessionDir(projectRoot, uuid string) string {
	return storage.GetSessionDir(projectRoot, uuid)
}

func (p *productionClearStorageLayout) GetSessionsDir(projectRoot string) string {
	return storage.GetSessionsDir(projectRoot)
}

// signalInterruptSubstring is the string to detect SIGINT-based termination.
const signalInterruptSubstring = "session terminated by signal interrupt"

// signalTerminatedSubstring is the string to detect SIGTERM-based termination.
const signalTerminatedSubstring = "session terminated by signal terminated"

// mapSignalExitCode maps a runtime error to appropriate signal exit codes.
// Returns the mapped exit code.
func mapSignalExitCode(runtimeExitCode int, err error) int {
	if err == nil {
		return runtimeExitCode
	}
	msg := err.Error()
	if strings.Contains(msg, signalInterruptSubstring) {
		return cmdutil.ExitSignalINT
	}
	if strings.Contains(msg, signalTerminatedSubstring) {
		return cmdutil.ExitSignalTERM
	}
	return runtimeExitCode
}
