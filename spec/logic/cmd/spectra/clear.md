# spectra clear Command

## Overview

The `spectra clear` command deletes session data from the `.spectra/sessions/` directory. It can delete a specific session by UUID or all sessions in the directory. When deleting all sessions, the command prompts the user for confirmation. The command does not check whether sessions are currently running; it deletes files regardless of session state.

## Behavior

### Deletion Flow

1. The `clear` command is invoked as `spectra clear` (delete all sessions) or `spectra clear --session-id <UUID>` (delete a specific session).
2. The command uses `SpectraFinder` to locate the project root directory (the directory containing `.spectra/`).
3. If `SpectraFinder` cannot find the project root, the command exits with code 1 and prints: `"Error: .spectra directory not found. Are you in a Spectra project?"`.
4. The command uses `StorageLayout` to compose the absolute path to the sessions directory: `.spectra/sessions/`.
5. **Case 1: Delete Specific Session** (if `--session-id` flag is provided):
   - The command composes the absolute path to the session directory: `.spectra/sessions/<SessionID>/`.
   - The command checks if the session directory exists using `os.Stat()`.
   - If the session directory does not exist, the command prints a warning: `"Warning: session '<SessionID>' not found, skipping"` and exits with code 0.
   - If the session directory exists, the command deletes the entire directory and all its contents recursively using `os.RemoveAll()`.
   - If deletion succeeds, the command prints: `"Session '<SessionID>' cleared successfully"` and exits with code 0.
   - If deletion fails (e.g., permission denied), the command prints: `"Error: failed to clear session '<SessionID>': <error>"` and exits with code 1.
6. **Case 2: Delete All Sessions** (if `--session-id` flag is not provided):
   - The command lists all entries in `.spectra/sessions/` using `os.ReadDir()`.
   - If the sessions directory does not exist, the command prints a warning: `"Warning: sessions directory not found, nothing to clear"` and exits with code 0.
   - If the sessions directory is empty (no subdirectories), the command prints: `"No sessions to clear"` and exits with code 0.
   - If the sessions directory contains one or more subdirectories, the command prompts the user for confirmation: `"Are you sure you want to delete all sessions? [y/N]: "`.
   - The command reads the user's input from stdin.
   - If the user enters `y` or `Y`, the command proceeds to delete all session directories.
   - If the user enters anything else (including empty input, `n`, `N`, or any other text), the command prints: `"Operation cancelled"` and exits with code 0.
   - For each subdirectory in `.spectra/sessions/`:
     - The command checks if the entry is a directory (not a file).
     - If the entry is a file, the command skips it (does not delete).
     - If the entry is a directory, the command deletes it recursively using `os.RemoveAll()`.
     - If deletion succeeds, the command prints: `"Session '<subdirectory-name>' cleared"`.
     - If deletion fails, the command prints: `"Error: failed to clear session '<subdirectory-name>': <error>"` and continues to the next directory.
   - After attempting to delete all session directories, the command prints: `"All sessions cleared successfully"` and exits with code 0 (even if some deletions failed; errors are reported individually).
7. The command does NOT delete the `.spectra/sessions/` directory itself. Only the subdirectories (session directories) are deleted.
8. The command does NOT check whether a session is currently running (e.g., checking for `runtime.sock` or session lock). It deletes files regardless of session state.

### Command Syntax

```
spectra clear
spectra clear --session-id <UUID>
```

### Usage Information

When invoked with `--help`:

```
Clear session data

Usage:
  spectra clear [flags]

Flags:
  --session-id string   UUID of the session to clear (if not provided, clears all sessions)
  --help                Show help information

Examples:
  spectra clear
  spectra clear --session-id 12345678-1234-1234-1234-123456789abc
```

### Success Output (stdout)

When deleting a specific session (success):

```
Session '12345678-1234-1234-1234-123456789abc' cleared successfully
```

When deleting all sessions (after confirmation):

```
Session '12345678-1234-1234-1234-123456789abc' cleared
Session '87654321-4321-4321-4321-cba987654321' cleared
All sessions cleared successfully
```

When no sessions exist:

```
No sessions to clear
```

When user cancels deletion:

```
Operation cancelled
```

### Warning Output (stdout)

When session not found:

```
Warning: session '12345678-1234-1234-1234-123456789abc' not found, skipping
```

When sessions directory not found:

