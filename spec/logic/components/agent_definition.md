# AgentDefinition

## Overview

An AgentDefinition describes the metadata for an AI agent role: Claude model configuration (`Model`, `Effort`), system prompt (`SystemPrompt`), tool permissions (`AllowedTools`, `DisallowedTools`), and execution context (`AgentRoot`). It is a pure immutable value object that validates its own field formats at construction time. AgentDefinition does not know about other AgentDefinitions, file systems, or runtime execution — uniqueness, file-to-role mapping, and directory existence are owned by the I/O loader layer.

The `Role` field is not parsed from YAML content. It is derived externally by the loader from the agent definition filename (e.g., `DefaultArchitect.yaml` → Role `DefaultArchitect`) and passed as a construction parameter.

## Boundaries

- Owns: construction-time validation of all fields (Role, Model, Effort, SystemPrompt, AgentRoot, AllowedTools, DisallowedTools).
- Owns: immutability guarantee for all fields after construction.
- Owns: AgentRoot relative-path format validation (not absolute, no drive letter prefix).
- Delegates: Role derivation from filename to the I/O loader layer.
- Delegates: Role uniqueness across agent definitions to the I/O loader layer (filesystem enforces via filename).
- Delegates: AgentRoot directory existence validation to the I/O loader layer.
- Delegates: Model, Effort, SystemPrompt, AllowedTools, and DisallowedTools semantic validation to Claude CLI (passthrough, no interpretation).
- Delegates: agent invocation and working-directory handling to the runtime.
- Must not: perform any I/O, network access, or filesystem operations.
- Must not: reference or import any module outside the `components` package.
- Must not: be constructed via struct literal — must use the provided constructor.
- Must not: parse or derive Role from YAML content.

## Dependencies

None. This value object depends only on Go standard library types.

Construction constraint: Must be constructed via `NewAgentDefinition(...)`. Direct struct literal or field assignment is forbidden.

## Behavior

1. Provides a constructor `NewAgentDefinition(role string, model string, effort string, systemPrompt string, agentRoot string, allowedTools []string, disallowedTools []string) (*AgentDefinition, error)` that validates all fields and returns an immutable AgentDefinition value.
2. Validates that `Role` is a non-empty PascalCase string (starts with uppercase letter, contains only alphanumeric characters).
3. Validates that `Model` is a non-empty string.
4. Validates that `Effort` is a non-empty string.
5. Validates that `SystemPrompt` is a non-empty string.
6. Validates that `AgentRoot` is a non-empty string that does not start with `/` and does not contain a drive letter prefix (e.g., `C:`).
7. Accepts `AllowedTools` as any string slice (including empty or nil, normalized to empty slice).
8. Accepts `DisallowedTools` as any string slice (including empty or nil, normalized to empty slice).
9. Returns a validation error if any constraint is violated. No AgentDefinition is created on failure.
10. All fields are immutable after construction.
11. Exposes all fields via exported getter methods.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Role | string | Non-empty, PascalCase (starts with uppercase letter, alphanumeric only). Provided by loader, derived from filename. | Yes |
| Model | string | Non-empty | Yes |
| Effort | string | Non-empty | Yes |
| SystemPrompt | string | Non-empty | Yes |
| AgentRoot | string | Non-empty, relative path (must not start with `/`, must not contain drive letter prefix) | Yes |
| AllowedTools | []string | Any string slice (nil normalized to empty slice) | Yes (default: `[]`) |
| DisallowedTools | []string | Any string slice (nil normalized to empty slice) | Yes (default: `[]`) |

## Outputs

### AgentDefinition Structure (accessed via getters)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Role | string | Non-empty, PascalCase | Agent role identifier (derived from filename by loader) |
| Model | string | Non-empty | Claude model identifier, passed to Claude CLI as-is |
| Effort | string | Non-empty | Claude effort level, passed to Claude CLI as-is |
| SystemPrompt | string | Non-empty | System prompt, passed to Claude CLI as-is |
| AgentRoot | string | Non-empty, relative path | Working directory for agent execution (relative to spectra root) |
| AllowedTools | []string | Non-nil slice | Tools explicitly allowed, passed to Claude CLI as-is |
| DisallowedTools | []string | Non-nil slice | Tools explicitly disallowed, passed to Claude CLI as-is |

