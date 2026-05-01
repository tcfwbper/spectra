# Test Specification: `runtime.go`

## Source File Under Test
`runtime/runtime.go`

## Test File
`runtime/runtime_test.go`

---

## `Runtime`

### Happy Path — Main Loop Flow (Session Completion)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SuccessfulCompletion` | `e2e` | Runs complete workflow session successfully from start to completion. | Test fixture creates temporary project directory with `.spectra/` structure and valid workflow definition; real SessionInitializer, SessionFinalizer, MessageRouter, RuntimeSocketManager with mocks for underlying stores | `WorkflowName="TestWorkflow"` | Session completes successfully with `Status="completed"`; SessionFinalizer prints success message to stdout; process exits with code 0 |
| `TestRun_CompletedSession_ExitCode0` | `unit` | Exits with code 0 when session completes successfully. | Test fixture; mock SessionInitializer returns Session with `Status="completed"` (simulating immediate completion); mock SessionFinalizer | `WorkflowName="TestWorkflow"` | Runtime exits with code 0 |
| `TestRun_SessionDoneNotification` | `unit` | Receives termination notification when Session.Done() called. | Test fixture; mock Session.Done() sends notification to terminationNotifier during workflow execution; mock SessionInitializer | `WorkflowName="TestWorkflow"` | Runtime receives notification from terminationNotifier; exits monitoring loop; calls SessionFinalizer; exits with code 0 |
| `TestRun_TerminationNotifierBufferSize2` | `unit` | Creates terminationNotifier channel with buffer size 2. | Test fixture; capture channel creation | `WorkflowName="TestWorkflow"` | terminationNotifier channel created with `cap=2` before calling SessionInitializer.Initialize() |

### Happy Path — Main Loop Flow (Session Failure)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_FailedSession_AgentError` | `e2e` | Handles session failure due to AgentError. | Test fixture creates project with workflow; agent returns error during execution | `WorkflowName="FailingWorkflow"` | Session fails with `Status="failed"`; SessionFinalizer prints error to stderr; process exits with code 1 |
| `TestRun_FailedSession_ExitCode1` | `unit` | Exits with code 1 when session fails. | Test fixture; mock SessionInitializer returns Session with `Status="failed"` | `WorkflowName="TestWorkflow"` | Runtime exits with code 1 |
| `TestRun_SessionFailNotification` | `unit` | Receives termination notification when Session.Fail() called. | Test fixture; trigger Session.Fail() during execution by simulating RuntimeError; mock SessionInitializer | `WorkflowName="TestWorkflow"` | Runtime receives notification from terminationNotifier; exits monitoring loop; calls SessionFinalizer; exits with code 1 |

### Happy Path — Socket Listener Management

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SocketListenerStarted` | `unit` | Starts socket listener after successful initialization. | Test fixture; mock RuntimeSocketManager.Listen() returns valid channels; track Listen() call | `WorkflowName="TestWorkflow"` | RuntimeSocketManager.Listen() called with MessageRouter.RouteMessage callback; listenerErrCh and listenerDoneCh returned |
| `TestRun_SocketListenerStopped` | `unit` | Stops socket listener on session completion. | Test fixture; mock RuntimeSocketManager with tracking; simulate session completion | `WorkflowName="TestWorkflow"` | RuntimeSocketManager.DeleteSocket() called; Runtime waits for listenerDoneCh closure before SessionFinalizer |
| `TestRun_MessageRouterInitialized` | `unit` | Initializes MessageRouter with correct dependencies. | Test fixture; capture MessageRouter initialization arguments | `WorkflowName="TestWorkflow"` | MessageRouter initialized with Session, EventProcessor, ErrorProcessor, terminationNotifier |
| `TestRun_ListenerGoroutineCompletes` | `unit` | Waits for listener goroutine to exit before finalizing. | Test fixture; mock RuntimeSocketManager; track goroutine lifecycle; listenerDoneCh closed after goroutine exit | `WorkflowName="TestWorkflow"` | Runtime blocks on listenerDoneCh read; after channel closure, proceeds to SessionFinalizer |

### Happy Path — Signal Handling (SIGINT)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SIGINT_GracefulShutdown` | `unit` | Initiates graceful shutdown on SIGINT. | Test fixture; simulate session running; send SIGINT signal to process | `WorkflowName="TestWorkflow"` then SIGINT | Runtime logs: `"Received signal interrupt. Initiating graceful shutdown."`; RuntimeSocketManager.DeleteSocket() called; session Status remains "running"; SessionFinalizer called; exits with code 1 |
| `TestRun_SIGINT_SessionStatusUnchanged` | `unit` | Session status remains "running" after SIGINT. | Test fixture; mock Session with `Status="running"`; send SIGINT | `WorkflowName="TestWorkflow"` then SIGINT | Session.Status remains "running" (not transitioned to "failed"); SessionFinalizer receives session with `Status="running"` |
| `TestRun_SIGINT_SocketDeletedBeforeFinalization` | `unit` | Deletes socket before calling SessionFinalizer on SIGINT. | Test fixture; mock RuntimeSocketManager; track call order; send SIGINT | `WorkflowName="TestWorkflow"` then SIGINT | RuntimeSocketManager.DeleteSocket() called before SessionFinalizer.Finalize() |
| `TestRun_SIGINT_LocksReleased` | `unit` | Releases all locks on SIGINT. | Test fixture; Session holds internal locks; send SIGINT | `WorkflowName="TestWorkflow"` then SIGINT | All locks automatically released when goroutines exit; no deadlock |

