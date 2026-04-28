package storage_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcfwbper/spectra/storage"
)

// TestSpectraFinder_FindInCurrentDir finds .spectra in the current directory immediately
func TestSpectraFinder_FindInCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	require.NoError(t, os.Chdir(tmpDir))

	result, err := storage.SpectraFinder("")
	assert.NoError(t, err)
	absDir, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absDir, result)
}

// TestSpectraFinder_FindWithExplicitStartDir finds .spectra when StartDir explicitly provided
func TestSpectraFinder_FindWithExplicitStartDir(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	result, err := storage.SpectraFinder(tmpDir)
	assert.NoError(t, err)
	absDir, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absDir, result)
}

// TestSpectraFinder_FindInParent searches upward and finds .spectra in parent directory
func TestSpectraFinder_FindInParent(t *testing.T) {
	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	spectraDir := filepath.Join(parentDir, ".spectra")

	require.NoError(t, os.MkdirAll(childDir, 0755))
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	result, err := storage.SpectraFinder(childDir)
	assert.NoError(t, err)
	absParent, err := filepath.Abs(parentDir)
	require.NoError(t, err)
	assert.Equal(t, absParent, result)
}

// TestSpectraFinder_FindMultipleLevelsUp searches upward through multiple directories
func TestSpectraFinder_FindMultipleLevelsUp(t *testing.T) {
	tmpDir := t.TempDir()
	rootDir := filepath.Join(tmpDir, "root")
	deepDir := filepath.Join(rootDir, "a", "b", "c")
	spectraDir := filepath.Join(rootDir, ".spectra")

	require.NoError(t, os.MkdirAll(deepDir, 0755))
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	result, err := storage.SpectraFinder(deepDir)
	assert.NoError(t, err)
	absRoot, err := filepath.Abs(rootDir)
	require.NoError(t, err)
	assert.Equal(t, absRoot, result)
}

// TestSpectraFinder_NearestSpectraWins returns nearest .spectra when multiple exist in hierarchy
func TestSpectraFinder_NearestSpectraWins(t *testing.T) {
	tmpDir := t.TempDir()
	rootDir := filepath.Join(tmpDir, "root")
	projectDir := filepath.Join(rootDir, "project")
	subdirDir := filepath.Join(projectDir, "subdir")

	require.NoError(t, os.MkdirAll(subdirDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(rootDir, ".spectra"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".spectra"), 0755))

	result, err := storage.SpectraFinder(subdirDir)
	assert.NoError(t, err)
	absProject, err := filepath.Abs(projectDir)
	require.NoError(t, err)
	assert.Equal(t, absProject, result)
}

// TestSpectraFinder_FollowsSymlinks follows symbolic links during upward traversal
func TestSpectraFinder_FollowsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows")
	}

	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real")
	subdirDir := filepath.Join(realDir, "subdir")
	linkDir := filepath.Join(tmpDir, "link")
	spectraDir := filepath.Join(realDir, ".spectra")

	require.NoError(t, os.MkdirAll(subdirDir, 0755))
	require.NoError(t, os.Mkdir(spectraDir, 0755))
	require.NoError(t, os.Symlink(subdirDir, linkDir))

	result, err := storage.SpectraFinder(linkDir)
	assert.NoError(t, err)
	absReal, err := filepath.Abs(realDir)
	require.NoError(t, err)
	assert.Equal(t, absReal, result)
}

// TestSpectraFinder_NotFoundReachesRoot returns error when .spectra not found after reaching filesystem root
func TestSpectraFinder_NotFoundReachesRoot(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := storage.SpectraFinder(tmpDir)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)spectra not initialized`, err.Error())
}

// TestSpectraFinder_SpectraIsFile continues searching upward if .spectra is a file, not a directory
func TestSpectraFinder_SpectraIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	spectraFile := filepath.Join(parentDir, ".spectra")

	require.NoError(t, os.MkdirAll(childDir, 0755))
	require.NoError(t, os.WriteFile(spectraFile, []byte("not a directory"), 0644))

	result, err := storage.SpectraFinder(childDir)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)spectra not initialized`, err.Error())
}

