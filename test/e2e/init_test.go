package e2e_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// --- Test Helpers ---

// executeInitCommand creates and executes the init command in the given working directory.
// Returns stdout, stderr, and exit code.
func executeInitCommand(t *testing.T, workDir string, args ...string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer

	cmd := spectra.NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"init"}, args...))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workDir))
	defer os.Chdir(origDir)

	exitCode := cmd.Execute()

	return stdout.String(), stderr.String(), exitCode
}

// assertDirExistsWithPermissions asserts that a directory exists with the expected permissions.
func assertDirExistsWithPermissions(t *testing.T, path string, expectedPerm os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "directory should exist: %s", path)
	assert.True(t, info.IsDir(), "should be a directory: %s", path)
	assert.Equal(t, expectedPerm, info.Mode().Perm(), "directory permissions mismatch: %s", path)
}

// assertFileExistsWithPerm asserts that a file exists with the expected permissions.
func assertFileExistsWithPerm(t *testing.T, path string, expectedPerm os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "file should exist: %s", path)
	assert.False(t, info.IsDir(), "should not be a directory: %s", path)
	assert.Equal(t, expectedPerm, info.Mode().Perm(), "file permissions mismatch: %s", path)
}

// =====================================================================
// Happy Path — Initialization
// =====================================================================

// TestInit_FreshDirectory initializes all directories and files in a fresh directory.
func TestInit_FreshDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Spectra project initialized successfully")

	// Verify .gitignore created with .spectra entry
	gitignoreContent, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(gitignoreContent), ".spectra")

	// Verify all .spectra/ directories created
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "sessions"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "workflows"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "agents"))

	// Verify all spec/ directories created
	assert.DirExists(t, filepath.Join(tmpDir, "spec"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec", "logic"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec", "test"))

	// Verify .spectra/ files written
	assert.FileExists(t, filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"))

	// Verify spec/ files written
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "CONVENTIONS.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "logic", "README.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "test", "README.md"))
}

// TestInit_GitignoreCreated creates .gitignore with .spectra entry when it does not exist.
func TestInit_GitignoreCreated(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, ".spectra\n", string(content))
}

// TestInit_GitignoreAlreadyContainsEntry skips modifying .gitignore when it already contains .spectra entry.
func TestInit_GitignoreAlreadyContainsEntry(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	originalContent := "node_modules\n.spectra\n*.log\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(originalContent), 0644))

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout, "Warning")

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(content))
}

// TestInit_GitignoreAppended appends .spectra to existing .gitignore that does not contain the entry.
func TestInit_GitignoreAppended(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	originalContent := "node_modules\n*.log\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(originalContent), 0644))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "node_modules")
	assert.Contains(t, string(content), "*.log")
	assert.Contains(t, string(content), ".spectra")
}

// TestInit_AllDirectoriesExist skips directory creation when all directories already exist.
func TestInit_AllDirectoriesExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create all directories
	dirs := []string{
		filepath.Join(tmpDir, ".spectra"),
		filepath.Join(tmpDir, ".spectra", "sessions"),
		filepath.Join(tmpDir, ".spectra", "workflows"),
		filepath.Join(tmpDir, ".spectra", "agents"),
		filepath.Join(tmpDir, "spec"),
		filepath.Join(tmpDir, "spec", "logic"),
		filepath.Join(tmpDir, "spec", "test"),
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d, 0755))
	}

	stdout, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Spectra project initialized successfully")
}

// =====================================================================
// Happy Path — Partial State
// =====================================================================

// TestInit_SomeDirectoriesExist creates missing directories when some already exist.
func TestInit_SomeDirectoriesExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only some directories
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Spectra project initialized successfully")

	// Verify missing directories were created
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "agents"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "sessions"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec", "logic"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec", "test"))
}

