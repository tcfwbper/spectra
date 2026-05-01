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
| `TestRunCommand_PositionalArgument` | `unit` | Executes workflow when workflow name is provided as positional argument. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/TestWorkflow.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run TestWorkflow` | Runtime.Run called with `"TestWorkflow"`; exit code 0 |
| `TestRunCommand_PositionalArgumentFromSubdirectory` | `unit` | Executes workflow from subdirectory (Runtime handles project root location). | Temporary test directory created programmatically within test fixture; `.spectra/` and `.spectra/workflows/SimpleSdd.yaml` created inside test fixture; subdirectory `subdir/nested/` created inside test fixture; test changes working directory to `subdir/nested/`; Runtime mocked to return exit code 0 | `run SimpleSdd` | Runtime.Run called with `"SimpleSdd"`; exit code 0 |

### Happy Path — Flag Argument

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_FlagArgument` | `unit` | Executes workflow when workflow name is provided via --workflow flag. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/MyWorkflow.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run --workflow MyWorkflow` | Runtime.Run called with `"MyWorkflow"`; exit code 0 |
| `TestRunCommand_FlagPrecedenceOverPositional` | `unit` | Flag takes precedence when both flag and positional argument are provided. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/FlagWorkflow.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run --workflow FlagWorkflow PositionalWorkflow` | Runtime.Run called with `"FlagWorkflow"` (not `"PositionalWorkflow"`); exit code 0 |

### Happy Path — Help Output

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_HelpFlag` | `unit` | Displays help information when --help flag is provided. | No setup required | `run --help` | Prints usage information containing `"Run a workflow"`, `"Usage:"`, `"Flags:"`, `"--workflow string"`, `"Examples:"`; Runtime.Run not called; exit code 0 |

### Happy Path — Runtime Exit Code Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeExitCodeZero` | `unit` | Propagates exit code 0 from Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Test.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | `run Test` | Runtime.Run called with `"Test"`; exit code 0 |
| `TestRunCommand_RuntimeExitCodeOne` | `unit` | Propagates exit code 1 from Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Test.yaml` file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run Test` | Runtime.Run called with `"Test"`; exit code 1 |

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

### Error Propagation — Project Root Not Found

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeReportsProjectRootNotFound` | `unit` | Propagates exit code 1 when Runtime fails to locate project root. | Temporary test directory created programmatically within test fixture; no `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run TestWorkflow` | Runtime.Run called with `"TestWorkflow"`; exit code 1 |

### Error Propagation — Runtime Initialization Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeReportsWorkflowNotFound` | `unit` | Propagates exit code 1 when Runtime cannot find workflow file. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created but no workflow file inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run NonExistent` | Runtime.Run called with `"NonExistent"`; exit code 1 |
| `TestRunCommand_RuntimeReportsInvalidYAML` | `unit` | Propagates exit code 1 when Runtime encounters invalid YAML. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/Invalid.yaml` with malformed YAML created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run Invalid` | Runtime.Run called with `"Invalid"`; exit code 1 |
| `TestRunCommand_RuntimeReportsWorkflowNotReadable` | `unit` | Propagates exit code 1 when Runtime cannot read workflow file. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run Restricted` | Runtime.Run called with `"Restricted"`; exit code 1 |

### Error Propagation — Runtime Execution Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_RuntimeReportsAgentError` | `unit` | Propagates exit code 1 when Runtime fails due to agent error. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and `.spectra/workflows/AgentFail.yaml` created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run AgentFail` | Runtime.Run called with `"AgentFail"`; exit code 1 |
| `TestRunCommand_RuntimeReportsSessionLockError` | `unit` | Propagates exit code 1 when Runtime detects another session is running. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 1 | `run Concurrent` | Runtime.Run called with `"Concurrent"`; exit code 1 |

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

### Edge Cases — Blocking Runtime

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_WaitsForRuntimeCompletion` | `unit` | Command waits for Runtime.Run to return before exiting. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and workflow file created inside test fixture; test changes working directory to test fixture; Runtime mocked to block for a short period (100-200ms) before returning exit code 0 | `run TestWorkflow` | Command does not return until Runtime.Run completes; exit code 0 |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_MultipleInvocationsIndependent` | `unit` | Multiple sequential invocations are independent. | Temporary test directory created programmatically within test fixture; `.spectra/` directory and workflow file created inside test fixture; test changes working directory to test fixture; Runtime mocked to return exit code 0 | First `run TestWorkflow`, then second `run TestWorkflow` | Both invocations succeed independently; each calls Runtime.Run with same parameters; both exit with code 0 |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestRunCommand_PassesCorrectWorkflowName` | `unit` | Passes workflow name exactly as provided to Runtime. | Temporary test directory created programmatically within test fixture; `.spectra/` directory created inside test fixture; test changes working directory to test fixture; Runtime mocked | `run My-Workflow_123` | Runtime.Run called with workflowName=`"My-Workflow_123"` (exact match, no transformation) |
