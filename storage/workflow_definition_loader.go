package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spectra-ai/spectra/components"
)

// AgentLoader is the interface that WorkflowDefinitionLoader depends on for
// agent-role referential integrity validation. Defined in the consumer package
// per Go convention.
type AgentLoader interface {
	Load(agentRole string) (*components.AgentDefinition, error)
}

// workflowYAML is the internal representation for strict YAML parsing of
// workflow definition files. Fields use camelCase yaml tags.
type workflowYAML struct {
	Description     string           `yaml:"description"`
	EntryNode       string           `yaml:"entryNode"`
	Nodes           []nodeYAML       `yaml:"nodes"`
	Transitions     []transitionYAML `yaml:"transitions"`
	ExitTransitions []transitionYAML `yaml:"exitTransitions"`
}

// nodeYAML represents a single node entry in the YAML nodes array.
type nodeYAML struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	AgentRole   string `yaml:"agentRole"`
	Description string `yaml:"description"`
}

// transitionYAML represents a single transition entry in the YAML arrays.
type transitionYAML struct {
	FromNode  string `yaml:"fromNode"`
	EventType string `yaml:"eventType"`
	ToNode    string `yaml:"toNode"`
}

// WorkflowDefinitionLoader provides read-only access to workflow definition
// YAML files stored in .spectra/workflows/. It is stateless, does not cache,
// and is safe for concurrent use.
type WorkflowDefinitionLoader struct {
	projectRoot string
	agentLoader AgentLoader
}

// NewWorkflowDefinitionLoader creates a WorkflowDefinitionLoader for the given
// project root directory and injected agent loader.
func NewWorkflowDefinitionLoader(projectRoot string, agentLoader AgentLoader) *WorkflowDefinitionLoader {
	return &WorkflowDefinitionLoader{
		projectRoot: projectRoot,
		agentLoader: agentLoader,
	}
}

// Load reads, parses, and validates a workflow definition YAML file. The name
// is derived from the workflowName parameter (filename without .yaml extension).
func (l *WorkflowDefinitionLoader) Load(workflowName string) (*components.WorkflowDefinition, error) {
	// Compose file path via StorageLayout.
	filePath := GetWorkflowPath(l.projectRoot, workflowName)

	// Read file.
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("workflow definition not found: %s", workflowName)
		}
		return nil, fmt.Errorf("failed to read workflow definition '%s': %v", workflowName, err)
	}

	// Parse YAML with strict mode (unknown fields rejected).
	var raw workflowYAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition '%s': %v", workflowName, err)
	}

	// Construct Nodes (fail fast on first error).
	nodes := make([]*components.Node, 0, len(raw.Nodes))
	for i, n := range raw.Nodes {
		node, err := components.NewNode(n.Name, n.Type, n.AgentRole, n.Description)
		if err != nil {
			if n.Name != "" {
				return nil, fmt.Errorf("workflow definition '%s' validation failed: node '%s': %v", workflowName, n.Name, err)
			}
			return nil, fmt.Errorf("workflow definition '%s' validation failed: node[%d]: %v", workflowName, i, err)
		}
		nodes = append(nodes, node)
	}

	// Construct Transitions (fail fast on first error).
	transitions := make([]*components.Transition, 0, len(raw.Transitions))
	for _, tr := range raw.Transitions {
		t, err := components.NewTransition(tr.FromNode, tr.EventType, tr.ToNode)
		if err != nil {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: transition (from '%s', event '%s', to '%s'): %v", workflowName, tr.FromNode, tr.EventType, tr.ToNode, err)
		}
		transitions = append(transitions, t)
	}

	// Construct ExitTransitions (fail fast on first error).
	exitTransitions := make([]*components.ExitTransition, 0, len(raw.ExitTransitions))
	for _, et := range raw.ExitTransitions {
		ext, err := components.NewExitTransition(et.FromNode, et.EventType, et.ToNode)
		if err != nil {
			return nil, fmt.Errorf("workflow definition '%s' validation failed: exit_transition (from '%s', event '%s', to '%s'): %v", workflowName, et.FromNode, et.EventType, et.ToNode, err)
		}
		exitTransitions = append(exitTransitions, ext)
	}

	// Construct WorkflowDefinition via constructor (name derived from filename).
	def, err := components.NewWorkflowDefinition(
		workflowName,
		raw.Description,
		raw.EntryNode,
		nodes,
		transitions,
		exitTransitions,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow definition '%s' validation failed: %v", workflowName, err)
	}

	// Validate agent_role referential integrity (fail fast on first error).
	for _, node := range nodes {
		if node.Type() == "agent" {
			_, err := l.agentLoader.Load(node.AgentRole())
			if err != nil {
				return nil, fmt.Errorf("workflow definition '%s' validation failed: node '%s' references invalid agent_role '%s': %v", workflowName, node.Name(), node.AgentRole(), err)
			}
		}
	}

	return def, nil
}
