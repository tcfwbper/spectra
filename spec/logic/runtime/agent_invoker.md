# AgentInvoker

## Overview

AgentInvoker is responsible for invoking a Claude CLI agent with the appropriate configuration, environment variables, and working directory. It constructs the `claude` command with flags derived from the AgentDefinition, manages Claude session ID lifecycle (generating new IDs when needed), and ensures the agent process starts successfully before returning. AgentInvoker executes asynchronously — it starts the Claude CLI process and returns immediately without waiting for the agent to complete its work. If any step fails (UUID generation, callback execution, or process start), AgentInvoker triggers a RuntimeError and attempts to clean up any started processes.

## Behavior

1. AgentInvoker receives the following inputs: `NodeName` (non-empty), `Message` (plain text user prompt), and `AgentDefinition`.
2. AgentInvoker calls `Session.GetSessionDataSafe("<NodeName>.ClaudeSessionID")` to retrieve the stored Claude session ID for this node.
3. If the key does not exist (returns `(nil, false)`), AgentInvoker generates a new UUID v4 as the Claude session ID and marks this as a new session (will use `--session-id` flag).
4. If the key exists (returns `(value, true)`), AgentInvoker validates that the value is a string type. If not, returns an error: `"invalid Claude session ID type for node '<NodeName>': expected string"`. If valid, uses this value as the Claude session ID and marks this as an existing session (will use `--resume` flag).
5. If UUID generation fails (when creating a new session), AgentInvoker returns an error: `"failed to generate Claude session ID"`. The caller is responsible for constructing a RuntimeError and calling `Session.Fail()` if appropriate.
6. If a new Claude session ID was generated, AgentInvoker calls `Session.UpdateSessionDataSafe("<NodeName>.ClaudeSessionID", generatedClaudeSessionID)` to store and persist it.
7. If `Session.UpdateSessionDataSafe()` returns an error, AgentInvoker returns an error: `"failed to update session with new Claude session ID: <error>"`. The caller is responsible for constructing a RuntimeError and calling `Session.Fail()` if appropriate.
8. At this point, `ClaudeSessionID` is guaranteed to be a non-empty string (either retrieved from session data or newly generated).
9. AgentInvoker uses the `ProjectRoot` provided at initialization (no SpectraFinder call).
10. AgentInvoker constructs the absolute working directory path by joining `ProjectRoot` with `AgentDefinition.AgentRoot`.
11. If the working directory does not exist or is not a directory, AgentInvoker returns an error: `"agent working directory not found or invalid: <path>"`. The caller is responsible for constructing a RuntimeError and calling `Session.Fail()` if appropriate.
12. AgentInvoker constructs the `claude` CLI command with the following structure:
    - Base command: `claude`
    - Permission flag: `--permission-mode bypassPermission`
    - Model flag: `--model <AgentDefinition.Model>`
    - Effort flag: `--effort <AgentDefinition.Effort>`
    - System prompt flag: `--system-prompt <AgentDefinition.SystemPrompt>`
    - (Conditional) Allowed tools: `--allowed-tools "<tool1>" "<tool2>" ...` if `AgentDefinition.AllowedTools` is non-empty
    - (Conditional) Disallowed tools: `--disallowed-tools "<tool1>" "<tool2>" ...` if `AgentDefinition.DisallowedTools` is non-empty
    - (Conditional) Resume flag: `--resume <ClaudeSessionID>` if the Claude session ID was retrieved from session data (existing session)
    - (Conditional) Session ID flag: `--session-id <ClaudeSessionID>` if the Claude session ID was newly generated (new session)
    - Print flag: `--print <Message>` where `<Message>` is passed as a separate argument (not shell-escaped, handled by Go's exec.Command)
13. AgentInvoker constructs the command using Go's `exec.Command` to avoid shell injection and automatic argument escaping. Each flag and value is passed as a separate argument to `exec.Command`.
14. AgentInvoker sets the command's working directory to the absolute path constructed in step 10 using `cmd.Dir`.
15. AgentInvoker injects two environment variables into the command's environment:
    - `SPECTRA_SESSION_ID=<Session.ID>`
    - `SPECTRA_CLAUDE_SESSION_ID=<ClaudeSessionID>`
    Environment variables are injected by appending to the parent process's environment using `cmd.Env = append(os.Environ(), "SPECTRA_SESSION_ID=...", "SPECTRA_CLAUDE_SESSION_ID=...")`. If these variables already exist in the parent environment, the appended values take precedence (last value in array wins).
16. AgentInvoker starts the Claude CLI process using `cmd.Start()` (non-blocking).
17. If `cmd.Start()` fails (e.g., `claude` command not found, permission denied, working directory inaccessible), AgentInvoker returns an error: `"failed to start Claude CLI process: <details>"`. No process cleanup is needed because the process never started. The caller is responsible for constructing a RuntimeError and calling `Session.Fail()` if appropriate.
18. If `cmd.Start()` succeeds, AgentInvoker stores a reference to the started process (`cmd.Process`) for potential cleanup.
19. AgentInvoker immediately returns `nil` (no error) after successfully starting the process. The Claude CLI process runs asynchronously in the background.
20. AgentInvoker does **not** capture or redirect the Claude CLI's stdout or stderr. Output streams are inherited from the parent process.
21. If an error occurs after `cmd.Start()` succeeds (e.g., a post-start validation step), AgentInvoker attempts to stop the process before returning the error. The termination mechanism is platform-specific:
    - **Unix/Linux/macOS**: Send SIGTERM to the process using `cmd.Process.Signal(syscall.SIGTERM)`. Wait for up to 5 seconds for the process to exit. If the process does not exit within 5 seconds, send SIGKILL using `cmd.Process.Kill()`.
    - **Windows**: Send SIGKILL immediately using `cmd.Process.Kill()` (Windows does not support SIGTERM).
    - Return an error: `"post-start validation failed: <details>"`. The caller is responsible for constructing a RuntimeError and calling `Session.Fail()` if appropriate.
22. AgentInvoker does not monitor the Claude CLI process lifecycle after startup. Process exit status and long-term lifecycle management are not part of AgentInvoker's responsibility.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| Session | *Session | Reference to the Session entity shared across all runtime components | Yes |
| ProjectRoot | string | Absolute path to the directory containing `.spectra` (already located by Runtime via SpectraFinder; passed in to avoid repeated upward filesystem traversal) | Yes |

**Note**: AgentInvoker does not invoke SpectraFinder itself. The project root is provided at initialization time by Runtime, which has already performed the filesystem traversal once during session initialization.

### For InvokeAgent Operation

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| NodeName | string | Non-empty, PascalCase (matches a workflow node name) | Yes |
| Message | string | Plain text user prompt, may contain any characters including quotes, newlines, etc. | Yes |
| AgentDefinition | AgentDefinition struct | Valid agent definition with all required fields (Role, Model, Effort, SystemPrompt, AgentRoot, AllowedTools, DisallowedTools) | Yes |

## Outputs

### Success Case

**Return value**: `nil` (no error)

**Side effects**:
- Claude CLI process started and running in the background
- If a new Claude session ID was generated, `SessionData[<NodeName>.ClaudeSessionID]` is updated and persisted via `Session.UpdateSessionDataSafe()`

### Error Cases

| Error Condition | Error Behavior |
|----------------|----------------|
| Claude session ID exists but is not a string type | Returns error: `"invalid Claude session ID type for node '<NodeName>': expected string"`. Caller handles error. |
| Working directory does not exist or is not a directory | Returns error: `"agent working directory not found or invalid: <path>"`. Caller handles error. |
| UUID generation fails | Returns error: `"failed to generate Claude session ID"`. Caller handles error. |
| `Session.UpdateSessionDataSafe()` returns error | Returns error: `"failed to update session with new Claude session ID: <error>"`. Caller handles error. |
| `cmd.Start()` fails | Returns error: `"failed to start Claude CLI process: <details>"`. Caller handles error. |
| Post-start error | Attempts platform-specific process termination (Unix: SIGTERM → SIGKILL; Windows: SIGKILL), returns error: `"post-start validation failed: <details>"`. Caller handles error. |

## Invariants

1. **ClaudeSessionID Non-Empty After Processing**: After step 8, `ClaudeSessionID` must be a non-empty string (either retrieved from session data or newly generated).

2. **Session Data Update Condition**: `Session.UpdateSessionDataSafe()` is called **if and only if** the Claude session ID was not found in session data (indicating a new Claude session needs to be created).

3. **Command Construction Safety**: AgentInvoker must use `exec.Command` with separate arguments for each flag and value. It must **never** construct a shell command string and pass it to a shell interpreter (e.g., `sh -c`).

4. **Argument Escaping**: AgentInvoker must **not** manually escape or quote arguments. Go's `exec.Command` handles argument passing safely when each argument is provided separately.

5. **Working Directory Absolute Path**: The working directory set via `cmd.Dir` must be an absolute path constructed by joining the spectra root (absolute) with `AgentDefinition.AgentRoot` (relative).

6. **Environment Variable Injection**: The environment must include both `SPECTRA_SESSION_ID` and `SPECTRA_CLAUDE_SESSION_ID` when the Claude CLI process starts.

7. **Resume vs. Session-ID Flag Mutual Exclusivity**:
   - If the Claude session ID was retrieved from session data (existing session), the command includes `--resume <ClaudeSessionID>` and **does not** include `--session-id`.
   - If the Claude session ID was newly generated (new session), the command includes `--session-id <ClaudeSessionID>` and **does not** include `--resume`.

8. **AllowedTools and DisallowedTools Conditional Inclusion**:
   - If `AgentDefinition.AllowedTools` is an empty array `[]`, the `--allowed-tools` flag is **omitted entirely** from the command.
   - If `AgentDefinition.DisallowedTools` is an empty array `[]`, the `--disallowed-tools` flag is **omitted entirely** from the command.
   - If non-empty, each tool string is passed as a separate quoted argument.

9. **Fail-Fast Error Handling**: AgentInvoker must return immediately upon the first error encountered. No subsequent steps are executed after an error. The caller is responsible for deciding whether to construct a RuntimeError and call `Session.Fail()`.

10. **Process Cleanup Obligation**: If `cmd.Start()` succeeds but a subsequent error occurs (before returning from AgentInvoker), the started process must be terminated using the platform-specific termination sequence (Unix: SIGTERM → wait → SIGKILL; Windows: SIGKILL).

11. **No Output Capture**: AgentInvoker must **not** redirect or capture stdout or stderr. The Claude CLI process inherits the parent process's output streams.

12. **Asynchronous Execution**: AgentInvoker returns immediately after successfully starting the Claude CLI process. It does **not** wait for the process to exit or monitor its status.

13. **Error Handling Delegation**: AgentInvoker returns errors to the caller without constructing RuntimeError or calling `Session.Fail()`. The caller (typically TransitionToNode or Runtime) is responsible for error handling and session state transitions.

14. **Thread Safety**: AgentInvoker is **thread-safe** for concurrent invocations targeting different nodes. Concurrent invocations modifying the same session but for different nodes (e.g., `NodeA` and `NodeB` updating `SessionData[NodeA.ClaudeSessionID]` and `SessionData[NodeB.ClaudeSessionID]` respectively) are safe because:
   - Each invocation updates a distinct `<NodeName>.ClaudeSessionID` key in SessionData.
   - `Session.UpdateSessionDataSafe()` uses the session-level write lock to serialize writes.
   - Session's "last write wins" behavior ensures consistency across concurrent writes.
   Callers invoking AgentInvoker concurrently for the **same node** within the same session must serialize invocations externally, as multiple Claude sessions for a single node are not supported.

## Edge Cases

- **Condition**: Claude session ID does not exist in session data (`Session.GetSessionDataSafe` returns `(nil, false)`).
  **Expected**: AgentInvoker generates a new UUID v4, calls `Session.UpdateSessionDataSafe("<NodeName>.ClaudeSessionID", newUUID)` to store and persist it, and uses `--session-id <newUUID>` in the command.

- **Condition**: Claude session ID exists in session data (`Session.GetSessionDataSafe` returns `(value, true)` and value is a string).
  **Expected**: AgentInvoker uses the stored Claude session ID as-is, does **not** call `Session.UpdateSessionDataSafe()`, and uses `--resume <ClaudeSessionID>` in the command.

- **Condition**: Claude session ID exists in session data but is not a string type.
  **Expected**: AgentInvoker returns an error: `"invalid Claude session ID type for node '<NodeName>': expected string"`. The caller handles the error.

- **Condition**: `AgentDefinition.AgentRoot` is `"."`.
  **Expected**: The working directory is set to the spectra root directory itself.

- **Condition**: `AgentDefinition.AgentRoot` is a multi-level relative path (e.g., `"spec/logic"`).
  **Expected**: AgentInvoker joins the spectra root with `"spec/logic"` to form the absolute working directory path.

- **Condition**: `AgentDefinition.AllowedTools` is `["Bash(*)", "Read(*)"]`.
  **Expected**: The command includes `--allowed-tools "Bash(*)" "Read(*)"` with each tool as a separate argument.

- **Condition**: `AgentDefinition.AllowedTools` is an empty array `[]`.
  **Expected**: The `--allowed-tools` flag is completely omitted from the command.

- **Condition**: `AgentDefinition.DisallowedTools` is an empty array `[]`.
  **Expected**: The `--disallowed-tools` flag is completely omitted from the command.

- **Condition**: `Message` contains double quotes, e.g., `He said "hello"`.
  **Expected**: Go's `exec.Command` handles the quotes safely. The message is passed as-is without manual escaping. The Claude CLI receives the exact message content.

- **Condition**: `Message` contains newlines and special characters, e.g., `Line1\nLine2\t$VAR`.
  **Expected**: Go's `exec.Command` preserves the message exactly. The Claude CLI receives the multi-line message with all special characters intact.

- **Condition**: UUID generation fails (e.g., due to lack of entropy, though extremely rare).
  **Expected**: AgentInvoker returns an error: `"failed to generate Claude session ID"`. No `Session.UpdateSessionDataSafe()` call is made, no command is executed. The caller handles the error.

- **Condition**: `Session.UpdateSessionDataSafe()` returns an error (e.g., validation failure, though unlikely for string values).
  **Expected**: AgentInvoker returns an error: `"failed to update session with new Claude session ID: <error>"`. No command is executed. The caller handles the error.

- **Condition**: `claude` command is not found in the system PATH.
  **Expected**: `cmd.Start()` fails with "executable file not found in $PATH". AgentInvoker returns an error: `"failed to start Claude CLI process: executable file not found in $PATH"`. The caller handles the error.

- **Condition**: Working directory (constructed from `AgentDefinition.AgentRoot`) does not exist.
  **Expected**: AgentInvoker detects this during step 11 and returns an error: `"agent working directory not found or invalid: <path>"`. No command is executed. The caller handles the error.

- **Condition**: Working directory exists but the current user lacks execute permissions.
  **Expected**: `cmd.Start()` fails with a permission error. AgentInvoker returns an error: `"failed to start Claude CLI process: permission denied"`. The caller handles the error.

- **Condition**: `cmd.Start()` succeeds, but a post-start validation step fails.
  **Expected**: AgentInvoker attempts platform-specific process termination. On Unix: sends SIGTERM, waits up to 5 seconds, then sends SIGKILL if needed. On Windows: sends SIGKILL immediately. Returns an error: `"post-start validation failed: <details>"`. The caller handles the error.

- **Condition**: Environment variables `SPECTRA_SESSION_ID` or `SPECTRA_CLAUDE_SESSION_ID` already exist in the parent process's environment.
  **Expected**: AgentInvoker appends the new values to the environment array. The appended values take precedence (last value in array wins). The Claude CLI process receives the AgentInvoker-provided values, not the parent's original values. AgentInvoker does **not** check for or deduplicate existing environment variables.

- **Condition**: Multiple agents are invoked concurrently for different nodes in the same session (e.g., `NodeA` and `NodeB` in parallel).
  **Expected**: Each AgentInvoker invocation is independent and thread-safe. Multiple Claude CLI processes run concurrently. Concurrent `Session.UpdateSessionDataSafe()` calls update distinct `<NodeName>.ClaudeSessionID` keys in SessionData. Session's internal write lock serializes writes, ensuring consistency.

- **Condition**: AgentInvoker is invoked concurrently for the same node in the same session.
  **Expected**: This violates the design assumption that each node has at most one Claude session per session. Callers must serialize such invocations externally. Concurrent invocations for the same node may result in race conditions or multiple Claude sessions being created for a single node, leading to undefined behavior.

- **Condition**: `AgentDefinition.Model` contains spaces or special characters (e.g., `"sonnet 4.0"`).
  **Expected**: Go's `exec.Command` passes the value safely as a separate argument. The Claude CLI is responsible for validating and interpreting the model identifier.

- **Condition**: `AgentDefinition.SystemPrompt` is a multi-line string with embedded quotes and special characters.
  **Expected**: Go's `exec.Command` passes the entire prompt as a single argument without modification. The Claude CLI receives the exact prompt content.

- **Condition**: The stored Claude session ID is not a valid UUID format.
  **Expected**: AgentInvoker does **not** validate UUID format. It passes the value as-is to the environment variables and command flags. The Claude CLI or downstream components are responsible for validation.

- **Condition**: AgentInvoker is invoked while a previous invocation's Claude process is still running for the same node (same Claude session ID).
  **Expected**: AgentInvoker proceeds with the new invocation. The `--resume` flag is used. The Claude CLI and its session management are responsible for handling concurrent or sequential invocations with the same session ID.

- **Condition**: `Session.UpdateSessionDataSafe()` is called with a `<NodeName>.ClaudeSessionID` key and a non-string value (programming error in AgentInvoker).
  **Expected**: `Session.UpdateSessionDataSafe()` validates the value type and returns an error. AgentInvoker returns the error to the caller. This should not occur in correct implementations.

- **Condition**: The parent process (spectra runtime) is terminated while Claude CLI processes are still running.
  **Expected**: Claude CLI processes are **not** automatically terminated (they are detached/independent processes). This may leave orphaned processes.

- **Condition**: `Session.UpdateSessionDataSafe()` panics (unlikely, but theoretically possible if Session is in an invalid state).
  **Expected**: The panic propagates up to the caller. AgentInvoker does **not** recover from panics. The caller (e.g., TransitionToNode or MessageRouter) is responsible for panic recovery and error handling.

- **Condition**: Process termination during cleanup fails (SIGTERM and SIGKILL both fail).
  **Expected**: AgentInvoker logs a warning or includes the termination failure details in the error message, but still returns the error. The process may remain as a zombie or orphaned process. The caller handles the error.

## Related

- [AgentDefinition](../components/agent_definition.md) - Defines agent configuration (model, effort, system prompt, tools, working directory)
- [Session](../entities/session/session.md) - Session stores Claude session IDs in SessionData with keys `<NodeName>.ClaudeSessionID` and provides thread-safe methods for data access
- [RuntimeError](../entities/runtime_error.md) - Caller constructs RuntimeError and calls `Session.Fail()` when AgentInvoker returns errors
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
