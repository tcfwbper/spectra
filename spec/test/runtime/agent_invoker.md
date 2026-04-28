# Test Specification: `agent_invoker.go`

## Source File Under Test
`runtime/agent_invoker.go`

## Test File
`runtime/agent_invoker_test.go`

---

## `AgentInvoker`

### Happy Path — New Session

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_NewSession_GeneratesSessionID` | `unit` | Generates new Claude session ID when not found in session data. | Mock session returns `(nil, false)` for `GetSessionDataSafe`; mock session accepts `UpdateSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; no pre-existing project files accessed | `NodeName="TestNode"`, `Message="test message"`, `AgentDefinition` with `AgentRoot="agents"` | Returns `nil`; `UpdateSessionDataSafe` called with `"TestNode.ClaudeSessionID"` and valid UUID v4; command uses `--session-id <UUID>` flag |
| `TestAgentInvoker_NewSession_StartsProcess` | `unit` | Starts Claude CLI process successfully for new session. | Mock session returns `(nil, false)` for `GetSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; mock Claude CLI executable in test PATH | `NodeName="TestNode"`, `Message="test message"`, `AgentDefinition` with `AgentRoot="agents"`, `Model="sonnet"`, `Effort="normal"` | Returns `nil`; process started successfully; command includes `--session-id` flag |

### Happy Path — Existing Session

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_ExistingSession_UsesStoredSessionID` | `unit` | Uses stored Claude session ID when found in session data. | Mock session returns `("existing-uuid-1234", true)` for `GetSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; mock Claude CLI executable in test PATH | `NodeName="TestNode"`, `Message="resume message"`, `AgentDefinition` with `AgentRoot="agents"` | Returns `nil`; `UpdateSessionDataSafe` not called; command uses `--resume existing-uuid-1234` flag |
| `TestAgentInvoker_ExistingSession_NoSessionDataUpdate` | `unit` | Does not update session data when session ID exists. | Mock session returns `("uuid-5678", true)` for `GetSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory | `NodeName="NodeA"`, `Message="message"`, `AgentDefinition` with `AgentRoot="agents"` | Returns `nil`; `UpdateSessionDataSafe` never invoked; process started with `--resume uuid-5678` |

### Happy Path — Command Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_CommandConstruction_AllFlags` | `unit` | Constructs command with all flags correctly. | Mock session with new session ID; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor to capture arguments | `NodeName="Node"`, `Message="prompt"`, `AgentDefinition` with `Model="sonnet"`, `Effort="high"`, `SystemPrompt="You are X"`, `AllowedTools=["Bash(*)", "Read(*)"]`, `DisallowedTools=["Write(*)"]`, `AgentRoot="agents"` | Command includes `--permission-mode bypassPermission`, `--model sonnet`, `--effort high`, `--system-prompt "You are X"`, `--allowed-tools "Bash(*)" "Read(*)"`, `--disallowed-tools "Write(*)"`, `--session-id <UUID>`, `--print prompt` |
| `TestAgentInvoker_CommandConstruction_EmptyToolArrays` | `unit` | Omits tool flags when arrays are empty. | Mock session with new session ID; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition` with `AllowedTools=[]`, `DisallowedTools=[]` | Command does not include `--allowed-tools` or `--disallowed-tools` flags |
| `TestAgentInvoker_CommandConstruction_WorkingDirectory` | `unit` | Sets working directory correctly. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/logic/` nested subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition` with `AgentRoot="agents/logic"` | Command working directory is `<ProjectRoot>/agents/logic` (absolute path) |
| `TestAgentInvoker_CommandConstruction_AgentRootDot` | `unit` | Handles AgentRoot "." as project root. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` subdirectory; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition` with `AgentRoot="."` | Command working directory is `<ProjectRoot>` (absolute path) |

### Happy Path — Environment Variables

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_EnvVars_Injected` | `unit` | Injects SPECTRA_SESSION_ID and SPECTRA_CLAUDE_SESSION_ID environment variables. | Mock session with `Session.ID="session-123"`; mock session returns new UUID; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `NodeName="Node"`, generated `ClaudeSessionID="claude-456"` | Command environment includes `SPECTRA_SESSION_ID=session-123` and `SPECTRA_CLAUDE_SESSION_ID=claude-456` |
| `TestAgentInvoker_EnvVars_OverrideParent` | `unit` | Injected environment variables override parent environment. | Parent environment set with `SPECTRA_SESSION_ID=old-value`; mock session with `Session.ID="new-session"`; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `NodeName="Node"` | Command environment has `SPECTRA_SESSION_ID=new-session` (parent value overridden) |

