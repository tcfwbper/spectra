# Test Specification: `init.go`

## Source File Under Test
`cmd/spectra/init.go`

## Test File
`cmd/spectra/init_test.go`

---

## `init` Command

### Happy Path — Initialization

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_FreshDirectory` | `e2e` | Initializes all directories and files in a fresh directory. | Create empty temp directory; embed all built-in workflow, agent, and spec files | Run `spectra init` in temp directory | Exit code 0; `.gitignore` created with `.spectra` entry; all `.spectra/` directories created; all `.spectra/` files written; all `spec/` directories created; all `spec/` files written; stdout contains `"Spectra project initialized successfully"` |
| `TestInit_GitignoreCreated` | `e2e` | Creates `.gitignore` with `.spectra` entry when it does not exist. | Create empty temp directory without `.gitignore`; embed built-in files | Run `spectra init` in temp directory | `.gitignore` exists with content `.spectra\n`; no warning or success message for `.gitignore` creation |
| `TestInit_GitignoreAlreadyContainsEntry` | `e2e` | Skips modifying `.gitignore` when it already contains `.spectra` entry. | Create temp directory with `.gitignore` containing `.spectra` entry | Run `spectra init` in temp directory | `.gitignore` content unchanged; no warning printed; initialization proceeds normally |
| `TestInit_GitignoreAppended` | `e2e` | Appends `.spectra` to existing `.gitignore` that does not contain the entry. | Create temp directory with `.gitignore` containing other entries but not `.spectra` | Run `spectra init` in temp directory | `.gitignore` contains appended `.spectra` entry; original content preserved; initialization proceeds normally |
| `TestInit_AllDirectoriesExist` | `e2e` | Skips directory creation when all directories already exist. | Create temp directory with all `.spectra/` and `spec/` directories pre-created | Run `spectra init` in temp directory | Exit code 0; no errors printed; all files written; stdout contains success message |

### Happy Path — Partial State

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_SomeDirectoriesExist` | `e2e` | Creates missing directories when some already exist. | Create temp directory with `.spectra/` and `.spectra/workflows/` but not `.spectra/agents/` or `.spectra/sessions/` | Run `spectra init` in temp directory | Exit code 0; missing directories created; existing directories unchanged; all files written |
| `TestInit_SomeBuiltinFilesExist` | `e2e` | Copies missing files and prints warnings for existing files. | Create temp directory with structure; create `.spectra/workflows/SimpleSdd.yaml` and `spec/ARCHITECTURE.md` | Run `spectra init` in temp directory | Exit code 0; warnings printed: `"Warning: workflow definition 'SimpleSdd.yaml' already exists, skipping"` and `"Warning: spec file 'ARCHITECTURE.md' already exists, skipping"`; other files written; success message printed |
| `TestInit_AllBuiltinFilesExist` | `e2e` | Prints warnings for all skipped files when all exist. | Create temp directory with complete structure and all built-in files | Run `spectra init` in temp directory | Exit code 0; warnings printed for each skipped file; no errors; success message printed |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_RepeatedInvocation` | `e2e` | Second invocation is idempotent with warnings for existing files. | Create temp directory; run `spectra init` once successfully | Run `spectra init` again in same directory | Exit code 0; warnings printed for all existing files; no errors; no files overwritten; success message printed |

### State Transitions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_PhasesExecuteInOrder` | `e2e` | Verifies phases execute in order: gitignore, .spectra dirs, .spectra files, spec dirs, spec files. | Create empty temp directory; embed built-in files | Run `spectra init` in temp directory | `.gitignore` created first; `.spectra/` directories exist before `.spectra/` files written; `spec/` directories exist before `spec/` files written; all operations succeed |

### Error Propagation — Phase 0 (.gitignore)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_GitignoreReadFails` | `e2e` | Returns error when reading `.gitignore` fails. | Create temp directory with `.gitignore` file with permissions `0000` (no read access) | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to read '.gitignore': permission denied"`; no `.spectra/` or `spec/` directories created |
| `TestInit_GitignoreWriteFails` | `e2e` | Returns error when updating `.gitignore` fails. | Create temp directory; make directory read-only after creating `.gitignore` without `.spectra` entry | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to update '.gitignore': permission denied"`; no `.spectra/` or `spec/` directories created |
| `TestInit_GitignoreBrokenSymlink` | `e2e` | Returns error when `.gitignore` is a broken symlink. | Create temp directory; create broken symlink `.gitignore` -> `nonexistent` | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to read '.gitignore': no such file or directory"`; no subsequent operations performed |

### Error Propagation — Phase 1 (.spectra directories)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_SpectraDirCreationFails` | `e2e` | Returns error when creating `.spectra/` directory fails. | Create temp directory; make directory read-only | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to create directory '.spectra': permission denied"`; no files written; `.gitignore` remains modified (if modified in Phase 0) |
| `TestInit_SpectraSubdirCreationFails` | `e2e` | Returns error when creating `.spectra/sessions/` fails after `.spectra/` succeeds. | Create temp directory; allow `.spectra/` creation but make it read-only before subdirectory creation | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to create directory '.spectra/sessions': permission denied"`; `.spectra/` exists; no `.spectra/` files written |
| `TestInit_SpectraExistsAsFile` | `e2e` | Returns error when `.spectra` exists as a regular file. | Create temp directory; create regular file `.spectra` | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to create directory '.spectra': file exists"` or similar; no subsequent operations performed |

