package spectra

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

// --- Fake: StorageLayout ---

// storageLayoutCall records a single invocation of a StorageLayout path method.
type storageLayoutCall struct {
	projectRoot string
	name        string
}

// fakeStorageLayout records calls to GetWorkflowPath and GetAgentPath, returning
// configured paths. Useful for verifying the BuiltinResourceCopier delegates
// path composition correctly.
type fakeStorageLayout struct {
	mu sync.Mutex

	// workflowPaths maps workflow name -> target path
	workflowPaths map[string]string

	// agentPaths maps agent name -> target path
	agentPaths map[string]string

	// Captured calls
	workflowCalls []storageLayoutCall
	agentCalls    []storageLayoutCall
}

func newFakeStorageLayout() *fakeStorageLayout {
	return &fakeStorageLayout{
		workflowPaths: make(map[string]string),
		agentPaths:    make(map[string]string),
	}
}

func (f *fakeStorageLayout) GetWorkflowPath(projectRoot, name string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.workflowCalls = append(f.workflowCalls, storageLayoutCall{
		projectRoot: projectRoot,
		name:        name,
	})
	if p, ok := f.workflowPaths[name]; ok {
		return p
	}
	// Default: mimic real layout
	return filepath.Join(projectRoot, ".spectra", "workflows", name+".yaml")
}

func (f *fakeStorageLayout) GetAgentPath(projectRoot, name string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.agentCalls = append(f.agentCalls, storageLayoutCall{
		projectRoot: projectRoot,
		name:        name,
	})
	if p, ok := f.agentPaths[name]; ok {
		return p
	}
	// Default: mimic real layout
	return filepath.Join(projectRoot, ".spectra", "agents", name+".yaml")
}

func (f *fakeStorageLayout) getWorkflowCalls() []storageLayoutCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]storageLayoutCall(nil), f.workflowCalls...)
}

func (f *fakeStorageLayout) getAgentCalls() []storageLayoutCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]storageLayoutCall(nil), f.agentCalls...)
}

// --- Fixture Builders ---

// buildMapFS creates a testing/fstest.MapFS from a map of relative paths to content.
func buildMapFS(files map[string]string) fstest.MapFS {
	m := fstest.MapFS{}
	for path, content := range files {
		m[path] = &fstest.MapFile{
			Data: []byte(content),
			Mode: 0644,
		}
	}
	return m
}

// buildEmptyMapFS creates a testing/fstest.MapFS containing only directories (no files).
func buildEmptyMapFS(dirs ...string) fstest.MapFS {
	m := fstest.MapFS{}
	for _, d := range dirs {
		m[d] = &fstest.MapFile{
			Mode: fs.ModeDir | 0755,
		}
	}
	return m
}

// ensureDir creates a directory at the given path, failing the test if it cannot.
func ensureDir(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755), "ensureDir: failed to create %s", path)
}

// writeFile creates a file at the given path with the given content and permissions.
func writeFile(t *testing.T, path string, content string, perm os.FileMode) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), perm), "writeFile: failed to write %s", path)
}

// readFileContent reads file content, failing the test on error.
func readFileContent(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "readFileContent: failed to read %s", path)
	return string(data)
}

// assertFileContent asserts that a file exists and its content matches expected.
func assertFileContent(t *testing.T, path string, expected string) {
	t.Helper()
	actual := readFileContent(t, path)
	require.Equal(t, expected, actual, "file content mismatch: %s", path)
}

// assertFilePermissions asserts that the file at path has the given permissions.
func assertFilePermissions(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "assertFilePermissions: stat failed for %s", path)
	actual := info.Mode().Perm()
	require.Equal(t, perm, actual, "permission mismatch for %s: want %o, got %o", path, perm, actual)
}

// assertFileExists asserts that a file (not directory) exists at path.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "assertFileExists: %s does not exist", path)
	require.False(t, info.IsDir(), "assertFileExists: %s is a directory, not a file", path)
}

// assertDirExists asserts that a directory exists at path.
func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "assertDirExists: %s does not exist", path)
	require.True(t, info.IsDir(), "assertDirExists: %s is not a directory", path)
}

// assertDirPermissions asserts that the directory at path has the given permissions.
func assertDirPermissions(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "assertDirPermissions: stat failed for %s", path)
	actual := info.Mode().Perm()
	require.Equal(t, perm, actual, "dir permission mismatch for %s: want %o, got %o", path, perm, actual)
}

// assertPathNotExists asserts that the given path does not exist.
func assertPathNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "assertPathNotExists: %s should not exist but does", path)
}
