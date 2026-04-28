# Test Specification: `builtin_resource_copier.go`

## Source File Under Test
`cmd/spectra/builtin_resource_copier.go`

## Test File
`cmd/spectra/builtin_resource_copier_test.go`

---

## `BuiltinResourceCopier`

### Happy Path — CopyWorkflows

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_AllNew` | `unit` | Copies all embedded workflow files when none exist at target paths. | Create temp project root with `.spectra/workflows/` directory; embed 2 workflow files (`SimpleSdd.yaml`, `Another.yaml`) | `projectRoot=<tempDir>` | Returns empty warnings slice and `nil` error; both files written with `0644` permissions |
| `TestCopyWorkflows_EmptyEmbedFS` | `unit` | Returns success with no files when embedded filesystem is empty. | Create temp project root with `.spectra/workflows/` directory; embed no workflow files | `projectRoot=<tempDir>` | Returns empty warnings slice and `nil` error; no files written |

### Happy Path — CopyAgents

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyAgents_AllNew` | `unit` | Copies all embedded agent files when none exist at target paths. | Create temp project root with `.spectra/agents/` directory; embed 2 agent files (`Architect.yaml`, `QaAnalyst.yaml`) | `projectRoot=<tempDir>` | Returns empty warnings slice and `nil` error; both files written with `0644` permissions |
| `TestCopyAgents_EmptyEmbedFS` | `unit` | Returns success with no files when embedded filesystem is empty. | Create temp project root with `.spectra/agents/` directory; embed no agent files | `projectRoot=<tempDir>` | Returns empty warnings slice and `nil` error; no files written |

