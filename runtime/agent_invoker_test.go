package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ensure imports are used (compile guard for scaffolded tests).
var (
	_ = assert.Equal
	_ = errors.New
)

// =============================================================================
// Test Helpers — AgentInvoker
// =============================================================================

// mockAgentDefinition provides test values for AgentDefinition getter methods
// used by AgentInvoker.Invoke. It decouples tests from components.AgentDefinition
// construction validation.
type mockAgentDefinition struct {
	model           string
	effort          string
	systemPrompt    string
	agentRoot       string
	allowedTools    []string
	disallowedTools []string
}

func (m *mockAgentDefinition) Model() string             { return m.model }
func (m *mockAgentDefinition) Effort() string            { return m.effort }
func (m *mockAgentDefinition) SystemPrompt() string      { return m.systemPrompt }
func (m *mockAgentDefinition) AgentRoot() string         { return m.agentRoot }
func (m *mockAgentDefinition) AllowedTools() []string    { return m.allowedTools }
func (m *mockAgentDefinition) DisallowedTools() []string { return m.disallowedTools }

// mockUUIDGenerator tracks UUID generation calls and returns configured values.
type mockUUIDGenerator struct {
	result string
	err    error
	called int
}

func (m *mockUUIDGenerator) Generate() (string, error) {
	m.called++
	return m.result, m.err
}

// mockCommandStarter captures exec.Command arguments and simulates cmd behavior.
type mockCommandStarter struct {
	// Captured fields
	path string
	args []string
	dir  string
	env  []string

	// Method call tracking
	startCalled  int
	runCalled    int
	outputCalled int
	waitCalled   int

	// Stdout/Stderr captures (nil means not redirected)
	stdoutSet bool
	stderrSet bool

	// Configured behavior
	startErr error
}

// newDefaultMockAgentDefinition returns a mock with standard test values.
func newDefaultMockAgentDefinition() *mockAgentDefinition {
	return &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       ".",
		allowedTools:    []string{},
		disallowedTools: []string{},
	}
}

// newDefaultMockUUIDGenerator returns a UUID generator that succeeds.
func newDefaultMockUUIDGenerator() *mockUUIDGenerator {
	return &mockUUIDGenerator{
		result: "generated-uuid",
	}
}

// newDefaultMockCommandStarter returns a command starter that succeeds.
func newDefaultMockCommandStarter() *mockCommandStarter {
	return &mockCommandStarter{}
}

// createTempProjectRoot creates a temp directory suitable as ProjectRoot.
func createTempProjectRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// createTempProjectRootWithSubdir creates a temp directory with a subdirectory.
func createTempProjectRootWithSubdir(t *testing.T, subdir string) string {
	t.Helper()
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, subdir), 0o755)
	require.NoError(t, err, "failed to create subdirectory %q", subdir)
	return dir
}

// createTempProjectRootWithFile creates a temp directory with a regular file (not a directory).
func createTempProjectRootWithFile(t *testing.T, filename string) string {
	t.Helper()
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, filename))
	require.NoError(t, err, "failed to create file %q", filename)
	f.Close()
	return dir
}

// containsArg checks if a string is present in a slice.
func containsArg(args []string, target string) bool {
	for _, a := range args {
		if a == target {
			return true
		}
	}
	return false
}

// containsEnvVar checks if env slice contains a specific KEY=VALUE entry.
func containsEnvVar(env []string, entry string) bool {
	for _, e := range env {
		if e == entry {
			return true
		}
	}
	return false
}

