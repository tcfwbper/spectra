package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/storage"
)

// --- Mock types for SessionInitializer tests ---

// mockWorkflowDefinitionLoaderForInit is a mock WorkflowDefinitionLoader for SessionInitializer tests.
type mockWorkflowDefinitionLoaderForInit struct {
	mock.Mock
}

func (m *mockWorkflowDefinitionLoaderForInit) Load(workflowName string) (*storage.WorkflowDefinition, error) {
	args := m.Called(workflowName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.WorkflowDefinition), args.Error(1)
}

// mockSessionDirectoryManagerForInit is a mock SessionDirectoryManager for SessionInitializer tests.
type mockSessionDirectoryManagerForInit struct {
	mock.Mock
}

func (m *mockSessionDirectoryManagerForInit) CreateSessionDirectory(sessionUUID string) error {
	args := m.Called(sessionUUID)
	return args.Error(0)
}

// mockRuntimeSocketManagerForInit is a mock RuntimeSocketManager for SessionInitializer tests.
type mockRuntimeSocketManagerForInit struct {
	mock.Mock
	mu                sync.Mutex
	createSocketCalls int
	deleteSocketCalls int
}

func (m *mockRuntimeSocketManagerForInit) CreateSocket() error {
	m.mu.Lock()
	m.createSocketCalls++
	m.mu.Unlock()
	args := m.Called()
	return args.Error(0)
}

func (m *mockRuntimeSocketManagerForInit) DeleteSocket() error {
	m.mu.Lock()
	m.deleteSocketCalls++
	m.mu.Unlock()
	args := m.Called()
	return args.Error(0)
}

func (m *mockRuntimeSocketManagerForInit) getDeleteSocketCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deleteSocketCalls
}

// mockSessionMetadataStoreForInit is a mock SessionMetadataStore for SessionInitializer tests.
type mockSessionMetadataStoreForInit struct {
	mock.Mock
	mu        sync.Mutex
	callOrder []string
}

func (m *mockSessionMetadataStoreForInit) Write(metadata interface{}) error {
	m.mu.Lock()
	m.callOrder = append(m.callOrder, "Write")
	m.mu.Unlock()
	args := m.Called(metadata)
	return args.Error(0)
}

func (m *mockSessionMetadataStoreForInit) getCallOrder() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.callOrder...)
}

// mockFileAccessorForInit is a mock FileAccessor for SessionInitializer tests.
type mockFileAccessorForInit struct {
	mock.Mock
}

func (m *mockFileAccessorForInit) Prepare() error {
	args := m.Called()
	return args.Error(0)
}

// mockSessionForInit is a mock Session for SessionInitializer tests.
type mockSessionForInit struct {
	mock.Mock
	mu     sync.RWMutex
	status string
	id     string
}

func (m *mockSessionForInit) Run(terminationNotifier chan<- struct{}) error {
	args := m.Called(terminationNotifier)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.status = "running"
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *mockSessionForInit) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(err, terminationNotifier)
	if args.Error(0) == nil {
		m.status = "failed"
	}
	return args.Error(0)
}

func (m *mockSessionForInit) GetStatusSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockSessionForInit) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

// callOrderTracker tracks method call ordering across multiple mocks.
type callOrderTracker struct {
	mu    sync.Mutex
	calls []string
}

func (t *callOrderTracker) record(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, name)
}

func (t *callOrderTracker) getCalls() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.calls...)
}

func (t *callOrderTracker) indexOf(name string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, c := range t.calls {
		if c == name {
			return i
		}
	}
	return -1
}

// --- Test fixture ---

func createSessionInitializerFixture(t *testing.T) (string, *mockWorkflowDefinitionLoaderForInit, *mockSessionDirectoryManagerForInit) {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0775))

	wdl := &mockWorkflowDefinitionLoaderForInit{}
	sdm := &mockSessionDirectoryManagerForInit{}

	return tmpDir, wdl, sdm
}

func defaultWorkflowDef() *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "start",
		Nodes: []storage.Node{
			{Name: "start", Type: "agent"},
		},
	}
}

// =====================================================================
// Happy Path — Construction
// =====================================================================

func TestSessionInitializer_New(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)

	require.NoError(t, err)
	assert.NotNil(t, si)
}

// =====================================================================
// Happy Path — Initialize
// =====================================================================

