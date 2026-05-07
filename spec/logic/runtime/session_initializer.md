# SessionInitializer

## Overview

SessionInitializer orchestrates the initialization flow for creating a new session. It receives a workflow name and a termination notifier channel, then performs the following sequence: generates a session UUID, logs the UUID for user visibility, loads the workflow definition, creates the session directory, constructs the Session entity, constructs per-session stores (SessionMetadataStore, EventStore), constructs a PersistentSession wrapper, and transitions the session to status="running" via PersistentSession.Run(). All persistence is handled automatically by PersistentSession — SessionInitializer does not perform manual store writes. The entire flow is governed by a context.Context with a 30-second timeout.

SessionInitializer does not create the runtime socket (socket lifecycle is owned by Runtime). SessionInitializer does not clean up partial resources on failure; the session directory and files remain on disk for inspection.

## Boundaries

- Owns: session UUID generation and logging.
- Owns: orchestration of initialization steps in sequence (load workflow → create directory → construct session → construct stores → construct PersistentSession → transition).
- Owns: context.Context timeout enforcement (30 seconds).
- Owns: RuntimeError construction and PersistentSession.Fail invocation when PersistentSession.Run() fails or timeout is exceeded after PersistentSession construction.
- Owns: construction of per-session stores (SessionMetadataStore, EventStore) using the generated session UUID.
- Owns: construction of PersistentSession wrapper (combining Session, stores, and Logger).
- Delegates: workflow definition loading to WorkflowDefinitionLoader.
- Delegates: session directory creation to SessionDirectoryManager.
- Delegates: Session entity construction to NewSession constructor.
- Delegates: all persistence to PersistentSession (automatic, non-fatal).
- Delegates: socket creation to Runtime (after SessionInitializer returns).
- Delegates: partial resource cleanup decisions to Runtime.
- Must not: create the runtime socket.
- Must not: clean up partial resources (directory, files) on failure.
- Must not: construct Session via struct literal — must use NewSession.
- Must not: call SessionMetadataStore.Write() or EventStore.Append() directly — all persistence is delegated to PersistentSession.
- Must not: construct WorkflowDefinitionLoader or SessionDirectoryManager internally (injected at construction).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `WorkflowDefinitionLoader` | Workflow loading | `Load(workflowName)` | Must not construct internally |
| `SessionDirectoryManager` | Directory creation | `CreateSessionDirectory(projectRoot, sessionUUID)` | Must not construct internally |
| `NewSession` | Session construction | Call constructor with validated inputs | Must not use struct literal |
| `NewSessionMetadataStore` | Store construction | `NewSessionMetadataStore(projectRoot, sessionUUID)` | Must not call `Write()` directly |
| `NewEventStore` | Store construction | `NewEventStore(projectRoot, sessionUUID, logger)` | Must not call `Append()` directly |
| `NewPersistentSession` | Wrapper construction | `NewPersistentSession(session, metadataStore, eventStore, logger)` | Must not use struct literal |
| `PersistentSession` | State container with auto-persist | `Run()`, `Fail(err, notifier)` | Must not call `Done()` |
| `RuntimeError` | Error entity | Construct via `NewRuntimeError` for failure cases | — |
| `Logger` | Structured logging | `Info(msg, args...)`, `Error(msg, args...)` | Must not use for session status output (that is SessionFinalizer's job) |
| `context.Context` | Timeout enforcement | `context.WithTimeout`, check `ctx.Err()` at checkpoints | — |

Construction constraint: SessionInitializer is constructed with `projectRoot`, `WorkflowDefinitionLoader`, `SessionDirectoryManager`, and `Logger` injected. Per-session stores (SessionMetadataStore, EventStore) and PersistentSession are constructed internally because they require the generated session UUID.

## Behavior

1. SessionInitializer is invoked by Runtime with inputs: `workflowName` (string) and `terminationNotifier` (chan<- struct{}).
2. Validates that `terminationNotifier` has a buffer capacity of at least 2. If the capacity is less than 2, returns an error immediately: `"terminationNotifier channel must have buffer capacity >= 2, got <actual-capacity>"`.
3. Creates a context.Context with a 30-second timeout: `ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)`. Defers `cancel()`.
4. Generates a new session UUID (UUID v4).
5. Logs the session UUID via `Logger.Info("session created", "sessionID", sessionUUID)`.
6. Checks `ctx.Err()`. If context is cancelled, returns a timeout error (see Timeout Handling).
7. Calls `WorkflowDefinitionLoader.Load(workflowName)` to load the workflow definition.
8. If the workflow definition fails to load, returns an error: `"failed to load workflow definition: <error>"`. No session entity is returned.
9. Checks `ctx.Err()`. If context is cancelled, returns a timeout error.
10. Calls `SessionDirectoryManager.CreateSessionDirectory(projectRoot, sessionUUID)` to create the session directory.
11. If directory creation fails, returns an error: `"failed to create session directory: <error>"`. No session entity is returned.
12. Checks `ctx.Err()`. If context is cancelled, returns a timeout error.
13. Constructs a Session entity via `NewSession(sessionUUID, workflowName, workflowDefinition.EntryNode(), now())`.
14. If Session construction fails, returns an error: `"failed to construct session: <error>"`. No session entity is returned.
15. Constructs `SessionMetadataStore` via `NewSessionMetadataStore(projectRoot, sessionUUID)`.
16. Constructs `EventStore` via `NewEventStore(projectRoot, sessionUUID, logger)`.
17. Constructs `PersistentSession` via `NewPersistentSession(session, metadataStore, eventStore, logger)`.
18. Checks `ctx.Err()`. If context is cancelled, constructs a RuntimeError with `Issuer="SessionInitializer"`, `Message="session initialization timed out"`, calls `PersistentSession.Fail(runtimeError, terminationNotifier)`, and returns the error with the failed PersistentSession in InitResult.
19. Calls `PersistentSession.Run()` to transition the session from status="initializing" to status="running". PersistentSession automatically persists metadata after a successful Run() (non-fatal if persistence fails).
20. If `PersistentSession.Run()` fails, constructs a RuntimeError with `Issuer="SessionInitializer"`, `Message="failed to transition session to running: <error>"`, calls `PersistentSession.Fail(runtimeError, terminationNotifier)`, and returns the error with the failed PersistentSession in InitResult.
21. Checks `ctx.Err()`. If context is cancelled after Run() succeeded, constructs a RuntimeError, calls `PersistentSession.Fail(runtimeError, terminationNotifier)`, and returns error with the failed PersistentSession.
22. Returns a successful InitResult containing the PersistentSession and WorkflowDefinition.

### Timeout Handling

1. A context.Context with 30-second timeout governs the entire initialization flow.
2. At each checkpoint between major I/O steps, SessionInitializer checks `ctx.Err()`.
3. If the context is cancelled before the PersistentSession is constructed (steps 6, 9, 12): returns an error `"session initialization timed out"` with no PersistentSession in InitResult.
4. If the context is cancelled after PersistentSession is constructed (steps 18, 21): constructs a RuntimeError with `Issuer="SessionInitializer"`, `Message="session initialization timed out"`, `SessionID=sessionUUID`, `FailingState=entryNode`, `OccurredAt=now()`, calls `PersistentSession.Fail(runtimeError, terminationNotifier)`, and returns the error with the failed PersistentSession in InitResult. PersistentSession automatically persists the failed state (non-fatal if persistence fails).
5. The 30-second timeout is hardcoded and not configurable.
6. The deferred `cancel()` ensures context resources are released regardless of outcome.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` (resolved by Runtime via SpectraFinder) | Yes |
| WorkflowDefinitionLoader | WorkflowDefinitionLoader | Injected loader for workflow definitions | Yes |
| SessionDirectoryManager | SessionDirectoryManager | Injected directory manager | Yes |
| Logger | logger.Logger | Non-nil Logger interface implementation | Yes |

### For Initialize Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowName | string | Non-empty, must reference a valid workflow definition file | Yes |
| TerminationNotifier | chan<- struct{} | Buffered channel with capacity >= 2 | Yes |

## Outputs

### InitResult Structure

| Field | Type | Description |
|-------|------|-------------|
| PersistentSession | *PersistentSession | Initialized PersistentSession wrapper. Non-nil on success (Status="running"). May be non-nil on failure (Status="failed") if failure occurred after PersistentSession construction. Nil if failure occurred before PersistentSession construction. |
| WorkflowDefinition | *WorkflowDefinition | Loaded workflow definition. Nil if failure occurred before loading. |
| Error | error | Nil on success. Non-nil on failure. |

### Error Cases

| Error Message Format | PersistentSession in InitResult? | Description |
|---------------------|------------------------|-------------|
| `"terminationNotifier channel must have buffer capacity >= 2, got <N>"` | No | Channel validation failed |
| `"session initialization timed out"` | Depends on timing | Timeout before PersistentSession construction → No. After → Yes (Status="failed"). |
| `"failed to load workflow definition: <error>"` | No | WorkflowDefinitionLoader failed |
| `"failed to create session directory: <error>"` | No | SessionDirectoryManager failed |
| `"failed to construct session: <error>"` | No | NewSession validation failed |
| `"failed to transition session to running: <error>"` | Yes (Status="failed") | PersistentSession.Run() failed |

## Invariants

1. **30-Second Timeout**: The entire initialization flow is governed by a context.Context with a 30-second timeout. Timeout is checked at each checkpoint between major I/O steps.

2. **Sequential Execution**: All initialization steps execute sequentially (not concurrently).

3. **Session Status Progression**: The session progresses: "initializing" → "running" (success) or "initializing" → "failed" (timeout or step failure after PersistentSession construction).

4. **No Socket Creation**: SessionInitializer must not create the runtime socket. Socket lifecycle is owned by Runtime.

5. **No Partial Cleanup**: SessionInitializer does not clean up partial resources on failure. Directory and files remain on disk.

6. **RuntimeError on Post-Construction Failure**: If any failure occurs after PersistentSession construction, SessionInitializer must construct a RuntimeError and call PersistentSession.Fail before returning.

7. **Persistence Delegated to PersistentSession**: SessionInitializer does not call SessionMetadataStore.Write() or EventStore.Append() directly. All persistence is automatic via PersistentSession. Persistence failures are non-fatal (logged by PersistentSession).

8. **Session UUID Logged Immediately**: The session UUID is logged via Logger.Info immediately after generation, before any I/O that could fail.

9. **InitResult Completeness**: On success, InitResult contains a non-nil PersistentSession (Status="running") and WorkflowDefinition. On failure, fields are populated up to the point of failure.

10. **Constructor Enforcement**: Session must be constructed via NewSession. Stores must be constructed via their respective constructors. PersistentSession must be constructed via NewPersistentSession.

11. **Timeout Value Hardcoded**: The 30-second timeout is not configurable.

12. **Context Cancellation Safety**: If context cancellation is detected and PersistentSession.Fail is called, but Fail returns an error (e.g., session already failed due to race), SessionInitializer still returns the timeout error. The first-error-wins invariant of Session is preserved.

## Edge Cases

- **Condition**: WorkflowDefinitionLoader fails (file not found, parse error, validation error).
  **Expected**: Returns error `"failed to load workflow definition: <error>"`. No session directory or resources created. No PersistentSession in InitResult.

- **Condition**: SessionDirectoryManager fails because `.spectra/sessions/` does not exist.
  **Expected**: Returns error `"failed to create session directory: <error>"` (wrapping ErrNotInitialized from SessionDirectoryManager).

- **Condition**: SessionDirectoryManager fails because session directory already exists (UUID collision).
  **Expected**: Returns error `"failed to create session directory: <error>"` (wrapping ErrSessionDirExists).

- **Condition**: NewSession constructor fails (invalid UUID format from UUID library — extremely unlikely).
  **Expected**: Returns error `"failed to construct session: <error>"`. No PersistentSession in InitResult.

- **Condition**: PersistentSession.Run() fails because status is not "initializing" (concurrent Fail from timeout).
  **Expected**: Constructs RuntimeError, calls PersistentSession.Fail (which may return "session already failed"). Returns error with failed PersistentSession.

- **Condition**: Timeout fires between PersistentSession construction and PersistentSession.Run().
  **Expected**: ctx.Err() detected at checkpoint. RuntimeError constructed. PersistentSession.Fail called. Returns error with failed PersistentSession (Status="failed"). PersistentSession auto-persists the failed state (non-fatal if persistence fails).

- **Condition**: Timeout fires after PersistentSession.Run() succeeds.
  **Expected**: ctx.Err() detected at checkpoint (step 21). RuntimeError constructed. PersistentSession.Fail called. Returns error with failed PersistentSession.

- **Condition**: PersistentSession auto-persist fails after Run() succeeds (disk full).
  **Expected**: PersistentSession logs the error internally. SessionInitializer is unaware. InitResult returned as success. Session operates normally in memory.

- **Condition**: TerminationNotifier channel has buffer capacity of 1.
  **Expected**: Returns error `"terminationNotifier channel must have buffer capacity >= 2, got 1"`. No initialization attempted.

- **Condition**: TerminationNotifier channel is nil.
  **Expected**: Returns error `"terminationNotifier channel must have buffer capacity >= 2, got 0"` (capacity of nil channel is 0).

- **Condition**: Multiple SessionInitializer instances run concurrently in different processes with the same workflow name.
  **Expected**: Each generates a unique session UUID. Session directories are created independently.

- **Condition**: Context timeout fires exactly when PersistentSession.Fail is being called for another reason.
  **Expected**: First Fail call wins. Subsequent Fail calls return "session already failed". SessionInitializer returns whichever error triggered first.

## Related

- [PersistentSession](./persistent_session.md) — Wrapper that auto-persists after mutations
- [Session](../entities/session/session.md) — Session entity, constructed via NewSession
- [Session Lifecycle](../entities/session/lifecycle.md) — Run and Fail methods
- [RuntimeError](../entities/runtime_error.md) — Error entity constructed on post-construction failures
- [WorkflowDefinitionLoader](../storage/workflow_definition_loader.md) — Loads workflow definitions
- [SessionDirectoryManager](../storage/session_directory_manager.md) — Creates session directories
- [SessionMetadataStore](../storage/session_metadata_store.md) — Persists session metadata
- [EventStore](../storage/event_store.md) — Persists event history
- [Logger](../logger/logger.md) — Structured logging interface
