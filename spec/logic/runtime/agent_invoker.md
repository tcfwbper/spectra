# AgentInvoker

## Overview

AgentInvoker is responsible for invoking a Claude CLI agent process with the appropriate configuration, environment variables, and working directory. It manages the Claude session ID lifecycle for each node (reading from or generating into session data), constructs the `claude` command with flags derived from the AgentDefinition, and starts the process asynchronously. AgentInvoker does not monitor the process after startup, does not capture process output, and does not handle RuntimeError construction or session failure — those are the caller's responsibilities.

## Boundaries

- Owns: Claude session ID lifecycle per node (read existing or generate new, persist to session data).
- Owns: CLI command construction (flags, arguments, working directory, environment variables).
- Owns: process startup via `exec.Command` + `cmd.Start()`.
- Owns: working directory validation (existence and directory-type check).
- Delegates: RuntimeError construction and `PersistentSession.Fail()` invocation to the runtime error-handling owner.
- Delegates: process lifecycle monitoring (exit detection, crash handling) to the event-driven runtime model (agents emit events via `spectra-agent event emit`).
- Delegates: semantic validation of AgentDefinition fields (model validity, tool existence) to Claude CLI.
- Delegates: UUID generation to a UUID generator dependency.
- Must not: block waiting for the Claude CLI process to exit.
- Must not: capture or redirect stdout/stderr of the Claude CLI process.
- Must not: construct RuntimeError or call `PersistentSession.Fail()`.
- Must not: perform shell command string construction — must use `exec.Command` with separate arguments.
- Must not: manually escape or quote arguments.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `PersistentSession` | State container with auto-persist | `GetSessionDataSafe()`, `UpdateSessionDataSafe()`, read `ID` | Must not call `Fail()`, `Run()`, `Done()`, or any lifecycle method |
| `AgentDefinition` | Configuration source | Read via getter methods (Model, Effort, SystemPrompt, AgentRoot, AllowedTools, DisallowedTools) | Must not modify or construct AgentDefinition |
| UUID generator | ID generation | Generate UUID v4 string | — |
| OS/exec | Process management | `exec.Command`, `cmd.Start()`, `cmd.Dir`, `cmd.Env` | Must not use `cmd.Run()` or `cmd.Output()` |
| Filesystem | Directory validation | Check existence and type of working directory | Must not create, delete, or modify any filesystem entry |

Construction constraint: AgentInvoker is initialized with a `PersistentSession` reference and `ProjectRoot` (absolute path). These are provided by the runtime layer which has already resolved the project root via SpectraFinder. AgentInvoker does not invoke SpectraFinder itself.

## Behavior

1. Receives `NodeName`, `Message`, and `AgentDefinition` as invocation parameters.
2. Calls `PersistentSession.GetSessionDataSafe("<NodeName>.ClaudeSessionID")` to retrieve the stored Claude session ID for this node.
3. If the key does not exist (returns `(nil, false)`), generates a new UUID v4 as the Claude session ID and marks this as a new session (will use `--session-id` flag).
4. If the key exists (returns `(value, true)`), validates that the value is a string type. If not a string, returns an error. If valid, uses this value as the Claude session ID and marks this as an existing session (will use `--resume` flag).
5. If UUID generation fails, returns an error immediately. No session data is written, no command is constructed.
6. If a new Claude session ID was generated, calls `PersistentSession.UpdateSessionDataSafe("<NodeName>.ClaudeSessionID", generatedClaudeSessionID)` to persist it. PersistentSession automatically persists the updated session data (non-fatal if persistence fails).
7. If `PersistentSession.UpdateSessionDataSafe()` returns an error (in-memory validation failure), returns that error immediately. No command is constructed.
8. Constructs the absolute working directory path by joining `ProjectRoot` with `AgentDefinition.AgentRoot()`.
9. Validates that the working directory exists and is a directory. If not, returns an error immediately.
10. Constructs the `claude` CLI command using `exec.Command` with separate arguments:
    - Base command: `claude`
    - `--permission-mode bypassPermissions`
    - `--model <AgentDefinition.Model()>`
    - `--effort <AgentDefinition.Effort()>`
    - `--system-prompt <AgentDefinition.SystemPrompt()>`
    - (Conditional) `--allowed-tools <tool1> <tool2> ...` if `AgentDefinition.AllowedTools()` is non-empty
    - (Conditional) `--disallowed-tools <tool1> <tool2> ...` if `AgentDefinition.DisallowedTools()` is non-empty
    - (Conditional) `--resume <ClaudeSessionID>` if existing session
    - (Conditional) `--session-id <ClaudeSessionID>` if new session
    - `--print <Message>` (Message as a separate argument)
