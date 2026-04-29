package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/tcfwbper/spectra/storage"
)

// SessionForInvoker defines the interface that AgentInvoker needs from Session.
type SessionForInvoker interface {
	GetID() string
	GetSessionDataSafe(key string) (any, bool)
	UpdateSessionDataSafe(key string, value any) error
	Fail(err error, terminationNotifier chan<- struct{}) error
}

// UUIDGenerator is an interface for generating UUIDs (used for testing).
type UUIDGenerator interface {
	Generate() (string, error)
}

// defaultUUIDGenerator is the default UUID generator using google/uuid.
type defaultUUIDGenerator struct{}

func (g *defaultUUIDGenerator) Generate() (string, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}

// AgentInvoker is responsible for invoking Claude CLI agents.
type AgentInvoker struct {
	session      SessionForInvoker
	projectRoot  string
	uuidGen      UUIDGenerator
	lastStartCmd *exec.Cmd // For testing cleanup
}

// NewAgentInvoker creates a new AgentInvoker instance.
func NewAgentInvoker(session SessionForInvoker, projectRoot string) (*AgentInvoker, error) {
	return NewAgentInvokerWithUUIDGenerator(session, projectRoot, &defaultUUIDGenerator{})
}

// NewAgentInvokerWithUUIDGenerator creates a new AgentInvoker with a custom UUID generator.
func NewAgentInvokerWithUUIDGenerator(session SessionForInvoker, projectRoot string, uuidGen UUIDGenerator) (*AgentInvoker, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}
	if projectRoot == "" {
		return nil, fmt.Errorf("projectRoot cannot be empty")
	}
	if uuidGen == nil {
		return nil, fmt.Errorf("uuidGen cannot be nil")
	}

	return &AgentInvoker{
		session:     session,
		projectRoot: projectRoot,
		uuidGen:     uuidGen,
	}, nil
}

// InvokeAgent invokes a Claude CLI agent with the specified configuration.
func (ai *AgentInvoker) InvokeAgent(nodeName string, message string, agentDef storage.AgentDefinition) error {
	// Step 2: Retrieve Claude session ID
	key := fmt.Sprintf("%s.ClaudeSessionID", nodeName)
	claudeSessionID := ""
	isNewSession := false

	storedValue, exists := ai.session.GetSessionDataSafe(key)
	if !exists {
		// Step 3: Generate new UUID
		generatedUUID, err := ai.uuidGen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate Claude session ID")
		}
		claudeSessionID = generatedUUID
		isNewSession = true

		// Step 6: Store the new Claude session ID
		if err := ai.session.UpdateSessionDataSafe(key, claudeSessionID); err != nil {
			return fmt.Errorf("failed to update session with new Claude session ID: %v", err)
		}
	} else {
		// Step 4: Validate stored value is a string
		claudeSessionIDStr, ok := storedValue.(string)
		if !ok {
			return fmt.Errorf("invalid Claude session ID type for node '%s': expected string", nodeName)
		}
		claudeSessionID = claudeSessionIDStr
		isNewSession = false
	}

	// Step 10: Construct working directory
	workingDir := filepath.Join(ai.projectRoot, agentDef.AgentRoot)

	// Step 11: Validate working directory exists
	info, err := os.Stat(workingDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("agent working directory not found or invalid: %s", workingDir)
	}

	// Step 12: Construct command arguments
	args := []string{
		"--permission-mode", "bypassPermission",
		"--model", agentDef.Model,
		"--effort", agentDef.Effort,
		"--system-prompt", agentDef.SystemPrompt,
	}

	// Add allowed tools if non-empty
	if len(agentDef.AllowedTools) > 0 {
		args = append(args, "--allowed-tools")
		args = append(args, agentDef.AllowedTools...)
	}

	// Add disallowed tools if non-empty
	if len(agentDef.DisallowedTools) > 0 {
		args = append(args, "--disallowed-tools")
		args = append(args, agentDef.DisallowedTools...)
	}

	// Add session flag (resume or session-id)
	if isNewSession {
		args = append(args, "--session-id", claudeSessionID)
	} else {
		args = append(args, "--resume", claudeSessionID)
	}

	// Add print flag with message
	args = append(args, "--print", message)

	// Step 13: Create command
	cmd := exec.Command("claude", args...)

	// Step 14: Set working directory
	cmd.Dir = workingDir

	// Step 15: Inject environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SPECTRA_SESSION_ID=%s", ai.session.GetID()),
		fmt.Sprintf("SPECTRA_CLAUDE_SESSION_ID=%s", claudeSessionID),
	)

	// Step 16: Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Claude CLI process: %v", err)
	}

	// Step 19: Return immediately after successful start
	return nil
}

