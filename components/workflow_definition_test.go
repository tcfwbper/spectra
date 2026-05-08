package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper: build minimal valid workflow components ---

// buildMinimalWorkflowComponents creates the minimal set of nodes, transitions,
// and exit transitions for a valid WorkflowDefinition (one human entry, one agent,
// one transition, one exit transition).
func buildMinimalWorkflowComponents(t *testing.T) (
	humanNode *Node,
	agentNode *Node,
	trans []*Transition,
	exitTrans []*ExitTransition,
) {
	t.Helper()

	var err error
	humanNode, err = NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err = NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	tr1, err := NewTransition("HumanStart", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	trans = []*Transition{tr1, tr2}
	exitTrans = []*ExitTransition{et}
	return
}

// --- Happy Path — Construction ---

func TestNewWorkflowDefinition_MinimalValid(t *testing.T) {
	humanNode, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	nodes := []*Node{humanNode, agentNode}

	wd, err := NewWorkflowDefinition(
		"SimpleFlow",
		"A simple flow",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)

	require.NoError(t, err)
	assert.Equal(t, "SimpleFlow", wd.Name())
	assert.Equal(t, "A simple flow", wd.Description())
	assert.Equal(t, "HumanStart", wd.EntryNode())
	assert.Len(t, wd.Nodes(), 2)
	assert.Len(t, wd.Transitions(), 2)
	assert.Len(t, wd.ExitTransitions(), 1)
}

func TestNewWorkflowDefinition_EmptyDescription(t *testing.T) {
	humanNode, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	nodes := []*Node{humanNode, agentNode}

	wd, err := NewWorkflowDefinition(
		"Flow",
		"",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)

	require.NoError(t, err)
	assert.Equal(t, "", wd.Description())
}

func TestNewWorkflowDefinition_MultipleTransitionsFromSameNode(t *testing.T) {
	human, err := NewNode("Human", "human", "", "")
	require.NoError(t, err)

	agent1, err := NewNode("AgentA", "agent", "RoleA", "")
	require.NoError(t, err)

	agent2, err := NewNode("AgentB", "agent", "RoleB", "")
	require.NoError(t, err)

	tr1, err := NewTransition("Human", "SubmitA", "AgentA")
	require.NoError(t, err)

	tr2, err := NewTransition("Human", "SubmitB", "AgentB")
	require.NoError(t, err)

	tr3, err := NewTransition("AgentA", "Done", "Human")
	require.NoError(t, err)

	tr4, err := NewTransition("AgentB", "Done", "Human")
	require.NoError(t, err)

	et1, err := NewExitTransition("AgentA", "Done", "Human")
	require.NoError(t, err)

	et2, err := NewExitTransition("AgentB", "Done", "Human")
	require.NoError(t, err)

	nodes := []*Node{human, agent1, agent2}
	trans := []*Transition{tr1, tr2, tr3, tr4}
	exitTrans := []*ExitTransition{et1, et2}

	wd, err := NewWorkflowDefinition(
		"BranchFlow",
		"",
		"Human",
		nodes,
		trans,
		exitTrans,
	)

	require.NoError(t, err)
	assert.Len(t, wd.Transitions(), 4)
}

func TestNewWorkflowDefinition_ExitTargetNodeNoOutgoing(t *testing.T) {
	entry, err := NewNode("Entry", "human", "", "")
	require.NoError(t, err)

	worker, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	receiver, err := NewNode("Receiver", "human", "", "")
	require.NoError(t, err)

	tr1, err := NewTransition("Entry", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "Receiver")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "Receiver")
	require.NoError(t, err)

	nodes := []*Node{entry, worker, receiver}
	trans := []*Transition{tr1, tr2}
	exitTrans := []*ExitTransition{et}

	wd, err := NewWorkflowDefinition(
		"ExitFlow",
		"",
		"Entry",
		nodes,
		trans,
		exitTrans,
	)

	require.NoError(t, err)
	// Receiver has no outgoing transitions but is accepted because it is an exit target.
	assert.Equal(t, "ExitFlow", wd.Name())
}

// --- Validation Failures — Name ---

func TestNewWorkflowDefinition_EmptyName(t *testing.T) {
	_, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)
	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"",
		"desc",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)

	require.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())
}

