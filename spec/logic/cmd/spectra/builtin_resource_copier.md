# BuiltinResourceCopier

## Overview

BuiltinResourceCopier is responsible for copying embedded built-in workflow, agent definition, and specification template files from the binary's embedded filesystem to the `.spectra/workflows/`, `.spectra/agents/`, and `spec/` directories during `spectra init`. It performs safe copying: files are only written if they do not already exist at the target path. If a file exists, the copier skips it and returns a warning message. The copier does not validate the content or structure of the embedded files.

## Behavior

1. BuiltinResourceCopier is initialized with three `embed.FS` instances: one for built-in workflows, one for built-in agents, and one for specification template files.
2. The embedded filesystems are populated at compile time using Go's `//go:embed` directive, referencing files in the `builtin/workflows/`, `builtin/agents/`, and `builtin/spec/` directories at the project root.
3. BuiltinResourceCopier provides three methods: `CopyWorkflows(projectRoot)`, `CopyAgents(projectRoot)`, and `CopySpecFiles(projectRoot)`.
4. `CopyWorkflows(projectRoot)` iterates over all embedded workflow YAML files in the `builtin/workflows/` embedded directory.
5. For each embedded workflow file (e.g., `DefaultLogicSpec.yaml`):
   - The copier uses `StorageLayout.GetWorkflowPath(projectRoot, workflowName)` to compose the target path.
   - The copier extracts the workflow name from the filename (removing the `.yaml` extension).
   - The copier checks if the target file already exists using `os.Stat()`.
   - If the file exists, the copier appends a warning to the result: `"Warning: workflow definition '<workflowName>.yaml' already exists, skipping"` and continues to the next file.
   - If the file does not exist, the copier reads the embedded file content and writes it to the target path using `os.WriteFile()` with permissions `0644` (rw-r--r--).
   - If the write operation succeeds, the copier continues to the next file.
   - If the write operation fails, the copier returns an error: `"failed to write built-in file '<targetPath>': <error>"` and stops processing.
6. `CopyAgents(projectRoot)` follows the same logic as `CopyWorkflows`, but for agent definition files in `builtin/agents/`.
7. `CopySpecFiles(projectRoot)` copies specification template files from the embedded `builtin/spec/` directory to the `spec/` directory in the project root.
8. For each embedded spec file (e.g., `builtin/spec/ARCHITECTURE.md`, `builtin/spec/logic/README.md`):
   - The copier preserves the relative directory structure within `spec/`.
   - The copier composes the target path by joining `projectRoot` with `spec/` and the relative path from `builtin/spec/`.
   - For example, `builtin/spec/logic/README.md` → `<projectRoot>/spec/logic/README.md`.
   - The copier checks if the target file already exists using `os.Stat()`.
   - If the file exists, the copier appends a warning to the result: `"Warning: spec file '<relativePath>' already exists, skipping"` (where `<relativePath>` is the path relative to `spec/`, e.g., `logic/README.md`) and continues to the next file.
   - If the file does not exist, the copier reads the embedded file content and writes it to the target path using `os.WriteFile()` with permissions `0644` (rw-r--r--).
   - If the write operation succeeds, the copier continues to the next file.
   - If the write operation fails, the copier returns an error: `"failed to write built-in file '<targetPath>': <error>"` and stops processing.
9. All three methods return two values: a list of warning messages (for skipped files) and an error (if any write operation failed).
10. If all files are successfully copied or skipped, the methods return the warning list and `nil` error.
11. The copier does not validate the YAML or Markdown syntax, structure, or naming conventions of the embedded files. It treats them as opaque binary data.
12. The copier assumes that the target directories (`.spectra/workflows/`, `.spectra/agents/`, and `spec/` subdirectories) already exist. It does not create directories.
13. The copier is stateless and can be called multiple times without side effects (idempotent).

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowsFS | embed.FS | Embedded filesystem containing `builtin/workflows/*.yaml` files | Yes |
| AgentsFS | embed.FS | Embedded filesystem containing `builtin/agents/*.yaml` files | Yes |
| SpecFilesFS | embed.FS | Embedded filesystem containing `builtin/spec/**/*.md` files | Yes |

