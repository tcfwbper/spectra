# Test Specification: `eventScanner.test.ts`

## Source File Under Test
`vscode/src/services/eventScanner.ts`

## Test File
`vscode/test/suite/eventScanner.test.ts`

---

## `EventScanner`

### Happy Path — scanEvents

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return array of event summaries from valid events file` | `unit` | Parses a well-formed events.jsonl file and returns event summary objects. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with two valid JSON lines: `{"Type":"ReviewNeeded","EmittedBy":"architect","Message":"done"}\n{"Type":"Error","EmittedBy":"runner","Message":"fail"}`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='abc-123'`, `logger` | Returns `[{Type:'ReviewNeeded', EmittedBy:'architect', Message:'done'}, {Type:'Error', EmittedBy:'runner', Message:'fail'}]` |
| `should return single-element array for file with one valid line and no trailing newline` | `unit` | Parses a file containing exactly one valid JSON line without a trailing newline. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `{"Type":"Info","EmittedBy":"node1","Message":"hello"}` (no newline at end). Create a mock logger. | `projectRoot='/project'`, `sessionId='s1'`, `logger` | Returns `[{Type:'Info', EmittedBy:'node1', Message:'hello'}]` |
| `should extract only Type, EmittedBy, and Message from lines with extra keys` | `unit` | Ignores extra JSON keys beyond the three required ones. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `{"Type":"X","EmittedBy":"Y","Message":"Z","extra":"ignored","count":42}\n`. Create a mock logger. | `projectRoot='/project'`, `sessionId='s2'`, `logger` | Returns `[{Type:'X', EmittedBy:'Y', Message:'Z'}]`; returned object does not contain `extra` or `count` keys |
| `should preserve file order in returned array` | `unit` | Returns events in the order they appear in the file (first line first). | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with three valid JSON lines with distinct `Type` values `A`, `B`, `C` in that order. Create a mock logger. | `projectRoot='/project'`, `sessionId='s3'`, `logger` | Returns array where index 0 has `Type:'A'`, index 1 has `Type:'B'`, index 2 has `Type:'C'` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return empty array when file does not exist` | `unit` | Returns empty array and warns when the events.jsonl file is missing. | Stub `fs.access` to reject with an error (file not found). Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='no-such'`, `logger` | Returns `[]`; `logger.warn` called once with a descriptive message |
| `should return empty array for completely empty file` | `unit` | Returns empty array without warning when the file has zero bytes. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `''`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='empty'`, `logger` | Returns `[]`; `logger.warn` not called |
| `should return empty array for file with only whitespace lines` | `unit` | Returns empty array without warning when the file contains only blank/whitespace lines. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `'  \n\n   \n'`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='blanks'`, `logger` | Returns `[]`; `logger.warn` not called |
| `should silently skip trailing newline` | `unit` | Trailing newline does not produce a warning or extra entry. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `{"Type":"A","EmittedBy":"B","Message":"C"}\n`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='trail'`, `logger` | Returns one-element array; `logger.warn` not called |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should warn and skip line when JSON parsing fails` | `unit` | Logs a warning and skips a malformed JSON line without throwing. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `not-json\n{"Type":"OK","EmittedBy":"n","Message":"m"}\n`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='bad'`, `logger` | Returns `[{Type:'OK', EmittedBy:'n', Message:'m'}]`; `logger.warn` called once (for the invalid line) |
| `should warn and skip line when required key is missing` | `unit` | Logs a warning and skips a line with valid JSON but a missing required key. | Stub `fs.access` to resolve successfully. Stub `fs.readFile` to resolve with `{"Type":"X","EmittedBy":"Y"}\n{"Type":"A","EmittedBy":"B","Message":"C"}\n`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='missing'`, `logger` | Returns `[{Type:'A', EmittedBy:'B', Message:'C'}]`; `logger.warn` called once (for the line missing `Message`) |
| `should never throw to the caller` | `unit` | Returns empty array and logs a warning even when the file cannot be read. | Stub `fs.access` to reject with a permission error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='noperm'`, `logger` | Returns `[]`; does not throw |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should construct correct file path from projectRoot and sessionId` | `unit` | Reads from the correct path combining projectRoot, `.spectra/sessions`, sessionId, and `events.jsonl`. | Stub `fs.access` to resolve. Stub `fs.readFile` with a spy that resolves with `''`. Create a mock logger. | `projectRoot='/my/root'`, `sessionId='sess-42'`, `logger` | `fs.readFile` called with path `/my/root/.spectra/sessions/sess-42/events.jsonl` and encoding `'utf-8'` (or equivalent) |
| `should call logger.warn with descriptive message on missing file` | `unit` | Warning message includes useful context about the missing file. | Stub `fs.access` to reject. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `sessionId='gone'`, `logger` | `logger.warn` called once; the message string is non-empty |
| `should not call any write operations on fs` | `unit` | The method performs no mutations to the filesystem. | Stub `fs.access` to resolve. Stub `fs.readFile` to resolve with valid content. Spy on `fs.writeFile`, `fs.mkdir`, `fs.unlink`. Create a mock logger. | `projectRoot='/project'`, `sessionId='s1'`, `logger` | None of `fs.writeFile`, `fs.mkdir`, `fs.unlink` are called |
