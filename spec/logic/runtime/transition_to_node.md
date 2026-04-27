# TransitionToNode

## Overview

TransitionToNode is responsible for executing the dispatch logic for transitioning to a single node in the workflow state machine. For **regular transitions** it performs node-type-specific actions (printing messages to stdout for human nodes, invoking AgentInvoker for agent nodes), then updates `Session.CurrentState`. For **exit transitions** it skips the node-type-specific action entirely (the workflow terminates immediately and the target node is never expected to do work), updates `Session.CurrentState` to the exit `to_node`, then calls `Session.Done`. TransitionToNode uses fail-fast semantics with internal error handling: any operation failure (except best-effort persistence) immediately constructs a RuntimeError, calls Session.Fail internally, and returns an error to the caller. The caller (EventProcessor) does not need to call Session.Fail again.

## Behavior

1. TransitionToNode is invoked by EventProcessor after a matching transition is found by TransitionEvaluator.
2. TransitionToNode receives the following inputs: `Message` (from Event.Message), `TargetNodeName` (from transition.ToNode), and `IsExitTransition` (boolean flag indicating whether this is an exit transition).
3. TransitionToNode loads the target node definition from WorkflowDefinition using `TargetNodeName`.
4. If the target node does not exist in the workflow, TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, `Message="target node not found: '<TargetNodeName>'"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"target node '<TargetNodeName>' not found in workflow"`. This should not occur if workflow validation is correct, but is handled defensively.
5. If `IsExitTransition == true`, TransitionToNode **skips the node-type-specific action**. The exit `to_node` is a terminal placeholder; no stdout print and no Claude CLI invocation occur. TransitionToNode proceeds directly to step 6.
6. If `IsExitTransition == false`, TransitionToNode performs node-type-specific actions based on the target node's `type` field:
   - **If `type == "human"`**:
     - TransitionToNode prints the `Message` to standard output (stdout) using a simple format: `"[Human Node: <TargetNodeName>] <Message>"`.
     - If `Message` is an empty string `""`, TransitionToNode prints: `"[Human Node: <TargetNodeName>] (no message)"`.
     - The output includes a newline at the end.
   - **If `type == "agent"`**:
     - TransitionToNode loads the AgentDefinition using AgentDefinitionLoader with `node.AgentRole` as the role identifier.
     - If AgentDefinitionLoader fails (agent not found or load error), TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, `Message="failed to load agent definition for role '<AgentRole>'"`, and details from the loader error, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to load agent definition for role '<AgentRole>': <error-details>"`.
     - TransitionToNode invokes `AgentInvoker.InvokeAgent(NodeName=TargetNodeName, Message=Message, AgentDefinition=loadedAgentDef)`.
     - If AgentInvoker returns an error, TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, `Message="failed to invoke agent for node '<TargetNodeName>'"`, and details from the AgentInvoker error, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to invoke agent for node '<TargetNodeName>': <error-details>"`.

7. After the (possibly skipped) action completes successfully, TransitionToNode calls `Session.UpdateCurrentStateSafe(TargetNodeName)` to update the session's current state.
8. `Session.UpdateCurrentStateSafe` follows a pure best-effort contract and always returns `nil` (empty input is logged as a warning and ignored; persistence failures are logged as warnings; valid input always succeeds in memory). TransitionToNode passes a workflow-validated `TargetNodeName`, so the no-op-on-empty branch is unreachable here. TransitionToNode does not check the return value.
9. If `IsExitTransition == true`, TransitionToNode calls `Session.Done(terminationNotifier)` to transition the session to "completed" status.
10. If `Session.Done` returns an error (e.g., status is not "running"), TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, `Message="failed to complete session after exit transition"`, and details from the Session.Done error, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to complete session after exit transition: <error-details>"`.

11. TransitionToNode returns `nil` (success).
12. TransitionToNode does not catch panics. Panic recovery is handled by MessageRouter.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Session | *Session | Reference to the Session entity shared across all runtime components | Yes |
| WorkflowDefinition | WorkflowDefinition | Valid, fully validated WorkflowDefinition loaded from storage | Yes |
| AgentDefinitionLoader | AgentDefinitionLoader | Loader for agent definitions | Yes |
| AgentInvoker | AgentInvoker | Invoker for starting Claude CLI agent processes | Yes |
| TerminationNotifier | chan<- struct{} | Channel for notifying the main loop of session termination (passed to Session methods) | Yes |

### For TransitionToNode Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Message | string | Plain text message from Event.Message, may be empty string `""`, may contain any characters including quotes, newlines, etc. | Yes |
| TargetNodeName | string | Non-empty, PascalCase, must reference a valid node name in WorkflowDefinition.Nodes | Yes |
| IsExitTransition | bool | `true` if this transition is an exit transition, `false` otherwise | Yes |

## Outputs

### Success Case

**Return value**: `nil` (no error)