### For CopyWorkflows Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### For CopyAgents Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |

### For CopySpecFiles Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the project root directory | Yes |

## Outputs

### For CopyWorkflows Method

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| Warnings | []string | List of warning messages for skipped files (may be empty) |
| Error | error | `nil` if all files were successfully copied or skipped |

**Error Case**:

| Field | Type | Description |
|-------|------|-------------|
| Warnings | []string | List of warning messages for files skipped before the error occurred |
| Error | error | Non-nil error if a write operation failed |

**Error Message Format**:

```
failed to write built-in file '<targetPath>': <underlying error>
```

**Warning Message Format**:

```
Warning: workflow definition '<workflowName>.yaml' already exists, skipping
```

### For CopyAgents Method

Same structure as `CopyWorkflows`, but for agent files.

**Warning Message Format**:

```
Warning: agent definition '<agentRole>.yaml' already exists, skipping
```

### For CopySpecFiles Method

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| Warnings | []string | List of warning messages for skipped files (may be empty) |
| Error | error | `nil` if all files were successfully copied or skipped |

**Error Case**:

| Field | Type | Description |
|-------|------|-------------|
| Warnings | []string | List of warning messages for files skipped before the error occurred |
| Error | error | Non-nil error if a write operation failed |

**Error Message Format**:

```
failed to write built-in file '<targetPath>': <underlying error>
```

**Warning Message Format**:

```
Warning: spec file '<relativePath>' already exists, skipping
```

Where `<relativePath>` is the path relative to `spec/` (e.g., `ARCHITECTURE.md`, `logic/README.md`, `test/README.md`).

## Invariants

1. **Safe Copying**: The copier must never overwrite existing files. If a file exists at the target path, it must be skipped.

2. **No Validation**: The copier must not validate the content, syntax, or structure of embedded files (YAML or Markdown). It treats them as opaque byte streams.

3. **No Directory Creation**: The copier assumes that target directories (`.spectra/workflows/`, `.spectra/agents/`, and `spec/` subdirectories) already exist. It does not create them.

4. **Stateless**: The copier must not maintain any internal state between method calls. It can be invoked multiple times safely.

5. **Fail-Fast**: If a write operation fails, the copier must return an error immediately and stop processing subsequent files.

6. **Warning Collection**: All warnings (skipped files) must be collected and returned, even if an error occurs later.

7. **File Permissions**: All written files must have permissions `0644` (rw-r--r--).

8. **Embedded Filesystem Access**: The copier accesses embedded files using the `embed.FS` API (`ReadFile`, `ReadDir`). It does not access the physical filesystem for source files.

9. **Filename-to-Name Extraction**: The workflow/agent name is extracted from the filename by removing the `.yaml` extension. For example, `DefaultLogicSpec.yaml` → `DefaultLogicSpec`. For spec files, the relative path within `builtin/spec/` is preserved when constructing the target path in `spec/`.

10. **StorageLayout Delegation**: The copier must use `StorageLayout` to compose target file paths for `.spectra/` files. For `spec/` files, the copier constructs paths by joining `projectRoot` with `spec/` and the relative path from the embedded filesystem.

11. **Directory Structure Preservation**: For spec files, the copier must preserve the directory structure from `builtin/spec/` to `spec/`. For example, `builtin/spec/logic/README.md` is copied to `spec/logic/README.md`, not `spec/README.md`.

## Edge Cases

- **Condition**: All embedded workflow files are successfully copied (no files exist at target paths).
  **Expected**: `CopyWorkflows()` returns an empty warnings list and `nil` error.

- **Condition**: Some embedded workflow files already exist at target paths, others do not.
  **Expected**: `CopyWorkflows()` copies non-existent files, skips existing files, and returns a warnings list containing one warning per skipped file. Error is `nil`.

