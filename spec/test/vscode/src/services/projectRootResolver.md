# Test Specification: `projectRootResolver.test.ts`

## Source File Under Test
`vscode/src/services/projectRootResolver.ts`

## Test File
`vscode/test/suite/projectRootResolver.test.ts`

---

## `ProjectRootResolver`

### Happy Path — resolve

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return workspace path when projectRoot config is not set` | `unit` | Returns the workspace folder path when no configuration override is present. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/home/user/project' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `undefined`. | *(none — static method)* | Returns `'/home/user/project'` |
| `should return joined path when projectRoot config is set` | `unit` | Joins the workspace path with the configured project root value. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/home/user/project' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `'sub/folder'`. | *(none — static method)* | Returns `'/home/user/project/sub/folder'` (result of `path.join`) |
| `should normalize path segments with dot-dot in config value` | `unit` | Joins and normalizes when config contains `..` segments. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/home/user/project' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `'../sibling'`. | *(none — static method)* | Returns `'/home/user/sibling'` (result of `path.join('/home/user/project', '../sibling')`) |

### Happy Path — isInitialized

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return true when .spectra directory exists` | `unit` | Returns true if stat succeeds for the .spectra path. | Stub `vscode.workspace.fs.stat` to resolve successfully (no error). Stub `vscode.Uri.file` to return a mock URI. | `ProjectRootResolver.isInitialized('/workspace')` | Returns `true` |
| `should return false when .spectra directory does not exist` | `unit` | Returns false if stat throws (directory missing). | Stub `vscode.workspace.fs.stat` to reject/throw an error (e.g., `FileNotFound`). Stub `vscode.Uri.file` to return a mock URI. | `ProjectRootResolver.isInitialized('/workspace')` | Returns `false` |
| `should construct URI with path.join of projectRoot and .spectra` | `unit` | Constructs the correct path for the stat call. | Stub `vscode.workspace.fs.stat` to resolve successfully. Spy on `vscode.Uri.file`. | `ProjectRootResolver.isInitialized('/my/project')` | `vscode.Uri.file` called with the result of `path.join('/my/project', '.spectra')` (i.e., `'/my/project/.spectra'`) |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return undefined when workspaceFolders is undefined` | `unit` | Returns undefined when VS Code has no workspace open. | Stub `vscode.workspace.workspaceFolders` to return `undefined`. | *(none — static method)* | Returns `undefined` |
| `should return undefined when workspaceFolders is empty array` | `unit` | Returns undefined when workspace folders array is empty. | Stub `vscode.workspace.workspaceFolders` to return `[]`. | *(none — static method)* | Returns `undefined` |
| `should return undefined when first workspace folder fsPath is empty string` | `unit` | Treats empty fsPath as falsy and returns undefined. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '' } }]`. | *(none — static method)* | Returns `undefined` |
| `should return workspace path when projectRoot config is null` | `unit` | Treats null config value as no value. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/workspace' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `null`. | *(none — static method)* | Returns `'/workspace'` |
| `should return workspace path when projectRoot config is empty string` | `unit` | Treats empty string config value as no value. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/workspace' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `''`. | *(none — static method)* | Returns `'/workspace'` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should read only the first workspace folder` | `unit` | Uses only the first workspace folder when multiple are present. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/first' } }, { uri: { fsPath: '/second' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `undefined`. | *(none — static method)* | Returns `'/first'` |
| `should call getConfiguration with spectra section` | `unit` | Reads configuration from the correct section and key. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/workspace' } }]`. Create a sinon spy on `vscode.workspace.getConfiguration`. Stub the returned config object's `get` method to return `'custom'`. | *(none — static method)* | `getConfiguration` called with `'spectra'`; `get` called with `'projectRoot'` |
| `should not create or write any file or directory` | `unit` | Method performs no mutation I/O. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/workspace' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `'sub'`. Spy on `vscode.workspace.fs.createDirectory` and `vscode.workspace.fs.writeFile`. | *(none — static method)* | No `createDirectory` or `writeFile` methods are called; returns `'/workspace/sub'` |
| `should call vscode.workspace.fs.stat in isInitialized` | `unit` | isInitialized delegates existence check to workspace.fs.stat. | Spy on `vscode.workspace.fs.stat`. Stub it to resolve successfully. Stub `vscode.Uri.file`. | `ProjectRootResolver.isInitialized('/workspace')` | `vscode.workspace.fs.stat` called exactly once with the URI for `'/workspace/.spectra'` |
