# Test Specification: `event_processor.go`

## Source File Under Test
`runtime/event_processor.go`

## Test File
`runtime/event_processor_test.go`

---

## `EventProcessor`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEventProcessor_New` | `unit` | Constructs EventProcessor with valid inputs. | Test fixture; mock Session, WorkflowDefinitionLoader, TransitionToNode, TerminationNotifier channel | `Session=<mock>`, `WorkflowDefinitionLoader=<mock>`, `TransitionToNode=<mock>`, `TerminationNotifier=<channel>` | Returns EventProcessor instance; no error |

### Happy Path — ProcessEvent (Agent Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_AgentNode_ValidClaudeSessionID` | `unit` | Processes event from agent node with matching Claude session ID. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock workflow with agent node; SessionData contains `agent_node.ClaudeSessionID=<uuid>`; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `type="event"`, `claudeSessionID=<uuid>`, `payload={eventType:"approved", message:"ok", payload:{}}` | Event recorded to EventStore; transition executed; returns RuntimeResponse with `status="success"`, message matching `/event 'approved' processed successfully.*session=.*currentState=.*sessionStatus=/i` |
| `TestProcessEvent_ComplexPayload` | `unit` | Handles event with complex nested payload. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `payload={eventType:"test", message:"msg", payload:{nested:{array:[1,2,3], bool:true, null:null}}}` | Event recorded with complete payload structure preserved; transition executed; returns success |
| `TestProcessEvent_OptionalFieldsPresent` | `unit` | Handles event with all optional fields. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with all optional fields including `claudeSessionID` and nested payload | Event recorded with all fields; transition executed; returns success |

### Happy Path — ProcessEvent (Human Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_HumanNode_EmptyClaudeSessionID` | `unit` | Processes event from human node with empty Claude session ID. | Mock Session with `Status="running"`, `CurrentState="human_node"`; mock workflow with human node; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `type="event"`, `claudeSessionID=""`, `payload={eventType:"continue", message:"proceed", payload:{}}` | Event recorded; transition executed; returns RuntimeResponse with `status="success"` |

### Happy Path — Event Recording

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_EventRecordedToStore` | `unit` | Records event to EventStore via Session.UpdateEventHistorySafe. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; mock EventStore tracks writes; TransitionEvaluator returns valid transition; TransitionToNode succeeds | Valid event RuntimeMessage | Session.UpdateEventHistorySafe called with Event containing: generated UUID, eventType from payload, message, payload, EmittedBy=CurrentState, EmittedAt=timestamp, SessionID |
| `TestProcessEvent_EventEmittedByAutoAssigned` | `unit` | Automatically assigns EmittedBy to Session.CurrentState. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; agent node; SessionData contains Claude session ID; mock EventStore tracks Event fields; TransitionEvaluator returns valid transition; TransitionToNode succeeds | Valid event RuntimeMessage (EmittedBy not in message) | Event has `EmittedBy="agent_node"` set automatically from Session.CurrentState |
| `TestProcessEvent_EventUUIDGenerated` | `unit` | Generates unique UUID for each event. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; TransitionEvaluator returns valid transition; TransitionToNode succeeds | Two sequential event RuntimeMessages | Two events recorded with different UUIDs; both valid UUID v4 format |

### Happy Path — Transition Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_TransitionToNodeCalled` | `unit` | Invokes TransitionToNode with correct parameters. | Mock Session with `Status="running"`, `CurrentState="node_a"`; TransitionEvaluator returns transition to "node_b" with `isExitTransition=false`; mock TransitionToNode tracks calls | Valid event RuntimeMessage | TransitionToNode.Transition called with `Message=<event-message>`, `TargetNodeName="node_b"`, `IsExitTransition=false`; returns success |
| `TestProcessEvent_ExitTransition` | `unit` | Handles exit transition correctly. | Mock Session with `Status="running"`; TransitionEvaluator returns transition with `isExitTransition=true`; TransitionToNode succeeds and sets Session.Status to "completed" | Valid event RuntimeMessage | TransitionToNode called with `IsExitTransition=true`; returns RuntimeResponse with message matching `/sessionStatus=completed/i` |
| `TestProcessEvent_CurrentStateUpdated` | `unit` | Verifies CurrentState updated after transition. | Mock Session with `Status="running"`, `CurrentState="node_a"`; TransitionEvaluator returns transition to "node_b"; TransitionToNode updates Session.CurrentState to "node_b" | Valid event RuntimeMessage | Session.CurrentState is "node_b" after transition; response message contains `currentState=node_b` |

