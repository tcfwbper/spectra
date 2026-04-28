# Test Specification: `agent_definition_loader.go`

## Source File Under Test
`storage/agent_definition_loader.go`

## Test File
`storage/agent_definition_loader_test.go`

---

## `AgentDefinitionLoader`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_New` | `unit` | Creates a new AgentDefinitionLoader with valid ProjectRoot. | Temp dir fixture with `.spectra/agents/` directory | `ProjectRoot=<temp_dir>` | Returns non-nil AgentDefinitionLoader; no error |

### Happy Path — Load

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_ValidDefinition` | `unit` | Loads a well-formed agent definition with all required fields. | Temp dir fixture with `.spectra/agents/Architect.yaml` containing all required fields; `agent_root: "."` directory exists | `AgentRole="Architect"` | Returns valid AgentDefinition with all fields populated; no error |
| `TestAgentDefinitionLoader_Load_WithOptionalTools` | `unit` | Loads agent definition with AllowedTools and DisallowedTools specified. | Temp dir fixture with agent YAML containing `allowed_tools: ["Read", "Write"]` and `disallowed_tools: ["Bash"]` | `AgentRole="Coder"` | Returns AgentDefinition with AllowedTools and DisallowedTools arrays populated; no error |
| `TestAgentDefinitionLoader_Load_WithoutOptionalTools` | `unit` | Loads agent definition without optional tool fields. | Temp dir fixture with agent YAML missing `allowed_tools` and `disallowed_tools` | `AgentRole="Reviewer"` | Returns AgentDefinition with AllowedTools and DisallowedTools as empty arrays; no error |
| `TestAgentDefinitionLoader_Load_RoleWithDigits` | `unit` | Accepts role with digits in valid PascalCase. | Temp dir fixture with agent YAML where `role: "V2Architect"` | `AgentRole="V2Architect"` | Returns AgentDefinition with Role="V2Architect"; no error |
| `TestAgentDefinitionLoader_Load_RoleWithConsecutiveUppercase` | `unit` | Accepts role with consecutive uppercase letters. | Temp dir fixture with agent YAML where `role: "QAReviewer"` | `AgentRole="QAReviewer"` | Returns AgentDefinition with Role="QAReviewer"; no error |
| `TestAgentDefinitionLoader_Load_SingleUppercaseLetter` | `unit` | Accepts single uppercase letter as role. | Temp dir fixture with agent YAML where `role: "A"` | `AgentRole="A"` | Returns AgentDefinition with Role="A"; no error |
| `TestAgentDefinitionLoader_Load_AgentRootDot` | `unit` | Accepts agent_root="." and validates ProjectRoot is a directory. | Temp dir fixture with agent YAML where `agent_root: "."` | `AgentRole="ProjectAgent"` | Returns AgentDefinition with AgentRoot="."; no error |
| `TestAgentDefinitionLoader_Load_AgentRootNestedPath` | `unit` | Accepts nested relative agent_root path. | Temp dir fixture with agent YAML where `agent_root: "agents/subdir"` and directory exists | `AgentRole="NestedAgent"` | Returns AgentDefinition with AgentRoot="agents/subdir"; no error |
| `TestAgentDefinitionLoader_Load_SystemPromptWithYAMLFrontMatter` | `unit` | Passes through system prompt with YAML front matter without validation. | Temp dir fixture with agent YAML where `system_prompt: "---\ntitle: Test\n---\nYou are..."` | `AgentRole="PromptAgent"` | Returns AgentDefinition with SystemPrompt containing YAML front matter; no error |
| `TestAgentDefinitionLoader_Load_InvalidModelValue` | `unit` | Passes through invalid model value without validation. | Temp dir fixture with agent YAML where `model: "invalid-model-123"` | `AgentRole="TestAgent"` | Returns AgentDefinition with Model="invalid-model-123"; no error |
| `TestAgentDefinitionLoader_Load_InvalidEffortValue` | `unit` | Passes through invalid effort value without validation. | Temp dir fixture with agent YAML where `effort: "ultra-mega-high"` | `AgentRole="TestAgent"` | Returns AgentDefinition with Effort="ultra-mega-high"; no error |
| `TestAgentDefinitionLoader_Load_InvalidToolIdentifiers` | `unit` | Passes through invalid tool identifiers without validation. | Temp dir fixture with agent YAML where `allowed_tools: ["Invalid(**)", "Bad@Tool"]` | `AgentRole="TestAgent"` | Returns AgentDefinition with AllowedTools containing invalid identifiers; no error |
| `TestAgentDefinitionLoader_Load_ConflictingToolLists` | `unit` | Allows the same tool in both AllowedTools and DisallowedTools. | Temp dir fixture with agent YAML where `allowed_tools: ["Read"]` and `disallowed_tools: ["Read"]` | `AgentRole="TestAgent"` | Returns AgentDefinition with "Read" in both lists; no error |

### Validation Failures — File Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_FileNotFound` | `unit` | Returns error when agent file does not exist. | Temp dir fixture with `.spectra/agents/` directory but no Architect.yaml | `AgentRole="Architect"` | Returns error matching `"agent definition not found: Architect"` |
| `TestAgentDefinitionLoader_Load_EmptyAgentRole` | `unit` | Returns error when agent role is empty string. | Temp dir fixture with `.spectra/agents/` directory | `AgentRole=""` | Returns error matching `"agent definition not found: "` |

