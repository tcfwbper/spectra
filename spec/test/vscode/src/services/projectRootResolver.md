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
| `should not perform any file system operations` | `unit` | Method performs no I/O — purely computational. | Stub `vscode.workspace.workspaceFolders` to return `[{ uri: { fsPath: '/workspace' } }]`. Stub `vscode.workspace.getConfiguration('spectra').get('projectRoot')` to return `'sub'`. Spy on `fs` module methods (e.g., `existsSync`, `mkdirSync`). | *(none — static method)* | No `fs` methods are called; returns `'/workspace/sub'` |
