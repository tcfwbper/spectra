# Test Specification: `agent_definition.go`

## Source File Under Test
`components/agent_definition.go`

## Test File
`components/agent_definition_test.go`

---

## `AgentDefinition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_ValidAgentAllFields` | `unit` | Creates AgentDefinition with all fields provided. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/spec/` directory exists; all file operations occur within test fixtures | `Role="QaReviewer"`, `Model="sonnet"`, `Effort="high"`, `SystemPrompt="You are a QA reviewer"`, `AgentRoot="spec"`, `AllowedTools=["Read(*)"]`, `DisallowedTools=["Bash(spectra *)"]` | Returns valid AgentDefinition; all fields match input; YAML file created at `<test-dir>/.spectra/agents/QaReviewer.yaml` |
| `TestAgentDefinition_EmptyToolLists` | `unit` | Creates AgentDefinition with empty AllowedTools and DisallowedTools. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="DefaultArchitect"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="Example prompt"`, `AgentRoot="."`, `AllowedTools=[]`, `DisallowedTools=[]` | Returns valid AgentDefinition; `AllowedTools=[]`, `DisallowedTools=[]` |
| `TestAgentDefinition_AgentRootCurrentDir` | `unit` | Creates AgentDefinition with AgentRoot set to current directory. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="RootAgent"`, `Model="opus"`, `Effort="low"`, `SystemPrompt="Root agent"`, `AgentRoot="."`, `AllowedTools=[]`, `DisallowedTools=[]` | Returns valid AgentDefinition; `AgentRoot="."` |
| `TestAgentDefinition_MultiLineSystemPrompt` | `unit` | Creates AgentDefinition with multi-line SystemPrompt. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="MultiLineAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="Line 1\nLine 2\nLine 3"`, `AgentRoot="."`, `AllowedTools=[]`, `DisallowedTools=[]` | Returns valid AgentDefinition; `SystemPrompt` preserves newlines |

### Happy Path — Load from YAML

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_LoadValidYAML` | `unit` | Loads AgentDefinition from valid YAML file. | Temporary test directory created; YAML file at `<test-dir>/.spectra/agents/Architect.yaml` with all required fields; `<test-dir>/spec/` directory exists; all file operations occur within test fixtures | Load agent with `Role="Architect"` | Returns valid AgentDefinition; all fields match YAML content |
| `TestAgentDefinition_LoadWithEmptyToolLists` | `unit` | Loads AgentDefinition with empty tool lists from YAML. | Temporary test directory created; YAML file at `<test-dir>/.spectra/agents/Agent.yaml` with `allowed_tools: []`, `disallowed_tools: []`; `<test-dir>/.` directory exists; all file operations occur within test fixtures | Load agent with `Role="Agent"` | Returns valid AgentDefinition; `AllowedTools=[]`, `DisallowedTools=[]` |
| `TestAgentDefinition_LoadWithMultiLinePrompt` | `unit` | Loads AgentDefinition with multi-line SystemPrompt from YAML. | Temporary test directory created; YAML file at `<test-dir>/.spectra/agents/Prompter.yaml` with multi-line `system_prompt` using YAML `\|-` syntax; `<test-dir>/.` directory exists; all file operations occur within test fixtures | Load agent with `Role="Prompter"` | Returns valid AgentDefinition; `SystemPrompt` preserves newlines and formatting |

### Validation Failures — Role

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_EmptyRole` | `unit` | Rejects AgentDefinition with empty Role. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role=""`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/role.*non-empty/i` |
| `TestAgentDefinition_RoleWithSpaces` | `unit` | Rejects AgentDefinition with Role containing spaces. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="Qa Reviewer"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/role.*PascalCase.*spaces.*special.*characters/i` |
| `TestAgentDefinition_RoleWithUnderscores` | `unit` | Rejects AgentDefinition with Role containing underscores. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="Qa_Reviewer"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/role.*PascalCase.*spaces.*special.*characters/i` |
| `TestAgentDefinition_RoleWithHyphens` | `unit` | Rejects AgentDefinition with Role containing hyphens. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="Qa-Reviewer"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/role.*PascalCase.*spaces.*special.*characters/i` |
| `TestAgentDefinition_RoleNotPascalCase` | `unit` | Rejects AgentDefinition with Role not in PascalCase. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="qaReviewer"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/role.*PascalCase/i` |

