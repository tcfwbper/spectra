package spectra_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
	"github.com/tcfwbper/spectra/storage"
)

// =====================================================================
// Happy Path — CopyWorkflows
// =====================================================================

// TestCopyWorkflows_AllNew copies all embedded workflow files when none exist at target paths.
func TestCopyWorkflows_AllNew(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
		"builtin/workflows/Another.yaml":   &fstest.MapFile{Data: []byte("name: Another\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	// Verify files written
	path1 := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	path2 := storage.GetWorkflowPath(tmpDir, "Another")
	assertFileExistsWithPermissions(t, path1, 0644)
	assertFileExistsWithPermissions(t, path2, 0644)
}

// TestCopyWorkflows_EmptyEmbedFS returns success with no files when embedded filesystem is empty.
func TestCopyWorkflows_EmptyEmbedFS(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows": &fstest.MapFile{Mode: os.ModeDir},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)
}

// =====================================================================
// Happy Path — CopyAgents
// =====================================================================

// TestCopyAgents_AllNew copies all embedded agent files when none exist at target paths.
func TestCopyAgents_AllNew(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
		"builtin/agents/QaAnalyst.yaml": &fstest.MapFile{Data: []byte("role: QaAnalyst\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	warnings, err := copier.CopyAgents(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	path1 := storage.GetAgentPath(tmpDir, "Architect")
	path2 := storage.GetAgentPath(tmpDir, "QaAnalyst")
	assertFileExistsWithPermissions(t, path1, 0644)
	assertFileExistsWithPermissions(t, path2, 0644)
}

// TestCopyAgents_EmptyEmbedFS returns success with no files when embedded filesystem is empty.
func TestCopyAgents_EmptyEmbedFS(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/agents": &fstest.MapFile{Mode: os.ModeDir},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	warnings, err := copier.CopyAgents(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)
}

// =====================================================================
// Happy Path — CopySpecFiles
// =====================================================================

// TestCopySpecFiles_AllNew copies all embedded spec files when none exist at target paths.
func TestCopySpecFiles_AllNew(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "logic"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "test"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("# Architecture\n")},
		"builtin/spec/CONVENTIONS.md":  &fstest.MapFile{Data: []byte("# Conventions\n")},
		"builtin/spec/logic/README.md": &fstest.MapFile{Data: []byte("# Logic\n")},
		"builtin/spec/test/README.md":  &fstest.MapFile{Data: []byte("# Test\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	assertFileExistsWithPermissions(t, filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"), 0644)
	assertFileExistsWithPermissions(t, filepath.Join(tmpDir, "spec", "CONVENTIONS.md"), 0644)
	assertFileExistsWithPermissions(t, filepath.Join(tmpDir, "spec", "logic", "README.md"), 0644)
	assertFileExistsWithPermissions(t, filepath.Join(tmpDir, "spec", "test", "README.md"), 0644)
}

// TestCopySpecFiles_EmptyEmbedFS returns success with no files when embedded filesystem is empty.
func TestCopySpecFiles_EmptyEmbedFS(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec": &fstest.MapFile{Mode: os.ModeDir},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)
}

// TestCopySpecFiles_PreservesDirectoryStructure preserves nested directory structure from embedded FS to target.
func TestCopySpecFiles_PreservesDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "logic"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "test"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("# Arch\n")},
		"builtin/spec/logic/README.md": &fstest.MapFile{Data: []byte("# Logic Spec\n")},
		"builtin/spec/test/README.md":  &fstest.MapFile{Data: []byte("# Test Spec\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	_, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)

	content1, err := os.ReadFile(filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Arch\n", string(content1))

	content2, err := os.ReadFile(filepath.Join(tmpDir, "spec", "logic", "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Logic Spec\n", string(content2))

	content3, err := os.ReadFile(filepath.Join(tmpDir, "spec", "test", "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Test Spec\n", string(content3))
}

// =====================================================================
// Idempotency
// =====================================================================

// TestCopyWorkflows_MultipleInvocations second invocation skips all files and returns warnings.
func TestCopyWorkflows_MultipleInvocations(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)

	// First invocation
	warnings1, err1 := copier.CopyWorkflows(tmpDir)
	require.NoError(t, err1)
	assert.Empty(t, warnings1)

	// Read content after first copy
	path := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	originalContent, err := os.ReadFile(path)
	require.NoError(t, err)

	// Second invocation
	warnings2, err2 := copier.CopyWorkflows(tmpDir)
	assert.NoError(t, err2)
	assert.Len(t, warnings2, 1)
	assert.Equal(t, "Warning: workflow definition 'SimpleSdd.yaml' already exists, skipping", warnings2[0])

	// File content unchanged
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, originalContent, content)
}

// TestCopyAgents_MultipleInvocations second invocation skips all files and returns warnings.
func TestCopyAgents_MultipleInvocations(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)

	// First invocation
	warnings1, err1 := copier.CopyAgents(tmpDir)
	require.NoError(t, err1)
	assert.Empty(t, warnings1)

	path := storage.GetAgentPath(tmpDir, "Architect")
	originalContent, err := os.ReadFile(path)
	require.NoError(t, err)

	// Second invocation
	warnings2, err2 := copier.CopyAgents(tmpDir)
	assert.NoError(t, err2)
	assert.Len(t, warnings2, 1)
	assert.Equal(t, "Warning: agent definition 'Architect.yaml' already exists, skipping", warnings2[0])

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, originalContent, content)
}

// TestCopySpecFiles_MultipleInvocations second invocation skips all files and returns warnings.
func TestCopySpecFiles_MultipleInvocations(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("# Arch\n")},
		"builtin/spec/CONVENTIONS.md":  &fstest.MapFile{Data: []byte("# Conv\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)

	// First invocation
	warnings1, err1 := copier.CopySpecFiles(tmpDir)
	require.NoError(t, err1)
	assert.Empty(t, warnings1)

	// Second invocation
	warnings2, err2 := copier.CopySpecFiles(tmpDir)
	assert.NoError(t, err2)
	assert.Len(t, warnings2, 2)

	// File content unchanged
	content, err := os.ReadFile(filepath.Join(tmpDir, "spec", "ARCHITECTURE.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Arch\n", string(content))
}

// =====================================================================
// Error Propagation
// =====================================================================

// TestCopyWorkflows_WriteFailsFirstFile returns error when writing the first workflow file fails.
func TestCopyWorkflows_WriteFailsFirstFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))
	require.NoError(t, os.Chmod(workflowsDir, 0555))
	t.Cleanup(func() { os.Chmod(workflowsDir, 0755) })

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.Error(t, err)
	assert.Empty(t, warnings)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Contains(t, err.Error(), "permission denied")
}

// TestCopyWorkflows_WriteFailsSecondFile returns error when writing the second workflow file fails after first succeeds.
func TestCopyWorkflows_WriteFailsSecondFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/First.yaml":  &fstest.MapFile{Data: []byte("name: First\n")},
		"builtin/workflows/Second.yaml": &fstest.MapFile{Data: []byte("name: Second\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)

	// Write first file manually, then make directory read-only
	firstPath := storage.GetWorkflowPath(tmpDir, "First")
	require.NoError(t, os.WriteFile(firstPath, []byte("name: First\n"), 0644))

	// Make directory read-only so writing second file fails
	require.NoError(t, os.Chmod(workflowsDir, 0555))
	t.Cleanup(func() { os.Chmod(workflowsDir, 0755) })

	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Len(t, warnings, 1, "should have one warning for skipped first file")
	// First file remains on disk
	_, statErr := os.Stat(firstPath)
	assert.NoError(t, statErr)
}

// TestCopyWorkflows_WriteFailsAfterSkip returns collected warnings and error when write fails after skipping existing files.
func TestCopyWorkflows_WriteFailsAfterSkip(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	// Create existing file
	firstPath := storage.GetWorkflowPath(tmpDir, "First")
	require.NoError(t, os.WriteFile(firstPath, []byte("existing"), 0644))

	embedFS := fstest.MapFS{
		"builtin/workflows/First.yaml":  &fstest.MapFile{Data: []byte("name: First\n")},
		"builtin/workflows/Second.yaml": &fstest.MapFile{Data: []byte("name: Second\n")},
	}

	// Make directory read-only so writing Second fails
	require.NoError(t, os.Chmod(workflowsDir, 0555))
	t.Cleanup(func() { os.Chmod(workflowsDir, 0755) })

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.Error(t, err)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "First.yaml")
}

// TestCopyAgents_WriteFailsPermissionDenied returns error when writing agent file fails due to permission denied.
func TestCopyAgents_WriteFailsPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.Chmod(agentsDir, 0555))
	t.Cleanup(func() { os.Chmod(agentsDir, 0755) })

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	warnings, err := copier.CopyAgents(tmpDir)

	assert.Error(t, err)
	assert.Empty(t, warnings)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Contains(t, err.Error(), "permission denied")
}

