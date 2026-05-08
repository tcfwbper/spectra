# ExitTransition

## Overview

An ExitTransition identifies a specific transition triple (FromNode, EventType, ToNode) that triggers workflow completion when traversed. It is a pure immutable value object that validates its own field formats at construction time. ExitTransition does not verify that a corresponding Transition exists, that the ToNode is a human-type node, or any other cross-component constraint — those are owned by the workflow-level aggregator.

## Boundaries

- Owns: construction-time validation of all fields (FromNode, EventType, ToNode).
- Owns: immutability guarantee for all fields after construction.
- Delegates: verification that a corresponding Transition exists (exact match on all three fields) to the workflow-level aggregator.
- Delegates: verification that ToNode references a human-type node to the workflow-level aggregator.
- Delegates: deduplication of ExitTransitions within a workflow to the workflow-level aggregator.
- Delegates: runtime completion behavior (marking session as completed, skipping node dispatch) to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `components` package.
- Must not: be constructed via struct literal — must use the provided constructor.

## Dependencies

None. This value object depends only on Go standard library types.

Construction constraint: Must be constructed via `NewExitTransition(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewExitTransition(fromNode string, eventType string, toNode string) (*ExitTransition, error)` that validates all fields and returns an immutable ExitTransition value.
2. Validates that `FromNode` is a non-empty PascalCase string.
3. Validates that `EventType` is a non-empty PascalCase string.
4. Validates that `ToNode` is a non-empty PascalCase string.
5. Returns a validation error if any constraint is violated. No ExitTransition is created on failure.
6. All fields are immutable after construction.
7. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| FromNode | string | Non-empty, PascalCase | Yes |
| EventType | string | Non-empty, PascalCase | Yes |
| ToNode | string | Non-empty, PascalCase | Yes |

## Outputs

### ExitTransition Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| FromNode | string | Non-empty, PascalCase | Source node of the exit transition |
| EventType | string | Non-empty, PascalCase | Event type that triggers workflow completion |
| ToNode | string | Non-empty, PascalCase | Destination node (workflow terminates here) |

### Error Output

| Condition | Error |
|-----------|-------|
| FromNode is empty | `"from_node cannot be empty"` |
| FromNode is not PascalCase | `"from_node must be PascalCase (start with uppercase, alphanumeric only)"` |
| EventType is empty | `"event_type cannot be empty"` |
| EventType is not PascalCase | `"event_type must be PascalCase (start with uppercase, alphanumeric only)"` |
| ToNode is empty | `"to_node cannot be empty"` |
| ToNode is not PascalCase | `"to_node must be PascalCase (start with uppercase, alphanumeric only)"` |

## Invariants

1. **FromNode PascalCase**: `FromNode` is always a non-empty PascalCase string after construction.
2. **EventType PascalCase**: `EventType` is always a non-empty PascalCase string after construction.
3. **ToNode PascalCase**: `ToNode` is always a non-empty PascalCase string after construction.
4. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
5. **Construction Only Via Constructor**: Must be constructed via `NewExitTransition`. Direct struct literal construction is forbidden.
6. **No Self-Loop Enforcement at This Level**: Unlike Transition, ExitTransition does not enforce FromNode != ToNode at its own level. This is because ExitTransition must correspond to an existing Transition (which already enforces no self-loop). The workflow-level aggregator validates this correspondence.

## Edge Cases

- Condition: `FromNode` is an empty string.
  Expected: Constructor returns error `"from_node cannot be empty"`. No ExitTransition is created.

- Condition: `FromNode` starts with a lowercase letter (e.g., `"humanApproval"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `EventType` is an empty string.
  Expected: Constructor returns error `"event_type cannot be empty"`. No ExitTransition is created.

- Condition: `EventType` contains special characters (e.g., `"Requirement-Approved"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `ToNode` is an empty string.
  Expected: Constructor returns error `"to_node cannot be empty"`. No ExitTransition is created.

- Condition: `FromNode == ToNode` (e.g., both are `"HumanApproval"`).
  Expected: Constructor accepts this. Self-loop prohibition is not enforced at the ExitTransition level; it is guaranteed by the corresponding Transition's invariant, validated by the workflow-level aggregator.

- Condition: All fields are valid PascalCase strings.
  Expected: Constructor returns a valid immutable ExitTransition.

## Related

- [Transition](./transition.md) — ExitTransitions must correspond to an existing Transition (validated by workflow-level aggregator)
- [Node](./node.md) — ExitTransitions reference Node names
- [Session](../entities/session/session.md) — Runtime marks session as completed when an ExitTransition is traversed
