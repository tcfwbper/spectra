package storage

import "path/filepath"

// Path constants for .spectra/ directory structure.
const (
	SpectraDir          = ".spectra"
	SessionsDir         = ".spectra/sessions"
	WorkflowsDir        = ".spectra/workflows"
	AgentsDir           = ".spectra/agents"
	SessionMetadataFile = "session.json"
	EventHistoryFile    = "events.jsonl"
	RuntimeSocketFile   = "runtime.sock"
)

// GetSpectraDir returns the absolute path to .spectra/.
func GetSpectraDir(projectRoot string) string {
	return filepath.Join(projectRoot, SpectraDir)
}

// GetSessionsDir returns the absolute path to .spectra/sessions/.
func GetSessionsDir(projectRoot string) string {
	return filepath.Join(projectRoot, SessionsDir)
}

// GetWorkflowsDir returns the absolute path to .spectra/workflows/.
func GetWorkflowsDir(projectRoot string) string {
	return filepath.Join(projectRoot, WorkflowsDir)
}

// GetAgentsDir returns the absolute path to .spectra/agents/.
func GetAgentsDir(projectRoot string) string {
	return filepath.Join(projectRoot, AgentsDir)
}

// GetSessionDir returns the absolute path to .spectra/sessions/<UUID>/.
func GetSessionDir(projectRoot, sessionUUID string) string {
	return filepath.Join(projectRoot, SessionsDir, sessionUUID)
}

// GetSessionMetadataPath returns the absolute path to .spectra/sessions/<UUID>/session.json.
func GetSessionMetadataPath(projectRoot, sessionUUID string) string {
	return filepath.Join(projectRoot, SessionsDir, sessionUUID, SessionMetadataFile)
}

// GetEventHistoryPath returns the absolute path to .spectra/sessions/<UUID>/events.jsonl.
func GetEventHistoryPath(projectRoot, sessionUUID string) string {
	return filepath.Join(projectRoot, SessionsDir, sessionUUID, EventHistoryFile)
}

// GetRuntimeSocketPath returns the absolute path to .spectra/sessions/<UUID>/runtime.sock.
func GetRuntimeSocketPath(projectRoot, sessionUUID string) string {
	return filepath.Join(projectRoot, SessionsDir, sessionUUID, RuntimeSocketFile)
}

// GetWorkflowPath returns the absolute path to .spectra/workflows/<WorkflowName>.yaml.
func GetWorkflowPath(projectRoot, workflowName string) string {
	return filepath.Join(projectRoot, WorkflowsDir, workflowName+".yaml")
}

// GetAgentPath returns the absolute path to .spectra/agents/<AgentRole>.yaml.
func GetAgentPath(projectRoot, agentRole string) string {
	return filepath.Join(projectRoot, AgentsDir, agentRole+".yaml")
}