### Happy Path — Message Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_Message_WithQuotes` | `unit` | Handles messages with double quotes safely. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `Message=`He said "hello"`` | Command includes `--print` with exact message content; no manual escaping |
| `TestAgentInvoker_Message_WithNewlines` | `unit` | Preserves newlines and special characters in message. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `Message="Line1\nLine2\t$VAR"` | Command includes `--print` with exact multi-line content including `\n`, `\t`, `$VAR` |
| `TestAgentInvoker_Message_MultilineSystemPrompt` | `unit` | Handles multi-line system prompts with special characters. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition.SystemPrompt="Line 1\nLine 2\n\"quoted\""` | Command includes `--system-prompt` with exact multi-line content |

### Happy Path — Asynchronous Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_ReturnsImmediately` | `unit` | Returns immediately after starting process without waiting. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; mock Claude CLI that sleeps for 2 seconds | `NodeName="Node"`, `Message="test"` | Returns `nil` in under 500ms; process still running in background |
| `TestAgentInvoker_NoOutputCapture` | `unit` | Does not redirect stdout or stderr. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `NodeName="Node"` | Command stdout and stderr are not redirected (inherit from parent) |

### Validation Failures — Session Data

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_SessionIDInvalidType` | `unit` | Returns error when stored Claude session ID is not a string. | Mock session returns `(123, true)` (int, not string) for `GetSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory | `NodeName="BadNode"` | Returns error matching `/invalid Claude session ID type for node 'BadNode': expected string/i` |
| `TestAgentInvoker_UpdateSessionDataFails` | `unit` | Returns error when UpdateSessionDataSafe fails. | Mock session returns `(nil, false)` for `GetSessionDataSafe`; mock session returns error from `UpdateSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory | `NodeName="Node"` | Returns error matching `/failed to update session with new Claude session ID/i` |
| `TestAgentInvoker_UpdateSessionDataFails_ErrorWrapping` | `unit` | Returns error with wrapped original error from UpdateSessionDataSafe. | Mock session returns `(nil, false)` for `GetSessionDataSafe`; mock session returns error `"validation failed: invalid key"` from `UpdateSessionDataSafe`; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory | `NodeName="Node"` | Returns error matching `/failed to update session with new Claude session ID:.*validation failed: invalid key/i`; error wraps original error |

### Validation Failures — Working Directory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_WorkingDirNotExist` | `unit` | Returns error when working directory does not exist. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` but no `nonexistent/` subdirectory; ProjectRoot set to fixture directory | `AgentDefinition.AgentRoot="nonexistent"` | Returns error matching `/agent working directory not found or invalid:.*nonexistent/i` |
| `TestAgentInvoker_WorkingDirIsFile` | `unit` | Returns error when working directory is a file, not a directory. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and regular file `file.txt`; ProjectRoot set to fixture directory | `AgentDefinition.AgentRoot="file.txt"` | Returns error matching `/agent working directory not found or invalid:.*file\.txt/i` |

### Validation Failures — Process Start

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_ClaudeCommandNotFound` | `unit` | Returns error when claude command not in PATH. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; PATH set to exclude claude executable | `NodeName="Node"` | Returns error matching `/failed to start Claude CLI process:.*executable file not found/i` |
| `TestAgentInvoker_WorkingDirNoExecutePermission` | `unit` | Returns error when working directory lacks execute permissions. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; `agents/` permissions set to `0000` within test fixture | `AgentDefinition.AgentRoot="agents"` | Returns error matching `/failed to start Claude CLI process:.*permission denied/i` |

### Validation Failures — UUID Generation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_UUIDGenerationFails` | `unit` | Returns error when UUID generation fails. | Mock session returns `(nil, false)` for `GetSessionDataSafe`; mock UUID generator that returns error; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory | `NodeName="Node"` | Returns error matching `/failed to generate Claude session ID/i` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_ErrorsDontCallSessionFail` | `unit` | Errors are returned without calling Session.Fail(). | Mock session that tracks `Session.Fail()` calls; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory; force error by missing working directory | `AgentDefinition.AgentRoot="missing"` | Returns error; `Session.Fail()` not called |