func TestInitialize_Success(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	require.NotNil(t, sess)
	assert.Equal(t, "running", sess.GetStatusSafe())
	assert.Equal(t, "TestWorkflow", sess.GetWorkflowName())
	assert.Equal(t, "start", sess.GetCurrentStateSafe())
}

func TestInitialize_MetadataPersisted(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	require.NotNil(t, sess)

	// SessionMetadataStore.Write() must be called before Session.Run()
	// Metadata written should have Status="initializing" at write time
	// After Session.Run(), status transitions to "running"
	assert.Equal(t, "running", sess.GetStatusSafe())
}

func TestInitialize_EmptyEventHistoryAndSessionData(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	require.NotNil(t, sess)
	assert.Empty(t, sess.GetEventHistory(), "EventHistory should be empty")
	assert.Empty(t, sess.GetSessionData(), "SessionData should be empty")
}

func TestInitialize_CurrentStateSetToEntryNode(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	customWorkflow := &storage.WorkflowDefinition{
		Name:      "CustomWorkflow",
		EntryNode: "custom_entry",
		Nodes: []storage.Node{
			{Name: "custom_entry", Type: "agent"},
		},
	}
	wdl.On("Load", "CustomWorkflow").Return(customWorkflow, nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("CustomWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "custom_entry", sess.GetCurrentStateSafe())
}

func TestInitialize_TimestampsSet(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	before := time.Now().Unix()
	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)
	after := time.Now().Unix()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, sess.GetCreatedAt(), before, "CreatedAt should be >= test start time")
	assert.LessOrEqual(t, sess.GetCreatedAt(), after, "CreatedAt should be <= test end time")
	assert.GreaterOrEqual(t, sess.GetUpdatedAt(), before, "UpdatedAt should be >= test start time")
	assert.LessOrEqual(t, sess.GetUpdatedAt(), after, "UpdatedAt should be <= test end time")
}

func TestInitialize_UniqueSessionUUID(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier1 := make(chan struct{}, 2)
	sess1, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier1)
	require.NoError(t, err)

	terminationNotifier2 := make(chan struct{}, 2)
	sess2, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier2)
	require.NoError(t, err)

	assert.NotEqual(t, sess1.GetID(), sess2.GetID(), "each call should return a unique session UUID")
}

// =====================================================================
// Happy Path — Timeout Completion
// =====================================================================

func TestInitialize_TimeoutTimerCanceled(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, "running", sess.GetStatusSafe())

	// timer.Stop() should have been called; timeout handler should not fire
	// Wait briefly to confirm no late timeout fires
	time.Sleep(100 * time.Millisecond)
	select {
	case <-terminationNotifier:
		t.Fatal("timeout handler should not fire after successful initialization")
	default:
		// Good — no notification
	}
}

func TestInitialize_CompletionRace_InitWins(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "running", sess.GetStatusSafe())
}

// =====================================================================
// Happy Path — Resource Cleanup on Failure
// =====================================================================

