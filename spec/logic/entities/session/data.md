# Session — Session Data Methods

## Overview

This file specifies the thread-safe access methods for `Session.SessionData`, the session-scoped key-value store: `UpdateSessionDataSafe` (write) and `GetSessionDataSafe` (read). These methods serialize access via the session-level read-write lock.

`SessionData` is consumed by agents (via spectra-agent's `event-emit`) and by runtime internals. Two namespace conventions are recognized:

- `logicSpec.<key>` — application-level outputs from agents.
- `<NodeName>.ClaudeSessionID` — Claude Code per-node session ID (string, used to resume the right Claude Code session for repeated visits to the same node).

Session does not perform persistence. The runtime caller is responsible for persisting state after mutations.

## Boundaries

- Owns: thread-safe read/write access to SessionData map, key validation, ClaudeSessionID type enforcement.
- Delegates: persistence of the updated state to the runtime caller.
- Delegates: namespace ownership enforcement (which agent writes which key) to the runtime caller.
- Must not: perform any I/O or persistence.
- Must not: reference or import any module outside the `entities` package.

## Dependencies

None beyond the session-level lock.

## Behavior

### `UpdateSessionDataSafe(key string, value any) error`

1. Validate `key` is non-empty. If empty, return error `"session data key cannot be empty"`.
2. **ClaudeSessionID type validation**: if `key` matches the suffix `".ClaudeSessionID"` (case-sensitive), validate that `value` is of dynamic type `string`. If not, return error `"ClaudeSessionID value must be a string, got <type>"`. The session does not validate that the prefix `<NodeName>` matches a workflow-defined node name; that is the caller's responsibility.
3. Acquire the session-level write lock.
4. Set `SessionData[key] = value` in memory.
5. Update `UpdatedAt = now()` in memory.
6. Release the write lock.
7. Return `nil`.

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
| `UpdateSessionDataSafe` | `error` | `nil` on success; non-nil only for validation failures (empty key, invalid ClaudeSessionID value type). |
| `GetSessionDataSafe` | `(any, bool)` | `bool` is `true` iff the key exists in the map. |

Validation error messages:

| Error Message | Condition |
|---|---|
| `session data key cannot be empty` | `key == ""` |
| `ClaudeSessionID value must be a string, got <type>` | Key ends in `.ClaudeSessionID` and `value` is not `string` |

## Invariants

1. **Write Serialization**: All concurrent `UpdateSessionDataSafe` calls are serialized by the write lock. Last write wins per key.
2. **Concurrent Reads**: `GetSessionDataSafe` uses the read lock; multiple readers may proceed concurrently. Readers block while a writer holds the write lock and observe a consistent snapshot.
3. **ClaudeSessionID Type Discipline**: For any key matching `<NodeName>.ClaudeSessionID`, the stored value is always a `string` (enforced by validation in step 2). Other key namespaces accept any value type.
4. **No Node-Name Validation**: The `<NodeName>` prefix in `<NodeName>.ClaudeSessionID` is **not** checked against the workflow definition. The caller is responsible for using the correct node name.
5. **UpdatedAt Refresh**: A successful in-memory write in `UpdateSessionDataSafe` refreshes `UpdatedAt` inside the same critical section.
6. **Validation Before Lock**: Argument validation (empty key, ClaudeSessionID type) occurs before lock acquisition.
7. **No Persistence**: This method never performs I/O. The runtime caller is responsible for persisting the updated state.

## Edge Cases

- **Condition**: Concurrent writes to same key.
  Expected: Serialized; last write wins.

- **Condition**: Concurrent reads during a write.
  Expected: Readers block until the writer releases.

- **Condition**: Empty key passed to `UpdateSessionDataSafe`.
  Expected: Returns validation error; no lock acquired, no mutation.

- **Condition**: Empty key passed to `GetSessionDataSafe`.
  Expected: Returns `(nil, false)` (Go map semantics; no validation needed).

- **Condition**: Nil value for a non-ClaudeSessionID key.
  Expected: Accepted; the map stores `nil`. `GetSessionDataSafe` returns `(nil, true)` — distinguishable from missing via the boolean.

- **Condition**: `<NodeName>.ClaudeSessionID` set to empty string `""`.
  Expected: Accepted (empty string is still a string). The session does not interpret content.

- **Condition**: Key with suffix `.ClaudeSessionID` and value of type `fmt.Stringer` but not actual `string`.
  Expected: Rejected. Type switch is strict on dynamic type `string`.

- **Condition**: Reading a key that does not exist.
  Expected: Returns `(nil, false)`.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md) — `Run`, `Done`, `Fail`
- [`getters.md`](./getters.md) — `GetMetadataSnapshotSafe`
