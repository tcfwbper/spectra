# Node

## Overview

A Node represents a discrete step in a workflow where either an AI agent or a human performs work. It is a pure immutable value object that validates its own field formats at construction time. Node does not know about workflow-level constraints (such as name uniqueness across a workflow or agent role existence); those are owned by the workflow-level aggregator.

## Boundaries

- Owns: construction-time validation of all fields (Name, Type, AgentRole, Description).
- Owns: immutability guarantee for all fields after construction.
- Owns: conditional format validation (AgentRole required when Type is agent, forbidden when Type is human).
- Delegates: name uniqueness within a workflow to the workflow-level aggregator.
- Delegates: referential integrity of AgentRole (whether the agent actually exists) to the I/O loader layer.
- Delegates: runtime dispatch behavior (invoking agents, waiting for human input) to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `components` package.
- Must not: be constructed via struct literal — must use the provided constructor.

## Dependencies

None. This value object depends only on Go standard library types.

Construction constraint: Must be constructed via `NewNode(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewNode(name string, nodeType string, agentRole string, description string) (*Node, error)` that validates all fields and returns an immutable Node value.
2. Validates that `Name` is a non-empty PascalCase string (starts with uppercase letter, contains only alphanumeric characters).
3. Validates that `Type` is exactly `"agent"` or `"human"`. No other values are accepted.
4. If `Type == "agent"`, validates that `AgentRole` is a non-empty PascalCase string.
5. If `Type == "human"`, validates that `AgentRole` is an empty string.
6. Accepts `Description` as any string including empty string. No validation on Description content.
7. Returns a validation error if any constraint is violated. No Node is created on failure.
8. All fields are immutable after construction.
9. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Name | string | Non-empty, PascalCase (starts with uppercase letter, alphanumeric only) | Yes |
| Type | string | Exactly `"agent"` or `"human"` | Yes |
| AgentRole | string | Non-empty PascalCase if Type is `"agent"`; must be empty if Type is `"human"` | Conditional |
| Description | string | Any string (empty allowed) | No |

## Outputs

### Node Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Name | string | Non-empty, PascalCase | Unique identifier for the node within a workflow |
| Type | string | `"agent"` or `"human"` | Actor type for this step |
| AgentRole | string | PascalCase or empty | The agent role to invoke (empty for human nodes) |
| Description | string | Any string | Human-readable description of the node's purpose |

### Error Output

| Condition | Error |
|-----------|-------|
| Name is empty | `"node name cannot be empty"` |
| Name is not PascalCase | `"node name must be PascalCase (start with uppercase, alphanumeric only)"` |
| Type is not `"agent"` or `"human"` | `"node type must be 'agent' or 'human'"` |
| Type is `"agent"` and AgentRole is empty | `"agent_role is required when type is 'agent'"` |
| Type is `"agent"` and AgentRole is not PascalCase | `"agent_role must be PascalCase (start with uppercase, alphanumeric only)"` |
| Type is `"human"` and AgentRole is not empty | `"agent_role must be empty when type is 'human'"` |

## Invariants

1. **Name PascalCase**: `Name` is always a non-empty PascalCase string after construction.
2. **Type Enum**: `Type` is always exactly `"agent"` or `"human"` after construction.
3. **AgentRole Conditional**: If `Type == "agent"`, `AgentRole` is always a non-empty PascalCase string. If `Type == "human"`, `AgentRole` is always an empty string.
4. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
5. **Construction Only Via Constructor**: Must be constructed via `NewNode`. Direct struct literal construction is forbidden.

## Edge Cases

- Condition: `Name` is an empty string.
  Expected: Constructor returns error `"node name cannot be empty"`. No Node is created.

- Condition: `Name` starts with a lowercase letter (e.g., `"reviewStep"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `Name` contains non-alphanumeric characters (e.g., `"Review-Step"`, `"review_step"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `Type` is an unrecognized value (e.g., `"bot"`, `""`).
  Expected: Constructor returns error `"node type must be 'agent' or 'human'"`.

- Condition: `Type == "agent"` but `AgentRole` is empty.
  Expected: Constructor returns error `"agent_role is required when type is 'agent'"`.

- Condition: `Type == "human"` but `AgentRole` is non-empty (e.g., `"Architect"`).
  Expected: Constructor returns error `"agent_role must be empty when type is 'human'"`.

- Condition: `Description` is an empty string.
  Expected: Constructor accepts this. Description is optional.

- Condition: `Type == "agent"` and `AgentRole` starts with lowercase (e.g., `"architect"`).
  Expected: Constructor returns error indicating AgentRole must be PascalCase.

## Related

- [Transition](./transition.md) — Transitions reference Node names as FromNode and ToNode
- [ExitTransition](./exit_transition.md) — ExitTransitions reference Node names
- [Event](../entities/event.md) — Events are emitted from Nodes (runtime responsibility)
- [Session](../entities/session/session.md) — Runtime tracks current node via Session.CurrentState