// TestCopySpecFiles_WriteFailsPermissionDenied returns error when writing spec file fails due to permission denied.
func TestCopySpecFiles_WriteFailsPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "spec")
	require.NoError(t, os.MkdirAll(specDir, 0755))
	require.NoError(t, os.Chmod(specDir, 0555))
	t.Cleanup(func() { os.Chmod(specDir, 0755) })

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("# Arch\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.Error(t, err)
	assert.Empty(t, warnings)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Contains(t, err.Error(), "spec/ARCHITECTURE.md")
}

// TestCopySpecFiles_TargetDirMissing returns error when target subdirectory does not exist for nested spec file.
func TestCopySpecFiles_TargetDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	// Create spec/ but not spec/logic/
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/logic/README.md": &fstest.MapFile{Data: []byte("# Logic\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	_, err := copier.CopySpecFiles(tmpDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Contains(t, err.Error(), "no such file or directory")
}

// =====================================================================
// Validation Failures
// =====================================================================

// TestCopyWorkflows_TargetDirMissing returns error when .spectra/workflows/ directory does not exist.
func TestCopyWorkflows_TargetDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	// Do NOT create .spectra/workflows/

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Contains(t, err.Error(), "no such file or directory")
}

// TestCopyAgents_TargetDirMissing returns error when .spectra/agents/ directory does not exist.
func TestCopyAgents_TargetDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	// Do NOT create .spectra/agents/

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	_, err := copier.CopyAgents(tmpDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	assert.Contains(t, err.Error(), "no such file or directory")
}

