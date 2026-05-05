# Test Specification: `agent_error_test.go`

## Source File Under Test
`entities/agent_error.go`

## Test File
`entities/agent_error_test.go`

---

## `AgentError`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentError_ValidInputs` | `unit` | Constructs an AgentError with all valid fields including a non-empty AgentRole. | | `agentRole="Reviewer"`, `message="agent failed"`, `detail=json.RawMessage('{"reason":"timeout"}')`, `occurredAt=1700000000`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Review"` | Returns no error; all getters return the provided values |
| `TestNewAgentError_EmptyAgentRole` | `unit` | Constructs an AgentError with empty AgentRole representing a human node. | | `agentRole=""`, `message="human error"`, `detail=nil`, `occurredAt=1700000000`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Approval"` | Returns no error; AgentRole getter returns `""` |
| `TestNewAgentError_SpecialCharsAgentRole` | `unit` | Constructs an AgentError with special characters in AgentRole. | | `agentRole="my agent/v2 (test)"`, `message="err"`, `detail=nil`, `occurredAt=1`, `sessionID="550e8400-e29b-41d4-a716-446655440000"`, `failingState="Init"` | Returns no error; AgentRole getter returns `"my agent/v2 (test)"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentError_PropagatesMessageError` | `unit` | Propagates validation error when Message is invalid. | | `agentRole="Agent"`, `message=""`, other fields valid | Returns validation error from SessionError |
| `TestNewAgentError_PropagatesSessionIDError` | `unit` | Propagates validation error when SessionID is invalid. | | `agentRole="Agent"`, `sessionID="bad"`, other fields valid | Returns validation error from SessionError |
| `TestNewAgentError_PropagatesOccurredAtError` | `unit` | Propagates validation error when OccurredAt is invalid. | | `agentRole="Agent"`, `occurredAt=0`, other fields valid | Returns validation error from SessionError |
| `TestNewAgentError_PropagatesDetailError` | `unit` | Propagates validation error when Detail is a JSON array. | | `agentRole="Agent"`, `detail=json.RawMessage('[]')`, other fields valid | Returns validation error from SessionError |
| `TestNewAgentError_PropagatesFailingStateError` | `unit` | Propagates validation error when FailingState is empty. | | `agentRole="Agent"`, `failingState=""`, other fields valid | Returns validation error from SessionError |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_Immutability` | `unit` | All fields including AgentRole remain unchanged after construction. | Construct a valid AgentError | Verify all getter values after construction | All getter values remain identical to construction inputs |
