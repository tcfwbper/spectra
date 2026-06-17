/**
 * Shared test helpers for extension.test.ts.
 *
 * Provides mock factories for all collaborators that the activate() function
 * orchestrates: vscode.window, vscode.commands, ProjectRootResolver,
 * SessionListController, SessionDetailController, and SpectraPanel.
 */
import * as sinon from "sinon";

import { activate, type ExtensionDeps } from "../../../src/extension";

// ─── Logger Mock ─────────────────────────────────────────────────────────────

/**
 * Logger interface matching the shape created inside activate().
 * The logger wraps an OutputChannel with { info, warn, error } methods.
 */
export interface MockLogger {
  info: sinon.SinonSpy;
  warn: sinon.SinonSpy;
  error: sinon.SinonSpy;
}

/**
 * Creates a mock logger.
 */
export function createMockLogger(): MockLogger {
  return {
    info: sinon.spy(),
    warn: sinon.spy(),
    error: sinon.spy(),
  };
}

// ─── OutputChannel Mock ──────────────────────────────────────────────────────

/**
 * Mock OutputChannel matching vscode.OutputChannel shape.
 */
export interface MockOutputChannel {
  appendLine: sinon.SinonSpy;
  dispose: sinon.SinonStub;
}

/**
 * Creates a mock OutputChannel.
 */
export function createMockOutputChannel(): MockOutputChannel {
  return {
    appendLine: sinon.spy(),
    dispose: sinon.stub(),
  };
}

// ─── Extension Context Mock ──────────────────────────────────────────────────

/**
 * Mock ExtensionContext matching vscode.ExtensionContext shape.
 */
export interface MockExtensionContext {
  subscriptions: any[];
  extensionUri: MockUri;
}

/**
 * Minimal URI mock.
 */
export interface MockUri {
  fsPath: string;
  scheme: string;
  path: string;
}

/**
 * Creates a mock ExtensionContext.
 */
export function createMockExtensionContext(
  extensionPath = "/test/extension",
): MockExtensionContext {
  return {
    subscriptions: [],
    extensionUri: {
      fsPath: extensionPath,
      scheme: "file",
      path: extensionPath,
    },
  };
}

// ─── Mock Panel ──────────────────────────────────────────────────────────────

/**
 * Callback holder for event-based subscriptions.
 */
type Callback<T> = (value: T) => void;

/**
 * Mock SpectraPanel instance with controllable event triggers.
 */
export interface MockPanel {
  showSessionList: sinon.SinonStub;
  showSessionDetail: sinon.SinonStub;
  dispose: sinon.SinonStub;
  onDidReceiveMessage: (listener: Callback<any>) => { dispose: () => void };
  onDidDispose: (listener: () => void) => { dispose: () => void };
  /** Trigger the onDidReceiveMessage callback. */
  triggerMessage: (msg: any) => void;
  /** Trigger the onDidDispose callback. */
  triggerDispose: () => void;
  /** All message listeners (for test inspection). */
  messageListeners: Array<Callback<any>>;
  /** All dispose listeners (for test inspection). */
  disposeListeners: Array<() => void>;
}

/**
 * Creates a mock SpectraPanel instance.
 */
