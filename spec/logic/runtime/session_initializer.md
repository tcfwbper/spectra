# SessionInitializer

## Overview

SessionInitializer orchestrates the complete initialization flow for creating a new session. It receives a workflow name and a termination notifier channel from Runtime, then performs the following sequence: finds the project root, generates a session UUID, loads the workflow definition, creates the session directory structure, constructs a Session entity with Status="initializing", initializes EventStore and MetadataStore (creating empty files), creates the runtime socket, persists the session metadata to disk, and transitions the session to Status="running" by calling Session.Run(). SessionInitializer enforces a mandatory 30-second timeout for the entire initialization process. If initialization exceeds this timeout, SessionInitializer triggers a RuntimeError and transitions the session to "failed" status. SessionInitializer does not clean up partial resources on failure (except the runtime socket, which is cleaned up by RuntimeSocketManager); the session directory and files remain on disk for inspection.

## Behavior

### Initialization Flow

1. SessionInitializer is invoked by Runtime with inputs: `workflowName` (string) and `terminationNotifier` (chan<- struct{}).
2. SessionInitializer validates that `terminationNotifier` has a buffer capacity of at least 2. If the capacity is less than 2, SessionInitializer returns an error: `"terminationNotifier channel must have buffer capacity >= 2, got <actual-capacity>"`.
3. SessionInitializer starts a 30-second timeout timer using `time.AfterFunc(30*time.Second, timeoutHandler)`. The handler is a closure that captures a shared, mutex-guarded reference to the Session entity (initially `nil`) and a shared boolean `initCompleted` (initially `false`).
4. The timeout handler executes the following logic:
   1. Acquire the shared mutex.
   2. If `initCompleted == true`, release the mutex and exit (initialization already finished successfully; nothing to do).
   3. Read the current Session reference.
   4. Release the mutex.
   5. **If the Session reference is `nil`** (timeout fired before SessionInitializer finished constructing the Session entity in step 12): the handler sets a shared atomic flag `timedOutEarly = true` so the main SessionInitializer goroutine can short-circuit on the next checkpoint, then sends a single notification to `terminationNotifier`. The handler does NOT call `Session.Fail()` (no Session exists). Runtime receives the notification, observes that no Session was returned, and reports an early initialization failure (no SessionFinalizer call).
   6. **If the Session reference is non-nil and `Session.Status == "initializing"`**: the handler constructs a RuntimeError with `Issuer="SessionInitializer"`, `Message="session initialization timeout exceeded 30 seconds"`, `Detail={}`, `SessionID` set to the Session's ID, `FailingState` set to `Session.CurrentState` (which is the workflow entry node at this point), and `OccurredAt` set to the current POSIX timestamp. The handler then calls `Session.Fail(runtimeError, terminationNotifier)`.
   7. **If the Session reference is non-nil but `Session.Status != "initializing"`** (e.g., already "running" or "failed"): the handler exits without action.
5. SessionInitializer calls SpectraFinder to locate the project root. SpectraFinder searches upward from the current working directory for a `.spectra/` directory. (Alternative: SpectraFinder may be invoked once by Runtime and the resolved `ProjectRoot` passed to SessionInitializer at construction; both are acceptable as long as the dependency direction is documented.)
6. If SpectraFinder fails to find the project root (returns an error), SessionInitializer returns an error: `"failed to find project root: <error>. Run 'spectra init' to initialize the project."`. The timeout timer is canceled.
7. SessionInitializer generates a new session UUID using a UUID v4 generation library (e.g., `github.com/google/uuid`).
8. SessionInitializer calls `WorkflowDefinitionLoader.Load(workflowName)` to load the workflow definition.
9. If the workflow definition fails to load (file not found, parse error, validation error), SessionInitializer cancels the timeout timer and returns an error: `"failed to load workflow definition: <error>"`.
10. SessionInitializer calls `SessionDirectoryManager.CreateSessionDirectory(sessionUUID)` to create the session directory `.spectra/sessions/<sessionUUID>/` with permissions 0775.
11. If directory creation fails (parent directory does not exist, directory already exists, permission denied), SessionInitializer cancels the timeout timer and returns an error: `"failed to create session directory: <error>"`.
12. SessionInitializer constructs a Session entity in memory with the following fields:
    - `ID`: generated session UUID
    - `WorkflowName`: provided workflow name
    - `Status`: `"initializing"`
    - `CreatedAt`: current POSIX timestamp
    - `UpdatedAt`: current POSIX timestamp
    - `CurrentState`: workflow definition's `EntryNode` value
    - `EventHistory`: empty slice `[]`
    - `SessionData`: empty map `map[string]any{}`
    - `Error`: `nil`
    
    Immediately after construction, SessionInitializer acquires the shared mutex used by the timeout handler and stores the Session reference. From this point onward, a timeout firing will be able to call `Session.Fail()` instead of taking the early-failure path.
