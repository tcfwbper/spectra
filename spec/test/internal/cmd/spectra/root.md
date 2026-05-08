# Test Specification: `root_test.go`

## Source File Under Test

`internal/cmd/spectra/root.go`

## Test File

`internal/cmd/spectra/root_test.go`

---

## `RootCommand`

### Happy Path — Execute

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRoot_NoArgs_PrintsUsage` | `unit` | Prints usage information when invoked without arguments. | Construct root command with subcommands registered. | No arguments | stdout contains `"spectra [command]"`; exit code 0 |
| `TestRoot_Help_PrintsUsage` | `unit` | Prints usage information when invoked with --help. | Construct root command with subcommands registered. | `--help` | stdout contains `"spectra [command]"`; exit code 0 |
| `TestRoot_Version_PrintsVersion` | `unit` | Prints version string when invoked with --version. | Construct root command with subcommands registered. | `--version` | stdout contains `"spectra version v0.1.0"`; exit code 0 |

### Happy Path — Subcommand Registration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRoot_RegistersInitSubcommand` | `unit` | Root command has init subcommand registered. | Construct root command. | N/A | Root command's subcommands include `"init"` |
| `TestRoot_RegistersRunSubcommand` | `unit` | Root command has run subcommand registered. | Construct root command. | N/A | Root command's subcommands include `"run"` |
| `TestRoot_RegistersClearSubcommand` | `unit` | Root command has clear subcommand registered. | Construct root command. | N/A | Root command's subcommands include `"clear"` |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRoot_UnknownSubcommand_ExitsWithError` | `unit` | Prints error and exits with code 1 for unknown subcommand. | Construct root command with subcommands registered. | `"foo"` | stderr contains `"Error: unknown command \"foo\" for \"spectra\""`; exit code 1 |

### Happy Path — Exit Code Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRoot_PropagatesSubcommandExitCode0` | `unit` | Propagates exit code 0 from a successful subcommand. | Register a stub subcommand that returns exit code 0. | Invoke stub subcommand | `Execute()` returns 0 |
| `TestRoot_PropagatesSubcommandExitCode1` | `unit` | Propagates exit code 1 from a failed subcommand. | Register a stub subcommand that returns exit code 1. | Invoke stub subcommand | `Execute()` returns 1 |
| `TestRoot_PropagatesExitCode130` | `unit` | Propagates exit code 130 unchanged from subcommand. | Register a stub subcommand that returns exit code 130. | Invoke stub subcommand | `Execute()` returns 130 |
| `TestRoot_PropagatesExitCode143` | `unit` | Propagates exit code 143 unchanged from subcommand. | Register a stub subcommand that returns exit code 143. | Invoke stub subcommand | `Execute()` returns 143 |
