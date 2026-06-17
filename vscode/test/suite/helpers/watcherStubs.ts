/**
 * Shared test helpers for stubbing VS Code FileSystemWatcher and EventEmitter
 * used by watcher services (EventWatcher, SessionWatcher, WorkflowWatcher).
 *
 * These helpers provide consistent watcher stubbing patterns so that
 * individual watcher test files remain focused on assertions.
 */
import * as sinon from "sinon";

/**
 * Callback type for VS Code event subscriptions.
 */
type EventListener<T> = (e: T) => void;

/**
 * A mock VS Code Event<T> that allows tests to register listeners
 * and trigger them programmatically.
 */
export interface MockEvent<T> {
  /** Register a listener; returns a disposable. */
  (listener: EventListener<T>): { dispose: sinon.SinonStub };
  /** Fire the event to all registered listeners (test utility). */
  fire(value: T): void;
  /** All registered listeners (test inspection). */
  listeners: Array<EventListener<T>>;
}

/**
 * Creates a mock VS Code Event<T>.
 * The returned function acts as the event accessor (like `watcher.onDidChange`).
 * Call `.fire(value)` to simulate the event firing.
 */
export function createMockEvent<T = void>(): MockEvent<T> {
  const listeners: Array<EventListener<T>> = [];

  const event = ((listener: EventListener<T>) => {
    listeners.push(listener);
    return { dispose: sinon.stub() };
  }) as MockEvent<T>;

  event.fire = (value: T) => {
    for (const listener of listeners) {
      listener(value);
    }
  };

  event.listeners = listeners;

  return event;
}

/**
 * Represents a mock VS Code FileSystemWatcher.
 */
export interface MockFileSystemWatcher {
  onDidCreate: MockEvent<any>;
  onDidChange: MockEvent<any>;
  onDidDelete: MockEvent<any>;
  dispose: sinon.SinonStub;
}

/**
 * Creates a mock FileSystemWatcher with all three event types
 * and a dispose stub.
 */
export function createMockFileSystemWatcher(): MockFileSystemWatcher {
  return {
    onDidCreate: createMockEvent<any>(),
    onDidChange: createMockEvent<any>(),
    onDidDelete: createMockEvent<any>(),
    dispose: sinon.stub(),
  };
}

/**
 * Represents the stubbed VS Code namespace used in watcher tests.
 * This provides the minimal surface needed without requiring the real
 * vscode module.
 */
export interface MockVscodeNamespace {
  workspace: {
    createFileSystemWatcher: sinon.SinonStub;
  };
  RelativePattern: sinon.SinonStub;
  EventEmitter: new () => MockEventEmitter;
}

/**
 * A mock VS Code EventEmitter<T> that tracks fire/dispose calls
 * and exposes an event accessor.
 */
export interface MockEventEmitter {
  event: MockEvent<void>;
  fire: sinon.SinonStub;
  dispose: sinon.SinonStub;
}

/**
 * Creates a mock EventEmitter that mimics vscode.EventEmitter<void>.
 */
export function createMockEventEmitter(): MockEventEmitter {
  const event = createMockEvent<void>();
  return {
    event,
    fire: sinon.stub().callsFake(() => {
      event.fire(undefined as any);
    }),
    dispose: sinon.stub(),
  };
}

/**
 * Creates a complete mock VS Code namespace suitable for watcher tests.
 * Returns the namespace plus references to the mock watcher and emitter
 * for assertion purposes.
 */
export function createWatcherTestContext(): {
  vscode: MockVscodeNamespace;
  mockWatcher: MockFileSystemWatcher;
  mockEmitter: MockEventEmitter;
  relativePatternArgs: any[];
} {
  const mockWatcher = createMockFileSystemWatcher();
  const mockEmitter = createMockEventEmitter();
  const relativePatternArgs: any[] = [];

  const vscode: MockVscodeNamespace = {
    workspace: {
      createFileSystemWatcher: sinon.stub().returns(mockWatcher),
    },
    RelativePattern: sinon.stub().callsFake((...args: any[]) => {
      relativePatternArgs.push(...args);
      return { pattern: args };
    }),
    EventEmitter: class {
      event = mockEmitter.event;
      fire = mockEmitter.fire;
      dispose = mockEmitter.dispose;
    } as any,
  };

  return { vscode, mockWatcher, mockEmitter, relativePatternArgs };
}
