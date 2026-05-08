package spectra

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/logger"
)

// --- Sentinel Errors ---

var (
	errFakeFinderFailure = errors.New("fake: spectra finder failed")
	errFakeGetwdFailure  = errors.New("fake: os.Getwd failed")
	errFakePhaseFailure  = errors.New("fake: phase failed")
)

// --- Fake: SpectraFinder for ClearCommand ---

// fakeSpectraFinderForClear records calls to Find() and returns configured values.
type fakeSpectraFinderForClear struct {
	projectRoot   string
	err           error
	findCallCount int
}

func (f *fakeSpectraFinderForClear) Find() (string, error) {
	f.findCallCount++
	if f.err != nil {
		return "", f.err
	}
	return f.projectRoot, nil
}

// --- Fake: StorageLayout for ClearCommand ---

// clearSessionDirCall records a single invocation of GetSessionDir.
type clearSessionDirCall struct {
	projectRoot string
	uuid        string
}

// fakeClearStorageLayout records calls and returns session directory paths.
type fakeClearStorageLayout struct {
	projectRoot string

	getSessionDirCalls      []clearSessionDirCall
	getSessionsDirCallCount int
	getSessionsDirOverride  string
}

func (f *fakeClearStorageLayout) GetSessionDir(projectRoot, uuid string) string {
	f.getSessionDirCalls = append(f.getSessionDirCalls, clearSessionDirCall{
		projectRoot: projectRoot,
		uuid:        uuid,
	})
	return filepath.Join(projectRoot, ".spectra", "sessions", uuid)
}

func (f *fakeClearStorageLayout) GetSessionsDir(projectRoot string) string {
	f.getSessionsDirCallCount++
	if f.getSessionsDirOverride != "" {
		return f.getSessionsDirOverride
	}
	return filepath.Join(projectRoot, ".spectra", "sessions")
}

// --- Fake: InitCommand Dependencies ---

// fakeGitignoreEnsurer records calls to Ensure().
type fakeGitignoreEnsurer struct {
	receivedProjectRoot string
	called              bool
	err                 error
	callOrder           *[]string // shared slice for ordering verification
}

func (f *fakeGitignoreEnsurer) Ensure(projectRoot string) error {
	f.called = true
	f.receivedProjectRoot = projectRoot
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "gitignore")
	}
	return f.err
}

// fakeDirectoryCreator records calls to CreateAll().
type fakeDirectoryCreator struct {
	receivedProjectRoot string
	called              bool
	err                 error
	callOrder           *[]string // shared slice for ordering verification
}

func (f *fakeDirectoryCreator) CreateAll(projectRoot string) error {
	f.called = true
	f.receivedProjectRoot = projectRoot
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "directories")
	}
	return f.err
}

// fakeBuiltinResourceCopier records calls to Copy methods.
type fakeBuiltinResourceCopier struct {
	receivedProjectRoot string

	copyWorkflowsCalled bool
	copyAgentsCalled    bool
	copySpecFilesCalled bool

	copyWorkflowsWarnings []string
	copyAgentsWarnings    []string
	copySpecFilesWarnings []string

	copyWorkflowsErr error
	copyAgentsErr    error
	copySpecFilesErr error

	callOrder *[]string // shared slice for ordering verification
}

func (f *fakeBuiltinResourceCopier) CopyWorkflows(projectRoot string) ([]string, error) {
	f.copyWorkflowsCalled = true
	f.receivedProjectRoot = projectRoot
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "workflows")
	}
	return f.copyWorkflowsWarnings, f.copyWorkflowsErr
}

func (f *fakeBuiltinResourceCopier) CopyAgents(projectRoot string) ([]string, error) {
	f.copyAgentsCalled = true
	f.receivedProjectRoot = projectRoot
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "agents")
	}
	return f.copyAgentsWarnings, f.copyAgentsErr
}

func (f *fakeBuiltinResourceCopier) CopySpecFiles(projectRoot string) ([]string, error) {
	f.copySpecFilesCalled = true
	f.receivedProjectRoot = projectRoot
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "specfiles")
	}
	return f.copySpecFilesWarnings, f.copySpecFilesErr
}

// fakeInitDeps bundles all dependencies for init command tests.
type fakeInitDeps struct {
	cwd              string
	getwdErr         error
	gitignoreEnsurer *fakeGitignoreEnsurer
	directoryCreator *fakeDirectoryCreator
	copier           *fakeBuiltinResourceCopier
	callOrder        []string
}

// newFakeInitDeps returns a fully configured fakeInitDeps with no errors
// and a default CWD of "/tmp/fake-project". All fakes share a single
// callOrder slice so phase ordering can be verified.
func newFakeInitDeps() *fakeInitDeps {
	deps := &fakeInitDeps{
		cwd:              "/tmp/fake-project",
		gitignoreEnsurer: &fakeGitignoreEnsurer{},
		directoryCreator: &fakeDirectoryCreator{},
		copier:           &fakeBuiltinResourceCopier{},
	}
	// Wire all fakes to the shared callOrder slice for ordering tests.
	deps.gitignoreEnsurer.callOrder = &deps.callOrder
	deps.directoryCreator.callOrder = &deps.callOrder
	deps.copier.callOrder = &deps.callOrder
	return deps
}

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

// assertPathNotExists asserts that the given path does not exist or is unreachable
// (e.g., a parent component is a file rather than a directory).
func assertPathNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.Error(t, err, "assertPathNotExists: %s should not exist but does", path)
}

// --- Fake: Runtime for RunCommand ---

// fakeRuntime records calls to Run() and returns configured values.
// Used by run_test.go to test the run subcommand in isolation.
type fakeRuntime struct {
	// Configuration
	exitCode int
	err      error

	// Captured state
	calledCount  int
	workflowName string
	loggerWasNil bool
}

func newFakeRuntime(exitCode int, err error) *fakeRuntime {
	return &fakeRuntime{
		exitCode: exitCode,
		err:      err,
	}
}

// Run satisfies the RunRuntime interface defined in run.go.
// It captures invocation details and returns the configured result.
func (f *fakeRuntime) Run(workflowName string, log logger.Logger) (int, error) {
	f.calledCount++
	f.workflowName = workflowName
	f.loggerWasNil = (log == nil)
	return f.exitCode, f.err
}
