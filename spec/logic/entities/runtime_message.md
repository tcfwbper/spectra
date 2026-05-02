# RuntimeMessage

## Overview

RuntimeMessage is the structured message format used for communication between spectra-agent (client) and RuntimeSocketManager (server) over the runtime socket. Each message carries a type identifier and a type-specific payload. RuntimeMessage is serialized as JSON and transmitted over Unix domain sockets (or named pipes on Windows).

## Behavior

1. RuntimeMessage is created by spectra-agent when invoking `spectra-agent event emit` or `spectra-agent error`.
2. spectra-agent serializes the RuntimeMessage to JSON and sends it over the runtime socket connection, terminated by a newline (`\n`).
3. RuntimeSocketManager reads the JSON message from the socket, validates the structure, and parses it into a RuntimeMessage struct.
4. RuntimeSocketManager validates that the `type` field is present and is one of the recognized message types.
5. RuntimeSocketManager validates that the `payload` field is present and is a valid JSON object.
6. If validation fails (missing fields, invalid JSON, unrecognized type), RuntimeSocketManager logs a warning, sends an error response to the client, and closes the connection without invoking the MessageHandler.
7. If validation succeeds, RuntimeSocketManager invokes the MessageHandler callback with the parsed RuntimeMessage struct.
8. The MessageHandler processes the message based on its `type` and returns a RuntimeResponse.
9. RuntimeMessage is not persisted to disk. It is a transient communication protocol structure.

## Inputs

### For Message Creation (by spectra-agent)

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Type | string | Non-empty, must be a recognized message type ("event" or "error") | Yes |
| Payload | JSON object | Valid JSON object, structure depends on `Type` | Yes |
| ClaudeSessionID | string | Any valid string, may be empty `""` | No (defaults to `""` if omitted) |

### Payload Structure for Type="event"

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| EventType | string | Non-empty, PascalCase, must be a workflow-defined event type | Yes |
| Message | string | Any valid string, may be empty `""` | Yes (defaults to `""` if omitted by user) |
| Payload | JSON object | Valid JSON object, may be empty `{}` | Yes (defaults to `{}` if omitted by user) |

### Payload Structure for Type="error"

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Message | string | Non-empty, human-readable error description | Yes |
| Detail | JSON object | Valid JSON object, may be `null` or `{}` | No (defaults to `{}` if omitted) |

> **Note**: The `agentRole` field is **not** part of the error wire payload. It is derived server-side by ErrorProcessor from the current node's `agent_role` field (see [ErrorProcessor](../runtime/error_processor.md)).

## Outputs

### RuntimeMessage Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Type | string | Non-empty, recognized message type | Message type identifier: "event" or "error" |
| Payload | JSON object | Valid JSON object, structure determined by `Type` | Type-specific message payload |
| ClaudeSessionID | string | Any valid string, defaults to `""` | Claude session identifier for validation; passed from `--claude-session-id` flag |

### JSON Serialization Format

**Example 1: Event emission from agent**
```json
{"type": "event", "claudeSessionID": "550e8400-e29b-41d4-a716-446655440000", "payload": {"eventType": "DraftCompleted", "message": "Logic specification draft is ready for review", "payload": {"fileCount": 3, "totalLines": 450}}}
```

**Example 2: Error report from agent**
```json
{"type": "error", "claudeSessionID": "550e8400-e29b-41d4-a716-446655440000", "payload": {"message": "Failed to load workflow definition", "detail": {"error": "file not found", "path": ".spectra/workflows/DefaultLogicSpec.yaml"}}}
```

**Example 3: Event from human node (empty claudeSessionID)**
```json
{"type": "event", "claudeSessionID": "", "payload": {"eventType": "RequirementProvided", "message": "", "payload": {}}}
```

**Example 4: Event without claudeSessionID field (defaults to empty)**
```json
{"type": "event", "payload": {"eventType": "RequirementProvided", "message": "", "payload": {}}}
```

## Invariants

1. **Type Non-Empty**: The `type` field must be a non-empty string.

2. **Type Recognition**: The `type` field must be one of the recognized message types: `"event"` or `"error"`. Any other value is rejected with an error response.

3. **Payload Presence**: The `payload` field must be present and must be a valid JSON object `{}`. It must not be a JSON primitive, array, or null.

