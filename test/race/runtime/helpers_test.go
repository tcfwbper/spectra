package runtime_race

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/runtime"
)

// =============================================================================
// Thread-Safe Mock Session
// =============================================================================

// raceSafeMockSession is a fully thread-safe Session mock for race tests.
type raceSafeMockSession struct {
	mu sync.Mutex

	// Configurable return values
	statusResult       string
	currentStateResult string
	sessionDataVal     any
	sessionDataOK      bool
	failErr            error
	doneErr            error

	// Call tracking
	failCalled               int
	updateCurrentStateCalled int
	updateCurrentStateInput  string
}

func (m *raceSafeMockSession) Run() error { return nil }

func (m *raceSafeMockSession) Done(notifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.doneErr
}

func (m *raceSafeMockSession) Fail(err error, notifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCalled++
	if m.failCalled == 1 {
		return m.failErr
	}
	// Second call returns "already failed" to simulate first-error-wins.
	return errAlreadyFailed
}

func (m *raceSafeMockSession) UpdateCurrentStateSafe(newState string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCurrentStateCalled++
	m.updateCurrentStateInput = newState
	m.currentStateResult = newState
	return nil
}

func (m *raceSafeMockSession) UpdateSessionDataSafe(key string, value any) error {
	return nil
}

func (m *raceSafeMockSession) UpdateEventHistorySafe(event entities.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *raceSafeMockSession) GetStatusSafe() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusResult
}

func (m *raceSafeMockSession) GetCurrentStateSafe() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentStateResult
}

func (m *raceSafeMockSession) GetErrorSafe() error {
	return nil
}

func (m *raceSafeMockSession) GetMetadataSnapshotSafe() session.SessionMetadata {
	m.mu.Lock()
	defer m.mu.Unlock()
	return session.SessionMetadata{
		ID:           testSessionID,
		WorkflowName: testWorkflowName,
		Status:       m.statusResult,
		CurrentState: m.currentStateResult,
	}
}

func (m *raceSafeMockSession) GetSessionDataSafe(key string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessionDataVal, m.sessionDataOK
}

// =============================================================================
// Thread-Safe Mock Metadata Store
// =============================================================================

type raceSafeMetadataStore struct {
	mu sync.Mutex
}

func (m *raceSafeMetadataStore) Write(meta session.SessionMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

// =============================================================================
// Thread-Safe Mock Event Store
// =============================================================================

type raceSafeEventStore struct {
	mu sync.Mutex
}

func (m *raceSafeEventStore) Append(event *entities.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

// =============================================================================
// Thread-Safe Mock TransitionToNode for EventProcessor
// =============================================================================

type raceSafeTransitionToNode struct {
	mu   sync.Mutex
	sess *raceSafeMockSession
}

func (m *raceSafeTransitionToNode) Execute(targetNodeName, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sess.mu.Lock()
	m.sess.currentStateResult = targetNodeName
	m.sess.mu.Unlock()
	return nil
}

// =============================================================================
// Mock EventProcessor for MessageRouter
// =============================================================================

type raceSafeEventProcessor struct {
	mu   sync.Mutex
	resp *entities.RuntimeResponse
}

func (m *raceSafeEventProcessor) ProcessEvent(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.resp
}

// =============================================================================
// Mock ErrorProcessor for MessageRouter
// =============================================================================

type raceSafeErrorProcessor struct {
	mu   sync.Mutex
	resp *entities.RuntimeResponse
}

func (m *raceSafeErrorProcessor) ProcessError(sessionUUID string, msg *entities.RuntimeMessage) *entities.RuntimeResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.resp
}

// =============================================================================
// Workflow Definition Mocks
// =============================================================================

type raceErrorProcessorWfDef struct {
	nodes []*components.Node
}

func (m *raceErrorProcessorWfDef) Nodes() []*components.Node {
	return m.nodes
}

type raceEventProcessorWfDef struct {
	nodes           []*components.Node
	transitions     []*components.Transition
	exitTransitions []*components.ExitTransition
}

func (m *raceEventProcessorWfDef) Nodes() []*components.Node {
	return m.nodes
}

func (m *raceEventProcessorWfDef) Transitions() []*components.Transition {
	return m.transitions
}

func (m *raceEventProcessorWfDef) ExitTransitions() []*components.ExitTransition {
	return m.exitTransitions
}

type raceTransitionWorkflowDef struct {
	nodes []*components.Node
}

func (m *raceTransitionWorkflowDef) Nodes() []*components.Node {
	return m.nodes
}

// =============================================================================
// Mock AgentDefLoader and AgentInvoker for TransitionToNode
// =============================================================================

type raceAgentDefLoader struct {
	mu  sync.Mutex
	def runtime.AgentDef
	err error
}

func (m *raceAgentDefLoader) Load(agentRole string) (runtime.AgentDef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.def, m.err
}

type raceAgentInvoker struct {
	mu  sync.Mutex
	err error
}

func (m *raceAgentInvoker) InvokeAgent(nodeName, message string, agentDef runtime.AgentDef) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.err
}

// =============================================================================
// Test Helpers
// =============================================================================

const (
	testSessionID    = "550e8400-e29b-41d4-a716-446655440000"
	testWorkflowName = "my-workflow"
)

var errAlreadyFailed = errorString("session already failed")

type errorString string

func (e errorString) Error() string { return string(e) }

func newRacePersistentSession(t *testing.T, sess *raceSafeMockSession) *runtime.PersistentSession {
	t.Helper()
	return runtime.NewPersistentSession(
		sess,
		&raceSafeMetadataStore{},
		&raceSafeEventStore{},
		logger.NewNopLogger(),
	)
}

func newTerminationChannel() chan struct{} {
	return make(chan struct{}, 2)
}

func mustNewNode(t *testing.T, name, nodeType, agentRole string) *components.Node {
	t.Helper()
	n, err := components.NewNode(name, nodeType, agentRole, "Test node "+name)
	require.NoError(t, err, "mustNewNode(%q, %q, %q)", name, nodeType, agentRole)
	return n
}

func mustNewTransition(t *testing.T, from, eventType, to string) *components.Transition {
	t.Helper()
	tr, err := components.NewTransition(from, eventType, to)
	require.NoError(t, err, "mustNewTransition(%q, %q, %q)", from, eventType, to)
	return tr
}

func mustNewEventRuntimeMessage(t *testing.T, claudeSessionID string, payload json.RawMessage) *entities.RuntimeMessage {
	t.Helper()
	msg, err := entities.NewRuntimeMessage("event", payload, claudeSessionID)
	require.NoError(t, err, "mustNewEventRuntimeMessage")
	return msg
}

func mustNewErrorRuntimeMessage(t *testing.T, claudeSessionID string, payload json.RawMessage) *entities.RuntimeMessage {
	t.Helper()
	msg, err := entities.NewRuntimeMessage("error", payload, claudeSessionID)
	require.NoError(t, err, "mustNewErrorRuntimeMessage")
	return msg
}

// Compile guards
var (
	_ runtime.Session              = (*raceSafeMockSession)(nil)
	_ runtime.SessionMetadataStore = (*raceSafeMetadataStore)(nil)
	_ runtime.EventStore           = (*raceSafeEventStore)(nil)
)
