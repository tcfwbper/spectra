/**
 * Shared test helpers for stubbing VS Code Webview and WebviewPanel APIs.
 *
 * Used by getWebviewContent.test.ts and spectraPanel.test.ts to create
 * consistent mock objects for the webview panel lifecycle.
 *
 * Scaffolded: depends on the production views surface
 *   (vscode/src/views/getWebviewContent.ts, vscode/src/views/spectraPanel.ts)
 *   being created before tests can compile and run.
 */
import * as sinon from "sinon";

// ─── Stub Webview ────────────────────────────────────────────────────────────

/**
 * Minimal stub of vscode.Webview for getWebviewContent tests.
 */
export interface StubWebview {
  cspSource: string;
  postMessage: sinon.SinonStub;
  onDidReceiveMessage: sinon.SinonStub;
  asWebviewUri: sinon.SinonStub;
  html: string;
  options: any;
}

/**
 * Default fake URI returned by asWebviewUri when no custom implementation is given.
 */
export const FAKE_CODICON_WEBVIEW_URI =
  "https://file+.vscode-resource.vscode-cdn.net/test/extension/node_modules/@vscode/codicons/dist/codicon.css";

/**
 * Creates a stub vscode.Webview with configurable cspSource.
 * The `asWebviewUri` stub returns a predictable local URI by default.
 */
export function createStubWebview(cspSource = "https://test.csp"): StubWebview {
  const asWebviewUri = sinon.stub().callsFake((uri: any) => {
    // Return a vscode-resource style URI based on the input path
    const path = uri?.path || uri?.fsPath || "/unknown";
    return {
      toString: () => `https://file+.vscode-resource.vscode-cdn.net${path}`,
    };
  });

  return {
    cspSource,
    postMessage: sinon.stub().resolves(true),
    onDidReceiveMessage: sinon.stub(),
    asWebviewUri,
    html: "",
    options: {},
  };
}

// ─── Stub Uri ────────────────────────────────────────────────────────────────

/**
 * Minimal stub of vscode.Uri for extensionUri parameters.
 */
export interface StubUri {
  fsPath: string;
  scheme: string;
  path: string;
  with: (change: { path: string }) => StubUri;
}

/**
 * Creates a stub vscode.Uri for the extension root.
 * Supports the `with({ path })` method used by `Uri.joinPath` patterns.
 */
export function createStubExtensionUri(fsPath = "/test/extension"): StubUri {
  const uri: StubUri = {
    fsPath,
    scheme: "file",
    path: fsPath,
    with(change: { path: string }) {
      return createStubExtensionUri(change.path);
    },
  };
  return uri;
}

/**
 * Creates a stub vscode.Uri.joinPath implementation for tests that need
 * to verify URI construction from extensionUri (e.g., codicon font path).
 * Returns a new StubUri with the joined path segments appended.
 */
export function stubUriJoinPath(
  base: StubUri,
  ...pathSegments: string[]
): StubUri {
  const joined = base.path + "/" + pathSegments.join("/");
  return createStubExtensionUri(joined);
}

// ─── Stub WebviewPanel ───────────────────────────────────────────────────────

/**
 * Callback type for panel event listeners.
 */
type DisposableListener<T> = (e: T) => void;

/**
 * Disposable registration result.
 */
interface StubDisposable {
  dispose: () => void;
}

/**
 * Minimal stub of vscode.WebviewPanel for SpectraPanel tests.
 */
export interface StubWebviewPanel {
  webview: StubWebview;
  reveal: sinon.SinonStub;
  dispose: sinon.SinonStub;
  onDidDispose: (listener: () => void) => StubDisposable;
  /** Test utility: trigger the onDidDispose callback. */
  triggerDispose: () => void;
  /** Test utility: trigger onDidReceiveMessage callback with a message. */
  triggerMessage: (msg: any) => void;
  /** All registered dispose listeners (test inspection). */
  disposeListeners: Array<() => void>;
  /** All registered message listeners (test inspection). */
  messageListeners: Array<DisposableListener<any>>;
}

/**
 * Creates a stub WebviewPanel with controllable event triggers.
 */
