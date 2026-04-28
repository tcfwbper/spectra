# Test Specification: `session_initializer.go`

## Source File Under Test
`runtime/session_initializer.go`

## Test File
`runtime/session_initializer_test.go`

---

## `SessionInitializer`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSessionInitializer_New` | `unit` | Constructs SessionInitializer with valid dependencies. | Test fixture creates temporary project directory structure with `.spectra/` and `.spectra/sessions/` directories | `ProjectRoot=<temp-dir>`, `WorkflowDefinitionLoader=<mock>`, `SessionDirectoryManager=<mock>` | Returns SessionInitializer instance; no error |

### Happy Path — Initialize

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_Success` | `unit` | Initializes session successfully with all resources created. | Test fixture creates temporary directory with `.spectra/` structure; mock WorkflowDefinitionLoader returns valid workflow with `EntryNode="start"`; mock SessionDirectoryManager creates directory; mock stores and socket manager succeed | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<buffered-channel-cap-2>` | Returns Session with `Status="running"`, `WorkflowName="TestWorkflow"`, `CurrentState="start"`, `EventHistory=[]`, `SessionData={}`; session directory, files, and socket created |
| `TestInitialize_MetadataPersisted` | `unit` | Persists session metadata with Status="initializing" before calling Session.Run(). | Test fixture; mock stores capture call order | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | SessionMetadataStore.Write() called before Session.Run(); metadata contains `Status="initializing"`; after Session.Run(), status transitions to "running" |
| `TestInitialize_EmptyEventHistoryAndSessionData` | `unit` | Initializes session with empty EventHistory and SessionData. | Test fixture; valid mocks | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Returns Session with `EventHistory=[]` (empty slice), `SessionData={}` (empty map) |
| `TestInitialize_CurrentStateSetToEntryNode` | `unit` | Sets CurrentState to workflow EntryNode. | Test fixture; mock workflow with `EntryNode="custom_entry"` | `WorkflowName="CustomWorkflow"`, `TerminationNotifier=<channel>` | Returns Session with `CurrentState="custom_entry"` |
| `TestInitialize_TimestampsSet` | `unit` | Sets CreatedAt and UpdatedAt timestamps. | Test fixture; valid mocks | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Returns Session with `CreatedAt` and `UpdatedAt` set to current POSIX timestamp (within 1 second tolerance) |
| `TestInitialize_UniqueSessionUUID` | `unit` | Generates unique session UUID. | Test fixture; call Initialize twice with same workflow | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` (called twice) | Each call returns Session with different UUID |

### Happy Path — Timeout Completion

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TimeoutTimerCanceled` | `unit` | Cancels timeout timer on successful initialization. | Test fixture; mock time.AfterFunc; track timer.Stop() calls | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Initialize completes successfully; timer.Stop() called; timeout handler does not fire |
| `TestInitialize_CompletionRace_InitWins` | `unit` | Handles race where initialization completes before timeout fires. | Test fixture; mock timer with 30-second duration; manually control timing to have initialization complete before timeout trigger | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Returns Session with `Status="running"`; timeout handler observes `initCompleted=true` and exits silently |

### Happy Path — Resource Cleanup on Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_SocketCleanupOnSessionRunFailure` | `unit` | Cleans up socket when Session.Run() fails. | Test fixture; mock Session.Run() to return error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | RuntimeSocketManager.DeleteSocket() called; Session transitioned to "failed"; returns error matching `/failed to transition session to running:/i` |
| `TestInitialize_SocketCleanupOnMetadataWriteFailure` | `unit` | Cleans up socket when SessionMetadataStore.Write() fails. | Test fixture; mock SessionMetadataStore.Write() returns error (disk full) | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to persist initial session metadata:/i` |
| `TestInitialize_SocketCleanupOnSocketCreateFailure` | `unit` | Cleans up partial socket on RuntimeSocketManager.CreateSocket() failure. | Test fixture; mock RuntimeSocketManager.CreateSocket() returns error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to create runtime socket:/i` |

### Happy Path — Timeout Enforcement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TimeoutAfterSessionConstructed` | `unit` | Timeout handler calls Session.Fail when Session is constructed but initialization not complete. | Test fixture; mock timer with 30-second duration; mock WorkflowDefinitionLoader that blocks; manually trigger timeout after Session constructed but before Session.Run() completes | `WorkflowName="SlowWorkflow"`, `TerminationNotifier=<channel>` | Timeout handler fires; Session is constructed and Status="initializing"; handler calls Session.Fail with RuntimeError `Issuer="SessionInitializer"`, `Message="session initialization timeout exceeded 30 seconds"`; TerminationNotifier receives signal |
| `TestInitialize_TimeoutBeforeSessionConstructed` | `unit` | Timeout handler sets timedOutEarly flag when Session not yet constructed. | Test fixture; mock timer with 30-second duration; mock SpectraFinder that blocks; manually trigger timeout before Session construction step | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout handler fires; Session reference is nil; handler sets `timedOutEarly=true`; sends one notification to TerminationNotifier; SessionInitializer observes flag at next checkpoint and returns error matching `/session initialization timeout exceeded 30 seconds before session entity was constructed/i` |

