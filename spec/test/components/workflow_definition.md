# Test Specification: `workflow_definition.go`

## Source File Under Test
`components/workflow_definition.go`

## Test File
`components/workflow_definition_test.go`

---

## `WorkflowDefinition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_ValidWorkflowAllFields` | `unit` | Creates WorkflowDefinition with all fields provided. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="DefaultLogicSpec"`, `Description="Simple workflow"`, `EntryNode="HumanRequirement"`, `ExitTransitions=[{FromNode:"HumanApproval", EventType:"RequirementApproved", ToNode:"HumanRequirement"}]`, `Nodes=[{Name:"HumanRequirement", Type:"human"}, {Name:"HumanApproval", Type:"human"}]`, `Transitions=[{FromNode:"HumanRequirement", EventType:"RequirementProvided", ToNode:"HumanApproval"}, {FromNode:"HumanApproval", EventType:"RequirementApproved", ToNode:"HumanRequirement"}]` | Returns valid WorkflowDefinition; all fields match input; YAML file created at `<test-dir>/.spectra/workflows/DefaultLogicSpec.yaml` |
| `TestWorkflowDefinition_EmptyDescription` | `unit` | Creates WorkflowDefinition with empty description. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="TestWorkflow"`, `Description=""`, `EntryNode="Start"`, valid ExitTransitions, Nodes, Transitions | Returns valid WorkflowDefinition; `Description=""` |
| `TestWorkflowDefinition_MultipleExitTransitions` | `unit` | Creates WorkflowDefinition with multiple exit transitions. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="MultiExit"`, `ExitTransitions=[{FromNode:"End1", EventType:"Done1", ToNode:"Start"}, {FromNode:"End2", EventType:"Done2", ToNode:"Start"}]`, corresponding nodes and transitions | Returns valid WorkflowDefinition; `ExitTransitions` contains both transitions |

### Happy Path — Load from YAML

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_LoadValidYAML` | `unit` | Loads WorkflowDefinition from valid YAML file. | Temporary test directory created; YAML file at `<test-dir>/.spectra/workflows/DefaultLogicSpec.yaml` with all required fields; all file operations occur within test fixtures | Load workflow with `Name="DefaultLogicSpec"` | Returns valid WorkflowDefinition; all fields match YAML content |
| `TestWorkflowDefinition_LoadWithEmptyDescription` | `unit` | Loads WorkflowDefinition with empty description from YAML. | Temporary test directory created; YAML file at `<test-dir>/.spectra/workflows/Test.yaml` with `description: ""`; all file operations occur within test fixtures | Load workflow with `Name="Test"` | Returns valid WorkflowDefinition; `Description=""` |

### Validation Failures — Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_EmptyName` | `unit` | Rejects WorkflowDefinition with empty Name. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name=""`, valid other fields | Returns error; error message matches `/workflow name.*non-empty/i` |
| `TestWorkflowDefinition_NameWithSpaces` | `unit` | Rejects WorkflowDefinition with Name containing spaces. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Default LogicSpec"`, valid other fields | Returns error; error message matches `/workflow name.*PascalCase.*spaces.*special.*characters/i` |
| `TestWorkflowDefinition_NameWithUnderscores` | `unit` | Rejects WorkflowDefinition with Name containing underscores. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Default_LogicSpec"`, valid other fields | Returns error; error message matches `/workflow name.*PascalCase.*spaces.*special.*characters/i` |
| `TestWorkflowDefinition_NameWithHyphens` | `unit` | Rejects WorkflowDefinition with Name containing hyphens. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Default-LogicSpec"`, valid other fields | Returns error; error message matches `/workflow name.*PascalCase.*spaces.*special.*characters/i` |
| `TestWorkflowDefinition_NameNotPascalCase` | `unit` | Rejects WorkflowDefinition with Name not in PascalCase. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="defaultLogicSpec"`, valid other fields | Returns error; error message matches `/workflow name.*PascalCase/i` |

### Validation Failures — EntryNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_EntryNodeNonExistent` | `unit` | Rejects WorkflowDefinition with EntryNode referencing non-existent node. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `EntryNode="NonExistent"`, `Nodes=[{Name:"Existing", Type:"human"}]` | Returns error; error message matches `/entry.*node.*NonExistent.*not found/i` |
| `TestWorkflowDefinition_EntryNodeNotHuman` | `unit` | Rejects WorkflowDefinition with EntryNode referencing agent node. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `EntryNode="AgentNode"`, `Nodes=[{Name:"AgentNode", Type:"agent", AgentRole:"Architect"}]` | Returns error; error message matches `/entry node.*AgentNode.*type.*human/i` |

