package runtime

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities/session"
)

// =============================================================================
// Test Helpers — TransitionToNode
// =============================================================================
//
// Production surface expected in runtime/transition_to_node.go:
//   - type TransitionToNode struct { ... }
//   - func NewTransitionToNode(ps *PersistentSession, wfDef TransitionWorkflowDef, loader TransitionAgentDefLoader, invoker TransitionAgentInvoker, opts ...TransitionToNodeOption) *TransitionToNode
//   - func (t *TransitionToNode) Execute(targetNodeName, message string) error
//   - type TransitionToNodeOption func(*TransitionToNode)
//   - func WithOutput(w io.Writer) TransitionToNodeOption
//   - type TransitionWorkflowDef interface { Nodes() []*components.Node }
//   - type TransitionAgentDefLoader interface { Load(agentRole string) (AgentDef, error) }
//   - type TransitionAgentInvoker interface { InvokeAgent(nodeName, message string, agentDef AgentDef) error }
// =============================================================================

// --- Mock: TransitionToNode WorkflowDefinition interface ---

type mockTransitionWorkflowDef struct {
	nodes []*components.Node
}

func (m *mockTransitionWorkflowDef) Nodes() []*components.Node {
	return m.nodes
}

// --- Mock: AgentDefinitionLoader interface for TransitionToNode ---

type mockTransitionAgentDefLoader struct {
	loadCalled int
	loadInput  string
	loadResult AgentDef
	loadErr    error
}

func (m *mockTransitionAgentDefLoader) Load(agentRole string) (AgentDef, error) {
	m.loadCalled++
	m.loadInput = agentRole
	return m.loadResult, m.loadErr
}

// --- Mock: AgentInvoker interface for TransitionToNode ---

type mockTransitionAgentInvoker struct {
	invokeAgentCalled   int
	invokeAgentNodeName string
	invokeAgentMessage  string
	invokeAgentDef      AgentDef
	invokeAgentErr      error
	invokeAgentCallback func() // optional: for call order tracking
}

func (m *mockTransitionAgentInvoker) InvokeAgent(nodeName, message string, agentDef AgentDef) error {
	m.invokeAgentCalled++
	m.invokeAgentNodeName = nodeName
	m.invokeAgentMessage = message
	m.invokeAgentDef = agentDef
	if m.invokeAgentCallback != nil {
		m.invokeAgentCallback()
	}
	return m.invokeAgentErr
}

// --- Recording mock session for call-order and lifecycle assertions ---

type recordingMockSession struct {
	mockSession
	mu       sync.Mutex
	callLog  []string
	callback func(method string)
}

func (r *recordingMockSession) UpdateCurrentStateSafe(newState string) error {
	r.mu.Lock()
	r.callLog = append(r.callLog, "UpdateCurrentStateSafe")
	r.mu.Unlock()
	if r.callback != nil {
		r.callback("UpdateCurrentStateSafe")
	}
	r.mockSession.updateCurrentStateCalled++
	r.mockSession.updateCurrentStateInput = newState
	return r.mockSession.updateCurrentStateErr
}

func (r *recordingMockSession) Run() error {
	r.mu.Lock()
	r.callLog = append(r.callLog, "Run")
	r.mu.Unlock()
	return r.mockSession.runErr
}

func (r *recordingMockSession) Done(notifier chan<- struct{}) error {
	r.mu.Lock()
	r.callLog = append(r.callLog, "Done")
	r.mu.Unlock()
	return r.mockSession.doneErr
}

func (r *recordingMockSession) Fail(err error, notifier chan<- struct{}) error {
	r.mu.Lock()
	r.callLog = append(r.callLog, "Fail")
	r.mu.Unlock()
	return r.mockSession.failErr
}

// --- Failing Writer for stdout error simulation ---

type failingWriter struct {
	err error
}

func (fw *failingWriter) Write(_ []byte) (int, error) {
	return 0, fw.err
}

// --- Thread-safe Buffer for concurrent stdout capture ---

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

// --- Fixture Builders ---

// mustNewNode creates a *components.Node for testing; fails test on error.
func mustNewNode(t *testing.T, name, nodeType, agentRole string) *components.Node {
	t.Helper()
	n, err := components.NewNode(name, nodeType, agentRole, fmt.Sprintf("Test node %s", name))
	require.NoError(t, err, "mustNewNode(%q, %q, %q)", name, nodeType, agentRole)
	return n
}

// newTransitionTestPersistentSession creates a PersistentSession backed by
// the given Session mock for TransitionToNode tests.
func newTransitionTestPersistentSession(sess Session) *PersistentSession {
	metaStore := newDefaultMockMetadataStore()
	evStore := newDefaultMockEventStore()
	log := newDefaultMockLogger()
	return NewPersistentSession(sess, metaStore, evStore, log)
}

