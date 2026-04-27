# AgentDefinition

## Overview

An AgentDefinition describes the metadata for an AI agent role, including its unique identifier (`role`), Claude model configuration (`model`, `effort`), system prompt (`system_prompt`), tool permissions (`allowed_tools`, `disallowed_tools`), and execution context (`agent_root`). Each agent is uniquely identified by its role and stored as a YAML file in `.spectra/agents/`. Agents are loaded from the user's `.spectra/agents/` directory. Built-in agents are copied to `.spectra/agents/` during `spectra init` if they do not already exist.

## Behavior

1. An AgentDefinition is loaded from `.spectra/agents/<role>.yaml` when a workflow references the agent via `Node.AgentRole`.
2. The runtime validates that `role` is a unique PascalCase identifier with no spaces or special characters.
3. During agent definition load (before any workflow execution), the runtime checks that the `agent_root` directory exists. If it does not exist, the runtime returns an error.
4. When an agent is invoked during workflow execution, the runtime changes the working directory to `<agent_root>` and passes the following to the Claude CLI:
   - `--model <model>` (value from `model` field)
   - `--effort <effort>` (value from `effort` field)
   - `--system-prompt <system_prompt>` (value from `system_prompt` field)
   - `allowed_tools` and `disallowed_tools` as command-line arguments
5. If two AgentDefinitions have the same `role`, the runtime returns an error during load.
6. `model`, `effort`, `system_prompt`, `allowed_tools`, and `disallowed_tools` are passed directly to Claude CLI without validation by the Spectra runtime. The Claude CLI is responsible for validating and interpreting these values.

## Inputs

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Role | string | Non-empty, PascalCase (Go naming conventions), unique across all agents, no spaces or special characters | Yes |
| Model | string | Non-empty, user-defined, passed to Claude CLI `--model` flag as-is | Yes |
| Effort | string | Non-empty, user-defined, passed to Claude CLI `--effort` flag as-is | Yes |
| SystemPrompt | string | Non-empty, multi-line string, passed to Claude CLI `--system-prompt` flag as-is | Yes |
| AgentRoot | string | Non-empty, relative path from the spectra root (e.g., `"."`, `"spec"`), must be a valid directory | Yes |
| AllowedTools | []string | Array of strings, may be empty, passed to Claude CLI as-is | Yes (default: `[]`) |
| DisallowedTools | []string | Array of strings, may be empty, passed to Claude CLI as-is | Yes (default: `[]`) |

## Outputs

### AgentDefinition Structure

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| Role | string | Unique, PascalCase | Agent role identifier |
| Model | string | Non-empty, user-defined | Claude model identifier (e.g., "sonnet", "opus") |
| Effort | string | Non-empty, user-defined | Claude effort level (e.g., "low", "medium", "high") |
| SystemPrompt | string | Non-empty, multi-line string | System prompt for the agent, passed to Claude CLI |
| AgentRoot | string | Relative path | Working directory for agent execution |
| AllowedTools | []string | List of tool identifiers | Tools explicitly allowed for this agent |
| DisallowedTools | []string | List of tool identifiers | Tools explicitly disallowed for this agent |

### File Format

**File path**: `.spectra/agents/<Role>.yaml`

**Example**:

`.spectra/agents/QaReviewer.yaml`:
```yaml
role: "QaReviewer"
model: "sonnet"
effort: "high"
system_prompt: |-
  You are an attentive QA reviewer for software specifications...
agent_root: "."
allowed_tools:
  - "Read(*)"
disallowed_tools:
  - "Bash(spectra *)"
  - "Edit(*)"
```

## Invariants

1. **Role Uniqueness**: No two agents in `.spectra/agents/` may have the same `Role`. The filesystem enforces this constraint, but the runtime must also validate role uniqueness when loading multiple agents.

2. **Role Format**: `Role` must follow Go PascalCase naming conventions (no spaces, underscores, or special characters).

3. **Model Passthrough**: `Model` is passed to Claude CLI `--model` flag without validation by Spectra. The Claude CLI is responsible for validating model identifiers.

4. **Effort Passthrough**: `Effort` is passed to Claude CLI `--effort` flag without validation by Spectra. The Claude CLI is responsible for validating effort values.

