# Test Specification: `workflowScanner.test.ts`

## Source File Under Test
`vscode/src/services/workflowScanner.ts`

## Test File
`vscode/test/suite/workflowScanner.test.ts`

---

## `WorkflowScanner`

### Happy Path — scanWorkflows

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return array of workflow names without yaml extension` | `unit` | Lists .yaml files and strips the extension from each name. | Stub `fs.readdir` to resolve with file entries `['build.yaml', 'deploy.yaml']` (both regular files). Create a mock logger. | `projectRoot='/project'`, `logger` | Returns `['build', 'deploy']` |
| `should include only yaml files and exclude other file types` | `unit` | Filters out non-.yaml files from the result. | Stub `fs.readdir` to resolve with entries `['workflow.yaml', 'README.md', 'config.json']` (all regular files). Create a mock logger. | `projectRoot='/project'`, `logger` | Returns `['workflow']` |
| `should return empty string element for file named dot-yaml only` | `unit` | A file named `.yaml` (no base name) produces an empty string element. | Stub `fs.readdir` to resolve with file entry `['.yaml']`. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns `['']` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return empty array when workflows directory does not exist` | `unit` | Warns and returns empty array when the workflows directory is missing. | Stub `fs.readdir` to reject with ENOENT error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` called once with a descriptive message |
| `should return empty array when directory exists but contains no yaml files` | `unit` | Returns empty array without warning when directory has only non-.yaml files. | Stub `fs.readdir` to resolve with entries `['README.md', 'notes.txt']` (all regular files). Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` not called |
| `should return empty array when directory is empty` | `unit` | Returns empty array without warning when directory has no entries. | Stub `fs.readdir` to resolve with `[]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; `logger.warn` not called |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should never throw to the caller` | `unit` | Returns empty array even when readdir fails with a permission error. | Stub `fs.readdir` to reject with EACCES (permission denied) error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `logger` | Returns `[]`; does not throw |
| `should exclude subdirectories even if they have yaml in their name` | `unit` | Only regular files are included; directories are excluded. | Stub `fs.readdir` to resolve with entries where `'subdir.yaml'` is a directory and `'real.yaml'` is a regular file. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns `['real']` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should construct correct workflows directory path` | `unit` | Calls readdir with the correct path combining projectRoot and `.spectra/workflows`. | Stub `fs.readdir` with a spy that resolves with `[]`. Create a mock logger. | `projectRoot='/my/root'`, `logger` | `fs.readdir` called with path `/my/root/.spectra/workflows` |
| `should not call any write operations on fs` | `unit` | The method performs no mutations to the filesystem. | Stub `fs.readdir` to resolve with `['test.yaml']` as file entry. Spy on `fs.writeFile`, `fs.mkdir`, `fs.unlink`. Create a mock logger. | `projectRoot='/project'`, `logger` | None of `fs.writeFile`, `fs.mkdir`, `fs.unlink` are called |
| `should not read file contents` | `unit` | The method never reads the content of any file in the directory. | Stub `fs.readdir` to resolve with `['a.yaml', 'b.yaml']` as file entries. Spy on `fs.readFile`. Create a mock logger. | `projectRoot='/project'`, `logger` | `fs.readFile` is never called |
| `should not recurse into subdirectories` | `unit` | Only top-level entries are considered; no recursive listing. | Stub `fs.readdir` to resolve with entries where `'nested'` is a directory and `'top.yaml'` is a file. Create a mock logger. | `projectRoot='/project'`, `logger` | Returns `['top']`; `fs.readdir` called exactly once (for the workflows directory only) |