// newRecordingSession creates a recordingMockSession with default config.
func newRecordingSession(callback func(string)) *recordingMockSession {
	return &recordingMockSession{
		mockSession: mockSession{
			id:           testSessionID,
			workflowName: testWorkflowName,
			getMetadataSnapshotResult: session.SessionMetadata{
				ID:           testSessionID,
				WorkflowName: testWorkflowName,
				Status:       "running",
				CreatedAt:    testCreatedAt,
				UpdatedAt:    testCreatedAt + 1,
				CurrentState: testEntryNode,
				SessionData:  map[string]any{},
			},
			getStatusResult:       "running",
			getCurrentStateResult: testEntryNode,
		},
		callback: callback,
	}
}

// =============================================================================
// Happy Path — Construction
// =============================================================================

func TestNewTransitionToNode_ValidDeps(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	require.NotNil(t, ttn)
}

// =============================================================================
// Happy Path — Execute
// =============================================================================

func TestTransitionToNode_Execute_HumanNode(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	var buf bytes.Buffer
	ttn := NewTransitionToNode(ps, wfDef, loader, invoker, WithOutput(&buf))
	err := ttn.Execute("HumanReview", "please review this")
	require.NoError(t, err)
	assert.Equal(t, "[HumanReview] please review this\n", buf.String())
	assert.Equal(t, 1, sess.updateCurrentStateCalled)
	assert.Equal(t, "HumanReview", sess.updateCurrentStateInput)
}

func TestTransitionToNode_Execute_HumanNodeEmptyMessage(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	var buf bytes.Buffer
	ttn := NewTransitionToNode(ps, wfDef, loader, invoker, WithOutput(&buf))
	err := ttn.Execute("HumanReview", "")
	require.NoError(t, err)
	assert.Equal(t, "[HumanReview] (no message)\n", buf.String())
	assert.Equal(t, 1, sess.updateCurrentStateCalled)
	assert.Equal(t, "HumanReview", sess.updateCurrentStateInput)
}

func TestTransitionToNode_Execute_HumanNodeSpecialChars(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	var buf bytes.Buffer
	message := "line1\n\"quoted\" $var"
	ttn := NewTransitionToNode(ps, wfDef, loader, invoker, WithOutput(&buf))
	err := ttn.Execute("HumanReview", message)
	require.NoError(t, err)
	expected := "[HumanReview] line1\n\"quoted\" $var\n"
	assert.Equal(t, expected, buf.String())
	assert.Equal(t, 1, sess.updateCurrentStateCalled)
	assert.Equal(t, "HumanReview", sess.updateCurrentStateInput)
}

func TestTransitionToNode_Execute_AgentNode(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "Coder", "agent", "Developer")},
	}
	agentDef := newDefaultMockAgentDefinition()
	loader := &mockTransitionAgentDefLoader{loadResult: agentDef}
	invoker := &mockTransitionAgentInvoker{}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	err := ttn.Execute("Coder", "implement feature")
	require.NoError(t, err)
	assert.Equal(t, 1, loader.loadCalled)
	assert.Equal(t, "Developer", loader.loadInput)
	assert.Equal(t, 1, invoker.invokeAgentCalled)
	assert.Equal(t, "Coder", invoker.invokeAgentNodeName)
	assert.Equal(t, "implement feature", invoker.invokeAgentMessage)
	assert.Equal(t, agentDef, invoker.invokeAgentDef)
	assert.Equal(t, 1, sess.updateCurrentStateCalled)
	assert.Equal(t, "Coder", sess.updateCurrentStateInput)
}

// =============================================================================
// Error Propagation — Execute
// =============================================================================

func TestTransitionToNode_Execute_NodeNotFound(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	err := ttn.Execute("NonExistent", "msg")
	require.Error(t, err)
	assert.Equal(t, "target node 'NonExistent' not found in workflow", err.Error())
	assert.Equal(t, 0, sess.updateCurrentStateCalled)
}

func TestTransitionToNode_Execute_AgentDefLoadFails(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "Coder", "agent", "MissingRole")},
	}
	loader := &mockTransitionAgentDefLoader{
		loadErr: errors.New("file not found"),
	}
	invoker := &mockTransitionAgentInvoker{}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	err := ttn.Execute("Coder", "msg")
	require.Error(t, err)
	assert.Equal(t, "failed to load agent definition for role 'MissingRole': file not found", err.Error())
	assert.Equal(t, 0, invoker.invokeAgentCalled)
	assert.Equal(t, 0, sess.updateCurrentStateCalled)
}