### Error Propagation — Phase 2 (.spectra files)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_WorkflowFileWriteFails` | `e2e` | Returns error when writing a workflow file fails. | Create temp directory with structure; make `.spectra/workflows/` read-only | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to write built-in file '.spectra/workflows/SimpleSdd.yaml': permission denied"`; `.spectra/` directories exist; no `spec/` directories created |
| `TestInit_AgentFileWriteFails` | `e2e` | Returns error when writing an agent file fails. | Create temp directory with structure; make `.spectra/agents/` read-only | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to write built-in file"` and mentions agent file path; workflow files may exist (if written before error) |
| `TestInit_WorkflowFileWriteFailsDiskFull` | `e2e` | Returns error when writing fails due to disk full. | Create temp directory with structure; simulate disk full condition using platform-specific approach (Linux: use small loop device or quota; macOS: use disk image; Windows: skip test with message "disk quota simulation not available"); skip if simulation not feasible | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to write built-in file"` and `"no space left on device"` or similar; partial state remains; or test skipped on unsupported platforms |

### Error Propagation — Phase 3 (spec directories)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_SpecDirCreationFails` | `e2e` | Returns error when creating `spec/` directory fails. | Create temp directory; make directory read-only after `.spectra/` files are written | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to create directory 'spec': permission denied"`; all `.spectra/` directories and files exist |
| `TestInit_SpecSubdirCreationFails` | `e2e` | Returns error when creating `spec/logic/` fails after `spec/` succeeds. | Create temp directory; allow `spec/` creation but make it read-only before subdirectory creation | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to create directory 'spec/logic': permission denied"`; `spec/` exists; all `.spectra/` directories and files exist |
| `TestInit_SpecExistsAsFile` | `e2e` | Returns error when `spec` exists as a regular file. | Create temp directory with complete `.spectra/` structure; create regular file `spec` | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to create directory 'spec': file exists"` or similar; all `.spectra/` directories and files exist |

### Error Propagation — Phase 4 (spec files)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_SpecFileWriteFails` | `e2e` | Returns error when writing a spec file fails. | Create temp directory with all directories; make `spec/` read-only | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to write built-in file"` and mentions spec file path; all directories exist; all `.spectra/` files exist |
| `TestInit_SpecFileWriteFailsNestedDir` | `e2e` | Returns error when writing nested spec file fails due to missing subdirectory. | Create temp directory with `.spectra/` structure complete and `spec/` directory; manually delete `spec/logic/` after Phase 3 completes (mock/test double) | Run `spectra init` in temp directory | Exit code 1; stderr contains `"Error: failed to write built-in file"` and `"no such file or directory"` for `spec/logic/README.md` |

### Validation Failures — .gitignore

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_GitignoreContainsVariation` | `e2e` | Appends `.spectra` when `.gitignore` contains `.spectra/` but not `.spectra`. | Create temp directory with `.gitignore` containing `.spectra/` entry | Run `spectra init` in temp directory | `.gitignore` contains both `.spectra/` (original) and `.spectra` (appended); initialization succeeds |
| `TestInit_GitignoreContainsCommented` | `e2e` | Appends `.spectra` when `.gitignore` contains only commented `# .spectra` entry. | Create temp directory with `.gitignore` containing `# .spectra` | Run `spectra init` in temp directory | `.gitignore` contains both `# .spectra` (original) and `.spectra` (appended); initialization succeeds |
| `TestInit_GitignoreWhitespaceMatch` | `e2e` | Skips modification when `.gitignore` line contains `.spectra` with leading/trailing spaces or tabs. | Create temp directory with `.gitignore` containing `  .spectra  ` (spaces) or `\t.spectra\t` (tabs) | Run `spectra init` in temp directory | `.gitignore` unchanged; initialization proceeds normally |
| `TestInit_GitignoreNonBreakingSpace` | `e2e` | Appends `.spectra` when `.gitignore` line contains non-breaking space around `.spectra`. | Create temp directory with `.gitignore` containing ` .spectra` (non-breaking space U+00A0) | Run `spectra init` in temp directory | `.gitignore` contains original line and appended `.spectra` (non-breaking space not trimmed) |
| `TestInit_GitignoreSymlinkFollowed` | `e2e` | Follows symlink and modifies target file when `.gitignore` is a symlink. | Create temp directory; create file `shared-gitignore`; create symlink `.gitignore` -> `shared-gitignore` | Run `spectra init` in temp directory | `shared-gitignore` modified to include `.spectra`; initialization succeeds |
| `TestInit_GitignoreNoTrailingNewline` | `e2e` | Appends `.spectra` with proper newline when `.gitignore` does not end with newline. | Create temp directory with `.gitignore` containing `*.log` without trailing newline | Run `spectra init` in temp directory | `.gitignore` contains `*.log\n.spectra\n` (newline added before `.spectra`) |
| `TestInit_GitignoreWithTrailingNewline` | `e2e` | Appends `.spectra` when `.gitignore` ends with newline. | Create temp directory with `.gitignore` containing `*.log\n` | Run `spectra init` in temp directory | `.gitignore` contains `*.log\n.spectra\n` |

