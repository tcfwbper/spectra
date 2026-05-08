package runtime

import (
	"fmt"
	"io"
	"os"

	"github.com/tcfwbper/spectra/components"
)

// TransitionWorkflowDef defines the read-only interface for workflow definitions
// consumed by TransitionToNode. It provides access to the nodes in the workflow.
type TransitionWorkflowDef interface {
	Nodes() []*components.Node
}

// TransitionAgentDefLoader defines the interface for loading agent definitions
// by agent role name. Used by TransitionToNode to retrieve the AgentDef for
// agent-type nodes.
type TransitionAgentDefLoader interface {
	Load(agentRole string) (AgentDef, error)
}

// TransitionAgentInvoker defines the interface for invoking an agent process.
// Used by TransitionToNode to start the agent for agent-type nodes.
type TransitionAgentInvoker interface {
	InvokeAgent(nodeName, message string, agentDef AgentDef) error
}

// TransitionToNode is responsible for executing the dispatch logic when
// transitioning to a target node in the workflow state machine. It performs
// node-type-specific actions and updates PersistentSession.CurrentState.
type TransitionToNode struct {
	ps      *PersistentSession
	wfDef   TransitionWorkflowDef
	loader  TransitionAgentDefLoader
	invoker TransitionAgentInvoker
	output  io.Writer
}

// TransitionToNodeOption is a functional option for configuring TransitionToNode.
type TransitionToNodeOption func(*TransitionToNode)

// WithOutput sets a custom writer for human node message output.
func WithOutput(w io.Writer) TransitionToNodeOption {
	return func(t *TransitionToNode) {
		t.output = w
	}
}

// NewTransitionToNode constructs a TransitionToNode with the given dependencies
// and applies functional options.
func NewTransitionToNode(ps *PersistentSession, wfDef TransitionWorkflowDef, loader TransitionAgentDefLoader, invoker TransitionAgentInvoker, opts ...TransitionToNodeOption) *TransitionToNode {
	ttn := &TransitionToNode{
		ps:      ps,
		wfDef:   wfDef,
		loader:  loader,
		invoker: invoker,
		output:  os.Stdout,
	}
	for _, opt := range opts {
		opt(ttn)
	}
	return ttn
}

// Execute performs the transition to the target node. It looks up the node,
// performs the node-type-specific action, and updates the session state.
func (ttn *TransitionToNode) Execute(targetNodeName, message string) error {
	// Step 1: Look up the target node in the workflow definition.
	var targetNode *components.Node
	for _, n := range ttn.wfDef.Nodes() {
		if n.Name() == targetNodeName {
			targetNode = n
			break
		}
	}
	if targetNode == nil {
		return fmt.Errorf("target node '%s' not found in workflow", targetNodeName)
	}

	// Step 2: Perform node-type-specific action.
	switch targetNode.Type() {
	case "human":
		displayMsg := message
		if displayMsg == "" {
			displayMsg = "(no message)"
		}
		if _, err := fmt.Fprintf(ttn.output, "[%s] %s\n", targetNodeName, displayMsg); err != nil {
			return err
		}
	case "agent":
		agentDef, err := ttn.loader.Load(targetNode.AgentRole())
		if err != nil {
			return fmt.Errorf("failed to load agent definition for role '%s': %s", targetNode.AgentRole(), err.Error())
		}
		if err := ttn.invoker.InvokeAgent(targetNodeName, message, agentDef); err != nil {
			return fmt.Errorf("failed to invoke agent for node '%s': %s", targetNodeName, err.Error())
		}
	}

	// Step 3: Update session state after successful action.
	if err := ttn.ps.UpdateCurrentStateSafe(targetNodeName); err != nil {
		return fmt.Errorf("failed to update current state: %s", err.Error())
	}

	return nil
}