// =====================================================================
// Boundary Values — File Names
// =====================================================================

// TestCopyWorkflows_MultipleDotsInFilename extracts workflow name correctly from filename with multiple dots.
func TestCopyWorkflows_MultipleDotsInFilename(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/My.Workflow.v2.yaml": &fstest.MapFile{Data: []byte("name: test\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)

	targetPath := storage.GetWorkflowPath(tmpDir, "My.Workflow.v2")
	_, statErr := os.Stat(targetPath)
	assert.NoError(t, statErr)
}

// TestCopyWorkflows_MixedCaseFilename preserves case in workflow filename.
func TestCopyWorkflows_MixedCaseFilename(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSDD.yaml":      &fstest.MapFile{Data: []byte("a")},
		"builtin/workflows/simpleWorkflow.yaml": &fstest.MapFile{Data: []byte("b")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)

	_, err1 := os.Stat(storage.GetWorkflowPath(tmpDir, "SimpleSDD"))
	assert.NoError(t, err1)
	_, err2 := os.Stat(storage.GetWorkflowPath(tmpDir, "simpleWorkflow"))
	assert.NoError(t, err2)
}

// TestCopyWorkflows_NoYamlExtension processes file without .yaml extension.
func TestCopyWorkflows_NoYamlExtension(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/README.txt": &fstest.MapFile{Data: []byte("readme content")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	// File copied as-is; name extraction uses StorageLayout which appends .yaml
	targetPath := storage.GetWorkflowPath(tmpDir, "README.txt")
	_, statErr := os.Stat(targetPath)
	assert.NoError(t, statErr)
}

