# Session — Safe Getters

## Overview

This file specifies the thread-safe getters on `Session`: `GetStatusSafe`, `GetCurrentStateSafe`, `GetErrorSafe`, and `GetMetadataSnapshotSafe`. They provide race-free read paths for fields that are mutated under the write lock by other Session methods; direct field access from outside the `session` package is disallowed.

All getters acquire the session-level **read** lock, copy the value(s), release the lock, and return the copy. Readers may run concurrently; they block only while a writer holds the write lock.

## Boundaries

- Owns: thread-safe read access to Session fields via read lock, consistent snapshot creation (GetMetadataSnapshotSafe).
- Must not: mutate any Session field (including `UpdatedAt`).
- Must not: perform any I/O or persistence.
- Must not: reference or import any module outside the `entities` package.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `SessionMetadata` | Embedded struct | Copy all fields for snapshot | Must not mutate |

## Behavior

### `GetStatusSafe() string`

1. Acquire the session-level read lock.
2. Copy `Status` to a local string.
3. Release the read lock.
4. Return the copy.

### `GetCurrentStateSafe() string`

1. Acquire the session-level read lock.
2. Copy `CurrentState` to a local string.
3. Release the read lock.
4. Return the copy.

### `GetErrorSafe() error`

1. Acquire the session-level read lock.
2. Copy the `Error` value (which is one of `nil`, `*AgentError`, or `*RuntimeError`) to a local variable.
3. Release the read lock.
4. Return the copy.

The returned `error` value (if non-nil) shares its underlying `*AgentError` / `*RuntimeError` allocation with the Session. Callers must treat the returned error as immutable. AgentError and RuntimeError have no exported mutators, so this is enforced by their public API.

### `GetMetadataSnapshotSafe() SessionMetadata`

1. Acquire the session-level read lock.
2. Create a shallow copy of the embedded `SessionMetadata` struct.
3. Create a shallow copy of the `SessionData` map (new map with same key-value pairs).
4. Assign the copied map to the snapshot's `SessionData` field.
5. Release the read lock.
6. Return the snapshot.

The returned `SessionMetadata` is a detached copy safe for serialization without holding the lock. The `SessionData` map is shallow-copied so that mutations to the returned map do not affect the Session's internal state. However, values inside the map (if they are reference types) still share underlying allocations — the runtime caller must not mutate map values after obtaining the snapshot.

## Inputs

None of the getters take parameters.

## Outputs

| Method | Returns | Notes |
|---|---|---|
| `GetStatusSafe` | `string` | One of `"initializing"`, `"running"`, `"completed"`, `"failed"`. |
| `GetCurrentStateSafe` | `string` | A workflow-defined node name; never empty for a successfully constructed session. |
| `GetErrorSafe` | `error` | `nil` if `Status != "failed"`; otherwise `*AgentError` or `*RuntimeError`. |
| `GetMetadataSnapshotSafe` | `SessionMetadata` | A detached copy of all persistable fields. Safe for serialization without holding the lock. |

## Invariants

1. **Concurrent Reads**: All getters use the read lock. Multiple goroutines may hold the read lock simultaneously.
2. **Snapshot Consistency**: Each getter returns the value as observed at the moment its read lock was held. Two separate `GetXxxSafe` calls do not provide a combined snapshot; callers that need multiple fields atomically must use `GetMetadataSnapshotSafe`.
3. **No Mutation**: Getters never modify any Session field, including `UpdatedAt`.
4. **Error Aliasing**: `GetErrorSafe` returns the same pointer stored in `Session.Error`. AgentError and RuntimeError are immutable by convention; callers must not mutate them.
5. **Map Isolation**: `GetMetadataSnapshotSafe` returns a shallow copy of the `SessionData` map. Inserting or deleting keys in the returned map does not affect the Session. However, mutable values (if any) are shared.
6. **No Persistence**: Getters never perform I/O.

## Edge Cases

- **Condition**: Called before `Run()`.
  Expected: `GetStatusSafe` returns `"initializing"`; `GetCurrentStateSafe` returns the entry node (set at construction); `GetErrorSafe` returns `nil`.

- **Condition**: Called after `Fail()`.
  Expected: `GetErrorSafe` returns the recorded `*AgentError` or `*RuntimeError`; `GetStatusSafe` returns `"failed"`.

- **Condition**: Called concurrently with a write.
  Expected: Read lock blocks until the write lock is released; the getter observes the post-write state.

- **Condition**: `GetMetadataSnapshotSafe` called while `SessionData` is empty.
  Expected: Returns a snapshot with an empty (non-nil) `SessionData` map.

- **Condition**: Caller mutates the map returned by `GetMetadataSnapshotSafe`.
  Expected: Session's internal `SessionData` is unaffected (map was shallow-copied).

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md) — writers of `Status` and `Error`
- [`current_state.md`](./current_state.md) — writer of `CurrentState`
- [`data.md`](./data.md) — writer of `SessionData`
- [SessionMetadata](./session_metadata.md) — the struct type returned by `GetMetadataSnapshotSafe`
