# PersistentSession

## Overview

PersistentSession is a wrapper around the Session entity that automatically triggers persistence after every successful in-memory mutation. It exposes the same thread-safe mutation and getter methods as Session, forwarding calls to the underlying Session entity and invoking the appropriate store (SessionMetadataStore or EventStore) upon success. Persistence failures are logged but never propagate as errors to the caller — the in-memory Session remains the single source of truth.

PersistentSession does not own session state logic (status transitions, validation, locking) — those remain in the Session entity. PersistentSession does not own store construction or session directory creation — those are handled by SessionInitializer before PersistentSession is constructed.

## Boundaries

- Owns: automatic persistence orchestration after each successful in-memory mutation.
- Owns: persistence failure logging (non-fatal, log-only).
- Owns: coordination between Session entity methods and store write operations.
- Delegates: all in-memory state management (status transitions, field validation, locking) to the underlying Session entity.
- Delegates: metadata serialization and file I/O to SessionMetadataStore.
- Delegates: event serialization and file I/O to EventStore.
- Delegates: structured logging output to Logger.
- Must not: implement session state logic (validation, status transition rules, locking).
- Must not: construct SessionMetadataStore or EventStore internally (injected at construction).
- Must not: construct the Session entity internally (injected at construction).
- Must not: return persistence errors to callers — only Session entity errors are returned.
- Must not: cache or buffer persistence operations — each mutation triggers immediate persistence.
- Must not: be constructed via struct literal — must use the provided constructor `NewPersistentSession`.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `Session` | In-memory state container | All exported thread-safe methods: `Run()`, `Done(notifier)`, `Fail(err, notifier)`, `UpdateEventHistorySafe(event)`, `UpdateCurrentStateSafe(newState)`, `UpdateSessionDataSafe(key, value)`, `GetStatusSafe()`, `GetCurrentStateSafe()`, `GetErrorSafe()`, `GetMetadataSnapshotSafe()`, `GetSessionDataSafe(key)` | Must not access unexported fields or bypass thread-safe methods |
| `SessionMetadataStore` | Metadata persistence | `Write(metadata)` | Must not call `Read()` during runtime |
| `EventStore` | Event persistence | `Append(event)` | Must not call `Read()` during runtime |
| `Logger` | Structured logging | `Error(msg string, args ...any)` for persistence failures | Must not use for routine operational logging |

Construction constraint: Must be constructed via `NewPersistentSession(session, metadataStore, eventStore, logger)`. Direct struct literal is forbidden. All dependencies are injected and must be non-nil.

## Behavior

### Construction

`NewPersistentSession(session *Session, metadataStore *SessionMetadataStore, eventStore *EventStore, logger logger.Logger) *PersistentSession`

1. Validates all inputs are non-nil. If any input is nil, panics with `"NewPersistentSession: <field> must not be nil"` (programming error — these are all required injected dependencies).
2. Returns a PersistentSession instance holding references to all four dependencies.

### Mutation Methods (auto-persist metadata)

Each of the following methods wraps the corresponding Session method. On success, it triggers a metadata persist. On persistence failure, it logs and returns the Session method's result (nil error).

#### `Run() error`

1. Calls `session.Run()`.
2. If `session.Run()` returns an error, returns that error immediately (no persist attempted).
3. Calls `metadataStore.Write(session.GetMetadataSnapshotSafe())`.
4. If `Write` fails, logs via `logger.Error("failed to persist session metadata after Run", "error", err, "sessionID", session.ID)`.
5. Returns `nil`.

#### `Done(terminationNotifier chan<- struct{}) error`

1. Calls `session.Done(terminationNotifier)`.
2. If `session.Done()` returns an error, returns that error immediately (no persist attempted).
3. Calls `metadataStore.Write(session.GetMetadataSnapshotSafe())`.
4. If `Write` fails, logs via `logger.Error("failed to persist session metadata after Done", "error", err, "sessionID", session.ID)`.
5. Returns `nil`.

#### `Fail(err error, terminationNotifier chan<- struct{}) error`

