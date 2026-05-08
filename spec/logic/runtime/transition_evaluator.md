# TransitionEvaluator

## Overview

TransitionEvaluator is a stateless pure function that determines the next state transition in a workflow state machine. Given a WorkflowDefinition, the current state (node name), and an event type, it finds the matching transition and determines whether it is an exit transition. TransitionEvaluator does not validate session state, modify any state, or perform semantic checks — it only performs transition lookup against the workflow definition's pre-validated transition graph.

## Boundaries

- Owns: transition lookup by (currentState, eventType) pair against WorkflowDefinition.Transitions.
- Owns: exit transition classification by checking WorkflowDefinition.ExitTransitions.
- Delegates: session status validation (e.g., session must be "running") to the caller (EventProcessor).
- Delegates: currentState validity verification to the caller.
- Delegates: eventType semantic validation (e.g., whether the event type is defined in the workflow) to the caller.
- Delegates: state mutation, event history recording, and agent invocation to the caller.
- Must not: modify WorkflowDefinition, session state, or any global state.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: maintain internal state between invocations.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `WorkflowDefinition` | Configuration source | Read Transitions and ExitTransitions via getter methods | Must not modify any field |
| `Transition` | Data object | Read FromNode, EventType, ToNode via getter methods | Must not modify |
| `ExitTransition` | Data object | Read FromNode, EventType, ToNode via getter methods | Must not modify |

No construction constraint: TransitionEvaluator is a stateless function (package-level function or method on a zero-state struct). No constructor is needed.

## Behavior

1. Receives WorkflowDefinition, currentState (string), and eventType (string) as parameters.
2. Iterates through WorkflowDefinition.Transitions() to find a transition where `FromNode() == currentState` and `EventType() == eventType`.
3. If no matching transition is found, returns `(nil, false)` — indicating no valid transition exists for this state+event combination.
4. If a matching transition is found, checks whether it is an exit transition by searching WorkflowDefinition.ExitTransitions() for an entry where all three fields match: `FromNode() == transition.FromNode()`, `EventType() == transition.EventType()`, `ToNode() == transition.ToNode()`.
5. If the transition is found in ExitTransitions, returns `(transition, true)` — indicating this is an exit transition that triggers workflow completion.
6. If the transition is not found in ExitTransitions, returns `(transition, false)` — indicating this is a regular transition.
7. Assumes WorkflowDefinition has been validated at construction time: no duplicate (FromNode, EventType) pairs exist in Transitions. If duplicates exist (bypassed validation), behavior is undefined (may return first match).

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowDefinition | WorkflowDefinition reference | Valid, fully validated, already-constructed WorkflowDefinition | Yes |
| CurrentState | string | Non-empty; should be a valid node name in the workflow (not validated by TransitionEvaluator) | Yes |
| EventType | string | Non-empty; should be a workflow-defined event type (not validated by TransitionEvaluator) | Yes |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| Transition | *Transition or nil | The matching transition, or nil if no match found |
| IsExitTransition | bool | `true` if the transition is an exit transition; `false` otherwise (also `false` when Transition is nil) |

### Return Cases

| Case | Transition | IsExitTransition | Meaning |
|------|-----------|-----------------|---------|
| Regular transition found | non-nil | false | Workflow advances to the transition's ToNode |
| Exit transition found | non-nil | true | Workflow should complete after this transition |
| No matching transition | nil | false | No valid transition for this state+event pair |

No error is returned. The "no match" case is communicated via nil Transition.

## Invariants

1. **Stateless**: Must not maintain any internal state between invocations. All necessary information is passed as parameters.

2. **Pure Function**: Must not modify WorkflowDefinition, session state, or any external state. Only reads input and returns values.

3. **No Session Validation**: Does not validate session status, currentState validity, or eventType validity. These are the caller's responsibility.

4. **No Error Return**: Never returns an error. The "no match" case is represented by `(nil, false)`, not an error value.

5. **Deterministic**: Given the same inputs (WorkflowDefinition, currentState, eventType), always produces the same output. This follows from WorkflowDefinition's invariant that no duplicate (FromNode, EventType) pairs exist.

6. **Exit Transition Classification**: A transition is classified as an exit transition if and only if all three fields (FromNode, EventType, ToNode) match an entry in WorkflowDefinition.ExitTransitions(). Partial matches do not count.

## Edge Cases

- Condition: No transition in WorkflowDefinition.Transitions() matches `FromNode == currentState` and `EventType == eventType`.
  Expected: Returns `(nil, false)`. Caller decides how to handle (e.g., return error to agent, log warning).

- Condition: A matching transition is found but is not in ExitTransitions.
  Expected: Returns `(transition, false)` — a regular transition.

- Condition: A matching transition is found and is also in ExitTransitions.
  Expected: Returns `(transition, true)` — an exit transition.

- Condition: `currentState` does not correspond to any node in the workflow (stale or invalid state).
  Expected: Lookup proceeds normally. If no transition matches, returns `(nil, false)`. TransitionEvaluator does not validate node existence.

- Condition: `eventType` is not a workflow-defined event type.
  Expected: Lookup proceeds normally. If no transition matches, returns `(nil, false)`. TransitionEvaluator does not validate event type existence.

- Condition: `currentState` or `eventType` is an empty string.
  Expected: Lookup proceeds normally. If no transition matches (highly likely for empty strings), returns `(nil, false)`.

- Condition: WorkflowDefinition contains duplicate (FromNode, EventType) transitions (violated construction validation).
  Expected: Behavior is undefined. May return the first match encountered. This case should never occur with properly constructed WorkflowDefinitions.

- Condition: An ExitTransition in WorkflowDefinition partially matches the found transition (same FromNode and EventType, different ToNode).
  Expected: Not considered a match. Exit transition classification requires all three fields to match exactly.

- Condition: WorkflowDefinition.ExitTransitions() contains an entry that has no corresponding entry in Transitions() (violated construction validation).
  Expected: This entry is never matched because the transition lookup in step 2 would not find it. No impact on TransitionEvaluator behavior.

## Related

- [WorkflowDefinition](../components/workflow_definition.md) — provides the transition graph and exit transition list
- [Transition](../components/transition.md) — the value object returned when a match is found
- [ExitTransition](../components/exit_transition.md) — used for exit transition classification
- [EventProcessor](./event_processor.md) — caller that invokes TransitionEvaluator after receiving an event
