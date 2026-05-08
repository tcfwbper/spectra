# Test Specification: `validate_claude_session_id_test.go`

## Source File Under Test

`runtime/validate_claude_session_id.go`

## Test File

`runtime/validate_claude_session_id_test.go`

---

## `ValidateClaudeSessionID`

### Happy Path — ValidateClaudeSessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestValidateClaudeSessionID_AgentNode_Matches` | `unit` | Returns nil when stored Claude session ID matches the provided value. | Create mock PersistentSession; stub `GetSessionDataSafe("myAgent.ClaudeSessionID")` to return `("abc-123", true)`. Create mock Node; stub `Type()` to return `"agent"` and `Name()` to return `"myAgent"`. | `ValidateClaudeSessionID(session, node, "abc-123")` | Returns `nil` |
| `TestValidateClaudeSessionID_HumanNode_EmptyID` | `unit` | Returns nil when human node receives an empty claude session ID. | Create mock PersistentSession. Create mock Node; stub `Type()` to return `"human"`. | `ValidateClaudeSessionID(session, node, "")` | Returns `nil` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestValidateClaudeSessionID_AgentNode_KeyNotFound` | `unit` | Returns error when session data key does not exist for agent node. | Create mock PersistentSession; stub `GetSessionDataSafe("myAgent.ClaudeSessionID")` to return `(nil, false)`. Create mock Node; stub `Type()` to return `"agent"` and `Name()` to return `"myAgent"`. | `ValidateClaudeSessionID(session, node, "abc-123")` | Returns error with message `"claude session ID not found for node 'myAgent'"` |
| `TestValidateClaudeSessionID_AgentNode_Mismatch` | `unit` | Returns error when stored session ID does not match provided value. | Create mock PersistentSession; stub `GetSessionDataSafe("myAgent.ClaudeSessionID")` to return `("expected-id", true)`. Create mock Node; stub `Type()` to return `"agent"` and `Name()` to return `"myAgent"`. | `ValidateClaudeSessionID(session, node, "wrong-id")` | Returns error with message `"claude session ID mismatch: expected expected-id but got wrong-id"` |
| `TestValidateClaudeSessionID_HumanNode_NonEmptyID` | `unit` | Returns error when human node receives a non-empty claude session ID. | Create mock PersistentSession. Create mock Node; stub `Type()` to return `"human"`. | `ValidateClaudeSessionID(session, node, "some-id")` | Returns error with message `"invalid claude session ID for human node: must be empty"` |
| `TestValidateClaudeSessionID_UnsupportedNodeType` | `unit` | Returns error for an unrecognized node type. | Create mock PersistentSession. Create mock Node; stub `Type()` to return `"unknown"`. | `ValidateClaudeSessionID(session, node, "any")` | Returns error with message `"unsupported node type 'unknown'"` |
