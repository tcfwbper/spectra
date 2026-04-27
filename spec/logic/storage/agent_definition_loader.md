# AgentDefinitionLoader

## Overview

AgentDefinitionLoader provides read-only access to agent definition files stored in `.spectra/agents/`. It loads and parses agent YAML files, validates all required fields (including verifying that `agent_root` directory exists), and returns a complete AgentDefinition structure. AgentDefinitionLoader does not create, modify, or delete agent files. It is stateless and does not cache definitions in memory.

## Behavior

1. AgentDefinitionLoader is initialized with a `ProjectRoot` path (absolute path to the directory containing `.spectra`).
2. AgentDefinitionLoader uses StorageLayout to compose the absolute path to agent YAML files.
3. When the `Load(agentRole)` method is called, AgentDefinitionLoader composes the file path using `StorageLayout.GetAgentPath(ProjectRoot, agentRole)`.
4. AgentDefinitionLoader attempts to open and read the agent file. If the file does not exist, it returns an error: `"agent definition not found: <agentRole>"`.
5. AgentDefinitionLoader parses the YAML content into an AgentDefinition structure using a YAML parser (e.g., `gopkg.in/yaml.v3`).
6. If YAML parsing fails (syntax errors, type mismatches), AgentDefinitionLoader returns an error: `"failed to parse agent definition '<agentRole>': <yaml error>"`.
7. After successful parsing, AgentDefinitionLoader validates all required fields are non-empty: `Role`, `Model`, `Effort`, `SystemPrompt`, `AgentRoot`.
8. If any required field is missing or empty, AgentDefinitionLoader returns an error: `"agent definition '<agentRole>' validation failed: missing required field '<field_name>'"`.
9. AgentDefinitionLoader validates that `Role` follows PascalCase naming conventions. The role must match the pattern `^[A-Z][a-zA-Z0-9]*$`: starts with an uppercase letter, followed by any combination of letters (upper or lower case) and digits, with no spaces, underscores, hyphens, or other special characters. If the validation fails, it returns an error: `"agent definition '<agentRole>' validation failed: role must be PascalCase with no spaces or special characters"`.
10. AgentDefinitionLoader validates that `AgentRoot` is a relative path (does not start with `/` or contain drive letters like `C:`). If it is absolute, it returns an error: `"agent definition '<agentRole>' validation failed: agent_root must be a relative path"`.
11. AgentDefinitionLoader composes the absolute path to `agent_root` by joining `ProjectRoot` with the `AgentRoot` value using `filepath.Join`.
12. AgentDefinitionLoader checks if the `agent_root` directory exists using `os.Stat`. If the directory does not exist, it returns an error: `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"`.
13. AgentDefinitionLoader validates that the `agent_root` path is a directory (not a regular file). If it is a file, it returns an error: `"agent definition '<agentRole>' validation failed: agent_root is not a directory: <absolutePath>"`.
14. `AllowedTools` and `DisallowedTools` are optional fields and default to empty arrays if not present. AgentDefinitionLoader does not validate tool names or conflicts between allowed and disallowed tools (this is the responsibility of Claude CLI).
15. AgentDefinitionLoader does not validate the content of `SystemPrompt` (e.g., does not check for YAML front matter). The prompt content is passed to Claude CLI as-is for interpretation.
16. If all validations pass, AgentDefinitionLoader returns the parsed AgentDefinition structure.
17. AgentDefinitionLoader does not cache agent definitions. Each `Load()` call reads from disk.
18. AgentDefinitionLoader is safe to call concurrently from multiple goroutines. No file locking is required as all operations are read-only.
19. AgentDefinitionLoader does not use FileAccessor because it does not need to prepare files. It directly uses `os.ReadFile()` or equivalent for read operations.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### For Load Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| AgentRole | string | Non-empty, PascalCase identifier, corresponds to `<AgentRole>.yaml` | Yes |

## Outputs

