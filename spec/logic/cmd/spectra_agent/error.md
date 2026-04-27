# error Subcommand

## Overview

The `error` subcommand reports an unrecoverable error to the workflow runtime, causing the session to transition to `"failed"` status. It constructs a `RuntimeMessage` with type "error", sends it to the session's runtime socket via `SocketClient`, and waits for the Runtime's response. The error message and optional detail are provided as command-line arguments and flags.

## Behavior

### Command Syntax

**Usage**: `spectra-agent error <message> --session-id <UUID> [--claude-session-id <UUID>] [--detail <json>]`

**Description**: Reports an unrecoverable error with optional detail to halt the workflow and mark the session as failed.

### Execution Flow

1. The subcommand receives the session ID and project root from the root command.
2. `<message>` is a required positional argument representing a human-readable error description.
3. `<message>` must be a non-empty string. If omitted or empty, the subcommand exits with code 1 and prints: `"Error: error message is required"`
4. `--detail <json>` is an optional flag. If omitted, it defaults to an empty JSON object `{}`.
5. `--claude-session-id <UUID>` is an optional flag. If omitted, it defaults to an empty string `""`. Any string value is accepted.
6. The subcommand validates that `--detail`, if provided, is valid JSON and is either a JSON object `{}` or `null`. JSON primitives (string, number, boolean) and arrays are rejected.
7. If `--detail` is invalid JSON or not an object/null, the subcommand exits with code 1 and prints: `"Error: --detail must be a JSON object or null"`
8. The subcommand constructs a `RuntimeMessage` JSON object:
   ```json
   {
     "type": "error",
     "claudeSessionID": "<claude-session-id>",
     "payload": {
       "message": "<message>",
       "detail": <json-object-or-null>
     }
   }
   ```
9. The subcommand invokes `SocketClient.Send(sessionID, projectRoot, runtimeMessage)` to send the message and receive the response.
10. `SocketClient` handles all socket operations (connect, send, receive, close) and returns a response or error.
11. If `SocketClient` returns a transport error (socket not found, connection failed, timeout, I/O error), the subcommand exits with code 2 and prints the error to stderr.
12. If `SocketClient` returns a Runtime response with status `"success"`, the subcommand prints `"Error reported successfully"` to stdout and exits with code 0.
13. If `SocketClient` returns a Runtime response with status `"error"`, the subcommand prints the error message to stderr (prefixed with `"Error: "`) and exits with code 3. This includes validation errors from Claude session ID mismatches.
14. If `SocketClient` returns a malformed response (invalid JSON, missing `status` field), the subcommand exits with code 3 and prints an appropriate error to stderr.

## Inputs

### From Root Command

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Session ID | string (UUID) | `--session-id` flag (parsed by root) | Yes |
| Project Root | string | Discovered by `SpectraFinder` in root command | Yes |

### Subcommand-Specific Parameters

| Parameter | Type | Constraints | Required | Default |
|-----------|------|-------------|----------|---------|
| `<message>` | string (positional) | Non-empty string | Yes | None |
| `--detail` | JSON string | Must parse as valid JSON (object or null) | No | `{}` |
| `--claude-session-id` | string | Any string, including empty | No | `""` |

## Outputs

### stdout (on success, exit code 0)

```
Error reported successfully
```

### stderr (on failure, exit codes 1/2/3)

**Exit Code 1 (Invocation Error) Examples**:
```
Error: error message is required
Error: --detail must be a JSON object or null
```

**Exit Code 2 (Transport Error) Examples**:
```
Error: socket file not found: /path/to/.spectra/sessions/<uuid>/runtime.sock
Error: connection refused: Runtime is not running for session <uuid>
Error: connection timeout after 30s
Error: failed to send message: <io-error>
Error: failed to read response: <io-error>
```

**Exit Code 3 (Runtime Execution Error) Examples**:
```
Error: session not found: <uuid>
Error: session terminated: session is in 'completed' status
Error: claude session ID mismatch: expected 550e8400-e29b-41d4-a716-446655440000 but got 660e8400-e29b-41d4-a716-446655440001
Error: invalid claude session ID for human node: must be empty
Error: claude session ID not found for node 'Architect'
Error: malformed response from Runtime: <json-parse-error>
Error: response missing 'status' field
```

### Exit Codes

Inherits exit code scheme from root command (0, 1, 2, 3).

## Invariants

1. **Error Message Required**: The `<message>` positional argument must be non-empty. This is validated before socket communication.

2. **Detail Type Constraint**: The `--detail` flag accepts only JSON objects `{}` or `null`. JSON primitives (strings, numbers, booleans) and arrays are rejected with exit code 1, matching the AgentError and RuntimeMessage entity constraints.

3. **Default Detail Value**: Omitted `--detail` defaults to `{}` (empty JSON object). Omitted `--claude-session-id` defaults to `""`.

4. **Message Format**: The constructed `RuntimeMessage` must conform to the protocol expected by `RuntimeSocketManager`.

5. **SocketClient Delegation**: All socket operations (connect, send, receive, close) are delegated to `SocketClient`. The subcommand does not perform direct socket I/O.

6. **Error Propagation**: Transport errors (exit code 2) and Runtime execution errors (exit code 3) are propagated from `SocketClient` without modification.

