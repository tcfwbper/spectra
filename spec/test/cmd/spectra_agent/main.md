# Test Specification: `main.go`

## Source File Under Test
`cmd/spectra_agent/main.go`

## Test File
`cmd/spectra_agent/main_test.go`

---

## `main`

### Happy Path — main

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMain_ExitZero` | `unit` | Process exits with code 0 when Execute returns 0. | Replace `os.Exit` with a capturing fake and stub `Execute` to return `0` | | Fake exit called once with code `0` |
| `TestMain_ExitNonZero` | `unit` | Process exits with the exact code returned by Execute. | Replace `os.Exit` with a capturing fake and stub `Execute` to return `1` | | Fake exit called once with code `1` |
| `TestMain_ExitCode2` | `unit` | Process propagates exit code 2 unchanged. | Replace `os.Exit` with a capturing fake and stub `Execute` to return `2` | | Fake exit called once with code `2` |
| `TestMain_ExitCode3` | `unit` | Process propagates exit code 3 unchanged. | Replace `os.Exit` with a capturing fake and stub `Execute` to return `3` | | Fake exit called once with code `3` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestMain_CallsExecuteExactlyOnce` | `unit` | main calls Execute exactly once per invocation. | Replace `os.Exit` with a no-op fake; stub `Execute` with a call counter returning `0` | | `Execute` call count equals `1` |
| `TestMain_NoOtherOsCalls` | `unit` | Source file only references os.Exit and no other os functions. | Parse `cmd/spectra_agent/main.go` with `go/ast`; collect all selector expressions on the `os` package identifier | | The only `os.X` reference found is `os.Exit` |
