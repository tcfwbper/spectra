package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — EnsureSessionsDirectory ---

func TestEnsureSessionsDirectory_CreatesWhenMissing(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSpectra(t)

	err := EnsureSessionsDirectory(dir)

	require.NoError(t, err)
	sessionsDir := filepath.Join(dir, ".spectra", "sessions")
	info, statErr := os.Stat(sessionsDir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestEnsureSessionsDirectory_AlreadyExists(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSessions(t)

	err := EnsureSessionsDirectory(dir)

	require.NoError(t, err)
}

// --- Idempotency ---

func TestEnsureSessionsDirectory_Idempotent(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSpectra(t)

	err1 := EnsureSessionsDirectory(dir)
	err2 := EnsureSessionsDirectory(dir)

	require.NoError(t, err1)
	require.NoError(t, err2)

	sessionsDir := filepath.Join(dir, ".spectra", "sessions")
	info, statErr := os.Stat(sessionsDir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}

// --- Error Propagation ---

func TestEnsureSessionsDirectory_SpectraDirMissing(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDir(t) // no .spectra/ directory

	err := EnsureSessionsDirectory(dir)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}

func TestEnsureSessionsDirectory_SpectraDirIsFile(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSpectraAsFile(t)

	err := EnsureSessionsDirectory(dir)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}

func TestEnsureSessionsDirectory_MkdirPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithReadOnlySpectra(t)

	err := EnsureSessionsDirectory(dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create sessions directory")
	assert.Contains(t, err.Error(), "permission denied")
}

// --- Happy Path — CreateSessionDirectory ---

func TestCreateSessionDirectory_Success(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSessions(t)
	uuid := "550e8400-e29b-41d4-a716-446655440000"

	err := CreateSessionDirectory(dir, uuid)

	require.NoError(t, err)
	sessionDir := filepath.Join(dir, ".spectra", "sessions", uuid)
	info, statErr := os.Stat(sessionDir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestCreateSessionDirectory_EnsuresParent(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSpectra(t) // has .spectra/ but not sessions/
	uuid := "550e8400-e29b-41d4-a716-446655440000"

	err := CreateSessionDirectory(dir, uuid)

	require.NoError(t, err)
	sessionsDir := filepath.Join(dir, ".spectra", "sessions")
	sessionDir := filepath.Join(sessionsDir, uuid)
	_, err1 := os.Stat(sessionsDir)
	_, err2 := os.Stat(sessionDir)
	assert.NoError(t, err1, ".spectra/sessions/ should exist")
	assert.NoError(t, err2, "session directory should exist")
}

// --- Validation Failures ---

func TestCreateSessionDirectory_AlreadyExists(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	uuid := "550e8400-e29b-41d4-a716-446655440000"
	dir := makeTempDirWithSessionDir(t, uuid)

	err := CreateSessionDirectory(dir, uuid)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionDirExists)
}

func TestCreateSessionDirectory_SpectraDirMissing(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDir(t) // no .spectra/

	err := CreateSessionDirectory(dir, "some-uuid")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}

func TestCreateSessionDirectory_MkdirFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithReadOnlySessions(t)

	err := CreateSessionDirectory(dir, "new-session-uuid")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create session directory:")
}

// --- Null / Empty Input ---

func TestCreateSessionDirectory_EmptyUUID(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSessions(t)

	// Should not panic; may return nil or os.PathError
	assert.NotPanics(t, func() {
		_ = CreateSessionDirectory(dir, "")
	})
}

// --- Boundary Values — sessionUUID ---

func TestCreateSessionDirectory_UUIDWithPathSeparator(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/session_directory_manager.go")

	dir := makeTempDirWithSessions(t)

	// Should not panic; path traversal is not prevented by SessionDirectoryManager
	assert.NotPanics(t, func() {
		err := CreateSessionDirectory(dir, "../escape")
		// No validation error expected; may succeed or fail at OS level
		_ = err
	})
}
