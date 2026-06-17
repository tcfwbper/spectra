// SessionWatcher — monitors all session.json files under
// <projectRoot>/.spectra/sessions/*/ for creation, modification, and deletion.
// Emits a debounced notification event.
//
// Logic spec: spec/logic/vscode/src/services/sessionWatcher.md

import type {
  Disposable,
  Event,
  IEventEmitter,
  IFileSystemWatcher,
} from "./eventWatcher";

/**
 * Injectable VS Code namespace dependencies required by SessionWatcher.
 */
export interface SessionWatcherDeps {
  createFileSystemWatcher(pattern: any): IFileSystemWatcher;
  createRelativePattern(base: string, pattern: string): any;
  createEventEmitter<T>(): IEventEmitter<T>;
}

/**
 * Monitors all session.json files under projectRoot/.spectra/sessions/
 * for creation, modification, and deletion. Emits a debounced notification.
 *
 * - Owns: FileSystemWatcher lifecycle, debounce logic, EventEmitter lifecycle.
 * - Delegates: session data reading to SessionScanner.
 * - Must not: read, write, create, or delete any file.
 */
export class SessionWatcher implements Disposable {
  private readonly _projectRoot: string;
  private readonly _emitter: IEventEmitter<void>;
  private readonly _watcher: IFileSystemWatcher;
  private _debounceTimer: ReturnType<typeof setTimeout> | null = null;
  private _disposed = false;

  /**
   * The public event that fires when any session.json file changes (debounced 300ms).
   */
  public readonly onDidChange: Event<void>;

  constructor(projectRoot: string, deps?: SessionWatcherDeps) {
    this._projectRoot = projectRoot;

    if (deps) {
      this._emitter = deps.createEventEmitter<void>();
      const pattern = deps.createRelativePattern(
        projectRoot,
        `.spectra/sessions/*/session.json`,
      );
      this._watcher = deps.createFileSystemWatcher(pattern);
    } else {
      throw new Error("SessionWatcher requires deps parameter");
    }

    this.onDidChange = this._emitter.event;

    // Subscribe to all three event types
    this._watcher.onDidCreate(() => this._handleFileChange());
    this._watcher.onDidChange(() => this._handleFileChange());
    this._watcher.onDidDelete(() => this._handleFileChange());
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
