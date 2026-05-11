package spectra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gitignorePath returns the .gitignore path for a given projectRoot.
func gitignorePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".gitignore")
}

// --- Happy Path — Ensure ---

func TestGitignoreEnsurer_Ensure_CreatesNewFile(t *testing.T) {

	projectRoot := t.TempDir()

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileExists(t, gitignorePath(projectRoot))
	assertFileContent(t, gitignorePath(projectRoot), ".spectra\n")
	assertFilePermissions(t, gitignorePath(projectRoot), 0644)
}

func TestGitignoreEnsurer_Ensure_AppendsWhenMissing_EndsWithNewline(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "node_modules\n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), "node_modules\n.spectra\n")
}

func TestGitignoreEnsurer_Ensure_AppendsWhenMissing_NoTrailingNewline(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "node_modules", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), "node_modules\n.spectra\n")
}

func TestGitignoreEnsurer_Ensure_AlreadyPresent(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "node_modules\n.spectra\n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), "node_modules\n.spectra\n")
}

// --- Idempotency ---

func TestGitignoreEnsurer_Ensure_Idempotent(t *testing.T) {

	projectRoot := t.TempDir()

	ensurer := NewGitignoreEnsurer()

	// First call — creates the file
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	// Second call — should not duplicate
	err = ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	content := readFileContent(t, gitignorePath(projectRoot))
	// Count occurrences of ".spectra" in lines
	lines := splitLines(content)
	count := 0
	for _, line := range lines {
		if trimLine(line) == ".spectra" {
			count++
		}
	}
	assert.Equal(t, 1, count, "expected exactly one .spectra entry, got %d", count)
}

// --- Boundary Values — Line Matching ---

func TestGitignoreEnsurer_Ensure_MatchesWithLeadingTrailingSpaces(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "  .spectra  \n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	// File should not be modified
	assertFileContent(t, gitignorePath(projectRoot), "  .spectra  \n")
}

func TestGitignoreEnsurer_Ensure_MatchesWithTabs(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "\t.spectra\t\n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	// File should not be modified
	assertFileContent(t, gitignorePath(projectRoot), "\t.spectra\t\n")
}

func TestGitignoreEnsurer_Ensure_DoesNotMatchSpectraSlash(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), ".spectra/\n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), ".spectra/\n.spectra\n")
}

func TestGitignoreEnsurer_Ensure_DoesNotMatchCommented(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "# .spectra\n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), "# .spectra\n.spectra\n")
}

func TestGitignoreEnsurer_Ensure_DoesNotTrimUnicodeWhitespace(t *testing.T) {

	projectRoot := t.TempDir()
	// NBSP (U+00A0) prefix — should NOT be trimmed, so line does not match
	writeFile(t, gitignorePath(projectRoot), " .spectra\n", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), " .spectra\n.spectra\n")
}

// --- Null / Empty Input ---

func TestGitignoreEnsurer_Ensure_EmptyFile(t *testing.T) {

	projectRoot := t.TempDir()
	writeFile(t, gitignorePath(projectRoot), "", 0644)

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	assertFileContent(t, gitignorePath(projectRoot), ".spectra\n")
}

// --- Error Propagation ---

func TestGitignoreEnsurer_Ensure_ReadPermissionDenied(t *testing.T) {

	projectRoot := t.TempDir()
	path := gitignorePath(projectRoot)
	writeFile(t, path, "content\n", 0000)
	t.Cleanup(func() {
		_ = os.Chmod(path, 0644)
	})

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read '.gitignore'")
}

func TestGitignoreEnsurer_Ensure_WritePermissionDenied(t *testing.T) {

	projectRoot := t.TempDir()
	path := gitignorePath(projectRoot)
	writeFile(t, path, "node_modules\n", 0444)
	t.Cleanup(func() {
		_ = os.Chmod(path, 0644)
	})

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update '.gitignore'")
}

func TestGitignoreEnsurer_Ensure_BrokenSymlink(t *testing.T) {

	projectRoot := t.TempDir()
	path := gitignorePath(projectRoot)
	// Create a symlink pointing to a non-existent target
	require.NoError(t, os.Symlink(filepath.Join(projectRoot, "nonexistent_target"), path))

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read '.gitignore'")
}

// --- Mock / Dependency Interaction ---

func TestGitignoreEnsurer_Ensure_FollowsSymlink(t *testing.T) {

	projectRoot := t.TempDir()
	realFilePath := filepath.Join(projectRoot, "real_gitignore")
	writeFile(t, realFilePath, "node_modules\n", 0644)

	// Create symlink .gitignore -> real_gitignore
	path := gitignorePath(projectRoot)
	require.NoError(t, os.Symlink(realFilePath, path))

	ensurer := NewGitignoreEnsurer()
	err := ensurer.Ensure(projectRoot)
	require.NoError(t, err)

	// Verify the symlink target was modified
	assertFileContent(t, realFilePath, "node_modules\n.spectra\n")
}

// --- Utility helpers for gitignore tests ---

// splitLines splits content into lines (including empty trailing element if ends with \n).
func splitLines(content string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

// trimLine trims spaces and tabs from a line (matching GitignoreEnsurer logic).
func trimLine(line string) string {
	// Trim only ASCII space and tab
	start := 0
	for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
		start++
	}
	end := len(line)
	for end > start && (line[end-1] == ' ' || line[end-1] == '\t') {
		end--
	}
	return line[start:end]
}