### Validation Failures — Model

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_EmptyModel` | `unit` | Rejects AgentDefinition with empty Model. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="TestAgent"`, `Model=""`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/model.*non-empty/i` |

### Validation Failures — Effort

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_EmptyEffort` | `unit` | Rejects AgentDefinition with empty Effort. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort=""`, `SystemPrompt="test"`, `AgentRoot="."` | Returns error; error message matches `/effort.*non-empty/i` |

### Validation Failures — SystemPrompt

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_EmptySystemPrompt` | `unit` | Rejects AgentDefinition with empty SystemPrompt. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt=""`, `AgentRoot="."` | Returns error; error message matches `/system_prompt.*non-empty/i` |

### Validation Failures — AgentRoot

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_EmptyAgentRoot` | `unit` | Rejects AgentDefinition with empty AgentRoot. | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot=""` | Returns error; error message matches `/agent_root.*non-empty/i` |
| `TestAgentDefinition_AbsolutePathAgentRoot` | `unit` | Rejects AgentDefinition with absolute path AgentRoot (Unix). | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="/usr/local"` | Returns error; error message matches `/agent_root.*relative.*path/i` |
| `TestAgentDefinition_AbsolutePathAgentRootWindows` | `unit` | Rejects AgentDefinition with absolute path AgentRoot (Windows drive letter). | Temporary test directory created; `.spectra/agents/` directory created in test directory; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="C:\\spectra"` | Returns error; error message matches `/agent_root.*relative.*path/i` |
| `TestAgentDefinition_NonExistentAgentRoot` | `unit` | Rejects AgentDefinition when AgentRoot directory does not exist. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/nonexistent/` directory does NOT exist; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="nonexistent"` | Returns error during agent definition load; error message matches `/agent_root.*directory.*not found/i` |

### Validation Failures — Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_DuplicateRole` | `unit` | Rejects loading multiple AgentDefinitions with same Role. | Temporary test directory created; YAML file at `<test-dir>/.spectra/agents/Architect.yaml`; second YAML loaded from different source with same `role: "Architect"`; all file operations occur within test fixtures | Load both agents with `Role="Architect"` | Second agent load returns error; error message matches `/agent.*Architect.*already exists/i` |

### Validation Failures — File Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_FileDoesNotExist` | `unit` | Returns error when agent YAML file does not exist. | Temporary test directory created; `.spectra/agents/` directory created but empty; all file operations occur within test fixtures | Load agent with `Role="NonExistent"` | Returns error; error message matches `/agent.*not found/i` |

### Validation Failures — Malformed YAML

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_MalformedYAML` | `unit` | Rejects AgentDefinition with malformed YAML syntax. | Temporary test directory created; YAML file at `<test-dir>/.spectra/agents/Broken.yaml` with invalid YAML syntax (unclosed quote); all file operations occur within test fixtures | Load agent with `Role="Broken"` | Returns parse error; error message indicates YAML syntax issue |
| `TestAgentDefinition_MissingRequiredField` | `unit` | Rejects AgentDefinition with missing required field (model). | Temporary test directory created; YAML file at `<test-dir>/.spectra/agents/Incomplete.yaml` missing `model` field; all file operations occur within test fixtures | Load agent with `Role="Incomplete"` | Returns error; error message matches `/model.*required/i` |

### Happy Path — Passthrough to Claude CLI

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_InvalidModelPassthrough` | `unit` | AgentDefinition with invalid Model is loaded without validation; validation deferred to Claude CLI. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="invalid-model"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns valid AgentDefinition; `Model="invalid-model"` stored without validation |
| `TestAgentDefinition_InvalidEffortPassthrough` | `unit` | AgentDefinition with invalid Effort is loaded without validation; validation deferred to Claude CLI. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="super-high"`, `SystemPrompt="test"`, `AgentRoot="."` | Returns valid AgentDefinition; `Effort="super-high"` stored without validation |
| `TestAgentDefinition_InvalidToolsPassthrough` | `unit` | AgentDefinition with invalid tool identifiers is loaded without validation; validation deferred to Claude CLI. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."`, `AllowedTools=["InvalidTool"]` | Returns valid AgentDefinition; `AllowedTools=["InvalidTool"]` stored without validation |
| `TestAgentDefinition_ToolConflictPassthrough` | `unit` | AgentDefinition with same tool in both AllowedTools and DisallowedTools is loaded without validation; conflict resolution deferred to Claude CLI. | Temporary test directory created; `.spectra/agents/` directory created in test directory; `<test-dir>/.` directory exists; all file operations occur within test fixtures | `Role="TestAgent"`, `Model="sonnet"`, `Effort="medium"`, `SystemPrompt="test"`, `AgentRoot="."`, `AllowedTools=["Read(*)"]`, `DisallowedTools=["Read(*)"]` | Returns valid AgentDefinition; both tool lists contain "Read(*)"; no validation error |

### Happy Path — YAML Serialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_ToYAML` | `unit` | AgentDefinition serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | AgentDefinition with all fields populated | YAML contains `role`, `model`, `effort`, `system_prompt`, `agent_root`, `allowed_tools`, `disallowed_tools` with correct values |
| `TestAgentDefinition_ToYAMLEmptyToolLists` | `unit` | AgentDefinition with empty tool lists serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | AgentDefinition with `AllowedTools=[]`, `DisallowedTools=[]` | YAML contains `allowed_tools: []`, `disallowed_tools: []` |
| `TestAgentDefinition_ToYAMLMultiLinePrompt` | `unit` | AgentDefinition with multi-line SystemPrompt serializes to YAML correctly. | Temporary test directory created; YAML output written to test directory; all file operations occur within test fixtures | AgentDefinition with `SystemPrompt="Line 1\nLine 2\nLine 3"` | YAML uses YAML multi-line syntax (`\|-`) and preserves newlines |