func TestNewWorkflowDefinition_NameStartsLowercase(t *testing.T) {
	_, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)
	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"defaultWorkflow",
		"desc",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)

	require.Error(t, err)
	assert.Equal(t, "name must be PascalCase (start with uppercase, alphanumeric only)", err.Error())
}

func TestNewWorkflowDefinition_NameContainsSpecialChar(t *testing.T) {
	_, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)
	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Work-Flow",
		"desc",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)

	require.Error(t, err)
	assert.Equal(t, "name must be PascalCase (start with uppercase, alphanumeric only)", err.Error())
}

// --- Null / Empty Input ---

func TestNewWorkflowDefinition_NilNodes(t *testing.T) {
	_, err := NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nil,
		[]*Transition{},
		[]*ExitTransition{},
	)

	require.Error(t, err)
	assert.Equal(t, "nodes cannot be empty", err.Error())
}

func TestNewWorkflowDefinition_EmptyNodes(t *testing.T) {
	_, err := NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		[]*Node{},
		[]*Transition{},
		[]*ExitTransition{},
	)

	require.Error(t, err)
	assert.Equal(t, "nodes cannot be empty", err.Error())
}

func TestNewWorkflowDefinition_NilTransitions(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		nil,
		[]*ExitTransition{},
	)

	require.Error(t, err)
	assert.Equal(t, "transitions cannot be empty", err.Error())
}

func TestNewWorkflowDefinition_EmptyTransitions(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{},
		[]*ExitTransition{},
	)

	require.Error(t, err)
	assert.Equal(t, "transitions cannot be empty", err.Error())
}

func TestNewWorkflowDefinition_NilExitTransitions(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	tr, err := NewTransition("HumanStart", "Submit", "Worker")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr},
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, "exit_transitions cannot be empty", err.Error())
}

func TestNewWorkflowDefinition_EmptyExitTransitions(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	tr, err := NewTransition("HumanStart", "Submit", "Worker")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr},
		[]*ExitTransition{},
	)

	require.Error(t, err)
	assert.Equal(t, "exit_transitions cannot be empty", err.Error())
}

// --- Validation Failures — Node Uniqueness ---

func TestNewWorkflowDefinition_DuplicateNodeName(t *testing.T) {
	node1, err := NewNode("Architect", "agent", "RoleA", "")
	require.NoError(t, err)

	node2, err := NewNode("Architect", "agent", "RoleB", "")
	require.NoError(t, err)

	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	tr, err := NewTransition("HumanStart", "Submit", "Architect")
	require.NoError(t, err)

	et, err := NewExitTransition("Architect", "Done", "HumanStart")
	require.NoError(t, err)

	nodes := []*Node{humanNode, node1, node2}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "duplicate node name: 'Architect'", err.Error())
}

// --- Validation Failures — EntryNode ---

func TestNewWorkflowDefinition_EntryNodeNotFound(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	tr, err := NewTransition("HumanStart", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"Missing",
		nodes,
		[]*Transition{tr, tr2},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "entry_node 'Missing' does not reference a valid node", err.Error())
}

func TestNewWorkflowDefinition_EntryNodeNotHuman(t *testing.T) {
	agentEntry, err := NewNode("AgentEntry", "agent", "Architect", "")
	require.NoError(t, err)

	humanNode, err := NewNode("HumanEnd", "human", "", "")
	require.NoError(t, err)

	tr, err := NewTransition("AgentEntry", "Done", "HumanEnd")
	require.NoError(t, err)

	et, err := NewExitTransition("AgentEntry", "Done", "HumanEnd")
	require.NoError(t, err)

	nodes := []*Node{agentEntry, humanNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"AgentEntry",
		nodes,
		[]*Transition{tr},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "entry_node 'AgentEntry' must have type 'human'", err.Error())
}

// --- Validation Failures — Transition Referential Integrity ---

func TestNewWorkflowDefinition_TransitionFromNodeNotFound(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	// Transition from a non-existent node.
	tr, err := NewTransition("NonExistent", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr, tr2},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "transition from_node 'NonExistent' does not reference a valid node", err.Error())
}

func TestNewWorkflowDefinition_TransitionToNodeNotFound(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	// Transition to a non-existent node.
	tr, err := NewTransition("HumanStart", "Submit", "NonExistent")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr, tr2},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "transition to_node 'NonExistent' does not reference a valid node", err.Error())
}

// --- Validation Failures — Transition Determinism ---

