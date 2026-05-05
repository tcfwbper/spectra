# RuntimeError

## Overview

RuntimeError is a failure entity representing an unrecoverable error raised by a runtime component during session execution (e.g., socket creation failure, state transition failure, panic in message processing). It embeds SessionError for shared error fields and adds the `Issuer` field to identify the runtime component that raised the error. RuntimeError is a pure data entity — it does not perform runtime orchestration (halting state machines, persisting, notifying humans). Those behaviors are owned by the runtime.

## Boundaries

- Owns: construction-time validation of `Issuer` (non-empty) and delegation to SessionError for shared field validation.
- Owns: immutability guarantee for all fields (including embedded SessionError fields) after construction.
- Delegates: shared field validation (Message, Detail, OccurredAt, SessionID, FailingState) to SessionError constructor.
- Delegates: semantic validation (session existence, FailingState validity, component name verification) to the runtime caller.
- Delegates: state machine halting, session status transition, persistence, and human notification to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be constructed via struct literal — must use the provided constructor.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `SessionError` | Embedded base struct | Embed and delegate shared field construction/validation | Must not bypass SessionError constructor |

Construction constraint: Must be constructed via `NewRuntimeError(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewRuntimeError(issuer string, message string, detail json.RawMessage, occurredAt int64, sessionID string, failingState string)` that validates all fields and returns an immutable RuntimeError value.
2. Validates that `Issuer` contains at least one non-whitespace character.
3. Delegates validation of Message, Detail, OccurredAt, SessionID, and FailingState to `NewSessionError`.
4. If Issuer validation fails, returns a validation error.
5. If SessionError construction fails, propagates the validation error.
6. All fields are immutable after construction.
7. Exposes all fields via exported getter methods (Issuer getter + all SessionError getters).

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Issuer | string | At least one non-whitespace character (e.g., "MessageRouter", "RuntimeSocketManager", "EventProcessor") | Yes |
| Message | string | At least one non-whitespace character | Yes |
| Detail | json.RawMessage | nil or valid JSON object | No (nil allowed) |
| OccurredAt | int64 | Positive integer (> 0) | Yes |
| SessionID | string | Valid UUID format | Yes |
| FailingState | string | Non-empty | Yes |

## Outputs

### RuntimeError Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Issuer | string | Non-empty (at least one non-whitespace character) | Runtime component that raised the error |
| Message | string | Non-empty | Human-readable error description |
| Detail | json.RawMessage | nil or valid JSON object | Additional error context (e.g., stack trace, system error) |
| OccurredAt | int64 | Positive integer | POSIX timestamp when the error occurred |
| SessionID | string | Valid UUID format | Associated session identifier |
| FailingState | string | Non-empty | State machine node where the error occurred |

### Error Output

| Condition | Error |
|-----------|-------|
| Issuer is empty or whitespace-only | Validation error |
| SessionError validation fails | Propagated validation error from SessionError |

## Invariants

1. **Issuer Non-Empty**: `Issuer` must contain at least one non-whitespace character. The entity does not validate that the issuer name matches a known runtime component; any non-empty string is accepted.
2. **SessionError Invariants Apply**: All invariants defined in SessionError apply to the embedded fields.
3. **Immutability**: Once constructed, no field (including Issuer) may be modified.
4. **Construction Only Via Constructor**: Must be constructed via `NewRuntimeError`. Direct struct literal construction is forbidden.

## Edge Cases

- Condition: `Issuer` is an empty string or contains only whitespace.
  Expected: Constructor returns a validation error. No RuntimeError is created.

- Condition: `Issuer` is an arbitrary non-empty string that does not match any known component name.
  Expected: Constructor accepts this as valid. Component name validation is not this entity's responsibility.

- Condition: Any shared field (Message, Detail, OccurredAt, SessionID, FailingState) violates SessionError constraints.
  Expected: Constructor returns the validation error from SessionError. No RuntimeError is created.

## Related

- [SessionError](./session_error.md) — embedded base struct providing shared fields and validation
- [AgentError](./agent_error.md) — sibling error entity for agent-originated failures
- [Event](./event.md) — events drive workflow progression; RuntimeError halts that mechanism (runtime responsibility)
