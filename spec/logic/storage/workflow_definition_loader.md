# WorkflowDefinitionLoader

## Overview

WorkflowDefinitionLoader provides read-only access to workflow definition files stored in `.spectra/workflows/`. It loads and parses workflow YAML files, validates all required fields and structural integrity (node references, transition integrity, exit transition constraints), and returns a complete WorkflowDefinition structure. WorkflowDefinitionLoader does not create, modify, or delete workflow files. It is stateless and does not cache definitions in memory.

## Behavior

1. WorkflowDefinitionLoader is initialized with a `ProjectRoot` path (absolute path to the directory containing `.spectra`).
2. WorkflowDefinitionLoader uses StorageLayout to compose the absolute path to workflow YAML files.
3. When the `Load(workflowName)` method is called, WorkflowDefinitionLoader composes the file path using `StorageLayout.GetWorkflowPath(ProjectRoot, workflowName)`.
4. WorkflowDefinitionLoader attempts to open and read the workflow file. If the file does not exist, it returns an error: `"workflow definition not found: <workflowName>"`.
5. WorkflowDefinitionLoader parses the YAML content into a WorkflowDefinition structure using a YAML parser (e.g., `gopkg.in/yaml.v3`).
6. If YAML parsing fails (syntax errors, type mismatches), WorkflowDefinitionLoader returns an error: `"failed to parse workflow definition '<workflowName>': <yaml error>"`.
7. After successful parsing, WorkflowDefinitionLoader validates all required fields are non-empty: `Name`, `EntryNode`, `ExitTransitions`, `Nodes`, `Transitions`.
8. If any required field is missing or empty, WorkflowDefinitionLoader returns an error: `"workflow definition '<workflowName>' validation failed: missing required field '<field_name>'"`.
9. WorkflowDefinitionLoader validates that `Name` follows PascalCase naming conventions. The name must match the pattern `^[A-Z][a-zA-Z0-9]*$`: starts with an uppercase letter, followed by any combination of letters (upper or lower case) and digits, with no spaces, underscores, hyphens, or other special characters. If the validation fails, it returns an error: `"workflow definition '<workflowName>' validation failed: name must be PascalCase with no spaces or special characters"`.
10. WorkflowDefinitionLoader validates that `EntryNode` references a valid `Node.Name` in the `Nodes` array. If not found, it returns an error: `"workflow definition '<workflowName>' validation failed: entry_node '<entryNode>' references non-existent node"`.
11. WorkflowDefinitionLoader validates that the node referenced by `EntryNode` has `type == "human"`. If not, it returns an error: `"workflow definition '<workflowName>' validation failed: entry_node '<entryNode>' must have type 'human', but has type '<actualType>'"`.
12. WorkflowDefinitionLoader validates that all `Node.Name` values are unique within the workflow. If duplicates are found, it returns an error: `"workflow definition '<workflowName>' validation failed: duplicate node name '<nodeName>'"`.
13. WorkflowDefinitionLoader validates that all `Transition.FromNode` and `Transition.ToNode` reference valid `Node.Name` values. If a reference is invalid, it returns an error: `"workflow definition '<workflowName>' validation failed: transition references non-existent node '<nodeName>'"`.
14. WorkflowDefinitionLoader validates that no transition has `FromNode == ToNode` (self-loop). If a self-loop is found, it returns an error: `"workflow definition '<workflowName>' validation failed: transition from_node and to_node must be different (node '<nodeName>', event '<eventType>')"`.
15. WorkflowDefinitionLoader validates that no two transitions share the same `(FromNode, EventType)` pair. If a duplicate is found, it returns an error: `"workflow definition '<workflowName>' validation failed: duplicate transition for event '<eventType>' from node '<nodeName>'"`.
16. WorkflowDefinitionLoader validates that no two ExitTransitions share the identical `(FromNode, EventType, ToNode)` triple. If a duplicate is found, it returns an error: `"workflow definition '<workflowName>' validation failed: duplicate exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>')"`.
17. WorkflowDefinitionLoader validates that all transitions in `ExitTransitions` match a transition defined in `Transitions` (exact match on `from_node`, `event_type`, and `to_node`). If an exit transition has no corresponding transition definition, it returns an error: `"workflow definition '<workflowName>' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition definition"`.
18. WorkflowDefinitionLoader validates that for each transition in `ExitTransitions`, the `to_node` references a node with `type == "human"`. If not, it returns an error: `"workflow definition '<workflowName>' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') must target a human node, but targets '<nodeName>' with type '<actualType>'"`.
19. WorkflowDefinitionLoader validates that every node not targeted by any exit transition has at least one outgoing transition. If a node has no outgoing transitions and is not an exit target, it returns an error: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' has no outgoing transitions and is not an exit target"`.
20. WorkflowDefinitionLoader validates that every node with `type == "agent"` references an `agent_role` that resolves to an existing agent definition. For each such node, WorkflowDefinitionLoader invokes the injected `AgentDefinitionLoader.Load(node.AgentRole)`. If the underlying load fails (file not found, parse error, or per-agent validation failure), it returns an error: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': <underlying error>"`. This guarantees that all agent references are resolvable at workflow load time, not at runtime.
21. If all validations pass, WorkflowDefinitionLoader returns the parsed WorkflowDefinition structure.
22. WorkflowDefinitionLoader does not cache workflow definitions. Each `Load()` call reads from disk and re-runs the agent-role checks via AgentDefinitionLoader.
23. WorkflowDefinitionLoader is safe to call concurrently from multiple goroutines. No file locking is required as all operations are read-only.
24. WorkflowDefinitionLoader does not use FileAccessor because it does not need to prepare files. It directly uses `os.ReadFile()` or equivalent for read operations.

