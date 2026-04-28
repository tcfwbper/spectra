# Test Specification: `node.go`

## Source File Under Test
`components/node.go`

## Test File
`components/node_test.go`

---

## `Node`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_ValidAgentNode` | `unit` | Creates Node with type agent and valid agent role. | Temporary test directory created; agent definition file exists at `<test-dir>/.spectra/agents/ArchitectReviewer.yaml` with `role: "ArchitectReviewer"`; all file operations occur within test fixtures | `Name="ArchitectReviewer"`, `Type="agent"`, `AgentRole="ArchitectReviewer"`, `Description="Review specs"` | Returns valid Node; all fields match input |
| `TestNode_ValidHumanNode` | `unit` | Creates Node with type human without agent role. | | `Name="HumanApproval"`, `Type="human"`, `Description="Human reviews output"` | Returns valid Node; `AgentRole=""` |
| `TestNode_EmptyDescription` | `unit` | Creates Node with empty description (defaults to empty string). | | `Name="TestNode"`, `Type="human"`, `Description=""` | Returns valid Node; `Description=""` |
| `TestNode_OmittedDescription` | `unit` | Creates Node with description omitted (defaults to empty string). | | `Name="TestNode"`, `Type="human"`, `Description` omitted | Returns valid Node; `Description=""` |

### Validation Failures — Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_EmptyName` | `unit` | Rejects Node with empty Name. | | `Name=""`, `Type="human"` | Returns error; error message matches `/name.*non-empty/i` |
| `TestNode_NameWithSpaces` | `unit` | Rejects Node with Name containing spaces. | | `Name="Review Step"`, `Type="human"` | Returns error; error message matches `/name.*PascalCase.*spaces/i` |
| `TestNode_NameWithUnderscores` | `unit` | Rejects Node with Name containing underscores. | | `Name="Review_Step"`, `Type="human"` | Returns error; error message matches `/name.*PascalCase.*special.*characters/i` |
| `TestNode_NameWithHyphens` | `unit` | Rejects Node with Name containing hyphens. | | `Name="Review-Step"`, `Type="human"` | Returns error; error message matches `/name.*PascalCase.*special.*characters/i` |
| `TestNode_NameNotPascalCase` | `unit` | Rejects Node with Name not in PascalCase. | | `Name="reviewStep"`, `Type="human"` | Returns error; error message matches `/name.*PascalCase/i` |

### Validation Failures — Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_EmptyType` | `unit` | Rejects Node with empty Type. | | `Name="TestNode"`, `Type=""` | Returns error; error message matches `/type.*required/i` |
| `TestNode_InvalidType` | `unit` | Rejects Node with invalid Type value. | | `Name="TestNode"`, `Type="service"` | Returns error; error message matches `/type.*agent.*human/i` |
| `TestNode_CaseSensitiveType` | `unit` | Rejects Node with incorrect Type casing. | | `Name="TestNode"`, `Type="Agent"` | Returns error; error message matches `/type.*agent.*human/i` |

### Validation Failures — AgentRole

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_AgentTypeWithoutRole` | `unit` | Rejects agent Node without AgentRole. | | `Name="TestNode"`, `Type="agent"`, `AgentRole=""` | Returns error; error message matches `/agent_role.*required.*agent/i` |
| `TestNode_AgentTypeWithOmittedRole` | `unit` | Rejects agent Node with AgentRole omitted. | | `Name="TestNode"`, `Type="agent"`, `AgentRole` omitted | Returns error; error message matches `/agent_role.*required.*agent/i` |
| `TestNode_HumanTypeWithRole` | `unit` | Rejects human Node with AgentRole provided. | | `Name="TestNode"`, `Type="human"`, `AgentRole="Reviewer"` | Returns error; error message matches `/agent_role.*empty.*human/i` |
| `TestNode_NonExistentAgentRole` | `unit` | Rejects agent Node with non-existent agent role. | Agent definition for "NonExistent" does not exist | `Name="TestNode"`, `Type="agent"`, `AgentRole="NonExistent"` | Returns error; error message matches `/agent.*NonExistent.*not found/i` |
| `TestNode_AgentRoleNotPascalCase` | `unit` | Rejects agent Node with AgentRole not in PascalCase. | | `Name="TestNode"`, `Type="agent"`, `AgentRole="architect_reviewer"` | Returns error; error message matches `/agent_role.*PascalCase/i` |

### Validation Failures — Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_DuplicateName` | `unit` | Rejects workflow with duplicate node names. | Workflow already contains node with `Name="Reviewer"` | Add second node with `Name="Reviewer"` | Returns error; error message matches `/duplicate.*node.*name.*Reviewer/i` |

