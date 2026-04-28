# Test Specification: `root.go`

## Source File Under Test
`cmd/spectra_agent/root.go`

## Test File
`cmd/spectra_agent/root_test.go`

---

## `RootCommand`

### Happy Path — Subcommand Dispatch

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_DispatchToEventEmit` | `unit` | Successfully dispatches to event emit subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock event emit handler registered | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Root command initializes successfully; dispatches to event emit handler; returns exit code from handler |
| `TestRootCommand_DispatchToError` | `unit` | Successfully dispatches to error subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock error handler registered | `args=["error", "test message", "--session-id", "<uuid>"]` | Root command initializes successfully; dispatches to error handler; returns exit code from handler |

### Happy Path — SpectraFinder Integration

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_FindsProjectRoot` | `unit` | Uses SpectraFinder to locate project root from current directory. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | SpectraFinder called with current working directory; project root identified; subcommand executed |
| `TestRootCommand_FindsProjectRootFromSubdir` | `unit` | Locates project root when invoked from subdirectory. | Temporary test directory created programmatically within test fixture; `.spectra/` and `subdir/nested/` directories created inside test fixture; test changes working directory to `<test-fixture>/subdir/nested/` | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | SpectraFinder traverses upward; finds `.spectra/` at test fixture root; returns correct project root |

### Happy Path — Usage Information

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_NoSubcommand` | `unit` | Prints usage information when invoked without subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["--session-id", "<uuid>"]` (no subcommand) | Prints usage information to stdout; exit code `0`; output contains `/spectra-agent - Interact with the Spectra workflow runtime/` |
| `TestRootCommand_HelpFlag` | `unit` | Prints usage information when invoked with --help. | Temporary test directory created programmatically within test fixture | `args=["--help"]` | Prints usage information to stdout; exit code `0`; output contains available commands and flags |
| `TestRootCommand_SubcommandHelp` | `unit` | Prints subcommand-specific help. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "--help"]` | Prints event subcommand help to stdout; exit code `0` |

### Validation Failures — Missing Required Flag

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_MissingSessionID` | `unit` | Returns exit code 1 when --session-id flag is missing. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent"]` (no --session-id) | Returns exit code `1`, stderr matches `/Error: --session-id flag is required/` |
| `TestRootCommand_EmptySessionID` | `unit` | Returns exit code 1 when --session-id flag is empty string. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", ""]` | Returns exit code `1`, stderr matches `/Error: --session-id flag is required/` |

### Validation Failures — Unknown Subcommand

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_UnknownSubcommand` | `unit` | Returns exit code 1 for unknown subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["foo", "--session-id", "<uuid>"]` | Returns exit code `1`, stderr matches `/Error: unknown command "foo" for "spectra-agent"/` |
| `TestRootCommand_UnknownNestedSubcommand` | `unit` | Returns exit code 1 for unknown nested subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "unknown", "--session-id", "<uuid>"]` | Returns exit code `1`, stderr indicates unknown subcommand under "event" |

### Validation Failures — Project Root Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_SpectraNotFound` | `unit` | Returns exit code 1 when .spectra directory not found in any ancestor. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Returns exit code `1`, stderr matches `/Error: \.spectra directory not found\. Are you in a Spectra project\?/` |
| `TestRootCommand_SpectraNotFoundFromRoot` | `unit` | Returns exit code 1 when simulating search from filesystem root with no .spectra. | Temporary test directory created programmatically within test fixture simulating filesystem root behavior; mock SpectraFinder configured to simulate traversal starting from root; no `.spectra/` directory in simulated hierarchy | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Returns exit code `1`, stderr matches `/Error: \.spectra directory not found\. Are you in a Spectra project\?/` |

### Validation Failures — Invalid Session ID Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_InvalidUUIDFormat` | `unit` | Accepts invalid UUID format and passes to subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "not-a-uuid"]` | Root command does not validate UUID; dispatches to subcommand with invalid value; subcommand or SocketClient returns appropriate error |

