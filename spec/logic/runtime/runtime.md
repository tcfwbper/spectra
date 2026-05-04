# Runtime

## Overview

Runtime is the main orchestrator invoked by the `spectra run` command. It receives a WorkflowName as input, bootstraps all runtime dependencies, initializes a new session, creates and manages the runtime socket, runs the main event loop to process workflow events, handles termination signals (session completion/failure, OS signals, socket errors), performs cleanup (socket deletion, listener shutdown), and invokes SessionFinalizer to print the final session status. Runtime returns an error to `spectra run` indicating the outcome. Runtime is the single entry point for workflow execution and coordinates the lifecycle of all runtime components. Runtime does not manage session directory creation or deletion (handled by SessionInitializer and `spectra clear` respectively).

## Behavior

### Initialization and Bootstrap

1. Runtime is invoked by `spectra run` with a single input: `workflowName` (string).
2. Runtime calls `SpectraFinder.Find()` (stateless package-level function) to locate the `.spectra` directory and obtain the `projectRoot` absolute path.
3. If `SpectraFinder.Find()` returns an error, Runtime immediately returns the error to `spectra run` without creating any resources. Error format: `"failed to locate project root: <error>"`.
4. Runtime creates a buffered channel `terminationNotifier` with capacity 2 (`make(chan struct{}, 2)`). This channel is used to receive termination notifications from Session lifecycle methods (Done/Fail) and SessionInitializer timeout. The capacity of 2 ensures that concurrent notifications (e.g., SessionInitializer timeout + Session.Fail due to socket error) do not block the sender.
5. Runtime constructs all runtime dependencies in the following order, using the obtained `projectRoot`:
   1. `WorkflowDefinitionLoader` (requires `projectRoot`)
   2. `SessionDirectoryManager` (requires `projectRoot`)
   3. `AgentDefinitionLoader` (requires `projectRoot`)
   4. `SessionInitializer` (requires `projectRoot`, `WorkflowDefinitionLoader`, `SessionDirectoryManager`)
6. If any dependency construction fails, Runtime returns an error to `spectra run` without creating any resources. Error format: `"failed to initialize runtime dependencies: <error>"`.
7. Runtime invokes `SessionInitializer.Initialize(workflowName, terminationNotifier)` to create and initialize the session.
8. `SessionInitializer.Initialize()` performs the following:
   - Generates a session UUID
   - Loads the workflow definition
   - Creates the session directory structure
   - Constructs the Session entity with Status="initializing"
   - Initializes EventStore and SessionMetadataStore
   - Persists initial session metadata to disk
   - Calls `Session.Run(terminationNotifier)` to transition Status to "running"
   - Enforces a 30-second timeout; if exceeded, triggers a RuntimeError via `Session.Fail()` or sends an early-timeout notification if Session entity has not yet been constructed
9. If `SessionInitializer.Initialize()` returns an error and `session == nil` (initialization failed before Session entity was constructed), Runtime returns the error to `spectra run` without calling SessionFinalizer. Error format: `"failed to initialize session: <error>"`.
10. If `SessionInitializer.Initialize()` returns an error but `session != nil` (initialization failed after Session entity was constructed, e.g., timeout, Session.Run failure), Runtime proceeds to cleanup and SessionFinalizer (steps 27-30) to print the session status, then returns the error.

### Socket Creation and Listener Startup

11. After SessionInitializer returns successfully with `session != nil` and `session.Status == "running"`, Runtime constructs additional dependencies that require the session UUID:
    1. `SessionMetadataStore` (requires `projectRoot`, `session.ID`)
    2. `EventStore` (requires `projectRoot`, `session.ID`)
    3. `RuntimeSocketManager` (requires `projectRoot`, `session.ID`)
    4. `AgentInvoker` (stateless utility, no construction needed; used by TransitionToNode)
    5. `TransitionToNode` (requires `session`, `WorkflowDefinitionLoader`, `AgentDefinitionLoader`, `AgentInvoker`, `terminationNotifier`)
    6. `EventProcessor` (requires `session`, `WorkflowDefinitionLoader`, `TransitionToNode`, `terminationNotifier`)
    7. `ErrorProcessor` (requires `session`, `WorkflowDefinitionLoader`, `terminationNotifier`)
    8. `MessageRouter` (requires `session`, `EventProcessor`, `ErrorProcessor`, `terminationNotifier`)
    9. `SessionFinalizer` (stateless utility, no construction needed; invoked at cleanup time)
