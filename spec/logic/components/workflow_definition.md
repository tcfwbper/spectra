# WorkflowDefinition

## Overview

A WorkflowDefinition is the workflow-level aggregator that assembles Nodes, Transitions, and ExitTransitions into a validated, immutable workflow graph. It performs all cross-component structural validation at construction time: node name uniqueness, transition referential integrity, exit transition correspondence, entry node type constraint, outgoing transition coverage, and reachability analysis. WorkflowDefinition does not know about agent definition existence, file systems, or runtime execution — those are owned by the I/O loader layer and the runtime respectively.

## Boundaries

- Owns: cross-component structural validation (all invariants listed below).
- Owns: immutability guarantee for all fields after construction.
- Owns: graph integrity (reachability, outgoing transitions, entry/exit constraints).
- Delegates: agent role existence validation (whether referenced AgentRoles correspond to real AgentDefinition files) to the I/O loader layer.
- Delegates: Name derivation from filename to the I/O loader layer (same pattern as AgentDefinition).
- Delegates: Name uniqueness across workflow definitions to the I/O loader layer (filesystem enforces via filename).
- Delegates: runtime state machine execution, session management, and event processing to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `components` package.
- Must not: be constructed via struct literal — must use the provided constructor.
- Must not: validate that AgentRole values correspond to existing agent definitions.
- Must not: parse or derive Name from YAML content.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `Node` | Child value object | Read via getter methods (Name, Type, AgentRole) | Must not construct Nodes internally; receives pre-constructed instances |
| `Transition` | Child value object | Read via getter methods (FromNode, EventType, ToNode) | Must not construct Transitions internally; receives pre-constructed instances |
| `ExitTransition` | Child value object | Read via getter methods (FromNode, EventType, ToNode) | Must not construct ExitTransitions internally; receives pre-constructed instances |

Construction constraint: Must be constructed via `NewWorkflowDefinition(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewWorkflowDefinition(name string, description string, entryNode string, nodes []*Node, transitions []*Transition, exitTransitions []*ExitTransition) (*WorkflowDefinition, error)` that validates all cross-component constraints and returns an immutable WorkflowDefinition.
2. Validates that `Name` is a non-empty PascalCase string.
3. Accepts `Description` as any string including empty string.
4. Validates that `Nodes` is non-empty.
5. Validates that `Transitions` is non-empty.
6. Validates that `ExitTransitions` is non-empty.
7. Validates that all Node names are unique within the provided nodes.
8. Validates that `EntryNode` references a Node name that exists in `Nodes`.
9. Validates that the node referenced by `EntryNode` has Type `"human"`.
10. Validates that all Transition FromNode and ToNode values reference existing Node names.
11. Validates that no two Transitions share the same (FromNode, EventType) pair (deterministic routing).
12. Validates that each ExitTransition (FromNode, EventType, ToNode) matches exactly one Transition in `Transitions` (same FromNode, EventType, and ToNode).
13. Validates that no two ExitTransitions are identical (same FromNode, EventType, ToNode).
14. Validates that for each ExitTransition, the ToNode references a Node with Type `"human"`.
15. Validates that every Node not targeted by any ExitTransition has at least one outgoing Transition.
16. Validates that every non-entry Node is reachable (has at least one incoming Transition).
17. Returns a validation error if any constraint is violated. No WorkflowDefinition is created on failure.
18. All fields are immutable after construction.
19. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Name | string | Non-empty, PascalCase (starts with uppercase letter, alphanumeric only). Provided by loader, derived from filename. | Yes |
| Description | string | Any string (empty allowed) | No |
| EntryNode | string | Must reference a Node.Name in Nodes; referenced Node must have Type `"human"` | Yes |
| Nodes | []*Node | Non-empty; all Node.Name values must be unique | Yes |
| Transitions | []*Transition | Non-empty; all (FromNode, EventType) pairs must be unique; all FromNode/ToNode must reference valid Node names | Yes |
| ExitTransitions | []*ExitTransition | Non-empty; each must correspond to an existing Transition; each ToNode must be a human-type Node | Yes |

## Outputs

### WorkflowDefinition Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Name | string | Non-empty, PascalCase | Workflow identifier (derived from filename by loader) |
| Description | string | Any string | Human-readable description |
| EntryNode | string | Valid Node.Name, human type | First node where the workflow begins |
| Nodes | []*Node | Non-empty, unique names | All nodes in the workflow |
| Transitions | []*Transition | Non-empty, deterministic routing | All state transitions |
| ExitTransitions | []*ExitTransition | Non-empty, each maps to a Transition | Transitions that trigger workflow completion |

### Error Output

| Condition | Error |
|-----------|-------|
| Name is empty | `"name cannot be empty"` |
| Name is not PascalCase | `"name must be PascalCase (start with uppercase, alphanumeric only)"` |
| Nodes is empty | `"nodes cannot be empty"` |
| Transitions is empty | `"transitions cannot be empty"` |
| ExitTransitions is empty | `"exit_transitions cannot be empty"` |
| Duplicate Node name | `"duplicate node name: '<name>'"` |
| EntryNode not found in Nodes | `"entry_node '<name>' does not reference a valid node"` |
| EntryNode references non-human node | `"entry_node '<name>' must have type 'human'"` |
| Transition FromNode not found | `"transition from_node '<name>' does not reference a valid node"` |
| Transition ToNode not found | `"transition to_node '<name>' does not reference a valid node"` |
| Duplicate (FromNode, EventType) | `"duplicate transition for event '<event_type>' from node '<from_node>'"` |
| ExitTransition has no corresponding Transition | `"exit_transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition"` |
| Duplicate ExitTransition | `"duplicate exit_transition (from_node: '<from>', event_type: '<type>', to_node: '<to>')"` |
| ExitTransition ToNode is not human | `"exit_transition to_node '<name>' must have type 'human'"` |
| Non-exit-target node has no outgoing transitions | `"node '<name>' has no outgoing transitions and is not an exit target"` |
| Non-entry node is unreachable | `"node '<name>' is unreachable (no incoming transitions)"` |

