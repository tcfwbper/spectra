# Test Specification: `runtime.go`

## Source File Under Test
`runtime/runtime.go`

## Test File
`runtime/runtime_test.go`

---

## `Runtime`

### Happy Path — Initialization and Bootstrap

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SuccessfulInitialization` | `unit` | Initializes runtime with valid workflow name and dependencies. | Test fixture creates temporary .spectra directory; mock SpectraFinder returns projectRoot; mock all dependencies succeed | `workflowName="TestWorkflow"` | Returns `nil`; Session.Status="completed"; SessionFinalizer prints success message |
| `TestRuntime_CreatesTerminationNotifier` | `unit` | Creates terminationNotifier channel with capacity 2. | Test fixture with mock dependencies; instrument Runtime to expose terminationNotifier | `workflowName="TestWorkflow"` | terminationNotifier created with `cap(terminationNotifier) == 2` |
| `TestRuntime_DependencyConstructionOrder` | `unit` | Constructs dependencies in specified order. | Test fixture with mock dependencies tracking construction order | `workflowName="TestWorkflow"` | Dependencies constructed in order: WorkflowDefinitionLoader, SessionDirectoryManager, AgentDefinitionLoader, SessionInitializer |
| `TestRuntime_PostSessionDependencyConstruction` | `unit` | Constructs post-session dependencies after SessionInitializer succeeds. | Test fixture; SessionInitializer returns successfully with session | `workflowName="TestWorkflow"` | SessionMetadataStore, EventStore, RuntimeSocketManager, TransitionToNode, EventProcessor, ErrorProcessor, MessageRouter constructed after session initialized |

### Happy Path — Session Lifecycle

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionCompletedSuccessfully` | `unit` | Returns nil when session completes successfully. | Test fixture; Session.Done called from TransitionToNode sends termination notification | `workflowName="TestWorkflow"` | Runtime returns `nil`; SessionFinalizer prints: `"Session <id> completed successfully. Workflow: TestWorkflow"` to stdout |
| `TestRuntime_SessionInitializerTransitionsToRunning` | `unit` | Session.Status transitions to "running" after SessionInitializer returns. | Test fixture; mock SessionInitializer calls Session.Run | `workflowName="TestWorkflow"` | SessionInitializer returns with `session.Status == "running"` |
| `TestRuntime_SocketCreatedAfterInitialization` | `unit` | Runtime socket created after SessionInitializer succeeds. | Test fixture creates temporary session directory; mock RuntimeSocketManager tracks CreateSocket call | `workflowName="TestWorkflow"` | RuntimeSocketManager.CreateSocket() called exactly once; socket path is `.spectra/sessions/<sessionUUID>/runtime.sock` |
| `TestRuntime_ListenerStartedAfterSocketCreated` | `unit` | Socket listener started after CreateSocket succeeds. | Test fixture; mock RuntimeSocketManager.CreateSocket succeeds; track Listen call | `workflowName="TestWorkflow"` | RuntimeSocketManager.Listen() called with MessageRouter.RouteMessage as messageHandler |

