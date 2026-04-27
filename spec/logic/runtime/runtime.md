# Runtime

## Overview

Runtime is the top-level entry point and main loop for the workflow execution system. It is invoked by the CLI command `spectra run <workflow-name>` and orchestrates the entire session lifecycle: locating the project root, creating the termination notifier channel, invoking SessionInitializer to create and initialize the session, starting the socket listener (MessageRouter) in a separate goroutine, monitoring for termination signals (completion, failure, or OS kill signals), and invoking SessionFinalizer to clean up and print the final status. Runtime binds to exactly one session per invocation and exits when the session reaches a terminal state or is forcibly terminated. Runtime is responsible for graceful shutdown on OS signals (SIGINT, SIGTERM), ensuring that locks are released and the session status is properly finalized.

## Behavior

### Main Loop Flow

1. Runtime is invoked by the CLI with a single input: `workflowName` (string, provided as a command-line argument to `spectra run <workflow-name>`).
2. Runtime creates a buffered termination notifier channel: `terminationNotifier := make(chan struct{}, 2)`. The buffer size is 2 to prevent blocking when both `Session.Done()` and `Session.Fail()` send notifications, or when a timeout handler and a normal termination occur simultaneously.
3. Runtime invokes `SessionInitializer.Initialize(workflowName, terminationNotifier)` to create and initialize the session.
4. If SessionInitializer returns an error, Runtime proceeds directly to step 18 (SessionFinalizer invocation) with the partially initialized session (if available). If no session entity exists (early failure before session creation), Runtime prints an error to stderr and exits with code 1.
5. If SessionInitializer succeeds and returns a Session entity with Status="running", Runtime proceeds to start the socket listener.
6. Runtime initializes MessageRouter with the Session, EventProcessor, ErrorProcessor, and terminationNotifier.
7. Runtime starts the socket listener by calling `listenerErrCh, listenerDoneCh, syncErr := RuntimeSocketManager.Listen(MessageRouter.RouteMessage)`. `Listen()` itself spawns the accept-loop goroutine and returns immediately with two channels: `listenerErrCh` (delivers asynchronous listener errors such as `accept`/`read` failures or invalid-frame protocol errors) and `listenerDoneCh` (closed by RuntimeSocketManager when the listener goroutine has fully exited).
8. If `syncErr` is non-nil (synchronous setup failure such as `bind`/`listen` failure or socket already exists), Runtime constructs a RuntimeError with `Issuer="Runtime"`, `Message="failed to start socket listener"`, `Detail` containing the error string under key `"error"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and proceeds to step 18 (SessionFinalizer). In this case the listener goroutine was never spawned and `listenerDoneCh` is already closed.
9. Runtime sets up OS signal handling using `signal.Notify()` to capture SIGINT and SIGTERM signals.
10. Runtime enters the main monitoring loop, using `select` to wait for one of the following events:
    - Termination notification from `terminationNotifier` (sent by `Session.Done()` or `Session.Fail()`)
    - Asynchronous listener error received on `listenerErrCh`
    - OS signal (SIGINT or SIGTERM)
11. If a termination notification is received:
    - Runtime exits the monitoring loop and proceeds to step 18 (SessionFinalizer).
11a. If an asynchronous listener error is received on `listenerErrCh`:
    - Runtime constructs a RuntimeError with `Issuer="Runtime"`, `Message="runtime socket listener error"`, `Detail` containing the error string under key `"error"`, `SessionID` set to `Session.ID`, `FailingState` set to `Session.GetCurrentStateSafe()`, and `OccurredAt` set to the current POSIX timestamp.
    - Runtime calls `Session.Fail(runtimeError, terminationNotifier)`. `Session.Fail` sends a notification to `terminationNotifier`; the next `select` iteration (or the same one, depending on scheduling) observes the termination signal and exits the loop.
    - Runtime exits the monitoring loop and proceeds to step 18 (SessionFinalizer).
12. If an OS signal is received (SIGINT or SIGTERM):
    - Runtime logs: `"Received signal <signal-name>. Initiating graceful shutdown."`
    - Runtime calls `RuntimeSocketManager.DeleteSocket()` to stop the listener and close the socket. This unblocks any pending socket operations.
    - Runtime does **not** transition the session to "failed" status. The session remains in its current status ("running" or "initializing").
    - Runtime releases any locks held by Session (Session's internal read-write lock is automatically released when goroutines exit).
    - Runtime proceeds to step 18 (SessionFinalizer).
13. After exiting the monitoring loop, Runtime proceeds to cleanup.
14. Runtime stops the socket listener by calling `RuntimeSocketManager.DeleteSocket()` (idempotent; safe to call even if already called in step 12).
15. Runtime waits for the listener goroutine to exit by reading from `listenerDoneCh`. RuntimeSocketManager closes `listenerDoneCh` when the listener goroutine has fully terminated (after socket close, all in-flight handlers returned). After `listenerDoneCh` is closed, Runtime drains any remaining errors from `listenerErrCh` (best-effort, non-blocking) and discards them; the session is already in a terminal state by this point so additional listener errors are not actionable.
16. Runtime ensures that all locks are released by this point. Session methods handle lock release automatically. No explicit unlock is required by Runtime.
17. Runtime invokes `SessionFinalizer.Finalize(session)` to print the final status and perform best-effort cleanup.
18. If SessionFinalizer was invoked due to an initialization error (step 4):
    - If a session entity exists but is in "initializing" or "failed" status, SessionFinalizer prints the error details.
    - If no session entity exists, Runtime prints to stderr: `"Failed to initialize session: <error>"` and exits with code 1 without calling SessionFinalizer.
19. After SessionFinalizer completes, Runtime exits.
20. Runtime exits with code 0 if `Session.Status == "completed"`, and code 1 if `Session.Status == "failed"` or if initialization failed.

### Goroutine Management

1. The socket listener runs in a separate goroutine spawned by Runtime at step 7.
2. The listener goroutine runs `RuntimeSocketManager.Listen()`, which blocks until the socket is closed or an error occurs.
3. When `RuntimeSocketManager.DeleteSocket()` is called (either in step 12 or step 14), the socket is closed, and the listener goroutine exits.
4. Runtime waits for the listener goroutine to exit using a done channel before invoking SessionFinalizer.
5. If the listener goroutine encounters an error (e.g., socket bind failure), it may trigger a RuntimeError via MessageRouter's panic recovery or return an error to Runtime (depending on the error type). Runtime handles this by calling `Session.Fail()` and proceeding to SessionFinalizer.

### Signal Handling

1. Runtime uses `signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)` to capture SIGINT (Ctrl+C) and SIGTERM (kill) signals.
2. When a signal is received, Runtime initiates graceful shutdown:
   - Stops the socket listener (prevents new connections).
   - Releases locks (automatically handled by Session methods and goroutine exit).
   - Does **not** transition the session to "failed". The session remains in its current status.
   - Calls SessionFinalizer to print the current status.
3. Runtime does **not** handle SIGKILL (kill -9), as SIGKILL cannot be caught by the process.
4. If a second signal is received during shutdown, Runtime exits immediately with code 130 (standard Unix convention for signal interruption) without waiting for graceful cleanup.

## Inputs

### For Run Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowName | string | Non-empty, PascalCase, must reference a valid workflow definition | Yes |

## Outputs

### Success Case

No return value (void). Runtime exits the process with exit code 0 if the session completes successfully.

### Error Cases

Runtime exits the process with the following exit codes:

| Exit Code | Description |
|-----------|-------------|
| 0 | Session completed successfully (Status="completed") |
| 1 | Session failed (Status="failed") or initialization failed |
| 130 | Terminated by second OS signal during shutdown |

### Console Output

Runtime prints the following to console:

**On OS Signal**:
```
Received signal <signal-name>. Initiating graceful shutdown.
```

**On Initialization Failure (no session entity)**:
```
Failed to initialize session: <error>
```

**Session Status (via SessionFinalizer)**:
- See SessionFinalizer specification for status output format.

## Invariants

1. **Single Session Binding**: Each Runtime invocation binds to exactly one session. Runtime does not support managing multiple sessions concurrently.

2. **Termination Notifier Buffer Size**: The terminationNotifier channel must have a buffer size of 2. This is validated by SessionInitializer.

3. **Graceful Shutdown on Signals**: Runtime must handle SIGINT and SIGTERM signals gracefully by stopping the socket listener and calling SessionFinalizer. The session status is not modified (remains "running" or "initializing").

4. **No Session Failure on Signal**: OS signals do **not** transition the session to "failed" status. The session remains in its current status at the time of the signal.

5. **Listener Goroutine Cleanup**: Runtime must wait for the listener goroutine to exit before calling SessionFinalizer. This ensures no lingering goroutines after Runtime exits.

6. **SessionFinalizer Always Called**: SessionFinalizer must be called in all termination paths (success, failure, signal, initialization error with session entity) except when no session entity exists (early initialization failure).

7. **Exit Code Consistency**: Runtime must exit with code 0 for completed sessions, code 1 for failed or initialization-failed sessions, and code 130 for double-signal termination.

8. **Lock Release**: All locks (Session's internal read-write lock, file locks in stores and socket manager) must be released before Runtime exits. This is handled automatically by component methods and goroutine exit.

9. **Idempotent Socket Deletion**: Runtime may call `RuntimeSocketManager.DeleteSocket()` multiple times (once in signal handler, once in cleanup). The method is idempotent and safe to call repeatedly.

10. **Listener Error Handling**: If the listener goroutine encounters an error, it should trigger a RuntimeError via MessageRouter or notify Runtime through an error channel. Runtime transitions the session to "failed" and proceeds to SessionFinalizer.

11. **SessionInitializer Error Handling**: If SessionInitializer fails, Runtime proceeds to SessionFinalizer if a session entity exists. Otherwise, Runtime prints an error and exits with code 1.

12. **Main Loop Blocking**: The main loop blocks on `select` until a termination notification or OS signal is received. No polling or busy-waiting is used.

13. **MessageRouter Initialization**: MessageRouter must be initialized with a valid Session, EventProcessor, ErrorProcessor, and terminationNotifier before being passed to RuntimeSocketManager.Listen().

14. **Channel Lifecycle Convention**: `terminationNotifier` is created by Runtime and is **never closed** by any component. The channel is garbage-collected when the Session entity becomes unreachable after Runtime exits. `listenerErrCh` is owned by RuntimeSocketManager and is **never closed** (consumers must not assume close means "no more errors"; instead, observe `listenerDoneCh` closure as the listener-shutdown signal). `listenerDoneCh` is the only channel that gets closed, and it is closed exactly once by RuntimeSocketManager when its listener goroutine has fully exited. This convention avoids the well-known "send on closed channel" panic class without requiring synchronization between concurrent senders.

## Edge Cases

- **Condition**: SessionInitializer fails to find the project root (SpectraFinder error).
  **Expected**: SessionInitializer returns an error. Runtime prints to stderr: `"Failed to initialize session: failed to find project root: <error>. Run 'spectra init' to initialize the project."` and exits with code 1. No session entity exists, so SessionFinalizer is not called.

- **Condition**: SessionInitializer fails to load the workflow definition (file not found, parse error).
  **Expected**: SessionInitializer returns an error. Runtime prints to stderr: `"Failed to initialize session: failed to load workflow definition: <error>"` and exits with code 1. No session entity exists, so SessionFinalizer is not called.

- **Condition**: SessionInitializer creates a session but fails during socket creation.
  **Expected**: SessionInitializer returns an error. A session entity exists with Status="initializing" or "failed". Runtime calls SessionFinalizer, which prints the session status to stderr. Runtime exits with code 1.

- **Condition**: SessionInitializer times out (exceeds 30 seconds).
  **Expected**: SessionInitializer's timeout handler calls `Session.Fail()`, which sends a notification to terminationNotifier. SessionInitializer may still return an error to Runtime. Runtime receives the termination notification, exits the monitoring loop (or never enters it if SessionInitializer has not returned), and calls SessionFinalizer. SessionFinalizer prints the timeout RuntimeError to stderr. Runtime exits with code 1.

- **Condition**: RuntimeSocketManager.Listen() fails to bind to the socket immediately.
  **Expected**: Runtime captures the error, constructs a RuntimeError with `Issuer="Runtime"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and proceeds to SessionFinalizer. SessionFinalizer prints the error. Runtime exits with code 1.