### Happy Path — Signal Handling (SIGTERM)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SIGTERM_GracefulShutdown` | `unit` | Initiates graceful shutdown on SIGTERM. | Test fixture; simulate session running; send SIGTERM signal | `WorkflowName="TestWorkflow"` then SIGTERM | Runtime logs: `"Received signal terminated. Initiating graceful shutdown."`; RuntimeSocketManager.DeleteSocket() called; SessionFinalizer called; exits with code 1 |
| `TestRun_SIGTERM_SessionStatusUnchanged` | `unit` | Session status remains unchanged after SIGTERM. | Test fixture; mock Session with `Status="running"`; send SIGTERM | `WorkflowName="TestWorkflow"` then SIGTERM | Session.Status remains "running"; SessionFinalizer receives session with `Status="running"` |

### Happy Path — Double Signal Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_DoubleSignal_ImmediateExit` | `unit` | Exits immediately with code 130 on second signal. | Test fixture; send SIGINT; during shutdown, send second SIGINT | `WorkflowName="TestWorkflow"` then SIGINT twice | First SIGINT initiates graceful shutdown; second SIGINT exits immediately with code 130; SessionFinalizer may not complete |
| `TestRun_DoubleSignal_NoWaitForFinalization` | `unit` | Does not wait for SessionFinalizer on double signal. | Test fixture; mock SessionFinalizer with delay; send SIGINT twice rapidly | `WorkflowName="TestWorkflow"` then SIGINT twice | Second SIGINT exits process immediately without waiting for SessionFinalizer to complete |

### Happy Path — Listener Error Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_ListenerErrorReceived` | `unit` | Handles asynchronous listener error from listenerErrCh. | Test fixture; mock RuntimeSocketManager sends error on listenerErrCh (e.g., "accept error") during session execution | `WorkflowName="TestWorkflow"` | Runtime receives error on listenerErrCh; constructs RuntimeError with `Issuer="Runtime"`, `Message="runtime socket listener error"`; calls Session.Fail(runtimeError, terminationNotifier); receives termination notification; exits monitoring loop; SessionFinalizer called; exits with code 1 |
| `TestRun_ListenerErrorAfterTermination_Discarded` | `unit` | Drains and discards listener errors after session terminates. | Test fixture; session completes; listenerErrCh receives error after termination notification | `WorkflowName="TestWorkflow"` | Runtime exits monitoring loop on termination notification; waits for listenerDoneCh; drains remaining errors from listenerErrCh (non-blocking, best-effort); errors discarded; SessionFinalizer called normally |

### Happy Path — Socket Deletion Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SocketDeleteCalledTwice` | `unit` | Calls RuntimeSocketManager.DeleteSocket() multiple times safely. | Test fixture; simulate SIGINT (calls DeleteSocket once); then cleanup step (calls DeleteSocket again) | `WorkflowName="TestWorkflow"` then SIGINT | RuntimeSocketManager.DeleteSocket() called twice; second call is no-op (idempotent); no error |