### Happy Path — Main Event Loop

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_EventLoopWaitsForTermination` | `unit` | Main loop blocks on select until first termination signal. | Test fixture; use sync primitives (WaitGroup) to verify Runtime enters select; Session.Done sends notification after verification | `workflowName="TestWorkflow"` | Runtime enters select statement; proceeds to cleanup after receiving termination notification; verified without unconditional delays |
| `TestRuntime_FirstSignalOnly` | `unit` | Processes only first termination signal and ignores subsequent signals. | Test fixture; Session.Done sends notification; listenerErrCh sends error 1ms later | `workflowName="TestWorkflow"` | Runtime proceeds to cleanup after first signal (Session.Done); listenerErrCh signal ignored |

### Happy Path — OS Signal Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SIGINT_GracefulShutdown` | `unit` | Handles SIGINT and initiates graceful shutdown. | Test fixture; send SIGINT after Session enters "running" state | `workflowName="TestWorkflow"` | Runtime logs: `"received signal interrupt, initiating graceful shutdown"`; Session.Fail NOT called; SessionFinalizer prints: `"Session <id> terminated with status 'running'. Workflow: TestWorkflow"` to stderr; returns error: `"session terminated by signal SIGINT"` |
| `TestRuntime_SIGTERM_GracefulShutdown` | `unit` | Handles SIGTERM and initiates graceful shutdown. | Test fixture; send SIGTERM after Session enters "running" state | `workflowName="TestWorkflow"` | Runtime logs: `"received signal terminated, initiating graceful shutdown"`; Session.Fail NOT called; SessionFinalizer prints status to stderr; returns error: `"session terminated by signal SIGTERM"` |
| `TestRuntime_SIGINT_StoresReceivedSignal` | `unit` | Stores received signal in receivedSignal variable. | Test fixture; send SIGINT; instrument Runtime to expose receivedSignal | `workflowName="TestWorkflow"` | receivedSignal set to syscall.SIGINT or os.Interrupt; error message constructed from receivedSignal |
| `TestRuntime_SIGINT_DuringInitializing` | `unit` | Handles SIGINT when Session.Status is "initializing". | Test fixture; send SIGINT during SessionInitializer execution before Session.Run called | `workflowName="TestWorkflow"` | Runtime proceeds to cleanup; SessionFinalizer prints: `"Session <id> terminated with status 'initializing'. Workflow: TestWorkflow"` to stderr; returns error: `"session terminated by signal SIGINT"` |

### Happy Path — Cleanup and Finalization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_CleanupOrder` | `unit` | Cleanup operations execute in specified order. | Test fixture; Session.Done sends notification; track cleanup operation sequence | `workflowName="TestWorkflow"` | Cleanup order: (1) RuntimeSocketManager.DeleteSocket, (2) wait for listenerDoneCh, (3) SessionFinalizer.Finalize |
| `TestRuntime_WaitsForListenerShutdown` | `unit` | Waits for listenerDoneCh to close before SessionFinalizer. | Test fixture; listenerDoneCh closes after verification that wait started; track execution order | `workflowName="TestWorkflow"` | Runtime waits for listenerDoneCh; SessionFinalizer called only after listenerDoneCh closes |
| `TestRuntime_ListenerShutdownTimeout` | `unit` | Proceeds to SessionFinalizer if listenerDoneCh timeout exceeds 2 seconds. | Test fixture; listenerDoneCh never closes; terminationNotifier must receive a signal (sent from inside the mock SessionInitializer or its goroutine) to unblock the main event loop and reach the cleanup phase — mutating session status alone is insufficient; inject two independent mock timers: an immediate-fire timer for the listener shutdown wait (simulating 2-second timeout), and a never-fire timer for the grace period (preventing forced-exit from firing during the test) | `workflowName="TestWorkflow"` | Runtime logs: `"listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer"`; SessionFinalizer called after simulated timeout; test completes quickly without any real delay |
| `TestRuntime_DeleteSocketIdempotent` | `unit` | DeleteSocket is idempotent; warnings logged but cleanup continues. | Test fixture; RuntimeSocketManager.DeleteSocket logs warning "socket not found" but returns nil | `workflowName="TestWorkflow"` | RuntimeSocketManager logs warning; Runtime continues to wait for listenerDoneCh and call SessionFinalizer; no error returned |

### Happy Path — Grace Period Enforcement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_GracePeriodEnforced` | `unit` | Enforces 5-second grace period for cleanup. | Test fixture; listenerDoneCh never closes (simulates cleanup blocked indefinitely); terminationNotifier must receive a signal (sent from inside the mock SessionInitializer or its goroutine) to unblock the main event loop — mutating session status alone is insufficient; inject two independent mock timers: a never-fire timer for the listener shutdown wait (keeping Runtime blocked there so the grace period can fire), and an immediate-fire timer for the grace period (simulating 5-second expiration); inject a no-op exit function so that forced-exit is signalled without terminating the test process, allowing Run() to return normally | `workflowName="TestWorkflow"` | Runtime logs: `"cleanup exceeded 5 second grace period, forcing exit"` after simulated timeout; Run() returns normally (via injectable exit mechanism) so assertions can execute; test completes quickly without any real delay |
| `TestRuntime_SecondSignalForcesExit` | `unit` | Second OS signal during grace period forces immediate exit. | Test fixture; send first SIGINT to trigger graceful shutdown; inject a never-fire timer for the listener shutdown wait so Runtime stays blocked in the cleanup wait (giving the grace period goroutine time to receive the second signal); send second SIGINT during the wait; inject a no-op exit function so that forced-exit is signalled without terminating the test process, allowing Run() to return normally and assertions to execute | `workflowName="TestWorkflow"` | Runtime logs: `"received signal interrupt, initiating graceful shutdown"` (first); Runtime logs: `"received second signal, forcing exit"` after second SIGINT; Run() returns normally (via injectable exit mechanism); log message order verified |

