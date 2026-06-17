/**
 * SessionListController — manages all state and actions for the Sessions list view.
 *
 * Logic spec: spec/logic/vscode/src/controllers/sessionListController.md
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
 * Minimal Watcher-like interface (subset used by the controller).
 */
export interface IListWatcher extends Disposable {
  onDidChange: Event<void>;
}

/**
 * Logger interface required by SessionListController.
 */
export interface SessionListControllerLogger {
  info(msg: string): void;
  warn(msg: string): void;
  error(msg: string): void;
}

/**
 * State pushed via onDidUpdate containing sessions and workflows.
 */
export interface SessionListState {
  sessions: any[];
  workflows: string[];
}

/**
 * Structured result describing the outcome of a termination attempt.
 */
export interface TerminationResult {
  terminated: boolean;
  method: "sigterm" | "sigkill" | "already_dead" | "not_spectra";
  error?: any;
}

/**
 * Injectable dependencies for SessionListController.
 * Enables full testability by decoupling from concrete service classes.
 */
export interface SessionListControllerDeps {
  /** Factory to construct a SessionWatcher. */
  createSessionWatcher(projectRoot: string): IListWatcher;
  /** Factory to construct a WorkflowWatcher. */
  createWorkflowWatcher(projectRoot: string): IListWatcher;
  /** SessionScanner.scan(projectRoot, logger) */
  scanSessions(projectRoot: string, logger: SessionListControllerLogger): Promise<any[]>;
  /** WorkflowScanner.scan(projectRoot, logger) */
  scanWorkflows(projectRoot: string, logger: SessionListControllerLogger): Promise<string[]>;
  /** SessionLauncher.launch(workflowName, projectRoot, logger) */
  launch(workflowName: string, projectRoot: string, logger: SessionListControllerLogger): Promise<void>;
  /** SessionTerminator.terminate(pid, logger) */
  terminate(pid: number, logger: SessionListControllerLogger): Promise<TerminationResult>;
  /** Factory to construct a vscode.EventEmitter<SessionListState>. */
  createStateEmitter(): IEventEmitter<SessionListState>;
  /** Factory to construct a vscode.EventEmitter<Error>. */
  createErrorEmitter(): IEventEmitter<Error>;
}

/**
 * Manages all state and actions for the Sessions list view.
 *
 * - Owns: constructing/disposing SessionWatcher and WorkflowWatcher,
 *   subscribing to onDidChange, triggering scans, assembling state,
 *   coalescing overlapping scans, firing onDidUpdate/onDidError, classifying
 *   termination results, guarding callbacks after dispose.
 * - Must not: read/write files, spawn/signal processes, or display UI elements.
 */
export class SessionListController implements Disposable {
  private readonly _projectRoot: string;
  private readonly _logger: SessionListControllerLogger;
  private readonly _deps: SessionListControllerDeps;
  private readonly _stateEmitter: IEventEmitter<SessionListState>;
  private readonly _errorEmitter: IEventEmitter<Error>;
  private readonly _sessionWatcher: IListWatcher;
  private readonly _workflowWatcher: IListWatcher;

  /** Public event: fires when session or workflow state changes. */
  public readonly onDidUpdate: Event<SessionListState>;
  /** Public event: fires for actionable errors. */
  public readonly onDidError: Event<Error>;

  /** Internal state. */
  private _sessions: any[] = [];
  private _workflows: string[] = [];

  /** Session scan coalescing. */
  private _sessionDirty = false;
  private _sessionScanning = false;

  /** Workflow scan coalescing. */
  private _workflowDirty = false;
  private _workflowScanning = false;

  private _disposed = false;

