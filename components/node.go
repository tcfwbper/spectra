package components

import "fmt"

// Node represents a discrete step in a workflow where either an AI agent or a
// human performs work. It is a pure immutable value object.
type Node struct {
	name        string
	nodeType    string
	agentRole   string
	description string
}

// NewNode constructs and validates a Node. Returns an error if any constraint
// is violated.
func NewNode(name string, nodeType string, agentRole string, description string) (*Node, error) {
	if name == "" {
		return nil, fmt.Errorf("node name cannot be empty")
	}
	if !isPascalCase(name) {
		return nil, fmt.Errorf("node name must be PascalCase (start with uppercase, alphanumeric only)")
	}

	if nodeType != "agent" && nodeType != "human" {
		return nil, fmt.Errorf("node type must be 'agent' or 'human'")
	}

	if nodeType == "agent" {
		if agentRole == "" {
			return nil, fmt.Errorf("agent_role is required when type is 'agent'")
		}
		if !isPascalCase(agentRole) {
			return nil, fmt.Errorf("agent_role must be PascalCase (start with uppercase, alphanumeric only)")
		}
	}

	if nodeType == "human" {
		if agentRole != "" {
			return nil, fmt.Errorf("agent_role must be empty when type is 'human'")
		}
	}

	return &Node{
		name:        name,
		nodeType:    nodeType,
		agentRole:   agentRole,
		description: description,
	}, nil
}

// Name returns the node's unique identifier within a workflow.
func (n *Node) Name() string { return n.name }

// Type returns the actor type for this step ("agent" or "human").
func (n *Node) Type() string { return n.nodeType }

// AgentRole returns the agent role to invoke (empty for human nodes).
func (n *Node) AgentRole() string { return n.agentRole }

// Description returns the human-readable description of the node's purpose.
func (n *Node) Description() string { return n.description }
