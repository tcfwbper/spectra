# Test Specification: `agent_definition_loader_test.go`

## Source File Under Test
`storage/agent_definition_loader.go`

## Test File
`storage/agent_definition_loader_test.go`

---

## `AgentDefinitionLoader`

### Happy Path — Load

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_ValidDefinition` | `unit` | Loads a well-formed YAML and returns a valid AgentDefinition. | Create temp dir with `.spectra/agents/MyAgent.yaml` containing valid YAML (model, effort, systemPrompt, agentRoot with camelCase keys). Create the agentRoot directory relative to projectRoot. | `agentRole="MyAgent"` | Returns `*AgentDefinition` with Role=`"MyAgent"`, all fields matching YAML content, nil error |
| `TestAgentDefinitionLoader_Load_RoleDerivedFromFilename` | `unit` | Role is derived from the agentRole parameter, not from YAML content. | Create temp dir with `.spectra/agents/Architect.yaml` containing valid YAML (no role field in YAML). Create the agentRoot directory. | `agentRole="Architect"` | Returns `*AgentDefinition` with Role=`"Architect"` |
| `TestAgentDefinitionLoader_Load_AllowedToolsAndDisallowedTools` | `unit` | Parses allowedTools and disallowedTools arrays from YAML. | Create temp dir with `.spectra/agents/Worker.yaml` containing allowedTools and disallowedTools arrays. Create the agentRoot directory. | `agentRole="Worker"` | Returns `*AgentDefinition` with AllowedTools and DisallowedTools matching YAML arrays |
| `TestAgentDefinitionLoader_Load_MissingToolsFields` | `unit` | Missing allowedTools and disallowedTools in YAML results in empty slices. | Create temp dir with `.spectra/agents/Simple.yaml` without allowedTools or disallowedTools fields. Create the agentRoot directory. | `agentRole="Simple"` | Returns `*AgentDefinition` with AllowedTools=`[]` and DisallowedTools=`[]` |
| `TestAgentDefinitionLoader_Load_AgentRootDot` | `unit` | AgentRoot "." resolves to projectRoot and passes validation. | Create temp dir with `.spectra/agents/RootAgent.yaml` containing `agentRoot: "."`. projectRoot itself is the directory. | `agentRole="RootAgent"` | Returns `*AgentDefinition` successfully |
| `TestAgentDefinitionLoader_Load_AgentRootSymlinkToDir` | `unit` | AgentRoot that is a symlink to a valid directory passes validation. | Create temp dir with `.spectra/agents/Linked.yaml`. Create a target directory and a symlink pointing to it as the agentRoot path. | `agentRole="Linked"` | Returns `*AgentDefinition` successfully |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_FileNotFound` | `unit` | Returns not-found error when YAML file does not exist. | Create temp dir with `.spectra/agents/` directory but no YAML file. | `agentRole="Missing"` | Returns error matching `"agent definition not found: Missing"` |
| `TestAgentDefinitionLoader_Load_ReadPermissionDenied` | `unit` | Returns wrapped read error on permission failure. | Create temp dir with `.spectra/agents/Locked.yaml` with permissions `0000`. | `agentRole="Locked"` | Returns error matching `"failed to read agent definition 'Locked':"` containing permission error |
| `TestAgentDefinitionLoader_Load_YamlSyntaxError` | `unit` | Returns parse error for syntactically invalid YAML. | Create temp dir with `.spectra/agents/Bad.yaml` containing invalid YAML (e.g., unclosed quote). | `agentRole="Bad"` | Returns error matching `"failed to parse agent definition 'Bad':"` |
| `TestAgentDefinitionLoader_Load_YamlUnknownField` | `unit` | Returns parse error when YAML contains unknown fields. | Create temp dir with `.spectra/agents/Extra.yaml` containing valid structure plus `customField: value`. | `agentRole="Extra"` | Returns error matching `"failed to parse agent definition 'Extra':"` |
| `TestAgentDefinitionLoader_Load_YamlSnakeCaseField` | `unit` | Rejects YAML with snake_case field names as unknown fields. | Create temp dir with `.spectra/agents/Snake.yaml` using `system_prompt` instead of `systemPrompt`. | `agentRole="Snake"` | Returns error matching `"failed to parse agent definition 'Snake':"` |
| `TestAgentDefinitionLoader_Load_ConstructorValidationFails` | `unit` | Returns wrapped validation error when NewAgentDefinition rejects input. | Create temp dir with `.spectra/agents/NoModel.yaml` containing empty model field. | `agentRole="NoModel"` | Returns error matching `"agent definition 'NoModel' validation failed:"` |
| `TestAgentDefinitionLoader_Load_AgentRootNotExists` | `unit` | Returns error when agentRoot directory does not exist on disk. | Create temp dir with `.spectra/agents/Orphan.yaml` referencing `agentRoot: "missing_dir"`. Do not create that directory. | `agentRole="Orphan"` | Returns error matching `"agent definition 'Orphan' validation failed: agent_root directory not found:"` |
| `TestAgentDefinitionLoader_Load_AgentRootIsFile` | `unit` | Returns error when agentRoot path points to a regular file. | Create temp dir with `.spectra/agents/FileRoot.yaml` referencing `agentRoot: "somefile"`. Create `somefile` as a regular file. | `agentRole="FileRoot"` | Returns error matching `"agent definition 'FileRoot' validation failed: agent_root is not a directory:"` |
| `TestAgentDefinitionLoader_Load_AgentRootSymlinkBroken` | `unit` | Returns error when agentRoot is a symlink to a non-existent path. | Create temp dir with `.spectra/agents/Broken.yaml` referencing `agentRoot: "broken_link"`. Create a dangling symlink at that path. | `agentRole="Broken"` | Returns error matching `"agent definition 'Broken' validation failed: agent_root directory not found:"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_EmptyAgentRole` | `unit` | Returns not-found error when agentRole is empty string. | Create temp dir with `.spectra/agents/` directory. | `agentRole=""` | Returns error matching `"agent definition not found: "` |
| `TestAgentDefinitionLoader_Load_EmptyYamlFile` | `unit` | Returns parse error for zero-byte YAML file. | Create temp dir with `.spectra/agents/Empty.yaml` as an empty file. | `agentRole="Empty"` | Returns error matching `"failed to parse agent definition 'Empty':"` |

### Boundary Values — agentRole

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_PathTraversal` | `unit` | Path separators in agentRole result in file-not-found. | Create temp dir with `.spectra/agents/` directory. | `agentRole="../malicious/agent"` | Returns error (file not found or read error) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_NoCaching` | `unit` | Second Load reflects file changes made after first Load. | Create temp dir with `.spectra/agents/Mutable.yaml` with initial content and agentRoot dir. After first Load, overwrite YAML with different model value. | Two sequential `Load("Mutable")` calls | First returns original model; second returns updated model |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentDefinitionLoader_Load_ConcurrentAccess` | `unit` | Multiple goroutines loading same agent succeed without interference. | Create temp dir with `.spectra/agents/Shared.yaml` containing valid content and agentRoot dir. | Launch multiple goroutines calling `Load("Shared")` concurrently | All goroutines return valid `*AgentDefinition` with nil error |