export function createStubWebviewPanel(
  cspSource = "https://test.csp",
): StubWebviewPanel {
  const disposeListeners: Array<() => void> = [];
  const messageListeners: Array<DisposableListener<any>> = [];

  const webview = createStubWebview(cspSource);

  // Wire onDidReceiveMessage to register listeners
  webview.onDidReceiveMessage = sinon
    .stub()
    .callsFake((listener: DisposableListener<any>) => {
      messageListeners.push(listener);
      return { dispose: () => {} };
    });

  return {
    webview,
    reveal: sinon.stub(),
    dispose: sinon.stub(),
    onDidDispose: (listener: () => void) => {
      disposeListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerDispose: () => {
      for (const l of [...disposeListeners]) {
        l();
      }
    },
    triggerMessage: (msg: any) => {
      for (const l of [...messageListeners]) {
        l(msg);
      }
    },
    disposeListeners,
    messageListeners,
  };
}

// ─── Stub Extension Context ──────────────────────────────────────────────────

/**
 * Minimal stub of vscode.ExtensionContext for panel tests.
 */
export interface StubExtensionContext {
  subscriptions: any[];
}

/**
 * Creates a stub ExtensionContext with an empty subscriptions array.
 */
export function createStubExtensionContext(): StubExtensionContext {
  return {
    subscriptions: [],
  };
}

// ─── Mock Logger ─────────────────────────────────────────────────────────────

/**
 * Logger interface matching the shape required by SpectraPanel.
 */
export interface MockPanelLogger {
  info: sinon.SinonSpy;
  warn: sinon.SinonSpy;
  error: sinon.SinonSpy;
}

/**
 * Creates a mock logger for panel tests.
 */
export function createMockPanelLogger(): MockPanelLogger {
  return {
    info: sinon.spy(),
    warn: sinon.spy(),
    error: sinon.spy(),
  };
}

// ─── Nonce Extraction Utility ────────────────────────────────────────────────

/**
 * Extracts the nonce value from a CSP meta tag in the HTML string.
 * Returns null if no nonce is found.
 */
export function extractNonceFromCsp(html: string): string | null {
  const match = html.match(/nonce-([a-f0-9]+)/);
  return match ? match[1] : null;
}

/**
 * Extracts the nonce value from a style tag's nonce attribute.
 * Returns null if not found.
 */
export function extractNonceFromStyleTag(html: string): string | null {
  const match = html.match(/<style\s+nonce="([^"]+)"/);
  return match ? match[1] : null;
}

/**
 * Extracts the nonce value from a script tag's nonce attribute.
 * Returns null if not found.
 */
export function extractNonceFromScriptTag(html: string): string | null {
  const match = html.match(/<script\s+nonce="([^"]+)"/);
  return match ? match[1] : null;
}

/**
 * Extracts the font-src directive value from a CSP meta tag in the HTML string.
 * Returns null if no font-src is found.
 */
export function extractFontSrcFromCsp(html: string): string | null {
  const match = html.match(/font-src\s+([^;"]+)/);
  return match ? match[1].trim() : null;
}

/**
 * Checks whether the HTML includes a codicon font reference (link or style).
 * Returns the matched string segment or null.
 */
export function extractCodiconReference(html: string): string | null {
  // Look for a <link> with codicon in href, or an @font-face referencing codicon
  const linkMatch = html.match(/<link[^>]*codicon[^>]*>/i);
  if (linkMatch) return linkMatch[0];
  const fontFaceMatch = html.match(/@font-face[^}]*codicon[^}]*/i);
  if (fontFaceMatch) return fontFaceMatch[0];
  return null;
}

// ─── HTML Content Assertions ─────────────────────────────────────────────────

/**
 * All element IDs expected in the webview content, per the test spec.
 */
export const EXPECTED_ELEMENT_IDS = [
  "page-not-initialized",
  "page-sessions",
  "page-detail",
  "btn-run",
  "btn-back",
  "btn-send",
  "session-list",
  "event-list",
  "workflow-select",
  "event-type-select",
  "event-message-input",
] as const;

// ─── Stub WebviewView (for SpectraViewProvider tests) ───────────────────────

/**
 * Callback type for WebviewView event listeners.
 */
type ViewDisposableListener<T> = (e: T) => void;

/**
 * Minimal stub of vscode.WebviewView for SpectraViewProvider tests.
 */
export interface StubWebviewView {
  webview: StubWebview;
  onDidDispose: (listener: () => void) => { dispose: () => void };
  /** Test utility: trigger the onDidDispose callback. */
  triggerDispose: () => void;
  /** Test utility: trigger onDidReceiveMessage callback with a message. */
  triggerMessage: (msg: any) => void;
  /** All registered dispose listeners (test inspection). */
  disposeListeners: Array<() => void>;
  /** All registered message listeners (test inspection). */
  messageListeners: Array<ViewDisposableListener<any>>;
}

/**
 * Creates a stub WebviewView with controllable event triggers.
 */
export function createStubWebviewView(
  cspSource = "https://test.csp",
): StubWebviewView {
  const disposeListeners: Array<() => void> = [];
  const messageListeners: Array<ViewDisposableListener<any>> = [];

  const webview = createStubWebview(cspSource);

  // Wire onDidReceiveMessage to register listeners
  webview.onDidReceiveMessage = sinon
    .stub()
    .callsFake((listener: ViewDisposableListener<any>) => {
      messageListeners.push(listener);
      return { dispose: () => {} };
    });

  return {
    webview,
    onDidDispose: (listener: () => void) => {
      disposeListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerDispose: () => {
      for (const l of [...disposeListeners]) {
        l();
      }
    },
    triggerMessage: (msg: any) => {
      for (const l of [...messageListeners]) {
        l(msg);
      }
    },
    disposeListeners,
    messageListeners,
  };
}

/**
 * Minimal stub of vscode.WebviewViewResolveContext.
 */
export interface StubWebviewViewResolveContext {
  state: any;
}

/**
 * Creates a stub WebviewViewResolveContext.
 */
export function createStubWebviewViewResolveContext(): StubWebviewViewResolveContext {
  return { state: undefined };
}

/**
 * Minimal stub of vscode.CancellationToken.
 */
export interface StubCancellationToken {
  isCancellationRequested: boolean;
  onCancellationRequested: sinon.SinonStub;
}

/**
 * Creates a stub CancellationToken.
 */
export function createStubCancellationToken(): StubCancellationToken {
  return {
    isCancellationRequested: false,
    onCancellationRequested: sinon.stub(),
  };
}
