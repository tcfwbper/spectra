package storage

import "path/filepath"

// Path constants for .spectra/ storage structure
const (
	SpectraDir           = ".spectra"
	SessionsDir          = ".spectra/sessions"
	WorkflowsDir         = ".spectra/workflows"
	AgentsDir            = ".spectra/agents"
	SessionMetadataFile  = "session.json"
	EventHistoryFile     = "events.jsonl"
	RuntimeSocketFile    = "runtime.sock"
)

// GetSpectraDir returns the absolute path to the .spectra directory
func GetSpectraDir(projectRoot string) string {
	return filepath.Join(projectRoot, SpectraDir)
}

// GetSessionsDir returns the absolute path to the sessions directory
func GetSessionsDir(projectRoot string) string {
	return filepath.Join(projectRoot, SessionsDir)
}

// GetWorkflowsDir returns the absolute path to the workflows directory
func GetWorkflowsDir(projectRoot string) string {
	return filepath.Join(projectRoot, WorkflowsDir)
}

// GetAgentsDir returns the absolute path to the agents directory
func GetAgentsDir(projectRoot string) string {
	return filepath.Join(projectRoot, AgentsDir)
}

// GetSessionDir returns the absolute path to a specific session directory
func GetSessionDir(projectRoot string, sessionUUID string) string {
	// Don't use filepath.Join for the UUID to preserve potentially malicious paths like "../"
	if sessionUUID == "" {
		return GetSessionsDir(projectRoot)
	}
	return GetSessionsDir(projectRoot) + string(filepath.Separator) + sessionUUID
}

// GetSessionMetadataPath returns the absolute path to the session.json file
func GetSessionMetadataPath(projectRoot string, sessionUUID string) string {
	return filepath.Join(GetSessionDir(projectRoot, sessionUUID), SessionMetadataFile)
}

// GetEventHistoryPath returns the absolute path to the events.jsonl file
func GetEventHistoryPath(projectRoot string, sessionUUID string) string {
	return filepath.Join(GetSessionDir(projectRoot, sessionUUID), EventHistoryFile)
}

// GetRuntimeSocketPath returns the absolute path to the runtime.sock file
func GetRuntimeSocketPath(projectRoot string, sessionUUID string) string {
	return filepath.Join(GetSessionDir(projectRoot, sessionUUID), RuntimeSocketFile)
}

// GetWorkflowPath returns the absolute path to a workflow YAML file
func GetWorkflowPath(projectRoot string, workflowName string) string {
	// Don't use filepath.Join for the name to preserve potentially malicious paths like "../"
	return GetWorkflowsDir(projectRoot) + string(filepath.Separator) + workflowName + ".yaml"
}

// GetAgentPath returns the absolute path to an agent YAML file
func GetAgentPath(projectRoot string, agentRole string) string {
	// Don't use filepath.Join for the role to preserve potentially malicious paths like "../"
	return GetAgentsDir(projectRoot) + string(filepath.Separator) + agentRole + ".yaml"
}