export function createMockPanel(): MockPanel {
  const messageListeners: Array<Callback<any>> = [];
  const disposeListeners: Array<() => void> = [];

  return {
    showSessionList: sinon.stub(),
    showSessionDetail: sinon.stub(),
    dispose: sinon.stub(),
    onDidReceiveMessage: (listener: Callback<any>) => {
      messageListeners.push(listener);
      return { dispose: () => {} };
    },
    onDidDispose: (listener: () => void) => {
      disposeListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerMessage: (msg: any) => {
      for (const l of [...messageListeners]) {
        l(msg);
      }
    },
    triggerDispose: () => {
      for (const l of [...disposeListeners]) {
        l();
      }
    },
    messageListeners,
    disposeListeners,
  };
}

// ─── Mock SessionListController ──────────────────────────────────────────────

/**
 * Mock SessionListController with controllable event triggers.
 */
export interface MockSessionListController {
  launch: sinon.SinonStub;
  terminate: sinon.SinonStub;
  dispose: sinon.SinonStub;
  onDidUpdate: (listener: Callback<any>) => { dispose: () => void };
  onDidError: (listener: Callback<any>) => { dispose: () => void };
  /** Trigger onDidUpdate callback. */
  triggerUpdate: (state: any) => void;
  /** Trigger onDidError callback. */
  triggerError: (err: any) => void;
  /** All update listeners (for test inspection). */
  updateListeners: Array<Callback<any>>;
  /** All error listeners (for test inspection). */
  errorListeners: Array<Callback<any>>;
}

/**
 * Creates a mock SessionListController.
 */
export function createMockSessionListController(): MockSessionListController {
  const updateListeners: Array<Callback<any>> = [];
  const errorListeners: Array<Callback<any>> = [];

  return {
    launch: sinon.stub(),
    terminate: sinon.stub(),
    dispose: sinon.stub(),
    onDidUpdate: (listener: Callback<any>) => {
      updateListeners.push(listener);
      return { dispose: () => {} };
    },
    onDidError: (listener: Callback<any>) => {
      errorListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerUpdate: (state: any) => {
      for (const l of [...updateListeners]) {
        l(state);
      }
    },
    triggerError: (err: any) => {
      for (const l of [...errorListeners]) {
        l(err);
      }
    },
    updateListeners,
    errorListeners,
  };
}

// ─── Mock SessionDetailController ────────────────────────────────────────────

/**
 * Mock SessionDetailController with controllable event triggers.
 */
export interface MockSessionDetailController {
  open: sinon.SinonStub;
  sendEvent: sinon.SinonStub;
  dispose: sinon.SinonStub;
  onDidUpdate: (listener: Callback<any>) => { dispose: () => void };
  onDidError: (listener: Callback<any>) => { dispose: () => void };
  /** Trigger onDidUpdate callback. */
  triggerUpdate: (state: any) => void;
  /** Trigger onDidError callback. */
  triggerError: (err: any) => void;
  /** All update listeners (for test inspection). */
  updateListeners: Array<Callback<any>>;
  /** All error listeners (for test inspection). */
  errorListeners: Array<Callback<any>>;
}

/**
 * Creates a mock SessionDetailController.
 */
export function createMockSessionDetailController(): MockSessionDetailController {
  const updateListeners: Array<Callback<any>> = [];
  const errorListeners: Array<Callback<any>> = [];

  return {
    open: sinon.stub(),
    sendEvent: sinon.stub(),
    dispose: sinon.stub(),
    onDidUpdate: (listener: Callback<any>) => {
      updateListeners.push(listener);
      return { dispose: () => {} };
    },
    onDidError: (listener: Callback<any>) => {
      errorListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerUpdate: (state: any) => {
      for (const l of [...updateListeners]) {
        l(state);
      }
    },
    triggerError: (err: any) => {
      for (const l of [...errorListeners]) {
        l(err);
      }
    },
    updateListeners,
    errorListeners,
  };
}

// ─── Mock vscode namespace ───────────────────────────────────────────────────

/**
 * Represents the mocked vscode.window and vscode.commands namespace
 * needed by extension.ts tests.
 */
export interface MockVscodeNamespace {
  window: {
    createOutputChannel: sinon.SinonStub;
    showErrorMessage: sinon.SinonStub;
  };
  commands: {
    registerCommand: sinon.SinonStub;
  };
}

/**
 * Creates a mock vscode namespace with stubs for window and commands.
 */
export function createMockVscodeNamespace(
  outputChannel: MockOutputChannel,
): MockVscodeNamespace {
  return {
    window: {
      createOutputChannel: sinon.stub().returns(outputChannel),
      showErrorMessage: sinon.stub(),
    },
    commands: {
      registerCommand: sinon.stub().returns({ dispose: () => {} }),
    },
  };
}

// ─── Full Test Fixture ───────────────────────────────────────────────────────

/**
 * A complete test fixture for extension.test.ts that assembles all mocks.
 */
export interface ExtensionTestFixture {
  context: MockExtensionContext;
  outputChannel: MockOutputChannel;
  vscode: MockVscodeNamespace;
  panel: MockPanel;
  sessionListController: MockSessionListController;
  sessionDetailController: MockSessionDetailController;
  projectRootResolveStub: sinon.SinonStub;
  spectraPanelCreateOrRevealStub: sinon.SinonStub;
  sessionListControllerConstructorStub: sinon.SinonStub;
  sessionDetailControllerConstructorStub: sinon.SinonStub;
}

/**
 * Creates a fully-assembled test fixture with default "happy path" wiring.
 *
 * @param projectRoot - The value ProjectRootResolver.resolve() will return.
 *   Pass `null` or `undefined` explicitly to simulate no workspace; the stub
 *   will return `undefined`. Omitting the argument defaults to "/workspace".
 */
export function createExtensionTestFixture(
  ...args: [string | undefined | null] | []
): ExtensionTestFixture {
  const projectRoot: string | undefined =
    args.length === 0 ? "/workspace" : (args[0] ?? undefined);
  const context = createMockExtensionContext();
  const outputChannel = createMockOutputChannel();
  const vscode = createMockVscodeNamespace(outputChannel);
  const panel = createMockPanel();
  const sessionListController = createMockSessionListController();
  const sessionDetailController = createMockSessionDetailController();

  const projectRootResolveStub = sinon.stub().returns(projectRoot);
  const spectraPanelCreateOrRevealStub = sinon.stub().returns(panel);
  const sessionListControllerConstructorStub = sinon
    .stub()
    .returns(sessionListController);
  const sessionDetailControllerConstructorStub = sinon
    .stub()
    .returns(sessionDetailController);

  return {
    context,
    outputChannel,
    vscode,
    panel,
    sessionListController,
    sessionDetailController,
    projectRootResolveStub,
    spectraPanelCreateOrRevealStub,
    sessionListControllerConstructorStub,
    sessionDetailControllerConstructorStub,
  };
}

// ─── Bridge: fixture → ExtensionDeps → activate() ────────────────────────────

/**
 * Converts an ExtensionTestFixture into the ExtensionDeps interface
 * expected by the production activate() function, then calls activate().
 *
 * This is the "staged scaffold replacement" wiring — tests call this helper
 * instead of importing activate() directly so that the fixture's mocks
 * are properly injected through the DI interface.
 */
export function activateWithFixture(fixture: ExtensionTestFixture): void {
  const deps: ExtensionDeps = {
    createOutputChannel: fixture.vscode.window.createOutputChannel,
    showErrorMessage: fixture.vscode.window.showErrorMessage,
    registerCommand: fixture.vscode.commands.registerCommand,
    resolveProjectRoot: fixture.projectRootResolveStub,
    createSessionListController: fixture.sessionListControllerConstructorStub,
    createSessionDetailController: fixture.sessionDetailControllerConstructorStub,
    createOrRevealPanel: fixture.spectraPanelCreateOrRevealStub,
  };

  activate(fixture.context, deps);
}
