/**
 * Shared test helpers for stubbing Node.js `fs/promises` methods
 * used by scanner services.
 *
 * These helpers provide consistent fs stubbing patterns so that
 * individual scanner test files remain focused on assertions.
 */
import * as sinon from "sinon";

/**
 * Minimal Dirent-like shape returned by fs.readdir with { withFileTypes: true }.
 */
export interface FakeDirent {
  name: string;
  isDirectory(): boolean;
  isFile(): boolean;
}

/**
 * Creates a FakeDirent representing a directory.
 */
export function fakeDirEntry(name: string): FakeDirent {
  return {
    name,
    isDirectory: () => true,
    isFile: () => false,
  };
}

/**
 * Creates a FakeDirent representing a regular file.
 */
export function fakeFileEntry(name: string): FakeDirent {
  return {
    name,
    isDirectory: () => false,
    isFile: () => true,
  };
}

/**
 * Creates a mock logger matching the `{ warn(msg: string): void }` interface.
 * Returns an object with a sinon spy on `warn`.
 */
export function createMockLogger(): { warn: sinon.SinonSpy } {
  return {
    warn: sinon.spy(),
  };
}

/**
 * Represents a stub collection for fs/promises operations.
 * Scanner tests stub these to control filesystem behavior.
 */
export interface FsStubs {
  access: sinon.SinonStub;
  readFile: sinon.SinonStub;
  readdir: sinon.SinonStub;
  writeFile: sinon.SinonStub;
  mkdir: sinon.SinonStub;
  unlink: sinon.SinonStub;
}

/**
 * Creates a complete set of fs/promises stubs.
 * By default all stubs reject — callers configure specific behaviors per test.
 */
export function createFsStubs(): FsStubs {
  return {
    access: sinon.stub(),
    readFile: sinon.stub(),
    readdir: sinon.stub(),
    writeFile: sinon.stub(),
    mkdir: sinon.stub(),
    unlink: sinon.stub(),
  };
}
