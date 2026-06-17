/**
 * Unit tests for WorkflowDefinitionParser.
 *
 * Test spec: spec/test/vscode/src/services/workflowDefinitionParser.md
 * Source under test: vscode/src/services/workflowDefinitionParser.ts
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createMockLogger,
  createFsStubs,
  type FsStubs,
} from "./helpers/fsStubs";

import { WorkflowDefinitionParser } from "../../src/services/workflowDefinitionParser";

/**
 * Builds a valid YAML string from structured input.
 * Uses plain string construction to avoid needing a YAML library in tests.
 */
function buildYaml(opts: {
  entryNode?: string | number;
  transitions?: Array<Record<string, unknown>>;
  extras?: Record<string, unknown>;
}): string {
  const lines: string[] = [];

  if (opts.extras) {
    for (const [key, value] of Object.entries(opts.extras)) {
      lines.push(`${key}: ${JSON.stringify(value)}`);
    }
  }

  if (opts.entryNode !== undefined) {
    if (typeof opts.entryNode === "number") {
      lines.push(`entryNode: ${opts.entryNode}`);
    } else {
      lines.push(`entryNode: '${opts.entryNode}'`);
    }
  }

  if (opts.transitions !== undefined) {
    lines.push("transitions:");
    if (opts.transitions.length === 0) {
      lines.push("  []");
    } else {
      for (const t of opts.transitions) {
        const parts: string[] = [];
        for (const [key, value] of Object.entries(t)) {
          if (typeof value === "number") {
            parts.push(`${key}: ${value}`);
          } else if (Array.isArray(value)) {
            parts.push(`${key}: ${JSON.stringify(value)}`);
          } else {
            parts.push(`${key}: '${value}'`);
          }
        }
        lines.push(`  - {${parts.join(", ")}}`);
      }
    }
  }

  return lines.join("\n");
}

/**
 * Creates a bound parseWorkflowDefinition function that injects fs stubs.
 *
 * Reason: staged scaffold replacement — production module now exists.
 */
function createParseWithStubs(fsStubs: FsStubs) {
  return async function parseWorkflowDefinition(
    projectRoot: string,
    workflowName: string,
    logger: { warn(msg: string): void },
  ): Promise<{ entryNode: string; eventTypes: string[] }> {
    return WorkflowDefinitionParser.parseWorkflowDefinition(
      projectRoot,
      workflowName,
      logger,
      fsStubs,
    );
  };
}

