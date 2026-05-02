# Test Specification: `workflow_definition_loader.go`

## Source File Under Test
`storage/workflow_definition_loader.go`

## Test File
`storage/workflow_definition_loader_test.go`

---

## `WorkflowDefinitionLoader`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_New` | `unit` | Creates a new WorkflowDefinitionLoader with valid ProjectRoot and AgentDefinitionLoader. | Temp dir fixture with `.spectra/workflows/` directory; mock AgentDefinitionLoader | `ProjectRoot=<temp_dir>`, `AgentLoader=<mock>` | Returns non-nil WorkflowDefinitionLoader; no error |

### Happy Path — Load

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_MinimalValidWorkflow` | `unit` | Loads a minimal valid workflow with single human node. | Temp dir fixture with `.spectra/workflows/Simple.yaml` containing name, entry_node (human type), one node, one transition, one exit_transition | `WorkflowName="Simple"` | Returns valid WorkflowDefinition with all fields populated; no error |
| `TestWorkflowDefinitionLoader_Load_MultiNodeWorkflow` | `unit` | Loads workflow with multiple nodes and transitions. | Temp dir fixture with workflow YAML containing 3 nodes (1 human, 2 agent), multiple transitions, exit transitions | `WorkflowName="Complex"` | Returns valid WorkflowDefinition with all nodes and transitions; no error |
| `TestWorkflowDefinitionLoader_Load_NameWithDigits` | `unit` | Accepts workflow name with digits in valid PascalCase. | Temp dir fixture with workflow YAML where `name: "V2Workflow"` | `WorkflowName="V2Workflow"` | Returns WorkflowDefinition with Name="V2Workflow"; no error |
| `TestWorkflowDefinitionLoader_Load_NameWithConsecutiveUppercase` | `unit` | Accepts workflow name with consecutive uppercase letters. | Temp dir fixture with workflow YAML where `name: "DefaultLOGICSPEC"` | `WorkflowName="DefaultLOGICSPEC"` | Returns WorkflowDefinition with Name="DefaultLOGICSPEC"; no error |
| `TestWorkflowDefinitionLoader_Load_SingleUppercaseLetter` | `unit` | Accepts single uppercase letter as workflow name. | Temp dir fixture with workflow YAML where `name: "A"` | `WorkflowName="A"` | Returns WorkflowDefinition with Name="A"; no error |
| `TestWorkflowDefinitionLoader_Load_WithOptionalDescription` | `unit` | Loads workflow with optional description field. | Temp dir fixture with workflow YAML containing `description: "Test workflow"` | `WorkflowName="Described"` | Returns WorkflowDefinition with Description="Test workflow"; no error |
| `TestWorkflowDefinitionLoader_Load_WithoutDescription` | `unit` | Loads workflow without optional description field. | Temp dir fixture with workflow YAML missing `description` | `WorkflowName="Simple"` | Returns WorkflowDefinition with Description as empty string; no error |
| `TestWorkflowDefinitionLoader_Load_ExitTargetWithOutgoingTransitions` | `unit` | Allows exit target nodes to have outgoing transitions. | Temp dir fixture with workflow YAML where exit target node has outgoing transitions defined; all non-exit-target nodes have outgoing transitions | `WorkflowName="ExitWithOut"` | Returns valid WorkflowDefinition; exit target with outgoing transitions allowed; no error |
| `TestWorkflowDefinitionLoader_Load_UnknownFieldsIgnored` | `unit` | Ignores unknown fields in YAML. | Temp dir fixture with workflow YAML containing all required fields plus `custom_metadata: "extra"` | `WorkflowName="Extended"` | Returns valid WorkflowDefinition; unknown fields ignored; no error |

### Happy Path — Agent Reference Validation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_AgentNodeReferencesValidAgent` | `unit` | Validates agent node references an existing valid agent. | Temp dir fixture with workflow containing agent node with `agent_role: "Architect"`; mock AgentDefinitionLoader returns success for "Architect" | `WorkflowName="WithAgent"` | AgentDefinitionLoader.Load("Architect") called; returns valid WorkflowDefinition; no error |
| `TestWorkflowDefinitionLoader_Load_MultipleAgentNodesValidated` | `unit` | Validates all agent nodes reference valid agents. | Temp dir fixture with workflow containing 3 agent nodes with different roles; mock AgentDefinitionLoader returns success for all | `WorkflowName="MultiAgent"` | AgentDefinitionLoader.Load called once per unique agent role; returns valid WorkflowDefinition; no error |
| `TestWorkflowDefinitionLoader_Load_DuplicateAgentRoleValidatedOnce` | `unit` | Validates each unique agent role only once even if used by multiple nodes. | Temp dir fixture with workflow containing 2 agent nodes both using `agent_role: "Coder"`; mock AgentDefinitionLoader | `WorkflowName="DupeRole"` | AgentDefinitionLoader.Load("Coder") called exactly once; returns valid WorkflowDefinition; no error |