### Validation Failures — Project Root Lookup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_ProjectRootNotFound` | `unit` | Exits with error when SpectraFinder fails to locate project root. | Test fixture creates temporary directory WITHOUT `.spectra/`; mock SpectraFinder returns error: "project root not found" | `WorkflowName="TestWorkflow"` | Runtime prints to stderr: `"Failed to locate project root: project root not found. Run 'spectra init' to initialize the project."`; exits with code 1; SessionInitializer NOT called; SessionFinalizer NOT called |
| `TestRun_ProjectRootNotFoundFromSubdirectory` | `unit` | Exits with error when SpectraFinder fails from subdirectory. | Test fixture creates temporary directory with deep subdirectory structure but NO `.spectra/` at any level; test changes working directory to deepest subdirectory; mock SpectraFinder returns error | `WorkflowName="TestWorkflow"` | Runtime prints to stderr: `"Failed to locate project root: <error>. Run 'spectra init' to initialize the project."`; exits with code 1; SessionInitializer NOT called |

### Validation Failures — Initialization Errors (No Session Entity)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_InitializationFails_WorkflowNotFound` | `unit` | Exits with error when workflow definition not found. | Test fixture; mock SpectraFinder succeeds; mock SessionInitializer returns error: "failed to load workflow definition: file not found"; no session entity | `WorkflowName="NonExistentWorkflow"` | Runtime prints to stderr: `"Failed to initialize session: failed to load workflow definition: file not found"`; exits with code 1; SessionFinalizer NOT called |
| `TestRun_InitializationFails_NoSessionEntity` | `unit` | Handles early initialization failure gracefully. | Test fixture; mock SpectraFinder succeeds; mock SessionInitializer returns error with nil session | `WorkflowName="TestWorkflow"` | Runtime prints error to stderr; exits with code 1; SessionFinalizer NOT called (no session to finalize) |

### Validation Failures — Initialization Errors (Session Entity Exists)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_InitializationFails_SessionExists_Initializing` | `unit` | Calls SessionFinalizer when initialization fails but session exists in initializing state. | Test fixture; mock SessionInitializer returns error AND partially initialized Session with `Status="initializing"` | `WorkflowName="TestWorkflow"` | Runtime proceeds to SessionFinalizer with partial session; SessionFinalizer prints status to stderr; exits with code 1 |
| `TestRun_InitializationFails_SessionExists_Failed` | `unit` | Calls SessionFinalizer when initialization fails and session in failed state. | Test fixture; mock SessionInitializer returns error AND Session with `Status="failed"` | `WorkflowName="TestWorkflow"` | Runtime calls SessionFinalizer; SessionFinalizer prints error details; exits with code 1 |
| `TestRun_InitializationFails_SocketCreationError` | `unit` | Handles socket creation failure during initialization. | Test fixture; mock SessionInitializer fails at socket creation step; returns Session with `Status="initializing"` | `WorkflowName="TestWorkflow"` | SessionFinalizer called with partial session; SessionFinalizer prints error; exits with code 1 |
| `TestRun_InitializationTimeout` | `unit` | Handles initialization timeout with session entity. | Test fixture; mock SessionInitializer timeout handler calls Session.Fail(); SessionInitializer returns error; Session exists with `Status="failed"` | `WorkflowName="TestWorkflow"` | SessionInitializer sends notification to terminationNotifier; Runtime receives notification or error return; calls SessionFinalizer; SessionFinalizer prints timeout RuntimeError; exits with code 1 |

### Validation Failures — Listener Synchronous Setup Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_ListenerSetup_BindFailure` | `unit` | Handles synchronous bind failure from RuntimeSocketManager.Listen(). | Test fixture; mock SessionInitializer succeeds; mock RuntimeSocketManager.Listen() returns syncErr="bind: address already in use", listenerDoneCh already closed | `WorkflowName="TestWorkflow"` | Runtime receives syncErr; constructs RuntimeError with `Issuer="Runtime"`, `Message="failed to start socket listener"`, `Detail["error"]="bind: address already in use"`; calls Session.Fail(runtimeError, terminationNotifier); proceeds to SessionFinalizer; SessionFinalizer prints error; exits with code 1 |
| `TestRun_ListenerSetup_SocketAlreadyExists` | `unit` | Handles synchronous error when socket file already exists. | Test fixture; mock RuntimeSocketManager.Listen() returns syncErr="socket file already exists", listenerDoneCh closed | `WorkflowName="TestWorkflow"` | Runtime calls Session.Fail with RuntimeError; proceeds to SessionFinalizer; exits with code 1 |
| `TestRun_ListenerSetup_ListenerNeverSpawned` | `unit` | Verifies listener goroutine never spawned on sync error. | Test fixture; mock RuntimeSocketManager.Listen() returns syncErr; listenerDoneCh already closed | `WorkflowName="TestWorkflow"` | Runtime proceeds to cleanup; reads from listenerDoneCh (already closed); no goroutine was spawned; SessionFinalizer called; exits with code 1 |

