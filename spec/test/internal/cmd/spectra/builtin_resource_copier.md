# Test Specification: `builtin_resource_copier_test.go`

## Source File Under Test

`internal/cmd/spectra/builtin_resource_copier.go`

## Test File

`internal/cmd/spectra/builtin_resource_copier_test.go`

---

## `BuiltinResourceCopier`

### Happy Path — CopyWorkflows

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyWorkflows_WritesAllFiles` | `unit` | Copies all embedded workflow files to target paths. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create the target directory `<projectRoot>/.spectra/workflows/`. Construct a test `embed.FS` (using `testing/fstest.MapFS`) containing `workflows/DefaultLogicSpec.yaml` and `workflows/DefaultTestSpec.yaml` with known content. Mock `StorageLayout.GetWorkflowPath` to return expected paths. | `projectRoot` = temp dir path | Returns empty warnings and `nil` error; both files exist at paths returned by `StorageLayout.GetWorkflowPath` with correct content and permissions `0644` |
| `TestBuiltinResourceCopier_CopyWorkflows_SkipsExisting` | `unit` | Skips existing workflow files and returns warnings. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create the target directory. Pre-create the target file for one workflow. Construct test `embed.FS` with that workflow. Mock `StorageLayout.GetWorkflowPath`. | `projectRoot` = temp dir path | Returns warning `"Warning: workflow definition '<name>.yaml' already exists, skipping"`; `nil` error; existing file content unchanged |

### Happy Path — CopyAgents

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyAgents_WritesAllFiles` | `unit` | Copies all embedded agent files to target paths. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create the target directory `<projectRoot>/.spectra/agents/`. Construct a test `embed.FS` containing `agents/TestAgent.yaml`. Mock `StorageLayout.GetAgentPath`. | `projectRoot` = temp dir path | Returns empty warnings and `nil` error; file exists at path returned by `StorageLayout.GetAgentPath` with correct content and permissions `0644` |
| `TestBuiltinResourceCopier_CopyAgents_SkipsExisting` | `unit` | Skips existing agent files and returns warnings. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create the target directory. Pre-create the target file. Construct test `embed.FS`. Mock `StorageLayout.GetAgentPath`. | `projectRoot` = temp dir path | Returns warning `"Warning: agent definition '<name>.yaml' already exists, skipping"`; `nil` error; existing file content unchanged |

