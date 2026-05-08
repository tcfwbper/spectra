package runtime

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// AgentDef defines the read-only interface for agent definitions consumed by
// AgentInvoker.Invoke. It provides access to agent configuration fields.
type AgentDef interface {
	Model() string
	Effort() string
	SystemPrompt() string
	AgentRoot() string
	AllowedTools() []string
	DisallowedTools() []string
}

// UUIDGenerator defines the interface for UUID generation.
type UUIDGenerator interface {
	Generate() (string, error)
}

// CommandHandle abstracts an exec.Cmd for testing. It exposes the minimal
// surface needed by AgentInvoker: Start, and field setters.
type CommandHandle interface {
	SetDir(dir string)
	SetEnv(env []string)
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)
	Start() error
}

// CommandStarter abstracts the creation of a command (exec.Command).
type CommandStarter interface {
	Command(name string, args ...string) CommandHandle
}

// --- Default implementations ---

// defaultUUIDGenerator generates UUID v4 strings using crypto/rand.
type defaultUUIDGenerator struct{}

func (d *defaultUUIDGenerator) Generate() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Set version (4) and variant (RFC 4122).
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// execCommandHandle wraps exec.Cmd to implement CommandHandle.
type execCommandHandle struct {
	cmd *exec.Cmd
}

func (h *execCommandHandle) SetDir(dir string) {
	h.cmd.Dir = dir
}

func (h *execCommandHandle) SetEnv(env []string) {
	h.cmd.Env = env
}

func (h *execCommandHandle) SetStdout(w io.Writer) {
	h.cmd.Stdout = w
}

func (h *execCommandHandle) SetStderr(w io.Writer) {
	h.cmd.Stderr = w
}

func (h *execCommandHandle) Start() error {
	return h.cmd.Start()
}

// defaultCommandStarter creates real exec.Cmd instances.
type defaultCommandStarter struct{}

func (d *defaultCommandStarter) Command(name string, args ...string) CommandHandle {
	return &execCommandHandle{cmd: exec.Command(name, args...)}
}

// --- AgentInvoker ---

// AgentInvoker is responsible for invoking a Claude CLI agent process with the
// appropriate configuration, environment variables, and working directory.
type AgentInvoker struct {
	ps          *PersistentSession
	projectRoot string
	uuidGen     UUIDGenerator
	cmdStarter  CommandStarter
}

// AgentInvokerOption is a functional option for configuring AgentInvoker.
type AgentInvokerOption func(*AgentInvoker)

// WithUUIDGenerator sets a custom UUID generator for the AgentInvoker.
func WithUUIDGenerator(gen UUIDGenerator) AgentInvokerOption {
	return func(ai *AgentInvoker) {
		ai.uuidGen = gen
	}
}

// WithCommandStarter sets a custom command starter for the AgentInvoker.
func WithCommandStarter(starter CommandStarter) AgentInvokerOption {
	return func(ai *AgentInvoker) {
		ai.cmdStarter = starter
	}
}

// NewAgentInvoker constructs an AgentInvoker with the given PersistentSession
// and project root path, and applies functional options.
func NewAgentInvoker(ps *PersistentSession, projectRoot string, opts ...AgentInvokerOption) *AgentInvoker {
	ai := &AgentInvoker{
		ps:          ps,
		projectRoot: projectRoot,
		uuidGen:     &defaultUUIDGenerator{},
		cmdStarter:  &defaultCommandStarter{},
	}
	for _, opt := range opts {
		opt(ai)
	}
	return ai
}

// Invoke starts a Claude CLI process for the given node with the specified
// message and agent definition. It manages the Claude session ID lifecycle
// (reading from or generating into session data), constructs the command with
// flags, and starts the process asynchronously.
func (ai *AgentInvoker) Invoke(nodeName, message string, agentDef AgentDef) error {
	// Step 1: Resolve Claude session ID.
	claudeSessionID, isExisting, err := ai.resolveClaudeSessionID(nodeName)
	if err != nil {
		return err
	}

	// Step 2: If new session, persist the Claude session ID.
	if !isExisting {
		key := nodeName + ".ClaudeSessionID"
		if err := ai.ps.UpdateSessionDataSafe(key, claudeSessionID); err != nil {
			return fmt.Errorf("failed to update session with new Claude session ID: %w", err)
		}
	}

	// Step 3: Validate working directory.
	workDir := filepath.Join(ai.projectRoot, agentDef.AgentRoot())
	info, statErr := os.Stat(workDir)
	if statErr != nil || !info.IsDir() {
		return fmt.Errorf("agent working directory not found or invalid: %s", workDir)
	}

	// Step 4: Build command arguments.
	args := ai.buildArgs(agentDef, claudeSessionID, isExisting, message)

	// Step 5: Create and configure command.
	cmd := ai.cmdStarter.Command("claude", args...)
	cmd.SetDir(workDir)
	cmd.SetEnv(ai.buildEnv(claudeSessionID))

	// Step 6: Start the process.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Claude CLI process: %w", err)
	}

	return nil
}

// resolveClaudeSessionID reads or generates the Claude session ID for a node.
func (ai *AgentInvoker) resolveClaudeSessionID(nodeName string) (string, bool, error) {
	key := nodeName + ".ClaudeSessionID"
	stored, ok := ai.ps.GetSessionDataSafe(key)
	if ok {
		// Key exists: validate that it's a string.
		strVal, isStr := stored.(string)
		if !isStr {
			return "", false, fmt.Errorf("invalid Claude session ID type for node '%s': expected string", nodeName)
		}
		return strVal, true, nil
	}

	// Key does not exist: generate a new UUID.
	id, err := ai.uuidGen.Generate()
	if err != nil {
		return "", false, fmt.Errorf("failed to generate Claude session ID")
	}
	return id, false, nil
}

// buildArgs constructs the CLI arguments for the claude command.
func (ai *AgentInvoker) buildArgs(agentDef AgentDef, claudeSessionID string, isExisting bool, message string) []string {
	args := []string{
		"--permission-mode", "bypassPermissions",
		"--model", agentDef.Model(),
		"--effort", agentDef.Effort(),
		"--system-prompt", agentDef.SystemPrompt(),
	}

	// Conditional: allowed tools.
	if tools := agentDef.AllowedTools(); len(tools) > 0 {
		args = append(args, "--allowed-tools")
		args = append(args, tools...)
	}

	// Conditional: disallowed tools.
	if tools := agentDef.DisallowedTools(); len(tools) > 0 {
		args = append(args, "--disallowed-tools")
		args = append(args, tools...)
	}

	// Session ID or resume.
	if isExisting {
		args = append(args, "--resume", claudeSessionID)
	} else {
		args = append(args, "--session-id", claudeSessionID)
	}

	// Message.
	args = append(args, "--print", message)

	return args
}

// buildEnv constructs the environment variables for the child process.
func (ai *AgentInvoker) buildEnv(claudeSessionID string) []string {
	env := os.Environ()
	env = append(env, "SPECTRA_SESSION_ID="+ai.ps.ID)
	env = append(env, "SPECTRA_CLAUDE_SESSION_ID="+claudeSessionID)
	return env
}
