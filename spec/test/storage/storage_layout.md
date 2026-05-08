# Test Specification: `storage_layout_test.go`

## Source File Under Test
`storage/storage_layout.go`

## Test File
`storage/storage_layout_test.go`

---

## `StorageLayout`

### Happy Path — GetSpectraDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSpectraDir_AbsolutePath` | `unit` | Returns correct absolute path to `.spectra` directory. | | `projectRoot="/home/user/project"` | Returns `"/home/user/project/.spectra"` |
| `TestGetSpectraDir_TrailingSlash` | `unit` | Normalizes trailing slash in projectRoot. | | `projectRoot="/home/user/project/"` | Returns path without double separators; equivalent to `"/home/user/project/.spectra"` |

### Happy Path — GetSessionsDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionsDir_AbsolutePath` | `unit` | Returns correct absolute path to `.spectra/sessions` directory. | | `projectRoot="/home/user/project"` | Returns `"/home/user/project/.spectra/sessions"` |

### Happy Path — GetWorkflowsDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetWorkflowsDir_AbsolutePath` | `unit` | Returns correct absolute path to `.spectra/workflows` directory. | | `projectRoot="/home/user/project"` | Returns `"/home/user/project/.spectra/workflows"` |

### Happy Path — GetAgentsDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetAgentsDir_AbsolutePath` | `unit` | Returns correct absolute path to `.spectra/agents` directory. | | `projectRoot="/home/user/project"` | Returns `"/home/user/project/.spectra/agents"` |

### Happy Path — GetSessionDir

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDir_ValidUUID` | `unit` | Returns correct absolute path to a session directory. | | `projectRoot="/home/user/project"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns `"/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000"` |

### Happy Path — GetSessionMetadataPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionMetadataPath_ValidUUID` | `unit` | Returns correct path to session.json within session directory. | | `projectRoot="/home/user/project"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns `"/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/session.json"` |

### Happy Path — GetEventHistoryPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetEventHistoryPath_ValidUUID` | `unit` | Returns correct path to events.jsonl within session directory. | | `projectRoot="/home/user/project"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns `"/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/events.jsonl"` |

### Happy Path — GetRuntimeSocketPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetRuntimeSocketPath_ValidUUID` | `unit` | Returns correct path to runtime.sock within session directory. | | `projectRoot="/home/user/project"`, `sessionUUID="550e8400-e29b-41d4-a716-446655440000"` | Returns `"/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/runtime.sock"` |

### Happy Path — GetWorkflowPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetWorkflowPath_ValidName` | `unit` | Returns correct path to a workflow YAML file. | | `projectRoot="/home/user/project"`, `workflowName="CodeReview"` | Returns `"/home/user/project/.spectra/workflows/CodeReview.yaml"` |

### Happy Path — GetAgentPath

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetAgentPath_ValidRole` | `unit` | Returns correct path to an agent YAML file. | | `projectRoot="/home/user/project"`, `agentRole="Architect"` | Returns `"/home/user/project/.spectra/agents/Architect.yaml"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSessionDir_EmptyUUID` | `unit` | Returns malformed path when sessionUUID is empty. | | `projectRoot="/home/user/project"`, `sessionUUID=""` | Returns a path ending with `.spectra/sessions/`; no error or panic |
| `TestGetSpectraDir_EmptyProjectRoot` | `unit` | Returns relative path when projectRoot is empty. | | `projectRoot=""` | Returns `.spectra`; no error or panic |
| `TestGetWorkflowPath_EmptyName` | `unit` | Returns malformed path when workflowName is empty. | | `projectRoot="/home/user/project"`, `workflowName=""` | Returns path ending with `.yaml`; no error or panic |
| `TestGetAgentPath_EmptyRole` | `unit` | Returns malformed path when agentRole is empty. | | `projectRoot="/home/user/project"`, `agentRole=""` | Returns path ending with `.yaml`; no error or panic |

### Boundary Values — projectRoot

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestGetSpectraDir_RelativeProjectRoot` | `unit` | Returns relative path when projectRoot is relative. | | `projectRoot="./project"` | Returns `"project/.spectra"` (filepath.Join normalizes `./`) |
| `TestGetSessionDir_PathSeparatorInUUID` | `unit` | UUID containing path separators produces a traversal path. | | `projectRoot="/home/user/project"`, `sessionUUID="../malicious"` | Returns `"/home/user/project/.spectra/sessions/../malicious"` or its normalized form; no error or panic |

