# Session — Entity

## Overview

A `Session` represents a single execution instance of a workflow. It is a pure in-memory state container that composes [SessionMetadata](./session_metadata.md) (the persistable subset: `ID`, `WorkflowName`, `Status`, `CreatedAt`, `UpdatedAt`, `CurrentState`, `SessionData`, `Error`) with runtime-only state: the chronological event log (`EventHistory`) and a read-write lock. Session does not know about persistence, sockets, or any runtime orchestration — those are owned by the runtime layer.

SessionMetadata is embedded (anonymous field) in Session, making all metadata fields directly accessible on the Session instance (e.g., `session.ID`, `session.Status`) without requiring a `Metadata` accessor.

This file defines the **entity struct, its fields, the lock, construction, and the cross-method invariants**. Per-method behavior is split across sibling files matching the planned Go file layout:

| Spec file | Go file | Methods |
|---|---|---|
| [`lifecycle.md`](./lifecycle.md) | `session/lifecycle.go` | `Run`, `Done`, `Fail` |
| [`data.md`](./data.md) | `session/data.go` | `UpdateSessionDataSafe`, `GetSessionDataSafe` |
| [`event_history.md`](./event_history.md) | `session/event_history.go` | `UpdateEventHistorySafe` |
| [`current_state.md`](./current_state.md) | `session/current_state.go` | `UpdateCurrentStateSafe` |
| [`getters.md`](./getters.md) | `session/getters.go` | `GetStatusSafe`, `GetCurrentStateSafe`, `GetErrorSafe`, `GetMetadataSnapshotSafe` |

## Boundaries

- Owns: in-memory state management (status transitions, event history, current state, session data), field validation at construction, and thread-safe access serialization via read-write lock.
- Owns: immutability of construction-time fields (`ID`, `WorkflowName`, `Pid`, `CreatedAt`).
- Owns: terminal-state notification via `terminationNotifier` channel (Done, Fail).
- Delegates: persistence of session state to the runtime caller. Session does not know about or invoke any persistence mechanism.
- Delegates: UUID generation, workflow definition lookup, and `terminationNotifier` channel creation to the runtime caller (provided as constructor arguments).
- Delegates: semantic validation of workflow node names, event types, and agent roles to the runtime caller.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be constructed via struct literal — must use the provided constructor `NewSession`.
- Must not: read from or write to any persistent store.

## Dependencies

None. This entity depends only on Go standard library types and sibling entity types within the `entities` package ([Event](../event.md), [AgentError](../agent_error.md), [RuntimeError](../runtime_error.md)).