### For Load Method

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| AgentDefinition | AgentDefinition struct | Fully parsed and validated agent definition structure |

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"agent definition not found: <agentRole>"` | The agent YAML file does not exist at the expected path |
| `"failed to read agent definition '<agentRole>': <error>"` | File read operation failed (e.g., permission denied) |
| `"failed to parse agent definition '<agentRole>': <yaml error>"` | YAML parsing failed due to syntax errors or type mismatches |
| `"agent definition '<agentRole>' validation failed: missing required field '<field_name>'"` | A required field is missing or empty after parsing |
| `"agent definition '<agentRole>' validation failed: role must be PascalCase with no spaces or special characters"` | Role field contains invalid characters or format |
| `"agent definition '<agentRole>' validation failed: agent_root must be a relative path"` | AgentRoot is an absolute path |
| `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"` | AgentRoot directory does not exist |
| `"agent definition '<agentRole>' validation failed: agent_root is not a directory: <absolutePath>"` | AgentRoot path points to a file instead of a directory |

## Invariants

1. **Read-Only**: AgentDefinitionLoader must never create, modify, or delete agent definition files.

2. **No FileAccessor**: AgentDefinitionLoader must not use FileAccessor because it does not prepare files. It directly reads existing files using standard file I/O.

3. **No Caching**: AgentDefinitionLoader must not cache agent definitions in memory. Each `Load()` call must read from disk.

4. **Stateless**: AgentDefinitionLoader must not maintain any internal state between `Load()` calls. It is safe to call `Load()` multiple times with different agent roles.

5. **Thread-Safe**: AgentDefinitionLoader must be safe to call concurrently from multiple goroutines without synchronization.

6. **No File Locking**: AgentDefinitionLoader must not acquire file locks. Read-only operations on immutable files do not require locking.

7. **Complete Validation**: AgentDefinitionLoader must validate all required fields and constraints defined in `logic/components/agent_definition.md` before returning an AgentDefinition.

8. **AgentRoot Existence Check**: AgentDefinitionLoader must verify that the `agent_root` directory exists at load time. If the directory does not exist, it must return an error.

9. **Detailed Error Messages**: All validation errors must include the agent role and the specific issue (field name, path, etc.).

10. **Fail Fast Validation**: If any validation fails, AgentDefinitionLoader must return an error immediately without attempting to proceed. Only the first validation error is reported. Users must fix errors iteratively. This design simplifies error handling and prevents cascading errors from obscuring the root cause.

11. **Path Composition Delegation**: AgentDefinitionLoader must delegate agent file path composition to StorageLayout. It must construct `agent_root` absolute paths by joining `ProjectRoot` with the relative `AgentRoot` value.

12. **YAML Parsing Only**: AgentDefinitionLoader must only support YAML format. It must not attempt to parse other formats (JSON, TOML, etc.).

13. **Tool List Passthrough**: AgentDefinitionLoader must not validate `AllowedTools` or `DisallowedTools` contents. These are passed to Claude CLI as-is.

14. **Model and Effort Passthrough**: AgentDefinitionLoader must not validate `Model` or `Effort` values. These are passed to Claude CLI as-is.

15. **SystemPrompt Content Passthrough**: AgentDefinitionLoader must not validate the content of `SystemPrompt` (including checks for YAML front matter, special characters, or formatting). The prompt is passed to Claude CLI as-is for interpretation.

16. **PascalCase Validation Pattern**: Role validation must follow the pattern `^[A-Z][a-zA-Z0-9]*$`. This pattern allows consecutive uppercase letters (e.g., `QAReviewer`), mixed case (e.g., `QaReviewer`), and digits (e.g., `V2Architect`), but prohibits spaces, underscores, hyphens, and other special characters.

## Edge Cases

- **Condition**: Agent file `.spectra/agents/<agentRole>.yaml` does not exist.
  **Expected**: AgentDefinitionLoader returns an error: `"agent definition not found: <agentRole>"`.

