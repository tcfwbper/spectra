package spectra_agent_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra_agent "github.com/tcfwbper/spectra/cmd/spectra_agent"
)

// setupRootTestFixture creates a temporary test directory with .spectra/ directory.
func setupRootTestFixture(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	return tmpDir
}

// setupRootTestFixtureWithSubdir creates a temporary test directory with .spectra/ and nested subdirectory.
func setupRootTestFixtureWithSubdir(t *testing.T, subdirPath string) (string, string) {
	t.Helper()
	tmpDir := setupRootTestFixture(t)
	subdir := filepath.Join(tmpDir, subdirPath)
	require.NoError(t, os.MkdirAll(subdir, 0755))
	return tmpDir, subdir
}

// setupRootTestFixtureNoSpectra creates a temporary test directory without .spectra/ directory.
func setupRootTestFixtureNoSpectra(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// executeRootCommand creates and executes the root command with given args, returning
// stdout, stderr, and exit code.
func executeRootCommand(t *testing.T, workDir string, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := spectra_agent.NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	// Change to the working directory for SpectraFinder
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workDir))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// --- Happy Path — Subcommand Dispatch ---

// TestRootCommand_DispatchToEventEmit successfully dispatches to event emit subcommand.
func TestRootCommand_DispatchToEventEmit(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	// Register mock event emit handler
	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Event emit handler should be called")
	assert.Equal(t, mockHandler.ExitCode(), exitCode)
}

// TestRootCommand_DispatchToError successfully dispatches to error subcommand.
func TestRootCommand_DispatchToError(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithErrorHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"error", "test message", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Error handler should be called")
	assert.Equal(t, mockHandler.ExitCode(), exitCode)
}

// --- Happy Path — SpectraFinder Integration ---

// TestRootCommand_FindsProjectRoot uses SpectraFinder to locate project root from current directory.
func TestRootCommand_FindsProjectRoot(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "Subcommand should be executed after finding project root")
}

// TestRootCommand_FindsProjectRootFromSubdir locates project root when invoked from subdirectory.
func TestRootCommand_FindsProjectRootFromSubdir(t *testing.T) {
	_, subdir := setupRootTestFixtureWithSubdir(t, filepath.Join("subdir", "nested"))
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(subdir))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled(), "SpectraFinder should traverse upward and find .spectra/")
}

// --- Happy Path — Usage Information ---

// TestRootCommand_NoSubcommand prints usage information when invoked without subcommand.
func TestRootCommand_NoSubcommand(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	stdout, _, exitCode := executeRootCommand(t, projectRoot, []string{"--session-id", sessionID})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "spectra-agent - Interact with the Spectra workflow runtime")
}

// TestRootCommand_HelpFlag prints usage information when invoked with --help.
func TestRootCommand_HelpFlag(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	stdout, _, exitCode := executeRootCommand(t, projectRoot, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Available Commands")
	assert.Contains(t, stdout, "--session-id")
}

// TestRootCommand_SubcommandHelp prints subcommand-specific help.
func TestRootCommand_SubcommandHelp(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	stdout, _, exitCode := executeRootCommand(t, projectRoot, []string{"event", "--help"})

	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, stdout)
}

// --- Validation Failures — Missing Required Flag ---

// TestRootCommand_MissingSessionID returns exit code 1 when --session-id flag is missing.
func TestRootCommand_MissingSessionID(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{"event", "emit", "MyEvent"})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --session-id flag is required`, stderr)
}

// TestRootCommand_EmptySessionID returns exit code 1 when --session-id flag is empty string.
func TestRootCommand_EmptySessionID(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{"event", "emit", "MyEvent", "--session-id", ""})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --session-id flag is required`, stderr)
}

// --- Validation Failures — Unknown Subcommand ---

// TestRootCommand_UnknownSubcommand returns exit code 1 for unknown subcommand.
func TestRootCommand_UnknownSubcommand(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{"foo", "--session-id", sessionID})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: unknown command "foo" for "spectra-agent"`, stderr)
}

// TestRootCommand_UnknownNestedSubcommand returns exit code 1 for unknown nested subcommand.
func TestRootCommand_UnknownNestedSubcommand(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{"event", "unknown", "--session-id", sessionID})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
}

// --- Validation Failures — Project Root Not Found ---

// TestRootCommand_SpectraNotFound returns exit code 1 when .spectra directory not found in any ancestor.
func TestRootCommand_SpectraNotFound(t *testing.T) {
	projectRoot := setupRootTestFixtureNoSpectra(t)
	sessionID := uuid.New().String()

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{"event", "emit", "MyEvent", "--session-id", sessionID})

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: \.spectra directory not found\. Are you in a Spectra project\?`, stderr)
}

