# Test Specification: `sessionLauncher.test.ts`

## Source File Under Test
`vscode/src/services/sessionLauncher.ts`

## Test File
`vscode/test/suite/sessionLauncher.test.ts`

---

## `SessionLauncher`

### Happy Path — launch

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should spawn detached process with correct arguments` | `unit` | Launches a workflow by spawning the binary with the correct argument list and options. | Stub `vscode.workspace.getConfiguration` to return `'spectra'` for `spectra.binaryPath`. Stub `child_process.spawn` to return a mock child process (EventEmitter with `unref` spy). Stub `crypto.randomUUID` to return `'test-uuid-1234'`. Provide mock logger. | `workflowName='myWorkflow'` | `spawn` called with binary `'spectra'`, args `['run', '--workflow', 'myWorkflow', '--session-id', 'test-uuid-1234']`, and options `{ detached: true, stdio: 'ignore' }` |
| `should call unref on the child process` | `unit` | The spawned process is unrefed so the extension host does not wait for it. | Stub config. Stub `spawn` to return a mock child process with `unref` spy. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | `childProcess.unref()` called exactly once |
| `should log info message with workflow name and session id` | `unit` | Logs a diagnostic info message upon successful spawn. | Stub config. Stub `spawn` to return a mock child process. Stub `crypto.randomUUID` to return `'uuid-abc'`. Provide a spy logger. | `workflowName='deploy'` | `logger.info` called with a string containing `'deploy'` and `'uuid-abc'` |
| `should resolve with void on successful spawn` | `unit` | Promise resolves with void after spawn succeeds. | Stub config. Stub `spawn` to return a mock child process (no error). Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | Promise resolves with `undefined` |

### Happy Path — configuration default

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should default to spectra when config is undefined` | `unit` | Falls back to `"spectra"` when configuration returns undefined. | Stub `vscode.workspace.getConfiguration` to return `undefined` for `spectra.binaryPath`. Stub `spawn` to return a mock child process. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | `spawn` called with binary `'spectra'` |
| `should default to spectra when config is empty string` | `unit` | Falls back to `"spectra"` when configuration returns empty string. | Stub `vscode.workspace.getConfiguration` to return `''` for `spectra.binaryPath`. Stub `spawn` to return a mock child process. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | `spawn` called with binary `'spectra'` |
| `should use custom binary path from configuration` | `unit` | Uses the configured path when it is a non-empty string. | Stub `vscode.workspace.getConfiguration` to return `'/usr/local/bin/spectra'` for `spectra.binaryPath`. Stub `spawn` to return a mock child process. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | `spawn` called with binary `'/usr/local/bin/spectra'` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should throw when spawn fails with ENOENT` | `unit` | Rejects the promise when the binary is not found. | Stub config to return `'/missing/spectra'`. Stub `spawn` to return a mock child process that emits `error` with `code: 'ENOENT'` synchronously. Provide mock logger. | `workflowName='wf'` | Promise rejects with an error whose message includes `'/missing/spectra'` |
| `should throw when spawn fails with EACCES` | `unit` | Rejects the promise when permission is denied. | Stub config to return `'/no-exec/spectra'`. Stub `spawn` to return a mock child process that emits `error` with `code: 'EACCES'` synchronously. Provide mock logger. | `workflowName='wf'` | Promise rejects with an error |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should generate a fresh UUID for every invocation` | `unit` | Each call generates a new UUID, never reusing previous ones. | Stub config. Stub `spawn` to return a mock child process. Stub `crypto.randomUUID` to return different values on successive calls (`'uuid-1'`, `'uuid-2'`). Provide mock logger. | Call `launch` twice with `workflowName='wf'` | First `spawn` called with args containing `'uuid-1'`; second `spawn` called with args containing `'uuid-2'` |
| `should spawn with detached true` | `unit` | Verifies detached option is set for process independence. | Stub config. Spy on `spawn` to capture options. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | `spawn` options include `detached: true` |
| `should spawn with stdio ignore` | `unit` | Verifies stdio is set to ignore for no pipe connection. | Stub config. Spy on `spawn` to capture options. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='wf'` | `spawn` options include `stdio: 'ignore'` |
| `should read configuration on every invocation` | `unit` | Configuration is read fresh each time, never cached. | Stub `vscode.workspace.getConfiguration` with a spy. Stub `spawn` to return a mock child process. Stub `crypto.randomUUID`. Provide mock logger. | Call `launch` twice | `getConfiguration` called at least twice (once per invocation) |
| `should pass workflowName with special characters as single argv element` | `unit` | Workflow names with spaces or special characters are passed verbatim. | Stub config. Spy on `spawn` to capture args. Stub `crypto.randomUUID`. Provide mock logger. | `workflowName='my workflow (v2)'` | `spawn` args contain `'my workflow (v2)'` as a single element after `'--workflow'` |