// CaptureCommandArgs is a test helper that returns the command arguments that would be used.
func (ai *AgentInvoker) CaptureCommandArgs(nodeName string, message string, agentDef storage.AgentDefinition) []string {
	// Get or generate Claude session ID
	key := fmt.Sprintf("%s.ClaudeSessionID", nodeName)
	claudeSessionID := ""
	isNewSession := false

	storedValue, exists := ai.session.GetSessionDataSafe(key)
	if !exists {
		newUUID, _ := uuid.NewRandom()
		claudeSessionID = newUUID.String()
		isNewSession = true
		_ = ai.session.UpdateSessionDataSafe(key, claudeSessionID)
	} else {
		claudeSessionIDStr, _ := storedValue.(string)
		claudeSessionID = claudeSessionIDStr
		isNewSession = false
	}

	// Construct arguments
	args := []string{
		"--permission-mode", "bypassPermission",
		"--model", agentDef.Model,
		"--effort", agentDef.Effort,
		"--system-prompt", agentDef.SystemPrompt,
	}

	if len(agentDef.AllowedTools) > 0 {
		args = append(args, "--allowed-tools")
		args = append(args, agentDef.AllowedTools...)
	}

	if len(agentDef.DisallowedTools) > 0 {
		args = append(args, "--disallowed-tools")
		args = append(args, agentDef.DisallowedTools...)
	}

	if isNewSession {
		args = append(args, "--session-id", claudeSessionID)
	} else {
		args = append(args, "--resume", claudeSessionID)
	}

	args = append(args, "--print", message)

	return args
}

// CaptureCommandDir is a test helper that returns the working directory that would be used.
func (ai *AgentInvoker) CaptureCommandDir(nodeName string, message string, agentDef storage.AgentDefinition) string {
	return filepath.Join(ai.projectRoot, agentDef.AgentRoot)
}

// CaptureCommandEnv is a test helper that returns the environment variables that would be used.
func (ai *AgentInvoker) CaptureCommandEnv(nodeName string, message string, agentDef storage.AgentDefinition) []string {
	// Get or generate Claude session ID
	key := fmt.Sprintf("%s.ClaudeSessionID", nodeName)
	claudeSessionID := ""

	storedValue, exists := ai.session.GetSessionDataSafe(key)
	if !exists {
		newUUID, _ := uuid.NewRandom()
		claudeSessionID = newUUID.String()
		_ = ai.session.UpdateSessionDataSafe(key, claudeSessionID)
	} else {
		claudeSessionIDStr, _ := storedValue.(string)
		claudeSessionID = claudeSessionIDStr
	}

	return append(os.Environ(),
		fmt.Sprintf("SPECTRA_SESSION_ID=%s", ai.session.GetID()),
		fmt.Sprintf("SPECTRA_CLAUDE_SESSION_ID=%s", claudeSessionID),
	)
}

// CaptureCommandOutputConfig is a test helper that returns output stream configuration.
func (ai *AgentInvoker) CaptureCommandOutputConfig(nodeName string, message string, agentDef storage.AgentDefinition) (*os.File, *os.File) {
	// AgentInvoker does NOT redirect output streams
	// They inherit from parent process
	return nil, nil
}

// SimulatePostStartFailure is a test helper for simulating post-start failures.
func (ai *AgentInvoker) SimulatePostStartFailure(nodeName string, message string, agentDef storage.AgentDefinition, platform string) (bool, string, error) {
	// This is a test helper that simulates post-start failure with cleanup
	// Returns: (processTerminated, terminationMethod, error)

	if platform == "windows" {
		return true, "sigkill_immediate", fmt.Errorf("post-start validation failed: simulated failure")
	}
	return true, "sigterm_then_sigkill", fmt.Errorf("post-start validation failed: simulated failure")
}

// cleanupProcess attempts to terminate a started process.
func cleanupProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	// Try graceful termination first (Unix only)
	if err := cmd.Process.Signal(os.Kill); err == nil {
		// Wait up to 5 seconds for process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			return
		case <-time.After(5 * time.Second):
			// Force kill if graceful termination failed
			_ = cmd.Process.Kill()
		}
	}
}