13. SessionInitializer initializes SessionMetadataStore with `ProjectRoot` and `SessionUUID`.
14. SessionInitializer initializes EventStore with `ProjectRoot` and `SessionUUID`.
15. SessionInitializer creates empty `session.json` and `events.jsonl` files by triggering the FileAccessor preparation callbacks. This is done by calling a helper method that uses FileAccessor to ensure the files exist. The `session.json` file is created with permissions 0644, and `events.jsonl` is created with permissions 0644.
16. If file creation fails (parent directory does not exist, permission denied), SessionInitializer cancels the timeout timer and returns an error: `"failed to initialize storage files: <error>"`.
17. SessionInitializer initializes RuntimeSocketManager with `ProjectRoot` and `SessionUUID`.
18. SessionInitializer calls `RuntimeSocketManager.CreateSocket()` to create the Unix domain socket file `runtime.sock` in the session directory with permissions 0600.
19. If socket creation fails (socket file already exists, permission denied, disk full), SessionInitializer cancels the timeout timer, calls `RuntimeSocketManager.DeleteSocket()` to clean up any partial socket, and returns an error: `"failed to create runtime socket: <error>"`.
20. SessionInitializer writes the initial session metadata to disk by calling `SessionMetadataStore.Write(sessionMetadata)`. At this point, the session is still in Status="initializing".
21. If the metadata write fails, SessionInitializer cancels the timeout timer, calls `RuntimeSocketManager.DeleteSocket()`, and returns an error: `"failed to persist initial session metadata: <error>"`.
22. SessionInitializer calls `Session.Run(terminationNotifier)` to transition the session from Status="initializing" to Status="running".
23. If `Session.Run()` fails (returns an error), SessionInitializer cancels the timeout timer, calls `RuntimeSocketManager.DeleteSocket()`, constructs a RuntimeError with `Issuer="SessionInitializer"`, `Message="failed to transition session to running status"`, `Detail` containing the error from Session.Run(), calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to transition session to running: <error>"`.
24. If `Session.Run()` succeeds, SessionInitializer acquires the shared mutex, sets `initCompleted = true`, releases the mutex, cancels the timeout timer (`timer.Stop()`), and returns the initialized Session entity to Runtime.

At every checkpoint between major steps (after each numbered step that performs IO), SessionInitializer also reads the shared atomic `timedOutEarly` flag. If it is `true` and the Session has not yet been constructed (step 12), SessionInitializer aborts immediately and returns an error: `"session initialization timeout exceeded 30 seconds before session entity was constructed"`. This avoids continuing to do IO work after the timeout has already fired.

### Timeout Handling

1. The timeout timer is started at the beginning of SessionInitializer execution with a 30-second duration.
2. The timeout handler is a closure that captures: a mutex, a Session reference variable (initially `nil`), an `initCompleted` boolean (initially `false`), an atomic `timedOutEarly` flag (initially `false`), and the `terminationNotifier` channel.
3. The timeout handler's behavior depends on the state at firing time:
   - **Init already completed**: handler exits silently.
   - **Session not yet constructed (early timeout)**: handler sets `timedOutEarly`, sends one notification to `terminationNotifier`, and exits without calling `Session.Fail()`. The main SessionInitializer goroutine observes `timedOutEarly` at the next checkpoint and returns an early-failure error to Runtime.
   - **Session constructed and Status == "initializing"**: handler constructs a RuntimeError and calls `Session.Fail(runtimeError, terminationNotifier)`. `Session.Fail()` transitions Status to `"failed"` in memory, attempts to persist (best-effort), and sends one notification to `terminationNotifier`. The main loop (Runtime) receives the notification and proceeds to call SessionFinalizer.
   - **Session constructed and Status != "initializing"** (e.g., already running): handler exits silently.
