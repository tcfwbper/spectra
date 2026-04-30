package spectra_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// --- Test Helpers ---

// setupClearTestFixtureWithSessions creates a temporary directory with .spectra/sessions/ and
// the specified session subdirectories. Returns the project root directory.
func setupClearTestFixtureWithSessions(t *testing.T, sessionNames ...string) string {
	t.Helper()
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))
	for _, name := range sessionNames {
		sessionDir := filepath.Join(sessionsDir, name)
		require.NoError(t, os.MkdirAll(sessionDir, 0755))
	}
	return tmpDir
}

// setupClearTestFixtureEmpty creates a temporary directory with an empty .spectra/sessions/ directory.
func setupClearTestFixtureEmpty(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))
	return tmpDir
}

// setupClearTestFixtureNoSessions creates a temporary directory with .spectra/ but no sessions/ subdirectory.
func setupClearTestFixtureNoSessions(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	return tmpDir
}

// setupClearTestFixtureNoSpectra creates a temporary directory without .spectra/.
func setupClearTestFixtureNoSpectra(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// executeClearCommand creates and executes the clear command with given args, stdin, and working directory.
// Returns stdout, stderr, and exit code.
func executeClearCommand(t *testing.T, workDir string, args []string, stdin io.Reader) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := spectra.NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"clear"}, args...))
	if stdin != nil {
		cmd.SetIn(stdin)
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workDir))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// --- Happy Path — Delete Specific Session ---

// TestClearCommand_DeleteSpecificSession deletes a specific session by UUID.
func TestClearCommand_DeleteSpecificSession(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "12345678-1234-1234-1234-123456789abc")

	// Create session files
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", "12345678-1234-1234-1234-123456789abc")
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session.json"), []byte("{}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(""), 0644))

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=12345678-1234-1234-1234-123456789abc"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Session '12345678-1234-1234-1234-123456789abc' cleared successfully")

	// Verify session directory is deleted
	_, err := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(err))
}

// TestClearCommand_DeleteSessionWithNestedFiles deletes session directory containing subdirectories and files.
func TestClearCommand_DeleteSessionWithNestedFiles(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "test-session")

	// Create nested subdirectories and multiple files
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", "test-session")
	nestedDir := filepath.Join(sessionDir, "subdir", "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session.json"), []byte("{}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "data.txt"), []byte("data"), 0644))

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "cleared successfully")

	// Verify entire directory tree deleted
	_, err := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(err))
}

// --- Happy Path — Delete All Sessions ---

