# Test Specification: `session_initializer_test.go`

## Source File Under Test

`runtime/session_initializer.go`

## Test File

`runtime/session_initializer_test.go`

---

## `SessionInitializer`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewSessionInitializer_ValidDeps` | `unit` | Constructs SessionInitializer with all valid dependencies. | Create mock WorkflowDefinitionLoader, mock SessionDirectoryManager, and mock Logger. | `NewSessionInitializer("/project/root", loader, dirMgr, logger)` | Returns non-nil `*SessionInitializer`; no panic |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_Initialize_TerminationNotifierCapacityOne` | `unit` | Returns error when terminationNotifier has capacity 1. | Create mock dependencies. | `si.Initialize("workflow", "", make(chan struct{}, 1))` | Returns error `"terminationNotifier channel must have buffer capacity >= 2, got 1"` |
| `TestSessionInitializer_Initialize_TerminationNotifierNil` | `unit` | Returns error when terminationNotifier is nil. | Create mock dependencies. | `si.Initialize("workflow", "", nil)` | Returns error `"terminationNotifier channel must have buffer capacity >= 2, got 0"` |
| `TestSessionInitializer_Initialize_TerminationNotifierUnbuffered` | `unit` | Returns error when terminationNotifier is unbuffered. | Create mock dependencies. | `si.Initialize("workflow", "", make(chan struct{}))` | Returns error `"terminationNotifier channel must have buffer capacity >= 2, got 0"` |
| `TestSessionInitializer_Initialize_InvalidSessionID` | `unit` | Returns error when sessionID is not a valid UUID format. | Create mock dependencies. | `si.Initialize("workflow", "not-a-uuid", make(chan struct{}, 2))` | Returns error `"invalid session ID: must be a valid UUID"`; nil PersistentSession in InitResult |
| `TestSessionInitializer_Initialize_InvalidSessionID_NumericString` | `unit` | Returns error when sessionID is a numeric string (not UUID). | Create mock dependencies. | `si.Initialize("workflow", "12345", make(chan struct{}, 2))` | Returns error `"invalid session ID: must be a valid UUID"`; nil PersistentSession in InitResult |