- **Condition**: All embedded workflow files already exist at target paths.
  **Expected**: `CopyWorkflows()` skips all files and returns a warnings list with one warning per file. Error is `nil`.

- **Condition**: Writing the first workflow file fails (permission denied).
  **Expected**: `CopyWorkflows()` returns an empty warnings list and an error: `"failed to write built-in file '.spectra/workflows/DefaultLogicSpec.yaml': permission denied"`. Subsequent files are not processed.

- **Condition**: Writing the second workflow file fails (disk full), after successfully copying the first file.
  **Expected**: `CopyWorkflows()` returns an error: `"failed to write built-in file '.spectra/workflows/AnotherWorkflow.yaml': no space left on device"`. The first file remains on disk. Warnings list is empty (no files were skipped).

- **Condition**: Writing the third workflow file fails, after skipping the first file (already exists) and copying the second file.
  **Expected**: `CopyWorkflows()` returns a warnings list containing one warning (for the first file) and an error for the third file. The second file remains on disk.

- **Condition**: Embedded filesystem contains no workflow files (empty `builtin/workflows/` directory).
  **Expected**: `CopyWorkflows()` returns an empty warnings list and `nil` error. No files are copied.

- **Condition**: Embedded filesystem contains a file with no `.yaml` extension (e.g., `README.txt`).
  **Expected**: The copier processes all files, including non-YAML files. It extracts the name by removing `.yaml` extension. For `README.txt`, the name is `README.txt` (no extension removed). StorageLayout composes the path `.spectra/workflows/README.txt.yaml`. The file is copied to this path.

- **Condition**: Embedded filesystem contains a file with multiple dots in the name (e.g., `My.Workflow.v2.yaml`).
  **Expected**: The copier extracts the name by removing the `.yaml` extension: `My.Workflow.v2`. StorageLayout composes the path `.spectra/workflows/My.Workflow.v2.yaml`. The file is copied.

- **Condition**: Embedded filesystem contains a file with uppercase and lowercase letters (e.g., `DefaultLOGICSPEC.yaml`, `defaultLogicSpec.yaml`).
  **Expected**: The copier treats filenames as-is without case conversion. Files are copied to `.spectra/workflows/DefaultLOGICSPEC.yaml` and `.spectra/workflows/defaultLogicSpec.yaml`.

- **Condition**: Target directory `.spectra/workflows/` does not exist.
  **Expected**: `os.WriteFile()` fails with "no such file or directory". The copier returns an error: `"failed to write built-in file '.spectra/workflows/DefaultLogicSpec.yaml': no such file or directory"`.

- **Condition**: `ProjectRoot` is a relative path (e.g., `"./project"`).
  **Expected**: StorageLayout composes relative paths. `os.Stat()` and `os.WriteFile()` operate on relative paths. This may succeed or fail depending on the current working directory. The copier does not validate `ProjectRoot` format.

- **Condition**: `ProjectRoot` is an empty string.
  **Expected**: StorageLayout composes paths like `workflows/DefaultLogicSpec.yaml` (relative to current directory). This may succeed or fail. The copier does not validate `ProjectRoot`.

- **Condition**: Embedded file content is empty (0 bytes).
  **Expected**: The copier writes an empty file to the target path. `os.WriteFile()` succeeds. The file exists but is empty.

- **Condition**: Embedded file content is invalid YAML (syntax errors, missing fields).
  **Expected**: The copier copies the file as-is without validation. The file will fail validation when loaded by WorkflowDefinitionLoader or AgentDefinitionLoader later.

- **Condition**: Embedded file has permissions `0600` (read-write for owner only) in the source directory.
  **Expected**: The copier ignores the source file's permissions. Written files always have permissions `0644`.

- **Condition**: Target file exists but is a directory, not a regular file.
  **Expected**: `os.Stat()` succeeds (directory exists). The copier skips the file and returns a warning. No error is returned (because the copier checks for existence, not file type).