### Happy Path — CopySpecFiles

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopySpecFiles_AllNew` | `unit` | Copies all embedded spec files when none exist at target paths. | Create temp project root with `spec/`, `spec/logic/`, `spec/test/` directories; embed 4 spec files (`ARCHITECTURE.md`, `CONVENTIONS.md`, `logic/README.md`, `test/README.md`) | `projectRoot=<tempDir>` | Returns empty warnings slice and `nil` error; all files written with `0644` permissions in correct subdirectories |
| `TestCopySpecFiles_EmptyEmbedFS` | `unit` | Returns success with no files when embedded filesystem is empty. | Create temp project root with `spec/` directory | `projectRoot=<tempDir>` | Returns empty warnings slice and `nil` error; no files written |
| `TestCopySpecFiles_PreservesDirectoryStructure` | `unit` | Preserves nested directory structure from embedded FS to target. | Create temp project root with `spec/`, `spec/logic/`, `spec/test/` directories; embed files at various depths (`ARCHITECTURE.md`, `logic/README.md`, `test/README.md`) | `projectRoot=<tempDir>` | Returns `nil` error; `spec/ARCHITECTURE.md`, `spec/logic/README.md`, `spec/test/README.md` exist with correct content |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_MultipleInvocations` | `unit` | Second invocation skips all files and returns warnings. | Create temp project root with `.spectra/workflows/` directory; embed 1 workflow file; invoke `CopyWorkflows` once successfully | `projectRoot=<tempDir>` (second invocation) | Returns warnings slice with 1 entry: `"Warning: workflow definition 'SimpleSdd.yaml' already exists, skipping"`; `nil` error; file content unchanged |
| `TestCopyAgents_MultipleInvocations` | `unit` | Second invocation skips all files and returns warnings. | Create temp project root with `.spectra/agents/` directory; embed 1 agent file; invoke `CopyAgents` once successfully | `projectRoot=<tempDir>` (second invocation) | Returns warnings slice with 1 entry: `"Warning: agent definition 'Architect.yaml' already exists, skipping"`; `nil` error; file content unchanged |
| `TestCopySpecFiles_MultipleInvocations` | `unit` | Second invocation skips all files and returns warnings. | Create temp project root with `spec/` directory; embed 2 spec files; invoke `CopySpecFiles` once successfully | `projectRoot=<tempDir>` (second invocation) | Returns warnings slice with 2 entries for each skipped file; `nil` error; file content unchanged |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_WriteFailsFirstFile` | `unit` | Returns error when writing the first workflow file fails. | Create temp project root; make `.spectra/workflows/` directory read-only (`0555`); embed 1 workflow file | `projectRoot=<tempDir>` | Returns empty warnings slice and error matching `"failed to write built-in file '.spectra/workflows/SimpleSdd.yaml': permission denied"` |
| `TestCopyWorkflows_WriteFailsSecondFile` | `unit` | Returns error when writing the second workflow file fails after first succeeds. | Create temp project root with `.spectra/workflows/` directory; embed 2 workflow files; after first file write, change directory to read-only | `projectRoot=<tempDir>` | Returns empty warnings slice and error matching `"failed to write built-in file"` for second file; first file remains on disk |
| `TestCopyWorkflows_WriteFailsAfterSkip` | `unit` | Returns collected warnings and error when write fails after skipping existing files. | Create temp project root with `.spectra/workflows/` directory; create existing workflow file `First.yaml`; embed 2 workflow files (`First.yaml`, `Second.yaml`); make directory read-only after checking first file | `projectRoot=<tempDir>` | Returns warnings slice with 1 entry for skipped file and error for failed write of second file |
| `TestCopyAgents_WriteFailsPermissionDenied` | `unit` | Returns error when writing agent file fails due to permission denied. | Create temp project root; make `.spectra/agents/` directory read-only (`0555`); embed 1 agent file | `projectRoot=<tempDir>` | Returns empty warnings slice and error matching `"failed to write built-in file '.spectra/agents/Architect.yaml': permission denied"` |
| `TestCopySpecFiles_WriteFailsPermissionDenied` | `unit` | Returns error when writing spec file fails due to permission denied. | Create temp project root; make `spec/` directory read-only (`0555`); embed 1 spec file (`ARCHITECTURE.md`) | `projectRoot=<tempDir>` | Returns empty warnings slice and error matching `"failed to write built-in file"` with path containing `spec/ARCHITECTURE.md` |
| `TestCopySpecFiles_TargetDirMissing` | `unit` | Returns error when target subdirectory does not exist for nested spec file. | Create temp project root with `spec/` directory but not `spec/logic/` subdirectory; embed file `logic/README.md` | `projectRoot=<tempDir>` | Returns error matching `"failed to write built-in file"` and `"no such file or directory"` for `spec/logic/README.md` |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_TargetDirMissing` | `unit` | Returns error when `.spectra/workflows/` directory does not exist. | Create temp project root without `.spectra/workflows/` directory; embed 1 workflow file | `projectRoot=<tempDir>` | Returns error matching `"failed to write built-in file"` and `"no such file or directory"` |
| `TestCopyAgents_TargetDirMissing` | `unit` | Returns error when `.spectra/agents/` directory does not exist. | Create temp project root without `.spectra/agents/` directory; embed 1 agent file | `projectRoot=<tempDir>` | Returns error matching `"failed to write built-in file"` and `"no such file or directory"` |