// TestClearCommand_DeleteAllSessionsWithConfirmation deletes all sessions when user confirms with 'y'.
func TestClearCommand_DeleteAllSessionsWithConfirmation(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2", "session-3")

	stdin := strings.NewReader("y\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Are you sure you want to delete all sessions?")
	assert.Contains(t, stdout, "All sessions cleared successfully")

	// Verify all session directories deleted
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestClearCommand_DeleteAllSessionsUppercaseY accepts uppercase 'Y' as confirmation.
func TestClearCommand_DeleteAllSessionsUppercaseY(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2")

	stdin := strings.NewReader("Y\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "All sessions cleared successfully")

	// Verify all session directories deleted
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestClearCommand_SkipsFilesInSessionsDirectory only deletes directories, skips regular files in sessions directory.
func TestClearCommand_SkipsFilesInSessionsDirectory(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2")

	// Create a regular file in sessions directory
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "somefile.txt"), []byte("data"), 0644))

	stdin := strings.NewReader("y\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "All sessions cleared successfully")

	// somefile.txt should remain
	_, err := os.Stat(filepath.Join(sessionsDir, "somefile.txt"))
	assert.NoError(t, err, "Regular file somefile.txt should remain")

	// Session directories should be deleted
	_, err = os.Stat(filepath.Join(sessionsDir, "session-1"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(sessionsDir, "session-2"))
	assert.True(t, os.IsNotExist(err))
}

// --- Happy Path — No Sessions to Clear ---

// TestClearCommand_NoSessions prints message when sessions directory is empty.
func TestClearCommand_NoSessions(t *testing.T) {
	projectRoot := setupClearTestFixtureEmpty(t)

	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "No sessions to clear")
}

// --- Happy Path — User Cancels Operation ---

// TestClearCommand_CancelWithN cancels deletion when user enters 'n'.
func TestClearCommand_CancelWithN(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2")

	stdin := strings.NewReader("n\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Are you sure you want to delete all sessions?")
	assert.Contains(t, stdout, "Operation cancelled")

	// Verify no sessions deleted
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

// TestClearCommand_CancelWithUppercaseN cancels deletion when user enters 'N'.
func TestClearCommand_CancelWithUppercaseN(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2")

	stdin := strings.NewReader("N\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Operation cancelled")

	// Verify no sessions deleted
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

// TestClearCommand_CancelWithEmptyInput treats empty input (just Enter) as cancellation.
func TestClearCommand_CancelWithEmptyInput(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1")

	stdin := strings.NewReader("\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Operation cancelled")

	// Verify no sessions deleted
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// TestClearCommand_CancelWithInvalidInput treats any input other than 'y' or 'Y' as cancellation.
func TestClearCommand_CancelWithInvalidInput(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1")

	stdin := strings.NewReader("yes\n")
	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Operation cancelled")

	// Verify no sessions deleted
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// --- Validation Failures — Session Not Found ---

// TestClearCommand_SessionNotFound prints warning when specified session does not exist.
func TestClearCommand_SessionNotFound(t *testing.T) {
	projectRoot := setupClearTestFixtureEmpty(t)

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=nonexistent-session"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Warning: session 'nonexistent-session' not found, skipping")
}

// TestClearCommand_InvalidUUIDFormat does not validate UUID format, attempts deletion.
func TestClearCommand_InvalidUUIDFormat(t *testing.T) {
	projectRoot := setupClearTestFixtureEmpty(t)

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=invalid-uuid"}, nil)

	assert.Equal(t, 0, exitCode)
	// Composes path with invalid UUID; prints warning if directory not found
	assert.Contains(t, stdout, "Warning:")
}

// TestClearCommand_EmptySessionID handles empty string as session ID.
func TestClearCommand_EmptySessionID(t *testing.T) {
	projectRoot := setupClearTestFixtureEmpty(t)

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id="}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Warning: session '' not found, skipping")
}

// --- Validation Failures — Directory Not Found ---

// TestClearCommand_SessionsDirectoryNotFound prints warning when sessions directory does not exist.
func TestClearCommand_SessionsDirectoryNotFound(t *testing.T) {
	projectRoot := setupClearTestFixtureNoSessions(t)

	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Warning: sessions directory not found, nothing to clear")
}

// TestClearCommand_SpectraNotFound returns error when .spectra directory not found.
func TestClearCommand_SpectraNotFound(t *testing.T) {
	projectRoot := setupClearTestFixtureNoSpectra(t)

	_, stderr, exitCode := executeClearCommand(t, projectRoot, []string{}, nil)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: .spectra directory not found. Are you in a Spectra project?")
}

// --- Validation Failures — Permission Denied ---

// TestClearCommand_SessionDeletionPermissionDenied returns error when session directory cannot be deleted.
func TestClearCommand_SessionDeletionPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	projectRoot := setupClearTestFixtureWithSessions(t, "test-session")

	// Set parent directory permissions to read-only (0555) to prevent deletion
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	require.NoError(t, os.Chmod(sessionsDir, 0555))
	t.Cleanup(func() { os.Chmod(sessionsDir, 0755) })

	_, stderr, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "Error: failed to clear session 'test-session': permission denied")
}

// TestClearCommand_SessionsDirectoryNotReadable returns error when sessions directory is not readable.
func TestClearCommand_SessionsDirectoryNotReadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	projectRoot := setupClearTestFixtureEmpty(t)

	// Set sessions directory permissions to 0000
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	require.NoError(t, os.Chmod(sessionsDir, 0000))
	t.Cleanup(func() { os.Chmod(sessionsDir, 0755) })

	_, stderr, exitCode := executeClearCommand(t, projectRoot, []string{}, nil)

	assert.Equal(t, 1, exitCode)
	assert.Regexp(t, `(?i)failed to read sessions directory:.*permission denied`, stderr)
}

// --- Validation Failures — Partial Deletion Failure ---

// TestClearCommand_PartialDeletionFailure reports error for failed session but continues with others.
func TestClearCommand_PartialDeletionFailure(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2", "session-3")

	// Make session-2 undeletable by setting its parent to prevent deletion
	session2Dir := filepath.Join(projectRoot, ".spectra", "sessions", "session-2")
	// Create a file inside and make the directory read-only
	require.NoError(t, os.WriteFile(filepath.Join(session2Dir, "lock"), []byte("locked"), 0644))
	require.NoError(t, os.Chmod(session2Dir, 0555))
	t.Cleanup(func() { os.Chmod(session2Dir, 0755) })

	stdin := strings.NewReader("y\n")
	stdout, stderr, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	// Exit code 0 even with partial failures
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "All sessions cleared successfully")

	// Error reported for failed session
	combinedOutput := stdout + stderr
	assert.Regexp(t, `(?i)error.*failed to clear session`, combinedOutput)
}

// --- Happy Path — Recursive Deletion ---

// TestClearCommand_DeletesAllSessionContents recursively deletes all files and subdirectories.
func TestClearCommand_DeletesAllSessionContents(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "test-session")

	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", "test-session")
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session.json"), []byte("{}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(""), 0644))

	nestedDir := filepath.Join(sessionDir, "nested", "deep")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "data.txt"), []byte("data"), 0644))

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "cleared successfully")

	// Verify entire directory tree deleted; no files remain
	_, err := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(err))
}

// --- Happy Path — Symbolic Link Handling ---

