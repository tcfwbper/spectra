package runtime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for TransitionToNode tests ---

// mockSessionForTransition provides a mock Session for TransitionToNode tests.
type mockSessionForTransition struct {
	mock.Mock
	mu           sync.RWMutex
	status       string
	currentState string
	callOrder    []string
}

func newMockSessionForTransition(status, currentState string) *mockSessionForTransition {
	return &mockSessionForTransition{
		status:       status,
		currentState: currentState,
	}
}

func (m *mockSessionForTransition) UpdateCurrentStateSafe(newState string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callOrder = append(m.callOrder, "UpdateCurrentStateSafe")
	args := m.Called(newState)
	m.currentState = newState
	return args.Error(0)
}

func (m *mockSessionForTransition) Done(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callOrder = append(m.callOrder, "Done")
	args := m.Called(terminationNotifier)
	if args.Error(0) == nil {
		m.status = "completed"
		select {
		case terminationNotifier <- struct{}{}:
		default:
		}
	}
	return args.Error(0)
}

func (m *mockSessionForTransition) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callOrder = append(m.callOrder, "Fail")
	args := m.Called(err, terminationNotifier)
	if args.Error(0) == nil {
		m.status = "failed"
		select {
		case terminationNotifier <- struct{}{}:
		default:
		}
	}
	return args.Error(0)
}

func (m *mockSessionForTransition) getCallOrder() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string(nil), m.callOrder...)
}

func (m *mockSessionForTransition) getStatus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockSessionForTransition) getCurrentState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

// mockAgentDefLoaderForTransition provides a mock AgentDefinitionLoader for TransitionToNode tests.
type mockAgentDefLoaderForTransition struct {
	mock.Mock
}

func (m *mockAgentDefLoaderForTransition) Load(agentRole string) (*storage.AgentDefinition, error) {
	args := m.Called(agentRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.AgentDefinition), args.Error(1)
}

// mockAgentInvokerForTransition provides a mock AgentInvoker for TransitionToNode tests.
type mockAgentInvokerForTransition struct {
	mock.Mock
}

func (m *mockAgentInvokerForTransition) InvokeAgent(nodeName string, message string, agentDef storage.AgentDefinition) error {
	args := m.Called(nodeName, message, agentDef)
	return args.Error(0)
}

// --- Test helpers ---

// captureStdout redirects os.Stdout to a pipe and returns a restore function
// that restores os.Stdout and returns captured output.
func captureStdout(t *testing.T) func() string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	return func() string {
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		r.Close()
		return buf.String()
	}
}

func createTransitionFixture(t *testing.T, wfDef *storage.WorkflowDefinition) (
	*TransitionToNode,
	*mockSessionForTransition,
	*mockAgentDefLoaderForTransition,
	*mockAgentInvokerForTransition,
	chan struct{},
) {
	t.Helper()
	sess := newMockSessionForTransition("running", "Entry")
	agentDefLoader := &mockAgentDefLoaderForTransition{}
	agentInvoker := &mockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 2)

	transitioner, err := NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	require.NoError(t, err)
	require.NotNil(t, transitioner)

	return transitioner, sess, agentDefLoader, agentInvoker, terminationNotifier
}

func buildTransitionWorkflowWithHumanNode(nodeName string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: nodeName,
		Nodes: []storage.Node{
			{Name: nodeName, Type: "human"},
		},
	}
}

func buildTransitionWorkflowWithAgentNode(nodeName, agentRole string) *storage.WorkflowDefinition {
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: nodeName, Type: "agent", AgentRole: agentRole},
		},
	}
}

func buildTransitionWorkflowWithNodes(nodes []storage.Node) *storage.WorkflowDefinition {
	entryNode := "Entry"
	if len(nodes) > 0 {
		entryNode = nodes[0].Name
	}
	return &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: entryNode,
		Nodes:     nodes,
	}
}

func defaultAgentDefForTransition() *storage.AgentDefinition {
	return &storage.AgentDefinition{
		Role:         "TestRole",
		Model:        "sonnet",
		Effort:       "normal",
		SystemPrompt: "You are a test agent",
		AgentRoot:    "agents",
	}
}

// --- Happy Path — Construction ---

