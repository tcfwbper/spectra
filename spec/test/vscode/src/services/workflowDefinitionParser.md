# Test Specification: `workflowDefinitionParser.test.ts`

## Source File Under Test
`vscode/src/services/workflowDefinitionParser.ts`

## Test File
`vscode/test/suite/workflowDefinitionParser.test.ts`

---

## `WorkflowDefinitionParser`

### Happy Path — parseWorkflowDefinition

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return entryNode and matching eventTypes from valid workflow` | `unit` | Parses a valid YAML with transitions matching entryNode and returns deduplicated eventTypes. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 'start', eventType: 'init'}, {fromNode: 'start', eventType: 'retry'}, {fromNode: 'other', eventType: 'done'}]`. Create a mock logger. | `projectRoot='/project'`, `workflowName='deploy'`, `logger` | Returns `{ entryNode: 'start', eventTypes: ['init', 'retry'] }` |
| `should deduplicate eventTypes when multiple transitions share the same eventType` | `unit` | Returns each eventType only once even when duplicated across transitions. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'begin'` and `transitions: [{fromNode: 'begin', eventType: 'trigger'}, {fromNode: 'begin', eventType: 'trigger'}, {fromNode: 'begin', eventType: 'other'}]`. Create a mock logger. | `projectRoot='/project'`, `workflowName='build'`, `logger` | Returns `{ entryNode: 'begin', eventTypes: ['trigger', 'other'] }` |
| `should return empty eventTypes when no transitions match entryNode` | `unit` | Returns empty array without warning when no transition has fromNode equal to entryNode. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 'middle', eventType: 'proceed'}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='test'`, `logger` | Returns `{ entryNode: 'start', eventTypes: [] }`; `logger.warn` not called |
| `should return empty eventTypes when transitions array is empty` | `unit` | Returns empty array without warning when transitions is a valid but empty array. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'start'` and `transitions: []`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='empty'`, `logger` | Returns `{ entryNode: 'start', eventTypes: [] }`; `logger.warn` not called |
| `should ignore unknown top-level keys in YAML` | `unit` | Extra top-level keys beyond entryNode and transitions do not affect parsing. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'start'`, `transitions: [{fromNode: 'start', eventType: 'go'}]`, `description: 'extra'`, `nodes: [...]`. Create a mock logger. | `projectRoot='/project'`, `workflowName='extra'`, `logger` | Returns `{ entryNode: 'start', eventTypes: ['go'] }` |
| `should ignore unknown keys in transition dicts` | `unit` | Extra keys in transition objects beyond fromNode and eventType are silently ignored. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 'start', eventType: 'run', toNode: 'end', guard: 'check'}]`. Create a mock logger. | `projectRoot='/project'`, `workflowName='rich'`, `logger` | Returns `{ entryNode: 'start', eventTypes: ['run'] }` |

### Error Propagation

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should return failure result when file does not exist` | `unit` | Returns failure result and logs warning when the YAML file is missing. | Stub `fs.readFile` to reject with an ENOENT error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='missing'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once with a descriptive message |
| `should return failure result when YAML syntax is invalid` | `unit` | Returns failure result and logs warning on malformed YAML. | Stub `fs.readFile` to resolve with invalid YAML content (e.g., `': invalid: [unterminated'`). Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='broken'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should return failure result when entryNode key is missing` | `unit` | Returns failure result and logs warning when YAML lacks entryNode. | Stub `fs.readFile` to resolve with YAML containing only `transitions: [{fromNode: 'a', eventType: 'b'}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='noentry'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should return failure result when transitions key is missing` | `unit` | Returns failure result and logs warning when YAML lacks transitions. | Stub `fs.readFile` to resolve with YAML containing only `entryNode: 'start'`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='notrans'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should return failure result when entryNode is not a string` | `unit` | Returns failure result and logs warning when entryNode has non-string value. | Stub `fs.readFile` to resolve with YAML containing `entryNode: 42` and `transitions: []`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='badentry'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should return failure result when transitions is not an array` | `unit` | Returns failure result and logs warning when transitions is a non-array value. | Stub `fs.readFile` to resolve with YAML containing `entryNode: 'start'` and `transitions: 'not-an-array'`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='badtrans'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should never throw an exception to the caller` | `unit` | Any internal error is caught and returns failure result instead of throwing. | Stub `fs.readFile` to reject with a generic Error. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='crash'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; does not throw |

### Validation Failures

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should fail fast when a transition dict is missing fromNode` | `unit` | Returns failure result immediately when any transition lacks fromNode. | Stub `fs.readFile` to resolve with YAML containing `entryNode: 'start'` and `transitions: [{eventType: 'go'}, {fromNode: 'start', eventType: 'valid'}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='badfrom'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should fail fast when a transition dict is missing eventType` | `unit` | Returns failure result immediately when any transition lacks eventType. | Stub `fs.readFile` to resolve with YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 'start'}, {fromNode: 'start', eventType: 'valid'}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='badevent'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should fail fast when fromNode in a transition is not a string` | `unit` | Returns failure result when fromNode has non-string value. | Stub `fs.readFile` to resolve with YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 123, eventType: 'go'}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='numfrom'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |
| `should fail fast when eventType in a transition is not a string` | `unit` | Returns failure result when eventType has non-string value. | Stub `fs.readFile` to resolve with YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 'start', eventType: ['array']}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='arrtype'`, `logger` | Returns `{ entryNode: '', eventTypes: [] }`; `logger.warn` called once |

### Mock / Dependency Interaction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `should construct correct file path from projectRoot and workflowName` | `unit` | Calls readFile with the correct path combining projectRoot, `.spectra/workflows`, and workflowName with .yaml extension. | Stub `fs.readFile` with a spy that resolves with valid YAML (`entryNode: 'x'`, `transitions: []`). Create a mock logger. | `projectRoot='/my/root'`, `workflowName='deploy'`, `logger` | `fs.readFile` called with path `/my/root/.spectra/workflows/deploy.yaml` |
| `should not call any write operations on fs` | `unit` | The method performs no mutations to the filesystem. | Stub `fs.readFile` to resolve with valid YAML. Spy on `fs.writeFile`, `fs.mkdir`, `fs.unlink`. Create a mock logger. | `projectRoot='/project'`, `workflowName='test'`, `logger` | None of `fs.writeFile`, `fs.mkdir`, `fs.unlink` are called |
| `should call logger.warn exactly once per failure` | `unit` | Each failure scenario calls logger.warn exactly once before returning. | Stub `fs.readFile` to reject with ENOENT. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='gone'`, `logger` | `logger.warn` called exactly once |
| `should not call logger.warn on successful parse` | `unit` | No warning is emitted when parsing succeeds. | Stub `fs.readFile` to resolve with valid YAML containing `entryNode: 'start'` and `transitions: [{fromNode: 'start', eventType: 'go'}]`. Create a mock logger with a `warn` spy. | `projectRoot='/project'`, `workflowName='ok'`, `logger` | `logger.warn` not called |
