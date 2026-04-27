# TransitionEvaluator

## Overview

TransitionEvaluator is a stateless function that determines the next state transition in a workflow state machine. Given a WorkflowDefinition, the current state (node name), and an event type, it finds the matching transition, determines whether the transition is an exit transition (triggers workflow completion), or returns nil if no matching transition exists. TransitionEvaluator does not validate session state or perform semantic checks; it only performs transition lookup based on the workflow definition. All session state validation is the caller's responsibility (typically EventProcessor).

## Behavior

1. TransitionEvaluator is invoked by EventProcessor after an event is received and basic validation (session status, event structure) has passed.
2. TransitionEvaluator receives three parameters: WorkflowDefinition, currentState (node name), and eventType.
3. TransitionEvaluator searches the WorkflowDefinition.Transitions array for a transition where `from_node == currentState` and `event_type == eventType`.
4. If exactly one matching transition is found, TransitionEvaluator proceeds to step 5. If no matching transition is found, TransitionEvaluator returns `(nil, false, nil)` to indicate "no matching transition".
5. TransitionEvaluator checks if the matching transition is an exit transition by searching WorkflowDefinition.ExitTransitions for a match with the same `from_node`, `event_type`, and `to_node`.
6. If the transition is found in ExitTransitions, TransitionEvaluator returns `(transition, true, nil)` to indicate "this is an exit transition".
7. If the transition is not found in ExitTransitions, TransitionEvaluator returns `(transition, false, nil)` to indicate "this is a regular transition".
8. TransitionEvaluator assumes the WorkflowDefinition has already been validated during loading. Specifically, it assumes there are no duplicate transitions (same `from_node` and `event_type` with different `to_node`) because such conflicts are rejected at workflow load time.
9. TransitionEvaluator is a pure function with no side effects. It does not modify the WorkflowDefinition, session state, or any global state.
10. TransitionEvaluator does not validate whether the session is in "running" status, whether the currentState is valid, or whether the eventType is workflow-defined. These validations are the caller's responsibility.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowDefinition | WorkflowDefinition | Valid, fully validated WorkflowDefinition loaded from storage | Yes |
| CurrentState | string | Non-empty, should be a workflow-defined node name (validation not performed by TransitionEvaluator) | Yes |
| EventType | string | Non-empty, should be a workflow-defined event type (validation not performed by TransitionEvaluator) | Yes |

## Outputs

### Success Cases

**Case 1: Regular Transition Found**

| Field | Type | Description |
|-------|------|-------------|
| Transition | *Transition | Pointer to the matching transition from WorkflowDefinition.Transitions |
| IsExitTransition | bool | `false` (this is a regular transition, not an exit transition) |
| Error | error | `nil` |

**Case 2: Exit Transition Found**

| Field | Type | Description |
|-------|------|-------------|
| Transition | *Transition | Pointer to the matching transition from WorkflowDefinition.Transitions |
| IsExitTransition | bool | `true` (this transition triggers workflow completion) |
| Error | error | `nil` |

**Case 3: No Matching Transition**

| Field | Type | Description |
|-------|------|-------------|
| Transition | *Transition | `nil` |
| IsExitTransition | bool | `false` (meaningless when Transition is nil) |
| Error | error | `nil` |

### Return Signature (Go-like pseudocode)

```
func EvaluateTransition(
    workflowDef WorkflowDefinition,
    currentState string,
    eventType string,
) (*Transition, bool, error)
```

**Return values**:
- `(*Transition, bool, nil)` - Found a matching transition. The bool indicates whether it's an exit transition.
- `(nil, false, nil)` - No matching transition found.

## Invariants

1. **Stateless Function**: TransitionEvaluator must not maintain any internal state between invocations. All necessary information is passed as parameters.

2. **Pure Function**: TransitionEvaluator must not modify the WorkflowDefinition, session state, or any global state. It only reads and returns values.