## Inputs

### For Initialization

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| ProjectRoot | string | Absolute path to the directory containing `.spectra` | Yes |
| AgentDefinitionLoader | AgentDefinitionLoader | Injected loader used to verify agent_role referential integrity for every agent node in the workflow at load time | Yes |

### For Load Method

| Field | Type | Constraints | Required |
|-------|------|-------------|----------|
| WorkflowName | string | Non-empty, PascalCase identifier, corresponds to `<WorkflowName>.yaml` | Yes |

## Outputs

### For Load Method

**Success Case**:

| Field | Type | Description |
|-------|------|-------------|
| WorkflowDefinition | WorkflowDefinition struct | Fully parsed and validated workflow definition structure |

**Error Cases**:

| Error Message Format | Description |
|---------------------|-------------|
| `"workflow definition not found: <workflowName>"` | The workflow YAML file does not exist at the expected path |
| `"failed to read workflow definition '<workflowName>': <error>"` | File read operation failed (e.g., permission denied) |
| `"failed to parse workflow definition '<workflowName>': <yaml error>"` | YAML parsing failed due to syntax errors or type mismatches |
| `"workflow definition '<workflowName>' validation failed: missing required field '<field_name>'"` | A required field is missing or empty after parsing |
| `"workflow definition '<workflowName>' validation failed: name must be PascalCase with no spaces or special characters"` | Name field contains invalid characters or format |
| `"workflow definition '<workflowName>' validation failed: entry_node '<entryNode>' references non-existent node"` | EntryNode references a node that does not exist in Nodes array |
| `"workflow definition '<workflowName>' validation failed: entry_node '<entryNode>' must have type 'human', but has type '<actualType>'"` | EntryNode references a node with type other than "human" |
| `"workflow definition '<workflowName>' validation failed: duplicate node name '<nodeName>'"` | Multiple nodes have the same Name |
| `"workflow definition '<workflowName>' validation failed: transition references non-existent node '<nodeName>'"` | A transition references a FromNode or ToNode that does not exist |
| `"workflow definition '<workflowName>' validation failed: transition from_node and to_node must be different (node '<nodeName>', event '<eventType>')"` | A transition has FromNode == ToNode (self-loop) |
| `"workflow definition '<workflowName>' validation failed: duplicate transition for event '<eventType>' from node '<nodeName>'"` | Two transitions share the same (FromNode, EventType) |
| `"workflow definition '<workflowName>' validation failed: duplicate exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>')"` | Two ExitTransitions share the identical triple |
| `"workflow definition '<workflowName>' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition definition"` | An exit transition does not match any transition in Transitions |
| `"workflow definition '<workflowName>' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') must target a human node, but targets '<nodeName>' with type '<actualType>'"` | An exit transition's to_node targets a non-human node |
| `"workflow definition '<workflowName>' validation failed: node '<nodeName>' has no outgoing transitions and is not an exit target"` | A non-exit-target node has no outgoing transitions |
| `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': <underlying error>"` | An agent node references an agent_role that AgentDefinitionLoader cannot load (missing file, parse error, or per-agent validation failure) |

