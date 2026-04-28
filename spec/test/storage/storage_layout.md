# Test Specification: `storage_layout.go`

## Source File Under Test
`storage/storage_layout.go`

## Test File
`storage/storage_layout_test.go`

---

## `StorageLayout`

### Happy Path — Path Constants

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestStorageLayout_SpectraDirConstant` | `unit` | Verify SpectraDir constant value. | | | `SpectraDir == ".spectra"` |
| `TestStorageLayout_SessionsDirConstant` | `unit` | Verify SessionsDir constant value. | | | `SessionsDir == ".spectra/sessions"` |
| `TestStorageLayout_WorkflowsDirConstant` | `unit` | Verify WorkflowsDir constant value. | | | `WorkflowsDir == ".spectra/workflows"` |
| `TestStorageLayout_AgentsDirConstant` | `unit` | Verify AgentsDir constant value. | | | `AgentsDir == ".spectra/agents"` |
| `TestStorageLayout_SessionMetadataFileConstant` | `unit` | Verify SessionMetadataFile constant value. | | | `SessionMetadataFile == "session.json"` |
| `TestStorageLayout_EventHistoryFileConstant` | `unit` | Verify EventHistoryFile constant value. | | | `EventHistoryFile == "events.jsonl"` |
| `TestStorageLayout_RuntimeSocketFileConstant` | `unit` | Verify RuntimeSocketFile constant value. | | | `RuntimeSocketFile == "runtime.sock"` |

### Happy Path — GetSpectraDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSpectraDir_AbsolutePath` | `unit` | Returns absolute path to .spectra directory. | | `ProjectRoot="/home/user/project"` | Returns `/home/user/project/.spectra` |
| `TestGetSpectraDir_TrailingSlash` | `unit` | Handles project root with trailing slash correctly. | | `ProjectRoot="/home/user/project/"` | Returns `/home/user/project/.spectra` without double slashes |

### Happy Path — GetSessionsDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionsDir_AbsolutePath` | `unit` | Returns absolute path to sessions directory. | | `ProjectRoot="/home/user/project"` | Returns `/home/user/project/.spectra/sessions` |

### Happy Path — GetWorkflowsDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetWorkflowsDir_AbsolutePath` | `unit` | Returns absolute path to workflows directory. | | `ProjectRoot="/home/user/project"` | Returns `/home/user/project/.spectra/workflows` |

### Happy Path — GetAgentsDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetAgentsDir_AbsolutePath` | `unit` | Returns absolute path to agents directory. | | `ProjectRoot="/home/user/project"` | Returns `/home/user/project/.spectra/agents` |

### Happy Path — GetSessionDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDir_ValidUUID` | `unit` | Returns absolute path to session directory with valid UUID. | | `ProjectRoot="/home/user/project"`, `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns `/home/user/project/.spectra/sessions/123e4567-e89b-12d3-a456-426614174000` |
| `TestGetSessionDir_UUIDWithUppercase` | `unit` | Preserves UUID case as provided. | | `ProjectRoot="/home/user/project"`, `SessionUUID="123E4567-E89B-12D3-A456-426614174000"` | Returns `/home/user/project/.spectra/sessions/123E4567-E89B-12D3-A456-426614174000` |

### Happy Path — GetSessionMetadataPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionMetadataPath_ValidUUID` | `unit` | Returns absolute path to session.json file. | | `ProjectRoot="/home/user/project"`, `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns `/home/user/project/.spectra/sessions/123e4567-e89b-12d3-a456-426614174000/session.json` |

### Happy Path — GetEventHistoryPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetEventHistoryPath_ValidUUID` | `unit` | Returns absolute path to events.jsonl file. | | `ProjectRoot="/home/user/project"`, `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns `/home/user/project/.spectra/sessions/123e4567-e89b-12d3-a456-426614174000/events.jsonl` |

### Happy Path — GetRuntimeSocketPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetRuntimeSocketPath_ValidUUID` | `unit` | Returns absolute path to runtime.sock file. | | `ProjectRoot="/home/user/project"`, `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns `/home/user/project/.spectra/sessions/123e4567-e89b-12d3-a456-426614174000/runtime.sock` |

### Happy Path — GetWorkflowPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetWorkflowPath_PascalCaseName` | `unit` | Returns absolute path to workflow YAML file. | | `ProjectRoot="/home/user/project"`, `WorkflowName="CodeReview"` | Returns `/home/user/project/.spectra/workflows/CodeReview.yaml` |

### Happy Path — GetAgentPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetAgentPath_PascalCaseRole` | `unit` | Returns absolute path to agent YAML file. | | `ProjectRoot="/home/user/project"`, `AgentRole="Architect"` | Returns `/home/user/project/.spectra/agents/Architect.yaml` |

### Boundary Values — Empty Inputs

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDir_EmptyUUID` | `unit` | Returns malformed path with empty UUID. | | `ProjectRoot="/home/user/project"`, `SessionUUID=""` | Returns `/home/user/project/.spectra/sessions/` (malformed, no validation performed) |
| `TestGetWorkflowPath_EmptyName` | `unit` | Returns malformed path with empty workflow name. | | `ProjectRoot="/home/user/project"`, `WorkflowName=""` | Returns `/home/user/project/.spectra/workflows/.yaml` (malformed, no validation performed) |
| `TestGetAgentPath_EmptyRole` | `unit` | Returns malformed path with empty agent role. | | `ProjectRoot="/home/user/project"`, `AgentRole=""` | Returns `/home/user/project/.spectra/agents/.yaml` (malformed, no validation performed) |

### Boundary Values — Relative Path Inputs

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSpectraDir_RelativePath` | `unit` | Returns relative path when project root is relative. | | `ProjectRoot="./project"` | Returns `./project/.spectra` (relative, no conversion to absolute) |
| `TestGetSessionDir_RelativePathProjectRoot` | `unit` | Returns relative path when project root is relative. | | `ProjectRoot="./project"`, `SessionUUID="123e4567-e89b-12d3-a456-426614174000"` | Returns `./project/.spectra/sessions/123e4567-e89b-12d3-a456-426614174000` (relative) |

### Boundary Values — Path Separators in Inputs

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDir_UUIDWithPathSeparator` | `unit` | Does not validate UUID format; passes through as-is. | | `ProjectRoot="/home/user/project"`, `SessionUUID="../malicious"` | Returns `/home/user/project/.spectra/sessions/../malicious` (potentially dangerous, no validation) |
| `TestGetWorkflowPath_NameWithPathSeparator` | `unit` | Does not validate workflow name; passes through as-is. | | `ProjectRoot="/home/user/project"`, `WorkflowName="../malicious/workflow"` | Returns `/home/user/project/.spectra/workflows/../malicious/workflow.yaml` (potentially dangerous) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestStorageLayout_IdempotentComposition` | `unit` | Multiple calls with same inputs return identical paths. | | Call `GetSessionDir("/home/user/project", "123e4567-e89b-12d3-a456-426614174000")` three times | All three calls return identical path string |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestStorageLayout_ConcurrentAccess` | `race` | Multiple goroutines call path composition methods concurrently. | | 10 goroutines each call `GetSessionDir`, `GetSessionMetadataPath`, `GetEventHistoryPath` with different UUIDs | All calls succeed; no data races detected; all paths correctly formed |

### Happy Path — Platform-Agnostic Path Separators

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestStorageLayout_PlatformSpecificSeparators` | `unit` | Uses platform-appropriate path separators. | | `ProjectRoot` with platform-specific format | Returned paths use platform-appropriate separators (`/` on Unix, `\` on Windows) |
