# Test Specification: `eventDispatcher.test.ts`

## Source File Under Test
`vscode/src/services/eventDispatcher.ts`

## Test File
`vscode/test/suite/eventDispatcher.test.ts`

---

## `EventDispatcher`

### Happy Path — dispatch

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should spawn spectra-agent with correct arguments` | `unit` | Dispatches an event by spawning the binary with the correct argument list. | Stub `vscode.workspace.getConfiguration` to return `'spectra-agent'` for `spectra.agentBinaryPath`. Stub `child_process.execFile` to return a mock child process (EventEmitter with no errors). Provide a mock logger. | `eventType='ReviewNeeded'`, `sessionId='abc-123'`, `message='hello world'` | `execFile` called with binary `'spectra-agent'` and args `['event', 'emit', 'ReviewNeeded', '--session-id', 'abc-123', '--message', 'hello world']`; promise resolves with `void` |
| `should log info message with event type and session id` | `unit` | Logs a diagnostic info message upon successful spawn. | Stub `vscode.workspace.getConfiguration` to return `'spectra-agent'`. Stub `execFile` to return a mock child process (no errors). Provide a spy logger. | `eventType='SessionStarted'`, `sessionId='uuid-1'`, `message='started'` | `logger.info` called with a string containing `'SessionStarted'` and `'uuid-1'` |
| `should resolve without waiting for child process exit` | `unit` | The promise resolves immediately after successful spawn without awaiting exit. | Stub config. Stub `execFile` to return a mock child process that never emits `exit`. Use fake timers. Provide mock logger. | `eventType='Ping'`, `sessionId='s1'`, `message='m'` | Promise resolves immediately without advancing timers |

### Happy Path — configuration default

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should default to spectra-agent when config is undefined` | `unit` | Falls back to `"spectra-agent"` when configuration returns undefined. | Stub `vscode.workspace.getConfiguration` to return `undefined` for `spectra.agentBinaryPath`. Stub `execFile` to return a mock child process. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='m'` | `execFile` called with binary `'spectra-agent'` |
| `should default to spectra-agent when config is empty string` | `unit` | Falls back to `"spectra-agent"` when configuration returns empty string. | Stub `vscode.workspace.getConfiguration` to return `''` for `spectra.agentBinaryPath`. Stub `execFile` to return a mock child process. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='m'` | `execFile` called with binary `'spectra-agent'` |
| `should use custom binary path from configuration` | `unit` | Uses the configured path when it is a non-empty string. | Stub `vscode.workspace.getConfiguration` to return `'/opt/bin/spectra-agent'` for `spectra.agentBinaryPath`. Stub `execFile` to return a mock child process. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='m'` | `execFile` called with binary `'/opt/bin/spectra-agent'` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should throw when spawn fails with ENOENT` | `unit` | Rejects the promise when the binary is not found. | Stub config to return `'/missing/spectra-agent'`. Stub `execFile` to return a mock child process that emits an `error` event with `code: 'ENOENT'` synchronously. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='m'` | Promise rejects with an error whose message includes `'/missing/spectra-agent'` |
| `should throw when spawn fails with EACCES` | `unit` | Rejects the promise when permission is denied. | Stub config to return `'/no-exec/spectra-agent'`. Stub `execFile` to return a mock child process that emits an `error` event with `code: 'EACCES'` synchronously. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='m'` | Promise rejects with an error |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should not use shell for spawn` | `unit` | Verifies execFile is called without a shell option to prevent injection. | Stub config. Spy on `execFile` to capture options. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='m'` | `execFile` is called without `shell: true` in options (or no options object with shell) |
| `should log warning on non-zero exit code` | `unit` | When the child process exits with a non-zero code, logger.warn is called. | Stub config. Stub `execFile` to return a mock child process. Provide a spy logger. Trigger the child process `exit` event with code `1` after spawn. | `eventType='E'`, `sessionId='s'`, `message='m'` | `logger.warn` called with a string containing exit code `1` |
| `should not throw on non-zero exit code` | `unit` | A non-zero exit code does not cause the resolved promise to reject. | Stub config. Stub `execFile` to return a mock child process. Provide mock logger. Let promise resolve, then emit `exit` with code `2`. | `eventType='E'`, `sessionId='s'`, `message='m'` | Promise already resolved with `void`; no unhandled rejection |
| `should pass special characters in message without shell interpretation` | `unit` | Special characters are passed verbatim as argv. | Stub config. Spy on `execFile` to capture args. Provide mock logger. | `eventType='E'`, `sessionId='s'`, `message='hello "world" \n $PATH'` | `execFile` called with the message as a single argument element containing the special characters verbatim |
| `should read configuration on every invocation` | `unit` | Configuration is read fresh each time, never cached. | Stub `vscode.workspace.getConfiguration` with a spy. Stub `execFile` to return a mock child process. Provide mock logger. | Call `dispatch` twice | `getConfiguration` called at least twice (once per invocation) |
