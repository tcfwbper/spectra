# Test Specification: `error_processor_test.go`

## Source File Under Test

`runtime/error_processor.go`

## Test File

`runtime/error_processor_test.go`

---

## `ErrorProcessor`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewErrorProcessor_ValidDeps` | `unit` | Constructs ErrorProcessor with all valid dependencies. | Create mock PersistentSession, mock WorkflowDefinition, and a buffered TerminationNotifier channel (cap >= 2). | `NewErrorProcessor(persistentSession, workflowDef, terminationNotifier)` | Returns non-nil `*ErrorProcessor`; no panic |

### Happy Path — ProcessError

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorProcessor_ProcessError_RunningStatus` | `unit` | Records error successfully when session status is "running". | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`, `Fail()` returns nil, `ID` returns `"sess-uuid"`. Mock WorkflowDefinition: `Nodes()` returns a node named `"nodeA"` with `Type()="agent"` and `AgentRole()="coder"`. Stub `ValidateClaudeSessionID` returns nil. Create RuntimeMessage with `Type()="error"`, `ClaudeSessionID()="cs-123"`, `Payload()={"message":"something failed","detail":"stack trace"}`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `SuccessResponse("error recorded \| session=sess-uuid \| failingState=nodeA \| agentRole=coder \| error=something failed")` |
| `TestErrorProcessor_ProcessError_InitializingStatus` | `unit` | Records error successfully when session status is "initializing". | Mock PersistentSession: `GetStatusSafe()` returns `"initializing"`, `GetCurrentStateSafe()` returns `"startNode"`, `Fail()` returns nil, `ID` returns `"sess-uuid"`. Mock WorkflowDefinition: `Nodes()` returns a node named `"startNode"` with `Type()="agent"` and `AgentRole()="init"`. Stub `ValidateClaudeSessionID` returns nil. Create RuntimeMessage with valid error payload `{"message":"init error"}`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `SuccessResponse(...)` containing `"error recorded"` |
| `TestErrorProcessor_ProcessError_HumanNode` | `unit` | Derives empty agentRole when current node is a human node. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"humanNode"`, `Fail()` returns nil, `ID` returns `"sess-uuid"`. Mock WorkflowDefinition: `Nodes()` returns a node named `"humanNode"` with `Type()="human"`. Stub `ValidateClaudeSessionID` returns nil. Create RuntimeMessage with valid error payload `{"message":"user error"}`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `SuccessResponse(...)` containing `"agentRole="` (empty agentRole); `Fail()` called with AgentError having `AgentRole=""` |
| `TestErrorProcessor_ProcessError_DetailNilOrMissing` | `unit` | Constructs AgentError with nil Detail when detail field is absent. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`, `Fail()` returns nil, `ID` returns `"sess-uuid"`. Mock WorkflowDefinition with matching node. Stub `ValidateClaudeSessionID` returns nil. Create RuntimeMessage with payload `{"message":"oops"}` (no detail field). | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns success; `Fail()` called with AgentError having `Detail == nil` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorProcessor_ProcessError_CallsValidateClaudeSessionID` | `unit` | Invokes ValidateClaudeSessionID with correct arguments before recording error. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`. Mock WorkflowDefinition with matching node. Stub `ValidateClaudeSessionID` to capture arguments and return nil. Mock `Fail()` returns nil. Create RuntimeMessage with `ClaudeSessionID()="cs-456"`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | `ValidateClaudeSessionID` called once with (persistentSession, currentNode, `"cs-456"`) |
| `TestErrorProcessor_ProcessError_FailCalledWithCorrectAgentError` | `unit` | Calls PersistentSession.Fail with correctly constructed AgentError and terminationNotifier. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeB"`, `ID` returns `"sess-uuid"`. Mock WorkflowDefinition: node `"nodeB"` with `Type()="agent"`, `AgentRole()="reviewer"`. Stub `ValidateClaudeSessionID` returns nil. Mock `Fail()` captures args, returns nil. Create RuntimeMessage with payload `{"message":"fail msg","detail":"some detail"}`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | `Fail()` called once with AgentError having `AgentRole="reviewer"`, `Message="fail msg"`, `Detail="some detail"`, `SessionID="sess-uuid"`, `FailingState="nodeB"`, `OccurredAt` set to a POSIX timestamp; second arg is terminationNotifier |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorProcessor_ProcessError_SessionCompleted` | `unit` | Returns error response when session status is "completed". | Mock PersistentSession: `GetStatusSafe()` returns `"completed"`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("session terminated: status is 'completed'")` |
| `TestErrorProcessor_ProcessError_SessionFailed` | `unit` | Returns error response when session status is "failed". | Mock PersistentSession: `GetStatusSafe()` returns `"failed"`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("session terminated: status is 'failed'")` |
| `TestErrorProcessor_ProcessError_ClaudeSessionIDValidationFails` | `unit` | Returns error response when ValidateClaudeSessionID fails. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`. Mock WorkflowDefinition with matching node. Stub `ValidateClaudeSessionID` returns `errors.New("session ID mismatch")`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("session ID mismatch")`; `Fail()` not called |
| `TestErrorProcessor_ProcessError_InvalidPayloadMissingMessage` | `unit` | Returns error response when payload is missing required 'message' field. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`. Mock WorkflowDefinition with matching node. Stub `ValidateClaudeSessionID` returns nil. Create RuntimeMessage with payload `{}` (no message field). | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("invalid error payload: missing required field 'message'")`; `Fail()` not called |
| `TestErrorProcessor_ProcessError_FailReturnsError` | `unit` | Returns error response when PersistentSession.Fail returns error. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`, `Fail()` returns `errors.New("session already failed")`, `ID` returns `"sess-uuid"`. Mock WorkflowDefinition with matching node. Stub `ValidateClaudeSessionID` returns nil. Create RuntimeMessage with valid error payload. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("failed to record error: session already failed")` |
| `TestErrorProcessor_ProcessError_NodeNotFound` | `unit` | Returns error response when current node is not in workflow definition. | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"unknownNode"`. Mock WorkflowDefinition: `Nodes()` returns no node matching `"unknownNode"`. | `ep.ProcessError("sess-uuid", runtimeMessage)` | Returns `ErrorResponse("current node 'unknownNode' not found in workflow definition")` |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorProcessor_ProcessError_ConcurrentFirstErrorWins` | `race` | First concurrent error call succeeds; subsequent calls get "session already failed". | Mock PersistentSession: `GetStatusSafe()` returns `"running"`, `GetCurrentStateSafe()` returns `"nodeA"`. Mock WorkflowDefinition with matching node. Stub `ValidateClaudeSessionID` returns nil. Mock `Fail()` succeeds on first call and returns `errors.New("session already failed")` on second call. Create two valid RuntimeMessages. | Call `ep.ProcessError(...)` concurrently from two goroutines. | One returns success response; the other returns `ErrorResponse("failed to record error: session already failed")`. No data race. |
