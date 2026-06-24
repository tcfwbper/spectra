# SessionMetadata — Entity

## Overview

SessionMetadata represents the persistable subset of a Session's state, excluding the event history and runtime-only fields. It captures the core identifying information, lifecycle status, timestamps, current workflow position, session-scoped data store, and any fatal error. SessionMetadata is a plain data structure embedded in Session — it has no methods, no locks, and no persistence logic of its own. Persistence is entirely owned by the runtime layer.

## Boundaries

- Owns: structural grouping of persistable session fields for serialization purposes.
- Delegates: all field mutations to Session entity methods (see [Session](./session.md)).
- Delegates: persistence (serialization, storage) to the runtime layer.
- Must not: have any methods of its own.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be instantiated independently — it is only valid when embedded in Session via `NewSession`.

## Dependencies

None. This entity depends only on Go standard library types and sibling entity types within the `entities` package ([AgentError](../agent_error.md), [RuntimeError](../runtime_error.md)).

## Behavior

1. SessionMetadata is a plain data structure with no methods of its own.
2. All mutations to SessionMetadata fields occur through Session entity methods (see [Session](./session.md)).
3. SessionMetadata is embedded in the Session entity, allowing direct field access on Session instances.
4. SessionMetadata instances are created only during Session construction by `NewSession`.
5. SessionMetadata is designed to be JSON-serializable for persistence.
6. The Error field uses the `omitempty` JSON tag, meaning it is omitted from serialization when nil.

## Inputs

SessionMetadata is not invoked; it is constructed as part of Session initialization. See [Session](./session.md) for construction details.

## Outputs

### SessionMetadata Struct

| Field | Type | Constraints | Description |
|---|---|---|---|
| `ID` | string (UUID) | Valid UUID format; immutable after construction | Unique session identifier |
| `WorkflowName` | string | Non-empty; immutable after construction | Name of the workflow being executed |
| `Pid` | int | > 0; immutable after construction | OS process ID of the `spectra run` process that owns this session |
| `Status` | string | Enum: `"initializing"`, `"running"`, `"completed"`, `"failed"` | Current execution status |
| `CreatedAt` | int64 (POSIX seconds) | > 0; immutable after construction | Timestamp at construction |
| `UpdatedAt` | int64 (POSIX seconds) | >= `CreatedAt` | Timestamp of last in-memory mutation |
| `CurrentState` | string | Non-empty; always a workflow-defined node name | Active node in the workflow state machine |
| `SessionData` | `map[string]any` | Never nil; see [data.md](./data.md) for namespace conventions | Session-scoped key-value store |
| `Error` | `error` (dynamic type `*AgentError` or `*RuntimeError`, nullable) | Non-nil iff `Status == "failed"`; omitempty in JSON | First fatal error |

## Invariants

1. **Type Integrity**: SessionMetadata is a value type (struct) with no embedded locks or channels. It is safe to copy for serialization.

2. **Status Enumeration**: `Status` must always be one of the four defined values: `"initializing"`, `"running"`, `"completed"`, or `"failed"`.

3. **Timestamp Ordering**: `0 < CreatedAt <= UpdatedAt` at all times.

4. **Error Correlation**: `Status == "failed"` iff `Error != nil`.

5. **Non-Empty ID**: `ID` must always be a valid UUID v4 format string.

6. **Non-Empty WorkflowName**: `WorkflowName` must always be a non-empty string.

7. **Positive Pid**: `Pid` must always be > 0. It is immutable after construction.

8. **Non-Empty CurrentState**: `CurrentState` must always be a non-empty string representing a workflow-defined node name.

9. **SessionData Never Nil**: `SessionData` must never be nil; it is initialized as an empty map `map[string]any{}` if there are no entries.

10. **JSON Serialization Compatibility**: All field types must be JSON-serializable. The Error field's dynamic type (`*AgentError` or `*RuntimeError`) exposes getter methods sufficient for external serializers; these entities are not required to implement `json.Marshaler` themselves. Serialization logic is owned by the persistence layer (SessionMetadataStore).

11. **Error Field Omitempty**: When Error is nil, it is omitted from JSON serialization (enforced via `omitempty` JSON struct tag).

12. **Pid No Omitempty**: The `Pid` field is always serialized in JSON (no omitempty). It is always > 0 for any valid session.

13. **No Standalone Construction**: SessionMetadata must not be instantiated independently — it is only valid when embedded in Session and initialized by `NewSession`.

## Edge Cases

- **Condition**: SessionMetadata is copied for serialization while Session methods are mutating it.
  Expected: This is prevented by Session's locking model. External code must use `GetMetadataSnapshotSafe` to obtain a consistent copy.

- **Condition**: `SessionData` contains a value that cannot be JSON-marshaled (e.g., a channel or function).
  Expected: The runtime layer's serialization will fail. Session entity does not validate JSON-serializability of SessionData values.

- **Condition**: `Error` field is nil and JSON serialization is requested.
  Expected: The error field is omitted from the JSON output entirely (no `"error": null` entry).

## Related

- [Session](./session.md) — Embeds SessionMetadata and provides mutation methods
- [AgentError](../agent_error.md), [RuntimeError](../runtime_error.md) — Possible types for the Error field
