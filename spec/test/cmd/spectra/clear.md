# Test Specification: `clear.go`

## Source File Under Test
`cmd/spectra/clear.go`

## Test File
`cmd/spectra/clear_test.go`

---

## `ClearCommand`

### Happy Path — Delete Specific Session

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_DeleteSpecificSession` | `unit` | Deletes a specific session by UUID. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/12345678-1234-1234-1234-123456789abc/` directory with `session.json` and `events.jsonl` files created inside test fixture | `--session-id=12345678-1234-1234-1234-123456789abc` | Prints `"Session '12345678-1234-1234-1234-123456789abc' cleared successfully"`; session directory deleted; exit code 0 |
| `TestClearCommand_DeleteSessionWithNestedFiles` | `unit` | Deletes session directory containing subdirectories and files. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/test-session/` with nested subdirectories and multiple files created inside test fixture | `--session-id=test-session` | Prints success message; entire directory tree deleted recursively; exit code 0 |

### Happy Path — Delete All Sessions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_DeleteAllSessionsWithConfirmation` | `unit` | Deletes all sessions when user confirms with 'y'. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with three session subdirectories created inside test fixture; stdin mocked to provide `"y\n"` | No `--session-id` flag | Prints confirmation prompt; prints cleared message for each session; prints `"All sessions cleared successfully"`; all session directories deleted; exit code 0 |
| `TestClearCommand_DeleteAllSessionsUppercaseY` | `unit` | Accepts uppercase 'Y' as confirmation. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with two session subdirectories created inside test fixture; stdin mocked to provide `"Y\n"` | No `--session-id` flag | All sessions deleted; prints success messages; exit code 0 |
| `TestClearCommand_SkipsFilesInSessionsDirectory` | `unit` | Only deletes directories, skips regular files in sessions directory. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with two session subdirectories and one regular file `somefile.txt` created inside test fixture; stdin mocked to provide `"y\n"` | No `--session-id` flag | Session directories deleted; `somefile.txt` remains; prints success for session directories only; exit code 0 |

### Happy Path — No Sessions to Clear

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_NoSessions` | `unit` | Prints message when sessions directory is empty. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` directory created but empty inside test fixture | No `--session-id` flag | Prints `"No sessions to clear"`; no confirmation prompt; exit code 0 |

### Happy Path — User Cancels Operation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_CancelWithN` | `unit` | Cancels deletion when user enters 'n'. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with two session subdirectories created inside test fixture; stdin mocked to provide `"n\n"` | No `--session-id` flag | Prints confirmation prompt; prints `"Operation cancelled"`; no sessions deleted; exit code 0 |
| `TestClearCommand_CancelWithUppercaseN` | `unit` | Cancels deletion when user enters 'N'. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with two session subdirectories created inside test fixture; stdin mocked to provide `"N\n"` | No `--session-id` flag | Prints `"Operation cancelled"`; no sessions deleted; exit code 0 |
| `TestClearCommand_CancelWithEmptyInput` | `unit` | Treats empty input (just Enter) as cancellation. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with one session subdirectory created inside test fixture; stdin mocked to provide `"\n"` | No `--session-id` flag | Prints `"Operation cancelled"`; no sessions deleted; exit code 0 |
| `TestClearCommand_CancelWithInvalidInput` | `unit` | Treats any input other than 'y' or 'Y' as cancellation. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with one session subdirectory created inside test fixture; stdin mocked to provide `"yes\n"` | No `--session-id` flag | Prints `"Operation cancelled"`; no sessions deleted; exit code 0 |

### Validation Failures — Session Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_SessionNotFound` | `unit` | Prints warning when specified session does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` directory created but empty inside test fixture | `--session-id=nonexistent-session` | Prints `"Warning: session 'nonexistent-session' not found, skipping"`; exit code 0 |
| `TestClearCommand_InvalidUUIDFormat` | `unit` | Does not validate UUID format, attempts deletion. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` directory created inside test fixture | `--session-id=invalid-uuid` | Composes path with invalid UUID; prints warning if directory not found; exit code 0 |
| `TestClearCommand_EmptySessionID` | `unit` | Handles empty string as session ID. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` directory created inside test fixture | `--session-id=""` | Composes path `.spectra/sessions//`; prints `"Warning: session '' not found, skipping"`; exit code 0 |