11. Sets `cmd.Dir` to the absolute working directory path.
12. Sets `cmd.Env` by appending `SPECTRA_SESSION_ID=<PersistentSession.ID>` and `SPECTRA_CLAUDE_SESSION_ID=<ClaudeSessionID>` to the parent process's environment (`os.Environ()`).
13. Calls `cmd.Start()` to start the process asynchronously.
14. If `cmd.Start()` fails, returns an error immediately.
15. Returns `nil` on success. The Claude CLI process runs independently in the background.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| PersistentSession | PersistentSession reference | Valid, constructed via NewPersistentSession | Yes |
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### For Invocation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| NodeName | string | Non-empty, PascalCase (matches a workflow node name) | Yes |
| Message | string | Plain text prompt; may contain any characters including quotes, newlines, special characters | Yes |
| AgentDefinition | AgentDefinition reference | Valid, already-constructed AgentDefinition | Yes |

## Outputs

### Success

| Field | Type | Description |
|-------|------|-------------|
| Error | error | `nil` |

**Side effects**:
- Claude CLI process started and running in the background (stdout/stderr inherited from parent process).
- If a new Claude session ID was generated, `SessionData["<NodeName>.ClaudeSessionID"]` is updated and persisted via `PersistentSession.UpdateSessionDataSafe()`.

### Error Cases

| Condition | Error Message |
|-----------|--------------|
| Claude session ID exists but is not a string type | `"invalid Claude session ID type for node '<NodeName>': expected string"` |
| UUID generation fails | `"failed to generate Claude session ID"` |
| `PersistentSession.UpdateSessionDataSafe()` fails | `"failed to update session with new Claude session ID: <underlying error>"` |
| Working directory does not exist or is not a directory | `"agent working directory not found or invalid: <path>"` |
| `cmd.Start()` fails | `"failed to start Claude CLI process: <underlying error>"` |

## Invariants

1. **ClaudeSessionID Non-Empty After Resolution**: After step 4 completes without error, ClaudeSessionID is guaranteed to be a non-empty string (either retrieved or newly generated).

2. **Session Data Write-If-And-Only-If New**: `PersistentSession.UpdateSessionDataSafe()` is called if and only if the Claude session ID was newly generated (key did not exist in session data).

3. **Command Construction Safety**: Must use `exec.Command` with each flag and value as separate arguments. Must never construct a shell command string or pass arguments to a shell interpreter.

4. **No Manual Escaping**: Must not manually quote or escape arguments. `exec.Command` handles argument passing safely.

5. **Working Directory Is Absolute**: `cmd.Dir` is always set to an absolute path (ProjectRoot joined with AgentDefinition.AgentRoot()).

6. **Environment Variable Injection**: The started process's environment always contains both `SPECTRA_SESSION_ID` and `SPECTRA_CLAUDE_SESSION_ID`. Appended values take precedence over any pre-existing values in the parent environment (last-value-wins behavior).

7. **Resume vs Session-ID Mutual Exclusivity**: The command includes exactly one of `--resume <ID>` (existing session) or `--session-id <ID>` (new session), never both.

8. **Conditional Tool Flags**: `--allowed-tools` is omitted entirely when `AllowedTools()` returns an empty slice. Same for `--disallowed-tools` and `DisallowedTools()`.

9. **Fail-Fast**: Returns immediately on first error. No subsequent steps execute after an error.

10. **No Output Capture**: Must not redirect or capture stdout/stderr. The Claude CLI process inherits the parent process's output streams.