### Validation Failures — SpectraFinder

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SpectraFinderFails_NotInitialized` | `unit` | Returns error when SpectraFinder cannot locate .spectra directory. | Test fixture; mock SpectraFinder.Find() returns error: "spectra not initialized" | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to locate project root: spectra not initialized"`; no resources created; SessionFinalizer NOT called |
| `TestRuntime_SpectraFinderFails_NoResources` | `unit` | No resources created when SpectraFinder fails. | Test fixture; mock SpectraFinder.Find() returns error | `workflowName="TestWorkflow"` | terminationNotifier NOT created; WorkflowDefinitionLoader NOT constructed; Runtime returns error immediately |

### Validation Failures — Dependency Construction (Pre-Session)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_WorkflowDefinitionLoaderConstructionFails` | `unit` | Returns error when WorkflowDefinitionLoader construction fails. | Test fixture; mock WorkflowDefinitionLoader constructor returns error: "failed to initialize" | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to initialize runtime dependencies: failed to initialize"`; no session created; SessionFinalizer NOT called |
| `TestRuntime_SessionDirectoryManagerConstructionFails` | `unit` | Returns error when SessionDirectoryManager construction fails. | Test fixture; mock SessionDirectoryManager constructor returns error | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to initialize runtime dependencies: <error>"`; no session created; SessionFinalizer NOT called |
| `TestRuntime_AgentDefinitionLoaderConstructionFails` | `unit` | Returns error when AgentDefinitionLoader construction fails. | Test fixture; mock AgentDefinitionLoader constructor returns error | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to initialize runtime dependencies: <error>"`; no session created; SessionFinalizer NOT called |
| `TestRuntime_SessionInitializerConstructionFails` | `unit` | Returns error when SessionInitializer construction fails. | Test fixture; mock SessionInitializer constructor returns error | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to initialize runtime dependencies: <error>"`; no session created; SessionFinalizer NOT called |

### Validation Failures — SessionInitializer

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionInitializerFails_NoSession` | `unit` | Returns error when SessionInitializer fails before Session entity created. | Test fixture; mock SessionInitializer.Initialize() returns error with `session == nil` | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to initialize session: <error>"`; SessionFinalizer NOT called |
| `TestRuntime_SessionInitializerFails_WithSession` | `unit` | Calls SessionFinalizer when SessionInitializer fails after Session entity created. | Test fixture; mock SessionInitializer.Initialize() returns error with `session != nil`, `session.Status="failed"` | `workflowName="TestWorkflow"` | Runtime proceeds to cleanup; SessionFinalizer called to print failure; Runtime returns error: `"failed to initialize session: <error>"` |
| `TestRuntime_SessionInitializerTimeout_BeforeSessionEntity` | `unit` | Handles early timeout before Session entity constructed. | Test fixture; SessionInitializer timeout fires before Session entity created; returns error with `session == nil` | `workflowName="TestWorkflow"` | Runtime returns error: `"failed to initialize session: session initialization timeout exceeded 30 seconds before session entity was constructed"`; SessionFinalizer NOT called |
| `TestRuntime_SessionInitializerTimeout_AfterSessionEntity` | `unit` | Handles timeout after Session entity constructed. | Test fixture; SessionInitializer timeout fires after Session entity created; calls Session.Fail; returns error with `session != nil` | `workflowName="TestWorkflow"` | Runtime receives termination notification; proceeds to cleanup; SessionFinalizer prints failure; Runtime returns error: `"failed to initialize session: session initialization timeout exceeded 30 seconds"` |