### Exit Code Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_PropagatesExitCode0` | `unit` | Propagates exit code 0 from successful subcommand without state modification. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock subcommand returns 0; test verifies root command state unchanged | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Returns exit code `0` unchanged; root command has no internal state to restore (stateless operation confirmed) |
| `TestRootCommand_PropagatesExitCode2` | `unit` | Propagates exit code 2 from subcommand transport error without state changes. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock subcommand returns 2; test verifies root command state unchanged | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Returns exit code `2` unchanged; stderr contains error from subcommand; root command has no state to roll back (stateless) |
| `TestRootCommand_PropagatesExitCode3` | `unit` | Propagates exit code 3 from subcommand runtime error without state changes. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock subcommand returns 3; test verifies root command state unchanged | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Returns exit code `3` unchanged; stderr contains error from subcommand; root command has no state to roll back (stateless) |

### Boundary Values — Edge Cases

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_SessionIDWithSpecialChars` | `unit` | Accepts session ID with special characters. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "abc-123-def-456"]` | Root command accepts value; passes to subcommand without modification |
| `TestRootCommand_VeryLongSessionID` | `unit` | Accepts very long session ID value. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "<1000-character-string>"]` | Root command accepts value; passes to subcommand without modification |
| `TestRootCommand_MultipleFlags` | `unit` | Handles multiple flags including global and subcommand-specific. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>", "--message", "test"]` | Parses both global `--session-id` and subcommand flag `--message` correctly |

### Validation Failures — Flag Combinations

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_DuplicateSessionIDFlag` | `unit` | Returns error when --session-id flag provided multiple times. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "<uuid1>", "--session-id", "<uuid2>"]` | Returns exit code `1`; stderr indicates duplicate flag error |
| `TestRootCommand_MalformedFlag` | `unit` | Returns error when flag is malformed. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id"]` (missing value) | Returns exit code `1`; stderr indicates missing flag value |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_RepeatedInvocation` | `unit` | Multiple invocations with same arguments produce consistent results. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | Execute root command three times with identical arguments | All three invocations return same exit code and produce same output |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_CallsSpectraFinder` | `unit` | Calls SpectraFinder before dispatching to subcommand. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock SpectraFinder injected | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Mock SpectraFinder called once with current working directory; result passed to subcommand |
| `TestRootCommand_DoesNotCallSocketClient` | `unit` | Root command does not perform socket operations directly. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; mock SocketClient injected | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | Mock SocketClient never called by root command; only called by subcommand handler |

### Error Output Format

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_ErrorPrefixFormat` | `unit` | All error messages are prefixed with "Error: ". | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created; test changes working directory to test fixture | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` | stderr output starts with `/^Error: /` |
| `TestRootCommand_ErrorOutputToStderr` | `unit` | Error messages printed to stderr, not stdout. | Temporary test directory created programmatically within test fixture; test changes working directory to test fixture | `args=["event", "emit", "MyEvent"]` (missing session-id) | Error message appears in stderr; stdout is empty or contains only usage info |

### State Isolation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_NoStateBetweenInvocations` | `unit` | Root command maintains no state between invocations. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture | First invocation with `args=["event", "emit", "Event1", "--session-id", "<uuid1>"]`, second with `args=["event", "emit", "Event2", "--session-id", "<uuid2>"]` | Second invocation does not use any state from first; each invocation independent |

### Environment Variable Behavior

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRootCommand_IgnoresEnvironmentVariables` | `unit` | Does not read SPECTRA_SESSION_ID environment variable. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; `SPECTRA_SESSION_ID=<uuid>` set in test environment | `args=["event", "emit", "MyEvent"]` (no --session-id flag) | Returns exit code `1`; requires explicit `--session-id` flag; ignores environment variable |
| `TestRootCommand_IgnoresClaudeSessionIDEnv` | `unit` | Does not read SPECTRA_CLAUDE_SESSION_ID environment variable. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; `SPECTRA_CLAUDE_SESSION_ID=<id>` set in test environment | `args=["event", "emit", "MyEvent", "--session-id", "<uuid>"]` (no --claude-session-id flag) | Subcommand receives empty or absent `--claude-session-id`; environment variable ignored |
