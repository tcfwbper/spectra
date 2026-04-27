# Node

## Overview

A Node represents a discrete step in a workflow where either an AI agent or a human performs work. Each node is identified by a unique name within the workflow and is associated with a type (`agent` or `human`) and an optional agent role. Nodes do not execute logic themselves; they define the actor responsible for the step and serve as anchors for transitions.

## Behavior

1. A Node is defined within a `WorkflowDefinition` as an element in the `Nodes` array.
2. When the workflow state machine transitions to a node, the runtime dispatches execution to the associated agent role (if `type == "agent"`) or waits for human input (if `type == "human"`).
3. The node remains active until an event is emitted that triggers a transition to another node.
4. If `type == "agent"`, the `agent_role` field must reference a valid `AgentDefinition.Role`.
5. If `type == "human"`, the `agent_role` field must be empty or omitted.
6. The runtime validates that all nodes have unique names within the workflow.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Name | string | Non-empty, PascalCase (Go naming conventions), unique within the workflow | Yes |
| Type | string | Enum: `"agent"`, `"human"` | Yes |
| AgentRole | string | Non-empty PascalCase string if `Type == "agent"`, must be empty if `Type == "human"`, must reference a valid `AgentDefinition.Role` | Conditional (required if `Type == "agent"`) |
| Description | string | May be empty string `""` | Yes (default: `""`) |

## Outputs

### Node Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Name | string | Unique within workflow, PascalCase | Node identifier |
| Type | string | Enum: `"agent"`, `"human"` | Actor type for this node |
| AgentRole | string | PascalCase, must reference valid `AgentDefinition.Role` if `Type == "agent"` | The agent role to invoke for this node |
| Description | string | May be empty | Human-readable description of the node's purpose |

### YAML Representation

```yaml
name: "ArchitectReviewer"
type: "agent"
agent_role: "ArchitectReviewer"
description: "Review the drafted logic specification"
```

```yaml
name: "HumanApproval"
type: "human"
description: "Human reviews and approves the final output"
```

## Invariants

1. **Name Uniqueness**: Within a workflow, all `Node.Name` values must be unique.

2. **Name Format**: `Name` must follow Go PascalCase naming conventions (no spaces, underscores, or special characters).

3. **Type Constraint**: `Type` must be exactly `"agent"` or `"human"`. No other values are allowed. Future extension is not supported.

4. **AgentRole Conditional Requirement**:
   - If `Type == "agent"`, `AgentRole` must be a non-empty PascalCase string.
   - If `Type == "human"`, `AgentRole` must be empty or omitted.

5. **AgentRole Referential Integrity**: If `AgentRole` is provided, it must reference a `Role` field of an existing `AgentDefinition` in `.spectra/agents/<role>.yaml`.

## Edge Cases

- **Condition**: `Name` contains spaces or special characters (e.g., `"Review-Step"`).
  **Expected**: Runtime rejects the node with a validation error: "node name must be PascalCase with no spaces or special characters".

- **Condition**: `Type == "agent"` but `AgentRole` is empty or omitted.
  **Expected**: Runtime rejects the node with an error: "agent_role is required when type is 'agent'".

- **Condition**: `Type == "human"` but `AgentRole` is provided.
  **Expected**: Runtime rejects the node with an error: "agent_role must be empty when type is 'human'".

- **Condition**: `AgentRole` references a non-existent agent definition.
  **Expected**: Runtime rejects the node with an error: "agent '<role>' not found in .spectra/agents/".

- **Condition**: Two nodes in the same workflow have the same `Name`.
  **Expected**: Runtime rejects the workflow with an error: "duplicate node name '<name>'".

- **Condition**: Node is defined but has no incoming or outgoing transitions.
  **Expected**: Runtime issues a warning for unreachable node but does not reject the workflow.

## Related

- [WorkflowDefinition](./workflow_definition.md) - Nodes are defined within workflows
- [Transition](./transition.md) - Transitions connect nodes
- [AgentDefinition](./agent_definition.md) - Nodes reference agent roles
- [Session](../entities/session/session.md) - Runtime tracks current node in session state
