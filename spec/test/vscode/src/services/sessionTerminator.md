# Test Specification: `sessionTerminator.test.ts`

## Source File Under Test
`vscode/src/services/sessionTerminator.ts`

## Test File
`vscode/test/suite/sessionTerminator.test.ts`

---

## `SessionTerminator`

### Happy Path — terminate

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return terminated with sigterm when process responds to SIGTERM` | `unit` | Process dies during grace period after SIGTERM. | Stub `vscode.workspace.getConfiguration` to return `'spectra'` for `spectra.binaryPath`. Stub `process.kill(pid, 0)` to succeed (process alive). Stub `child_process.execFile` for `ps` to return `'spectra\n'`. Stub `process.kill(pid, 'SIGTERM')` to succeed. Use fake timers. After first poll (500ms), stub `process.kill(pid, 0)` to throw ESRCH (process dead). Provide mock logger. | `pid=1234` | Returns `{ terminated: true, method: 'sigterm' }` |
| `should return terminated with sigkill when process ignores SIGTERM` | `unit` | Process survives grace period and is killed with SIGKILL. | Stub config to return `'spectra'`. Stub `process.kill(pid, 0)` to always succeed during grace period. Stub `execFile` for `ps` to return `'spectra\n'`. Stub `process.kill(pid, 'SIGTERM')` to succeed. Use fake timers. After SIGKILL, stub `process.kill(pid, 0)` to throw ESRCH. Provide mock logger. | `pid=5678` | After advancing fake timers 5000ms (grace period), `process.kill(pid, 'SIGKILL')` is called; returns `{ terminated: true, method: 'sigkill' }` |
| `should log info when SIGTERM is sent` | `unit` | Logs SIGTERM delivery. | Stub config. Stub process alive. Stub `ps` to match. Stub SIGTERM to succeed. Use fake timers. Make process die on first poll. Provide spy logger. | `pid=100` | `logger.info` called with a string indicating SIGTERM sent |
| `should log warn when escalating to SIGKILL` | `unit` | Logs SIGKILL escalation. | Stub config. Stub process alive for entire grace period. Stub `ps` to match. Stub SIGTERM to succeed. Use fake timers. Advance past grace period. Provide spy logger. | `pid=100` | `logger.warn` called with a string indicating SIGKILL escalation |

### Happy Path — already dead

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return already_dead when process does not exist` | `unit` | PID not alive at initial check. | Stub config. Stub `process.kill(pid, 0)` to throw error with `code: 'ESRCH'`. Provide mock logger. | `pid=9999` | Returns `{ terminated: true, method: 'already_dead' }` |
| `should return already_dead when process dies between check and SIGTERM` | `unit` | Race condition: process dies after liveness check but before signal delivery. | Stub config. Stub `process.kill(pid, 0)` to succeed on first call. Stub `execFile` for `ps` to return `'spectra\n'`. Stub `process.kill(pid, 'SIGTERM')` to throw error with `code: 'ESRCH'`. Provide mock logger. | `pid=4321` | Returns `{ terminated: true, method: 'already_dead' }` |

