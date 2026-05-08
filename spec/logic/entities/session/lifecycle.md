# Session — Lifecycle Methods

## Overview

This file specifies the three lifecycle-transition methods on the Session entity: `Run`, `Done`, and `Fail`. These methods are the only legitimate way to mutate `Session.Status` and (in the case of `Fail`) `Session.Error`. All three are thread-safe via the session-level read-write lock. `Done` and `Fail` additionally send exactly one notification to a caller-provided `terminationNotifier` channel, which unblocks the runtime's main loop.

Session does not perform persistence. The runtime caller is responsible for persisting state after these methods return successfully.

## Boundaries

- Owns: status transition logic, status precondition validation, error recording (Fail), and termination notification (Done, Fail).
- Delegates: persistence of the updated state to the runtime caller.
- Delegates: channel creation and lifecycle to the runtime.
- Must not: perform any I/O or persistence.
- Must not: reference or import any module outside the `entities` package.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `AgentError` | Error type for Fail | Type-switch to validate dynamic type | Must not construct AgentError |
| `RuntimeError` | Error type for Fail | Type-switch to validate dynamic type | Must not construct RuntimeError |

## Behavior

### `Run() error`

Transitions the session from `"initializing"` to `"running"`.

1. Acquire the session-level write lock.
2. Validate `Status == "initializing"`. If not, release the lock and return error: `"cannot run session: status is '<actual-status>', expected 'initializing'"`.
3. Update `Status = "running"` in memory.
4. Update `UpdatedAt = now()` in memory (POSIX seconds).
5. Release the write lock.
6. Return `nil`.

`Run` does **not** accept or send to `terminationNotifier`.

### `Done(terminationNotifier chan<- struct{}) error`

Transitions the session from `"running"` to `"completed"` and notifies the runtime.

1. Acquire the session-level write lock.
2. Validate `Status == "running"`. If not, release the lock and return error: `"cannot complete session: status is '<actual-status>', expected 'running'"`.
3. Update `Status = "completed"` in memory.
4. Update `UpdatedAt = now()` in memory.
5. Release the write lock.
6. Send one notification: `terminationNotifier <- struct{}{}`. The channel is expected to be buffered (capacity >= 2), so this send is non-blocking under correct usage.
7. Return `nil`.

### `Fail(err error, terminationNotifier chan<- struct{}) error`

Transitions the session to `"failed"` from any non-terminal status, records the error, and notifies the runtime. The first error wins.

1. If `err == nil`, return error `"error cannot be nil"` immediately (no lock acquired).
2. Validate `err` is `*AgentError` or `*RuntimeError` via type switch. If not, return error `"invalid error type: must be *AgentError or *RuntimeError"`. Validation happens before lock acquisition.
3. Acquire the session-level write lock.
4. If `Status == "failed"`, release the lock and return error `"session already failed"`. The existing `Error` field is preserved (first error wins).
5. If `Status == "completed"`, release the lock and return error `"cannot fail session: status is 'completed', workflow already terminated"`.
6. Update `Status = "failed"` in memory.
7. Set `Session.Error = err` in memory.
8. Update `UpdatedAt = now()` in memory.
9. Release the write lock.
10. Send one notification: `terminationNotifier <- struct{}{}` (non-blocking under correct usage).
11. Return `nil`.

`Fail` may be called from `"initializing"` (e.g., initialization timeout) or `"running"` (runtime errors, agent errors). It explicitly rejects both terminal statuses to enforce Terminal State Finality.

## Inputs

### `Run`

No parameters.

### `Done`

| Field | Type | Constraints | Required |
|---|---|---|---|
| terminationNotifier | chan<- struct{} | Non-nil, buffered, capacity >= 2 | Yes |

### `Fail`

| Field | Type | Constraints | Required |
|---|---|---|---|
| err | error | Must be non-nil and of dynamic type `*AgentError` or `*RuntimeError` | Yes |
| terminationNotifier | chan<- struct{} | Non-nil, buffered, capacity >= 2 | Yes |

