# BuiltinResourceCopier

## Overview

BuiltinResourceCopier copies embedded built-in workflow, agent definition, and specification template files to their target locations during `spectra init`. It performs safe copying: files are only written if they do not already exist. If a file exists, it is skipped with a warning. The copier does not validate file content.

## Boundaries

- Owns: iterating embedded filesystem entries, checking target file existence, writing files.
- Owns: warning message generation for skipped files.
- Owns: fail-fast on write errors.
- Delegates: orchestration (when to call) to init command.
- Delegates: path composition for `.spectra/` targets to StorageLayout.
- Must not: create directories (assumes target directories already exist).
- Must not: overwrite existing files.
- Must not: validate YAML or Markdown content.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `storage.StorageLayout` | Path composition | `GetWorkflowPath(projectRoot, name)`, `GetAgentPath(projectRoot, name)` | Must not use session-related functions |
| `internal/builtin` | Embedded filesystems | Read `embed.FS` variables for workflows, agents, spec files | Must not modify |

Construction constraint: BuiltinResourceCopier is constructed with three `embed.FS` references from the `internal/builtin` package. These are populated at compile time via Go's `//go:embed` directive.

### Embedded File Sources

The `internal/builtin` package provides:

```go
//go:embed workflows/*.yaml
var Workflows embed.FS

//go:embed agents/*.yaml
var Agents embed.FS

//go:embed spec/ARCHITECTURE.md
//go:embed spec/CONVENTIONS.md
//go:embed spec/logic/README.md
//go:embed spec/test/README.md
var SpecFiles embed.FS
```

## Behavior

### CopyWorkflows

1. `CopyWorkflows(projectRoot string) (warnings []string, err error)`.
2. Iterates over all `.yaml` files in the embedded `workflows/` directory.
3. For each file:
   - Extracts workflow name by removing `.yaml` extension (e.g., `DefaultLogicSpec.yaml` → `DefaultLogicSpec`).
   - Composes target path via `StorageLayout.GetWorkflowPath(projectRoot, workflowName)`.
   - If target file exists (`os.Stat` succeeds): appends warning `"Warning: workflow definition '<name>.yaml' already exists, skipping"` and continues.
   - If target file does not exist: reads embedded content and writes to target with permissions `0644`.
   - If write fails: returns accumulated warnings and error `"failed to write built-in file '<targetPath>': <error>"`.
4. Returns warnings and nil error on success.

### CopyAgents

5. `CopyAgents(projectRoot string) (warnings []string, err error)`.
6. Same logic as CopyWorkflows but for agent definition files in `agents/`.
7. Warning format: `"Warning: agent definition '<name>.yaml' already exists, skipping"`.

### CopySpecFiles

8. `CopySpecFiles(projectRoot string) (warnings []string, err error)`.
9. Iterates over all files in the embedded `spec/` directory (preserving subdirectory structure).
10. For each file:
    - Composes target path: `filepath.Join(projectRoot, "spec", relativePath)` where `relativePath` is the path relative to `spec/` in the embedded FS (e.g., `ARCHITECTURE.md`, `logic/README.md`).
    - If target file exists: appends warning `"Warning: spec file '<relativePath>' already exists, skipping"` and continues.
    - If target file does not exist: reads embedded content and writes to target with permissions `0644`.
    - If write fails: returns accumulated warnings and error.
11. Returns warnings and nil error on success.

## Inputs

### For Construction

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowsFS | embed.FS | From `internal/builtin.Workflows` | Yes |
| AgentsFS | embed.FS | From `internal/builtin.Agents` | Yes |
| SpecFilesFS | embed.FS | From `internal/builtin.SpecFiles` | Yes |

### For Copy Methods

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| projectRoot | string | Absolute path to project root | Yes |

## Outputs

### For All Copy Methods

| Field | Type | Description |
|-------|------|-------------|
| warnings | []string | Warning messages for skipped files (may be empty) |
| error | error | nil if all files copied/skipped, non-nil on write failure |

### Warning Message Formats

| Method | Format |
|--------|--------|
| CopyWorkflows | `Warning: workflow definition '<name>.yaml' already exists, skipping` |
| CopyAgents | `Warning: agent definition '<name>.yaml' already exists, skipping` |
| CopySpecFiles | `Warning: spec file '<relativePath>' already exists, skipping` |

### Error Message Format

```
failed to write built-in file '<targetPath>': <underlying error>
```

## Invariants

1. **Safe Copy**: Must never overwrite existing files.
2. **No Validation**: Does not validate content of embedded files.
3. **No Directory Creation**: Assumes target directories already exist. Returns error if write fails due to missing directory.
4. **Fail-Fast**: First write failure stops processing. Subsequent files are not attempted.
5. **Warning Accumulation**: Warnings are accumulated even if an error occurs later.
6. **File Permissions**: Written files use `0644`.
7. **Stateless**: No state between method calls. Idempotent.
8. **StorageLayout for .spectra Paths**: Uses StorageLayout for workflow and agent target paths. Uses direct filepath.Join for spec paths.
9. **Directory Structure Preservation**: For spec files, subdirectory structure from embedded FS is preserved in target.

## Edge Cases

- Condition: All target files already exist.
  Expected: Returns warnings for each file, nil error.

- Condition: Embedded filesystem has no files (empty directory).
  Expected: Returns empty warnings, nil error.

- Condition: Target directory does not exist (e.g., `.spectra/workflows/` missing).
  Expected: Write fails. Returns error.

- Condition: Disk full during write.
  Expected: Returns error for that file. Previously written files remain.

- Condition: Target file exists as a directory.
  Expected: `os.Stat` succeeds. Skipped with warning.

- Condition: Embedded file content is empty (0 bytes).
  Expected: Writes empty file to target.

## Related

- [init](./init.md) - Orchestrator that calls BuiltinResourceCopier in Phase 2
- [StorageLayout](../../../../storage/storage_layout.md) - Path composition for .spectra/ targets
- [internal/builtin] - Embedded filesystem source (abstract package reference)