### Validation Failures — Empty Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_EmptyWorkflowName` | `unit` | Rejects empty workflow name. | Test fixture; mock SessionInitializer validates input | `WorkflowName=""` | SessionInitializer returns error: "workflowName must be non-empty"; Runtime prints error to stderr; exits with code 1 |

### State Transitions — Terminal States

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_StatusCompleted_ExitCode0` | `unit` | Exits with code 0 for completed session. | Test fixture; mock Session with `Status="completed"` | `WorkflowName="TestWorkflow"` | SessionFinalizer called; Runtime exits with code 0 |
| `TestRun_StatusFailed_ExitCode1` | `unit` | Exits with code 1 for failed session. | Test fixture; mock Session with `Status="failed"` | `WorkflowName="TestWorkflow"` | SessionFinalizer called; Runtime exits with code 1 |
| `TestRun_StatusInitializing_SIGINT_ExitCode1` | `unit` | Exits with code 1 when SIGINT received during initialization. | Test fixture; Session in `Status="initializing"`; send SIGINT | `WorkflowName="TestWorkflow"` then SIGINT | Session Status remains "initializing"; SessionFinalizer prints initializing status; exits with code 1 |

### Edge Cases — Termination Race Conditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_CompletionAndSIGINT_Race` | `unit` | Handles race between session completion and SIGINT signal. | Test fixture; trigger Session.Done() and SIGINT simultaneously; use select in monitoring loop | `WorkflowName="TestWorkflow"` | Runtime select receives whichever event arrives first; if termination notification, proceeds with normal completion (exit code 0); if SIGINT, proceeds with graceful shutdown (exit code 1); either outcome acceptable; no panic |
| `TestRun_ListenerErrorAndCompletion_Race` | `unit` | Handles race between listener error and session completion. | Test fixture; listenerErrCh receives error at same time Session.Done() sends notification | `WorkflowName="TestWorkflow"` | Runtime select receives whichever event arrives first; if listenerErrCh, calls Session.Fail then exits; if terminationNotifier, exits normally; deterministic based on select choice |

### Edge Cases — Channel Buffer Management

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_TerminationNotifierBufferNotExhausted` | `unit` | Verifies terminationNotifier buffer does not fill under normal operation. | Test fixture; Session.Done() sends one notification; monitor channel | `WorkflowName="TestWorkflow"` | Session.Done() sends exactly one notification; buffer size 2 accommodates notification; Runtime reads notification; no blocking |
| `TestRun_TerminationNotifierNeverClosed` | `unit` | Verifies terminationNotifier channel never closed. | Test fixture; track channel lifecycle; session completes | `WorkflowName="TestWorkflow"` | terminationNotifier created by Runtime; never closed by any component; garbage-collected after Runtime exits |
| `TestRun_ListenerErrChNeverClosed` | `unit` | Verifies listenerErrCh never closed by RuntimeSocketManager. | Test fixture; mock RuntimeSocketManager; track listenerErrCh lifecycle | `WorkflowName="TestWorkflow"` | listenerErrCh owned by RuntimeSocketManager; never closed; consumers must observe listenerDoneCh closure as shutdown signal |
| `TestRun_ListenerDoneChClosedExactlyOnce` | `unit` | Verifies listenerDoneCh closed exactly once when goroutine exits. | Test fixture; mock RuntimeSocketManager; track listenerDoneCh close events | `WorkflowName="TestWorkflow"` | RuntimeSocketManager closes listenerDoneCh exactly once when listener goroutine exits; Runtime reads from channel (receives zero value) |

### Edge Cases — SessionFinalizer Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SessionFinalizerPrintFails` | `unit` | Handles SessionFinalizer print failure gracefully. | Test fixture; redirect stdout/stderr to closed pipe; Session completes | `WorkflowName="TestWorkflow"` | SessionFinalizer print operations fail silently; Runtime proceeds to exit; exits with appropriate code based on session status |
| `TestRun_SessionFinalizerSocketDeleteWarning` | `unit` | Ignores socket deletion warning from SessionFinalizer. | Test fixture; mock RuntimeSocketManager.DeleteSocket() logs warning "socket not found"; Session completes | `WorkflowName="TestWorkflow"` | RuntimeSocketManager logs warning; SessionFinalizer continues; Runtime exits with code 0 |

