/**
 * Shared test helpers for SessionDetailController and SessionListController tests.
 *
 * Provides mock factories for all collaborators that the controllers delegate to:
 * EventWatcher, SessionWatcher, WorkflowWatcher, EventScanner, SessionScanner,
 * WorkflowDefinitionParser, WorkflowScanner, EventDispatcher, SessionLauncher,
 * SessionTerminator, and vscode.EventEmitter.
 *
 * Scaffolded: The controller source files do not yet exist. These helpers are
 * designed around the interfaces defined in the logic specs so that tests
 * compile and provide structural coverage once the production surface is created.
 */
import * as sinon from "sinon";

// ─── Logger ───────────────────────────────────────────────────────────────────

/**
 * Logger interface matching the shape required by both controllers.
 */
export interface MockControllerLogger {
  info: sinon.SinonSpy;
  warn: sinon.SinonSpy;
  error: sinon.SinonSpy;
}

/**
 * Creates a mock logger with sinon spies on info, warn, and error.
 */
export function createMockControllerLogger(): MockControllerLogger {
  return {
    info: sinon.spy(),
    warn: sinon.spy(),
    error: sinon.spy(),
  };
}

// ─── Deferred Promise ─────────────────────────────────────────────────────────

/**
 * A deferred promise that can be resolved/rejected externally.
 * Used to control async timing in concurrency tests.
 */
export interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (reason?: any) => void;
}

/**
 * Creates a deferred promise.
 */
export function createDeferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  let reject!: (reason?: any) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

// ─── EventEmitter Mock (vscode.EventEmitter<T>) ──────────────────────────────

/**
 * Callback type for event subscriptions.
 */
type EventListener<T> = (e: T) => void;

/**
 * Mock vscode.EventEmitter<T> that tracks fire/dispose calls and provides
 * an event accessor function that registers listeners.
 */
export interface MockTypedEventEmitter<T> {
  /** The `.event` accessor — a function that registers listeners. */
  event: (listener: EventListener<T>) => { dispose: () => void };
  /** Fire method — invokes all registered listeners with the value. */
  fire: sinon.SinonStub;
  /** Dispose spy. */
  dispose: sinon.SinonStub;
  /** All currently registered listeners (test inspection). */
  listeners: Array<EventListener<T>>;
}

/**
 * Creates a mock vscode.EventEmitter<T>.
 */
export function createMockTypedEventEmitter<T>(): MockTypedEventEmitter<T> {
  const listeners: Array<EventListener<T>> = [];

  const event = (listener: EventListener<T>) => {
    listeners.push(listener);
    return { dispose: () => {} };
  };

  const fire = sinon.stub().callsFake((value: T) => {
    for (const l of listeners) {
      l(value);
    }
  });

  return {
    event,
    fire,
    dispose: sinon.stub(),
    listeners,
  };
}

// ─── EventWatcher Mock ────────────────────────────────────────────────────────

/**
 * A mock EventWatcher instance for SessionDetailController tests.
 */
export interface MockEventWatcherInstance {
  /** Controllable onDidChange — call .fire() to simulate file change. */
  onDidChange: (listener: () => void) => { dispose: () => void };
  /** Programmatically trigger onDidChange listeners (test utility). */
  triggerChange: () => void;
  /** Dispose spy. */
  dispose: sinon.SinonStub;
}

/**
 * Creates a mock EventWatcher instance.
 */