### Validation Failures — ExitTransitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_DuplicateExitTransition` | `unit` | Rejects WorkflowDefinition with duplicate exit transitions (same `from_node`, `event_type`, `to_node`). | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `ExitTransitions=[{FromNode:"A", EventType:"Done", ToNode:"B"}, {FromNode:"A", EventType:"Done", ToNode:"B"}]`, matching transition in `Transitions` | Returns error; error message matches `/duplicate exit transition.*event_type.*Done.*from_node.*A.*to_node.*B/i` |
| `TestWorkflowDefinition_EmptyExitTransitions` | `unit` | Rejects WorkflowDefinition with empty ExitTransitions array. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `ExitTransitions=[]`, valid other fields | Returns error; error message matches `/at least one exit transition required/i` |
| `TestWorkflowDefinition_ExitTransitionNoMatch` | `unit` | Rejects WorkflowDefinition when exit transition does not match any defined transition. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `ExitTransitions=[{FromNode:"A", EventType:"Done", ToNode:"B"}]`, `Transitions=[{FromNode:"A", EventType:"Start", ToNode:"B"}]` (mismatch) | Returns error; error message matches `/exit transition.*from_node.*A.*event_type.*Done.*to_node.*B.*no corresponding transition/i` |
| `TestWorkflowDefinition_ExitTransitionTargetsAgent` | `unit` | Rejects WorkflowDefinition when exit transition targets agent node. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `ExitTransitions=[{FromNode:"A", EventType:"Done", ToNode:"AgentNode"}]`, `Nodes` includes `{Name:"AgentNode", Type:"agent", AgentRole:"Architect"}`, matching transition in `Transitions` | Returns error; error message matches `/exit transition.*to_node.*AgentNode.*must target.*human.*type.*agent/i` |

### Validation Failures — Nodes

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_EmptyNodes` | `unit` | Rejects WorkflowDefinition with empty Nodes array. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Nodes=[]`, valid other fields | Returns error; error message matches `/at least one node required/i` or similar validation message |
| `TestWorkflowDefinition_DuplicateNodeNames` | `unit` | Rejects WorkflowDefinition with duplicate node names. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Nodes=[{Name:"Node1", Type:"human"}, {Name:"Node1", Type:"agent", AgentRole:"Architect"}]` | Returns error; error message matches `/duplicate.*node.*name.*Node1/i` |

### Validation Failures — Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_DuplicateTransition` | `unit` | Rejects WorkflowDefinition with two transitions sharing same `from_node` and `event_type`. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Transitions=[{FromNode:"A", EventType:"Done", ToNode:"B"}, {FromNode:"A", EventType:"Done", ToNode:"C"}]`, valid nodes | Returns error; error message matches `/duplicate transition.*event.*Done.*node.*A/i` |
| `TestWorkflowDefinition_EmptyTransitions` | `unit` | Rejects WorkflowDefinition with empty Transitions array. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Transitions=[]`, valid other fields | Returns error; error message matches `/at least one transition required/i` or similar validation message |
| `TestWorkflowDefinition_TransitionFromNodeNonExistent` | `unit` | Rejects WorkflowDefinition when transition references non-existent FromNode. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Transitions=[{FromNode:"NonExistent", EventType:"Event", ToNode:"Existing"}]`, `Nodes=[{Name:"Existing", Type:"human"}]` | Returns error; error message matches `/transition.*undefined.*node.*NonExistent/i` |
| `TestWorkflowDefinition_TransitionToNodeNonExistent` | `unit` | Rejects WorkflowDefinition when transition references non-existent ToNode. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Transitions=[{FromNode:"Existing", EventType:"Event", ToNode:"NonExistent"}]`, `Nodes=[{Name:"Existing", Type:"human"}]` | Returns error; error message matches `/transition.*undefined.*node.*NonExistent/i` |
| `TestWorkflowDefinition_NodeNoOutgoingTransition` | `unit` | Rejects WorkflowDefinition when non-exit-target node has no outgoing transitions. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Nodes=[{Name:"Isolated", Type:"human"}, {Name:"Start", Type:"human"}, {Name:"End", Type:"human"}]`, `Transitions=[{FromNode:"Start", EventType:"Go", ToNode:"Isolated"}, {FromNode:"Start", EventType:"Finish", ToNode:"End"}]`, `ExitTransitions=[{FromNode:"Start", EventType:"Finish", ToNode:"End"}]` (Isolated is not exit target and has no outgoing transitions) | Returns error; error message matches `/node.*Isolated.*no outgoing transitions.*not.*exit target/i` |

### Validation Failures — File Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_FileDoesNotExist` | `unit` | Returns error when workflow YAML file does not exist. | Temporary test directory created; `.spectra/workflows/` directory created but empty; all file operations occur within test fixtures | Load workflow with `Name="NonExistent"` | Returns error; error message matches `/workflow.*not found/i` |

