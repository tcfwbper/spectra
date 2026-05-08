# cmd/spectra main

## Overview

The process entry point for the `spectra` CLI binary. Its sole responsibility is to invoke the root command's `Execute()` function and terminate the process with the returned exit code. It contains no business logic, flag parsing, or I/O beyond calling `os.Exit`.

## Boundaries

- Owns: process entry (`func main`) and `os.Exit` invocation.
- Delegates: all CLI behavior (flag parsing, subcommand dispatch, output) to `internal/cmd/spectra.Execute`.
- Must not: parse arguments, read environment variables, perform I/O, or handle signals.
- Must not: contain any logic beyond the single `os.Exit(Execute())` call.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `internal/cmd/spectra` | Root command package | Call `Execute() int` | Must not call any other function from this package |
| `os` | Standard library | Call `os.Exit(code)` | Must not use any other `os` function |

Construction constraint: This is a `package main` with a single `func main()`. No constructors, factories, or adapters are involved.

## Behavior

1. Calls `internal/cmd/spectra.Execute()` which returns an integer exit code.
2. Passes the returned exit code to `os.Exit`.

## Inputs

None. All inputs (CLI arguments, environment) are consumed by the delegated root command.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| Process exit code | int | Propagated unchanged from `Execute()` |

## Invariants

1. **Single Statement**: The function body contains exactly one effective statement: `os.Exit(Execute())`.
2. **No Logic**: No conditionals, loops, error handling, or variable declarations beyond the call.
3. **Exit Code Passthrough**: The exit code from `Execute()` is never modified.

## Edge Cases

- Condition: `Execute()` returns 0.
  Expected: Process exits with code 0.

- Condition: `Execute()` returns non-zero (1, 130, 143).
  Expected: Process exits with that exact code.

## Related

- [internal/cmd/spectra root](../../internal/cmd/spectra/root.md) — the root command that owns all CLI behavior
