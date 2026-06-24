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
 * for creation, modification, and deletion. Also monitors session directory
 * creation and deletion. Emits a debounced notification.
 *
 * - Owns: FileSystemWatcher lifecycle (two watchers), debounce logic, EventEmitter lifecycle.
 * - Delegates: session data reading to SessionScanner.
 * - Must not: read, write, create, or delete any file.
 */
export class SessionWatcher implements Disposable {
  private readonly _projectRoot: string;
  private readonly _emitter: IEventEmitter<void>;
  private readonly _fileWatcher: IFileSystemWatcher;
  private readonly _dirWatcher: IFileSystemWatcher;
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

      // File watcher: session.json files
      const filePattern = deps.createRelativePattern(
        projectRoot,
        `.spectra/sessions/*/session.json`,
      );
      this._fileWatcher = deps.createFileSystemWatcher(filePattern);

      // Directory watcher: session directories
      const dirPattern = deps.createRelativePattern(
        projectRoot,
        `.spectra/sessions/*`,
      );
      this._dirWatcher = deps.createFileSystemWatcher(dirPattern);
    } else {
      throw new Error("SessionWatcher requires deps parameter");
    }

    this.onDidChange = this._emitter.event;

    // Subscribe to file watcher events (create, change, delete)
    this._fileWatcher.onDidCreate(() => this._handleFileChange());
    this._fileWatcher.onDidChange(() => this._handleFileChange());
    this._fileWatcher.onDidDelete(() => this._handleFileChange());

    // Subscribe to directory watcher events (create, delete)
    this._dirWatcher.onDidCreate(() => this._handleFileChange());
    this._dirWatcher.onDidDelete(() => this._handleFileChange());
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

    this._fileWatcher.dispose();
    this._dirWatcher.dispose();
    this._emitter.dispose();
  }
}