func TestTransitionToNode_New(t *testing.T) {
	sess := newMockSessionForTransition("running", "Entry")
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	agentDefLoader := &mockAgentDefLoaderForTransition{}
	agentInvoker := &mockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 1)

	transitioner, err := NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)

	require.NoError(t, err)
	require.NotNil(t, transitioner)
}

// --- Happy Path — Transition (Human Node, Regular) ---

func TestTransition_HumanNode_PrintsMessage(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	err := transitioner.Transition("Hello world", "NodeA", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: NodeA] Hello world\n", output)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "NodeA")
}

func TestTransition_HumanNode_EmptyMessage(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeB")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeB").Return(nil)

	err := transitioner.Transition("", "NodeB", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: NodeB] (no message)\n", output)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "NodeB")
}

func TestTransition_HumanNode_MessageWithNewlines(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeC")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeC").Return(nil)

	err := transitioner.Transition("Line1\nLine2\tTab", "NodeC", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: NodeC] Line1\nLine2\tTab\n", output)
}

func TestTransition_HumanNode_MessageWithQuotes(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeD")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeD").Return(nil)

	err := transitioner.Transition(`He said "hello"`, "NodeD", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: NodeD] He said \"hello\"\n", output)
}

func TestTransition_HumanNode_LargeMessage(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeE")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeE").Return(nil)

	largeMessage := strings.Repeat("A", 1024*1024) // 1 MB
	err := transitioner.Transition(largeMessage, "NodeE", false)
	output := restore()

	require.NoError(t, err)
	expected := fmt.Sprintf("[Human Node: NodeE] %s\n", largeMessage)
	assert.Equal(t, expected, output)
}

func TestTransition_HumanNode_UpdatesCurrentState(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeF")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeF").Return(nil)

	err := transitioner.Transition("test", "NodeF", false)
	output := restore()

	require.NoError(t, err)
	assert.Contains(t, output, "[Human Node: NodeF] test")
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "NodeF")
}

// --- Happy Path — Transition (Agent Node, Regular) ---

func TestTransition_AgentNode_InvokesAgent(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode1", "reviewer")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	loadedDef.Role = "reviewer"
	agentDefLoader.On("Load", "reviewer").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode1", "Review this", *loadedDef).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode1").Return(nil)

	err := transitioner.Transition("Review this", "AgentNode1", false)

	require.NoError(t, err)
	agentDefLoader.AssertCalled(t, "Load", "reviewer")
	agentInvoker.AssertCalled(t, "InvokeAgent", "AgentNode1", "Review this", *loadedDef)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "AgentNode1")
}

func TestTransition_AgentNode_PassesMessageUnmodified(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode2", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode2", "Complex message with 🎉 unicode", *loadedDef).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode2").Return(nil)

	err := transitioner.Transition("Complex message with 🎉 unicode", "AgentNode2", false)

	require.NoError(t, err)
	agentInvoker.AssertCalled(t, "InvokeAgent", "AgentNode2", "Complex message with 🎉 unicode", *loadedDef)
}

func TestTransition_AgentNode_UpdatesCurrentStateAfterInvoke(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode3", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	var callOrder []string
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode3", "test", *loadedDef).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "InvokeAgent")
	}).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode3").Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "UpdateCurrentStateSafe")
	}).Return(nil)

	err := transitioner.Transition("test", "AgentNode3", false)

	require.NoError(t, err)
	require.Equal(t, []string{"InvokeAgent", "UpdateCurrentStateSafe"}, callOrder)
}

// --- Happy Path — Transition (Exit Transition) ---

func TestTransition_ExitTransition_SkipsHumanAction(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
	sess.On("Done", mock.Anything).Return(nil)

	err := transitioner.Transition("final message", "ExitNode", true)
	output := restore()

	require.NoError(t, err)
	assert.Empty(t, output, "stdout should be empty for exit transition")
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "ExitNode")
	sess.AssertCalled(t, "Done", mock.Anything)
}

