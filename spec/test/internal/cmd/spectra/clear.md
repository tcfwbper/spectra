# Test Specification: `clear_test.go`

## Source File Under Test

`internal/cmd/spectra/clear.go`

## Test File

`internal/cmd/spectra/clear_test.go`

---

## `ClearCommand`

### Happy Path — Clear Specific Sessions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_SingleUUID_Exists` | `unit` | Deletes a single existing session directory when user confirms. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/<UUID>/` subdirectory. Stub `SpectraFinder.Find()` to return project root. Stub `StorageLayout.GetSessionDir()` to return session path. Provide `strings.NewReader("y\n")` as stdin for `ConfirmPrompt`. | positional args: `["<UUID>"]` | Session directory removed; stdout contains `"Session '<UUID>' cleared"` |
| `TestClear_MultipleUUIDs_AllExist` | `unit` | Deletes multiple existing session directories when user confirms. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/<UUID1>/` and `.spectra/sessions/<UUID2>/`. Stub `SpectraFinder.Find()` to return project root. Stub `StorageLayout.GetSessionDir()` for each UUID. Provide `strings.NewReader("y\n")` as stdin. | positional args: `["<UUID1>", "<UUID2>"]` | Both directories removed; stdout contains `"Session '<UUID1>' cleared"` and `"Session '<UUID2>' cleared"` |
| `TestClear_ConfirmationPrompt_ListsUUIDs` | `unit` | Confirmation prompt lists all provided UUIDs. | Create temp dir as project root using `t.TempDir()`. Stub `SpectraFinder.Find()` to return project root. Provide `strings.NewReader("n\n")` as stdin. | positional args: `["abc-123", "def-456"]` | Prompt output contains `"Are you sure you want to delete the following sessions?"`, `"- abc-123"`, `"- def-456"` |

### Happy Path — Clear All Sessions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_NoArgs_DeletesAll` | `unit` | Deletes all session subdirectories when user confirms. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/sess1/` and `.spectra/sessions/sess2/`. Stub `SpectraFinder.Find()` to return project root. Stub `StorageLayout.GetSessionsDir()` to return sessions dir. Provide `strings.NewReader("y\n")` as stdin. | no positional args | Both directories removed; stdout contains `"Session 'sess1' cleared"`, `"Session 'sess2' cleared"`, `"All sessions cleared successfully"` |
| `TestClear_NoArgs_SkipsFiles` | `unit` | Skips regular files in sessions directory, only deletes subdirectories. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/sess1/` (directory) and `.spectra/sessions/somefile.txt` (regular file). Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. Provide `strings.NewReader("y\n")` as stdin. | no positional args | `sess1/` removed; `somefile.txt` still exists; stdout contains `"Session 'sess1' cleared"` |
| `TestClear_NoArgs_PromptText` | `unit` | Shows correct prompt when deleting all sessions. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/sess1/`. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. Provide `strings.NewReader("n\n")` as stdin. | no positional args | Prompt output contains `"Are you sure you want to delete all sessions? [y/N]: "` |

### Happy Path — User Cancels

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_SpecificUUIDs_UserDeclinesN` | `unit` | No deletion when user enters "n" for specific UUIDs. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/<UUID>/`. Stub `SpectraFinder.Find()`. Provide `strings.NewReader("n\n")` as stdin. | positional args: `["<UUID>"]` | Session directory still exists; stdout contains `"Operation cancelled"` |
| `TestClear_AllSessions_UserDeclinesN` | `unit` | No deletion when user enters "n" for delete-all. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/sess1/`. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. Provide `strings.NewReader("n\n")` as stdin. | no positional args | Session directory still exists; stdout contains `"Operation cancelled"` |
| `TestClear_UserEntersYes_TreatedAsRejection` | `unit` | "yes" is treated as rejection (only single "y" accepted). | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/<UUID>/`. Stub `SpectraFinder.Find()`. Provide `strings.NewReader("yes\n")` as stdin. | positional args: `["<UUID>"]` | Session directory still exists; stdout contains `"Operation cancelled"` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_SpectraFinderFails` | `unit` | Exits with code 1 when SpectraFinder cannot find project root. | Stub `SpectraFinder.Find()` to return error. | any args | stderr contains `"Error: .spectra directory not found. Are you in a Spectra project?"`; exit code 1 |
| `TestClear_NoArgs_SessionsDirNotExist` | `unit` | Prints warning and exits 0 when sessions directory does not exist. | Create temp dir as project root using `t.TempDir()`. Do not create `.spectra/sessions/`. Stub `SpectraFinder.Find()` to return project root. Stub `StorageLayout.GetSessionsDir()` to return non-existent path. | no positional args | stdout contains `"Warning: sessions directory not found, nothing to clear"`; exit code 0 |
| `TestClear_NoArgs_ReadDirFails` | `unit` | Exits with code 1 when reading sessions directory fails. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/` with permissions `0000`. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. Cleanup: restore permissions in `t.Cleanup`. | no positional args | stderr contains `"Error: failed to read sessions directory:"`; exit code 1 |
| `TestClear_SpecificUUID_DeletionFails` | `unit` | Reports error for failed deletion and continues to next UUID. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/<UUID1>/` with read-only parent or use OS-level lock to prevent deletion. Create `.spectra/sessions/<UUID2>/` (deletable). Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionDir()`. Provide `strings.NewReader("y\n")` as stdin. Cleanup: restore permissions in `t.Cleanup`. | positional args: `["<UUID1>", "<UUID2>"]` | stderr contains `"Error: failed to clear session '<UUID1>':"` ; stdout contains `"Session '<UUID2>' cleared"`; exit code 0 |
| `TestClear_NoArgs_PartialDeletionFailure` | `unit` | Reports per-session errors without summary when some deletions fail in delete-all mode. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/sess1/` (undeletable via permission) and `.spectra/sessions/sess2/` (deletable). Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. Provide `strings.NewReader("y\n")` as stdin. Cleanup: restore permissions in `t.Cleanup`. | no positional args | stderr contains `"Error: failed to clear session 'sess1':"` ; stdout contains `"Session 'sess2' cleared"`; stdout does not contain `"All sessions cleared successfully"`; exit code 0 |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_NoArgs_EmptySessionsDir` | `unit` | Prints "No sessions to clear" when sessions directory exists but is empty. | Create temp dir as project root using `t.TempDir()`. Create empty `.spectra/sessions/` directory. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. | no positional args | stdout contains `"No sessions to clear"`; exit code 0; no confirmation prompt shown |
| `TestClear_NoArgs_OnlyFilesInSessionsDir` | `unit` | Prints "No sessions to clear" when sessions directory contains only regular files. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/` with only a regular file inside. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionsDir()`. | no positional args | stdout contains `"No sessions to clear"`; exit code 0 |
| `TestClear_SpecificUUID_NotFound` | `unit` | Prints warning for non-existent UUID after confirmation. | Create temp dir as project root using `t.TempDir()`. Stub `SpectraFinder.Find()`. Stub `StorageLayout.GetSessionDir()` to return a path that does not exist. Provide `strings.NewReader("y\n")` as stdin. | positional args: `["nonexistent-uuid"]` | stdout contains `"Warning: session 'nonexistent-uuid' not found, skipping"`; exit code 0 |
| `TestClear_EOF_Stdin` | `unit` | Treats EOF stdin as rejection. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/<UUID>/`. Stub `SpectraFinder.Find()`. Provide `strings.NewReader("")` (EOF) as stdin. | positional args: `["<UUID>"]` | Session directory still exists; stdout contains `"Operation cancelled"`; exit code 0 |