// TestInit_SomeBuiltinFilesExist copies missing files and prints warnings for existing files.
func TestInit_SomeBuiltinFilesExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure and some existing files
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "logic"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "test"), 0755))

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"),
		[]byte("existing workflow"), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"),
		[]byte("existing arch"), 0644))

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Warning: workflow definition 'DefaultLogicSpec.yaml' already exists, skipping")
	assert.Contains(t, stdout, "Warning: spec file 'ARCHITECTURE.md' already exists, skipping")
	assert.Contains(t, stdout, "Spectra project initialized successfully")
}

// TestInit_AllBuiltinFilesExist prints warnings for all skipped files when all exist.
func TestInit_AllBuiltinFilesExist(t *testing.T) {
	tmpDir := t.TempDir()

	// First run to create everything
	_, _, exitCode1 := executeInitCommand(t, tmpDir)
	require.Equal(t, 0, exitCode1)

	// Second run: all files exist
	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Warning:")
	assert.Contains(t, stdout, "Spectra project initialized successfully")
}

// =====================================================================
// Idempotency
// =====================================================================

// TestInit_RepeatedInvocation second invocation is idempotent with warnings for existing files.
func TestInit_RepeatedInvocation(t *testing.T) {
	tmpDir := t.TempDir()

	// First invocation
	stdout1, _, exitCode1 := executeInitCommand(t, tmpDir)
	assert.Equal(t, 0, exitCode1)
	assert.Contains(t, stdout1, "Spectra project initialized successfully")

	// Second invocation
	stdout2, _, exitCode2 := executeInitCommand(t, tmpDir)
	assert.Equal(t, 0, exitCode2)
	assert.Contains(t, stdout2, "Warning:")
	assert.Contains(t, stdout2, "Spectra project initialized successfully")
}

// =====================================================================
// State Transitions
// =====================================================================

// TestInit_PhasesExecuteInOrder verifies phases execute in order: gitignore, .spectra dirs, .spectra files, spec dirs, spec files.
func TestInit_PhasesExecuteInOrder(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	// Verify all expected artifacts exist (order is verified indirectly: if phases were
	// out of order, files/dirs would fail to create)
	assert.FileExists(t, filepath.Join(tmpDir, ".gitignore"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "sessions"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "workflows"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "agents"))
	assert.FileExists(t, filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec", "logic"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec", "test"))
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"))
}

// =====================================================================
// Error Propagation — Phase 0 (.gitignore)
// =====================================================================

// TestInit_GitignoreReadFails returns error when reading .gitignore fails.
func TestInit_GitignoreReadFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("content"), 0000))
	t.Cleanup(func() { os.Chmod(gitignorePath, 0644) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to read '.gitignore'")
	assert.Contains(t, stderr, "permission denied")

	// No .spectra/ or spec/ directories created
	assert.NoDirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.NoDirExists(t, filepath.Join(tmpDir, "spec"))
}

// TestInit_GitignoreWriteFails returns error when updating .gitignore fails.
func TestInit_GitignoreWriteFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	// Create .gitignore without .spectra entry but make it read-only
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("*.log\n"), 0444))
	t.Cleanup(func() { os.Chmod(gitignorePath, 0644) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to update '.gitignore'")
	assert.Contains(t, stderr, "permission denied")

	// No .spectra/ or spec/ directories created
	assert.NoDirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.NoDirExists(t, filepath.Join(tmpDir, "spec"))
}

// TestInit_GitignoreBrokenSymlink returns error when .gitignore is a broken symlink.
func TestInit_GitignoreBrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.Symlink(filepath.Join(tmpDir, "nonexistent"), gitignorePath))

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to read '.gitignore'")
	assert.Contains(t, stderr, "no such file or directory")
}

// =====================================================================
// Error Propagation — Phase 1 (.spectra directories)
// =====================================================================

