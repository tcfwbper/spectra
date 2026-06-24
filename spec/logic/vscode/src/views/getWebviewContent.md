# getWebviewContent

## Overview

Generates the complete HTML string for the Spectra sidebar webview. Produces a self-contained document with inline CSS and JavaScript that renders three pages (not-initialized notice, sessions list, and session detail), handles client-side page routing via `window.addEventListener('message', ...)`, and communicates with the extension host via `vscode.postMessage(...)`. Does not perform any I/O beyond what the VS Code webview API provides.

## Boundaries

- Owns: generating a CSP-compliant HTML document, embedding inline styles and scripts, rendering the DOM structure for all three pages (not-initialized, sessions list, session detail), implementing client-side message handling and page switching, implementing button cooldown (2-second lock) logic, implementing the send-button guard (entryNode + running), and wiring all `vscode.postMessage` calls.
- Delegates: actual state data provision to the extension host (received via `postMessage`).
- Delegates: view lifecycle management to SpectraViewProvider.
- Must not: perform any filesystem I/O.
- Must not: fetch external resources (all content is inline; no external fonts or CDN resources).
- Must not: inject raw user data into HTML strings (all dynamic content is rendered via DOM manipulation in the embedded JS, not via string interpolation).
- Must not: use `eval()` or inline event handlers (`onclick` attributes) — all event binding is done in the `<script>` block.

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| `vscode.Webview` | Webview reference | `webview.cspSource` (for CSP header) | Must not call `postMessage` or subscribe to events |
| `vscode.Uri` | Extension URI | Used for `localResourceRoots` context | — |
| `crypto` (Node.js) | Nonce generation | `randomBytes` or equivalent for CSP nonce | — |

Construction constraint: This is a standalone exported function, not a class. Signature: `getWebviewContent(webview: vscode.Webview, extensionUri: vscode.Uri): string`.

## Behavior

### HTML Structure

1. Generates a random nonce (16+ bytes, hex-encoded) for the Content Security Policy.
2. Produces `<!DOCTYPE html>` with a `<meta>` CSP tag: `default-src 'none'; style-src 'nonce-${nonce}'; script-src 'nonce-${nonce}';`.
3. Embeds a single `<style nonce="${nonce}">` block with all CSS.
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
   - The stop button is rendered as a circular icon button with a pulsing animation to convey "running" state:
     - Shape: a small circle (20×20px) with a semi-transparent accent background (`var(--vscode-progressBar-background)` at 20% opacity).
     - Inner icon: a centered 8×8px square (solid `var(--vscode-progressBar-background)` color) representing the stop symbol.
     - Pulse animation: a CSS `@keyframes pulse` animation that smoothly scales the circle between 1.0 and 1.15 (period: 2s, infinite loop, ease-in-out). Conveys that the session is actively running.
     - On hover: the circle background opacity increases to 40%, and the pulse animation pauses (providing a clear "clickable" affordance).
     - On click: the pulse animation stops, the button enters cooldown.
     - No text label, no codicon dependency for this button.
   - The stop button triggers `terminateSession` with the session's `pid`.
   - The stop button has a 2-second cooldown after each click (disabled state: pulse animation stops, circle and inner square become grey/faded, hover effect suppressed).
   - The stop button is only rendered for sessions with `status === 'running'`.
   - Clicking anywhere on the row (except the stop button) triggers `navigateToDetail` with `sessionId` and `workflowName`.

### Session Detail Page DOM (id: `page-detail`)

The detail page uses a full-height flex column layout (`display: flex; flex-direction: column; height: 100%;`). The event history fills all available vertical space, and the input controls are pinned to the bottom.

8. A back button (id: `btn-back`) at the top-left, rendered as a 28×28px square icon-only button containing an inline SVG chevron-left icon. The SVG uses `currentColor` for stroke so it inherits the text color from the VS Code theme. The button uses `display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; flex-shrink: 0;` — it does NOT stretch to fill the row. No text label. On hover: subtle background highlight (`var(--vscode-toolbar-hoverBackground)`) with `border-radius: 4px`. Triggers `navigateToList`.