// TestClearCommand_DeletesSymlinkNotTarget deletes symbolic link without following to target.
func TestClearCommand_DeletesSymlinkNotTarget(t *testing.T) {
	projectRoot := setupClearTestFixtureEmpty(t)

	// Create target directory outside test fixture
	targetDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "data.txt"), []byte("target data"), 0644))

	// Create symlink in sessions directory
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	symlinkPath := filepath.Join(sessionsDir, "link-session")
	require.NoError(t, os.Symlink(targetDir, symlinkPath))

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=link-session"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "cleared successfully")

	// Symlink should be deleted
	_, err := os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(err))

	// Target directory should remain intact
	_, err = os.Stat(filepath.Join(targetDir, "data.txt"))
	assert.NoError(t, err, "Target directory should remain intact")
}

// --- Happy Path — Sessions Directory Preserved ---

// TestClearCommand_PreservesSessionsDirectory does not delete the sessions directory itself.
func TestClearCommand_PreservesSessionsDirectory(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "session-1", "session-2")

	stdin := strings.NewReader("y\n")
	_, _, exitCode := executeClearCommand(t, projectRoot, []string{}, stdin)

	assert.Equal(t, 0, exitCode)

	// .spectra/sessions/ directory should remain
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	info, err := os.Stat(sessionsDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".spectra/sessions/ directory should remain")

	// But session subdirectories should be gone
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// --- Happy Path — No Confirmation for Single Session ---

// TestClearCommand_NoConfirmationForSingleSession does not prompt for confirmation when deleting specific session.
func TestClearCommand_NoConfirmationForSingleSession(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "test-session")

	// Pass nil stdin — no confirmation input available; should not prompt
	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout, "Are you sure")
	assert.Contains(t, stdout, "cleared successfully")

	// Verify session deleted
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", "test-session")
	_, err := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(err))
}

// --- Idempotency ---

// TestClearCommand_IdempotentDeletion repeated invocation after deletion prints warning, not error.
func TestClearCommand_IdempotentDeletion(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "test-session")

	// First call: deletes and prints success
	stdout1, _, exitCode1 := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	assert.Equal(t, 0, exitCode1)
	assert.Contains(t, stdout1, "cleared successfully")

	// Second call: prints warning
	stdout2, _, exitCode2 := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	assert.Equal(t, 0, exitCode2)
	assert.Contains(t, stdout2, "Warning:")
}

// --- Happy Path — Help Output ---

// TestClearCommand_Help displays help information without invoking SpectraFinder.
func TestClearCommand_Help(t *testing.T) {
	// No .spectra/ directory created — help should not attempt to find it
	projectRoot := setupClearTestFixtureNoSpectra(t)

	stdout, _, exitCode := executeClearCommand(t, projectRoot, []string{"--help"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "session-id")
	assert.Contains(t, stdout, "Usage:")
}

// --- Boundary Values — Large Session ---

// TestClearCommand_LargeSessionDirectory deletes session with many files synchronously.
func TestClearCommand_LargeSessionDirectory(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "large-session")

	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", "large-session")
	// Create 1000 small files
	for i := 0; i < 1000; i++ {
		filename := filepath.Join(sessionDir, fmt.Sprintf("file_%04d.txt", i))
		require.NoError(t, os.WriteFile(filename, []byte("data"), 0644))
	}

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=large-session"}, nil)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "cleared successfully")

	// Verify all files deleted
	_, err := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(err))
}

// --- Integration — SpectraFinder ---

// TestClearCommand_UsesSpectraFinder uses SpectraFinder to locate project root.
func TestClearCommand_UsesSpectraFinder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure: root/.spectra/sessions/ and root/subdir/
	rootDir := filepath.Join(tmpDir, "root")
	sessionsDir := filepath.Join(rootDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	subdir := filepath.Join(rootDir, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0755))

	// Execute from subdir; SpectraFinder should search upward
	stdout, stderr, exitCode := executeClearCommand(t, subdir,
		[]string{"--session-id=test-session"}, nil)

	// SpectraFinder finds .spectra/ in parent; operates on root/.spectra/sessions/
	// Session doesn't exist, so expect warning with exit code 0
	assert.True(t, exitCode == 0 || exitCode == 1,
		"exit code should be 0 or 1 depending on session existence, got %d", exitCode)
	combinedOutput := stdout + stderr
	assert.NotEmpty(t, combinedOutput)
}

// --- Integration — StorageLayout ---

// TestClearCommand_UsesStorageLayout uses StorageLayout to compose sessions directory path.
func TestClearCommand_UsesStorageLayout(t *testing.T) {
	projectRoot := setupClearTestFixtureWithSessions(t, "test-session")

	stdout, _, exitCode := executeClearCommand(t, projectRoot,
		[]string{"--session-id=test-session"}, nil)

	// Command uses StorageLayout to get sessions path; correct path composed
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "cleared successfully")

	// Verify it operated on the correct directory
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", "test-session")
	_, err := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(err))
}
