package runtime

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
)

// =============================================================================
// Test Helpers — SessionInitializer
// =============================================================================
//
// Production surface expected in runtime/session_initializer.go:
//   - type SessionInitializer struct { ... }
//   - func NewSessionInitializer(projectRoot string, loader WorkflowLoader,
//       dirMgr SessionDirManager, logger logger.Logger) *SessionInitializer
//   - func (si *SessionInitializer) Initialize(workflowName string,
//       terminationNotifier chan<- struct{}) InitResult
//   - type InitResult struct {
//       PersistentSession *PersistentSession
//       WorkflowDefinition *components.WorkflowDefinition
//       Error error
//     }
//
// Interfaces expected (defined in session_initializer.go or a shared file):
//   - type WorkflowLoader interface { Load(workflowName string) (*components.WorkflowDefinition, error) }
//   - type SessionDirManager interface { CreateSessionDirectory(projectRoot, sessionUUID string) error }
//
// The SessionInitializer orchestrates session creation:
//   1. Validate terminationNotifier capacity >= 2
//   2. Generate UUID, log it
//   3. Load workflow definition
//   4. Create session directory
//   5. Construct Session, stores, PersistentSession
//   6. Call PersistentSession.Run()
//   7. Return InitResult
// =============================================================================

// --- Mock: WorkflowLoader ---

type mockWorkflowLoader struct {
	mu         sync.Mutex
	loadCalled int
	loadInput  string
	loadResult *components.WorkflowDefinition
	loadErr    error
	// onLoad is an optional hook called during Load (for context cancellation tests)
	onLoad func()
}

func (m *mockWorkflowLoader) Load(workflowName string) (*components.WorkflowDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadCalled++
	m.loadInput = workflowName
	if m.onLoad != nil {
		m.onLoad()
	}
	return m.loadResult, m.loadErr
}

// --- Mock: SessionDirManager ---

type mockSessionDirManager struct {
	mu                          sync.Mutex
	createSessionDirCalled      int
	createSessionDirProjectRoot string
	createSessionDirUUID        string
	createSessionDirErr         error
}

func (m *mockSessionDirManager) CreateSessionDirectory(projectRoot, sessionUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createSessionDirCalled++
	m.createSessionDirProjectRoot = projectRoot
	m.createSessionDirUUID = sessionUUID
	return m.createSessionDirErr
}

// --- Fixture Builder: SessionInitializer ---

type sessionInitializerFixture struct {
	projectRoot string
	loader      *mockWorkflowLoader
	dirMgr      *mockSessionDirManager
	logger      *mockLogger
}

func newSessionInitializerFixture(t *testing.T) *sessionInitializerFixture {
	t.Helper()
	return &sessionInitializerFixture{
		projectRoot: "/project/root",
		loader:      &mockWorkflowLoader{},
		dirMgr:      &mockSessionDirManager{},
		logger:      newDefaultMockLogger(),
	}
}

// withWorkflowLoaderSuccess configures the loader to return a valid WorkflowDefinition.
func (f *sessionInitializerFixture) withWorkflowLoaderSuccess(t *testing.T) *sessionInitializerFixture {
	t.Helper()
	wfDef := mustNewWorkflowDefinition(t)
	f.loader.loadResult = wfDef
	f.loader.loadErr = nil
	return f
}

// withDirManagerSuccess configures the directory manager to succeed.
func (f *sessionInitializerFixture) withDirManagerSuccess() *sessionInitializerFixture {
	f.dirMgr.createSessionDirErr = nil
	return f
}

// mustNewWorkflowDefinition creates a minimal valid WorkflowDefinition for tests.
func mustNewWorkflowDefinition(t *testing.T) *components.WorkflowDefinition {
	t.Helper()
	// Minimal valid workflow: one human entry node, one agent node, one transition, one exit transition
	entryNode, err := components.NewNode("Start", "human", "", "Entry node")
	require.NoError(t, err)
	agentNode, err := components.NewNode("Worker", "agent", "Coder", "Agent node")
	require.NoError(t, err)
	transition, err := components.NewTransition("Start", "TaskAssigned", "Worker")
	require.NoError(t, err)
	backTransition, err := components.NewTransition("Worker", "TaskCompleted", "Start")
	require.NoError(t, err)
	exitTrans, err := components.NewExitTransition("Worker", "TaskCompleted", "Start")
	require.NoError(t, err)

	wfDef, err := components.NewWorkflowDefinition(
		"TestWorkflow",
		"A test workflow",
		"Start",
		[]*components.Node{entryNode, agentNode},
		[]*components.Transition{transition, backTransition},
		[]*components.ExitTransition{exitTrans},
	)
	require.NoError(t, err)
	return wfDef
}

// --- Assertion Helpers: UUID validation ---

// isValidUUID checks if a string matches UUID v4 format.
func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewSessionInitializer_ValidDeps(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer does not exist yet")

	f := newSessionInitializerFixture(t)

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)

	// Assert
	require.NotNil(t, si)
}

// =============================================================================
// Validation Failures
// =============================================================================

func TestSessionInitializer_Initialize_TerminationNotifierCapacityOne(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", make(chan struct{}, 1))

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 1", result.Error.Error())
}

func TestSessionInitializer_Initialize_TerminationNotifierNil(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", nil)

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 0", result.Error.Error())
}

func TestSessionInitializer_Initialize_TerminationNotifierUnbuffered(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", make(chan struct{}))

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 0", result.Error.Error())
}

// =============================================================================
// Happy Path — Initialize
// =============================================================================

