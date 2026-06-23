/**
 * SpectraViewProvider — implements vscode.WebviewViewProvider to provide a
 * WebviewView in the Activity Bar sidebar.
 *
 * Logic spec: spec/logic/vscode/src/views/spectraViewProvider.md
 *
 * Acts as a thin transport layer that posts state to the webview and forwards
 * incoming messages to subscribers. Does not own business logic, state
 * computation, or UI rendering.
 */
import { getWebviewContent as defaultGetWebviewContent } from "./getWebviewContent";

/**
 * Minimal Disposable interface matching vscode.Disposable.
 */
export interface Disposable {
  dispose(): void;
}

/**
 * Minimal Event type matching vscode.Event<T>.
 */
export type Event<T> = (listener: (e: T) => void) => Disposable;

/**
 * Minimal EventEmitter interface matching vscode.EventEmitter<T>.
 */
export interface IEventEmitter<T> {
  event: Event<T>;
  fire(data: T): void;
  dispose(): void;
}

/**
 * Logger interface required by SpectraViewProvider.
 */
export interface ViewProviderLogger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * Minimal Webview interface.
 */
export interface IWebview {
  cspSource: string;
  html: string;
  options: any;
  postMessage(message: any): PromiseLike<boolean>;
  onDidReceiveMessage: (listener: (e: any) => void) => Disposable;
}

/**
 * Minimal WebviewView interface.
 */
export interface IWebviewView {
  webview: IWebview;
  onDidDispose: (listener: () => void) => Disposable;
}

/**
 * Minimal Uri interface.
 */
export interface IUri {
  fsPath: string;
  scheme: string;
  path: string;
}

/**
 * Minimal WebviewViewResolveContext interface.
 */
export interface IWebviewViewResolveContext {
  state: any;
}

/**
 * Minimal CancellationToken interface.
 */
export interface ICancellationToken {
  isCancellationRequested: boolean;
}

/**
 * State for the sessions list page.
 */
export interface SessionListState {
  sessions: any[];
  workflows: string[];
}

/**
 * State for the session detail page.
 */
export interface SessionDetailState {
  sessionId: string;
  workflowName: string;
  entryNode: string;
  currentState: string;
  status: string;
  pid: number;
  eventTypes: string[];
  events: any[];
}

/**
 * Injectable dependencies for SpectraViewProvider.
 * Enables testing without real vscode APIs.
 */
export interface SpectraViewProviderDeps {
  getWebviewContent(webview: IWebview, extensionUri: IUri): string;
}

/**
 * Simple EventEmitter implementation for use without vscode runtime.
 */
class SimpleEventEmitter<T> implements IEventEmitter<T> {
  private _listeners: Array<(e: T) => void> = [];
  private _disposed = false;

  public readonly event: Event<T> = (listener: (e: T) => void): Disposable => {
    if (this._disposed) {
      return { dispose: () => {} };
    }
    this._listeners.push(listener);
    return {
      dispose: () => {
        const idx = this._listeners.indexOf(listener);
        if (idx >= 0) {
          this._listeners.splice(idx, 1);
        }
      },
    };
  };

  fire(data: T): void {
    if (this._disposed) {
      return;
    }
    for (const listener of [...this._listeners]) {
      listener(data);
    }
  }

  dispose(): void {
    this._disposed = true;
    this._listeners = [];
  }
}

/**
 * Implements vscode.WebviewViewProvider to provide a WebviewView in the
 * Activity Bar sidebar. Posts state to the webview and forwards incoming
 * messages to subscribers.
 */
export class SpectraViewProvider implements Disposable {
  private readonly _extensionUri: IUri;
  private readonly _logger: ViewProviderLogger;
  private readonly _deps: SpectraViewProviderDeps;
  private readonly _messageEmitter: IEventEmitter<any>;
  private _view: IWebviewView | null = null;
  private _currentPage: "sessions" | "detail" | "notInitialized" = "sessions";
  private _pendingMessage: any | null = null;