export function createMockEventWatcherInstance(): MockEventWatcherInstance {
  const changeListeners: Array<() => void> = [];

  return {
    onDidChange: (listener: () => void) => {
      changeListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerChange: () => {
      for (const l of changeListeners) {
        l();
      }
    },
    dispose: sinon.stub(),
  };
}

// ─── SessionWatcher / WorkflowWatcher Mocks ──────────────────────────────────

/**
 * A mock watcher instance (SessionWatcher or WorkflowWatcher) for
 * SessionListController tests.
 */
export interface MockListWatcherInstance {
  /** Controllable onDidChange — call .fire() to simulate file change. */
  onDidChange: (listener: () => void) => { dispose: () => void };
  /** Programmatically trigger onDidChange listeners (test utility). */
  triggerChange: () => void;
  /** Dispose spy. */
  dispose: sinon.SinonStub;
}

/**
 * Creates a mock watcher instance (same shape for SessionWatcher/WorkflowWatcher).
 */
export function createMockListWatcherInstance(): MockListWatcherInstance {
  const changeListeners: Array<() => void> = [];

  return {
    onDidChange: (listener: () => void) => {
      changeListeners.push(listener);
      return { dispose: () => {} };
    },
    triggerChange: () => {
      for (const l of changeListeners) {
        l();
      }
    },
    dispose: sinon.stub(),
  };
}

// ─── SessionDetailController Dependencies ─────────────────────────────────────

/**
 * Injectable dependencies for SessionDetailController.
 * The controller production code is expected to accept this shape for testability.
 *
 * Scaffolded: exact interface name/shape will be confirmed when the controller
 * source is created. This follows the pattern established by EventWatcher, etc.
 */
export interface SessionDetailControllerDeps {
  /** Factory to construct an EventWatcher. */
  createEventWatcher: sinon.SinonStub;
  /** Static method: EventScanner.scan(projectRoot, sessionId, logger) */
  scanEvents: sinon.SinonStub;
  /** Static method: SessionScanner.scan(projectRoot, logger) */
  scanSessions: sinon.SinonStub;
  /** Static method: WorkflowDefinitionParser.parse(projectRoot, workflowName, logger) */
  parseWorkflowDefinition: sinon.SinonStub;
  /** Static method: EventDispatcher.dispatch(eventType, sessionId, message, logger) */
  dispatchEvent: sinon.SinonStub;
  /** Factory to construct a vscode.EventEmitter<SessionDetailState>. */
  createStateEmitter: sinon.SinonStub;
  /** Factory to construct a vscode.EventEmitter<Error>. */
  createErrorEmitter: sinon.SinonStub;
}

/**
 * Creates stubbed dependencies for SessionDetailController.
 * Returns both the deps object and references to the mock emitters/watcher
 * for assertion purposes.
 */
export function createSessionDetailControllerDeps(): {
  deps: SessionDetailControllerDeps;
  stateEmitter: MockTypedEventEmitter<any>;
  errorEmitter: MockTypedEventEmitter<Error>;
  eventWatcher: MockEventWatcherInstance;
} {
  const stateEmitter = createMockTypedEventEmitter<any>();
  const errorEmitter = createMockTypedEventEmitter<Error>();
  const eventWatcher = createMockEventWatcherInstance();

  const deps: SessionDetailControllerDeps = {
    createEventWatcher: sinon.stub().returns(eventWatcher),
    scanEvents: sinon.stub().resolves([]),
    scanSessions: sinon.stub().resolves([]),
    parseWorkflowDefinition: sinon.stub().resolves({ entryNode: "", eventTypes: [] }),
    dispatchEvent: sinon.stub().resolves(),
    createStateEmitter: sinon.stub().returns(stateEmitter),
    createErrorEmitter: sinon.stub().returns(errorEmitter),
  };

  return { deps, stateEmitter, errorEmitter, eventWatcher };
}

// ─── SessionListController Dependencies ───────────────────────────────────────

/**
 * Injectable dependencies for SessionListController.
 * The controller production code is expected to accept this shape for testability.
 *
 * Scaffolded: exact interface name/shape will be confirmed when the controller
 * source is created.
 */
export interface SessionListControllerDeps {
  /** Factory to construct a SessionWatcher. */
  createSessionWatcher: sinon.SinonStub;
  /** Factory to construct a WorkflowWatcher. */
  createWorkflowWatcher: sinon.SinonStub;
  /** Static method: SessionScanner.scan(projectRoot, logger) */
  scanSessions: sinon.SinonStub;
  /** Static method: WorkflowScanner.scan(projectRoot, logger) */
  scanWorkflows: sinon.SinonStub;
  /** Static method: SessionLauncher.launch(workflowName, logger) */
  launch: sinon.SinonStub;
  /** Static method: SessionTerminator.terminate(pid, logger) */
  terminate: sinon.SinonStub;
  /** Factory to construct a vscode.EventEmitter<SessionListState>. */
  createStateEmitter: sinon.SinonStub;
  /** Factory to construct a vscode.EventEmitter<Error>. */
  createErrorEmitter: sinon.SinonStub;
}

/**
 * Creates stubbed dependencies for SessionListController.
 * Returns both the deps object and references to the mock emitters/watchers.
 */
export function createSessionListControllerDeps(): {
  deps: SessionListControllerDeps;
  stateEmitter: MockTypedEventEmitter<any>;
  errorEmitter: MockTypedEventEmitter<Error>;
  sessionWatcher: MockListWatcherInstance;
  workflowWatcher: MockListWatcherInstance;
} {
  const stateEmitter = createMockTypedEventEmitter<any>();
  const errorEmitter = createMockTypedEventEmitter<Error>();
  const sessionWatcher = createMockListWatcherInstance();
  const workflowWatcher = createMockListWatcherInstance();

  const deps: SessionListControllerDeps = {
    createSessionWatcher: sinon.stub().returns(sessionWatcher),
    createWorkflowWatcher: sinon.stub().returns(workflowWatcher),
    scanSessions: sinon.stub().resolves([]),
    scanWorkflows: sinon.stub().resolves([]),
    launch: sinon.stub().resolves(),
    terminate: sinon.stub().resolves({ method: "sigterm", terminated: true }),
    createStateEmitter: sinon.stub().returns(stateEmitter),
    createErrorEmitter: sinon.stub().returns(errorEmitter),
  };

  return { deps, stateEmitter, errorEmitter, sessionWatcher, workflowWatcher };
}
