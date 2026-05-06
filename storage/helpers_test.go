package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Test Constants ---

const (
	testProjectRoot  = "/home/user/project"
	testSessionUUID  = "550e8400-e29b-41d4-a716-446655440000"
	testWorkflowName = "CodeReview"
	testAgentRole    = "Architect"
)

// --- Fixture Builders ---

// makeTempDir creates a temporary directory for the test and registers cleanup.
func makeTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// makeTempDirWithSpectra creates a temp dir containing a `.spectra/` subdirectory.
func makeTempDirWithSpectra(t *testing.T) string {
	t.Helper()
	dir := makeTempDir(t)
	spectraDir := filepath.Join(dir, ".spectra")
	if err := os.Mkdir(spectraDir, 0755); err != nil {
		t.Fatalf("makeTempDirWithSpectra: failed to create .spectra: %v", err)
	}
	return dir
}

// makeTempDirWithSessions creates a temp dir containing `.spectra/sessions/`.
func makeTempDirWithSessions(t *testing.T) string {
	t.Helper()
	dir := makeTempDirWithSpectra(t)
	sessionsDir := filepath.Join(dir, ".spectra", "sessions")
	if err := os.Mkdir(sessionsDir, 0755); err != nil {
		t.Fatalf("makeTempDirWithSessions: failed to create sessions: %v", err)
	}
	return dir
}

// makeTempDirWithSessionDir creates a temp dir containing `.spectra/sessions/<uuid>/`.
func makeTempDirWithSessionDir(t *testing.T, uuid string) string {
	t.Helper()
	dir := makeTempDirWithSessions(t)
	sessionDir := filepath.Join(dir, ".spectra", "sessions", uuid)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatalf("makeTempDirWithSessionDir: failed to create session dir: %v", err)
	}
	return dir
}

// makeTempFile creates a temporary file at the given path (parent must exist).
func makeTempFile(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("makeTempFile: failed to create file %s: %v", path, err)
	}
	f.Close()
}

// makeTempDirWithSpectraAsFile creates a temp dir containing `.spectra` as a regular file.
func makeTempDirWithSpectraAsFile(t *testing.T) string {
	t.Helper()
	dir := makeTempDir(t)
	spectraPath := filepath.Join(dir, ".spectra")
	makeTempFile(t, spectraPath)
	return dir
}

// makeTempDirWithReadOnlySpectra creates a temp dir with `.spectra/` set to 0555 (no write).
func makeTempDirWithReadOnlySpectra(t *testing.T) string {
	t.Helper()
	dir := makeTempDirWithSpectra(t)
	spectraDir := filepath.Join(dir, ".spectra")
	if err := os.Chmod(spectraDir, 0555); err != nil {
		t.Fatalf("makeTempDirWithReadOnlySpectra: failed to chmod .spectra: %v", err)
	}
	t.Cleanup(func() {
		// Restore write permission for cleanup
		os.Chmod(spectraDir, 0755)
	})
	return dir
}

// makeTempDirWithReadOnlySessions creates a temp dir with `.spectra/sessions/` set to 0555.
func makeTempDirWithReadOnlySessions(t *testing.T) string {
	t.Helper()
	dir := makeTempDirWithSessions(t)
	sessionsDir := filepath.Join(dir, ".spectra", "sessions")
	if err := os.Chmod(sessionsDir, 0555); err != nil {
		t.Fatalf("makeTempDirWithReadOnlySessions: failed to chmod sessions: %v", err)
	}
	t.Cleanup(func() {
		// Restore write permission for cleanup
		os.Chmod(sessionsDir, 0755)
	})
	return dir
}
