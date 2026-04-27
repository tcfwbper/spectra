# Test Specification: `agent_error.go`

## Source File Under Test
`entities/agent_error.go`

## Test File
`entities/agent_error_test.go`

---

## `AgentError`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_ValidConstruction` | `unit` | Creates AgentError with all valid required fields. | | `AgentRole="architect"`, `Message="task failed"`, `Detail={"reason": "timeout"}`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns valid AgentError; all fields match input |
| `TestAgentError_EmptyAgentRole` | `unit` | Creates AgentError with empty AgentRole for human node. | | `AgentRole=""`, `Message="user cancelled"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="human_input"`, `OccurredAt=1714147200` | Returns valid AgentError; `AgentRole=""` |
| `TestAgentError_NullDetail` | `unit` | Creates AgentError with null Detail. | | `AgentRole="tester"`, `Message="validation failed"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="test"`, `OccurredAt=1714147200` | Returns valid AgentError; `Detail=null` |
| `TestAgentError_EmptyDetail` | `unit` | Creates AgentError with empty JSON object Detail. | | `AgentRole="architect"`, `Message="error"`, `Detail={}`, `SessionID=<valid-uuid>`, `FailingState="init"`, `OccurredAt=1714147200` | Returns valid AgentError; `Detail={}` |

### Validation Failures — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_EmptyMessage` | `unit` | Rejects AgentError with empty message. | | `AgentRole="architect"`, `Message=""`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns error; error message matches `/message.*non-empty/i` |
| `TestAgentError_WhitespaceOnlyMessage` | `unit` | Rejects AgentError with whitespace-only message. | | `AgentRole="architect"`, `Message="   "`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns error; error message matches `/message.*whitespace/i` |

### Validation Failures — Detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_InvalidDetailJSON` | `unit` | Rejects AgentError with malformed JSON in Detail. | | `AgentRole="architect"`, `Message="error"`, `Detail=[]byte("{invalid json}")`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns error; error message matches `/JSON.*parse|unmarshal/i` |

### Validation Failures — SessionID

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_NonExistentSession` | `unit` | Rejects AgentError with non-existent SessionID. | Session with given UUID does not exist | `AgentRole="architect"`, `Message="error"`, `Detail=null`, `SessionID=<non-existent-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns error; error message matches `/session.*not found/i` |
| `TestAgentError_FailedSessionID` | `unit` | Rejects AgentError for session with Status=failed. | Session exists with `Status="failed"` | `AgentRole="architect"`, `Message="error"`, `Detail=null`, `SessionID=<failed-session-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns error; error message matches `/session.*terminated/i`; warning logged |
| `TestAgentError_CompletedSessionID` | `unit` | Rejects AgentError for session with Status=completed. | Session exists with `Status="completed"` | `AgentRole="architect"`, `Message="error"`, `Detail=null`, `SessionID=<completed-session-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns error; error message matches `/session.*terminated/i`; warning logged |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_TransitionsSessionToFailed` | `unit` | Session Status transitions to failed when AgentError is raised. | Session exists with `Status="running"`, `CurrentState="review"` | AgentError with valid fields; `SessionID=<session-uuid>`, `FailingState="review"` | Session `Status="failed"`; `CurrentState` unchanged; `Error` field set to AgentError instance |
| `TestAgentError_InitializingSessionFails` | `unit` | Session in initializing status transitions to failed. | Session exists with `Status="initializing"`, `CurrentState="entry"` | AgentError with `SessionID=<session-uuid>`, `FailingState="entry"` | Session `Status="failed"`; `CurrentState="entry"`; `FailingState="entry"` |
| `TestAgentError_RunningSessionFails` | `unit` | Session in running status transitions to failed. | Session exists with `Status="running"`, `CurrentState="processing"` | AgentError with `SessionID=<session-uuid>`, `FailingState="processing"` | Session `Status="failed"`; `CurrentState="processing"`; `FailingState="processing"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_FieldsImmutable` | `unit` | AgentError fields cannot be modified after creation. | AgentError instance created | Attempt to modify `Message`, `Detail`, or other fields | Field modification attempt fails or has no effect; original values remain |

### Happy Path — ErrorProcessor Derivation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_AgentRoleDerivedFromNode` | `unit` | AgentRole is derived from current node's agent_role by ErrorProcessor. | Session at agent node with `agent_role="architect"` | Error raised without AgentRole in wire payload | AgentError created with `AgentRole="architect"` derived from node definition |
| `TestAgentError_EmptyAgentRoleForHumanNode` | `unit` | AgentRole is empty string for human nodes. | Session at human node (no agent_role defined) | Error raised from human node | AgentError created with `AgentRole=""` |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_MultipleErrorsSerialized` | `race` | Multiple simultaneous errors are serialized; first error wins. | Session with `Status="running"` | Two AgentError instances raised simultaneously for same session | First error to acquire session lock recorded in session's `Error` field; second error logged but does not overwrite; session `Status="failed"` |

### Happy Path — Persistence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_PersistedToDisk` | `unit` | AgentError is persisted to session error log. | Temporary test directory created; session files placed within this directory; all file operations occur within test fixtures | Valid AgentError raised | Error details written to error log file within test directory; session metadata updated on disk |
| `TestAgentError_SessionDeletion` | `unit` | AgentError removed when session is deleted. | Temporary test directory created; session exists with recorded AgentError in test directory | Delete session | AgentError file removed from filesystem; subsequent error queries return "session not found" |