### Boundary Values — File Names

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_MultipleDotsInFilename` | `unit` | Extracts workflow name correctly from filename with multiple dots. | Create temp project root with `.spectra/workflows/` directory; embed workflow file `My.Workflow.v2.yaml` | `projectRoot=<tempDir>` | Returns `nil` error; file written to `.spectra/workflows/My.Workflow.v2.yaml` |
| `TestCopyWorkflows_MixedCaseFilename` | `unit` | Preserves case in workflow filename. | Create temp project root with `.spectra/workflows/` directory; embed workflow files `SimpleSDD.yaml`, `simpleWorkflow.yaml` | `projectRoot=<tempDir>` | Returns `nil` error; files written to `.spectra/workflows/SimpleSDD.yaml` and `.spectra/workflows/simpleWorkflow.yaml` |
| `TestCopyWorkflows_NoYamlExtension` | `unit` | Processes file without `.yaml` extension. | Create temp project root with `.spectra/workflows/` directory; embed file `README.txt` | `projectRoot=<tempDir>` | Returns `nil` error; file written (name extraction behavior depends on StorageLayout; file copied as-is) |
| `TestCopySpecFiles_NestedPath` | `unit` | Handles spec file in nested subdirectory. | Create temp project root with `spec/logic/` directory; embed file `logic/README.md` | `projectRoot=<tempDir>` | Returns `nil` error; file written to `spec/logic/README.md` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_EmptyProjectRoot` | `unit` | Handles empty project root path. | Create temp directory; embed 1 workflow file | `projectRoot=""` | Behavior depends on StorageLayout; may return error or write to relative path |
| `TestCopyWorkflows_RelativeProjectRoot` | `unit` | Handles relative project root path. | Create temp directory; embed 1 workflow file | `projectRoot="./project"` | Composes relative paths; writes files relative to current working directory |
| `TestCopyWorkflows_EmptyFileContent` | `unit` | Copies embedded file with empty content. | Create temp project root with `.spectra/workflows/` directory; embed workflow file with 0 bytes content | `projectRoot=<tempDir>` | Returns `nil` error; empty file written with `0644` permissions |
| `TestCopySpecFiles_EmptyFileContent` | `unit` | Copies embedded spec file with empty content. | Create temp project root with `spec/` directory; embed spec file with 0 bytes content | `projectRoot=<tempDir>` | Returns `nil` error; empty file written with `0644` permissions |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_UsesStorageLayout` | `unit` | Verifies copier uses StorageLayout to compose target paths. | Create temp project root with `.spectra/workflows/` directory; mock or spy on StorageLayout.GetWorkflowPath; embed 1 workflow file `SimpleSdd.yaml` | `projectRoot=<tempDir>` | `GetWorkflowPath` called with `projectRoot` and `"SimpleSdd"`; file written to returned path |
| `TestCopyAgents_UsesStorageLayout` | `unit` | Verifies copier uses StorageLayout to compose target paths for agents. | Create temp project root with `.spectra/agents/` directory; mock or spy on StorageLayout.GetAgentPath; embed 1 agent file `Architect.yaml` | `projectRoot=<tempDir>` | `GetAgentPath` called with `projectRoot` and `"Architect"`; file written to returned path |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_FileContentIndependent` | `unit` | Verifies written file content matches embedded file exactly. | Create temp project root with `.spectra/workflows/` directory; embed workflow file with specific YAML content (valid or invalid) | `projectRoot=<tempDir>` | Returns `nil` error; written file content is byte-for-byte identical to embedded file |
| `TestCopyAgents_FileContentIndependent` | `unit` | Verifies written agent file content matches embedded file exactly. | Create temp project root with `.spectra/agents/` directory; embed agent file with specific YAML content | `projectRoot=<tempDir>` | Returns `nil` error; written file content is byte-for-byte identical to embedded file |
| `TestCopySpecFiles_FileContentIndependent` | `unit` | Verifies written spec file content matches embedded file exactly. | Create temp project root with `spec/` directory; embed spec file with specific Markdown content | `projectRoot=<tempDir>` | Returns `nil` error; written file content is byte-for-byte identical to embedded file |

