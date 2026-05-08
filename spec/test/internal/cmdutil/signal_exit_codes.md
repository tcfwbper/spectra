# Test Specification: `signal_exit_codes_test.go`

## Source File Under Test
`internal/cmdutil/signal_exit_codes.go`

## Test File
`internal/cmdutil/signal_exit_codes_test.go`

---

## Signal Exit Code Constants

### Happy Path — Constants

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestExitSignalINT_Value` | `unit` | ExitSignalINT equals 130 (128 + 2). | None. | | `ExitSignalINT == 130` |
| `TestExitSignalTERM_Value` | `unit` | ExitSignalTERM equals 143 (128 + 15). | None. | | `ExitSignalTERM == 143` |

### Boundary Values — No Overlap With Base Exit Codes

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestSignalExitCodes_NoOverlapWithBaseCodes` | `unit` | Signal exit codes do not overlap with base exit codes (0, 1, 2, 3). | None. | | `ExitSignalINT` and `ExitSignalTERM` are not in `{0, 1, 2, 3}` |
| `TestSignalExitCodes_Unique` | `unit` | All signal exit code constants have distinct values. | None. | | `ExitSignalINT != ExitSignalTERM` |