- **Condition**: Session completes successfully (Status="completed") via an exit transition.
  **Expected**: `Session.Done()` sends a notification to terminationNotifier. Runtime receives the notification, exits the monitoring loop, stops the socket listener, and calls SessionFinalizer. SessionFinalizer prints to stdout: `"Session <SessionID> completed successfully. Workflow: <WorkflowName>"`. Runtime exits with code 0.

- **Condition**: Session fails (Status="failed") due to an AgentError or RuntimeError.
  **Expected**: `Session.Fail()` sends a notification to terminationNotifier. Runtime receives the notification, exits the monitoring loop, stops the socket listener, and calls SessionFinalizer. SessionFinalizer prints the error to stderr. Runtime exits with code 1.

- **Condition**: User presses Ctrl+C (SIGINT) while session is running.
  **Expected**: Runtime's signal handler receives SIGINT, logs `"Received signal interrupt. Initiating graceful shutdown."`, calls `RuntimeSocketManager.DeleteSocket()` to stop the listener, releases locks, calls SessionFinalizer. SessionFinalizer prints the current session status (likely "running"). Runtime exits with code 1 (because session is not "completed").

- **Condition**: User sends SIGTERM to the Runtime process.
  **Expected**: Same as SIGINT. Runtime initiates graceful shutdown, stops the listener, calls SessionFinalizer, and exits with code 1.

