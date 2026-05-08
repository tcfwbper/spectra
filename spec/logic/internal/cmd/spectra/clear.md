# spectra clear

## Overview

The `spectra clear` command deletes session data from the `.spectra/sessions/` directory. It accepts 0 or more session UUIDs as positional arguments. If UUIDs are provided, only those sessions are deleted. If no UUIDs are provided, all sessions are deleted. In both cases, the command prompts for user confirmation before proceeding. The command does not check whether sessions are currently running.

## Boundaries

- Owns: positional argument parsing (0-N UUIDs).
- Owns: listing sessions directory contents (for delete-all case).
- Owns: composing session directory paths via StorageLayout.
- Owns: recursive deletion of session directories.
- Owns: individual error reporting (per-session) and continuation.
- Owns: output messages (success, warning, error).
- Delegates: project root discovery to SpectraFinder.
- Delegates: confirmation prompt to `cmdutil.ConfirmPrompt`.
- Delegates: error message formatting to `cmdutil.ErrorFormatter`.
- Must not: delete the `.spectra/sessions/` directory itself.
- Must not: check whether a session is currently running.
- Must not: validate UUID format.
- Must not: delete files directly within `.spectra/sessions/` (only subdirectories).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `storage.SpectraFinder` | Project root discovery | `Find()` | — |
| `storage.StorageLayout` | Path composition | `GetSessionsDir(projectRoot)`, `GetSessionDir(projectRoot, uuid)` | Must not use other path functions |
| `cmdutil.ConfirmPrompt` | User confirmation | `ConfirmPrompt(reader, writer, prompt)` | — |
| `cmdutil.ErrorFormatter` | Output formatting | `FormatError(msg)`, `FormatWarning(msg)` | — |

Construction constraint: Registered as a Cobra subcommand of the root command.

## Behavior

### Project Root Discovery

1. Calls `SpectraFinder.Find()` to locate the project root.
2. If SpectraFinder fails, prints `"Error: .spectra directory not found. Are you in a Spectra project?"` to stderr and exits with code 1.

### Case 1: Delete Specific Sessions (positional arguments provided)

3. One or more UUIDs are provided as positional arguments: `spectra clear <UUID1> [<UUID2> ...]`.
4. Prints confirmation prompt listing the UUIDs:
   ```
   Are you sure you want to delete the following sessions?
     - <UUID1>
     - <UUID2>
   [y/N]: 
   ```
5. Calls `cmdutil.ConfirmPrompt` with the formatted prompt.
6. If user does not confirm, prints `"Operation cancelled"` and exits with code 0.
7. If user confirms, iterates over each UUID:
   - Composes the session directory path via `StorageLayout.GetSessionDir(projectRoot, uuid)`.
   - Checks if the directory exists via `os.Stat()`.
   - If the directory does not exist, prints `"Warning: session '<UUID>' not found, skipping"` and continues.
   - If the directory exists, deletes it recursively via `os.RemoveAll()`.
   - If deletion succeeds, prints `"Session '<UUID>' cleared"`.
   - If deletion fails, prints `"Error: failed to clear session '<UUID>': <error>"` to stderr and continues.
8. After processing all UUIDs, exits with code 0.

### Case 2: Delete All Sessions (no positional arguments)

9. No UUIDs provided: `spectra clear`.
10. Reads the sessions directory via `os.ReadDir()` on `StorageLayout.GetSessionsDir(projectRoot)`.
11. If the sessions directory does not exist, prints `"Warning: sessions directory not found, nothing to clear"` and exits with code 0.
12. If reading the directory fails, prints `"Error: failed to read sessions directory: <error>"` to stderr and exits with code 1.
13. Filters entries to only directories (skips regular files).
14. If no directories exist, prints `"No sessions to clear"` and exits with code 0.
15. Prints confirmation prompt:
    ```
    Are you sure you want to delete all sessions? [y/N]: 
    ```