### Validation Failures — File Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_FileNotFound` | `unit` | Returns error when workflow file does not exist. | Temp dir fixture with `.spectra/workflows/` directory but no Simple.yaml | `WorkflowName="Simple"` | Returns error matching `"workflow definition not found: Simple"` |
| `TestWorkflowDefinitionLoader_Load_EmptyWorkflowName` | `unit` | Returns error when workflow name is empty string. | Temp dir fixture with `.spectra/workflows/` directory | `WorkflowName=""` | Returns error matching `"workflow definition not found: "` |

### Validation Failures — File Read Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_PermissionDenied` | `unit` | Returns error when file exists but is not readable. | Temp dir fixture with `.spectra/workflows/Simple.yaml` with permissions set to 0000 | `WorkflowName="Simple"` | Returns error matching `"failed to read workflow definition 'Simple': permission denied"` |

### Validation Failures — YAML Parsing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_EmptyFile` | `unit` | Returns error when workflow file is completely empty. | Temp dir fixture with empty `.spectra/workflows/Simple.yaml` | `WorkflowName="Simple"` | Returns error matching `"failed to parse workflow definition 'Simple': EOF"` |
| `TestWorkflowDefinitionLoader_Load_InvalidYAMLSyntax` | `unit` | Returns error when YAML has syntax errors. | Temp dir fixture with `.spectra/workflows/Simple.yaml` containing invalid YAML with incorrect indentation | `WorkflowName="Simple"` | Returns error matching `"failed to parse workflow definition 'Simple': yaml: line"` and includes line/column info |

### Validation Failures — Missing Required Fields

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_MissingName` | `unit` | Returns error when name field is missing or empty. | Temp dir fixture with workflow YAML missing `name` field | `WorkflowName="Simple"` | Returns error matching `"workflow definition 'Simple' validation failed: missing required field 'name'"` |
| `TestWorkflowDefinitionLoader_Load_MissingEntryNode` | `unit` | Returns error when entry_node field is missing or empty. | Temp dir fixture with workflow YAML missing `entry_node` field | `WorkflowName="Simple"` | Returns error matching `"workflow definition 'Simple' validation failed: missing required field 'entry_node'"` |
| `TestWorkflowDefinitionLoader_Load_MissingNodes` | `unit` | Returns error when nodes array is missing or empty. | Temp dir fixture with workflow YAML with empty `nodes: []` | `WorkflowName="Simple"` | Returns error matching `"workflow definition 'Simple' validation failed: missing required field 'nodes'"` |
| `TestWorkflowDefinitionLoader_Load_MissingTransitions` | `unit` | Returns error when transitions array is missing or empty. | Temp dir fixture with workflow YAML with empty `transitions: []` | `WorkflowName="Simple"` | Returns error matching `"workflow definition 'Simple' validation failed: missing required field 'transitions'"` |
| `TestWorkflowDefinitionLoader_Load_MissingExitTransitions` | `unit` | Returns error when exit_transitions array is missing or empty. | Temp dir fixture with workflow YAML with empty `exit_transitions: []` | `WorkflowName="Simple"` | Returns error matching `"workflow definition 'Simple' validation failed: missing required field 'exit_transitions'"` |

### Validation Failures — Name Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_NameWithSpaces` | `unit` | Returns error when name contains spaces. | Temp dir fixture with workflow YAML where `name: "Default LogicSpec"` | `WorkflowName="Default LogicSpec"` | Returns error matching `"workflow definition 'Default LogicSpec' validation failed: name must be PascalCase with no spaces or special characters"` |
| `TestWorkflowDefinitionLoader_Load_NameWithUnderscore` | `unit` | Returns error when name contains underscores. | Temp dir fixture with workflow YAML where `name: "Default_LogicSpec"` | `WorkflowName="Default_LogicSpec"` | Returns error matching `"workflow definition 'Default_LogicSpec' validation failed: name must be PascalCase with no spaces or special characters"` |
| `TestWorkflowDefinitionLoader_Load_NameWithHyphen` | `unit` | Returns error when name contains hyphens. | Temp dir fixture with workflow YAML where `name: "Default-LogicSpec"` | `WorkflowName="Default-LogicSpec"` | Returns error matching `"workflow definition 'Default-LogicSpec' validation failed: name must be PascalCase with no spaces or special characters"` |
| `TestWorkflowDefinitionLoader_Load_NameWithDot` | `unit` | Returns error when name contains dots. | Temp dir fixture with workflow YAML where `name: "Default.LogicSpec"` | `WorkflowName="Default.LogicSpec"` | Returns error matching `"workflow definition 'Default.LogicSpec' validation failed: name must be PascalCase with no spaces or special characters"` |
| `TestWorkflowDefinitionLoader_Load_NameStartsLowercase` | `unit` | Returns error when name starts with lowercase letter. | Temp dir fixture with workflow YAML where `name: "defaultLogicSpec"` | `WorkflowName="defaultLogicSpec"` | Returns error matching `"workflow definition 'defaultLogicSpec' validation failed: name must be PascalCase with no spaces or special characters"` |