- **Condition**: Target file exists but is not readable (permission denied).
  **Expected**: `os.Stat()` succeeds (file exists, stat does not require read permissions). The copier skips the file and returns a warning.

- **Condition**: Multiple goroutines call `CopyWorkflows()` concurrently with the same `ProjectRoot`.
  **Expected**: Both goroutines attempt to write the same files. Race conditions may occur. This is not a supported use case; the copier is not thread-safe for concurrent writes to the same target directory.

- **Condition**: `CopyWorkflows()` is called multiple times sequentially with the same `ProjectRoot`.
  **Expected**: First call copies all files. Second call skips all files (all exist) and returns warnings. This is idempotent behavior.

- **Condition**: All embedded spec files are successfully copied (no files exist at target paths).
  **Expected**: `CopySpecFiles()` returns an empty warnings list and `nil` error.

- **Condition**: Some embedded spec files already exist at target paths, others do not.
  **Expected**: `CopySpecFiles()` copies non-existent files, skips existing files, and returns a warnings list containing one warning per skipped file. Error is `nil`.

- **Condition**: All embedded spec files already exist at target paths.
  **Expected**: `CopySpecFiles()` skips all files and returns a warnings list with one warning per file. Error is `nil`.

- **Condition**: Writing `spec/ARCHITECTURE.md` fails (permission denied).
  **Expected**: `CopySpecFiles()` returns an error: `"failed to write built-in file '<projectRoot>/spec/ARCHITECTURE.md': permission denied"`. Subsequent files are not processed. Warnings list contains warnings for any files that were skipped before the error.

- **Condition**: Writing `spec/logic/README.md` fails (disk full), after successfully copying `spec/ARCHITECTURE.md` and `spec/CONVENTIONS.md`.
  **Expected**: `CopySpecFiles()` returns an error: `"failed to write built-in file '<projectRoot>/spec/logic/README.md': no space left on device"`. The first two files remain on disk.

- **Condition**: Target directory `spec/logic/` does not exist when attempting to write `spec/logic/README.md`.
  **Expected**: `os.WriteFile()` fails with "no such file or directory". The copier returns an error: `"failed to write built-in file '<projectRoot>/spec/logic/README.md': no such file or directory"`.

- **Condition**: `spec/ARCHITECTURE.md` exists as a directory, not a regular file.
  **Expected**: `os.Stat()` succeeds (directory exists). The copier skips the file and returns a warning: `"Warning: spec file 'ARCHITECTURE.md' already exists, skipping"`. No error is returned.

- **Condition**: Embedded `builtin/spec/` filesystem contains no files (empty directory).
  **Expected**: `CopySpecFiles()` returns an empty warnings list and `nil` error. No files are copied.

- **Condition**: Embedded spec file content is empty (0 bytes).
  **Expected**: The copier writes an empty file to the target path. `os.WriteFile()` succeeds. The file exists but is empty.

- **Condition**: Embedded spec file content is invalid Markdown (e.g., malformed syntax).
  **Expected**: The copier copies the file as-is without validation. The file will not be validated by the system (Markdown files are not validated like YAML files).

- **Condition**: Embedded `builtin/spec/` contains subdirectories beyond `logic/` and `test/` (e.g., `builtin/spec/extra/file.md`).
  **Expected**: The copier preserves the directory structure and copies the file to `spec/extra/file.md`. If the `spec/extra/` directory does not exist, the write operation fails.

- **Condition**: `CopySpecFiles()` is called multiple times sequentially with the same `ProjectRoot`.
  **Expected**: First call copies all files. Second call skips all files (all exist) and returns warnings. This is idempotent behavior.

## Related

- [init Subcommand](./init.md) - Uses BuiltinResourceCopier to initialize projects
- [StorageLayout](../../storage/storage_layout.md) - Provides path composition for target files
- [WorkflowDefinitionLoader](../../storage/workflow_definition_loader.md) - Loads and validates workflow files after copying
- [AgentDefinitionLoader](../../storage/agent_definition_loader.md) - Loads and validates agent files after copying
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - Framework architecture overview