9. Below the back button: an event history container (id: `event-list`) styled as a chat-style conversation view.
   - The container uses `flex: 1; overflow-y: auto;` (fills remaining vertical space, scrolls when content overflows).
   - Each event entry is rendered as a chat bubble:
     - Alignment: events where `EmittedBy` equals the session's `entryNode` value are right-aligned (user's own messages); all other events (agent-emitted) are left-aligned. The `entryNode` value is available from the `showDetail` state payload and stored in a module-level variable.
     - Above the bubble: a small label displaying the event `Type` field in muted/secondary color (`var(--vscode-descriptionForeground)`), font-size 11px.
     - Bubble styling: rounded corners (border-radius: 12px), padding 8px 12px, max-width 80% of container width.
     - Left-aligned bubbles use `var(--vscode-editorWidget-background)` background.
     - Right-aligned bubbles use `var(--vscode-button-background)` at 20% opacity (or a distinguishable secondary color).
     - Message text inside the bubble uses `word-wrap: break-word; white-space: pre-wrap; overflow-wrap: break-word;` to ensure long text wraps correctly and preserves line breaks.
     - Text color: `var(--vscode-editor-foreground)` for readability.
     - Bubbles have vertical spacing (margin-bottom: 8px) between each entry.
   - The container auto-scrolls to the bottom when new events arrive (on each `showDetail` re-render).

10. Below the event history (pinned to bottom): a vertical stack (id: `detail-controls`) containing the event controls.
    - The container has `margin-top: 8px;` to provide visual separation between the last message bubble and the input controls.
    - The container has `padding-right: 8px;` to provide consistent right-side spacing that aligns its content's right edge with the rest of the sidebar content.
    - First row: a flex row (`display: flex; align-items: center; gap: 8px;`) containing:
      - An event-type `<select>` dropdown (id: `event-type-select`) using `flex: 1; min-width: 0;` (fills remaining horizontal space, shrinks gracefully).
      - A "Send" button (id: `btn-send`) with fixed width (`flex-shrink: 0; width: auto; padding: 4px 12px;`), right-aligned by flex layout.
      - The entire row adapts to sidebar width: dropdown stretches/shrinks, Send button stays fixed.
    - Second row (below the first row, margin-top: 8px): a `<textarea>` (id: `event-message-input`) for multi-line message input.
      - Height: 3 rows (approximately 72px, set via `rows="3"` attribute and matching CSS `height`).
      - Width: 100% of the container (respects the container's padding-right, so its right edge aligns with the Send button's right edge).
      - CSS: `resize: vertical;` (user may drag to enlarge vertically), `white-space: pre-wrap; word-wrap: break-word;`.
    - The dropdown is populated dynamically when `showDetail` state arrives (from `eventTypes`).
    - The "Send" button triggers `sendEvent` with the selected `eventType` and the textarea text.
    - The "Send" button has a 2-second cooldown after each click (disabled + grey styling during cooldown).
    - The "Send" button has an additional guard: it is enabled only when `currentState === entryNode` AND `status === 'running'`. Otherwise it is disabled with grey styling.
    - Both guards must pass simultaneously for the button to be clickable: no active cooldown AND the state condition is met.

### Client-Side JavaScript Behavior

11. Acquires VS Code API via `const vscode = acquireVsCodeApi()`.
12. Registers `window.addEventListener('message', handler)` to receive messages from the extension host.
13a. On receiving `{ type: 'showNotInitialized' }`:
    - Adds `.hidden` class to `page-sessions` and `page-detail`; removes `.hidden` class from `page-not-initialized`.
13. On receiving `{ type: 'showSessions', state }`:
    - Adds `.hidden` class to `page-detail` and `page-not-initialized`; removes `.hidden` class from `page-sessions`.
    - Populates the `workflow-select` dropdown with `state.workflows`.
    - Clears and rebuilds the `session-list` container from `state.sessions`.
    - Re-evaluates stop button visibility (only for `status === 'running'` sessions).
