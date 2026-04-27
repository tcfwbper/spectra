# Session — Session Data Methods

## Overview

This file specifies the thread-safe access methods for `Session.SessionData`, the session-scoped key-value store: `UpdateSessionDataSafe` (write) and `GetSessionDataSafe` (read). These methods serialize access via the session-level read-write lock and use the memory-first, best-effort-persistence model.

`SessionData` is consumed by agents (via spectra-agent's `event-emit`) and by Runtime internals. Two namespace conventions are recognized by Runtime:

- `logicSpec.<key>` — application-level outputs from agents.
- `<NodeName>.ClaudeSessionID` — Claude Code per-node session ID (string, used by `AgentInvoker` to resume the right Claude Code session for repeated visits to the same node).

## Behavior

### `UpdateSessionDataSafe(key string, value any) error`

1. Validate `key` is non-empty. If empty, return error `"session data key cannot be empty"`.
2. **ClaudeSessionID type validation**: if `key` matches the suffix `".ClaudeSessionID"` (case-sensitive), validate that `value` is of dynamic type `string`. If not, return error `"ClaudeSessionID value must be a string, got <type>"`. The session does not validate that the prefix `<NodeName>` matches a workflow-defined node name; that is the caller's responsibility.
3. Acquire the session-level write lock.
4. Set `SessionData[key] = value` in memory.
5. Update `UpdatedAt = now()` in memory.
6. Release the write lock.
7. Write the updated session metadata to SessionMetadataStore.
8. If the write fails, log a warning. Do not return an error. The in-memory map is authoritative.
9. Return `nil`.

### `GetSessionDataSafe(key string) (any, bool)`

1. Acquire the session-level read lock.
2. `value, ok := SessionData[key]`.
3. Release the read lock.
4. Return `(value, ok)`.

`GetSessionDataSafe` does not refresh `UpdatedAt` (it is a pure read).

## Inputs

### `UpdateSessionDataSafe`

| Field | Type | Constraints | Required |
|---|---|---|---|
| key | string | Non-empty | Yes |
| value | any | If key has suffix `.ClaudeSessionID`, must be a `string` | Yes (may be nil for non-ClaudeSessionID keys) |

### `GetSessionDataSafe`

| Field | Type | Constraints | Required |
|---|---|---|---|
| key | string | Any string (including empty); empty key always returns `(nil, false)` | Yes |

## Outputs

| Method | Returns | Notes |
|---|---|---|
| `UpdateSessionDataSafe` | `error` | `nil` on success; non-nil only for validation failures (empty key, invalid ClaudeSessionID value type). Persistence failures are logged, not returned. |
| `GetSessionDataSafe` | `(any, bool)` | `bool` is `true` iff the key exists in the map. |

Validation error messages:

| Error Message | Condition |
|---|---|
| `session data key cannot be empty` | `key == ""` |
| `ClaudeSessionID value must be a string, got <type>` | Key ends in `.ClaudeSessionID` and `value` is not `string` |

## Invariants

1. **Write Serialization**: All concurrent `UpdateSessionDataSafe` calls are serialized by the write lock. Last write wins per key.
2. **Concurrent Reads**: `GetSessionDataSafe` uses the read lock; multiple readers may proceed concurrently. Readers block while a writer holds the write lock and observe a consistent snapshot.
3. **Memory Authoritative**: In-memory `SessionData` is the single source of truth. SessionMetadataStore persistence failures never revert the in-memory mutation.
4. **ClaudeSessionID Type Discipline**: For any key matching `<NodeName>.ClaudeSessionID`, the stored value is always a `string` (enforced by validation in step 2). Other key namespaces accept any value type.
5. **No Node-Name Validation**: The `<NodeName>` prefix in `<NodeName>.ClaudeSessionID` is **not** checked against the workflow definition. The caller (typically `AgentInvoker`) is responsible for using the correct node name.
6. **UpdatedAt Refresh**: A successful in-memory write in `UpdateSessionDataSafe` refreshes `UpdatedAt` inside the same critical section.
7. **Validation Before Lock**: Argument validation (empty key, ClaudeSessionID type) occurs before lock acquisition.

## Edge Cases

- **Concurrent writes to same key**: serialized; last write wins.
- **Concurrent reads during a write**: readers block until the writer releases.
- **Empty key**: write returns validation error; read returns `(nil, false)` (Go map semantics; no validation needed).
- **Nil value for a non-ClaudeSessionID key**: accepted; the map stores `nil`. `GetSessionDataSafe` returns `(nil, true)` — distinguishable from missing via the boolean.
- **`<NodeName>.ClaudeSessionID` set to empty string `""`**: accepted (empty string is still a string). The session does not interpret content.
- **Key with suffix `.ClaudeSessionID` and value of type `string`-compatible interface (e.g., `fmt.Stringer`) but not actual `string`**: rejected. Type switch is strict on dynamic type `string`.
- **Reading a key that does not exist**: returns `(nil, false)`.
- **SessionMetadataStore write fails**: logged warning, in-memory map updated, no error returned.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md) — `Run`, `Done`, `Fail`
- [SessionMetadataStore](../../storage/session_metadata_store.md) — persistence target
- [AgentInvoker](../../runtime/agent_invoker.md) — primary writer/reader of `<NodeName>.ClaudeSessionID`