## Invariants

1. **Read-Only**: WorkflowDefinitionLoader must never create, modify, or delete workflow definition files.

2. **No FileAccessor**: WorkflowDefinitionLoader must not use FileAccessor because it does not prepare files. It directly reads existing files using standard file I/O.

3. **No Caching**: WorkflowDefinitionLoader must not cache workflow definitions in memory. Each `Load()` call must read from disk.

4. **Stateless**: WorkflowDefinitionLoader must not maintain any internal state between `Load()` calls. It is safe to call `Load()` multiple times with different workflow names.

5. **Thread-Safe**: WorkflowDefinitionLoader must be safe to call concurrently from multiple goroutines without synchronization.

6. **No File Locking**: WorkflowDefinitionLoader must not acquire file locks. Read-only operations on immutable files do not require locking.

7. **Complete Validation**: WorkflowDefinitionLoader must validate all structural integrity constraints defined in `logic/components/workflow_definition.md`, `logic/components/transition.md`, `logic/components/exit_transition.md`, and `logic/components/node.md` before returning a WorkflowDefinition. This includes: required fields, PascalCase naming, EntryNode validity & type, node uniqueness, transition node-reference integrity, no self-loop transitions, transition uniqueness on `(from_node, event_type)`, ExitTransition uniqueness, ExitTransition correspondence to a defined transition, ExitTransition `to_node` type==human, non-exit-target outgoing transition requirement, and agent_role referential integrity (via AgentDefinitionLoader).

8. **Detailed Error Messages**: All validation errors must include the workflow name and the specific issue (field name, node name, event type, etc.).

9. **Fail Fast Validation**: If any validation fails, WorkflowDefinitionLoader must return an error immediately without attempting to proceed. Only the first validation error is reported. Users must fix errors iteratively. This design simplifies error handling and prevents cascading errors from obscuring the root cause.

10. **Path Composition Delegation**: WorkflowDefinitionLoader must delegate all path composition to StorageLayout. It must not construct paths manually.

11. **YAML Parsing Only**: WorkflowDefinitionLoader must only support YAML format. It must not attempt to parse other formats (JSON, TOML, etc.).

12. **PascalCase Validation Pattern**: Name validation must follow the pattern `^[A-Z][a-zA-Z0-9]*$`. This pattern allows consecutive uppercase letters (e.g., `SimpleSdd`), mixed case (e.g., `SimpleWorkflow`), and digits (e.g., `V2Workflow`), but prohibits spaces, underscores, hyphens, and other special characters.

13. **Agent Reference Integrity at Load Time**: Every node with `type=="agent"` must have an `agent_role` whose corresponding YAML file exists and passes AgentDefinitionLoader's full validation. WorkflowDefinitionLoader resolves these references at load time by invoking `AgentDefinitionLoader.Load(agentRole)` for each agent node. The Runtime never reaches a state where it has an in-memory WorkflowDefinition referencing a non-existent or invalid agent.