1. Calls `session.Fail(err, terminationNotifier)`.
2. If `session.Fail()` returns an error, returns that error immediately (no persist attempted).
3. Calls `metadataStore.Write(session.GetMetadataSnapshotSafe())`.
4. If `Write` fails, logs via `logger.Error("failed to persist session metadata after Fail", "error", err, "sessionID", session.ID)`.
5. Returns `nil`.

#### `UpdateCurrentStateSafe(newState string) error`

1. Calls `session.UpdateCurrentStateSafe(newState)`.
2. If `session.UpdateCurrentStateSafe()` returns an error, returns that error immediately (no persist attempted).
3. Calls `metadataStore.Write(session.GetMetadataSnapshotSafe())`.
4. If `Write` fails, logs via `logger.Error("failed to persist session metadata after UpdateCurrentStateSafe", "error", err, "sessionID", session.ID)`.
5. Returns `nil`.

#### `UpdateSessionDataSafe(key string, value any) error`

1. Calls `session.UpdateSessionDataSafe(key, value)`.
2. If `session.UpdateSessionDataSafe()` returns an error, returns that error immediately (no persist attempted).
3. Calls `metadataStore.Write(session.GetMetadataSnapshotSafe())`.
4. If `Write` fails, logs via `logger.Error("failed to persist session metadata after UpdateSessionDataSafe", "error", err, "sessionID", session.ID, "key", key)`.
5. Returns `nil`.

### Mutation Method (auto-persist event)

#### `UpdateEventHistorySafe(event Event) error`

1. Calls `session.UpdateEventHistorySafe(event)`.
2. If `session.UpdateEventHistorySafe()` returns an error, returns that error immediately (no persist attempted).
3. Calls `eventStore.Append(event)` to persist the single event.
4. If `Append` fails, logs via `logger.Error("failed to persist event", "error", err, "sessionID", session.ID, "eventID", event.ID())`.
5. Calls `metadataStore.Write(session.GetMetadataSnapshotSafe())` to persist updated metadata (UpdatedAt changed).
6. If `Write` fails, logs via `logger.Error("failed to persist session metadata after UpdateEventHistorySafe", "error", err, "sessionID", session.ID)`.
7. Returns `nil`.

### Getter Methods (pass-through, no persist)

The following methods are pure pass-throughs to the underlying Session. No persistence is triggered.

- `GetStatusSafe() string` → delegates to `session.GetStatusSafe()`.
- `GetCurrentStateSafe() string` → delegates to `session.GetCurrentStateSafe()`.
- `GetErrorSafe() error` → delegates to `session.GetErrorSafe()`.
- `GetMetadataSnapshotSafe() SessionMetadata` → delegates to `session.GetMetadataSnapshotSafe()`.
- `GetSessionDataSafe(key string) (any, bool)` → delegates to `session.GetSessionDataSafe(key)`.

### Direct Field Access (pass-through, no persist)

- `ID` → delegates to `session.ID` (immutable field, read-only access).
- `WorkflowName` → delegates to `session.WorkflowName` (immutable field, read-only access).

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| session | *Session | Valid, constructed via NewSession | Yes |
| metadataStore | *SessionMetadataStore | Valid, constructed via NewSessionMetadataStore | Yes |
| eventStore | *EventStore | Valid, constructed via NewEventStore | Yes |
| logger | logger.Logger | Non-nil Logger interface implementation | Yes |

### For Mutation Methods

Same as the corresponding Session method inputs (see sibling session spec files).

## Outputs

### For Construction

| Field | Type | Description |
|-------|------|-------------|
| persistentSession | *PersistentSession | Configured instance wrapping Session with automatic persistence |

No error — constructor panics on nil inputs (programming error).

### For Mutation Methods

Same as the corresponding Session method outputs. Persistence failures are never surfaced as return errors.

### For Getter Methods

Same as the corresponding Session method outputs.

## Invariants

1. **Persistence Is Non-Fatal**: Persistence failures must never be returned as errors to callers. Only Session entity validation/precondition errors are returned.

2. **Persist After Every Successful Mutation**: Every successful in-memory mutation triggers an immediate persistence attempt. No mutations are silently skipped.

3. **In-Memory Is Authoritative**: The underlying Session entity's in-memory state is the single source of truth. Persistence is best-effort for debugging and post-session inspection.

