# Test Specification: `run.go`

## Source File Under Test
`cmd/spectra/run.go`

## Test File
`cmd/spectra/run_test.go`

---

## `RunCommand`

### Happy Path — Positional Argument

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_PositionalArgument` | `unit` | Executes workflow when workflow name is provided as positional argument. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/TestWorkflow.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run TestWorkflow` | Runtime.Run called with correct project root and `"TestWorkflow"`; exit code 0 |
| `TestRunCommand_PositionalArgumentFromSubdirectory` | `unit` | Locates project root from subdirectory and executes workflow. | Temporary test directory created programmatically within test fixture; `.spectra/` and `.spectra/workflows/SimpleSdd.yaml` created inside test fixture; subdirectory `subdir/nested/` created inside test fixture; test changes working directory to `subdir/nested/` | `run SimpleSdd` | SpectraFinder locates project root correctly; Runtime.Run called with project root and `"SimpleSdd"`; exit code 0 |

### Happy Path — Flag Argument

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_FlagArgument` | `unit` | Executes workflow when workflow name is provided via --workflow flag. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/MyWorkflow.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run --workflow MyWorkflow` | Runtime.Run called with correct project root and `"MyWorkflow"`; exit code 0 |
| `TestRunCommand_FlagPrecedenceOverPositional` | `unit` | Flag takes precedence when both flag and positional argument are provided. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/FlagWorkflow.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run --workflow FlagWorkflow PositionalWorkflow` | Runtime.Run called with `"FlagWorkflow"` (not `"PositionalWorkflow"`); exit code 0 |

### Happy Path — Help Output

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_HelpFlag` | `unit` | Displays help information when --help flag is provided. | No setup required | `run --help` | Prints usage information containing `"Run a workflow"`, `"Usage:"`, `"Flags:"`, `"--workflow string"`, `"Examples:"`; Runtime.Run not called; exit code 0 |

### Happy Path — Runtime Output Forwarding

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_ForwardsStdout` | `unit` | Forwards Runtime stdout to command stdout. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Test.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to write `"workflow output\n"` to stdout and return exit code 0 | `run Test` | Command stdout contains `"workflow output\n"`; exit code 0 |
| `TestRunCommand_ForwardsStderr` | `unit` | Forwards Runtime stderr to command stderr. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Test.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to write `"workflow error\n"` to stderr and return exit code 1 | `run Test` | Command stderr contains `"workflow error\n"`; exit code 1 |
| `TestRunCommand_ForwardsBothStreams` | `unit` | Forwards both stdout and stderr from Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Test.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to write to both stdout and stderr | `run Test` | Command stdout and stderr contain Runtime's respective output; exit code matches Runtime exit code |

### Validation Failures — Missing Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_NoWorkflowName` | `unit` | Returns error when no workflow name is provided. | No setup required | `run` | Stderr contains `"workflow name"` and `"required"`; Runtime.Run not called; exit code 1 |
| `TestRunCommand_NoWorkflowNameWithFlag` | `unit` | Returns error when --workflow flag is provided without value. | No setup required | `run --workflow` | Prints error to stderr; Runtime.Run not called; exit code 1 |

### Validation Failures — Empty Workflow Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_EmptyWorkflowNamePositional` | `unit` | Returns error when positional workflow name is empty string. | No setup required | `run ""` | Stderr contains `"workflow name"` and `"empty"`; Runtime.Run not called; exit code 1 |
| `TestRunCommand_EmptyWorkflowNameFlag` | `unit` | Returns error when --workflow flag value is empty string. | No setup required | `run --workflow ""` | Stderr contains `"workflow name"` and `"empty"`; Runtime.Run not called; exit code 1 |
| `TestRunCommand_WhitespaceWorkflowName` | `unit` | Returns error when workflow name contains only whitespace. | No setup required | `run "   "` | Stderr contains `"workflow name"` and `"empty"` after trimming; Runtime.Run not called; exit code 1 |

### Validation Failures — Too Many Arguments

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_TooManyArguments` | `unit` | Returns error when multiple positional arguments are provided. | No setup required | `run Workflow1 Workflow2` | Stderr contains `"too many arguments"`; Runtime.Run not called; exit code 1 |
| `TestRunCommand_TooManyArgumentsWithThree` | `unit` | Returns error when three positional arguments are provided. | No setup required | `run Workflow1 Workflow2 Workflow3` | Stderr contains `"too many arguments"`; Runtime.Run not called; exit code 1 |

### Validation Failures — Project Root Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_SpectraNotFound` | `unit` | Returns error when .spectra directory is not found. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created inside test fixture; test changes working directory to test fixture | `run TestWorkflow` | Stderr contains `".spectra"` and `"not found"`; Runtime.Run not called; exit code 1 |
| `TestRunCommand_SpectraNotFoundInParent` | `unit` | Returns error when .spectra directory is not found in any parent directory. | Temporary test directory created programmatically within test fixture; deep subdirectory structure created but no `.spectra/` at any level inside test fixture; test changes working directory to deepest subdirectory | `run TestWorkflow` | Stderr contains `".spectra"` and `"not found"`; Runtime.Run not called; exit code 1 |