5. **SystemPrompt Passthrough**: `SystemPrompt` is passed to Claude CLI `--system-prompt` flag without validation by Spectra. The runtime does not parse or interpret the prompt content, including any formatting, special characters, or YAML front matter.

6. **AgentRoot Validity**: `AgentRoot` must be a valid relative path from the spectra root and must reference an existing directory.

7. **Tool List Passthrough**: `AllowedTools` and `DisallowedTools` are passed to Claude CLI without modification or validation by Spectra. The Claude CLI is responsible for interpreting and enforcing these constraints.

8. **No AgentRoot Absolute Paths**: `AgentRoot` must not be an absolute path (must not start with `/` or contain drive letters like `C:`).

9. **Built-in Agent Copy Behavior**: During `spectra init`, built-in agents are copied to `.spectra/agents/` only if a file with the same name does not already exist. Existing files are skipped without error.

## Edge Cases

- **Condition**: AgentDefinition file `.spectra/agents/<role>.yaml` does not exist.
  **Expected**: Runtime returns "agent not found" error when a workflow references the role.

- **Condition**: `Role` contains spaces or special characters (e.g., `"QA-Analyst"`).
  **Expected**: Runtime rejects the agent definition with an error: "role must be PascalCase with no spaces or special characters".

- **Condition**: Two agents have the same `Role`.
  **Expected**: The filesystem enforces uniqueness (same filename). If the runtime attempts to load multiple agents with the same role from different sources, it returns an error: "agent '<role>' already exists".

- **Condition**: `Model` is an empty string.
  **Expected**: Runtime rejects the agent definition with an error: "model must be non-empty".

- **Condition**: `Model` contains an invalid Claude model identifier (e.g., `"invalid-model"`).
  **Expected**: The Spectra runtime passes the value to Claude CLI without validation. The Claude CLI returns an error during agent invocation if the model is invalid.

- **Condition**: `Effort` is an empty string.
  **Expected**: Runtime rejects the agent definition with an error: "effort must be non-empty".

- **Condition**: `Effort` contains an invalid value.
  **Expected**: The Spectra runtime passes the value to Claude CLI without validation. The Claude CLI returns an error during agent invocation if the effort value is invalid.

- **Condition**: `SystemPrompt` is an empty string.
  **Expected**: Runtime rejects the agent definition with an error: "system_prompt must be non-empty".

- **Condition**: `SystemPrompt` contains YAML front matter (content delimited by `---`).
  **Expected**: The runtime does not validate prompt content. The front matter is passed to Claude CLI as-is, which may interpret it as part of the prompt or return an error.

- **Condition**: `AgentRoot` is an absolute path (e.g., `"/usr/local"`).
  **Expected**: Runtime rejects the agent definition during load with an error: "agent_root must be a relative path".

- **Condition**: `AgentRoot` references a non-existent directory.
  **Expected**: Runtime returns an error during agent definition load (before any workflow execution): "agent_root directory not found: <path>".

- **Condition**: `AllowedTools` and `DisallowedTools` both contain the same tool.
  **Expected**: The Spectra runtime does not validate or resolve this conflict. It passes both arrays to Claude CLI as-is. The Claude CLI is responsible for handling the conflict.

- **Condition**: `AllowedTools` and `DisallowedTools` are both empty.
  **Expected**: The runtime allows this configuration. The Claude CLI determines the default behavior (likely allowing or denying all tools based on its own rules).

- **Condition**: `AllowedTools` or `DisallowedTools` contain invalid tool identifiers.
  **Expected**: The Spectra runtime does not validate tool names. Invalid tools are passed to Claude CLI, which may return an error during agent execution.

- **Condition**: AgentDefinition YAML is malformed or missing required fields.
  **Expected**: Runtime rejects the agent definition with a parse error indicating the specific issue.

- **Condition**: During `spectra init`, a built-in agent file already exists in `.spectra/agents/`.
  **Expected**: Runtime skips copying that file and proceeds without error. Existing user-defined agents are preserved.

## Related

- [Node](./node.md) - Nodes reference agent roles
- [WorkflowDefinition](./workflow_definition.md) - Workflows use agents in their nodes
- [Session](../entities/session/session.md) - Runtime dispatches agents during session execution
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
