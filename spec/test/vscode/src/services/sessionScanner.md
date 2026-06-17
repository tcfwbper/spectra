# Test Specification: `sessionScanner.test.ts`

## Source File Under Test
`vscode/src/services/sessionScanner.ts`

## Test File
`vscode/test/suite/sessionScanner.test.ts`

---

## `SessionScanner`

### Happy Path — scanSessions

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return sorted array of session summaries` | `unit` | Reads multiple session directories and returns summaries sorted by createdAt descending. | Stub `fs.access` to resolve (directory exists). Stub `fs.readdir` to return `['sess-1', 'sess-2']` with `withFileTypes` entries marked as directories. Stub `fs.readFile` for `sess-1/session.json` to resolve with `{"id":"sess-1","workflowName":"build","createdAt":1000,"pid":100,"status":"completed","currentState":"done"}`. Stub `fs.readFile` for `sess-2/session.json` to resolve with `{"id":"sess-2","workflowName":"test","createdAt":2000,"pid":200,"status":"running","currentState":"execute"}`. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns `[{id:'sess-2', workflowName:'test', createdAt:2000, pid:200, status:'running', currentState:'execute'}, {id:'sess-1', workflowName:'build', createdAt:1000, pid:100, status:'completed', currentState:'done'}]` |
| `should extract only the six required fields from session.json` | `unit` | Ignores extra keys beyond the six required ones. | Stub `fs.access` to resolve. Stub `fs.readdir` to return one directory entry `['s1']`. Stub `fs.readFile` for `s1/session.json` to resolve with `{"id":"s1","workflowName":"w","createdAt":500,"pid":1,"status":"running","currentState":"init","extra":"ignored","debug":true}`. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns one-element array; returned object contains only `id`, `workflowName`, `createdAt`, `pid`, `status`, `currentState` — no `extra` or `debug` keys |
| `should include sessions with same createdAt without error` | `unit` | Both sessions are included when they share the same createdAt value. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['a', 'b']` as directories. Stub both `session.json` files with `createdAt:1000` but different `id` values. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns array of length 2; both sessions are present (relative order unspecified) |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return empty array when sessions directory does not exist` | `unit` | Warns and returns empty array when the sessions directory is missing. | Stub `fs.access` to reject with file-not-found error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` called once with a descriptive message |
| `should return empty array when sessions directory is empty` | `unit` | Returns empty array without warning when directory exists but has no entries. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `[]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` not called |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should warn and skip session when session.json is missing` | `unit` | Logs warning and skips a session directory that has no session.json file. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['s1']` as directory. Stub `fs.readFile` for `s1/session.json` to reject with file-not-found error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` called once |
| `should warn and skip session when session.json is malformed JSON` | `unit` | Logs warning and skips when JSON parsing fails. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['s1']` as directory. Stub `fs.readFile` for `s1/session.json` to resolve with `'not valid json{'`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` called once |
| `should warn and skip session when required key is missing` | `unit` | Logs warning and skips when JSON is valid but a required key is absent. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['s1']` as directory. Stub `fs.readFile` for `s1/session.json` to resolve with `{"id":"s1","workflowName":"w","createdAt":100,"pid":1,"status":"running"}` (missing `currentState`). Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` called once |
| `should never throw to the caller` | `unit` | Returns empty array even when directory access fails unexpectedly. | Stub `fs.access` to reject with a permission error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; does not throw |
| `should skip regular files in sessions directory` | `unit` | Only traverses subdirectories; regular files at the top level are ignored. | Stub `fs.access` to resolve. Stub `fs.readdir` to return entries where `['notes.txt']` is a file and `['s1']` is a directory. Stub `fs.readFile` for `s1/session.json` to resolve with valid JSON containing all six fields. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns one-element array for `s1`; `logger.warn` not called; `notes.txt` not processed |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should construct correct sessions directory path` | `unit` | Checks for existence of the correct path combining projectRoot and `.spectra/sessions`. | Stub `fs.access` with a spy that resolves. Stub `fs.readdir` to return `[]`. Create a mock logger. | `projectRoot='/my/root'`, `logger` | `fs.access` (or equivalent existence check) called with path `/my/root/.spectra/sessions` |
| `should read session.json from each session subdirectory` | `unit` | Reads the correct file path for each discovered session directory. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['sess-a', 'sess-b']` as directories. Stub `fs.readFile` with a spy that resolves with valid JSON. Create a mock logger. | `projectRoot='/root'`, `logger` | `fs.readFile` called with `/root/.spectra/sessions/sess-a/session.json` and `/root/.spectra/sessions/sess-b/session.json` |
| `should not call any write operations on fs` | `unit` | The method performs no mutations to the filesystem. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['s1']` as directory. Stub `fs.readFile` to resolve with valid JSON. Spy on `fs.writeFile`, `fs.mkdir`, `fs.unlink`. Create a mock logger. | `projectRoot='/project'`, `logger` | None of `fs.writeFile`, `fs.mkdir`, `fs.unlink` are called |

### Ordering — createdAt descending

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should sort sessions by createdAt in descending order` | `unit` | Most recent session appears first in the returned array. | Stub `fs.access` to resolve. Stub `fs.readdir` to return `['old', 'mid', 'new']` as directories. Stub `fs.readFile` for each with `createdAt` values `100`, `500`, `900` respectively. All have valid required fields. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns array where index 0 has `createdAt:900`, index 1 has `createdAt:500`, index 2 has `createdAt:100` |
