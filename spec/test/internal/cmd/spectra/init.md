# Test Specification: `init_test.go`

## Source File Under Test

`internal/cmd/spectra/init.go`

## Test File

`internal/cmd/spectra/init_test.go`

---

## `InitCommand`

### Happy Path — Init

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_AllPhasesSucceed` | `unit` | Prints success message when all phases complete without error. | Mock `GitignoreEnsurer.Ensure()` to return nil. Mock `DirectoryCreator.CreateAll()` to return nil. Mock `BuiltinResourceCopier.CopyWorkflows()`, `CopyAgents()`, `CopySpecFiles()` to return nil warnings and nil error. Stub `os.Getwd()` to return a temp dir path. | (no flags) | stdout contains `"Spectra project initialized successfully"`; exit code 0 |
| `TestInit_Phase2_WarningsPrinted` | `unit` | Prints warnings returned by BuiltinResourceCopier to stdout. | Mock `GitignoreEnsurer.Ensure()` to return nil. Mock `DirectoryCreator.CreateAll()` to return nil. Mock `BuiltinResourceCopier.CopyWorkflows()` to return warnings `["workflow X already exists, skipping"]` and nil error. Mock `CopyAgents()` and `CopySpecFiles()` to return nil. Stub `os.Getwd()`. | (no flags) | stdout contains `"workflow X already exists, skipping"` and `"Spectra project initialized successfully"`; exit code 0 |

### Ordering — Phase Sequencing

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_PhasesExecuteInOrder` | `unit` | Phases execute in order: gitignore, directories, files. | Create a shared call-order slice. Mock `GitignoreEnsurer.Ensure()` to append "gitignore" to slice. Mock `DirectoryCreator.CreateAll()` to append "directories". Mock `BuiltinResourceCopier.CopyWorkflows()` to append "workflows". Mock `CopyAgents()` to append "agents". Mock `CopySpecFiles()` to append "specfiles". Stub `os.Getwd()`. | (no flags) | Call-order slice is `["gitignore", "directories", "workflows", "agents", "specfiles"]` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_GetwdFails` | `unit` | Exits with code 1 when os.Getwd() fails. | Stub or inject `os.Getwd()` to return error. | (no flags) | stderr contains `"Error: failed to determine working directory:"`; exit code 1 |
| `TestInit_Phase0_GitignoreFails` | `unit` | Exits with code 1 when GitignoreEnsurer fails. | Mock `GitignoreEnsurer.Ensure()` to return error `"permission denied"`. Stub `os.Getwd()`. | (no flags) | stderr contains error message; exit code 1; `DirectoryCreator.CreateAll()` not called |
| `TestInit_Phase1_DirectoryCreatorFails` | `unit` | Exits with code 1 when DirectoryCreator fails. | Mock `GitignoreEnsurer.Ensure()` to return nil. Mock `DirectoryCreator.CreateAll()` to return error. Stub `os.Getwd()`. | (no flags) | stderr contains error message; exit code 1; `BuiltinResourceCopier` methods not called |
| `TestInit_Phase2a_CopyWorkflowsFails` | `unit` | Exits with code 1 when CopyWorkflows fails. | Mock `GitignoreEnsurer.Ensure()` to return nil. Mock `DirectoryCreator.CreateAll()` to return nil. Mock `BuiltinResourceCopier.CopyWorkflows()` to return error. Stub `os.Getwd()`. | (no flags) | stderr contains error message; exit code 1; `CopyAgents()` and `CopySpecFiles()` not called |
| `TestInit_Phase2b_CopyAgentsFails` | `unit` | Exits with code 1 when CopyAgents fails. | Mock `GitignoreEnsurer.Ensure()` to return nil. Mock `DirectoryCreator.CreateAll()` to return nil. Mock `CopyWorkflows()` to return nil. Mock `CopyAgents()` to return error. Stub `os.Getwd()`. | (no flags) | stderr contains error message; exit code 1; `CopySpecFiles()` not called |
| `TestInit_Phase2c_CopySpecFilesFails` | `unit` | Exits with code 1 when CopySpecFiles fails. | Mock `GitignoreEnsurer.Ensure()` to return nil. Mock `DirectoryCreator.CreateAll()` to return nil. Mock `CopyWorkflows()` and `CopyAgents()` to return nil. Mock `CopySpecFiles()` to return error. Stub `os.Getwd()`. | (no flags) | stderr contains error message; exit code 1 |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_PassesProjectRootToAllPhases` | `unit` | All phase functions receive the CWD as projectRoot argument. | Stub `os.Getwd()` to return `/fake/project`. Mock all phase collaborators to record received `projectRoot`. | (no flags) | `GitignoreEnsurer.Ensure("/fake/project")`, `DirectoryCreator.CreateAll("/fake/project")`, `CopyWorkflows("/fake/project")`, `CopyAgents("/fake/project")`, `CopySpecFiles("/fake/project")` each called with `/fake/project` |
| `TestInit_FailFast_Phase0_SkipsSubsequent` | `unit` | No subsequent phase is called after Phase 0 failure. | Mock `GitignoreEnsurer.Ensure()` to return error. Mock all other collaborators. Stub `os.Getwd()`. | (no flags) | `DirectoryCreator.CreateAll()` never called; `BuiltinResourceCopier` methods never called |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_ReInitialization_WarningsOnly` | `unit` | Re-initialization succeeds with warnings when all resources already exist. | Mock `GitignoreEnsurer.Ensure()` to return nil (idempotent). Mock `DirectoryCreator.CreateAll()` to return nil (idempotent). Mock `BuiltinResourceCopier.CopyWorkflows()`, `CopyAgents()`, `CopySpecFiles()` to each return warnings and nil error. Stub `os.Getwd()`. | (no flags) | stdout contains all warnings and `"Spectra project initialized successfully"`; exit code 0 |
