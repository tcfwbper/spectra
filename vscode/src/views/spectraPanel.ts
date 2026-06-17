/**
 * SpectraPanel — manages the lifecycle of a single WebviewPanel (singleton).
 *
 * Logic spec: spec/logic/vscode/src/views/spectraPanel.md
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
 * Logger interface required by SpectraPanel.
 */
export interface SpectraPanelLogger {
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
  postMessage(message: any): PromiseLike<boolean>;
  onDidReceiveMessage: (listener: (e: any) => void) => Disposable;
}

/**
 * Minimal WebviewPanel interface.
 */
export interface IWebviewPanel {
  webview: IWebview;
  reveal(): void;
  dispose(): void;
  onDidDispose: (listener: () => void) => Disposable;
}

/**
 * Minimal ExtensionContext interface.
 */
export interface IExtensionContext {
  subscriptions: Disposable[];
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
 * Injectable dependencies for SpectraPanel.
 * Enables testing without real vscode APIs.
 */
export interface SpectraPanelDeps {
  createWebviewPanel(
    viewType: string,
    title: string,
    showOptions: number,
    options: {
      enableScripts: boolean;
      retainContextWhenHidden: boolean;
      localResourceRoots: IUri[];
    },
  ): IWebviewPanel;
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
 * Manages the lifecycle of a single WebviewPanel (singleton pattern) and
 * provides bidirectional message routing between the webview and extension host.
 */
export class SpectraPanel implements Disposable {
  private static _instance: SpectraPanel | null = null;
  private static _deps: SpectraPanelDeps | null = null;

  private readonly _panel: IWebviewPanel;
  private readonly _logger: SpectraPanelLogger;
  private readonly _messageEmitter: IEventEmitter<any>;
  private readonly _disposeEmitter: IEventEmitter<void>;
  private _disposed = false;
  private _currentPage: "sessions" | "detail" = "sessions";

  /** Public event: fires when the webview posts a message to the extension. */
  public readonly onDidReceiveMessage: Event<any>;
  /** Public event: fires when the panel is closed or disposed. */
  public readonly onDidDispose: Event<void>;

  private constructor(
    panel: IWebviewPanel,
    logger: SpectraPanelLogger,
  ) {
    this._panel = panel;
    this._logger = logger;

    this._messageEmitter = new SimpleEventEmitter<any>();
    this._disposeEmitter = new SimpleEventEmitter<void>();

    this.onDidReceiveMessage = this._messageEmitter.event;
    this.onDidDispose = this._disposeEmitter.event;

    // Subscribe to webview messages
    panel.webview.onDidReceiveMessage((msg: any) => {
      if (!this._disposed) {
        this._messageEmitter.fire(msg);
      }
    });

    // Subscribe to panel dispose
    panel.onDidDispose(() => {
      this._disposed = true;
      this._disposeEmitter.fire(undefined as any);
      this._messageEmitter.dispose();
      this._disposeEmitter.dispose();
      SpectraPanel._instance = null;
      this._logger.info("SpectraPanel disposed");
    });
  }

  /**
   * Creates or reveals the singleton SpectraPanel.
   *
   * @param context - Extension context for disposable registration.
   * @param extensionUri - Extension root URI.
   * @param logger - Logger instance.
   * @param deps - Optional injectable dependencies for testing.
   */
  static createOrReveal(
    context: IExtensionContext,
    extensionUri: IUri,
    logger: SpectraPanelLogger,
    deps?: SpectraPanelDeps,
  ): SpectraPanel {
    // Store deps for use during construction
    if (deps) {
      SpectraPanel._deps = deps;
    }

    // If a live instance already exists, reveal and return
    if (SpectraPanel._instance !== null) {
      SpectraPanel._instance._panel.reveal();
      return SpectraPanel._instance;
    }

    const resolvedDeps = SpectraPanel._deps;

    // Create the panel
    const panel = resolvedDeps
      ? resolvedDeps.createWebviewPanel("spectra", "Spectra", 1, {
          enableScripts: true,
          retainContextWhenHidden: true,
          localResourceRoots: [extensionUri],
        })
      : defaultCreateWebviewPanel("spectra", "Spectra", 1, {
          enableScripts: true,
          retainContextWhenHidden: true,
          localResourceRoots: [extensionUri],
        });

    // Set HTML content
    const html = resolvedDeps
      ? resolvedDeps.getWebviewContent(panel.webview, extensionUri)
      : defaultGetWebviewContent(panel.webview as any, extensionUri as any);
    panel.webview.html = html;

    // Create instance
    const instance = new SpectraPanel(panel, logger);

    // Store as singleton
    SpectraPanel._instance = instance;

    // Register with context
    context.subscriptions.push(instance);

    logger.info("SpectraPanel created");

    return instance;
  }

  /**
   * Resets the static singleton for testing purposes.
   */
  static _resetForTest(): void {
    SpectraPanel._instance = null;
    SpectraPanel._deps = null;
  }

  /**
   * Posts the session list state to the webview.
   */
  showSessionList(state: SessionListState): void {
    if (this._disposed) {
      return;
    }
    this._currentPage = "sessions";
    this._panel.webview.postMessage({ type: "showSessions", state });
  }

  /**
   * Posts the session detail state to the webview.
   */
  showSessionDetail(state: SessionDetailState): void {
    if (this._disposed) {
      return;
    }
    this._currentPage = "detail";
    this._panel.webview.postMessage({ type: "showDetail", state });
  }

  /**
   * Disposes the underlying panel.
   */
  dispose(): void {
    if (this._disposed) {
      return;
    }
    this._panel.dispose();
  }
}

/**
 * Default panel factory (placeholder for when no DI is provided).
 * In a real VS Code extension, this would call vscode.window.createWebviewPanel.
 */
function defaultCreateWebviewPanel(
  viewType: string,
  title: string,
  showOptions: number,
  options: {
    enableScripts: boolean;
    retainContextWhenHidden: boolean;
    localResourceRoots: IUri[];
  },
): IWebviewPanel {
  throw new Error(
    "SpectraPanel: no createWebviewPanel dependency provided. " +
      "In production, pass vscode.window.createWebviewPanel via deps.",
  );
}
