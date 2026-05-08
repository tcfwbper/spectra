# TransitionToNode

## Overview

TransitionToNode is responsible for executing the dispatch logic when transitioning to a target node in the workflow state machine. It loads the target node definition, performs node-type-specific actions (printing messages to stdout for human nodes, invoking AgentInvoker for agent nodes), and updates `PersistentSession.CurrentState` (which automatically persists the new state). TransitionToNode does not manage session lifecycle (Run, Done, Fail), does not evaluate transitions, and does not know whether a transition is an exit transition.

## Boundaries

- Owns: target node lookup from WorkflowDefinition.
- Owns: node-type dispatch (stdout print for human, AgentInvoker invocation for agent).
- Owns: loading AgentDefinition via AgentDefinitionLoader for agent nodes.
- Owns: updating `PersistentSession.CurrentState` after successful dispatch (which automatically persists).
- Delegates: transition evaluation (finding which transition matches) to TransitionEvaluator / caller.
- Delegates: session lifecycle transitions (Run, Done, Fail) to the caller (EventProcessor).
- Delegates: RuntimeError construction and `PersistentSession.Fail` invocation to the caller.
- Delegates: persistence to PersistentSession (automatic via UpdateCurrentStateSafe).
- Delegates: process lifecycle monitoring after agent startup to the event-driven runtime model.
- Must not: call `PersistentSession.Fail`, `PersistentSession.Done`, or `PersistentSession.Run`.
- Must not: construct RuntimeError.
- Must not: hold or use a termination notifier channel.
- Must not: determine whether a transition is an exit transition.
- Must not: call SessionMetadataStore.Write() or EventStore.Append() directly.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State container with auto-persist | `UpdateCurrentStateSafe(newState)` | Must not call `Fail()`, `Done()`, `Run()`, or any lifecycle method |
| `WorkflowDefinition` | Node lookup source | Read `Nodes()` to find target node by name | Must not modify |
| `AgentDefinitionLoader` | Agent definition retrieval | `Load(agentRole)` | Must not cache results or modify loader state |
| `AgentInvoker` | Agent process startup | `InvokeAgent(NodeName, Message, AgentDefinition)` | Must not monitor process after startup |
| stdout | Human output | Write formatted message | Must not read from stdin |

Construction constraint: TransitionToNode is initialized with references to `PersistentSession`, `WorkflowDefinition`, `AgentDefinitionLoader`, and `AgentInvoker`. These are provided by the runtime layer at session startup. TransitionToNode does not construct these dependencies internally.

## Behavior

1. TransitionToNode is invoked by EventProcessor after a matching transition is found.
2. Receives `TargetNodeName` and `Message` as invocation parameters.
3. Loads the target node definition from `WorkflowDefinition.Nodes()` by matching `Node.Name() == TargetNodeName`.
4. If the target node does not exist in the workflow, returns an error: `"target node '<TargetNodeName>' not found in workflow"`.
5. Performs node-type-specific actions based on the target node's `Type()`:
   - **If `Type() == "human"`**:
     - Prints a formatted message to stdout: `"[<TargetNodeName>] <Message>\n"`.
     - If `Message` is an empty string, prints: `"[<TargetNodeName>] (no message)\n"`.
   - **If `Type() == "agent"`**:
     - Calls `AgentDefinitionLoader.Load(node.AgentRole())` to load the agent definition.
     - If loading fails, returns an error: `"failed to load agent definition for role '<AgentRole>': <error-details>"`.
     - Calls `AgentInvoker.InvokeAgent(NodeName=TargetNodeName, Message=Message, AgentDefinition=loadedDef)`.
     - If AgentInvoker returns an error, returns an error: `"failed to invoke agent for node '<TargetNodeName>': <error-details>"`.
6. After the node-type action completes successfully, calls `PersistentSession.UpdateCurrentStateSafe(TargetNodeName)`. PersistentSession automatically persists the updated metadata (non-fatal if persistence fails).
7. If `UpdateCurrentStateSafe` returns an error (in-memory validation failure), returns that error: `"failed to update current state: <error-details>"`.
8. Returns `nil` (success).
9. TransitionToNode does not catch panics. Panic recovery is handled by the caller's infrastructure.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | PersistentSession reference | Valid, constructed via NewPersistentSession | Yes |
| WorkflowDefinition | WorkflowDefinition reference | Valid, fully validated WorkflowDefinition | Yes |
| AgentDefinitionLoader | AgentDefinitionLoader reference | Initialized with correct projectRoot | Yes |
| AgentInvoker | AgentInvoker reference | Initialized with PersistentSession and ProjectRoot | Yes |

### For Invocation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| TargetNodeName | string | Non-empty, PascalCase, must reference a valid node in WorkflowDefinition.Nodes | Yes |
| Message | string | Plain text from Event.Message; may be empty string, may contain any characters including quotes, newlines, etc. | Yes |

## Outputs

### Success Case

**Return value**: `nil` (no error)

