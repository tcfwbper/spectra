/**
 * Unit tests for SessionScanner.
 *
 * Test spec: spec/test/vscode/src/services/sessionScanner.md
 * Source under test: vscode/src/services/sessionScanner.ts
 *
 * Scaffolded: The production module `sessionScanner.ts` does not exist yet.
 * Tests are structured to be compile-ready once SessionScanner is implemented
 * and exports `SessionScanner` with a static `scanSessions` method.
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

/**
 * TODO: Replace with actual import once production module exists:
 *   import { SessionScanner } from "../../src/services/sessionScanner";
 *
 * Scaffolded interface matching the logic spec contract.
 * The static method signature is:
 *   static async scanSessions(projectRoot: string, logger: { warn(msg: string): void }): Promise<SessionSummary[]>
 */
interface SessionSummary {
  id: string;
  workflowName: string;
  createdAt: number;
  pid: number;
  status: string;
  currentState: string;
}

/**
 * Scaffold: provides a placeholder `scanSessions` that exercises the fs stubs.
 * This will be replaced by the real import once the production file exists.
 *
 * Missing production symbol: SessionScanner (from ../../src/services/sessionScanner)
 */
function createScanSessionsWithStubs(fsStubs: FsStubs) {
  return async function scanSessions(
    projectRoot: string,
    logger: { warn(msg: string): void },
  ): Promise<SessionSummary[]> {
    const sessionsDir = `${projectRoot}/.spectra/sessions`;

    // Check if sessions directory exists
    try {
      await fsStubs.access(sessionsDir);
    } catch {
      logger.warn(`Sessions directory not found: ${sessionsDir}`);
      return [];
    }

    // Read directory entries
    let entries: Array<{
      name: string;
      isDirectory(): boolean;
      isFile(): boolean;
    }>;
    try {
      entries = await fsStubs.readdir(sessionsDir, { withFileTypes: true });
    } catch {
      logger.warn(`Cannot read sessions directory: ${sessionsDir}`);
      return [];
    }

    // Process each subdirectory
    const results: SessionSummary[] = [];
    const requiredKeys = [
      "id",
      "workflowName",
      "createdAt",
      "pid",
      "status",
      "currentState",
    ];

    for (const entry of entries) {
      if (!entry.isDirectory()) {
        continue;
      }

      const sessionJsonPath = `${sessionsDir}/${entry.name}/session.json`;

      let content: string;
      try {
        content = await fsStubs.readFile(sessionJsonPath, "utf-8");
      } catch {
        logger.warn(`Cannot read session.json: ${sessionJsonPath}`);
        continue;
      }

      let parsed: any;
      try {
        parsed = JSON.parse(content);
      } catch {
        logger.warn(`Invalid JSON in session.json: ${sessionJsonPath}`);
        continue;
      }

      // Validate required keys
      const missing = requiredKeys.filter((k) => parsed[k] === undefined);
      if (missing.length > 0) {
        logger.warn(
          `Missing required keys in session.json: ${missing.join(", ")}`,
        );
        continue;
      }

      results.push({
        id: parsed.id,
        workflowName: parsed.workflowName,
        createdAt: parsed.createdAt,
        pid: parsed.pid,
        status: parsed.status,
        currentState: parsed.currentState,
      });
    }

    // Sort by createdAt descending
    results.sort((a, b) => b.createdAt - a.createdAt);

    return results;
  };
}

/**
 * Helper to build a valid session.json content string.
 */
function buildSessionJson(
  overrides: Partial<SessionSummary> & { id: string },
): string {
  const defaults = {
    workflowName: "default",
    createdAt: 1000,
    pid: 1,
    status: "running",
    currentState: "init",
  };
  return JSON.stringify({ ...defaults, ...overrides });
}