### Validation Failures — Malformed YAML

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_MalformedYAML` | `unit` | Rejects WorkflowDefinition with malformed YAML syntax. | Temporary test directory created; YAML file at `<test-dir>/.spectra/workflows/Broken.yaml` with invalid YAML syntax (unclosed quote); all file operations occur within test fixtures | Load workflow with `Name="Broken"` | Returns parse error; error message indicates YAML syntax issue |
| `TestWorkflowDefinition_MissingRequiredField` | `unit` | Rejects WorkflowDefinition with missing required field (entry_node). | Temporary test directory created; YAML file at `<test-dir>/.spectra/workflows/Incomplete.yaml` missing `entry_node` field; all file operations occur within test fixtures | Load workflow with `Name="Incomplete"` | Returns error; error message matches `/entry.*node.*required/i` |

### Validation Failures — Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_DuplicateName` | `unit` | Rejects loading multiple WorkflowDefinitions with same Name. | Temporary test directory created; YAML file at `<test-dir>/.spectra/workflows/DefaultLogicSpec.yaml`; second YAML loaded from different source with same `name: "DefaultLogicSpec"`; all file operations occur within test fixtures | Load both workflows with `Name="DefaultLogicSpec"` | Second workflow load returns error; error message matches `/workflow.*DefaultLogicSpec.*already exists/i` |

### Happy Path — Exit Transition Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_ExitTransitionTriggersCompletion` | `unit` | Session completes when exit transition is traversed. | Temporary test directory created; workflow loaded with exit transition from "HumanApproval" to "HumanRequirement" on "RequirementApproved"; session at `CurrentState="HumanApproval"`, `Status="running"`; all file operations occur within test fixtures | Emit event: `Type="RequirementApproved"` | Session `CurrentState` transitions to "HumanRequirement" in memory; session `Status` immediately set to `"completed"` in memory; persistence to disk is best-effort; target node does not execute |
| `TestWorkflowDefinition_ExitTargetNodeNotExecuted` | `unit` | Target node of exit transition does not execute when exit transition is traversed. | Temporary test directory created; workflow with exit transition to human node "Final"; mock tracker for node execution; session at exit transition source; all file operations occur within test fixtures | Emit event matching exit transition | Session transitions to "Final" and marks completed; "Final" node actions are NOT executed; no agent/human interaction triggered |

### Happy Path — Session Initialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_SessionStartsAtEntryNode` | `unit` | New session starts at EntryNode with initializing status. | Temporary test directory created; workflow loaded with `EntryNode="HumanRequirement"`; all file operations occur within test fixtures | Create new session from workflow | Session created with `CurrentState="HumanRequirement"`, `Status="initializing"` |