### Validation Failures — Session Status

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_StatusInitializing` | `unit` | Rejects event when session status is initializing. | Mock Session with `Status="initializing"` | Valid event RuntimeMessage | Returns RuntimeResponse with `status="error"`, `message="session not ready: status is 'initializing'"` |
| `TestProcessEvent_StatusCompleted` | `unit` | Rejects event when session status is completed. | Mock Session with `Status="completed"` | Valid event RuntimeMessage | Returns RuntimeResponse with `status="error"`, `message="session not ready: status is 'completed'"` |
| `TestProcessEvent_StatusFailed` | `unit` | Rejects event when session status is failed. | Mock Session with `Status="failed"` | Valid event RuntimeMessage | Returns RuntimeResponse with `status="error"`, `message="session not ready: status is 'failed'"` |

### Validation Failures — Claude Session ID (Agent Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_AgentNode_ClaudeSessionIDNotFound` | `unit` | Rejects event when Claude session ID not found in SessionData. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock workflow with agent node; SessionData does NOT contain `agent_node.ClaudeSessionID` | RuntimeMessage with `type="event"`, `claudeSessionID=<uuid>` | Returns RuntimeResponse with `status="error"`, `message="claude session ID not found for node 'agent_node'"`; no event recorded |
| `TestProcessEvent_AgentNode_ClaudeSessionIDMismatch` | `unit` | Rejects event when Claude session ID does not match. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; SessionData contains `agent_node.ClaudeSessionID=<uuid-1>` | RuntimeMessage with `claudeSessionID=<uuid-2>` (different UUID) | Returns RuntimeResponse with `status="error"`, message matching `/claude session ID mismatch: expected <uuid-1> but got <uuid-2>/i`; no event recorded |

### Validation Failures — Claude Session ID (Human Node)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_HumanNode_NonEmptyClaudeSessionID` | `unit` | Rejects event from human node with non-empty Claude session ID. | Mock Session with `Status="running"`, `CurrentState="human_node"`; mock workflow with human node | RuntimeMessage with `type="event"`, `claudeSessionID=<uuid>` (non-empty) | Returns RuntimeResponse with `status="error"`, `message="invalid claude session ID for human node: must be empty"`; no event recorded |

### Validation Failures — Workflow Definition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_WorkflowDefinitionNotFound` | `unit` | Returns error when workflow definition cannot be loaded. | Mock Session with `Status="running"`; mock WorkflowDefinitionLoader programmatically returns error simulating file not found (no actual file I/O) | Valid event RuntimeMessage | Returns RuntimeResponse with `status="error"`, message matching `/failed to load workflow definition:/i` |
| `TestProcessEvent_WorkflowDefinitionParseError` | `unit` | Returns error when workflow definition has parse error. | Mock Session with `Status="running"`; mock WorkflowDefinitionLoader programmatically returns parse error (no actual file I/O) | Valid event RuntimeMessage | Returns RuntimeResponse with `status="error"`, message matching `/failed to load workflow definition:/i` |

### Validation Failures — Message Payload

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_MissingEventTypeField` | `unit` | Returns error when eventType field is missing from payload. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID | RuntimeMessage with `type="event"`, `payload={message:"test"}` (eventType missing) | Returns RuntimeResponse with `status="error"`, message matching `/invalid event payload: missing eventType/i` |

### Error Propagation — Event Recording Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_EventStoreWriteFailure` | `unit` | Triggers session failure when EventStore write fails. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe returns error (disk full) | Valid event RuntimeMessage | Session.Fail called with RuntimeError `Issuer="EventProcessor"`, `Message="failed to record event"`; returns RuntimeResponse with `status="error"`, message matching `/failed to record event:/i`; transition NOT attempted |