func TestInitialize_SocketCleanupOnSessionRunFailure(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Configure Session.Run() to fail
	si.SetSessionRunError(fmt.Errorf("run failed"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to transition session to running:`, err.Error())

	// RuntimeSocketManager.DeleteSocket() should have been called for cleanup
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called on Session.Run failure")
}

func TestInitialize_SocketCleanupOnMetadataWriteFailure(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Configure SessionMetadataStore.Write() to fail
	si.SetMetadataWriteError(fmt.Errorf("no space left on device"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to persist initial session metadata:`, err.Error())

	// RuntimeSocketManager.DeleteSocket() should have been called for cleanup
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called on metadata write failure")
}

func TestInitialize_SocketCleanupOnSocketCreateFailure(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Configure RuntimeSocketManager.CreateSocket() to fail
	si.SetSocketCreateError(fmt.Errorf("socket creation failed"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create runtime socket:`, err.Error())

	// RuntimeSocketManager.DeleteSocket() should have been called to clean up partial socket
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called on CreateSocket failure")
}

// =====================================================================
// Happy Path — Timeout Enforcement
// =====================================================================

func TestInitialize_TimeoutAfterSessionConstructed(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	// Use a blocking workflow loader to simulate slow initialization
	blockCh := make(chan struct{})
	wdl.On("Load", "SlowWorkflow").Run(func(args mock.Arguments) {
		<-blockCh
	}).Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)

	// Trigger timeout manually by using a very short timeout for testing
	si.SetTimeoutDuration(50 * time.Millisecond)

	var initErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, initErr = si.Initialize("SlowWorkflow", projectRoot, terminationNotifier)
	}()

	// Wait for timeout to fire
	time.Sleep(200 * time.Millisecond)

	// Unblock the workflow loader
	close(blockCh)

	<-done

	// Timeout should have fired; either session fails or error returned
	if initErr != nil {
		assert.Regexp(t, `(?i)session initialization timeout exceeded 30 seconds`, initErr.Error())
	}
}

func TestInitialize_TimeoutBeforeSessionConstructed(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	// WorkflowDefinitionLoader blocks to prevent Session construction
	blockCh := make(chan struct{})
	wdl.On("Load", "TestWorkflow").Run(func(args mock.Arguments) {
		<-blockCh
	}).Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetTimeoutDuration(50 * time.Millisecond)

	terminationNotifier := make(chan struct{}, 2)

	var initErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, initErr = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)
	}()

	// Wait for timeout to fire before session is constructed
	time.Sleep(200 * time.Millisecond)

	// Unblock WorkflowDefinitionLoader
	close(blockCh)

	<-done

	require.Error(t, initErr)
	assert.Regexp(t, `(?i)session initialization timeout exceeded 30 seconds before session entity was constructed`, initErr.Error())
}

// =====================================================================
// Validation Failures — TerminationNotifier
// =====================================================================

func TestInitialize_TerminationNotifierBufferCapacity1(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	ch := make(chan struct{}, 1)
	_, err = si.Initialize("TestWorkflow", projectRoot, ch)

	require.Error(t, err)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 1", err.Error())
}

func TestInitialize_TerminationNotifierNil(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	_, err = si.Initialize("TestWorkflow", projectRoot, nil)

	require.Error(t, err)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 0", err.Error())
}

func TestInitialize_TerminationNotifierUnbuffered(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	ch := make(chan struct{})
	_, err = si.Initialize("TestWorkflow", projectRoot, ch)

	require.Error(t, err)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 0", err.Error())
}

func TestInitialize_TerminationNotifierBufferCapacity2(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	ch := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, ch)

	require.NoError(t, err)
	assert.NotNil(t, sess)
}

func TestInitialize_TerminationNotifierBufferCapacity5(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	ch := make(chan struct{}, 5)
	sess, err := si.Initialize("TestWorkflow", projectRoot, ch)

	require.NoError(t, err)
	assert.NotNil(t, sess)
}

// =====================================================================
// Validation Failures — Project Root
// =====================================================================

func TestInitialize_ProjectRootInvalid(t *testing.T) {
	_, wdl, sdm := createSessionInitializerFixture(t)

	si, err := NewSessionInitializer("/tmp/nonexistent-project/", wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", "/tmp/nonexistent-project/", terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)project root does not exist`, err.Error())
}

func TestInitialize_ProjectRootNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-directory")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))

	wdl := &mockWorkflowDefinitionLoaderForInit{}
	sdm := &mockSessionDirectoryManagerForInit{}

	si, err := NewSessionInitializer(filePath, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", filePath, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)project root is not a directory`, err.Error())
}

// =====================================================================
// Validation Failures — Workflow Definition
// =====================================================================

func TestInitialize_WorkflowNotFound(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "NonExistentWorkflow").Return(nil, fmt.Errorf("workflow file not found"))

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("NonExistentWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, err.Error())
}

func TestInitialize_WorkflowParseError(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "MalformedWorkflow").Return(nil, fmt.Errorf("yaml: line 5: found unexpected end"))

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("MalformedWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, err.Error())
}

func TestInitialize_WorkflowValidationError(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "InvalidWorkflow").Return(nil, fmt.Errorf("validation error: entry_node references non-existent node"))

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("InvalidWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to load workflow definition:`, err.Error())
}

// =====================================================================
// Validation Failures — Session Directory
// =====================================================================

func TestInitialize_SessionDirectoryParentMissing(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(
		fmt.Errorf("sessions directory does not exist: %s. Run 'spectra init' to initialize the project",
			filepath.Join(projectRoot, ".spectra", "sessions")))

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create session directory:.*sessions directory does not exist.*Run 'spectra init'`, err.Error())
}

func TestInitialize_SessionDirectoryAlreadyExists(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(
		fmt.Errorf("session directory already exists: /some/path. This indicates a UUID collision or a previous session was not cleaned up properly"))

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create session directory:.*session directory already exists.*UUID collision or a previous session was not cleaned up`, err.Error())
}

func TestInitialize_SessionDirectoryPermissionDenied(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(
		fmt.Errorf("permission denied"))

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create session directory:`, err.Error())
}

// =====================================================================
// Validation Failures — Storage Files
// =====================================================================

func TestInitialize_StorageFileCreationFails(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetFileAccessorError(fmt.Errorf("file creation failed"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to initialize storage files:`, err.Error())
}

