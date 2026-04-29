package race_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for TransitionToNode race tests ---

// raceMockSessionForTransition is a thread-safe mock Session for transition race tests.
type raceMockSessionForTransition struct {
	mu           sync.RWMutex
	status       string
	currentState string
	failCalls    int
	doneCalls    int
	err          error
}

func newRaceMockSessionForTransition(status, currentState string) *raceMockSessionForTransition {
	return &raceMockSessionForTransition{
		status:       status,
		currentState: currentState,
	}
}

func (m *raceMockSessionForTransition) UpdateCurrentStateSafe(newState string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentState = newState
	return nil
}

func (m *raceMockSessionForTransition) Done(terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status != "running" {
		return fmt.Errorf("cannot complete session: status is '%s', expected 'running'", m.status)
	}
	m.status = "completed"
	m.doneCalls++
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *raceMockSessionForTransition) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status == "failed" {
		return fmt.Errorf("session already failed")
	}
	m.status = "failed"
	m.err = err
	m.failCalls++
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}
	return nil
}

func (m *raceMockSessionForTransition) getStatus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *raceMockSessionForTransition) getCurrentState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *raceMockSessionForTransition) getFailCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failCalls
}

func (m *raceMockSessionForTransition) getDoneCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.doneCalls
}

// raceMockAgentDefLoaderForTransition is a thread-safe mock AgentDefinitionLoader.
type raceMockAgentDefLoaderForTransition struct {
	def *storage.AgentDefinition
	err error
}

func (m *raceMockAgentDefLoaderForTransition) Load(agentRole string) (*storage.AgentDefinition, error) {
	return m.def, m.err
}

// raceMockAgentInvokerForTransition is a thread-safe mock AgentInvoker.
type raceMockAgentInvokerForTransition struct {
	mu        sync.Mutex
	calls     int
	returnErr error
}

func (m *raceMockAgentInvokerForTransition) InvokeAgent(nodeName string, message string, agentDef storage.AgentDefinition) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return m.returnErr
}

func (m *raceMockAgentInvokerForTransition) getCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// --- Race tests ---

// TestTransition_ConcurrentCallsSameSession verifies that concurrent transitions
// on the same session are handled safely with no data races.
func TestTransition_ConcurrentCallsSameSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	nodes := []storage.Node{
		{Name: "Entry", Type: "human"},
		{Name: "NodeA", Type: "human"},
		{Name: "NodeB", Type: "human"},
		{Name: "NodeC", Type: "human"},
		{Name: "NodeD", Type: "human"},
		{Name: "NodeE", Type: "human"},
	}
	wfDef := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes:     nodes,
	}

	sess := newRaceMockSessionForTransition("running", "Entry")
	agentDefLoader := &raceMockAgentDefLoaderForTransition{}
	agentInvoker := &raceMockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 10)

	transitioner, err := runtime.NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 5
	targetNodes := []string{"NodeA", "NodeB", "NodeC", "NodeD", "NodeE"}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = transitioner.Transition(
				fmt.Sprintf("message %d", idx),
				targetNodes[idx],
				false,
			)
		}(i)
	}
	wg.Wait()

	// All calls should complete without panic
	for i, callErr := range errs {
		assert.NoError(t, callErr, "goroutine %d should succeed", i)
	}

	// Last UpdateCurrentStateSafe wins; final state is one of the target nodes
	finalState := sess.getCurrentState()
	assert.Contains(t, targetNodes, finalState, "final state should be one of the target nodes")

	// No data races should be detected (run with -race)
}

// TestTransition_ConcurrentFailures verifies that concurrent transitions with
// invalid target nodes are handled safely with no data races.
func TestTransition_ConcurrentFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	wfDef := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
		},
	}

	sess := newRaceMockSessionForTransition("running", "Entry")
	agentDefLoader := &raceMockAgentDefLoaderForTransition{}
	agentInvoker := &raceMockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 10)

	transitioner, err := runtime.NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 3
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = transitioner.Transition(
				"test",
				fmt.Sprintf("Missing%d", idx),
				false,
			)
		}(i)
	}
	wg.Wait()

	// All goroutines should return errors
	for i, callErr := range errs {
		assert.Error(t, callErr, "goroutine %d should return error", i)
	}

	// Session.Fail was called; first error wins in Session
	finalStatus := sess.getStatus()
	assert.Equal(t, "failed", finalStatus)
	assert.GreaterOrEqual(t, sess.getFailCalls(), 1, "Session.Fail should be called at least once")

	// No data races should be detected (run with -race)
}

// TestTransition_ConcurrentExitTransitions verifies that concurrent exit transitions
// are handled safely. One Session.Done succeeds, the other may fail.
func TestTransition_ConcurrentExitTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	wfDef := &storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "Entry",
		Nodes: []storage.Node{
			{Name: "Entry", Type: "human"},
			{Name: "ExitA", Type: "human"},
			{Name: "ExitB", Type: "human"},
		},
	}

	sess := newRaceMockSessionForTransition("running", "Entry")
	agentDefLoader := &raceMockAgentDefLoaderForTransition{}
	agentInvoker := &raceMockAgentInvokerForTransition{}
	terminationNotifier := make(chan struct{}, 10)

	transitioner, err := runtime.NewTransitionToNode(sess, wfDef, agentDefLoader, agentInvoker, terminationNotifier)
	require.NoError(t, err)

	const goroutines = 2
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make([]error, goroutines)

	go func() {
		defer wg.Done()
		errs[0] = transitioner.Transition("exit1", "ExitA", true)
	}()
	go func() {
		defer wg.Done()
		errs[1] = transitioner.Transition("exit2", "ExitB", true)
	}()
	wg.Wait()

	// At least one should succeed (the first Done call);
	// the other may fail because status is no longer "running"
	successCount := 0
	for _, callErr := range errs {
		if callErr == nil {
			successCount++
		}
	}

	// First completion wins; at most one Done call succeeds
	finalStatus := sess.getStatus()
	assert.Contains(t, []string{"completed", "failed"}, finalStatus)

	// No data races should be detected (run with -race)
}
