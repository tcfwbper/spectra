# Test Specification: `exit_transition.go`

## Source File Under Test
`components/exit_transition.go`

## Test File
`components/exit_transition_test.go`

---

## `ExitTransition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_ValidConstruction` | `unit` | Creates ExitTransition matching existing transition to human node. | Workflow with nodes "HumanApproval" (human), "HumanRequirement" (human); transition from "HumanApproval" on "RequirementApproved" to "HumanRequirement" | `FromNode="HumanApproval"`, `EventType="RequirementApproved"`, `ToNode="HumanRequirement"` | Returns valid ExitTransition; all fields match input |
| `TestExitTransition_MultipleExitTransitions` | `unit` | Multiple ExitTransitions coexist in workflow. | Workflow with transitions: "HumanApproval" → "HumanRequirement" on "RequirementApproved" and "HumanApproval" → "HumanRequirement" on "SpecificationRejected" (both to human nodes) | Add both as ExitTransitions | Both ExitTransitions valid; workflow validation succeeds |
| `TestExitTransition_DifferentNodesMultipleExits` | `unit` | ExitTransitions from different nodes. | Workflow with transitions: "HumanApproval" → "HumanRequirement" on "Approved" and "QualityGate" → "HumanReview" on "CriticalFailure" (both to human nodes) | Add both as ExitTransitions | Both ExitTransitions valid; workflow validation succeeds |

### Validation Failures — Transition Existence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_NoMatchingTransition` | `unit` | Rejects ExitTransition with no corresponding transition. | Workflow with transition from "A" to "B" on "Event1" only | ExitTransition: `FromNode="A"`, `EventType="Event2"`, `ToNode="B"` | Returns error; error message matches `/exit transition.*Event2.*no.*corresponding.*transition/i` |
| `TestExitTransition_PartialMatchFromNode` | `unit` | Rejects ExitTransition where FromNode and EventType match but ToNode differs. | Workflow with transition from "A" to "B" on "Event" | ExitTransition: `FromNode="A"`, `EventType="Event"`, `ToNode="C"` | Returns error; error message matches `/exit transition.*no.*corresponding.*transition/i` |
| `TestExitTransition_PartialMatchEventType` | `unit` | Rejects ExitTransition where FromNode and ToNode match but EventType differs. | Workflow with transition from "A" to "B" on "Event1" | ExitTransition: `FromNode="A"`, `EventType="Event2"`, `ToNode="B"` | Returns error; error message matches `/exit transition.*no.*corresponding.*transition/i` |

### Validation Failures — Target Node Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_ToAgentNode` | `unit` | Rejects ExitTransition with ToNode referencing agent node. | Workflow with nodes "HumanApproval" (human), "ArchitectReviewer" (agent); transition from "HumanApproval" to "ArchitectReviewer" on "Approved" | ExitTransition: `FromNode="HumanApproval"`, `EventType="Approved"`, `ToNode="ArchitectReviewer"` | Returns error; error message matches `/exit transition.*must target.*human.*node.*ArchitectReviewer.*agent/i` |

### Validation Failures — Node References

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_FromNodeNonExistent` | `unit` | Rejects ExitTransition with non-existent FromNode. | Workflow with node "Reviewer" only | `FromNode="NonExistent"`, `EventType="Event"`, `ToNode="Reviewer"` | Returns error; error message matches `/exit transition.*non-existent.*node.*NonExistent/i` |
| `TestExitTransition_ToNodeNonExistent` | `unit` | Rejects ExitTransition with non-existent ToNode. | Workflow with node "Approval" only | `FromNode="Approval"`, `EventType="Event"`, `ToNode="NonExistent"` | Returns error; error message matches `/exit transition.*non-existent.*node.*NonExistent/i` |

### Validation Failures — EventType

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_EmptyEventType` | `unit` | Rejects ExitTransition with empty EventType. | Workflow with nodes "A" (human) and "B" (human); transition from "A" to "B" on "Event" | `FromNode="A"`, `EventType=""`, `ToNode="B"` | Returns error; error message matches `/event_type.*non-empty/i` |
| `TestExitTransition_EventTypeWithSpaces` | `unit` | Rejects ExitTransition with EventType containing spaces. | Workflow with nodes "A" (human) and "B" (human) | `FromNode="A"`, `EventType="Task Done"`, `ToNode="B"` | Returns error; error message matches `/event_type.*PascalCase.*spaces/i` |
| `TestExitTransition_EventTypeNotPascalCase` | `unit` | Rejects ExitTransition with EventType not in PascalCase. | Workflow with nodes "A" (human) and "B" (human) | `FromNode="A"`, `EventType="task_done"`, `ToNode="B"` | Returns error; error message matches `/event_type.*PascalCase/i` |

