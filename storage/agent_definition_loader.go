package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// AgentDefinition represents an agent configuration loaded from YAML.
type AgentDefinition struct {
	Role            string   `yaml:"role"`
	Model           string   `yaml:"model"`
	Effort          string   `yaml:"effort"`
	SystemPrompt    string   `yaml:"system_prompt"`
	AgentRoot       string   `yaml:"agent_root"`
	AllowedTools    []string `yaml:"allowed_tools"`
	DisallowedTools []string `yaml:"disallowed_tools"`
}

// AgentDefinitionLoader loads agent definitions from .spectra/agents/.
type AgentDefinitionLoader struct {
	projectRoot string
}

// NewAgentDefinitionLoader creates a new AgentDefinitionLoader
func NewAgentDefinitionLoader(projectRoot string) *AgentDefinitionLoader {
	return &AgentDefinitionLoader{projectRoot: projectRoot}
}

// Load loads an agent definition from disk.
func (l *AgentDefinitionLoader) Load(agentRole string) (*AgentDefinition, error) {
	// Compose the file path
	agentPath := GetAgentPath(l.projectRoot, agentRole)

	// Read the file
	data, err := os.ReadFile(agentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent definition not found: %s", agentRole)
		}
		return nil, fmt.Errorf("failed to read agent definition '%s': %w", agentRole, err)
	}

	// Check for empty file before parsing
	if len(data) == 0 {
		return nil, fmt.Errorf("failed to parse agent definition '%s': EOF", agentRole)
	}

	// Parse YAML
	var def AgentDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse agent definition '%s': %w", agentRole, err)
	}

	// Validate required fields
	if def.Role == "" {
		return nil, fmt.Errorf("agent definition '%s' validation failed: missing required field 'role'", agentRole)
	}
	if def.Model == "" {
		return nil, fmt.Errorf("agent definition '%s' validation failed: missing required field 'model'", agentRole)
	}
	if def.Effort == "" {
		return nil, fmt.Errorf("agent definition '%s' validation failed: missing required field 'effort'", agentRole)
	}
	if def.SystemPrompt == "" {
		return nil, fmt.Errorf("agent definition '%s' validation failed: missing required field 'system_prompt'", agentRole)
	}
	if def.AgentRoot == "" {
		return nil, fmt.Errorf("agent definition '%s' validation failed: missing required field 'agent_root'", agentRole)
	}

	// Validate role format (PascalCase)
	pascalCasePattern := regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	if !pascalCasePattern.MatchString(def.Role) {
		return nil, fmt.Errorf("agent definition '%s' validation failed: role must be PascalCase with no spaces or special characters", agentRole)
	}

	// Validate agent_root is a relative path
	// Check for absolute paths (starts with / or drive letter like C:)
	if filepath.IsAbs(def.AgentRoot) || (len(def.AgentRoot) >= 2 && def.AgentRoot[1] == ':') {
		return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root must be a relative path", agentRole)
	}

	// Validate agent_root directory exists
	agentRootPath := filepath.Join(l.projectRoot, def.AgentRoot)
	info, err := os.Stat(agentRootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root directory not found: %s", agentRole, agentRootPath)
		}
		return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root directory not found: %s", agentRole, agentRootPath)
	}

	// Validate agent_root is a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root is not a directory: %s", agentRole, agentRootPath)
	}

	return &def, nil
}