12. If any post-session dependency construction fails, Runtime constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to initialize post-session dependencies"`, `Detail` containing the construction error, `SessionID=session.ID`, `FailingState=session.CurrentState`, and `OccurredAt` set to the current POSIX timestamp. Runtime calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status, then proceeds to cleanup and SessionFinalizer (steps 27-30), then returns an error: `"failed to initialize post-session dependencies: <error>"`.
13. Runtime calls `RuntimeSocketManager.CreateSocket()` to create the runtime socket file at `.spectra/sessions/<sessionUUID>/runtime.sock` with permissions 0600.
14. If `RuntimeSocketManager.CreateSocket()` returns an error (socket file already exists, permission denied, disk full), Runtime constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to create runtime socket"`, `Detail` containing the socket creation error, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status, then proceeds to cleanup and SessionFinalizer (steps 27-30), then returns an error: `"failed to create runtime socket: <error>"`.
15. Runtime calls `RuntimeSocketManager.Listen(messageHandler)` to start the socket listener. The `messageHandler` is the `MessageRouter.RouteMessage` method (MessageRouter implements the MessageHandler callback interface).
16. `RuntimeSocketManager.Listen()` returns three values: `(listenerErrCh <-chan error, listenerDoneCh <-chan struct{}, err error)`.
    - `listenerErrCh`: buffered channel (capacity 1) that receives at most one fatal listener error after the listener has started. **Never closed** by RuntimeSocketManager.
    - `listenerDoneCh`: unbuffered channel that is closed exactly once when the listener goroutine has fully exited (after DeleteSocket is called or after a fatal error). This serves as a shutdown barrier.
    - `err`: synchronous error returned if the initial bind/listen fails before the listener goroutine starts.
17. If `RuntimeSocketManager.Listen()` returns a synchronous error (`err != nil`), Runtime constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to start socket listener"`, `Detail` containing the listener error, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status, then proceeds to cleanup and SessionFinalizer (steps 27-30), then returns an error: `"failed to start socket listener: <error>"`.

### Main Event Loop and Termination Signal Handling

18. After the socket listener starts successfully, Runtime enters the main event loop. Runtime uses a `select` statement to monitor multiple termination sources:
    - `terminationNotifier`: receives session termination notifications from Session lifecycle methods (Done/Fail) or SessionInitializer timeout
    - `listenerErrCh`: receives fatal listener errors from RuntimeSocketManager
    - OS signal channel (SIGINT, SIGTERM): receives OS signals for graceful shutdown
19. Runtime sets up OS signal handling using `signal.Notify()` to catch SIGINT and SIGTERM. On Windows, SIGTERM may not be available; Runtime should handle this platform difference gracefully (e.g., only register SIGINT on Windows).
20. Runtime stores the received OS signal (if any) in a variable `receivedSignal` (type `os.Signal`, initially `nil`) for use in error message construction later.
21. Runtime's main loop blocks on `select` waiting for the **first** signal from any of these sources:
    ```
    select {
    case <-terminationNotifier:
        // Session reached terminal status (completed/failed) or SessionInitializer timeout
    case err := <-listenerErrCh:
        // Fatal listener error (e.g., accept loop failure)
    case sig := <-signalCh:
        // OS signal (SIGINT/SIGTERM)
        receivedSignal = sig
    }
    ```
22. **Case 1: Session Termination Notification** (`<-terminationNotifier`):
    - Runtime logs: `"received session termination notification"`
    - Runtime proceeds directly to cleanup (step 28)
