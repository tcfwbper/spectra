# Session — Event History Method

## Overview

This file specifies `UpdateEventHistorySafe`, the only method that mutates `Session.EventHistory`. Events are first validated for required fields, then appended in memory under the session-level write lock. Validation is strict because event-history integrity is critical for audit and downstream transition evaluation: an invalid event in memory could corrupt the workflow state machine.

Session does not perform persistence. The runtime caller is responsible for persisting events after this method returns successfully.

There is intentionally **no** `GetEventHistorySafe` getter at this layer; the runtime accesses `EventHistory` through its own coordination patterns.

## Boundaries

- Owns: required-field validation of events before append, thread-safe append to EventHistory, UpdatedAt refresh.
- Delegates: event construction (including stamping `EmittedBy` and `SessionID`) to the runtime caller.
- Delegates: persistence of the appended event to the runtime caller.
- Must not: perform any I/O or persistence.
- Must not: reference or import any module outside the `entities` package.
- Must not: construct or modify Event entities.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `Event` | Sibling entity | Read fields via getters for validation | Must not construct or mutate Events |

## Behavior

### `UpdateEventHistorySafe(event Event) error`

1. **Required-field validation** (before any lock is acquired):
   - `event.ID()` non-empty string.
   - `event.Type()` non-empty string.
   - `event.SessionID()` non-empty string.
   - `event.EmittedAt()` > 0 (positive POSIX timestamp).
   - `event.EmittedBy()` non-empty string (set by the runtime caller to the session's `CurrentState` at emission time).
   
   `event.Message()` and `event.Payload()` are **not** validated here (their semantics are application-defined; an empty Message or empty-object Payload is legal).
   
   If any required field is missing or invalid, return error `"invalid event: <field> is required"` (where `<field>` is the first failing field name in the order listed above).
2. Acquire the session-level write lock.
3. Append `event` to `EventHistory` in memory.
4. Update `UpdatedAt = now()` in memory.
5. Release the write lock.
6. Return `nil`.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| event | Event | All required-field validations above must pass | Yes |

See [Event](../event.md) for the full Event struct.

## Outputs

Returns `error`. Possible errors:

| Error Message | Condition |
|---|---|
| `invalid event: ID is required` | `event.ID() == ""` |
| `invalid event: Type is required` | `event.Type() == ""` |
| `invalid event: SessionID is required` | `event.SessionID() == ""` |
| `invalid event: EmittedAt is required` | `event.EmittedAt() <= 0` |
| `invalid event: EmittedBy is required` | `event.EmittedBy() == ""` |

## Invariants

1. **Write Serialization**: Concurrent `UpdateEventHistorySafe` calls are serialized by the write lock; events are appended in lock-acquisition order.
2. **Append-Only**: Once appended, events are never removed or reordered in memory.
3. **Validation Before Lock**: Required-field validation runs before the write lock is acquired, so invalid events do not contend on the lock and never enter memory.
4. **Strict Required Fields**: `ID`, `Type`, `SessionID`, `EmittedAt`, `EmittedBy` are required. `Message` and `Payload` are not validated.
5. **UpdatedAt Refresh**: A successful in-memory append refreshes `UpdatedAt` inside the same critical section.
6. **Caller Sets EmittedBy and SessionID**: The runtime caller is responsible for stamping `EmittedBy` (= `Session.CurrentState` at emission time) and `SessionID` (= `Session.ID`) before calling this method. The session does not derive them.
7. **No Persistence**: This method never performs I/O. The runtime caller is responsible for persisting the event.

## Edge Cases

- **Condition**: `Message` is empty string.
  Expected: Accepted (Message is application-level).

- **Condition**: `Payload` is an empty JSON object `{}`.
  Expected: Accepted.

- **Condition**: Validation failure on a required field.
  Expected: Returns error immediately; event is not appended; lock is not acquired.

- **Condition**: Concurrent appends.
  Expected: Serialized; chronological order in `EventHistory` matches the order in which `UpdateEventHistorySafe` acquired the lock.

- **Condition**: Event with `EmittedAt` in the future or before `Session.CreatedAt`.
  Expected: Not validated here; accepted. Time consistency is the caller's responsibility.

- **Condition**: Same `event.ID()` appended twice (programming error).
  Expected: Both copies are stored; this method does not deduplicate.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md), [`data.md`](./data.md), [`current_state.md`](./current_state.md), [`getters.md`](./getters.md)
- [Event](../event.md)
