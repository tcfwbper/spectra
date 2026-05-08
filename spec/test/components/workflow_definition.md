# Test Specification: `workflow_definition_test.go`

## Source File Under Test
`components/workflow_definition.go`

## Test File
`components/workflow_definition_test.go`

---

## `WorkflowDefinition`

### Happy Path — Construction

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_MinimalValid` | `unit` | Constructs a minimal valid WorkflowDefinition with one human entry node, one agent node, one transition, and one exit transition. | Create Nodes: `humanNode=NewNode("HumanStart","human","","")`, `agentNode=NewNode("Worker","agent","Architect","")`. Create Transition: `t=NewTransition("HumanStart","Submit","Worker")`. Create ExitTransition: `et=NewExitTransition("Worker","Done","HumanStart")`. Also add a Transition matching the exit: `t2=NewTransition("Worker","Done","HumanStart")`. | `name="SimpleFlow"`, `description="A simple flow"`, `entryNode="HumanStart"`, `nodes=[humanNode, agentNode]`, `transitions=[t, t2]`, `exitTransitions=[et]` | Returns no error; all getters return the provided values |
| `TestNewWorkflowDefinition_EmptyDescription` | `unit` | Accepts empty string Description. | Same node/transition setup as minimal valid case | `name="Flow"`, `description=""`, other fields valid | Returns no error; Description getter returns `""` |
| `TestNewWorkflowDefinition_MultipleTransitionsFromSameNode` | `unit` | Accepts multiple transitions from the same node with different EventTypes. | Create Nodes: `human=NewNode("Human","human","","")`, `agent1=NewNode("AgentA","agent","RoleA","")`, `agent2=NewNode("AgentB","agent","RoleB","")`. Create Transitions from Human with different events, plus exit transition back to Human. | `name="BranchFlow"`, valid nodes/transitions/exitTransitions | Returns no error; Transitions getter contains all transitions |
| `TestNewWorkflowDefinition_ExitTargetNodeNoOutgoing` | `unit` | Accepts a node targeted by an ExitTransition that has no outgoing transitions (exempt from coverage). | Create Nodes: `entry=NewNode("Entry","human","","")`, `worker=NewNode("Worker","agent","Architect","")`, `receiver=NewNode("Receiver","human","","")`. Transitions: `t1=NewTransition("Entry","Submit","Worker")`, `t2=NewTransition("Worker","Done","Receiver")`. ExitTransition: `et=NewExitTransition("Worker","Done","Receiver")`. `Receiver` is reachable via `t2`, is an exit target, and has zero outgoing transitions. | `name="ExitFlow"`, `entryNode="Entry"`, `nodes=[entry, worker, receiver]`, `transitions=[t1, t2]`, `exitTransitions=[et]` | Returns no error; `Receiver` has no outgoing transitions but is accepted because it is an exit target |

### Validation Failures — Name

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_EmptyName` | `unit` | Rejects empty string Name. | Valid nodes/transitions/exitTransitions setup | `name=""` with other fields valid | Returns error `"name cannot be empty"` |
| `TestNewWorkflowDefinition_NameStartsLowercase` | `unit` | Rejects Name starting with a lowercase letter. | Valid nodes/transitions/exitTransitions setup | `name="defaultWorkflow"` with other fields valid | Returns error `"name must be PascalCase (start with uppercase, alphanumeric only)"` |
| `TestNewWorkflowDefinition_NameContainsSpecialChar` | `unit` | Rejects Name with non-alphanumeric characters. | Valid nodes/transitions/exitTransitions setup | `name="Work-Flow"` with other fields valid | Returns error `"name must be PascalCase (start with uppercase, alphanumeric only)"` |

### Null / Empty Input

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_NilNodes` | `unit` | Rejects nil Nodes slice. | | `nodes=nil` with other fields valid | Returns error `"nodes cannot be empty"` |
| `TestNewWorkflowDefinition_EmptyNodes` | `unit` | Rejects empty Nodes slice. | | `nodes=[]` with other fields valid | Returns error `"nodes cannot be empty"` |
| `TestNewWorkflowDefinition_NilTransitions` | `unit` | Rejects nil Transitions slice. | Valid nodes setup | `transitions=nil` with other fields valid | Returns error `"transitions cannot be empty"` |
| `TestNewWorkflowDefinition_EmptyTransitions` | `unit` | Rejects empty Transitions slice. | Valid nodes setup | `transitions=[]` with other fields valid | Returns error `"transitions cannot be empty"` |
| `TestNewWorkflowDefinition_NilExitTransitions` | `unit` | Rejects nil ExitTransitions slice. | Valid nodes and transitions setup | `exitTransitions=nil` with other fields valid | Returns error `"exit_transitions cannot be empty"` |
| `TestNewWorkflowDefinition_EmptyExitTransitions` | `unit` | Rejects empty ExitTransitions slice. | Valid nodes and transitions setup | `exitTransitions=[]` with other fields valid | Returns error `"exit_transitions cannot be empty"` |

### Validation Failures — Node Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_DuplicateNodeName` | `unit` | Rejects Nodes with duplicate names. | Create two Nodes both named `"Architect"` (different types/roles are irrelevant). | `nodes=[node1, node2]` with duplicate name `"Architect"` | Returns error `"duplicate node name: 'Architect'"` |