14. On receiving `{ type: 'showDetail', state }`:
    - Adds `.hidden` class to `page-sessions` and `page-not-initialized`; removes `.hidden` class from `page-detail`.
    - Stores `state.entryNode`, `state.currentState`, and `state.status` in module-level variables for the send-button guard.
    - Populates the `event-type-select` dropdown with `state.eventTypes`.
    - Clears and rebuilds the `event-list` container from `state.events` as chat bubbles:
      - For each event: creates a wrapper div with appropriate alignment class (right-aligned if `event.EmittedBy === entryNode`, left-aligned otherwise).
      - Renders the `event.Type` label above the bubble using `textContent` (never innerHTML).
      - Renders the `event.Message` text inside the bubble using `textContent` (preserves text safely; CSS `white-space: pre-wrap` handles line breaks).
      - After rebuilding, scrolls the `event-list` container to the bottom (`scrollTop = scrollHeight`).
    - Re-evaluates the send button's enabled/disabled state based on the guard condition.
14a. On receiving `{ type: 'sendResult', success }`:
    - If `success` is `true`: clears the `event-message-input` textarea value (sets to empty string).
    - If `success` is `false`: does nothing (textarea retains the user's message so they can retry).

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

- Must include a Content Security Policy meta tag with `default-src 'none'`, nonce-gated `style-src` and `script-src`. No `font-src` directive is needed (no external fonts are loaded; icons are inline SVG).
- Must never inject dynamic data (user content, session IDs, messages) via string interpolation into the HTML template — all dynamic rendering happens via DOM manipulation in the embedded JS after `message` events.
- Must not use inline event handlers (`onclick`, `onsubmit`, etc.) — all event binding is in the script block.
- Must produce valid HTML5 (`<!DOCTYPE html>`).
- The nonce must be cryptographically random and unique per invocation.
- All three pages exist in the DOM simultaneously; visibility is toggled via a `.hidden` CSS class that applies `display: none !important`. Removing the class restores the element's intrinsic display mode (e.g., `flex` for the detail page). Each page element sets its own layout display in its base CSS rule; the `.hidden` class overrides it when applied.
- All buttons with cooldown must show a visual disabled state (grey/light-grey color) while locked.
- The send button must be disabled (grey) whenever `currentState !== entryNode` OR `status !== 'running'`, regardless of cooldown state.
- Event history bubbles must use `textContent` for rendering message text — never `innerHTML` — to prevent XSS and ensure all characters display correctly.
- Event history bubbles must apply `word-wrap: break-word; white-space: pre-wrap; overflow-wrap: break-word;` to ensure long unbroken strings wrap and line breaks are preserved.
- The stop button pulse animation must use CSS `@keyframes` (not JavaScript timers) for performance and battery efficiency.
- The textarea must only be cleared upon receiving a `sendResult` message with `success: true` — never on button click alone.
- The `detail-controls` container must apply `margin-top: 8px` to visually separate the message bubbles from the input controls.
- The `detail-controls` container must apply `padding-right: 8px` so the textarea's right edge aligns with the Send button's right edge.

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

- Condition: An event message contains very long text without spaces (e.g., a URL or file path).
  Expected: The chat bubble wraps the text using `overflow-wrap: break-word`. The bubble does not overflow its max-width or the container boundary.

- Condition: An event message contains newline characters.
  Expected: Line breaks are preserved and rendered visually due to `white-space: pre-wrap`. No manual `<br>` injection is needed.

- Condition: The event history has many messages that exceed the visible area.
  Expected: The container scrolls vertically. On each `showDetail` re-render, the container auto-scrolls to the bottom to show the most recent message.

- Condition: The sidebar is narrowed such that the event-type dropdown and Send button row has minimal space.
  Expected: The event-type dropdown shrinks (respecting `flex: 1; min-width: 0;`) while the Send button maintains its fixed width. The row does not overflow.

- Condition: `sendResult` with `success: true` arrives.
  Expected: The textarea is cleared. The event-type dropdown retains its current selection.

- Condition: `sendResult` with `success: false` arrives.
  Expected: The textarea retains its content. The user can retry sending without re-typing.

- Condition: `sendResult` arrives while the user has already started typing a new message (race between fast typing and async result).
  Expected: On `success: true`, the textarea is still cleared (the new partial input is lost). This is acceptable because the prior send succeeded and the user's new typing occurred during the brief async gap.

## Related

- [SpectraViewProvider](./spectraViewProvider.md) — Calls this function to set `webview.html` during `resolveWebviewView`.
- [SessionListController](../controllers/sessionListController.md) — Defines `SessionListState` shape pushed to the sessions page.
- [SessionDetailController](../controllers/sessionDetailController.md) — Defines `SessionDetailState` shape pushed to the detail page.