### Error Propagation — No Matching Transition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_NoMatchingTransition` | `unit` | Returns error when no transition found for event. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns `(nil, false, nil)` | RuntimeMessage with `eventType="unknown"` | Event already recorded; returns RuntimeResponse with `status="error"`, message matching `/no transition found for event 'unknown' from node '<CurrentState>'/i`; session remains in "running" status |
| `TestProcessEvent_SessionRemainsRunningAfterNoTransition` | `unit` | Session remains in running status when no transition found. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns no transition | Valid event RuntimeMessage | Event recorded; returns error response; Session.Status remains "running"; Session.Fail NOT called |

### Error Propagation — Transition Execution Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_TransitionToNodeFails` | `unit` | Returns error when TransitionToNode fails. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode returns error (already handled Session.Fail internally) | Valid event RuntimeMessage | Event recorded; returns RuntimeResponse with `status="error"`, message matching `/transition failed:/i`; EventProcessor does NOT call Session.Fail again |
| `TestProcessEvent_TransitionToNodeSessionFailCalledInternally` | `unit` | Verifies EventProcessor does not call Session.Fail when TransitionToNode fails. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode fails and calls Session.Fail internally; mock tracks Session.Fail calls | Valid event RuntimeMessage | Session.Fail called exactly once (by TransitionToNode, not by EventProcessor); EventProcessor returns error response |

### Boundary Values — Event Message Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_VeryLargeEventMessage` | `unit` | Handles very large event message (5 MB). | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `payload={eventType:"test", message:<5MB-string>, payload:{}}` | Event recorded with complete message; transition executed; returns success |
| `TestProcessEvent_VeryLargePayload` | `unit` | Handles very large payload structure (5 MB). | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `payload={eventType:"test", message:"msg", payload:{data:<5MB-structure>}}` | Event recorded with complete payload; transition executed; returns success |
| `TestProcessEvent_UnicodeInEvent` | `unit` | Handles Unicode characters in event fields. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `payload={eventType:"测试", message:"emoji 🎉", payload:{}}` | Event recorded with Unicode preserved; transition executed; returns success |

### Boundary Values — Field Values

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_MinimalEventPayload` | `unit` | Handles event with only required fields. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `payload={eventType:"test"}` (minimal) | Event recorded; transition executed; returns success |
| `TestProcessEvent_EmptyNestedPayload` | `unit` | Handles empty nested payload object. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; Session.UpdateEventHistorySafe succeeds; TransitionEvaluator returns valid transition; TransitionToNode succeeds | RuntimeMessage with `payload={eventType:"test", message:"msg", payload:{}}` | Event recorded with empty nested payload; transition executed; returns success |

### Concurrent Behaviour — Multiple Simultaneous Events

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_ConcurrentEvents` | `race` | Handles concurrent events from multiple agents safely. | Mock Session with `Status="running"`; Session.UpdateEventHistorySafe uses write lock; 5 goroutines call ProcessEvent simultaneously | 5 concurrent event RuntimeMessages | All 5 events recorded to EventHistory (serialized by lock); transitions executed; no data races detected |
| `TestProcessEvent_EventRecordingSerialized` | `race` | Event recording serialized via session-level write lock. | Mock Session with `Status="running"`; 3 goroutines call ProcessEvent; monitor lock acquisition order | 3 concurrent event RuntimeMessages | Events recorded in order of lock acquisition; Session.UpdateEventHistorySafe calls serialized; all complete successfully |

