package spectraagent

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tcfwbper/spectra/internal/cmdutil"
	"github.com/tcfwbper/spectra/storage"
)

// SpectraFinder defines the interface for project root discovery.
type SpectraFinder interface {
	FindProjectRoot(startDir string) (string, error)
}

// SendAndHandler defines the interface for sending messages and handling responses.
type SendAndHandler interface {
	SendAndHandle(sessionID, projectRoot string, message any, successText string) (int, string, string)
}

// defaultSpectraFinder is the production implementation using storage.FindSpectraRoot.
type defaultSpectraFinder struct{}

func (d *defaultSpectraFinder) FindProjectRoot(startDir string) (string, error) {
	return storage.FindSpectraRoot(startDir)
}

// defaultSendAndHandler is the production implementation that delegates to cmdutil.PublicSendAndHandle.
type defaultSendAndHandler struct{}

func (d *defaultSendAndHandler) SendAndHandle(sessionID, projectRoot string, message any, successText string) (int, string, string) {
	return cmdutil.PublicSendAndHandle(sessionID, projectRoot, message, successText)
}

// Options configures the root command with injectable dependencies.
type Options struct {
	Finder SpectraFinder
	Sender SendAndHandler
	Args   []string
}

// Execute builds and runs the spectra-agent command tree with production defaults.
// It returns the process exit code.
func Execute() int {
	return ExecuteWithOptions(Options{
		Finder: &defaultSpectraFinder{},
		Sender: &defaultSendAndHandler{},
		Args:   os.Args[1:],
	})
}

// ExecuteWithOptions builds and runs the spectra-agent command tree with the given options.
// It returns the process exit code.
func ExecuteWithOptions(opts Options) int {
	var stdoutBuf, stderrBuf bytes.Buffer

	rootCmd := buildRootCmd(opts, &stdoutBuf, &stderrBuf)
	rootCmd.SetArgs(opts.Args)

	err := rootCmd.Execute()

	// Print captured output.
	if stdoutBuf.Len() > 0 {
		fmt.Fprint(os.Stdout, stdoutBuf.String())
	}
	if stderrBuf.Len() > 0 {
		fmt.Fprint(os.Stderr, stderrBuf.String())
	}

	if err != nil {
		return 1
	}
	return 0
}

// RunForResult builds and runs the command tree capturing output.
// Returns the exit code, stdout string, and stderr string.
func RunForResult(opts Options) (exitCode int, stdout string, stderr string) {
	var stdoutBuf, stderrBuf bytes.Buffer

	rootCmd := buildRootCmd(opts, &stdoutBuf, &stderrBuf)
	rootCmd.SetArgs(opts.Args)

	err := rootCmd.Execute()

	exitCode = 0
	if err != nil {
		// Check if the error carries a specific exit code via the context.
		if code, ok := exitCodeFromError(err); ok {
			exitCode = code
		} else {
			exitCode = 1
			// If the error was not already written to stderr (e.g., unknown subcommand),
			// write it now.
			if stderrBuf.Len() == 0 {
				fmt.Fprintf(&stderrBuf, "Error: %s\n", err)
			}
		}
	}

	return exitCode, stdoutBuf.String(), stderrBuf.String()
}

// exitCodeError wraps an error with a specific exit code.
type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string { return e.err.Error() }
func (e *exitCodeError) Unwrap() error { return e.err }

func newExitCodeError(code int, msg string) error {
	return &exitCodeError{code: code, err: errors.New(msg)}
}

func exitCodeFromError(err error) (int, bool) {
	var ece *exitCodeError
	if errors.As(err, &ece) {
		return ece.code, true
	}
	return 0, false
}

// cmdContext holds shared state between the root command and subcommands.
type cmdContext struct {
	sessionID   string
	projectRoot string
	finder      SpectraFinder
	sender      SendAndHandler
}

func buildRootCmd(opts Options, stdoutBuf, stderrBuf *bytes.Buffer) *cobra.Command {
	ctx := &cmdContext{
		finder: opts.Finder,
		sender: opts.Sender,
	}

	rootCmd := &cobra.Command{
		Use:           "spectra-agent [command]",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate --session-id.
			sessionID, _ := cmd.Flags().GetString("session-id")
			if strings.TrimSpace(sessionID) == "" {
				fmt.Fprintln(stderrBuf, "Error: --session-id flag is required")
				return newExitCodeError(1, "--session-id flag is required")
			}
			ctx.sessionID = sessionID

			// Discover project root.
			projectRoot, err := ctx.finder.FindProjectRoot("")
			if err != nil {
				if errors.Is(err, storage.ErrNotInitialized) || err.Error() == storage.ErrNotInitialized.Error() {
					fmt.Fprintln(stderrBuf, "Error: .spectra directory not found. Are you in a Spectra project?")
					return newExitCodeError(1, ".spectra directory not found. Are you in a Spectra project?")
				}
				fmt.Fprintf(stderrBuf, "Error: %s\n", err)
				return newExitCodeError(1, err.Error())
			}
			ctx.projectRoot = projectRoot

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// No subcommand: print usage.
			cmd.SetOut(stdoutBuf)
			stdoutBuf.WriteString(cmd.UsageString())
			return nil
		},
	}

	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)

	// Register persistent flags.
	rootCmd.PersistentFlags().String("session-id", "", "Session ID for communication")

	// Override Cobra's built-in help to run through PersistentPreRunE before showing help.
	// When --session-id is provided, this ensures validation and FindProjectRoot always execute.
	// When --session-id is absent, help is shown directly (the spec says --help bypasses validation).
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Only run PersistentPreRunE if --session-id is provided.
		sessionID, _ := cmd.Flags().GetString("session-id")
		if strings.TrimSpace(sessionID) != "" && rootCmd.PersistentPreRunE != nil {
			if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
				// Validation failed; error already written to stderrBuf.
				return
			}
		}
		// Print usage to stdout.
		stdoutBuf.WriteString(cmd.UsageString())
	})
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no-op"})

	// Register subcommands.
	rootCmd.AddCommand(newErrorCmd(ctx, stdoutBuf, stderrBuf))
	rootCmd.AddCommand(newEventCmd(ctx, stdoutBuf, stderrBuf))

	return rootCmd
}
