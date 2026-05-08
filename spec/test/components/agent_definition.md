# Test Specification: `agent_definition_test.go`

## Source File Under Test
`components/agent_definition.go`

## Test File
`components/agent_definition_test.go`

---

## `AgentDefinition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentDefinition_AllFieldsValid` | `unit` | Constructs an AgentDefinition with all valid fields including non-empty tool slices. | | `role="DefaultArchitect"`, `model="claude-sonnet-4-20250514"`, `effort="high"`, `systemPrompt="You are an architect."`, `agentRoot="spec"`, `allowedTools=[]string{"Bash","Read"}`, `disallowedTools=[]string{"Write"}` | Returns no error; all getters return the provided values |
| `TestNewAgentDefinition_EmptyAllowedTools` | `unit` | Accepts empty slice for AllowedTools. | | `role="Writer"`, `model="claude-sonnet-4-20250514"`, `effort="low"`, `systemPrompt="Write docs."`, `agentRoot="docs"`, `allowedTools=[]string{}`, `disallowedTools=[]string{}` | Returns no error; AllowedTools getter returns `[]string{}` |
| `TestNewAgentDefinition_NilAllowedTools` | `unit` | Normalizes nil AllowedTools to empty slice. | | `role="Writer"`, `model="claude-sonnet-4-20250514"`, `effort="low"`, `systemPrompt="Write docs."`, `agentRoot="docs"`, `allowedTools=nil`, `disallowedTools=nil` | Returns no error; AllowedTools getter returns `[]string{}`; DisallowedTools getter returns `[]string{}` |
| `TestNewAgentDefinition_AgentRootDot` | `unit` | Accepts `"."` as a valid relative AgentRoot. | | `role="Root"`, `model="claude-sonnet-4-20250514"`, `effort="medium"`, `systemPrompt="Prompt."`, `agentRoot="."`, `allowedTools=nil`, `disallowedTools=nil` | Returns no error; AgentRoot getter returns `"."` |
| `TestNewAgentDefinition_AgentRootNestedPath` | `unit` | Accepts a nested relative path for AgentRoot. | | `role="Nested"`, `model="claude-sonnet-4-20250514"`, `effort="medium"`, `systemPrompt="Prompt."`, `agentRoot="src/internal/pkg"`, `allowedTools=nil`, `disallowedTools=nil` | Returns no error; AgentRoot getter returns `"src/internal/pkg"` |
| `TestNewAgentDefinition_OverlappingTools` | `unit` | Accepts same tool in both AllowedTools and DisallowedTools without error. | | `role="Tester"`, `model="claude-sonnet-4-20250514"`, `effort="high"`, `systemPrompt="Test."`, `agentRoot="test"`, `allowedTools=[]string{"Bash"}`, `disallowedTools=[]string{"Bash"}` | Returns no error; both slices contain `"Bash"` |
| `TestNewAgentDefinition_UnrecognizedModelAccepted` | `unit` | Accepts unrecognized model string without validation (passthrough). | | `role="Agent"`, `model="invalid-model"`, `effort="high"`, `systemPrompt="Prompt."`, `agentRoot="src"`, `allowedTools=nil`, `disallowedTools=nil` | Returns no error; Model getter returns `"invalid-model"` |

### Validation Failures — Role

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentDefinition_EmptyRole` | `unit` | Rejects empty string Role. | | `role=""` with other fields valid | Returns error `"role cannot be empty"` |
| `TestNewAgentDefinition_RoleStartsLowercase` | `unit` | Rejects Role starting with a lowercase letter. | | `role="defaultArchitect"` with other fields valid | Returns error `"role must be PascalCase (start with uppercase, alphanumeric only)"` |
| `TestNewAgentDefinition_RoleContainsHyphen` | `unit` | Rejects Role with non-alphanumeric characters (hyphen). | | `role="Default-Architect"` with other fields valid | Returns error `"role must be PascalCase (start with uppercase, alphanumeric only)"` |
| `TestNewAgentDefinition_RoleContainsUnderscore` | `unit` | Rejects Role with non-alphanumeric characters (underscore). | | `role="default_architect"` with other fields valid | Returns error `"role must be PascalCase (start with uppercase, alphanumeric only)"` |

### Validation Failures — Model

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentDefinition_EmptyModel` | `unit` | Rejects empty string Model. | | `model=""` with other fields valid | Returns error `"model cannot be empty"` |

### Validation Failures — Effort

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentDefinition_EmptyEffort` | `unit` | Rejects empty string Effort. | | `effort=""` with other fields valid | Returns error `"effort cannot be empty"` |

### Validation Failures — SystemPrompt

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentDefinition_EmptySystemPrompt` | `unit` | Rejects empty string SystemPrompt. | | `systemPrompt=""` with other fields valid | Returns error `"system_prompt cannot be empty"` |

### Validation Failures — AgentRoot

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewAgentDefinition_EmptyAgentRoot` | `unit` | Rejects empty string AgentRoot. | | `agentRoot=""` with other fields valid | Returns error `"agent_root cannot be empty"` |
| `TestNewAgentDefinition_AgentRootAbsoluteUnix` | `unit` | Rejects AgentRoot starting with `/`. | | `agentRoot="/usr/local"` with other fields valid | Returns error `"agent_root must be a relative path"` |
| `TestNewAgentDefinition_AgentRootDriveLetter` | `unit` | Rejects AgentRoot containing a drive letter prefix. | | `agentRoot="C:\\projects"` with other fields valid | Returns error `"agent_root must be a relative path"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_Immutability` | `unit` | All fields remain unchanged after construction; no exported setters exist. | Construct a valid AgentDefinition with `role="Architect"`, `model="claude-sonnet-4-20250514"`, `effort="high"`, `systemPrompt="You are an architect."`, `agentRoot="spec"`, `allowedTools=[]string{"Bash"}`, `disallowedTools=[]string{"Write"}` | Access all getters after construction | All getter values remain identical to construction inputs |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_AllowedToolsCopySemantics` | `unit` | Mutation of the original AllowedTools slice after construction does not affect the stored value. | Construct a valid AgentDefinition with `allowedTools=[]string{"Bash","Read"}` | Mutate the original slice after construction (e.g., `allowedTools[0]="Mutated"`) | AllowedTools getter still returns `[]string{"Bash","Read"}` |
| `TestAgentDefinition_DisallowedToolsCopySemantics` | `unit` | Mutation of the original DisallowedTools slice after construction does not affect the stored value. | Construct a valid AgentDefinition with `disallowedTools=[]string{"Write"}` | Mutate the original slice after construction (e.g., `disallowedTools[0]="Mutated"`) | DisallowedTools getter still returns `[]string{"Write"}` |