### Edge Cases — Session Metadata Persistence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_MetadataPersistenceFails_NonBlocking` | `unit` | Continues when session metadata persistence fails (best-effort). | Test fixture; mock SessionMetadataStore.Write() logs warning but does not return error during Session.Done(); Session completes in-memory | `WorkflowName="TestWorkflow"` | Session.Done() logs warning; in-memory status is "completed"; SessionFinalizer prints success message; Runtime exits with code 0 |

### Edge Cases — Listener Goroutine Panic

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_MessageRouterPanic` | `unit` | Handles panic in MessageRouter via panic recovery. | Test fixture; mock MessageRouter.RouteMessage panics with "unexpected nil"; MessageRouter implements panic recovery | `WorkflowName="TestWorkflow"` | MessageRouter panic recovery triggers RuntimeError; Session.Fail() called; terminationNotifier receives notification; Runtime exits monitoring loop; SessionFinalizer prints RuntimeError with panic details; exits with code 1 |

### Edge Cases — Listener In-Flight Messages

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_ListenerProcessingMessage_SocketDeleted` | `unit` | Handles socket deletion while listener processing message. | Test fixture; listener goroutine currently processing message; Runtime calls RuntimeSocketManager.DeleteSocket() | `WorkflowName="TestWorkflow"` then SIGINT during message processing | Socket closed; connection interrupted; listener goroutine handles error gracefully or MessageRouter panic recovery triggered; goroutine exits and closes listenerDoneCh; Runtime waits for listenerDoneCh; SessionFinalizer called |

### Edge Cases — Runtime Crash (Unhandled)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_RuntimeCrash_NoCleanup` | `unit` | Documents behavior when Runtime process crashes. | Test fixture creates project; manually terminate process with kill -9 | `WorkflowName="TestWorkflow"` then kill -9 | Process immediately terminated by OS; no cleanup performed; runtime socket file remains on disk; session files remain; on next session creation with different UUID, system works normally; crash recovery not specified in design |

### Boundary Values — Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_WorkflowNamePascalCase` | `unit` | Accepts valid PascalCase workflow name. | Test fixture; mock SessionInitializer validates workflow name format | `WorkflowName="ValidWorkflowName"` | SessionInitializer accepts name; session initializes successfully |
| `TestRun_WorkflowNameSingleWord` | `unit` | Accepts single-word workflow name. | Test fixture | `WorkflowName="Workflow"` | Session initializes successfully |

### Idempotency — Runtime Invocations

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_MultipleInvocations_DifferentSessions` | `e2e` | Multiple Runtime invocations create independent sessions. | Test fixture creates project; call Runtime.Run() twice with same workflow name in sequence (second call after first exits) | `WorkflowName="TestWorkflow"` (called twice) | Each invocation creates new session with unique UUID; socket file created and deleted for each session; sessions independent; both complete successfully |

### Concurrent Behaviour — Single Session Per Invocation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SingleSessionBinding` | `unit` | Verifies Runtime binds to exactly one session per invocation. | Test fixture; track Session entity creation | `WorkflowName="TestWorkflow"` | Runtime creates exactly one Session entity; Runtime does not manage multiple sessions concurrently |

### Mock / Dependency Interaction — SpectraFinder

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SpectraFinderCalled` | `unit` | Calls SpectraFinder to locate project root. | Test fixture creates temporary directory with `.spectra/`; test changes working directory to test fixture; mock SpectraFinder tracks call | `WorkflowName="TestWorkflow"` | SpectraFinder called to locate project root from current working directory; returns project root path |
| `TestRun_SpectraFinderFromSubdirectory` | `unit` | Calls SpectraFinder from subdirectory. | Test fixture creates temporary directory with `.spectra/`; subdirectory `sub/nested/` created inside test fixture; test changes working directory to `sub/nested/`; mock SpectraFinder tracks call | `WorkflowName="TestWorkflow"` | SpectraFinder called from `sub/nested/`; returns parent directory containing `.spectra/` |

### Mock / Dependency Interaction — SessionInitializer

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SessionInitializerCalled` | `unit` | Calls SessionInitializer.Initialize with correct arguments including resolved project root. | Test fixture; mock SpectraFinder returns `/tmp/test-project/`; mock SessionInitializer tracks Initialize() call | `WorkflowName="TestWorkflow"` | SessionInitializer.Initialize() called with `WorkflowName="TestWorkflow"`, `ProjectRoot="/tmp/test-project/"`, `TerminationNotifier=<channel-cap-2>` |
| `TestRun_SessionInitializerReturnsSession` | `unit` | Receives Session entity from SessionInitializer. | Test fixture; mock SpectraFinder succeeds; mock SessionInitializer returns Session with `ID="abc-123"`, `Status="running"` | `WorkflowName="TestWorkflow"` | Runtime receives Session; proceeds to start socket listener |