3. **No Session Validation**: TransitionEvaluator does not validate session status, currentState validity, or eventType validity. The caller is responsible for these validations.

4. **No Duplicate Transitions**: TransitionEvaluator assumes the WorkflowDefinition has been validated and does not contain duplicate transitions (same `from_node` and `event_type`). If duplicates exist, TransitionEvaluator's behavior is undefined (may return the first match).

5. **Exit Transition Lookup**: TransitionEvaluator determines exit transitions by checking if the matching transition exists in WorkflowDefinition.ExitTransitions (matching all three fields: `from_node`, `event_type`, `to_node`).

6. **No Error Return**: TransitionEvaluator never returns an error. It returns `nil` for the transition when no match is found, allowing the caller to decide how to handle the "no match" case.

## Edge Cases

- **Condition**: No transition in WorkflowDefinition.Transitions matches `from_node == currentState` and `event_type == eventType`.
  **Expected**: TransitionEvaluator returns `(nil, false, nil)`. The caller (EventProcessor) should interpret this as "no valid transition" and respond with an error (typically exit code 3: "no transition found").

- **Condition**: A matching transition is found in WorkflowDefinition.Transitions, but it is not listed in WorkflowDefinition.ExitTransitions.
  **Expected**: TransitionEvaluator returns `(transition, false, nil)` to indicate a regular transition.

- **Condition**: A matching transition is found in WorkflowDefinition.Transitions and is also listed in WorkflowDefinition.ExitTransitions.
  **Expected**: TransitionEvaluator returns `(transition, true, nil)` to indicate an exit transition.

- **Condition**: WorkflowDefinition.Transitions contains multiple transitions with the same `from_node` and `event_type` but different `to_node` (this violates workflow validation, but may occur if validation is bypassed).
  **Expected**: TransitionEvaluator's behavior is undefined. It may return the first matching transition encountered. Workflow validation should prevent this case from occurring.

- **Condition**: currentState is not a valid node name in the workflow (e.g., typo, stale session state).
  **Expected**: TransitionEvaluator performs the lookup as usual. If no transition matches, it returns `(nil, false, nil)`. The caller should have validated currentState before invoking TransitionEvaluator.

- **Condition**: eventType is not a workflow-defined event type.
  **Expected**: TransitionEvaluator performs the lookup as usual. If no transition matches, it returns `(nil, false, nil)`. The caller should have validated eventType semantics before invoking TransitionEvaluator.

- **Condition**: WorkflowDefinition is `nil` or empty (this violates the input constraints).
  **Expected**: TransitionEvaluator's behavior is undefined. The caller must ensure a valid WorkflowDefinition is passed.

- **Condition**: currentState or eventType is an empty string `""`.
  **Expected**: TransitionEvaluator performs the lookup as usual. If no transition matches, it returns `(nil, false, nil)`. The caller should have validated these fields before invoking TransitionEvaluator.

- **Condition**: WorkflowDefinition.ExitTransitions contains a transition that does not exist in WorkflowDefinition.Transitions (this violates workflow validation).
  **Expected**: This case cannot occur if workflow validation is properly implemented. If it does occur, TransitionEvaluator will never match the exit transition (because the transition lookup in step 3 will fail first).

- **Condition**: A transition in WorkflowDefinition.ExitTransitions partially matches (e.g., same `from_node` and `event_type`, but different `to_node`).
  **Expected**: TransitionEvaluator does not consider this a match. Exit transitions must match all three fields exactly.

## Related

- [WorkflowDefinition](../components/workflow_definition.md) - Defines transitions and exit transitions
- [Transition](../components/transition.md) - Defines individual state transitions
- [ExitTransition](../components/exit_transition.md) - Identifies transitions that trigger workflow completion
- [Session](../entities/session/session.md) - Provides the currentState used by TransitionEvaluator
- [Event](../entities/event.md) - Provides the eventType used by TransitionEvaluator
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