### Happy Path — configuration default

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should default to spectra when config is undefined` | `unit` | Falls back to `"spectra"` when configuration returns undefined. | Stub `vscode.workspace.getConfiguration` to return `undefined` for `spectra.binaryPath`. Stub process alive. Stub `execFile` for `ps` to return `'spectra\n'`. Stub signals. Use fake timers. Make process die on poll. Provide mock logger. | `pid=100` | `ps` output matched against basename `'spectra'`; termination proceeds |
| `should default to spectra when config is empty string` | `unit` | Falls back to `"spectra"` when configuration returns empty string. | Stub `vscode.workspace.getConfiguration` to return `''` for `spectra.binaryPath`. Stub process alive. Stub `execFile` for `ps` to return `'spectra\n'`. Stub signals. Use fake timers. Make process die on poll. Provide mock logger. | `pid=100` | `ps` output matched against basename `'spectra'`; termination proceeds |

### Boundary Values — command name matching

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should match when ps reports basename of configured path` | `unit` | Configured path is absolute; basename matches ps output. | Stub config to return `'/usr/local/bin/spectra'`. Stub process alive. Stub `execFile` for `ps` to return `'spectra\n'`. Stub signals. Use fake timers. Make process die on poll. Provide mock logger. | `pid=100` | Proceeds with termination; returns `{ terminated: true, method: 'sigterm' }` |
| `should match when ps reports literal spectra regardless of config` | `unit` | Even with a custom config path, literal `"spectra"` always matches. | Stub config to return `'/opt/custom/spectra-dev'`. Stub process alive. Stub `execFile` for `ps` to return `'spectra\n'`. Provide mock logger. Use fake timers. Make process die on poll. | `pid=100` | Proceeds with termination |
| `should match custom binary name from config basename` | `unit` | Custom binary name matches when ps reports it. | Stub config to return `'/opt/bin/spectra-dev'`. Stub process alive. Stub `execFile` for `ps` to return `'spectra-dev\n'`. Stub signals. Use fake timers. Make process die on poll. Provide mock logger. | `pid=100` | Proceeds with termination; returns `{ terminated: true, method: 'sigterm' }` |
| `should return not_spectra when command name does not match` | `unit` | Process command name does not match expected binary. | Stub config to return `'spectra'`. Stub process alive. Stub `execFile` for `ps` to return `'node\n'`. Provide mock logger. | `pid=100` | Returns `{ terminated: false, method: 'not_spectra' }`. No signal sent. |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return error result when SIGTERM fails with EPERM` | `unit` | Permission denied on signal delivery. | Stub config. Stub process alive. Stub `execFile` for `ps` to return `'spectra\n'`. Stub `process.kill(pid, 'SIGTERM')` to throw error with `code: 'EPERM'`. Provide mock logger. | `pid=100` | Returns `{ terminated: false, method: 'sigterm', error: '<descriptive message>' }` |
| `should return not_spectra with error when ps command fails` | `unit` | The `ps` command is unavailable or errors. | Stub config. Stub process alive. Stub `execFile` for `ps` to reject with an error. Provide spy logger. | `pid=100` | `logger.error` called; returns `{ terminated: false, method: 'not_spectra', error: '<descriptive message>' }` |
| `should never throw to caller` | `unit` | All outcomes are expressed as TerminationResult, never thrown. | Stub config. Stub `process.kill(pid, 0)` to throw an unexpected error (not ESRCH). Provide mock logger. | `pid=100` | Promise resolves (does not reject) with a result containing an `error` field |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should poll liveness every 500ms during grace period` | `unit` | Verifies the polling interval during the SIGTERM grace period. | Stub config. Stub process alive. Stub `ps` to match. Stub SIGTERM to succeed. Use fake timers. Keep process alive until 2500ms, then die. Provide mock logger. | `pid=100` | `process.kill(pid, 0)` called at approximately 500ms intervals; process confirmed dead after 5th poll (2500ms) |
| `should use 5 second grace period before SIGKILL` | `unit` | Verifies the exact grace period duration. | Stub config. Stub process alive always. Stub `ps` to match. Stub SIGTERM to succeed. Use fake timers. Provide mock logger. | `pid=100` | `process.kill(pid, 'SIGKILL')` called after exactly 5000ms; not called before 5000ms |
| `should wait 500ms after SIGKILL to confirm death` | `unit` | Verifies confirmation wait after SIGKILL. | Stub config. Stub process alive during grace period. Stub `ps` to match. Stub SIGTERM to succeed. Use fake timers. After SIGKILL, stub `process.kill(pid, 0)` to throw ESRCH. Provide mock logger. | `pid=100` | After SIGKILL, a 500ms wait occurs before final liveness check and result is returned |
| `should read configuration on every invocation` | `unit` | Configuration is read fresh each time, never cached. | Stub `vscode.workspace.getConfiguration` with a spy. Stub process alive. Stub `ps` to match. Stub signals. Use fake timers. Make process die on poll. Provide mock logger. | Call `terminate` twice with `pid=100` | `getConfiguration` called at least twice (once per invocation) |
| `should not send any signal when process is not_spectra` | `unit` | No SIGTERM or SIGKILL sent when command name mismatches. | Stub config. Stub process alive. Stub `ps` to return `'nginx\n'`. Spy on `process.kill`. Provide mock logger. | `pid=100` | `process.kill` called only with signal `0` (liveness check); never called with `'SIGTERM'` or `'SIGKILL'` |