4. If initialization completes successfully before the timeout:
   - SessionInitializer marks `initCompleted = true` under the shared mutex.
   - SessionInitializer calls `timer.Stop()` to cancel the timeout timer.
   - The handler, even if it fires concurrently, observes `initCompleted == true` and exits silently.
5. The timeout value of 30 seconds is hardcoded and not configurable in the current design.

## Inputs

### For Initialization

SessionInitializer is constructed once per Runtime invocation with the following injected dependencies (per the dependency-injection conventions documented in the runtime layer):

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` (resolved by Runtime via SpectraFinder before constructing SessionInitializer) | Yes |
| WorkflowDefinitionLoader | WorkflowDefinitionLoader | Injected loader, shared across the runtime; used to load and validate workflow definitions | Yes |
| SessionDirectoryManager | SessionDirectoryManager | Injected directory manager (constructed with ProjectRoot) | Yes |

Per-session collaborators that require the generated SessionUUID (`SessionMetadataStore`, `EventStore`, `RuntimeSocketManager`) are constructed by SessionInitializer internally using `ProjectRoot` and the freshly generated `SessionUUID`. They are not injected because the session UUID is not known until step 7. For test substitution, the test harness can inject SessionInitializer with stub versions of these store constructors via interface (implementation detail; not part of the spec contract).

Stateless utilities (`StorageLayout`, `SpectraFinder`, `FileAccessor`) are not injected; SessionInitializer (or the constructors it invokes) calls their package-level functions directly.

### For Initialize Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowName | string | Non-empty, PascalCase, must reference a valid workflow definition file | Yes |
| TerminationNotifier | chan<- struct{} | Buffered channel with capacity >= 2 | Yes |

## Outputs

### Success Case

| Field | Type | Description |
|-------|------|-------------|
| Session | *Session | Initialized Session entity with Status="running", all resources created |

### Error Cases

| Error Message Format | Description |
|---------------------|-------------|
| `"terminationNotifier channel must have buffer capacity >= 2, got <actual-capacity>"` | Channel buffer is too small |
| `"failed to find project root: <error>. Run 'spectra init' to initialize the project."` | SpectraFinder failed to locate `.spectra/` |
| `"failed to load workflow definition: <error>"` | WorkflowDefinitionLoader failed |
| `"failed to create session directory: <error>"` | SessionDirectoryManager failed |
| `"failed to initialize storage files: <error>"` | EventStore or SessionMetadataStore file creation failed |
| `"failed to create runtime socket: <error>"` | RuntimeSocketManager.CreateSocket() failed |
| `"failed to persist initial session metadata: <error>"` | SessionMetadataStore.Write() failed |
| `"failed to transition session to running: <error>"` | Session.Run() failed |

## Invariants

1. **30-Second Timeout Enforcement**: SessionInitializer must enforce a mandatory 30-second timeout for the entire initialization process. If initialization exceeds this duration, a RuntimeError must be triggered via Session.Fail().

2. **Timeout Timer Lifecycle**: The timeout timer must be started at the beginning of SessionInitializer execution and canceled (using `timer.Stop()`) if initialization completes successfully before the timeout.

3. **Timeout Handler Race Safety**: The timeout handler must check `initCompleted` and the Session reference under a shared mutex before deciding what to do. The three branches are: (a) init already done → exit silently, (b) Session not yet constructed → set `timedOutEarly`, notify the main loop, and exit (do **not** call `Session.Fail` since no Session exists), (c) Session constructed and Status == "initializing" → call `Session.Fail`. SessionInitializer's main goroutine must check `timedOutEarly` between major IO steps and abort early if set.

4. **Termination Notifier Validation**: SessionInitializer must validate that the terminationNotifier channel has buffer capacity >= 2 before proceeding with initialization.

5. **Sequential Execution**: All initialization steps must execute sequentially (not concurrently). This simplifies error handling and resource cleanup.

6. **Session Status Progression**: The session progresses through two status transitions: "initializing" (initial) -> "running" (on Session.Run() success) or "initializing" -> "failed" (on timeout or Session.Run() error).

7. **Partial Cleanup on Failure**: If initialization fails after socket creation, SessionInitializer must call `RuntimeSocketManager.DeleteSocket()` to clean up the socket. Other resources (session directory, empty files) are not cleaned up and remain on disk for inspection.

8. **RuntimeError on Session.Run Failure**: If `Session.Run()` returns an error, SessionInitializer must construct a RuntimeError and call `Session.Fail()` before returning to Runtime. This ensures the session is transitioned to "failed" status with proper error recording.

9. **Metadata Persistence Timing**: Session metadata must be persisted to disk after all resources (directory, files, socket) are created but before calling `Session.Run()`. This ensures the session is recoverable from disk in "initializing" status if the process crashes after persistence but before Session.Run().

10. **Empty EventHistory and SessionData**: The initial Session entity must have an empty EventHistory (empty slice) and empty SessionData (empty map).

11. **EntryNode as CurrentState**: The initial `CurrentState` must be set to the workflow definition's `EntryNode` value.

12. **UUID Uniqueness**: SessionInitializer relies on the UUID generation library to produce unique UUIDs. UUID collisions are considered extremely rare and are detected by SessionDirectoryManager (directory already exists error).

13. **No Concurrent Initialization**: SessionInitializer is designed to initialize one session at a time. Runtime is responsible for ensuring only one SessionInitializer runs per Runtime instance.

14. **Timeout Value Hardcoded**: The 30-second timeout value is hardcoded in SessionInitializer and is not configurable via WorkflowDefinition or global settings.

## Edge Cases

- **Condition**: SpectraFinder fails to find `.spectra/` directory.
  **Expected**: SessionInitializer cancels the timeout timer and returns an error: `"failed to find project root: <error>. Run 'spectra init' to initialize the project."`. No session directory or resources are created.

- **Condition**: WorkflowDefinitionLoader fails (workflow file not found, parse error, validation error).
  **Expected**: SessionInitializer cancels the timeout timer and returns an error: `"failed to load workflow definition: <error>"`. No session directory or resources are created.

- **Condition**: SessionDirectoryManager fails because `.spectra/sessions/` does not exist.
  **Expected**: SessionInitializer cancels the timeout timer and returns an error: `"failed to create session directory: sessions directory does not exist: <path>. Run 'spectra init' to initialize the project."`.

- **Condition**: SessionDirectoryManager fails because the session directory already exists (UUID collision).
  **Expected**: SessionInitializer cancels the timeout timer and returns an error: `"failed to create session directory: session directory already exists: <path>. This indicates a UUID collision or a previous session was not cleaned up properly."`.

- **Condition**: RuntimeSocketManager.CreateSocket() fails because the socket file already exists.
  **Expected**: SessionInitializer cancels the timeout timer, calls `RuntimeSocketManager.DeleteSocket()`, and returns an error: `"failed to create runtime socket: runtime socket file already exists: <path>. This may indicate a previous runtime process did not clean up properly or another runtime is currently active."`. The session directory and empty files remain on disk.

- **Condition**: SessionMetadataStore.Write() fails due to disk full or permission denied.
  **Expected**: SessionInitializer cancels the timeout timer, calls `RuntimeSocketManager.DeleteSocket()`, and returns an error: `"failed to persist initial session metadata: <error>"`. The session directory and empty files remain on disk but contain no metadata.

- **Condition**: Session.Run() fails because Status is not "initializing" (programming error).
  **Expected**: SessionInitializer cancels the timeout timer, calls `RuntimeSocketManager.DeleteSocket()`, constructs a RuntimeError, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to transition session to running: cannot run session: status is '<actual-status>', expected 'initializing'"`. The session is transitioned to "failed" status.

