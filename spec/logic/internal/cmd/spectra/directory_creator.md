# DirectoryCreator

## Overview

DirectoryCreator creates the `.spectra/` and `spec/` directory structures required for a Spectra project. It creates each directory individually with idempotent behavior: if a directory already exists, it is silently skipped. If creation fails, an error is returned immediately (fail-fast).

## Boundaries

- Owns: creating the fixed set of project directories with correct permissions.
- Owns: idempotent directory creation (skip if exists).
- Owns: fail-fast error on first creation failure.
- Delegates: orchestration (when to call) to init command.
- Must not: create files.
- Must not: delete directories.
- Must not: validate directory contents.

## Dependencies

None. Uses only Go standard library (`os`, `path/filepath`).

## Behavior

1. `CreateAll(projectRoot string) error`.
2. Creates the following directories in order, each with permissions `0755`:
   - `<projectRoot>/.spectra/`
   - `<projectRoot>/.spectra/sessions/`
   - `<projectRoot>/.spectra/workflows/`
   - `<projectRoot>/.spectra/agents/`
   - `<projectRoot>/spec/`
   - `<projectRoot>/spec/logic/`
   - `<projectRoot>/spec/test/`
3. For each directory:
   - If the directory already exists (and is a directory), skips silently.
   - If the directory does not exist, creates it with `os.Mkdir` and permissions `0755`.
   - If creation fails (permission denied, disk full, path exists as file), returns error: `"failed to create directory '<relative-path>': <error>"`.
4. If all directories are created/exist, returns nil.
5. Relative path in error messages is relative to `projectRoot` (e.g., `.spectra/sessions`, `spec/logic`).

## Inputs

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| projectRoot | string | Absolute path to project root directory | Yes |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| error | error | nil on success, descriptive error on first failure |

### Error Message Format

```
failed to create directory '<relative-path>': <underlying error>
```

## Invariants

1. **Idempotent**: Existing directories are silently skipped. No warnings or errors.
2. **Fail-Fast**: First directory creation failure stops processing. Subsequent directories are not attempted.
3. **Fixed Order**: Directories are created in the order listed. Parent directories are created before children.
4. **Permissions 0755**: All newly created directories have permissions `0755`.
5. **No File Creation**: Only creates directories, never files.
6. **No Deletion**: Never deletes existing directories or their contents.
7. **Partial State on Failure**: Directories created before the failure remain on disk.

## Edge Cases

- Condition: All directories already exist.
  Expected: Returns nil. No modifications.

- Condition: `.spectra/` exists but `.spectra/sessions/` does not.
  Expected: Skips `.spectra/`, creates `.spectra/sessions/`.

- Condition: `.spectra` exists as a file (not directory).
  Expected: `os.Mkdir` fails. Returns `"failed to create directory '.spectra': file exists"`.

- Condition: Permission denied on `.spectra/` creation.
  Expected: Returns error. No subsequent directories attempted.

- Condition: `spec/` exists but `spec/logic/` does not.
  Expected: Skips `spec/`, creates `spec/logic/`.

- Condition: CWD is read-only.
  Expected: First directory creation fails. Returns error.

## Related

- [init](./init.md) - Orchestrator that calls DirectoryCreator in Phase 1
- [StorageLayout](../../../../storage/storage_layout.md) - Defines `.spectra/` path constants (not used directly by DirectoryCreator; paths are hardcoded for init simplicity)
