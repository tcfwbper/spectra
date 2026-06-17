# Test Specification: `getWebviewContent.test.ts`

## Source File Under Test
`vscode/src/views/getWebviewContent.ts`

## Test File
`vscode/test/suite/getWebviewContent.test.ts`

---

## `getWebviewContent`

### Happy Path — getWebviewContent

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return a valid HTML5 document` | `unit` | Output starts with DOCTYPE and contains html/head/body tags. | Create a stub `vscode.Webview` with `cspSource` returning `'https://test.csp'`. Create a stub `vscode.Uri` for extensionUri. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string starts with `<!DOCTYPE html>` and contains `<html`, `<head`, `<body` tags |
| `should include CSP meta tag with nonce-gated style-src and script-src` | `unit` | CSP meta tag is correctly formed. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains a `<meta` tag with `default-src 'none'`, `style-src 'nonce-` and `script-src 'nonce-` |
| `should include a style block with matching nonce` | `unit` | Style tag uses the same nonce as CSP. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains exactly one `<style nonce="..."` where the nonce value matches the nonce in the CSP meta tag |
| `should include a script block with matching nonce` | `unit` | Script tag uses the same nonce as CSP. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains exactly one `<script nonce="..."` where the nonce value matches the nonce in the CSP meta tag |
| `should generate a unique nonce per invocation` | `unit` | Nonce changes between calls. | Create a stub `vscode.Webview` with `cspSource`. | Call `getWebviewContent(stubWebview, stubExtensionUri)` twice | The nonce extracted from the first result differs from the nonce extracted from the second result |
| `should contain header element with text Spectra` | `unit` | Header displays application name. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains a header element with the text content `Spectra` |
| `should contain sessions list page element` | `unit` | DOM contains page-sessions. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="page-sessions"` |
| `should contain session detail page element` | `unit` | DOM contains page-detail. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="page-detail"` |
| `should contain workflow-select dropdown` | `unit` | Sessions page has the workflow dropdown. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains a `<select` element with `id="workflow-select"` |
| `should contain Run button` | `unit` | Sessions page has the Run button. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="btn-run"` |
| `should contain session-list container` | `unit` | Sessions page has the session list container. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="session-list"` |
| `should contain back button on detail page` | `unit` | Detail page has the back button. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="btn-back"` |
| `should contain event-list container on detail page` | `unit` | Detail page has the event list. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="event-list"` |
| `should contain event-type-select dropdown on detail page` | `unit` | Detail page has the event type dropdown. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains a `<select` element with `id="event-type-select"` |
| `should contain event-message-input on detail page` | `unit` | Detail page has the text input. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an `<input` element with `id="event-message-input"` |
| `should contain Send button on detail page` | `unit` | Detail page has the Send button. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains an element with `id="btn-send"` |
| `should include acquireVsCodeApi call in script` | `unit` | Client JS acquires the API. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains `acquireVsCodeApi()` |
| `should include message event listener in script` | `unit` | Client JS listens for messages. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string contains `addEventListener('message'` or `addEventListener("message"` |
| `should not contain inline event handlers` | `unit` | No onclick attributes in HTML. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string does not contain `onclick=`, `onsubmit=`, `onchange=`, or `onkeydown=` |
| `should not use eval in script` | `unit` | No eval usage. | Create a stub `vscode.Webview` with `cspSource`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string does not contain `eval(` |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should access webview.cspSource` | `unit` | CSP source is read from webview. | Create a stub `vscode.Webview` with `cspSource` as a sinon property spy returning `'https://csp.test'`. | `getWebviewContent(stubWebview, stubExtensionUri)` | `cspSource` getter was accessed at least once |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should produce valid HTML when cspSource is empty string` | `unit` | Empty cspSource does not break generation. | Create a stub `vscode.Webview` with `cspSource` returning `''`. | `getWebviewContent(stubWebview, stubExtensionUri)` | Returned string starts with `<!DOCTYPE html>` and contains CSP meta tag (nonce-based parts still present) |

### Idempotency

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should produce structurally consistent output on repeated calls` | `unit` | Same structure each time (aside from nonce). | Create a stub `vscode.Webview` with `cspSource`. | Call `getWebviewContent(stubWebview, stubExtensionUri)` twice | Both results contain the same set of element IDs (`page-sessions`, `page-detail`, `btn-run`, `btn-back`, `btn-send`, `session-list`, `event-list`, `workflow-select`, `event-type-select`, `event-message-input`) |
