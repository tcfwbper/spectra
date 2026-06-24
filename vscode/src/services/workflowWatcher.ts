/**
 * WorkflowWatcher — monitors all *.yaml files under
 * <projectRoot>/.spectra/workflows/ for creation and deletion (not modification).
 * Emits a debounced notification event.
 *
 * Logic spec: spec/logic/vscode/src/services/workflowWatcher.md
 */

import type {
  Disposable,
  Event,
  IEventEmitter,
  IFileSystemWatcher,
} from "./eventWatcher";

/**
 * Injectable VS Code namespace dependencies required by WorkflowWatcher.
 */
export interface WorkflowWatcherDeps {
  createFileSystemWatcher(pattern: any): IFileSystemWatcher;
  createRelativePattern(base: string, pattern: string): any;
  createEventEmitter<T>(): IEventEmitter<T>;
}

/**
 * Monitors all *.yaml files under <projectRoot>/.spectra/workflows/
 * for creation and deletion only. Emits a debounced notification.
 *
 * - Owns: FileSystemWatcher lifecycle, debounce logic, EventEmitter lifecycle.
 * - Delegates: workflow listing to WorkflowScanner.
 * - Must not: read, write, create, or delete any file.
 * - Must not: fire notifications on file content modifications.
 */
export class WorkflowWatcher implements Disposable {
  private readonly _projectRoot: string;
  private readonly _emitter: IEventEmitter<void>;
  private readonly _watcher: IFileSystemWatcher;
  private _debounceTimer: ReturnType<typeof setTimeout> | null = null;
  private _disposed = false;

  /**
   * The public event that fires when yaml files are created or deleted (debounced 300ms).
   */
  public readonly onDidChange: Event<void>;

  constructor(projectRoot: string, deps?: WorkflowWatcherDeps) {
    this._projectRoot = projectRoot;

    if (deps) {
      this._emitter = deps.createEventEmitter<void>();
      const pattern = deps.createRelativePattern(
        projectRoot,
        `.spectra/workflows/*.yaml`,
      );
      this._watcher = deps.createFileSystemWatcher(pattern);
    } else {
      throw new Error("WorkflowWatcher requires deps parameter");
    }

    this.onDidChange = this._emitter.event;

    // Subscribe only to onDidCreate and onDidDelete (NOT onDidChange)
    this._watcher.onDidCreate(() => this._handleFileChange());
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
