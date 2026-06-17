/**
 * Extension entry point — activation and deactivation for the Spectra VS Code extension.
 *
 * Logic spec: spec/logic/vscode/src/extension.md
 *
 * Acts purely as a composition root and message router. Does not own business logic,
 * state computation, or I/O.
 *
 * - Owns: creating the logger, resolving project root, constructing controllers and panel,
 *   wiring onDidUpdate/onDidError/onDidReceiveMessage/onDidDispose subscriptions,
 *   caching last-known SessionListState, routing webview messages, registering the
 *   spectra.openPanel command, and pushing all disposables to context.subscriptions.
 */

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
  sendEvent(eventType: string, message: string): void;
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
 * Injectable dependencies for the activate function.
 * In production these are satisfied by real vscode APIs and service constructors.
 * In tests these are replaced with stubs/mocks.
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
 * @param context - The VS Code extension context.
 * @param deps - Injectable dependencies (optional; for testing).
 */
export function activate(context: IExtensionContext, deps: ExtensionDeps): void {
  // Step 1: Create OutputChannel named 'Spectra'
  const outputChannel = deps.createOutputChannel("Spectra");

  // Step 2: Wrap in logger adapter
  const logger = createLogger(outputChannel);

  // Step 3: Log activation start
  logger.info("Spectra extension activating...");

  // Step 4: Resolve project root
  const projectRoot = deps.resolveProjectRoot();

  // Step 5: If projectRoot is undefined, show error and return early
  if (projectRoot === undefined) {
    deps.showErrorMessage("Spectra: No workspace folder open.");
    logger.error("No workspace folder open. Activation aborted.");
    context.subscriptions.push(outputChannel);
    return;
  }

  // Step 6: Create SessionListController
  const sessionListController = deps.createSessionListController(projectRoot, logger);

  // Step 7: Create SessionDetailController
  const sessionDetailController = deps.createSessionDetailController(projectRoot, logger);

  // Step 8: Create or reveal panel
  const panel = deps.createOrRevealPanel(context, context.extensionUri, logger);

  // Step 9: Initialize cached session list state
  let cachedSessionListState: any = null;

  // Step 10: Subscribe sessionListController.onDidUpdate
  const listUpdateSub = sessionListController.onDidUpdate((state: any) => {
    cachedSessionListState = state;
    panel.showSessionList(state);
  });

  // Step 11: Subscribe sessionDetailController.onDidUpdate
  const detailUpdateSub = sessionDetailController.onDidUpdate((state: any) => {
    panel.showSessionDetail(state);
  });

  // Step 12: Subscribe sessionListController.onDidError
  const listErrorSub = sessionListController.onDidError((error: any) => {
    deps.showErrorMessage(error.message);
  });

  // Step 13: Subscribe sessionDetailController.onDidError
  const detailErrorSub = sessionDetailController.onDidError((error: any) => {
    deps.showErrorMessage(error.message);
  });

  // Step 14: Subscribe panel.onDidReceiveMessage and route
  const messageSub = panel.onDidReceiveMessage((msg: any) => {
    switch (msg.command) {
      case "navigateToDetail":
        sessionDetailController.open(msg.sessionId, msg.workflowName);
        break;
      case "navigateToList":
        if (cachedSessionListState !== null) {
          panel.showSessionList(cachedSessionListState);
        }
        break;
      case "launchSession":
        sessionListController.launch(msg.workflowName);
        break;
      case "terminateSession":
        sessionListController.terminate(msg.pid);
        break;
      case "sendEvent":
        sessionDetailController.sendEvent(msg.eventType, msg.message);
        break;
      default:
        logger.warn(`Unrecognized webview command: ${msg.command}`);
        break;
    }
  });

  // Step 15: Subscribe panel.onDidDispose
  const disposeSub = panel.onDidDispose(() => {
    sessionListController.dispose();
    sessionDetailController.dispose();
  });

  // Step 16: Register spectra.openPanel command
  const commandDisposable = deps.registerCommand("spectra.openPanel", () => {
    deps.createOrRevealPanel(context, context.extensionUri, logger);
  });

  // Step 17: Push all disposables to context.subscriptions
  context.subscriptions.push(outputChannel);
  context.subscriptions.push(sessionListController);
  context.subscriptions.push(sessionDetailController);
  context.subscriptions.push(panel);
  context.subscriptions.push(commandDisposable);
  context.subscriptions.push(listUpdateSub);
  context.subscriptions.push(detailUpdateSub);
  context.subscriptions.push(listErrorSub);
  context.subscriptions.push(detailErrorSub);
  context.subscriptions.push(messageSub);
  context.subscriptions.push(disposeSub);

  // Step 18: Log successful activation
  logger.info(`Spectra extension activated successfully. Project root: ${projectRoot}`);
}

/**
 * Deactivates the Spectra extension.
 * Empty function body — all cleanup is handled by context.subscriptions disposal.
 */
export function deactivate(): void {
  // Step 19: Empty function body.
}
