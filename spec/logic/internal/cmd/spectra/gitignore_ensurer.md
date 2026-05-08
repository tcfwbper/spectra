# GitignoreEnsurer

## Overview

GitignoreEnsurer ensures that `.gitignore` in the project root contains a `.spectra` entry. If `.gitignore` does not exist, it creates one. If it exists but does not contain `.spectra`, it appends the entry. If `.spectra` is already present, it does nothing. This module handles all `.gitignore`-related I/O for the init command.

## Boundaries

- Owns: reading `.gitignore`, checking for `.spectra` entry, creating or appending to `.gitignore`.
- Owns: line-by-line matching logic (trim spaces/tabs, exact match).
- Owns: newline handling (ensure valid file format after append).
- Delegates: orchestration decision (when to call) to init command.
- Must not: print any success or warning messages (silent operation).
- Must not: modify any other files.
- Must not: create directories.

## Dependencies

None. Uses only Go standard library (`os`, `bufio`, `strings`).

## Behavior

1. `Ensure(projectRoot string) error`.
2. Composes the `.gitignore` path as `filepath.Join(projectRoot, ".gitignore")`.
3. If `.gitignore` is a symbolic link, follows the symlink and operates on the target file.
4. If `.gitignore` does not exist:
   - Creates `.gitignore` with permissions `0644` containing exactly `.spectra\n`.
   - Returns nil.
5. If `.gitignore` exists:
   - Reads the file content line-by-line using `bufio.Scanner`.
   - For each line, trims leading/trailing spaces (` `) and tabs (`\t`). Other Unicode whitespace is NOT trimmed.
   - If any trimmed line equals exactly `.spectra`, returns nil (already present).
   - If no line matches after reading all lines:
     - If the file does not end with a newline, appends `\n.spectra\n`.
     - If the file ends with a newline, appends `.spectra\n`.
   - Returns nil on success.
6. If reading `.gitignore` fails, returns error: `"failed to read '.gitignore': <error>"`.
7. If writing/creating `.gitignore` fails, returns error: `"failed to update '.gitignore': <error>"`.

## Inputs

| Parameter | Type | Constraints | Required |
|-----------|------|-------------|----------|
| projectRoot | string | Absolute path to project root directory | Yes |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| error | error | nil on success, descriptive error on failure |

### Error Message Formats

| Format | Condition |
|--------|-----------|
| `"failed to read '.gitignore': <error>"` | Read operation failed |
| `"failed to update '.gitignore': <error>"` | Write/create operation failed |

## Invariants

1. **Exact Match**: Matches `.spectra` exactly after trimming only spaces and tabs. Does not match `.spectra/`, `# .spectra`, or other variations.
2. **Trim Only Spaces and Tabs**: Only ASCII space (` `) and tab (`\t`) are trimmed. Other Unicode whitespace (NBSP, etc.) is not trimmed.
3. **No Overwrite**: Never overwrites existing `.gitignore` content. Only appends or creates.
4. **Valid File Format**: Ensures the resulting file is valid (no missing newline between last existing line and new entry).
5. **Silent Operation**: Does not print any messages to stdout or stderr. Communicates only via return value.
6. **Symlink Following**: If `.gitignore` is a symlink, operates on the target file.
7. **File Permissions**: Created files use `0644`.
8. **Cross-Platform**: Writes `\n` (LF). Git accepts both LF and CRLF on all platforms. `bufio.Scanner` handles both `\n` and `\r\n` when reading.

## Edge Cases

- Condition: `.gitignore` does not exist.
  Expected: Creates file with content `.spectra\n`. Returns nil.

- Condition: `.gitignore` exists with `.spectra` already present.
  Expected: Returns nil. No modification.

- Condition: `.gitignore` exists with `  .spectra  ` (whitespace-padded).
  Expected: Matches after trim. Returns nil.

- Condition: `.gitignore` exists without `.spectra`, file ends with newline.
  Expected: Appends `.spectra\n`.

- Condition: `.gitignore` exists without `.spectra`, file does NOT end with newline.
  Expected: Appends `\n.spectra\n`.

- Condition: `.gitignore` contains `.spectra/` but not `.spectra`.
  Expected: Appends `.spectra\n` (`.spectra/` does not match).

- Condition: `.gitignore` contains `# .spectra` (commented).
  Expected: Appends `.spectra\n` (commented line does not match).

- Condition: `.gitignore` is a broken symbolic link.
  Expected: Returns `"failed to read '.gitignore': no such file or directory"`.

- Condition: `.gitignore` is read-only (`0444`).
  Expected: Read succeeds. If append needed, returns `"failed to update '.gitignore': permission denied"`.

- Condition: Disk full when creating/appending.
  Expected: Returns `"failed to update '.gitignore': no space left on device"`.

## Related

- [init](./init.md) - Orchestrator that calls GitignoreEnsurer in Phase 0
