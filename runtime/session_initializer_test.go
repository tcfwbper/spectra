package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/storage"
)

// =============================================================================
// Test Helpers — SessionInitializer
// =============================================================================
//
// Production surface expected in runtime/session_initializer.go:
//   - type SessionInitializer struct { ... }
//   - func NewSessionInitializer(projectRoot string, loader WorkflowLoader,
//       dirMgr SessionDirManager, logger logger.Logger) *SessionInitializer
//   - func (si *SessionInitializer) Initialize(workflowName string, sessionID string,
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
//   2. Determine session UUID (validate user-provided or generate new)
//   3. Log session UUID with source ("user" or "generated")
//   4. Load workflow definition
//   5. Create session directory
//   6. Construct Session, stores, PersistentSession
//   7. Call PersistentSession.Run()
//   8. Return InitResult
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
// isValidUUID is now defined in session_initializer.go (production code).

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewSessionInitializer_ValidDeps(t *testing.T) {
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
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", "", make(chan struct{}, 1))

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 1", result.Error.Error())
}

func TestSessionInitializer_Initialize_TerminationNotifierNil(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", "", nil)

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 0", result.Error.Error())
}

func TestSessionInitializer_Initialize_TerminationNotifierUnbuffered(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", "", make(chan struct{}))

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "terminationNotifier channel must have buffer capacity >= 2, got 0", result.Error.Error())
}

func TestSessionInitializer_Initialize_InvalidSessionID(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", "not-a-uuid", make(chan struct{}, 2))

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "invalid session ID: must be a valid UUID", result.Error.Error())
	assert.Nil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_InvalidSessionID_NumericString(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("workflow", "12345", make(chan struct{}, 2))

	// Assert
	require.Error(t, result.Error)
	assert.Equal(t, "invalid session ID: must be a valid UUID", result.Error.Error())
	assert.Nil(t, result.PersistentSession)
}

// =============================================================================
// Happy Path — Initialize
// =============================================================================

func TestSessionInitializer_Initialize_Success_GeneratedUUID(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("my-wf", "", make(chan struct{}, 2))

	// Assert
	require.NoError(t, result.Error)
	require.NotNil(t, result.PersistentSession)
	require.NotNil(t, result.WorkflowDefinition)
	assert.Equal(t, "running", result.PersistentSession.GetStatusSafe())
	// Logger.Info called with "session created", valid sessionID, and source "generated"
	assertLogHasMessage(t, f.logger.infoCalls, "session created")
	assertSessionCreatedLogWithSource(t, f.logger.infoCalls, "generated")
}

func TestSessionInitializer_Initialize_Success_UserProvidedUUID(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("my-wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))

	// Assert
	require.NoError(t, result.Error)
	require.NotNil(t, result.PersistentSession)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result.PersistentSession.ID)
	// Logger.Info called with "session created", the user-provided sessionID, and source "user"
	assertSessionCreatedLogWithSessionID(t, f.logger.infoCalls, "550e8400-e29b-41d4-a716-446655440000")
	assertSessionCreatedLogWithSource(t, f.logger.infoCalls, "user")
}

// =============================================================================
// Mock / Dependency Interaction
// =============================================================================

func TestSessionInitializer_Initialize_LogsSessionID_Generated(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: Logger.Info called with "session created" containing a valid UUID and source "generated"
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
	assertSessionCreatedLogWithSource(t, f.logger.infoCalls, "generated")

	// Assert: logged before any call to WorkflowDefinitionLoader or SessionDirectoryManager
	assert.Equal(t, 1, f.loader.loadCalled)
}

func TestSessionInitializer_Initialize_LogsSessionID_User(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))

	// Assert: Logger.Info called with "session created", the user UUID, and source "user"
	assertSessionCreatedLogWithSessionID(t, f.logger.infoCalls, "550e8400-e29b-41d4-a716-446655440000")
	assertSessionCreatedLogWithSource(t, f.logger.infoCalls, "user")

	// Assert: logged before any call to WorkflowDefinitionLoader or SessionDirectoryManager
	assert.Equal(t, 1, f.loader.loadCalled)
}

func TestSessionInitializer_Initialize_CallsLoadWithWorkflowName(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("target-wf", "", make(chan struct{}, 2))

	// Assert
	assert.Equal(t, 1, f.loader.loadCalled)
	assert.Equal(t, "target-wf", f.loader.loadInput)
}

