package components

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkflowDefinition describes the structure and behavior of an event-driven state machine.
type WorkflowDefinition struct {
	name            string
	description     string
	entryNode       string
	exitTransitions []*ExitTransition
	nodes           []*Node
	transitions     []*Transition
}

// workflowDefinitionYAML is used for YAML marshaling/unmarshaling
type workflowDefinitionYAML struct {
	Name            string                 `yaml:"name"`
	Description     string                 `yaml:"description"`
	EntryNode       string                 `yaml:"entry_node"`
	ExitTransitions []exitTransitionYAML   `yaml:"exit_transitions"`
	Nodes           []nodeYAML             `yaml:"nodes"`
	Transitions     []transitionYAML       `yaml:"transitions"`
}

type exitTransitionYAML struct {
	FromNode  string `yaml:"from_node"`
	EventType string `yaml:"event_type"`
	ToNode    string `yaml:"to_node"`
}

type nodeYAML struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	AgentRole   string `yaml:"agent_role,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type transitionYAML struct {
	FromNode  string `yaml:"from_node"`
	EventType string `yaml:"event_type"`
	ToNode    string `yaml:"to_node"`
}

// NewWorkflowDefinition creates a new WorkflowDefinition with the given parameters.
func NewWorkflowDefinition(name, description, entryNode string, exitTransitions []*ExitTransition, nodes []*Node, transitions []*Transition) (*WorkflowDefinition, error) {
	// Validate name
	if name == "" {
		return nil, fmt.Errorf("workflow name must be non-empty")
	}
	if !pascalCasePattern.MatchString(name) {
		if hasSpaces(name) || hasSpecialChars(name) {
			return nil, fmt.Errorf("workflow name must be PascalCase with no spaces or special characters")
		}
		return nil, fmt.Errorf("workflow name must be PascalCase")
	}

	// Validate exit transitions non-empty
	if len(exitTransitions) == 0 {
		return nil, fmt.Errorf("at least one exit transition required")
	}

	// Validate nodes non-empty
	if len(nodes) == 0 {
		return nil, fmt.Errorf("at least one node required")
	}

	// Validate transitions non-empty
	if len(transitions) == 0 {
		return nil, fmt.Errorf("at least one transition required")
	}

	// Build node map for validation
	nodeMap := make(map[string]*Node)
	for _, node := range nodes {
		if _, exists := nodeMap[node.GetName()]; exists {
			return nil, fmt.Errorf("duplicate node name: %s", node.GetName())
		}
		nodeMap[node.GetName()] = node
	}

	// Validate entry node
	entryNodeObj, exists := nodeMap[entryNode]
	if !exists {
		return nil, fmt.Errorf("entry node '%s' not found", entryNode)
	}
	if entryNodeObj.GetType() != "human" {
		return nil, fmt.Errorf("entry node '%s' must have type 'human', but has type '%s'", entryNode, entryNodeObj.GetType())
	}

	// Build transition map for validation
	transitionMap := make(map[string]bool)
	for _, t := range transitions {
		key := fmt.Sprintf("%s|%s|%s", t.GetFromNode(), t.GetEventType(), t.GetToNode())
		transitionMap[key] = true

		// Validate transition nodes exist
		if _, exists := nodeMap[t.GetFromNode()]; !exists {
			return nil, fmt.Errorf("transition references undefined node: %s", t.GetFromNode())
		}
		if _, exists := nodeMap[t.GetToNode()]; !exists {
			return nil, fmt.Errorf("transition references undefined node: %s", t.GetToNode())
		}
	}

	// Validate exit transitions
	exitTargetNodes := make(map[string]bool)
	for _, et := range exitTransitions {
		key := fmt.Sprintf("%s|%s|%s", et.GetFromNode(), et.GetEventType(), et.GetToNode())
		if !transitionMap[key] {
			return nil, fmt.Errorf("exit transition (from_node: %s, event_type: %s, to_node: %s) has no corresponding transition", et.GetFromNode(), et.GetEventType(), et.GetToNode())
		}

		// Validate exit transition target is human
		toNode := nodeMap[et.GetToNode()]
		if toNode.GetType() != "human" {
			return nil, fmt.Errorf("exit transition to_node '%s' must target a human node, but has type '%s'", et.GetToNode(), toNode.GetType())
		}

		exitTargetNodes[et.GetToNode()] = true
	}

	// Build incoming transition map
	incomingTransitions := make(map[string]bool)
	for _, t := range transitions {
		incomingTransitions[t.GetToNode()] = true
	}

	// Build outgoing transition map
	outgoingTransitions := make(map[string]bool)
	for _, t := range transitions {
		outgoingTransitions[t.GetFromNode()] = true
	}

	// Validate all nodes (except entry node) have incoming transitions
	for _, node := range nodes {
		nodeName := node.GetName()
		if nodeName != entryNode && !incomingTransitions[nodeName] {
			return nil, fmt.Errorf("unreachable node '%s' (no incoming transitions)", nodeName)
		}
	}

	// Validate nodes not targeted by exit transitions have outgoing transitions
	for _, node := range nodes {
		nodeName := node.GetName()
		if !exitTargetNodes[nodeName] && !outgoingTransitions[nodeName] {
			return nil, fmt.Errorf("node '%s' has no outgoing transitions and is not an exit target", nodeName)
		}
	}

	return &WorkflowDefinition{
		name:            name,
		description:     description,
		entryNode:       entryNode,
		exitTransitions: exitTransitions,
		nodes:           nodes,
		transitions:     transitions,
	}, nil
}

// LoadWorkflowDefinition loads a WorkflowDefinition from a YAML file.
func LoadWorkflowDefinition(yamlPath string) (*WorkflowDefinition, error) {
	// Check if file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workflow not found")
	}

	// Read file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading workflow file: %w", err)
	}

	// Parse YAML
	var yamlData workflowDefinitionYAML
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("parsing workflow YAML: %w", err)
	}

	// Validate required fields from YAML
	if yamlData.EntryNode == "" {
		return nil, fmt.Errorf("entry_node is required")
	}

	// Convert YAML nodes to Node objects
	nodes := make([]*Node, 0, len(yamlData.Nodes))
	for _, nodeYAML := range yamlData.Nodes {
		node, err := NewNode(nodeYAML.Name, nodeYAML.Type, nodeYAML.AgentRole, nodeYAML.Description)
		if err != nil {
			return nil, fmt.Errorf("creating node: %w", err)
		}
		nodes = append(nodes, node)
	}

	// Convert YAML transitions to Transition objects
	transitions := make([]*Transition, 0, len(yamlData.Transitions))
	for _, transitionYAML := range yamlData.Transitions {
		transition, err := NewTransition(transitionYAML.FromNode, transitionYAML.EventType, transitionYAML.ToNode)
		if err != nil {
			return nil, fmt.Errorf("creating transition: %w", err)
		}
		transitions = append(transitions, transition)
	}

	// Convert YAML exit transitions to ExitTransition objects
	exitTransitions := make([]*ExitTransition, 0, len(yamlData.ExitTransitions))
	for _, exitTransitionYAML := range yamlData.ExitTransitions {
		exitTransition, err := NewExitTransition(exitTransitionYAML.FromNode, exitTransitionYAML.EventType, exitTransitionYAML.ToNode)
		if err != nil {
			return nil, fmt.Errorf("creating exit transition: %w", err)
		}
		exitTransitions = append(exitTransitions, exitTransition)
	}

	// Create workflow definition
	workflow, err := NewWorkflowDefinition(
		yamlData.Name,
		yamlData.Description,
		yamlData.EntryNode,
		exitTransitions,
		nodes,
		transitions,
	)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

// SaveToFile saves the WorkflowDefinition to a YAML file.
func (w *WorkflowDefinition) SaveToFile(yamlPath string) error {
	// Convert nodes to YAML format
	nodesYAML := make([]nodeYAML, 0, len(w.nodes))
	for _, node := range w.nodes {
		nodeYAML := nodeYAML{
			Name:        node.GetName(),
			Type:        node.GetType(),
			AgentRole:   node.GetAgentRole(),
			Description: node.GetDescription(),
		}
		nodesYAML = append(nodesYAML, nodeYAML)
	}

	// Convert transitions to YAML format
	transitionsYAML := make([]transitionYAML, 0, len(w.transitions))
	for _, transition := range w.transitions {
		transitionYAML := transitionYAML{
			FromNode:  transition.GetFromNode(),
			EventType: transition.GetEventType(),
			ToNode:    transition.GetToNode(),
		}
		transitionsYAML = append(transitionsYAML, transitionYAML)
	}

	// Convert exit transitions to YAML format
	exitTransitionsYAML := make([]exitTransitionYAML, 0, len(w.exitTransitions))
	for _, exitTransition := range w.exitTransitions {
		exitTransitionYAML := exitTransitionYAML{
			FromNode:  exitTransition.GetFromNode(),
			EventType: exitTransition.GetEventType(),
			ToNode:    exitTransition.GetToNode(),
		}
		exitTransitionsYAML = append(exitTransitionsYAML, exitTransitionYAML)
	}

	yamlData := workflowDefinitionYAML{
		Name:            w.name,
		Description:     w.description,
		EntryNode:       w.entryNode,
		ExitTransitions: exitTransitionsYAML,
		Nodes:           nodesYAML,
		Transitions:     transitionsYAML,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&yamlData)
	if err != nil {
		return fmt.Errorf("marshaling workflow to YAML: %w", err)
	}

	// Adjust description formatting for empty strings
	if w.description == "" {
		dataStr := string(data)
		dataStr = strings.Replace(dataStr, "description: \"\"\n", "description: \"\"\n", 1)
		data = []byte(dataStr)
	}

	// Write to file
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("writing workflow file: %w", err)
	}

	return nil
}

// GetName returns the workflow name.
func (w *WorkflowDefinition) GetName() string {
	return w.name
}

// GetDescription returns the workflow description.
func (w *WorkflowDefinition) GetDescription() string {
	return w.description
}

// GetEntryNode returns the workflow entry node.
func (w *WorkflowDefinition) GetEntryNode() string {
	return w.entryNode
}

// GetExitTransitions returns the workflow exit transitions.
func (w *WorkflowDefinition) GetExitTransitions() []*ExitTransition {
	return w.exitTransitions
}

// GetNodes returns the workflow nodes.
func (w *WorkflowDefinition) GetNodes() []*Node {
	return w.nodes
}

// GetTransitions returns the workflow transitions.
func (w *WorkflowDefinition) GetTransitions() []*Transition {
	return w.transitions
}

// WorkflowRegistry manages a collection of WorkflowDefinitions and enforces uniqueness.
type WorkflowRegistry struct {
	workflows map[string]*WorkflowDefinition
}

// NewWorkflowRegistry creates a new WorkflowRegistry.
func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		workflows: make(map[string]*WorkflowDefinition),
	}
}

// Register adds a WorkflowDefinition to the registry.
func (r *WorkflowRegistry) Register(workflow *WorkflowDefinition) error {
	if _, exists := r.workflows[workflow.name]; exists {
		return fmt.Errorf("workflow '%s' already exists", workflow.name)
	}
	r.workflows[workflow.name] = workflow
	return nil
}

// Get retrieves a WorkflowDefinition by name.
func (r *WorkflowRegistry) Get(name string) (*WorkflowDefinition, bool) {
	workflow, exists := r.workflows[name]
	return workflow, exists
}