## Edge Cases

- **Condition**: Workflow file `.spectra/workflows/<workflowName>.yaml` does not exist.
  **Expected**: WorkflowDefinitionLoader returns an error: `"workflow definition not found: <workflowName>"`.

- **Condition**: Workflow file exists but is empty.
  **Expected**: YAML parser fails, WorkflowDefinitionLoader returns: `"failed to parse workflow definition '<workflowName>': EOF"`.

- **Condition**: Workflow file contains invalid YAML syntax (e.g., incorrect indentation).
  **Expected**: YAML parser fails with line/column info, WorkflowDefinitionLoader returns: `"failed to parse workflow definition '<workflowName>': yaml: line 5: mapping values are not allowed in this context"`.

- **Condition**: Workflow file has valid YAML but missing required field `Name`.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: missing required field 'name'"`.

- **Condition**: `Name` field contains spaces (e.g., `"Simple SDD"`).
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: name must be PascalCase with no spaces or special characters"`.

- **Condition**: `Name` field contains underscores (e.g., `"Simple_SDD"`).
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: name must be PascalCase with no spaces or special characters"`.

- **Condition**: `Name` field contains hyphens or other special characters (e.g., `"Simple-SDD"`, `"Simple.SDD"`).
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: name must be PascalCase with no spaces or special characters"`.

- **Condition**: `Name` field is valid PascalCase with consecutive uppercase letters (e.g., `"SimpleSDD"`).
  **Expected**: WorkflowDefinitionLoader accepts this as valid.

- **Condition**: `Name` field contains digits (e.g., `"V2Workflow"`, `"Workflow2024"`).
  **Expected**: WorkflowDefinitionLoader accepts this as valid (digits are allowed after the first character).

- **Condition**: `Name` field starts with a lowercase letter (e.g., `"simpleSDD"`).
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: name must be PascalCase with no spaces or special characters"`.

- **Condition**: `Name` field is a single uppercase letter (e.g., `"A"`).
  **Expected**: WorkflowDefinitionLoader accepts this as valid (matches pattern `^[A-Z][a-zA-Z0-9]*$`).

- **Condition**: `EntryNode` field references a node name that does not exist in `Nodes`.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: entry_node '<entryNode>' references non-existent node"`.

- **Condition**: `EntryNode` references a node with `type == "agent"`.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: entry_node '<nodeName>' must have type 'human', but has type 'agent'"`.

- **Condition**: Two nodes in `Nodes` array have the same `Name`.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: duplicate node name '<nodeName>'"`.

- **Condition**: A transition's `FromNode` references a non-existent node.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: transition references non-existent node '<nodeName>'"`.

- **Condition**: A transition in `ExitTransitions` does not match any transition defined in `Transitions`.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') has no corresponding transition definition"`.

- **Condition**: A transition has `from_node == to_node` (self-loop).
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: transition from_node and to_node must be different (node '<nodeName>', event '<eventType>')"`.

- **Condition**: Two transitions share the same `(from_node, event_type)` pair (with different `to_node`).
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: duplicate transition for event '<eventType>' from node '<nodeName>'"`.

- **Condition**: Two ExitTransitions have the identical `(from_node, event_type, to_node)` triple.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: duplicate exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>')"`.

- **Condition**: An agent node's `agent_role` references an AgentDefinition file that does not exist.
  **Expected**: AgentDefinitionLoader returns `"agent definition not found: <agentRole>"`. WorkflowDefinitionLoader wraps this and returns: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': agent definition not found: <agentRole>"`.

- **Condition**: An agent node's `agent_role` references an AgentDefinition that exists but fails per-agent validation (e.g., agent_root directory missing).
  **Expected**: AgentDefinitionLoader returns its specific validation error. WorkflowDefinitionLoader wraps it and returns: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' references invalid agent_role '<agentRole>': <agent validation error>"`.