### Happy Path — CopySpecFiles

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopySpecFiles_WritesAllFiles` | `unit` | Copies all embedded spec files preserving subdirectory structure. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `<projectRoot>/spec/`, `<projectRoot>/spec/logic/`, `<projectRoot>/spec/test/` directories. Construct a test `embed.FS` containing `spec/ARCHITECTURE.md`, `spec/CONVENTIONS.md`, `spec/logic/README.md`, `spec/test/README.md`. | `projectRoot` = temp dir path | Returns empty warnings and `nil` error; all files exist at `<projectRoot>/spec/<relativePath>` with correct content and permissions `0644` |
| `TestBuiltinResourceCopier_CopySpecFiles_SkipsExisting` | `unit` | Skips existing spec files and returns warnings. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create target directories. Pre-create `<projectRoot>/spec/ARCHITECTURE.md`. Construct test `embed.FS`. | `projectRoot` = temp dir path | Returns warning `"Warning: spec file 'ARCHITECTURE.md' already exists, skipping"`; `nil` error; existing file content unchanged |
| `TestBuiltinResourceCopier_CopySpecFiles_PreservesSubdirectoryStructure` | `unit` | Files in subdirectories are written to corresponding subdirectories in target. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create `<projectRoot>/spec/logic/` directory. Construct a test `embed.FS` containing `spec/logic/README.md`. | `projectRoot` = temp dir path | Returns `nil` error; file exists at `<projectRoot>/spec/logic/README.md` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyWorkflows_WriteError` | `unit` | Returns error when write fails (target directory missing). | Create a temporary directory as `projectRoot` using `t.TempDir()`. Do NOT create `<projectRoot>/.spectra/workflows/`. Construct test `embed.FS` with a workflow file. Mock `StorageLayout.GetWorkflowPath` to return path under the missing directory. | `projectRoot` = temp dir path | Returns error containing `"failed to write built-in file"` |
| `TestBuiltinResourceCopier_CopyAgents_WriteError` | `unit` | Returns error when write fails for agent files. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Do NOT create `<projectRoot>/.spectra/agents/`. Construct test `embed.FS` with an agent file. Mock `StorageLayout.GetAgentPath` to return path under the missing directory. | `projectRoot` = temp dir path | Returns error containing `"failed to write built-in file"` |
| `TestBuiltinResourceCopier_CopySpecFiles_WriteError` | `unit` | Returns error when write fails for spec files. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Do NOT create `<projectRoot>/spec/` directory. Construct test `embed.FS` with a spec file. | `projectRoot` = temp dir path | Returns error containing `"failed to write built-in file"` |
| `TestBuiltinResourceCopier_CopyWorkflows_FailFast` | `unit` | Stops processing after first write failure; returns accumulated warnings. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Construct test `embed.FS` with two workflow files. Mock `StorageLayout.GetWorkflowPath` to return a valid path for the first file (pre-create that file to generate a warning) and an invalid path for the second (missing parent dir). | `projectRoot` = temp dir path | Returns accumulated warning for the first file AND error for the second file; no files written after the failure |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyWorkflows_AllExist` | `unit` | Returns warnings for all files and nil error when all target files exist. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create target directory. Pre-create all target workflow files. Construct test `embed.FS`. Mock `StorageLayout.GetWorkflowPath`. | `projectRoot` = temp dir path | Returns one warning per file; `nil` error; no files modified |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyWorkflows_EmptyFS` | `unit` | Returns empty warnings and nil error when embedded FS has no files. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Construct an empty test `embed.FS` (no files in `workflows/`). | `projectRoot` = temp dir path | Returns empty warnings and `nil` error |
| `TestBuiltinResourceCopier_CopyWorkflows_EmptyFileContent` | `unit` | Writes empty file when embedded file has zero bytes. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create target directory. Construct test `embed.FS` with `workflows/Empty.yaml` containing empty content. Mock `StorageLayout.GetWorkflowPath`. | `projectRoot` = temp dir path | Returns `nil` error; target file exists with 0 bytes and permissions `0644` |

### Boundary Values — Target Exists as Directory

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyWorkflows_TargetIsDirectory` | `unit` | Skips with warning when target path exists as a directory. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create target workflows directory. Create a directory at the target file path (i.e., where the workflow file would be written). Construct test `embed.FS`. Mock `StorageLayout.GetWorkflowPath`. | `projectRoot` = temp dir path | Returns warning for the skipped entry; `nil` error |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestBuiltinResourceCopier_CopyWorkflows_UsesStorageLayout` | `unit` | Calls `StorageLayout.GetWorkflowPath` with correct arguments for each file. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create target directory. Construct test `embed.FS` with `workflows/MyWorkflow.yaml`. Set up mock `StorageLayout` that records calls. | `projectRoot` = temp dir path | Mock `StorageLayout.GetWorkflowPath` called with `(projectRoot, "MyWorkflow")` |
| `TestBuiltinResourceCopier_CopyAgents_UsesStorageLayout` | `unit` | Calls `StorageLayout.GetAgentPath` with correct arguments for each file. | Create a temporary directory as `projectRoot` using `t.TempDir()`. Create target directory. Construct test `embed.FS` with `agents/MyAgent.yaml`. Set up mock `StorageLayout` that records calls. | `projectRoot` = temp dir path | Mock `StorageLayout.GetAgentPath` called with `(projectRoot, "MyAgent")` |
