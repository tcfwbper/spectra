# Session — Current State Method

## Overview

This file specifies `UpdateCurrentStateSafe`, the thread-safe mutator for `Session.CurrentState`. `CurrentState` holds the name of the workflow node that is currently active; it is the authoritative pointer used by `EventProcessor` (to stamp `Event.EmittedBy`) and by `TransitionEvaluator` (to look up applicable transitions from). It is mutated only by `TransitionToNode` after evaluating and dispatching a transition.

The method does not validate that the new state name exists in the workflow; that responsibility belongs to the caller (which has just selected the name from a validated workflow definition). This keeps `Session` independent of the workflow definition. The method follows a **pure best-effort** pattern: it never returns a non-nil error. Empty input and persistence failures are both logged as warnings and otherwise ignored.

## Behavior

### `UpdateCurrentStateSafe(newState string) error`

1. If `newState == ""`, log a warning (`"UpdateCurrentStateSafe called with empty newState; in-memory state unchanged"`) and return `nil` without acquiring the lock or mutating any field. The caller is presumed to have already validated against the workflow definition; an empty value indicates a programming bug, not a runtime condition to be reported up the stack.
2. Acquire the session-level write lock.
3. Set `CurrentState = newState` in memory.
4. Update `UpdatedAt = now()` in memory.
5. Release the write lock.
6. Write the updated session metadata to SessionMetadataStore.
7. If the write fails, log a warning. Do not return an error.
8. Return `nil`.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| newState | string | Caller guarantees a valid, non-empty workflow node name; empty values are tolerated as a no-op-with-warning | Yes |

## Outputs

Returns `error`. **Always returns `nil`.** The signature uses `error` for symmetry with other Session mutators; no condition produces a non-nil value.

Warnings logged (not returned):

| Warning Message | Condition |
|---|---|
| `UpdateCurrentStateSafe called with empty newState; in-memory state unchanged` | `newState == ""` |
| `UpdateCurrentStateSafe persistence failed: <error>` | SessionMetadataStore write failed |

## Invariants

1. **Always Returns nil**: `UpdateCurrentStateSafe` never returns a non-nil error. Callers may safely ignore the return value.
2. **Write Serialization**: Concurrent `UpdateCurrentStateSafe` calls with non-empty `newState` are serialized by the write lock. Last write wins.
3. **No Workflow Validation**: This method does not consult the workflow definition. It accepts any non-empty string. The caller (`TransitionToNode`) supplies a name already validated against the workflow.
4. **Empty newState Is a No-Op**: Empty `newState` does not mutate in-memory state and does not touch persistence; the warning log is the only side effect. This avoids replacing a valid `CurrentState` with garbage on caller error.
5. **Memory Authoritative**: For non-empty `newState`, in-memory `CurrentState` is the single source of truth. Persistence failures never revert the in-memory mutation.
6. **UpdatedAt Refresh**: A successful in-memory write refreshes `UpdatedAt` inside the same critical section.

## Edge Cases

- **Empty `newState`**: warning logged; in-memory state unchanged; returns `nil`. Callers must not rely on this behavior to clear `CurrentState` (there is no API for that).
- **`newState` equal to current `CurrentState`** (self-transition): accepted; write occurs (idempotent at the in-memory level), `UpdatedAt` advances. Self-loops in workflow definitions are rejected at workflow load time, so under correct usage this should not occur via TransitionToNode; nevertheless the method does not block it.
- **`newState` is an unknown node name**: accepted at this layer; the caller is responsible for validation. Subsequent `EventProcessor` reads of `CurrentState` will then look for transitions that do not exist, leading to a `no_transition_found` error returned to the agent (per [Transition](../../components/transition.md)).
- **SessionMetadataStore write fails**: logged warning; in-memory `CurrentState` updated; no error returned.
- **Concurrent updates**: serialized; the order in which calls observe the lock determines the final value.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md), [`data.md`](./data.md), [`event_history.md`](./event_history.md), [`getters.md`](./getters.md)
- [TransitionToNode](../../runtime/transition_to_node.md) — sole caller
