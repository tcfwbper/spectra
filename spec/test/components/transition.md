# Test Specification: `transition.go`

## Source File Under Test
`components/transition.go`

## Test File
`components/transition_test.go`

---

## `Transition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ValidAgentToAgent` | `unit` | Creates Transition from agent node to another agent node. | Workflow with nodes "Architect" (agent) and "ArchitectReviewer" (agent) | `FromNode="Architect"`, `EventType="DraftCompleted"`, `ToNode="ArchitectReviewer"` | Returns valid Transition; all fields match input |
| `TestTransition_ValidAgentToHuman` | `unit` | Creates Transition from agent node to human node. | Workflow with nodes "ArchitectReviewer" (agent) and "HumanApproval" (human) | `FromNode="ArchitectReviewer"`, `EventType="ReviewApproved"`, `ToNode="HumanApproval"` | Returns valid Transition; all fields match input |
| `TestTransition_ValidHumanToAgent` | `unit` | Creates Transition from human node to agent node. | Workflow with nodes "HumanApproval" (human) and "Architect" (agent) | `FromNode="HumanApproval"`, `EventType="RequirementReceived"`, `ToNode="Architect"` | Returns valid Transition; all fields match input |
| `TestTransition_MultipleFromSameNode` | `unit` | Multiple transitions from same node with different event types. | Workflow with nodes "Architect", "HumanApproval", "ArchitectReviewer" | Add two transitions: `FromNode="Architect"`, `EventType="AmbiguousSpecFound"`, `ToNode="HumanApproval"` and `FromNode="Architect"`, `EventType="DraftCompleted"`, `ToNode="ArchitectReviewer"` | Both transitions valid; coexist in workflow |

### Validation Failures — Node References

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_FromNodeNonExistent` | `unit` | Rejects Transition with non-existent FromNode. | Workflow with node "Reviewer" only | `FromNode="NonExistent"`, `EventType="Event"`, `ToNode="Reviewer"` | Returns error; error message matches `/transition.*undefined.*node.*NonExistent/i` |
| `TestTransition_ToNodeNonExistent` | `unit` | Rejects Transition with non-existent ToNode. | Workflow with node "Architect" only | `FromNode="Architect"`, `EventType="Event"`, `ToNode="NonExistent"` | Returns error; error message matches `/transition.*undefined.*node.*NonExistent/i` |
| `TestTransition_BothNodesNonExistent` | `unit` | Rejects Transition with both nodes non-existent. | Workflow with no nodes | `FromNode="Node1"`, `EventType="Event"`, `ToNode="Node2"` | Returns error; error message matches `/transition.*undefined.*node/i` |

### Validation Failures — Self-Loop

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_SelfLoop` | `unit` | Rejects Transition where FromNode equals ToNode. | Workflow with node "Processor" | `FromNode="Processor"`, `EventType="Continue"`, `ToNode="Processor"` | Returns error; error message matches `/from_node.*to_node.*different/i` |

### Validation Failures — EventType

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_EmptyEventType` | `unit` | Rejects Transition with empty EventType. | Workflow with nodes "A" and "B" | `FromNode="A"`, `EventType=""`, `ToNode="B"` | Returns error; error message matches `/event_type.*non-empty/i` |
| `TestTransition_EventTypeWithSpaces` | `unit` | Rejects Transition with EventType containing spaces. | Workflow with nodes "A" and "B" | `FromNode="A"`, `EventType="Task Completed"`, `ToNode="B"` | Returns error; error message matches `/event_type.*PascalCase.*spaces/i` |
| `TestTransition_EventTypeNotPascalCase` | `unit` | Rejects Transition with EventType not in PascalCase. | Workflow with nodes "A" and "B" | `FromNode="A"`, `EventType="task_completed"`, `ToNode="B"` | Returns error; error message matches `/event_type.*PascalCase/i` |

### Validation Failures — Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_DuplicateTransition` | `unit` | Rejects duplicate transition with same FromNode and EventType. | Workflow with existing transition: `FromNode="A"`, `EventType="Done"`, `ToNode="B"` | Add second transition: `FromNode="A"`, `EventType="Done"`, `ToNode="C"` | Returns error; error message matches `/duplicate.*transition.*Done.*node.*A/i` |

