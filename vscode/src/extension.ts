/**
 * Extension entry point — activation and deactivation for the Spectra VS Code extension.
 *
 * Logic spec: spec/logic/vscode/src/extension.md
 *
 * Acts purely as a composition root and message router. Does not own business logic,
 * state computation, or I/O.
 *
 * - Owns: creating the logger, resolving project root, checking project initialization,
 *   constructing controllers and view provider, wiring onDidUpdate/onDidError/onDidReceiveMessage
 *   subscriptions, caching last-known SessionListState, routing webview messages,
 *   registering SpectraViewProvider with VS Code, and pushing all disposables to
 *   context.subscriptions.
 */

import { SessionListController } from "./controllers/sessionListController";
import { SessionDetailController } from "./controllers/sessionDetailController";
import { SpectraViewProvider } from "./views/spectraViewProvider";
import { ProjectRootResolver } from "./services/projectRootResolver";
import { SessionScanner } from "./services/sessionScanner";
import { WorkflowScanner } from "./services/workflowScanner";
import { SessionWatcher } from "./services/sessionWatcher";
import { WorkflowWatcher } from "./services/workflowWatcher";
import { EventWatcher } from "./services/eventWatcher";
import { EventScanner } from "./services/eventScanner";
import { SessionLauncher } from "./services/sessionLauncher";
import { SessionTerminator } from "./services/sessionTerminator";
import { EventDispatcher } from "./services/eventDispatcher";
import { WorkflowDefinitionParser } from "./services/workflowDefinitionParser";

/**
 * Logger interface providing severity-tagged output.
 */
export interface Logger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * Minimal OutputChannel interface.
 */
export interface IOutputChannel {
  appendLine(value: string): void;
  dispose(): void;
}

/**
 * Minimal Disposable interface.
 */
export interface IDisposable {
  dispose(): void;
}

/**
 * Interface for the session list controller as seen by the extension entry point.
 */
export interface ISessionListController extends IDisposable {
  onDidUpdate: (listener: (state: any) => void) => IDisposable;
  onDidError: (listener: (err: any) => void) => IDisposable;
  launch(workflowName: string): void;
  terminate(pid: number): void;
}

/**
 * Interface for the session detail controller as seen by the extension entry point.
 */
export interface ISessionDetailController extends IDisposable {
  onDidUpdate: (listener: (state: any) => void) => IDisposable;
  onDidError: (listener: (err: any) => void) => IDisposable;
  open(sessionId: string, workflowName: string): void;
  sendEvent(eventType: string, message: string): Promise<boolean>;
}

/**
 * Interface for the panel as seen by the extension entry point.
 */
export interface IPanel extends IDisposable {
  showSessionList(state: any): void;
  showSessionDetail(state: any): void;
  onDidReceiveMessage: (listener: (msg: any) => void) => IDisposable;
  onDidDispose: (listener: () => void) => IDisposable;
}

/**
 * Interface for the view provider as seen by the extension entry point.
 */
export interface IViewProvider extends IDisposable {
  showSessionList(state: any): void;
  showSessionDetail(state: any): void;
  showNotInitialized(): void;
  postSendResult(success: boolean): void;
  onDidReceiveMessage: (listener: (msg: any) => void) => IDisposable;
}

/**
 * Injectable dependencies for the activate function.
 * All fields are optional — when omitted, production defaults are used.
 * In tests, individual fields can be replaced with stubs/mocks.
 *
 * The legacy required form (ExtensionDeps) is maintained for backward compatibility
 * with existing test infrastructure. New tests should use the optional ActivateDeps form.
 */