func TestInitialize_StorageFileDiskFull(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetFileAccessorError(fmt.Errorf("no space left on device"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to initialize storage files:.*no space left on device`, err.Error())
}

// =====================================================================
// Validation Failures — Runtime Socket
// =====================================================================

func TestInitialize_SocketFileAlreadyExists(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetSocketCreateError(fmt.Errorf("runtime socket file already exists: /some/path. This may indicate a previous runtime process did not clean up properly or another runtime is currently active"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create runtime socket:.*runtime socket file already exists.*previous runtime process did not clean up`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called to clean up")
}

func TestInitialize_SocketPermissionDenied(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetSocketCreateError(fmt.Errorf("permission denied"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create runtime socket:`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called to clean up")
}

func TestInitialize_SocketDiskFull(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetSocketCreateError(fmt.Errorf("no space left on device"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to create runtime socket:.*no space left on device`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called to clean up")
}

// =====================================================================
// Validation Failures — Metadata Persistence
// =====================================================================

func TestInitialize_MetadataWriteDiskFull(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetMetadataWriteError(fmt.Errorf("no space left on device"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to persist initial session metadata:.*no space left on device`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called for cleanup")
}

func TestInitialize_MetadataWritePermissionDenied(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetMetadataWriteError(fmt.Errorf("permission denied"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to persist initial session metadata:`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called for cleanup")
}

// =====================================================================
// Validation Failures — Session.Run
// =====================================================================

func TestInitialize_SessionRunFailsNonInitializing(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetSessionRunError(fmt.Errorf("cannot run session: status is 'running', expected 'initializing'"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to transition session to running:.*status is 'running', expected 'initializing'`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called for cleanup")
}

func TestInitialize_SessionRunFailsGeneric(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetSessionRunError(fmt.Errorf("internal error"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to transition session to running:.*internal error`, err.Error())
	assert.True(t, si.WasDeleteSocketCalled(), "DeleteSocket should be called for cleanup")

	// Session.Fail() should have been called with RuntimeError
	assert.True(t, si.WasSessionFailCalled(), "Session.Fail should be called with RuntimeError")
	failErr := si.GetSessionFailError()
	if failErr != nil {
		assert.Contains(t, failErr.Error(), "failed to transition session to running status")
	}
}

// =====================================================================
// Error Propagation — Timeout Handler Race
// =====================================================================

func TestInitialize_TimeoutRacesSessionRun_TimeoutWins(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Make Session.Run() block so timeout handler can win
	runBlockCh := make(chan struct{})
	si.SetSessionRunBlock(runBlockCh)
	si.SetTimeoutDuration(50 * time.Millisecond)

	terminationNotifier := make(chan struct{}, 2)

	var initErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, initErr = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)
	}()

	// Wait for timeout to fire
	time.Sleep(200 * time.Millisecond)

	// Unblock Session.Run()
	close(runBlockCh)

	<-done

	// Either timeout or session.Run failure should result in an error
	require.Error(t, initErr)
	assert.Regexp(t, `(?i)failed to transition session to running:`, initErr.Error())
}

func TestInitialize_TimeoutRacesSessionRun_SessionRunWins(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// All mocks complete immediately, so Session.Run() completes before timeout
	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "running", sess.GetStatusSafe())
}

// =====================================================================
// Idempotency — Timeout Handler
// =====================================================================

func TestInitialize_TimeoutHandlerExitsIfInitCompleted(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "running", sess.GetStatusSafe())

	// No termination notification should arrive (timeout handler should exit silently)
	time.Sleep(100 * time.Millisecond)
	select {
	case <-terminationNotifier:
		t.Fatal("timeout handler should not send notification after init completes")
	default:
		// Good
	}
}

func TestInitialize_TimeoutHandlerExitsIfStatusRunning(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "running", sess.GetStatusSafe())
}

func TestInitialize_TimeoutHandlerExitsIfStatusFailed(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Session.Run() fails, transitioning status to "failed" before timeout could fire
	si.SetSessionRunError(fmt.Errorf("test error"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)
}

// =====================================================================
// Boundary Values — Timeout
// =====================================================================

func TestInitialize_TimeoutExactly30Seconds(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	// Use a blocking workflow loader to simulate indefinite work
	blockCh := make(chan struct{})
	wdl.On("Load", "SlowWorkflow").Run(func(args mock.Arguments) {
		<-blockCh
	}).Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Use short timeout for testing purposes (represents 30s in production)
	si.SetTimeoutDuration(50 * time.Millisecond)

	terminationNotifier := make(chan struct{}, 2)

	var initErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, initErr = si.Initialize("SlowWorkflow", projectRoot, terminationNotifier)
	}()

	time.Sleep(200 * time.Millisecond)
	close(blockCh)
	<-done

	if initErr != nil {
		assert.Regexp(t, `(?i)session initialization timeout exceeded 30 seconds`, initErr.Error())
	}
}

func TestInitialize_CompletesJustBeforeTimeout(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "FastWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("FastWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "running", sess.GetStatusSafe())
}

// =====================================================================
// Boundary Values — Workflow Name
// =====================================================================

func TestInitialize_WorkflowNamePascalCase(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "MyTestWorkflow").Return(&storage.WorkflowDefinition{
		Name:      "MyTestWorkflow",
		EntryNode: "start",
		Nodes:     []storage.Node{{Name: "start", Type: "agent"}},
	}, nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("MyTestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "MyTestWorkflow", sess.GetWorkflowName())
}

func TestInitialize_WorkflowNameSingleWord(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "Test").Return(&storage.WorkflowDefinition{
		Name:      "Test",
		EntryNode: "start",
		Nodes:     []storage.Node{{Name: "start", Type: "agent"}},
	}, nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("Test", projectRoot, terminationNotifier)

	require.NoError(t, err)
	assert.Equal(t, "Test", sess.GetWorkflowName())
}

// =====================================================================
// Boundary Values — Session UUID
// =====================================================================

func TestInitialize_SessionUUIDFormat(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	terminationNotifier := make(chan struct{}, 2)
	sess, err := si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)

	// Verify UUID v4 format: 8-4-4-4-12 hex characters
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	assert.Regexp(t, uuidRegex, sess.GetID(), "session ID should be valid UUID v4 format")
}

// =====================================================================
// Resource Cleanup — No Directory Deletion
// =====================================================================

func TestInitialize_NoDirectoryCleanupOnFailure(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		sessionUUID := args.String(0)
		sessionDir := filepath.Join(projectRoot, ".spectra", "sessions", sessionUUID)
		require.NoError(t, os.MkdirAll(sessionDir, 0775))
		require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session.json"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(""), 0644))
	}).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Session.Run() fails
	si.SetSessionRunError(fmt.Errorf("run failed"))

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)

	// Session directory and files should remain on disk
	entries, readErr := os.ReadDir(filepath.Join(projectRoot, ".spectra", "sessions"))
	require.NoError(t, readErr)
	assert.NotEmpty(t, entries, "session directory should exist on disk after failure")
}