- **Condition**: User sends SIGKILL (kill -9) to the Runtime process.
  **Expected**: The process is immediately terminated by the OS. No cleanup is performed. The runtime socket file remains on disk. On next session creation, SessionDirectoryManager may detect the residual socket and return an error (if the same UUID is generated, which is extremely unlikely).

- **Condition**: User presses Ctrl+C twice in rapid succession.
  **Expected**: The first SIGINT triggers graceful shutdown. The second SIGINT is received during shutdown. Runtime exits immediately with code 130 without waiting for SessionFinalizer to complete.

- **Condition**: Session transitions to "completed" and sends a notification to terminationNotifier at the exact moment the user presses Ctrl+C.
  **Expected**: `select` in the monitoring loop receives whichever event arrives first. If the termination notification arrives first, Runtime proceeds with normal completion. If the signal arrives first, Runtime proceeds with graceful shutdown. Both are acceptable outcomes. The session status is deterministic based on which `select` case is chosen.

- **Condition**: Listener goroutine panics due to a bug in MessageRouter.
  **Expected**: MessageRouter implements panic recovery and triggers a RuntimeError, which calls `Session.Fail()` and sends a notification to terminationNotifier. Runtime receives the notification, exits the monitoring loop, and calls SessionFinalizer. SessionFinalizer prints the RuntimeError with panic details. Runtime exits with code 1.

