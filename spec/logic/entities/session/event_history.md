# Session — Event History Method

## Overview

This file specifies `UpdateEventHistorySafe`, the only method that mutates `Session.EventHistory`. Events are first validated for required fields, then appended in memory under the session-level write lock, then persisted to EventStore (best-effort). Validation is stricter here than for `SessionData` because event-history integrity is critical for audit and for downstream transition evaluation: an invalid event in memory could corrupt the workflow state machine. Persistence failures, however, are still tolerated (logged as warnings).

There is intentionally **no** `GetEventHistorySafe` getter at this layer; transition evaluation reads `EventHistory` indirectly via the `Session` reference passed to `TransitionEvaluator` (which is a pure function executed while the EventProcessor holds the appropriate read pattern via the session reference).

## Behavior

### `UpdateEventHistorySafe(event Event) error`

1. **Required-field validation** (before any lock is acquired):
   - `event.ID` non-empty string.
   - `event.Type` non-empty string.
   - `event.SessionID` non-empty string.
   - `event.EmittedAt` > 0 (positive POSIX timestamp).
   - `event.EmittedBy` non-empty string (set by EventProcessor to the session's `CurrentState` at emission time).
   
   `event.Message` and `event.Payload` are **not** validated here (their semantics are application-defined; an empty Message or nil Payload is legal).
   
   If any required field is missing or invalid, return error `"invalid event: <field> is required"` (where `<field>` is the first failing field name).
2. Acquire the session-level write lock.
3. Append `event` to `EventHistory` in memory.
4. Update `UpdatedAt = now()` in memory.
5. Release the write lock.
6. Attempt to write the event to EventStore.
7. If the EventStore write fails, log a warning. Do not return an error. The in-memory `EventHistory` is authoritative.
8. Return `nil`.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| event | Event | All required-field validations above must pass | Yes |

See [Event](../event.md) for the full Event struct.

## Outputs

Returns `error`. Possible errors:

| Error Message | Condition |
|---|---|
| `invalid event: ID is required` | `event.ID == ""` |
| `invalid event: Type is required` | `event.Type == ""` |
| `invalid event: SessionID is required` | `event.SessionID == ""` |
| `invalid event: EmittedAt is required` | `event.EmittedAt <= 0` |
| `invalid event: EmittedBy is required` | `event.EmittedBy == ""` |

EventStore persistence errors are never returned; they are only logged.

## Invariants

1. **Write Serialization**: Concurrent `UpdateEventHistorySafe` calls are serialized by the write lock; events are appended in lock-acquisition order.
2. **Append-Only**: Once appended, events are never removed or reordered in memory.
3. **Memory Authoritative**: In-memory `EventHistory` is the single source of truth for transition evaluation. EventStore persistence failures never revert the in-memory append.
4. **Validation Before Lock**: Required-field validation runs before the write lock is acquired, so invalid events do not contend on the lock and never enter memory.
5. **Strict Required Fields**: `ID`, `Type`, `SessionID`, `EmittedAt`, `EmittedBy` are required. `Message` and `Payload` are not validated.
6. **UpdatedAt Refresh**: A successful in-memory append refreshes `UpdatedAt` inside the same critical section.
7. **Caller Sets EmittedBy and SessionID**: `EventProcessor` is responsible for stamping `EmittedBy` (= `Session.CurrentState` at emission time) and `SessionID` (= `Session.ID`) before calling this method. The session does not derive them.

## Edge Cases

- **Empty `Message`**: accepted (Message is application-level).
- **Nil `Payload`**: accepted.
- **Validation failure**: caller (typically EventProcessor) constructs a `RuntimeError` and calls `Session.Fail()`.
- **EventStore write fails (disk full, permission denied, etc.)**: logged warning; in-memory append succeeds; method returns `nil`.
- **Concurrent appends**: serialized; chronological order in `EventHistory` matches the order in which `UpdateEventHistorySafe` returned successfully.
- **Event with `EmittedAt` in the future or before `Session.CreatedAt`**: not validated here; accepted. Time consistency is the caller's responsibility.
- **Same `event.ID` appended twice** (programming error): both copies are stored; this method does not deduplicate.

## Related

- [`session.md`](./session.md) — entity struct, top-level invariants
- [`lifecycle.md`](./lifecycle.md), [`data.md`](./data.md), [`current_state.md`](./current_state.md), [`getters.md`](./getters.md)
- [Event](../event.md)
- [EventStore](../../storage/event_store.md)
- [EventProcessor](../../runtime/event_processor.md) — sole caller in the runtime layer