// TestCopySpecFiles_NestedPath handles spec file in nested subdirectory.
func TestCopySpecFiles_NestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec", "logic"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/logic/README.md": &fstest.MapFile{Data: []byte("# Logic\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	_, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)

	content, readErr := os.ReadFile(filepath.Join(tmpDir, "spec", "logic", "README.md"))
	require.NoError(t, readErr)
	assert.Equal(t, "# Logic\n", string(content))
}

// =====================================================================
// Null / Empty Input
// =====================================================================

// TestCopyWorkflows_EmptyProjectRoot handles empty project root path.
func TestCopyWorkflows_EmptyProjectRoot(t *testing.T) {
	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	// Behavior depends on StorageLayout; may return error or write to relative path
	_, _ = copier.CopyWorkflows("")
}

// TestCopyWorkflows_RelativeProjectRoot handles relative project root path.
func TestCopyWorkflows_RelativeProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	workflowsDir := filepath.Join(projectDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)

	// Change to tmpDir so "./project" is a valid relative path
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	warnings, copyErr := copier.CopyWorkflows("./project")
	assert.NoError(t, copyErr)
	assert.Empty(t, warnings)
}

// TestCopyWorkflows_EmptyFileContent copies embedded file with empty content.
func TestCopyWorkflows_EmptyFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/Empty.yaml": &fstest.MapFile{Data: []byte{}},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	path := storage.GetWorkflowPath(tmpDir, "Empty")
	assertFileExistsWithPermissions(t, path, 0644)

	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Empty(t, content)
}

// TestCopySpecFiles_EmptyFileContent copies embedded spec file with empty content.
func TestCopySpecFiles_EmptyFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/EMPTY.md": &fstest.MapFile{Data: []byte{}},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	path := filepath.Join(tmpDir, "spec", "EMPTY.md")
	assertFileExistsWithPermissions(t, path, 0644)

	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Empty(t, content)
}

// =====================================================================
// Mock / Dependency Interaction
// =====================================================================

// TestCopyWorkflows_UsesStorageLayout verifies copier uses StorageLayout to compose target paths.
func TestCopyWorkflows_UsesStorageLayout(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)
	require.NoError(t, err)

	// Verify file is at the path StorageLayout would compose
	expectedPath := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	_, statErr := os.Stat(expectedPath)
	assert.NoError(t, statErr)
}

// TestCopyAgents_UsesStorageLayout verifies copier uses StorageLayout to compose target paths for agents.
func TestCopyAgents_UsesStorageLayout(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	_, err := copier.CopyAgents(tmpDir)
	require.NoError(t, err)

	expectedPath := storage.GetAgentPath(tmpDir, "Architect")
	_, statErr := os.Stat(expectedPath)
	assert.NoError(t, statErr)
}

// =====================================================================
// Data Independence (Copy Semantics)
// =====================================================================

// TestCopyWorkflows_FileContentIndependent verifies written file content matches embedded file exactly.
func TestCopyWorkflows_FileContentIndependent(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embeddedContent := []byte("name: SimpleSdd\nsteps:\n  - invalid yaml {{{\n")
	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: embeddedContent},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)
	require.NoError(t, err)

	path := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, embeddedContent, content)
}

// TestCopyAgents_FileContentIndependent verifies written agent file content matches embedded file exactly.
func TestCopyAgents_FileContentIndependent(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	embeddedContent := []byte("role: Architect\nprompt: some prompt\n")
	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: embeddedContent},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	_, err := copier.CopyAgents(tmpDir)
	require.NoError(t, err)

	path := storage.GetAgentPath(tmpDir, "Architect")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, embeddedContent, content)
}

// TestCopySpecFiles_FileContentIndependent verifies written spec file content matches embedded file exactly.
func TestCopySpecFiles_FileContentIndependent(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	embeddedContent := []byte("# Architecture\n\nSome markdown content here.\n")
	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: embeddedContent},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	_, err := copier.CopySpecFiles(tmpDir)
	require.NoError(t, err)

	path := filepath.Join(tmpDir, "spec", "ARCHITECTURE.md")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, embeddedContent, content)
}

// =====================================================================
// Not Immutable
// =====================================================================

// TestCopyWorkflows_InvalidYAMLContent copies file with invalid YAML content without validation.
func TestCopyWorkflows_InvalidYAMLContent(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	malformedYAML := []byte("{{invalid: yaml: [broken\n")
	embedFS := fstest.MapFS{
		"builtin/workflows/Bad.yaml": &fstest.MapFile{Data: malformedYAML},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	path := storage.GetWorkflowPath(tmpDir, "Bad")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, malformedYAML, content)
}