### Invariants — FailingState Consistency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_FailingStateMatchesCurrentState` | `unit` | FailingState must match CurrentState at error time. | Session with `CurrentState="review"` | AgentError with `FailingState="review"` | AgentError recorded; `FailingState="review"` matches session `CurrentState` |
| `TestAgentError_CurrentStateUnchanged` | `unit` | CurrentState does not change when error occurs. | Session with `CurrentState="processing"` | AgentError raised | Session `CurrentState` remains "processing"; `FailingState="processing"` |

### Happy Path — Terminal Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_StatusPermanentlyFailed` | `unit` | Session Status remains failed and cannot transition. | Session transitioned to `Status="failed"` by AgentError | Attempt any status transition | Status remains "failed"; transition rejected |
| `TestAgentError_NoAutomaticRetry` | `unit` | Runtime does not automatically retry failed session. | Session with `Status="failed"` due to AgentError; mock clock advanced by 5 seconds | Query session status after time advancement | Session remains failed; no retry attempted |
| `TestAgentError_ManualRecoveryRejected` | `unit` | Recovery requests for failed session are rejected. | Session with `Status="failed"` | Request session recovery | Returns error; error message matches `/recovery not supported|cannot recover/i` and mentions creating new session |

### Happy Path — Sensitive Data

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_SensitiveDataPersisted` | `unit` | Detail with sensitive info is persisted as-is (agent responsibility to sanitize). | | AgentError with `Detail={"api_key": "secret123"}` | Detail persisted exactly as provided; no sanitization by runtime |

### Boundary Values — Message

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_LargeMessage` | `unit` | Accepts AgentError with very large message string. | | `AgentRole="architect"`, `Message=<1MB string>`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns valid AgentError; message stored correctly |
| `TestAgentError_UnicodeMessage` | `unit` | Accepts AgentError with Unicode characters in message. | | `AgentRole="architect"`, `Message="Error: 任务失败 🔥"`, `Detail=null`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns valid AgentError; Unicode preserved correctly |

### Boundary Values — Detail

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_LargeDetail` | `unit` | Accepts AgentError with very large Detail JSON object. | | `AgentRole="architect"`, `Message="error"`, `Detail=<10MB JSON object>`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns valid AgentError; Detail stored correctly |
| `TestAgentError_DeepNestedDetail` | `unit` | Accepts AgentError with deeply nested JSON in Detail. | | `AgentRole="architect"`, `Message="error"`, `Detail=<JSON nested 100 levels deep>`, `SessionID=<valid-uuid>`, `FailingState="review"`, `OccurredAt=1714147200` | Returns valid AgentError; nested structure preserved |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentError_CLIInvocation` | `e2e` | Agent raises error via spectra-agent CLI. | Temporary test directory created; runtime running with session files in test directory; session active with socket in test directory; all file operations occur within test fixtures | Execute `spectra-agent error "task failed" --session-id <uuid> --detail '{"code": 500}'` | Command succeeds; AgentError recorded in test directory; session transitions to failed; human notified |
| `TestAgentError_CLIMissingSessionID` | `e2e` | CLI rejects error without session-id flag. | Temporary test directory created; runtime running; all file operations occur within test fixtures | Execute `spectra-agent error "task failed"` | Command fails; error message matches `/session-id.*required/i` |