// argsContainSequence checks if a sequence of args appears in order.
func argsContainSequence(args []string, seq ...string) bool {
	if len(seq) == 0 {
		return true
	}
	for i := 0; i <= len(args)-len(seq); i++ {
		match := true
		for j, s := range seq {
			if args[i+j] != s {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewAgentInvoker_ValidDeps(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and constructor NewAgentInvoker do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	// Act
	_ = ps
	_ = projectRoot
	// invoker := NewAgentInvoker(ps, projectRoot)

	// Assert: returns non-nil *AgentInvoker; no panic
	// require.NotNil(t, invoker)
}

// =============================================================================
// Happy Path — Invoke
// =============================================================================

func TestAgentInvoker_Invoke_NewSession(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := newDefaultMockUUIDGenerator()
	uuidGen.result = "generated-uuid"
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("MyNode", "hello", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.True(t, containsArg(cmdStarter.args, "--session-id"))
	// assert.True(t, containsArg(cmdStarter.args, "generated-uuid"))
	// assert.False(t, containsArg(cmdStarter.args, "--resume"))
	// assert.Equal(t, 1, sess.updateSessionDataCalled)
	// assert.Equal(t, "MyNode.ClaudeSessionID", sess.updateSessionDataInputKey)
	// assert.Equal(t, "generated-uuid", sess.updateSessionDataInputVal)
}

func TestAgentInvoker_Invoke_ExistingSession(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = "existing-id"
	sess.getSessionDataResultOK = true
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("MyNode", "hello", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.True(t, containsArg(cmdStarter.args, "--resume"))
	// assert.True(t, containsArg(cmdStarter.args, "existing-id"))
	// assert.False(t, containsArg(cmdStarter.args, "--session-id"))
	// assert.Equal(t, 0, sess.updateSessionDataCalled)
}

func TestAgentInvoker_Invoke_WithAllowedTools(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       ".",
		allowedTools:    []string{"Bash(*)", "Read(*)"},
		disallowedTools: []string{},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.True(t, containsArg(cmdStarter.args, "--allowed-tools"))
	// assert.True(t, containsArg(cmdStarter.args, "Bash(*)"))
	// assert.True(t, containsArg(cmdStarter.args, "Read(*)"))
}

func TestAgentInvoker_Invoke_WithDisallowedTools(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       ".",
		allowedTools:    []string{},
		disallowedTools: []string{"Write(*)"},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.True(t, containsArg(cmdStarter.args, "--disallowed-tools"))
	// assert.True(t, containsArg(cmdStarter.args, "Write(*)"))
}

func TestAgentInvoker_Invoke_OmitsToolFlagsWhenEmpty(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.False(t, containsArg(cmdStarter.args, "--allowed-tools"))
	// assert.False(t, containsArg(cmdStarter.args, "--disallowed-tools"))
}

func TestAgentInvoker_Invoke_SubdirectoryAgentRoot(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRootWithSubdir(t, "spec/logic")

	agentDef := &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       "spec/logic",
		allowedTools:    []string{},
		disallowedTools: []string{},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.Equal(t, filepath.Join(projectRoot, "spec/logic"), cmdStarter.dir)
}

func TestAgentInvoker_Invoke_DotAgentRoot(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition() // AgentRoot = "."
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.Equal(t, projectRoot, cmdStarter.dir)
}

func TestAgentInvoker_Invoke_EnvironmentVariables(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.id = "sess-42"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "claude-uuid"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.True(t, containsEnvVar(cmdStarter.env, "SPECTRA_SESSION_ID=sess-42"))
	// assert.True(t, containsEnvVar(cmdStarter.env, "SPECTRA_CLAUDE_SESSION_ID=claude-uuid"))
}

func TestAgentInvoker_Invoke_EnvOverridesExisting(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.id = "new-sess"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "new-uuid"}
	cmdStarter := newDefaultMockCommandStarter()

	// Set parent env to include pre-existing value
	t.Setenv("SPECTRA_SESSION_ID", "old-val")

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert: both old and new values in env (last-value-wins)
	// require.NoError(t, err)
	// assert.True(t, containsEnvVar(cmdStarter.env, "SPECTRA_SESSION_ID=new-sess"))
}

func TestAgentInvoker_Invoke_MessageWithSpecialChars(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	message := "line1\n\"quoted\" $var"

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter
	_ = message

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", message, agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.True(t, containsArg(cmdStarter.args, message))
}

func TestAgentInvoker_Invoke_CommandStructure(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.id = "sess-1"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := &mockAgentDefinition{
		model:           "opus",
		effort:          "low",
		systemPrompt:    "sys prompt",
		agentRoot:       ".",
		allowedTools:    []string{},
		disallowedTools: []string{},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("TestNode", "msg text", agentDef)

	// Assert: command path is "claude", args contain flags in expected order
	// require.NoError(t, err)
	// assert.Equal(t, "claude", cmdStarter.path)
	// assert.True(t, argsContainSequence(cmdStarter.args, "--permission-mode", "bypassPermission"))
	// assert.True(t, argsContainSequence(cmdStarter.args, "--model", "opus"))
	// assert.True(t, argsContainSequence(cmdStarter.args, "--effort", "low"))
	// assert.True(t, argsContainSequence(cmdStarter.args, "--system-prompt", "sys prompt"))
	// assert.True(t, argsContainSequence(cmdStarter.args, "--session-id", "uuid-1"))
	// assert.True(t, argsContainSequence(cmdStarter.args, "--print", "msg text"))
}

// =============================================================================
// Error Propagation — Invoke
// =============================================================================

func TestAgentInvoker_Invoke_InvalidSessionIDType(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: stored session ID is not a string (integer)
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = 123
	sess.getSessionDataResultOK = true
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()

	_ = ps
	_ = projectRoot
	_ = agentDef

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot)
	// err := invoker.Invoke("BadNode", "msg", agentDef)

	// Assert
	// require.Error(t, err)
	// assert.Equal(t, "invalid Claude session ID type for node 'BadNode': expected string", err.Error())
}

func TestAgentInvoker_Invoke_UUIDGenerationFails(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: UUID generator fails
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{err: errors.New("entropy exhausted")}

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to generate Claude session ID")
	// assert.Equal(t, 0, sess.updateSessionDataCalled)
}

func TestAgentInvoker_Invoke_UpdateSessionDataFails(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: UpdateSessionDataSafe returns error
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	sess.updateSessionDataErr = errors.New("validation failed")
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to update session with new Claude session ID")
	// assert.Equal(t, 0, cmdStarter.startCalled)
}

func TestAgentInvoker_Invoke_WorkingDirNotExist(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: AgentRoot references nonexistent path
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t) // no subdirectory "nonexistent/path" created

	agentDef := &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       "nonexistent/path",
		allowedTools:    []string{},
		disallowedTools: []string{},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.Error(t, err)
	// expectedPath := filepath.Join(projectRoot, "nonexistent/path")
	// assert.Contains(t, err.Error(), "agent working directory not found or invalid: "+expectedPath)
}

func TestAgentInvoker_Invoke_WorkingDirIsFile(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: "afile" is a file, not a directory
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRootWithFile(t, "afile")

	agentDef := &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       "afile",
		allowedTools:    []string{},
		disallowedTools: []string{},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.Error(t, err)
	// expectedPath := filepath.Join(projectRoot, "afile")
	// assert.Contains(t, err.Error(), "agent working directory not found or invalid: "+expectedPath)
}

func TestAgentInvoker_Invoke_CmdStartFails(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: Start() returns error
	sess := newDefaultMockSession()
	sess.id = "sess-1"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := &mockCommandStarter{startErr: errors.New("exec: not found")}

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to start Claude CLI process")
}

// =============================================================================
// Mock / Dependency Interaction — Invoke
// =============================================================================

func TestAgentInvoker_Invoke_NoUpdateWhenExistingSession(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = "existing-id"
	sess.getSessionDataResultOK = true
	sess.id = "sess-1"
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("ResumeNode", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.Equal(t, 0, sess.updateSessionDataCalled)
}

func TestAgentInvoker_Invoke_DoesNotCallCmdRunOrOutput(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.id = "sess-1"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.Equal(t, 1, cmdStarter.startCalled)
	// assert.Equal(t, 0, cmdStarter.runCalled)
	// assert.Equal(t, 0, cmdStarter.outputCalled)
	// assert.Equal(t, 0, cmdStarter.waitCalled)
}

func TestAgentInvoker_Invoke_UUIDGeneratorCalledOnce(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.id = "sess-1"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.Equal(t, 1, uuidGen.called)
}

func TestAgentInvoker_Invoke_NoStdoutStderrRedirect(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup
	sess := newDefaultMockSession()
	sess.id = "sess-1"
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert
	// require.NoError(t, err)
	// assert.False(t, cmdStarter.stdoutSet)
	// assert.False(t, cmdStarter.stderrSet)
}

// =============================================================================
// State Transitions — Invoke
// =============================================================================

func TestAgentInvoker_Invoke_FailFast_AfterUUIDError(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: UUID generation fails
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{err: errors.New("entropy exhausted")}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert: error returned; no session update; no command created
	// require.Error(t, err)
	// assert.Equal(t, 0, sess.updateSessionDataCalled)
	// assert.Equal(t, 0, cmdStarter.startCalled)
}

func TestAgentInvoker_Invoke_FailFast_AfterUpdateError(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: UpdateSessionDataSafe fails
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	sess.updateSessionDataErr = errors.New("update failed")
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t)

	agentDef := newDefaultMockAgentDefinition()
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert: error returned; no command created
	// require.Error(t, err)
	// assert.Equal(t, 0, cmdStarter.startCalled)
}

func TestAgentInvoker_Invoke_FailFast_AfterWorkDirError(t *testing.T) {
	t.Skip("scaffolded: production type AgentInvoker and Invoke method do not exist yet")

	// Setup: working directory doesn't exist
	sess := newDefaultMockSession()
	sess.getSessionDataResultVal = nil
	sess.getSessionDataResultOK = false
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	ps := NewPersistentSession(sess, metaStore, evStore, log)
	projectRoot := createTempProjectRoot(t) // no "missing" subdirectory

	agentDef := &mockAgentDefinition{
		model:           "sonnet",
		effort:          "high",
		systemPrompt:    "prompt",
		agentRoot:       "missing",
		allowedTools:    []string{},
		disallowedTools: []string{},
	}
	uuidGen := &mockUUIDGenerator{result: "uuid-1"}
	cmdStarter := newDefaultMockCommandStarter()

	_ = ps
	_ = projectRoot
	_ = agentDef
	_ = uuidGen
	_ = cmdStarter

	// Act
	// invoker := NewAgentInvoker(ps, projectRoot, WithUUIDGenerator(uuidGen), WithCommandStarter(cmdStarter))
	// err := invoker.Invoke("Node1", "msg", agentDef)

	// Assert: error returned; no command started
	// require.Error(t, err)
	// assert.Equal(t, 0, cmdStarter.startCalled)
}