// TestSpectraFinder_StartDirNotExist returns error when StartDir does not exist
func TestSpectraFinder_StartDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "nonexistent", "directory")

	result, err := storage.SpectraFinder(nonExistentDir)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)invalid start directory.*nonexistent`, err.Error())
}

// TestSpectraFinder_StartDirIsFile returns error when StartDir is a file, not a directory
func TestSpectraFinder_StartDirIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))

	result, err := storage.SpectraFinder(testFile)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)invalid start directory.*file\.txt`, err.Error())
}

// TestSpectraFinder_PermissionDeniedOnParent returns error when parent directory is not readable
func TestSpectraFinder_PermissionDeniedOnParent(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "parent")
	childDir := filepath.Join(parentDir, "child")

	require.NoError(t, os.MkdirAll(childDir, 0755))
	require.NoError(t, os.Chmod(parentDir, 0000))
	defer func() { _ = os.Chmod(parentDir, 0755) }()

	result, err := storage.SpectraFinder(childDir)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)spectra not initialized`, err.Error())
}

// TestSpectraFinder_SymlinkLoop returns error when encountering symbolic link loop during traversal
func TestSpectraFinder_SymlinkLoop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows")
	}

	tmpDir := t.TempDir()
	linkA := filepath.Join(tmpDir, "a")
	linkB := filepath.Join(tmpDir, "b")

	require.NoError(t, os.Symlink(linkB, linkA))
	require.NoError(t, os.Symlink(linkA, linkB))

	result, err := storage.SpectraFinder(linkA)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)spectra not initialized`, err.Error())
}

// TestSpectraFinder_StartAtFilesystemRoot returns error when starting at filesystem root with no .spectra
func TestSpectraFinder_StartAtFilesystemRoot(t *testing.T) {
	result, err := storage.SpectraFinder("/")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Regexp(t, `(?i)spectra not initialized`, err.Error())
}

// TestSpectraFinder_RelativeStartDir resolves relative StartDir to absolute path before searching
func TestSpectraFinder_RelativeStartDir(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	require.NoError(t, os.Chdir(tmpDir))

	result, err := storage.SpectraFinder(".")
	assert.NoError(t, err)
	absDir, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absDir, result)
}

// TestSpectraFinder_RelativeStartDirWithParent resolves relative path with parent references
func TestSpectraFinder_RelativeStartDirWithParent(t *testing.T) {
	tmpDir := t.TempDir()
	rootDir := filepath.Join(tmpDir, "root")
	deepDir := filepath.Join(rootDir, "a", "b")
	spectraDir := filepath.Join(rootDir, ".spectra")

	require.NoError(t, os.MkdirAll(deepDir, 0755))
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	require.NoError(t, os.Chdir(deepDir))

	result, err := storage.SpectraFinder("..")
	assert.NoError(t, err)
	absRoot, err := filepath.Abs(rootDir)
	require.NoError(t, err)
	assert.Equal(t, absRoot, result)
}

// TestSpectraFinder_RepeatedSearch tests multiple searches from same directory return identical results
func TestSpectraFinder_RepeatedSearch(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	result1, err1 := storage.SpectraFinder(tmpDir)
	assert.NoError(t, err1)

	result2, err2 := storage.SpectraFinder(tmpDir)
	assert.NoError(t, err2)

	result3, err3 := storage.SpectraFinder(tmpDir)
	assert.NoError(t, err3)

	assert.Equal(t, result1, result2)
	assert.Equal(t, result2, result3)
}

// TestSpectraFinder_ReturnsAbsolutePath tests returned ProjectRoot is always an absolute path
func TestSpectraFinder_ReturnsAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.Mkdir(spectraDir, 0755))

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	require.NoError(t, os.Chdir(tmpDir))

	result, err := storage.SpectraFinder(".")
	assert.NoError(t, err)
	assert.True(t, filepath.IsAbs(result) || filepath.VolumeName(result) != "")
}

// TestSpectraFinder_NoCaching tests finder performs fresh search on each invocation, no caching
func TestSpectraFinder_NoCaching(t *testing.T) {
	tmpDir := t.TempDir()
	spectraDir := filepath.Join(tmpDir, ".spectra")

	result1, err1 := storage.SpectraFinder(tmpDir)
	assert.Error(t, err1)
	assert.Empty(t, result1)

	require.NoError(t, os.Mkdir(spectraDir, 0755))

	result2, err2 := storage.SpectraFinder(tmpDir)
	assert.NoError(t, err2)
	absDir, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absDir, result2)
}
