package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentDefinition describes the metadata for an AI agent role.
type AgentDefinition struct {
	role            string
	model           string
	effort          string
	systemPrompt    string
	agentRoot       string
	allowedTools    []string
	disallowedTools []string
}

// agentDefinitionYAML is used for YAML marshaling/unmarshaling
type agentDefinitionYAML struct {
	Role            string   `yaml:"role"`
	Model           string   `yaml:"model"`
	Effort          string   `yaml:"effort"`
	SystemPrompt    string   `yaml:"system_prompt"`
	AgentRoot       string   `yaml:"agent_root"`
	AllowedTools    []string `yaml:"allowed_tools"`
	DisallowedTools []string `yaml:"disallowed_tools"`
}

// NewAgentDefinition creates a new AgentDefinition with the given parameters.
func NewAgentDefinition(role, model, effort, systemPrompt, agentRoot string, allowedTools, disallowedTools []string) (*AgentDefinition, error) {
	// Validate role
	if role == "" {
		return nil, fmt.Errorf("role must be non-empty")
	}
	if !pascalCasePattern.MatchString(role) {
		if hasSpaces(role) || hasSpecialChars(role) {
			return nil, fmt.Errorf("role must be PascalCase with no spaces or special characters")
		}
		return nil, fmt.Errorf("role must be PascalCase")
	}

	// Validate model
	if model == "" {
		return nil, fmt.Errorf("model must be non-empty")
	}

	// Validate effort
	if effort == "" {
		return nil, fmt.Errorf("effort must be non-empty")
	}

	// Validate system_prompt
	if systemPrompt == "" {
		return nil, fmt.Errorf("system_prompt must be non-empty")
	}

	// Validate agent_root
	if agentRoot == "" {
		return nil, fmt.Errorf("agent_root must be non-empty")
	}
	if filepath.IsAbs(agentRoot) || isWindowsAbsPath(agentRoot) {
		return nil, fmt.Errorf("agent_root must be a relative path")
	}

	return &AgentDefinition{
		role:            role,
		model:           model,
		effort:          effort,
		systemPrompt:    systemPrompt,
		agentRoot:       agentRoot,
		allowedTools:    allowedTools,
		disallowedTools: disallowedTools,
	}, nil
}

// LoadAgentDefinition loads an AgentDefinition from a YAML file.
func LoadAgentDefinition(yamlPath string) (*AgentDefinition, error) {
	// Check if file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("agent not found")
	}

	// Read file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading agent file: %w", err)
	}

	// Parse YAML
	var yamlData agentDefinitionYAML
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("parsing agent YAML: %w", err)
	}

	// Validate required fields from YAML
	if yamlData.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	// Create agent definition
	agent, err := NewAgentDefinition(
		yamlData.Role,
		yamlData.Model,
		yamlData.Effort,
		yamlData.SystemPrompt,
		yamlData.AgentRoot,
		yamlData.AllowedTools,
		yamlData.DisallowedTools,
	)
	if err != nil {
		return nil, err
	}

	// Verify agent_root directory exists
	// Get the directory containing the YAML file
	yamlDir := filepath.Dir(yamlPath)
	// Navigate up to find spectra root (parent of .spectra)
	spectraRoot := filepath.Dir(filepath.Dir(yamlDir))
	agentRootPath := filepath.Join(spectraRoot, agent.agentRoot)
	if _, err := os.Stat(agentRootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("agent_root directory not found: %s", agent.agentRoot)
	}

	return agent, nil
}

// SaveToFile saves the AgentDefinition to a YAML file.
func (a *AgentDefinition) SaveToFile(yamlPath string) error {
	yamlData := agentDefinitionYAML{
		Role:            a.role,
		Model:           a.model,
		Effort:          a.effort,
		SystemPrompt:    a.systemPrompt,
		AgentRoot:       a.agentRoot,
		AllowedTools:    a.allowedTools,
		DisallowedTools: a.disallowedTools,
	}

	// Marshal to YAML with custom formatting for multi-line strings
	data, err := yaml.Marshal(&yamlData)
	if err != nil {
		return fmt.Errorf("marshaling agent to YAML: %w", err)
	}

	// Check if system_prompt contains newlines and adjust formatting
	if strings.Contains(a.systemPrompt, "\n") {
		// Re-marshal with literal style for system_prompt
		var node yaml.Node
		if err := node.Encode(&yamlData); err != nil {
			return fmt.Errorf("encoding agent to YAML node: %w", err)
		}

		// Find system_prompt field and set it to literal style
		for i := 0; i < len(node.Content[0].Content); i += 2 {
			if node.Content[0].Content[i].Value == "system_prompt" {
				node.Content[0].Content[i+1].Style = yaml.LiteralStyle
				break
			}
		}

		data, err = yaml.Marshal(&node)
		if err != nil {
			return fmt.Errorf("marshaling agent with literal style: %w", err)
		}
	}

	// Write to file
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("writing agent file: %w", err)
	}

	return nil
}

// GetRole returns the agent role.
func (a *AgentDefinition) GetRole() string {
	return a.role
}

// GetModel returns the agent model.
func (a *AgentDefinition) GetModel() string {
	return a.model
}

// GetEffort returns the agent effort.
func (a *AgentDefinition) GetEffort() string {
	return a.effort
}

// GetSystemPrompt returns the agent system prompt.
func (a *AgentDefinition) GetSystemPrompt() string {
	return a.systemPrompt
}

// GetAgentRoot returns the agent root directory.
func (a *AgentDefinition) GetAgentRoot() string {
	return a.agentRoot
}

// GetAllowedTools returns the list of allowed tools.
func (a *AgentDefinition) GetAllowedTools() []string {
	return a.allowedTools
}

// GetDisallowedTools returns the list of disallowed tools.
func (a *AgentDefinition) GetDisallowedTools() []string {
	return a.disallowedTools
}

// AgentRegistry manages a collection of AgentDefinitions and enforces uniqueness.
type AgentRegistry struct {
	agents map[string]*AgentDefinition
}

// NewAgentRegistry creates a new AgentRegistry.
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentDefinition),
	}
}

// Register adds an AgentDefinition to the registry.
func (r *AgentRegistry) Register(agent *AgentDefinition) error {
	if _, exists := r.agents[agent.role]; exists {
		return fmt.Errorf("agent '%s' already exists", agent.role)
	}
	r.agents[agent.role] = agent
	return nil
}

// Get retrieves an AgentDefinition by role.
func (r *AgentRegistry) Get(role string) (*AgentDefinition, bool) {
	agent, exists := r.agents[role]
	return agent, exists
}

// isWindowsAbsPath checks if the path is a Windows absolute path (e.g., C:\path)
func isWindowsAbsPath(path string) bool {
	if len(path) < 3 {
		return false
	}
	// Check for drive letter pattern: X:\
	if (path[0] >= 'A' && path[0] <= 'Z' || path[0] >= 'a' && path[0] <= 'z') &&
		path[1] == ':' &&
		(path[2] == '\\' || path[2] == '/') {
		return true
	}
	return false
}
