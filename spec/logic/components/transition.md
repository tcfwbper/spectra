# Transition

## Overview

A Transition defines an event-driven state change in a workflow. It specifies the source node (`from_node`), the event type that triggers the transition (`event_type`), and the destination node (`to_node`). Transitions are unconditional and always occur when the specified event is emitted from the source node. Transitions are the only mechanism by which the workflow state machine progresses from one node to another.

## Behavior

1. A Transition is defined within a `WorkflowDefinition` as an element in the `Transitions` array.
2. When an event is emitted in a session, the runtime looks for a transition where `from_node` matches the current node and `event_type` matches the emitted event's type.
3. If exactly one matching transition is found, the runtime unconditionally transitions the session to the `to_node`.
4. If multiple transitions match (same `from_node` and `event_type`), the runtime rejects the workflow definition with a validation error: "duplicate transition for event '<type>' from node '<node>'".
5. If no transition is defined from the current node for the emitted event type, the session status goes to `failed` with error: "no transition defined for event '<type>' from node '<node>'".
6. The runtime validates that `from_node` and `to_node` reference valid node names in the workflow's `Nodes` array.
7. The runtime validates that `from_node` and `to_node` are different. Self-loop transitions (`from_node == to_node`) are not allowed and must be rejected with error: "transition from_node and to_node must be different".
8. The runtime validates that `event_type` is a valid event type (PascalCase, defined in the system or workflow-specific).

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| FromNode | string | Non-empty, must reference a valid `Node.Name` in the workflow's `Nodes` array, must be different from `ToNode` | Yes |
| EventType | string | Non-empty, PascalCase (Go naming conventions), must be a defined event type | Yes |
| ToNode | string | Non-empty, must reference a valid `Node.Name` in the workflow's `Nodes` array, must be different from `FromNode` | Yes |

## Outputs

### Transition Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| FromNode | string | Valid `Node.Name`, must differ from `ToNode` | Source node of the transition |
| EventType | string | PascalCase, valid event type | Event that triggers this transition |
| ToNode | string | Valid `Node.Name`, must differ from `FromNode` | Destination node of the transition |

### YAML Representation

**Example 1: Agent node to another agent node**
```yaml
from_node: "Architect"
event_type: "DraftCompleted"
to_node: "ArchitectReviewer"
```

**Example 2: Review node to human approval**
```yaml
from_node: "ArchitectReviewer"
event_type: "ReviewApproved"
to_node: "HumanApproval"
```

**Example 3: Human node to agent node**
```yaml
from_node: "HumanApproval"
event_type: "RequirementReceived"
to_node: "Architect"
```

**Example 4: Multiple transitions from the same node (different event types)**
```yaml
from_node: "Architect"
event_type: "AmbiguousSpecFound"
to_node: "HumanApproval"

from_node: "Architect"
event_type: "DraftCompleted"
to_node: "ArchitectReviewer"
```

## Invariants

1. **Node Referential Integrity**: `FromNode` and `ToNode` must both reference valid `Node.Name` values defined in the workflow's `Nodes` array.

2. **Event Type Validation**: `EventType` must be a valid event type identifier (PascalCase, non-empty, defined in the system or workflow-specific).

3. **No Self-Loop**: `FromNode` must not equal `ToNode`. Self-loop transitions are prohibited to prevent infinite loops in the workflow state machine.

4. **Transition Uniqueness**: Within a workflow, at most one transition may exist for any given pair of (`FromNode`, `EventType`). Duplicate transitions are rejected during workflow validation.

## Edge Cases

- **Condition**: `FromNode` or `ToNode` references a non-existent node.
  **Expected**: Runtime rejects the workflow with a validation error: "transition references undefined node '<name>'".

- **Condition**: `EventType` is not a valid PascalCase identifier (e.g., contains spaces or special characters).
  **Expected**: Runtime rejects the transition with an error: "event_type must be PascalCase with no spaces".

- **Condition**: A transition where `FromNode == ToNode`.
  **Expected**: Runtime rejects the workflow with an error: "transition from_node and to_node must be different".

- **Condition**: Multiple transitions exist with the same `FromNode` and `EventType`.
  **Expected**: Runtime rejects the workflow with a validation error: "duplicate transition for event '<type>' from node '<node>'".

- **Condition**: Session receives an event that has no transition defined from the current node.
  **Expected**: EventProcessor records the event in EventHistory (audit trail) and returns a RuntimeResponse with `status="error"` and `message="no transition found for event '<type>' from node '<node>'"`. **The session remains in `running` status**; the agent or human may retry with a different event. The session is not automatically failed.

## Related

- [WorkflowDefinition](./workflow_definition.md) - Transitions are defined within workflows
- [Node](./node.md) - Transitions connect nodes
- [Event](../entities/event.md) - Events trigger transitions
- [Session](../entities/session/session.md) - Runtime evaluates transitions to update session state