```
Warning: sessions directory not found, nothing to clear
```

### Error Output (stderr)

When `.spectra` not found:

```
Error: .spectra directory not found. Are you in a Spectra project?
```

When session deletion fails:

```
Error: failed to clear session '12345678-1234-1234-1234-123456789abc': permission denied
```

### Confirmation Prompt

When deleting all sessions:

```
Are you sure you want to delete all sessions? [y/N]: 
```

User must type `y` or `Y` and press Enter to confirm. Any other input (including just pressing Enter) cancels the operation.

## Inputs

### Flags

| Flag | Type | Constraints | Required | Default |
|------|------|-------------|----------|---------|
| `--session-id` | string (UUID) | Valid UUID v4 format (not validated by command) | No | None |
| `--help` | boolean | N/A | No | false |

### Environment

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Current Working Directory | string | Process environment | Yes (implicit, used by SpectraFinder) |

### User Input (for delete all confirmation)

| Input | Type | Source | Required |
|-------|------|--------|----------|
| Confirmation | string | stdin (read from terminal) | Yes (if deleting all sessions) |

## Outputs

### stdout

- Success messages for cleared sessions
- Warning messages for missing sessions/directories
- Confirmation prompt (when deleting all sessions)
- Cancellation message (if user declines confirmation)

### stderr

- Error messages for failed operations

### Exit Codes

| Code | Meaning | Trigger Conditions |
|------|---------|-------------------|
| 0 | Success | Session(s) cleared successfully, or no sessions to clear, or user cancelled operation, or session not found (warning) |
| 1 | Error | `.spectra` not found, or session deletion failed |

### Filesystem Changes

- Session directories deleted: `.spectra/sessions/<SessionID>/` (and all contents)
- The `.spectra/sessions/` directory itself is NOT deleted

## Invariants

1. **SpectraFinder Requirement**: The command must use SpectraFinder to locate the project root. It must not assume `.spectra/` exists in the current directory.

2. **Sessions Directory Preservation**: The command must never delete the `.spectra/sessions/` directory itself. Only subdirectories (session directories) are deleted.

3. **No Session State Checking**: The command must not check whether a session is currently running. It deletes files regardless of session state.

4. **Confirmation for Delete All**: When deleting all sessions (no `--session-id` flag), the command must prompt the user for confirmation and only proceed if the user explicitly confirms with `y` or `Y`.

5. **No Confirmation for Single Session**: When deleting a specific session (with `--session-id` flag), the command must not prompt for confirmation. It proceeds immediately.

6. **Idempotent Delete**: If a session directory does not exist, the command must print a warning and exit with code 0 (not an error).

7. **Individual Error Reporting**: When deleting all sessions, if one session deletion fails, the command must report the error and continue attempting to delete other sessions. It must not abort on first error.

8. **Recursive Deletion**: Session directories are deleted recursively, including all files (`session.json`, `events.jsonl`, `runtime.sock`, etc.).

9. **File vs Directory Handling**: When listing entries in `.spectra/sessions/`, the command must only delete directories. Files (if any) are skipped.

10. **UUID Validation**: The command does not validate the UUID format. If an invalid UUID is provided, the command composes a path with the invalid UUID and attempts to delete it. If the directory does not exist, it prints a warning.

## Edge Cases

- **Condition**: User invokes `spectra clear --session-id 12345678-1234-1234-1234-123456789abc` (session exists).
  **Expected**: Delete the session directory and print `"Session '12345678-1234-1234-1234-123456789abc' cleared successfully"`. Exit with code 0.

- **Condition**: User invokes `spectra clear --session-id 12345678-1234-1234-1234-123456789abc` (session does not exist).
  **Expected**: Print `"Warning: session '12345678-1234-1234-1234-123456789abc' not found, skipping"`. Exit with code 0.

- **Condition**: User invokes `spectra clear` (no sessions exist).
  **Expected**: Print `"No sessions to clear"`. Exit with code 0. Do not prompt for confirmation.

- **Condition**: User invokes `spectra clear` (multiple sessions exist), then enters `y` at the confirmation prompt.
  **Expected**: Delete all session directories, print success messages, and exit with code 0.

- **Condition**: User invokes `spectra clear` (multiple sessions exist), then enters `n` at the confirmation prompt.
  **Expected**: Print `"Operation cancelled"`. Exit with code 0. No sessions are deleted.