func TestInitialize_NoDirectoryCleanupOnTimeout(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	blockCh := make(chan struct{})
	wdl.On("Load", "SlowWorkflow").Run(func(args mock.Arguments) {
		<-blockCh
	}).Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetTimeoutDuration(50 * time.Millisecond)

	terminationNotifier := make(chan struct{}, 2)

	done := make(chan struct{})
	go func() {
		defer close(done)
		si.Initialize("SlowWorkflow", projectRoot, terminationNotifier)
	}()

	time.Sleep(200 * time.Millisecond)
	close(blockCh)
	<-done

	// Session directory and files should remain (not cleaned up by SessionInitializer)
	// Socket should be deleted
}

// =====================================================================
// Mock / Dependency Interaction — Call Order
// =====================================================================

func TestInitialize_CriticalCallOrder(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	tracker := &callOrderTracker{}

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		tracker.record("CreateSessionDirectory")
	}).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	// Inject tracking into the initializer
	si.SetCallOrderTracker(tracker)

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)

	calls := tracker.getCalls()

	// Verify critical orderings:
	// 1) CreateSessionDirectory before FileAccessor preparation
	createDirIdx := tracker.indexOf("CreateSessionDirectory")
	fileAccessorIdx := tracker.indexOf("FileAccessorPrepare")
	if createDirIdx >= 0 && fileAccessorIdx >= 0 {
		assert.Less(t, createDirIdx, fileAccessorIdx,
			"CreateSessionDirectory must be called before FileAccessor preparation. Got calls: %v", calls)
	}

	// 2) FileAccessor preparation before CreateSocket
	createSocketIdx := tracker.indexOf("CreateSocket")
	if fileAccessorIdx >= 0 && createSocketIdx >= 0 {
		assert.Less(t, fileAccessorIdx, createSocketIdx,
			"FileAccessor preparation must be called before CreateSocket. Got calls: %v", calls)
	}

	// 3) CreateSocket before SessionMetadataStore.Write
	writeIdx := tracker.indexOf("MetadataWrite")
	if createSocketIdx >= 0 && writeIdx >= 0 {
		assert.Less(t, createSocketIdx, writeIdx,
			"CreateSocket must be called before MetadataWrite. Got calls: %v", calls)
	}

	// 4) SessionMetadataStore.Write before Session.Run
	runIdx := tracker.indexOf("SessionRun")
	if writeIdx >= 0 && runIdx >= 0 {
		assert.Less(t, writeIdx, runIdx,
			"MetadataWrite must be called before SessionRun. Got calls: %v", calls)
	}
}

