# Test Specification: `root.go`

## Source File Under Test
`cmd/spectra/root.go`

## Test File
`cmd/spectra/root_test.go`

---

## `RootCommand`

### Happy Path — Display Help

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_NoSubcommand` | `unit` | Displays usage information when invoked without subcommand. | None | No arguments | Prints usage information to stdout including available commands; exit code 0 |
| `TestRootCommand_HelpFlag` | `unit` | Displays usage information when invoked with --help. | None | `--help` | Prints usage information to stdout including available commands, flags, and usage examples; exit code 0 |

### Happy Path — Display Version

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_VersionFlag` | `unit` | Displays version string when invoked with --version. | None | `--version` | Prints `"spectra version <version-string>"` to stdout; exit code 0 |
| `TestRootCommand_VersionFormat` | `unit` | Version string follows semantic versioning format. | None | `--version` | Output matches pattern `/spectra version v\d+\.\d+\.\d+/`; exit code 0 |

### Happy Path — Subcommand Delegation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_InitSubcommand` | `unit` | Delegates to init subcommand successfully. | Mock init subcommand handler registered | `init` | Init subcommand handler invoked; no errors from root command |
| `TestRootCommand_RunSubcommand` | `unit` | Delegates to run subcommand successfully. | Mock run subcommand handler registered | `run` | Run subcommand handler invoked; no errors from root command |
| `TestRootCommand_ClearSubcommand` | `unit` | Delegates to clear subcommand successfully. | Mock clear subcommand handler registered | `clear` | Clear subcommand handler invoked; no errors from root command |

### Happy Path — Exit Code Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_PropagatesSubcommandSuccess` | `unit` | Propagates exit code 0 from successful subcommand. | Mock subcommand that returns exit code 0 | Subcommand invocation | Root command exits with code 0 |
| `TestRootCommand_PropagatesSubcommandError` | `unit` | Propagates exit code 1 from failed subcommand. | Mock subcommand that returns exit code 1 | Subcommand invocation | Root command exits with code 1 |

### Validation Failures — Unknown Subcommand

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_UnknownSubcommand` | `unit` | Returns error for unknown subcommand. | None | `unknown-command` | Prints `"Error: unknown command \"unknown-command\" for \"spectra\""` to stderr; exit code 1 |
| `TestRootCommand_MultipleUnknownSubcommands` | `unit` | Returns error for first unknown subcommand. | None | `foo bar` | Prints `"Error: unknown command \"foo\" for \"spectra\""` to stderr; exit code 1 |

### Validation Failures — Invalid Flags

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_InvalidGlobalFlag` | `unit` | Returns error for unknown global flag. | None | `--invalid-flag` | Prints error to stderr mentioning unknown flag; exit code 1 |
| `TestRootCommand_SubcommandInvalidFlag` | `unit` | Subcommand handles flag parsing, error propagated. | None | `init --invalid-flag` | Subcommand returns error for invalid flag; root command propagates error; exit code 1 |

### Stateless Execution

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_StatelessInvocations` | `unit` | Multiple invocations are independent with no shared state. | None | First invocation: `--version`, second invocation: `--help` | Each invocation produces correct output; no state leakage between invocations |

### Error Output Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_ErrorPrefix` | `unit` | Error messages are prefixed with "Error: ". | None | `unknown-command` | stderr output starts with `"Error: "`; exit code 1 |
| `TestRootCommand_ErrorToStderr` | `unit` | Error messages are printed to stderr, not stdout. | None | `unknown-command` | Error message appears on stderr; stdout is empty; exit code 1 |

### Usage Information Content

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_UsageIncludesAllSubcommands` | `unit` | Usage information lists all three subcommands. | None | `--help` | stdout contains "init", "run", and "clear" in Available Commands section |
| `TestRootCommand_UsageIncludesDescription` | `unit` | Usage information includes command description. | None | `--help` | stdout contains "Framework for defining and executing flexible AI agent workflows" or similar description |
| `TestRootCommand_UsageIncludesFlags` | `unit` | Usage information lists available flags. | None | `--help` | stdout contains "--help" and "--version" in Flags section |
| `TestRootCommand_UsageIncludesExamples` | `unit` | Usage information includes usage examples. | None | `--help` | stdout contains usage pattern `"spectra [command]"` and help text for subcommands |

### Happy Path — Cobra Framework

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_UsesCobra` | `unit` | Root command is implemented using Cobra library. | None | Code inspection or behavior verification | Command structure follows Cobra patterns; subcommands registered via Cobra API |

### Integration — Subcommand Flags

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_SubcommandWithFlags` | `unit` | Passes flags to subcommand correctly. | Mock clear subcommand handler that accepts --session-id flag | `clear --session-id=test-uuid` | Subcommand receives --session-id flag with value "test-uuid"; subcommand executes correctly |

### Boundary Values — Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_EmptyArguments` | `unit` | Treats empty arguments same as no arguments. | None | Empty args array | Prints usage information to stdout; exit code 0 |

### Happy Path — Help for Subcommands

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_SubcommandHelp` | `unit` | Shows help for specific subcommand. | None | `init --help` | Prints init subcommand help information; exit code 0 |
| `TestRootCommand_HelpSubcommandSyntax` | `unit` | Supports help command syntax. | None | `help init` | Prints init subcommand help information; exit code 0 |

### Concurrent Behaviour

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_ConcurrentInvocations` | `race` | Multiple concurrent invocations execute independently. | Test fixture with temporary directory created programmatically | 10 goroutines each invoke root command with different subcommands/flags | All invocations complete; no data races; correct outputs for each invocation |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_PropagatesSubcommandError` | `unit` | Propagates detailed error from subcommand. | Mock subcommand that returns error with specific message | Subcommand invocation | Root command includes subcommand error message in output; exit code 1 |
