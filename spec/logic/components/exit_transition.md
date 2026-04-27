# ExitTransition

## Overview

An ExitTransition identifies a specific transition in a workflow that triggers workflow completion when traversed. It is defined by the triple (`from_node`, `event_type`, `to_node`) and must reference a valid transition in the workflow's `Transitions` array. ExitTransitions provide an explicit, edge-based mechanism for determining workflow termination, which is particularly important in workflows with cycles where node-based termination would be ambiguous.

## Behavior

1. An ExitTransition is defined within a `WorkflowDefinition` as an element in the `ExitTransitions` array.
2. Each ExitTransition specifies a `from_node`, `event_type`, and `to_node` that must exactly match a transition defined in the workflow's `Transitions` array.
3. When the runtime processes an event, it first determines the target transition by looking up (current node, event_type) in the workflow's `Transitions` array to find the `to_node`. Then it checks if this complete transition triple (current node, event_type, to_node) matches any ExitTransition.
4. If a match is found, the runtime transitions `CurrentState` to the `to_node` in memory, then immediately marks the session `Status` as `"completed"` in memory. Persistence to disk is best-effort.
5. The target node (`to_node`) does not execute any agent or human actions when reached via an ExitTransition. TransitionToNode skips the node-type-specific dispatch (no stdout print for human nodes, no Claude CLI invocation for agent nodes) and proceeds directly to updating CurrentState and calling Session.Done.
6. The runtime validates that each ExitTransition's `to_node` references a node with `type == "human"`.
7. The runtime validates that each ExitTransition corresponds to exactly one transition in the `Transitions` array (exact match on all three fields).
8. If multiple ExitTransitions are defined, any one of them being traversed will trigger workflow completion.
9. An ExitTransition can only trigger completion once per session. Once traversed, the session immediately transitions to `"completed"` status.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| FromNode | string | Non-empty, must reference a valid `Node.Name` in the workflow's `Nodes` array | Yes |
| EventType | string | Non-empty, PascalCase (Go naming conventions), must be a defined event type | Yes |
| ToNode | string | Non-empty, must reference a valid `Node.Name` in the workflow's `Nodes` array, referenced node must have `type == "human"` | Yes |

## Outputs

### ExitTransition Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| FromNode | string | Valid `Node.Name` | Source node of the exit transition |
| EventType | string | PascalCase, valid event type | Event that triggers workflow completion |
| ToNode | string | Valid `Node.Name`, must have `type == "human"` | Destination node (workflow terminates here) |

### YAML Representation

**Example 1: Exit transition from human approval**
```yaml
from_node: "HumanApproval"
event_type: "RequirementApproved"
to_node: "HumanRequirement"
```

**Example 2: Exit transition for rejection**
```yaml
from_node: "HumanApproval"
event_type: "SpecificationRejected"
to_node: "HumanRequirement"
```

**Example 3: Multiple exit transitions**
```yaml
exit_transitions:
  - from_node: "HumanApproval"
    event_type: "RequirementApproved"
    to_node: "HumanRequirement"
  - from_node: "HumanApproval"
    event_type: "SpecificationRejected"
    to_node: "HumanRequirement"
  - from_node: "QualityGate"
    event_type: "CriticalFailure"
    to_node: "HumanReview"
```

## Invariants

1. **Transition Existence**: Each ExitTransition must correspond to exactly one transition in the workflow's `Transitions` array, with an exact match on `from_node`, `event_type`, and `to_node`.

2. **Target Node Type**: The `to_node` of every ExitTransition must reference a node with `type == "human"`. Exit transitions to agent nodes are prohibited.

3. **Node Referential Integrity**: `FromNode` and `ToNode` must both reference valid `Node.Name` values defined in the workflow's `Nodes` array.

4. **Event Type Validation**: `EventType` must be a valid event type identifier (PascalCase, non-empty, defined in the workflow).

5. **One-Time Trigger**: Each ExitTransition triggers workflow completion at most once per session. Once an ExitTransition is traversed, the session immediately transitions to `"completed"` status.

6. **In-Memory State Completion**: When an ExitTransition is traversed, the runtime must update both `CurrentState` (to `to_node`) and session `Status` (to `"completed"`) in memory. These updates occur in sequence: first `CurrentState`, then `Status`. Persistence to SessionMetadataStore is best-effort and may occur in separate write operations. The in-memory state is authoritative.

7. **No Duplicate ExitTransitions**: Within a workflow, no two ExitTransitions may have identical (`from_node`, `event_type`, `to_node`) triples.

8. **Any-One Completion**: If multiple ExitTransitions are defined, traversing any one of them triggers workflow completion. They are logically OR-ed, not AND-ed.

## Edge Cases

- **Condition**: An ExitTransition references a `from_node`, `event_type`, and `to_node` that do not match any transition in `Transitions`.
  **Expected**: Runtime rejects the workflow with a validation error: "exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition definition".

- **Condition**: An ExitTransition's `to_node` references a node with `type == "agent"`.
  **Expected**: Runtime rejects the workflow with a validation error: "exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') must target a human node, but targets '<node>' with type 'agent'".

- **Condition**: An ExitTransition's `from_node` or `to_node` references a non-existent node.
  **Expected**: Runtime rejects the workflow with a validation error: "exit transition references non-existent node '<node>'".

- **Condition**: An ExitTransition's `event_type` is not a valid PascalCase identifier (e.g., contains spaces or special characters).
  **Expected**: Runtime rejects the workflow with a validation error: "event_type must be PascalCase with no spaces".

- **Condition**: Two ExitTransitions have identical (`from_node`, `event_type`, `to_node`) values.
  **Expected**: Runtime rejects the workflow with a validation error: "duplicate exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>')".

- **Condition**: A transition is defined in `Transitions` but not in `ExitTransitions`, and that transition is traversed.
  **Expected**: Runtime processes it as a normal transition. The session continues running. Only transitions explicitly defined in `ExitTransitions` trigger completion.

- **Condition**: A workflow defines a cycle where a node targeted by an ExitTransition also has outgoing transitions.
  **Expected**: Runtime allows this configuration but may issue a warning: "exit target node '<node>' has outgoing transitions that will never be used". The outgoing transitions are ignored when the node is reached via an ExitTransition.

- **Condition**: An ExitTransition is traversed in a session.
  **Expected**: Runtime transitions `CurrentState` to the `to_node` in memory, then immediately marks `Status` as `"completed"` in memory. Persistence to SessionMetadataStore is best-effort and may occur in separate writes. The `to_node` does not execute any agent or human actions. The session terminates.

- **Condition**: `ExitTransitions` array is empty.
  **Expected**: Runtime rejects the workflow with a validation error: "at least one exit transition required".

- **Condition**: Session attempts to emit an event that would match an ExitTransition, but the session is not at the `from_node`.
  **Expected**: The ExitTransition is not triggered. The runtime processes the event according to the normal transition rules (which may result in a "no transition found" error if the event is invalid for the current node).

- **Condition**: A workflow has multiple ExitTransitions from the same `from_node` with different event types.
  **Expected**: Runtime allows this. Each ExitTransition is evaluated independently. Traversing any one of them triggers completion.


## Related

- [WorkflowDefinition](./workflow_definition.md) - ExitTransitions are defined within workflows
- [Transition](./transition.md) - ExitTransitions must reference valid transitions
- [Node](./node.md) - ExitTransitions reference nodes
- [Session](../entities/session/session.md) - Runtime evaluates ExitTransitions to determine session completion
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