### Validation Failures — Directory Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_SessionsDirectoryNotFound` | `unit` | Prints warning when sessions directory does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created but no `sessions/` subdirectory inside test fixture | No `--session-id` flag | Prints `"Warning: sessions directory not found, nothing to clear"`; exit code 0 |
| `TestClearCommand_SpectraNotFound` | `unit` | Returns error when .spectra directory not found. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created inside test fixture; test changes working directory to test fixture | No `--session-id` flag | Prints `"Error: .spectra directory not found. Are you in a Spectra project?"`; exit code 1 |

### Validation Failures — Permission Denied

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_SessionDeletionPermissionDenied` | `unit` | Returns error when session directory cannot be deleted. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/test-session/` directory created inside test fixture; parent directory permissions set to read-only (0555) within test fixture | `--session-id=test-session` | Prints `"Error: failed to clear session 'test-session': permission denied"`; exit code 1 |
| `TestClearCommand_SessionsDirectoryNotReadable` | `unit` | Returns error when sessions directory is not readable. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` directory created inside test fixture; permissions set to `0000` within test fixture | No `--session-id` flag | Prints error matching `/failed to read sessions directory:.*permission denied/i`; exit code 1 |

### Validation Failures — Partial Deletion Failure

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_PartialDeletionFailure` | `unit` | Reports error for failed session but continues with others. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with three session subdirectories created inside test fixture; second session directory permissions set to prevent deletion within test fixture; stdin mocked to provide `"y\n"` | No `--session-id` flag | Prints error for failed session; prints success for other sessions; prints `"All sessions cleared successfully"`; exit code 0 |

### Happy Path — Recursive Deletion

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_DeletesAllSessionContents` | `unit` | Recursively deletes all files and subdirectories. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/test-session/` with `session.json`, `events.jsonl`, nested subdirectories with files created inside test fixture | `--session-id=test-session` | Entire directory tree deleted; no files remain; prints success message; exit code 0 |

### Happy Path — Symbolic Link Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_DeletesSymlinkNotTarget` | `unit` | Deletes symbolic link without following to target. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/link-session/` created as symlink to external directory inside test fixture; target directory exists outside test fixture | `--session-id=link-session` | Symlink deleted; target directory remains intact; prints success message; exit code 0 |

### Happy Path — Sessions Directory Preserved

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_PreservesSessionsDirectory` | `unit` | Does not delete the sessions directory itself. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` with two session subdirectories created inside test fixture; stdin mocked to provide `"y\n"` | No `--session-id` flag | Session subdirectories deleted; `.spectra/sessions/` directory remains; exit code 0 |

### Happy Path — No Confirmation for Single Session

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_NoConfirmationForSingleSession` | `unit` | Does not prompt for confirmation when deleting specific session. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/test-session/` directory created inside test fixture | `--session-id=test-session` | No confirmation prompt; session deleted immediately; prints success message; exit code 0 |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_IdempotentDeletion` | `unit` | Repeated invocation after deletion prints warning, not error. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/test-session/` directory created inside test fixture | First call: `--session-id=test-session`, then second call: `--session-id=test-session` | First call deletes and prints success; second call prints warning; both exit with code 0 |

### Happy Path — Help Output

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_Help` | `unit` | Displays help information without invoking SpectraFinder. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created | `--help` | Prints usage information including flags and examples; exit code 0; no attempt to find `.spectra/` |

### Boundary Values — Large Session

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_LargeSessionDirectory` | `unit` | Deletes session with many files synchronously. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/large-session/` with 1000 small files created inside test fixture | `--session-id=large-session` | All files deleted; prints success message; deletion completes without timeout; exit code 0 |

### Integration — SpectraFinder

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_UsesSpectraFinder` | `unit` | Uses SpectraFinder to locate project root. | Temporary test directory created programmatically within test fixture; nested directory structure `root/.spectra/sessions/` and `root/subdir/` created inside test fixture; test changes working directory to `root/subdir/` | `--session-id=test-session` from `root/subdir/` | SpectraFinder searches upward; finds `.spectra/` in parent; operates on `root/.spectra/sessions/`; prints appropriate message; exit code 0 or 1 depending on session existence |

### Integration — StorageLayout

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClearCommand_UsesStorageLayout` | `unit` | Uses StorageLayout to compose sessions directory path. | Temporary test directory created programmatically within test fixture; `.spectra/sessions/` directory created inside test fixture | `--session-id=test-session` | Command uses StorageLayout to get sessions path; correct path composed; operates on correct directory |
