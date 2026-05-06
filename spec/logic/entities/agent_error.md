# AgentError

## Overview

AgentError is a failure entity representing an unrecoverable error raised by an agent (or from a human node) during workflow execution. It embeds SessionError for shared error fields and adds the `AgentRole` field to identify the source agent. AgentError is a pure data entity — it does not perform runtime orchestration (halting state machines, persisting, notifying humans). Those behaviors are owned by the runtime.

## Boundaries

- Owns: construction-time validation of `AgentRole` and delegation to SessionError for shared field validation.
- Owns: immutability guarantee for all fields (including embedded SessionError fields) after construction.
- Delegates: shared field validation (Message, Detail, OccurredAt, SessionID, FailingState) to SessionError constructor.
- Delegates: semantic validation (session existence, workflow-defined role check, FailingState validity) to the runtime caller.
- Delegates: state machine halting, session status transition, persistence, and human notification to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be constructed via struct literal — must use the provided constructor.
- Must not: implement custom MarshalJSON — serialization of AgentError for persistence is entirely owned by SessionMetadataStore, which reads fields via getter methods.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `SessionError` | Embedded base struct | Embed and delegate shared field construction/validation | Must not bypass SessionError constructor |

Construction constraint: Must be constructed via `NewAgentError(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewAgentError(agentRole string, message string, detail json.RawMessage, occurredAt int64, sessionID string, failingState string)` that validates all fields and returns an immutable AgentError value.
2. Accepts `AgentRole` as any string including empty string (empty string represents a human node origin).
3. Delegates validation of Message, Detail, OccurredAt, SessionID, and FailingState to `NewSessionError`.
4. If SessionError construction fails, propagates the validation error.
5. All fields are immutable after construction.
6. Exposes all fields via exported getter methods (AgentRole getter + all SessionError getters).

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| AgentRole | string | Any string. Empty string `""` represents human node origin. | Yes (but may be empty) |
| Message | string | At least one non-whitespace character | Yes |
| Detail | json.RawMessage | nil or valid JSON object | No (nil allowed) |
| OccurredAt | int64 | Positive integer (> 0) | Yes |
| SessionID | string | Valid UUID format | Yes |
| FailingState | string | Non-empty | Yes |

## Outputs

### AgentError Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| AgentRole | string | Any string (empty for human nodes) | Agent role that raised the error, derived from current node's definition |
| Message | string | Non-empty | Human-readable error description |
| Detail | json.RawMessage | nil or valid JSON object | Additional error context |
| OccurredAt | int64 | Positive integer | POSIX timestamp when the error occurred |
| SessionID | string | Valid UUID format | Associated session identifier |
| FailingState | string | Non-empty | State machine node where the error occurred |

### Error Output

| Condition | Error |
|-----------|-------|
| SessionError validation fails | Propagated validation error from SessionError |

## Invariants

1. **AgentRole Accepts Any String**: `AgentRole` may be any string value including empty string. No format or non-empty constraint.
2. **SessionError Invariants Apply**: All invariants defined in SessionError apply to the embedded fields.
3. **Immutability**: Once constructed, no field (including AgentRole) may be modified.
4. **Construction Only Via Constructor**: Must be constructed via `NewAgentError`. Direct struct literal construction is forbidden.

## Edge Cases

- Condition: `AgentRole` is an empty string.
  Expected: Constructor accepts this as valid. Represents error originating from a human node.

- Condition: `AgentRole` contains arbitrary characters (spaces, special characters).
  Expected: Constructor accepts this as valid. Semantic validation of role names is not this entity's responsibility.

- Condition: Any shared field (Message, Detail, OccurredAt, SessionID, FailingState) violates SessionError constraints.
  Expected: Constructor returns the validation error from SessionError. No AgentError is created.

## Related

- [SessionError](./session_error.md) — embedded base struct providing shared fields and validation
- [RuntimeError](./runtime_error.md) — sibling error entity for runtime-originated failures
- [Event](./event.md) — events drive workflow progression; AgentError halts that mechanism (runtime responsibility)