func TestNewWorkflowDefinition_DuplicateFromNodeEventType(t *testing.T) {
	humanNode, err := NewNode("HumanApproval", "human", "", "")
	require.NoError(t, err)

	agent1, err := NewNode("AgentA", "agent", "RoleA", "")
	require.NoError(t, err)

	agent2, err := NewNode("AgentB", "agent", "RoleB", "")
	require.NoError(t, err)

	// Two transitions from same node with same event type.
	tr1, err := NewTransition("HumanApproval", "Approve", "AgentA")
	require.NoError(t, err)

	tr2, err := NewTransition("HumanApproval", "Approve", "AgentB")
	require.NoError(t, err)

	tr3, err := NewTransition("AgentA", "Done", "HumanApproval")
	require.NoError(t, err)

	tr4, err := NewTransition("AgentB", "Done", "HumanApproval")
	require.NoError(t, err)

	et, err := NewExitTransition("AgentA", "Done", "HumanApproval")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agent1, agent2}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanApproval",
		nodes,
		[]*Transition{tr1, tr2, tr3, tr4},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "duplicate transition for event 'Approve' from node 'HumanApproval'", err.Error())
}

// --- Validation Failures — ExitTransition Correspondence ---

func TestNewWorkflowDefinition_ExitTransitionNoCorrespondingTransition(t *testing.T) {
	humanNode, err := NewNode("B", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("A", "agent", "Architect", "")
	require.NoError(t, err)

	// Transition that does NOT match the exit transition.
	tr, err := NewTransition("B", "Submit", "A")
	require.NoError(t, err)

	// ExitTransition with no corresponding Transition triple.
	et, err := NewExitTransition("A", "Done", "B")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"B",
		nodes,
		[]*Transition{tr},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "exit_transition (from_node: 'A', event_type: 'Done', to_node: 'B') has no corresponding transition", err.Error())
}

// --- Validation Failures — ExitTransition Uniqueness ---

func TestNewWorkflowDefinition_DuplicateExitTransition(t *testing.T) {
	humanNode, err := NewNode("Human", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	tr1, err := NewTransition("Human", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "Human")
	require.NoError(t, err)

	et1, err := NewExitTransition("Worker", "Done", "Human")
	require.NoError(t, err)

	et2, err := NewExitTransition("Worker", "Done", "Human")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"Human",
		nodes,
		[]*Transition{tr1, tr2},
		[]*ExitTransition{et1, et2},
	)

	require.Error(t, err)
	assert.Equal(t, "duplicate exit_transition (from_node: 'Worker', event_type: 'Done', to_node: 'Human')", err.Error())
}

// --- Validation Failures — ExitTransition Target Type ---

func TestNewWorkflowDefinition_ExitTransitionToNodeNotHuman(t *testing.T) {
	humanNode, err := NewNode("Human", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("AgentNode", "agent", "RoleA", "")
	require.NoError(t, err)

	agentNode2, err := NewNode("AgentNode2", "agent", "RoleB", "")
	require.NoError(t, err)

	tr1, err := NewTransition("Human", "Submit", "AgentNode")
	require.NoError(t, err)

	tr2, err := NewTransition("AgentNode", "Forward", "AgentNode2")
	require.NoError(t, err)

	tr3, err := NewTransition("AgentNode2", "Done", "Human")
	require.NoError(t, err)

	// ExitTransition pointing to an agent-type node.
	et, err := NewExitTransition("AgentNode", "Forward", "AgentNode2")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode, agentNode2}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"Human",
		nodes,
		[]*Transition{tr1, tr2, tr3},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "exit_transition to_node 'AgentNode2' must have type 'human'", err.Error())
}

// --- Validation Failures — Outgoing Transition Coverage ---

func TestNewWorkflowDefinition_NodeNoOutgoingNotExitTarget(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	orphanNode, err := NewNode("Orphan", "agent", "Writer", "")
	require.NoError(t, err)

	// Transitions: HumanStart -> Worker, Worker -> Orphan, Worker -> HumanStart (exit)
	tr1, err := NewTransition("HumanStart", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Forward", "Orphan")
	require.NoError(t, err)

	tr3, err := NewTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	// Orphan has an incoming transition but no outgoing transitions and is not an exit target.
	nodes := []*Node{humanNode, agentNode, orphanNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr1, tr2, tr3},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "node 'Orphan' has no outgoing transitions and is not an exit target", err.Error())
}