### Validation Failures — Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_Duplicate` | `unit` | Rejects duplicate ExitTransitions with identical triples. | Workflow with existing ExitTransition: `FromNode="A"`, `EventType="Done"`, `ToNode="B"` (both human nodes) | Add second ExitTransition: `FromNode="A"`, `EventType="Done"`, `ToNode="B"` | Returns error; error message matches `/duplicate.*exit transition.*Done.*A.*B/i` |

### Validation Failures — Empty Array

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_EmptyArray` | `unit` | Rejects workflow with empty ExitTransitions array. | Workflow with nodes and transitions | `ExitTransitions` array is empty | Returns error; error message matches `/at least one.*exit transition.*required/i` |

### Happy Path — Workflow Completion

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_TriggersCompletion` | `unit` | ExitTransition triggers workflow completion when traversed. | Workflow with ExitTransition from "HumanApproval" to "HumanRequirement" on "Approved" (both human); session at `CurrentState="HumanApproval"`, `Status="running"` | Emit event: `Type="Approved"` | Session `CurrentState` transitions to "HumanRequirement"; session `Status` set to `"completed"`; session terminates |
| `TestExitTransition_StateUpdateBeforeCompletion` | `unit` | CurrentState updates to ToNode before Status changes to completed. | Workflow with ExitTransition from "A" to "B" on "Exit" (both human); session at `CurrentState="A"`, `Status="running"` | Emit event: `Type="Exit"` | In-memory state updates in sequence: first `CurrentState="B"`, then `Status="completed"` |
| `TestExitTransition_NoNodeExecution` | `unit` | Target node does not execute actions when reached via ExitTransition. | Workflow with ExitTransition to human node "Final"; human node would print to stdout on normal entry; session at source node | Emit event triggering ExitTransition | Node "Final" reached; `CurrentState="Final"`; no stdout print; session immediately completed |
| `TestExitTransition_OneTimeCompletion` | `unit` | ExitTransition triggers completion only once per session. | Workflow with ExitTransition; session at source node | Emit event triggering ExitTransition | Session transitions to completed; session terminates; cannot emit further events |

### Happy Path — Any-One Completion

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_FirstMatchCompletes` | `unit` | First matching ExitTransition triggers completion when multiple defined. | Workflow with 2 ExitTransitions from "A": on "Exit1" to "B", on "Exit2" to "C" (all human); session at `CurrentState="A"` | Emit event: `Type="Exit1"` | Session transitions to "B" and completes; "Exit2" ExitTransition not evaluated |
| `TestExitTransition_AnyExitTriggers` | `unit` | Any of multiple ExitTransitions can trigger completion. | Workflow with 2 ExitTransitions from different nodes; session at second source node | Emit event matching second ExitTransition | Session completes via second ExitTransition; first ExitTransition not required |

### Happy Path — Non-Exit Transition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_NormalTransitionContinues` | `unit` | Non-exit transition does not trigger completion. | Workflow with transition from "A" to "B" on "Event" (not in ExitTransitions); session at `CurrentState="A"` | Emit event: `Type="Event"` | Session transitions to "B"; `Status` remains `"running"`; session continues |

### Validation Failures — Wrong Node State

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_EventFromWrongNode` | `unit` | ExitTransition not triggered when session not at FromNode. | Workflow with ExitTransition from "B" to "C" on "Exit"; session at `CurrentState="A"` | Emit event: `Type="Exit"` | Returns error; error message matches `/no.*transition.*Exit.*node.*A/i`; session remains at "A"; ExitTransition not triggered |

### Happy Path — Persistence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_BestEffortPersistence` | `unit` | State persisted to disk on best-effort basis. | Temporary test directory created; workflow with ExitTransition; session files in test directory; session at source node; all file operations occur within test fixtures | Emit event triggering ExitTransition | In-memory state: `CurrentState` and `Status` updated; disk persistence attempted; in-memory state authoritative; if disk write fails, session still completed in memory |
| `TestExitTransition_SeparateWrites` | `unit` | CurrentState and Status may persist in separate write operations. | Temporary test directory created; workflow with ExitTransition; session files in test directory; session at source node; all file operations occur within test fixtures | Emit event triggering ExitTransition; introduce write delay between updates | `CurrentState` persisted first, then `Status`; separate writes allowed; in-memory state authoritative |

