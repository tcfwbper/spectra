/**
 * Unit tests for EventScanner.
 *
 * Test spec: spec/test/vscode/src/services/eventScanner.md
 * Source under test: vscode/src/services/eventScanner.ts
 *
 * Scaffolded: The production module `eventScanner.ts` does not exist yet.
 * Tests are structured to be compile-ready once EventScanner is implemented
 * and exports `EventScanner` with a static `scanEvents` method.
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockLogger,
  createFsStubs,
  type FsStubs,
} from "./helpers/fsStubs";

/**
 * TODO: Replace with actual import once production module exists:
 *   import { EventScanner } from "../../src/services/eventScanner";
 *
 * Scaffolded interface matching the logic spec contract.
 * The static method signature is:
 *   static async scanEvents(projectRoot: string, sessionId: string, logger: { warn(msg: string): void }): Promise<EventSummary[]>
 */
interface EventSummary {
  Type: string;
  EmittedBy: string;
  Message: string;
}

/**
 * Scaffold: provides a placeholder `scanEvents` that exercises the fs stubs.
 * This will be replaced by the real import once the production file exists.
 *
 * Missing production symbol: EventScanner (from ../../src/services/eventScanner)
 */
function createScanEventsWithStubs(fsStubs: FsStubs) {
  return async function scanEvents(
    projectRoot: string,
    sessionId: string,
    logger: { warn(msg: string): void },
  ): Promise<EventSummary[]> {
    const path = `${projectRoot}/.spectra/sessions/${sessionId}/events.jsonl`;

    // Check file existence
    try {
      await fsStubs.access(path);
    } catch {
      logger.warn(`Events file not found: ${path}`);
      return [];
    }

    // Read file content
    let content: string;
    try {
      content = await fsStubs.readFile(path, "utf-8");
    } catch {
      logger.warn(`Cannot read events file: ${path}`);
      return [];
    }

    // Parse lines
    const lines = content.split("\n");
    const results: EventSummary[] = [];

    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed === "") {
        continue;
      }

      let parsed: any;
      try {
        parsed = JSON.parse(trimmed);
      } catch {
        logger.warn(`Invalid JSON in events file: ${trimmed}`);
        continue;
      }

      if (!parsed.Type || !parsed.EmittedBy || !parsed.Message) {
        if (
          parsed.Type === undefined ||
          parsed.EmittedBy === undefined ||
          parsed.Message === undefined
        ) {
          logger.warn(`Missing required key in event line`);
          continue;
        }
      }

      results.push({
        Type: parsed.Type,
        EmittedBy: parsed.EmittedBy,
        Message: parsed.Message,
      });
    }

    return results;
  };
}

