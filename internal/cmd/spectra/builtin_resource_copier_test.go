package spectra

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — CopyWorkflows ---

func TestBuiltinResourceCopier_CopyWorkflows_WritesAllFiles(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	workflowsFS := buildMapFS(map[string]string{
		"workflows/DefaultLogicSpec.yaml": "logic-spec-content",
		"workflows/DefaultTestSpec.yaml":  "test-spec-content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	// Verify files written to paths from StorageLayout
	path1 := layout.GetWorkflowPath(projectRoot, "DefaultLogicSpec")
	path2 := layout.GetWorkflowPath(projectRoot, "DefaultTestSpec")
	assertFileContent(t, path1, "logic-spec-content")
	assertFileContent(t, path2, "test-spec-content")
	assertFilePermissions(t, path1, 0644)
	assertFilePermissions(t, path2, 0644)
}

func TestBuiltinResourceCopier_CopyWorkflows_SkipsExisting(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	workflowsFS := buildMapFS(map[string]string{
		"workflows/DefaultLogicSpec.yaml": "new-content",
	})

	layout := newFakeStorageLayout()
	targetPath := layout.GetWorkflowPath(projectRoot, "DefaultLogicSpec")
	writeFile(t, targetPath, "existing-content", 0644)

	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Equal(t, "Warning: workflow definition 'DefaultLogicSpec.yaml' already exists, skipping", warnings[0])

	// Content unchanged
	assertFileContent(t, targetPath, "existing-content")
}

// --- Happy Path — CopyAgents ---

func TestBuiltinResourceCopier_CopyAgents_WritesAllFiles(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "agents"))

	agentsFS := buildMapFS(map[string]string{
		"agents/TestAgent.yaml": "agent-content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, agentsFS, fstest.MapFS{}, layout)

	warnings, err := copier.CopyAgents(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	targetPath := layout.GetAgentPath(projectRoot, "TestAgent")
	assertFileContent(t, targetPath, "agent-content")
	assertFilePermissions(t, targetPath, 0644)
}

func TestBuiltinResourceCopier_CopyAgents_SkipsExisting(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "agents"))

	agentsFS := buildMapFS(map[string]string{
		"agents/TestAgent.yaml": "new-content",
	})

	layout := newFakeStorageLayout()
	targetPath := layout.GetAgentPath(projectRoot, "TestAgent")
	writeFile(t, targetPath, "existing-content", 0644)

	copier := NewBuiltinResourceCopier(fstest.MapFS{}, agentsFS, fstest.MapFS{}, layout)

	warnings, err := copier.CopyAgents(projectRoot)
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Equal(t, "Warning: agent definition 'TestAgent.yaml' already exists, skipping", warnings[0])

	// Content unchanged
	assertFileContent(t, targetPath, "existing-content")
}

// --- Happy Path — CopySpecFiles ---

func TestBuiltinResourceCopier_CopySpecFiles_WritesAllFiles(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, "spec"))
	ensureDir(t, filepath.Join(projectRoot, "spec", "logic"))
	ensureDir(t, filepath.Join(projectRoot, "spec", "test"))

	specFS := buildMapFS(map[string]string{
		"spec/ARCHITECTURE.md": "arch-content",
		"spec/CONVENTIONS.md":  "conv-content",
		"spec/logic/README.md": "logic-readme",
		"spec/test/README.md":  "test-readme",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, fstest.MapFS{}, specFS, layout)

	warnings, err := copier.CopySpecFiles(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	assertFileContent(t, filepath.Join(projectRoot, "spec", "ARCHITECTURE.md"), "arch-content")
	assertFileContent(t, filepath.Join(projectRoot, "spec", "CONVENTIONS.md"), "conv-content")
	assertFileContent(t, filepath.Join(projectRoot, "spec", "logic", "README.md"), "logic-readme")
	assertFileContent(t, filepath.Join(projectRoot, "spec", "test", "README.md"), "test-readme")
	assertFilePermissions(t, filepath.Join(projectRoot, "spec", "ARCHITECTURE.md"), 0644)
	assertFilePermissions(t, filepath.Join(projectRoot, "spec", "CONVENTIONS.md"), 0644)
	assertFilePermissions(t, filepath.Join(projectRoot, "spec", "logic", "README.md"), 0644)
	assertFilePermissions(t, filepath.Join(projectRoot, "spec", "test", "README.md"), 0644)
}

func TestBuiltinResourceCopier_CopySpecFiles_SkipsExisting(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, "spec"))

	// Pre-create one file
	writeFile(t, filepath.Join(projectRoot, "spec", "ARCHITECTURE.md"), "existing", 0644)

	specFS := buildMapFS(map[string]string{
		"spec/ARCHITECTURE.md": "new-content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, fstest.MapFS{}, specFS, layout)

	warnings, err := copier.CopySpecFiles(projectRoot)
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Equal(t, "Warning: spec file 'ARCHITECTURE.md' already exists, skipping", warnings[0])

	// Content unchanged
	assertFileContent(t, filepath.Join(projectRoot, "spec", "ARCHITECTURE.md"), "existing")
}

func TestBuiltinResourceCopier_CopySpecFiles_PreservesSubdirectoryStructure(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, "spec", "logic"))

	specFS := buildMapFS(map[string]string{
		"spec/logic/README.md": "logic-readme-content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, fstest.MapFS{}, specFS, layout)

	warnings, err := copier.CopySpecFiles(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	assertFileExists(t, filepath.Join(projectRoot, "spec", "logic", "README.md"))
	assertFileContent(t, filepath.Join(projectRoot, "spec", "logic", "README.md"), "logic-readme-content")
}

// --- Error Propagation ---

func TestBuiltinResourceCopier_CopyWorkflows_WriteError(t *testing.T) {


	projectRoot := t.TempDir()
	// Do NOT create .spectra/workflows/ — write should fail

	workflowsFS := buildMapFS(map[string]string{
		"workflows/MyWorkflow.yaml": "content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	_ = warnings // May or may not be empty depending on ordering
}

func TestBuiltinResourceCopier_CopyAgents_WriteError(t *testing.T) {


	projectRoot := t.TempDir()
	// Do NOT create .spectra/agents/ — write should fail

	agentsFS := buildMapFS(map[string]string{
		"agents/MyAgent.yaml": "content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, agentsFS, fstest.MapFS{}, layout)

	warnings, err := copier.CopyAgents(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	_ = warnings
}

func TestBuiltinResourceCopier_CopySpecFiles_WriteError(t *testing.T) {


	projectRoot := t.TempDir()
	// Do NOT create spec/ directory — write should fail

	specFS := buildMapFS(map[string]string{
		"spec/ARCHITECTURE.md": "content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, fstest.MapFS{}, specFS, layout)

	warnings, err := copier.CopySpecFiles(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write built-in file")
	_ = warnings
}

func TestBuiltinResourceCopier_CopyWorkflows_FailFast(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	// Two workflow files: first one exists (generates warning), second one goes to invalid path
	workflowsFS := buildMapFS(map[string]string{
		"workflows/ExistingWorkflow.yaml": "content-a",
		"workflows/FailingWorkflow.yaml":  "content-b",
	})

	layout := newFakeStorageLayout()

	// Pre-create first file to generate a warning
	firstPath := layout.GetWorkflowPath(projectRoot, "ExistingWorkflow")
	writeFile(t, firstPath, "pre-existing", 0644)

	// Point second workflow to an invalid path (missing parent dir)
	layout.workflowPaths["FailingWorkflow"] = filepath.Join(projectRoot, "nonexistent", "dir", "FailingWorkflow.yaml")

	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.Error(t, err)
	// Should have accumulated warning for the first (existing) file
	assert.NotEmpty(t, warnings)
	assert.Contains(t, err.Error(), "failed to write built-in file")
}

// --- Idempotency ---

func TestBuiltinResourceCopier_CopyWorkflows_AllExist(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	workflowsFS := buildMapFS(map[string]string{
		"workflows/DefaultLogicSpec.yaml": "content-a",
		"workflows/DefaultTestSpec.yaml":  "content-b",
	})

	layout := newFakeStorageLayout()

	// Pre-create all target files
	writeFile(t, layout.GetWorkflowPath(projectRoot, "DefaultLogicSpec"), "existing-a", 0644)
	writeFile(t, layout.GetWorkflowPath(projectRoot, "DefaultTestSpec"), "existing-b", 0644)

	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.NoError(t, err)
	assert.Len(t, warnings, 2)

	// Content unchanged
	assertFileContent(t, layout.GetWorkflowPath(projectRoot, "DefaultLogicSpec"), "existing-a")
	assertFileContent(t, layout.GetWorkflowPath(projectRoot, "DefaultTestSpec"), "existing-b")
}

// --- Null / Empty Input ---

func TestBuiltinResourceCopier_CopyWorkflows_EmptyFS(t *testing.T) {


	projectRoot := t.TempDir()

	// Empty FS — no workflow files
	workflowsFS := buildEmptyMapFS("workflows")

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, warnings)
}

func TestBuiltinResourceCopier_CopyWorkflows_EmptyFileContent(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	// Workflow file with empty content
	workflowsFS := buildMapFS(map[string]string{
		"workflows/Empty.yaml": "",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	targetPath := layout.GetWorkflowPath(projectRoot, "Empty")
	assertFileExists(t, targetPath)
	info, statErr := os.Stat(targetPath)
	require.NoError(t, statErr)
	assert.Equal(t, int64(0), info.Size())
	assertFilePermissions(t, targetPath, 0644)
}

// --- Boundary Values — Target Exists as Directory ---

func TestBuiltinResourceCopier_CopyWorkflows_TargetIsDirectory(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	workflowsFS := buildMapFS(map[string]string{
		"workflows/DirWorkflow.yaml": "content",
	})

	layout := newFakeStorageLayout()

	// Create a directory where the file would be written
	targetPath := layout.GetWorkflowPath(projectRoot, "DirWorkflow")
	ensureDir(t, targetPath)

	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	warnings, err := copier.CopyWorkflows(projectRoot)
	require.NoError(t, err)
	assert.NotEmpty(t, warnings)
}

// --- Mock / Dependency Interaction ---

func TestBuiltinResourceCopier_CopyWorkflows_UsesStorageLayout(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "workflows"))

	workflowsFS := buildMapFS(map[string]string{
		"workflows/MyWorkflow.yaml": "wf-content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(workflowsFS, fstest.MapFS{}, fstest.MapFS{}, layout)

	_, _ = copier.CopyWorkflows(projectRoot)

	calls := layout.getWorkflowCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, projectRoot, calls[0].projectRoot)
	assert.Equal(t, "MyWorkflow", calls[0].name)
}

func TestBuiltinResourceCopier_CopyAgents_UsesStorageLayout(t *testing.T) {


	projectRoot := t.TempDir()
	ensureDir(t, filepath.Join(projectRoot, ".spectra", "agents"))

	agentsFS := buildMapFS(map[string]string{
		"agents/MyAgent.yaml": "agent-content",
	})

	layout := newFakeStorageLayout()
	copier := NewBuiltinResourceCopier(fstest.MapFS{}, agentsFS, fstest.MapFS{}, layout)

	_, _ = copier.CopyAgents(projectRoot)

	calls := layout.getAgentCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, projectRoot, calls[0].projectRoot)
	assert.Equal(t, "MyAgent", calls[0].name)
}