func TestSessionInitializer_Initialize_CallsCreateSessionDirectory(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert
	assert.Equal(t, 1, f.dirMgr.createSessionDirCalled)
	assert.Equal(t, "/project/root", f.dirMgr.createSessionDirProjectRoot)
	assert.True(t, isValidUUID(f.dirMgr.createSessionDirUUID), "sessionUUID passed to CreateSessionDirectory should be valid UUID")
}

func TestSessionInitializer_Initialize_CallsCreateSessionDirectoryWithUserUUID(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	_ = si.Initialize("wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))

	// Assert: user-provided UUID is passed to SessionDirectoryManager
	assert.Equal(t, 1, f.dirMgr.createSessionDirCalled)
	assert.Equal(t, "/project/root", f.dirMgr.createSessionDirProjectRoot)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", f.dirMgr.createSessionDirUUID)
}

// =============================================================================
// Error Propagation
// =============================================================================

func TestSessionInitializer_Initialize_WorkflowLoadFails(t *testing.T) {
	f := newSessionInitializerFixture(t)
	f.loader.loadErr = errors.New("file not found")
	f.loader.loadResult = nil

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("bad-wf", "", make(chan struct{}, 2))

	// Assert
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to load workflow definition: file not found")
	assert.Nil(t, result.PersistentSession)
	assert.Nil(t, result.WorkflowDefinition)
}

func TestSessionInitializer_Initialize_DirectoryCreationFails(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t)
	f.dirMgr.createSessionDirErr = errors.New("permission denied")

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to create session directory: permission denied")
	assert.Nil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_DirectoryExistsWithUserUUID(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t)
	f.dirMgr.createSessionDirErr = storage.ErrSessionDirExists

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))

	// Assert: error wraps ErrSessionDirExists
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to create session directory:")
	assert.ErrorIs(t, result.Error, storage.ErrSessionDirExists)
	assert.Nil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_SessionConstructionFails(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	// Inject a failing session factory.
	si.sessionFactory = func(id, workflowName, entryNode string, createdAt int64) (Session, error) {
		return nil, errors.New("invalid session parameters")
	}
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: returns wrapped error, no PersistentSession created.
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to construct session: invalid session parameters")
	assert.Nil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_RunFails(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	// Inject a session factory that returns a mock session whose Run() fails.
	si.sessionFactory = func(id, workflowName, entryNode string, createdAt int64) (Session, error) {
		ms := newDefaultMockSession()
		ms.id = id
		ms.workflowName = workflowName
		ms.runErr = errors.New("status not initializing")
		ms.getStatusResult = "initializing"
		return ms, nil
	}
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: error returned, PersistentSession non-nil (failed), Fail was called.
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to transition session to running: status not initializing")
	require.NotNil(t, result.PersistentSession)
}

// =============================================================================
// State Transitions
// =============================================================================

func TestSessionInitializer_Initialize_TimeoutBeforePersistentSession(t *testing.T) {
	// Scenario: context expires before PersistentSession is constructed.
	// We inject a context factory that returns an already-cancelled context so
	// the first ctx.Err() check (step 6) detects cancellation.
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	// Inject a context that is already cancelled.
	si.contextFactory = func() (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // already cancelled
		return ctx, cancel
	}
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: timeout error, no PersistentSession.
	require.Error(t, result.Error)
	assert.Equal(t, "session initialization timed out", result.Error.Error())
	assert.Nil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_TimeoutAfterPersistentSession(t *testing.T) {
	// Scenario: context expires at step 18 checkpoint (after PersistentSession construction).
	// We use a session factory that cancels the context when called, so by step 18
	// the context is cancelled.
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)

	var cancel context.CancelFunc
	si.contextFactory = func() (context.Context, context.CancelFunc) {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		return ctx, cancel
	}
	// Inject a session factory that cancels context after returning (simulates
	// timeout occurring between session construction and step 18 check).
	si.sessionFactory = func(id, workflowName, entryNode string, createdAt int64) (Session, error) {
		ms := newDefaultMockSession()
		ms.id = id
		ms.workflowName = workflowName
		ms.getStatusResult = "initializing"
		// Cancel context to trigger timeout at step 18.
		cancel()
		return ms, nil
	}
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: timeout error with PersistentSession present (failed).
	require.Error(t, result.Error)
	assert.Equal(t, "session initialization timed out", result.Error.Error())
	require.NotNil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_TimeoutAfterRunSucceeds(t *testing.T) {
	// Scenario: context expires at step 21 checkpoint (after Run succeeds).
	// We inject a session whose Run() succeeds but cancels the context so that
	// the step 21 check detects cancellation.
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)

	var cancel context.CancelFunc
	si.contextFactory = func() (context.Context, context.CancelFunc) {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		return ctx, cancel
	}
	// Inject a session factory whose Run() cancels the context on success.
	si.sessionFactory = func(id, workflowName, entryNode string, createdAt int64) (Session, error) {
		ms := newDefaultMockSession()
		ms.id = id
		ms.workflowName = workflowName
		ms.getStatusResult = "initializing"
		ms.runErr = nil
		return &cancelOnRunSession{mockSession: ms, cancelFn: &cancel}, nil
	}
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: timeout error with PersistentSession present (failed after Run succeeded).
	require.Error(t, result.Error)
	assert.Equal(t, "session initialization timed out", result.Error.Error())
	require.NotNil(t, result.PersistentSession)
}

func TestSessionInitializer_Initialize_RunFailsRuntimeErrorDetails(t *testing.T) {
	// Scenario: Verify the RuntimeError constructed when Run fails has correct fields.
	// We use a mock session that records the error passed to Fail.
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	var capturedFailErr error
	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	si.sessionFactory = func(id, workflowName, entryNode string, createdAt int64) (Session, error) {
		ms := newDefaultMockSession()
		ms.id = id
		ms.workflowName = workflowName
		ms.getStatusResult = "initializing"
		ms.runErr = errors.New("cannot transition")
		return &capturingFailSession{mockSession: ms, captured: &capturedFailErr}, nil
	}
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: error returned with correct message.
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to transition session to running: cannot transition")
	require.NotNil(t, result.PersistentSession)

	// Assert: RuntimeError passed to Fail has correct fields.
	require.NotNil(t, capturedFailErr)
	rtErr, ok := capturedFailErr.(*entities.RuntimeError)
	require.True(t, ok, "error passed to Fail should be *entities.RuntimeError")
	assert.Equal(t, "SessionInitializer", rtErr.Issuer())
	assert.Contains(t, rtErr.Message(), "failed to transition session to running: cannot transition")
	assert.True(t, isValidUUID(rtErr.SessionID()), "RuntimeError.SessionID should be valid UUID")
}

// =============================================================================
// Boundary Values — TerminationNotifier
// =============================================================================

func TestSessionInitializer_Initialize_TerminationNotifierCapacityTwo(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", "", make(chan struct{}, 2))

	// Assert: succeeds with capacity exactly 2
	require.NoError(t, result.Error)
}

func TestSessionInitializer_Initialize_TerminationNotifierCapacityLarge(t *testing.T) {
	f := newSessionInitializerFixture(t).withWorkflowLoaderSuccess(t).withDirManagerSuccess()

	// Act
	si := NewSessionInitializer(f.projectRoot, f.loader, f.dirMgr, f.logger)
	result := si.Initialize("wf", "", make(chan struct{}, 10))

	// Assert: succeeds with capacity > 2
	require.NoError(t, result.Error)
}

// =============================================================================
// Test Helper Types — SessionInitializer
// =============================================================================

// cancelOnRunSession wraps a mockSession and cancels the context when Run()
// succeeds. This simulates a timeout occurring between Run() success and step 21.
type cancelOnRunSession struct {
	*mockSession
	cancelFn *context.CancelFunc
}

func (c *cancelOnRunSession) Run() error {
	err := c.mockSession.Run()
	if err == nil && c.cancelFn != nil && *c.cancelFn != nil {
		(*c.cancelFn)()
	}
	return err
}

// capturingFailSession wraps a mockSession and captures the error passed to Fail.
type capturingFailSession struct {
	*mockSession
	captured *error
}

func (c *capturingFailSession) Fail(err error, notifier chan<- struct{}) error {
	if c.captured != nil {
		*c.captured = err
	}
	return c.mockSession.Fail(err, notifier)
}