- **Condition**: Agent file exists but is empty.
  **Expected**: YAML parser fails, AgentDefinitionLoader returns: `"failed to parse agent definition '<agentRole>': EOF"`.

- **Condition**: Agent file contains invalid YAML syntax (e.g., incorrect indentation).
  **Expected**: YAML parser fails with line/column info, AgentDefinitionLoader returns: `"failed to parse agent definition '<agentRole>': yaml: line 3: mapping values are not allowed in this context"`.

- **Condition**: Agent file has valid YAML but missing required field `Role`.
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: missing required field 'role'"`.

- **Condition**: `Role` field contains spaces (e.g., `"QA Reviewer"`).
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: role must be PascalCase with no spaces or special characters"`.

- **Condition**: `Role` field contains underscores (e.g., `"QA_Reviewer"`).
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: role must be PascalCase with no spaces or special characters"`.

- **Condition**: `Role` field contains hyphens or other special characters (e.g., `"QA-Reviewer"`, `"QA.Reviewer"`).
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: role must be PascalCase with no spaces or special characters"`.

- **Condition**: `Role` field is valid PascalCase with consecutive uppercase letters (e.g., `"QAReviewer"`).
  **Expected**: AgentDefinitionLoader accepts this as valid.

- **Condition**: `Role` field contains digits (e.g., `"V2Architect"`, `"Q1Reviewer"`).
  **Expected**: AgentDefinitionLoader accepts this as valid (digits are allowed after the first character).

- **Condition**: `Role` field starts with a lowercase letter (e.g., `"qaReviewer"`).
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: role must be PascalCase with no spaces or special characters"`.

- **Condition**: `Role` field is a single uppercase letter (e.g., `"A"`).
  **Expected**: AgentDefinitionLoader accepts this as valid (matches pattern `^[A-Z][a-zA-Z0-9]*$`).

- **Condition**: `Model` field is empty.
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: missing required field 'model'"`.

- **Condition**: `Model` field contains an invalid Claude model identifier (e.g., `"invalid-model"`).
  **Expected**: AgentDefinitionLoader allows this. The model value is passed to Claude CLI as-is. Claude CLI will return an error during agent invocation if the model is invalid.

- **Condition**: `Effort` field is empty.
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: missing required field 'effort'"`.

- **Condition**: `Effort` field contains an invalid value (e.g., `"ultra-mega-high"`).
  **Expected**: AgentDefinitionLoader allows this. The effort value is passed to Claude CLI as-is. Claude CLI will return an error during agent invocation if the effort is invalid.

- **Condition**: `SystemPrompt` field is empty.
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: missing required field 'system_prompt'"`.

- **Condition**: `SystemPrompt` contains YAML front matter (e.g., `"---\ntitle: Prompt\n---\nYou are..."`).
  **Expected**: AgentDefinitionLoader allows this. The prompt content is passed to Claude CLI as-is without validation. Claude CLI is responsible for interpreting or rejecting the content.

- **Condition**: `AgentRoot` is an absolute path (e.g., `"/usr/local/bin"`).
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: agent_root must be a relative path"`.

- **Condition**: `AgentRoot` is a relative path but the directory does not exist.
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"`.

- **Condition**: `AgentRoot` points to a regular file instead of a directory.
  **Expected**: AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: agent_root is not a directory: <absolutePath>"`.

- **Condition**: `AgentRoot` is `"."` (current directory).
  **Expected**: AgentDefinitionLoader validates that `ProjectRoot` itself is a directory. If valid, the AgentDefinition is returned successfully.

- **Condition**: `AllowedTools` and `DisallowedTools` are both empty arrays.
  **Expected**: AgentDefinitionLoader allows this. Both fields default to empty arrays if not present or are explicitly empty.

