# Test Specification: `exit_codes_test.go`

## Source File Under Test
`internal/cmdutil/exit_codes.go`

## Test File
`internal/cmdutil/exit_codes_test.go`

---

## Exit Code Constants

### Happy Path — Constants

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitSuccess_Value` | `unit` | ExitSuccess equals 0. | None. | | `ExitSuccess == 0` |
| `TestExitInvocationError_Value` | `unit` | ExitInvocationError equals 1. | None. | | `ExitInvocationError == 1` |
| `TestExitTransportError_Value` | `unit` | ExitTransportError equals 2. | None. | | `ExitTransportError == 2` |
| `TestExitRuntimeError_Value` | `unit` | ExitRuntimeError equals 3. | None. | | `ExitRuntimeError == 3` |

### Boundary Values — No Overlap

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitCodes_Unique` | `unit` | All exit code constants have distinct values. | None. | | `ExitSuccess`, `ExitInvocationError`, `ExitTransportError`, `ExitRuntimeError` are all distinct |