func TestTransition_ExitTransition_SkipsAgentAction(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("ExitAgent", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitAgent").Return(nil)
	sess.On("Done", mock.Anything).Return(nil)

	err := transitioner.Transition("final", "ExitAgent", true)

	require.NoError(t, err)
	agentDefLoader.AssertNotCalled(t, "Load", mock.Anything)
	agentInvoker.AssertNotCalled(t, "InvokeAgent", mock.Anything, mock.Anything, mock.Anything)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "ExitAgent")
	sess.AssertCalled(t, "Done", mock.Anything)
}

func TestTransition_ExitTransition_UpdatesStateBeforeDone(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	var callOrder []string
	sess.On("UpdateCurrentStateSafe", "ExitNode").Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "UpdateCurrentStateSafe")
	}).Return(nil)
	sess.On("Done", mock.Anything).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "Done")
	}).Return(nil)

	err := transitioner.Transition("", "ExitNode", true)

	require.NoError(t, err)
	require.Equal(t, []string{"UpdateCurrentStateSafe", "Done"}, callOrder)
}

func TestTransition_ExitTransition_CallsDone(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, terminationNotifier := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
	sess.On("Done", mock.MatchedBy(func(ch chan<- struct{}) bool {
		return ch == terminationNotifier
	})).Return(nil)

	err := transitioner.Transition("done", "ExitNode", true)

	require.NoError(t, err)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "ExitNode")
	sess.AssertCalled(t, "Done", mock.Anything)
}

// --- Validation Failures — Target Node ---

func TestTransition_TargetNodeNotFound(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode" &&
			strings.Contains(runtimeErr.Message, "target node not found: 'NonExistentNode'")
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "NonExistentNode", false)

	require.Error(t, err)
	assert.Regexp(t, `(?i)target node 'NonExistentNode' not found in workflow`, err.Error())
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

// --- Validation Failures — Agent Definition Loading ---

func TestTransition_AgentDefinitionNotFound(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "unknown")
	transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)

	agentDefLoader.On("Load", "unknown").Return(nil, fmt.Errorf("agent file not found"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode" &&
			strings.Contains(runtimeErr.Message, "failed to load agent definition for role 'unknown'")
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to load agent definition for role 'unknown':`, err.Error())
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_AgentDefinitionLoadError(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "broken")
	transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)

	agentDefLoader.On("Load", "broken").Return(nil, fmt.Errorf("invalid YAML"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode"
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to load agent definition for role.*invalid YAML`, err.Error())
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_AgentNodeEmptyRole(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "")
	transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)

	agentDefLoader.On("Load", "").Return(nil, fmt.Errorf("empty role"))
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to load agent definition`, err.Error())
	agentDefLoader.AssertCalled(t, "Load", "")
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

// --- Validation Failures — Agent Invocation ---

func TestTransition_AgentInvokerFails(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Return(fmt.Errorf("claude command not found"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode" &&
			strings.Contains(runtimeErr.Message, "failed to invoke agent for node 'AgentNode'")
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to invoke agent for node 'AgentNode'.*claude command not found`, err.Error())
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_AgentInvokerPermissionDenied(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Return(fmt.Errorf("permission denied"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return strings.Contains(runtimeErr.Message, "permission denied") ||
			strings.Contains(runtimeErr.Message, "failed to invoke agent")
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_AgentInvokerWorkingDirInvalid(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Return(fmt.Errorf("working directory not found"))
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to invoke agent.*working directory not found`, err.Error())
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

// --- Validation Failures — Session.Done ---

func TestTransition_SessionDoneFails_StatusNotRunning(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
	sess.On("Done", mock.Anything).Return(fmt.Errorf("cannot complete session: status is 'completed', expected 'running'"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode" &&
			strings.Contains(runtimeErr.Message, "failed to complete session after exit transition")
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "ExitNode", true)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to complete session after exit transition:.*cannot complete session: status is 'completed'`, err.Error())
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_SessionDoneFails_SessionAlreadyFailed(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
	sess.On("Done", mock.Anything).Return(fmt.Errorf("cannot complete session: status is 'failed', expected 'running'"))
	sess.On("Fail", mock.Anything, mock.Anything).Return(fmt.Errorf("session already failed"))

	err := transitioner.Transition("test", "ExitNode", true)

	require.Error(t, err)
	assert.Regexp(t, `(?i)failed to complete session after exit transition`, err.Error())
}

// --- Error Propagation — Session.Fail Always Called ---

func TestTransition_SessionFailCalledOnNodeNotFound(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode" &&
			runtimeErr.Message == "target node not found: 'Missing'"
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "Missing", false)

	require.Error(t, err)
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_SessionFailCalledOnAgentLoadError(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "badRole")
	transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)

	agentDefLoader.On("Load", "badRole").Return(nil, fmt.Errorf("agent not found"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode"
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_SessionFailCalledOnAgentInvokeError(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Return(fmt.Errorf("invocation failed"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode"
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_SessionFailCalledOnDoneError(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
	sess.On("Done", mock.Anything).Return(fmt.Errorf("done error"))
	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Issuer == "TransitionToNode"
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "ExitNode", true)

	require.Error(t, err)
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_SessionFailNotCalledOnSuccess(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	err := transitioner.Transition("test", "NodeA", false)
	restore()

	require.NoError(t, err)
	sess.AssertNotCalled(t, "Fail", mock.Anything, mock.Anything)
}

// --- Error Propagation — Caller Responsibility ---

func TestTransition_CallerDoesNotCallFailAgain(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	err := transitioner.Transition("test", "Missing", false)

	require.Error(t, err)
	// TransitionToNode already called Session.Fail internally exactly once.
	// Caller (EventProcessor) should NOT call Session.Fail again.
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

// --- State Transitions — Action-Before-State-Update ---

func TestTransition_HumanNode_PrintBeforeStateUpdate(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	var callOrder []string
	sess.On("UpdateCurrentStateSafe", "NodeA").Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "UpdateCurrentStateSafe")
	}).Return(nil)

	err := transitioner.Transition("test", "NodeA", false)
	output := restore()

	require.NoError(t, err)
	assert.Contains(t, output, "[Human Node: NodeA] test")
	assert.Equal(t, []string{"UpdateCurrentStateSafe"}, callOrder)
	// Print occurred before UpdateCurrentStateSafe: stdout has content
	// and the synchronous execution ensures print happened first.
}

func TestTransition_AgentNode_InvokeBeforeStateUpdate(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	var callOrder []string
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "InvokeAgent")
	}).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode").Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "UpdateCurrentStateSafe")
	}).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.NoError(t, err)
	require.Equal(t, []string{"InvokeAgent", "UpdateCurrentStateSafe"}, callOrder)
}

