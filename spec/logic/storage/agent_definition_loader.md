# AgentDefinitionLoader

## Overview

AgentDefinitionLoader provides read-only access to agent definition YAML files stored in `.spectra/agents/`. It reads and parses a single agent YAML file, derives the `Role` from the filename, constructs an AgentDefinition via the `NewAgentDefinition` constructor, and performs I/O-layer-specific validation (AgentRoot directory existence). AgentDefinitionLoader does not create, modify, or delete agent files. It is stateless and does not cache definitions in memory.

## Boundaries

- Owns: reading agent YAML files from disk via standard file I/O.
- Owns: strict YAML parsing with unknown field rejection.
- Owns: deriving the `Role` value from the YAML filename (stripping `.yaml` extension).
- Owns: AgentRoot directory existence and type validation (must exist and be a directory).
- Owns: error wrapping with layered context (read / parse / validation phases).
- Delegates: all field-level format validation (PascalCase, non-empty, relative path format) to `NewAgentDefinition` constructor.
- Delegates: path composition to StorageLayout.
- Delegates: project root discovery to the caller (receives `projectRoot` as input).
- Must not: create, modify, or delete agent definition files.
- Must not: bypass `NewAgentDefinition` constructor (must not use struct literals or direct field assignment).
- Must not: cache agent definitions in memory.
- Must not: validate Model, Effort, SystemPrompt, AllowedTools, or DisallowedTools content (passthrough to Claude CLI).
- Must not: use FileAccessor (read-only operation, no file preparation needed).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| StorageLayout | Path composition | `GetAgentPath(projectRoot, agentRole)` | Must not call any other StorageLayout function |
| `NewAgentDefinition` | Constructor | Call with parsed fields to obtain validated `*AgentDefinition` | Must not construct AgentDefinition via struct literal |
| OS filesystem | File reading | `os.ReadFile()`, `os.Stat()` | Must not write, create, or delete |

Construction constraint: AgentDefinitionLoader is initialized with a `projectRoot` string. No struct literal bypass is needed for the loader itself (it is a simple struct with a single field), but it must use `NewAgentDefinition` for its output.

## Behavior

1. AgentDefinitionLoader is initialized with a `projectRoot` path (absolute path to the directory containing `.spectra`).
2. When `Load(agentRole)` is called, composes the file path using `StorageLayout.GetAgentPath(projectRoot, agentRole)`.
3. Reads the file using `os.ReadFile`. If the file does not exist (`os.ErrNotExist`), returns error: `"agent definition not found: <agentRole>"`. If another read error occurs (e.g., permission denied), returns error: `"failed to read agent definition '<agentRole>': <error>"`.
4. Parses the YAML content using a strict YAML decoder with unknown field rejection enabled (`yaml.Decoder` with `KnownFields(true)`). If parsing fails, returns error: `"failed to parse agent definition '<agentRole>': <yaml error>"`.
5. Derives the `Role` value from the `agentRole` parameter (which corresponds to the filename without `.yaml` extension). The Role is not read from YAML content.
6. Calls `NewAgentDefinition(role, model, effort, systemPrompt, agentRoot, allowedTools, disallowedTools)` with the derived Role and parsed YAML fields. If the constructor returns an error, returns error: `"agent definition '<agentRole>' validation failed: <constructor error>"`.
7. After successful construction, validates that the AgentRoot directory exists on the filesystem. Composes the absolute path by joining `projectRoot` with the `AgentRoot` value from the constructed AgentDefinition using `filepath.Join`.
8. Calls `os.Stat` on the composed AgentRoot path. If the path does not exist, returns error: `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"`.
9. If the path exists but is not a directory, returns error: `"agent definition '<agentRole>' validation failed: agent_root is not a directory: <absolutePath>"`.
10. If all validations pass, returns the constructed `*AgentDefinition`.
11. Each `Load()` call reads from disk independently. No caching.
12. AgentDefinitionLoader is safe to call concurrently from multiple goroutines. No file locking is required.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### For Load Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| agentRole | string | Non-empty, corresponds to `<agentRole>.yaml` filename | Yes |

## Outputs

### For Load Method

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| AgentDefinition | *AgentDefinition | Fully constructed and validated agent definition |

**Error Cases**:

| Error Message Format | Phase | Description |
|---------------------|-------|-------------|
| `"agent definition not found: <agentRole>"` | Read | YAML file does not exist |
| `"failed to read agent definition '<agentRole>': <error>"` | Read | File read failed (permission denied, etc.) |
| `"failed to parse agent definition '<agentRole>': <yaml error>"` | Parse | YAML parsing failed (syntax error, type mismatch, unknown field) |
| `"agent definition '<agentRole>' validation failed: <constructor error>"` | Validation | NewAgentDefinition constructor rejected the input |
| `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"` | Validation | AgentRoot directory does not exist |
| `"agent definition '<agentRole>' validation failed: agent_root is not a directory: <absolutePath>"` | Validation | AgentRoot path is not a directory |