### Resource Cleanup — Post-Start Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_PostStartFailure_TerminatesProcess_Unix` | `unit` | Terminates process with SIGTERM then SIGKILL on Unix after post-start failure. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; mock process/signal handlers for testing cleanup logic; mock post-start validation that fails; Unix platform | `NodeName="Node"` | Returns error matching `/post-start validation failed/i`; mock verifies SIGTERM sent, 5-second wait, then SIGKILL if needed |
| `TestAgentInvoker_PostStartFailure_TerminatesProcess_Windows` | `unit` | Terminates process with SIGKILL immediately on Windows after post-start failure. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; mock process/signal handlers for testing cleanup logic; mock post-start validation that fails; Windows platform | `NodeName="Node"` | Returns error matching `/post-start validation failed/i`; mock verifies SIGKILL sent immediately |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_ConcurrentInvocationsDifferentNodes` | `race` | Multiple concurrent invocations for different nodes are thread-safe. | Mock session with thread-safe implementations; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; 10 goroutines | Each goroutine invokes with different `NodeName` (`NodeA`, `NodeB`, ..., `NodeJ`) | All invocations succeed; no data races; each node has distinct Claude session ID stored |
| `TestAgentInvoker_ConcurrentSameNode_RaceCondition` | `race` | Concurrent invocations for same node may cause race conditions (documented behavior). | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; 5 goroutines | All goroutines invoke with same `NodeName="SameNode"` | Verifies no panic occurs; exact behavior undefined (may generate multiple UUIDs or race conditions); documents current implementation |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_RepeatedInvocationExistingSession` | `unit` | Repeated invocations with existing session ID behave identically. | Mock session returns same UUID for all `GetSessionDataSafe` calls; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory | Call AgentInvoker three times with same `NodeName` | All three calls succeed; all use same `--resume <UUID>` flag; `UpdateSessionDataSafe` never called |

### Boundary Values — Tools Arrays

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_AllowedToolsSingleItem` | `unit` | Handles single item in AllowedTools array. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition.AllowedTools=["Bash(*)"]` | Command includes `--allowed-tools "Bash(*)"` as single argument |
| `TestAgentInvoker_AllowedToolsMultipleItems` | `unit` | Handles multiple items in AllowedTools array. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition.AllowedTools=["Bash(*)", "Read(*)", "Edit(*)"]` | Command includes `--allowed-tools "Bash(*)" "Read(*)" "Edit(*)"` as separate arguments |
| `TestAgentInvoker_AllowedToolsNilVsEmpty` | `unit` | Distinguishes between nil and empty array for AllowedTools. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | Test both: `AllowedTools=nil` and `AllowedTools=[]` | Both cases omit `--allowed-tools` flag from command |

### Boundary Values — Model and Effort

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_ModelWithSpaces` | `unit` | Handles model identifiers with spaces. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition.Model="sonnet 4.0"` | Command includes `--model "sonnet 4.0"` passed safely as separate argument |
| `TestAgentInvoker_EffortSpecialChars` | `unit` | Handles effort values with special characters. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory; command interceptor | `AgentDefinition.Effort="high-priority"` | Command includes `--effort high-priority` passed safely |

### Boundary Values — Paths

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_AgentRootMultiLevel` | `unit` | Handles multi-level relative AgentRoot paths. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and nested `spec/logic/` subdirectories; ProjectRoot set to fixture directory | `AgentDefinition.AgentRoot="spec/logic"` | Command working directory is `<ProjectRoot>/spec/logic` (absolute path) |
| `TestAgentInvoker_ProjectRootAbsolutePath` | `unit` | ProjectRoot is always treated as absolute path. | Mock session; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory absolute path | `ProjectRoot=<fixture-dir-abs-path>`, `AgentDefinition.AgentRoot="agents"` | Command working directory is `<fixture-dir-abs-path>/agents` |

### Not Immutable — Session Data

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_SessionDataMutated` | `unit` | Session data is mutated by storing new Claude session ID. | Mock session with initial empty SessionData; isolated temporary fixture directory created using `t.TempDir()` with `.spectra/` and `agents/` subdirectories; ProjectRoot set to fixture directory | `NodeName="NewNode"` | `UpdateSessionDataSafe` called with `"NewNode.ClaudeSessionID"` and generated UUID; session data updated |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestAgentInvoker_CallsGetSessionDataSafe` | `unit` | Calls Session.GetSessionDataSafe with correct key. | Mock session that tracks method calls; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory | `NodeName="TestNode"` | `GetSessionDataSafe` called with key `"TestNode.ClaudeSessionID"` |
| `TestAgentInvoker_CallsUpdateSessionDataSafeOnNewSession` | `unit` | Calls Session.UpdateSessionDataSafe only for new sessions. | Mock session returns `(nil, false)` for `GetSessionDataSafe`; mock session that tracks method calls; isolated temporary fixture directory created using `t.TempDir()`; ProjectRoot set to fixture directory | `NodeName="Node"` | `UpdateSessionDataSafe` called exactly once with key `"Node.ClaudeSessionID"` and valid UUID string |