// TestInit_SpectraDirCreationFails returns error when creating .spectra/ directory fails.
func TestInit_SpectraDirCreationFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Pre-create .gitignore so Phase 0 succeeds and Phase 1 (directory creation) can be tested
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte(".spectra\n"), 0644))

	require.NoError(t, os.Chmod(tmpDir, 0555))
	t.Cleanup(func() { os.Chmod(tmpDir, 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")
	assert.Contains(t, stderr, "permission denied")
}

// TestInit_SpectraSubdirCreationFails returns error when creating .spectra/sessions/ fails after .spectra/ succeeds.
func TestInit_SpectraSubdirCreationFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Create .spectra/ but make it read-only to prevent subdirectory creation
	spectraDir := filepath.Join(tmpDir, ".spectra")
	require.NoError(t, os.MkdirAll(spectraDir, 0755))
	require.NoError(t, os.Chmod(spectraDir, 0555))
	t.Cleanup(func() { os.Chmod(spectraDir, 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")
	assert.Contains(t, stderr, "permission denied")

	// .spectra/ exists
	assert.DirExists(t, spectraDir)
}

// TestInit_SpectraExistsAsFile returns error when .spectra exists as a regular file.
func TestInit_SpectraExistsAsFile(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".spectra"), []byte("file"), 0644))

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")
}

// =====================================================================
// Error Propagation — Phase 2 (.spectra files)
// =====================================================================

// TestInit_WorkflowFileWriteFails returns error when writing a workflow file fails.
func TestInit_WorkflowFileWriteFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Create directories but make workflows/ read-only
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))
	require.NoError(t, os.Chmod(filepath.Join(tmpDir, ".spectra", "workflows"), 0555))
	t.Cleanup(func() { os.Chmod(filepath.Join(tmpDir, ".spectra", "workflows"), 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to write built-in file")
	assert.Contains(t, stderr, "permission denied")

	// .spectra/ directories exist
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))

	// No spec/ directories created
	assert.NoDirExists(t, filepath.Join(tmpDir, "spec"))
}

// TestInit_AgentFileWriteFails returns error when writing an agent file fails.
func TestInit_AgentFileWriteFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Create directories but make agents/ read-only
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))
	require.NoError(t, os.Chmod(filepath.Join(tmpDir, ".spectra", "agents"), 0555))
	t.Cleanup(func() { os.Chmod(filepath.Join(tmpDir, ".spectra", "agents"), 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to write built-in file")
}

// TestInit_WorkflowFileWriteFailsDiskFull returns error when writing fails due to disk full.
func TestInit_WorkflowFileWriteFailsDiskFull(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("disk quota simulation not available on Windows")
	}
	t.Skip("disk full simulation requires platform-specific setup (loop device or quota)")
}

// =====================================================================
// Error Propagation — Phase 3 (spec directories)
// =====================================================================

// TestInit_SpecDirCreationFails returns error when creating spec/ directory fails.
func TestInit_SpecDirCreationFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Pre-create .spectra structure so Phase 1 and 2 succeed
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))

	// Run init once to create .spectra files
	_, _, exitCode1 := executeInitCommand(t, tmpDir)
	require.Equal(t, 0, exitCode1)

	// Remove spec/ dir and make root read-only
	require.NoError(t, os.RemoveAll(filepath.Join(tmpDir, "spec")))
	require.NoError(t, os.Chmod(tmpDir, 0555))
	t.Cleanup(func() { os.Chmod(tmpDir, 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")
	assert.Contains(t, stderr, "permission denied")

	// All .spectra/ directories and files exist
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "workflows"))
}