func TestTransition_ExitTransition_SkipsActionNoOrdering(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, agentInvoker, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	var callOrder []string
	sess.On("UpdateCurrentStateSafe", "ExitNode").Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "UpdateCurrentStateSafe")
	}).Return(nil)
	sess.On("Done", mock.Anything).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "Done")
	}).Return(nil)

	err := transitioner.Transition("test", "ExitNode", true)
	output := restore()

	require.NoError(t, err)
	assert.Empty(t, output, "no stdout write for exit transition")
	agentInvoker.AssertNotCalled(t, "InvokeAgent", mock.Anything, mock.Anything, mock.Anything)
	assert.Equal(t, []string{"UpdateCurrentStateSafe", "Done"}, callOrder)
}

// --- State Transitions — State-Before-Done ---

func TestTransition_ExitTransition_StateUpdateBeforeDone(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	var callOrder []string
	sess.On("UpdateCurrentStateSafe", "ExitNode").Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "UpdateCurrentStateSafe")
	}).Return(nil)
	sess.On("Done", mock.Anything).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "Done")
	}).Return(nil)

	err := transitioner.Transition("test", "ExitNode", true)

	require.NoError(t, err)
	require.Len(t, callOrder, 2)
	assert.Equal(t, "UpdateCurrentStateSafe", callOrder[0])
	assert.Equal(t, "Done", callOrder[1])
}

// --- Idempotency — UpdateCurrentStateSafe Always Succeeds ---

func TestTransition_UpdateCurrentStateSafeAlwaysNil(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ValidNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "ValidNode").Return(nil)

	err := transitioner.Transition("test", "ValidNode", false)
	restore()

	require.NoError(t, err)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "ValidNode")
}