Construction constraint: Must be constructed via `NewSession(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

### Construction

`NewSession(id string, workflowName string, entryNode string, pid int, createdAt int64) (*Session, error)`

Validates all inputs and returns an initialized Session. Initial values:

**SessionMetadata fields** (embedded):

| Field | Initial Value |
|---|---|
| `ID` | Validated `id` parameter (UUID format) |
| `WorkflowName` | Validated `workflowName` parameter |
| `Pid` | Validated `pid` parameter (> 0) |
| `Status` | `"initializing"` |
| `CreatedAt` | Validated `createdAt` parameter (> 0) |
| `UpdatedAt` | Same value as `CreatedAt` |
| `CurrentState` | Validated `entryNode` parameter |
| `SessionData` | `map[string]any{}` |
| `Error` | `nil` |

**Runtime-only fields**:

| Field | Initial Value |
|---|---|
| `EventHistory` | `[]Event{}` |
| `mu` | `sync.RWMutex{}` (zero value) |

Validation rules:
1. `id` must be a valid UUID format string. If not, return error `"invalid session ID: must be a valid UUID"`.
2. `workflowName` must be non-empty. If empty, return error `"workflow name cannot be empty"`.
3. `entryNode` must be non-empty. If empty, return error `"entry node cannot be empty"`.
4. `pid` must be > 0. If not, return error `"pid must be a positive integer"`.
5. `createdAt` must be > 0. If not, return error `"createdAt must be a positive POSIX timestamp"`.

After construction, all field access must go through the thread-safe methods listed in the table above.

### Lifecycle Summary

1. `Status == "initializing"` from construction until [`Run`](./lifecycle.md) succeeds.
2. `Status == "running"` after `Run`. Workflow events flow through the runtime, which invokes [`UpdateEventHistorySafe`](./event_history.md), [`UpdateCurrentStateSafe`](./current_state.md), and ultimately [`Done`](./lifecycle.md).
3. `Status == "completed"` after `Done`, or `Status == "failed"` after [`Fail`](./lifecycle.md) is called from any non-terminal status.

### Threading Model

Session uses a single `sync.RWMutex` per entity. All mutating methods acquire the write lock; all `GetXxxSafe` methods acquire the read lock. Argument validation runs **before** lock acquisition. Notifications on `terminationNotifier` are sent **after** the write lock is released, so a slow receiver cannot stall other Session readers.

## Inputs

Session construction inputs are documented in the Construction section above. Per-method inputs are documented in each sibling spec file.

## Outputs

### Session Struct

Session embeds [SessionMetadata](./session_metadata.md) as an anonymous field, plus runtime-only fields:

**Embedded from SessionMetadata**:

| Field | Type | Constraints | Description |
|---|---|---|---|
| `ID` | string (UUID) | Valid UUID format; immutable after construction | Unique session identifier |
| `WorkflowName` | string | Non-empty; immutable after construction | Name of the workflow being executed |
| `Pid` | int | > 0; immutable after construction | OS process ID of the `spectra run` process that owns this session |
| `Status` | string | Enum: `"initializing"`, `"running"`, `"completed"`, `"failed"` | Current execution status |
| `CreatedAt` | int64 (POSIX seconds) | > 0; immutable after construction | Timestamp at construction |
| `UpdatedAt` | int64 (POSIX seconds) | >= `CreatedAt` | Timestamp of last in-memory mutation; refreshed by every mutating method |
| `CurrentState` | string | Non-empty; always a workflow-defined node name | Active node in the workflow state machine |
| `SessionData` | `map[string]any` | Never nil; see [`data.md`](./data.md) for namespace conventions | Session-scoped key-value store |
| `Error` | `error` (dynamic type `*AgentError` or `*RuntimeError`, nullable) | Non-nil iff `Status == "failed"` | First fatal error (first-error-wins) |

**Runtime-only fields**:

| Field | Type | Constraints | Description |
|---|---|---|---|
| `EventHistory` | `[]Event` | Append-only, ordered by append time | Chronological log of emitted events; mutated only by [`UpdateEventHistorySafe`](./event_history.md) |
| `mu` | `sync.RWMutex` | Not serialized | Read-write lock protecting all Session state |

### Constructor Error Output

| Error Message | Condition |
|---|---|
| `invalid session ID: must be a valid UUID` | `id` is not valid UUID format |
| `workflow name cannot be empty` | `workflowName == ""` |
| `entry node cannot be empty` | `entryNode == ""` |
| `pid must be a positive integer` | `pid <= 0` |
| `createdAt must be a positive POSIX timestamp` | `createdAt <= 0` |

## Invariants

These are the cross-method invariants. Per-method invariants live in the corresponding spec files.

1. **Status–State Consistency**:
   - `Status == "initializing"` ⇒ `CurrentState` is the workflow entry node (as provided at construction).
   - `Status == "running"` ⇒ `CurrentState` is any workflow-defined node.
   - `Status == "completed"` ⇒ reached only via [`Done`](./lifecycle.md).
   - `Status == "failed"` ⇒ `Error != nil`; reached only via [`Fail`](./lifecycle.md).

2. **Timestamp Ordering**: `0 < CreatedAt <= UpdatedAt` at all times. Every successful in-memory mutation refreshes `UpdatedAt`.

3. **Error Correlation**: `Status == "failed"` iff `Error != nil`.

4. **State Machine Initialization**: A newly constructed session has `Status == "initializing"` and `CurrentState == entryNode`.

5. **Event History Immutability**: Once appended via [`UpdateEventHistorySafe`](./event_history.md), events are never removed, reordered, or modified.

6. **SessionData Namespace Discipline**: Keys with namespace prefixes (`logicSpec.*`, `<NodeName>.ClaudeSessionID`, `<NodeName>.PID`) follow the conventions documented in [`data.md`](./data.md). The Session does not enforce namespace ownership.

7. **ClaudeSessionID Type Discipline**: Keys matching `<NodeName>.ClaudeSessionID` always hold `string` values; this is enforced by [`UpdateSessionDataSafe`](./data.md) at write time.

8. **PID Type Discipline**: Keys matching `<NodeName>.PID` always hold `int` values; this is enforced by [`UpdateSessionDataSafe`](./data.md) at write time.

9. **Terminal State Finality**: Once `Status` reaches `"completed"` or `"failed"`, no lifecycle method may transition it elsewhere. [`Run`](./lifecycle.md) and [`Done`](./lifecycle.md) reject all non-matching statuses via preconditions. [`Fail`](./lifecycle.md) explicitly rejects both `"completed"` and `"failed"`.

10. **Thread-Safety via Read-Write Lock**: All access to `Status`, `CurrentState`, `EventHistory`, `SessionData`, `Error`, and `UpdatedAt` after construction goes through the methods listed in this index. Direct field access from outside the `session` package is disallowed.

11. **First Error Wins**: The first successful [`Fail`](./lifecycle.md) call is authoritative. Subsequent `Fail` calls are rejected with `"session already failed"`.

12. **No Persistence**: Session never performs I/O. All state is purely in-memory. The runtime caller is responsible for persisting state if desired.

13. **Construction Only Via Constructor**: Must be constructed via `NewSession`. Direct struct literal construction is forbidden. The constructor validates all inputs and establishes all invariants.

## Edge Cases

Cross-method edge cases are listed below. Per-method edge cases live in the corresponding spec files.

- **Condition**: Constructor receives an invalid UUID for `id`.
  Expected: Returns validation error. No Session is created.

- **Condition**: Constructor receives empty `workflowName` or `entryNode`.
  Expected: Returns validation error. No Session is created.

- **Condition**: Constructor receives `pid <= 0`.
  Expected: Returns validation error `"pid must be a positive integer"`. No Session is created.

- **Condition**: Constructor receives `createdAt <= 0`.
  Expected: Returns validation error. No Session is created.

- **Condition**: Multiple goroutines call mutating methods concurrently.
  Expected: All calls are serialized by the write lock. No data races occur. Last write wins for same-field mutations.

## Related

- Per-method spec files (table at the top of this document).
- [SessionMetadata](./session_metadata.md) — Embedded persistable state structure
- [Event](../event.md), [AgentError](../agent_error.md), [RuntimeError](../runtime_error.md)