### Boundary Values — Directory Names

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_CurrentDirIsRoot` | `e2e` | Handles initialization when current directory is filesystem root. | Create temp directory with subdirectory `mock-root/` with permissions `0555` to simulate root-like restrictions; run test within fixture only (never use actual system root) | Run `spectra init` in mock-root directory | Exit code 1; stderr contains permission denied error |
| `TestInit_CurrentDirIsReadOnly` | `e2e` | Returns error when current directory is read-only. | Create temp directory; set permissions to `0555` (read-only); skip test if platform does not support permission restrictions (e.g., Windows without admin privileges) | Run `spectra init` in temp directory | Exit code 1; stderr contains permission denied error; or test skipped with message "read-only permissions not supported on this platform" |

### Boundary Values — Nested Project

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_NestedProject` | `e2e` | Creates nested Spectra project in subdirectory of existing project. | Create temp directory (fixture root); run `spectra init` in fixture root; create subdirectory `nested/` within fixture root; change working directory to `nested/` subdirectory | Run `spectra init` in `nested/` subdirectory | Exit code 0; `nested/.spectra/` and `nested/spec/` created; parent `.spectra/` unchanged; success message printed |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_NoArguments` | `e2e` | Runs successfully with no command-line arguments. | Create empty temp directory | Run `spectra init` (no additional arguments) | Exit code 0; initialization proceeds normally in current directory |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_UsesBuiltinResourceCopier` | `unit` | Verifies init command uses BuiltinResourceCopier for copying files. | Mock BuiltinResourceCopier; create temp directory | Run init command logic (unit test) | `CopyWorkflows`, `CopyAgents`, and `CopySpecFiles` methods called with correct project root |
| `TestInit_UsesEmbedFS` | `unit` | Verifies init command uses embed.FS for built-in files. | Inspect embedded filesystems; create temp directory | Run init command logic (unit test) | Embedded filesystems `builtinWorkflows`, `builtinAgents`, and `builtinSpecFiles` are populated from `//go:embed` directives |

### Resource Cleanup

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_DirectoryPermissions` | `e2e` | Verifies created directories have correct permissions. | Create empty temp directory | Run `spectra init` in temp directory | All created directories have permissions `0755` (rwxr-xr-x) |
| `TestInit_FilePermissions` | `e2e` | Verifies created files have correct permissions. | Create empty temp directory | Run `spectra init` in temp directory | `.gitignore` and all created files have permissions `0644` (rw-r--r--) |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_BuiltinFilesNotValidated` | `e2e` | Copies built-in files without validating YAML or Markdown syntax. | Embed built-in workflow file with invalid YAML syntax (mock/test double) | Run `spectra init` in temp directory | Exit code 0; invalid YAML file written to disk; no validation error; success message printed |
| `TestInit_BuiltinSpecFilesNotValidated` | `e2e` | Copies built-in spec files without validating Markdown syntax. | Embed built-in spec file with malformed Markdown (mock/test double) | Run `spectra init` in temp directory | Exit code 0; malformed Markdown file written to disk; no validation error; success message printed |

### Ordering — Phase Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_PhaseOrderingGitignoreFirst` | `e2e` | Verifies `.gitignore` modification happens before directory creation. | Create temp directory; instrument or log operations | Run `spectra init` in temp directory | `.gitignore` created/modified before any `.spectra/` directories created |
| `TestInit_PhaseOrderingSpectraDirsBeforeFiles` | `e2e` | Verifies `.spectra/` directories created before `.spectra/` files written. | Create temp directory; instrument or log operations | Run `spectra init` in temp directory | All `.spectra/` directories exist before any `.spectra/` files written |
| `TestInit_PhaseOrderingSpecDirsBeforeFiles` | `e2e` | Verifies `spec/` directories created before `spec/` files written. | Create temp directory; instrument or log operations | Run `spectra init` in temp directory | All `spec/` directories exist before any `spec/` files written |
| `TestInit_PhaseOrderingSpectraDirsBeforeSpecDirs` | `e2e` | Verifies `.spectra/` directories and files created before `spec/` directories. | Create temp directory; instrument or log operations | Run `spectra init` in temp directory | All `.spectra/` directories and files exist before `spec/` directory created |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestInit_ConcurrentInvocations` | `race` | Handles concurrent invocations in same directory (race condition). | Create temp directory | Run two `spectra init` commands concurrently in same directory | Both commands may succeed or one may fail with file exists error; no data corruption; filesystem remains consistent; completes within 5 seconds (no deadlock) |