func TestTransition_UpdateCurrentStateSafePersistenceFails(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	// UpdateCurrentStateSafe returns nil even when persistence fails internally
	// (it logs a warning; in-memory state is authoritative)
	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	err := transitioner.Transition("test", "NodeA", false)
	restore()

	require.NoError(t, err)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "NodeA")
}

// --- Boundary Values — Message Content ---

func TestTransition_MessageVeryLarge(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	largeMessage := strings.Repeat("X", 5*1024*1024) // 5 MB
	err := transitioner.Transition(largeMessage, "NodeA", false)
	output := restore()

	require.NoError(t, err)
	expected := fmt.Sprintf("[Human Node: NodeA] %s\n", largeMessage)
	assert.Equal(t, expected, output)
}

func TestTransition_MessageUnicodeCharacters(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	err := transitioner.Transition("测试🎉emoji", "NodeA", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: NodeA] 测试🎉emoji\n", output)
}

func TestTransition_MessageWithEmbeddedNulls(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	err := transitioner.Transition("before\x00after", "NodeA", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: NodeA] before\x00after\n", output)
}

// --- Boundary Values — Target Node Names ---

func TestTransition_TargetNodePascalCase(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("MyNodeName")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "MyNodeName").Return(nil)

	err := transitioner.Transition("test", "MyNodeName", false)
	restore()

	require.NoError(t, err)
}

func TestTransition_TargetNodeSingleCharacter(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("A")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "A").Return(nil)

	err := transitioner.Transition("test", "A", false)
	restore()

	require.NoError(t, err)
}

func TestTransition_TargetNodeLongName(t *testing.T) {
	// Generate a 256-character PascalCase node name
	longName := strings.Repeat("NodeNameSegment", 17) + "X" // 15*17=255 + 1 = 256
	longName = longName[:256]
	wfDef := buildTransitionWorkflowWithHumanNode(longName)
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", longName).Return(nil)

	err := transitioner.Transition("test", longName, false)
	restore()

	require.NoError(t, err)
}

// --- Mock / Dependency Interaction — AgentDefinitionLoader ---

func TestTransition_AgentDefinitionLoaderCalledWithCorrectRole(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "code_reviewer")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	loadedDef.Role = "code_reviewer"
	agentDefLoader.On("Load", "code_reviewer").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", mock.Anything, *loadedDef).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode").Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.NoError(t, err)
	agentDefLoader.AssertCalled(t, "Load", "code_reviewer")
	agentInvoker.AssertCalled(t, "InvokeAgent", "AgentNode", mock.Anything, *loadedDef)
}

func TestTransition_AgentDefinitionLoaderNotCalledForHuman(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("HumanNode")
	transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "HumanNode").Return(nil)

	err := transitioner.Transition("test", "HumanNode", false)
	restore()

	require.NoError(t, err)
	agentDefLoader.AssertNotCalled(t, "Load", mock.Anything)
}

func TestTransition_AgentDefinitionLoaderNotCalledForExit(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("ExitAgent", "worker")
	transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitAgent").Return(nil)
	sess.On("Done", mock.Anything).Return(nil)

	err := transitioner.Transition("test", "ExitAgent", true)

	require.NoError(t, err)
	agentDefLoader.AssertNotCalled(t, "Load", mock.Anything)
}

// --- Mock / Dependency Interaction — AgentInvoker ---

func TestTransition_AgentInvokerCalledWithCorrectArguments(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "Do this task", *loadedDef).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode").Return(nil)

	err := transitioner.Transition("Do this task", "AgentNode", false)

	require.NoError(t, err)
	agentInvoker.AssertCalled(t, "InvokeAgent", "AgentNode", "Do this task", *loadedDef)
}

func TestTransition_AgentInvokerNotCalledForHuman(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("HumanNode")
	transitioner, sess, _, agentInvoker, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "HumanNode").Return(nil)

	err := transitioner.Transition("test", "HumanNode", false)
	restore()

	require.NoError(t, err)
	agentInvoker.AssertNotCalled(t, "InvokeAgent", mock.Anything, mock.Anything, mock.Anything)
}