### Happy Path — Initialize

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_Initialize_Success_GeneratedUUID` | `unit` | Successfully initializes session with auto-generated UUID and returns running PersistentSession. | Mock WorkflowDefinitionLoader.Load("my-wf") returns valid WorkflowDefinition with EntryNode()="start". Mock SessionDirectoryManager.CreateSessionDirectory returns nil. Mock Logger. Use fake clock if needed. | `si.Initialize("my-wf", "", make(chan struct{}, 2))` | Returns InitResult with non-nil PersistentSession (status="running"), non-nil WorkflowDefinition, nil Error; Logger.Info called with `"session created"`, `"sessionID"` containing a valid UUID, and `"source", "generated"` |
| `TestSessionInitializer_Initialize_Success_UserProvidedUUID` | `unit` | Successfully initializes session using user-provided UUID. | Mock WorkflowDefinitionLoader.Load("my-wf") returns valid WorkflowDefinition with EntryNode()="start". Mock SessionDirectoryManager.CreateSessionDirectory returns nil. Mock Logger. | `si.Initialize("my-wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))` | Returns InitResult with non-nil PersistentSession whose ID is `"550e8400-e29b-41d4-a716-446655440000"`, nil Error; Logger.Info called with `"session created"`, `"sessionID", "550e8400-e29b-41d4-a716-446655440000"`, and `"source", "user"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_Initialize_LogsSessionID_Generated` | `unit` | Logs session UUID with source "generated" immediately after generation before any I/O. | Mock all dependencies to succeed. Mock Logger records call order. | `si.Initialize("wf", "", make(chan struct{}, 2))` | `Logger.Info` called with `"session created"`, `"sessionID"` arg containing a valid UUID, and `"source", "generated"`, before any call to WorkflowDefinitionLoader or SessionDirectoryManager |
| `TestSessionInitializer_Initialize_LogsSessionID_User` | `unit` | Logs session UUID with source "user" when user-provided UUID is used. | Mock all dependencies to succeed. Mock Logger records call order. | `si.Initialize("wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))` | `Logger.Info` called with `"session created"`, `"sessionID", "550e8400-e29b-41d4-a716-446655440000"`, and `"source", "user"`, before any call to WorkflowDefinitionLoader or SessionDirectoryManager |
| `TestSessionInitializer_Initialize_CallsLoadWithWorkflowName` | `unit` | Passes workflow name to WorkflowDefinitionLoader.Load. | Mock WorkflowDefinitionLoader.Load records args. Other mocks succeed. | `si.Initialize("target-wf", "", make(chan struct{}, 2))` | `WorkflowDefinitionLoader.Load` called with `"target-wf"` |
| `TestSessionInitializer_Initialize_CallsCreateSessionDirectory` | `unit` | Passes projectRoot and generated UUID to SessionDirectoryManager. | Mock SessionDirectoryManager.CreateSessionDirectory records args. Other mocks succeed. | `si.Initialize("wf", "", make(chan struct{}, 2))` | `SessionDirectoryManager.CreateSessionDirectory` called with `"/project/root"` and a valid UUID string |
| `TestSessionInitializer_Initialize_CallsCreateSessionDirectoryWithUserUUID` | `unit` | Passes user-provided UUID to SessionDirectoryManager. | Mock SessionDirectoryManager.CreateSessionDirectory records args. Other mocks succeed. | `si.Initialize("wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))` | `SessionDirectoryManager.CreateSessionDirectory` called with `"/project/root"` and `"550e8400-e29b-41d4-a716-446655440000"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_Initialize_WorkflowLoadFails` | `unit` | Returns error when workflow definition loading fails. | Mock WorkflowDefinitionLoader.Load returns `errors.New("file not found")`. | `si.Initialize("bad-wf", "", make(chan struct{}, 2))` | Returns InitResult with Error containing `"failed to load workflow definition: file not found"`, nil PersistentSession, nil WorkflowDefinition |
| `TestSessionInitializer_Initialize_DirectoryCreationFails` | `unit` | Returns error when session directory creation fails. | Mock WorkflowDefinitionLoader.Load succeeds. Mock SessionDirectoryManager.CreateSessionDirectory returns `errors.New("permission denied")`. | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns InitResult with Error containing `"failed to create session directory: permission denied"`, nil PersistentSession |
| `TestSessionInitializer_Initialize_DirectoryExistsWithUserUUID` | `unit` | Returns error when user-provided UUID collides with existing session directory. | Mock WorkflowDefinitionLoader.Load succeeds. Mock SessionDirectoryManager.CreateSessionDirectory returns `ErrSessionDirExists`. | `si.Initialize("wf", "550e8400-e29b-41d4-a716-446655440000", make(chan struct{}, 2))` | Returns InitResult with Error containing `"failed to create session directory:"` wrapping `ErrSessionDirExists`, nil PersistentSession |
| `TestSessionInitializer_Initialize_SessionConstructionFails` | `unit` | Returns error when NewSession constructor fails. | Mock WorkflowDefinitionLoader.Load succeeds. Mock SessionDirectoryManager succeeds. Inject a NewSession constructor that returns error. | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns InitResult with Error containing `"failed to construct session: <error>"`, nil PersistentSession |
| `TestSessionInitializer_Initialize_RunFails` | `unit` | Constructs RuntimeError and calls Fail when PersistentSession.Run fails. | Mock all steps succeed until PersistentSession.Run() which returns an error. | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns InitResult with non-nil PersistentSession (status="failed"), Error containing `"failed to transition session to running"` |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_Initialize_TimeoutBeforePersistentSession` | `unit` | Returns timeout error without PersistentSession when context expires before construction. | Inject a fake context or a slow mock: Mock WorkflowDefinitionLoader.Load triggers context cancellation (e.g., use an already-cancelled context via injectable context factory or make Load block until timeout). | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns InitResult with Error containing `"session initialization timed out"`, nil PersistentSession |
| `TestSessionInitializer_Initialize_TimeoutAfterPersistentSession` | `unit` | Constructs RuntimeError and fails PersistentSession when context expires after construction. | Inject a context that is cancelled at the checkpoint after PersistentSession is constructed (step 18). Use a mock that cancels context at the right moment. | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns InitResult with non-nil PersistentSession (status="failed"), Error containing `"session initialization timed out"` |
| `TestSessionInitializer_Initialize_TimeoutAfterRunSucceeds` | `unit` | Fails PersistentSession when context expires after Run succeeds. | Mock all steps succeed. Cancel context at step 21 checkpoint (after Run). | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns InitResult with non-nil PersistentSession (status="failed"), Error containing `"session initialization timed out"` |
| `TestSessionInitializer_Initialize_RunFailsRuntimeErrorDetails` | `unit` | RuntimeError constructed with correct fields when Run fails. | Mock PersistentSession.Run() returns error. Capture the RuntimeError passed to PersistentSession.Fail. | `si.Initialize("wf", "", make(chan struct{}, 2))` | RuntimeError passed to Fail has `Issuer="SessionInitializer"`, Message contains `"failed to transition session to running"` |

### Boundary Values — TerminationNotifier

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_Initialize_TerminationNotifierCapacityTwo` | `unit` | Accepts terminationNotifier with exactly capacity 2. | Create mock dependencies that all succeed. | `si.Initialize("wf", "", make(chan struct{}, 2))` | Returns successful InitResult with nil Error |
| `TestSessionInitializer_Initialize_TerminationNotifierCapacityLarge` | `unit` | Accepts terminationNotifier with capacity greater than 2. | Create mock dependencies that all succeed. | `si.Initialize("wf", "", make(chan struct{}, 10))` | Returns successful InitResult with nil Error |