### Validation Failures — Post-Session Dependencies

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_PostSessionDependencyFails` | `unit` | Calls Session.Fail when post-session dependency construction fails. | Test fixture; SessionInitializer succeeds; EventStore construction fails with error: `"failed to open events file"` | `workflowName="TestWorkflow"` | Runtime constructs RuntimeError with `Issuer="Runtime"`, `Message="failed to initialize post-session dependencies"`, `Detail={"error":"failed to open events file"}`, `SessionID=session.ID`, `FailingState=session.CurrentState`, `OccurredAt=<POSIX timestamp>`; calls Session.Fail; SessionFinalizer prints failure; returns error: `"failed to initialize post-session dependencies: failed to open events file"` |
| `TestRuntime_RuntimeSocketManagerConstructionFails` | `unit` | Handles RuntimeSocketManager construction failure. | Test fixture; SessionInitializer succeeds; RuntimeSocketManager construction fails | `workflowName="TestWorkflow"` | Runtime calls Session.Fail with RuntimeError; SessionFinalizer prints failure; returns error: `"failed to initialize post-session dependencies: <error>"` |
| `TestRuntime_RuntimeErrorDetailFieldPopulated` | `unit` | Verifies RuntimeError.Detail contains underlying error as JSON. | Test fixture; SessionInitializer succeeds; EventStore construction fails with error: `"disk full"` | `workflowName="TestWorkflow"` | RuntimeError constructed with `Detail={"underlying_error":"disk full"}` or similar JSON structure; Detail is valid JSON; calls Session.Fail; SessionFinalizer prints Detail as compact JSON to stderr |
| `TestRuntime_RuntimeErrorOccurredAtTimestamp` | `unit` | Verifies RuntimeError.OccurredAt is a valid POSIX timestamp. | Test fixture; post-session dependency fails; capture RuntimeError construction | `workflowName="TestWorkflow"` | RuntimeError.OccurredAt is set to current POSIX timestamp; timestamp is within ±5 seconds of current time (reasonable tolerance) |

### Validation Failures — Socket Creation and Listener

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_CreateSocketFails_SocketAlreadyExists` | `unit` | Handles socket file already exists error with detailed troubleshooting message. | Test fixture; socket manager's CreateSocket() operation returns error: `"runtime socket file already exists: /tmp/.spectra/sessions/abc-123/runtime.sock"`; note: CreateSocket() must be part of the injectable socket manager interface so its error can be injected in tests — if CreateSocket() is not in the interface, this test cannot exercise the error path | `workflowName="TestWorkflow"` | Runtime constructs RuntimeError with `Message="failed to create runtime socket"`, `Detail=<error>`; calls Session.Fail; SessionFinalizer prints failure to stderr; returns error starting with: `"failed to create runtime socket: runtime socket file already exists: /tmp/.spectra/sessions/abc-123/runtime.sock. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm /tmp/.spectra/sessions/abc-123/runtime.sock"` |
| `TestRuntime_CreateSocketFails_PermissionDenied` | `unit` | Handles permission denied error during socket creation. | Test fixture; socket manager's CreateSocket() operation returns error: `"permission denied"`; CreateSocket() must be part of the injectable socket manager interface | `workflowName="TestWorkflow"` | Runtime constructs RuntimeError; calls Session.Fail; SessionFinalizer prints failure; returns error: `"failed to create runtime socket: permission denied"` |
| `TestRuntime_ListenerStartFails_BindError` | `unit` | Handles bind/listen failure on socket. | Test fixture; RuntimeSocketManager.Listen() returns synchronous error (err != nil) | `workflowName="TestWorkflow"` | Runtime constructs RuntimeError with `Message="failed to start socket listener"`; calls Session.Fail; SessionFinalizer prints failure; returns error: `"failed to start socket listener: <error>"` |