### Validation Failures — TerminationNotifier

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TerminationNotifierBufferCapacity1` | `unit` | Rejects channel with buffer capacity 1. | Test fixture | `WorkflowName="TestWorkflow"`, `TerminationNotifier=make(chan struct{}, 1)` | Returns error: `"terminationNotifier channel must have buffer capacity >= 2, got 1"` |
| `TestInitialize_TerminationNotifierNil` | `unit` | Rejects nil channel. | Test fixture | `WorkflowName="TestWorkflow"`, `TerminationNotifier=nil` | Returns error: `"terminationNotifier channel must have buffer capacity >= 2, got 0"` |
| `TestInitialize_TerminationNotifierUnbuffered` | `unit` | Rejects unbuffered channel. | Test fixture | `WorkflowName="TestWorkflow"`, `TerminationNotifier=make(chan struct{})` | Returns error: `"terminationNotifier channel must have buffer capacity >= 2, got 0"` |
| `TestInitialize_TerminationNotifierBufferCapacity2` | `unit` | Accepts channel with buffer capacity 2. | Test fixture; valid mocks | `WorkflowName="TestWorkflow"`, `TerminationNotifier=make(chan struct{}, 2)` | Returns Session successfully; no channel validation error |
| `TestInitialize_TerminationNotifierBufferCapacity5` | `unit` | Accepts channel with buffer capacity greater than 2. | Test fixture; valid mocks | `WorkflowName="TestWorkflow"`, `TerminationNotifier=make(chan struct{}, 5)` | Returns Session successfully; no channel validation error |

### Validation Failures — Project Root

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_ProjectRootNotFound` | `unit` | Returns error when SpectraFinder fails to find project root. | Test fixture creates temporary directory WITHOUT `.spectra/`; mock SpectraFinder returns error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to find project root:.*Run 'spectra init' to initialize the project/i` |

### Validation Failures — Workflow Definition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_WorkflowNotFound` | `unit` | Returns error when workflow definition file not found. | Test fixture; mock WorkflowDefinitionLoader returns file not found error | `WorkflowName="NonExistentWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to load workflow definition:/i` |
| `TestInitialize_WorkflowParseError` | `unit` | Returns error when workflow definition has parse error. | Test fixture; mock WorkflowDefinitionLoader returns YAML parse error | `WorkflowName="MalformedWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to load workflow definition:/i` |
| `TestInitialize_WorkflowValidationError` | `unit` | Returns error when workflow definition fails validation. | Test fixture; mock WorkflowDefinitionLoader returns validation error | `WorkflowName="InvalidWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to load workflow definition:/i` |

### Validation Failures — Session Directory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_SessionDirectoryParentMissing` | `unit` | Returns error when `.spectra/sessions/` directory does not exist. | Test fixture creates `.spectra/` but NOT `.spectra/sessions/`; mock SessionDirectoryManager returns error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to create session directory:.*sessions directory does not exist.*Run 'spectra init'/i` |
| `TestInitialize_SessionDirectoryAlreadyExists` | `unit` | Returns error when session directory already exists (UUID collision). | Test fixture; mock SessionDirectoryManager returns "directory already exists" error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to create session directory:.*session directory already exists.*UUID collision or previous session was not cleaned up/i` |
| `TestInitialize_SessionDirectoryPermissionDenied` | `unit` | Returns error when permission denied creating session directory. | Test fixture creates `.spectra/sessions/` with read-only permissions; mock SessionDirectoryManager returns permission error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to create session directory:/i` |

### Validation Failures — Storage Files

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_StorageFileCreationFails` | `unit` | Returns error when EventStore or SessionMetadataStore file creation fails. | Test fixture; mock FileAccessor preparation callback returns error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to initialize storage files:/i` |
| `TestInitialize_StorageFileDiskFull` | `unit` | Returns error when disk is full during file creation. | Test fixture; mock FileAccessor returns "no space left on device" error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; returns error matching `/failed to initialize storage files:.*no space left on device/i` |