- **Condition**: terminationNotifier channel fills up (buffer size 2) before Runtime starts monitoring.
  **Expected**: This should not occur under normal operation. `Session.Done()` and `Session.Fail()` send at most one notification each. If both are called (programming error), the buffer accommodates both. If a third send is attempted (should never happen), it blocks until Runtime starts reading from the channel.

- **Condition**: SessionFinalizer fails to print to stdout or stderr (e.g., file descriptor closed).
  **Expected**: Print operations may fail silently. SessionFinalizer does not check for errors. Runtime proceeds to exit.

- **Condition**: RuntimeSocketManager.DeleteSocket() logs a warning during cleanup (step 14).
  **Expected**: The warning is logged by RuntimeSocketManager. Runtime ignores the warning and proceeds to SessionFinalizer. SessionFinalizer prints the session status. Runtime exits normally.

- **Condition**: Session metadata persistence fails during `Session.Done()` or `Session.Fail()` (best-effort persistence).
  **Expected**: Session methods log warnings but do not return errors. The in-memory session status is authoritative. SessionFinalizer prints the in-memory status. Runtime exits with the appropriate exit code based on the in-memory status.

- **Condition**: Runtime process crashes (e.g., segfault) before SessionFinalizer is called.
  **Expected**: No cleanup is performed. The runtime socket file and session files remain on disk. The session status on disk may not reflect the in-memory status (if persistence failed). On next invocation, SessionInitializer detects the residual socket file (if the same UUID is generated) and returns an error. Crash recovery logic is not specified in this design.

- **Condition**: Listener goroutine is still processing a message when Runtime calls RuntimeSocketManager.DeleteSocket().
  **Expected**: The socket is closed, which interrupts the connection. MessageRouter's panic recovery (if triggered) or the listener's normal error handling gracefully exits the goroutine. Runtime waits for the goroutine to signal the done channel, then proceeds to SessionFinalizer.

- **Condition**: terminationNotifier channel is garbage-collected before Runtime starts monitoring (programming error).
  **Expected**: This should not occur. terminationNotifier is passed to SessionInitializer and retained by Session methods. As long as the Session entity is reachable, the channel is not garbage-collected.

## Related

- [SessionInitializer](./session_initializer.md) - Initializes the session
- [SessionFinalizer](./session_finalizer.md) - Finalizes the session and prints status
- [Session](../entities/session/session.md) - Session entity and lifecycle methods
- [MessageRouter](./message_router.md) - Routes messages to processors
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Manages socket lifecycle and listener
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture and workflow runtime
