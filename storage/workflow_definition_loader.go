package storage

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Node represents a node in a workflow.
type Node struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	AgentRole   string `yaml:"agent_role,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Transition represents a state transition in a workflow.
type Transition struct {
	FromNode  string `yaml:"from_node"`
	EventType string `yaml:"event_type"`
	ToNode    string `yaml:"to_node"`
}

// ExitTransition represents a workflow exit transition.
type ExitTransition struct {
	FromNode  string `yaml:"from_node"`
	EventType string `yaml:"event_type"`
	ToNode    string `yaml:"to_node"`
}

// WorkflowDefinition represents a workflow configuration loaded from YAML.
type WorkflowDefinition struct {
	Name            string           `yaml:"name"`
	Description     string           `yaml:"description,omitempty"`
	EntryNode       string           `yaml:"entry_node"`
	ExitTransitions []ExitTransition `yaml:"exit_transitions"`
	Nodes           []Node           `yaml:"nodes"`
	Transitions     []Transition     `yaml:"transitions"`
}

// AgentLoader is an interface for loading agent definitions
type AgentLoader interface {
	Load(agentRole string) (*AgentDefinition, error)
}

// WorkflowDefinitionLoader loads workflow definitions from .spectra/workflows/.
type WorkflowDefinitionLoader struct {
	projectRoot string
	agentLoader AgentLoader
}

// NewWorkflowDefinitionLoader creates a new WorkflowDefinitionLoader
func NewWorkflowDefinitionLoader(projectRoot string, agentLoader AgentLoader) *WorkflowDefinitionLoader {
	return &WorkflowDefinitionLoader{
		projectRoot: projectRoot,
		agentLoader: agentLoader,
	}
}

// Load loads a workflow definition from disk.
func (l *WorkflowDefinitionLoader) Load(workflowName string) (*WorkflowDefinition, error) {
	// Compose the file path
	workflowPath := GetWorkflowPath(l.projectRoot, workflowName)

	// Read the file
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("workflow definition not found: %s", workflowName)
		}
		return nil, fmt.Errorf("failed to read workflow definition '%s': %w", workflowName, err)
	}

	// Check for empty file before parsing
	if len(data) == 0 {
		return nil, fmt.Errorf("failed to parse workflow definition '%s': EOF", workflowName)
	}

	// Parse YAML
	var def WorkflowDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition '%s': %w", workflowName, err)
	}

	// Validate required fields
	if def.Name == "" {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: missing required field 'name'", workflowName)
	}
	if def.EntryNode == "" {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: missing required field 'entry_node'", workflowName)
	}
	if len(def.Nodes) == 0 {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: missing required field 'nodes'", workflowName)
	}
	if len(def.Transitions) == 0 {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: missing required field 'transitions'", workflowName)
	}
	if len(def.ExitTransitions) == 0 {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: missing required field 'exit_transitions'", workflowName)
	}

	// Validate name format (PascalCase)
	pascalCasePattern := regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	if !pascalCasePattern.MatchString(def.Name) {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: name must be PascalCase with no spaces or special characters", workflowName)
	}

	// Build node map for quick lookup
	nodeMap := make(map[string]*Node)
	for i := range def.Nodes {
		node := &def.Nodes[i]
		if _, exists := nodeMap[node.Name]; exists {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: duplicate node name '%s'", workflowName, node.Name)
		}
		nodeMap[node.Name] = node
	}

	// Validate entry node exists
	entryNode, exists := nodeMap[def.EntryNode]
	if !exists {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: entry_node '%s' references non-existent node", workflowName, def.EntryNode)
	}

	// Validate entry node is human type
	if entryNode.Type != "human" {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: entry_node '%s' must have type 'human', but has type '%s'", workflowName, def.EntryNode, entryNode.Type)
	}

	// Validate all transitions reference valid nodes
	transitionKeys := make(map[string]bool)
	nodesWithIncoming := make(map[string]bool)
	for _, t := range def.Transitions {
		// Check from_node exists
		if _, exists := nodeMap[t.FromNode]; !exists {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: transition references non-existent node '%s'", workflowName, t.FromNode)
		}
		// Check to_node exists
		if _, exists := nodeMap[t.ToNode]; !exists {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: transition references non-existent node '%s'", workflowName, t.ToNode)
		}
		// Check no self-loop
		if t.FromNode == t.ToNode {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: transition from_node and to_node must be different (node '%s', event '%s')", workflowName, t.FromNode, t.EventType)
		}
		// Check no duplicate (from_node, event_type) pairs
		key := t.FromNode + "|" + t.EventType
		if transitionKeys[key] {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: duplicate transition for event '%s' from node '%s'", workflowName, t.EventType, t.FromNode)
		}
		transitionKeys[key] = true
		nodesWithIncoming[t.ToNode] = true
	}

	// Validate exit transitions
	exitTransitionSet := make(map[string]bool)
	exitTargets := make(map[string]bool)
	for _, et := range def.ExitTransitions {
		// Check for duplicate exit transitions
		exitKey := fmt.Sprintf("%s|%s|%s", et.FromNode, et.EventType, et.ToNode)
		if exitTransitionSet[exitKey] {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: duplicate exit transition (from_node: '%s', event_type: '%s', to_node: '%s')", workflowName, et.FromNode, et.EventType, et.ToNode)
		}
		exitTransitionSet[exitKey] = true

		// Check that exit transition has corresponding transition definition
		transKey := et.FromNode + "|" + et.EventType
		if !transitionKeys[transKey] {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: exit transition (from_node: '%s', event_type: '%s', to_node: '%s') has no corresponding transition definition", workflowName, et.FromNode, et.EventType, et.ToNode)
		}

		// Verify the exact match of to_node in transitions
		found := false
		for _, t := range def.Transitions {
			if t.FromNode == et.FromNode && t.EventType == et.EventType && t.ToNode == et.ToNode {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: exit transition (from_node: '%s', event_type: '%s', to_node: '%s') has no corresponding transition definition", workflowName, et.FromNode, et.EventType, et.ToNode)
		}

		// Check that to_node is human type
		toNode, exists := nodeMap[et.ToNode]
		if exists && toNode.Type != "human" {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: exit transition (from_node: '%s', event_type: '%s', to_node: '%s') must target a human node, but targets '%s' with type '%s'", workflowName, et.FromNode, et.EventType, et.ToNode, et.ToNode, toNode.Type)
		}

		exitTargets[et.ToNode] = true
	}

	// Validate that non-exit-target nodes have outgoing transitions
	// Exception: unreachable nodes (no incoming transitions) are allowed
	for nodeName := range nodeMap {
		// Skip if it's an exit target
		if exitTargets[nodeName] {
			continue
		}
		// Skip if the node is unreachable (no incoming transitions and not the entry node)
		if !nodesWithIncoming[nodeName] && nodeName != def.EntryNode {
			continue
		}
		// Check if it has outgoing transitions
		hasOutgoing := false
		for _, t := range def.Transitions {
			if t.FromNode == nodeName {
				hasOutgoing = true
				break
			}
		}
		if !hasOutgoing {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: node '%s' has no outgoing transitions and is not an exit target", workflowName, nodeName)
		}
	}

	// Validate agent references
	validatedAgents := make(map[string]bool)
	for _, node := range def.Nodes {
		if node.Type == "agent" {
			// Only validate each agent role once
			if !validatedAgents[node.AgentRole] {
				_, err := l.agentLoader.Load(node.AgentRole)
				if err != nil {
					return nil, fmt.Errorf("workflow definition '%s' validation failed: node '%s' references invalid agent_role '%s': %w", workflowName, node.Name, node.AgentRole, err)
				}
				validatedAgents[node.AgentRole] = true
			}
		}
	}

	return &def, nil
}