- **Condition**: `AllowedTools` and `DisallowedTools` both contain the same tool identifier.
  **Expected**: AgentDefinitionLoader allows this. Tool list conflicts are not validated. Claude CLI is responsible for resolving conflicts.

- **Condition**: `AllowedTools` or `DisallowedTools` contain invalid tool identifiers (e.g., `"InvalidTool(**)"`).
  **Expected**: AgentDefinitionLoader allows this. Tool identifiers are not validated. Claude CLI will handle invalid tools during agent execution.

- **Condition**: `AllowedTools` and `DisallowedTools` fields are missing from YAML.
  **Expected**: AgentDefinitionLoader sets both fields to empty arrays (default values).

- **Condition**: `AgentRole` contains path separators (e.g., `"../malicious/agent"`).
  **Expected**: StorageLayout composes the path as-is, potentially pointing outside `.spectra/agents/`. File read fails with "not found" or accesses unintended file. AgentDefinitionLoader does not validate agent role format; caller is responsible.

- **Condition**: `AgentRole` is an empty string.
  **Expected**: StorageLayout produces a malformed path. File read fails with an error: `"agent definition not found: "`.

- **Condition**: File read operation fails with "permission denied".
  **Expected**: AgentDefinitionLoader returns: `"failed to read agent definition '<agentRole>': permission denied"`.

- **Condition**: Multiple goroutines call `Load()` for the same agent role simultaneously.
  **Expected**: Both goroutines independently read and parse the file. Both succeed without interference. No file locking is used.

- **Condition**: Multiple goroutines call `Load()` for different agent roles simultaneously.
  **Expected**: All goroutines succeed independently. AgentDefinitionLoader is stateless and thread-safe.

- **Condition**: User modifies the agent file while a `Load()` operation is in progress.
  **Expected**: The `Load()` operation reads whatever file state the OS provides (may be partially old, partially new depending on OS buffering). No guarantees about consistency. This is acceptable as agent definitions should not be modified during runtime.

- **Condition**: User modifies the agent file between two `Load()` calls.
  **Expected**: The second `Load()` call returns the updated agent definition. AgentDefinitionLoader does not cache, so changes are immediately reflected.

- **Condition**: YAML file contains fields not defined in AgentDefinition structure (e.g., user-defined metadata).
  **Expected**: YAML parser ignores unknown fields (assuming standard unmarshaling behavior). AgentDefinitionLoader returns the defined fields only.

- **Condition**: `AgentRoot` directory exists but is not readable (permission denied).
  **Expected**: `os.Stat` succeeds (stat only checks metadata, not read permissions). AgentDefinitionLoader returns the AgentDefinition successfully. Claude CLI will encounter permission errors when attempting to change to the directory during agent execution.

- **Condition**: `AgentRoot` directory is a symbolic link to a valid directory.
  **Expected**: `os.Stat` follows the symbolic link and verifies the target is a directory. AgentDefinitionLoader returns the AgentDefinition successfully.

- **Condition**: `AgentRoot` directory is a symbolic link to a non-existent path.
  **Expected**: `os.Stat` fails with "not found". AgentDefinitionLoader returns: `"agent definition '<agentRole>' validation failed: agent_root directory not found: <absolutePath>"`.

- **Condition**: `ProjectRoot` is a relative path (e.g., `"./project"`).
  **Expected**: AgentDefinitionLoader uses it as-is. Path composition may produce incorrect paths. Caller is responsible for providing an absolute `ProjectRoot`.

- **Condition**: `ProjectRoot` does not contain a `.spectra/` directory.
  **Expected**: File read fails with "not found" error. AgentDefinitionLoader does not validate the existence of `.spectra/`.

## Related

- [AgentDefinition](../components/agent_definition.md) - Defines the AgentDefinition structure and validation rules
- [StorageLayout](./storage_layout.md) - Provides path composition for agent files
- [WorkflowDefinitionLoader](./workflow_definition_loader.md) - Similar loader for workflow definitions
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
