/**
 * getWebviewContent — generates the complete HTML string for the Spectra sidebar webview.
 *
 * Logic spec: spec/logic/vscode/src/views/getWebviewContent.md
 *
 * Produces a self-contained document with inline CSS and JavaScript that renders
 * three pages (not-initialized notice, sessions list, and session detail),
 * handles client-side page routing via window.addEventListener('message', ...),
 * and communicates with the extension host via vscode.postMessage(...).
 */
import * as crypto from "crypto";

/**
 * Minimal Webview interface for the subset of vscode.Webview used here.
 */
interface Webview {
  cspSource: string;
}

/**
 * Minimal Uri interface for the subset of vscode.Uri used here.
 */
interface Uri {
  fsPath: string;
  scheme: string;
  path: string;
}

/**
 * Generates a complete HTML5 document string with CSP-compliant inline CSS and
 * JavaScript for the Spectra webview. Contains two pages (sessions list and
 * session detail) controlled by postMessage from the extension host.
 *
 * @param webview - The VS Code webview instance (used for cspSource).
 * @param extensionUri - The extension root URI.
 * @returns A complete HTML document string ready to assign to webview.html.
 */
export function getWebviewContent(webview: Webview, extensionUri: Uri): string {
  const nonce = crypto.randomBytes(16).toString("hex");
  const cspSource = webview.cspSource;

  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'nonce-${nonce}'; script-src 'nonce-${nonce}';">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style nonce="${nonce}">
body {
  font-family: var(--vscode-font-family, sans-serif);
  padding: 0;
  margin: 0;
  color: var(--vscode-foreground, #ccc);
  background: var(--vscode-editor-background, #1e1e1e);
}
h1 {
  margin: 0;
  padding: 12px 16px;
  font-size: 1.2em;
}
.page {
  padding: 0 16px 16px;
}
.hidden {
  display: none;
}
.row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
}
select, input {
  padding: 4px 8px;
  background: var(--vscode-input-background, #3c3c3c);
  color: var(--vscode-input-foreground, #ccc);
  border: 1px solid var(--vscode-input-border, #555);
}
button {
  padding: 4px 12px;
  cursor: pointer;
  background: var(--vscode-button-background, #0e639c);
  color: var(--vscode-button-foreground, #fff);
  border: none;
}
button:disabled {
  background: #555;
  color: #999;
  cursor: default;
}
.session-row {
  padding: 8px;
  margin-bottom: 4px;
  cursor: pointer;
  border: 1px solid var(--vscode-panel-border, #444);
}
.session-row:hover {
  background: var(--vscode-list-hoverBackground, #2a2d2e);
}
.session-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.session-subtitle {
  font-size: 0.85em;
  opacity: 0.7;
  margin-top: 2px;
}
.event-entry {
  padding: 4px 0;
  font-size: 0.9em;
  border-bottom: 1px solid var(--vscode-panel-border, #333);
}
#btn-back {
  margin-bottom: 12px;
}
.not-initialized-text {
  color: var(--vscode-descriptionForeground, #999);
  text-align: center;
  padding: 32px 16px;
}
</style>
</head>
<body>
<h1>Spectra</h1>

<div id="page-not-initialized" class="page hidden">
  <div class="not-initialized-text">Please run <code>spectra init</code> to initialize the project.</div>
</div>

<div id="page-sessions" class="page">
  <div class="row">
    <select id="workflow-select"></select>
    <button id="btn-run">Run</button>
  </div>
  <div id="session-list"></div>
</div>

<div id="page-detail" class="page hidden">
  <button id="btn-back">&larr; Back</button>
  <div id="event-list"></div>
  <div class="row">
    <select id="event-type-select"></select>
    <input id="event-message-input" type="text" placeholder="Message">
    <button id="btn-send">Send</button>
  </div>
</div>

<script nonce="${nonce}">
(function() {
  const vscode = acquireVsCodeApi();

  const pageNotInitialized = document.getElementById('page-not-initialized');
  const pageSessions = document.getElementById('page-sessions');
  const pageDetail = document.getElementById('page-detail');
  const workflowSelect = document.getElementById('workflow-select');
  const btnRun = document.getElementById('btn-run');
  const sessionList = document.getElementById('session-list');
  const btnBack = document.getElementById('btn-back');
  const eventList = document.getElementById('event-list');
  const eventTypeSelect = document.getElementById('event-type-select');
  const eventMessageInput = document.getElementById('event-message-input');
  const btnSend = document.getElementById('btn-send');

  let entryNode = '';
  let currentState = '';
  let status = '';
  let sendCooldown = false;

  function applyCooldown(btn, durationMs) {
    btn.disabled = true;
    setTimeout(function() {
      if (btn === btnSend) {
        reevaluateSendButton();
      } else {
        btn.disabled = false;
      }
    }, durationMs);
  }

  function reevaluateSendButton() {
    sendCooldown = false;
    const guardMet = (currentState === entryNode && status === 'running');
    btnSend.disabled = !guardMet || sendCooldown;
  }

  function applySendCooldown() {
    sendCooldown = true;
    btnSend.disabled = true;
    setTimeout(function() {
      sendCooldown = false;
      reevaluateSendButton();
    }, 2000);
  }

  btnRun.addEventListener('click', function() {
    if (btnRun.disabled) return;
    const workflowName = workflowSelect.value || '';
    vscode.postMessage({ command: 'launchSession', workflowName: workflowName });
    applyCooldown(btnRun, 2000);
  });

  btnBack.addEventListener('click', function() {
    vscode.postMessage({ command: 'navigateToList' });
  });

  btnSend.addEventListener('click', function() {
    if (btnSend.disabled) return;
    const eventType = eventTypeSelect.value || '';
    const message = eventMessageInput.value || '';
    vscode.postMessage({ command: 'sendEvent', eventType: eventType, message: message });
    applySendCooldown();
  });

  window.addEventListener('message', function(event) {
    const msg = event.data;
    if (!msg) return;

    if (msg.type === 'showNotInitialized') {
      pageNotInitialized.classList.remove('hidden');
      pageSessions.classList.add('hidden');
      pageDetail.classList.add('hidden');
    }

    if (msg.type === 'showSessions') {
      pageSessions.classList.remove('hidden');
      pageNotInitialized.classList.add('hidden');
      pageDetail.classList.add('hidden');

      const state = msg.state || {};
      const workflows = state.workflows || [];
      const sessions = state.sessions || [];

      workflowSelect.innerHTML = '';
      workflows.forEach(function(wf) {
        const opt = document.createElement('option');
        opt.value = wf;
        opt.textContent = wf;
        workflowSelect.appendChild(opt);
      });

      sessionList.innerHTML = '';
      sessions.forEach(function(s) {
        const row = document.createElement('div');
        row.className = 'session-row';

        const titleLine = document.createElement('div');
        titleLine.className = 'session-title';

        const label = document.createElement('span');
        label.textContent = s.workflowName + '-' + (s.id || '').substring(0, 8);
        titleLine.appendChild(label);

        if (s.status === 'running') {
          const stopBtn = document.createElement('button');
          stopBtn.textContent = 'Stop';
          stopBtn.addEventListener('click', function(e) {
            e.stopPropagation();
            if (stopBtn.disabled) return;
            vscode.postMessage({ command: 'terminateSession', pid: s.pid });
            applyCooldown(stopBtn, 2000);
          });
          titleLine.appendChild(stopBtn);
        }

        const subtitle = document.createElement('div');
        subtitle.className = 'session-subtitle';
        subtitle.textContent = s.status || '';

        row.appendChild(titleLine);
        row.appendChild(subtitle);

        row.addEventListener('click', function() {
          vscode.postMessage({ command: 'navigateToDetail', sessionId: s.id, workflowName: s.workflowName });
        });

        sessionList.appendChild(row);
      });
    }

    if (msg.type === 'showDetail') {
      pageDetail.classList.remove('hidden');
      pageNotInitialized.classList.add('hidden');
      pageSessions.classList.add('hidden');

      const state = msg.state || {};
      entryNode = state.entryNode || '';
      currentState = state.currentState || '';
      status = state.status || '';

      const eventTypes = state.eventTypes || [];
      const events = state.events || [];

      eventTypeSelect.innerHTML = '';
      eventTypes.forEach(function(et) {
        const opt = document.createElement('option');
        opt.value = et;
        opt.textContent = et;
        eventTypeSelect.appendChild(opt);
      });

      eventList.innerHTML = '';
      events.forEach(function(ev) {
        const entry = document.createElement('div');
        entry.className = 'event-entry';
        entry.textContent = '[' + (ev.type || '') + '] ' + (ev.emittedBy || '') + ': ' + (ev.message || '');
        eventList.appendChild(entry);
      });

      if (!sendCooldown) {
        const guardMet = (currentState === entryNode && status === 'running');
        btnSend.disabled = !guardMet;
      }
    }
  });
})();
</script>
</body>
</html>`;
}
