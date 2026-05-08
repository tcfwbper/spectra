# error Subcommand

## Overview

The `error` subcommand parses error-specific arguments (`<message>`, `--detail`, `--claude-session-id`), validates them, constructs the RuntimeMessage wire-format struct for error type, and delegates sending to `cmdutil.SendAndHandle`. It does not perform socket I/O directly.

## Boundaries

- Owns: subcommand Cobra definition and flag registration for `error`.
- Owns: positional argument `<message>` validation (non-empty).
- Owns: `--detail` JSON validation (must be a JSON object or null).
- Owns: RuntimeMessage wire-format struct construction for error type.
- Delegates: `--session-id` validation and project root discovery to the root command.
- Delegates: message sending, response interpretation, and exit code mapping to `cmdutil.SendAndHandle`.
- Must not: perform socket I/O.
- Must not: validate `--claude-session-id` format (any string accepted, Runtime validates).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `cmdutil.SendAndHandle` | Send and response handling | `SendAndHandle(sessionID, projectRoot, message, successText)` | Must not call SocketClient directly |
| Root command context | Provides sessionID and projectRoot | Read sessionID, projectRoot from shared context | Must not modify |

Construction constraint: Registered as a Cobra subcommand directly under root during command tree initialization. No separate constructor needed.

## Behavior

1. Defines Cobra command `error` with usage `"spectra-agent error <message> [flags]"`.
2. Registers flags: `--detail` (string, default `"{}"`), `--claude-session-id` (string, default `""`).
3. In `RunE`, reads `<message>` from `args[0]`. If missing (no positional args), returns error `"error message is required"` (exit code 1).
4. Validates that `<message>` is non-empty.
5. If `--detail` is provided, parses it as JSON. Validates it is either a JSON object `{}` or JSON `null`. Primitives (string, number, boolean) and arrays are rejected. If invalid, returns error `"--detail must be a JSON object or null"` (exit code 1).
6. Constructs the message struct matching the RuntimeMessage wire format:
   ```json
   {
     "type": "error",
     "claudeSessionID": "<claude-session-id-flag-value>",
     "payload": {
       "message": "<message>",
       "detail": <json-object-or-null>
     }
   }
   ```
7. Calls `cmdutil.SendAndHandle(sessionID, projectRoot, message, "Error reported successfully")`.
8. Returns the exit code, prints stdout/stderr as appropriate.

## Inputs

| Parameter | Type | Source | Constraints | Required | Default |
|-----------|------|--------|-------------|----------|---------|
| `<message>` | string | Positional arg | Non-empty | Yes | — |
| `--detail` | string | Flag | Must parse as JSON object or null | No | `"{}"` |
| `--claude-session-id` | string | Flag | Any string | No | `""` |
| sessionID | string | Root command context | Non-empty (validated by root) | Yes | — |
| projectRoot | string | Root command context | Absolute path (validated by root) | Yes | — |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| Exit code | int | 0, 1, 2, or 3 (from SendAndHandle or local validation) |
| stdout | string | `"Error reported successfully"` on success |
| stderr | string | Error message on failure |

## Invariants

1. **Error Message Required**: `<message>` must be non-empty. Validated before SendAndHandle is called.
2. **Detail Must Be JSON Object or Null**: `--detail` must parse as JSON object `{}` or `null`. Primitives and arrays are rejected with exit code 1.
3. **Default Values**: Omitted `--detail` → `{}`. Omitted `--claude-session-id` → `""`.
4. **No Direct Socket I/O**: All communication is delegated to SendAndHandle.
5. **Fixed Success Message**: On success, always outputs `"Error reported successfully"` regardless of Runtime's response message.
6. **Whitespace-Only Message Accepted**: The subcommand does not reject whitespace-only messages. Semantic validation is the Runtime's responsibility.

## Edge Cases

- Condition: `spectra-agent error` without `<message>`.
  Expected: Exit code 1, stderr `"Error: error message is required"`.

- Condition: `<message>` is whitespace-only (e.g., `"   "`).
  Expected: Accepted. Sent as-is. Semantic validation is Runtime's responsibility.

- Condition: `--detail '{"stack": "...", "code": 500}'` (valid JSON object).
  Expected: Accepted. Sent in message payload.

- Condition: `--detail 'null'` (JSON null).
  Expected: Accepted. Detail set to `null` in payload.

- Condition: `--detail '"string"'` (JSON primitive).
  Expected: Exit code 1, stderr `"Error: --detail must be a JSON object or null"`.

- Condition: `--detail '[1, 2, 3]'` (JSON array).
  Expected: Exit code 1, stderr `"Error: --detail must be a JSON object or null"`.

- Condition: `--detail '{invalid}'` (invalid JSON).
  Expected: Exit code 1, stderr `"Error: --detail must be a JSON object or null"`.

- Condition: `--detail` omitted.
  Expected: Detail defaults to `{}`.

- Condition: `--claude-session-id` omitted.
  Expected: ClaudeSessionID set to `""` in wire message.

## Related

- [Root](./root.md) — parent command, provides sessionID and projectRoot
- [SendAndHandle](../../cmdutil/send_and_handle.md) — handles send and response interpretation
- [RuntimeMessage wire format](../../../entities/runtime_message.md) — server-side entity that parses this message
- [ErrorProcessor](../../../runtime/error_processor.md) — server-side handler for error messages