### Validation Failures — Runtime Socket

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_SocketFileAlreadyExists` | `unit` | Returns error when socket file already exists. | Test fixture; mock RuntimeSocketManager.CreateSocket() returns "file already exists" error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to create runtime socket:.*runtime socket file already exists.*previous runtime process did not clean up/i` |
| `TestInitialize_SocketPermissionDenied` | `unit` | Returns error when permission denied creating socket. | Test fixture; mock RuntimeSocketManager.CreateSocket() returns permission error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to create runtime socket:/i` |
| `TestInitialize_SocketDiskFull` | `unit` | Returns error when disk is full creating socket. | Test fixture; mock RuntimeSocketManager.CreateSocket() returns "no space left on device" error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to create runtime socket:.*no space left on device/i` |

### Validation Failures — Metadata Persistence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_MetadataWriteDiskFull` | `unit` | Returns error when SessionMetadataStore.Write() fails due to disk full. | Test fixture; mock SessionMetadataStore.Write() returns "no space left on device" error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to persist initial session metadata:.*no space left on device/i` |
| `TestInitialize_MetadataWritePermissionDenied` | `unit` | Returns error when SessionMetadataStore.Write() fails due to permission denied. | Test fixture; mock SessionMetadataStore.Write() returns permission error | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; returns error matching `/failed to persist initial session metadata:/i` |

### Validation Failures — Session.Run

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_SessionRunFailsNonInitializing` | `unit` | Returns error when Session.Run() fails because Status is not "initializing". | Test fixture; mock Session.Run() returns error "cannot run session: status is 'running', expected 'initializing'" | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; Session.Fail() called with RuntimeError; returns error matching `/failed to transition session to running:.*status is 'running', expected 'initializing'/i` |
| `TestInitialize_SessionRunFailsGeneric` | `unit` | Returns error when Session.Run() fails for generic reason. | Test fixture; mock Session.Run() returns error "internal error" | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout timer canceled; RuntimeSocketManager.DeleteSocket() called; Session.Fail() called with RuntimeError `Issuer="SessionInitializer"`, `Message="failed to transition session to running status"`; returns error matching `/failed to transition session to running:.*internal error/i` |

### Error Propagation — Timeout Handler Race

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TimeoutRacesSessionRun_TimeoutWins` | `unit` | Handles race where timeout handler transitions to "failed" before Session.Run(). | Test fixture; mock timer; Session.Run() set to block; manually trigger timeout before Session.Run() completes; use goroutine synchronization to control race | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Timeout handler calls Session.Fail(), transitions to "failed"; Session.Run() returns error "cannot run session: status is 'failed'"; SessionInitializer calls Session.Fail() again (no-op or error); returns error matching `/failed to transition session to running:/i` |
| `TestInitialize_TimeoutRacesSessionRun_SessionRunWins` | `unit` | Handles race where Session.Run() transitions to "running" before timeout fires. | Test fixture; mock timer; Session.Run() completes immediately; manually trigger timeout after Session.Run() completes and initCompleted set; use goroutine synchronization to control race | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Session.Run() transitions to "running"; initCompleted set to true; timeout handler fires, observes `initCompleted=true` or `Status="running"`, exits silently; returns Session with `Status="running"` |

### Idempotency — Timeout Handler

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TimeoutHandlerExitsIfInitCompleted` | `unit` | Timeout handler exits silently if initialization already completed. | Test fixture; instrument handler; initialization completes before timeout | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Initialize completes successfully; timeout handler fires after completion; handler acquires mutex, observes `initCompleted=true`, exits without calling Session.Fail() or sending notification |
| `TestInitialize_TimeoutHandlerExitsIfStatusRunning` | `unit` | Timeout handler exits silently if Session.Status is "running". | Test fixture; Session.Run() completes before timeout fires | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Session transitions to "running"; timeout handler fires after; handler reads `Session.Status="running"`, exits without action |
| `TestInitialize_TimeoutHandlerExitsIfStatusFailed` | `unit` | Timeout handler exits silently if Session.Status is already "failed". | Test fixture; Session.Fail() called by another error before timeout fires | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Session transitions to "failed" due to other error; timeout handler fires; handler reads `Session.Status="failed"`, exits without calling Session.Fail() again |

### Boundary Values — Timeout

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TimeoutExactly30Seconds` | `unit` | Timeout fires at exactly 30 seconds. | Test fixture; mock timer with 30-second duration; workflow loader set to block indefinitely; manually trigger timeout | `WorkflowName="SlowWorkflow"`, `TerminationNotifier=<channel>` | Timeout handler fires; Session.Fail() called with RuntimeError `Message="session initialization timeout exceeded 30 seconds"` |
| `TestInitialize_CompletesJustBeforeTimeout` | `unit` | Initialization completes just before timeout. | Test fixture; mock timer with 30-second duration; all mocks complete immediately; verify timeout does NOT trigger before completion | `WorkflowName="FastWorkflow"`, `TerminationNotifier=<channel>` | Initialization completes successfully; timer.Stop() called before timeout fires; returns Session with `Status="running"` |

