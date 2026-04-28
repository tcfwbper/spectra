# WorkflowDefinition

## Overview

A WorkflowDefinition describes the structure and behavior of an event-driven state machine for orchestrating AI agents and human collaboration. It defines the nodes (agent or human actors), transitions (event-driven state changes), and entry/exit points of the workflow. Exit points are defined as specific transitions (ExitTransitions) that trigger workflow completion when traversed. Each workflow is uniquely identified by its name and stored as a YAML file in `.spectra/workflows/`. Workflows are loaded from the user's `.spectra/workflows/` directory. Built-in workflows are copied to `.spectra/workflows/` during `spectra init` if they do not already exist.

## Behavior

1. A WorkflowDefinition is loaded from `.spectra/workflows/<name>.yaml` when `spectra run --workflow <name>` is invoked.
2. The runtime validates that `entry_node` references a valid node name in `nodes` and that the referenced node has `type == "human"`.
3. The runtime validates that all transitions in `exit_transitions` reference valid transitions defined in `transitions` (matching `from_node`, `event_type`, and `to_node`).
4. The runtime validates that for each transition in `exit_transitions`, the `to_node` references a node with `type == "human"`.
5. The runtime validates that all `transitions` reference valid node names and event types.
6. The runtime ensures that every node not targeted by any exit transition has at least one outgoing transition.
7. When a session starts, `CurrentState` is set to the entry node and `Status` is set to `"initializing"`.
8. When an exit transition is traversed (i.e., the session is at `from_node` and the `event_type` is emitted), the runtime first transitions `CurrentState` to the `to_node` in memory, then immediately marks the session `Status` as `"completed"` in memory. Persistence to disk is best-effort. The target node does not execute any agent or human actions.
9. If an undefined event is received or no transition condition matches, EventProcessor records the event (audit trail) and returns an error response. The session remains in `running` status; the caller may retry with a different event.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Name | string | Non-empty, PascalCase (Go naming conventions), unique across all workflows, no spaces or special characters | Yes |
| Description | string | May be empty string `""` | Yes (default: `""`) |
| EntryNode | string | Must reference a `Node.Name` in `Nodes` array, and the referenced node must have `type == "human"` | Yes |
| ExitTransitions | []ExitTransition | Non-empty array, each element must reference a valid transition in `Transitions` (matching `from_node`, `event_type`, `to_node`), and each `to_node` must have `type == "human"` | Yes |
| Nodes | []Node | Non-empty array, all `Node.Name` must be unique | Yes |
| Transitions | []Transition | Non-empty array, all transitions must reference valid nodes and event types | Yes |

## Outputs

### WorkflowDefinition Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Name | string | Non-empty, unique, PascalCase | Workflow identifier |
| Description | string | May be empty | Human-readable description of the workflow's purpose |
| EntryNode | string | Valid `Node.Name`, referenced node must have `type == "human"` | The first node where the workflow begins |
| ExitTransitions | []ExitTransition | Non-empty, each must reference a valid transition in `Transitions`, each `to_node` must have `type == "human"` | Transitions that trigger workflow completion when traversed |
| Nodes | []Node | Non-empty, unique names | Array of all nodes in the workflow |
| Transitions | []Transition | Non-empty | Array of all state transitions driven by events |

### File Format

**File path**: `.spectra/workflows/<Name>.yaml`

**Example**:
```yaml
name: "SimpleSdd"
description: "A simplified specification-driven development workflow"
entry_node: "HumanRequirement"
exit_transitions:
  - from_node: "HumanApproval"
    event_type: "RequirementApproved"
    to_node: "HumanRequirement"
  - from_node: "HumanApproval"
    event_type: "SpecificationRejected"
    to_node: "HumanRequirement"
nodes:
  - name: "HumanRequirement"
    type: "human"
    description: "Human provides initial requirements"
  - name: "Architect"
    type: "agent"
    agent_role: "Architect"
    description: "AI architect drafts logic specification"
  - name: "ArchitectReviewer"
    type: "agent"
    agent_role: "ArchitectReviewer"
    description: "AI reviewer evaluates the specification"
  - name: "HumanApproval"
    type: "human"
    description: "Human reviews and approves or rejects the specification"
transitions:
  - from_node: "HumanRequirement"
    event_type: "RequirementProvided"
    to_node: "Architect"
  - from_node: "Architect"
    event_type: "DraftCompleted"
    to_node: "ArchitectReviewer"
  - from_node: "ArchitectReviewer"
    event_type: "ReviewPassed"
    to_node: "HumanApproval"
  - from_node: "ArchitectReviewer"
    event_type: "RevisionNeeded"
    to_node: "Architect"
  - from_node: "HumanApproval"
    event_type: "RequirementApproved"
    to_node: "HumanRequirement"
  - from_node: "HumanApproval"
    event_type: "SpecificationRejected"
    to_node: "HumanRequirement"
```

## Invariants

1. **Name Uniqueness**: No two workflows in `.spectra/workflows/` may have the same `Name`. The filesystem enforces this constraint, but the runtime must also validate name uniqueness when loading multiple workflows.

2. **Name Format**: `Name` must follow Go PascalCase naming conventions (no spaces, underscores, or special characters).

3. **Entry Node Validity**: `EntryNode` must reference a `Node.Name` that exists in `Nodes`.

