package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — SpectraFinder ---

func TestSpectraFinder_FoundInStartDir(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDirWithSpectra(t)

	result, err := FindSpectraRoot(dir)

	require.NoError(t, err)
	assert.Equal(t, dir, result)
}

func TestSpectraFinder_FoundInParent(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDirWithSpectra(t)
	subDir := filepath.Join(dir, "sub", "deep")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	result, err := FindSpectraRoot(subDir)

	require.NoError(t, err)
	assert.Equal(t, dir, result)
}

func TestSpectraFinder_FoundNearestAncestor(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDir(t)
	// Create .spectra at root level
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".spectra"), 0755))
	// Create inner with its own .spectra
	innerDir := filepath.Join(dir, "inner")
	require.NoError(t, os.Mkdir(innerDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(innerDir, ".spectra"), 0755))
	// Create child under inner
	childDir := filepath.Join(innerDir, "child")
	require.NoError(t, os.Mkdir(childDir, 0755))

	result, err := FindSpectraRoot(childDir)

	require.NoError(t, err)
	assert.Equal(t, innerDir, result)
}

func TestSpectraFinder_EmptyStartDirUsesCwd(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go with Getwd seam")

	dir := makeTempDirWithSpectra(t)

	// Change working directory to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	result, findErr := FindSpectraRoot("")

	require.NoError(t, findErr)
	assert.Equal(t, dir, result)
}

func TestSpectraFinder_RelativeStartDir(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDirWithSpectra(t)

	// Change to parent of dir so relative path works
	parentDir := filepath.Dir(dir)
	baseName := filepath.Base(dir)
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(parentDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	result, findErr := FindSpectraRoot(baseName)

	require.NoError(t, findErr)
	assert.Equal(t, dir, result)
}

// --- Error Propagation ---

func TestSpectraFinder_NotFoundReturnsErrNotInitialized(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDir(t)
	deepDir := filepath.Join(dir, "a", "b", "c")
	require.NoError(t, os.MkdirAll(deepDir, 0755))

	_, err := FindSpectraRoot(deepDir)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}

func TestSpectraFinder_PermissionDeniedReturnsErrNotInitialized(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDir(t)
	restricted := filepath.Join(dir, "restricted")
	require.NoError(t, os.Mkdir(restricted, 0755))
	child := filepath.Join(restricted, "child")
	require.NoError(t, os.Mkdir(child, 0755))
	// Restrict parent so traversal upward fails
	require.NoError(t, os.Chmod(restricted, 0000))
	t.Cleanup(func() { os.Chmod(restricted, 0755) })

	_, err := FindSpectraRoot(child)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}

// --- Validation Failures ---

func TestSpectraFinder_NonExistentStartDir(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	_, err := FindSpectraRoot("/tmp/non-existent-path-abc123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start directory")
	assert.Contains(t, err.Error(), "/tmp/non-existent-path-abc123")
}

func TestSpectraFinder_StartDirIsFile(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDir(t)
	filePath := filepath.Join(dir, "afile.txt")
	makeTempFile(t, filePath)

	_, err := FindSpectraRoot(filePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start directory")
	assert.Contains(t, err.Error(), filePath)
}

func TestSpectraFinder_GetwdFails(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go with Getwd seam/test hook")

	// This test requires either:
	// 1. A test seam to inject a failing Getwd, or
	// 2. Changing to a directory that is then removed (OS-dependent behavior)
	// Left as scaffolded until the production surface reveals the seam approach.

	// Placeholder: If a Getwd seam exists, inject failure and verify error message.
	// _, err := FindSpectraRoot("")
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to get working directory")
}

// --- Boundary Values — startDir ---

func TestSpectraFinder_StartDirIsRoot(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	_, err := FindSpectraRoot("/")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}

func TestSpectraFinder_SpectraExistsAsFile(t *testing.T) {
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDir(t)
	// Create parent with .spectra as directory
	parentDir := filepath.Join(dir, "parent")
	require.NoError(t, os.Mkdir(parentDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(parentDir, ".spectra"), 0755))
	// Create child with .spectra as file
	childDir := filepath.Join(parentDir, "child")
	require.NoError(t, os.Mkdir(childDir, 0755))
	makeTempFile(t, filepath.Join(childDir, ".spectra"))

	result, err := FindSpectraRoot(childDir)

	require.NoError(t, err)
	assert.Equal(t, parentDir, result)
}

func TestSpectraFinder_SymlinkLoop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink test not reliable on Windows")
	}
	t.Skip("scaffolded: awaiting production source storage/spectra_finder.go")

	dir := makeTempDir(t)
	aDir := filepath.Join(dir, "a")
	require.NoError(t, os.Mkdir(aDir, 0755))
	// Create symlink loop: a/link -> a
	linkPath := filepath.Join(aDir, "link")
	require.NoError(t, os.Symlink(aDir, linkPath))

	_, err := FindSpectraRoot(linkPath)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInitialized)
}
