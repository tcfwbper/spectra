/**
 * Shared test helpers for extension.test.ts.
 *
 * Provides mock factories for all collaborators that the activate() function
 * orchestrates: OutputChannel, ViewProvider, SessionListController,
 * SessionDetailController, and all deps fields.
 *
 * Architecture: The production activate(context, deps?) uses an all-optional
 * ActivateDeps interface. When a field is missing, production defaults are used.
 * Tests inject stubs via the deps parameter to avoid requiring the vscode module.
 *
 * Test spec: spec/test/vscode/src/extension.md
 */
import * as sinon from "sinon";

import { activate, type ActivateDeps } from "../../../src/extension";

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

// ─── Mock ViewProvider ──────────────────────────────────────────────────────

/**
 * Callback holder for event-based subscriptions.
 */
type Callback<T> = (value: T) => void;

/**
 * Mock SpectraViewProvider instance with controllable event triggers.
 */
export interface MockViewProvider {
  showSessionList: sinon.SinonStub;
  showSessionDetail: sinon.SinonStub;
  showNotInitialized: sinon.SinonStub;
  dispose: sinon.SinonStub;
  onDidReceiveMessage: (listener: Callback<any>) => { dispose: () => void };
  /** Trigger the onDidReceiveMessage callback. */
  triggerMessage: (msg: any) => void;
  /** All message listeners (for test inspection). */
  messageListeners: Array<Callback<any>>;
}

/**
 * Creates a mock SpectraViewProvider instance.
 */
export function createMockViewProvider(): MockViewProvider {
  const messageListeners: Array<Callback<any>> = [];

  return {
    showSessionList: sinon.stub(),
    showSessionDetail: sinon.stub(),
    showNotInitialized: sinon.stub(),
    dispose: sinon.stub(),
    onDidReceiveMessage: (listener: Callback<any>) => {
      messageListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerMessage: (msg: any) => {
      for (const l of [...messageListeners]) {
        l(msg);
      }
    },
    messageListeners,
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

// ─── Full Test Fixture ───────────────────────────────────────────────────────

/**
 * A complete test fixture for extension.test.ts that assembles all mocks.
 *
 * The fixture provides:
 * - A pre-built deps object (ActivateDeps) ready to pass to activate().
 * - Direct access to mock instances for assertion (viewProvider, controllers, etc).
 * - Trigger methods on controllers and viewProvider for simulating events.
 */
export interface ExtensionTestFixture {
  context: MockExtensionContext;
  outputChannel: MockOutputChannel;
  viewProvider: MockViewProvider;
  sessionListController: MockSessionListController;
  sessionDetailController: MockSessionDetailController;
  /** The fully-assembled deps object. Modify individual fields before calling activate. */
  deps: ActivateDeps;
  /** Spies/stubs for verifying dep calls */
  spies: {
    resolveProjectRoot: sinon.SinonStub;
    showErrorMessage: sinon.SinonStub;
    registerCommand: sinon.SinonStub;
    createViewProvider: sinon.SinonStub;
    registerWebviewViewProvider: sinon.SinonStub;
    createSessionListController: sinon.SinonStub;
    createSessionDetailController: sinon.SinonStub;
  };
}

/**
 * Creates a fully-assembled test fixture with default "happy path" wiring.
 *
 * All deps fields are pre-wired so that activate(context, deps) succeeds.
 * Tests can override individual deps fields before calling activateWithFixture.
 *
 * @param projectRoot - The value resolveProjectRoot() will return.
 *   Pass `undefined` to simulate no workspace. Defaults to "/workspace".
 */
export function createExtensionTestFixture(
  projectRoot?: string | undefined,
): ExtensionTestFixture {
  // Use "/workspace" only when no argument is provided (arguments.length === 0).
  // Explicit `undefined` must be preserved to test the "no workspace" path.
  const resolvedProjectRoot = arguments.length === 0 ? "/workspace" : projectRoot;
  const context = createMockExtensionContext();
  const outputChannel = createMockOutputChannel();
  const viewProvider = createMockViewProvider();
  const sessionListController = createMockSessionListController();
  const sessionDetailController = createMockSessionDetailController();

  const resolveProjectRoot = sinon.stub().returns(resolvedProjectRoot);
  const showErrorMessage = sinon.stub();
  const registerCommand = sinon.stub().returns({ dispose: () => {} });
  const createViewProvider = sinon.stub().returns(viewProvider);
  const registerWebviewViewProvider = sinon
    .stub()
    .returns({ dispose: () => {} });
  const createSessionListController = sinon
    .stub()
    .returns(sessionListController);
  const createSessionDetailController = sinon
    .stub()
    .returns(sessionDetailController);

  const deps: ActivateDeps = {
    outputChannel,
    showErrorMessage,
    registerCommand,
    resolveProjectRoot,
    createViewProvider,
    registerWebviewViewProvider,
    createSessionListController,
    createSessionDetailController,
  };

  return {
    context,
    outputChannel,
    viewProvider,
    sessionListController,
    sessionDetailController,
    deps,
    spies: {
      resolveProjectRoot,
      showErrorMessage,
      registerCommand,
      createViewProvider,
      registerWebviewViewProvider,
      createSessionListController,
      createSessionDetailController,
    },
  };
}

// ─── Bridge: fixture → activate() ────────────────────────────────────────────

/**
 * Calls activate(context, deps) using the fixture's assembled deps.
 *
 * The production activate() function accepts (context, deps?) where deps
 * is an ActivateDeps object with all-optional fields. When a field is present,
 * it is used instead of the production default.
 */
export function activateWithFixture(fixture: ExtensionTestFixture): void {
  activate(fixture.context, fixture.deps);
}