4. **No Buffering or Batching**: Each mutation triggers its own independent persistence call. Mutations are not batched or deferred.

5. **Event Persist Uses Append**: `UpdateEventHistorySafe` persists only the single new event via `EventStore.Append(event)`, not the entire history.

6. **Metadata Persist Uses Full Snapshot**: All metadata persists use `SessionMetadataStore.Write(session.GetMetadataSnapshotSafe())`, which is a full overwrite of `session.json`.

7. **Getter Pass-Through**: Getter methods do not trigger persistence and are pure delegates to the Session entity.

8. **Construction Panic on Nil**: Constructor panics (not returns error) if any dependency is nil. This is a programming error, not a runtime condition.

9. **Log Context**: All persistence failure logs include `sessionID`. Event persist failures additionally include `eventID`. SessionData persist failures additionally include the `key`.

10. **No Store Read During Runtime**: PersistentSession must not call `SessionMetadataStore.Read()` or `EventStore.Read()` during runtime operation. Those APIs exist for external tools only.

11. **UpdateEventHistorySafe Persists Both**: After a successful event append, PersistentSession persists both the event (via EventStore.Append) and the metadata (via SessionMetadataStore.Write), because UpdatedAt has changed. Failure of either is logged independently.

12. **Thread Safety Inherited**: PersistentSession does not introduce its own locking. Thread safety is provided by the underlying Session entity's read-write lock and the stores' file-level locks.

## Edge Cases

- **Condition**: `session.Run()` returns a precondition error (e.g., status is not "initializing").
  **Expected**: PersistentSession returns the error immediately. No persist attempted.

- **Condition**: `metadataStore.Write()` fails after `session.Run()` succeeds.
  **Expected**: Error logged. `Run()` returns nil. Session is in "running" status in memory.

- **Condition**: `eventStore.Append()` fails after `session.UpdateEventHistorySafe()` succeeds.
  **Expected**: Error logged. Event is in memory but not on disk. `UpdateEventHistorySafe()` returns nil. Metadata persist still attempted.

- **Condition**: Both `eventStore.Append()` and `metadataStore.Write()` fail in `UpdateEventHistorySafe`.
  **Expected**: Both errors logged independently. Method returns nil.

- **Condition**: `session.Fail()` returns "session already failed".
  **Expected**: PersistentSession returns that error. No persist attempted (the session state did not change).

- **Condition**: Concurrent mutations via PersistentSession from multiple goroutines.
  **Expected**: Session's internal lock serializes in-memory mutations. Store file-level locks serialize persistence writes. Order of persistence may differ from order of in-memory mutations, but last-write-wins semantics of SessionMetadataStore make this safe.

- **Condition**: Constructor called with nil session.
  **Expected**: Panics with `"NewPersistentSession: session must not be nil"`.

- **Condition**: Constructor called with nil metadataStore.
  **Expected**: Panics with `"NewPersistentSession: metadataStore must not be nil"`.

- **Condition**: Constructor called with nil eventStore.
  **Expected**: Panics with `"NewPersistentSession: eventStore must not be nil"`.

- **Condition**: Constructor called with nil logger.
  **Expected**: Panics with `"NewPersistentSession: logger must not be nil"`.

- **Condition**: Persistence fails consistently (disk full).
  **Expected**: Every mutation logs an error. Session continues operating normally in memory. Log accumulates failure records for human inspection.

## Related

- [Session](../entities/session/session.md) — Underlying in-memory state entity
- [Session Lifecycle](../entities/session/lifecycle.md) — Run, Done, Fail methods
- [Session Event History](../entities/session/event_history.md) — UpdateEventHistorySafe method
- [Session Current State](../entities/session/current_state.md) — UpdateCurrentStateSafe method
- [Session Data](../entities/session/data.md) — UpdateSessionDataSafe method
- [SessionMetadataStore](../storage/session_metadata_store.md) — Metadata persistence store
- [EventStore](../storage/event_store.md) — Event persistence store
- [Logger](../logger/logger.md) — Structured logging interface
- [SessionInitializer](./session_initializer.md) — Constructs PersistentSession