// TestCopyAgents_InvalidYAMLContent copies agent file with invalid YAML content without validation.
func TestCopyAgents_InvalidYAMLContent(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	malformedYAML := []byte("{{invalid: yaml: [broken\n")
	embedFS := fstest.MapFS{
		"builtin/agents/Bad.yaml": &fstest.MapFile{Data: malformedYAML},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	warnings, err := copier.CopyAgents(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	path := storage.GetAgentPath(tmpDir, "Bad")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, malformedYAML, content)
}

// TestCopySpecFiles_InvalidMarkdownContent copies spec file with invalid Markdown content without validation.
func TestCopySpecFiles_InvalidMarkdownContent(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	malformedMD := []byte("### [broken](link\n```unclosed code block\n")
	embedFS := fstest.MapFS{
		"builtin/spec/BAD.md": &fstest.MapFile{Data: malformedMD},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)
	assert.Empty(t, warnings)

	path := filepath.Join(tmpDir, "spec", "BAD.md")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, malformedMD, content)
}

// =====================================================================
// Resource Cleanup
// =====================================================================

// TestCopyWorkflows_FilePermissions verifies written workflow files have correct permissions.
func TestCopyWorkflows_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	_, err := copier.CopyWorkflows(tmpDir)
	require.NoError(t, err)

	path := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	assertFileExistsWithPermissions(t, path, 0644)
}

// TestCopyAgents_FilePermissions verifies written agent files have correct permissions.
func TestCopyAgents_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	_, err := copier.CopyAgents(tmpDir)
	require.NoError(t, err)

	path := storage.GetAgentPath(tmpDir, "Architect")
	assertFileExistsWithPermissions(t, path, 0644)
}

// TestCopySpecFiles_FilePermissions verifies written spec files have correct permissions.
func TestCopySpecFiles_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("# Arch\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	_, err := copier.CopySpecFiles(tmpDir)
	require.NoError(t, err)

	path := filepath.Join(tmpDir, "spec", "ARCHITECTURE.md")
	assertFileExistsWithPermissions(t, path, 0644)
}

// =====================================================================
// State Transitions
// =====================================================================

