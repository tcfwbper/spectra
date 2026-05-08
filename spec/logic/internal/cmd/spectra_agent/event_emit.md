# event emit Subcommand

## Overview

The `event emit` subcommand parses event-specific arguments (`<type>`, `--message`, `--payload`, `--claude-session-id`), validates them, constructs the RuntimeMessage wire-format struct, and delegates sending to `cmdutil.SendAndHandle`. It does not perform socket I/O directly.

## Boundaries

- Owns: subcommand Cobra definition and flag registration for `event emit`.
- Owns: positional argument `<type>` validation (non-empty).
- Owns: `--payload` JSON validation (must be a JSON object).
- Owns: RuntimeMessage wire-format struct construction for event type.
- Delegates: `--session-id` validation and project root discovery to the root command.
- Delegates: message sending, response interpretation, and exit code mapping to `cmdutil.SendAndHandle`.
- Must not: perform socket I/O.
- Must not: validate event type against the workflow definition (that is the Runtime's responsibility).
- Must not: validate `--claude-session-id` format (any string accepted, Runtime validates).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `cmdutil.SendAndHandle` | Send and response handling | `SendAndHandle(sessionID, projectRoot, message, successText)` | Must not call SocketClient directly |
| Root command context | Provides sessionID and projectRoot | Read sessionID, projectRoot from shared context | Must not modify |

Construction constraint: Registered as a Cobra subcommand under `event` during root command initialization. No separate constructor needed.

## Behavior

1. Defines Cobra command `emit` nested under `event`, with usage `"spectra-agent event emit <type> [flags]"`.
2. Registers flags: `--message` (string, default `""`), `--payload` (string, default `"{}"`), `--claude-session-id` (string, default `""`).
3. In `RunE`, reads `<type>` from `args[0]`. If missing (no positional args), returns error `"event type is required"` (exit code 1).
4. Validates that `<type>` is non-empty.
5. If `--payload` is provided, parses it as JSON. Validates it is a JSON object (starts with `{`, parses as `map[string]any`). If invalid, returns error `"--payload must be a valid JSON object, e.g., {}"` (exit code 1).
6. Constructs the message struct matching the RuntimeMessage wire format:
   ```json
   {
     "type": "event",
     "claudeSessionID": "<claude-session-id-flag-value>",
     "payload": {
       "eventType": "<type>",
       "message": "<message-flag-value>",
       "payload": <parsed-json-object>
     }
   }
   ```
7. Calls `cmdutil.SendAndHandle(sessionID, projectRoot, message, "Event emitted successfully")`.
8. Returns the exit code, prints stdout/stderr as appropriate.

## Inputs

| Parameter | Type | Source | Constraints | Required | Default |
|-----------|------|--------|-------------|----------|---------|
| `<type>` | string | Positional arg | Non-empty | Yes | — |
| `--message` | string | Flag | Any string | No | `""` |
| `--payload` | string | Flag | Must parse as JSON object `{}` | No | `"{}"` |
| `--claude-session-id` | string | Flag | Any string | No | `""` |
| sessionID | string | Root command context | Non-empty (validated by root) | Yes | — |
| projectRoot | string | Root command context | Absolute path (validated by root) | Yes | — |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| Exit code | int | 0, 1, 2, or 3 (from SendAndHandle or local validation) |
| stdout | string | `"Event emitted successfully"` on success |
| stderr | string | Error message on failure |

## Invariants

1. **Event Type Required**: `<type>` must be non-empty. Validated before SendAndHandle is called.
2. **Payload Must Be JSON Object**: `--payload` must parse as a JSON object. Primitives and arrays are rejected with exit code 1.
3. **Default Values**: Omitted `--message` → `""`. Omitted `--payload` → `{}`. Omitted `--claude-session-id` → `""`.
4. **No Direct Socket I/O**: All communication is delegated to SendAndHandle.
5. **Fixed Success Message**: On success, always outputs `"Event emitted successfully"` regardless of Runtime's response message.

## Edge Cases

- Condition: `spectra-agent event emit` without `<type>`.
  Expected: Exit code 1, stderr `"Error: event type is required"`.

- Condition: `--payload '{"key": "value"}'` (valid JSON object).
  Expected: Accepted. Sent in message payload.

- Condition: `--payload '"string"'` (JSON primitive).
  Expected: Exit code 1, stderr `"Error: --payload must be a valid JSON object, e.g., {}"`.

- Condition: `--payload '[1, 2, 3]'` (JSON array).
  Expected: Exit code 1, stderr `"Error: --payload must be a valid JSON object, e.g., {}"`.

- Condition: `--payload '{invalid}'` (invalid JSON).
  Expected: Exit code 1, stderr `"Error: --payload must be a valid JSON object, e.g., {}"`.

- Condition: `--message ''` (explicit empty string).
  Expected: Accepted. Message field set to `""`.

- Condition: `--claude-session-id` omitted.
  Expected: ClaudeSessionID set to `""` in wire message.

## Related

- [Root](./root.md) — parent command, provides sessionID and projectRoot
- [SendAndHandle](../../cmdutil/send_and_handle.md) — handles send and response interpretation
- [RuntimeMessage wire format](../../../entities/runtime_message.md) — server-side entity that parses this message
- [RuntimeSocketManager](../../../storage/runtime_socket_manager.md) — server-side protocol handler