## Outputs

All three methods return `error`. Possible errors:

| Method | Error Message | Condition |
|---|---|---|
| `Run` | `cannot run session: status is '<X>', expected 'initializing'` | Status mismatch |
| `Done` | `cannot complete session: status is '<X>', expected 'running'` | Status mismatch |
| `Fail` | `error cannot be nil` | `err` parameter is nil |
| `Fail` | `invalid error type: must be *AgentError or *RuntimeError` | Wrong dynamic type |
| `Fail` | `session already failed` | `Status == "failed"` already |
| `Fail` | `cannot fail session: status is 'completed', workflow already terminated` | `Status == "completed"` already |

## Invariants

1. **Terminal State Finality**: Once `Status` reaches `"completed"` or `"failed"`, no method in this file may transition it elsewhere. `Run` and `Done` enforce this by status precondition. `Fail` enforces it by explicitly rejecting both `"completed"` and `"failed"` statuses.
2. **First Error Wins**: After `Fail` succeeds once, subsequent `Fail` calls with any other error are rejected; the original `Session.Error` is preserved.
3. **Single Notification**: Each successful `Done` or `Fail` call sends exactly one value on `terminationNotifier`. `Run` sends nothing and does not accept a channel.
4. **UpdatedAt Refresh**: Every successful in-memory mutation refreshes `UpdatedAt` to the current POSIX timestamp inside the same critical section.
5. **No Lock During Send**: `terminationNotifier` sends occur **after** the write lock is released, so a slow receiver cannot stall other Session readers.
6. **Validation Before Lock**: `Fail`'s argument validation (nil and type-switch) happens before the lock is acquired, so invalid arguments do not contend on the lock.
7. **No Persistence**: These methods never perform I/O. The runtime caller is responsible for persisting the updated state.

## Edge Cases

- **Condition**: `Run` called when `Status` is `"running"`, `"completed"`, or `"failed"`.
  Expected: Returns precondition error; status untouched.

- **Condition**: `Done` called when `Status` is `"initializing"`, `"completed"`, or `"failed"`.
  Expected: Returns precondition error; status untouched.

- **Condition**: `Fail` called when `Status` is `"completed"`.
  Expected: Returns error `"cannot fail session: status is 'completed', workflow already terminated"`; status untouched.

- **Condition**: `Fail` called when `Status` is `"failed"`.
  Expected: Returns error `"session already failed"`; preserves first error.

- **Condition**: `Fail` called with nil `terminationNotifier`.
  Expected: Panics on send (step 10). This is a programming error; the runtime is responsible for always supplying a non-nil buffered channel.

- **Condition**: `terminationNotifier` is full (programming error — would require >= 3 notifications).
  Expected: Send blocks indefinitely. Under correct usage the channel capacity >= 2 and at most one Done xor Fail succeeds per session, so this cannot occur.

- **Condition**: Concurrent `Run` and `Fail` during initialization.
  Expected: Serialized by the write lock. If `Fail` acquires the lock first, `Run` observes status `"failed"` and returns its precondition error. If `Run` acquires the lock first, `Fail` subsequently succeeds from `"running"`, records the error, and the final status is `"failed"`.

## Related

- [`session.md`](./session.md) — entity struct, fields, top-level invariants
- [`getters.md`](./getters.md) — `GetStatusSafe`, `GetCurrentStateSafe`, `GetErrorSafe`
- [`current_state.md`](./current_state.md) — `UpdateCurrentStateSafe`
- [`event_history.md`](./event_history.md) — `UpdateEventHistorySafe`
- [`data.md`](./data.md) — `UpdateSessionDataSafe`, `GetSessionDataSafe`
- [AgentError](../agent_error.md), [RuntimeError](../runtime_error.md) — accepted error types for `Fail`