7. **Success Message**: On success, the subcommand always prints `"Error reported successfully"` to stdout, regardless of the Runtime's response message content.

## Edge Cases

- **Condition**: User invokes `spectra-agent error` without providing `<message>`.
  **Expected**: Exit with code 1, print `"Error: error message is required"` to stderr.

- **Condition**: User provides `<message>` with only whitespace (e.g., `"   "`).
  **Expected**: The subcommand accepts the value (whitespace-only validation is Runtime's responsibility). Message is sent successfully.

- **Condition**: User provides `--detail 'null'` (JSON null).
  **Expected**: Detail is set to `null` and the error is reported successfully.

- **Condition**: User provides `--detail '{"stack": "...", "code": 500}'` (valid JSON object).
  **Expected**: Detail is set to the object and the error is reported successfully.

- **Condition**: User provides `--detail '"error detail string"'` (JSON primitive string).
  **Expected**: Exit with code 1, print `"Error: --detail must be a JSON object or null"` to stderr.

- **Condition**: User provides `--detail '[1, 2, 3]'` (JSON array).
  **Expected**: Exit with code 1, print `"Error: --detail must be a JSON object or null"` to stderr.

- **Condition**: User provides `--detail '{invalid json}'`.
  **Expected**: Exit with code 1, print `"Error: --detail must be a JSON object or null"` to stderr.

- **Condition**: User omits `--detail` flag.
  **Expected**: Detail defaults to `{}` and the error is reported successfully.

- **Condition**: Socket file does not exist (session not running or session ID invalid).
  **Expected**: `SocketClient` returns transport error. Exit with code 2, print socket-not-found error to stderr.

- **Condition**: Runtime responds with `{"status": "success", "message": ""}` (empty message).
  **Expected**: Exit with code 0, print `"Error reported successfully"` to stdout (default success message).

- **Condition**: Runtime responds with `{"status": "success", "message": "Session marked as failed"}`.
  **Expected**: Exit with code 0, print `"Error reported successfully"` to stdout (ignore custom message, use default).

- **Condition**: Runtime responds with `{"status": "error", "message": "session not found: abc-123"}`.
  **Expected**: Exit with code 3, print `"Error: session not found: abc-123"` to stderr.

- **Condition**: Runtime responds with `{"status": "error", "message": "session terminated: session already failed"}`.
  **Expected**: Exit with code 3, print `"Error: session terminated: session already failed"` to stderr.

- **Condition**: User provides `--claude-session-id '550e8400-e29b-41d4-a716-446655440000'` for an agent node, but the Runtime's stored Claude session ID does not match.
  **Expected**: Runtime validates and rejects. Exit with code 3, print `"Error: claude session ID mismatch: expected <stored-uuid> but got 550e8400-e29b-41d4-a716-446655440000"` to stderr. The error is not recorded.

- **Condition**: User provides `--claude-session-id '550e8400-e29b-41d4-a716-446655440000'` for a human node.
  **Expected**: Runtime validates and rejects. Exit with code 3, print `"Error: invalid claude session ID for human node: must be empty"` to stderr. The error is not recorded.

- **Condition**: User omits `--claude-session-id` for an agent node (defaults to empty string).
  **Expected**: Runtime validates and rejects (because agent nodes require non-empty Claude session ID). Exit with code 3, print `"Error: claude session ID mismatch: expected <stored-uuid> but got "` to stderr. The error is not recorded.

- **Condition**: User omits `--claude-session-id` for a human node (defaults to empty string).
  **Expected**: Runtime validates and accepts. Error is reported successfully.

- **Condition**: User provides `--claude-session-id ''` (empty string) for a human node.
  **Expected**: Runtime validates and accepts. Error is reported successfully.

- **Condition**: Runtime responds with malformed JSON.
  **Expected**: `SocketClient` detects malformed response. Exit with code 3, print JSON parse error to stderr.

- **Condition**: Runtime responds with valid JSON but missing `status` field.
  **Expected**: `SocketClient` detects missing field. Exit with code 3, print `"Error: response missing 'status' field"` to stderr.

## Expected Usage Pattern

This subcommand is normally invoked from inside a Claude agent process spawned by AgentInvoker. The agent's system prompt instructs it to forward `$SPECTRA_SESSION_ID` and `$SPECTRA_CLAUDE_SESSION_ID` from its environment as flags:

```
spectra-agent error "<human-readable error>" \
  --session-id "$SPECTRA_SESSION_ID" \
  --claude-session-id "$SPECTRA_CLAUDE_SESSION_ID" \
  [--detail '{"...":"..."}']
```

For human nodes, `--claude-session-id` must be omitted or empty. Calling without `--claude-session-id` from an agent node will be rejected by Runtime with exit code 3 (Claude session ID mismatch).

See [spectra-agent root](./root.md#caller-responsibility-expected-usage-pattern) for the complete contract.

## Related

- [Root Command](./root.md) - Handles global flags and initialization
- [SocketClient](./client.md) - Performs socket communication
- [AgentError Entity](../../entities/agent_error.md) - Error reporting structure and behavior
- [RuntimeSocketManager](../../storage/runtime_socket_manager.md) - Server-side message handling