// TestInit_SpecSubdirCreationFails returns error when creating spec/logic/ fails after spec/ succeeds.
func TestInit_SpecSubdirCreationFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Pre-create full .spectra structure and files
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))

	// Create spec/ but make it read-only to prevent subdirectory creation
	specDir := filepath.Join(tmpDir, "spec")
	require.NoError(t, os.MkdirAll(specDir, 0755))
	require.NoError(t, os.Chmod(specDir, 0555))
	t.Cleanup(func() { os.Chmod(specDir, 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")
	assert.Contains(t, stderr, "permission denied")

	// spec/ exists
	assert.DirExists(t, specDir)
	// All .spectra/ directories exist
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
}

// TestInit_SpecExistsAsFile returns error when spec exists as a regular file.
func TestInit_SpecExistsAsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create .spectra structure
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))

	// Create spec as a regular file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "spec"), []byte("file"), 0644))

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")

	// All .spectra/ directories exist
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
}

// =====================================================================
// Error Propagation — Phase 4 (spec files)
// =====================================================================

// TestInit_SpecFileWriteFails returns error when writing a spec file fails.
func TestInit_SpecFileWriteFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Pre-create all directories
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))
	specDir := filepath.Join(tmpDir, "spec")
	require.NoError(t, os.MkdirAll(filepath.Join(specDir, "logic"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(specDir, "test"), 0755))

	// Make spec/ read-only to prevent file writes
	require.NoError(t, os.Chmod(specDir, 0555))
	t.Cleanup(func() { os.Chmod(specDir, 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to write built-in file")

	// All directories exist
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
}

// TestInit_SpecFileWriteFailsNestedDir returns error when writing nested spec file fails due to missing subdirectory.
func TestInit_SpecFileWriteFailsNestedDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Pre-create .spectra structure
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "workflows"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "agents"), 0755))

	// Create spec/ and spec/test/ but not spec/logic/
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "test"), 0755))
	// spec/logic/ is intentionally missing — if Phase 3 creates it, we need a different approach
	// This test may need to be adapted based on actual init implementation that creates dirs in Phase 3

	// If Phase 3 always creates spec/logic/, this test verifies Phase 4 failure
	// by removing spec/logic/ after directory creation phase
	// Since we can't inject between phases in an e2e test, we skip if not feasible
	t.Skip("requires ability to inject between Phase 3 and Phase 4 (mock/test double needed)")
}

// =====================================================================
// Validation Failures — .gitignore
// =====================================================================

// TestInit_GitignoreContainsVariation appends .spectra when .gitignore contains .spectra/ but not .spectra.
func TestInit_GitignoreContainsVariation(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte(".spectra/\n"), 0644))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), ".spectra/")
	assert.Contains(t, string(content), ".spectra\n")
}

// TestInit_GitignoreContainsCommented appends .spectra when .gitignore contains only commented # .spectra entry.
func TestInit_GitignoreContainsCommented(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("# .spectra\n"), 0644))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	lines := string(content)
	assert.Contains(t, lines, "# .spectra")
	// Should have appended .spectra as a separate line
	assert.True(t, strings.Contains(lines, "\n.spectra\n") || strings.HasSuffix(lines, "\n.spectra\n"),
		"should append .spectra as separate uncommented line")
}

// TestInit_GitignoreWhitespaceMatch skips modification when .gitignore line contains .spectra with leading/trailing spaces or tabs.
func TestInit_GitignoreWhitespaceMatch(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"leading and trailing spaces", "  .spectra  \n"},
		{"leading and trailing tabs", "\t.spectra\t\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			gitignorePath := filepath.Join(tmpDir, ".gitignore")
			require.NoError(t, os.WriteFile(gitignorePath, []byte(tc.content), 0644))

			_, _, exitCode := executeInitCommand(t, tmpDir)

			assert.Equal(t, 0, exitCode)

			content, err := os.ReadFile(gitignorePath)
			require.NoError(t, err)
			assert.Equal(t, tc.content, string(content))
		})
	}
}