// TestRootCommand_SpectraNotFoundFromRoot returns exit code 1 when simulating search from filesystem root with no .spectra.
func TestRootCommand_SpectraNotFoundFromRoot(t *testing.T) {
	// Create a fixture that simulates a directory tree with no .spectra anywhere
	projectRoot := setupRootTestFixtureNoSpectra(t)
	sessionID := uuid.New().String()

	// Use a mock SpectraFinder configured to simulate traversal from root
	mockFinder := spectra_agent.NewMockSpectraFinder(func(startDir string) (string, error) {
		return "", fmt.Errorf("spectra not initialized")
	})

	cmd := spectra_agent.NewRootCommandWithFinder(mockFinder)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: \.spectra directory not found\. Are you in a Spectra project\?`, stderr.String())
}

// --- Validation Failures — Invalid Session ID Format ---

// TestRootCommand_InvalidUUIDFormat accepts invalid UUID format and passes to subcommand.
func TestRootCommand_InvalidUUIDFormat(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", "not-a-uuid"})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	// Root command should not validate UUID — it should dispatch to subcommand
	assert.True(t, mockHandler.WasCalled(), "Root command should not validate UUID format")
}

// --- Exit Code Propagation ---

// TestRootCommand_PropagatesExitCode0 propagates exit code 0 from successful subcommand without state modification.
func TestRootCommand_PropagatesExitCode0(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandlerWithExitCode(0)
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	assert.Equal(t, 0, exitCode)
}

// TestRootCommand_PropagatesExitCode2 propagates exit code 2 from subcommand transport error without state changes.
func TestRootCommand_PropagatesExitCode2(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandlerWithExitCode(2)
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	assert.Equal(t, 2, exitCode)
}

// TestRootCommand_PropagatesExitCode3 propagates exit code 3 from subcommand runtime error without state changes.
func TestRootCommand_PropagatesExitCode3(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandlerWithExitCode(3)
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	assert.Equal(t, 3, exitCode)
}

// --- Boundary Values — Edge Cases ---

// TestRootCommand_SessionIDWithSpecialChars accepts session ID with special characters.
func TestRootCommand_SessionIDWithSpecialChars(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", "abc-123-def-456"})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled())
}

// TestRootCommand_VeryLongSessionID accepts very long session ID value.
func TestRootCommand_VeryLongSessionID(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	longID := strings.Repeat("a", 1000)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", longID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled())
}

// TestRootCommand_MultipleFlags handles multiple flags including global and subcommand-specific.
func TestRootCommand_MultipleFlags(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID, "--message", "test"})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.True(t, mockHandler.WasCalled())
}

// --- Validation Failures — Flag Combinations ---

// TestRootCommand_DuplicateSessionIDFlag returns error when --session-id flag provided multiple times.
func TestRootCommand_DuplicateSessionIDFlag(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID1 := uuid.New().String()
	sessionID2 := uuid.New().String()

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{
		"event", "emit", "MyEvent", "--session-id", sessionID1, "--session-id", sessionID2,
	})

	assert.Equal(t, 2, exitCode)
	assert.NotEmpty(t, stderr)
}

// TestRootCommand_MalformedFlag returns error when flag is malformed.
func TestRootCommand_MalformedFlag(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{
		"event", "emit", "MyEvent", "--session-id",
	})

	assert.Equal(t, 1, exitCode)
	assert.NotEmpty(t, stderr)
}

// --- Idempotency ---

// TestRootCommand_RepeatedInvocation multiple invocations with same arguments produce consistent results.
func TestRootCommand_RepeatedInvocation(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	args := []string{"event", "emit", "MyEvent", "--session-id", sessionID}

	var exitCodes []int
	for i := 0; i < 3; i++ {
		mockHandler := spectra_agent.NewMockSubcommandHandlerWithExitCode(0)
		cmd := spectra_agent.NewRootCommandWithHandlers(
			spectra_agent.WithEventEmitHandler(mockHandler),
		)

		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs(args)

		origDir, _ := os.Getwd()
		require.NoError(t, os.Chdir(projectRoot))

		exitCode := cmd.Execute()
		exitCodes = append(exitCodes, exitCode)

		os.Chdir(origDir)
	}

	assert.Equal(t, exitCodes[0], exitCodes[1])
	assert.Equal(t, exitCodes[1], exitCodes[2])
}

// --- Mock / Dependency Interaction ---

// TestRootCommand_CallsSpectraFinder calls SpectraFinder before dispatching to subcommand.
func TestRootCommand_CallsSpectraFinder(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	finderCalled := false
	mockFinder := spectra_agent.NewMockSpectraFinder(func(startDir string) (string, error) {
		finderCalled = true
		return projectRoot, nil
	})

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithFinderAndHandlers(
		mockFinder,
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.True(t, finderCalled, "SpectraFinder should be called")
	assert.True(t, mockHandler.WasCalled(), "Subcommand should be called after SpectraFinder")
}

// TestRootCommand_DoesNotCallSocketClient root command does not perform socket operations directly.
func TestRootCommand_DoesNotCallSocketClient(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	mockSocketClient := spectra_agent.NewMockSocketClient()
	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
		spectra_agent.WithSocketClient(mockSocketClient),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	assert.False(t, mockSocketClient.WasCalled(), "Root command should not call SocketClient directly")
}

// --- Error Output Format ---

// TestRootCommand_ErrorPrefixFormat all error messages are prefixed with "Error: ".
func TestRootCommand_ErrorPrefixFormat(t *testing.T) {
	projectRoot := setupRootTestFixtureNoSpectra(t)
	sessionID := uuid.New().String()

	_, stderr, _ := executeRootCommand(t, projectRoot, []string{"event", "emit", "MyEvent", "--session-id", sessionID})

	assert.Regexp(t, `^Error: `, stderr)
}

// TestRootCommand_ErrorOutputToStderr error messages printed to stderr, not stdout.
func TestRootCommand_ErrorOutputToStderr(t *testing.T) {
	projectRoot := setupRootTestFixture(t)

	stdout, stderr, _ := executeRootCommand(t, projectRoot, []string{"event", "emit", "MyEvent"})

	// Error should be in stderr
	assert.NotEmpty(t, stderr)
	// Stdout should be empty or contain only usage info, not the error
	assert.NotContains(t, stdout, "Error: --session-id flag is required")
}

// --- State Isolation ---

// TestRootCommand_NoStateBetweenInvocations root command maintains no state between invocations.
func TestRootCommand_NoStateBetweenInvocations(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID1 := uuid.New().String()
	sessionID2 := uuid.New().String()

	// First invocation
	mockHandler1 := spectra_agent.NewMockSubcommandHandlerWithExitCode(0)
	cmd1 := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler1),
	)
	var stdout1, stderr1 bytes.Buffer
	cmd1.SetOut(&stdout1)
	cmd1.SetErr(&stderr1)
	cmd1.SetArgs([]string{"event", "emit", "Event1", "--session-id", sessionID1})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))

	exitCode1 := cmd1.Execute()

	// Second invocation with different args
	mockHandler2 := spectra_agent.NewMockSubcommandHandlerWithExitCode(0)
	cmd2 := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler2),
	)
	var stdout2, stderr2 bytes.Buffer
	cmd2.SetOut(&stdout2)
	cmd2.SetErr(&stderr2)
	cmd2.SetArgs([]string{"event", "emit", "Event2", "--session-id", sessionID2})

	exitCode2 := cmd2.Execute()
	os.Chdir(origDir)

	// Both should succeed independently
	assert.Equal(t, 0, exitCode1)
	assert.Equal(t, 0, exitCode2)
	assert.True(t, mockHandler1.WasCalled())
	assert.True(t, mockHandler2.WasCalled())
}

// --- Environment Variable Behavior ---

// TestRootCommand_IgnoresEnvironmentVariables does not read SPECTRA_SESSION_ID environment variable.
func TestRootCommand_IgnoresEnvironmentVariables(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	t.Setenv("SPECTRA_SESSION_ID", sessionID)

	_, stderr, exitCode := executeRootCommand(t, projectRoot, []string{"event", "emit", "MyEvent"})

	// Should still require --session-id flag
	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `Error: --session-id flag is required`, stderr)
}

// TestRootCommand_IgnoresClaudeSessionIDEnv does not read SPECTRA_CLAUDE_SESSION_ID environment variable.
func TestRootCommand_IgnoresClaudeSessionIDEnv(t *testing.T) {
	projectRoot := setupRootTestFixture(t)
	sessionID := uuid.New().String()

	t.Setenv("SPECTRA_CLAUDE_SESSION_ID", "some-claude-id")

	mockHandler := spectra_agent.NewMockSubcommandHandler()
	cmd := spectra_agent.NewRootCommandWithHandlers(
		spectra_agent.WithEventEmitHandler(mockHandler),
	)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"event", "emit", "MyEvent", "--session-id", sessionID})

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(origDir)

	cmd.Execute()

	// Subcommand should be called; the env var for claude-session-id should be ignored
	assert.True(t, mockHandler.WasCalled())
}
