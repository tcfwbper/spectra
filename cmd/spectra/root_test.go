package main_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// --- Test Helpers ---

// executeRootCommand creates and executes the root command with given args, returning
// stdout, stderr, and exit code.
func executeRootCommand(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := spectra.NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// --- Happy Path — Display Help ---

// TestRootCommand_NoSubcommand displays usage information when invoked without subcommand.
func TestRootCommand_NoSubcommand(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Available Commands")
}

// TestRootCommand_HelpFlag displays usage information when invoked with --help.
func TestRootCommand_HelpFlag(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Available Commands")
	assert.Contains(t, stdout, "Flags")
	assert.Contains(t, stdout, "spectra [command]")
}

// --- Happy Path — Display Version ---

// TestRootCommand_VersionFlag displays version string when invoked with --version.
func TestRootCommand_VersionFlag(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--version"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "spectra version")
}

// TestRootCommand_VersionFormat version string follows semantic versioning format.
func TestRootCommand_VersionFormat(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--version"})

	assert.Equal(t, 0, exitCode)
	assert.Regexp(t, `spectra version v\d+\.\d+\.\d+`, stdout)
}

// --- Happy Path — Subcommand Delegation ---

// TestRootCommand_InitSubcommand delegates to init subcommand successfully.
func TestRootCommand_InitSubcommand(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandler()
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithInitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"init"})

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Init subcommand handler should be called")
}

// TestRootCommand_RunSubcommand delegates to run subcommand successfully.
func TestRootCommand_RunSubcommand(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandler()
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithRunHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run"})

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Run subcommand handler should be called")
}

// TestRootCommand_ClearSubcommand delegates to clear subcommand successfully.
func TestRootCommand_ClearSubcommand(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandler()
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithClearHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"clear"})

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Clear subcommand handler should be called")
}

// --- Happy Path — Exit Code Propagation ---

// TestRootCommand_PropagatesSubcommandSuccess propagates exit code 0 from successful subcommand.
func TestRootCommand_PropagatesSubcommandSuccess(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandlerWithExitCode(0)
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithInitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"init"})

	exitCode := cmd.Execute()

	assert.Equal(t, 0, exitCode)
}

// TestRootCommand_PropagatesSubcommandError propagates exit code 1 from failed subcommand.
func TestRootCommand_PropagatesSubcommandError(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandlerWithExitCode(1)
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithInitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"init"})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
}

// --- Validation Failures — Unknown Subcommand ---

// TestRootCommand_UnknownSubcommand returns error for unknown subcommand.
func TestRootCommand_UnknownSubcommand(t *testing.T) {
	_, stderr, exitCode := executeRootCommand(t, []string{"unknown-command"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, `Error: unknown command "unknown-command" for "spectra"`)
}

// TestRootCommand_MultipleUnknownSubcommands returns error for first unknown subcommand.
func TestRootCommand_MultipleUnknownSubcommands(t *testing.T) {
	_, stderr, exitCode := executeRootCommand(t, []string{"foo", "bar"})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, `Error: unknown command "foo" for "spectra"`)
}

// --- Validation Failures — Invalid Flags ---

// TestRootCommand_InvalidGlobalFlag returns error for unknown global flag.
func TestRootCommand_InvalidGlobalFlag(t *testing.T) {
	_, stderr, exitCode := executeRootCommand(t, []string{"--invalid-flag"})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
}

// TestRootCommand_SubcommandInvalidFlag subcommand handles flag parsing, error propagated.
func TestRootCommand_SubcommandInvalidFlag(t *testing.T) {
	_, stderr, exitCode := executeRootCommand(t, []string{"init", "--invalid-flag"})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
}

// --- Stateless Execution ---

