# WorkflowDefinitionLoader

## Overview

WorkflowDefinitionLoader provides read-only access to workflow definition YAML files stored in `.spectra/workflows/`. It reads and parses a workflow YAML file, derives the `Name` from the filename, constructs all child value objects (Node, Transition, ExitTransition) via their constructors, constructs a WorkflowDefinition via `NewWorkflowDefinition`, and validates agent_role referential integrity via an injected agent loader interface. WorkflowDefinitionLoader does not create, modify, or delete workflow files. It is stateless and does not cache definitions in memory.

## Boundaries

- Owns: reading workflow YAML files from disk via standard file I/O.
- Owns: strict YAML parsing with unknown field rejection.
- Owns: deriving the `Name` value from the YAML filename (stripping `.yaml` extension).
- Owns: constructing child value objects (Node, Transition, ExitTransition) from parsed YAML data.
- Owns: agent_role referential integrity validation (every agent node's role must load successfully via injected loader).
- Owns: error wrapping with layered context (read / parse / validation phases).
- Delegates: field-level format validation for Node, Transition, ExitTransition to their respective constructors.
- Delegates: cross-component structural validation (uniqueness, reachability, determinism) to `NewWorkflowDefinition` constructor.
- Delegates: agent definition field validation to the injected agent loader (which in turn delegates to `NewAgentDefinition`).
- Delegates: path composition to StorageLayout.
- Must not: create, modify, or delete workflow definition files.
- Must not: bypass constructors (must not use struct literals for Node, Transition, ExitTransition, or WorkflowDefinition).
- Must not: cache workflow definitions in memory.
- Must not: validate cross-component constraints that are owned by WorkflowDefinition constructor.
- Must not: use FileAccessor (read-only operation, no file preparation needed).

## Dependencies

| Collaborator | Role | Allowed Interaction | Forbidden Interaction |
|---|---|---|---|
| StorageLayout | Path composition | `GetWorkflowPath(projectRoot, workflowName)` | Must not call other StorageLayout functions |
| `NewNode` | Constructor | Call with parsed fields for each node | Must not construct Node via struct literal |
| `NewTransition` | Constructor | Call with parsed fields for each transition | Must not construct Transition via struct literal |
| `NewExitTransition` | Constructor | Call with parsed fields for each exit transition | Must not construct ExitTransition via struct literal |
| `NewWorkflowDefinition` | Constructor | Call with derived Name, parsed Description, EntryNode, and constructed child slices | Must not construct WorkflowDefinition via struct literal |
| AgentLoader interface | Agent reference validation | `Load(agentRole) (*AgentDefinition, error)` | Must not construct AgentDefinitionLoader internally |
| OS filesystem | File reading | `os.ReadFile()` | Must not write, create, or delete |

Construction constraint: WorkflowDefinitionLoader is initialized with `projectRoot` and an agent loader interface. The interface is defined in the WorkflowDefinitionLoader's own package (per Go convention: interfaces belong in the consumer package).

Agent loader interface definition:

```
type AgentLoader interface {
    Load(agentRole string) (*AgentDefinition, error)
}
```

## Behavior

1. WorkflowDefinitionLoader is initialized with a `projectRoot` path and an `AgentLoader` interface.
2. When `Load(workflowName)` is called, composes the file path using `StorageLayout.GetWorkflowPath(projectRoot, workflowName)`.
3. Reads the file using `os.ReadFile`. If the file does not exist (`os.ErrNotExist`), returns error: `"workflow definition not found: <workflowName>"`. If another read error occurs, returns error: `"failed to read workflow definition '<workflowName>': <error>"`.
4. Parses the YAML content using a strict YAML decoder with unknown field rejection enabled (`yaml.Decoder` with `KnownFields(true)`). If parsing fails, returns error: `"failed to parse workflow definition '<workflowName>': <yaml error>"`.
5. Derives the `Name` value from the `workflowName` parameter (filename without `.yaml` extension). The YAML file does not contain a `name` field; strict parsing will reject any YAML that includes one.
6. Constructs each Node by calling `NewNode(name, nodeType, agentRole, description)` with parsed fields. If any constructor fails, returns error: `"workflow definition '<workflowName>' validation failed: node '<nodeName>': <constructor error>"`. If the node name cannot be determined from YAML (empty or missing), uses the zero-based index as fallback: `"workflow definition '<workflowName>' validation failed: node[<index>]: <constructor error>"`. Fails fast on first error.
7. Constructs each Transition by calling `NewTransition(fromNode, eventType, toNode)` with parsed fields. If any constructor fails, returns error: `"workflow definition '<workflowName>' validation failed: transition (from '<fromNode>', event '<eventType>', to '<toNode>'): <constructor error>"`. If fields are empty, uses whatever values are available. Fails fast on first error.
8. Constructs each ExitTransition by calling `NewExitTransition(fromNode, eventType, toNode)` with parsed fields. If any constructor fails, returns error: `"workflow definition '<workflowName>' validation failed: exit_transition (from '<fromNode>', event '<eventType>', to '<toNode>'): <constructor error>"`. Fails fast on first error.
9. Calls `NewWorkflowDefinition(name, description, entryNode, nodes, transitions, exitTransitions)` with the derived Name, parsed Description, parsed EntryNode string, and constructed child slices. If the constructor returns an error, returns error: `"workflow definition '<workflowName>' validation failed: <constructor error>"`.
10. After successful WorkflowDefinition construction, validates agent_role referential integrity. For each Node with Type `"agent"`, invokes `AgentLoader.Load(node.AgentRole())`. If the load fails, returns error: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': <underlying error>"`. Fails fast on first error.
11. If all validations pass, returns the constructed `*WorkflowDefinition`.
12. Each `Load()` call reads from disk and re-runs all validations including agent_role checks. No caching.
13. WorkflowDefinitionLoader is safe to call concurrently from multiple goroutines. No file locking is required.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| projectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| agentLoader | AgentLoader | Interface with `Load(agentRole string) (*AgentDefinition, error)` method | Yes |

### For Load Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| workflowName | string | Non-empty, corresponds to `<workflowName>.yaml` filename | Yes |

## Outputs

### For Load Method

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| WorkflowDefinition | *WorkflowDefinition | Fully constructed and validated workflow definition with verified agent references |

**Error Cases**:

| Error Message Format | Phase | Description |
|---------------------|-------|-------------|
| `"workflow definition not found: <workflowName>"` | Read | YAML file does not exist |
| `"failed to read workflow definition '<workflowName>': <error>"` | Read | File read failed (permission denied, etc.) |
| `"failed to parse workflow definition '<workflowName>': <yaml error>"` | Parse | YAML parsing failed (syntax error, type mismatch, unknown field) |
| `"workflow definition '<workflowName>' validation failed: node '<nodeName>': <constructor error>"` | Validation | Node constructor rejected input |
| `"workflow definition '<workflowName>' validation failed: node[<index>]: <constructor error>"` | Validation | Node constructor rejected input (name unavailable, fallback to index) |
| `"workflow definition '<workflowName>' validation failed: transition (from '<fromNode>', event '<eventType>', to '<toNode>'): <constructor error>"` | Validation | Transition constructor rejected input |
| `"workflow definition '<workflowName>' validation failed: exit_transition (from '<fromNode>', event '<eventType>', to '<toNode>'): <constructor error>"` | Validation | ExitTransition constructor rejected input |
| `"workflow definition '<workflowName>' validation failed: <constructor error>"` | Validation | WorkflowDefinition constructor rejected input (structural integrity) |
| `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': <underlying error>"` | Validation | Agent loader failed to load referenced agent_role |

## Invariants

1. **Read-Only**: Must never create, modify, or delete workflow definition files.
2. **Constructor Required**: Must construct all value objects (Node, Transition, ExitTransition, WorkflowDefinition) exclusively via their constructors. Direct struct literals are forbidden.
3. **No Caching**: Each `Load()` call reads from disk and re-runs all validations. No in-memory caching.
4. **Stateless**: No internal state between `Load()` calls.
5. **Thread-Safe**: Safe to call concurrently without synchronization.
6. **No File Locking**: Read-only operations do not require locks.
7. **Strict YAML Parsing**: Unknown fields in YAML must be rejected (KnownFields mode).
8. **Name From Filename Only**: Name is derived from the filename parameter. The workflow YAML file does not contain a `name` field; if present, strict parsing will reject it as an unknown field.
9. **Fail Fast**: Returns immediately on first error at each construction phase (nodes → transitions → exit transitions → workflow → agent refs).
10. **Path Composition Delegation**: Must use StorageLayout for workflow file path composition.
11. **Agent Reference Integrity at Load Time**: Every agent node's `agent_role` must be loadable via the injected AgentLoader. The runtime never receives a WorkflowDefinition with unresolvable agent references.
12. **Layered Error Context**: Errors must include phase context and identify the specific problematic element (node name/index, transition fields, etc.).
13. **Interface-Based Agent Loader**: AgentLoader is an interface defined in the WorkflowDefinitionLoader package, not a concrete type dependency.
14. **YAML camelCase Fields**: YAML fields use camelCase naming (e.g., `entryNode`, `agentRole`, `fromNode`, `eventType`, `toNode`, `exitTransitions`). Struct tags enforce this mapping.

## Edge Cases

- Condition: Workflow file `.spectra/workflows/<workflowName>.yaml` does not exist.
  Expected: Returns error `"workflow definition not found: <workflowName>"`.

- Condition: Workflow file exists but is empty (0 bytes).
  Expected: YAML parse fails. Returns `"failed to parse workflow definition '<workflowName>': EOF"`.

- Condition: YAML contains an unknown field (e.g., `customField: value`).
  Expected: Strict parser rejects it. Returns parse error.

- Condition: YAML contains a `name` field (legacy or mistakenly added).
  Expected: Strict parser rejects it as an unknown field. Returns parse error. Name is derived exclusively from filename.

- Condition: YAML uses snake_case field names (e.g., `entry_node` instead of `entryNode`).
  Expected: Strict parser treats as unknown field and rejects. Returns parse error.

- Condition: YAML is syntactically invalid.
  Expected: Returns `"failed to parse workflow definition '<workflowName>': <yaml syntax error>"`.

- Condition: A node in YAML has an empty `name` field.
  Expected: `NewNode` constructor returns `"node name cannot be empty"`. Loader wraps with index fallback: `"workflow definition '<workflowName>' validation failed: node[0]: node name cannot be empty"`.

- Condition: A node has a valid `name` but invalid `type` (e.g., `"bot"`).
  Expected: `NewNode` constructor returns `"node type must be 'agent' or 'human'"`. Loader wraps with node name: `"workflow definition '<workflowName>' validation failed: node '<nodeName>': node type must be 'agent' or 'human'"`.

- Condition: A transition has `fromNode == toNode`.
  Expected: `NewTransition` constructor returns `"from_node and to_node must be different"`. Loader wraps with transition context.

- Condition: Two transitions share the same (fromNode, eventType) pair.
  Expected: `NewWorkflowDefinition` constructor returns duplicate transition error. Loader wraps as: `"workflow definition '<workflowName>' validation failed: <constructor error>"`.

- Condition: EntryNode references a node with type `"agent"`.
  Expected: `NewWorkflowDefinition` constructor returns entry node type error. Loader wraps.

- Condition: A non-entry node is unreachable (no incoming transitions).
  Expected: `NewWorkflowDefinition` constructor returns reachability error. Loader wraps.

- Condition: An agent node's `agentRole` references a non-existent agent definition file.
  Expected: AgentLoader returns `"agent definition not found: <agentRole>"`. Loader wraps as: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': agent definition not found: <agentRole>"`.

- Condition: An agent node's `agentRole` references an agent that exists but fails its own validation (e.g., agent_root missing).
  Expected: AgentLoader returns its specific error. Loader wraps with node and role context.

- Condition: ExitTransitions array is empty in YAML.
  Expected: Parsed as empty slice. `NewWorkflowDefinition` constructor returns `"exit_transitions cannot be empty"`. Loader wraps.

- Condition: Nodes array is empty in YAML.
  Expected: Parsed as empty/nil slice. `NewWorkflowDefinition` constructor returns `"nodes cannot be empty"`. Loader wraps.

- Condition: Transitions array is empty in YAML.
  Expected: Parsed as empty/nil slice. `NewWorkflowDefinition` constructor returns `"transitions cannot be empty"`. Loader wraps.

- Condition: `workflowName` contains path separators (e.g., `"../malicious/workflow"`).
  Expected: StorageLayout composes path as-is. File read likely fails. Loader does not validate workflowName format; caller is responsible.

- Condition: `workflowName` is an empty string.
  Expected: StorageLayout produces malformed path. File read fails. Returns `"workflow definition not found: "`.

- Condition: File read fails with permission denied.
  Expected: Returns `"failed to read workflow definition '<workflowName>': permission denied"`.

- Condition: Multiple goroutines call `Load()` for the same workflow simultaneously.
  Expected: Both independently read, parse, and validate. Both succeed without interference.

- Condition: YAML file is modified between two `Load()` calls.
  Expected: Second call reflects updated content. No caching.

- Condition: `Description` field is missing from YAML.
  Expected: Parsed as empty string. `NewWorkflowDefinition` accepts empty description. Succeeds.

- Condition: An exit transition's `toNode` references a non-human node.
  Expected: `NewWorkflowDefinition` constructor returns exit transition target type error. Loader wraps.

- Condition: A node targeted by an exit transition has outgoing transitions defined.
  Expected: `NewWorkflowDefinition` accepts this. Exit-target nodes with outgoing transitions are valid.

- Condition: `ProjectRoot` does not contain a `.spectra/` directory.
  Expected: File read fails with not-found error.

## Related

- [WorkflowDefinition](../components/workflow_definition.md) — Value object constructed by this loader; owns cross-component structural validation
- [Node](../components/node.md) — Child value object constructed during loading
- [Transition](../components/transition.md) — Child value object constructed during loading
- [ExitTransition](../components/exit_transition.md) — Child value object constructed during loading
- [AgentDefinitionLoader](./agent_definition_loader.md) — Concrete implementation of the AgentLoader interface
- [StorageLayout](./storage_layout.md) — Provides path composition for workflow files
