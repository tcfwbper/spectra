/**
 * WorkflowDefinitionParser — reads and parses a workflow definition YAML file,
 * extracts eventType values from transitions whose fromNode equals the entryNode,
 * deduplicates them, and returns the result.
 *
 * Logic spec: spec/logic/vscode/src/services/workflowDefinitionParser.md
 */

import * as path from "path";

/**
 * Result type for workflow definition parsing.
 */
export interface WorkflowParseResult {
  entryNode: string;
  eventTypes: string[];
}

/**
 * Minimal filesystem interface required by WorkflowDefinitionParser.
 * In production this is satisfied by Node.js fs/promises; in tests a stub is passed.
 */
export interface WorkflowDefinitionParserFs {
  readFile(path: string, encoding: string): Promise<string>;
}

/** Failure result constant. */
const FAILURE_RESULT: WorkflowParseResult = { entryNode: "", eventTypes: [] };

/**
 * Provides a static method that reads and parses a workflow definition YAML file.
 *
 * - Owns: reading a single workflow YAML file from disk.
 * - Owns: validating presence of required top-level keys (entryNode, transitions).
 * - Owns: validating that each transition dict contains required keys (fromNode, eventType).
 * - Owns: filtering transitions by fromNode === entryNode and collecting unique eventType values.
 * - Owns: graceful degradation — returns empty result on failure after logging a warning.
 * - Must not: write, create, or delete any file or directory.
 * - Must not: throw exceptions to the caller.
 */
export class WorkflowDefinitionParser {
  /**
   * Parses a workflow definition file and extracts event types for the entry node.
   *
   * @param projectRoot - Absolute path to the project root.
   * @param workflowName - Name of the workflow (without .yaml extension).
   * @param logger - Logger with a warn method for reporting issues.
   * @param fs - Injectable filesystem dependency (defaults to Node.js fs/promises).
   * @returns WorkflowParseResult containing entryNode and deduplicated eventTypes.
   */
  static async parseWorkflowDefinition(
    projectRoot: string,
    workflowName: string,
    logger: { warn(msg: string): void },
    fs?: WorkflowDefinitionParserFs,
  ): Promise<WorkflowParseResult> {
    try {
      const fsImpl =
        fs ?? ((await import("fs/promises")) as unknown as WorkflowDefinitionParserFs);

      const filePath = path.join(
        projectRoot,
        ".spectra",
        "workflows",
        `${workflowName}.yaml`,
      );

      // Step 3: Read the file
      let content: string;
      try {
        content = await fsImpl.readFile(filePath, "utf-8");
      } catch {
        logger.warn(
          `Workflow definition file not found: ${filePath}`,
        );
        return { ...FAILURE_RESULT };
      }

      // Step 5: Parse YAML
      // eslint-disable-next-line @typescript-eslint/no-require-imports
      const yaml = require("js-yaml");
      let parsed: unknown;
      try {
        parsed = yaml.load(content);
      } catch {
        logger.warn(
          `Invalid YAML syntax in workflow definition: ${filePath}`,
        );
        return { ...FAILURE_RESULT };
      }

      // Step 7: Validate entryNode
      if (
        typeof parsed !== "object" ||
        parsed === null ||
        !("entryNode" in parsed)
      ) {
        logger.warn(
          `Workflow definition missing required key 'entryNode': ${filePath}`,
        );
        return { ...FAILURE_RESULT };
      }

      const doc = parsed as Record<string, unknown>;

      if (typeof doc.entryNode !== "string") {
        logger.warn(
          `Workflow definition 'entryNode' must be a string: ${filePath}`,
        );
        return { ...FAILURE_RESULT };
      }

      const entryNode = doc.entryNode;

      // Step 8: Validate transitions
      if (!("transitions" in doc)) {
        logger.warn(
          `Workflow definition missing required key 'transitions': ${filePath}`,
        );
        return { ...FAILURE_RESULT };
      }

      if (!Array.isArray(doc.transitions)) {
        logger.warn(
          `Workflow definition 'transitions' must be an array: ${filePath}`,
        );
        return { ...FAILURE_RESULT };
      }

      const transitions = doc.transitions as unknown[];

      // Step 9: Validate each transition element
      for (const t of transitions) {
        if (typeof t !== "object" || t === null) {
          logger.warn(
            `Workflow definition contains invalid transition entry: ${filePath}`,
          );
          return { ...FAILURE_RESULT };
        }

        const transition = t as Record<string, unknown>;

        if (!("fromNode" in transition)) {
          logger.warn(
            `Workflow definition transition missing required key 'fromNode': ${filePath}`,
          );
          return { ...FAILURE_RESULT };
        }

        if (typeof transition.fromNode !== "string") {
          logger.warn(
            `Workflow definition transition 'fromNode' must be a string: ${filePath}`,
          );
          return { ...FAILURE_RESULT };
        }

        if (!("eventType" in transition)) {
          logger.warn(
            `Workflow definition transition missing required key 'eventType': ${filePath}`,
          );
          return { ...FAILURE_RESULT };
        }

        if (typeof transition.eventType !== "string") {
          logger.warn(
            `Workflow definition transition 'eventType' must be a string: ${filePath}`,
          );
          return { ...FAILURE_RESULT };
        }
      }

      // Step 10-12: Filter, collect, and deduplicate eventTypes
      const seen = new Set<string>();
      const eventTypes: string[] = [];

      for (const t of transitions) {
        const transition = t as Record<string, unknown>;
        if (transition.fromNode === entryNode) {
          const eventType = transition.eventType as string;
          if (!seen.has(eventType)) {
            seen.add(eventType);
            eventTypes.push(eventType);
          }
        }
      }

      return { entryNode, eventTypes };
    } catch {
      // Must never throw to the caller
      return { ...FAILURE_RESULT };
    }
  }
}