- **Condition**: User invokes `spectra clear` (multiple sessions exist), then presses Enter without typing anything.
  **Expected**: Treat as cancellation. Print `"Operation cancelled"`. Exit with code 0. No sessions are deleted.

- **Condition**: User invokes `spectra clear` (multiple sessions exist), then enters `yes` (not `y` or `Y`).
  **Expected**: Treat as cancellation. Print `"Operation cancelled"`. Exit with code 0. No sessions are deleted.

- **Condition**: `.spectra/sessions/` directory does not exist.
  **Expected**: Print `"Warning: sessions directory not found, nothing to clear"`. Exit with code 0.

- **Condition**: `.spectra/sessions/` contains a regular file (not a directory), e.g., `.spectra/sessions/somefile.txt`.
  **Expected**: The command skips the file (does not delete it). Only directories are deleted.

- **Condition**: User invokes `spectra clear --session-id invalid-uuid` (not a valid UUID format).
  **Expected**: The command does not validate UUID format. It composes the path `.spectra/sessions/invalid-uuid/` and attempts to delete it. If it does not exist, print `"Warning: session 'invalid-uuid' not found, skipping"`.

- **Condition**: User invokes `spectra clear --session-id 12345678-1234-1234-1234-123456789abc` while a session is currently running (has `runtime.sock` or lock).
  **Expected**: The command does not check session state. It attempts to delete the directory. If the OS allows deletion (no open file handles), the session directory is deleted. If the OS prevents deletion (file in use), the command prints an error: `"Error: failed to clear session '<SessionID>': resource busy"` and exits with code 1.

- **Condition**: User invokes `spectra clear` and one session deletion fails (permission denied), but other sessions succeed.
  **Expected**: Print an error for the failed session, print success messages for other sessions, and exit with code 0 (because some deletions succeeded).

- **Condition**: User invokes `spectra clear` from a directory without `.spectra/`.
  **Expected**: SpectraFinder fails. Print `"Error: .spectra directory not found. Are you in a Spectra project?"` and exit with code 1.

- **Condition**: `.spectra/sessions/` directory exists but is not readable (permission denied).
  **Expected**: `os.ReadDir()` fails. Print `"Error: failed to read sessions directory: permission denied"` and exit with code 1.

- **Condition**: User invokes `spectra clear --session-id 12345678-1234-1234-1234-123456789abc` and the session directory is not empty (contains files and subdirectories).
  **Expected**: `os.RemoveAll()` deletes the directory and all its contents recursively. Print `"Session '12345678-1234-1234-1234-123456789abc' cleared successfully"`.

- **Condition**: User invokes `spectra clear --session-id ""` (empty string).
  **Expected**: The command composes the path `.spectra/sessions//`. This path is invalid. `os.Stat()` returns "not found". Print `"Warning: session '' not found, skipping"` and exit with code 0.

- **Condition**: User invokes `spectra clear --help`.
  **Expected**: Print usage information to stdout and exit with code 0. Do not invoke SpectraFinder or delete any sessions.

- **Condition**: User invokes `spectra clear` on a Windows system and stdin is not a terminal (e.g., piped input).
  **Expected**: The confirmation prompt is printed, but reading from stdin may fail or return EOF. The command should handle this gracefully by treating it as cancellation (exit with code 0, print "Operation cancelled").

- **Condition**: User invokes `spectra clear` and the sessions directory contains a symbolic link to a directory.
  **Expected**: The command deletes the symbolic link itself (not the target directory). `os.RemoveAll()` deletes symlinks without following them.

- **Condition**: User invokes `spectra clear` and the sessions directory contains a large session (e.g., 10 GB of data).
  **Expected**: The command deletes all files recursively. This may take time. The command does not provide progress updates. Deletion is synchronous.

## Related

- [SpectraFinder](../../storage/spectra_finder.md) - Locates the project root
- [StorageLayout](../../storage/storage_layout.md) - Provides path composition for sessions directory
- [Runtime](../../runtime/runtime.md) - Creates and manages sessions
- [SessionDirectoryManager](../../storage/session_directory_manager.md) - Creates session directories
- [init Subcommand](./init.md) - Initialize a Spectra project
- [run Subcommand](./run.md) - Run a workflow (creates sessions)
- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - System architecture overview
