# Session — Session Data Methods

## Overview

This file specifies the thread-safe access methods for `Session.SessionData`, the session-scoped key-value store: `UpdateSessionDataSafe` (write) and `GetSessionDataSafe` (read). These methods serialize access via the session-level read-write lock.

`SessionData` is consumed by agents (via spectra-agent's `event-emit`) and by runtime internals. Three namespace conventions are recognized:

- `logicSpec.<key>` — application-level outputs from agents.
- `<NodeName>.ClaudeSessionID` — Claude Code per-node session ID (string, used to resume the right Claude Code session for repeated visits to the same node).
- `<NodeName>.PID` — OS process ID of the Claude CLI process spawned for this node (integer, used by ClaudeProcessCleaner to terminate orphaned processes on shutdown).

Session does not perform persistence. The runtime caller is responsible for persisting state after mutations.

## Boundaries

- Owns: thread-safe read/write access to SessionData map, key validation, ClaudeSessionID type enforcement, PID type enforcement.
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
3. **PID type validation**: if `key` matches the suffix `".PID"` (case-sensitive), validate that `value` is of dynamic type `int`. If not, return error `"PID value must be an int, got <type>"`. The session does not validate that the prefix `<NodeName>` matches a workflow-defined node name; that is the caller's responsibility.
4. Acquire the session-level write lock.
5. Set `SessionData[key] = value` in memory.
6. Update `UpdatedAt = now()` in memory.
7. Release the write lock.
8. Return `nil`.

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
| value | any | If key has suffix `.ClaudeSessionID`, must be a `string`. If key has suffix `.PID`, must be an `int`. | Yes (may be nil for non-ClaudeSessionID/non-PID keys) |

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
| `PID value must be an int, got <type>` | Key ends in `.PID` and `value` is not `int` |

## Invariants

1. **Write Serialization**: All concurrent `UpdateSessionDataSafe` calls are serialized by the write lock. Last write wins per key.
2. **Concurrent Reads**: `GetSessionDataSafe` uses the read lock; multiple readers may proceed concurrently. Readers block while a writer holds the write lock and observe a consistent snapshot.
3. **ClaudeSessionID Type Discipline**: For any key matching `<NodeName>.ClaudeSessionID`, the stored value is always a `string` (enforced by validation in step 2).
4. **PID Type Discipline**: For any key matching `<NodeName>.PID`, the stored value is always an `int` (enforced by validation in step 3). Other key namespaces accept any value type.
5. **No Node-Name Validation**: The `<NodeName>` prefix in `<NodeName>.ClaudeSessionID` and `<NodeName>.PID` is **not** checked against the workflow definition. The caller is responsible for using the correct node name.
6. **UpdatedAt Refresh**: A successful in-memory write in `UpdateSessionDataSafe` refreshes `UpdatedAt` inside the same critical section.
7. **Validation Before Lock**: Argument validation (empty key, ClaudeSessionID type, PID type) occurs before lock acquisition.
8. **No Persistence**: This method never performs I/O. The runtime caller is responsible for persisting the updated state.

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

- **Condition**: `<NodeName>.PID` set to a valid positive integer.
  Expected: Accepted. The session does not validate PID range or process existence.

- **Condition**: `<NodeName>.PID` set to 0 or negative integer.
  Expected: Accepted (still an `int`). The session does not validate PID semantics, only type.

- **Condition**: Key with suffix `.PID` and value of type `int64` or `float64` but not `int`.
  Expected: Rejected. Type switch is strict on dynamic type `int`.

- **Condition**: Reading a key that does not exist.
  Expected: Returns `(nil, false)`.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md) — `Run`, `Done`, `Fail`
- [`getters.md`](./getters.md) — `GetMetadataSnapshotSafe`