- **Condition**: An exit transition's `to_node` targets a node with `type == "agent"`.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: exit transition (from_node: '<from>', event_type: '<type>', to_node: '<to>') must target a human node, but targets '<nodeName>' with type 'agent'"`.

- **Condition**: A node that is not an exit target has no outgoing transitions.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: node '<nodeName>' has no outgoing transitions and is not an exit target"`.

- **Condition**: `ExitTransitions` array is empty.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: missing required field 'exit_transitions'"` (treating empty array as missing).

- **Condition**: `Nodes` array is empty.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: missing required field 'nodes'"` (treating empty array as missing).

- **Condition**: `Transitions` array is empty.
  **Expected**: WorkflowDefinitionLoader returns: `"workflow definition '<workflowName>' validation failed: missing required field 'transitions'"` (treating empty array as missing).

- **Condition**: `WorkflowName` contains path separators (e.g., `"../malicious/workflow"`).
  **Expected**: StorageLayout composes the path as-is, potentially pointing outside `.spectra/workflows/`. File read fails with "not found" or accesses unintended file. WorkflowDefinitionLoader does not validate workflow name format; caller is responsible.

- **Condition**: `WorkflowName` is an empty string.
  **Expected**: StorageLayout produces a malformed path. File read fails with an error: `"workflow definition not found: "`.

- **Condition**: File read operation fails with "permission denied".
  **Expected**: WorkflowDefinitionLoader returns: `"failed to read workflow definition '<workflowName>': permission denied"`.

- **Condition**: Multiple goroutines call `Load()` for the same workflow simultaneously.
  **Expected**: Both goroutines independently read and parse the file. Both succeed without interference. No file locking is used.

- **Condition**: Multiple goroutines call `Load()` for different workflows simultaneously.
  **Expected**: All goroutines succeed independently. WorkflowDefinitionLoader is stateless and thread-safe.

- **Condition**: User modifies the workflow file while a `Load()` operation is in progress.
  **Expected**: The `Load()` operation reads whatever file state the OS provides (may be partially old, partially new depending on OS buffering). No guarantees about consistency. This is acceptable as workflows should not be modified during runtime.

- **Condition**: User modifies the workflow file between two `Load()` calls.
  **Expected**: The second `Load()` call returns the updated workflow definition. WorkflowDefinitionLoader does not cache, so changes are immediately reflected.

- **Condition**: YAML file contains fields not defined in WorkflowDefinition structure (e.g., user-defined metadata).
  **Expected**: YAML parser ignores unknown fields (assuming standard unmarshaling behavior). WorkflowDefinitionLoader returns the defined fields only.

- **Condition**: `Description` field is missing or empty.
  **Expected**: WorkflowDefinitionLoader allows this. `Description` is optional (defaults to empty string).

- **Condition**: A node is unreachable (no incoming transitions).
  **Expected**: WorkflowDefinitionLoader allows this. Unreachable nodes are not a validation failure (they may be intentionally unused or for future expansion).

- **Condition**: A node targeted by an exit transition has outgoing transitions defined.
  **Expected**: WorkflowDefinitionLoader allows this. Exit target nodes with outgoing transitions are not a validation failure (the Runtime will ignore those transitions).

- **Condition**: `ProjectRoot` is a relative path (e.g., `"./project"`).
  **Expected**: WorkflowDefinitionLoader uses it as-is. Path composition may produce incorrect paths. Caller is responsible for providing an absolute `ProjectRoot`.

- **Condition**: `ProjectRoot` does not contain a `.spectra/` directory.
  **Expected**: File read fails with "not found" error. WorkflowDefinitionLoader does not validate the existence of `.spectra/`.

## Related

- [WorkflowDefinition](../components/workflow_definition.md) - Defines the WorkflowDefinition structure and validation rules
- [StorageLayout](./storage_layout.md) - Provides path composition for workflow files
- [AgentDefinitionLoader](./agent_definition_loader.md) - Similar loader for agent definitions
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Framework architecture overview