### Validation Failures — EntryNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_EntryNodeNotFound` | `unit` | Returns error when entry_node references non-existent node. | Temp dir fixture with workflow YAML where `entry_node: "NonExistent"` but node not in nodes array | `WorkflowName="BadEntry"` | Returns error matching `"workflow definition 'BadEntry' validation failed: entry_node 'NonExistent' references non-existent node"` |
| `TestWorkflowDefinitionLoader_Load_EntryNodeNotHumanType` | `unit` | Returns error when entry_node references agent node. | Temp dir fixture with workflow YAML where entry_node references node with `type: "agent"` | `WorkflowName="AgentEntry"` | Returns error matching `"workflow definition 'AgentEntry' validation failed: entry_node '<nodeName>' must have type 'human', but has type 'agent'"` |

### Validation Failures — Node Integrity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_DuplicateNodeName` | `unit` | Returns error when multiple nodes have the same name. | Temp dir fixture with workflow YAML containing two nodes with `name: "Review"` | `WorkflowName="DuplicateNode"` | Returns error matching `"workflow definition 'DuplicateNode' validation failed: duplicate node name 'Review'"` |

### Validation Failures — Transition Integrity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_TransitionFromNodeNotFound` | `unit` | Returns error when transition from_node references non-existent node. | Temp dir fixture with workflow YAML where transition has `from_node: "Ghost"` not in nodes array | `WorkflowName="BadFrom"` | Returns error matching `"workflow definition 'BadFrom' validation failed: transition references non-existent node 'Ghost'"` |
| `TestWorkflowDefinitionLoader_Load_TransitionToNodeNotFound` | `unit` | Returns error when transition to_node references non-existent node. | Temp dir fixture with workflow YAML where transition has `to_node: "Ghost"` not in nodes array | `WorkflowName="BadTo"` | Returns error matching `"workflow definition 'BadTo' validation failed: transition references non-existent node 'Ghost'"` |
| `TestWorkflowDefinitionLoader_Load_TransitionSelfLoop` | `unit` | Returns error when transition has from_node equal to to_node. | Temp dir fixture with workflow YAML where transition has `from_node: "Review"` and `to_node: "Review"` | `WorkflowName="SelfLoop"` | Returns error matching `"workflow definition 'SelfLoop' validation failed: transition from_node and to_node must be different (node 'Review', event '<eventType>')"` |
| `TestWorkflowDefinitionLoader_Load_DuplicateTransitionKey` | `unit` | Returns error when two transitions share same from_node and event_type. | Temp dir fixture with workflow YAML where two transitions have `from_node: "Review"` and `event_type: "approve"` | `WorkflowName="DupeTrans"` | Returns error matching `"workflow definition 'DupeTrans' validation failed: duplicate transition for event 'approve' from node 'Review'"` |
| `TestWorkflowDefinitionLoader_Load_NodeWithoutOutgoingTransitions` | `unit` | Returns error when non-exit-target node has no outgoing transitions. | Temp dir fixture with workflow YAML where reachable node "Isolated" has incoming transitions but no outgoing transitions and is not an exit target | `WorkflowName="Isolated"` | Returns error matching `"workflow definition 'Isolated' validation failed: node 'Isolated' has no outgoing transitions and is not an exit target"` |

### Validation Failures — Node Reachability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_UnreachableNode` | `unit` | Returns error when non-entry node has no incoming transitions. | Temp dir fixture with workflow YAML containing a node with no incoming transitions | `WorkflowName="Unreachable"` | Returns error matching `"workflow definition 'Unreachable' validation failed: node 'Isolated' is unreachable (no incoming transitions)"` |

### Validation Failures — ExitTransition Integrity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_DuplicateExitTransition` | `unit` | Returns error when two exit transitions share identical triple. | Temp dir fixture with workflow YAML where two exit transitions have identical `from_node`, `event_type`, and `to_node` | `WorkflowName="DupeExit"` | Returns error matching `"workflow definition 'DupeExit' validation failed: duplicate exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>')"` |
| `TestWorkflowDefinitionLoader_Load_ExitTransitionNoCorrespondingTransition` | `unit` | Returns error when exit transition has no matching transition definition. | Temp dir fixture with workflow YAML where exit transition triple doesn't match any transition in transitions array | `WorkflowName="Orphan"` | Returns error matching `"workflow definition 'Orphan' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition definition"` |
| `TestWorkflowDefinitionLoader_Load_ExitTransitionTargetsAgentNode` | `unit` | Returns error when exit transition to_node references agent node. | Temp dir fixture with workflow YAML where exit transition `to_node` references node with `type: "agent"` | `WorkflowName="AgentExit"` | Returns error matching `"workflow definition 'AgentExit' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') must target a human node, but targets '<nodeName>' with type 'agent'"` |

