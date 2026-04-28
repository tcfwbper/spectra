# Test Specification: `error_processor.go`

## Source File Under Test
`runtime/error_processor.go`

## Test File
`runtime/error_processor_test.go`

---

## `ErrorProcessor`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestErrorProcessor_New` | `unit` | Constructs ErrorProcessor with valid inputs. | Test fixture; mock Session, WorkflowDefinitionLoader, TerminationNotifier channel | `Session=<mock>`, `WorkflowDefinitionLoader=<mock>`, `TerminationNotifier=<channel>` | Returns ErrorProcessor instance; no error |

### Happy Path — ProcessError (Agent Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_AgentNode_ValidClaudeSessionID` | `unit` | Processes error from agent node with matching Claude session ID. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock workflow with agent node having `agent_role="reviewer"`; SessionData contains `agent_node.ClaudeSessionID=<uuid>` | RuntimeMessage with `type="error"`, `claudeSessionID=<uuid>`, `payload={message:"test error", detail:{}}` | Returns RuntimeResponse with `status="success"`, message matching `/error recorded.*session=.*failingState=agent_node.*agentRole=reviewer/i`; Session.Fail called with AgentError |
| `TestProcessError_InitializingStatus` | `unit` | Accepts error when session status is initializing. | Mock Session with `Status="initializing"`, `CurrentState="entry_node"`; mock workflow with agent node; SessionData contains Claude session ID | RuntimeMessage with `type="error"`, matching `claudeSessionID`, `payload={message:"init error", detail:{}}` | Returns RuntimeResponse with `status="success"`; Session transitions to "failed"; FailingState set to "entry_node" |
| `TestProcessError_DetailFieldPresent` | `unit` | Processes error with detail field. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `payload={message:"error", detail:{code:500, context:"test"}}` | Returns RuntimeResponse with `status="success"`; AgentError contains detail with code and context |
| `TestProcessError_DetailFieldNull` | `unit` | Handles null detail field as empty object. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `payload={message:"error", detail:null}` | Returns RuntimeResponse with `status="success"`; AgentError detail is empty object `{}` |

### Happy Path — ProcessError (Human Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_HumanNode_EmptyClaudeSessionID` | `unit` | Processes error from human node with empty Claude session ID. | Mock Session with `Status="running"`, `CurrentState="human_node"`; mock workflow with human node (no agent_role field) | RuntimeMessage with `type="error"`, `claudeSessionID=""`, `payload={message:"human error", detail:{}}` | Returns RuntimeResponse with `status="success"`, message matching `/agentRole=""/i` or containing empty agentRole; AgentError has `AgentRole=""` |

### Happy Path — TerminationNotifier

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_TerminationNotifierSignaled` | `unit` | Signals termination notifier after successful error recording. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; TerminationNotifier channel monitored | Valid error RuntimeMessage | Session.Fail called; TerminationNotifier receives signal; returns RuntimeResponse with `status="success"` |

### Validation Failures — Session Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_StatusCompleted` | `unit` | Rejects error when session status is completed. | Mock Session with `Status="completed"` | Valid error RuntimeMessage | Returns RuntimeResponse with `status="error"`, `message="session terminated: status is 'completed'"` |
| `TestProcessError_StatusFailed` | `unit` | Rejects error when session status is failed. | Mock Session with `Status="failed"` | Valid error RuntimeMessage | Returns RuntimeResponse with `status="error"`, `message="session terminated: status is 'failed'"` |

### Validation Failures — Claude Session ID (Agent Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_AgentNode_ClaudeSessionIDNotFound` | `unit` | Rejects error when Claude session ID not found in SessionData. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock workflow with agent node; SessionData does NOT contain `agent_node.ClaudeSessionID` | RuntimeMessage with `type="error"`, `claudeSessionID=<uuid>` | Returns RuntimeResponse with `status="error"`, `message="claude session ID not found for node 'agent_node'"` |
| `TestProcessError_AgentNode_ClaudeSessionIDMismatch` | `unit` | Rejects error when Claude session ID does not match. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; SessionData contains `agent_node.ClaudeSessionID=<uuid-1>` | RuntimeMessage with `claudeSessionID=<uuid-2>` (different UUID) | Returns RuntimeResponse with `status="error"`, message matching `/claude session ID mismatch: expected <uuid-1> but got <uuid-2>/i` |

### Validation Failures — Claude Session ID (Human Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_HumanNode_NonEmptyClaudeSessionID` | `unit` | Rejects error from human node with non-empty Claude session ID. | Mock Session with `Status="running"`, `CurrentState="human_node"`; mock workflow with human node | RuntimeMessage with `type="error"`, `claudeSessionID=<uuid>` (non-empty) | Returns RuntimeResponse with `status="error"`, `message="invalid claude session ID for human node: must be empty"` |

### Validation Failures — Workflow Definition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_WorkflowDefinitionNotFound` | `unit` | Returns error when workflow definition cannot be loaded. | Mock Session with `Status="running"`; mock WorkflowDefinitionLoader programmatically returns error simulating file not found (no actual file I/O) | Valid error RuntimeMessage | Returns RuntimeResponse with `status="error"`, message matching `/failed to load workflow definition:/i` |
| `TestProcessError_WorkflowDefinitionParseError` | `unit` | Returns error when workflow definition has parse error. | Mock Session with `Status="running"`; mock WorkflowDefinitionLoader programmatically returns parse error (no actual file I/O) | Valid error RuntimeMessage | Returns RuntimeResponse with `status="error"`, message matching `/failed to load workflow definition:/i` |

### Validation Failures — Message Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_MissingMessageField` | `unit` | Returns error when message field is missing from payload. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `type="error"`, `payload={detail:{}}` (message missing) | Returns RuntimeResponse with `status="error"`, message matching `/invalid error payload: missing required field/i` |