### Concurrent Behaviour — Concurrent Event and Error

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_ConcurrentEventAndError` | `race` | Verifies EventProcessor's reliance on Session's thread-safety when concurrent with ErrorProcessor. | Mock Session with `Status="running"`; Session uses internal lock for thread-safe state access; EventProcessor and ErrorProcessor (minimal mock) run concurrently sharing the same Session | Event processed in one goroutine; error in another goroutine | Session's internal locking prevents data races; if error wins, Session.Status becomes "failed" and event validation fails; if event wins, error succeeds after event; verifies EventProcessor correctly uses Session's thread-safe methods |

### Mock / Dependency Interaction — Session Methods

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_SessionGetSessionDataSafeCalled` | `unit` | Verifies Session.GetSessionDataSafe called for agent node. | Mock Session with `Status="running"`, `CurrentState="agent_node"`; mock tracks GetSessionDataSafe calls | RuntimeMessage with `claudeSessionID=<uuid>` | Session.GetSessionDataSafe called with key `"agent_node.ClaudeSessionID"`; returns stored value |
| `TestProcessEvent_SessionUpdateEventHistorySafeCalled` | `unit` | Verifies Session.UpdateEventHistorySafe called to record event. | Mock Session with `Status="running"`; agent node; SessionData contains Claude session ID; mock tracks UpdateEventHistorySafe calls; TransitionEvaluator returns valid transition; TransitionToNode succeeds | Valid event RuntimeMessage | Session.UpdateEventHistorySafe called once with Event containing correct fields |
| `TestProcessEvent_SessionMethodsNotDirectlyModified` | `unit` | Verifies EventProcessor uses Session methods, not direct field access. | Mock Session with private fields; only methods exposed; monitor for direct field access attempts | Valid event RuntimeMessage | All Session interactions via methods (GetSessionDataSafe, UpdateEventHistorySafe); no direct field modifications |

### Mock / Dependency Interaction — TransitionEvaluator

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_TransitionEvaluatorCalled` | `unit` | Verifies TransitionEvaluator invoked with correct parameters. | Mock Session with `Status="running"`, `CurrentState="node_a"`; monitor TransitionEvaluator calls (package-level function) | RuntimeMessage with `eventType="approved"` | TransitionEvaluator.EvaluateTransition called with `WorkflowDefinition=<loaded>`, `CurrentState="node_a"`, `EventType="approved"` |

### Mock / Dependency Interaction — TransitionToNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_TransitionToNodeParameters` | `unit` | Verifies TransitionToNode receives correct parameters. | Mock Session with `Status="running"`; TransitionEvaluator returns transition to "target_node" with `isExitTransition=false`; mock TransitionToNode tracks parameters | RuntimeMessage with `message="proceed"` | TransitionToNode.Transition called with `Message="proceed"`, `TargetNodeName="target_node"`, `IsExitTransition=false` |
| `TestProcessEvent_TransitionToNodeExitTransition` | `unit` | Verifies TransitionToNode receives isExitTransition flag. | Mock Session with `Status="running"`; TransitionEvaluator returns exit transition with `isExitTransition=true`; mock TransitionToNode tracks parameters | Valid event RuntimeMessage | TransitionToNode.Transition called with `IsExitTransition=true` |

### Mock / Dependency Interaction — WorkflowDefinitionLoader

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_WorkflowDefinitionLoaderCalled` | `unit` | Verifies WorkflowDefinitionLoader invoked to load workflow. | Mock Session with `Status="running"`; mock WorkflowDefinitionLoader tracks calls | Valid event RuntimeMessage | WorkflowDefinitionLoader called with session's workflow name |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_CurrentStateTransitionedByTransitionToNode` | `unit` | Verifies CurrentState transition delegated to TransitionToNode. | Mock Session with `Status="running"`, `CurrentState="node_a"`; TransitionToNode updates CurrentState to "node_b" | Valid event RuntimeMessage | EventProcessor does NOT directly modify CurrentState; TransitionToNode performs update; final CurrentState is "node_b" |
| `TestProcessEvent_StatusTransitionedByTransitionToNode` | `unit` | Verifies Session.Status transition delegated to TransitionToNode. | Mock Session with `Status="running"`; TransitionEvaluator returns exit transition; TransitionToNode calls Session.Done to set Status to "completed" | Valid event RuntimeMessage | EventProcessor does NOT directly modify Status; TransitionToNode performs update; final Status is "completed" |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestProcessEvent_NoCleanupPerformed` | `unit` | Verifies EventProcessor does not perform cleanup operations. | Mock Session with `Status="running"`; mock RuntimeSocketManager and SessionFinalizer to verify methods NOT called; track all method invocations | Valid event RuntimeMessage | EventProcessor does NOT invoke socket deletion methods or SessionFinalizer; mock verification confirms only Session data methods and TransitionToNode called |