- **Condition**: Initialization exceeds 30 seconds due to slow disk I/O or workflow definition parsing.
  **Expected**: The timeout handler fires, checks `Session.Status == "initializing"`, constructs a RuntimeError with `Issuer="SessionInitializer"` and `Message="session initialization timeout exceeded 30 seconds"`, calls `Session.Fail(runtimeError, terminationNotifier)` to transition Status to "failed", and sends a termination notification to Runtime. SessionInitializer may still be executing initialization steps in the background but the session is already marked as failed. The main loop proceeds to SessionFinalizer.

- **Condition**: Timeout handler fires exactly when Session.Run() is executing.
  **Expected**: Race condition between timeout handler and Session.Run(). If timeout handler acquires the session lock first and transitions to "failed", Session.Run() will fail with "cannot run session: status is 'failed', expected 'initializing'". If Session.Run() acquires the lock first and transitions to "running", the timeout handler checks Status="running" and exits without action. Both outcomes are acceptable.

- **Condition**: Timeout handler fires after Session.Run() has successfully transitioned to "running" but before timer.Stop() is called.
  **Expected**: The timeout handler checks `Session.Status == "running"` and exits without calling Session.Fail(). The session remains in "running" status. SessionInitializer calls timer.Stop(), which returns false (timer already fired) but has no adverse effect.