### Error Propagation — Runtime Initialization Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeReportsWorkflowNotFound` | `unit` | Forwards Runtime error when workflow file does not exist. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created but no workflow file inside test fixture; test changes working directory to test fixture; Runtime mocked to write `"Error: workflow definition not found"` to stderr and return exit code 1 | `run NonExistent` | Command stderr contains `"Error: workflow definition not found"`; exit code 1 |
| `TestRunCommand_RuntimeReportsInvalidYAML` | `unit` | Forwards Runtime error when workflow file has invalid YAML syntax. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Invalid.yaml` with malformed YAML created inside test fixture; test changes working directory to test fixture; Runtime mocked to write YAML parse error to stderr and return exit code 1 | `run Invalid` | Command stderr contains YAML parse error from Runtime; exit code 1 |
| `TestRunCommand_RuntimeReportsWorkflowNotReadable` | `unit` | Forwards Runtime error when workflow file cannot be read due to permissions. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked to write `"Error: permission denied reading workflow file"` to stderr and return exit code 1 | `run Restricted` | Command stderr contains permission error from Runtime; exit code 1 |

### Error Propagation — Runtime Execution Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeReportsAgentError` | `unit` | Forwards Runtime error when workflow fails due to agent error. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/AgentFail.yaml` created inside test fixture; test changes working directory to test fixture; Runtime mocked to simulate agent failure with stderr output and return exit code 1 | `run AgentFail` | Command stderr contains agent error message from Runtime; exit code 1 |
| `TestRunCommand_RuntimeReportsSessionLockError` | `unit` | Forwards Runtime error when another session is already running. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked to write `"Error: another workflow session is already running"` to stderr and return exit code 1 | `run Concurrent` | Command stderr contains session lock error from Runtime; exit code 1 |
| `TestRunCommand_RuntimeRunReturnsError` | `unit` | Converts Runtime.Run error return to exit code 1 and stderr message. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Test.yaml` created inside test fixture; test changes working directory to test fixture; Runtime.Run mocked to return Go error (e.g., `errors.New("runtime internal error")`) | `run Test` | Command writes error message to stderr; exit code 1 |

### Edge Cases — Special Workflow Names

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_WorkflowNameWithPathSeparators` | `unit` | Passes workflow name with path separators to Runtime without validation. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked to handle the workflow name | `run ../malicious/workflow` | Runtime.Run called with `"../malicious/workflow"`; command does not validate or sanitize the name; Runtime's behavior determines outcome |
| `TestRunCommand_WorkflowNameWithSpecialCharacters` | `unit` | Passes workflow name with special characters to Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked | `run Work@flow#123` | Runtime.Run called with `"Work@flow#123"`; exit code matches Runtime exit code |
| `TestRunCommand_WorkflowNameWithUnicode` | `unit` | Passes workflow name with Unicode characters to Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked | `run 工作流程` | Runtime.Run called with `"工作流程"`; exit code matches Runtime exit code |

### Edge Cases — Signal Handling

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_PropagatesSIGINT` | `unit` | Propagates SIGINT signal to Runtime subprocess. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and workflow file created inside test fixture; test changes working directory to test fixture; Runtime mocked to record signal delivery and return graceful shutdown exit code; SIGINT sent to mock Runtime (not test process itself) to ensure test isolation | `run TestWorkflow`, then send SIGINT to mock Runtime | Mock Runtime records SIGINT delivery; command exits with Runtime's exit code indicating graceful shutdown |

### Edge Cases — Streaming Output

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_HandlesStreamingOutput` | `unit` | Does not timeout or buffer output from workflow that produces streaming output. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and workflow file created inside test fixture; test changes working directory to test fixture; Runtime mocked to simulate streaming behavior by producing multiple output lines over a SHORT period (100-200ms with output every 10-20ms) | `run StreamingWorkflow` | All Runtime output forwarded in real-time without buffering; command waits for Runtime to complete; exit code matches Runtime exit code |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_MultipleInvocationsIndependent` | `unit` | Multiple sequential invocations are independent. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and workflow file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | First `run TestWorkflow`, then second `run TestWorkflow` | Both invocations succeed independently; each calls Runtime.Run with same parameters; both exit with code 0 |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_CallsSpectraFinder` | `unit` | Invokes SpectraFinder to locate project root. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; SpectraFinder and Runtime mocked | `run TestWorkflow` | SpectraFinder.Find() called from current working directory; SpectraFinder returns project root; Runtime.Run called with returned project root |
| `TestRunCommand_PassesCorrectProjectRoot` | `unit` | Passes project root from SpectraFinder to Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created at path `/tmp/test-project/.spectra/` inside test fixture; subdirectory `sub/` created inside test fixture; test changes working directory to `sub/`; SpectraFinder mocked to return `/tmp/test-project/`; Runtime mocked | `run TestWorkflow` | Runtime.Run called with projectRoot=`"/tmp/test-project/"` and workflowName=`"TestWorkflow"` |
| `TestRunCommand_PassesCorrectWorkflowName` | `unit` | Passes workflow name exactly as provided to Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked | `run My-Workflow_123` | Runtime.Run called with workflowName=`"My-Workflow_123"` (exact match, no transformation) |