### Not Immutable

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_InvalidYAMLContent` | `unit` | Copies file with invalid YAML content without validation. | Create temp project root with `.spectra/workflows/` directory; embed workflow file with malformed YAML content | `projectRoot=<tempDir>` | Returns `nil` error; invalid YAML file written to disk (no validation performed) |
| `TestCopyAgents_InvalidYAMLContent` | `unit` | Copies agent file with invalid YAML content without validation. | Create temp project root with `.spectra/agents/` directory; embed agent file with malformed YAML content | `projectRoot=<tempDir>` | Returns `nil` error; invalid YAML file written to disk (no validation performed) |
| `TestCopySpecFiles_InvalidMarkdownContent` | `unit` | Copies spec file with invalid Markdown content without validation. | Create temp project root with `spec/` directory; embed spec file with malformed Markdown | `projectRoot=<tempDir>` | Returns `nil` error; file written to disk (no validation performed) |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_FilePermissions` | `unit` | Verifies written workflow files have correct permissions. | Create temp project root with `.spectra/workflows/` directory; embed 1 workflow file | `projectRoot=<tempDir>` | Returns `nil` error; written file has permissions `0644` (rw-r--r--) |
| `TestCopyAgents_FilePermissions` | `unit` | Verifies written agent files have correct permissions. | Create temp project root with `.spectra/agents/` directory; embed 1 agent file | `projectRoot=<tempDir>` | Returns `nil` error; written file has permissions `0644` (rw-r--r--) |
| `TestCopySpecFiles_FilePermissions` | `unit` | Verifies written spec files have correct permissions. | Create temp project root with `spec/` directory; embed 1 spec file | `projectRoot=<tempDir>` | Returns `nil` error; written file has permissions `0644` (rw-r--r--) |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_MixedExistingAndNew` | `unit` | Copies new files and skips existing files in single invocation. | Create temp project root with `.spectra/workflows/` directory; create existing file `First.yaml`; embed 3 workflow files (`First.yaml`, `Second.yaml`, `Third.yaml`) | `projectRoot=<tempDir>` | Returns warnings slice with 1 entry for `First.yaml`; `nil` error; `Second.yaml` and `Third.yaml` written; `First.yaml` unchanged |
| `TestCopyAgents_MixedExistingAndNew` | `unit` | Copies new agent files and skips existing files in single invocation. | Create temp project root with `.spectra/agents/` directory; create existing file `Architect.yaml`; embed 2 agent files (`Architect.yaml`, `QaAnalyst.yaml`) | `projectRoot=<tempDir>` | Returns warnings slice with 1 entry for `Architect.yaml`; `nil` error; `QaAnalyst.yaml` written; `Architect.yaml` unchanged |
| `TestCopySpecFiles_MixedExistingAndNew` | `unit` | Copies new spec files and skips existing files in single invocation. | Create temp project root with `spec/` directory; create existing file `ARCHITECTURE.md`; embed 2 spec files (`ARCHITECTURE.md`, `CONVENTIONS.md`) | `projectRoot=<tempDir>` | Returns warnings slice with 1 entry for `ARCHITECTURE.md`; `nil` error; `CONVENTIONS.md` written; `ARCHITECTURE.md` unchanged |

### Catch Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestCopyWorkflows_ExistingFileIsDirectory` | `unit` | Skips when target path exists as directory. | Create temp project root with `.spectra/workflows/` directory; create subdirectory `.spectra/workflows/SimpleSdd.yaml/`; embed workflow file `SimpleSdd.yaml` | `projectRoot=<tempDir>` | Returns warnings slice with entry: `"Warning: workflow definition 'SimpleSdd.yaml' already exists, skipping"`; `nil` error (os.Stat succeeds for directory) |
| `TestCopyAgents_ExistingFileIsDirectory` | `unit` | Skips when target path exists as directory. | Create temp project root with `.spectra/agents/` directory; create subdirectory `.spectra/agents/Architect.yaml/`; embed agent file `Architect.yaml` | `projectRoot=<tempDir>` | Returns warnings slice with entry: `"Warning: agent definition 'Architect.yaml' already exists, skipping"`; `nil` error |
| `TestCopySpecFiles_ExistingFileIsDirectory` | `unit` | Skips when target spec file path exists as directory. | Create temp project root with `spec/` directory; create subdirectory `spec/ARCHITECTURE.md/`; embed spec file `ARCHITECTURE.md` | `projectRoot=<tempDir>` | Returns warnings slice with entry: `"Warning: spec file 'ARCHITECTURE.md' already exists, skipping"`; `nil` error |
| `TestCopyWorkflows_ExistingFileUnreadable` | `unit` | Skips when target file exists but is not readable. | Create temp project root with `.spectra/workflows/` directory; create file `SimpleSdd.yaml` with permissions `0000`; embed workflow file `SimpleSdd.yaml` | `projectRoot=<tempDir>` | Returns warnings slice with entry for skipped file; `nil` error (os.Stat does not require read permission) |
