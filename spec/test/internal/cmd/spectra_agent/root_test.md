# Test Specification: `root_test.go`

## Source File Under Test
`internal/cmd/spectra_agent/root.go`

## Test File
`internal/cmd/spectra_agent/root_test.go`

---

## `Execute`

### Happy Path — Execute

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExecute_NoSubcommandPrintsUsage` | `unit` | Prints usage and exits 0 when invoked without a subcommand. | Create a temp directory with `.spectra/` subdirectory as project root. Stub `storage.SpectraFinder.FindProjectRoot` to return that directory. | args: `["--session-id", "abc123"]` | Exit code 0; stdout contains `"spectra-agent [command]"` |
| `TestExecute_HelpFlag` | `unit` | Prints usage and exits 0 when --help is passed. | Create a temp directory with `.spectra/` subdirectory. Stub `SpectraFinder.FindProjectRoot` to return that directory. | args: `["--help"]` | Exit code 0; stdout contains usage text |

### Validation Failures — session-id

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExecute_MissingSessionID` | `unit` | Returns exit code 1 when --session-id is not provided. | Stub `SpectraFinder.FindProjectRoot` to return a valid directory. | args: `["error", "msg"]` (no `--session-id`) | Exit code 1; stderr contains `"--session-id flag is required"` |
| `TestExecute_EmptySessionID` | `unit` | Returns exit code 1 when --session-id is an empty string. | Stub `SpectraFinder.FindProjectRoot` to return a valid directory. | args: `["--session-id", "", "error", "msg"]` | Exit code 1; stderr contains `"--session-id flag is required"` |
| `TestExecute_InvalidUUIDSessionIDAccepted` | `unit` | Accepts a non-UUID session-id without error. | Create a temp directory with `.spectra/` subdirectory. Stub `SpectraFinder.FindProjectRoot` to return that directory. | args: `["--session-id", "not-a-uuid", "--help"]` | Exit code 0; no error about session-id format |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExecute_ProjectRootNotFound` | `unit` | Returns exit code 1 when SpectraFinder returns ErrNotInitialized. | Stub `SpectraFinder.FindProjectRoot` to return `ErrNotInitialized`. | args: `["--session-id", "abc", "error", "msg"]` | Exit code 1; stderr contains `".spectra directory not found. Are you in a Spectra project?"` |
| `TestExecute_UnknownSubcommand` | `unit` | Returns exit code 1 when an unknown subcommand is given. | Create a temp directory with `.spectra/` subdirectory. Stub `SpectraFinder.FindProjectRoot` to return that directory. | args: `["--session-id", "abc", "unknown"]` | Exit code 1; stderr contains unknown command error |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExecute_PropagatesSubcommandExitCode2` | `unit` | Propagates exit code 2 from subcommand unchanged. | Stub `SpectraFinder.FindProjectRoot` to return a valid project root. Mock `cmdutil.SendAndHandle` to return exit code 2. | args: `["--session-id", "abc", "error", "some error"]` | Exit code 2 |
| `TestExecute_PropagatesSubcommandExitCode3` | `unit` | Propagates exit code 3 from subcommand unchanged. | Stub `SpectraFinder.FindProjectRoot` to return a valid project root. Mock `cmdutil.SendAndHandle` to return exit code 3. | args: `["--session-id", "abc", "error", "some error"]` | Exit code 3 |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExecute_CallsSpectraFinderWithEmptyString` | `unit` | Calls SpectraFinder.FindProjectRoot with empty string (use CWD). | Mock `SpectraFinder.FindProjectRoot` to capture arguments and return a valid path. | args: `["--session-id", "abc", "--help"]` | `FindProjectRoot` called with `startDir=""` |
| `TestExecute_PropagatesSessionIDToSubcommand` | `unit` | Makes sessionID available to subcommand context. | Stub `SpectraFinder.FindProjectRoot` to return `/tmp/project`. Mock `cmdutil.SendAndHandle` to capture `sessionID` argument and return exit code 0. | args: `["--session-id", "my-sess-123", "error", "msg"]` | `SendAndHandle` called with `sessionID="my-sess-123"` |
| `TestExecute_PropagatesProjectRootToSubcommand` | `unit` | Makes projectRoot available to subcommand context. | Stub `SpectraFinder.FindProjectRoot` to return `/tmp/my-project`. Mock `cmdutil.SendAndHandle` to capture `projectRoot` argument and return exit code 0. | args: `["--session-id", "abc", "error", "msg"]` | `SendAndHandle` called with `projectRoot="/tmp/my-project"` |
