# Test Specification: `workflow_definition_loader_test.go`

## Source File Under Test
`storage/workflow_definition_loader.go`

## Test File
`storage/workflow_definition_loader_test.go`

---

## `WorkflowDefinitionLoader`

### Happy Path — Load

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_ValidDefinition` | `unit` | Loads a well-formed workflow YAML and returns a valid WorkflowDefinition. | Create temp dir with `.spectra/workflows/MyWorkflow.yaml` containing valid YAML (description, entryNode, nodes with agent and human types, transitions, exitTransitions using camelCase keys). Inject mock AgentLoader that returns success for all agent roles. | `workflowName="MyWorkflow"` | Returns `*WorkflowDefinition` with Name=`"MyWorkflow"`, all fields matching YAML content, nil error |
| `TestWorkflowDefinitionLoader_Load_NameDerivedFromFilename` | `unit` | Name is derived from the workflowName parameter, not from YAML content. | Create temp dir with `.spectra/workflows/CodeReview.yaml` containing valid YAML (no name field). Inject mock AgentLoader returning success. | `workflowName="CodeReview"` | Returns `*WorkflowDefinition` with Name=`"CodeReview"` |
| `TestWorkflowDefinitionLoader_Load_EmptyDescription` | `unit` | Missing description field is accepted as empty string. | Create temp dir with `.spectra/workflows/Minimal.yaml` without description field but all other required fields. Inject mock AgentLoader returning success. | `workflowName="Minimal"` | Returns `*WorkflowDefinition` with Description=`""`, nil error |
| `TestWorkflowDefinitionLoader_Load_MultipleAgentNodes` | `unit` | All agent nodes have their roles validated via AgentLoader. | Create temp dir with `.spectra/workflows/Multi.yaml` containing multiple agent-type nodes with different agentRole values. Inject mock AgentLoader that records calls and returns success. | `workflowName="Multi"` | Returns `*WorkflowDefinition`; mock AgentLoader received Load call for each unique agent role |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_FileNotFound` | `unit` | Returns not-found error when YAML file does not exist. | Create temp dir with `.spectra/workflows/` directory but no YAML file. Inject mock AgentLoader. | `workflowName="Missing"` | Returns error matching `"workflow definition not found: Missing"` |
| `TestWorkflowDefinitionLoader_Load_ReadPermissionDenied` | `unit` | Returns wrapped read error on permission failure. | Create temp dir with `.spectra/workflows/Locked.yaml` with permissions `0000`. Inject mock AgentLoader. | `workflowName="Locked"` | Returns error matching `"failed to read workflow definition 'Locked':"` containing permission error |
| `TestWorkflowDefinitionLoader_Load_YamlSyntaxError` | `unit` | Returns parse error for syntactically invalid YAML. | Create temp dir with `.spectra/workflows/Bad.yaml` containing invalid YAML. Inject mock AgentLoader. | `workflowName="Bad"` | Returns error matching `"failed to parse workflow definition 'Bad':"` |
| `TestWorkflowDefinitionLoader_Load_YamlUnknownField` | `unit` | Returns parse error when YAML contains unknown fields. | Create temp dir with `.spectra/workflows/Extra.yaml` containing valid structure plus `customField: value`. Inject mock AgentLoader. | `workflowName="Extra"` | Returns error matching `"failed to parse workflow definition 'Extra':"` |
| `TestWorkflowDefinitionLoader_Load_YamlSnakeCaseField` | `unit` | Rejects YAML with snake_case field names as unknown fields. | Create temp dir with `.spectra/workflows/Snake.yaml` using `entry_node` instead of `entryNode`. Inject mock AgentLoader. | `workflowName="Snake"` | Returns error matching `"failed to parse workflow definition 'Snake':"` |
| `TestWorkflowDefinitionLoader_Load_NodeConstructorFails_WithName` | `unit` | Returns wrapped error with node name when node constructor fails. | Create temp dir with `.spectra/workflows/BadNode.yaml` containing a node with valid name but invalid type (e.g., `"bot"`). Inject mock AgentLoader. | `workflowName="BadNode"` | Returns error matching `"workflow definition 'BadNode' validation failed: node 'NodeName':"` |
| `TestWorkflowDefinitionLoader_Load_NodeConstructorFails_EmptyName` | `unit` | Returns wrapped error with index fallback when node name is empty. | Create temp dir with `.spectra/workflows/NoName.yaml` containing a node with empty name field. Inject mock AgentLoader. | `workflowName="NoName"` | Returns error matching `"workflow definition 'NoName' validation failed: node[0]:"` |
| `TestWorkflowDefinitionLoader_Load_TransitionConstructorFails` | `unit` | Returns wrapped error with transition context when transition constructor fails. | Create temp dir with `.spectra/workflows/BadTrans.yaml` containing a transition with fromNode == toNode. Inject mock AgentLoader. | `workflowName="BadTrans"` | Returns error matching `"workflow definition 'BadTrans' validation failed: transition (from"` |
| `TestWorkflowDefinitionLoader_Load_ExitTransitionConstructorFails` | `unit` | Returns wrapped error with exit transition context when constructor fails. | Create temp dir with `.spectra/workflows/BadExit.yaml` containing an exit transition with invalid fields. Inject mock AgentLoader. | `workflowName="BadExit"` | Returns error matching `"workflow definition 'BadExit' validation failed: exit_transition (from"` |
| `TestWorkflowDefinitionLoader_Load_WorkflowDefinitionConstructorFails` | `unit` | Returns wrapped error when WorkflowDefinition constructor rejects structural constraints. | Create temp dir with `.spectra/workflows/BadGraph.yaml` with valid nodes/transitions individually but duplicate (fromNode, eventType) pair. Inject mock AgentLoader. | `workflowName="BadGraph"` | Returns error matching `"workflow definition 'BadGraph' validation failed:"` |
| `TestWorkflowDefinitionLoader_Load_AgentRoleNotFound` | `unit` | Returns error when AgentLoader fails to load a referenced agent role. | Create temp dir with `.spectra/workflows/BadRef.yaml` containing valid workflow with an agent node referencing `"NonExistent"`. Inject mock AgentLoader that returns error for `"NonExistent"`. | `workflowName="BadRef"` | Returns error matching `"workflow definition 'BadRef' validation failed: node 'NodeName' references invalid agent_role 'NonExistent':"` |
| `TestWorkflowDefinitionLoader_Load_AgentRoleValidationFails` | `unit` | Returns error when AgentLoader returns a validation error for the referenced agent. | Create temp dir with `.spectra/workflows/InvalidAgent.yaml` containing valid workflow. Inject mock AgentLoader that returns specific validation error (e.g., agent_root missing). | `workflowName="InvalidAgent"` | Returns error matching `"workflow definition 'InvalidAgent' validation failed: node"` containing the underlying agent error |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_EmptyWorkflowName` | `unit` | Returns not-found error when workflowName is empty string. | Create temp dir with `.spectra/workflows/` directory. Inject mock AgentLoader. | `workflowName=""` | Returns error matching `"workflow definition not found: "` |
| `TestWorkflowDefinitionLoader_Load_EmptyYamlFile` | `unit` | Returns parse error for zero-byte YAML file. | Create temp dir with `.spectra/workflows/Empty.yaml` as an empty file. Inject mock AgentLoader. | `workflowName="Empty"` | Returns error matching `"failed to parse workflow definition 'Empty':"` |
| `TestWorkflowDefinitionLoader_Load_EmptyNodesArray` | `unit` | Returns validation error when nodes array is empty. | Create temp dir with `.spectra/workflows/NoNodes.yaml` with `nodes: []`. Inject mock AgentLoader. | `workflowName="NoNodes"` | Returns error matching `"workflow definition 'NoNodes' validation failed:"` containing nodes-related error |
| `TestWorkflowDefinitionLoader_Load_EmptyTransitionsArray` | `unit` | Returns validation error when transitions array is empty. | Create temp dir with `.spectra/workflows/NoTrans.yaml` with `transitions: []` and valid nodes. Inject mock AgentLoader. | `workflowName="NoTrans"` | Returns error matching `"workflow definition 'NoTrans' validation failed:"` containing transitions-related error |
| `TestWorkflowDefinitionLoader_Load_EmptyExitTransitionsArray` | `unit` | Returns validation error when exitTransitions array is empty. | Create temp dir with `.spectra/workflows/NoExit.yaml` with `exitTransitions: []` and valid nodes/transitions. Inject mock AgentLoader. | `workflowName="NoExit"` | Returns error matching `"workflow definition 'NoExit' validation failed:"` containing exit_transitions-related error |

### Boundary Values — workflowName

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_PathTraversal` | `unit` | Path separators in workflowName result in file-not-found. | Create temp dir with `.spectra/workflows/` directory. Inject mock AgentLoader. | `workflowName="../malicious/workflow"` | Returns error (file not found or read error) |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_AgentLoaderNotCalledForHumanNodes` | `unit` | AgentLoader is not invoked for human-type nodes. | Create temp dir with `.spectra/workflows/HumanOnly.yaml` containing only human-type nodes with valid workflow structure. Inject mock AgentLoader that fails if called. | `workflowName="HumanOnly"` | Returns `*WorkflowDefinition` successfully; AgentLoader never called |
| `TestWorkflowDefinitionLoader_Load_AgentLoaderFailFast` | `unit` | Stops agent validation on first failure without checking remaining agents. | Create temp dir with `.spectra/workflows/TwoAgents.yaml` containing two agent nodes. Inject mock AgentLoader that fails for the first agent role encountered. | `workflowName="TwoAgents"` | Returns error for the first agent; AgentLoader called at most once |
| `TestWorkflowDefinitionLoader_Load_NodeConstructionFailFast` | `unit` | Stops node construction on first failure without constructing remaining nodes. | Create temp dir with `.spectra/workflows/MultiNode.yaml` with first node invalid and second node valid. Inject mock AgentLoader. | `workflowName="MultiNode"` | Returns error for first node only |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_NoCaching` | `unit` | Second Load reflects file changes made after first Load. | Create temp dir with `.spectra/workflows/Mutable.yaml` with initial content. Inject mock AgentLoader returning success. After first Load, overwrite YAML with different description. | Two sequential `Load("Mutable")` calls | First returns original description; second returns updated description |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinitionLoader_Load_ConcurrentAccess` | `unit` | Multiple goroutines loading same workflow succeed without interference. | Create temp dir with `.spectra/workflows/Shared.yaml` containing valid content. Inject mock AgentLoader returning success. | Launch multiple goroutines calling `Load("Shared")` concurrently | All goroutines return valid `*WorkflowDefinition` with nil error |