### Error Output

| Condition | Error |
|-----------|-------|
| Role is empty | `"role cannot be empty"` |
| Role is not PascalCase | `"role must be PascalCase (start with uppercase, alphanumeric only)"` |
| Model is empty | `"model cannot be empty"` |
| Effort is empty | `"effort cannot be empty"` |
| SystemPrompt is empty | `"system_prompt cannot be empty"` |
| AgentRoot is empty | `"agent_root cannot be empty"` |
| AgentRoot starts with `/` | `"agent_root must be a relative path"` |
| AgentRoot contains drive letter prefix | `"agent_root must be a relative path"` |

## Invariants

1. **Role PascalCase**: `Role` is always a non-empty PascalCase string after construction.
2. **Model Non-Empty**: `Model` is always a non-empty string after construction.
3. **Effort Non-Empty**: `Effort` is always a non-empty string after construction.
4. **SystemPrompt Non-Empty**: `SystemPrompt` is always a non-empty string after construction.
5. **AgentRoot Relative**: `AgentRoot` is always a non-empty relative path (no leading `/`, no drive letter) after construction.
6. **Tools Non-Nil**: `AllowedTools` and `DisallowedTools` are always non-nil slices after construction (may be empty but never nil).
7. **Immutability**: Once constructed, no field may be modified. All access is via exported getter methods.
8. **Construction Only Via Constructor**: Must be constructed via `NewAgentDefinition`. Direct struct literal construction is forbidden.
9. **No Role in YAML**: The Role value does not originate from YAML file content. It is provided externally by the loader (derived from filename).
10. **Passthrough Semantics**: Model, Effort, SystemPrompt, AllowedTools, and DisallowedTools are stored as-is without interpretation. Semantic validation is Claude CLI's responsibility.

## Edge Cases

- Condition: `Role` is an empty string.
  Expected: Constructor returns error `"role cannot be empty"`. No AgentDefinition is created.

- Condition: `Role` starts with a lowercase letter (e.g., `"defaultArchitect"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `Role` contains non-alphanumeric characters (e.g., `"Default-Architect"`, `"default_architect"`).
  Expected: Constructor returns error indicating PascalCase is required.

- Condition: `Model` is an empty string.
  Expected: Constructor returns error `"model cannot be empty"`.

- Condition: `Effort` is an empty string.
  Expected: Constructor returns error `"effort cannot be empty"`.

- Condition: `SystemPrompt` is an empty string.
  Expected: Constructor returns error `"system_prompt cannot be empty"`.

- Condition: `AgentRoot` is an empty string.
  Expected: Constructor returns error `"agent_root cannot be empty"`.

- Condition: `AgentRoot` is an absolute path (e.g., `"/usr/local"`, `"/home/user"`).
  Expected: Constructor returns error `"agent_root must be a relative path"`.

- Condition: `AgentRoot` contains a drive letter (e.g., `"C:\\projects"`).
  Expected: Constructor returns error `"agent_root must be a relative path"`.

- Condition: `AgentRoot` is `"."` (current directory).
  Expected: Constructor accepts this as a valid relative path.

- Condition: `AllowedTools` is nil.
  Expected: Constructor normalizes to empty slice `[]string{}`. No error.

- Condition: `DisallowedTools` is nil.
  Expected: Constructor normalizes to empty slice `[]string{}`. No error.

- Condition: `AllowedTools` and `DisallowedTools` both contain the same tool string.
  Expected: Constructor accepts this without error. Conflict resolution is Claude CLI's responsibility.

- Condition: `Model` contains an unrecognized value (e.g., `"invalid-model"`).
  Expected: Constructor accepts this. Semantic validation is Claude CLI's responsibility.

## Related

- [Node](./node.md) — Nodes reference agent roles via AgentRole field (format only, no existence check at components layer)
- [WorkflowDefinition](./workflow_definition.md) — Workflows contain nodes that reference agent roles
