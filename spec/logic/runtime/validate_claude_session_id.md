# ValidateClaudeSessionID

## Overview

ValidateClaudeSessionID is a package-level helper function that validates the `claudeSessionID` from a RuntimeMessage against the current node's requirements. It encapsulates the shared validation logic used by both EventProcessor and ErrorProcessor, eliminating duplication. ValidateClaudeSessionID does not modify session state, record events, or perform any lifecycle transitions.

## Boundaries

- Owns: Claude session ID validation logic based on current node type.
- Delegates: node type determination to the caller (who provides the node definition).
- Delegates: session data access to Session's thread-safe methods.
- Must not: modify session state or session data.
- Must not: record events or errors.
- Must not: perform lifecycle transitions (Fail, Done, Run).
- Must not: perform any I/O or filesystem operations.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State container with auto-persist | `GetSessionDataSafe(key)` (read only) | Must not call any mutating method |
| `Node` | Configuration source | Read `Type()` and `Name()` via getters | Must not modify |

No construction constraint: ValidateClaudeSessionID is a stateless package-level function.

## Behavior

1. Receives the PersistentSession reference, the current Node definition, and the `claudeSessionID` string from the RuntimeMessage.
2. If the node's `Type() == "agent"`:
   - Calls `PersistentSession.GetSessionDataSafe("<Node.Name()>.ClaudeSessionID")` to retrieve the stored Claude session ID.
   - If the key does not exist (returns `(nil, false)`), returns an error: `"claude session ID not found for node '<Node.Name()>'"`.
   - If the stored value does not match the provided `claudeSessionID`, returns an error: `"claude session ID mismatch: expected <stored-value> but got <provided-value>"`.
   - If the stored value matches, returns nil (validation passed).
3. If the node's `Type() == "human"`:
   - If `claudeSessionID` is not an empty string `""`, returns an error: `"invalid claude session ID for human node: must be empty"`.
   - If `claudeSessionID` is an empty string, returns nil (validation passed).
4. For any other node type (should not occur with validated WorkflowDefinitions), returns an error: `"unsupported node type '<Type()>'"`.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | PersistentSession reference | Valid, constructed via NewPersistentSession | Yes |
| Node | Node reference | Valid node definition from WorkflowDefinition | Yes |
| ClaudeSessionID | string | Any string (from RuntimeMessage.ClaudeSessionID()) | Yes |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| Error | error | nil if validation passes; descriptive error message if validation fails |

### Error Cases

| Condition | Error Message |
|-----------|--------------|
| Agent node, key not found | `"claude session ID not found for node '<NodeName>'"` |
| Agent node, mismatch | `"claude session ID mismatch: expected <stored> but got <provided>"` |
| Human node, non-empty ID | `"invalid claude session ID for human node: must be empty"` |
| Unknown node type | `"unsupported node type '<Type>'"` |

## Invariants

1. **Read-Only**: Must not modify any session state. Only reads via `PersistentSession.GetSessionDataSafe`.
2. **Stateless**: No internal state between invocations. All information is passed as parameters.
3. **Pure Validation**: Returns nil on success, error on failure. No side effects.
4. **Node Type Dispatch**: Validation rules are determined solely by `Node.Type()`. Agent nodes require matching stored ID; human nodes require empty string.

## Edge Cases

- Condition: Agent node, `GetSessionDataSafe` returns `(nil, false)`.
  Expected: Returns error `"claude session ID not found for node '<NodeName>'"`.

- Condition: Agent node, stored value matches provided `claudeSessionID`.
  Expected: Returns nil.

- Condition: Agent node, stored value does not match provided `claudeSessionID`.
  Expected: Returns error with both expected and actual values in the message.

- Condition: Human node, `claudeSessionID` is empty string.
  Expected: Returns nil.

- Condition: Human node, `claudeSessionID` is non-empty.
  Expected: Returns error `"invalid claude session ID for human node: must be empty"`.

- Condition: Node type is neither "agent" nor "human" (should not happen with validated workflows).
  Expected: Returns error `"unsupported node type '<Type>'"`.

## Related

- [EventProcessor](./event_processor.md) — caller that uses this function before event recording
- [ErrorProcessor](./error_processor.md) — caller that uses this function before error recording
- [Session data](../entities/session/data.md) — provides `GetSessionDataSafe` method
- [Node](../components/node.md) — provides node type and name
