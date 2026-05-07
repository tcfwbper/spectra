# Test Specification: `transition_evaluator_test.go`

## Source File Under Test

`runtime/transition_evaluator.go`

## Test File

`runtime/transition_evaluator_test.go`

---

## `TransitionEvaluator`

### Happy Path — EvaluateTransition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvaluateTransition_RegularTransition` | `unit` | Returns matching transition with IsExitTransition=false for a regular transition. | Create mock WorkflowDefinition with Transitions() returning one transition: FromNode="A", EventType="done", ToNode="B". ExitTransitions() returns empty slice. | `EvaluateTransition(wfDef, "A", "done")` | Returns `(transition, false)` where transition.ToNode()=="B" |
| `TestEvaluateTransition_ExitTransition` | `unit` | Returns matching transition with IsExitTransition=true for an exit transition. | Create mock WorkflowDefinition with Transitions() returning one transition: FromNode="B", EventType="complete", ToNode="End". ExitTransitions() returns one entry: FromNode="B", EventType="complete", ToNode="End". | `EvaluateTransition(wfDef, "B", "complete")` | Returns `(transition, true)` where transition.ToNode()=="End" |
| `TestEvaluateTransition_NoMatch` | `unit` | Returns nil and false when no transition matches the state+event pair. | Create mock WorkflowDefinition with Transitions() returning one transition: FromNode="A", EventType="done", ToNode="B". ExitTransitions() returns empty slice. | `EvaluateTransition(wfDef, "A", "error")` | Returns `(nil, false)` |
| `TestEvaluateTransition_MultipleTransitions_CorrectMatch` | `unit` | Selects the correct transition from multiple entries. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"], [FromNode="A", EventType="error", ToNode="C"], [FromNode="B", EventType="done", ToNode="D"]. ExitTransitions() returns empty slice. | `EvaluateTransition(wfDef, "A", "error")` | Returns `(transition, false)` where transition.ToNode()=="C" |
| `TestEvaluateTransition_MultipleExitTransitions` | `unit` | Correctly identifies exit transition among multiple exit entries. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="X", EventType="done", ToNode="Y"], [FromNode="Y", EventType="finish", ToNode="End"]. ExitTransitions() returns: [FromNode="Y", EventType="finish", ToNode="End"]. | `EvaluateTransition(wfDef, "Y", "finish")` | Returns `(transition, true)` where transition.ToNode()=="End" |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvaluateTransition_EmptyCurrentState` | `unit` | Returns nil and false when currentState is empty string. | Create mock WorkflowDefinition with Transitions() returning one transition: FromNode="A", EventType="done", ToNode="B". | `EvaluateTransition(wfDef, "", "done")` | Returns `(nil, false)` |
| `TestEvaluateTransition_EmptyEventType` | `unit` | Returns nil and false when eventType is empty string. | Create mock WorkflowDefinition with Transitions() returning one transition: FromNode="A", EventType="done", ToNode="B". | `EvaluateTransition(wfDef, "A", "")` | Returns `(nil, false)` |
| `TestEvaluateTransition_EmptyTransitionsList` | `unit` | Returns nil and false when WorkflowDefinition has no transitions. | Create mock WorkflowDefinition with Transitions() returning empty slice. ExitTransitions() returns empty slice. | `EvaluateTransition(wfDef, "A", "done")` | Returns `(nil, false)` |

### Boundary Values — ExitTransition Classification

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvaluateTransition_PartialExitMatch_DifferentToNode` | `unit` | Does not classify as exit when ExitTransition has same FromNode and EventType but different ToNode. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"]. ExitTransitions() returns: [FromNode="A", EventType="done", ToNode="C"]. | `EvaluateTransition(wfDef, "A", "done")` | Returns `(transition, false)` where transition.ToNode()=="B" |
| `TestEvaluateTransition_PartialExitMatch_DifferentEventType` | `unit` | Does not classify as exit when ExitTransition has same FromNode and ToNode but different EventType. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"]. ExitTransitions() returns: [FromNode="A", EventType="error", ToNode="B"]. | `EvaluateTransition(wfDef, "A", "done")` | Returns `(transition, false)` where transition.ToNode()=="B" |
| `TestEvaluateTransition_PartialExitMatch_DifferentFromNode` | `unit` | Does not classify as exit when ExitTransition has same EventType and ToNode but different FromNode. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"]. ExitTransitions() returns: [FromNode="X", EventType="done", ToNode="B"]. | `EvaluateTransition(wfDef, "A", "done")` | Returns `(transition, false)` where transition.ToNode()=="B" |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvaluateTransition_RepeatedCalls_SameResult` | `unit` | Produces identical output on repeated calls with same inputs. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"]. ExitTransitions() returns empty slice. | Call `EvaluateTransition(wfDef, "A", "done")` twice | Both calls return `(transition, false)` with transition.ToNode()=="B"; results are identical |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestEvaluateTransition_DoesNotModifyWorkflowDefinition` | `unit` | Does not modify WorkflowDefinition or any of its contained objects. | Create mock WorkflowDefinition; assert no setter or mutator methods are called on WorkflowDefinition, Transition, or ExitTransition mocks. | `EvaluateTransition(wfDef, "A", "done")` | No mutator calls recorded on any mock |
| `TestEvaluateTransition_InvalidCurrentState_NoError` | `unit` | Does not error or panic for a currentState that is not a known node. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"]. ExitTransitions() returns empty slice. | `EvaluateTransition(wfDef, "NonExistentNode", "done")` | Returns `(nil, false)`; no panic |
| `TestEvaluateTransition_InvalidEventType_NoError` | `unit` | Does not error or panic for an eventType that is not defined in workflow. | Create mock WorkflowDefinition with Transitions() returning: [FromNode="A", EventType="done", ToNode="B"]. ExitTransitions() returns empty slice. | `EvaluateTransition(wfDef, "A", "undefinedEvent")` | Returns `(nil, false)`; no panic |