export interface ActivateDeps {
  /** Pre-constructed OutputChannel. When provided, createOutputChannel is not called. */
  outputChannel?: IOutputChannel;
  /** Factory to create an OutputChannel by name. */
  createOutputChannel?: (name: string) => IOutputChannel;
  /** Display an error message to the user. */
  showErrorMessage?: (msg: string) => void;
  /** Register a VS Code command. */
  registerCommand?: (id: string, handler: (...args: any[]) => any) => IDisposable;
  /** Resolve the project root directory. */
  resolveProjectRoot?: () => string | undefined;
  /** Check whether the project is initialized (.spectra/ exists). */
  isInitialized?: (projectRoot: string) => boolean;
  /** Factory to create a SessionListController. */
  createSessionListController?: (projectRoot: string, logger: Logger) => ISessionListController;
  /** Factory to create a SessionDetailController. */
  createSessionDetailController?: (projectRoot: string, logger: Logger) => ISessionDetailController;
  /** Factory to create a SpectraViewProvider. */
  createViewProvider?: (extensionUri: any, logger: Logger) => IViewProvider;
  /** Register a WebviewViewProvider with VS Code. */
  registerWebviewViewProvider?: (viewType: string, provider: IViewProvider, options: any) => IDisposable;
  /** Legacy: create or reveal panel. Kept for backward compatibility during transition. */
  createOrRevealPanel?: (context: any, extensionUri: any, logger: Logger) => IPanel;
}

/**
 * Legacy required-deps interface for backward compatibility with existing test helpers.
 * @deprecated Use ActivateDeps (all-optional) for new code.
 */
export interface ExtensionDeps {
  createOutputChannel(name: string): IOutputChannel;
  showErrorMessage(msg: string): void;
  registerCommand(id: string, handler: (...args: any[]) => any): IDisposable;
  resolveProjectRoot(): string | undefined;
  createSessionListController(projectRoot: string, logger: Logger): ISessionListController;
  createSessionDetailController(projectRoot: string, logger: Logger): ISessionDetailController;
  createOrRevealPanel(context: any, extensionUri: any, logger: Logger): IPanel;
}

/**
 * Minimal extension context interface.
 */
export interface IExtensionContext {
  subscriptions: IDisposable[];
  extensionUri: any;
}

/**
 * Creates a logger that wraps an OutputChannel with severity-tagged output.
 */
function createLogger(outputChannel: IOutputChannel): Logger {
  return {
    info(msg: string): void {
      outputChannel.appendLine(`[INFO] ${msg}`);
    },
    warn(msg: string): void {
      outputChannel.appendLine(`[WARN] ${msg}`);
    },
    error(msg: string): void {
      outputChannel.appendLine(`[ERROR] ${msg}`);
    },
  };
}

/**
 * Activates the Spectra extension.
 *
 * Accepts an optional `deps` parameter. When `deps` is undefined or partially
 * provided, production defaults are used for any missing fields (merge-with-defaults).
 * VS Code calls activate(context) with no second argument; tests may pass mock deps.
 *
 * @param context - The VS Code extension context.
 * @param deps - Injectable dependencies (optional; for testing).
 */
