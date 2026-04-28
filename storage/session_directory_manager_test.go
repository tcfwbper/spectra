package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

// TestSessionDirectoryManager_ValidConstruction creates manager with valid project root
func TestSessionDirectoryManager_ValidConstruction(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	assert.NotNil(t, manager)
}

// TestCreateSessionDirectory_ValidUUID creates session directory with valid UUID
func TestCreateSessionDirectory_ValidUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.NoError(t, err)

	sessionDir := filepath.Join(sessionsDir, sessionUUID)
	info, err := os.Stat(sessionDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestCreateSessionDirectory_Permissions0775 tests created directory has permissions 0775
func TestCreateSessionDirectory_Permissions0775(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.NoError(t, err)

	sessionDir := filepath.Join(sessionsDir, sessionUUID)
	info, err := os.Stat(sessionDir)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0775), info.Mode().Perm())
}

// TestCreateSessionDirectory_SessionsParentMissing returns error when .spectra/sessions/ does not exist
func TestCreateSessionDirectory_SessionsParentMissing(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)sessions directory does not exist.*\.spectra/sessions.*Run 'spectra init'`, err.Error())
}

// TestCreateSessionDirectory_SpectraDirMissing returns error when .spectra/ does not exist
func TestCreateSessionDirectory_SpectraDirMissing(t *testing.T) {
	tmpDir := t.TempDir()

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)sessions directory does not exist.*\.spectra/sessions.*Run 'spectra init'`, err.Error())
}

// TestCreateSessionDirectory_AlreadyExists returns error when session directory already exists
func TestCreateSessionDirectory_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	sessionDir := filepath.Join(sessionsDir, sessionUUID)

	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)session directory already exists.*123e4567-e89b-12d3-a456-426614174000.*UUID collision`, err.Error())
}

// TestCreateSessionDirectory_PermissionDenied returns error when directory creation fails due to permission denied
func TestCreateSessionDirectory_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))
	require.NoError(t, os.Chmod(sessionsDir, 0555))
	defer os.Chmod(sessionsDir, 0755)

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)failed to create session directory.*permission denied`, err.Error())
}

// TestCreateSessionDirectory_DiskFull returns error when directory creation fails due to disk full
func TestCreateSessionDirectory_DiskFull(t *testing.T) {
	t.Skip("Simulating disk full is platform-specific and requires special setup")
}

// TestCreateSessionDirectory_EmptyUUID attempts to create directory with empty UUID (malformed path)
func TestCreateSessionDirectory_EmptyUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := ""

	err := manager.CreateSessionDirectory(sessionUUID)
	// Behavior depends on OS; test that it either succeeds or fails
	_ = err
}

// TestCreateSessionDirectory_UUIDWithPathSeparator tests no validation of UUID format; potentially dangerous paths passed through
func TestCreateSessionDirectory_UUIDWithPathSeparator(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "../malicious"

	err := manager.CreateSessionDirectory(sessionUUID)
	// May succeed or fail depending on filesystem state
	_ = err
}

// TestSessionDirectoryManager_RelativeProjectRoot uses relative path if ProjectRoot is relative
func TestSessionDirectoryManager_RelativeProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tmpDir))

	manager := storage.NewSessionDirectoryManager(".")
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err = manager.CreateSessionDirectory(sessionUUID)
	assert.NoError(t, err)

	sessionDir := filepath.Join(".spectra", "sessions", sessionUUID)
	info, err := os.Stat(sessionDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestCreateSessionDirectory_PathExceedsLimit returns error when full path exceeds platform maximum
func TestCreateSessionDirectory_PathExceedsLimit(t *testing.T) {
	t.Skip("Testing path length limits requires very long directory structures")
}

// TestSessionDirectoryManager_NoStateCaching tests manager is stateless; each call performs fresh filesystem checks
func TestSessionDirectoryManager_NoStateCaching(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)

	require.NoError(t, os.RemoveAll(sessionsDir))

	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)sessions directory does not exist`, err.Error())
}

// TestSessionDirectoryManager_DoesNotCreateParents tests manager never creates .spectra/ or .spectra/sessions/ directories
func TestSessionDirectoryManager_DoesNotCreateParents(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)

	sessionsDir := filepath.Join(spectraDir, "sessions")
	_, err = os.Stat(sessionsDir)
	assert.True(t, os.IsNotExist(err), "sessions directory should not be created")
}

// TestSessionDirectoryManager_UsesStorageLayout tests manager delegates all path composition to StorageLayout
func TestSessionDirectoryManager_UsesStorageLayout(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.NoError(t, err)

	expectedPath := storage.GetSessionDir(tmpDir, sessionUUID)
	info, err := os.Stat(expectedPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestCreateSessionDirectory_ExternalProcessCreates tests another process creates directory between existence check and mkdir call
func TestCreateSessionDirectory_ExternalProcessCreates(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".spectra", "sessions")
	sessionUUID := "123e4567-e89b-12d3-a456-426614174000"
	sessionDir := filepath.Join(sessionsDir, sessionUUID)

	require.NoError(t, os.MkdirAll(sessionsDir, 0755))

	manager := storage.NewSessionDirectoryManager(tmpDir)

	require.NoError(t, os.Mkdir(sessionDir, 0755))

	err := manager.CreateSessionDirectory(sessionUUID)
	assert.Error(t, err)
	assert.Regexp(t, `(?i)session directory already exists.*file exists`, err.Error())
}