### Validation Failures — File Read Errors

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_PermissionDenied` | `unit` | Returns error when file exists but is not readable. | Temp dir fixture with `.spectra/agents/Architect.yaml` with permissions set to 0000 | `AgentRole="Architect"` | Returns error matching `"failed to read agent definition 'Architect': permission denied"` |

### Validation Failures — YAML Parsing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_EmptyFile` | `unit` | Returns error when agent file is completely empty. | Temp dir fixture with empty `.spectra/agents/Architect.yaml` | `AgentRole="Architect"` | Returns error matching `"failed to parse agent definition 'Architect': EOF"` |
| `TestAgentDefinitionLoader_Load_InvalidYAMLSyntax` | `unit` | Returns error when YAML has syntax errors. | Temp dir fixture with `.spectra/agents/Architect.yaml` containing `"role:\n  - invalid:\nbroken"` | `AgentRole="Architect"` | Returns error matching `"failed to parse agent definition 'Architect': yaml: line"` and includes line/column info |
| `TestAgentDefinitionLoader_Load_UnknownFieldsIgnored` | `unit` | Ignores unknown fields in YAML. | Temp dir fixture with agent YAML containing all required fields plus `custom_metadata: "extra"` | `AgentRole="Architect"` | Returns valid AgentDefinition; unknown fields ignored; no error |

### Validation Failures — Missing Required Fields

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_MissingRole` | `unit` | Returns error when role field is missing or empty. | Temp dir fixture with agent YAML missing `role` field | `AgentRole="Architect"` | Returns error matching `"agent definition 'Architect' validation failed: missing required field 'role'"` |
| `TestAgentDefinitionLoader_Load_MissingModel` | `unit` | Returns error when model field is missing or empty. | Temp dir fixture with agent YAML missing `model` field | `AgentRole="Architect"` | Returns error matching `"agent definition 'Architect' validation failed: missing required field 'model'"` |
| `TestAgentDefinitionLoader_Load_MissingEffort` | `unit` | Returns error when effort field is missing or empty. | Temp dir fixture with agent YAML missing `effort` field | `AgentRole="Architect"` | Returns error matching `"agent definition 'Architect' validation failed: missing required field 'effort'"` |
| `TestAgentDefinitionLoader_Load_MissingSystemPrompt` | `unit` | Returns error when system_prompt field is missing or empty. | Temp dir fixture with agent YAML missing `system_prompt` field | `AgentRole="Architect"` | Returns error matching `"agent definition 'Architect' validation failed: missing required field 'system_prompt'"` |
| `TestAgentDefinitionLoader_Load_MissingAgentRoot` | `unit` | Returns error when agent_root field is missing or empty. | Temp dir fixture with agent YAML missing `agent_root` field | `AgentRole="Architect"` | Returns error matching `"agent definition 'Architect' validation failed: missing required field 'agent_root'"` |

### Validation Failures — Role Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_RoleWithSpaces` | `unit` | Returns error when role contains spaces. | Temp dir fixture with agent YAML where `role: "QA Reviewer"` | `AgentRole="QAReviewer"` | Returns error matching `"agent definition 'QAReviewer' validation failed: role must be PascalCase with no spaces or special characters"` |
| `TestAgentDefinitionLoader_Load_RoleWithUnderscore` | `unit` | Returns error when role contains underscores. | Temp dir fixture with agent YAML where `role: "QA_Reviewer"` | `AgentRole="QA_Reviewer"` | Returns error matching `"agent definition 'QA_Reviewer' validation failed: role must be PascalCase with no spaces or special characters"` |
| `TestAgentDefinitionLoader_Load_RoleWithHyphen` | `unit` | Returns error when role contains hyphens. | Temp dir fixture with agent YAML where `role: "QA-Reviewer"` | `AgentRole="QA-Reviewer"` | Returns error matching `"agent definition 'QA-Reviewer' validation failed: role must be PascalCase with no spaces or special characters"` |
| `TestAgentDefinitionLoader_Load_RoleWithDot` | `unit` | Returns error when role contains dots. | Temp dir fixture with agent YAML where `role: "QA.Reviewer"` | `AgentRole="QA.Reviewer"` | Returns error matching `"agent definition 'QA.Reviewer' validation failed: role must be PascalCase with no spaces or special characters"` |
| `TestAgentDefinitionLoader_Load_RoleStartsLowercase` | `unit` | Returns error when role starts with lowercase letter. | Temp dir fixture with agent YAML where `role: "qaReviewer"` | `AgentRole="qaReviewer"` | Returns error matching `"agent definition 'qaReviewer' validation failed: role must be PascalCase with no spaces or special characters"` |

