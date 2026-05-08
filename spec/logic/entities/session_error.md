# SessionError

## Overview

SessionError is a shared base structure embedded by both AgentError and RuntimeError. It holds the common fields and validation logic for all session-halting error entities. SessionError does not exist as a standalone entity in the system — it is always accessed through AgentError or RuntimeError. It does not perform any runtime orchestration (halting, persisting, notifying); it only defines the shared structural contract and construction-time validation.

## Boundaries

- Owns: shared field storage (Message, Detail, OccurredAt, SessionID, FailingState) and their construction-time format validation.
- Owns: immutability guarantee for all shared fields after construction.
- Delegates: semantic validation (session existence, FailingState validity) to the runtime caller.
- Delegates: persistence, state machine halting, and human notification to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be instantiated directly as a standalone entity — it is only valid when embedded in AgentError or RuntimeError.

## Dependencies

None. This entity depends only on Go standard library types (`json.RawMessage`).

## Behavior

1. Provides a constructor `NewSessionError(message string, detail json.RawMessage, occurredAt int64, sessionID string, failingState string)` that validates all fields and returns an immutable value.
2. Validates that `Message` contains at least one non-whitespace character.
3. Validates that `Detail` is either nil (representing JSON null) or a valid JSON object (starts with `{`). Rejects JSON primitives and arrays.
4. Validates that `OccurredAt` is a positive integer (> 0).
5. Validates that `SessionID` is a valid UUID format string.
6. Validates that `FailingState` is a non-empty string.
7. Returns a validation error if any constraint is violated.
8. All fields are immutable after construction.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Message | string | At least one non-whitespace character | Yes |
| Detail | json.RawMessage | nil (null) or valid JSON object (`{...}`). Must not be a primitive or array. | No (nil allowed) |
| OccurredAt | int64 | Positive integer (> 0) | Yes |
| SessionID | string | Valid UUID format | Yes |
| FailingState | string | Non-empty | Yes |

## Outputs

### SessionError Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Message | string | Non-empty (at least one non-whitespace character) | Human-readable error description |
| Detail | json.RawMessage | nil or valid JSON object | Additional error context |
| OccurredAt | int64 | Positive integer | POSIX timestamp when the error occurred |
| SessionID | string | Valid UUID format | Associated session identifier |
| FailingState | string | Non-empty | State machine node where the error occurred |

### Error Output

| Condition | Error |
|-----------|-------|
| Message is empty or whitespace-only | Validation error |
| Detail is a JSON primitive or array | Validation error |
| Detail is invalid JSON | Validation error |
| OccurredAt <= 0 | Validation error |
| SessionID is not valid UUID format | Validation error |
| FailingState is empty | Validation error |

## Invariants

1. **Message Non-Empty**: `Message` must contain at least one non-whitespace character after construction.
2. **Detail Type Constraint**: `Detail` must be nil or a valid JSON object. It must never be a JSON primitive or array.
3. **Timestamp Positive**: `OccurredAt` must be a positive integer (> 0).
4. **SessionID Format**: `SessionID` must be a valid UUID format string.
5. **FailingState Non-Empty**: `FailingState` must be a non-empty string.
6. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
7. **Construction Only Via Constructor**: Must be constructed via `NewSessionError`. Direct struct literal construction is forbidden.
8. **No Standalone Use**: SessionError must not be instantiated as a standalone entity — it is only valid when embedded in AgentError or RuntimeError.

## Edge Cases

- Condition: `Message` is an empty string or contains only whitespace.
  Expected: Constructor returns a validation error. No SessionError is created.

- Condition: `Detail` is a JSON array (e.g., `[]` or `[1,2,3]`).
  Expected: Constructor returns a validation error. No SessionError is created.

- Condition: `Detail` is a JSON primitive (e.g., `"string"`, `123`, `true`, `null` as raw bytes).
  Expected: Constructor returns a validation error. No SessionError is created.

- Condition: `Detail` is nil.
  Expected: Constructor accepts this as valid (represents JSON null / not provided).

- Condition: `Detail` is invalid JSON bytes (e.g., `{broken`).
  Expected: Constructor returns a validation error. No SessionError is created.

- Condition: `OccurredAt` is 0 or negative.
  Expected: Constructor returns a validation error. No SessionError is created.

- Condition: `SessionID` is not a valid UUID format (e.g., `"not-a-uuid"`).
  Expected: Constructor returns a validation error. No SessionError is created.

- Condition: `FailingState` is an empty string.
  Expected: Constructor returns a validation error. No SessionError is created.

## Related

- [AgentError](./agent_error.md) — embeds SessionError, adds AgentRole field
- [RuntimeError](./runtime_error.md) — embeds SessionError, adds Issuer field