export function activate(context: IExtensionContext, deps: ActivateDeps | ExtensionDeps = {}): void {
  const resolvedDeps = deps;

  // Production defaults — lazily require vscode only when deps don't supply their own factories
  const hasCustomDeps = (resolvedDeps as ActivateDeps).outputChannel || (resolvedDeps as ActivateDeps).createOutputChannel || (resolvedDeps as ExtensionDeps).createOutputChannel;
  const vscode = hasCustomDeps ? undefined : require("vscode");

  // Step 1: Create or use provided OutputChannel named 'Spectra'
  const outputChannel: IOutputChannel =
    (resolvedDeps as ActivateDeps).outputChannel ??
    ((resolvedDeps as ActivateDeps).createOutputChannel ?? (resolvedDeps as ExtensionDeps).createOutputChannel ?? ((name: string) => vscode.window.createOutputChannel(name)))("Spectra");

  // Step 2: Wrap in logger adapter
  const logger = createLogger(outputChannel);

  // Step 3: Log activation start
  logger.info("Spectra extension activating...");

  // Step 4: Resolve project root
  const resolveProjectRoot = (resolvedDeps as ActivateDeps).resolveProjectRoot ?? (resolvedDeps as ExtensionDeps).resolveProjectRoot ?? (() => ProjectRootResolver.resolve(vscode.workspace));
  const projectRoot = resolveProjectRoot();

  // Resolve showErrorMessage
  const showErrorMessage = (resolvedDeps as ActivateDeps).showErrorMessage ?? (resolvedDeps as ExtensionDeps).showErrorMessage ?? ((msg: string) => vscode.window.showErrorMessage(msg));

  // Step 5: If projectRoot is undefined, show error and return early
  if (projectRoot === undefined) {
    showErrorMessage("Spectra: No workspace folder open.");
    logger.error("No workspace folder open. Activation aborted.");
    context.subscriptions.push(outputChannel);
    return;
  }

  // Resolve registerCommand
  const registerCommand = (resolvedDeps as ActivateDeps).registerCommand ?? (resolvedDeps as ExtensionDeps).registerCommand ?? ((id: string, handler: (...args: any[]) => any) => vscode.commands.registerCommand(id, handler));

  // Production vscode deps helper for watchers
  function createVscodeWatcherDeps() {
    return {
      createFileSystemWatcher: (pattern: any) => vscode.workspace.createFileSystemWatcher(pattern),
      createRelativePattern: (base: string, pattern: string) => new vscode.RelativePattern(base, pattern),
      createEventEmitter: () => new vscode.EventEmitter(),
    };
  }

  function createVscodeConfigGetter() {
    return () => vscode.workspace.getConfiguration("spectra");
  }

  // Step 6: Create SessionListController
  const createSessionListController = (resolvedDeps as ActivateDeps).createSessionListController ?? (resolvedDeps as ExtensionDeps).createSessionListController ?? ((root: string, log: Logger) => {
    const watcherDeps = createVscodeWatcherDeps();
    return new SessionListController(root, log, {
      createSessionWatcher: (pr: string) => new SessionWatcher(pr, watcherDeps),
      createWorkflowWatcher: (pr: string) => new WorkflowWatcher(pr, watcherDeps),
      scanSessions: (pr: string, l: any) => SessionScanner.scanSessions(pr, l),
      scanWorkflows: (pr: string, l: any) => WorkflowScanner.scanWorkflows(pr, l),
      launch: (wf: string, pr: string, l: any) => SessionLauncher.launch(wf, pr, l, {
        getConfiguration: createVscodeConfigGetter(),
        spawn: require("child_process").spawn,
        randomUUID: require("crypto").randomUUID,
      }),
      terminate: (pid: number, l: any) => SessionTerminator.terminate(pid, l, {
        getConfiguration: createVscodeConfigGetter(),
        processKill: (p: number, s: string | number) => process.kill(p, s),
        execFile: require("child_process").execFile,
      }),
      createStateEmitter: () => new vscode.EventEmitter(),
      createErrorEmitter: () => new vscode.EventEmitter(),
    });
  });
  const sessionListController = createSessionListController(projectRoot, logger);

  // Step 7: Create SessionDetailController
  const createSessionDetailController = (resolvedDeps as ActivateDeps).createSessionDetailController ?? (resolvedDeps as ExtensionDeps).createSessionDetailController ?? ((root: string, log: Logger) => {
    const watcherDeps = createVscodeWatcherDeps();
    return new SessionDetailController(root, log, {
      createEventWatcher: (pr: string, sid: string) => new EventWatcher(pr, sid, watcherDeps),
      scanEvents: (pr: string, sid: string, l: any) => EventScanner.scanEvents(pr, sid, l),
      scanSessions: (pr: string, l: any) => SessionScanner.scanSessions(pr, l),
      parseWorkflowDefinition: (pr: string, wf: string, l: any) => WorkflowDefinitionParser.parseWorkflowDefinition(pr, wf, l),
      dispatchEvent: (et: string, sid: string, msg: string, pr: string, l: any) => EventDispatcher.dispatch(et, sid, msg, pr, l, {
        getConfiguration: createVscodeConfigGetter(),
        execFile: require("child_process").execFile,
      }),
      createStateEmitter: () => new vscode.EventEmitter(),
      createErrorEmitter: () => new vscode.EventEmitter(),
    });
  });
  const sessionDetailController = createSessionDetailController(projectRoot, logger);

  // Step 8: Create ViewProvider and register for sidebar
  const createViewProvider = (resolvedDeps as ActivateDeps).createViewProvider ?? ((uri: any, log: Logger) => new SpectraViewProvider(uri, log));
  const viewProvider = createViewProvider(context.extensionUri, logger);

  const registerWebviewViewProvider = (resolvedDeps as ActivateDeps).registerWebviewViewProvider ?? ((viewType: string, provider: any, options: any) => vscode.window.registerWebviewViewProvider(viewType, provider, options));
  const viewProviderRegistration = registerWebviewViewProvider("spectra.chatView", viewProvider, { webviewOptions: { retainContextWhenHidden: true } });

  // Step 9: Initialize cached session list state
  let cachedSessionListState: any = null;

  // Step 9a: Initialize activePage tracking
  let activePage: "sessions" | "detail" = "sessions";

  // Step 10: Subscribe sessionListController.onDidUpdate
  const listUpdateSub = sessionListController.onDidUpdate((state: any) => {
    cachedSessionListState = state;
    if (activePage === "sessions") {
      viewProvider.showSessionList(state);
    }
  });

  // Step 11: Subscribe sessionDetailController.onDidUpdate
  const detailUpdateSub = sessionDetailController.onDidUpdate((state: any) => {
    viewProvider.showSessionDetail(state);
  });

  // Step 12: Subscribe sessionListController.onDidError
  const listErrorSub = sessionListController.onDidError((error: any) => {
    showErrorMessage(error.message);
  });

  // Step 13: Subscribe sessionDetailController.onDidError
  const detailErrorSub = sessionDetailController.onDidError((error: any) => {
    showErrorMessage(error.message);
  });

  // Step 14: Subscribe viewProvider.onDidReceiveMessage and route
  const messageSub = viewProvider.onDidReceiveMessage((msg: any) => {
    switch (msg.command) {
      case "navigateToDetail":
        activePage = "detail";
        sessionDetailController.open(msg.sessionId, msg.workflowName);
        break;
      case "navigateToList":
        activePage = "sessions";
        if (cachedSessionListState !== null) {
          viewProvider.showSessionList(cachedSessionListState);
        }
        break;
      case "launchSession":
        sessionListController.launch(msg.workflowName);
        break;
      case "terminateSession":
        sessionListController.terminate(msg.pid);
        break;
      case "sendEvent":
        sessionDetailController.sendEvent(msg.eventType, msg.message).then((result: boolean) => {
          viewProvider.postSendResult(result);
        });
        break;
      default:
        logger.warn(`Unrecognized webview command: ${msg.command}`);
        break;
    }
  });

  // Step 15: Register spectra.openPanel command
  const commandDisposable = registerCommand("spectra.openPanel", () => {
    // no-op — view is managed by VS Code sidebar
  });

  // Step 16: Push all disposables to context.subscriptions
  context.subscriptions.push(outputChannel);
  context.subscriptions.push(sessionListController);
  context.subscriptions.push(sessionDetailController);
  context.subscriptions.push(viewProvider);
  context.subscriptions.push(viewProviderRegistration);
  context.subscriptions.push(commandDisposable);
  context.subscriptions.push(listUpdateSub);
  context.subscriptions.push(detailUpdateSub);
  context.subscriptions.push(listErrorSub);
  context.subscriptions.push(detailErrorSub);
  context.subscriptions.push(messageSub);

  // Step 17: Log successful activation
  logger.info(`Spectra extension activated successfully. Project root: ${projectRoot}`);
}

/**
 * Deactivates the Spectra extension.
 * Empty function body — all cleanup is handled by context.subscriptions disposal.
 */
export function deactivate(): void {
  // Empty function body.
}