### Boundary Values — Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_WorkflowNamePascalCase` | `unit` | Accepts PascalCase workflow name. | Test fixture; valid mocks | `WorkflowName="MyTestWorkflow"`, `TerminationNotifier=<channel>` | Returns Session with `WorkflowName="MyTestWorkflow"` |
| `TestInitialize_WorkflowNameSingleWord` | `unit` | Accepts single-word workflow name. | Test fixture; valid mocks | `WorkflowName="Test"`, `TerminationNotifier=<channel>` | Returns Session with `WorkflowName="Test"` |

### Boundary Values — Session UUID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_SessionUUIDFormat` | `unit` | Generated session UUID is valid UUID v4 format. | Test fixture; valid mocks | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Returns Session with `ID` matching UUID v4 format (8-4-4-4-12 hex characters) |

### Resource Cleanup — No Directory Deletion

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_NoDirectoryCleanupOnFailure` | `unit` | Session directory and files remain on disk after failure. | Test fixture; mock Session.Run() returns error; capture filesystem state | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Initialize fails; session directory exists on disk; empty `session.json` and `events.jsonl` files exist; socket deleted |
| `TestInitialize_NoDirectoryCleanupOnTimeout` | `unit` | Session directory and files remain on disk after timeout. | Test fixture; timeout fires before completion; capture filesystem state | `WorkflowName="SlowWorkflow"`, `TerminationNotifier=<channel>` | Timeout fires; Session.Fail() called; session directory and files remain on disk; socket deleted |

### Concurrent Behaviour — Multiple Initializers

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_ConcurrentInitializers` | `race` | Multiple SessionInitializers in different goroutines produce unique session UUIDs. | Test fixture; spawn 10 goroutines, each calling Initialize with same workflow name | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` (10 instances) | All 10 initializations succeed; all session UUIDs are unique; all session directories created independently |

### Mock / Dependency Interaction — Call Order

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_CriticalCallOrder` | `unit` | Verifies critical ordering constraints for correctness. | Test fixture; mock all dependencies with call tracking | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | Critical orderings verified: 1) SessionDirectoryManager.CreateSessionDirectory called before FileAccessor preparation (directory must exist before files), 2) FileAccessor preparation called before RuntimeSocketManager.CreateSocket (files before socket), 3) RuntimeSocketManager.CreateSocket called before SessionMetadataStore.Write (socket before metadata), 4) SessionMetadataStore.Write called before Session.Run (metadata persisted before status transition) |

### Mock / Dependency Interaction — TerminationNotifier

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TerminationNotifierPassedToSessionFail` | `unit` | Verifies TerminationNotifier is passed to Session.Fail() on timeout. | Test fixture; timeout fires; monitor Session.Fail() call | `WorkflowName="SlowWorkflow"`, `TerminationNotifier=<channel>` | Timeout handler calls `Session.Fail(runtimeError, terminationNotifier)` with correct channel |
| `TestInitialize_TerminationNotifierPassedToSessionRun` | `unit` | Verifies TerminationNotifier is passed to Session.Run(). | Test fixture; monitor Session.Run() call | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | SessionInitializer calls `Session.Run(terminationNotifier)` with correct channel |

### Mock / Dependency Interaction — RuntimeError Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInitialize_TimeoutRuntimeErrorFields` | `unit` | Verifies RuntimeError fields when timeout fires. | Test fixture; timeout fires after Session constructed; monitor Session.Fail() call | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | RuntimeError has `Issuer="SessionInitializer"`, `Message="session initialization timeout exceeded 30 seconds"`, `Detail={}`, `SessionID=<session-uuid>`, `FailingState=<EntryNode>`, `OccurredAt` within 1 second of timeout |
| `TestInitialize_SessionRunFailureRuntimeErrorFields` | `unit` | Verifies RuntimeError fields when Session.Run() fails. | Test fixture; mock Session.Run() returns error "test error"; monitor Session.Fail() call | `WorkflowName="TestWorkflow"`, `TerminationNotifier=<channel>` | RuntimeError has `Issuer="SessionInitializer"`, `Message="failed to transition session to running status"`, `Detail` contains original error from Session.Run() |