4. **Payload Structure Validation**: The structure of `payload` must conform to the schema defined for the given `type`. RuntimeSocketManager performs basic structure validation; semantic validation is performed by the MessageHandler.

5. **JSON Compliance**: RuntimeMessage must serialize to valid JSON and must be terminated by a newline (`\n`) when transmitted over the socket.

6. **Session ID Omission**: RuntimeMessage does not contain a `sessionID` field. The session is identified by the socket path (`.spectra/sessions/<SessionUUID>/runtime.sock`), which spectra-agent uses to connect to the correct RuntimeSocketManager instance.

7. **ClaudeSessionID Default Value**: If the `claudeSessionID` field is omitted from the JSON message, RuntimeSocketManager must treat it as an empty string `""`. This field is optional in the wire format but always present in the parsed RuntimeMessage struct.

8. **Message Size Limit**: RuntimeMessage serialized size must not exceed 10 MB. Messages exceeding this limit are rejected by RuntimeSocketManager.

## Edge Cases

- **Condition**: `type` field is missing from the JSON message.
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "missing required field 'type'", logs a warning, sends an error response to the client, and closes the connection.

- **Condition**: `type` field is an empty string `""`.
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "type field must not be empty", logs a warning, sends an error response, and closes the connection.

- **Condition**: `type` field is an unrecognized value (e.g., "unknown", "legacy").
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "invalid message type 'unknown'", logs a warning, sends an error response, and closes the connection.

- **Condition**: `payload` field is missing from the JSON message.
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "missing required field 'payload'", logs a warning, sends an error response, and closes the connection.

- **Condition**: `payload` is a JSON primitive (e.g., `"string"`, `123`, `true`, `null`).
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "payload must be a JSON object", logs a warning, sends an error response, and closes the connection.

- **Condition**: `payload` is a JSON array (e.g., `[]`, `[1, 2, 3]`).
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "payload must be a JSON object", logs a warning, sends an error response, and closes the connection.

- **Condition**: For `type="event"`, `payload.eventType` is missing.
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "event payload missing required field 'eventType'", logs a warning, sends an error response, and closes the connection.

- **Condition**: For `type="event"`, `payload.message` is missing.
  **Expected**: RuntimeSocketManager treats the missing field as an empty string `""` and accepts the message. This aligns with the Event entity default behavior.

- **Condition**: For `type="event"`, `payload.payload` is missing.
  **Expected**: RuntimeSocketManager treats the missing field as an empty JSON object `{}` and accepts the message. This aligns with the Event entity default behavior.

- **Condition**: For `type="error"`, `payload.message` is missing or empty.
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "error payload missing required field 'message'", logs a warning, sends an error response, and closes the connection.

- **Condition**: For `type="error"`, `payload.detail` is missing.
  **Expected**: RuntimeSocketManager treats the missing field as an empty JSON object `{}` and accepts the message.

- **Condition**: `claudeSessionID` field is missing from the JSON message.
  **Expected**: RuntimeSocketManager treats the missing field as an empty string `""` and accepts the message.

- **Condition**: `claudeSessionID` is a non-string value (e.g., `null`, number, object).
  **Expected**: RuntimeSocketManager rejects the message with a validation error: "claudeSessionID must be a string", logs a warning, sends an error response, and closes the connection.

- **Condition**: RuntimeMessage serialized JSON exceeds 10 MB.
  **Expected**: RuntimeSocketManager detects the size violation before fully reading the message, rejects it with a size limit error, logs a warning, sends an error response (if possible), and closes the connection.

- **Condition**: RuntimeMessage JSON is malformed (e.g., missing closing brace, invalid escape sequence).
  **Expected**: RuntimeSocketManager rejects the message with a JSON parse error, logs a warning, sends an error response (if possible), and closes the connection.

## Related

- [RuntimeResponse](./runtime_response.md) - Response structure returned by RuntimeSocketManager to spectra-agent
- [RuntimeSocketManager](../storage/runtime_socket_manager.md) - Server-side socket handler that parses and validates RuntimeMessage
- [Event](./event.md) - Event entity created from RuntimeMessage with `type="event"`
- [AgentError](./agent_error.md) - AgentError entity created from RuntimeMessage with `type="error"`
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