### Validation Failures — EntryNode

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_EntryNodeNotFound` | `unit` | Rejects EntryNode referencing a name not in Nodes. | Valid nodes that do not include a node named `"Missing"` | `entryNode="Missing"` | Returns error `"entry_node 'Missing' does not reference a valid node"` |
| `TestNewWorkflowDefinition_EntryNodeNotHuman` | `unit` | Rejects EntryNode referencing a node with Type `"agent"`. | Create agent node named `"AgentEntry"` and include it in Nodes | `entryNode="AgentEntry"` | Returns error `"entry_node 'AgentEntry' must have type 'human'"` |

### Validation Failures — Transition Referential Integrity

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_TransitionFromNodeNotFound` | `unit` | Rejects Transition with FromNode not in Nodes. | Create Transition with `fromNode="NonExistent"` pointing to a valid node | Valid nodes without `"NonExistent"` | Returns error `"transition from_node 'NonExistent' does not reference a valid node"` |
| `TestNewWorkflowDefinition_TransitionToNodeNotFound` | `unit` | Rejects Transition with ToNode not in Nodes. | Create Transition with `toNode="NonExistent"` from a valid node | Valid nodes without `"NonExistent"` | Returns error `"transition to_node 'NonExistent' does not reference a valid node"` |

### Validation Failures — Transition Determinism

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_DuplicateFromNodeEventType` | `unit` | Rejects two Transitions with the same (FromNode, EventType) pair. | Create two Transitions: both from `"HumanApproval"` with EventType `"Approve"` but different ToNodes | Valid nodes including `"HumanApproval"` and target nodes | Returns error `"duplicate transition for event 'Approve' from node 'HumanApproval'"` |

### Validation Failures — ExitTransition Correspondence

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_ExitTransitionNoCorrespondingTransition` | `unit` | Rejects ExitTransition with no matching Transition triple. | Create an ExitTransition `(FromNode:"A", EventType:"Done", ToNode:"B")` with no matching Transition having the same triple | Valid nodes `A` and `B` exist | Returns error `"exit_transition (from_node: 'A', event_type: 'Done', to_node: 'B') has no corresponding transition"` |

### Validation Failures — ExitTransition Uniqueness

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_DuplicateExitTransition` | `unit` | Rejects two identical ExitTransitions. | Create two identical ExitTransitions `(FromNode:"Worker", EventType:"Done", ToNode:"Human")` with a corresponding Transition | Valid nodes and matching Transition exist | Returns error `"duplicate exit_transition (from_node: 'Worker', event_type: 'Done', to_node: 'Human')"` |

### Validation Failures — ExitTransition Target Type

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_ExitTransitionToNodeNotHuman` | `unit` | Rejects ExitTransition whose ToNode is not a human-type node. | Create ExitTransition pointing to an agent-type node. Matching Transition exists. | ExitTransition `toNode` references agent-type node `"AgentNode"` | Returns error `"exit_transition to_node 'AgentNode' must have type 'human'"` |

### Validation Failures — Outgoing Transition Coverage

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_NodeNoOutgoingNotExitTarget` | `unit` | Rejects a non-exit-target node with no outgoing transitions. | Create a node `"Orphan"` that is reachable (has incoming transition) but has no outgoing transitions and is not targeted by any ExitTransition | Node `"Orphan"` in Nodes with no outgoing Transition | Returns error `"node 'Orphan' has no outgoing transitions and is not an exit target"` |

### Validation Failures — Reachability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestNewWorkflowDefinition_UnreachableNode` | `unit` | Rejects a non-entry node with no incoming transitions. | Create a node `"Island"` that has outgoing transitions but no incoming transitions (not the entry node) | Node `"Island"` in Nodes with no Transition having `toNode="Island"` | Returns error `"node 'Island' is unreachable (no incoming transitions)"` |

### Immutability

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_Immutability` | `unit` | All fields remain unchanged after construction; no exported setters exist. | Construct a valid WorkflowDefinition with known values | Access all getters after construction | All getter values remain identical to construction inputs |

### Data Independence (Copy Semantics)

| Test ID | Category | Description | Setup | Input | Expected |
|---|---|---|---|---|---|
| `TestWorkflowDefinition_NodeSliceCopySemantics` | `unit` | Mutation of the original Nodes slice after construction does not affect the stored value. | Construct a valid WorkflowDefinition; keep reference to original nodes slice | Mutate the original nodes slice after construction (e.g., set element to nil) | Nodes getter still returns the original node values |
| `TestWorkflowDefinition_TransitionSliceCopySemantics` | `unit` | Mutation of the original Transitions slice after construction does not affect the stored value. | Construct a valid WorkflowDefinition; keep reference to original transitions slice | Mutate the original transitions slice after construction | Transitions getter still returns the original transition values |
| `TestWorkflowDefinition_ExitTransitionSliceCopySemantics` | `unit` | Mutation of the original ExitTransitions slice after construction does not affect the stored value. | Construct a valid WorkflowDefinition; keep reference to original exit transitions slice | Mutate the original exit transitions slice after construction | ExitTransitions getter still returns the original exit transition values |
