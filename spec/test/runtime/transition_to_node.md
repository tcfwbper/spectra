# Test Specification: `transition_to_node_test.go`

## Source File Under Test

`runtime/transition_to_node.go`

## Test File

`runtime/transition_to_node_test.go`

---

## `TransitionToNode`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewTransitionToNode_ValidDeps` | `unit` | Constructs TransitionToNode with all valid dependencies. | Create mock PersistentSession, mock WorkflowDefinition, mock AgentDefinitionLoader, and mock AgentInvoker. | `NewTransitionToNode(session, workflowDef, agentDefLoader, agentInvoker)` | Returns non-nil `*TransitionToNode`; no panic |

### Happy Path — Execute

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionToNode_Execute_HumanNode` | `unit` | Prints formatted message to stdout and updates state for a human node. | Mock WorkflowDefinition.Nodes() returns a node with Name()="HumanReview" and Type()="human". Capture stdout via buffer. Mock PersistentSession.UpdateCurrentStateSafe returns nil. | `Execute("HumanReview", "please review this")` | Returns `nil`; stdout contains `"[HumanReview] please review this\n"`; PersistentSession.UpdateCurrentStateSafe called with `"HumanReview"` |
| `TestTransitionToNode_Execute_HumanNodeEmptyMessage` | `unit` | Prints "(no message)" placeholder when message is empty for human node. | Mock WorkflowDefinition.Nodes() returns a node with Name()="HumanReview" and Type()="human". Capture stdout via buffer. Mock PersistentSession.UpdateCurrentStateSafe returns nil. | `Execute("HumanReview", "")` | Returns `nil`; stdout contains `"[HumanReview] (no message)\n"`; PersistentSession.UpdateCurrentStateSafe called with `"HumanReview"` |
| `TestTransitionToNode_Execute_HumanNodeSpecialChars` | `unit` | Prints message with newlines and special characters as-is. | Mock WorkflowDefinition.Nodes() returns a node with Name()="HumanReview" and Type()="human". Capture stdout via buffer. Mock PersistentSession.UpdateCurrentStateSafe returns nil. | `Execute("HumanReview", "line1\n\"quoted\" $var")` | Returns `nil`; stdout contains `"[HumanReview] line1\n\"quoted\" $var\n"` preserving all characters; PersistentSession.UpdateCurrentStateSafe called with `"HumanReview"` |
| `TestTransitionToNode_Execute_AgentNode` | `unit` | Loads agent definition and invokes agent for an agent node. | Mock WorkflowDefinition.Nodes() returns a node with Name()="Coder" and Type()="agent" and AgentRole()="developer". Mock AgentDefinitionLoader.Load("developer") returns a valid AgentDefinition. Mock AgentInvoker.InvokeAgent("Coder", "implement feature", agentDef) returns nil. Mock PersistentSession.UpdateCurrentStateSafe returns nil. | `Execute("Coder", "implement feature")` | Returns `nil`; AgentDefinitionLoader.Load called with `"developer"`; AgentInvoker.InvokeAgent called with `("Coder", "implement feature", loadedDef)`; PersistentSession.UpdateCurrentStateSafe called with `"Coder"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionToNode_Execute_NodeNotFound` | `unit` | Returns error when target node does not exist in workflow. | Mock WorkflowDefinition.Nodes() returns nodes that do not include "NonExistent". | `Execute("NonExistent", "msg")` | Returns error with message `"target node 'NonExistent' not found in workflow"`; PersistentSession.UpdateCurrentStateSafe not called |
| `TestTransitionToNode_Execute_AgentDefLoadFails` | `unit` | Returns error when AgentDefinitionLoader fails. | Mock WorkflowDefinition.Nodes() returns a node with Name()="Coder" and Type()="agent" and AgentRole()="missing-role". Mock AgentDefinitionLoader.Load("missing-role") returns an error "file not found". | `Execute("Coder", "msg")` | Returns error with message `"failed to load agent definition for role 'missing-role': file not found"`; AgentInvoker.InvokeAgent not called; PersistentSession.UpdateCurrentStateSafe not called |
| `TestTransitionToNode_Execute_AgentInvokeFails` | `unit` | Returns error when AgentInvoker returns error. | Mock WorkflowDefinition.Nodes() returns a node with Name()="Coder" and Type()="agent" and AgentRole()="developer". Mock AgentDefinitionLoader.Load("developer") returns a valid AgentDefinition. Mock AgentInvoker.InvokeAgent returns an error "claude not in PATH". Mock PersistentSession.UpdateCurrentStateSafe not called. | `Execute("Coder", "msg")` | Returns error with message `"failed to invoke agent for node 'Coder': claude not in PATH"`; PersistentSession.UpdateCurrentStateSafe not called |
| `TestTransitionToNode_Execute_UpdateStateFails` | `unit` | Returns error when UpdateCurrentStateSafe fails after successful action. | Mock WorkflowDefinition.Nodes() returns a node with Name()="HumanReview" and Type()="human". Capture stdout via buffer. Mock PersistentSession.UpdateCurrentStateSafe returns an error "validation failed". | `Execute("HumanReview", "msg")` | Returns error with message `"failed to update current state: validation failed"`; stdout still contains `"[HumanReview] msg\n"` (action not rolled back) |
| `TestTransitionToNode_Execute_StdoutWriteFails` | `unit` | Returns error when stdout write fails for human node. | Mock WorkflowDefinition.Nodes() returns a node with Name()="HumanReview" and Type()="human". Inject a writer that returns an error on Write (e.g., closed pipe). | `Execute("HumanReview", "msg")` | Returns an error; PersistentSession.UpdateCurrentStateSafe not called |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionToNode_Execute_ActionBeforeStateUpdate` | `unit` | Verifies node-type action executes before state update by recording call order. | Mock WorkflowDefinition.Nodes() returns a node with Name()="Coder" and Type()="agent" and AgentRole()="dev". Mock AgentDefinitionLoader.Load returns valid def. Use a recording mock for AgentInvoker.InvokeAgent and PersistentSession.UpdateCurrentStateSafe to capture invocation order. Both return nil. | `Execute("Coder", "msg")` | Returns `nil`; AgentInvoker.InvokeAgent was called before PersistentSession.UpdateCurrentStateSafe |
| `TestTransitionToNode_Execute_NoLifecycleMethodsCalled` | `unit` | Verifies TransitionToNode never calls Fail, Done, or Run on PersistentSession. | Mock WorkflowDefinition.Nodes() returns a node with Name()="HumanReview" and Type()="human". Capture stdout via buffer. Mock PersistentSession with recording mock that tracks all method calls. UpdateCurrentStateSafe returns nil. | `Execute("HumanReview", "msg")` | Returns `nil`; PersistentSession.Fail, PersistentSession.Done, and PersistentSession.Run never called |
| `TestTransitionToNode_Execute_AgentNodeNoStateUpdateOnInvokeError` | `unit` | Verifies state is not updated when agent invocation fails. | Mock WorkflowDefinition.Nodes() returns a node with Name()="Coder" and Type()="agent" and AgentRole()="dev". Mock AgentDefinitionLoader.Load returns valid def. Mock AgentInvoker.InvokeAgent returns error. Mock PersistentSession with recording mock. | `Execute("Coder", "msg")` | Returns error; PersistentSession.UpdateCurrentStateSafe never called |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionToNode_Execute_ConcurrentCalls` | `race` | Concurrent calls to Execute for the same session serialize state updates. | Mock WorkflowDefinition.Nodes() returns nodes "NodeA" (human) and "NodeB" (human). Capture stdout via buffer (thread-safe). Mock PersistentSession.UpdateCurrentStateSafe serializes via internal lock and records calls. | Launch two goroutines: one calls `Execute("NodeA", "a")`, another calls `Execute("NodeB", "b")` concurrently. | Both calls return `nil`; PersistentSession.UpdateCurrentStateSafe called exactly twice (once with "NodeA", once with "NodeB"); no data race detected by race detector |