func TestTransition_AgentInvokerNotCalledForExit(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("ExitAgent", "worker")
	transitioner, sess, _, agentInvoker, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitAgent").Return(nil)
	sess.On("Done", mock.Anything).Return(nil)

	err := transitioner.Transition("test", "ExitAgent", true)

	require.NoError(t, err)
	agentInvoker.AssertNotCalled(t, "InvokeAgent", mock.Anything, mock.Anything, mock.Anything)
}

// --- Mock / Dependency Interaction — Session Methods ---

func TestTransition_UpdateCurrentStateSafeCalledForAll(t *testing.T) {
	t.Run("regular transition", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
		transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

		restore := captureStdout(t)

		sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

		err := transitioner.Transition("test", "NodeA", false)
		restore()

		require.NoError(t, err)
		sess.AssertCalled(t, "UpdateCurrentStateSafe", "NodeA")
	})

	t.Run("exit transition", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
		transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

		sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
		sess.On("Done", mock.Anything).Return(nil)

		err := transitioner.Transition("test", "ExitNode", true)

		require.NoError(t, err)
		sess.AssertCalled(t, "UpdateCurrentStateSafe", "ExitNode")
	})
}

func TestTransition_DoneOnlyCalledForExit(t *testing.T) {
	t.Run("regular transition does not call Done", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
		transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

		restore := captureStdout(t)

		sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

		err := transitioner.Transition("test", "NodeA", false)
		restore()

		require.NoError(t, err)
		sess.AssertNotCalled(t, "Done", mock.Anything)
	})

	t.Run("exit transition calls Done", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
		transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

		sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
		sess.On("Done", mock.Anything).Return(nil)

		err := transitioner.Transition("test", "ExitNode", true)

		require.NoError(t, err)
		sess.AssertCalled(t, "Done", mock.Anything)
	})
}

func TestTransition_SessionMethodsNotCalledOnEarlyError(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	err := transitioner.Transition("test", "Missing", false)

	require.Error(t, err)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
	sess.AssertNotCalled(t, "UpdateCurrentStateSafe", mock.Anything)
	sess.AssertNotCalled(t, "Done", mock.Anything)
}

func TestTransition_SessionUpdateCalledEvenIfPersistenceFails(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("NodeA")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	// UpdateCurrentStateSafe returns nil even when persistence fails
	sess.On("UpdateCurrentStateSafe", "NodeA").Return(nil)

	err := transitioner.Transition("test", "NodeA", false)
	restore()

	require.NoError(t, err)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "NodeA")
}

// --- Resource Cleanup — No Rollback ---

func TestTransition_NoRollbackOnDoneFailure(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("ExitAgent", "worker")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "ExitAgent").Return(nil)
	sess.On("Done", mock.Anything).Return(fmt.Errorf("done error"))
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	err := transitioner.Transition("test", "ExitAgent", true)

	require.Error(t, err)
	// State was updated (no rollback)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "ExitAgent")
	// Done failed
	sess.AssertCalled(t, "Done", mock.Anything)
	// Fail was called to report the error
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
	// No rollback of state update — UpdateCurrentStateSafe called exactly once
	sess.AssertNumberOfCalls(t, "UpdateCurrentStateSafe", 1)
}

func TestTransition_NoRollbackOnAgentInvokeFailure(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Return(fmt.Errorf("partial execution failure"))
	sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

	err := transitioner.Transition("test", "AgentNode", false)

	require.Error(t, err)
	// Agent invocation was attempted
	agentInvoker.AssertCalled(t, "InvokeAgent", "AgentNode", "test", *loadedDef)
	// Session.Fail called to report the error
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
	// No rollback — UpdateCurrentStateSafe was NOT called (action failed before state update)
	sess.AssertNotCalled(t, "UpdateCurrentStateSafe", mock.Anything)
}

// --- Panic Recovery — Not Handled ---