**Side effects**:
- For regular transitions to human nodes: Message printed to stdout
- For regular transitions to agent nodes: Claude CLI process started via AgentInvoker
- For exit transitions: no stdout print, no agent invocation
- `Session.CurrentState` updated to `TargetNodeName` (in memory and persisted via UpdateCurrentStateSafe)
- If `IsExitTransition == true`: `Session.Status` transitioned to "completed" (in memory and persisted via Session.Done)

### Error Cases

All errors are handled internally by TransitionToNode, which constructs a RuntimeError, calls Session.Fail, and returns an error to the caller. The caller (EventProcessor) should NOT call Session.Fail again.

| Error Condition | Error Behavior |
|----------------|----------------|
| Target node not found in workflow | Constructs RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, returns error: `"target node '<TargetNodeName>' not found in workflow"`. |
| AgentDefinitionLoader fails to load agent | Constructs RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, returns error: `"failed to load agent definition for role '<AgentRole>': <error-details>"`. |
| AgentInvoker returns error | Constructs RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, returns error: `"failed to invoke agent for node '<TargetNodeName>': <error-details>"`. |
| Session.Done returns error (exit transition case) | Constructs RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, returns error: `"failed to complete session after exit transition: <error-details>"`. |

## Invariants

1. **Action-Before-State-Update (Regular Transitions Only)**: For regular transitions, TransitionToNode must execute the node-type-specific action (print message or invoke agent) before updating `Session.CurrentState`. For exit transitions, the action is skipped entirely; only CurrentState update and Session.Done occur.

2. **State-Before-Done**: For exit transitions, TransitionToNode must update `Session.CurrentState` to the target node before calling `Session.Done`. The session completes with `CurrentState` set to the exit transition's target node.

3. **Fail-Fast Non-Persistence Operations**: All operations except persistence (node lookup, agent definition loading, agent invocation, Session.Done status validation) must use fail-fast semantics. Any failure immediately triggers RuntimeError construction, Session.Fail call, and error return to caller.