// --- Validation Failures — Reachability ---

func TestNewWorkflowDefinition_UnreachableNode(t *testing.T) {
	humanNode, err := NewNode("HumanStart", "human", "", "")
	require.NoError(t, err)

	agentNode, err := NewNode("Worker", "agent", "Architect", "")
	require.NoError(t, err)

	// Island has outgoing transitions but no incoming (not entry node).
	islandNode, err := NewNode("Island", "agent", "Writer", "")
	require.NoError(t, err)

	tr1, err := NewTransition("HumanStart", "Submit", "Worker")
	require.NoError(t, err)

	tr2, err := NewTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	// Island has outgoing but no incoming.
	tr3, err := NewTransition("Island", "Send", "HumanStart")
	require.NoError(t, err)

	et, err := NewExitTransition("Worker", "Done", "HumanStart")
	require.NoError(t, err)

	nodes := []*Node{humanNode, agentNode, islandNode}

	_, err = NewWorkflowDefinition(
		"Flow",
		"desc",
		"HumanStart",
		nodes,
		[]*Transition{tr1, tr2, tr3},
		[]*ExitTransition{et},
	)

	require.Error(t, err)
	assert.Equal(t, "node 'Island' is unreachable (no incoming transitions)", err.Error())
}

// --- Immutability ---

func TestWorkflowDefinition_Immutability(t *testing.T) {
	humanNode, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	nodes := []*Node{humanNode, agentNode}

	wd, err := NewWorkflowDefinition(
		"SimpleFlow",
		"A simple flow",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)
	require.NoError(t, err)

	// First access.
	assert.Equal(t, "SimpleFlow", wd.Name())
	assert.Equal(t, "A simple flow", wd.Description())
	assert.Equal(t, "HumanStart", wd.EntryNode())
	assert.Len(t, wd.Nodes(), 2)
	assert.Len(t, wd.Transitions(), 2)
	assert.Len(t, wd.ExitTransitions(), 1)

	// Second access — ensure no mutation between calls.
	assert.Equal(t, "SimpleFlow", wd.Name())
	assert.Equal(t, "A simple flow", wd.Description())
	assert.Equal(t, "HumanStart", wd.EntryNode())
	assert.Len(t, wd.Nodes(), 2)
	assert.Len(t, wd.Transitions(), 2)
	assert.Len(t, wd.ExitTransitions(), 1)
}

// --- Data Independence (Copy Semantics) ---

func TestWorkflowDefinition_NodeSliceCopySemantics(t *testing.T) {
	humanNode, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	nodes := []*Node{humanNode, agentNode}

	wd, err := NewWorkflowDefinition(
		"SimpleFlow",
		"desc",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)
	require.NoError(t, err)

	// Mutate the original nodes slice.
	nodes[0] = nil

	// Getter must still return the original values.
	retrieved := wd.Nodes()
	require.Len(t, retrieved, 2)
	assert.Equal(t, "HumanStart", retrieved[0].Name())
}

func TestWorkflowDefinition_TransitionSliceCopySemantics(t *testing.T) {
	humanNode, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	nodes := []*Node{humanNode, agentNode}

	wd, err := NewWorkflowDefinition(
		"SimpleFlow",
		"desc",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)
	require.NoError(t, err)

	// Mutate the original transitions slice.
	trans[0] = nil

	// Getter must still return the original values.
	retrieved := wd.Transitions()
	require.Len(t, retrieved, 2)
	assert.Equal(t, "HumanStart", retrieved[0].FromNode())
}

func TestWorkflowDefinition_ExitTransitionSliceCopySemantics(t *testing.T) {
	humanNode, agentNode, trans, exitTrans := buildMinimalWorkflowComponents(t)
	nodes := []*Node{humanNode, agentNode}

	wd, err := NewWorkflowDefinition(
		"SimpleFlow",
		"desc",
		"HumanStart",
		nodes,
		trans,
		exitTrans,
	)
	require.NoError(t, err)

	// Mutate the original exit transitions slice.
	exitTrans[0] = nil

	// Getter must still return the original values.
	retrieved := wd.ExitTransitions()
	require.Len(t, retrieved, 1)
	assert.Equal(t, "Worker", retrieved[0].FromNode())
}
