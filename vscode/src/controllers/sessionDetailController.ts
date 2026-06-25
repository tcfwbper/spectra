/**
 * SessionDetailController — manages state and interactions for a single
 * Session detail view.
 *
 * Logic spec: spec/logic/vscode/src/controllers/sessionDetailController.md
 */

/**
 * Minimal Disposable interface matching vscode.Disposable.
 */
export interface Disposable {
  dispose(): void;
}

/**
 * Minimal Event interface matching vscode.Event<T>.
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
 * Minimal EventWatcher-like interface (subset used by the controller).
 */
export interface IEventWatcher extends Disposable {
  onDidChange: Event<void>;
}

/**
 * Logger interface required by SessionDetailController.
 */
export interface SessionDetailControllerLogger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * State pushed via onDidUpdate after open() or a re-scan.
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
 * Injectable dependencies for SessionDetailController.
 * Enables full testability by decoupling from concrete service classes.
 */
export interface SessionDetailControllerDeps {
  /** Factory to construct an EventWatcher. */
  createEventWatcher(projectRoot: string, sessionId: string): IEventWatcher;
  /** EventScanner.scan(projectRoot, sessionId, logger) */
  scanEvents(projectRoot: string, sessionId: string, logger: SessionDetailControllerLogger): Promise<any[]>;
  /** SessionScanner.scan(projectRoot, logger) */
  scanSessions(projectRoot: string, logger: SessionDetailControllerLogger): Promise<any[]>;
  /** WorkflowDefinitionParser.parse(projectRoot, workflowName, logger) */
  parseWorkflowDefinition(projectRoot: string, workflowName: string, logger: SessionDetailControllerLogger): Promise<{ entryNode: string; eventTypes: string[] }>;
  /** EventDispatcher.dispatch(eventType, sessionId, message, projectRoot, logger) */
  dispatchEvent(eventType: string, sessionId: string, message: string, projectRoot: string, logger: SessionDetailControllerLogger): Promise<void>;
  /** Factory to construct a vscode.EventEmitter<SessionDetailState>. */
  createStateEmitter(): IEventEmitter<SessionDetailState>;
  /** Factory to construct a vscode.EventEmitter<Error>. */
  createErrorEmitter(): IEventEmitter<Error>;
}

/**
 * Manages state and interactions for a single Session detail view.
 *
 * - Owns: constructing/disposing EventWatcher instances, subscribing to onDidChange,
 *   triggering scans, assembling SessionDetailState, coalescing overlapping scans,
 *   maintaining generation counter, dispatching events, suppressing callbacks after dispose.
 * - Must not: read/write files, spawn processes, or display UI elements.
 */
export class SessionDetailController implements Disposable {
  private readonly _projectRoot: string;
  private readonly _logger: SessionDetailControllerLogger;
  private readonly _deps: SessionDetailControllerDeps;
  private readonly _fallbackScanDelayMs: number;
  private readonly _stateEmitter: IEventEmitter<SessionDetailState>;
  private readonly _errorEmitter: IEventEmitter<Error>;

  /** Public event: fires when session detail state changes. */
  public readonly onDidUpdate: Event<SessionDetailState>;
  /** Public event: fires for actionable errors. */
  public readonly onDidError: Event<Error>;

  private _currentWatcher: IEventWatcher | null = null;
  private _currentSessionId: string | null = null;
  private _currentWorkflowName: string | null = null;
  private _generation = 0;
  private _dirty = false;
  private _scanning = false;
  private _disposed = false;
  private _fallbackTimer: ReturnType<typeof setTimeout> | null = null;

  /** Cached workflow parse results for use in re-scans. */
  private _entryNode = "";
  private _eventTypes: string[] = [];

  constructor(
    projectRoot: string,
    logger: SessionDetailControllerLogger,
    deps: SessionDetailControllerDeps,
    fallbackScanDelayMs?: number,
  ) {
    this._projectRoot = projectRoot;
    this._logger = logger;
    this._deps = deps;
    this._fallbackScanDelayMs = fallbackScanDelayMs ?? 800;

    this._stateEmitter = deps.createStateEmitter();
    this._errorEmitter = deps.createErrorEmitter();

    this.onDidUpdate = this._stateEmitter.event;
    this.onDidError = this._errorEmitter.event;
  }

  /**
   * Opens a session for detail viewing. Creates a new EventWatcher,
   * parses the workflow definition, scans events and sessions, and
   * fires onDidUpdate with the assembled state.
   */
  async open(sessionId: string, workflowName: string): Promise<void> {
    if (this._disposed) {
      return;
    }

    // Dispose previous watcher
    if (this._currentWatcher !== null) {
      this._currentWatcher.dispose();
      this._currentWatcher = null;
    }

    // Cancel any pending fallback timer
    if (this._fallbackTimer !== null) {
      clearTimeout(this._fallbackTimer);
      this._fallbackTimer = null;
    }

    // Increment generation
    this._generation += 1;
    const openGeneration = this._generation;

    // Store current session info
    this._currentSessionId = sessionId;
    this._currentWorkflowName = workflowName;

    // Reset scan state
    this._dirty = false;
    this._scanning = false;

    // Construct EventWatcher (throws propagate to caller)
    const watcher = this._deps.createEventWatcher(this._projectRoot, sessionId);
    this._currentWatcher = watcher;

    // Subscribe to onDidChange
    const subscriptionGeneration = openGeneration;
    watcher.onDidChange(() => {
      this._onDidChange(subscriptionGeneration);
    });

    // Perform initial scan
    const [parseResult, events, sessions] = await Promise.all([
      this._deps.parseWorkflowDefinition(this._projectRoot, workflowName, this._logger),
      this._deps.scanEvents(this._projectRoot, sessionId, this._logger),
      this._deps.scanSessions(this._projectRoot, this._logger),
    ]);

    // Check generation (a newer open may have taken over)
    if (this._generation !== openGeneration) {
      return;
    }

    // Store parse results for subsequent re-scans
    this._entryNode = parseResult.entryNode;
    this._eventTypes = parseResult.eventTypes;

    // Find matching session
    const matchingSession = sessions.find((s: any) => s.id === sessionId);
    const currentState = matchingSession ? matchingSession.currentState : "";
    const status = matchingSession ? matchingSession.status : "initializing";
    const pid = matchingSession ? matchingSession.pid : 0;

    // Assemble and fire state
    if (!this._disposed) {
      this._stateEmitter.fire({
        sessionId,
        workflowName,
        entryNode: this._entryNode,
        currentState,
        status,
        pid,
        eventTypes: this._eventTypes,
        events,
      });
    }
  }