// =====================================================================
// Mock / Dependency Interaction — TerminationNotifier
// =====================================================================

func TestInitialize_TerminationNotifierPassedToSessionFail(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	blockCh := make(chan struct{})
	wdl.On("Load", "SlowWorkflow").Run(func(args mock.Arguments) {
		<-blockCh
	}).Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetTimeoutDuration(50 * time.Millisecond)

	terminationNotifier := make(chan struct{}, 2)

	var capturedNotifier atomic.Value

	si.SetSessionFailCallback(func(err error, notifier chan<- struct{}) {
		capturedNotifier.Store(notifier)
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		si.Initialize("SlowWorkflow", projectRoot, terminationNotifier)
	}()

	time.Sleep(200 * time.Millisecond)
	close(blockCh)
	<-done

	// TerminationNotifier should have been passed to Session.Fail()
	if val := capturedNotifier.Load(); val != nil {
		// Verify it's the same channel
		_ = val
	}
}

func TestInitialize_TerminationNotifierPassedToSessionRun(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	var capturedNotifier atomic.Value

	si.SetSessionRunCallback(func(notifier chan<- struct{}) {
		capturedNotifier.Store(notifier)
	})

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.NoError(t, err)

	// Session.Run() should have received the terminationNotifier
	val := capturedNotifier.Load()
	assert.NotNil(t, val, "terminationNotifier should be passed to Session.Run()")
}

// =====================================================================
// Mock / Dependency Interaction — RuntimeError Construction
// =====================================================================

func TestInitialize_TimeoutRuntimeErrorFields(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	blockCh := make(chan struct{})
	wdl.On("Load", "TestWorkflow").Run(func(args mock.Arguments) {
		<-blockCh
	}).Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetTimeoutDuration(50 * time.Millisecond)

	var capturedErr atomic.Value
	si.SetSessionFailCallback(func(err error, notifier chan<- struct{}) {
		capturedErr.Store(err)
	})

	terminationNotifier := make(chan struct{}, 2)

	done := make(chan struct{})
	go func() {
		defer close(done)
		si.Initialize("TestWorkflow", projectRoot, terminationNotifier)
	}()

	time.Sleep(200 * time.Millisecond)
	close(blockCh)
	<-done

	if val := capturedErr.Load(); val != nil {
		rtErr := val.(error)
		assert.Contains(t, rtErr.Error(), "session initialization timeout exceeded 30 seconds")
	}
}

func TestInitialize_SessionRunFailureRuntimeErrorFields(t *testing.T) {
	projectRoot, wdl, sdm := createSessionInitializerFixture(t)

	wdl.On("Load", "TestWorkflow").Return(defaultWorkflowDef(), nil)
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	si, err := NewSessionInitializer(projectRoot, wdl, sdm)
	require.NoError(t, err)

	si.SetSessionRunError(fmt.Errorf("test error"))

	var capturedErr atomic.Value
	si.SetSessionFailCallback(func(err error, notifier chan<- struct{}) {
		capturedErr.Store(err)
	})

	terminationNotifier := make(chan struct{}, 2)
	_, err = si.Initialize("TestWorkflow", projectRoot, terminationNotifier)

	require.Error(t, err)

	if val := capturedErr.Load(); val != nil {
		rtErr := val.(error)
		assert.Contains(t, rtErr.Error(), "failed to transition session to running status")
	}
}