### Validation Failures — Unused Outgoing Transitions Warning

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_TargetNodeHasOutgoingTransitions` | `unit` | Issues warning when exit target node has outgoing transitions. | Workflow with ExitTransition to node "Final"; node "Final" (human) has outgoing transition to "NextNode" | Validate workflow | Returns warning message matching `/exit target.*Final.*outgoing transitions.*never.*used/i`; workflow not rejected; warning logged |

### Happy Path — Partial Persistence Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_PartialDiskPersistenceFailure` | `unit` | In-memory state authoritative when Status disk write fails. | Temporary test directory created; workflow with ExitTransition; session files in test directory; mock disk write failure for Status update only; all file operations occur within test fixtures | Emit event triggering ExitTransition; CurrentState disk write succeeds; Status disk write fails | In-memory: both `CurrentState` and `Status` updated correctly; session completed in memory; CurrentState persisted to disk; Status persistence failed; in-memory state authoritative; session cannot accept further events |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_FieldsImmutable` | `unit` | ExitTransition fields cannot be modified after creation. | ExitTransition instance created | Attempt to modify `FromNode`, `EventType`, or `ToNode` | Field modification attempt fails or has no effect; original values remain |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_ImplementsInterface` | `unit` | ExitTransition type implements expected interface. | | ExitTransition instance created | ExitTransition satisfies ExitTransition interface contract (GetFromNode, GetEventType, GetToNode methods) |

### Happy Path — YAML Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_ToYAML` | `unit` | ExitTransition serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | ExitTransition with `FromNode="HumanApproval"`, `EventType="RequirementApproved"`, `ToNode="HumanRequirement"` | YAML contains `from_node: "HumanApproval"`, `event_type: "RequirementApproved"`, `to_node: "HumanRequirement"` |
| `TestExitTransition_FromYAML` | `unit` | YAML deserializes to ExitTransition correctly. | Temporary test directory created; YAML file in test directory; all file operations occur within test fixtures | YAML: `from_node: "A"`, `event_type: "Exit"`, `to_node: "B"` | ExitTransition created with matching fields |
| `TestExitTransition_MultipleToYAML` | `unit` | Multiple ExitTransitions serialize to YAML array correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | Three ExitTransitions in workflow | YAML contains `exit_transitions:` array with all three; order preserved |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_ListExitTransitionsInWorkflow` | `e2e` | CLI lists all exit transitions in a workflow. | Temporary test directory created; workflow definition in test directory with 2 ExitTransitions; all file operations occur within test fixtures | Execute `spectra workflow exit-transitions list --workflow <workflow-id>` | Command succeeds; output lists both ExitTransitions with from_node, event_type, to_node |
| `TestExitTransition_ValidateWorkflowWithExitTransitions` | `e2e` | CLI validates workflow containing exit transitions. | Temporary test directory created; valid workflow definition in test directory with nodes, transitions, and ExitTransitions; all file operations occur within test fixtures | Execute `spectra workflow validate --workflow <workflow-id>` | Command succeeds; no errors reported |
| `TestExitTransition_WorkflowCompletesViaExit` | `e2e` | End-to-end workflow completes via ExitTransition. | Temporary test directory created; workflow with ExitTransition; session running in test directory; all file operations occur within test fixtures | Execute workflow until ExitTransition event emitted | Workflow completes; session status shows `"completed"`; final state at exit target node |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_SimultaneousExitAndNormalEvent` | `race` | Exit event and normal event emitted simultaneously are serialized. | Workflow with ExitTransition on "Exit" and normal transition on "Continue" from node "Hub"; session at `CurrentState="Hub"` | Emit "Exit" and "Continue" simultaneously | First event processed; if "Exit" first, session completes; if "Continue" first, session transitions and "Exit" fails (not at "Hub"); events serialized by session lock |
| `TestExitTransition_MultipleSimultaneousExitEvents` | `race` | Multiple exit events for different ExitTransitions emitted simultaneously are serialized. | Workflow with 2 ExitTransitions from node "Hub": on "Exit1" to "Final1" (human), on "Exit2" to "Final2" (human); session at `CurrentState="Hub"` | Emit "Exit1" and "Exit2" simultaneously | First event processed completes session; second event fails (session already completed); only one ExitTransition triggers; events serialized by session lock |

### Ordering — Definition Order

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitTransition_OrderPreservedInWorkflow` | `unit` | ExitTransitions in workflow array preserve definition order. | Temporary test directory created; workflow definition in test directory; all file operations occur within test fixtures | Add 3 ExitTransitions in order: E1, E2, E3 | Query workflow; ExitTransitions returned in order: E1, E2, E3 |
