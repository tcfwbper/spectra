# RuntimeMessage

## Overview

RuntimeMessage is the structured message entity used for communication between spectra-agent (client) and the runtime (server) over the runtime socket. Each message carries a type identifier, a type-specific payload, and an optional Claude session identifier. RuntimeMessage is a pure data entity — it does not perform serialization, transmission, size checking, or payload semantic validation. Those behaviors are owned by the transport layer.

## Boundaries

- Owns: construction-time validation of all fields (Type, Payload, ClaudeSessionID).
- Owns: immutability guarantee for all fields after construction.
- Delegates: payload internal structure validation (e.g., eventType presence for event messages) to the runtime message handler.
- Delegates: serialization, wire-format (newline termination), transmission, and size limit enforcement to the transport layer.
- Delegates: semantic validation (session existence, event type defined in workflow) to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be constructed via struct literal — must use the provided constructor.
- Must not: validate payload internal structure based on the message type.

## Dependencies

None. This entity depends only on Go standard library types (`json.RawMessage`).

Construction constraint: Must be constructed via `NewRuntimeMessage(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewRuntimeMessage(msgType string, payload json.RawMessage, claudeSessionID string)` that validates all fields and returns an immutable RuntimeMessage value.
2. Validates that `Type` is a non-empty string and is one of the recognized message types: `"event"` or `"error"`.
3. Validates that `Payload` is a valid JSON object (starts with `{`). Must not be nil, a JSON primitive, or a JSON array.
4. Accepts `ClaudeSessionID` as any string including empty string (empty string is the default when omitted by the wire-format caller).
5. Returns a validation error if any constraint is violated.
6. All fields are immutable after construction.
7. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Type | string | Non-empty, must be one of: `"event"`, `"error"` | Yes |
| Payload | json.RawMessage | Valid JSON object (`{...}`). Must not be nil, a primitive, or an array. | Yes |
| ClaudeSessionID | string | Any valid string (empty string allowed) | Yes (but may be empty) |

## Outputs

### RuntimeMessage Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Type | string | `"event"` or `"error"` | Message type identifier |
| Payload | json.RawMessage | Valid JSON object | Type-specific message payload |
| ClaudeSessionID | string | Any string (may be empty) | Claude session identifier for agent validation |

### Error Output

| Condition | Error |
|-----------|-------|
| Type is empty | Validation error |
| Type is not `"event"` or `"error"` | Validation error |
| Payload is nil | Validation error |
| Payload is a JSON primitive or array | Validation error |
| Payload is invalid JSON | Validation error |

## Invariants

1. **Type Recognition**: `Type` must be one of the recognized message types: `"event"` or `"error"`. No other values are allowed.
2. **Payload JSON Object**: `Payload` must be a valid JSON object. It must not be nil, a JSON primitive, or a JSON array.
3. **ClaudeSessionID Accepts Any String**: `ClaudeSessionID` may be any string value including empty string. No format constraint.
4. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
5. **Construction Only Via Constructor**: Must be constructed via `NewRuntimeMessage`. Direct struct literal construction is forbidden.
6. **No Payload Semantic Validation**: RuntimeMessage does not validate the internal structure of `Payload`. It only ensures `Payload` is a syntactically valid JSON object.

## Edge Cases

- Condition: `Type` is an empty string.
  Expected: Constructor returns a validation error. No RuntimeMessage is created.

- Condition: `Type` is an unrecognized value (e.g., `"unknown"`, `"legacy"`, `"warning"`).
  Expected: Constructor returns a validation error indicating the type is not recognized.

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

- Condition: `ClaudeSessionID` is an empty string.
  Expected: Constructor accepts this as valid.

## Related

- [RuntimeResponse](./runtime_response.md) — response entity returned to spectra-agent after message processing
- [Event](./event.md) — event entity created from RuntimeMessage with type `"event"` (by the runtime)
- [AgentError](./agent_error.md) — agent error entity created from RuntimeMessage with type `"error"` (by the runtime)