describe("WorkflowDefinitionParser", () => {
  let sandbox: sinon.SinonSandbox;
  let fsStubs: FsStubs;
  let parseWorkflowDefinition: ReturnType<typeof createParseWithStubs>;

  beforeEach(() => {
    sandbox = sinon.createSandbox();
    fsStubs = createFsStubs();
    parseWorkflowDefinition = createParseWithStubs(fsStubs);
  });

  afterEach(() => {
    sandbox.restore();
  });

  describe("Happy Path — parseWorkflowDefinition", () => {
    it("should return entryNode and matching eventTypes from valid workflow", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [
          { fromNode: "start", eventType: "init" },
          { fromNode: "start", eventType: "retry" },
          { fromNode: "other", eventType: "done" },
        ],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "deploy",
        logger,
      );

      expect(result).to.deep.equal({
        entryNode: "start",
        eventTypes: ["init", "retry"],
      });
    });

    it("should deduplicate eventTypes when multiple transitions share the same eventType", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "begin",
        transitions: [
          { fromNode: "begin", eventType: "trigger" },
          { fromNode: "begin", eventType: "trigger" },
          { fromNode: "begin", eventType: "other" },
        ],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition("/project", "build", logger);

      expect(result).to.deep.equal({
        entryNode: "begin",
        eventTypes: ["trigger", "other"],
      });
    });

    it("should return empty eventTypes when no transitions match entryNode", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [{ fromNode: "middle", eventType: "proceed" }],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition("/project", "test", logger);

      expect(result).to.deep.equal({ entryNode: "start", eventTypes: [] });
      expect(logger.warn.called).to.be.false;
    });

    it("should return empty eventTypes when transitions array is empty", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition("/project", "empty", logger);

      expect(result).to.deep.equal({ entryNode: "start", eventTypes: [] });
      expect(logger.warn.called).to.be.false;
    });

    it("should ignore unknown top-level keys in YAML", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [{ fromNode: "start", eventType: "go" }],
        extras: { description: "extra", nodes: ["a", "b"] },
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition("/project", "extra", logger);

      expect(result).to.deep.equal({ entryNode: "start", eventTypes: ["go"] });
    });

    it("should ignore unknown keys in transition dicts", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [
          {
            fromNode: "start",
            eventType: "run",
            toNode: "end",
            guard: "check",
          },
        ],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition("/project", "rich", logger);

      expect(result).to.deep.equal({ entryNode: "start", eventTypes: ["run"] });
    });
  });

  describe("Error Propagation", () => {
    it("should return failure result when file does not exist", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const enoent = Object.assign(new Error("ENOENT"), { code: "ENOENT" });
      fsStubs.readFile.rejects(enoent);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "missing",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
      expect(logger.warn.firstCall.args[0]).to.be.a("string").and.not.be.empty;
    });

    it("should return failure result when YAML syntax is invalid", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      fsStubs.readFile.resolves("': invalid: [unterminated");
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "broken",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should return failure result when entryNode key is missing", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = "transitions:\n  - {fromNode: 'a', eventType: 'b'}";
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "noentry",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should return failure result when transitions key is missing", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = "entryNode: 'start'";
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "notrans",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should return failure result when entryNode is not a string", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: 42,
        transitions: [],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "badentry",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should return failure result when transitions is not an array", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = "entryNode: 'start'\ntransitions: 'not-an-array'";
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "badtrans",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should never throw an exception to the caller", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      fsStubs.readFile.rejects(new Error("generic failure"));
      const logger = createMockLogger();

      // Must not throw — should return failure result
      const result = await parseWorkflowDefinition("/project", "crash", logger);

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
    });
  });

  describe("Validation Failures", () => {
    it("should fail fast when a transition dict is missing fromNode", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [
          { eventType: "go" },
          { fromNode: "start", eventType: "valid" },
        ],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "badfrom",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should fail fast when a transition dict is missing eventType", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [
          { fromNode: "start" },
          { fromNode: "start", eventType: "valid" },
        ],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "badevent",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should fail fast when fromNode in a transition is not a string", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [{ fromNode: 123, eventType: "go" }],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "numfrom",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });

    it("should fail fast when eventType in a transition is not a string", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [{ fromNode: "start", eventType: ["array"] }],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      const result = await parseWorkflowDefinition(
        "/project",
        "arrtype",
        logger,
      );

      expect(result).to.deep.equal({ entryNode: "", eventTypes: [] });
      expect(logger.warn.calledOnce).to.be.true;
    });
  });

  describe("Mock / Dependency Interaction", () => {
    it("should construct correct file path from projectRoot and workflowName", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "x",
        transitions: [],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      await parseWorkflowDefinition("/my/root", "deploy", logger);

      expect(
        fsStubs.readFile.calledWith("/my/root/.spectra/workflows/deploy.yaml"),
      ).to.be.true;
    });

    it("should not call any write operations on fs", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [{ fromNode: "start", eventType: "go" }],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      await parseWorkflowDefinition("/project", "test", logger);

      expect(fsStubs.writeFile.called).to.be.false;
      expect(fsStubs.mkdir.called).to.be.false;
      expect(fsStubs.unlink.called).to.be.false;
    });

    it("should call logger.warn exactly once per failure", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const enoent = Object.assign(new Error("ENOENT"), { code: "ENOENT" });
      fsStubs.readFile.rejects(enoent);
      const logger = createMockLogger();

      await parseWorkflowDefinition("/project", "gone", logger);

      expect(logger.warn.callCount).to.equal(1);
    });

    it("should not call logger.warn on successful parse", async () => {
      // Scaffolded: depends on WorkflowDefinitionParser.parseWorkflowDefinition
      const yaml = buildYaml({
        entryNode: "start",
        transitions: [{ fromNode: "start", eventType: "go" }],
      });
      fsStubs.readFile.resolves(yaml);
      const logger = createMockLogger();

      await parseWorkflowDefinition("/project", "ok", logger);

      expect(logger.warn.called).to.be.false;
    });
  });
});
