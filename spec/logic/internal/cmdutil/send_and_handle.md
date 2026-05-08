# SendAndHandle

## Overview

SendAndHandle is a shared function that encapsulates the common "serialize message, send via SocketClient, interpret response, produce user-facing output" flow used by all spectra-agent subcommands. Each subcommand builds a message payload and success text, then delegates the transport and response handling to this function. SendAndHandle does not parse subcommand arguments — it operates on an already-constructed message.

## Boundaries

- Owns: JSON serialization of the message struct into wire-format bytes.
- Owns: invoking SocketClient.Send and interpreting the result.
- Owns: producing the final stdout/stderr output strings and exit code.
- Delegates: socket communication to SocketClient.
- Delegates: message construction (field population) to the calling subcommand.
- Delegates: argument parsing and validation to the calling subcommand.
- Must not: parse CLI flags or positional arguments.
- Must not: perform socket I/O directly.
- Must not: know which specific subcommand is calling it.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `SocketClient` | Transport | `Send(sessionID, projectRoot, messageBytes)` | Must not access SocketClient internals |
| `ErrorFormatter` | Output formatting | `FormatError(msg)` | — |

Construction constraint: Package-level function. No instantiation needed.

## Behavior

1. `SendAndHandle(sessionID, projectRoot string, message any, successText string) (exitCode int, stdout string, stderr string)`.
2. Serializes `message` to JSON bytes. If serialization fails (programming error — e.g., unencodable type), returns exit code 1, stderr with `"failed to serialize message: <error>"`.
3. Calls `SocketClient.Send(sessionID, projectRoot, jsonBytes)`.
4. If SocketClient returns exit code 2 (transport error), returns exit code 2, empty stdout, stderr set to the formatted error from SocketClient.
5. If SocketClient returns exit code 3 and a nil Response (malformed response / missing fields), returns exit code 3, empty stdout, stderr set to the formatted error from SocketClient.
6. If SocketClient returns exit code 3 and a non-nil Response with status "error", returns exit code 3, empty stdout, stderr set to `FormatError(response.Message)`.
7. If SocketClient returns exit code 0 (success), returns exit code 0, stdout set to `successText`, empty stderr.

## Inputs

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| sessionID | string | Non-empty | Yes |
| projectRoot | string | Absolute path | Yes |
| message | any (struct) | JSON-serializable struct matching RuntimeMessage wire format | Yes |
| successText | string | Non-empty human-readable success message | Yes |

### Message Wire Format

The `message` parameter must serialize to the RuntimeMessage wire format expected by RuntimeSocketManager:

```json
{
  "type": "event" | "error",
  "claudeSessionID": "<string>",
  "payload": { ... }
}
```

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| exitCode | int | 0, 1, 2, or 3 |
| stdout | string | Text to print to stdout (non-empty only on success) |
| stderr | string | Text to print to stderr (non-empty only on failure) |

## Invariants

1. **Transparent Exit Codes**: Exit codes from SocketClient are passed through unchanged. Only serialization failure produces exit code 1.
2. **Success Text on Success Only**: `stdout` is non-empty only when exitCode is 0.
3. **Error Text on Failure Only**: `stderr` is non-empty only when exitCode is non-zero.
4. **No Argument Parsing**: SendAndHandle does not access CLI flags or arguments.
5. **Single Responsibility**: Handles only the send-and-interpret flow. Message construction is the caller's job.

## Edge Cases

- Condition: `message` contains a field that cannot be serialized to JSON (e.g., channel, function).
  Expected: Returns exit code 1, stderr `"Error: failed to serialize message: <error>"`.

- Condition: SocketClient returns exit code 2 with error "socket file not found".
  Expected: Returns exit code 2, stderr `"Error: socket file not found: <path>"`.

- Condition: SocketClient returns exit code 0 with a success Response.
  Expected: Returns exit code 0, stdout set to `successText`.

- Condition: SocketClient returns exit code 3 with Response.Message = "session not found: abc".
  Expected: Returns exit code 3, stderr `"Error: session not found: abc"`.

- Condition: SocketClient returns exit code 3 with nil Response (malformed response).
  Expected: Returns exit code 3, stderr from SocketClient error (e.g., `"Error: malformed response from Runtime: ..."`).

## PublicSendAndHandle

`PublicSendAndHandle` is an exported convenience wrapper that wires production dependencies (a real SocketClient using storage layout for socket path resolution, and `FormatError` as the error formatter) and delegates to `SendAndHandle`. It exists so that packages outside `cmdutil` (e.g., `spectra-agent` production adapters) can invoke the send-and-handle flow without accessing unexported types.

### Behavior

1. `PublicSendAndHandle(sessionID, projectRoot string, message any, successText string) (exitCode int, stdout string, stderr string)`.
2. Constructs a production SocketClient that resolves socket paths via `storage.GetRuntimeSocketPath`.
3. Calls `SendAndHandle(productionClient, FormatError, sessionID, projectRoot, message, successText)`.
4. Returns the result unchanged.

### Boundaries

- Owns: wiring production SocketClient and error formatter.
- Delegates: all send-and-handle logic to `SendAndHandle`.
- Must not: contain any logic beyond construction and delegation.

## Related

- [SocketClient](./socket_client.md) — performs the actual socket communication
- [ExitCodes](./exit_codes.md) — exit code constants
- [ErrorFormatter](./error_formatter.md) — formats error messages
- [EventEmit](../cmd/spectra_agent/event_emit.md) — caller that builds event messages
- [Error](../cmd/spectra_agent/error_cmd.md) — caller that builds error messages
