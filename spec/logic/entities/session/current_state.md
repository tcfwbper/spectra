# Session — Current State Method

## Overview

This file specifies `UpdateCurrentStateSafe`, the thread-safe mutator for `Session.CurrentState`. `CurrentState` holds the name of the workflow node that is currently active. It is mutated by the runtime after evaluating and dispatching a transition.

The method does not validate that the new state name exists in the workflow; that responsibility belongs to the caller (which has just selected the name from a validated workflow definition). This keeps Session independent of the workflow definition.

Session does not perform persistence. The runtime caller is responsible for persisting state after this method returns successfully.

## Boundaries

- Owns: thread-safe mutation of `CurrentState`, empty-input rejection, UpdatedAt refresh.
- Delegates: workflow node name validation to the runtime caller.
- Delegates: persistence of the updated state to the runtime caller.
- Must not: perform any I/O or persistence.
- Must not: reference or import any module outside the `entities` package.
- Must not: consult any workflow definition.

## Dependencies

None beyond the session-level lock.

## Behavior

### `UpdateCurrentStateSafe(newState string) error`

1. If `newState == ""`, return error `"current state cannot be empty"` without acquiring the lock or mutating any field.
2. Acquire the session-level write lock.
3. Set `CurrentState = newState` in memory.
4. Update `UpdatedAt = now()` in memory.
5. Release the write lock.
6. Return `nil`.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| newState | string | Non-empty; caller guarantees a valid workflow node name | Yes |

## Outputs

Returns `error`. Possible errors:

| Error Message | Condition |
|---|---|
| `current state cannot be empty` | `newState == ""` |

## Invariants

1. **Non-Empty Enforcement**: `UpdateCurrentStateSafe` rejects empty strings with an error. `CurrentState` is never set to empty after construction.
2. **Write Serialization**: Concurrent `UpdateCurrentStateSafe` calls with non-empty `newState` are serialized by the write lock. Last write wins.
3. **No Workflow Validation**: This method does not consult the workflow definition. It accepts any non-empty string. The caller supplies a name already validated against the workflow.
4. **UpdatedAt Refresh**: A successful in-memory write refreshes `UpdatedAt` inside the same critical section.
5. **Validation Before Lock**: Empty-string check occurs before lock acquisition.
6. **No Persistence**: This method never performs I/O. The runtime caller is responsible for persisting the updated state.

## Edge Cases

- **Condition**: `newState` is empty.
  Expected: Returns error `"current state cannot be empty"`; no lock acquired, no mutation.

- **Condition**: `newState` equals current `CurrentState` (self-transition).
  Expected: Accepted; write occurs (idempotent at the value level), `UpdatedAt` advances.

- **Condition**: `newState` is an unknown node name.
  Expected: Accepted at this layer; the caller is responsible for validation.

- **Condition**: Concurrent updates.
  Expected: Serialized by the write lock; last writer wins.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md), [`data.md`](./data.md), [`event_history.md`](./event_history.md), [`getters.md`](./getters.md)
