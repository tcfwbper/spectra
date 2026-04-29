package components

import (
	"fmt"
	"regexp"
)

var pascalCasePattern = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

// Node represents a discrete step in a workflow where either an AI agent or a human performs work.
type Node struct {
	name        string
	nodeType    string
	agentRole   string
	description string
}

// NewNode creates a new Node with the given parameters.
func NewNode(name, nodeType, agentRole, description string) (*Node, error) {
	// Validate name
	if name == "" {
		return nil, fmt.Errorf("node name must be non-empty")
	}
	if !pascalCasePattern.MatchString(name) {
		if hasSpaces(name) {
			return nil, fmt.Errorf("node name must be PascalCase with no spaces")
		}
		if hasSpecialChars(name) {
			return nil, fmt.Errorf("node name must be PascalCase with no special characters")
		}
		return nil, fmt.Errorf("node name must be PascalCase")
	}

	// Validate type
	if nodeType == "" {
		return nil, fmt.Errorf("node type is required")
	}
	if nodeType != "agent" && nodeType != "human" {
		return nil, fmt.Errorf("node type must be either 'agent' or 'human'")
	}

	// Validate agent_role constraints
	switch nodeType {
	case "agent":
		if agentRole == "" {
			return nil, fmt.Errorf("agent_role is required when node type is 'agent'")
		}
		if !pascalCasePattern.MatchString(agentRole) {
			return nil, fmt.Errorf("agent_role must be PascalCase")
		}
	case "human":
		if agentRole != "" {
			return nil, fmt.Errorf("agent_role must be empty when node type is 'human'")
		}
	}

	return &Node{
		name:        name,
		nodeType:    nodeType,
		agentRole:   agentRole,
		description: description,
	}, nil
}

// GetName returns the node name.
func (n *Node) GetName() string {
	return n.name
}

// GetType returns the node type.
func (n *Node) GetType() string {
	return n.nodeType
}

// GetAgentRole returns the agent role.
func (n *Node) GetAgentRole() string {
	return n.agentRole
}

// GetDescription returns the node description.
func (n *Node) GetDescription() string {
	return n.description
}

func hasSpaces(s string) bool {
	for _, r := range s {
		if r == ' ' {
			return true
		}
	}
	return false
}

func hasSpecialChars(s string) bool {
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			return true
		}
	}
	return false
}
