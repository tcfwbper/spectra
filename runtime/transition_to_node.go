package runtime

import (
	"fmt"

	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// SessionForTransition defines the interface that TransitionToNode needs from Session.
type SessionForTransition interface {
	UpdateCurrentStateSafe(newState string) error
	Done(terminationNotifier chan<- struct{}) error
	Fail(err error, terminationNotifier chan<- struct{}) error
}

// AgentDefinitionLoaderForTransition defines the interface for loading agent definitions.
type AgentDefinitionLoaderForTransition interface {
	Load(agentRole string) (*storage.AgentDefinition, error)
}

// AgentInvokerForTransition defines the interface for invoking agents.
type AgentInvokerForTransition interface {
	InvokeAgent(nodeName string, message string, agentDef storage.AgentDefinition) error
}

// TransitionToNode is responsible for executing the dispatch logic for transitioning
// to a single node in the workflow state machine.
type TransitionToNode struct {
	session             SessionForTransition
	workflowDef         *storage.WorkflowDefinition
	agentDefLoader      AgentDefinitionLoaderForTransition
	agentInvoker        AgentInvokerForTransition
	terminationNotifier chan<- struct{}
}

// NewTransitionToNode creates a new TransitionToNode instance.
func NewTransitionToNode(
	sess SessionForTransition,
	wfDef *storage.WorkflowDefinition,
	agentDefLoader AgentDefinitionLoaderForTransition,
	agentInvoker AgentInvokerForTransition,
	terminationNotifier chan<- struct{},
) (*TransitionToNode, error) {
	if sess == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}
	if wfDef == nil {
		return nil, fmt.Errorf("workflow definition cannot be nil")
	}
	if agentDefLoader == nil {
		return nil, fmt.Errorf("agent definition loader cannot be nil")
	}
	if agentInvoker == nil {
		return nil, fmt.Errorf("agent invoker cannot be nil")
	}
	if terminationNotifier == nil {
		return nil, fmt.Errorf("termination notifier cannot be nil")
	}

	return &TransitionToNode{
		session:             sess,
		workflowDef:         wfDef,
		agentDefLoader:      agentDefLoader,
		agentInvoker:        agentInvoker,
		terminationNotifier: terminationNotifier,
	}, nil
}

// Transition executes the transition to the target node.
func (t *TransitionToNode) Transition(message string, targetNodeName string, isExitTransition bool) error {
	// Step 3: Load the target node definition from WorkflowDefinition
	targetNode := t.findNode(targetNodeName)
	if targetNode == nil {
		// Step 4: Target node not found - construct RuntimeError and call Session.Fail
		runtimeErr := t.createRuntimeError(
			fmt.Sprintf("target node not found: '%s'", targetNodeName),
			nil,
		)
		_ = t.session.Fail(runtimeErr, t.terminationNotifier)
		return fmt.Errorf("target node '%s' not found in workflow", targetNodeName)
	}

	// Step 5-6: Perform node-type-specific actions (skip if exit transition)
	if !isExitTransition {
		if err := t.executeNodeAction(message, targetNodeName, targetNode); err != nil {
			return err
		}
	}

	// Step 7: Update Session.CurrentState
	_ = t.session.UpdateCurrentStateSafe(targetNodeName)

	// Step 9-10: Handle exit transition
	if isExitTransition {
		if err := t.session.Done(t.terminationNotifier); err != nil {
			// Session.Done failed - construct RuntimeError and call Session.Fail
			runtimeErr := t.createRuntimeError(
				"failed to complete session after exit transition",
				err,
			)
			_ = t.session.Fail(runtimeErr, t.terminationNotifier)
			return fmt.Errorf("failed to complete session after exit transition: %v", err)
		}
	}

	// Step 11: Return success
	return nil
}

// findNode finds a node by name in the workflow definition.
func (t *TransitionToNode) findNode(nodeName string) *storage.Node {
	for i := range t.workflowDef.Nodes {
		if t.workflowDef.Nodes[i].Name == nodeName {
			return &t.workflowDef.Nodes[i]
		}
	}
	return nil
}

// executeNodeAction performs the node-type-specific action.
func (t *TransitionToNode) executeNodeAction(message string, targetNodeName string, targetNode *storage.Node) error {
	switch targetNode.Type {
	case "human":
		// Print message to stdout
		displayMessage := message
		if message == "" {
			displayMessage = "(no message)"
		}
		fmt.Printf("[Human Node: %s] %s\n", targetNodeName, displayMessage)
		return nil

	case "agent":
		// Load agent definition
		agentDef, err := t.agentDefLoader.Load(targetNode.AgentRole)
		if err != nil {
			// Agent definition load failed - construct RuntimeError and call Session.Fail
			runtimeErr := t.createRuntimeError(
				fmt.Sprintf("failed to load agent definition for role '%s'", targetNode.AgentRole),
				err,
			)
			_ = t.session.Fail(runtimeErr, t.terminationNotifier)
			return fmt.Errorf("failed to load agent definition for role '%s': %v", targetNode.AgentRole, err)
		}

		// Invoke agent
		if err := t.agentInvoker.InvokeAgent(targetNodeName, message, *agentDef); err != nil {
			// Agent invocation failed - construct RuntimeError and call Session.Fail
			runtimeErr := t.createRuntimeError(
				fmt.Sprintf("failed to invoke agent for node '%s'", targetNodeName),
				err,
			)
			_ = t.session.Fail(runtimeErr, t.terminationNotifier)
			return fmt.Errorf("failed to invoke agent for node '%s': %v", targetNodeName, err)
		}
		return nil

	default:
		return fmt.Errorf("unknown node type: %s", targetNode.Type)
	}
}

// createRuntimeError creates a RuntimeError with the specified message and details.
func (t *TransitionToNode) createRuntimeError(message string, detailErr error) *session.RuntimeError {
	return &session.RuntimeError{
		Issuer:  "TransitionToNode",
		Message: message,
	}
}