## Invariants

1. **Read-Only**: Must never create, modify, or delete agent definition files.
2. **Constructor Required**: Must construct AgentDefinition exclusively via `NewAgentDefinition`. Direct struct literal is forbidden.
3. **No Caching**: Each `Load()` call reads from disk. No in-memory caching.
4. **Stateless**: No internal state between `Load()` calls.
5. **Thread-Safe**: Safe to call concurrently without synchronization.
6. **No File Locking**: Read-only operations do not require locks.
7. **Strict YAML Parsing**: Unknown fields in YAML must be rejected (KnownFields mode).
8. **Role From Filename**: Role is derived from the filename parameter, never from YAML content.
9. **Fail Fast**: Returns immediately on first error. Does not attempt further validation after a failure.
10. **Path Composition Delegation**: Must use StorageLayout for agent file path composition.
11. **AgentRoot Existence Check**: Must verify AgentRoot directory exists after successful construction.
12. **Layered Error Context**: Errors must include phase context (read/parse/validation) and the agent role.
13. **YAML camelCase Fields**: YAML fields use camelCase naming (e.g., `systemPrompt`, `agentRoot`, `allowedTools`, `disallowedTools`). Struct tags enforce this mapping.

## Edge Cases

- Condition: Agent file `.spectra/agents/<agentRole>.yaml` does not exist.
  Expected: Returns error `"agent definition not found: <agentRole>"`.

- Condition: Agent file exists but is empty (0 bytes).
  Expected: YAML parse fails. Returns `"failed to parse agent definition '<agentRole>': EOF"`.

- Condition: YAML contains an unknown field (e.g., `customField: value`).
  Expected: Strict parser rejects it. Returns `"failed to parse agent definition '<agentRole>': <unknown field error>"`.

- Condition: YAML uses snake_case field names (e.g., `system_prompt` instead of `systemPrompt`).
  Expected: Strict parser treats as unknown field and rejects. Returns parse error.

- Condition: YAML is syntactically invalid (bad indentation, unclosed quotes).
  Expected: Returns `"failed to parse agent definition '<agentRole>': <yaml syntax error>"`.

- Condition: YAML is valid but `model` field is missing (empty string after parse).
  Expected: `NewAgentDefinition` constructor returns `"model cannot be empty"`. Loader wraps as: `"agent definition '<agentRole>' validation failed: model cannot be empty"`.

- Condition: Parsed `agentRoot` is an absolute path (e.g., `/usr/local`).
  Expected: `NewAgentDefinition` constructor returns `"agent_root must be a relative path"`. Loader wraps as validation error.

- Condition: Constructor succeeds but `AgentRoot` directory does not exist on disk.
  Expected: Returns `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"`.

- Condition: `AgentRoot` path points to a regular file instead of a directory.
  Expected: Returns `"agent definition '<agentRole>' validation failed: agent_root is not a directory: <absolutePath>"`.

- Condition: `AgentRoot` is `"."` (current directory) and projectRoot is a valid directory.
  Expected: `filepath.Join(projectRoot, ".")` resolves to projectRoot. Stat succeeds. Returns AgentDefinition successfully.

- Condition: `AgentRoot` directory is a symbolic link to a valid directory.
  Expected: `os.Stat` follows symlink. Succeeds. Returns AgentDefinition.

- Condition: `AgentRoot` directory is a symbolic link to a non-existent path.
  Expected: `os.Stat` fails. Returns `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"`.

- Condition: File read fails with permission denied.
  Expected: Returns `"failed to read agent definition '<agentRole>': permission denied"`.

- Condition: `agentRole` contains path separators (e.g., `"../malicious/agent"`).
  Expected: StorageLayout composes path as-is. File read likely fails with not-found. Loader does not validate agentRole format; caller is responsible.

- Condition: `agentRole` is an empty string.
  Expected: StorageLayout produces malformed path. File read fails. Returns `"agent definition not found: "`.

- Condition: `allowedTools` and `disallowedTools` are missing from YAML.
  Expected: Parsed as nil. Constructor normalizes to empty slices. Succeeds.

- Condition: Multiple goroutines call `Load()` for the same agent role simultaneously.
  Expected: Both independently read and parse. Both succeed without interference.

- Condition: YAML file is modified between two `Load()` calls.
  Expected: Second call reflects updated content. No caching.

- Condition: `ProjectRoot` does not contain a `.spectra/` directory.
  Expected: File read fails with not-found error.

## Related

- [AgentDefinition](../components/agent_definition.md) — Value object constructed by this loader; owns field validation
- [StorageLayout](./storage_layout.md) — Provides path composition for agent files
- [WorkflowDefinitionLoader](./workflow_definition_loader.md) — Consumes AgentDefinitionLoader for agent_role referential integrity
