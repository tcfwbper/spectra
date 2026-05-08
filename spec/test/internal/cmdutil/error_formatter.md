# Test Specification: `error_formatter_test.go`

## Source File Under Test
`internal/cmdutil/error_formatter.go`

## Test File
`internal/cmdutil/error_formatter_test.go`

---

## `FormatError`

### Happy Path — FormatError

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFormatError_BasicMessage` | `unit` | Returns message prefixed with "Error: ". | None. | `msg="file not found"` | Returns `"Error: file not found"` |
| `TestFormatError_MessageWithSpaces` | `unit` | Preserves internal whitespace in message. | None. | `msg="could not connect to server"` | Returns `"Error: could not connect to server"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFormatError_EmptyString` | `unit` | Returns prefix only when message is empty. | None. | `msg=""` | Returns `"Error: "` |

### Boundary Values — msg

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFormatError_AlreadyPrefixed` | `unit` | Does not deduplicate existing "Error: " prefix in msg. | None. | `msg="Error: something"` | Returns `"Error: Error: something"` |

## `FormatWarning`

### Happy Path — FormatWarning

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFormatWarning_BasicMessage` | `unit` | Returns message prefixed with "Warning: ". | None. | `msg="deprecated flag"` | Returns `"Warning: deprecated flag"` |
| `TestFormatWarning_MessageWithSpaces` | `unit` | Preserves internal whitespace in message. | None. | `msg="config value is missing"` | Returns `"Warning: config value is missing"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFormatWarning_EmptyString` | `unit` | Returns prefix only when message is empty. | None. | `msg=""` | Returns `"Warning: "` |

### Boundary Values — msg

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestFormatWarning_AlreadyPrefixed` | `unit` | Does not deduplicate existing "Warning: " prefix in msg. | None. | `msg="Warning: something"` | Returns `"Warning: Warning: something"` |