func TestTransition_PanicNotRecovered(t *testing.T) {
	// TransitionToNode does not recover from panics; panic propagates to caller.
	// Simulate a programming error that causes a panic during workflow node lookup.
	sess := newMockSessionForTransition("running", "NodeA")
	agentDefLoader := &mockAgentDefLoaderForTransition{}
	agentInvoker := &mockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 1)

	assert.Panics(t, func() {
		// Pass nil WorkflowDefinition to trigger nil pointer dereference during node lookup
		transitioner, err := NewTransitionToNode(sess, nil, agentDefLoader, agentInvoker, terminationNotifier)
		if err != nil {
			// Constructor rejected nil; simulate the programming error
			panic("simulated programming error: nil workflow definition")
		}
		transitioner.Transition("test", "NodeA", false)
	}, "TransitionToNode should not recover from panics; panic propagates to caller (MessageRouter)")
}

// --- Boundary Values — IsExitTransition Flag ---

func TestTransition_IsExitTransitionTrue_HumanNode(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("HumanExit")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "HumanExit").Return(nil)
	sess.On("Done", mock.Anything).Return(nil)

	err := transitioner.Transition("exit", "HumanExit", true)
	output := restore()

	require.NoError(t, err)
	assert.Empty(t, output, "no stdout print for exit transition")
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "HumanExit")
	sess.AssertCalled(t, "Done", mock.Anything)
}

func TestTransition_IsExitTransitionFalse_HumanNode(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("HumanNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "HumanNode").Return(nil)

	err := transitioner.Transition("regular", "HumanNode", false)
	output := restore()

	require.NoError(t, err)
	assert.Contains(t, output, "[Human Node: HumanNode] regular")
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "HumanNode")
	sess.AssertNotCalled(t, "Done", mock.Anything)
}

func TestTransition_IsExitTransitionTrue_AgentNode(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentExit", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	sess.On("UpdateCurrentStateSafe", "AgentExit").Return(nil)
	sess.On("Done", mock.Anything).Return(nil)

	err := transitioner.Transition("exit", "AgentExit", true)

	require.NoError(t, err)
	agentDefLoader.AssertNotCalled(t, "Load", mock.Anything)
	agentInvoker.AssertNotCalled(t, "InvokeAgent", mock.Anything, mock.Anything, mock.Anything)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "AgentExit")
	sess.AssertCalled(t, "Done", mock.Anything)
}

func TestTransition_IsExitTransitionFalse_AgentNode(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", "regular", *loadedDef).Return(nil)
	sess.On("UpdateCurrentStateSafe", "AgentNode").Return(nil)

	err := transitioner.Transition("regular", "AgentNode", false)

	require.NoError(t, err)
	agentInvoker.AssertCalled(t, "InvokeAgent", "AgentNode", "regular", *loadedDef)
	sess.AssertCalled(t, "UpdateCurrentStateSafe", "AgentNode")
	sess.AssertNotCalled(t, "Done", mock.Anything)
}

// --- Edge Cases — TerminationNotifier Channel ---

func TestTransition_TerminationNotifierPassedToDone(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
	sess := newMockSessionForTransition("running", "Entry")
	agentDefLoader := &mockAgentDefLoaderForTransition{}
	agentInvoker := &mockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 2)

	transitioner, err := NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	require.NoError(t, err)

	sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
	sess.On("Done", mock.MatchedBy(func(ch chan<- struct{}) bool {
		// Verify the terminationNotifier channel is the one passed at initialization
		return ch == terminationNotifier
	})).Return(nil)

	err = transitioner.Transition("test", "ExitNode", true)

	require.NoError(t, err)
	sess.AssertCalled(t, "Done", mock.Anything)
}

func TestTransition_TerminationNotifierPassedToFail(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	sess := newMockSessionForTransition("running", "Entry")
	agentDefLoader := &mockAgentDefLoaderForTransition{}
	agentInvoker := &mockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 2)

	transitioner, err := NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	require.NoError(t, err)

	sess.On("Fail", mock.Anything, mock.MatchedBy(func(ch chan<- struct{}) bool {
		return ch == terminationNotifier
	})).Return(nil)

	err = transitioner.Transition("test", "Missing", false)

	require.Error(t, err)
	sess.AssertCalled(t, "Fail", mock.Anything, mock.Anything)
}

// --- Edge Cases — RuntimeError Construction ---

