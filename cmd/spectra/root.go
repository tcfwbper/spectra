package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "v0.1.0"

// exitError is a custom error that carries an exit code
type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

// SubcommandHandler defines the interface for subcommand handlers
type SubcommandHandler interface {
	Execute() int
}

// RootCommand wraps a cobra.Command to provide an Execute method returning exit code
type RootCommand struct {
	*cobra.Command
}

// Execute runs the root command and returns exit code
func (rc *RootCommand) Execute() int {
	err := rc.Command.Execute()
	if err != nil {
		return 1
	}
	return 0
}

// HandlerOption is a function that configures handlers
type HandlerOption func(*rootConfig)

type rootConfig struct {
	initHandler  SubcommandHandler
	runHandler   SubcommandHandler
	clearHandler SubcommandHandler
}

// WithInitHandler sets the init subcommand handler
func WithInitHandler(handler SubcommandHandler) HandlerOption {
	return func(cfg *rootConfig) {
		cfg.initHandler = handler
	}
}

// WithRunHandler sets the run subcommand handler
func WithRunHandler(handler SubcommandHandler) HandlerOption {
	return func(cfg *rootConfig) {
		cfg.runHandler = handler
	}
}

// WithClearHandler sets the clear subcommand handler
func WithClearHandler(handler SubcommandHandler) HandlerOption {
	return func(cfg *rootConfig) {
		cfg.clearHandler = handler
	}
}

// NewRootCommand creates the root command for spectra CLI
func NewRootCommand() *RootCommand {
	return NewRootCommandWithHandlers()
}

// NewRootCommandWithHandlers creates the root command with optional handler overrides
func NewRootCommandWithHandlers(opts ...HandlerOption) *RootCommand {
	cfg := &rootConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	rootCmd := &cobra.Command{
		Use:     "spectra",
		Short:   "Framework for defining and executing flexible AI agent workflows",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Set custom version template
	rootCmd.SetVersionTemplate("spectra version " + version + "\n")

	// Add init subcommand
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Spectra project",
		RunE: func(cmd *cobra.Command, args []string) error {
			var exitCode int
			if cfg.initHandler != nil {
				exitCode = cfg.initHandler.Execute()
			} else {
				// Default implementation
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to get current directory: %v\n", err)
					return &exitError{code: 1}
				}
				copier := NewBuiltinResourceCopier(builtinWorkflows, builtinAgents, builtinSpecFiles)
				handler := NewInitHandlerWithOutput(cwd, copier, cmd.OutOrStdout(), cmd.ErrOrStderr())
				exitCode = handler.Execute()
			}
			if exitCode != 0 {
				return &exitError{code: exitCode}
			}
			return nil
		},
	}
	rootCmd.AddCommand(initCmd)

	// Add run subcommand
	var runCmd *cobra.Command
	if cfg.runHandler != nil {
		// Use test handler wrapper to get options
		if wrapper, ok := cfg.runHandler.(*runHandlerWrapper); ok {
			runCmd = newRunCommand(wrapper.opts)
		} else {
			// Fallback for legacy test handlers
			runCmd = &cobra.Command{
				Use:   "run",
				Short: "Run a workflow",
				RunE: func(cmd *cobra.Command, args []string) error {
					exitCode := cfg.runHandler.Execute()
					if exitCode != 0 {
						return &exitError{code: exitCode}
					}
					return nil
				},
			}
		}
	} else {
		// Default implementation
		runCmd = newRunCommand(nil)
	}
	rootCmd.AddCommand(runCmd)

	// Add clear subcommand
	clearCmd := newClearCommand(cfg.clearHandler)
	rootCmd.AddCommand(clearCmd)

	return &RootCommand{Command: rootCmd}
}
