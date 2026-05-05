# RuntimeResponse

## Overview

RuntimeResponse is the structured response entity returned by the runtime to spectra-agent after processing a RuntimeMessage. It indicates whether the message was successfully processed or encountered an error, along with a human-readable message. RuntimeResponse is a pure data entity — it does not perform serialization, transmission, or connection management. Those behaviors are owned by the transport layer.

## Boundaries

- Owns: construction-time validation of all fields (Status, Message).
- Owns: immutability guarantee for all fields after construction.
- Delegates: serialization, wire-format (newline termination), transmission, and connection lifecycle to the transport layer.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `entities` package.
- Must not: be constructed via struct literal — must use the provided factory functions.

## Dependencies

None. This entity depends only on Go standard library types.

Construction constraint: Must be constructed via `SuccessResponse(message string)` or `ErrorResponse(message string)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a factory function `SuccessResponse(message string)` that creates an immutable RuntimeResponse with status `"success"` and the given message.
2. Provides a factory function `ErrorResponse(message string)` that creates an immutable RuntimeResponse with status `"error"` and the given message.
3. Accepts `Message` as any string including empty string.
4. All fields are immutable after construction.
5. Exposes all fields via exported getter methods.

## Inputs

### SuccessResponse

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Message | string | Any valid string (empty string allowed) | Yes (but may be empty) |

### ErrorResponse

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Message | string | Any valid string (empty string allowed) | Yes (but may be empty) |

## Outputs

### RuntimeResponse Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Status | string | `"success"` or `"error"` | Indicates whether the message processing succeeded or failed |
| Message | string | Any string (may be empty) | Human-readable result description or error details |

## Invariants

1. **Status Validity**: `Status` is always exactly `"success"` or `"error"`. No other values are possible because construction is only via the two factory functions.
2. **Message String**: `Message` must always be a valid string. It may be empty but must not be null (Go strings cannot be null, so this is naturally satisfied).
3. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
4. **Construction Only Via Factory Functions**: Must be constructed via `SuccessResponse` or `ErrorResponse`. Direct struct literal construction is forbidden.

## Edge Cases

- Condition: `Message` is an empty string.
  Expected: Factory function accepts this as valid. Empty messages are allowed for both success and error responses.

- Condition: `Message` contains newline characters (`\n`).
  Expected: Factory function accepts this as valid. The message content is not constrained by wire-format concerns (those are the transport layer's responsibility).

## Related

- [RuntimeMessage](./runtime_message.md) — request entity sent by spectra-agent to the runtime