### Validation Failures — Agent Reference Integrity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_AgentNodeReferencesNonExistentAgent` | `unit` | Returns error when agent node references non-existent agent file. | Temp dir fixture with workflow containing agent node with `agent_role: "Ghost"`; mock AgentDefinitionLoader returns "not found" error | `WorkflowName="BadAgent"` | Returns error matching `"workflow definition 'BadAgent' validation failed: node '<nodeName>' references invalid agent_role 'Ghost': agent definition not found: Ghost"` |
| `TestWorkflowDefinitionLoader_Load_AgentNodeReferencesInvalidAgent` | `unit` | Returns error when agent node references agent that fails validation. | Temp dir fixture with workflow containing agent node; mock AgentDefinitionLoader returns validation error (e.g., missing agent_root directory) | `WorkflowName="InvalidAgent"` | Returns error matching `"workflow definition 'InvalidAgent' validation failed: node '<nodeName>' references invalid agent_role '<role>': <agent validation error>"` |

### Validation Failures — Path Injection

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_WorkflowNameWithPathTraversal` | `unit` | Handles workflow name with path traversal characters. | Temp dir fixture with `.spectra/workflows/` directory | `WorkflowName="../malicious/workflow"` | File read fails; returns error matching `"workflow definition not found:"` or accesses unintended file |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_RepeatedCalls` | `unit` | Multiple Load calls with same workflow name return identical results. | Temp dir fixture with valid `.spectra/workflows/Simple.yaml`; mock AgentDefinitionLoader | Call `Load("Simple")` three times | All three calls return identical WorkflowDefinition values; AgentDefinitionLoader called each time (no caching) |
| `TestWorkflowDefinitionLoader_Load_FileModifiedBetweenCalls` | `unit` | Load reflects file changes between calls (no caching). | Temp dir fixture with valid workflow YAML; mock AgentDefinitionLoader | Load once, modify file on disk, Load again | Second Load returns updated content; no caching |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_ConcurrentSameWorkflow` | `race` | Multiple goroutines load the same workflow concurrently. | Temp dir fixture with valid workflow YAML; thread-safe mock AgentDefinitionLoader | 10 goroutines call `Load("Simple")` simultaneously | All calls succeed with identical results; no data races; no file locking conflicts |
| `TestWorkflowDefinitionLoader_Load_ConcurrentDifferentWorkflows` | `race` | Multiple goroutines load different workflows concurrently. | Temp dir fixture with 5 different workflow YAML files; thread-safe mock AgentDefinitionLoader | 10 goroutines each load different workflows simultaneously | All calls succeed with correct respective results; no data races |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_UsesStorageLayout` | `unit` | Verifies WorkflowDefinitionLoader calls StorageLayout.GetWorkflowPath with correct arguments. | Mock StorageLayout; temp dir fixture | `ProjectRoot=<temp_dir>`, `WorkflowName="Simple"` | StorageLayout.GetWorkflowPath called with ProjectRoot and "Simple"; file read from returned path |
| `TestWorkflowDefinitionLoader_Load_CallsAgentDefinitionLoaderForEachAgentNode` | `unit` | Verifies WorkflowDefinitionLoader calls AgentDefinitionLoader.Load for each agent node. | Temp dir fixture with workflow containing 2 agent nodes with roles "A" and "B"; mock AgentDefinitionLoader | `WorkflowName="TwoAgents"` | AgentDefinitionLoader.Load("A") and Load("B") both called with correct roles |

### Boundary Values — ProjectRoot

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_New_RelativeProjectRoot` | `unit` | Accepts relative ProjectRoot path. | Temp dir fixture with relative path `.spectra/workflows/`; mock AgentDefinitionLoader | `ProjectRoot="./project"` | WorkflowDefinitionLoader created; path composition may be relative; no error |
| `TestWorkflowDefinitionLoader_Load_ProjectRootWithoutSpectraDir` | `unit` | Handles missing .spectra directory. | Temp dir fixture without `.spectra/` directory; mock AgentDefinitionLoader | `WorkflowName="Simple"` | File read fails; returns error matching `"workflow definition not found: Simple"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_AgentDefinitionLoaderErrorPropagated` | `unit` | Propagates errors from AgentDefinitionLoader with workflow context. | Temp dir fixture with workflow containing agent node; mock AgentDefinitionLoader returns specific error | `WorkflowName="PropError"` | Returns error wrapping underlying agent loader error with workflow and node context |
