package spectra_agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/storage"
)

// SubcommandHandler defines the interface for subcommand execution.
type SubcommandHandler interface {
	Execute() int
}

// CommandOption is a functional option for configuring the root command.
type CommandOption func(*rootCommandConfig)

// rootCommandConfig holds configuration for the root command.
type rootCommandConfig struct {
	finder           func(string) (string, error)
	eventEmitHandler SubcommandHandler
	errorHandler     SubcommandHandler
	socketClient     interface{ WasCalled() bool }
}

// WithEventEmitHandler sets the event emit handler.
func WithEventEmitHandler(handler SubcommandHandler) CommandOption {
	return func(cfg *rootCommandConfig) {
		cfg.eventEmitHandler = handler
	}
}

// WithErrorHandler sets the error handler.
func WithErrorHandler(handler SubcommandHandler) CommandOption {
	return func(cfg *rootCommandConfig) {
		cfg.errorHandler = handler
	}
}

// WithSocketClient sets the socket client (for testing).
func WithSocketClient(client interface{ WasCalled() bool }) CommandOption {
	return func(cfg *rootCommandConfig) {
		cfg.socketClient = client
	}
}

// exitCodeError is an error that carries an exit code.
type exitCodeError struct {
	code int
	msg  string
}

func (e *exitCodeError) Error() string {
	return e.msg
}

// RootCommand represents the root command.
type RootCommand struct {
	*cobra.Command
	config      *rootCommandConfig
	sessionID   string
	projectRoot string
}

// NewRootCommand creates a new root command with default configuration.
func NewRootCommand() *RootCommand {
	return newRootCommandWithConfig(&rootCommandConfig{
		finder: storage.SpectraFinder,
	})
}

// NewRootCommandWithHandlers creates a new root command with custom handlers.
func NewRootCommandWithHandlers(opts ...CommandOption) *RootCommand {
	cfg := &rootCommandConfig{
		finder: storage.SpectraFinder,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return newRootCommandWithConfig(cfg)
}

// NewRootCommandWithFinder creates a new root command with a custom finder.
func NewRootCommandWithFinder(finder func(string) (string, error)) *RootCommand {
	return newRootCommandWithConfig(&rootCommandConfig{
		finder: finder,
	})
}

// NewRootCommandWithFinderAndHandlers creates a new root command with custom finder and handlers.
func NewRootCommandWithFinderAndHandlers(finder func(string) (string, error), opts ...CommandOption) *RootCommand {
	cfg := &rootCommandConfig{
		finder: finder,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return newRootCommandWithConfig(cfg)
}

func newRootCommandWithConfig(cfg *rootCommandConfig) *RootCommand {
	rc := &RootCommand{
		config: cfg,
	}

	cmd := &cobra.Command{
		Use:   "spectra-agent",
		Short: "spectra-agent - Interact with the Spectra workflow runtime",
		Long:  "spectra-agent - Interact with the Spectra workflow runtime",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip validation for help commands
			if cmd.Name() == "help" || cmd.Name() == "__complete" || cmd.Name() == "__completeNoDesc" {
				return nil
			}

			// Check if this is just printing help
			if len(args) == 0 && cmd.Name() == "spectra-agent" {
				return nil
			}

			// Validate session ID
			if rc.sessionID == "" {
				return fmt.Errorf("Error: --session-id flag is required")
			}

			// Find project root
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("Error: .spectra directory not found. Are you in a Spectra project?")
			}

			projectRoot, err := cfg.finder(cwd)
			if err != nil {
				return fmt.Errorf("Error: .spectra directory not found. Are you in a Spectra project?")
			}

			rc.projectRoot = projectRoot
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Add persistent flags
	cmd.PersistentFlags().StringVar(&rc.sessionID, "session-id", "", "Session UUID (required)")

	// Add subcommands
	rc.addEventCommand(cmd)
	rc.addErrorCommand(cmd)

	rc.Command = cmd
	return rc
}

// addEventCommand adds the event subcommand.
func (rc *RootCommand) addEventCommand(parent *cobra.Command) {
	eventCmd := &cobra.Command{
		Use:   "event",
		Short: "Emit events to the workflow runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand specified, show help
			if len(args) == 0 {
				return fmt.Errorf("Error: must specify a subcommand")
			}
			// Unknown subcommand
			return fmt.Errorf("Error: unknown subcommand \"%s\" for \"event\"", args[0])
		},
	}

	emitCmd := &cobra.Command{
		Use:   "emit [eventType]",
		Short: "Emit an event",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rc.config.eventEmitHandler != nil {
				exitCode := rc.config.eventEmitHandler.Execute()
				if exitCode != 0 {
					return &exitCodeError{code: exitCode}
				}
				return nil
			}
			return fmt.Errorf("Error: event emit handler not implemented")
		},
	}

	// Add --message flag for emit subcommand
	emitCmd.Flags().String("message", "", "Message to include with event")

	eventCmd.AddCommand(emitCmd)
	parent.AddCommand(eventCmd)
}

// addErrorCommand adds the error subcommand.
func (rc *RootCommand) addErrorCommand(parent *cobra.Command) {
	errorCmd := &cobra.Command{
		Use:   "error [message]",
		Short: "Report unrecoverable errors to the workflow runtime",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rc.config.errorHandler != nil {
				exitCode := rc.config.errorHandler.Execute()
				if exitCode != 0 {
					return &exitCodeError{code: exitCode}
				}
				return nil
			}
			return fmt.Errorf("Error: error handler not implemented")
		},
	}

	parent.AddCommand(errorCmd)
}

// Execute runs the root command and returns the exit code.
func (rc *RootCommand) Execute() int {
	if err := rc.Command.Execute(); err != nil {
		// Check if it's an exit code error from a subcommand
		if ecErr, ok := err.(*exitCodeError); ok {
			return ecErr.code
		}

		errMsg := err.Error()
		// Add "Error: " prefix if not already present
		if !strings.HasPrefix(errMsg, "Error: ") {
			errMsg = "Error: " + errMsg
		}
		_, _ = fmt.Fprintf(rc.Command.ErrOrStderr(), "%s\n", errMsg)
		return 1
	}
	return 0
}