23. **Case 2: Listener Error** (`<-listenerErrCh`):
    - Runtime logs: `"listener error: <error>"`
    - If `session.Status` is not "completed" or "failed" (checked via `Session.GetStatusSafe()`), Runtime constructs a RuntimeError with `Issuer="Runtime"`, `Message="listener error"`, `Detail` containing the listener error, calls `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status. `Session.Fail()` sends a notification to `terminationNotifier`, but Runtime ignores it (first signal already received).
    - If `session.Status` is already "completed" or "failed" (race condition: session terminated concurrently with listener error), Runtime skips calling `Session.Fail()` (session already terminal).
    - Runtime proceeds to cleanup (step 28)
24. **Case 3: OS Signal** (`<-signalCh`):
    - Runtime receives the OS signal and stores it in `receivedSignal`
    - Runtime logs: `"received signal <signal-name>, initiating graceful shutdown"` where `<signal-name>` is the string representation of the signal (e.g., "interrupt" for SIGINT, "terminated" for SIGTERM)
    - Runtime does **not** call `Session.Fail()`. The session status remains in its current state ("initializing" or "running"). SessionFinalizer will handle non-terminal status by printing: `"Session <SessionID> terminated with status '<Status>'. Workflow: <WorkflowName>"`.
    - Runtime proceeds to cleanup (step 28)
25. After receiving the first termination signal, Runtime **does not** continue monitoring other channels. Only the first signal is processed.
26. Runtime also sets up a grace period timer for cleanup operations. After receiving a termination signal, Runtime starts a 5-second timer. If cleanup (steps 28-33) does not complete within 5 seconds, Runtime logs a warning: `"cleanup exceeded 5 second grace period, forcing exit"` and exits immediately with the appropriate error or exit code. Partial cleanup (e.g., socket deletion succeeded but SessionFinalizer not yet called) is acceptable in this case.
27. If a second OS signal (e.g., second Ctrl+C) is received during the grace period, Runtime immediately logs: `"received second signal, forcing exit"` and exits without further cleanup.

### Cleanup and SessionFinalizer

28. Runtime calls `RuntimeSocketManager.DeleteSocket()` to stop the listener, close all active connections, and delete the socket file.
29. `RuntimeSocketManager.DeleteSocket()` is idempotent and logs warnings on failure but does not return an error. Even if socket deletion fails, Runtime continues to the next cleanup step.
30. Runtime waits for `listenerDoneCh` to close (using `<-listenerDoneCh` or `select` with grace period timer) to ensure the listener goroutine has fully exited before proceeding. This prevents race conditions where SessionFinalizer prints status while connections are still being processed.
31. If waiting for `listenerDoneCh` exceeds 2 seconds (sub-timeout within the 5-second grace period), Runtime logs a warning: `"listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer"` and continues without waiting further.
32. Runtime invokes `SessionFinalizer.Finalize(session)` to print the final session status to stdout (if completed) or stderr (if failed or non-terminal).
33. `SessionFinalizer.Finalize()` does not return an error. All print operations are best-effort.
34. After SessionFinalizer completes, Runtime determines the return value based on `session.Status` (read via `Session.GetStatusSafe()` for thread safety) and `receivedSignal`:
    - If `session.Status == "completed"`: Runtime returns `nil` (success)
    - If `session.Status == "failed"`: Runtime returns an error: `"session failed: <session.Error.Message>"`
    - If `session.Status == "initializing"` or `"running"` (non-terminal, due to OS signal) and `receivedSignal != nil`:
      - If `receivedSignal` is SIGINT (syscall.SIGINT or os.Interrupt): Runtime returns an error: `"session terminated by signal SIGINT"`
      - If `receivedSignal` is SIGTERM (syscall.SIGTERM): Runtime returns an error: `"session terminated by signal SIGTERM"`
      - For any other signal: Runtime returns an error: `"session terminated by signal <signal-name>"`
    - If `session.Status == "initializing"` or `"running"` but `receivedSignal == nil` (should not happen in normal operation): Runtime returns an error: `"session terminated with status '<Status>'"`

### Error Propagation to spectra run

35. Runtime returns a Go `error` type to `spectra run`. The `spectra run` command is responsible for converting the error to an appropriate exit code based on the error message:
    - `nil` error → exit code 0 (success)
    - Error message `"session terminated by signal SIGINT"` → exit code 130 (128 + 2, standard Unix convention)
    - Error message `"session terminated by signal SIGTERM"` → exit code 143 (128 + 15, standard Unix convention)
    - All other non-nil errors → exit code 1 (generic failure)
36. The error message returned by Runtime includes sufficient context for debugging (session ID, workflow name, error details). `spectra run` may choose to print the error message to stderr before exiting.

## Inputs

### For Invocation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowName | string | Non-empty, PascalCase, must reference a valid workflow definition file | Yes |

## Outputs

### Success Case

| Field | Type | Description |
|-------|------|-------------|
| Error | nil | Session completed successfully |

### Error Cases

All error cases return a non-nil Go `error`:

| Error Message Format | Description |
|---------------------|-------------|
| `"failed to locate project root: <error>"` | SpectraFinder failed to locate `.spectra` directory |
| `"failed to initialize runtime dependencies: <error>"` | Dependency construction failed before session initialization |
| `"failed to initialize session: <error>"` | SessionInitializer failed before Session entity was constructed (no SessionFinalizer invocation) |
| `"failed to initialize post-session dependencies: <error>"` | Dependency construction failed after session was created (SessionFinalizer invoked) |
| `"failed to create runtime socket: <error>"` | RuntimeSocketManager.CreateSocket failed (SessionFinalizer invoked) |
| `"failed to start socket listener: <error>"` | RuntimeSocketManager.Listen failed (SessionFinalizer invoked) |
| `"session failed: <session.Error.Message>"` | Session reached "failed" status (SessionFinalizer invoked) |
| `"session terminated with status '<Status>'"` | Session terminated due to OS signal with non-terminal status (SessionFinalizer invoked) |

## Invariants

1. **Single Entry Point**: Runtime is the single entry point for workflow execution. It is invoked once per `spectra run` command and manages exactly one session lifecycle.

2. **ProjectRoot Discovery**: Runtime must call `SpectraFinder.Find()` at the beginning of execution to obtain `projectRoot`. `projectRoot` is not passed as input from `spectra run`.

3. **Dependency Construction Order**: Runtime must construct dependencies in the order specified in step 5 and step 11. Dependencies requiring the session UUID must be constructed after SessionInitializer returns successfully.

4. **Termination Notifier Capacity**: Runtime must create `terminationNotifier` with buffer capacity >= 2 to prevent blocking senders. SessionInitializer validates this capacity.

5. **First Signal Only**: Runtime processes only the **first** termination signal received from `terminationNotifier`, `listenerErrCh`, or `signalCh`. Subsequent signals are ignored.

6. **Session.Fail on Runtime Errors**: If Runtime encounters errors after the Session entity is constructed (socket creation failure, listener startup failure, listener error), Runtime must construct a RuntimeError and call `Session.Fail(runtimeError, terminationNotifier)` to transition the session to "failed" status before cleanup.

7. **No Session.Fail on OS Signals**: When an OS signal (SIGINT/SIGTERM) is received, Runtime must **not** call `Session.Fail()`. The session status remains unchanged. Runtime stores the received signal in `receivedSignal` for error message construction. SessionFinalizer handles non-terminal status by printing: `"Session <SessionID> terminated with status '<Status>'. Workflow: <WorkflowName>"`.

8. **Cleanup Order**: Runtime must perform cleanup in the following order: (1) DeleteSocket, (2) wait for listenerDoneCh, (3) invoke SessionFinalizer.

9. **SessionFinalizer Invocation**: Runtime must invoke SessionFinalizer if and only if the Session entity was constructed (i.e., `session != nil`). If SessionInitializer fails before constructing the Session entity, SessionFinalizer is **not** invoked.

10. **Listener Shutdown Barrier**: Runtime must wait for `listenerDoneCh` to close before invoking SessionFinalizer. This ensures all in-flight message processing has completed. A 2-second sub-timeout is enforced to prevent indefinite blocking.

11. **Grace Period Enforcement**: Runtime enforces a 5-second grace period for cleanup operations (steps 28-34). If cleanup exceeds this duration, Runtime logs a warning and exits immediately.

12. **Second Signal Force Exit**: If a second OS signal is received during the grace period, Runtime immediately exits without further cleanup.

13. **Idempotent Cleanup**: `RuntimeSocketManager.DeleteSocket()` is idempotent. Calling it multiple times or when the socket does not exist is safe.

14. **Best-Effort Cleanup**: Socket deletion failures and SessionFinalizer failures do not prevent Runtime from proceeding to return an error. All cleanup operations are best-effort.

15. **Thread-Safe Session Access**: Runtime must use Session's thread-safe methods (`GetStatusSafe()`, `GetCurrentStateSafe()`, `Fail()`) when accessing session state from the main loop. Direct field access is prohibited.

16. **Channel Lifecycle Convention**: `terminationNotifier` is **never closed** by any component. It is sent on at most twice per session lifetime (once by Done xor Fail, optionally once more by an early-timeout path that does not call Fail). `listenerErrCh` is **never closed** by RuntimeSocketManager. `listenerDoneCh` is closed exactly once by RuntimeSocketManager when the listener goroutine exits.

17. **Error Context Sufficiency**: All errors returned by Runtime must include sufficient context for debugging: session ID (if available), workflow name, signal type (if terminated by signal), and the underlying error message.

18. **Platform Signal Compatibility**: Runtime must handle OS signal differences between Unix-like systems (SIGINT, SIGTERM) and Windows (SIGINT only). On Windows, Runtime should only register SIGINT if SIGTERM is not available.

19. **Signal-Specific Error Messages**: When an OS signal terminates the session, Runtime must return an error message that identifies the specific signal type: `"session terminated by signal SIGINT"` for SIGINT (syscall.SIGINT or os.Interrupt), `"session terminated by signal SIGTERM"` for SIGTERM (syscall.SIGTERM), or `"session terminated by signal <signal-name>"` for other signals. This enables `spectra run` to map signals to appropriate exit codes.

## Edge Cases

- **Condition**: SpectraFinder fails to locate `.spectra` directory (project not initialized).
  **Expected**: Runtime returns an error: `"failed to locate project root: spectra not initialized"` without creating any resources. SessionFinalizer is not invoked.

- **Condition**: WorkflowDefinitionLoader fails to load the workflow definition (file not found, parse error).
  **Expected**: SessionInitializer returns an error with `session == nil`. Runtime returns an error: `"failed to initialize session: failed to load workflow definition: <error>"` without calling SessionFinalizer.

- **Condition**: SessionDirectoryManager fails to create the session directory (parent directory does not exist, permission denied).
  **Expected**: SessionInitializer returns an error with `session == nil`. Runtime returns an error: `"failed to initialize session: failed to create session directory: <error>"` without calling SessionFinalizer.

- **Condition**: SessionInitializer timeout fires after Session entity is constructed but before Session.Run succeeds.
  **Expected**: SessionInitializer's timeout handler calls `Session.Fail(runtimeError, terminationNotifier)` to transition Status to "failed" and sends a notification. SessionInitializer returns an error with `session != nil`. Runtime receives the termination notification in the main loop (or before entering the loop if initialization is still in progress), proceeds to cleanup, invokes SessionFinalizer to print the failure, and returns an error: `"failed to initialize session: session initialization timeout exceeded 30 seconds"`.

- **Condition**: SessionInitializer timeout fires before Session entity is constructed (early timeout).
  **Expected**: SessionInitializer's timeout handler sets `timedOutEarly`, sends a notification to `terminationNotifier`, and returns an error with `session == nil`. Runtime does not enter the main loop, does not call SessionFinalizer, and returns an error: `"failed to initialize session: session initialization timeout exceeded 30 seconds before session entity was constructed"`.

- **Condition**: RuntimeSocketManager.CreateSocket fails because the socket file already exists.
  **Expected**: Runtime constructs a RuntimeError, calls `Session.Fail(runtimeError, terminationNotifier)`, proceeds to cleanup, invokes SessionFinalizer to print the failure, and returns an error: `"failed to create runtime socket: runtime socket file already exists: <path>. This may indicate a previous runtime process did not clean up properly or another runtime is currently active. Verify no runtime process is running (e.g., ps aux | grep spectra), then remove the socket file manually with: rm <path>"`.

- **Condition**: RuntimeSocketManager.Listen fails due to bind/listen error.
  **Expected**: Runtime constructs a RuntimeError, calls `Session.Fail(runtimeError, terminationNotifier)`, proceeds to cleanup, invokes SessionFinalizer to print the failure, and returns an error: `"failed to start socket listener: failed to listen on runtime socket: <error>"`.

- **Condition**: Session completes successfully (workflow reaches exit transition).
  **Expected**: TransitionToNode calls `Session.Done(terminationNotifier)` to transition Status to "completed" and send a notification. Runtime receives the notification in the main loop, proceeds to cleanup, invokes SessionFinalizer to print: `"Session <SessionID> completed successfully. Workflow: <WorkflowName>"` to stdout, and returns `nil`.

- **Condition**: Agent reports an error via spectra-agent.
  **Expected**: ErrorProcessor constructs an AgentError, calls `Session.Fail(agentError, terminationNotifier)` to transition Status to "failed" and send a notification. Runtime receives the notification in the main loop, proceeds to cleanup, invokes SessionFinalizer to print the failure to stderr, and returns an error: `"session failed: <agentError.Message>"`.

- **Condition**: User presses Ctrl+C (SIGINT) while session is running.
  **Expected**: Runtime receives SIGINT on `signalCh`, stores it in `receivedSignal`, logs: `"received signal interrupt, initiating graceful shutdown"`, does **not** call `Session.Fail()`, proceeds to cleanup (DeleteSocket, wait for listenerDoneCh), invokes SessionFinalizer to print: `"Session <SessionID> terminated with status 'running'. Workflow: <WorkflowName>"` to stderr, and returns an error: `"session terminated by signal SIGINT"`.

- **Condition**: User presses Ctrl+C twice (second signal during grace period).
  **Expected**: Runtime receives the first SIGINT and begins cleanup. During cleanup, the second SIGINT is received on `signalCh`. Runtime detects the second signal, logs: `"received second signal, forcing exit"`, and exits immediately without completing cleanup or calling SessionFinalizer. Exit code is 1.

- **Condition**: Cleanup exceeds the 5-second grace period.
  **Expected**: Runtime's grace period timer fires, Runtime logs: `"cleanup exceeded 5 second grace period, forcing exit"`, and exits immediately. SessionFinalizer may or may not have been called depending on where cleanup was blocked. Exit code is 1.

- **Condition**: Listener error occurs (e.g., accept loop failure) while session is running.
  **Expected**: RuntimeSocketManager sends an error on `listenerErrCh`. Runtime receives the error, logs: `"listener error: <error>"`, constructs a RuntimeError, calls `Session.Fail(runtimeError, terminationNotifier)` to transition Status to "failed", proceeds to cleanup, invokes SessionFinalizer to print the failure, and returns an error: `"session failed: listener error"`.

- **Condition**: Listener error occurs concurrently with Session.Done (race condition).
  **Expected**: Runtime's main loop receives the **first** signal (either from `terminationNotifier` or `listenerErrCh`). If `terminationNotifier` wins, Runtime proceeds to cleanup without calling `Session.Fail()`. If `listenerErrCh` wins, Runtime checks `session.Status` (via `GetStatusSafe()`). If Status is already "completed" (Session.Done finished first), Runtime skips calling `Session.Fail()` (session already terminal) and proceeds to cleanup. SessionFinalizer prints the completed status.

- **Condition**: RuntimeSocketManager.DeleteSocket fails to remove the socket file (permission denied, filesystem error).
  **Expected**: `DeleteSocket()` logs a warning: `"failed to delete runtime socket: <error>. The socket file may need to be manually removed."` but does not return an error. Runtime continues to wait for `listenerDoneCh` and invoke SessionFinalizer. The socket file remains on disk.

- **Condition**: Waiting for listenerDoneCh exceeds the 2-second sub-timeout.
  **Expected**: Runtime logs: `"listener shutdown exceeded 2 seconds, proceeding to SessionFinalizer"` and continues to invoke SessionFinalizer without waiting further. The listener goroutine may still be running in the background but is detached.

- **Condition**: SessionFinalizer panics or fails to print output (stdout/stderr closed).
  **Expected**: SessionFinalizer does not return errors and does not implement panic recovery. If it panics, the panic propagates to Runtime. Runtime should implement panic recovery around SessionFinalizer invocation (deferred recovery) to log the panic and proceed to return an error: `"session failed: <session.Error.Message>"` or `"session terminated with status '<Status>'"` depending on the session status.

- **Condition**: Multiple runtime processes are started concurrently with the same workflow name.
  **Expected**: Each generates a unique session UUID. Session directories and socket files are created independently. No conflict occurs unless two runtimes attempt to use the same session directory (UUID collision, extremely rare).

- **Condition**: Session reaches "failed" status due to agent error, and concurrently the user presses Ctrl+C.
  **Expected**: Runtime's main loop receives the **first** signal. If `terminationNotifier` wins, Runtime proceeds to cleanup with `session.Status == "failed"`. SessionFinalizer prints the failed status. If `signalCh` wins, Runtime proceeds to cleanup without calling `Session.Fail()`. `session.Status` is already "failed" (Session.Fail finished just before signal). SessionFinalizer prints the failed status. In both cases, Runtime returns an error: `"session failed: <error-message>"`.

- **Condition**: Runtime constructs a RuntimeError but Session.Fail returns an error (session already failed).
  **Expected**: Runtime logs a warning: `"attempted to fail session but session already failed: <error>"` and proceeds to cleanup. SessionFinalizer prints the first error (first-error-wins policy enforced by Session.Fail).

- **Condition**: listenerErrCh never sends an error (normal operation).
  **Expected**: Runtime's main loop never receives a signal from `listenerErrCh`. Termination is driven by `terminationNotifier` or `signalCh`. `listenerErrCh` is ignored. The channel is garbage-collected after Runtime exits.

- **Condition**: listenerDoneCh is already closed when Runtime waits for it (race condition: DeleteSocket called and listener exited before wait).
  **Expected**: `<-listenerDoneCh` returns immediately (receiving from a closed channel returns the zero value without blocking). Runtime proceeds to SessionFinalizer without delay.

- **Condition**: Session entity has Status="initializing" when SIGINT is received (signal during SessionInitializer execution).
  **Expected**: Runtime receives SIGINT, stores it in `receivedSignal`, does not call `Session.Fail()`, proceeds to cleanup, invokes SessionFinalizer to print: `"Session <SessionID> terminated with status 'initializing'. Workflow: <WorkflowName>"` to stderr, and returns an error: `"session terminated by signal SIGINT"`.

- **Condition**: User sends SIGTERM to the runtime process while session is running (Unix/Linux/macOS only).
  **Expected**: Runtime receives SIGTERM on `signalCh`, stores it in `receivedSignal`, logs: `"received signal terminated, initiating graceful shutdown"`, does **not** call `Session.Fail()`, proceeds to cleanup, invokes SessionFinalizer to print: `"Session <SessionID> terminated with status 'running'. Workflow: <WorkflowName>"` to stderr, and returns an error: `"session terminated by signal SIGTERM"`.

- **Condition**: Session entity has Status="failed" and Session.Error is a RuntimeError (not AgentError).
  **Expected**: SessionFinalizer prints the failure with RuntimeError details: `"Error: <Message>", "Issuer: <Issuer>", "State: <FailingState>", "Detail: <Detail as JSON>"`. Runtime returns an error: `"session failed: <RuntimeError.Message>"`.

- **Condition**: Runtime dependencies (WorkflowDefinitionLoader, AgentDefinitionLoader) are constructed but unused (workflow has no agents).
  **Expected**: Dependencies are constructed but not invoked. No error occurs. Session completes successfully.

## Related

- [SessionInitializer](./session_initializer.md) - Initializes session and enforces 30-second timeout
- [SessionFinalizer](./session_finalizer.md) - Prints final session status
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Manages runtime socket lifecycle
- [MessageRouter](./message_router.md) - Routes incoming messages to processors
- [EventProcessor](./event_processor.md) - Processes event messages
- [ErrorProcessor](./error_processor.md) - Processes error messages
- [TransitionToNode](./transition_to_node.md) - Executes state transitions
- [Session](../entities/session/session.md) - Session entity and lifecycle methods
- [SpectraFinder](../storage/spectra_finder.md) - Locates project root
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture and session lifecycle
