package builtin

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — Workflows ---

func TestWorkflows_ContainsYAMLFiles(t *testing.T) {
	matches, err := fs.Glob(Workflows, "workflows/*.yaml")

	require.NoError(t, err)
	assert.NotEmpty(t, matches, "expected at least one .yaml file in workflows/")
}

func TestWorkflows_FilesAreReadable(t *testing.T) {
	matches, err := fs.Glob(Workflows, "workflows/*.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	for _, path := range matches {
		content, err := fs.ReadFile(Workflows, path)
		assert.NoError(t, err, "failed to read %s", path)
		assert.Greater(t, len(content), 0, "file %s is empty", path)
	}
}

func TestWorkflows_PreservesDirectoryStructure(t *testing.T) {
	entries, err := fs.ReadDir(Workflows, "workflows")

	require.NoError(t, err)
	assert.NotEmpty(t, entries, "expected non-empty directory entries under workflows/")
}

// --- Happy Path — Agents ---

func TestAgents_ContainsYAMLFiles(t *testing.T) {
	matches, err := fs.Glob(Agents, "agents/*.yaml")

	require.NoError(t, err)
	assert.NotEmpty(t, matches, "expected at least one .yaml file in agents/")
}

func TestAgents_FilesAreReadable(t *testing.T) {
	matches, err := fs.Glob(Agents, "agents/*.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	for _, path := range matches {
		content, err := fs.ReadFile(Agents, path)
		assert.NoError(t, err, "failed to read %s", path)
		assert.Greater(t, len(content), 0, "file %s is empty", path)
	}
}

func TestAgents_PreservesDirectoryStructure(t *testing.T) {
	entries, err := fs.ReadDir(Agents, "agents")

	require.NoError(t, err)
	assert.NotEmpty(t, entries, "expected non-empty directory entries under agents/")
}

// --- Happy Path — SpecFiles ---

func TestSpecFiles_ContainsFiles(t *testing.T) {
	entries, err := fs.ReadDir(SpecFiles, "spec")

	require.NoError(t, err)
	assert.NotEmpty(t, entries, "expected non-empty directory entries under spec/")
}

func TestSpecFiles_FilesAreReadable(t *testing.T) {
	err := fs.WalkDir(SpecFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, readErr := fs.ReadFile(SpecFiles, path)
		assert.NoError(t, readErr, "failed to read %s", path)
		assert.Greater(t, len(content), 0, "file %s is empty", path)
		return nil
	})
	require.NoError(t, err)
}

func TestSpecFiles_PreservesNestedDirectoryStructure(t *testing.T) {
	var paths []string
	err := fs.WalkDir(SpecFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	require.NoError(t, err)

	// At least one file path should have a nested subdirectory after "spec/"
	// e.g., "spec/logic/README.md" has more than one separator after "spec/"
	hasNested := false
	for _, p := range paths {
		after, found := strings.CutPrefix(p, "spec/")
		if found && strings.Contains(after, "/") {
			hasNested = true
			break
		}
	}
	assert.True(t, hasNested, "expected at least one file with nested subdirectory structure under spec/; got paths: %v", paths)
}

// --- Null / Empty Input ---

func TestWorkflows_NonYAMLFilesExcluded(t *testing.T) {
	matches, err := fs.Glob(Workflows, "workflows/*")
	require.NoError(t, err)

	for _, path := range matches {
		assert.True(t, strings.HasSuffix(path, ".yaml"),
			"expected all files to end in .yaml, but found: %s", path)
	}
}

func TestAgents_NonYAMLFilesExcluded(t *testing.T) {
	matches, err := fs.Glob(Agents, "agents/*")
	require.NoError(t, err)

	for _, path := range matches {
		assert.True(t, strings.HasSuffix(path, ".yaml"),
			"expected all files to end in .yaml, but found: %s", path)
	}
}

// --- Immutability ---

// writerFS is an interface used only to test that embed.FS does NOT implement write operations.
type writerFS interface {
	WriteFile(name string, data []byte) error
}

func TestWorkflows_ReadOnly(t *testing.T) {
	// embed.FS should not implement any write interface.
	var v any = Workflows
	_, ok := v.(writerFS)
	assert.False(t, ok, "embed.FS should not implement write interface")
}