### Boundary Values — UUIDs

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_EmptyStringUUID` | `unit` | Empty string UUID is passed to StorageLayout without validation. | Create temp dir as project root using `t.TempDir()`. Stub `SpectraFinder.Find()`. Stub `StorageLayout.GetSessionDir()` to return a path based on empty string. Provide `strings.NewReader("y\n")` as stdin. | positional args: `[""]` | Warning printed (stat likely fails); exit code 0 |
| `TestClear_MixedExistentAndNonExistent` | `unit` | Deletes existing session and warns about missing session in single invocation. | Create temp dir as project root using `t.TempDir()`. Create `.spectra/sessions/exists-uuid/`. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionDir()`. Provide `strings.NewReader("y\n")` as stdin. | positional args: `["exists-uuid", "missing-uuid"]` | `exists-uuid/` removed; stdout contains `"Session 'exists-uuid' cleared"` and `"Warning: session 'missing-uuid' not found, skipping"`; exit code 0 |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestClear_CallsSpectraFinder` | `unit` | Calls SpectraFinder.Find() to locate project root. | Mock `SpectraFinder` with `Find()` returning project root. | any args | `SpectraFinder.Find()` called exactly once |
| `TestClear_CallsStorageLayoutGetSessionDir` | `unit` | Calls StorageLayout.GetSessionDir for each provided UUID. | Create temp dir. Stub `SpectraFinder.Find()`. Mock `StorageLayout`. Provide `strings.NewReader("y\n")` as stdin. | positional args: `["uuid1", "uuid2"]` | `StorageLayout.GetSessionDir(projectRoot, "uuid1")` and `StorageLayout.GetSessionDir(projectRoot, "uuid2")` each called once |
| `TestClear_CallsStorageLayoutGetSessionsDir` | `unit` | Calls StorageLayout.GetSessionsDir when no args provided. | Create temp dir. Stub `SpectraFinder.Find()`. Mock `StorageLayout`. Create sessions dir with subdirectory. Provide `strings.NewReader("y\n")` as stdin. | no positional args | `StorageLayout.GetSessionsDir(projectRoot)` called exactly once |
| `TestClear_CallsConfirmPrompt` | `unit` | Calls ConfirmPrompt before any deletion. | Create temp dir. Create `.spectra/sessions/<UUID>/`. Stub `SpectraFinder.Find()` and `StorageLayout.GetSessionDir()`. Mock `ConfirmPrompt` to return false. | positional args: `["<UUID>"]` | `ConfirmPrompt` called once; no directories deleted |