### Error Propagation — Session.Fail Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_SessionFailReturnsError` | `unit` | Returns error when Session.Fail returns error. | Mock Session with `Status="running"`; Session.Fail returns error "session already failed" | Valid error RuntimeMessage | Returns RuntimeResponse with `status="error"`, message matching `/failed to record error:.*session already failed/i` |

### Error Propagation — Persistence Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_PersistenceFailureBestEffort` | `unit` | Continues when SessionMetadataStore write fails during Session.Fail. | Mock Session with `Status="running"`; SessionMetadataStore.Write returns error (disk full); Session.Fail logs warning but returns nil | Valid error RuntimeMessage | Returns RuntimeResponse with `status="success"`; session remains in "failed" status in memory; warning logged about persistence failure |

### Boundary Values — Error Message Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_VeryLargeErrorMessage` | `unit` | Handles very large error message (5 MB). | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `payload={message:<5MB-string>, detail:{}}` | Returns RuntimeResponse with `status="success"`; AgentError stored with complete message |
| `TestProcessError_VeryLargeDetail` | `unit` | Handles very large detail structure (5 MB stack trace). | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `payload={message:"error", detail:{stack:<5MB-string>}}` | Returns RuntimeResponse with `status="success"`; AgentError stored with complete detail |
| `TestProcessError_UnicodeInMessage` | `unit` | Handles Unicode characters in error message. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `payload={message:"错误: emoji 🚨", detail:{}}` | Returns RuntimeResponse with `status="success"`; AgentError preserves Unicode |

### Boundary Values — Field Values

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_EmptyAgentRole` | `unit` | Handles human node with empty agent role correctly. | Mock Session with `Status="running"`, `CurrentState="human_node"`; mock workflow with human node | RuntimeMessage with `claudeSessionID=""`, `payload={message:"error", detail:{}}` | Returns RuntimeResponse with `status="success"`; AgentError has `AgentRole=""` |
| `TestProcessError_DetailFieldMissing` | `unit` | Handles missing detail field as empty object. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `payload={message:"error"}` (detail missing) | Returns RuntimeResponse with `status="success"`; AgentError detail is empty object `{}` |

### Idempotency — First Error Wins

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_FirstErrorWins` | `unit` | Second error rejected when session already failed. | Mock Session with `Status="running"`; first call transitions to "failed"; Session.Fail on second call returns error | Two sequential error RuntimeMessages | First returns `status="success"`; second returns `status="error"`, message matching `/failed to record error:.*session already failed/i` |

### Concurrent Behaviour — Multiple Simultaneous Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_ConcurrentErrors` | `race` | Handles concurrent errors from multiple agents safely. | Mock Session with `Status="running"`; Session.Fail uses write lock; 5 goroutines call ProcessError simultaneously | 5 concurrent error RuntimeMessages | One succeeds (first to acquire lock); other 4 return error `"session already failed"`; no data races detected |
| `TestProcessError_ConcurrentErrorAndEvent` | `race` | Verifies ErrorProcessor's reliance on Session's thread-safety when concurrent with EventProcessor. | Mock Session with `Status="running"`; Session uses internal lock for thread-safe state access; ErrorProcessor and EventProcessor (minimal mock) run concurrently sharing the same Session | Error processed in one goroutine; event in another goroutine | Session's internal locking prevents data races; if error wins, event sees `Status="failed"` and rejects; if event wins, error succeeds; verifies ErrorProcessor correctly uses Session's thread-safe methods |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_SessionGetSessionDataSafeCalled` | `unit` | Verifies Session.GetSessionDataSafe called for agent node. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock tracks GetSessionDataSafe calls | RuntimeMessage with `claudeSessionID=<uuid>` | Session.GetSessionDataSafe called with key `"agent_node.ClaudeSessionID"`; returns stored value |
| `TestProcessError_SessionFailCalledWithAgentError` | `unit` | Verifies Session.Fail called with constructed AgentError. | Mock Session with `Status="running"`; mock tracks Session.Fail arguments | Valid error RuntimeMessage | Session.Fail called once with AgentError containing: correct agentRole, message, detail, SessionID, FailingState, and OccurredAt timestamp |
| `TestProcessError_WorkflowDefinitionLoaderCalled` | `unit` | Verifies WorkflowDefinitionLoader invoked to load workflow. | Mock Session with `Status="running"`; mock WorkflowDefinitionLoader tracks calls | Valid error RuntimeMessage | WorkflowDefinitionLoader called with session's workflow name |
| `TestProcessError_AgentRoleDerivedFromNode` | `unit` | Verifies agentRole derived from node definition, not message. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock workflow with agent node having `agent_role="tester"`; RuntimeMessage does NOT contain agentRole field | Valid error RuntimeMessage | AgentError constructed with `AgentRole="tester"` from node definition |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_CurrentStateNotChanged` | `unit` | Verifies CurrentState remains unchanged after error. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock tracks CurrentState | Valid error RuntimeMessage | Session.CurrentState remains "agent_node" after error processing; FailingState in AgentError is "agent_node" |
| `TestProcessError_StatusTransitionToFailed` | `unit` | Verifies Session.Status transitions to failed. | Mock Session with `Status="running"`; Session.Fail updates Status to "failed" | Valid error RuntimeMessage | Session.Status is "failed" after Session.Fail called |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessError_NoCleanupPerformed` | `unit` | Verifies ErrorProcessor does not perform cleanup operations. | Mock Session with `Status="running"`; monitor for socket deletion, stdout printing, etc. | Valid error RuntimeMessage | ErrorProcessor does NOT delete socket or print to stdout; only calls Session.Fail and returns response |