### Happy Path — Transition Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_TriggersStateChange` | `unit` | Transition executes when matching event emitted. | Workflow with transition from "Processing" to "Complete" on "Done"; session at `CurrentState="Processing"` | Emit event: `Type="Done"`, `SessionID=<session-uuid>` | Session transitions to `CurrentState="Complete"`; event recorded in EventHistory |
| `TestTransition_UnconditionalExecution` | `unit` | Transition always executes when event matches. | Workflow with transition from "Start" to "End" on "Finish"; session at `CurrentState="Start"` | Emit event: `Type="Finish"` multiple times | Each emission triggers transition; no conditions evaluated |

### Validation Failures — No Matching Transition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_NoMatchForEvent` | `unit` | Session remains running when event has no matching transition. | Workflow with transition from "A" to "B" on "Event1" only; session at `CurrentState="A"` | Emit event: `Type="Event2"` | Returns RuntimeResponse with `status="error"`, `message` matches `/no.*transition.*Event2.*node.*A/i`; session `Status` remains `"running"` (distinct from RuntimeResponse.status); event recorded in EventHistory; session remains at "A" |
| `TestTransition_NoMatchFromDifferentNode` | `unit` | Event valid for different node does not trigger transition. | Workflow with transition from "B" to "C" on "Proceed"; session at `CurrentState="A"` | Emit event: `Type="Proceed"` | Returns error; error message matches `/no.*transition.*Proceed.*node.*A/i`; session remains at "A" |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_FieldsImmutable` | `unit` | Transition fields cannot be modified after creation. | Transition instance created | Attempt to modify `FromNode`, `EventType`, or `ToNode` | Field modification attempt fails or has no effect; original values remain |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ImplementsInterface` | `unit` | Transition type implements expected interface. | | Transition instance created | Transition satisfies Transition interface contract (GetFromNode, GetEventType, GetToNode methods) |

### Happy Path — YAML Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ToYAML` | `unit` | Transition serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | Transition with `FromNode="Architect"`, `EventType="DraftCompleted"`, `ToNode="Reviewer"` | YAML contains `from_node: "Architect"`, `event_type: "DraftCompleted"`, `to_node: "Reviewer"` |
| `TestTransition_FromYAML` | `unit` | YAML deserializes to Transition correctly. | Temporary test directory created; YAML file in test directory; all file operations occur within test fixtures | YAML: `from_node: "A"`, `event_type: "Event"`, `to_node: "B"` | Transition created with matching fields |
| `TestTransition_MultipleTransitionsToYAML` | `unit` | Multiple transitions serialize to YAML array correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | Two transitions in workflow | YAML contains array with both transitions; order preserved |

### Happy Path — Workflow Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_AddedToWorkflow` | `unit` | Transition successfully added to workflow Transitions array. | Temporary test directory created; workflow definition in test directory with nodes "A" and "B"; all file operations occur within test fixtures | Add transition: `FromNode="A"`, `EventType="Proceed"`, `ToNode="B"` | Transition appears in workflow's `Transitions` array; workflow validation succeeds |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_ListTransitionsInWorkflow` | `e2e` | CLI lists all transitions in a workflow. | Temporary test directory created; workflow definition in test directory with 3 transitions; all file operations occur within test fixtures | Execute `spectra workflow transitions list --workflow <workflow-id>` | Command succeeds; output lists all 3 transitions with from_node, event_type, to_node |
| `TestTransition_ValidateWorkflowWithTransitions` | `e2e` | CLI validates workflow containing transitions. | Temporary test directory created; valid workflow definition in test directory with nodes and transitions; all file operations occur within test fixtures | Execute `spectra workflow validate --workflow <workflow-id>` | Command succeeds; no errors reported |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_SimultaneousEvents` | `race` | Multiple events triggering different transitions are serialized. | Workflow with two transitions from node "Hub": on "Event1" to "A", on "Event2" to "B"; session at `CurrentState="Hub"` | Emit "Event1" and "Event2" simultaneously | First event processed transitions session; second event fails (session no longer at "Hub"); events serialized by session lock |

### Ordering — Definition Order

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestTransition_OrderPreservedInWorkflow` | `unit` | Transitions in workflow Transitions array preserve definition order. | Temporary test directory created; workflow definition in test directory; all file operations occur within test fixtures | Add 3 transitions in order: T1, T2, T3 | Query workflow; transitions returned in order: T1, T2, T3 |