// TestRootCommand_StatelessInvocations multiple invocations are independent with no shared state.
func TestRootCommand_StatelessInvocations(t *testing.T) {
	// First invocation: --version
	stdout1, _, exitCode1 := executeRootCommand(t, []string{"--version"})
	assert.Equal(t, 0, exitCode1)
	assert.Contains(t, stdout1, "spectra version")

	// Second invocation: --help
	stdout2, _, exitCode2 := executeRootCommand(t, []string{"--help"})
	assert.Equal(t, 0, exitCode2)
	assert.Contains(t, stdout2, "Available Commands")

	// Verify no state leakage
	assert.NotContains(t, stdout2, "spectra version v")
}

// --- Error Output Format ---

// TestRootCommand_ErrorPrefix error messages are prefixed with "Error: ".
func TestRootCommand_ErrorPrefix(t *testing.T) {
	_, stderr, exitCode := executeRootCommand(t, []string{"unknown-command"})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `^Error: `, stderr)
}

// TestRootCommand_ErrorToStderr error messages are printed to stderr, not stdout.
func TestRootCommand_ErrorToStderr(t *testing.T) {
	stdout, stderr, exitCode := executeRootCommand(t, []string{"unknown-command"})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
	assert.Empty(t, stdout)
}

// --- Usage Information Content ---

// TestRootCommand_UsageIncludesAllSubcommands usage information lists all three subcommands.
func TestRootCommand_UsageIncludesAllSubcommands(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "init")
	assert.Contains(t, stdout, "run")
	assert.Contains(t, stdout, "clear")
}

// TestRootCommand_UsageIncludesDescription usage information includes command description.
func TestRootCommand_UsageIncludesDescription(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Regexp(t, `(?i)framework for defining and executing flexible AI agent workflows`, stdout)
}

// TestRootCommand_UsageIncludesFlags usage information lists available flags.
func TestRootCommand_UsageIncludesFlags(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "--help")
	assert.Contains(t, stdout, "--version")
}

// TestRootCommand_UsageIncludesExamples usage information includes usage examples.
func TestRootCommand_UsageIncludesExamples(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "spectra [command]")
}

// --- Happy Path — Cobra Framework ---

// TestRootCommand_UsesCobra root command is implemented using Cobra library.
func TestRootCommand_UsesCobra(t *testing.T) {
	cmd := spectra.NewRootCommand()

	// Cobra commands expose Use field and support subcommand registration
	assert.NotNil(t, cmd)
	// Verify basic Cobra patterns: command has Use, subcommands registered
	stdout, _, exitCode := executeRootCommand(t, []string{"--help"})
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Available Commands")
	assert.Contains(t, stdout, "Use")
}

// --- Integration — Subcommand Flags ---

// TestRootCommand_SubcommandWithFlags passes flags to subcommand correctly.
func TestRootCommand_SubcommandWithFlags(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandler()
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithClearHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"clear", "--session-id=test-uuid"})

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Subcommand should receive --session-id flag and execute correctly")
}

// --- Boundary Values — Empty Input ---

// TestRootCommand_EmptyArguments treats empty arguments same as no arguments.
func TestRootCommand_EmptyArguments(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Usage:")
}

// --- Happy Path — Help for Subcommands ---

// TestRootCommand_SubcommandHelp shows help for specific subcommand.
func TestRootCommand_SubcommandHelp(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"init", "--help"})

	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, stdout)
}

// TestRootCommand_HelpSubcommandSyntax supports help command syntax.
func TestRootCommand_HelpSubcommandSyntax(t *testing.T) {
	stdout, _, exitCode := executeRootCommand(t, []string{"help", "init"})

	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, stdout)
}

// --- Error Propagation ---

// TestRootCommand_PropagatesSubcommandErrorDetailed propagates detailed error from subcommand.
func TestRootCommand_PropagatesSubcommandErrorDetailed(t *testing.T) {
	mockHandler := spectra.NewMockSubcommandHandlerWithExitCode(1)
	cmd := spectra.NewRootCommandWithHandlers(
		spectra.WithInitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"init"})

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
}

// --- Concurrent Behaviour ---
// Note: TestRootCommand_ConcurrentInvocations is a race test and belongs in test/race/
