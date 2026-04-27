# Session — Lifecycle Methods

## Overview

This file specifies the three lifecycle-transition methods on the Session entity: `Run`, `Done`, and `Fail`. These methods are the only legitimate way to mutate `Session.Status` and (in the case of `Fail`) `Session.Error`. All three are thread-safe via the session-level read-write lock and follow the **memory-first, best-effort-persistence** model: in-memory state is updated under the write lock, the lock is released, and persistence is then attempted; persistence failures are logged as warnings and never block the operation.

`Done` and `Fail` additionally send exactly one notification to the `terminationNotifier` channel (created and owned by Runtime), which unblocks Runtime's main `select` loop.

## Behavior

### `Run(terminationNotifier chan<- struct{}) error`

Transitions the session from `"initializing"` to `"running"` and persists to SessionMetadataStore.

1. Acquire the session-level write lock.
2. Validate `Status == "initializing"`. If not, release the lock and return error: `"cannot run session: status is '<actual-status>', expected 'initializing'"`.
3. Update `Status = "running"` in memory.
4. Update `UpdatedAt = now()` in memory (POSIX seconds).
5. Release the write lock.
6. Write the updated session metadata to SessionMetadataStore.
7. If the write fails, log a warning. Do not return an error. The in-memory state remains authoritative.
8. Return `nil`.

`Run` does **not** send a notification to `terminationNotifier`; only `Done` and `Fail` do.

### `Done(terminationNotifier chan<- struct{}) error`

Transitions the session from `"running"` to `"completed"`, persists, and notifies the main loop.

1. Acquire the session-level write lock.
2. Validate `Status == "running"`. If not, release the lock and return error: `"cannot complete session: status is '<actual-status>', expected 'running'"`.
3. Update `Status = "completed"` in memory.
4. Update `UpdatedAt = now()` in memory.
5. Release the write lock.
6. Write the updated session metadata to SessionMetadataStore.
7. If the write fails, log a warning. Do not return an error.
8. Send one notification: `terminationNotifier <- struct{}{}`. The channel is required to be buffered with capacity >= 2 (validated by SessionInitializer), so this send is non-blocking under correct usage.
9. Return `nil`.

### `Fail(err error, terminationNotifier chan<- struct{}) error`

Transitions the session to `"failed"` from any non-terminal status, records the error, persists, and notifies the main loop. The first error wins. Terminal states (`"completed"` or `"failed"`) cannot be changed.

1. If `err == nil`, return error `"error cannot be nil"` immediately (no lock acquired).
2. Validate `err` is `*AgentError` or `*RuntimeError` via type switch:
   ```go
   switch err.(type) {
   case *AgentError, *RuntimeError:
       // ok
   default:
       return errors.New("invalid error type: must be *AgentError or *RuntimeError")
   }
   ```
   Validation happens before lock acquisition.
3. Acquire the session-level write lock.
4. If `Status == "failed"`, release the lock and return error `"session already failed"`. The existing `Error` field is preserved (first error wins).
5. If `Status == "completed"`, release the lock and return error `"cannot fail session: status is 'completed', workflow already terminated"`.
6. Update `Status = "failed"` in memory.
7. Set `Session.Error = err` in memory.
8. Update `UpdatedAt = now()` in memory.
9. Release the write lock.
10. Write the updated session metadata to SessionMetadataStore.
11. If the write fails, log a warning. Do not return an error.
12. Send one notification: `terminationNotifier <- struct{}{}` (non-blocking under correct usage).
13. Return `nil`.

`Fail` may be called from `"initializing"` (timeout from SessionInitializer) or `"running"` (runtime errors). It explicitly rejects both terminal statuses (`"completed"` and `"failed"`) to enforce Terminal State Finality.

## Inputs

### `Run`

| Field | Type | Constraints | Required |
|---|---|---|---|
| terminationNotifier | chan<- struct{} | Buffered, capacity >= 2 | Yes (unused but accepted for symmetry) |

### `Done`

| Field | Type | Constraints | Required |
|---|---|---|---|
| terminationNotifier | chan<- struct{} | Buffered, capacity >= 2 | Yes |

### `Fail`

| Field | Type | Constraints | Required |
|---|---|---|---|
| err | error | Must be non-nil and of dynamic type `*AgentError` or `*RuntimeError` | Yes |
| terminationNotifier | chan<- struct{} | Buffered, capacity >= 2 | Yes |

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

Persistence errors from SessionMetadataStore are **never returned**; they are only logged.

## Invariants

1. **Terminal State Finality**: Once `Status` reaches `"completed"` or `"failed"`, no method in this file may transition it elsewhere. `Run` and `Done` enforce this by status precondition. `Fail` enforces it by explicitly rejecting both `"completed"` and `"failed"` statuses.
2. **First Error Wins**: After `Fail` succeeds once, subsequent `Fail` calls with any other error are rejected; the original `Session.Error` is preserved.
3. **Memory Before Disk**: All three methods update in-memory state under the write lock, release the lock, then attempt persistence. Persistence failure never reverts the in-memory mutation.
4. **Single Notification**: Each successful `Done` or `Fail` call sends exactly one value on `terminationNotifier`. `Run` sends nothing.
5. **UpdatedAt Refresh**: Every successful in-memory mutation in this file refreshes `UpdatedAt` to the current POSIX timestamp. This is performed inside the same critical section as the field being mutated, so `UpdatedAt` is always consistent with the latest in-memory state.
6. **No Lock During Send**: `terminationNotifier` sends occur **after** the write lock is released, so a slow receiver cannot stall other Session readers.
7. **Validation Before Lock**: `Fail`'s argument validation (nil and type-switch) happens before the lock is acquired, so invalid arguments do not contend on the lock.

## Edge Cases

- **`Run` on `"running"`/`"completed"`/`"failed"`**: returns the precondition error; status untouched.
- **`Done` on `"initializing"`/`"completed"`/`"failed"`**: returns the precondition error; status untouched.
- **`Fail` on `"completed"`**: returns error `"cannot fail session: status is 'completed', workflow already terminated"`; status untouched. This enforces Terminal State Finality.
- **`Fail` on `"failed"`**: returns error `"session already failed"`; preserves first error (first-error-wins). This also enforces Terminal State Finality.
- **`Fail` with nil `terminationNotifier`**: panics on send (step 12). This is a programming error; SessionInitializer/Runtime are responsible for always supplying a non-nil channel. `Run` is safe with nil since it never sends.
- **`terminationNotifier` already full** (programming error — would require >= 3 notifications): send blocks indefinitely. Per Invariant 17 of [`session.md`](./session.md), the channel is created with capacity 2 by Runtime and is never sent more than twice (`Done` xor `Fail`), so this should not occur.
- **SessionMetadataStore write fails**: logged warning, no error returned, in-memory state authoritative.
- **Concurrent `Run` and `Fail`** during initialization: serialized by the write lock. Whichever acquires the lock first wins; the other observes the new status and returns its respective precondition error. Correctness is preserved.

## Related

- [`session.md`](./session.md) — entity struct, fields, top-level invariants
- [`getters.md`](./getters.md) — `GetStatusSafe`, `GetCurrentStateSafe`, `GetErrorSafe`
- [`current_state.md`](./current_state.md) — `UpdateCurrentStateSafe`
- [`event_history.md`](./event_history.md) — `UpdateEventHistorySafe`
- [`data.md`](./data.md) — `UpdateSessionDataSafe`, `GetSessionDataSafe`
- [SessionInitializer](../../runtime/session_initializer.md), [Runtime](../../runtime/runtime.md), [SessionFinalizer](../../runtime/session_finalizer.md)
