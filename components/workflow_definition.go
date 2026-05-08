package components

import "fmt"

// WorkflowDefinition is the workflow-level aggregator that assembles Nodes,
// Transitions, and ExitTransitions into a validated, immutable workflow graph.
type WorkflowDefinition struct {
	name            string
	description     string
	entryNode       string
	nodes           []*Node
	transitions     []*Transition
	exitTransitions []*ExitTransition
}

// NewWorkflowDefinition constructs and validates a WorkflowDefinition. Returns
// an error if any cross-component constraint is violated.
func NewWorkflowDefinition(
	name string,
	description string,
	entryNode string,
	nodes []*Node,
	transitions []*Transition,
	exitTransitions []*ExitTransition,
) (*WorkflowDefinition, error) {
	// Validate name.
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if !isPascalCase(name) {
		return nil, fmt.Errorf("name must be PascalCase (start with uppercase, alphanumeric only)")
	}

	// Validate non-empty collections.
	if len(nodes) == 0 {
		return nil, fmt.Errorf("nodes cannot be empty")
	}
	if len(transitions) == 0 {
		return nil, fmt.Errorf("transitions cannot be empty")
	}
	if len(exitTransitions) == 0 {
		return nil, fmt.Errorf("exit_transitions cannot be empty")
	}

	// Build node lookup and check uniqueness.
	nodeMap := make(map[string]*Node, len(nodes))
	for _, n := range nodes {
		if _, exists := nodeMap[n.Name()]; exists {
			return nil, fmt.Errorf("duplicate node name: '%s'", n.Name())
		}
		nodeMap[n.Name()] = n
	}

	// Validate entry node.
	entryNodeObj, exists := nodeMap[entryNode]
	if !exists {
		return nil, fmt.Errorf("entry_node '%s' does not reference a valid node", entryNode)
	}
	if entryNodeObj.Type() != "human" {
		return nil, fmt.Errorf("entry_node '%s' must have type 'human'", entryNode)
	}

	// Validate transition referential integrity and determinism.
	transKeySet := make(map[string]bool, len(transitions))
	for _, tr := range transitions {
		if _, ok := nodeMap[tr.FromNode()]; !ok {
			return nil, fmt.Errorf("transition from_node '%s' does not reference a valid node", tr.FromNode())
		}
		if _, ok := nodeMap[tr.ToNode()]; !ok {
			return nil, fmt.Errorf("transition to_node '%s' does not reference a valid node", tr.ToNode())
		}

		key := tr.FromNode() + "|" + tr.EventType()
		if transKeySet[key] {
			return nil, fmt.Errorf("duplicate transition for event '%s' from node '%s'", tr.EventType(), tr.FromNode())
		}
		transKeySet[key] = true
	}

	// Build transition triple set for exit transition correspondence check.
	transTripleSet := make(map[string]bool, len(transitions))
	for _, tr := range transitions {
		triple := tr.FromNode() + "|" + tr.EventType() + "|" + tr.ToNode()
		transTripleSet[triple] = true
	}

	// Validate exit transitions: uniqueness, correspondence, and target type.
	exitTripleSet := make(map[string]bool, len(exitTransitions))
	exitTargetNodes := make(map[string]bool)

	for _, et := range exitTransitions {
		triple := et.FromNode() + "|" + et.EventType() + "|" + et.ToNode()

		// Check uniqueness.
		if exitTripleSet[triple] {
			return nil, fmt.Errorf("duplicate exit_transition (from_node: '%s', event_type: '%s', to_node: '%s')", et.FromNode(), et.EventType(), et.ToNode())
		}
		exitTripleSet[triple] = true

		// Check correspondence.
		if !transTripleSet[triple] {
			return nil, fmt.Errorf("exit_transition (from_node: '%s', event_type: '%s', to_node: '%s') has no corresponding transition", et.FromNode(), et.EventType(), et.ToNode())
		}

		// Check target type is human.
		toNodeObj := nodeMap[et.ToNode()]
		if toNodeObj.Type() != "human" {
			return nil, fmt.Errorf("exit_transition to_node '%s' must have type 'human'", et.ToNode())
		}

		exitTargetNodes[et.ToNode()] = true
	}

	// Validate outgoing transition coverage: every node not targeted by an exit
	// transition must have at least one outgoing transition.
	outgoing := make(map[string]bool, len(nodes))
	for _, tr := range transitions {
		outgoing[tr.FromNode()] = true
	}
	for _, n := range nodes {
		if !exitTargetNodes[n.Name()] && !outgoing[n.Name()] {
			return nil, fmt.Errorf("node '%s' has no outgoing transitions and is not an exit target", n.Name())
		}
	}

	// Validate reachability: every non-entry node must have at least one
	// incoming transition.
	incoming := make(map[string]bool, len(nodes))
	for _, tr := range transitions {
		incoming[tr.ToNode()] = true
	}
	for _, n := range nodes {
		if n.Name() == entryNode {
			continue
		}
		if !incoming[n.Name()] {
			return nil, fmt.Errorf("node '%s' is unreachable (no incoming transitions)", n.Name())
		}
	}

	// Copy slices for immutability.
	nodesCopy := make([]*Node, len(nodes))
	copy(nodesCopy, nodes)

	transCopy := make([]*Transition, len(transitions))
	copy(transCopy, transitions)

	exitTransCopy := make([]*ExitTransition, len(exitTransitions))
	copy(exitTransCopy, exitTransitions)

	return &WorkflowDefinition{
		name:            name,
		description:     description,
		entryNode:       entryNode,
		nodes:           nodesCopy,
		transitions:     transCopy,
		exitTransitions: exitTransCopy,
	}, nil
}

// Name returns the workflow identifier.
func (wd *WorkflowDefinition) Name() string { return wd.name }

// Description returns the human-readable description.
func (wd *WorkflowDefinition) Description() string { return wd.description }

// EntryNode returns the entry node name.
func (wd *WorkflowDefinition) EntryNode() string { return wd.entryNode }

// Nodes returns a copy of the nodes slice.
func (wd *WorkflowDefinition) Nodes() []*Node {
	out := make([]*Node, len(wd.nodes))
	copy(out, wd.nodes)
	return out
}

// Transitions returns a copy of the transitions slice.
func (wd *WorkflowDefinition) Transitions() []*Transition {
	out := make([]*Transition, len(wd.transitions))
	copy(out, wd.transitions)
	return out
}

// ExitTransitions returns a copy of the exit transitions slice.
func (wd *WorkflowDefinition) ExitTransitions() []*ExitTransition {
	out := make([]*ExitTransition, len(wd.exitTransitions))
	copy(out, wd.exitTransitions)
	return out
}
