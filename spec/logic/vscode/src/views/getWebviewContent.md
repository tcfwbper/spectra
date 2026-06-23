# getWebviewContent

## Overview

Generates the complete HTML string for the Spectra sidebar webview. Produces a self-contained document with inline CSS and JavaScript that renders three pages (not-initialized notice, sessions list, and session detail), handles client-side page routing via `window.addEventListener('message', ...)`, and communicates with the extension host via `vscode.postMessage(...)`. Does not perform any I/O beyond what the VS Code webview API provides.

## Boundaries

- Owns: generating a CSP-compliant HTML document, embedding inline styles and scripts, rendering the DOM structure for all three pages (not-initialized, sessions list, session detail), implementing client-side message handling and page switching, implementing button cooldown (2-second lock) logic, implementing the send-button guard (entryNode + running), and wiring all `vscode.postMessage` calls.
- Delegates: actual state data provision to the extension host (received via `postMessage`).
- Delegates: view lifecycle management to SpectraViewProvider.
- Must not: perform any filesystem I/O.
- Must not: fetch external resources (all content is inline; codicon font is loaded from the extension's local assets, not from an external URL).
- Must not: inject raw user data into HTML strings (all dynamic content is rendered via DOM manipulation in the embedded JS, not via string interpolation).
- Must not: use `eval()` or inline event handlers (`onclick` attributes) — all event binding is done in the `<script>` block.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.Webview` | Webview reference | `webview.cspSource` (for CSP header and font-src) | Must not call `postMessage` or subscribe to events |
| `vscode.Uri` | Extension URI | Used to derive codicon font URI and `localResourceRoots` context | — |
| `crypto` (Node.js) | Nonce generation | `randomBytes` or equivalent for CSP nonce | — |
| Codicon font (`@vscode/codicons`) | Icon glyphs | Referenced via CSS `@font-face` or `<link>` with a URI derived from `extensionUri` | Must not fetch from external CDN |

Construction constraint: This is a standalone exported function, not a class. Signature: `getWebviewContent(webview: vscode.Webview, extensionUri: vscode.Uri): string`.

## Behavior

### HTML Structure

1. Generates a random nonce (16+ bytes, hex-encoded) for the Content Security Policy.
2. Produces `<!DOCTYPE html>` with a `<meta>` CSP tag: `default-src 'none'; style-src 'nonce-${nonce}'; script-src 'nonce-${nonce}'; font-src ${webview.cspSource};`.
3. Embeds a single `<style nonce="${nonce}">` block with all CSS.
3a. Embeds a `<link>` or `<style>` reference to the VS Code codicon font (from the extension's `node_modules/@vscode/codicons` or bundled asset), gated by the same nonce.
4. Embeds a single `<script nonce="${nonce}">` block with all client-side JavaScript.

### Not Initialized Page DOM (id: `page-not-initialized`)

5a. A centered container displaying the text: "Please run `spectra init` to initialize the project."
5b. The text uses a muted/secondary color consistent with VS Code's sidebar theme.
5c. This page is hidden by default; shown only when a `showNotInitialized` message is received.

### Sessions List Page DOM (id: `page-sessions`)

5. A header element displaying the text "Spectra" (top-left aligned).
6. Below the header: a flex row containing a `<select>` dropdown (id: `workflow-select`) and a "Run" button (id: `btn-run`).
   - The row uses `display: flex; align-items: center; gap: 8px;`.
   - The dropdown uses `flex: 1; min-width: 0;` (fills remaining space, shrinks gracefully).
   - The "Run" button uses `flex-shrink: 0;` (fixed size, right-aligned by flex layout).
   - The dropdown is populated dynamically when `showSessions` state arrives.
   - The "Run" button triggers `launchSession` with the selected workflow name.
   - The "Run" button has a 2-second cooldown after each click (disabled + grey styling during cooldown).
7. Below the dropdown row: a container (id: `session-list`) displaying session rows.
   - Each session row is a clickable flex block with two lines:
     - Line 1: a flex row with the session label (`<WorkflowName>-<first 8 chars of session ID>`) on the left and a stop button on the far right.
       - The session label uses `flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;` (truncates with "..." when space is insufficient).
       - The stop button uses `flex-shrink: 0;` (never compressed, always visible and clickable).
     - Line 2 (subtitle): the session's `status` value.
   - The stop button is rendered as a codicon icon button: a square, transparent-background button displaying the `codicon-close` icon (or `codicon-debug-stop`). On hover: a subtle light-grey background appears (`var(--vscode-toolbar-hoverBackground)`). No text label.
   - The stop button triggers `terminateSession` with the session's `pid`.
   - The stop button has a 2-second cooldown after each click (disabled state: icon becomes faded/grey, hover background suppressed).
   - The stop button is only rendered for sessions with `status === 'running'`.
   - Clicking anywhere on the row (except the stop button) triggers `navigateToDetail` with `sessionId` and `workflowName`.

### Session Detail Page DOM (id: `page-detail`)

8. A back button (id: `btn-back`) in the top-left corner that triggers `navigateToList`.
9. Below the back button: an event history container (id: `event-list`) displaying event entries.
   - Each entry shows: `[Type] EmittedBy: Message`.
10. Below the event history: a row containing an event-type `<select>` dropdown (id: `event-type-select`), a text input (id: `event-message-input`), and a "Send" button (id: `btn-send`).
    - The dropdown is populated dynamically when `showDetail` state arrives (from `eventTypes`).
    - The "Send" button triggers `sendEvent` with the selected `eventType` and the input text.
    - The "Send" button has a 2-second cooldown after each click (disabled + grey styling during cooldown).
    - The "Send" button has an additional guard: it is enabled only when `currentState === entryNode` AND `status === 'running'`. Otherwise it is disabled with grey styling.
    - Both guards must pass simultaneously for the button to be clickable: no active cooldown AND the state condition is met.

### Client-Side JavaScript Behavior

11. Acquires VS Code API via `const vscode = acquireVsCodeApi()`.
12. Registers `window.addEventListener('message', handler)` to receive messages from the extension host.
13a. On receiving `{ type: 'showNotInitialized' }`:
    - Hides `page-sessions` and `page-detail`, shows `page-not-initialized`.
13. On receiving `{ type: 'showSessions', state }`:
    - Hides `page-detail` and `page-not-initialized`, shows `page-sessions`.
    - Populates the `workflow-select` dropdown with `state.workflows`.
    - Clears and rebuilds the `session-list` container from `state.sessions`.
    - Re-evaluates stop button visibility (only for `status === 'running'` sessions).
14. On receiving `{ type: 'showDetail', state }`:
    - Hides `page-sessions` and `page-not-initialized`, shows `page-detail`.
    - Stores `state.entryNode`, `state.currentState`, and `state.status` in module-level variables for the send-button guard.
    - Populates the `event-type-select` dropdown with `state.eventTypes`.
    - Clears and rebuilds the `event-list` container from `state.events`.
    - Re-evaluates the send button's enabled/disabled state based on the guard condition.
15. Button cooldown implementation:
    - On button click: immediately sets `button.disabled = true` and applies the disabled CSS class.
    - After 2000ms (`setTimeout`): removes `disabled` attribute and the disabled CSS class — unless the button is also held by the send-button guard (in which case it stays disabled).
16. Send-button guard re-evaluation:
    - Called on every `showDetail` message arrival.
    - Called when cooldown expires for the send button.
    - Logic: `btn-send.disabled = !(currentState === entryNode && status === 'running') || cooldownActive`.

### postMessage Calls (webview → extension)

17. `vscode.postMessage({ command: 'navigateToDetail', sessionId, workflowName })` — on session row click.
18. `vscode.postMessage({ command: 'navigateToList' })` — on back button click.
19. `vscode.postMessage({ command: 'launchSession', workflowName })` — on Run button click.
20. `vscode.postMessage({ command: 'terminateSession', pid })` — on stop button click.
21. `vscode.postMessage({ command: 'sendEvent', eventType, message })` — on Send button click.

## Inputs

| Field | Type | Constraints | Required |
|---|---|---|---|
| webview | `vscode.Webview` | Valid webview instance from the sidebar view | Yes |
| extensionUri | `vscode.Uri` | Extension root URI | Yes |

## Outputs

| Field | Type | Description |
|---|---|---|
| result | `string` | Complete HTML document string ready to assign to `webview.html`. |

## Invariants

- Must include a Content Security Policy meta tag with `default-src 'none'`, nonce-gated `style-src` and `script-src`, and `font-src ${webview.cspSource}` (for codicon font loading from extension local assets).
- Must never inject dynamic data (user content, session IDs, messages) via string interpolation into the HTML template — all dynamic rendering happens via DOM manipulation in the embedded JS after `message` events.
- Must not use inline event handlers (`onclick`, `onsubmit`, etc.) — all event binding is in the script block.
- Must produce valid HTML5 (`<!DOCTYPE html>`).
- The nonce must be cryptographically random and unique per invocation.
- All three pages exist in the DOM simultaneously; visibility is toggled via CSS (`display: none` / `display: block`).
- All buttons with cooldown must show a visual disabled state (grey/light-grey color) while locked.
- The send button must be disabled (grey) whenever `currentState !== entryNode` OR `status !== 'running'`, regardless of cooldown state.

## Edge Cases

- Condition: `showSessions` state arrives with an empty `workflows` array.
  Expected: The `workflow-select` dropdown is empty. The "Run" button remains clickable but will post a `launchSession` with an empty string (extension.ts handles validation).

- Condition: `showSessions` state arrives with an empty `sessions` array.
  Expected: The `session-list` container is empty. No session rows are rendered.

- Condition: `showDetail` state arrives with an empty `eventTypes` array.
  Expected: The `event-type-select` dropdown is empty. The send button is disabled (no event type to select).

- Condition: `showDetail` state arrives with an empty `events` array.
  Expected: The `event-list` container is empty. No event entries are rendered.

- Condition: `showDetail` state arrives with `currentState !== entryNode`.
  Expected: The send button is disabled with grey styling. Cooldown state is irrelevant.

- Condition: `showDetail` state arrives with `status !== 'running'` (e.g., `'completed'`).
  Expected: The send button is disabled with grey styling.

- Condition: User clicks the send button, cooldown activates, then a new `showDetail` arrives with `currentState !== entryNode`.
  Expected: After cooldown expires, the button remains disabled because the guard condition is not met.

- Condition: User clicks the send button, cooldown activates, then a new `showDetail` arrives where the guard condition is met.
  Expected: After cooldown expires, the button becomes enabled (both conditions satisfied).

- Condition: `showDetail` arrives while the send button is in cooldown.
  Expected: The guard state variables are updated. The button remains disabled during cooldown. When cooldown expires, the guard is re-evaluated with the latest state.

- Condition: User rapidly clicks the Run button before cooldown activates.
  Expected: Only the first click fires `postMessage` (button is disabled immediately on first click). Second click is blocked.

- Condition: A session row has `status === 'completed'` or `status === 'failed'`.
  Expected: No stop button is rendered for that row.

- Condition: A session row has `status === 'initializing'`.
  Expected: No stop button is rendered (only `'running'` shows the stop button).

- Condition: The sidebar is narrowed such that the session label text cannot fit alongside the stop button.
  Expected: The session label text truncates with an ellipsis ("..."). The stop button remains fully visible and clickable at its fixed size. No content overflows or overlaps the row boundary.

- Condition: The sidebar is narrowed such that the workflow dropdown row has minimal space.
  Expected: The dropdown shrinks (respecting `min-width: 0`) while the Run button maintains its fixed size. The dropdown text may be clipped by the browser's native select rendering.

## Related

- [SpectraViewProvider](./spectraViewProvider.md) — Calls this function to set `webview.html` during `resolveWebviewView`.
- [SessionListController](../controllers/sessionListController.md) — Defines `SessionListState` shape pushed to the sessions page.
- [SessionDetailController](../controllers/sessionDetailController.md) — Defines `SessionDetailState` shape pushed to the detail page.
