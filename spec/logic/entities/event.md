# Event

## Overview

An Event is a typed signal entity that drives state transitions in the workflow state machine. Each event carries a workflow-defined type identifier, an optional message string, an optional JSON payload, and metadata about its origin and timing. Event is a pure data entity — it does not perform state transitions, message delivery, or session manipulation. Those behaviors are owned by the runtime.

## Boundaries

- Owns: construction-time validation of all fields (ID, Type, Message, Payload, EmittedBy, EmittedAt, SessionID).
- Owns: immutability guarantee for all fields after construction.
- Delegates: semantic validation (session existence, event type defined in workflow, EmittedBy matches actual current state) to the runtime caller.
- Delegates: state transition evaluation, event history management, message delivery, and session status checks to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be constructed via struct literal — must use the provided constructor.
- Must not: modify session state or evaluate transitions.
- Must not: implement custom MarshalJSON — serialization of Event for persistence is entirely owned by EventStore, which reads fields via getter methods.

## Dependencies

None. This entity depends only on Go standard library types (`json.RawMessage`).

Construction constraint: Must be constructed via `NewEvent(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewEvent(id string, eventType string, message string, payload json.RawMessage, emittedBy string, emittedAt int64, sessionID string)` that validates all fields and returns an immutable Event value.
2. Validates that `ID` is a valid UUID format string.
3. Validates that `Type` is a non-empty string matching PascalCase format (starts with uppercase letter, contains only alphanumeric characters).
4. Accepts `Message` as any string including empty string (empty string is the default when omitted by caller).
5. Validates that `Payload` is a valid JSON object (starts with `{`). Must not be nil, a JSON primitive, or a JSON array.
6. Validates that `EmittedBy` is a non-empty string.
7. Validates that `EmittedAt` is a positive integer (> 0).
8. Validates that `SessionID` is a valid UUID format string.
9. Returns a validation error if any constraint is violated.
10. All fields are immutable after construction.
11. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ID | string | Valid UUID format | Yes |
| Type | string | Non-empty, PascalCase (starts with uppercase letter, alphanumeric only) | Yes |
| Message | string | Any valid string (empty string allowed) | Yes (but may be empty) |
| Payload | json.RawMessage | Valid JSON object (`{...}`). Must not be nil, a primitive, or an array. | Yes |
| EmittedBy | string | Non-empty string (node name) | Yes |
| EmittedAt | int64 | Positive integer (> 0) | Yes |
| SessionID | string | Valid UUID format | Yes |

## Outputs

### Event Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| ID | string | Valid UUID format | Unique identifier for the event |
| Type | string | Non-empty, PascalCase | Event type identifier (workflow-defined) |
| Message | string | Any string (may be empty) | Optional message for the event recipient |
| Payload | json.RawMessage | Valid JSON object | Event-specific data |
| EmittedBy | string | Non-empty | Node name from which the event was emitted |
| EmittedAt | int64 | Positive integer | POSIX timestamp when the event was emitted |
| SessionID | string | Valid UUID format | Associated session identifier |

### Error Output

| Condition | Error |
|-----------|-------|
| ID is not valid UUID format | Validation error |
| Type is empty | Validation error |
| Type does not match PascalCase format | Validation error |
| Payload is nil | Validation error |
| Payload is a JSON primitive or array | Validation error |
| Payload is invalid JSON | Validation error |
| EmittedBy is empty | Validation error |
| EmittedAt <= 0 | Validation error |
| SessionID is not valid UUID format | Validation error |

## Invariants

1. **ID Format**: `ID` must be a valid UUID format string.
2. **Type PascalCase**: `Type` must be non-empty and match PascalCase format (starts with uppercase letter, alphanumeric only).
3. **Message String**: `Message` must always be a valid string. It may be empty but must not be null (Go strings cannot be null, so this is naturally satisfied).
4. **Payload JSON Object**: `Payload` must be a valid JSON object. It must not be nil, a JSON primitive, or a JSON array.
5. **EmittedBy Non-Empty**: `EmittedBy` must be a non-empty string.
6. **Timestamp Positive**: `EmittedAt` must be a positive integer (> 0).
7. **SessionID Format**: `SessionID` must be a valid UUID format string.
8. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
9. **Construction Only Via Constructor**: Must be constructed via `NewEvent`. Direct struct literal construction is forbidden.

## Edge Cases

- Condition: `Type` is an empty string.
  Expected: Constructor returns a validation error. No Event is created.

- Condition: `Type` starts with a lowercase letter (e.g., `"reviewNeeded"`).
  Expected: Constructor returns a validation error indicating PascalCase is required.

- Condition: `Type` contains non-alphanumeric characters (e.g., `"Review-Needed"`, `"review_needed"`).
  Expected: Constructor returns a validation error indicating PascalCase is required.

- Condition: `Payload` is nil.
  Expected: Constructor returns a validation error. Payload must be a JSON object, not nil.

- Condition: `Payload` is a JSON array (e.g., `[]` or `[1,2,3]`).
  Expected: Constructor returns a validation error indicating payload must be a JSON object.

- Condition: `Payload` is a JSON primitive (e.g., `"string"`, `123`, `true`, `null`).
  Expected: Constructor returns a validation error indicating payload must be a JSON object.

- Condition: `Payload` is `{}` (empty JSON object).
  Expected: Constructor accepts this as valid.

- Condition: `Payload` is invalid JSON bytes.
  Expected: Constructor returns a validation error.

- Condition: `Message` is an empty string.
  Expected: Constructor accepts this as valid. Empty string is the default when the caller omits the message.

- Condition: `EmittedBy` is an empty string.
  Expected: Constructor returns a validation error. No Event is created.

- Condition: `ID` is not a valid UUID format.
  Expected: Constructor returns a validation error. No Event is created.

## Related

- [AgentError](./agent_error.md) — agent errors halt the event-driven workflow progression (runtime responsibility)
- [RuntimeError](./runtime_error.md) — runtime errors halt the event-driven workflow progression (runtime responsibility)