16. Calls `cmdutil.ConfirmPrompt` with the prompt.
17. If user does not confirm, prints `"Operation cancelled"` and exits with code 0.
18. If user confirms, iterates over each directory entry:
    - Deletes it recursively via `os.RemoveAll()`.
    - If deletion succeeds, prints `"Session '<directory-name>' cleared"`.
    - If deletion fails, prints `"Error: failed to clear session '<directory-name>': <error>"` to stderr and continues.
19. After processing all directories:
    - If zero deletions failed, prints `"All sessions cleared successfully"`.
    - If one or more deletions failed, does not print an additional summary (per-session error messages are sufficient).
    - Exits with code 0.

## Inputs

### Positional Arguments

| Argument | Type | Constraints | Required | Description |
|----------|------|-------------|----------|-------------|
| UUIDs | []string | 0 or more, no format validation | No | Session UUIDs to delete |

### Flags

| Flag | Type | Required | Default |
|------|------|----------|---------|
| `--help` | boolean | No | false |

### User Input

| Input | Type | Source | Description |
|-------|------|--------|-------------|
| Confirmation | string | stdin | `y` or `Y` to confirm deletion |

## Outputs

### stdout

- Confirmation prompt
- Success messages per session cleared
- Warning messages for missing sessions
- "No sessions to clear" message
- "Operation cancelled" message
- "All sessions cleared successfully" message

### stderr

- Error messages for failed operations

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | Sessions cleared, or no sessions, or user cancelled, or missing sessions warned |
| 1 | Error | SpectraFinder failed, or sessions directory unreadable |

## Invariants

1. **SpectraFinder Required**: Must use SpectraFinder to locate project root. Does not assume `.spectra/` is in CWD.
2. **Sessions Directory Preserved**: Must never delete `.spectra/sessions/` itself.
3. **No Session State Check**: Does not check if sessions are running.
4. **Confirmation Always Required**: Both delete-specific and delete-all require user confirmation.
5. **Individual Error Reporting**: When one session deletion fails, reports error and continues to next session.
6. **Recursive Deletion**: Session directories deleted recursively including all contents.
7. **Directory Only**: Only deletes directory entries in `.spectra/sessions/`. Files are skipped.
8. **No UUID Validation**: UUIDs are passed as-is to StorageLayout. No format checking.
9. **Exit Code 0 on Partial Success**: Even if some deletions fail, exits with code 0 (errors reported individually).

## Edge Cases

- Condition: No positional arguments, sessions directory is empty.
  Expected: Prints "No sessions to clear", exits with code 0. No confirmation prompt.

- Condition: No positional arguments, user enters `n` at confirmation.
  Expected: Prints "Operation cancelled", exits with code 0.

- Condition: Positional argument with non-existent UUID.
  Expected: After confirmation, prints `"Warning: session '<UUID>' not found, skipping"`. Continues processing.

- Condition: Multiple UUIDs provided, one exists and one does not.
  Expected: After confirmation, deletes existing one, warns about missing one.

- Condition: Deletion fails due to permission denied.
  Expected: Prints error for that session, continues to next.

- Condition: `.spectra/sessions/` contains regular files (not directories).
  Expected: Files are skipped in delete-all mode.

- Condition: stdin is EOF (piped empty input).
  Expected: ConfirmPrompt returns false. Prints "Operation cancelled", exits with code 0.

- Condition: User enters `yes` (not `y`).
  Expected: Treated as rejection. Prints "Operation cancelled", exits with code 0.

- Condition: `.spectra` not found.
  Expected: Prints error to stderr, exits with code 1.

- Condition: Session directory is a symlink to another directory.
  Expected: `os.RemoveAll()` removes the symlink, not the target.

- Condition: Empty string UUID provided (e.g., `spectra clear ""`).
  Expected: No UUID validation. StorageLayout produces malformed path. Stat likely fails, warning printed.

## Related

- [root](./root.md) - Parent command
- [SpectraFinder](../../../../storage/spectra_finder.md) - Locates project root
- [StorageLayout](../../../../storage/storage_layout.md) - Path composition
- [ConfirmPrompt](../../cmdutil/confirm_prompt.md) - User confirmation utility
- [ErrorFormatter](../../cmdutil/error_formatter.md) - Error message formatting