### Happy Path — Event Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_UndefinedEventRecorded` | `unit` | Undefined event is recorded in audit trail and returns error without terminating session. | Temporary test directory created; workflow loaded; session at `CurrentState="A"`, `Status="running"`; only transition from "A" is on "Event1"; all file operations occur within test fixtures | Emit event: `Type="Event2"` (no matching transition) | EventProcessor records event in EventHistory; returns RuntimeResponse with `status="error"`, message matches `/no.*transition/i`; session `Status` remains `"running"`; caller may retry |
| `TestWorkflowDefinition_NoMatchingTransitionPreservesState` | `unit` | Session state unchanged when event has no matching transition. | Temporary test directory created; workflow loaded; session at `CurrentState="NodeA"`; no transition from "NodeA" on "UnknownEvent"; all file operations occur within test fixtures | Emit event: `Type="UnknownEvent"` | Session remains at `CurrentState="NodeA"`; `Status` remains `"running"`; error response returned |

### Happy Path — Node Outgoing Transition Validation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_ExitTargetNodeMayLackOutgoing` | `unit` | Node targeted by exit transition is allowed to have no outgoing transitions. | Temporary test directory created; `.spectra/workflows/` directory created in test directory; all file operations occur within test fixtures | `Name="Test"`, `Nodes=[{Name:"Start", Type:"human"}, {Name:"End", Type:"human"}]`, `Transitions=[{FromNode:"Start", EventType:"Go", ToNode:"End"}, {FromNode:"Start", EventType:"Exit", ToNode:"End"}]`, `ExitTransitions=[{FromNode:"Start", EventType:"Exit", ToNode:"End"}]` (End is exit target, has no outgoing) | Workflow validation succeeds; no error |
| `TestWorkflowDefinition_ExitTargetWithOutgoingWarning` | `unit` | Warning issued when exit target node has outgoing transitions. | Temporary test directory created; workflow with exit transition to "Final"; "Final" has outgoing transition to "Start"; all file operations occur within test fixtures | Validate workflow | Returns warning message matching `/exit target.*Final.*outgoing transitions.*never be used/i`; workflow not rejected |

### Validation Failures — Unreachable Node

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_UnreachableNodeRejected` | `unit` | Returns error for node with no incoming transitions (except entry node). | Temporary test directory created; workflow with node "Isolated" that has no incoming transitions and is not the entry node; all file operations occur within test fixtures | Validate workflow | Returns error matching `/unreachable.*node.*Isolated/i`; workflow rejected |

### Happy Path — YAML Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_ToYAML` | `unit` | WorkflowDefinition serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | WorkflowDefinition with all fields populated | YAML contains `name`, `description`, `entry_node`, `exit_transitions`, `nodes`, `transitions` with correct values and structure |
| `TestWorkflowDefinition_ToYAMLEmptyDescription` | `unit` | WorkflowDefinition with empty description serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | WorkflowDefinition with `Description=""` | YAML contains `description: ""` |
| `TestWorkflowDefinition_ToYAMLMultipleExitTransitions` | `unit` | WorkflowDefinition with multiple exit transitions serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | WorkflowDefinition with 3 exit transitions | YAML `exit_transitions` array contains all 3 transitions; order preserved |