  /**
   * Sends an event via EventDispatcher.
   * Returns true on success, false on failure or when disposed.
   */
  async sendEvent(eventType: string, message: string): Promise<boolean> {
    if (this._disposed) {
      return false;
    }

    try {
      await this._deps.dispatchEvent(eventType, this._currentSessionId!, message, this._projectRoot, this._logger);

      // Schedule fallback timer if a session is open (currentWatcher is not null)
      if (this._currentWatcher !== null) {
        // Debounce: cancel any existing fallback timer
        if (this._fallbackTimer !== null) {
          clearTimeout(this._fallbackTimer);
          this._fallbackTimer = null;
        }

        const timerGeneration = this._generation;
        this._fallbackTimer = setTimeout(() => {
          this._fallbackTimer = null;
          if (this._disposed || this._generation !== timerGeneration) {
            return;
          }
          this._logger.info(
            `fallback scan triggered after sendEvent for session ${this._currentSessionId}`,
          );
          this._runFallbackScan(timerGeneration);
        }, this._fallbackScanDelayMs);
      }

      return true;
    } catch (err: any) {
      if (!this._disposed) {
        this._logger.error(err.message);
        this._errorEmitter.fire(err);
      }
      return false;
    }
  }

  /**
   * Disposes all resources.
   */
  dispose(): void {
    if (this._disposed) {
      return;
    }
    this._disposed = true;

    // Cancel any pending fallback timer
    if (this._fallbackTimer !== null) {
      clearTimeout(this._fallbackTimer);
      this._fallbackTimer = null;
    }

    if (this._currentWatcher !== null) {
      this._currentWatcher.dispose();
      this._currentWatcher = null;
    }

    this._stateEmitter.dispose();
    this._errorEmitter.dispose();
  }

  /**
   * Internal scan routine triggered by onDidChange.
   */
  private _onDidChange(subscriptionGeneration: number): void {
    if (this._disposed) {
      return;
    }

    // Check if this subscription is still valid
    if (this._generation !== subscriptionGeneration) {
      return;
    }

    // Coalesce: if a scan is in-flight, just set dirty
    if (this._scanning) {
      this._dirty = true;
      return;
    }

    this._runScan();
  }

  /**
   * Runs a scan cycle, re-running if dirty flag is set during execution.
   */
  private async _runScan(): Promise<void> {
    this._scanning = true;
    const scanGeneration = this._generation;

    try {
      const [events, sessions] = await Promise.all([
        this._deps.scanEvents(this._projectRoot, this._currentSessionId!, this._logger),
        this._deps.scanSessions(this._projectRoot, this._logger),
      ]);

      // Check generation
      if (this._generation !== scanGeneration) {
        this._scanning = false;
        return;
      }

      // Check disposed
      if (this._disposed) {
        this._scanning = false;
        return;
      }

      // Find matching session
      const matchingSession = sessions.find((s: any) => s.id === this._currentSessionId);
      const currentState = matchingSession ? matchingSession.currentState : "";
      const status = matchingSession ? matchingSession.status : "initializing";
      const pid = matchingSession ? matchingSession.pid : 0;

      // Fire state
      this._stateEmitter.fire({
        sessionId: this._currentSessionId!,
        workflowName: this._currentWorkflowName!,
        entryNode: this._entryNode,
        currentState,
        status,
        pid,
        eventTypes: this._eventTypes,
        events,
      });
    } finally {
      this._scanning = false;
    }

    // If dirty flag was set during the scan, run again
    if (this._dirty) {
      this._dirty = false;
      if (!this._disposed && this._generation === scanGeneration) {
        this._runScan();
      }
    }
  }

  /**
   * Runs a fallback scan triggered by the fallback timer.
   * Uses the same coalescing mechanism (dirty flag) as onDidChange.
   * Catches errors and logs via logger.error without firing onDidError.
   */
  private _runFallbackScan(timerGeneration: number): void {
    if (this._disposed) {
      return;
    }
    if (this._generation !== timerGeneration) {
      return;
    }

    // Use the same coalescing mechanism: if a scan is in-flight, set dirty
    if (this._scanning) {
      this._dirty = true;
      return;
    }

    this._runScan().catch((err: any) => {
      if (!this._disposed) {
        this._logger.error(err.message);
      }
    });
  }
}
