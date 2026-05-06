# Test Specification: `node_test.go`

## Source File Under Test
`components/node.go`

## Test File
`components/node_test.go`

---

## `Node`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewNode_AgentTypeValid` | `unit` | Constructs an agent-type Node with all valid fields. | | `name="ReviewStep"`, `nodeType="agent"`, `agentRole="Architect"`, `description="Reviews code"` | Returns no error; all getters return the provided values |
| `TestNewNode_HumanTypeValid` | `unit` | Constructs a human-type Node with valid fields and empty AgentRole. | | `name="HumanApproval"`, `nodeType="human"`, `agentRole=""`, `description="Waits for approval"` | Returns no error; all getters return the provided values |
| `TestNewNode_EmptyDescription` | `unit` | Accepts empty string Description as valid. | | `name="Draft"`, `nodeType="agent"`, `agentRole="Writer"`, `description=""` | Returns no error; Description getter returns `""` |

### Validation Failures — Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewNode_EmptyName` | `unit` | Rejects empty string Name. | | `name=""` with other fields valid | Returns error `"node name cannot be empty"` |
| `TestNewNode_NameStartsLowercase` | `unit` | Rejects Name starting with a lowercase letter. | | `name="reviewStep"` with other fields valid | Returns error indicating PascalCase is required |
| `TestNewNode_NameContainsHyphen` | `unit` | Rejects Name with non-alphanumeric characters (hyphen). | | `name="Review-Step"` with other fields valid | Returns error indicating PascalCase is required |
| `TestNewNode_NameContainsUnderscore` | `unit` | Rejects Name with non-alphanumeric characters (underscore). | | `name="review_step"` with other fields valid | Returns error indicating PascalCase is required |

### Validation Failures — Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewNode_EmptyType` | `unit` | Rejects empty string Type. | | `nodeType=""` with other fields valid | Returns error `"node type must be 'agent' or 'human'"` |
| `TestNewNode_InvalidType` | `unit` | Rejects unrecognized Type value. | | `nodeType="bot"` with other fields valid | Returns error `"node type must be 'agent' or 'human'"` |

### Validation Failures — AgentRole

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewNode_AgentTypeEmptyRole` | `unit` | Rejects agent-type Node with empty AgentRole. | | `name="Review"`, `nodeType="agent"`, `agentRole=""`, `description=""` | Returns error `"agent_role is required when type is 'agent'"` |
| `TestNewNode_AgentTypeRoleStartsLowercase` | `unit` | Rejects agent-type Node with AgentRole starting lowercase. | | `name="Review"`, `nodeType="agent"`, `agentRole="architect"`, `description=""` | Returns error indicating AgentRole must be PascalCase |
| `TestNewNode_HumanTypeNonEmptyRole` | `unit` | Rejects human-type Node with non-empty AgentRole. | | `name="Approval"`, `nodeType="human"`, `agentRole="Architect"`, `description=""` | Returns error `"agent_role must be empty when type is 'human'"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNode_Immutability` | `unit` | All fields remain unchanged after construction; no exported setters exist. | Construct a valid Node with `name="Draft"`, `nodeType="agent"`, `agentRole="Writer"`, `description="Drafts content"` | Access all getters after construction | All getter values remain identical to construction inputs |