11. **Asynchronous Non-Blocking**: Returns immediately after `cmd.Start()` succeeds. Does not call `cmd.Wait()` or monitor the process.

12. **Error Delegation**: Returns plain Go errors. Must not construct RuntimeError or call `PersistentSession.Fail()`.

13. **Thread Safety**: Safe for concurrent invocations targeting different nodes within the same session. Each invocation operates on a distinct `<NodeName>.ClaudeSessionID` key. `PersistentSession.UpdateSessionDataSafe()` serializes concurrent writes via session's internal lock. Concurrent invocations for the same node are not supported and must be serialized externally by the caller.

## Edge Cases

- Condition: `PersistentSession.GetSessionDataSafe` returns `(nil, false)` (key does not exist).
  Expected: Generates new UUID v4, persists via `UpdateSessionDataSafe`, uses `--session-id` flag.

- Condition: `PersistentSession.GetSessionDataSafe` returns `(value, true)` where value is a valid string.
  Expected: Uses stored value as Claude session ID, does not call `UpdateSessionDataSafe`, uses `--resume` flag.

- Condition: `PersistentSession.GetSessionDataSafe` returns `(value, true)` where value is not a string.
  Expected: Returns error `"invalid Claude session ID type for node '<NodeName>': expected string"`.

- Condition: `AgentDefinition.AgentRoot()` returns `"."`.
  Expected: Working directory is set to ProjectRoot itself.

- Condition: `AgentDefinition.AgentRoot()` returns a multi-level path (e.g., `"spec/logic"`).
  Expected: Working directory is ProjectRoot joined with `"spec/logic"`.

- Condition: `AgentDefinition.AllowedTools()` returns `["Bash(*)", "Read(*)"]`.
  Expected: Command includes `--allowed-tools Bash(*) Read(*)` with each tool as a separate argument.

- Condition: `AgentDefinition.AllowedTools()` returns empty slice.
  Expected: `--allowed-tools` flag is entirely omitted.

- Condition: `AgentDefinition.DisallowedTools()` returns empty slice.
  Expected: `--disallowed-tools` flag is entirely omitted.

- Condition: `Message` contains double quotes, newlines, or shell metacharacters.
  Expected: `exec.Command` passes the message as-is without modification. Claude CLI receives exact content.

- Condition: `claude` command is not found in system PATH.
  Expected: `cmd.Start()` fails. Returns error `"failed to start Claude CLI process: ..."`.

- Condition: Working directory does not exist on filesystem.
  Expected: Returns error `"agent working directory not found or invalid: <path>"` before command construction.

- Condition: UUID generation fails (extremely rare — entropy exhaustion).
  Expected: Returns error `"failed to generate Claude session ID"`. No session data written, no command executed.

- Condition: `PersistentSession.UpdateSessionDataSafe()` returns error.
  Expected: Returns error. No command executed.

- Condition: Environment variables `SPECTRA_SESSION_ID` or `SPECTRA_CLAUDE_SESSION_ID` already exist in parent environment.
  Expected: Appended values take precedence (last-value-wins). No deduplication is performed.

- Condition: Multiple agents invoked concurrently for different nodes in same session.
  Expected: Each invocation is independent. Concurrent `UpdateSessionDataSafe` calls target distinct keys and are serialized by session's lock.

- Condition: Stored Claude session ID is not valid UUID format.
  Expected: AgentInvoker does not validate UUID format of stored values. Passes as-is. Claude CLI is responsible for validation.

- Condition: Parent process terminates while Claude CLI processes are running.
  Expected: Claude CLI processes are not automatically terminated. They may become orphaned. This is outside AgentInvoker's responsibility.

## Related

- [AgentDefinition](../components/agent_definition.md) — provides agent configuration (model, effort, system prompt, tools, agent root)
- [Session](../entities/session/session.md) — stores Claude session IDs in SessionData; provides thread-safe read/write methods
- [RuntimeError](../entities/runtime_error.md) — constructed by the runtime error-handling owner when AgentInvoker's caller propagates an error
- [TransitionToNode](./transition_to_node.md) — direct caller; invokes AgentInvoker and propagates errors upward
