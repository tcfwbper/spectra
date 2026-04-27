# Session — Safe Getters

## Overview

This file specifies the thread-safe scalar getters on `Session`: `GetStatusSafe`, `GetCurrentStateSafe`, and `GetErrorSafe`. They provide race-free read paths for fields that are mutated under the write lock by `lifecycle.go` and `current_state.go`; direct field access from outside the `session` package is disallowed.

All three getters acquire the session-level **read** lock, copy the scalar value (or pointer for `Error`), release the lock, and return the copy. Readers may run concurrently; they block only while a writer holds the write lock.

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
2. Copy the `Error` pointer (which is one of `nil`, `*AgentError`, or `*RuntimeError`) to a local variable.
3. Release the read lock.
4. Return the copy.

The returned `error` value (if non-nil) shares its underlying `*AgentError` / `*RuntimeError` allocation with the Session. Callers must treat the returned error as immutable. AgentError and RuntimeError have no exported mutators, so this is enforced by their public API.

## Inputs

None of the three getters take parameters.

## Outputs

| Method | Returns | Notes |
|---|---|---|
| `GetStatusSafe` | `string` | One of `"initializing"`, `"running"`, `"completed"`, `"failed"`. |
| `GetCurrentStateSafe` | `string` | A workflow-defined node name; never empty for a successfully initialized session. |
| `GetErrorSafe` | `error` | `nil` if `Status != "failed"`; otherwise `*AgentError` or `*RuntimeError`. |

## Invariants

1. **Concurrent Reads**: All three getters use the read lock. Multiple goroutines may hold the read lock simultaneously.
2. **Snapshot Consistency**: Each getter returns the value as observed at the moment its read lock was held. Two separate `GetXxxSafe` calls do not provide a combined snapshot; callers that read `Status` and `CurrentState` sequentially may observe values from two different points in time. The current call sites (Runtime's signal-handler and listener-error paths) tolerate this.
3. **No Mutation**: Getters never modify any Session field, including `UpdatedAt`.
4. **Error Aliasing**: `GetErrorSafe` returns the same pointer that is stored in `Session.Error`. AgentError and RuntimeError are immutable by convention; callers must not mutate them via reflection or unsafe.

## Edge Cases

- **Called before `Run()`**: `GetStatusSafe` returns `"initializing"`; `GetCurrentStateSafe` returns the workflow's entry node (set by SessionInitializer in step 12); `GetErrorSafe` returns `nil`.
- **Called after `Fail()`**: `GetErrorSafe` returns the recorded `*AgentError` or `*RuntimeError`; `GetStatusSafe` returns `"failed"`.
- **Called concurrently with a write**: read lock blocks until the write lock is released; the getter observes the post-write state.
- **Called on a session whose `Error` is set but `Status` is not yet `"failed"`** (cannot occur: `Fail` mutates both under one critical section): not reachable.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md) — writers of `Status` and `Error`
- [`current_state.md`](./current_state.md) — writer of `CurrentState`
- [Runtime](../../runtime/runtime.md), [SessionFinalizer](../../runtime/session_finalizer.md) — primary readers