  /** Public event: fires when the webview posts a message to the extension. */
  public readonly onDidReceiveMessage: Event<any>;

  /**
   * Constructs a SpectraViewProvider.
   *
   * @param extensionUri - Extension root URI.
   * @param logger - Logger instance.
   * @param deps - Optional injectable dependencies for testing.
   */
  constructor(
    extensionUri: IUri,
    logger: ViewProviderLogger,
    deps?: SpectraViewProviderDeps,
  ) {
    this._extensionUri = extensionUri;
    this._logger = logger;
    this._deps = deps || {
      getWebviewContent: (webview, uri) =>
        defaultGetWebviewContent(webview as any, uri as any),
    };

    // Step 2: Create EventEmitter and expose event
    this._messageEmitter = new SimpleEventEmitter<any>();
    this.onDidReceiveMessage = this._messageEmitter.event;
  }

  /**
   * Called by VS Code when the webview view is resolved (made visible).
   *
   * @param webviewView - The webview view provided by VS Code.
   * @param context - The resolve context.
   * @param token - Cancellation token.
   */
  resolveWebviewView(
    webviewView: IWebviewView,
    context: IWebviewViewResolveContext,
    token: ICancellationToken,
  ): void {
    // Step 5: Store view reference
    this._view = webviewView;

    // Step 6: Configure webview options
    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this._extensionUri],
    };

    // Step 7: Set HTML content
    webviewView.webview.html = this._deps.getWebviewContent(
      webviewView.webview,
      this._extensionUri,
    );

    // Step 8: Subscribe to webview messages
    webviewView.webview.onDidReceiveMessage((msg: any) => {
      this._messageEmitter.fire(msg);
    });

    // Step 9: Subscribe to view dispose
    webviewView.onDidDispose(() => {
      this._view = null;
      this._logger.info("SpectraViewProvider: view disposed");
    });

    // Step 10: Deliver pending message if any
    if (this._pendingMessage !== null) {
      webviewView.webview.postMessage(this._pendingMessage);
      this._pendingMessage = null;
    }

    // Step 11: Log view resolution
    this._logger.info("SpectraViewProvider: view resolved");
  }

  /**
   * Posts the session list state to the webview.
   */
  showSessionList(state: SessionListState): void {
    // Step 12: Set current page
    this._currentPage = "sessions";
    // Step 13: Construct message
    const message = { type: "showSessions", state };
    // Step 14: If view is null, store as pending
    if (this._view === null) {
      this._pendingMessage = message;
      return;
    }
    // Step 15: Post message
    this._view.webview.postMessage(message);
  }

  /**
   * Posts the session detail state to the webview.
   */
  showSessionDetail(state: SessionDetailState): void {
    // Step 16: Set current page
    this._currentPage = "detail";
    // Step 17: Construct message
    const message = { type: "showDetail", state };
    // Step 18: If view is null, store as pending
    if (this._view === null) {
      this._pendingMessage = message;
      return;
    }
    // Step 19: Post message
    this._view.webview.postMessage(message);
  }

  /**
   * Posts the not-initialized message to the webview.
   */
  showNotInitialized(): void {
    // Step 20: Set current page
    this._currentPage = "notInitialized";
    // Step 21: Construct message
    const message = { type: "showNotInitialized" };
    // Step 22: If view is null, store as pending
    if (this._view === null) {
      this._pendingMessage = message;
      return;
    }
    // Step 23: Post message
    this._view.webview.postMessage(message);
  }

  /**
   * Posts the send result to the webview.
   * Ephemeral — does not store as pendingMessage when view is null.
   */
  postSendResult(success: boolean): void {
    const message = { type: "sendResult", success };
    if (this._view === null) {
      return;
    }
    this._view.webview.postMessage(message);
  }

  /**
   * Disposes the SpectraViewProvider.
   */
  dispose(): void {
    // Step 24: Dispose the EventEmitter
    this._messageEmitter.dispose();
    // Step 25: Set view to null
    this._view = null;
    // Step 26: Set pendingMessage to null
    this._pendingMessage = null;
  }
}