describe("EventScanner", () => {
  let sandbox: sinon.SinonSandbox;
  let fsStubs: FsStubs;
  let scanEvents: ReturnType<typeof createScanEventsWithStubs>;

  beforeEach(() => {
    sandbox = sinon.createSandbox();
    fsStubs = createFsStubs();
    scanEvents = createScanEventsWithStubs(fsStubs);
  });

  afterEach(() => {
    sandbox.restore();
  });

  describe("Happy Path — scanEvents", () => {
    it("should return array of event summaries from valid events file", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves(
        '{"Type":"ReviewNeeded","EmittedBy":"architect","Message":"done"}\n{"Type":"Error","EmittedBy":"runner","Message":"fail"}',
      );
      const logger = createMockLogger();

      const result = await scanEvents("/project", "abc-123", logger);

      expect(result).to.deep.equal([
        { Type: "ReviewNeeded", EmittedBy: "architect", Message: "done" },
        { Type: "Error", EmittedBy: "runner", Message: "fail" },
      ]);
    });

    it("should return single-element array for file with one valid line and no trailing newline", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves(
        '{"Type":"Info","EmittedBy":"node1","Message":"hello"}',
      );
      const logger = createMockLogger();

      const result = await scanEvents("/project", "s1", logger);

      expect(result).to.deep.equal([
        { Type: "Info", EmittedBy: "node1", Message: "hello" },
      ]);
    });

    it("should extract only Type, EmittedBy, and Message from lines with extra keys", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves(
        '{"Type":"X","EmittedBy":"Y","Message":"Z","extra":"ignored","count":42}\n',
      );
      const logger = createMockLogger();

      const result = await scanEvents("/project", "s2", logger);

      expect(result).to.have.lengthOf(1);
      expect(result[0]).to.deep.equal({
        Type: "X",
        EmittedBy: "Y",
        Message: "Z",
      });
      expect(result[0]).to.not.have.property("extra");
      expect(result[0]).to.not.have.property("count");
    });

    it("should preserve file order in returned array", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves(
        '{"Type":"A","EmittedBy":"n","Message":"m"}\n{"Type":"B","EmittedBy":"n","Message":"m"}\n{"Type":"C","EmittedBy":"n","Message":"m"}',
      );
      const logger = createMockLogger();

      const result = await scanEvents("/project", "s3", logger);

      expect(result).to.have.lengthOf(3);
      expect(result[0].Type).to.equal("A");
      expect(result[1].Type).to.equal("B");
      expect(result[2].Type).to.equal("C");
    });
  });

  describe("Null / Empty Input", () => {
    it("should return empty array when file does not exist", async () => {
      fsStubs.access.rejects(new Error("ENOENT: file not found"));
      const logger = createMockLogger();

      const result = await scanEvents("/project", "no-such", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.calledOnce).to.be.true;
      expect(logger.warn.firstCall.args[0]).to.be.a("string").and.not.be.empty;
    });

    it("should return empty array for completely empty file", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves("");
      const logger = createMockLogger();

      const result = await scanEvents("/project", "empty", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.called).to.be.false;
    });

    it("should return empty array for file with only whitespace lines", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves("  \n\n   \n");
      const logger = createMockLogger();

      const result = await scanEvents("/project", "blanks", logger);

      expect(result).to.deep.equal([]);
      expect(logger.warn.called).to.be.false;
    });

    it("should silently skip trailing newline", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves('{"Type":"A","EmittedBy":"B","Message":"C"}\n');
      const logger = createMockLogger();

      const result = await scanEvents("/project", "trail", logger);

      expect(result).to.have.lengthOf(1);
      expect(logger.warn.called).to.be.false;
    });
  });

  describe("Error Propagation", () => {
    it("should warn and skip line when JSON parsing fails", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves(
        'not-json\n{"Type":"OK","EmittedBy":"n","Message":"m"}\n',
      );
      const logger = createMockLogger();

      const result = await scanEvents("/project", "bad", logger);

      expect(result).to.deep.equal([
        { Type: "OK", EmittedBy: "n", Message: "m" },
      ]);
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should warn and skip line when required key is missing", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves(
        '{"Type":"X","EmittedBy":"Y"}\n{"Type":"A","EmittedBy":"B","Message":"C"}\n',
      );
      const logger = createMockLogger();

      const result = await scanEvents("/project", "missing", logger);

      expect(result).to.deep.equal([
        { Type: "A", EmittedBy: "B", Message: "C" },
      ]);
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should never throw to the caller", async () => {
      fsStubs.access.rejects(new Error("EACCES: permission denied"));
      const logger = createMockLogger();

      const result = await scanEvents("/project", "noperm", logger);

      expect(result).to.deep.equal([]);
    });
  });

  describe("Mock / Dependency Interaction", () => {
    it("should construct correct file path from projectRoot and sessionId", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves("");
      const logger = createMockLogger();

      await scanEvents("/my/root", "sess-42", logger);

      // Verify the path passed to readFile
      const expectedPath = "/my/root/.spectra/sessions/sess-42/events.jsonl";
      // access is called first with the path
      expect(fsStubs.access.calledWith(expectedPath)).to.be.true;
    });

    it("should call logger.warn with descriptive message on missing file", async () => {
      fsStubs.access.rejects(new Error("ENOENT"));
      const logger = createMockLogger();

      await scanEvents("/project", "gone", logger);

      expect(logger.warn.calledOnce).to.be.true;
      expect(logger.warn.firstCall.args[0]).to.be.a("string").and.not.be.empty;
    });

    it("should not call any write operations on fs", async () => {
      fsStubs.access.resolves();
      fsStubs.readFile.resolves('{"Type":"A","EmittedBy":"B","Message":"C"}\n');
      const logger = createMockLogger();

      await scanEvents("/project", "s1", logger);

      expect(fsStubs.writeFile.called).to.be.false;
      expect(fsStubs.mkdir.called).to.be.false;
      expect(fsStubs.unlink.called).to.be.false;
    });
  });
});