### Mock / Dependency Interaction — SessionFinalizer

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_SessionFinalizerCalled` | `unit` | Calls SessionFinalizer.Finalize with Session entity. | Test fixture; mock SessionFinalizer tracks Finalize() call; Session completes | `WorkflowName="TestWorkflow"` | SessionFinalizer.Finalize() called with Session entity |
| `TestRun_SessionFinalizerCalledOnAllPaths` | `unit` | Verifies SessionFinalizer called on all termination paths. | Test fixture; test success, failure, SIGINT paths separately | `WorkflowName="TestWorkflow"` (3 scenarios) | SessionFinalizer.Finalize() called in all cases: completion, failure, SIGINT; only exception is early initialization failure with no session entity |

### Mock / Dependency Interaction — RuntimeSocketManager

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_RuntimeSocketManagerListenCalled` | `unit` | Calls RuntimeSocketManager.Listen with MessageRouter callback. | Test fixture; mock RuntimeSocketManager tracks Listen() call | `WorkflowName="TestWorkflow"` | RuntimeSocketManager.Listen() called with callback function `MessageRouter.RouteMessage` |
| `TestRun_RuntimeSocketManagerDeleteSocketCalled` | `unit` | Calls RuntimeSocketManager.DeleteSocket on cleanup. | Test fixture; mock RuntimeSocketManager tracks DeleteSocket() call; Session completes | `WorkflowName="TestWorkflow"` | RuntimeSocketManager.DeleteSocket() called after session termination, before SessionFinalizer |

### Mock / Dependency Interaction — MessageRouter

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_MessageRouterInitializedWithDependencies` | `unit` | Initializes MessageRouter with all required dependencies. | Test fixture; capture MessageRouter initialization | `WorkflowName="TestWorkflow"` | MessageRouter initialized with Session, EventProcessor, ErrorProcessor, terminationNotifier |

### Resource Cleanup — Locks

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_LocksReleasedOnExit` | `unit` | All locks released before Runtime exits. | Test fixture; track Session internal locks and file locks; Session completes | `WorkflowName="TestWorkflow"` | All locks released automatically by component methods and goroutine exit; no deadlock |
| `TestRun_LocksReleasedOnSIGINT` | `unit` | All locks released on SIGINT graceful shutdown. | Test fixture; track locks; send SIGINT | `WorkflowName="TestWorkflow"` then SIGINT | All locks released; Session internal lock released; file locks in stores released; RuntimeSocketManager file lock released |

### Asynchronous Flow — Listener Goroutine

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_ListenerGoroutineSpawned` | `unit` | Verifies listener goroutine spawned after successful Listen() call. | Test fixture; mock RuntimeSocketManager.Listen() spawns goroutine; track goroutine lifecycle | `WorkflowName="TestWorkflow"` | RuntimeSocketManager.Listen() spawns accept-loop goroutine; goroutine runs until socket closed or error; goroutine closes listenerDoneCh on exit |
| `TestRun_MainLoopBlocksOnSelect` | `unit` | Main monitoring loop blocks on select without polling. | Test fixture; monitor CPU usage; Session runs without events | `WorkflowName="TestWorkflow"` (long-running) | Main loop blocks on select; no busy-waiting; CPU usage minimal while idle |

### Ordering — Cleanup Sequence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRun_CleanupOrder_StopListener_WaitGoroutine_Finalize` | `unit` | Cleanup follows correct sequence: stop listener, wait for goroutine, finalize. | Test fixture; track call order; Session completes | `WorkflowName="TestWorkflow"` | RuntimeSocketManager.DeleteSocket() called first; then wait for listenerDoneCh closure; then SessionFinalizer.Finalize() called; order guaranteed |
| `TestRun_CleanupOrder_DrainErrors_AfterDoneChannel` | `unit` | Drains listenerErrCh after listenerDoneCh closure. | Test fixture; listenerErrCh has errors; Session completes | `WorkflowName="TestWorkflow"` | Runtime waits for listenerDoneCh closure; then drains remaining errors from listenerErrCh (non-blocking); errors discarded; SessionFinalizer called |