// TestCopyWorkflows_MixedExistingAndNew copies new files and skips existing files in single invocation.
func TestCopyWorkflows_MixedExistingAndNew(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	// Create existing file
	firstPath := storage.GetWorkflowPath(tmpDir, "First")
	require.NoError(t, os.WriteFile(firstPath, []byte("existing content"), 0644))

	embedFS := fstest.MapFS{
		"builtin/workflows/First.yaml":  &fstest.MapFile{Data: []byte("new first\n")},
		"builtin/workflows/Second.yaml": &fstest.MapFile{Data: []byte("new second\n")},
		"builtin/workflows/Third.yaml":  &fstest.MapFile{Data: []byte("new third\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "First.yaml")

	// First.yaml unchanged
	content, readErr := os.ReadFile(firstPath)
	require.NoError(t, readErr)
	assert.Equal(t, "existing content", string(content))

	// Second.yaml and Third.yaml written
	_, err2 := os.Stat(storage.GetWorkflowPath(tmpDir, "Second"))
	assert.NoError(t, err2)
	_, err3 := os.Stat(storage.GetWorkflowPath(tmpDir, "Third"))
	assert.NoError(t, err3)
}

// TestCopyAgents_MixedExistingAndNew copies new agent files and skips existing files in single invocation.
func TestCopyAgents_MixedExistingAndNew(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create existing file
	archPath := storage.GetAgentPath(tmpDir, "Architect")
	require.NoError(t, os.WriteFile(archPath, []byte("existing"), 0644))

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("new architect\n")},
		"builtin/agents/QaAnalyst.yaml": &fstest.MapFile{Data: []byte("new qa\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	warnings, err := copier.CopyAgents(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "Architect.yaml")

	// Architect.yaml unchanged
	content, readErr := os.ReadFile(archPath)
	require.NoError(t, readErr)
	assert.Equal(t, "existing", string(content))

	// QaAnalyst.yaml written
	_, statErr := os.Stat(storage.GetAgentPath(tmpDir, "QaAnalyst"))
	assert.NoError(t, statErr)
}

// TestCopySpecFiles_MixedExistingAndNew copies new spec files and skips existing files in single invocation.
func TestCopySpecFiles_MixedExistingAndNew(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	// Create existing file
	archPath := filepath.Join(tmpDir, "spec", "ARCHITECTURE.md")
	require.NoError(t, os.WriteFile(archPath, []byte("existing arch"), 0644))

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("new arch\n")},
		"builtin/spec/CONVENTIONS.md":  &fstest.MapFile{Data: []byte("new conv\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "ARCHITECTURE.md")

	// ARCHITECTURE.md unchanged
	content, readErr := os.ReadFile(archPath)
	require.NoError(t, readErr)
	assert.Equal(t, "existing arch", string(content))

	// CONVENTIONS.md written
	_, statErr := os.Stat(filepath.Join(tmpDir, "spec", "CONVENTIONS.md"))
	assert.NoError(t, statErr)
}

// =====================================================================
// Catch Behaviour
// =====================================================================

// TestCopyWorkflows_ExistingFileIsDirectory skips when target path exists as directory.
func TestCopyWorkflows_ExistingFileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	// Create a directory at the target path
	targetPath := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	require.NoError(t, os.MkdirAll(targetPath, 0755))

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Equal(t, "Warning: workflow definition 'SimpleSdd.yaml' already exists, skipping", warnings[0])
}

// TestCopyAgents_ExistingFileIsDirectory skips when target path exists as directory.
func TestCopyAgents_ExistingFileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".spectra", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create a directory at the target path
	targetPath := storage.GetAgentPath(tmpDir, "Architect")
	require.NoError(t, os.MkdirAll(targetPath, 0755))

	embedFS := fstest.MapFS{
		"builtin/agents/Architect.yaml": &fstest.MapFile{Data: []byte("role: Architect\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, embedFS, nil)
	warnings, err := copier.CopyAgents(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Equal(t, "Warning: agent definition 'Architect.yaml' already exists, skipping", warnings[0])
}

// TestCopySpecFiles_ExistingFileIsDirectory skips when target spec file path exists as directory.
func TestCopySpecFiles_ExistingFileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "spec"), 0755))

	// Create a directory at the target path
	targetPath := filepath.Join(tmpDir, "spec", "ARCHITECTURE.md")
	require.NoError(t, os.MkdirAll(targetPath, 0755))

	embedFS := fstest.MapFS{
		"builtin/spec/ARCHITECTURE.md": &fstest.MapFile{Data: []byte("# Arch\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(nil, nil, embedFS)
	warnings, err := copier.CopySpecFiles(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Equal(t, "Warning: spec file 'ARCHITECTURE.md' already exists, skipping", warnings[0])
}

// TestCopyWorkflows_ExistingFileUnreadable skips when target file exists but is not readable.
func TestCopyWorkflows_ExistingFileUnreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test requires non-root user")
	}

	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".spectra", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	// Create file with no permissions
	targetPath := storage.GetWorkflowPath(tmpDir, "SimpleSdd")
	require.NoError(t, os.WriteFile(targetPath, []byte("data"), 0000))
	t.Cleanup(func() { os.Chmod(targetPath, 0644) })

	embedFS := fstest.MapFS{
		"builtin/workflows/SimpleSdd.yaml": &fstest.MapFile{Data: []byte("name: SimpleSdd\n")},
	}

	copier := spectra.NewBuiltinResourceCopier(embedFS, nil, nil)
	warnings, err := copier.CopyWorkflows(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "SimpleSdd.yaml")
}

// =====================================================================
// Test Helpers
// =====================================================================

// assertFileExistsWithPermissions asserts that the file at path exists and has the expected permissions.
func assertFileExistsWithPermissions(t *testing.T, path string, expectedPerm os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "file should exist: %s", path)
	assert.Equal(t, expectedPerm, info.Mode().Perm(), "file permissions mismatch: %s", path)
}