4. **Best-Effort Persistence**: Persistence operations (via Session.UpdateCurrentStateSafe, Session.Done's internal persistence) use best-effort semantics. Persistence failures are logged but do not halt execution.

5. **Single Responsibility**: TransitionToNode focuses on transition logic and dispatch. It does not manage event recording, transition evaluation, or human notifications beyond stdout printing.

6. **Session State Modification via Methods**: TransitionToNode must not directly modify Session structure's internal fields. All state updates must be performed via Session's thread-safe methods (UpdateCurrentStateSafe, Done, Fail).

7. **Internal Error Handling**: TransitionToNode handles all errors internally by calling Session.Fail before returning errors to the caller. The caller must NOT call Session.Fail again. This prevents duplicate failure notifications and ensures consistent error recording.

8. **No Rollback**: TransitionToNode does not implement rollback logic. If a step fails after partial completion (e.g., agent process started but Session.Done fails), the session transitions to "failed" status without undoing previous steps.

9. **Human Node Message Format**: For human nodes, the stdout output format is fixed: `"[Human Node: <TargetNodeName>] <Message>\n"`. This format is not configurable.

10. **Agent Node Message Passthrough**: For agent nodes, the `Message` is passed directly to AgentInvoker without modification or validation beyond what AgentInvoker performs.

11. **Exit Transition No-Op Target**: When `IsExitTransition == true`, the node-type-specific dispatch is skipped regardless of the target node's `type`. Workflow validation already enforces that exit `to_node` must be `type=="human"`, but even so, TransitionToNode never prints to stdout for exit-transition human targets and never invokes AgentInvoker for exit-transition targets. The exit `to_node` is a labeled terminus, not an executing node.

12. **UpdateCurrentStateSafe Always Succeeds**: Session.UpdateCurrentStateSafe follows a pure best-effort contract and always returns `nil`. For workflow-validated, non-empty `TargetNodeName` (the only input TransitionToNode supplies), the in-memory write is guaranteed; persistence is best-effort. TransitionToNode treats this as a guaranteed success operation.

## Edge Cases

- **Condition**: `TargetNodeName` references a node that does not exist in WorkflowDefinition.Nodes.
  **Expected**: TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"target node '<TargetNodeName>' not found in workflow"`. This should not occur if workflow validation is correct.

- **Condition**: Target node is a human node and `Message` is an empty string `""`.
  **Expected**: TransitionToNode prints to stdout: `"[Human Node: <TargetNodeName>] (no message)\n"`.

- **Condition**: Target node is a human node and `Message` contains newlines and special characters.
  **Expected**: TransitionToNode prints the message as-is, preserving all characters. The output is: `"[Human Node: <TargetNodeName>] <Message>\n"` where `<Message>` includes all newlines and special characters.

- **Condition**: Target node is a human node and `Message` is very large (e.g., 1 MB).
  **Expected**: TransitionToNode prints the entire message to stdout. Performance may degrade, but no error occurs unless stdout write fails (e.g., pipe closed). If stdout write fails, this is a system-level error and may cause a panic or OS-level signal; TransitionToNode does not handle this.

- **Condition**: Target node is an agent node and `node.AgentRole` is empty or invalid.
  **Expected**: AgentDefinitionLoader returns an error. TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns the error. This should not occur if workflow validation ensures agent nodes have valid agent_role fields.

- **Condition**: Target node is an agent node and AgentDefinitionLoader fails (agent file not found, parse error).
  **Expected**: TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to load agent definition for role '<AgentRole>': <error-details>"`.

- **Condition**: Target node is an agent node and AgentInvoker.InvokeAgent returns an error (e.g., `claude` command not found, working directory invalid, UUID generation failure).
  **Expected**: TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns an error: `"failed to invoke agent for node '<TargetNodeName>': <error-details>"`.

- **Condition**: `Session.UpdateCurrentStateSafe` is called (always succeeds for workflow-validated input due to pure best-effort contract).
  **Expected**: The method always returns `nil`. TransitionToNode continues execution without checking the return value. The in-memory `Session.CurrentState` is updated and authoritative.

- **Condition**: `IsExitTransition == true` and `Session.Done` is called, but `Session.Status != "running"`.
  **Expected**: `Session.Done` returns an error: `"cannot complete session: status is '<actual-status>', expected 'running'"`. TransitionToNode constructs a RuntimeError with `Issuer="TransitionToNode"`, calls `Session.Fail(runtimeError, terminationNotifier)`, and returns the error: `"failed to complete session after exit transition: <error-details>"`.

- **Condition**: `IsExitTransition == true` and target node is a human node.
  **Expected**: TransitionToNode skips the stdout print, updates `Session.CurrentState` to the exit `to_node`, then calls `Session.Done`. Nothing is printed for the exit target.

- **Condition**: `IsExitTransition == true` and target node is an agent node.
  **Expected**: This configuration is rejected by workflow validation (exit `to_node` must be human). If validation is bypassed and TransitionToNode receives such a transition, it still skips AgentInvoker entirely, updates CurrentState, and calls Session.Done. No Claude CLI process is started.

- **Condition**: `IsExitTransition == false` (regular transition).
  **Expected**: TransitionToNode performs the node-type-specific action, updates `Session.CurrentState`, and returns without calling `Session.Done`. The session remains in "running" status.

- **Condition**: TransitionToNode is invoked concurrently for the same session (due to concurrent event emissions).
  **Expected**: Concurrent invocations are serialized by Session's internal write locks (UpdateCurrentStateSafe, Done, Fail all use write locks). The last successful TransitionToNode call wins for `CurrentState`. If one invocation fails and calls Session.Fail, subsequent invocations may fail due to session status no longer being "running" (though TransitionToNode itself does not validate status; EventProcessor does).

- **Condition**: AgentInvoker.InvokeAgent starts a Claude CLI process, but the process exits immediately with an error (e.g., invalid model, permission denied).
  **Expected**: AgentInvoker returns success (it only ensures the process starts, not that it completes successfully). TransitionToNode returns success. Agent execution failures after startup are reported by the agent emitting an error event (handled by ErrorProcessor) and are outside TransitionToNode's scope.

- **Condition**: Session.Fail is called (by TransitionToNode or concurrently by another component) while TransitionToNode is executing.
  **Expected**: If Session.Fail is called before TransitionToNode's Session.Done call, Session.Done will fail with `"cannot complete session: status is 'failed', expected 'running'"`. TransitionToNode will call Session.Fail again, which will fail with `"session already failed"` but not overwrite the first error. TransitionToNode returns an error, which EventProcessor propagates in the RuntimeResponse.

- **Condition**: TerminationNotifier channel is closed or nil when TransitionToNode calls Session.Done or Session.Fail.
  **Expected**: If the channel is closed, the send operation in Session.Done/Fail will panic. If the channel is nil, the send operation will panic. These panics are not caught by TransitionToNode; panic recovery is handled by MessageRouter. In production, the caller must ensure terminationNotifier is a valid, open, buffered channel.

## Related

- [EventProcessor](./event_processor.md) - Invokes TransitionToNode after finding a matching transition
- [Session](../entities/session/session.md) - Session provides UpdateCurrentStateSafe, Done, and Fail methods
- [AgentInvoker](./agent_invoker.md) - Invoked by TransitionToNode to start agent processes
- [AgentDefinitionLoader](../storage/agent_definition_loader.md) - Loads agent definitions for agent nodes
- [WorkflowDefinition](../components/workflow_definition.md) - Provides node definitions
- [Node](../components/node.md) - Node structure with type and agent_role fields
- [Event](../entities/event.md) - Source of the `Message` delivered by TransitionToNode
- [RuntimeError](../entities/runtime_error.md) - Constructed by TransitionToNode on failure
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