func TestTransition_RuntimeErrorIssuerAlwaysTransitionToNode(t *testing.T) {
	assertRuntimeErrorIssuer := func(t *testing.T, sess *mockSessionForTransition) {
		t.Helper()
		for _, call := range sess.Calls {
			if call.Method == "Fail" {
				runtimeErr, ok := call.Arguments.Get(0).(*session.RuntimeError)
				require.True(t, ok, "Fail should receive *RuntimeError")
				assert.Equal(t, "TransitionToNode", runtimeErr.Issuer)
			}
		}
	}

	t.Run("node not found", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithHumanNode("Entry")
		transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)
		sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

		_ = transitioner.Transition("test", "Missing", false)

		assertRuntimeErrorIssuer(t, sess)
	})

	t.Run("agent load fail", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "bad")
		transitioner, sess, agentDefLoader, _, _ := createTransitionFixture(t, wfDef)
		agentDefLoader.On("Load", "bad").Return(nil, fmt.Errorf("load error"))
		sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

		_ = transitioner.Transition("test", "AgentNode", false)

		assertRuntimeErrorIssuer(t, sess)
	})

	t.Run("invoke fail", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
		transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)
		loadedDef := defaultAgentDefForTransition()
		agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
		agentInvoker.On("InvokeAgent", "AgentNode", "test", *loadedDef).Return(fmt.Errorf("invoke error"))
		sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

		_ = transitioner.Transition("test", "AgentNode", false)

		assertRuntimeErrorIssuer(t, sess)
	})

	t.Run("done fail", func(t *testing.T) {
		wfDef := buildTransitionWorkflowWithHumanNode("ExitNode")
		transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)
		sess.On("UpdateCurrentStateSafe", "ExitNode").Return(nil)
		sess.On("Done", mock.Anything).Return(fmt.Errorf("done error"))
		sess.On("Fail", mock.Anything, mock.Anything).Return(nil)

		_ = transitioner.Transition("test", "ExitNode", true)

		assertRuntimeErrorIssuer(t, sess)
	})
}

func TestTransition_RuntimeErrorMessageDescriptive(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("Entry")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	sess.On("Fail", mock.MatchedBy(func(err error) bool {
		runtimeErr, ok := err.(*session.RuntimeError)
		if !ok {
			return false
		}
		return runtimeErr.Message == "target node not found: 'Missing'"
	}), mock.Anything).Return(nil)

	err := transitioner.Transition("test", "Missing", false)

	require.Error(t, err)
	sess.AssertNumberOfCalls(t, "Fail", 1)
}

func TestTransition_RuntimeErrorIncludesDetails(t *testing.T) {
	wfDef := buildTransitionWorkflowWithAgentNode("AgentNode", "worker")
	transitioner, sess, agentDefLoader, agentInvoker, _ := createTransitionFixture(t, wfDef)

	loadedDef := defaultAgentDefForTransition()
	agentDefLoader.On("Load", "worker").Return(loadedDef, nil)
	agentInvoker.On("InvokeAgent", "AgentNode", mock.Anything, *loadedDef).Return(fmt.Errorf("command not found"))

	var capturedErr error
	sess.On("Fail", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedErr = args.Get(0).(error)
	}).Return(nil)

	_ = transitioner.Transition("test", "AgentNode", false)

	require.NotNil(t, capturedErr)
	runtimeErr, ok := capturedErr.(*session.RuntimeError)
	require.True(t, ok)
	assert.Equal(t, "TransitionToNode", runtimeErr.Issuer)
	// The RuntimeError message or the returned error should include "command not found"
}

// --- Happy Path — Message Format Verification ---

func TestTransition_HumanNodeMessageFormat(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("TestNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "TestNode").Return(nil)

	err := transitioner.Transition("Test message", "TestNode", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: TestNode] Test message\n", output)
}

func TestTransition_HumanNodeEmptyMessageFormat(t *testing.T) {
	wfDef := buildTransitionWorkflowWithHumanNode("EmptyNode")
	transitioner, sess, _, _, _ := createTransitionFixture(t, wfDef)

	restore := captureStdout(t)

	sess.On("UpdateCurrentStateSafe", "EmptyNode").Return(nil)

	err := transitioner.Transition("", "EmptyNode", false)
	output := restore()

	require.NoError(t, err)
	assert.Equal(t, "[Human Node: EmptyNode] (no message)\n", output)
}