### Validation Failures — Listener Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_ListenerErrorDuringRuntime` | `unit` | Handles fatal listener error during session execution. | Test fixture; Session enters "running" state; listenerErrCh sends error: "accept loop failure" | `workflowName="TestWorkflow"` | Runtime logs: `"listener error: accept loop failure"`; constructs RuntimeError with `Message="listener error"`; calls Session.Fail; SessionFinalizer prints failure; returns error: `"session failed: listener error"` |
| `TestRuntime_ListenerErrorWhenSessionAlreadyCompleted` | `unit` | Skips Session.Fail when listener error occurs but session already terminal. | Test fixture; Session.Done transitions to "completed"; listenerErrCh sends error concurrently | `workflowName="TestWorkflow"` | Runtime receives listenerErrCh error (race: first signal); checks session.Status via GetStatusSafe(); Status is "completed"; Runtime skips Session.Fail; proceeds to cleanup; SessionFinalizer prints completed status |
| `TestRuntime_ListenerErrorWhenSessionAlreadyFailed` | `unit` | Skips Session.Fail when listener error occurs but session already failed. | Test fixture; Session.Fail transitions to "failed"; listenerErrCh sends error concurrently | `workflowName="TestWorkflow"` | Runtime checks session.Status; Status is "failed"; Runtime skips Session.Fail (already terminal); proceeds to cleanup; SessionFinalizer prints failed status |

### Error Propagation — Session Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionFailedWithAgentError` | `unit` | Returns error when session fails due to agent error. | Test fixture; ErrorProcessor calls Session.Fail with AgentError; terminationNotifier receives signal | `workflowName="TestWorkflow"` | Runtime receives termination notification; proceeds to cleanup; SessionFinalizer prints AgentError details to stderr; returns error: `"session failed: <agentError.Message>"` |
| `TestRuntime_SessionFailedWithRuntimeError` | `unit` | Returns error when session fails due to runtime error. | Test fixture; Runtime constructs RuntimeError and calls Session.Fail | `workflowName="TestWorkflow"` | Runtime proceeds to cleanup; SessionFinalizer prints RuntimeError details to stderr; returns error: `"session failed: <runtimeError.Message>"` |

### Error Propagation — Return Values

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_ReturnsNilOnSuccess` | `unit` | Returns nil when session completes successfully. | Test fixture; Session.Done transitions to "completed" | `workflowName="TestWorkflow"` | Runtime returns `nil` |
| `TestRuntime_ReturnsSIGINTError` | `unit` | Returns SIGINT-specific error message. | Test fixture; send SIGINT; session terminates with "running" status | `workflowName="TestWorkflow"` | Runtime returns error: `"session terminated by signal SIGINT"` |
| `TestRuntime_ReturnsSIGTERMError` | `unit` | Returns SIGTERM-specific error message. | Test fixture; send SIGTERM; session terminates with "running" status | `workflowName="TestWorkflow"` | Runtime returns error: `"session terminated by signal SIGTERM"` |
| `TestRuntime_ReturnsSessionFailedError` | `unit` | Returns session-failed error with error message. | Test fixture; Session.Fail with AgentError{Message:"validation failed"} | `workflowName="TestWorkflow"` | Runtime returns error: `"session failed: validation failed"` |
| `TestRuntime_ReturnsNonTerminalStatusError` | `unit` | Returns non-terminal status error when receivedSignal is nil. | Test fixture; simulate edge case where session terminates with "running" but receivedSignal is nil | `workflowName="TestWorkflow"` | Runtime returns error: `"session terminated with status 'running'"` |

### Boundary Values — Signal Timing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionDoneBeforeEventLoop` | `unit` | Handles Session.Done notification sent before main loop starts. | Test fixture; SessionInitializer calls Session.Done immediately; terminationNotifier receives signal before select statement | `workflowName="TestWorkflow"` | Runtime enters main loop; terminationNotifier already has signal (buffered); Runtime receives signal immediately; proceeds to cleanup |
| `TestRuntime_MultipleConcurrentTerminationSignals` | `unit` | Handles concurrent notifications from Session.Done and timeout. | Test fixture; Session.Done and SessionInitializer timeout both send to terminationNotifier concurrently | `workflowName="TestWorkflow"` | terminationNotifier has capacity 2; both signals buffered; Runtime receives first signal; proceeds to cleanup; second signal remains in buffer (ignored) |

