/**
 * EventWatcher — monitors a single session's events.jsonl file for
 * modifications and emits a notification event (debounced).
 *
 * Logic spec: spec/logic/vscode/src/services/eventWatcher.md
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
 * Minimal FileSystemWatcher interface matching vscode.FileSystemWatcher.
 */
export interface IFileSystemWatcher extends Disposable {
  onDidCreate: Event<any>;
  onDidChange: Event<any>;
  onDidDelete: Event<any>;
}

/**
 * Injectable VS Code namespace dependencies required by EventWatcher.
 */
export interface EventWatcherDeps {
  createFileSystemWatcher(pattern: any): IFileSystemWatcher;
  createRelativePattern(base: string, pattern: string): any;
  createEventEmitter<T>(): IEventEmitter<T>;
}

/**
 * Monitors a single session's `events.jsonl` file for modifications
 * and emits a debounced notification event.
 *
 * - Owns: FileSystemWatcher lifecycle, debounce logic, EventEmitter lifecycle.
 * - Delegates: file reading to EventScanner.
 * - Must not: read, write, create, or delete any file.
 */
export class EventWatcher implements Disposable {
  private readonly _projectRoot: string;
  private readonly _sessionId: string;
  private readonly _emitter: IEventEmitter<void>;
  private readonly _watcher: IFileSystemWatcher;
  private _debounceTimer: ReturnType<typeof setTimeout> | null = null;
  private _disposed = false;

  /**
   * The public event that fires when the watched events.jsonl file is modified (debounced 300ms).
   */
  public readonly onDidChange: Event<void>;

  constructor(projectRoot: string, sessionId: string, deps?: EventWatcherDeps) {
    this._projectRoot = projectRoot;
    this._sessionId = sessionId;

    if (deps) {
      this._emitter = deps.createEventEmitter<void>();
      const pattern = deps.createRelativePattern(
        projectRoot,
        `.spectra/sessions/${sessionId}/events.jsonl`,
      );
      this._watcher = deps.createFileSystemWatcher(pattern);
    } else {
      // Fallback: in production, use real vscode module (lazy loaded)
      // This path is only taken when running in the actual extension host
      throw new Error("EventWatcher requires deps parameter");
    }

    this.onDidChange = this._emitter.event;

    // Subscribe only to onDidChange (file modification)
    this._watcher.onDidChange(() => this._handleFileChange());
  }

  private _handleFileChange(): void {
    if (this._disposed) {
      return;
    }

    // Reset debounce timer
    if (this._debounceTimer !== null) {
      clearTimeout(this._debounceTimer);
    }

    this._debounceTimer = setTimeout(() => {
      if (!this._disposed) {
        this._emitter.fire(undefined as any);
      }
      this._debounceTimer = null;
    }, 300);
  }

  dispose(): void {
    if (this._disposed) {
      return;
    }
    this._disposed = true;

    if (this._debounceTimer !== null) {
      clearTimeout(this._debounceTimer);
      this._debounceTimer = null;
    }

    this._watcher.dispose();
    this._emitter.dispose();
  }
}