### Happy Path — Workflow Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_AddedToWorkflow` | `unit` | Node successfully added to workflow Nodes array. | Temporary test directory created; empty workflow definition in test directory; all file operations occur within test fixtures | Add node with `Name="TestNode"`, `Type="human"` | Node appears in workflow's `Nodes` array; workflow validation succeeds |
| `TestNode_MultipleNodesInWorkflow` | `unit` | Multiple nodes coexist in workflow. | Temporary test directory created; workflow definition in test directory; all file operations occur within test fixtures | Add nodes: `Name="Agent1"` (agent), `Name="Human1"` (human), `Name="Agent2"` (agent) | All three nodes in workflow's `Nodes` array; all unique |

### Validation Failures — Unreachable Node

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_UnreachableNode` | `unit` | Returns error for node with no incoming transitions (unreachable). | Temporary test directory created; workflow definition in test directory with node "Isolated" having no incoming transitions; all file operations occur within test fixtures | Validate workflow | Returns error message matching `/unreachable.*node.*Isolated/i`; workflow rejected |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_FieldsImmutable` | `unit` | Node fields cannot be modified after creation. | Node instance created | Attempt to modify `Name`, `Type`, `AgentRole`, or `Description` | Field modification attempt fails or has no effect; original values remain |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_ImplementsNodeInterface` | `unit` | Node type implements the expected Node interface or base type. | | Node instance created | Node satisfies Node interface contract (GetName, GetType, GetAgentRole, GetDescription methods) |

### Happy Path — YAML Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_AgentNodeToYAML` | `unit` | Agent Node serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | Node with `Name="Reviewer"`, `Type="agent"`, `AgentRole="Reviewer"`, `Description="Reviews"` | YAML contains `name: "Reviewer"`, `type: "agent"`, `agent_role: "Reviewer"`, `description: "Reviews"` |
| `TestNode_HumanNodeToYAML` | `unit` | Human Node serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | Node with `Name="Approval"`, `Type="human"`, `Description="Approves"` | YAML contains `name: "Approval"`, `type: "human"`, `description: "Approves"`; no `agent_role` field |
| `TestNode_YAMLToAgentNode` | `unit` | YAML deserializes to agent Node correctly. | Temporary test directory created; YAML file in test directory with agent node; all file operations occur within test fixtures | YAML: `name: "Reviewer"`, `type: "agent"`, `agent_role: "Reviewer"` | Node created with matching fields |
| `TestNode_YAMLToHumanNode` | `unit` | YAML deserializes to human Node correctly. | Temporary test directory created; YAML file in test directory with human node; all file operations occur within test fixtures | YAML: `name: "Approval"`, `type: "human"` | Node created with `Type="human"`, `AgentRole=""` |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_ListNodesInWorkflow` | `e2e` | CLI lists all nodes in a workflow. | Temporary test directory created; workflow definition in test directory with 3 nodes; all file operations occur within test fixtures | Execute `spectra workflow nodes list --workflow <workflow-id>` | Command succeeds; output lists all 3 nodes with names and types |
| `TestNode_ValidateWorkflowWithNodes` | `e2e` | CLI validates workflow containing nodes. | Temporary test directory created; valid workflow definition in test directory; all file operations occur within test fixtures | Execute `spectra workflow validate --workflow <workflow-id>` | Command succeeds; no errors reported |