describe("SessionScanner", () => {
  let sandbox: sinon.SinonSandbox;
  let fsStubs: FsStubs;
  let scanSessions: ReturnType<typeof createScanSessionsWithStubs>;

  beforeEach(() => {
    sandbox = sinon.createSandbox();
    fsStubs = createFsStubs();
    scanSessions = createScanSessionsWithStubs(fsStubs);
  });

  afterEach(() => {
    sandbox.restore();
  });

  describe("Happy Path — scanSessions", () => {
    it("should return sorted array of session summaries", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([
        fakeDirEntry("sess-1"),
        fakeDirEntry("sess-2"),
      ]);
      fsStubs.readFile
        .withArgs(
          "/project/.spectra/sessions/sess-1/session.json",
          sinon.match.any,
        )
        .resolves(
          JSON.stringify({
            id: "sess-1",
            workflowName: "build",
            createdAt: 1000,
            pid: 100,
            status: "completed",
            currentState: "done",
          }),
        );
      fsStubs.readFile
        .withArgs(
          "/project/.spectra/sessions/sess-2/session.json",
          sinon.match.any,
        )
        .resolves(
          JSON.stringify({
            id: "sess-2",
            workflowName: "test",
            createdAt: 2000,
            pid: 200,
            status: "running",
            currentState: "execute",
          }),
        );
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([
        {
          id: "sess-2",
          workflowName: "test",
          createdAt: 2000,
          pid: 200,
          status: "running",
          currentState: "execute",
        },
        {
          id: "sess-1",
          workflowName: "build",
          createdAt: 1000,
          pid: 100,
          status: "completed",
          currentState: "done",
        },
      ]);
    });

    it("should extract only the six required fields from session.json", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([fakeDirEntry("s1")]);
      fsStubs.readFile.resolves(
        JSON.stringify({
          id: "s1",
          workflowName: "w",
          createdAt: 500,
          pid: 1,
          status: "running",
          currentState: "init",
          extra: "ignored",
          debug: true,
        }),
      );
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.have.lengthOf(1);
      expect(result[0]).to.deep.equal({
        id: "s1",
        workflowName: "w",
        createdAt: 500,
        pid: 1,
        status: "running",
        currentState: "init",
      });
      expect(result[0]).to.not.have.property("extra");
      expect(result[0]).to.not.have.property("debug");
    });

    it("should include sessions with same createdAt without error", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([fakeDirEntry("a"), fakeDirEntry("b")]);
      fsStubs.readFile
        .withArgs("/project/.spectra/sessions/a/session.json", sinon.match.any)
        .resolves(buildSessionJson({ id: "a", createdAt: 1000 }));
      fsStubs.readFile
        .withArgs("/project/.spectra/sessions/b/session.json", sinon.match.any)
        .resolves(buildSessionJson({ id: "b", createdAt: 1000 }));
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.have.lengthOf(2);
      const ids = result.map((s) => s.id).sort();
      expect(ids).to.deep.equal(["a", "b"]);
    });
  });

  describe("Null / Empty Input", () => {
    it("should return empty array when sessions directory does not exist", async () => {
      fsStubs.access.rejects(new Error("ENOENT: file not found"));
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.calledOnce).to.be.true;
      expect(logger.warn.firstCall.args[0]).to.be.a("string").and.not.be.empty;
    });

    it("should return empty array when sessions directory is empty", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([]);
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.called).to.be.false;
    });
  });

  describe("Error Propagation", () => {
    it("should warn and skip session when session.json is missing", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([fakeDirEntry("s1")]);
      fsStubs.readFile.rejects(new Error("ENOENT: no such file"));
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should warn and skip session when session.json is malformed JSON", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([fakeDirEntry("s1")]);
      fsStubs.readFile.resolves("not valid json{");
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should warn and skip session when required key is missing", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([fakeDirEntry("s1")]);
      fsStubs.readFile.resolves(
        JSON.stringify({
          id: "s1",
          workflowName: "w",
          createdAt: 100,
          pid: 1,
          status: "running",
          // missing currentState
        }),
      );
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should never throw to the caller", async () => {
      fsStubs.access.rejects(new Error("EACCES: permission denied"));
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.deep.equal([]);
    });

    it("should skip regular files in sessions directory", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([
        fakeFileEntry("notes.txt"),
        fakeDirEntry("s1"),
      ]);
      fsStubs.readFile.resolves(
        buildSessionJson({
          id: "s1",
          workflowName: "w",
          createdAt: 1000,
          pid: 1,
          status: "running",
          currentState: "init",
        }),
      );
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.have.lengthOf(1);
      expect(result[0].id).to.equal("s1");
      expect(logger.warn.called).to.be.false;
    });
  });

  describe("Mock / Dependency Interaction", () => {
    it("should construct correct sessions directory path", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([]);
      const logger = createMockLogger();

      await scanSessions("/my/root", logger);

      expect(fsStubs.access.calledWith("/my/root/.spectra/sessions")).to.be
        .true;
    });

    it("should read session.json from each session subdirectory", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([
        fakeDirEntry("sess-a"),
        fakeDirEntry("sess-b"),
      ]);
      fsStubs.readFile.resolves(
        buildSessionJson({
          id: "x",
          workflowName: "w",
          createdAt: 1000,
          pid: 1,
          status: "running",
          currentState: "init",
        }),
      );
      const logger = createMockLogger();

      await scanSessions("/root", logger);

      expect(
        fsStubs.readFile.calledWith(
          "/root/.spectra/sessions/sess-a/session.json",
          sinon.match.any,
        ),
      ).to.be.true;
      expect(
        fsStubs.readFile.calledWith(
          "/root/.spectra/sessions/sess-b/session.json",
          sinon.match.any,
        ),
      ).to.be.true;
    });

    it("should not call any write operations on fs", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([fakeDirEntry("s1")]);
      fsStubs.readFile.resolves(
        buildSessionJson({
          id: "s1",
          workflowName: "w",
          createdAt: 1000,
          pid: 1,
          status: "running",
          currentState: "init",
        }),
      );
      const logger = createMockLogger();

      await scanSessions("/project", logger);

      expect(fsStubs.writeFile.called).to.be.false;
      expect(fsStubs.mkdir.called).to.be.false;
      expect(fsStubs.unlink.called).to.be.false;
    });
  });

  describe("Ordering — createdAt descending", () => {
    it("should sort sessions by createdAt in descending order", async () => {
      fsStubs.access.resolves();
      fsStubs.readdir.resolves([
        fakeDirEntry("old"),
        fakeDirEntry("mid"),
        fakeDirEntry("new"),
      ]);
      fsStubs.readFile
        .withArgs(
          "/project/.spectra/sessions/old/session.json",
          sinon.match.any,
        )
        .resolves(buildSessionJson({ id: "old", createdAt: 100 }));
      fsStubs.readFile
        .withArgs(
          "/project/.spectra/sessions/mid/session.json",
          sinon.match.any,
        )
        .resolves(buildSessionJson({ id: "mid", createdAt: 500 }));
      fsStubs.readFile
        .withArgs(
          "/project/.spectra/sessions/new/session.json",
          sinon.match.any,
        )
        .resolves(buildSessionJson({ id: "new", createdAt: 900 }));
      const logger = createMockLogger();

      const result = await scanSessions("/project", logger);

      expect(result).to.have.lengthOf(3);
      expect(result[0].createdAt).to.equal(900);
      expect(result[1].createdAt).to.equal(500);
      expect(result[2].createdAt).to.equal(100);
    });
  });
});
