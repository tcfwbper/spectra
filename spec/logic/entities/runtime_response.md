# RuntimeResponse

## Overview

RuntimeResponse is the structured response format returned by RuntimeSocketManager to spectra-agent after processing a RuntimeMessage. It indicates whether the message was successfully processed or encountered an error, along with a human-readable message. RuntimeResponse is used exclusively for socket communication protocol and is not used for internal Runtime component communication. Internal components use Go error types and custom error structs for richer error handling.

## Behavior

1. RuntimeResponse is created by the MessageHandler callback after processing a RuntimeMessage.
2. The MessageHandler returns a RuntimeResponse struct indicating success or error.
3. RuntimeSocketManager serializes the RuntimeResponse to JSON and sends it back to the spectra-agent client over the same socket connection, terminated by a newline (`\n`).
4. After sending the response, RuntimeSocketManager closes the connection to signal completion.
5. spectra-agent reads the JSON response, parses it, and uses the `status` field to determine the exit code (0 for success, non-zero for error).
6. RuntimeResponse is not persisted to disk. It is a transient communication protocol structure.

## Inputs

### For Response Creation (by MessageHandler)

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Status | string | Enum: "success" or "error" | Yes |
| Message | string | Human-readable message, may be empty `""` | Yes |

## Outputs

### RuntimeResponse Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Status | string | Enum: "success" or "error" | Indicates whether the message processing succeeded or failed |
| Message | string | Human-readable text, may be empty `""` | Describes the result or provides error details |

### JSON Serialization Format

**Example 1: Success response**
```json
{"status": "success", "message": "Event 'DraftCompleted' recorded successfully"}
```

**Example 2: Error response**
```json
{"status": "error", "message": "session not ready: status is 'initializing'"}
```

**Example 3: Success with empty message**
```json
{"status": "success", "message": ""}
```

**Example 4: Error with detailed message**
```json
{"status": "error", "message": "no transition found for event 'UnknownEvent' from node 'Architect'"}
```

## Invariants

1. **Status Validation**: The `status` field must be exactly "success" or "error". No other values are allowed.

2. **Message String Compliance**: The `message` field must always be a valid string. It may be empty `""`, but must not be null.

3. **JSON Compliance**: RuntimeResponse must serialize to valid JSON and must be terminated by a newline (`\n`) when transmitted over the socket.

4. **Single Response per Connection**: Each socket connection receives exactly one RuntimeResponse before the connection is closed by RuntimeSocketManager.

5. **Protocol-Only Usage**: RuntimeResponse is used exclusively for socket communication between RuntimeSocketManager and spectra-agent. Internal Runtime components (EventProcessor, ErrorProcessor, TransitionEvaluator, etc.) use Go error types and custom error structs for error handling.

6. **No Additional Fields**: RuntimeResponse contains only `status` and `message` fields. It does not carry additional context (e.g., error codes, event IDs, stack traces). Rich error details are handled internally by Runtime components and mapped to a simple message string at the socket boundary.

## Edge Cases

- **Condition**: MessageHandler returns a RuntimeResponse with `status` set to an invalid value (e.g., "warning", "unknown").
  **Expected**: RuntimeSocketManager detects the invalid status during serialization and replaces it with `{"status": "error", "message": "internal error: invalid response status"}` before sending to the client. It also logs a warning about the invalid status.

- **Condition**: MessageHandler returns a RuntimeResponse with `message` set to `null`.
  **Expected**: RuntimeSocketManager detects the null value during serialization and replaces it with an empty string `""` before sending to the client. It also logs a warning about the null message.

- **Condition**: MessageHandler returns a RuntimeResponse with `message` containing newline characters (`\n`).
  **Expected**: RuntimeSocketManager serializes the message as-is, properly escaped in JSON (e.g., `"message": "line1\nline2"`). The newline inside the message does not interfere with the message terminator newline that follows the JSON object.

- **Condition**: MessageHandler returns a RuntimeResponse with an empty `message` field (`""`).
  **Expected**: RuntimeSocketManager sends the response as-is: `{"status": "success", "message": ""}` or `{"status": "error", "message": ""}`. Empty messages are valid.

- **Condition**: Sending the RuntimeResponse to the client fails due to an I/O error (e.g., client disconnected prematurely).
  **Expected**: RuntimeSocketManager logs a warning: "failed to send response to client: <error>" and closes the connection. No retry is attempted.

- **Condition**: MessageHandler panics while processing the message.
  **Expected**: RuntimeSocketManager recovers from the panic, logs the error with stack trace, sends a RuntimeResponse with `{"status": "error", "message": "internal server error"}`, and closes the connection.

- **Condition**: RuntimeResponse JSON serialization exceeds 10 MB (extremely unlikely given the simple structure).
  **Expected**: RuntimeSocketManager logs an error: "response serialization exceeded size limit" and closes the connection without sending a response. This is a pathological case (e.g., message field contains gigabytes of text).

## Related

- [RuntimeMessage](./runtime_message.md) - Request structure sent by spectra-agent to RuntimeSocketManager
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Server-side socket handler that sends RuntimeResponse
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
