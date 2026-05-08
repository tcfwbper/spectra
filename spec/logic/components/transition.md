# Transition

## Overview

A Transition defines an event-driven edge between two nodes in a workflow. It specifies the source node (`FromNode`), the event type that triggers the transition (`EventType`), and the destination node (`ToNode`). Transition is a pure immutable value object that validates its own field formats at construction time. It does not evaluate transitions at runtime, check node existence, or enforce uniqueness across a collection — those are owned by the workflow-level aggregator and the runtime respectively.

## Boundaries

- Owns: construction-time validation of all fields (FromNode, EventType, ToNode).
- Owns: immutability guarantee for all fields after construction.
- Owns: self-loop prohibition (FromNode must differ from ToNode).
- Delegates: node existence validation (FromNode and ToNode reference valid Node.Name values) to the workflow-level aggregator.
- Delegates: transition uniqueness within a workflow (at most one transition per FromNode + EventType pair) to the workflow-level aggregator.
- Delegates: runtime transition evaluation and session state mutation to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `components` package.
- Must not: be constructed via struct literal — must use the provided constructor.

## Dependencies

None. This value object depends only on Go standard library types.

Construction constraint: Must be constructed via `NewTransition(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewTransition(fromNode string, eventType string, toNode string) (*Transition, error)` that validates all fields and returns an immutable Transition value.
2. Validates that `FromNode` is a non-empty PascalCase string.
3. Validates that `EventType` is a non-empty PascalCase string.
4. Validates that `ToNode` is a non-empty PascalCase string.
5. Validates that `FromNode != ToNode`. Self-loop transitions are prohibited.
6. Returns a validation error if any constraint is violated. No Transition is created on failure.
7. All fields are immutable after construction.
8. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| FromNode | string | Non-empty, PascalCase | Yes |
| EventType | string | Non-empty, PascalCase | Yes |
| ToNode | string | Non-empty, PascalCase | Yes |

Additional constraint: `FromNode` must differ from `ToNode`.

## Outputs

### Transition Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| FromNode | string | Non-empty, PascalCase | Source node of the transition |
| EventType | string | Non-empty, PascalCase | Event type that triggers this transition |
| ToNode | string | Non-empty, PascalCase | Destination node of the transition |

### Error Output

| Condition | Error |
|-----------|-------|
| FromNode is empty | `"from_node cannot be empty"` |
| FromNode is not PascalCase | `"from_node must be PascalCase (start with uppercase, alphanumeric only)"` |
| EventType is empty | `"event_type cannot be empty"` |
| EventType is not PascalCase | `"event_type must be PascalCase (start with uppercase, alphanumeric only)"` |
| ToNode is empty | `"to_node cannot be empty"` |
| ToNode is not PascalCase | `"to_node must be PascalCase (start with uppercase, alphanumeric only)"` |
| FromNode == ToNode | `"from_node and to_node must be different"` |

## Invariants

1. **FromNode PascalCase**: `FromNode` is always a non-empty PascalCase string after construction.
2. **EventType PascalCase**: `EventType` is always a non-empty PascalCase string after construction.
3. **ToNode PascalCase**: `ToNode` is always a non-empty PascalCase string after construction.
4. **No Self-Loop**: `FromNode != ToNode` always holds after construction.
5. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
6. **Construction Only Via Constructor**: Must be constructed via `NewTransition`. Direct struct literal construction is forbidden.

## Edge Cases

- Condition: `FromNode` is an empty string.
  Expected: Constructor returns error `"from_node cannot be empty"`. No Transition is created.

- Condition: `FromNode` starts with a lowercase letter (e.g., `"architect"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `EventType` is an empty string.
  Expected: Constructor returns error `"event_type cannot be empty"`. No Transition is created.

- Condition: `EventType` contains non-alphanumeric characters (e.g., `"Draft-Completed"`, `"draft_completed"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `ToNode` is an empty string.
  Expected: Constructor returns error `"to_node cannot be empty"`. No Transition is created.

- Condition: `FromNode == ToNode` (e.g., both are `"Architect"`).
  Expected: Constructor returns error `"from_node and to_node must be different"`. No Transition is created.

- Condition: All fields are valid and distinct.
  Expected: Constructor returns a valid immutable Transition.

## Related

- [Node](./node.md) — Transitions reference Node names (referential integrity validated by workflow-level aggregator)
- [ExitTransition](./exit_transition.md) — ExitTransitions reference Transition triples
- [Event](../entities/event.md) — Events carry the EventType that matches a Transition
- [Session](../entities/session/session.md) — Runtime uses Transitions to advance Session.CurrentState
