package spectra

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Clear Specific Sessions ---

func TestClear_SingleUUID_Exists(t *testing.T) {
	projectRoot := t.TempDir()
	uuid := "test-uuid-1234"
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", uuid)
	ensureDir(t, sessionDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Session directory removed; stdout contains "Session '<UUID>' cleared"
	assertPathNotExists(t, sessionDir)
	assert.Contains(t, stdout.String(), "Session '"+uuid+"' cleared")
}

func TestClear_MultipleUUIDs_AllExist(t *testing.T) {
	projectRoot := t.TempDir()
	uuid1 := "uuid-aaa"
	uuid2 := "uuid-bbb"
	sessionDir1 := filepath.Join(projectRoot, ".spectra", "sessions", uuid1)
	sessionDir2 := filepath.Join(projectRoot, ".spectra", "sessions", uuid2)
	ensureDir(t, sessionDir1)
	ensureDir(t, sessionDir2)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Both directories removed; stdout contains both messages
	assertPathNotExists(t, sessionDir1)
	assertPathNotExists(t, sessionDir2)
	assert.Contains(t, stdout.String(), "Session '"+uuid1+"' cleared")
	assert.Contains(t, stdout.String(), "Session '"+uuid2+"' cleared")
}

func TestClear_ConfirmationPrompt_ListsUUIDs(t *testing.T) {
	projectRoot := t.TempDir()

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("n\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Prompt output contains UUID listing
	output := stdout.String()
	assert.Contains(t, output, "Are you sure you want to delete the following sessions?")
	assert.Contains(t, output, "- abc-123")
	assert.Contains(t, output, "- def-456")
}

// --- Happy Path — Clear All Sessions ---

func TestClear_NoArgs_DeletesAll(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, filepath.Join(sessionsDir, "sess1"))
	ensureDir(t, filepath.Join(sessionsDir, "sess2"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Both directories removed; stdout contains messages
	assertPathNotExists(t, filepath.Join(sessionsDir, "sess1"))
	assertPathNotExists(t, filepath.Join(sessionsDir, "sess2"))
	assert.Contains(t, stdout.String(), "Session 'sess1' cleared")
	assert.Contains(t, stdout.String(), "Session 'sess2' cleared")
	assert.Contains(t, stdout.String(), "All sessions cleared successfully")
}

func TestClear_NoArgs_SkipsFiles(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, filepath.Join(sessionsDir, "sess1"))
	writeFile(t, filepath.Join(sessionsDir, "somefile.txt"), "data", 0644)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: sess1/ removed; somefile.txt still exists
	assertPathNotExists(t, filepath.Join(sessionsDir, "sess1"))
	assertFileExists(t, filepath.Join(sessionsDir, "somefile.txt"))
	assert.Contains(t, stdout.String(), "Session 'sess1' cleared")
}

func TestClear_NoArgs_PromptText(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, filepath.Join(sessionsDir, "sess1"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("n\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Prompt contains specific text
	assert.Contains(t, stdout.String(), "Are you sure you want to delete all sessions? [y/N]: ")
}

// --- Happy Path — User Cancels ---

func TestClear_SpecificUUIDs_UserDeclinesN(t *testing.T) {
	projectRoot := t.TempDir()
	uuid := "test-uuid"
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", uuid)
	ensureDir(t, sessionDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("n\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Session directory still exists; stdout contains "Operation cancelled"
	assertDirExists(t, sessionDir)
	assert.Contains(t, stdout.String(), "Operation cancelled")
}

func TestClear_AllSessions_UserDeclinesN(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, filepath.Join(sessionsDir, "sess1"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("n\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Session directory still exists; stdout contains "Operation cancelled"
	assertDirExists(t, filepath.Join(sessionsDir, "sess1"))
	assert.Contains(t, stdout.String(), "Operation cancelled")
}

func TestClear_UserEntersYes_TreatedAsRejection(t *testing.T) {
	projectRoot := t.TempDir()
	uuid := "test-uuid"
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", uuid)
	ensureDir(t, sessionDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("yes\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: "yes" treated as rejection, directory still exists
	assertDirExists(t, sessionDir)
	assert.Contains(t, stdout.String(), "Operation cancelled")
}

// --- Error Propagation ---

func TestClear_SpectraFinderFails(t *testing.T) {
	finder := &fakeSpectraFinderForClear{err: errFakeFinderFailure}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: stderr contains error message; exit code 1
	assert.Contains(t, stderr.String(), "Error: .spectra directory not found. Are you in a Spectra project?")
}

func TestClear_NoArgs_SessionsDirNotExist(t *testing.T) {
	projectRoot := t.TempDir()
	// Do NOT create .spectra/sessions/

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: stdout contains warning; exit code 0
	assert.Contains(t, stdout.String(), "Warning: sessions directory not found, nothing to clear")
	assert.Empty(t, stderr.String())
}

func TestClear_NoArgs_ReadDirFails(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, sessionsDir)
	require.NoError(t, os.Chmod(sessionsDir, 0000))
	t.Cleanup(func() {
		_ = os.Chmod(sessionsDir, 0755)
	})

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: stderr contains error; exit code 1
	assert.Contains(t, stderr.String(), "Error: failed to read sessions directory:")
}

func TestClear_SpecificUUID_DeletionFails(t *testing.T) {
	projectRoot := t.TempDir()
	uuid1 := "undeletable-uuid"
	uuid2 := "deletable-uuid"
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	sessionDir1 := filepath.Join(sessionsDir, uuid1)
	sessionDir2 := filepath.Join(sessionsDir, uuid2)
	ensureDir(t, sessionDir1)
	ensureDir(t, sessionDir2)

	// Place a child inside sessionDir1 and make sessionDir1 non-writable so
	// os.RemoveAll cannot remove its contents, causing deletion to fail for
	// uuid1 only. The parent (sessionsDir) retains write permission so uuid2
	// can still be deleted.
	writeFile(t, filepath.Join(sessionDir1, "lock"), "x", 0644)
	require.NoError(t, os.Chmod(sessionDir1, 0555))
	t.Cleanup(func() {
		_ = os.Chmod(sessionDir1, 0755)
	})

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: error for uuid1, success for uuid2
	assert.Contains(t, stderr.String(), "Error: failed to clear session '"+uuid1+"':")
	assert.Contains(t, stdout.String(), "Session '"+uuid2+"' cleared")
}

func TestClear_NoArgs_PartialDeletionFailure(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	sess1Dir := filepath.Join(sessionsDir, "sess1")
	ensureDir(t, sess1Dir)
	ensureDir(t, filepath.Join(sessionsDir, "sess2"))

	// Place a child inside sess1 and remove write permission on sess1 so
	// os.RemoveAll cannot remove its contents, while the parent (sessionsDir)
	// retains write permission allowing sess2 to be deleted.
	writeFile(t, filepath.Join(sess1Dir, "lock"), "x", 0644)
	require.NoError(t, os.Chmod(sess1Dir, 0555))
	t.Cleanup(func() {
		_ = os.Chmod(sess1Dir, 0755)
	})

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: error for sess1, success for sess2, no summary
	assert.Contains(t, stderr.String(), "Error: failed to clear session 'sess1':")
	assert.Contains(t, stdout.String(), "Session 'sess2' cleared")
	assert.NotContains(t, stdout.String(), "All sessions cleared successfully")
}

// --- Null / Empty Input ---

func TestClear_NoArgs_EmptySessionsDir(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, sessionsDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: stdout contains "No sessions to clear"; no confirmation prompt
	assert.Contains(t, stdout.String(), "No sessions to clear")
	assert.NotContains(t, stdout.String(), "[y/N]")
}

func TestClear_NoArgs_OnlyFilesInSessionsDir(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, sessionsDir)
	writeFile(t, filepath.Join(sessionsDir, "somefile.txt"), "data", 0644)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: stdout contains "No sessions to clear"
	assert.Contains(t, stdout.String(), "No sessions to clear")
}

func TestClear_SpecificUUID_NotFound(t *testing.T) {
	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "sessions"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: stdout contains warning about missing UUID
	assert.Contains(t, stdout.String(), "Warning: session 'nonexistent-uuid' not found, skipping")
}

func TestClear_EOF_Stdin(t *testing.T) {
	projectRoot := t.TempDir()
	uuid := "test-uuid"
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", uuid)
	ensureDir(t, sessionDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("") // EOF
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Session directory still exists; stdout contains "Operation cancelled"
	assertDirExists(t, sessionDir)
	assert.Contains(t, stdout.String(), "Operation cancelled")
}

// --- Boundary Values — UUIDs ---

func TestClear_EmptyStringUUID(t *testing.T) {
	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "sessions"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: Warning printed (stat likely fails); exit code 0
	// Empty string UUID is passed to StorageLayout without validation
}

func TestClear_MixedExistentAndNonExistent(t *testing.T) {
	projectRoot := t.TempDir()
	existsDir := filepath.Join(projectRoot, ".spectra", "sessions", "exists-uuid")
	ensureDir(t, existsDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: existing deleted, missing warned
	assertPathNotExists(t, existsDir)
	assert.Contains(t, stdout.String(), "Session 'exists-uuid' cleared")
	assert.Contains(t, stdout.String(), "Warning: session 'missing-uuid' not found, skipping")
}

// --- Mock / Dependency Interaction ---

func TestClear_CallsSpectraFinder(t *testing.T) {
	projectRoot := t.TempDir()

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: SpectraFinder.Find() called exactly once
	assert.Equal(t, 1, finder.findCallCount)
}

func TestClear_CallsStorageLayoutGetSessionDir(t *testing.T) {
	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "sessions", "uuid1"))
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "sessions", "uuid2"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: GetSessionDir called for each UUID
	require.Len(t, layout.getSessionDirCalls, 2)
	assert.Equal(t, clearSessionDirCall{projectRoot: projectRoot, uuid: "uuid1"}, layout.getSessionDirCalls[0])
	assert.Equal(t, clearSessionDirCall{projectRoot: projectRoot, uuid: "uuid2"}, layout.getSessionDirCalls[1])
}

func TestClear_CallsStorageLayoutGetSessionsDir(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".spectra", "sessions")
	ensureDir(t, filepath.Join(sessionsDir, "sess1"))

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: GetSessionsDir called exactly once
	assert.Equal(t, 1, layout.getSessionsDirCallCount)
}

func TestClear_CallsConfirmPrompt(t *testing.T) {
	projectRoot := t.TempDir()
	uuid := "test-uuid"
	sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", uuid)
	ensureDir(t, sessionDir)

	finder := &fakeSpectraFinderForClear{projectRoot: projectRoot}
	layout := &fakeClearStorageLayout{projectRoot: projectRoot}
	stdin := strings.NewReader("n\n") // Mock ConfirmPrompt returns false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_ = finder
	_ = layout
	_ = stdin
	_ = stdout
	_ = stderr

	t.Skip("scaffolded: production symbol NewClearCommand (clear.go) does not exist yet")

	// Expected: ConfirmPrompt called once; no directories deleted
	assertDirExists(t, sessionDir)
}