**Side effects**:
- For human nodes: formatted message printed to stdout.
- For agent nodes: Claude CLI process started via AgentInvoker (running in background).
- `PersistentSession.CurrentState` updated to `TargetNodeName` in memory (and automatically persisted).

### Error Cases

All errors are returned as plain errors to the caller. TransitionToNode does not construct RuntimeError or call Session.Fail.

| Error Condition | Error Message |
|----------------|--------------|
| Target node not found in workflow | `"target node '<TargetNodeName>' not found in workflow"` |
| AgentDefinitionLoader fails | `"failed to load agent definition for role '<AgentRole>': <error-details>"` |
| AgentInvoker returns error | `"failed to invoke agent for node '<TargetNodeName>': <error-details>"` |
| UpdateCurrentStateSafe returns error | `"failed to update current state: <error-details>"` |

## Invariants

1. **Action-Before-State-Update**: TransitionToNode must execute the node-type-specific action (print message or invoke agent) before updating `Session.CurrentState`. If the action fails, CurrentState is not updated.

2. **Fail-Fast**: Any operation failure immediately returns an error to the caller. No subsequent steps execute after a failure.

3. **No Lifecycle Management**: TransitionToNode must not call `PersistentSession.Fail`, `PersistentSession.Done`, or `PersistentSession.Run`. All lifecycle decisions belong to the caller.

4. **No RuntimeError Construction**: TransitionToNode returns plain Go errors. The caller is responsible for wrapping them into RuntimeError if needed.

5. **Session State Modification via Method**: TransitionToNode must not directly modify Session's internal fields. State update must go through `PersistentSession.UpdateCurrentStateSafe`.

6. **No Rollback**: If a step fails after partial completion (e.g., agent started but state update fails), TransitionToNode does not undo previous steps. It returns an error and the caller decides how to handle.

7. **Human Node Message Format**: Stdout output format is `"[<TargetNodeName>] <Message>\n"`. This format is not configurable.

8. **Agent Node Message Passthrough**: `Message` is passed directly to AgentInvoker without modification.

9. **Uniform Behavior**: TransitionToNode treats all transitions identically regardless of whether they are exit transitions. It always executes the node-type action and updates state.

10. **No Termination Notification**: TransitionToNode does not hold or use any termination channel. Termination signaling is the caller's responsibility.

## Edge Cases

- Condition: `TargetNodeName` references a node that does not exist in WorkflowDefinition.Nodes.
  Expected: Returns error `"target node '<TargetNodeName>' not found in workflow"`. No action is performed, no state is updated.

- Condition: Target node is a human node and `Message` is an empty string.
  Expected: Prints to stdout: `"[<TargetNodeName>] (no message)\n"`.

- Condition: Target node is a human node and `Message` contains newlines and special characters.
  Expected: Prints the message as-is. Output: `"[<TargetNodeName>] <Message>\n"` preserving all characters.

- Condition: Target node is an agent node and `node.AgentRole()` references a non-existent agent definition.
  Expected: AgentDefinitionLoader returns error. TransitionToNode returns `"failed to load agent definition for role '<AgentRole>': <error-details>"`. No agent process started, no state updated.

- Condition: Target node is an agent node and AgentInvoker fails (e.g., `claude` not in PATH, working directory invalid).
  Expected: Returns `"failed to invoke agent for node '<TargetNodeName>': <error-details>"`. No state updated.

- Condition: Node-type action succeeds but `UpdateCurrentStateSafe` returns error (empty string — unreachable since TargetNodeName is non-empty, but defensively handled).
  Expected: Returns `"failed to update current state: <error-details>"`. Action side-effects (stdout print or agent process) are not rolled back.

- Condition: Stdout write fails (e.g., pipe closed) during human node message print.
  Expected: Returns an error. State is not updated.

- Condition: Multiple concurrent calls to TransitionToNode for the same session.
  Expected: `PersistentSession.UpdateCurrentStateSafe` serializes state writes via session's internal lock. Last successful call wins for CurrentState. Node-type actions (agent startup, stdout print) execute independently without coordination.

- Condition: AgentInvoker starts a process that immediately exits with error.
  Expected: AgentInvoker returns success (only ensures startup). TransitionToNode returns success. Post-startup failures are handled by the event-driven model (agent emits error event).

## Related

- [PersistentSession](./persistent_session.md) — State container with automatic persistence
- [TransitionEvaluator](./transition_evaluator.md) — Finds matching transitions; invoked by the caller before TransitionToNode
- [AgentInvoker](./agent_invoker.md) — Invoked by TransitionToNode to start agent processes
- [AgentDefinitionLoader](../storage/agent_definition_loader.md) — Loads agent definitions for agent nodes
- [WorkflowDefinition](../components/workflow_definition.md) — Provides node definitions for lookup
- [Node](../components/node.md) — Node structure with Type and AgentRole fields
- [Session current_state](../entities/session/current_state.md) — UpdateCurrentStateSafe method spec (wrapped by PersistentSession)
- [Session lifecycle](../entities/session/lifecycle.md) — Done and Fail methods owned by caller (wrapped by PersistentSession)
