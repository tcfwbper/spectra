# Runtime

## Overview

Runtime is the top-level orchestrator invoked by `spectra run`. It receives a workflow name, an optional session ID, and a Logger, bootstraps all dependencies, initializes a session, creates the runtime socket, performs the initial dispatch of the entry node, runs the main event loop, handles termination signals (session completion/failure, listener errors, OS signals), enforces a grace period for cleanup, and returns an exit code with an optional error. Runtime is the single entry point for workflow execution and coordinates the lifecycle of all runtime components.

Runtime does not manage session directory creation or deletion (SessionInitializer and `spectra clear` respectively), does not validate messages (RuntimeSocketManager and processors), and does not evaluate transitions (TransitionEvaluator).

## Boundaries

- Owns: overall orchestration sequence (bootstrap → init → socket → dispatch → loop → cleanup → finalize).
- Owns: dependency construction and wiring of all runtime components.
- Owns: `terminationNotifier` channel creation and capacity guarantee.
- Owns: OS signal registration (SIGINT, SIGTERM) and signal-based termination.
- Owns: grace period enforcement (5-second timer) for cleanup operations.
- Owns: second signal force exit.
- Owns: initial dispatch of the entry node after session initialization.
- Owns: RuntimeError construction for post-session failures (socket, listener, initial dispatch).
- Owns: cleanup ordering (DeleteSocket → wait listenerDoneCh → SessionFinalizer).
- Owns: exit code propagation from SessionFinalizer to caller.
- Delegates: project root discovery to SpectraFinder.
- Delegates: session initialization (UUID, workflow load, directory, session entity, stores, PersistentSession) to SessionInitializer.
- Delegates: socket lifecycle (create, listen, delete) to RuntimeSocketManager.
- Delegates: message routing to MessageRouter.
- Delegates: node dispatch (stdout print, agent invocation, state update) to TransitionToNode.
- Delegates: final status reporting and exit code determination to SessionFinalizer.
- Delegates: persistence to PersistentSession (automatic, non-fatal).
- Must not: directly modify session state fields (use PersistentSession methods only).
- Must not: call SessionMetadataStore.Write() or EventStore.Append() directly.
- Must not: evaluate transitions or validate message semantics.
- Must not: construct AgentError (that is ErrorProcessor's responsibility).
- Must not: call PersistentSession.Done (that is EventProcessor's responsibility).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `SpectraFinder` | Project root discovery | `Find()` | — |
| `SessionInitializer` | Session bootstrap | `Initialize(workflowName, sessionID, terminationNotifier)` | Must not call after initialization |
| `PersistentSession` | State container with auto-persist | `Fail(err, notifier)`, `GetStatusSafe()`, read `ID`, `WorkflowName` | Must not call `Run()`, `Done()`, must not modify fields directly, must not call stores directly |
| `RuntimeSocketManager` | Socket lifecycle | `CreateSocket()`, `Listen(handler)`, `DeleteSocket()` | Must not construct via struct literal |
| `MessageRouter` | Message dispatch | Pass as MessageHandler to Listen | Must not invoke Handle directly |
| `TransitionToNode` | Node dispatch | `Transition(targetNodeName, message)` | Must not access internal state |
| `SessionFinalizer` | Final reporting | `Finalize(persistentSession)` returning exit code | Must not call before cleanup |
| `RuntimeError` | Error entity | Construct via `NewRuntimeError` for post-session failures | — |
| `Logger` | Structured logging | `Info(msg, args...)`, `Warn(msg, args...)`, `Error(msg, args...)` | Must not use for session status output (SessionFinalizer's job) |
| `WorkflowDefinition` | Configuration source | Read `EntryNode()` for initial dispatch | Must not modify |

Construction constraint: Runtime is an exported function `Run(workflowName string, sessionID string, logger logger.Logger) (int, error)`. It is not a struct. All dependencies are constructed internally during execution. Logger is the only externally injected dependency. `sessionID` is an optional parameter: empty string means SessionInitializer will auto-generate a UUID.

## Behavior

### Initialization and Bootstrap

1. Runtime is invoked by `spectra run` with inputs: `workflowName` (string), `sessionID` (string, may be empty), and `logger` (logger.Logger).
2. Calls `SpectraFinder.Find()` to locate the `.spectra` directory and obtain the `projectRoot` absolute path.
3. If `SpectraFinder.Find()` returns an error, returns `(1, error)` with message: `"failed to locate project root: <error>"`. No resources created.
4. Creates a buffered channel `terminationNotifier` with capacity 2 (`make(chan struct{}, 2)`).
5. Constructs pre-session dependencies using `projectRoot`:
   1. `WorkflowDefinitionLoader` (requires `projectRoot`)
   2. `SessionDirectoryManager` (requires `projectRoot`)
6. If any dependency construction fails, returns `(1, error)` with message: `"failed to initialize runtime dependencies: <error>"`.
7. Constructs `SessionInitializer` with `projectRoot`, `WorkflowDefinitionLoader`, `SessionDirectoryManager`, and `logger`.
8. Invokes `SessionInitializer.Initialize(workflowName, sessionID, terminationNotifier)` which returns an `InitResult` containing `PersistentSession`, `WorkflowDefinition`, and `Error`.
9. If `InitResult.Error != nil` and `InitResult.PersistentSession == nil` (failure before session entity construction), returns `(1, error)` with message: `"failed to initialize session: <error>"`. SessionFinalizer is not invoked.
10. If `InitResult.Error != nil` and `InitResult.PersistentSession != nil` (failure after session entity construction), proceeds to cleanup and SessionFinalizer (steps 33-39), then returns the exit code from SessionFinalizer and error: `"failed to initialize session: <error>"`.

### Post-Session Dependency Construction

11. After SessionInitializer returns successfully (InitResult.Error == nil), Runtime constructs post-session dependencies using the PersistentSession and WorkflowDefinition:
    1. `AgentDefinitionLoader` (requires `projectRoot`)
    2. `RuntimeSocketManager` via `NewRuntimeSocketManager(projectRoot, persistentSession.ID, logger)`
    3. `AgentInvoker` (requires `persistentSession.ID`, `projectRoot`)
    4. `TransitionToNode` (requires `PersistentSession`, `WorkflowDefinition`, `AgentDefinitionLoader`, `AgentInvoker`)
    5. `EventProcessor` (requires `PersistentSession`, `WorkflowDefinition`, `TransitionToNode`, `terminationNotifier`)
    6. `ErrorProcessor` (requires `PersistentSession`, `WorkflowDefinition`, `terminationNotifier`)
    7. `MessageRouter` (requires `PersistentSession`, `EventProcessor`, `ErrorProcessor`, `terminationNotifier`, `logger`)
    8. `SessionFinalizer` (requires `logger`)
12. If any post-session dependency construction fails, constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to initialize post-session dependencies"`, `Detail` containing the construction error, `SessionID=persistentSession.ID`, `FailingState=persistentSession.GetCurrentStateSafe()`, `OccurredAt=now()`. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Proceeds to cleanup and SessionFinalizer (steps 33-39), returns exit code from SessionFinalizer and error: `"failed to initialize post-session dependencies: <error>"`.

### Socket Creation and Listener Startup

13. Calls `RuntimeSocketManager.CreateSocket()` to create the runtime socket file.
14. If `CreateSocket()` returns an error, constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to create runtime socket"`, `Detail` containing the error. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Proceeds to cleanup and SessionFinalizer (steps 33-39), returns exit code and error: `"failed to create runtime socket: <error>"`.
15. Calls `RuntimeSocketManager.Listen(messageRouter)` to start the socket listener.
16. `Listen()` returns `(listenerErrCh, listenerDoneCh, err)`.
17. If `Listen()` returns a synchronous error (`err != nil`), constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to start socket listener"`, `Detail` containing the error. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Proceeds to cleanup and SessionFinalizer (steps 33-39), returns exit code and error: `"failed to start socket listener: <error>"`.

### Initial Dispatch

18. After the socket listener starts successfully, Runtime performs the initial dispatch of the entry node.
19. Retrieves the entry node name from `WorkflowDefinition.EntryNode()`.
20. Constructs a default message for the initial dispatch: `"Workflow started. You are the first node and may begin your work. To transition, run: spectra-agent event emit <type> --session-id <SessionID> [--message <message>] [--claude-session-id <UUID>] [--payload <json>]"` where `<SessionID>` is replaced with the actual session UUID.
21. Calls `TransitionToNode.Transition(entryNodeName, defaultMessage)`.
22. If `TransitionToNode.Transition()` returns an error, constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to dispatch entry node"`, `Detail` containing the error, `SessionID=persistentSession.ID`, `FailingState=entryNodeName`, `OccurredAt=now()`. Calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. Proceeds to cleanup and SessionFinalizer (steps 33-39), returns exit code and error: `"failed to dispatch entry node: <error>"`.

### Main Event Loop and Termination Signal Handling

23. After the initial dispatch succeeds, Runtime enters the main event loop using a `select` statement to monitor:
    - `terminationNotifier`: session termination notifications from PersistentSession lifecycle methods (Done/Fail)
    - `listenerErrCh`: fatal listener errors from RuntimeSocketManager
    - OS signal channel (`signalCh`): SIGINT, SIGTERM
24. Runtime registers OS signal handling using `signal.Notify()` for SIGINT and SIGTERM. On Windows, only SIGINT is registered if SIGTERM is unavailable.
25. Runtime stores the received OS signal (if any) in a variable `receivedSignal` (type `os.Signal`, initially `nil`).
26. Runtime's main loop blocks on `select` waiting for the first signal:
    - **Case: `<-terminationNotifier`** — Session reached terminal status (completed/failed). Logs: `"received session termination notification"`. Proceeds to cleanup.
    - **Case: `err := <-listenerErrCh`** — Fatal listener error. Logs: `"listener error: <error>"`. If `PersistentSession.GetStatusSafe()` is not "completed" or "failed", constructs a RuntimeError with `Issuer="Runtime"`, `Message="listener error"`, `Detail` containing the error, and calls `PersistentSession.Fail(runtimeError, terminationNotifier)`. If status is already terminal, skips Fail. Proceeds to cleanup.
    - **Case: `sig := <-signalCh`** — OS signal received. Stores `sig` in `receivedSignal`. Logs: `"received signal <signal-name>, initiating graceful shutdown"`. Does **not** call `PersistentSession.Fail()`. Proceeds to cleanup.
27. After receiving the first termination signal, Runtime does not continue monitoring other channels.

### Grace Period and Second Signal

28. After receiving a termination signal, Runtime starts a 5-second grace period timer for cleanup operations.
29. Runtime sets up monitoring for a second OS signal in a separate goroutine during the grace period.
30. If cleanup (steps 33-39) does not complete within 5 seconds, logs: `"cleanup exceeded 5 second grace period, forcing exit"` and returns `(1, error)` with message: `"cleanup timeout"`. Partial cleanup is acceptable.
31. If a second OS signal is received during the grace period, logs: `"received second signal, forcing exit"` and returns `(1, error)` with message: `"forced exit by second signal"`. No further cleanup.

### Cleanup and SessionFinalizer

32. Runtime stops OS signal notification (`signal.Stop()`).
33. Calls `RuntimeSocketManager.DeleteSocket()` to stop the listener, close all active connections, and delete the socket file.
34. `DeleteSocket()` is idempotent and does not return an error. Even if socket deletion fails internally, Runtime continues.
35. Waits for `listenerDoneCh` to close (using `<-listenerDoneCh` or `select` with a 2-second sub-timeout) to ensure the listener goroutine has fully exited.
36. If waiting for `listenerDoneCh` exceeds 2 seconds, logs: `"listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer"` and continues without waiting further.
37. Invokes `SessionFinalizer.Finalize(persistentSession)` which returns an exit code (int).
38. SessionFinalizer logs the final session status via Logger. It does not return errors.
39. Runtime determines the final return value:
    - If `receivedSignal != nil` (OS signal terminated the session): returns `(exitCode, error)` where error is `"session terminated by signal <signal-name>"` and exitCode is from SessionFinalizer (which returns 1 for non-terminal status).
    - If `receivedSignal == nil` and SessionFinalizer returns exit code 0: returns `(0, nil)`.
    - If `receivedSignal == nil` and SessionFinalizer returns exit code 1: returns `(1, error)` where error is `"session failed: <persistentSession.GetErrorSafe().Error()>"` if error is non-nil, or `"session terminated with non-terminal status"` otherwise.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| workflowName | string | Non-empty, must reference a valid workflow definition file | Yes |
| sessionID | string | Valid UUID format if non-empty; empty string means auto-generate | No |
| logger | logger.Logger | Non-nil Logger interface implementation | Yes |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| exitCode | int | 0 on success, 1 on failure. Determined by SessionFinalizer when session exists, or hardcoded 1 for early failures. |
| error | error | Nil on success. Non-nil with descriptive message on failure. Used by `spectra run` for stderr output. |

### Error Cases

| Error Message Format | Exit Code | Description |
|---------------------|-----------|-------------|
| `"failed to locate project root: <error>"` | 1 | SpectraFinder failed |
| `"failed to initialize runtime dependencies: <error>"` | 1 | Pre-session dependency construction failed |
| `"failed to initialize session: <error>"` | 1 | SessionInitializer failed (exit code from SessionFinalizer if session exists) |
| `"failed to initialize post-session dependencies: <error>"` | from SessionFinalizer | Post-session dependency construction failed |
| `"failed to create runtime socket: <error>"` | from SessionFinalizer | CreateSocket failed |
| `"failed to start socket listener: <error>"` | from SessionFinalizer | Listen failed |
| `"failed to dispatch entry node: <error>"` | from SessionFinalizer | TransitionToNode failed for entry node |
| `"session failed: <error message>"` | from SessionFinalizer | Session reached "failed" status |
| `"session terminated by signal <signal-name>"` | from SessionFinalizer | OS signal terminated runtime |
| `"session terminated with non-terminal status"` | from SessionFinalizer | Unexpected non-terminal status |
| `"cleanup timeout"` | 1 | Grace period exceeded |
| `"forced exit by second signal"` | 1 | Second OS signal forced exit |

## Invariants

1. **Single Entry Point**: Runtime is invoked once per `spectra run` and manages exactly one session lifecycle.

2. **Function Form**: Runtime is an exported function, not a struct. All dependencies are constructed internally.

3. **Logger and SessionID Injection**: Logger and sessionID are the only externally provided dependencies. All other dependencies are constructed within Runtime using projectRoot and session UUID. Runtime passes sessionID through to SessionInitializer without validation.

4. **ProjectRoot Discovery**: Runtime must call `SpectraFinder.Find()` at the beginning of execution. projectRoot is not passed as input.

5. **Termination Notifier Capacity**: Runtime must create `terminationNotifier` with buffer capacity exactly 2 to prevent blocking senders.

6. **Dependency Construction Order**: Pre-session dependencies are constructed before SessionInitializer. Post-session dependencies (requiring session UUID) are constructed after SessionInitializer returns successfully.

7. **First Signal Only**: Runtime processes only the first termination signal from `terminationNotifier`, `listenerErrCh`, or `signalCh`. Subsequent signals on these channels are not processed in the main loop.

8. **PersistentSession.Fail on Runtime Errors**: If Runtime encounters errors after PersistentSession is constructed (socket, listener, initial dispatch, listener error), Runtime must construct a RuntimeError and call PersistentSession.Fail before cleanup.

9. **No PersistentSession.Fail on OS Signals**: When an OS signal is received, Runtime must not call PersistentSession.Fail. Session status remains unchanged (non-terminal).

10. **Cleanup Order**: Runtime must perform cleanup in order: (1) signal.Stop, (2) DeleteSocket, (3) wait listenerDoneCh, (4) SessionFinalizer.

11. **SessionFinalizer Invocation Condition**: Runtime invokes SessionFinalizer if and only if PersistentSession is non-nil. If initialization fails before PersistentSession construction, SessionFinalizer is not invoked.

12. **Listener Shutdown Barrier**: Runtime must wait for `listenerDoneCh` before invoking SessionFinalizer (2-second sub-timeout enforced).

13. **Grace Period Enforcement**: A 5-second grace period governs all cleanup operations after the first termination signal.

14. **Second Signal Force Exit**: A second OS signal during the grace period causes immediate return without further cleanup.

15. **Idempotent Cleanup**: DeleteSocket is idempotent. Runtime may call it even if socket was never created or already deleted (no-op).

16. **Channel Lifecycle**: `terminationNotifier` is never closed by any component. `listenerErrCh` is never closed by RuntimeSocketManager. `listenerDoneCh` is closed exactly once by RuntimeSocketManager.

17. **Initial Dispatch**: Runtime must dispatch the entry node after successful session initialization and listener startup. The default message includes the session UUID and transition instructions.

18. **Exit Code Source**: Exit code is determined by SessionFinalizer when PersistentSession exists. For early failures (before PersistentSession), exit code is hardcoded to 1.

19. **Platform Signal Compatibility**: On Windows, only SIGINT is registered if SIGTERM is unavailable.

20. **No PersistentSession.Done or PersistentSession.Run**: Runtime must not call PersistentSession.Done (EventProcessor's responsibility) or PersistentSession.Run (SessionInitializer's responsibility).

21. **Thread-Safe Session Access**: Runtime uses PersistentSession's exported methods for all state access.

## Edge Cases

- **Condition**: SpectraFinder fails to locate `.spectra` directory (project not initialized).
  **Expected**: Returns `(1, "failed to locate project root: <error>")` without creating any resources. SessionFinalizer not invoked.

- **Condition**: SessionInitializer fails before PersistentSession construction (workflow not found, directory creation failed).
  **Expected**: Returns `(1, "failed to initialize session: <error>")`. SessionFinalizer not invoked.

- **Condition**: SessionInitializer fails after PersistentSession construction (timeout, Run() failure).
  **Expected**: Proceeds to cleanup and SessionFinalizer. Returns exit code from SessionFinalizer and error.

- **Condition**: Post-session dependency construction fails.
  **Expected**: Constructs RuntimeError, calls PersistentSession.Fail, proceeds to cleanup and SessionFinalizer. Returns exit code and error.

- **Condition**: RuntimeSocketManager.CreateSocket fails (socket already exists, permission denied).
  **Expected**: Constructs RuntimeError, calls PersistentSession.Fail, proceeds to cleanup and SessionFinalizer. Returns exit code and error.

- **Condition**: RuntimeSocketManager.Listen fails (bind error).
  **Expected**: Constructs RuntimeError, calls PersistentSession.Fail, proceeds to cleanup and SessionFinalizer. Returns exit code and error.

- **Condition**: Initial dispatch of entry node fails (agent definition not found, AgentInvoker fails).
  **Expected**: Constructs RuntimeError with FailingState=entryNodeName, calls PersistentSession.Fail, proceeds to cleanup and SessionFinalizer. Returns exit code and error.

- **Condition**: Session completes successfully (exit transition reached).
  **Expected**: EventProcessor calls PersistentSession.Done, notification sent to terminationNotifier. Runtime receives notification, proceeds to cleanup, SessionFinalizer returns 0. Runtime returns `(0, nil)`.

- **Condition**: Agent reports error via spectra-agent.
  **Expected**: ErrorProcessor calls PersistentSession.Fail, notification sent. Runtime receives notification, proceeds to cleanup, SessionFinalizer returns 1. Runtime returns `(1, "session failed: <message>")`.

- **Condition**: User presses Ctrl+C (SIGINT) while session is running.
  **Expected**: Runtime receives SIGINT, stores in receivedSignal, logs shutdown message, does not call Fail, proceeds to cleanup and SessionFinalizer. SessionFinalizer logs non-terminal status warning. Returns `(1, "session terminated by signal interrupt")`.

- **Condition**: User presses Ctrl+C twice (second signal during grace period).
  **Expected**: First SIGINT begins cleanup. Second SIGINT detected during grace period. Logs force exit message. Returns `(1, "forced exit by second signal")` immediately.

- **Condition**: Cleanup exceeds 5-second grace period.
  **Expected**: Timer fires. Logs timeout warning. Returns `(1, "cleanup timeout")` immediately. Partial cleanup acceptable.

- **Condition**: Listener error occurs while session is running.
  **Expected**: Runtime receives error from listenerErrCh, constructs RuntimeError, calls PersistentSession.Fail, proceeds to cleanup and SessionFinalizer.

- **Condition**: Listener error occurs concurrently with session completion (race).
  **Expected**: First signal wins in select. If terminationNotifier wins, cleanup proceeds normally. If listenerErrCh wins, Runtime checks status — if already "completed", skips Fail.

- **Condition**: PersistentSession.Fail returns error (session already failed by concurrent operation).
  **Expected**: Runtime logs warning: `"attempted to fail session but session already in terminal state: <error>"` and proceeds to cleanup. SessionFinalizer reports the first error (first-error-wins policy).

- **Condition**: listenerDoneCh wait exceeds 2-second sub-timeout.
  **Expected**: Logs warning and proceeds to SessionFinalizer. Listener goroutine may still be running in background.

- **Condition**: listenerDoneCh already closed when Runtime waits (DeleteSocket finished fast).
  **Expected**: `<-listenerDoneCh` returns immediately. Proceeds to SessionFinalizer.

- **Condition**: SIGTERM received on Unix while session is running.
  **Expected**: Same as SIGINT but signal name is "terminated". Returns `(1, "session terminated by signal terminated")`.

- **Condition**: SIGTERM on Windows (not available).
  **Expected**: Only SIGINT registered. SIGTERM is not caught. Process terminated by OS without graceful shutdown.

- **Condition**: Multiple runtime processes started concurrently with same workflow.
  **Expected**: Each generates unique session UUID via SessionInitializer. No conflict.

- **Condition**: Session is in "initializing" status when SIGINT is received (signal during SessionInitializer timeout path where PersistentSession exists but Run() failed).
  **Expected**: Runtime receives signal, does not call Fail, proceeds to cleanup. SessionFinalizer logs non-terminal status. Returns `(1, "session terminated by signal interrupt")`.

- **Condition**: DeleteSocket fails internally (permission error on socket file).
  **Expected**: DeleteSocket logs warning internally. Runtime continues to wait for listenerDoneCh and invoke SessionFinalizer.

- **Condition**: SessionFinalizer called with PersistentSession whose status is "running" and no signal received (should not happen in normal operation but covered defensively).
  **Expected**: SessionFinalizer logs non-terminal status warning. Returns exit code 1. Runtime returns `(1, "session terminated with non-terminal status")`.

## Related

- [SessionInitializer](./session_initializer.md) — Initializes session with 30-second timeout
- [PersistentSession](./persistent_session.md) — State container with automatic persistence
- [SessionFinalizer](./session_finalizer.md) — Final status logging and exit code determination
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) — Socket lifecycle management
- [MessageRouter](./message_router.md) — Routes incoming messages to processors
- [EventProcessor](./event_processor.md) — Processes event messages
- [ErrorProcessor](./error_processor.md) — Processes error messages
- [TransitionToNode](./transition_to_node.md) — Executes node dispatch and state update
- [SpectraFinder](../storage/spectra_finder.md) — Locates project root
- [Logger](../logger/logger.md) — Structured logging interface