### Boundary Values — Empty Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_EmptyWorkflowName` | `unit` | Passes empty workflow name to SessionInitializer. | Test fixture; mock SessionInitializer validates workflow name | `workflowName=""` | SessionInitializer returns error: "workflow name is empty"; Runtime returns error: `"failed to initialize session: workflow name is empty"`; SessionFinalizer NOT called |

### Idempotency — Cleanup Operations

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_MultipleDeleteSocketCalls` | `unit` | DeleteSocket is idempotent. | Test fixture; DeleteSocket called multiple times in cleanup path | `workflowName="TestWorkflow"` | First DeleteSocket deletes socket; subsequent calls are no-ops; no error returned; cleanup continues |

### Concurrent Behaviour — Race Conditions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionDoneAndListenerErrorConcurrent` | `race` | Handles concurrent Session.Done and listener error. | Test fixture; Session.Done and listenerErrCh send signals concurrently | `workflowName="TestWorkflow"` | Runtime receives first signal (non-deterministic); proceeds to cleanup; second signal ignored; no race condition on session state access (GetStatusSafe used) |
| `TestRuntime_SessionDoneAndSIGINTConcurrent` | `race` | Handles concurrent Session.Done and SIGINT. | Test fixture; Session.Done called; SIGINT sent concurrently | `workflowName="TestWorkflow"` | Runtime receives first signal; proceeds to cleanup; appropriate error message and SessionFinalizer output based on which signal won |
| `TestRuntime_ListenerDoneChAlreadyClosed` | `race` | Handles listenerDoneCh already closed when wait starts. | Test fixture; DeleteSocket called and listenerDoneCh closes before Runtime waits | `workflowName="TestWorkflow"` | `<-listenerDoneCh` returns immediately (closed channel); Runtime proceeds to SessionFinalizer without delay |