### Happy Path — Agent Invocation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_InvokesClaudeWithCorrectArgs` | `unit` | Runtime passes AgentDefinition fields to Claude CLI with correct flags. | Temporary test directory created; AgentDefinition loaded; mock Claude CLI invocation tracked; all file operations occur within test fixtures | Invoke agent with `Role="TestAgent"` during workflow execution | Claude CLI invoked with `--model sonnet`, `--effort medium`, `--system-prompt "test"`, allowed_tools and disallowed_tools flags; working directory changed to `<test-dir>/AgentRoot` |
| `TestAgentDefinition_ChangesWorkingDirectory` | `unit` | Runtime changes working directory to AgentRoot before invoking Claude CLI. | Temporary test directory created; AgentDefinition with `AgentRoot="spec"`; track working directory changes; all file operations occur within test fixtures | Invoke agent with `Role="TestAgent"` | Working directory changed to `<test-dir>/spec` before Claude CLI invocation |

### Happy Path — Built-in Agent Copy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_BuiltinCopiedDuringInit` | `e2e` | Built-in agents are copied to `.spectra/agents/` during `spectra init`. | Temporary test directory created; no `.spectra/agents/` directory exists; all file operations occur within test fixtures | Execute `spectra init` | Built-in agent files copied to `<test-dir>/.spectra/agents/`; files readable and valid |
| `TestAgentDefinition_ExistingAgentNotOverwritten` | `e2e` | Existing agent file is not overwritten during `spectra init`. | Temporary test directory created; `.spectra/agents/QaReviewer.yaml` exists with custom content; all file operations occur within test fixtures | Execute `spectra init` | `QaReviewer.yaml` content unchanged; other built-in agents copied; no error returned |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_FieldsImmutable` | `unit` | AgentDefinition fields cannot be modified after creation. | AgentDefinition instance created | Attempt to modify `Role`, `Model`, `Effort`, `SystemPrompt`, `AgentRoot`, `AllowedTools`, or `DisallowedTools` | Field modification attempt fails or has no effect; original values remain |

### Type Hierarchy

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_ImplementsInterface` | `unit` | AgentDefinition type implements expected interface. | | AgentDefinition instance created | AgentDefinition satisfies AgentDefinition interface contract (GetRole, GetModel, GetEffort, GetSystemPrompt, GetAgentRoot, GetAllowedTools, GetDisallowedTools methods) |

### Happy Path — Workflow Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_ReferencedByNode` | `unit` | AgentDefinition successfully referenced by workflow Node. | Temporary test directory created; AgentDefinition file at `<test-dir>/.spectra/agents/Architect.yaml`; workflow definition with agent Node referencing `AgentRole="Architect"`; all file operations occur within test fixtures | Load workflow and validate nodes | Node with `AgentRole="Architect"` references valid AgentDefinition; workflow validation succeeds |

### Happy Path — CLI Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinition_ListAgents` | `e2e` | CLI lists all available agents. | Temporary test directory created; multiple agent definition files in `<test-dir>/.spectra/agents/`; all file operations occur within test fixtures | Execute `spectra agent list` | Command succeeds; output lists all agents with role names and agent_root paths |
| `TestAgentDefinition_ShowAgentDetails` | `e2e` | CLI shows details for specific agent. | Temporary test directory created; agent definition file at `<test-dir>/.spectra/agents/Architect.yaml`; all file operations occur within test fixtures | Execute `spectra agent show --role Architect` | Command succeeds; output displays all fields: role, model, effort, system_prompt, agent_root, allowed_tools, disallowed_tools |
| `TestAgentDefinition_ValidateAgent` | `e2e` | CLI validates agent definition file. | Temporary test directory created; valid agent definition file at `<test-dir>/.spectra/agents/TestAgent.yaml`; `<test-dir>/spec/` directory exists; all file operations occur within test fixtures | Execute `spectra agent validate --role TestAgent` | Command succeeds; no errors reported |
| `TestAgentDefinition_ValidateAgentInvalidRoot` | `e2e` | CLI validation fails for agent with non-existent AgentRoot. | Temporary test directory created; agent definition file at `<test-dir>/.spectra/agents/BadAgent.yaml` with `agent_root: "nonexistent"`; directory does NOT exist; all file operations occur within test fixtures | Execute `spectra agent validate --role BadAgent` | Command fails; error message matches `/agent_root.*directory.*not found/i` |