  constructor(
    projectRoot: string,
    logger: SessionListControllerLogger,
    deps: SessionListControllerDeps,
  ) {
    this._projectRoot = projectRoot;
    this._logger = logger;
    this._deps = deps;

    // Create watchers
    this._sessionWatcher = deps.createSessionWatcher(projectRoot);
    this._workflowWatcher = deps.createWorkflowWatcher(projectRoot);

    // Create emitters
    this._stateEmitter = deps.createStateEmitter();
    this._errorEmitter = deps.createErrorEmitter();

    this.onDidUpdate = this._stateEmitter.event;
    this.onDidError = this._errorEmitter.event;

    // Subscribe to watcher events
    this._sessionWatcher.onDidChange(() => this._onSessionChange());
    this._workflowWatcher.onDidChange(() => this._onWorkflowChange());

    // Kick off initial async scan (non-blocking)
    this._runSessionScan();
    this._runWorkflowScan();
  }

  /**
   * Launches a new workflow session via SessionLauncher.
   */
  async launch(workflowName: string): Promise<void> {
    try {
      await this._deps.launch(workflowName, this._projectRoot, this._logger);
    } catch (err: any) {
      if (!this._disposed) {
        this._logger.error(err.message);
        this._errorEmitter.fire(err);
      }
    }
  }

  /**
   * Terminates a running session via SessionTerminator.
   */
  async terminate(pid: number): Promise<void> {
    const result = await this._deps.terminate(pid, this._logger);

    if (this._disposed) {
      return;
    }

    // Classify result
    if (result.method === "already_dead") {
      return; // success
    }

    if (result.terminated && (result.method === "sigterm" || result.method === "sigkill")) {
      return; // success
    }

    // Error cases
    if (result.method === "not_spectra") {
      const err = new Error(`Process ${pid} no longer belongs to Spectra (PID reuse suspected)`);
      this._logger.error(err.message);
      this._errorEmitter.fire(err);
      return;
    }

    if (!result.terminated && result.error) {
      const errMsg = result.error instanceof Error ? result.error.message : String(result.error);
      const err = new Error(`Failed to terminate process ${pid}: ${errMsg}`);
      this._logger.error(err.message);
      this._errorEmitter.fire(err);
      return;
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

    this._sessionWatcher.dispose();
    this._workflowWatcher.dispose();
    this._stateEmitter.dispose();
    this._errorEmitter.dispose();
  }

  /**
   * Called when SessionWatcher fires onDidChange.
   */
  private _onSessionChange(): void {
    if (this._disposed) {
      return;
    }

    if (this._sessionScanning) {
      this._sessionDirty = true;
      return;
    }

    this._runSessionScan();
  }

  /**
   * Called when WorkflowWatcher fires onDidChange.
   */
  private _onWorkflowChange(): void {
    if (this._disposed) {
      return;
    }

    if (this._workflowScanning) {
      this._workflowDirty = true;
      return;
    }

    this._runWorkflowScan();
  }

  /**
   * Runs a session scan cycle, re-running if dirty flag is set.
   */
  private async _runSessionScan(): Promise<void> {
    this._sessionScanning = true;

    try {
      const sessions = await this._deps.scanSessions(this._projectRoot, this._logger);

      if (this._disposed) {
        this._sessionScanning = false;
        return;
      }

      this._sessions = sessions;
      this._stateEmitter.fire({
        sessions: this._sessions,
        workflows: this._workflows,
      });
    } finally {
      this._sessionScanning = false;
    }

    // If dirty flag was set during scan, run again
    if (this._sessionDirty) {
      this._sessionDirty = false;
      if (!this._disposed) {
        this._runSessionScan();
      }
    }
  }

  /**
   * Runs a workflow scan cycle, re-running if dirty flag is set.
   */
  private async _runWorkflowScan(): Promise<void> {
    this._workflowScanning = true;

    try {
      const workflows = await this._deps.scanWorkflows(this._projectRoot, this._logger);

      if (this._disposed) {
        this._workflowScanning = false;
        return;
      }

      this._workflows = workflows;
      this._stateEmitter.fire({
        sessions: this._sessions,
        workflows: this._workflows,
      });
    } finally {
      this._workflowScanning = false;
    }

    // If dirty flag was set during scan, run again
    if (this._workflowDirty) {
      this._workflowDirty = false;
      if (!this._disposed) {
        this._runWorkflowScan();
      }
    }
  }
}