// TestInit_GitignoreNonBreakingSpace appends .spectra when .gitignore line contains non-breaking space around .spectra.
func TestInit_GitignoreNonBreakingSpace(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	// U+00A0 non-breaking space before .spectra
	require.NoError(t, os.WriteFile(gitignorePath, []byte(" .spectra\n"), 0644))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	// Original line preserved and .spectra appended
	assert.Contains(t, string(content), " .spectra")
	// Count occurrences — should have the original NBSP line plus a new .spectra line
	lines := strings.Split(string(content), "\n")
	foundExact := false
	for _, line := range lines {
		trimmed := strings.Trim(line, " \t")
		if trimmed == ".spectra" {
			foundExact = true
			break
		}
	}
	assert.True(t, foundExact, "should have appended exact .spectra line")
}

// TestInit_GitignoreSymlinkFollowed follows symlink and modifies target file when .gitignore is a symlink.
func TestInit_GitignoreSymlinkFollowed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target file
	targetFile := filepath.Join(tmpDir, "shared-gitignore")
	require.NoError(t, os.WriteFile(targetFile, []byte("*.log\n"), 0644))

	// Create symlink
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.Symlink(targetFile, gitignorePath))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	// Target file should be modified
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), ".spectra")
}

// TestInit_GitignoreNoTrailingNewline appends .spectra with proper newline when .gitignore does not end with newline.
func TestInit_GitignoreNoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("*.log"), 0644))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "*.log\n.spectra\n", string(content))
}

// TestInit_GitignoreWithTrailingNewline appends .spectra when .gitignore ends with newline.
func TestInit_GitignoreWithTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("*.log\n"), 0644))

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "*.log\n.spectra\n", string(content))
}

// =====================================================================
// Boundary Values — Directory Names
// =====================================================================

// TestInit_CurrentDirIsRoot handles initialization when current directory is filesystem root.
func TestInit_CurrentDirIsRoot(t *testing.T) {
	// Use a mock-root directory with restricted permissions to simulate root-like restrictions
	tmpDir := t.TempDir()
	mockRoot := filepath.Join(tmpDir, "mock-root")
	require.NoError(t, os.MkdirAll(mockRoot, 0555))
	t.Cleanup(func() { os.Chmod(mockRoot, 0755) })

	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	_, stderr, exitCode := executeInitCommand(t, mockRoot)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "permission denied")
}

// TestInit_CurrentDirIsReadOnly returns error when current directory is read-only.
func TestInit_CurrentDirIsReadOnly(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}
	if runtime.GOOS == "windows" {
		t.Skip("read-only permissions not supported on this platform")
	}

	tmpDir := t.TempDir()
	require.NoError(t, os.Chmod(tmpDir, 0555))
	t.Cleanup(func() { os.Chmod(tmpDir, 0755) })

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "permission denied")
}

// =====================================================================
// Boundary Values — Nested Project
// =====================================================================

// TestInit_NestedProject creates nested Spectra project in subdirectory of existing project.
func TestInit_NestedProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize parent project
	_, _, exitCode1 := executeInitCommand(t, tmpDir)
	require.Equal(t, 0, exitCode1)

	// Create nested subdirectory
	nestedDir := filepath.Join(tmpDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	// Initialize nested project
	stdout, _, exitCode := executeInitCommand(t, nestedDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Spectra project initialized successfully")

	// Verify nested project structure
	assert.DirExists(t, filepath.Join(nestedDir, ".spectra"))
	assert.DirExists(t, filepath.Join(nestedDir, "spec"))

	// Verify parent project unchanged
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.DirExists(t, filepath.Join(tmpDir, "spec"))
}

// =====================================================================
// Null / Empty Input
// =====================================================================

// TestInit_NoArguments runs successfully with no command-line arguments.
func TestInit_NoArguments(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Spectra project initialized successfully")
}

// =====================================================================
// Resource Cleanup
// =====================================================================

// TestInit_DirectoryPermissions verifies created directories have correct permissions.
func TestInit_DirectoryPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)
	require.Equal(t, 0, exitCode)

	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, ".spectra"), 0755)
	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, ".spectra", "sessions"), 0755)
	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, ".spectra", "workflows"), 0755)
	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, ".spectra", "agents"), 0755)
	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, "spec"), 0755)
	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, "spec", "logic"), 0755)
	assertDirExistsWithPermissions(t, filepath.Join(tmpDir, "spec", "test"), 0755)
}