## Invariants

1. **Name PascalCase**: `Name` is always a non-empty PascalCase string after construction.
2. **Nodes Non-Empty**: `Nodes` always contains at least one element.
3. **Transitions Non-Empty**: `Transitions` always contains at least one element.
4. **ExitTransitions Non-Empty**: `ExitTransitions` always contains at least one element.
5. **Node Name Uniqueness**: All Node.Name values in `Nodes` are unique.
6. **Entry Node Validity**: `EntryNode` always references a Node.Name that exists in `Nodes` and has Type `"human"`.
7. **Transition Referential Integrity**: All Transition FromNode and ToNode values reference valid Node names in `Nodes`.
8. **Transition Determinism**: No two Transitions share the same (FromNode, EventType) pair.
9. **ExitTransition Correspondence**: Every ExitTransition (FromNode, EventType, ToNode) matches exactly one Transition in `Transitions`.
10. **ExitTransition Uniqueness**: No two ExitTransitions are identical.
11. **ExitTransition Target Constraint**: For every ExitTransition, the ToNode references a Node with Type `"human"`.
12. **Outgoing Transition Coverage**: Every Node not targeted by any ExitTransition has at least one outgoing Transition.
13. **Reachability**: Every non-entry Node has at least one incoming Transition.
14. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
15. **Construction Only Via Constructor**: Must be constructed via `NewWorkflowDefinition`. Direct struct literal construction is forbidden.
16. **Name From Filename Only**: The Name value is provided externally by the loader (derived from filename). The workflow YAML file does not contain a `name` field.
17. **No AgentRole Existence Check**: WorkflowDefinition does not verify that Node.AgentRole values correspond to existing AgentDefinition instances. This is the I/O loader layer's responsibility.

## Edge Cases

- Condition: `Name` is an empty string.
  Expected: Constructor returns error `"name cannot be empty"`. No WorkflowDefinition is created.

- Condition: `Name` starts with lowercase (e.g., `"defaultWorkflow"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `Nodes` is an empty slice.
  Expected: Constructor returns error `"nodes cannot be empty"`.

- Condition: `Nodes` is nil.
  Expected: Constructor returns error `"nodes cannot be empty"`.

- Condition: `Transitions` is an empty slice or nil.
  Expected: Constructor returns error `"transitions cannot be empty"`.

- Condition: `ExitTransitions` is an empty slice or nil.
  Expected: Constructor returns error `"exit_transitions cannot be empty"`.

- Condition: Two Nodes have the same Name (e.g., two nodes both named `"Architect"`).
  Expected: Constructor returns error `"duplicate node name: 'Architect'"`.

- Condition: `EntryNode` references a name not present in Nodes.
  Expected: Constructor returns error indicating the entry node does not reference a valid node.

- Condition: `EntryNode` references a Node with Type `"agent"`.
  Expected: Constructor returns error indicating entry node must have type human.

- Condition: A Transition's FromNode references a Node name not in Nodes.
  Expected: Constructor returns error indicating the from_node does not reference a valid node.

- Condition: A Transition's ToNode references a Node name not in Nodes.
  Expected: Constructor returns error indicating the to_node does not reference a valid node.

- Condition: Two Transitions share (FromNode: `"HumanApproval"`, EventType: `"Approve"`).
  Expected: Constructor returns error indicating duplicate transition.

- Condition: An ExitTransition (FromNode: `"A"`, EventType: `"Done"`, ToNode: `"B"`) has no corresponding Transition with the same triple.
  Expected: Constructor returns error indicating no corresponding transition.

- Condition: Two identical ExitTransitions.
  Expected: Constructor returns error indicating duplicate exit_transition.

- Condition: An ExitTransition's ToNode references a Node with Type `"agent"`.
  Expected: Constructor returns error indicating exit_transition to_node must have type human.

- Condition: A non-exit-target Node has no outgoing Transitions.
  Expected: Constructor returns error indicating the node has no outgoing transitions.

- Condition: A Node targeted by an ExitTransition has no outgoing Transitions.
  Expected: Constructor accepts this. Exit-target nodes are exempt from the outgoing transition requirement.

- Condition: A non-entry Node has no incoming Transitions (unreachable).
  Expected: Constructor returns error indicating the node is unreachable.

- Condition: `Description` is an empty string.
  Expected: Constructor accepts this. Description is optional.

## Related

- [Node](./node.md) — Child value objects composed into the workflow
- [Transition](./transition.md) — Child value objects defining edges in the workflow graph
- [ExitTransition](./exit_transition.md) — Child value objects identifying completion triggers
- [AgentDefinition](./agent_definition.md) — Referenced by Node.AgentRole (existence validated by I/O loader, not here)