### State Transitions — Session Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionStatusInitializingToRunning` | `unit` | Session.Status transitions from "initializing" to "running". | Test fixture; SessionInitializer calls Session.Run successfully | `workflowName="TestWorkflow"` | SessionInitializer returns with `session.Status == "running"`; Runtime proceeds to socket creation |
| `TestRuntime_SessionStatusRunningToCompleted` | `unit` | Session.Status transitions from "running" to "completed". | Test fixture; TransitionToNode calls Session.Done | `workflowName="TestWorkflow"` | Session.Status becomes "completed"; Runtime returns nil |
| `TestRuntime_SessionStatusRunningToFailed` | `unit` | Session.Status transitions from "running" to "failed". | Test fixture; ErrorProcessor calls Session.Fail | `workflowName="TestWorkflow"` | Session.Status becomes "failed"; Runtime returns error: `"session failed: <message>"` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_MessageRouterReceivesMessageHandler` | `unit` | RuntimeSocketManager.Listen receives MessageRouter.RouteMessage as handler. | Test fixture; mock RuntimeSocketManager tracks Listen arguments | `workflowName="TestWorkflow"` | RuntimeSocketManager.Listen() called with `messageHandler == MessageRouter.RouteMessage` |
| `TestRuntime_TerminationNotifierPassedToDependencies` | `unit` | terminationNotifier passed to SessionInitializer and processors. | Test fixture; mock dependencies track terminationNotifier argument | `workflowName="TestWorkflow"` | Same terminationNotifier channel passed to SessionInitializer.Initialize, TransitionToNode, EventProcessor, ErrorProcessor, MessageRouter |
| `TestRuntime_SessionFinalizerCalledWithSession` | `unit` | SessionFinalizer.Finalize called with correct session entity. | Test fixture; mock SessionFinalizer tracks Finalize arguments | `workflowName="TestWorkflow"` | SessionFinalizer.Finalize() called with `session == <initialized-session>` |
| `TestRuntime_SessionFailAttemptOnTerminalSession_FirstErrorPreserved` | `unit` | Logs warning when attempting Session.Fail on already-failed session and preserves first error. | Test fixture; Session.Fail called first with AgentError{Message:"validation failed"}; Session.Fail returns error: `"session already failed"` on second attempt; Runtime attempts second Session.Fail with RuntimeError{Message:"listener error"} | `workflowName="TestWorkflow"` | Runtime logs warning: `"attempted to fail session but session already failed: session already failed"`; proceeds to cleanup; session.Error still contains first AgentError with Message="validation failed"; second RuntimeError NOT stored |
| `TestRuntime_UsesGetStatusSafe` | `unit` | Verifies Runtime uses Session.GetStatusSafe() for thread-safe access. | Test fixture; mock Session tracks method calls; listener error occurs | `workflowName="TestWorkflow"` | Mock Session.GetStatusSafe() called to check status before Session.Fail; direct field access to session.Status NOT observed |
| `TestRuntime_UsesGetCurrentStateSafe` | `unit` | Verifies Runtime uses Session.GetCurrentStateSafe() when constructing RuntimeError. | Test fixture; mock Session tracks method calls; post-session dependency fails | `workflowName="TestWorkflow"` | Mock Session.GetCurrentStateSafe() called when constructing RuntimeError.FailingState; direct field access to session.CurrentState NOT observed |

### Resource Cleanup — Panic Recovery

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionFinalizerPanics_WithDeferredRecovery` | `unit` | Recovers from SessionFinalizer panic using deferred recovery. | Test fixture; mock SessionFinalizer.Finalize() panics with `"print failed"`; Session.Status="completed" | `workflowName="TestWorkflow"` | Runtime's deferred recover() catches panic; logs panic message with stack trace; after catching the panic, Finalize() is NOT called a second time (calling it again would re-panic with no recovery); cleanup completes (DeleteSocket already called before SessionFinalizer); returns `nil` (session completed successfully despite panic); panic does not propagate to caller |
| `TestRuntime_SessionFinalizerPanics_FailedSession` | `unit` | Recovers from SessionFinalizer panic when session failed. | Test fixture; mock SessionFinalizer.Finalize() panics; Session.Status="failed", Session.Error=AgentError{Message:"agent error"} | `workflowName="TestWorkflow"` | Runtime recovers from panic; logs panic message; returns error: `"session failed: agent error"` (error based on session.Status and session.Error) |

### Platform Compatibility — Signal Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_WindowsSIGINTOnly` | `unit` | Registers only SIGINT on Windows platform. | Test fixture; simulate Windows environment (GOOS=windows); verify signal.Notify arguments | `workflowName="TestWorkflow"` | signal.Notify() called with only SIGINT (not SIGTERM); SIGTERM gracefully skipped |
| `TestRuntime_UnixSignalRegistration` | `unit` | Registers both SIGINT and SIGTERM on Unix-like platforms. | Test fixture; simulate Unix/Linux/macOS environment (GOOS=linux or darwin); verify signal.Notify arguments | `workflowName="TestWorkflow"` | signal.Notify() called with both SIGINT and SIGTERM; both signals monitored in main loop |

### Edge Cases — Channel Lifecycle

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_TerminationNotifierNeverClosed` | `unit` | terminationNotifier is never closed by any component. | Test fixture; instrument Runtime and dependencies to verify close() is never called on terminationNotifier | `workflowName="TestWorkflow"` | terminationNotifier is sent to but never closed; no close() call observed |
| `TestRuntime_ListenerErrChNeverClosed` | `unit` | listenerErrCh is never closed by RuntimeSocketManager. | Test fixture; instrument RuntimeSocketManager to verify listenerErrCh is never closed | `workflowName="TestWorkflow"` | listenerErrCh is sent to (or not) but never closed by RuntimeSocketManager |
| `TestRuntime_ListenerDoneChClosedOnce` | `unit` | listenerDoneCh closed exactly once when listener exits. | Test fixture; RuntimeSocketManager listenerDoneCh closes after DeleteSocket | `workflowName="TestWorkflow"` | listenerDoneCh closed exactly once; Runtime wait succeeds |

### Edge Cases — Multiple Runtime Instances

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_ConcurrentRuntimesDifferentWorkflows` | `e2e` | Multiple runtime instances with different workflow names run independently. | Test fixture creates two isolated temporary directories; each directory has its own .spectra project initialization with workflow definitions; spawn two Runtime instances in these isolated fixtures | `workflowName="Workflow1"` (in tmpDir1) and `workflowName="Workflow2"` (in tmpDir2) | Both runtimes execute independently in isolated fixtures; generate unique session UUIDs; no conflict; both complete successfully; no writes to pre-existing project directories |
| `TestRuntime_ConcurrentRuntimesSameWorkflow_UniqueSessionIDs` | `e2e` | Multiple runtime instances with same workflow name generate unique session IDs. | Test fixture creates isolated temporary directory with .spectra initialization; spawn two Runtime instances with same workflow name concurrently in same isolated fixture | `workflowName="TestWorkflow"` (both instances) | Each generates unique session UUID; session directories created independently in isolated fixture; no UUID collision; no writes to pre-existing project directories |

