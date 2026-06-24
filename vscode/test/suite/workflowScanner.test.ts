/**
 * Unit tests for WorkflowScanner.
 *
 * Test spec: spec/test/vscode/src/services/workflowScanner.md
 * Source under test: vscode/src/services/workflowScanner.ts
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockLogger,
  createFsStubs,
  fakeDirEntry,
  fakeFileEntry,
  type FsStubs,
} from "./helpers/fsStubs";

import { WorkflowScanner } from "../../src/services/workflowScanner";

/**
 * Creates a bound scanWorkflows function that injects fs stubs into the
 * production WorkflowScanner.scanWorkflows static method.
 */
function createScanWorkflowsWithStubs(fsStubs: FsStubs) {
  return function scanWorkflows(
    projectRoot: string,
    logger: { warn(msg: string): void },
  ) {
    return WorkflowScanner.scanWorkflows(projectRoot, logger, fsStubs);
  };
}

describe("WorkflowScanner", () => {
  let sandbox: sinon.SinonSandbox;
  let fsStubs: FsStubs;
  let scanWorkflows: ReturnType<typeof createScanWorkflowsWithStubs>;

  beforeEach(() => {
    sandbox = sinon.createSandbox();
    fsStubs = createFsStubs();
    scanWorkflows = createScanWorkflowsWithStubs(fsStubs);
  });

  afterEach(() => {
    sandbox.restore();
  });

  describe("Happy Path — scanWorkflows", () => {
    it("should return array of workflow names without yaml extension", async () => {
      fsStubs.readdir.resolves([
        fakeFileEntry("build.yaml"),
        fakeFileEntry("deploy.yaml"),
      ]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal(["build", "deploy"]);
    });

    it("should include only yaml files and exclude other file types", async () => {
      fsStubs.readdir.resolves([
        fakeFileEntry("workflow.yaml"),
        fakeFileEntry("README.md"),
        fakeFileEntry("config.json"),
      ]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal(["workflow"]);
    });

    it("should return empty string element for file named dot-yaml only", async () => {
      fsStubs.readdir.resolves([fakeFileEntry(".yaml")]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal([""]);
    });
  });

  describe("Null / Empty Input", () => {
    it("should return empty array when workflows directory does not exist", async () => {
      const enoent = Object.assign(new Error("ENOENT"), { code: "ENOENT" });
      fsStubs.readdir.rejects(enoent);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.calledOnce).to.be.true;
      expect(logger.warn.firstCall.args[0]).to.be.a("string").and.not.be.empty;
    });

    it("should return empty array when directory exists but contains no yaml files", async () => {
      fsStubs.readdir.resolves([
        fakeFileEntry("README.md"),
        fakeFileEntry("notes.txt"),
      ]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.called).to.be.false;
    });

    it("should return empty array when directory is empty", async () => {
      fsStubs.readdir.resolves([]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.called).to.be.false;
    });
  });

  describe("Error Propagation", () => {
    it("should never throw to the caller", async () => {
      const eacces = Object.assign(new Error("EACCES"), { code: "EACCES" });
      fsStubs.readdir.rejects(eacces);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal([]);
    });

    it("should exclude subdirectories even if they have yaml in their name", async () => {
      fsStubs.readdir.resolves([
        fakeDirEntry("subdir.yaml"),
        fakeFileEntry("real.yaml"),
      ]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal(["real"]);
    });
  });

  describe("Mock / Dependency Interaction", () => {
    it("should construct correct workflows directory path", async () => {
      fsStubs.readdir.resolves([]);
      const logger = createMockLogger();

      await scanWorkflows("/my/root", logger);

      expect(
        fsStubs.readdir.calledWith(
          "/my/root/.spectra/workflows",
          sinon.match.any,
        ),
      ).to.be.true;
    });

    it("should not call any write operations on fs", async () => {
      fsStubs.readdir.resolves([fakeFileEntry("test.yaml")]);
      const logger = createMockLogger();

      await scanWorkflows("/project", logger);

      expect(fsStubs.writeFile.called).to.be.false;
      expect(fsStubs.mkdir.called).to.be.false;
      expect(fsStubs.unlink.called).to.be.false;
    });

    it("should not read file contents", async () => {
      fsStubs.readdir.resolves([
        fakeFileEntry("a.yaml"),
        fakeFileEntry("b.yaml"),
      ]);
      const logger = createMockLogger();

      await scanWorkflows("/project", logger);

      expect(fsStubs.readFile.called).to.be.false;
    });

    it("should not recurse into subdirectories", async () => {
      fsStubs.readdir.resolves([
        fakeDirEntry("nested"),
        fakeFileEntry("top.yaml"),
      ]);
      const logger = createMockLogger();

      const result = await scanWorkflows("/project", logger);

      expect(result).to.deep.equal(["top"]);
      // readdir should be called exactly once (for the workflows directory only)
      expect(fsStubs.readdir.calledOnce).to.be.true;
    });
  });
});