### Validation Failures — AgentRoot Path

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_AgentRootAbsolutePath` | `unit` | Returns error when agent_root is an absolute path. | Temp dir fixture with agent YAML where `agent_root: "/usr/local/bin"` | `AgentRole="TestAgent"` | Returns error matching `"agent definition 'TestAgent' validation failed: agent_root must be a relative path"` |
| `TestAgentDefinitionLoader_Load_AgentRootWithDriveLetter` | `unit` | Returns error when agent_root contains Windows drive letter. | Temp dir fixture with agent YAML where `agent_root: "C:\\Users\\test"` | `AgentRole="TestAgent"` | Returns error matching `"agent definition 'TestAgent' validation failed: agent_root must be a relative path"` |
| `TestAgentDefinitionLoader_Load_AgentRootDirectoryNotFound` | `unit` | Returns error when agent_root directory does not exist. | Temp dir fixture with agent YAML where `agent_root: "nonexistent/dir"` but directory doesn't exist | `AgentRole="TestAgent"` | Returns error matching `"agent definition 'TestAgent' validation failed: agent_root directory not found:"` and includes absolute path |
| `TestAgentDefinitionLoader_Load_AgentRootIsFile` | `unit` | Returns error when agent_root points to a file instead of directory. | Temp dir fixture with agent YAML where `agent_root: "somefile.txt"` and file exists but is not a directory | `AgentRole="TestAgent"` | Returns error matching `"agent definition 'TestAgent' validation failed: agent_root is not a directory:"` and includes absolute path |
| `TestAgentDefinitionLoader_Load_AgentRootSymlinkToDirectory` | `unit` | Accepts agent_root as symlink to valid directory. | Temp dir fixture with agent YAML where `agent_root: "link"` and symlink points to valid directory | `AgentRole="TestAgent"` | Returns valid AgentDefinition; symlink followed; no error |
| `TestAgentDefinitionLoader_Load_AgentRootSymlinkToNonexistent` | `unit` | Returns error when agent_root is symlink to non-existent path. | Temp dir fixture with agent YAML where `agent_root: "broken_link"` and symlink points to non-existent path | `AgentRole="TestAgent"` | Returns error matching `"agent definition 'TestAgent' validation failed: agent_root directory not found:"` |
| `TestAgentDefinitionLoader_Load_AgentRootUnreadableDirectory` | `unit` | Accepts agent_root directory with no read permissions. | Temp dir fixture with agent YAML where `agent_root: "restricted"` and directory exists with 0000 permissions | `AgentRole="TestAgent"` | Returns valid AgentDefinition; permission check deferred to runtime; no error |

### Validation Failures — Path Injection

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_AgentRoleWithPathTraversal` | `unit` | Handles agent role with path traversal characters. | Temp dir fixture with `.spectra/agents/` directory | `AgentRole="../malicious/agent"` | File read fails; returns error matching `"agent definition not found:"` or accesses unintended file |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_RepeatedCalls` | `unit` | Multiple Load calls with same agent role return identical results. | Temp dir fixture with valid `.spectra/agents/Architect.yaml` | Call `Load("Architect")` three times | All three calls return identical AgentDefinition values; no caching evidence |
| `TestAgentDefinitionLoader_Load_FileModifiedBetweenCalls` | `unit` | Load reflects file changes between calls (no caching). | Temp dir fixture with valid agent YAML | Load once, modify file on disk, Load again | Second Load returns updated content; no caching |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_ConcurrentSameRole` | `race` | Multiple goroutines load the same agent role concurrently. | Temp dir fixture with valid agent YAML | 10 goroutines call `Load("Architect")` simultaneously | All calls succeed with identical results; no data races; no file locking conflicts |
| `TestAgentDefinitionLoader_Load_ConcurrentDifferentRoles` | `race` | Multiple goroutines load different agent roles concurrently. | Temp dir fixture with 5 different agent YAML files | 10 goroutines each load different agents simultaneously | All calls succeed with correct respective results; no data races |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_UsesStorageLayout` | `unit` | Verifies AgentDefinitionLoader calls StorageLayout.GetAgentPath with correct arguments. | Mock StorageLayout; temp dir fixture | `ProjectRoot=<temp_dir>`, `AgentRole="Architect"` | StorageLayout.GetAgentPath called with ProjectRoot and "Architect"; file read from returned path |

### Boundary Values — ProjectRoot

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_New_RelativeProjectRoot` | `unit` | Accepts relative ProjectRoot path. | Temp dir fixture with relative path `.spectra/agents/` | `ProjectRoot="./project"` | AgentDefinitionLoader created; path composition may be relative; no error |
| `TestAgentDefinitionLoader_Load_ProjectRootWithoutSpectraDir` | `unit` | Handles missing .spectra directory. | Temp dir fixture without `.spectra/` directory | `AgentRole="Architect"` | File read fails; returns error matching `"agent definition not found: Architect"` |