4. **Entry Node Type Constraint**: The node referenced by `EntryNode` must have `type == "human"`. This is validated during workflow definition load.

5. **Exit Transitions Non-Empty**: `ExitTransitions` must contain at least one element. A workflow without exit transitions cannot complete successfully. Empty arrays (`[]`) are treated as validation errors equivalent to missing the field entirely, as they represent an invalid workflow configuration.

6. **Exit Transitions Integrity**: All transitions in `ExitTransitions` must reference valid transitions defined in `Transitions` (matching `from_node`, `event_type`, and `to_node` exactly).

7. **Nodes Non-Empty**: `Nodes` must contain at least one element. A workflow without nodes is meaningless. Empty arrays (`[]`) are treated as validation errors equivalent to missing the field entirely.

8. **Transitions Non-Empty**: `Transitions` must contain at least one element. A workflow without transitions cannot progress. Empty arrays (`[]`) are treated as validation errors equivalent to missing the field entirely.

9. **Exit Transitions Target Constraint**: For each transition in `ExitTransitions`, the `to_node` must reference a node with `type == "human"`. This is validated during workflow definition load.

10. **Node Name Uniqueness**: All `Node.Name` values in `Nodes` must be unique within the workflow.

11. **Transition Integrity**: All `Transition.FromNode` and `Transition.ToNode` must reference valid `Node.Name` values in `Nodes`.

12. **Non-Exit Node Outgoing Transition**: Every node that is not targeted by any exit transition must have at least one outgoing transition (at least one `Transition` where `FromNode` equals the node's name). Nodes targeted by exit transitions (i.e., `to_node` of transitions defined in `ExitTransitions`) are exempt from this requirement.

13. **Event Type Validation**: All `Transition.EventType` values must be valid Event types (PascalCase, non-empty, workflow-specific).

14. **Built-in Workflow Copy Behavior**: During `spectra init`, built-in workflows are copied to `.spectra/workflows/` only if a file with the same name does not already exist. Existing files are skipped without error.

## Edge Cases

- **Condition**: Workflow file `.spectra/workflows/<name>.yaml` does not exist.
  **Expected**: Runtime returns "workflow not found" error.

- **Condition**: `Name` contains spaces or special characters (e.g., `"Simple-SDD"`, `"simple_sdd"`).
  **Expected**: Runtime rejects the workflow definition with an error: "workflow name must be PascalCase with no spaces or special characters".

- **Condition**: `EntryNode` references a non-existent node.
  **Expected**: Runtime rejects the workflow definition with a validation error during load.

- **Condition**: `EntryNode` references a node with `type == "agent"`.
  **Expected**: Runtime rejects the workflow definition with an error: "entry node '<name>' must have type 'human'".

- **Condition**: `ExitTransitions` is an empty array.
  **Expected**: Runtime rejects the workflow definition with an error: "at least one exit transition required".

- **Condition**: A transition in `ExitTransitions` does not match any transition defined in `Transitions`.
  **Expected**: Runtime rejects the workflow definition with an error: "exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition definition".

- **Condition**: A transition in `ExitTransitions` has a `to_node` with `type == "agent"`.
  **Expected**: Runtime rejects the workflow definition with an error: "exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') must target a human node, but targets '<node>' with type 'agent'".

- **Condition**: A non-entry node is unreachable (no incoming transitions).
  **Expected**: Runtime rejects the workflow definition with an error: "node '<name>' is unreachable (no incoming transitions)".

- **Condition**: A node not targeted by any exit transition has no outgoing transitions.
  **Expected**: Runtime rejects the workflow definition with an error: "node '<name>' has no outgoing transitions and is not an exit target".

- **Condition**: A node targeted by an exit transition has outgoing transitions defined.
  **Expected**: Runtime allows this configuration but issues a warning: "exit target node '<name>' has outgoing transitions that will never be used".

- **Condition**: Two workflows have the same `Name`.
  **Expected**: The filesystem enforces uniqueness (same filename). If the runtime attempts to load multiple workflows with the same name from different sources, it returns an error: "workflow '<name>' already exists".

- **Condition**: Session receives an event with no matching transition from the current node.
  **Expected**: EventProcessor records the event in EventHistory (audit trail) and returns a RuntimeResponse with `status="error"`. The session remains in `running` status; the caller may retry with a different event.

- **Condition**: An exit transition is traversed (event matching an exit transition is emitted from the corresponding `from_node`).
  **Expected**: Runtime transitions `CurrentState` to the `to_node` specified by the exit transition, then immediately sets `Status` to `"completed"`. The target node does not execute any actions.

- **Condition**: Workflow YAML is malformed or missing required fields.
  **Expected**: Runtime rejects the workflow with a parse error indicating the specific issue.

- **Condition**: During `spectra init`, a built-in workflow file already exists in `.spectra/workflows/`.
  **Expected**: Runtime skips copying that file and proceeds without error. Existing user-defined workflows are preserved.

## Related

- [Node](./node.md) - Individual nodes in the workflow
- [Transition](./transition.md) - Event-driven state transitions between nodes
- [ExitTransition](./exit_transition.md) - Transitions that trigger workflow completion
- [Session](../entities/session/session.md) - Runtime instance of a workflow
- [Event](../entities/event.md) - Events drive transitions in workflows
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
