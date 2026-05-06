package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/tcfwbper/spectra/components"
)

// agentYAML is the internal representation for strict YAML parsing of agent
// definition files. Fields use camelCase yaml tags matching the YAML schema.
type agentYAML struct {
	Model           string   `yaml:"model"`
	Effort          string   `yaml:"effort"`
	SystemPrompt    string   `yaml:"systemPrompt"`
	AgentRoot       string   `yaml:"agentRoot"`
	AllowedTools    []string `yaml:"allowedTools"`
	DisallowedTools []string `yaml:"disallowedTools"`
}

// AgentDefinitionLoader provides read-only access to agent definition YAML
// files stored in .spectra/agents/. It is stateless, does not cache, and is
// safe for concurrent use.
type AgentDefinitionLoader struct {
	projectRoot string
}

// NewAgentDefinitionLoader creates an AgentDefinitionLoader for the given
// project root directory.
func NewAgentDefinitionLoader(projectRoot string) *AgentDefinitionLoader {
	return &AgentDefinitionLoader{projectRoot: projectRoot}
}

// Load reads, parses, and validates an agent definition YAML file. The role is
// derived from the agentRole parameter (filename without .yaml extension).
func (l *AgentDefinitionLoader) Load(agentRole string) (*components.AgentDefinition, error) {
	// Compose file path via StorageLayout.
	filePath := GetAgentPath(l.projectRoot, agentRole)

	// Read file.
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("agent definition not found: %s", agentRole)
		}
		return nil, fmt.Errorf("failed to read agent definition '%s': %v", agentRole, err)
	}

	// Parse YAML with strict mode (unknown fields rejected).
	var raw agentYAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse agent definition '%s': %v", agentRole, err)
	}

	// Construct AgentDefinition via constructor (role derived from filename).
	def, err := components.NewAgentDefinition(
		agentRole,
		raw.Model,
		raw.Effort,
		raw.SystemPrompt,
		raw.AgentRoot,
		raw.AllowedTools,
		raw.DisallowedTools,
	)
	if err != nil {
		return nil, fmt.Errorf("agent definition '%s' validation failed: %v", agentRole, err)
	}

	// Validate AgentRoot directory existence.
	agentRootAbs := filepath.Join(l.projectRoot, def.AgentRoot())
	info, err := os.Stat(agentRootAbs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root directory not found: %s", agentRole, agentRootAbs)
		}
		return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root directory not found: %s", agentRole, agentRootAbs)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("agent definition '%s' validation failed: agent_root is not a directory: %s", agentRole, agentRootAbs)
	}

	return def, nil
}