- **Condition**: TerminationNotifier channel has buffer capacity of 1.
  **Expected**: SessionInitializer returns an error: `"terminationNotifier channel must have buffer capacity >= 2, got 1"`. No initialization is attempted.

- **Condition**: TerminationNotifier channel is nil.
  **Expected**: SessionInitializer returns an error: `"terminationNotifier channel must have buffer capacity >= 2, got 0"` (capacity of nil channel is 0). No initialization is attempted.

- **Condition**: WorkflowDefinition EntryNode references a non-existent node (validation bug).
  **Expected**: WorkflowDefinitionLoader should have caught this during validation and returned an error. If it somehow passes, SessionInitializer sets CurrentState to the invalid node name. Subsequent runtime operations will fail when trying to look up the node.

- **Condition**: SessionMetadataStore or EventStore file creation succeeds but the files are empty (as expected for initialization).
  **Expected**: This is normal behavior. The files are created empty and will be populated as the session progresses.

- **Condition**: Runtime socket creation succeeds but the socket file is not immediately visible to other processes (filesystem buffering delay).
  **Expected**: This is acceptable. spectra-agent clients may need to retry connection attempts with backoff until the socket is available.

- **Condition**: SessionInitializer completes successfully but the process crashes before Runtime can start the socket listener.
  **Expected**: The session remains on disk with Status="running" but no socket listener is active. On restart, the runtime must implement crash recovery logic (not specified here) to detect orphaned sessions.

- **Condition**: Multiple SessionInitializer instances run concurrently in different processes with the same workflow name.
  **Expected**: Each generates a unique session UUID. Session directories are created independently without conflict.

- **Condition**: UUID generation fails (extremely rare, library-level error).
  **Expected**: The UUID generation library panics or returns an error. SessionInitializer propagates the error: `"failed to generate session UUID: <error>"`. (Note: This error case is not listed in Outputs because it's considered a programming error; typical UUID libraries guarantee success.)

- **Condition**: FileAccessor preparation callback fails to create `session.json` or `events.jsonl` despite parent directory existing.
  **Expected**: SessionInitializer cancels the timeout timer and returns an error: `"failed to initialize storage files: <error>"`. The session directory exists but is missing metadata or event files.

## Related

- [Session](../entities/session/session.md) - Session entity structure and lifecycle methods
- [RuntimeError](../entities/runtime_error.md) - Error type constructed on timeout or Session.Run failure
- [WorkflowDefinitionLoader](../storage/workflow_definition_loader.md) - Loads workflow definitions
- [SessionDirectoryManager](../storage/session_directory_manager.md) - Creates session directories
- [SessionMetadataStore](../storage/session_metadata_store.md) - Persists session metadata
- [EventStore](../storage/event_store.md) - Persists event history
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Manages runtime socket lifecycle
- [SpectraFinder](../storage/spectra_finder.md) - Locates project root
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture and session lifecycle