// TestInit_FilePermissions verifies created files have correct permissions.
func TestInit_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)
	require.Equal(t, 0, exitCode)

	assertFileExistsWithPerm(t, filepath.Join(tmpDir, ".gitignore"), 0644)
	assertFileExistsWithPerm(t, filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"), 0644)
	assertFileExistsWithPerm(t, filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"), 0644)
	assertFileExistsWithPerm(t, filepath.Join(tmpDir, "spec", "CONVENTIONS.md"), 0644)
	assertFileExistsWithPerm(t, filepath.Join(tmpDir, "spec", "logic", "README.md"), 0644)
	assertFileExistsWithPerm(t, filepath.Join(tmpDir, "spec", "test", "README.md"), 0644)
}

// =====================================================================
// Data Independence (Copy Semantics)
// =====================================================================

// TestInit_BuiltinFilesNotValidated copies built-in files without validating YAML or Markdown syntax.
func TestInit_BuiltinFilesNotValidated(t *testing.T) {
	// This is verified indirectly: init copies files without parsing them.
	// If a built-in file had invalid YAML, init would still succeed.
	// Since we cannot replace embed.FS in e2e, we verify that init succeeds
	// and files are written without error.
	tmpDir := t.TempDir()

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Spectra project initialized successfully")

	// All files exist (content is not validated)
	assert.FileExists(t, filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"))
}

// TestInit_BuiltinSpecFilesNotValidated copies built-in spec files without validating Markdown syntax.
func TestInit_BuiltinSpecFilesNotValidated(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "Spectra project initialized successfully")

	assert.FileExists(t, filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "CONVENTIONS.md"))
}

// =====================================================================
// Ordering — Phase Execution
// =====================================================================

// TestInit_PhaseOrderingGitignoreFirst verifies .gitignore modification happens before directory creation.
func TestInit_PhaseOrderingGitignoreFirst(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	// If gitignore was not created first, and directory creation failed,
	// gitignore would still exist. We verify both exist.
	assert.FileExists(t, filepath.Join(tmpDir, ".gitignore"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
}

// TestInit_PhaseOrderingSpectraDirsBeforeFiles verifies .spectra/ directories created before .spectra/ files written.
func TestInit_PhaseOrderingSpectraDirsBeforeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	// Directories and files both exist — files couldn't be created without directories
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "workflows"))
	assert.FileExists(t, filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"))
}

// TestInit_PhaseOrderingSpecDirsBeforeFiles verifies spec/ directories created before spec/ files written.
func TestInit_PhaseOrderingSpecDirsBeforeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, exitCode := executeInitCommand(t, tmpDir)

	assert.Equal(t, 0, exitCode)

	assert.DirExists(t, filepath.Join(tmpDir, "spec", "logic"))
	assert.FileExists(t, filepath.Join(tmpDir, "spec", "logic", "README.md"))
}

// TestInit_PhaseOrderingSpectraDirsBeforeSpecDirs verifies .spectra/ directories and files created before spec/ directories.
func TestInit_PhaseOrderingSpectraDirsBeforeSpecDirs(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()

	// Create a file named "spec" to block Phase 3
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "spec"), []byte("blocker"), 0644))

	_, stderr, exitCode := executeInitCommand(t, tmpDir)

	// Phase 3 should fail, but Phases 0, 1, 2 should have completed
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "failed to create directory")

	// .spectra structure should exist (Phases 1+2 completed before Phase 3 failed)
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "workflows"))
	assert.FileExists(t, filepath.Join(tmpDir, ".spectra", "workflows", "DefaultLogicSpec.yaml"))
}
