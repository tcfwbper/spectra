/**
 * Shared test helpers for stubbing child_process, process.kill, crypto,
 * and VS Code configuration used by EventDispatcher, SessionLauncher,
 * and SessionTerminator tests.
 *
 * These helpers isolate process-spawning concerns so that individual
 * test files remain focused on assertions.
 */
import * as sinon from "sinon";

/**
 * Callback type for Node.js EventEmitter-style event handlers.
 */
type EventHandler = (...args: any[]) => void;

/**
 * A mock child process that behaves as a minimal EventEmitter
 * with `error` and `exit` events, plus optional `unref`.
 */
export interface MockChildProcess {
  /** Register an event handler. */
  on(event: string, handler: EventHandler): MockChildProcess;
  /** Programmatically emit an event (test utility). */
  emit(event: string, ...args: any[]): void;
  /** Spy for unref (used by SessionLauncher). */
  unref: sinon.SinonStub;
  /** Internal map of registered handlers (test inspection). */
  _handlers: Map<string, EventHandler[]>;
}

/**
 * Creates a mock child process with EventEmitter-like on/emit and an unref stub.
 */
export function createMockChildProcess(): MockChildProcess {
  const handlers = new Map<string, EventHandler[]>();

  const cp: MockChildProcess = {
    _handlers: handlers,
    on(event: string, handler: EventHandler) {
      if (!handlers.has(event)) {
        handlers.set(event, []);
      }
      handlers.get(event)!.push(handler);
      return cp;
    },
    emit(event: string, ...args: any[]) {
      const list = handlers.get(event);
      if (list) {
        for (const h of list) {
          h(...args);
        }
      }
    },
    unref: sinon.stub(),
  };

  return cp;
}

/**
 * Logger interface matching the shape required by EventDispatcher,
 * SessionLauncher, and SessionTerminator.
 */
export interface MockLogger {
  info: sinon.SinonSpy;
  warn: sinon.SinonSpy;
  error: sinon.SinonSpy;
}

/**
 * Creates a mock logger with sinon spies on info, warn, and error.
 */
export function createMockServiceLogger(): MockLogger {
  return {
    info: sinon.spy(),
    warn: sinon.spy(),
    error: sinon.spy(),
  };
}

/**
 * Dependencies interface for EventDispatcher tests.
 * Production code is expected to accept this shape for testability.
 */
export interface EventDispatcherDeps {
  getConfiguration: sinon.SinonStub;
  execFile: sinon.SinonStub;
}

/**
 * Creates stubbed dependencies for EventDispatcher.
 * @param binaryPath - Value returned by config.get('agentBinaryPath')
 * @param mockChild - The mock child process returned by execFile
 */
export function createEventDispatcherDeps(
  binaryPath: string | undefined,
  mockChild: MockChildProcess,
): EventDispatcherDeps {
  const configGetStub = sinon.stub();
  configGetStub.withArgs("agentBinaryPath").returns(binaryPath);

  const getConfigurationStub = sinon.stub().returns({ get: configGetStub });
  const execFileStub = sinon.stub().returns(mockChild);

  return {
    getConfiguration: getConfigurationStub,
    execFile: execFileStub,
  };
}

/**
 * Dependencies interface for SessionLauncher tests.
 * Production code is expected to accept this shape for testability.
 */
export interface SessionLauncherDeps {
  getConfiguration: sinon.SinonStub;
  spawn: sinon.SinonStub;
  randomUUID: sinon.SinonStub;
}

/**
 * Creates stubbed dependencies for SessionLauncher.
 * @param binaryPath - Value returned by config.get('binaryPath')
 * @param mockChild - The mock child process returned by spawn
 * @param uuid - Value returned by randomUUID
 */
export function createSessionLauncherDeps(
  binaryPath: string | undefined,
  mockChild: MockChildProcess,
  uuid: string,
): SessionLauncherDeps {
  const configGetStub = sinon.stub();
  configGetStub.withArgs("binaryPath").returns(binaryPath);

  const getConfigurationStub = sinon.stub().returns({ get: configGetStub });
  const spawnStub = sinon.stub().returns(mockChild);
  const randomUUIDStub = sinon.stub().returns(uuid);

  return {
    getConfiguration: getConfigurationStub,
    spawn: spawnStub,
    randomUUID: randomUUIDStub,
  };
}

/**
 * Dependencies interface for SessionTerminator tests.
 * Production code is expected to accept this shape for testability.
 */
export interface SessionTerminatorDeps {
  getConfiguration: sinon.SinonStub;
  processKill: sinon.SinonStub;
  execFile: sinon.SinonStub;
}

/**
 * Creates stubbed dependencies for SessionTerminator.
 * @param binaryPath - Value returned by config.get('binaryPath')
 */
export function createSessionTerminatorDeps(
  binaryPath: string | undefined,
): SessionTerminatorDeps {
  const configGetStub = sinon.stub();
  configGetStub.withArgs("binaryPath").returns(binaryPath);

  const getConfigurationStub = sinon.stub().returns({ get: configGetStub });
  const processKillStub = sinon.stub();
  const execFileStub = sinon.stub();

  return {
    getConfiguration: getConfigurationStub,
    processKill: processKillStub,
    execFile: execFileStub,
  };
}

/**
 * Helper to make a processKill stub throw ESRCH (process not found).
 */
export function makeProcessNotFound(
  processKillStub: sinon.SinonStub,
  pid?: number,
): void {
  const err: any = new Error("kill ESRCH");
  err.code = "ESRCH";
  if (pid !== undefined) {
    processKillStub.withArgs(pid, 0).throws(err);
  } else {
    processKillStub.throws(err);
  }
}

/**
 * Helper to make a processKill stub throw EPERM (permission denied).
 */
export function makeProcessPermDenied(
  processKillStub: sinon.SinonStub,
  pid: number,
  signal: string,
): void {
  const err: any = new Error("kill EPERM");
  err.code = "EPERM";
  processKillStub.withArgs(pid, signal).throws(err);
}

/**
 * Helper to make execFile for `ps` resolve with a given command name.
 */
export function makePsReturn(
  execFileStub: sinon.SinonStub,
  commandName: string,
): void {
  execFileStub.callsFake(
    (
      _cmd: string,
      _args: string[],
      callback: (err: Error | null, stdout: string) => void,
    ) => {
      callback(null, commandName + "\n");
    },
  );
}

/**
 * Helper to make execFile for `ps` reject with an error.
 */
export function makePsFail(
  execFileStub: sinon.SinonStub,
  errorMessage: string,
): void {
  execFileStub.callsFake(
    (
      _cmd: string,
      _args: string[],
      callback: (err: Error | null, stdout: string) => void,
    ) => {
      callback(new Error(errorMessage), "");
    },
  );
}
