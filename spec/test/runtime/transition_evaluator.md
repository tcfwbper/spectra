# Test Specification: `transition_evaluator.go`

## Source File Under Test
`runtime/transition_evaluator.go`

## Test File
`runtime/transition_evaluator_test.go`

---

## `TransitionEvaluator`

### Happy Path — Regular Transition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_RegularTransition_Found` | `unit` | Finds matching regular transition in workflow definition. | Mock WorkflowDefinition with transition `from_node="A"`, `event_type="success"`, `to_node="B"` in Transitions array; transition not in ExitTransitions | `CurrentState="A"`, `EventType="success"` | Returns `(transition, false, nil)` where transition points to matched entry; `IsExitTransition=false` |
| `TestTransitionEvaluator_RegularTransition_MultipleInDefinition` | `unit` | Finds correct transition when multiple transitions exist. | Mock WorkflowDefinition with transitions: `A->B (success)`, `A->C (failure)`, `B->D (done)`; none in ExitTransitions | `CurrentState="A"`, `EventType="failure"` | Returns transition `A->C (failure)`; `IsExitTransition=false` |

### Happy Path — Exit Transition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_ExitTransition_Found` | `unit` | Finds matching exit transition and marks it correctly. | Mock WorkflowDefinition with transition `from_node="Final"`, `event_type="complete"`, `to_node="End"` in both Transitions and ExitTransitions | `CurrentState="Final"`, `EventType="complete"` | Returns `(transition, true, nil)` where transition points to matched entry; `IsExitTransition=true` |
| `TestTransitionEvaluator_ExitTransition_ExactMatch` | `unit` | Exit transition requires exact match on all three fields. | Mock WorkflowDefinition with transition `A->B (success)` in Transitions; ExitTransitions has `A->C (success)` | `CurrentState="A"`, `EventType="success"` | Returns `(transition A->B, false, nil)`; `IsExitTransition=false` (ExitTransitions entry doesn't match all fields) |

### Happy Path — No Transition Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_NoMatch_ReturnsNil` | `unit` | Returns nil when no matching transition found. | Mock WorkflowDefinition with transitions: `A->B (success)`, `B->C (next)` | `CurrentState="A"`, `EventType="failure"` | Returns `(nil, false, nil)` |
| `TestTransitionEvaluator_NoMatch_EmptyTransitions` | `unit` | Returns nil when Transitions array is empty. | Mock WorkflowDefinition with empty Transitions array | `CurrentState="Any"`, `EventType="any"` | Returns `(nil, false, nil)` |

### Happy Path — Transition Lookup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_MatchesFromNodeAndEventType` | `unit` | Matches transition based on from_node and event_type only. | Mock WorkflowDefinition with transition `from_node="Start"`, `event_type="init"`, `to_node="Process"` | `CurrentState="Start"`, `EventType="init"` | Returns matched transition regardless of `to_node` value |
| `TestTransitionEvaluator_FirstMatchReturned` | `unit` | Returns first matching transition when duplicates exist (undefined behavior). | Mock WorkflowDefinition with duplicate transitions: `A->B (go)`, `A->C (go)` (violates validation) | `CurrentState="A"`, `EventType="go"` | Verifies no panic occurs; returns first matched transition; documents current implementation behavior (undefined per spec) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_RepeatedCalls_IdenticalResults` | `unit` | Multiple calls with same inputs return identical results. | Mock WorkflowDefinition with transition `A->B (success)` | Call three times with `CurrentState="A"`, `EventType="success"` | All three calls return identical transition pointer and `IsExitTransition=false` |
| `TestTransitionEvaluator_RepeatedCalls_NoStateChange` | `unit` | Repeated calls do not modify WorkflowDefinition. | Mock WorkflowDefinition with transition `A->B (success)` | Call TransitionEvaluator 10 times | WorkflowDefinition remains unchanged; no mutations |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_SequentialTransitions` | `unit` | Evaluates sequential state transitions correctly. | Mock WorkflowDefinition with transitions: `A->B (next)`, `B->C (next)`, `C->End (done)` where `C->End` is in ExitTransitions | First call: `CurrentState="A"`, `EventType="next"`; second call: `CurrentState="B"`, `EventType="next"`; third call: `CurrentState="C"`, `EventType="done"` | First returns `A->B, false`; second returns `B->C, false`; third returns `C->End, true` |

### Validation Failures — Invalid State

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_InvalidCurrentState_NoMatch` | `unit` | Returns nil when currentState is not a valid node name. | Mock WorkflowDefinition with transitions for nodes `A`, `B`, `C` | `CurrentState="InvalidNode"`, `EventType="success"` | Returns `(nil, false, nil)` (no validation error, just no match) |
| `TestTransitionEvaluator_EmptyCurrentState_NoMatch` | `unit` | Returns nil when currentState is empty string. | Mock WorkflowDefinition with transitions: `A->B (go)` | `CurrentState=""`, `EventType="go"` | Returns `(nil, false, nil)` (performs lookup, no match) |

### Validation Failures — Invalid Event Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_InvalidEventType_NoMatch` | `unit` | Returns nil when eventType is not defined in workflow. | Mock WorkflowDefinition with transitions using event types `success`, `failure` | `CurrentState="A"`, `EventType="unknown_event"` | Returns `(nil, false, nil)` (no validation error, just no match) |
| `TestTransitionEvaluator_EmptyEventType_NoMatch` | `unit` | Returns nil when eventType is empty string. | Mock WorkflowDefinition with transitions: `A->B (success)` | `CurrentState="A"`, `EventType=""` | Returns `(nil, false, nil)` (performs lookup, no match) |

### Boundary Values — Transitions Array

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_SingleTransition` | `unit` | Handles WorkflowDefinition with single transition. | Mock WorkflowDefinition with one transition: `A->B (go)` | `CurrentState="A"`, `EventType="go"` | Returns matched transition; `IsExitTransition=false` |
| `TestTransitionEvaluator_LargeTransitionsArray` | `unit` | Efficiently handles large Transitions array. | Mock WorkflowDefinition with 1000 transitions; target transition `State500->State501 (event500)` at index 500 | `CurrentState="State500"`, `EventType="event500"` | Returns matched transition; performs lookup efficiently |

### Boundary Values — Exit Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_EmptyExitTransitions` | `unit` | Handles empty ExitTransitions array. | Mock WorkflowDefinition with transitions but empty ExitTransitions array | `CurrentState="A"`, `EventType="success"` | Returns `(transition, false, nil)` (all transitions are regular) |
| `TestTransitionEvaluator_AllTransitionsAreExit` | `unit` | Handles when all transitions are exit transitions. | Mock WorkflowDefinition where all transitions in Transitions are also in ExitTransitions | Any valid `CurrentState` and `EventType` | Returns `(transition, true, nil)` for all matches |

### Boundary Values — Edge Case States

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_CaseSensitiveNodeNames` | `unit` | Node names are case-sensitive. | Mock WorkflowDefinition with transitions: `Start->B (go)`, `start->C (go)` | `CurrentState="Start"`, `EventType="go"` | Returns `Start->B` (not `start->C`); case matters |
| `TestTransitionEvaluator_CaseSensitiveEventTypes` | `unit` | Event types are case-sensitive. | Mock WorkflowDefinition with transitions: `A->B (Success)`, `A->C (success)` | `CurrentState="A"`, `EventType="success"` | Returns `A->C` (not `A->B`); case matters |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_WorkflowDefinitionUnmodified` | `unit` | WorkflowDefinition is not modified by evaluation. | Mock WorkflowDefinition with transitions; capture copy before evaluation | Call TransitionEvaluator | WorkflowDefinition matches original copy exactly; no mutations |
| `TestTransitionEvaluator_NoGlobalState` | `unit` | Function maintains no internal state between calls. | Mock WorkflowDefinition | First call: `CurrentState="A"`, `EventType="go"`; second call: `CurrentState="B"`, `EventType="stop"` | Second call does not reference or depend on first call's inputs or results |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_ReturnsNilNotError` | `unit` | Returns nil transition (not error) when no match found. | Mock WorkflowDefinition with transitions | `CurrentState="X"`, `EventType="nonexistent"` | Returns `(nil, false, nil)` where error field is `nil` (not an error condition) |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_AccessesTransitionsArray` | `unit` | Reads Transitions array from WorkflowDefinition. | Mock WorkflowDefinition that tracks field access | `CurrentState="A"`, `EventType="success"` | `WorkflowDefinition.Transitions` is accessed for lookup |
| `TestTransitionEvaluator_AccessesExitTransitionsArray` | `unit` | Reads ExitTransitions array when checking exit status. | Mock WorkflowDefinition that tracks field access; transition found in Transitions | `CurrentState="A"`, `EventType="success"` | `WorkflowDefinition.ExitTransitions` is accessed after finding match in Transitions |
| `TestTransitionEvaluator_NoSessionAccess` | `unit` | Does not access Session object (stateless). | Mock WorkflowDefinition; mock Session object that tracks access | `CurrentState="A"`, `EventType="success"` | Session object is never accessed; function is stateless |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_ConcurrentCalls` | `race` | Multiple goroutines call TransitionEvaluator concurrently. | Mock WorkflowDefinition with transitions; 100 goroutines | Each goroutine calls with various `CurrentState` and `EventType` values | All calls succeed; no data races; WorkflowDefinition unchanged |
| `TestTransitionEvaluator_ConcurrentReadsSameTransition` | `race` | Multiple goroutines read same transition concurrently. | Mock WorkflowDefinition with transition `A->B (go)`; 50 goroutines | All goroutines call with `CurrentState="A"`, `EventType="go"` | All calls return same transition pointer; no races; all return identical results |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_NeverReturnsError` | `unit` | Function never returns error (always returns nil error). | Mock WorkflowDefinition with any configuration; test various invalid inputs | `CurrentState` and `EventType` with various valid/invalid values | All calls return `error=nil` (third return value) |

### Atomic Replacement

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_ReturnedTransitionIsReference` | `unit` | Returned transition is a reference to the original in WorkflowDefinition. | Mock WorkflowDefinition with transition `A->B (go)` | `CurrentState="A"`, `EventType="go"` | Returned transition pointer points to same memory as original in Transitions array |

### Happy Path — Exit Transition Matching

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_ExitTransition_AllFieldsMatch` | `unit` | Exit transition matches only when all three fields match exactly. | Mock WorkflowDefinition with transition `from_node="End"`, `event_type="complete"`, `to_node="Exit"` in Transitions; ExitTransitions has same entry | `CurrentState="End"`, `EventType="complete"` | Returns `(transition, true, nil)` where all fields (`from_node`, `event_type`, `to_node`) match ExitTransitions entry |
| `TestTransitionEvaluator_ExitTransition_PartialMatch_NotExit` | `unit` | Partial match in ExitTransitions does not mark as exit. | Mock WorkflowDefinition with transition `A->B (success)` in Transitions; ExitTransitions has `A->B (failure)` (different event_type) | `CurrentState="A"`, `EventType="success"` | Returns `(transition A->B success, false, nil)`; not marked as exit (partial match doesn't count) |

### Boundary Values — Nil Inputs

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransitionEvaluator_WorkflowDefinitionNil_Undefined` | `unit` | Behavior undefined when WorkflowDefinition is nil (documented). | `WorkflowDefinition=nil` | `CurrentState="A"`, `EventType="go"` | Behavior undefined; may panic or return nil (caller must pass valid WorkflowDefinition) |