func TestTransitionToNode_Execute_AgentInvokeFails(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "Coder", "agent", "Developer")},
	}
	agentDef := newDefaultMockAgentDefinition()
	loader := &mockTransitionAgentDefLoader{loadResult: agentDef}
	invoker := &mockTransitionAgentInvoker{
		invokeAgentErr: errors.New("claude not in PATH"),
	}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	err := ttn.Execute("Coder", "msg")
	require.Error(t, err)
	assert.Equal(t, "failed to invoke agent for node 'Coder': claude not in PATH", err.Error())
	assert.Equal(t, 0, sess.updateCurrentStateCalled)
}

func TestTransitionToNode_Execute_UpdateStateFails(t *testing.T) {
	sess := newDefaultMockSession()
	sess.updateCurrentStateErr = errors.New("validation failed")
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	var buf bytes.Buffer
	ttn := NewTransitionToNode(ps, wfDef, loader, invoker, WithOutput(&buf))
	err := ttn.Execute("HumanReview", "msg")
	require.Error(t, err)
	assert.Equal(t, "failed to update current state: validation failed", err.Error())
	assert.Equal(t, "[HumanReview] msg\n", buf.String())
}

func TestTransitionToNode_Execute_StdoutWriteFails(t *testing.T) {
	sess := newDefaultMockSession()
	ps := newTransitionTestPersistentSession(sess)
	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	fw := &failingWriter{err: errors.New("broken pipe")}
	ttn := NewTransitionToNode(ps, wfDef, loader, invoker, WithOutput(fw))
	err := ttn.Execute("HumanReview", "msg")
	require.Error(t, err)
	assert.Equal(t, 0, sess.updateCurrentStateCalled)
}

// =============================================================================
// Mock / Dependency Interaction — Execute
// =============================================================================

func TestTransitionToNode_Execute_ActionBeforeStateUpdate(t *testing.T) {
	var callOrder []string
	var mu sync.Mutex

	rec := newRecordingSession(func(method string) {
		mu.Lock()
		callOrder = append(callOrder, method)
		mu.Unlock()
	})
	ps := newTransitionTestPersistentSession(rec)

	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "Coder", "agent", "Dev")},
	}
	agentDef := newDefaultMockAgentDefinition()
	loader := &mockTransitionAgentDefLoader{loadResult: agentDef}
	invoker := &mockTransitionAgentInvoker{
		invokeAgentCallback: func() {
			mu.Lock()
			callOrder = append(callOrder, "InvokeAgent")
			mu.Unlock()
		},
	}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	err := ttn.Execute("Coder", "msg")
	require.NoError(t, err)
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, callOrder, 2)
	assert.Equal(t, "InvokeAgent", callOrder[0])
	assert.Equal(t, "UpdateCurrentStateSafe", callOrder[1])
}

func TestTransitionToNode_Execute_NoLifecycleMethodsCalled(t *testing.T) {
	rec := newRecordingSession(nil)
	ps := newTransitionTestPersistentSession(rec)

	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "HumanReview", "human", "")},
	}
	loader := &mockTransitionAgentDefLoader{}
	invoker := &mockTransitionAgentInvoker{}

	var buf bytes.Buffer
	ttn := NewTransitionToNode(ps, wfDef, loader, invoker, WithOutput(&buf))
	err := ttn.Execute("HumanReview", "msg")
	require.NoError(t, err)
	for _, method := range rec.callLog {
		assert.NotEqual(t, "Run", method)
		assert.NotEqual(t, "Done", method)
		assert.NotEqual(t, "Fail", method)
	}
}

func TestTransitionToNode_Execute_AgentNodeNoStateUpdateOnInvokeError(t *testing.T) {
	rec := newRecordingSession(nil)
	ps := newTransitionTestPersistentSession(rec)

	wfDef := &mockTransitionWorkflowDef{
		nodes: []*components.Node{mustNewNode(t, "Coder", "agent", "Dev")},
	}
	agentDef := newDefaultMockAgentDefinition()
	loader := &mockTransitionAgentDefLoader{loadResult: agentDef}
	invoker := &mockTransitionAgentInvoker{
		invokeAgentErr: errors.New("agent failed"),
	}

	ttn := NewTransitionToNode(ps, wfDef, loader, invoker)
	err := ttn.Execute("Coder", "msg")
	require.Error(t, err)
	for _, method := range rec.callLog {
		assert.NotEqual(t, "UpdateCurrentStateSafe", method)
	}
}

// =============================================================================
// Compile guards — interface satisfaction checks
// =============================================================================

var (
	_ io.Writer = (*failingWriter)(nil)
	_ io.Writer = (*syncBuffer)(nil)
)