func TestSessionInitializer_Initialize_Success(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("my-wf", make(chan struct{}, 2))

	// Assert
	require.NoError(t, result.Error)
	require.NotNil(t, result.PersistentSession)
	require.NotNil(t, result.WorkflowDefinition)
	assert.Equal(t, "running", result.PersistentSession.GetStatusSafe())
	// Logger.Info called with "session created" and a valid sessionID
	assertLogHasMessage(t, f.logger.infoCalls, "session created")
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestSessionInitializer_Initialize_LogsSessionID(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs call-order tracking seam")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("wf", make(chan struct{}, 2))

	// Assert: Logger.Info called with "session created" containing a valid UUID
	require.NotEmpty(t, f.logger.infoCalls)
	found := false
	for _, call := range f.logger.infoCalls {
		if call.msg == "session created" {
			// Find "sessionID" arg
			for i := 0; i+1 < len(call.args); i += 2 {
				if call.args[i] == "sessionID" {
					sid, ok := call.args[i+1].(string)
					require.True(t, ok, "sessionID arg should be string")
					assert.True(t, isValidUUID(sid), "sessionID should be valid UUID, got: %s", sid)
					found = true
				}
			}
		}
	}
	assert.True(t, found, "expected Logger.Info called with 'session created' and sessionID")

	// Assert: logged before any call to WorkflowDefinitionLoader or SessionDirectoryManager
	// This requires call-order tracking which the current mocks partially support
	// via the loadCalled count being 1 (meaning Load was called after the log).
	assert.Equal(t, 1, f.loader.loadCalled)
}

func TestSessionInitializer_Initialize_CallsLoadWithWorkflowName(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("target-wf", make(chan struct{}, 2))

	// Assert
	assert.Equal(t, 1, f.loader.loadCalled)
	assert.Equal(t, "target-wf", f.loader.loadInput)
}

func TestSessionInitializer_Initialize_CallsCreateSessionDirectory(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("wf", make(chan struct{}, 2))

	// Assert
	assert.Equal(t, 1, f.dirMgr.createSessionDirCalled)
	assert.Equal(t, "/project/root", f.dirMgr.createSessionDirProjectRoot)
	assert.True(t, isValidUUID(f.dirMgr.createSessionDirUUID), "sessionUUID passed to CreateSessionDirectory should be valid UUID")
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestSessionInitializer_Initialize_WorkflowLoadFails(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t)
	f.loader.loadErr = errors.New("file not found")
	f.loader.loadResult = nil

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("bad-wf", make(chan struct{}, 2))

	// Assert
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to load workflow definition: file not found")
	assert.Nil(t, result.PersistentSession)
	assert.Nil(t, result.WorkflowDefinition)
}

func TestSessionInitializer_Initialize_DirectoryCreationFails(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t)
	f.dirMgr.createSessionDirErr = errors.New("permission denied")

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to create session directory: permission denied")
	assert.Nil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_SessionConstructionFails(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs injectable NewSession constructor seam")

	// This test requires that the SessionInitializer uses an injectable
	// NewSession constructor (or a seam that can be replaced in tests).
	// Since the production surface does not yet exist, we scaffold the intent.

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	// TODO: inject a failing NewSession constructor when seam is available
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert (placeholder — will be refined once seam is known)
	_ = result
}

func TestSessionInitializer_Initialize_RunFails(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs injectable PersistentSession or Session.Run() failure seam")

	// When PersistentSession.Run() fails, SessionInitializer should construct
	// a RuntimeError and call PersistentSession.Fail, then return error with
	// the failed PersistentSession.

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert (placeholder)
	_ = result
}

// =============================================================================
// State Transitions
// =============================================================================

func TestSessionInitializer_Initialize_TimeoutBeforePersistentSession(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs injectable context factory or context cancellation seam")

	// Scenario: context expires before PersistentSession is constructed.
	// This requires either:
	//   - An injectable context factory in SessionInitializer
	//   - A mock WorkflowDefinitionLoader.Load that triggers context cancellation
	//   - An already-cancelled context passed via some mechanism

	f := newSessionInitializerFixture(t)
	// Configure loader to "block" (simulate timeout trigger)
	// f.loader.onLoad = func() { /* cancel context somehow */ }

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert (placeholder — needs context seam)
	_ = result
}

func TestSessionInitializer_Initialize_TimeoutAfterPersistentSession(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs context cancellation seam at step 18 checkpoint")

	// Scenario: context expires at step 18 checkpoint (after PersistentSession construction).

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert (placeholder)
	_ = result
}

func TestSessionInitializer_Initialize_TimeoutAfterRunSucceeds(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs context cancellation seam at step 21 checkpoint")

	// Scenario: context expires at step 21 checkpoint (after Run succeeds).

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert (placeholder)
	_ = result
}

func TestSessionInitializer_Initialize_RunFailsRuntimeErrorDetails(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet; needs seam to capture RuntimeError passed to Fail")

	// Scenario: Verify the RuntimeError constructed when Run fails has correct fields.

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert (placeholder — needs to inspect RuntimeError passed to Fail)
	_ = result
}

// =============================================================================
// Boundary Values — TerminationNotifier
// =============================================================================

func TestSessionInitializer_Initialize_TerminationNotifierCapacityTwo(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 2))

	// Assert: succeeds with capacity exactly 2
	require.NoError(t, result.Error)
}

func TestSessionInitializer_Initialize_TerminationNotifierCapacityLarge(t *testing.T) {
	t.Skip("scaffolded: production surface NewSessionInitializer/Initialize does not exist yet")

	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", make(chan struct{}, 10))

	// Assert: succeeds with capacity > 2
	require.NoError(t, result.Error)
}