### Happy Path — Built-in Workflow Copy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_BuiltinCopiedDuringInit` | `e2e` | Built-in workflows are copied to `.spectra/workflows/` during `spectra init`. | Temporary test directory created; no `.spectra/workflows/` directory exists; all file operations occur within test fixtures | Execute `spectra init` | Built-in workflow files copied to `<test-dir>/.spectra/workflows/`; files readable and valid |
| `TestWorkflowDefinition_ExistingWorkflowNotOverwritten` | `e2e` | Existing workflow file is not overwritten during `spectra init`. | Temporary test directory created; `.spectra/workflows/DefaultLogicSpec.yaml` exists with custom content; all file operations occur within test fixtures | Execute `spectra init` | `DefaultLogicSpec.yaml` content unchanged; other built-in workflows copied; no error returned |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_FieldsImmutable` | `unit` | WorkflowDefinition fields cannot be modified after creation. | WorkflowDefinition instance created | Attempt to modify `Name`, `Description`, `EntryNode`, `ExitTransitions`, `Nodes`, or `Transitions` | Field modification attempt fails or has no effect; original values remain |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_ImplementsInterface` | `unit` | WorkflowDefinition type implements expected interface. | | WorkflowDefinition instance created | WorkflowDefinition satisfies WorkflowDefinition interface contract (GetName, GetDescription, GetEntryNode, GetExitTransitions, GetNodes, GetTransitions methods) |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_ListWorkflows` | `e2e` | CLI lists all available workflows. | Temporary test directory created; multiple workflow definition files in `<test-dir>/.spectra/workflows/`; all file operations occur within test fixtures | Execute `spectra workflow list` | Command succeeds; output lists all workflows with names and descriptions |
| `TestWorkflowDefinition_ShowWorkflowDetails` | `e2e` | CLI shows details for specific workflow. | Temporary test directory created; workflow definition file at `<test-dir>/.spectra/workflows/DefaultLogicSpec.yaml`; all file operations occur within test fixtures | Execute `spectra workflow show --workflow DefaultLogicSpec` | Command succeeds; output displays all fields: name, description, entry_node, exit_transitions, nodes, transitions |
| `TestWorkflowDefinition_ValidateWorkflow` | `e2e` | CLI validates workflow definition file. | Temporary test directory created; valid workflow definition file at `<test-dir>/.spectra/workflows/TestWorkflow.yaml`; all file operations occur within test fixtures | Execute `spectra workflow validate --workflow TestWorkflow` | Command succeeds; no errors reported |
| `TestWorkflowDefinition_ValidateWorkflowInvalidEntryNode` | `e2e` | CLI validation fails for workflow with invalid entry node. | Temporary test directory created; workflow definition file at `<test-dir>/.spectra/workflows/BadWorkflow.yaml` with `entry_node` referencing non-existent node; all file operations occur within test fixtures | Execute `spectra workflow validate --workflow BadWorkflow` | Command fails; error message matches `/entry.*node.*not found/i` |
| `TestWorkflowDefinition_RunWorkflow` | `e2e` | CLI runs workflow and creates session at entry node. | Temporary test directory created; valid workflow definition file at `<test-dir>/.spectra/workflows/TestWorkflow.yaml`; all file operations occur within test fixtures | Execute `spectra run --workflow TestWorkflow` | Command succeeds; session created with `CurrentState` set to workflow's `EntryNode`; `Status="initializing"` |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_SimultaneousSessionCreation` | `race` | Multiple sessions can be created from same workflow definition concurrently. | Temporary test directory created; workflow definition loaded; all file operations occur within test fixtures | Create 3 sessions concurrently from same workflow | All 3 sessions created successfully; each has unique SessionID; all start at same EntryNode |

### Ordering — Node and Transition Order

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_NodeOrderPreserved` | `unit` | Nodes in workflow Nodes array preserve definition order. | Temporary test directory created; workflow definition in test directory; all file operations occur within test fixtures | Add 5 nodes in order: N1, N2, N3, N4, N5 | Query workflow; nodes returned in order: N1, N2, N3, N4, N5 |
| `TestWorkflowDefinition_TransitionOrderPreserved` | `unit` | Transitions in workflow Transitions array preserve definition order. | Temporary test directory created; workflow definition in test directory; all file operations occur within test fixtures | Add 5 transitions in order: T1, T2, T3, T4, T5 | Query workflow; transitions returned in order: T1, T2, T3, T4, T5 |
| `TestWorkflowDefinition_ExitTransitionOrderPreserved` | `unit` | Exit transitions in workflow ExitTransitions array preserve definition order. | Temporary test directory created; workflow definition in test directory; all file operations occur within test fixtures | Add 3 exit transitions in order: E1, E2, E3 | Query workflow; exit transitions returned in order: E1, E2, E3 |
