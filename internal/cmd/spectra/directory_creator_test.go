package spectra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allDirectories returns the full list of directories that CreateAll should create,
// in the expected creation order.
func allDirectories() []string {
	return []string{
		".spectra",
		".spectra/sessions",
		".spectra/workflows",
		".spectra/agents",
		"spec",
		"spec/logic",
		"spec/test",
	}
}

// createAllDirs pre-creates all expected directories in projectRoot.
func createAllDirs(t *testing.T, projectRoot string) {
	t.Helper()
	for _, d := range allDirectories() {
		ensureDir(t, filepath.Join(projectRoot, d))
	}
}

// --- Happy Path — CreateAll ---

func TestDirectoryCreator_CreateAll_AllNew(t *testing.T) {
	projectRoot := t.TempDir()

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.NoError(t, err)

	for _, d := range allDirectories() {
		dirPath := filepath.Join(projectRoot, d)
		assertDirExists(t, dirPath)
		assertDirPermissions(t, dirPath, 0755)
	}
}

func TestDirectoryCreator_CreateAll_PartialExist(t *testing.T) {
	projectRoot := t.TempDir()

	// Pre-create only .spectra/ and spec/
	ensureDir(t, filepath.Join(projectRoot, ".spectra"))
	ensureDir(t, filepath.Join(projectRoot, "spec"))

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.NoError(t, err)

	// All directories should exist
	for _, d := range allDirectories() {
		dirPath := filepath.Join(projectRoot, d)
		assertDirExists(t, dirPath)
		assertDirPermissions(t, dirPath, 0755)
	}
}

// --- Idempotency ---

func TestDirectoryCreator_CreateAll_AllExist(t *testing.T) {
	projectRoot := t.TempDir()
	createAllDirs(t, projectRoot)

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.NoError(t, err)

	// All directories still exist with correct permissions
	for _, d := range allDirectories() {
		dirPath := filepath.Join(projectRoot, d)
		assertDirExists(t, dirPath)
		assertDirPermissions(t, dirPath, 0755)
	}
}

func TestDirectoryCreator_CreateAll_CalledTwice(t *testing.T) {
	projectRoot := t.TempDir()

	creator := NewDirectoryCreator()

	// First call
	err := creator.CreateAll(projectRoot)
	require.NoError(t, err)

	// Second call — idempotent
	err = creator.CreateAll(projectRoot)
	require.NoError(t, err)

	for _, d := range allDirectories() {
		dirPath := filepath.Join(projectRoot, d)
		assertDirExists(t, dirPath)
		assertDirPermissions(t, dirPath, 0755)
	}
}

// --- Ordering — Directory Creation ---

func TestDirectoryCreator_CreateAll_ParentBeforeChild(t *testing.T) {
	projectRoot := t.TempDir()

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.NoError(t, err)

	// Verify children exist inside parents
	assertDirExists(t, filepath.Join(projectRoot, ".spectra", "sessions"))
	assertDirExists(t, filepath.Join(projectRoot, ".spectra", "workflows"))
	assertDirExists(t, filepath.Join(projectRoot, ".spectra", "agents"))
	assertDirExists(t, filepath.Join(projectRoot, "spec", "logic"))
	assertDirExists(t, filepath.Join(projectRoot, "spec", "test"))
}

// --- Error Propagation ---

func TestDirectoryCreator_CreateAll_PathExistsAsFile(t *testing.T) {
	projectRoot := t.TempDir()

	// Create a regular file at the .spectra path
	writeFile(t, filepath.Join(projectRoot, ".spectra"), "I am a file", 0644)

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory '.spectra'")
}

func TestDirectoryCreator_CreateAll_PermissionDenied(t *testing.T) {
	projectRoot := t.TempDir()

	// Make projectRoot read-only
	require.NoError(t, os.Chmod(projectRoot, 0555))
	t.Cleanup(func() {
		_ = os.Chmod(projectRoot, 0755)
	})

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory '.spectra'")
}

func TestDirectoryCreator_CreateAll_FailFastStopsProcessing(t *testing.T) {
	projectRoot := t.TempDir()

	// Block .spectra directory creation by placing a file there
	writeFile(t, filepath.Join(projectRoot, ".spectra"), "blocker", 0644)

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".spectra")

	// Subsequent directories should NOT exist
	assertPathNotExists(t, filepath.Join(projectRoot, ".spectra", "sessions"))
	assertPathNotExists(t, filepath.Join(projectRoot, ".spectra", "workflows"))
	assertPathNotExists(t, filepath.Join(projectRoot, ".spectra", "agents"))
	assertPathNotExists(t, filepath.Join(projectRoot, "spec"))
	assertPathNotExists(t, filepath.Join(projectRoot, "spec", "logic"))
	assertPathNotExists(t, filepath.Join(projectRoot, "spec", "test"))
}

// --- Boundary Values — projectRoot ---

func TestDirectoryCreator_CreateAll_NestedChildMissing(t *testing.T) {
	projectRoot := t.TempDir()

	// Pre-create only .spectra/ (parent) — children missing
	ensureDir(t, filepath.Join(projectRoot, ".spectra"))

	creator := NewDirectoryCreator()
	err := creator.CreateAll(projectRoot)
	require.NoError(t, err)

	// Children should now exist
	assertDirExists(t, filepath.Join(projectRoot, ".spectra", "sessions"))
	assertDirPermissions(t, filepath.Join(projectRoot, ".spectra", "sessions"), 0755)
	assertDirExists(t, filepath.Join(projectRoot, ".spectra", "workflows"))
	assertDirPermissions(t, filepath.Join(projectRoot, ".spectra", "workflows"), 0755)
	assertDirExists(t, filepath.Join(projectRoot, ".spectra", "agents"))
	assertDirPermissions(t, filepath.Join(projectRoot, ".spectra", "agents"), 0755)
}