### Edge Cases — Listener Shutdown Timing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_ListenerDoneCh_ImmediateClose` | `unit` | Handles listenerDoneCh closing before DeleteSocket returns. | Test fixture; listenerDoneCh closes immediately when DeleteSocket called (race: listener exits very fast) | `workflowName="TestWorkflow"` | Runtime waits for listenerDoneCh; channel already closed; wait returns immediately; SessionFinalizer called without delay |

### Edge Cases — Empty Error Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_SessionFailedWithEmptyAgentErrorMessage` | `unit` | Handles AgentError with empty Message field. | Test fixture; Session.Fail with AgentError{Message:"", AgentRole:"reviewer", FailingState:"review_node"} | `workflowName="TestWorkflow"` | Runtime returns error: `"session failed: "` (empty message preserved); SessionFinalizer prints to stderr: `"Error: "` |
| `TestRuntime_SessionFailedWithEmptyRuntimeErrorMessage` | `unit` | Handles RuntimeError with empty Message field. | Test fixture; Session.Fail with RuntimeError{Message:"", Issuer:"MessageRouter", FailingState:"node1"} | `workflowName="TestWorkflow"` | Runtime returns error: `"session failed: "` (empty message preserved); SessionFinalizer prints to stderr: `"Error: "` |

### Error Propagation — spectra run Exit Code Mapping

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRuntime_ExitCode0_SessionCompleted` | `e2e` | Returns nil for successful exit code 0 mapping. | Test fixture creates isolated temporary directory with .spectra initialization; Runtime.Run() completes session successfully | `workflowName="TestWorkflow"` | Runtime returns `nil`; `spectra run` command converts to exit code 0 |
| `TestRuntime_ExitCode1_GenericFailure` | `e2e` | Returns generic error for exit code 1 mapping. | Test fixture creates isolated temporary directory; Runtime.Run() returns error: `"failed to initialize session: <error>"` | `workflowName="TestWorkflow"` | Runtime returns error; `spectra run` command converts to exit code 1; error printed to stderr |
| `TestRuntime_ExitCode1_SessionFailed` | `e2e` | Returns session-failed error for exit code 1 mapping. | Test fixture creates isolated temporary directory; Session.Fail called with AgentError | `workflowName="TestWorkflow"` | Runtime returns error: `"session failed: validation failed"`; `spectra run` command converts to exit code 1; error printed to stderr |
| `TestRuntime_ExitCode130_SIGINT` | `e2e` | Returns SIGINT error for exit code 130 mapping. | Test fixture creates isolated temporary directory; send SIGINT to Runtime during execution | `workflowName="TestWorkflow"` | Runtime returns error: `"session terminated by signal SIGINT"`; `spectra run` command converts to exit code 130 (128 + 2); error printed to stderr |
| `TestRuntime_ExitCode143_SIGTERM` | `e2e` | Returns SIGTERM error for exit code 143 mapping. | Test fixture creates isolated temporary directory; send SIGTERM to Runtime during execution | `workflowName="TestWorkflow"` | Runtime returns error: `"session terminated by signal SIGTERM"`; `spectra run` command converts to exit code 143 (128 + 15); error printed to stderr |
