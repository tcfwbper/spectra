package runtime_race

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcfwbper/spectra/components"
	"github.com/tcfwbper/spectra/logger"
	"github.com/tcfwbper/spectra/runtime"
)

// =============================================================================
// Concurrent Behaviour — TransitionToNode
// =============================================================================
//
// Production surface required:
//   - runtime.NewTransitionToNode(ps, wfDef, loader, invoker, opts...)
//   - runtime.TransitionToNode.Execute(targetNodeName, message string) error
//   - runtime.WithOutput(w io.Writer) TransitionToNodeOption
//   - runtime.TransitionWorkflowDef interface
//   - runtime.TransitionAgentDefLoader interface
//   - runtime.TransitionAgentInvoker interface
//
// This test verifies concurrent calls to Execute for the same session
// serialize state updates without data races.
// =============================================================================

func TestTransitionToNode_Execute_ConcurrentCalls(t *testing.T) {
	// Setup:
	// - Mock WorkflowDefinition.Nodes() returns nodes "NodeA" (human) and "NodeB" (human).
	// - Capture stdout via thread-safe buffer.
	// - Mock PersistentSession.UpdateCurrentStateSafe serializes via internal lock and records calls.
	sess := &raceSafeMockSession{
		statusResult:       "running",
		currentStateResult: "Start",
	}

	ps := runtime.NewPersistentSession(
		sess,
		&raceSafeMetadataStore{},
		&raceSafeEventStore{},
		logger.NewNopLogger(),
	)

	wfDef := &raceTransitionWorkflowDef{
		nodes: []*components.Node{
			mustNewNode(t, "NodeA", "human", ""),
			mustNewNode(t, "NodeB", "human", ""),
		},
	}

	loader := &raceAgentDefLoader{}
	invoker := &raceAgentInvoker{}

	var buf safeBuffer
	ttn := runtime.NewTransitionToNode(ps, wfDef, loader, invoker, runtime.WithOutput(&buf))

	// Act:
	// - Launch two goroutines: one calls Execute("NodeA", "a"), another calls Execute("NodeB", "b") concurrently.
	var wg sync.WaitGroup
	var err1, err2 error
	wg.Add(2)
	go func() { defer wg.Done(); err1 = ttn.Execute("NodeA", "a") }()
	go func() { defer wg.Done(); err2 = ttn.Execute("NodeB", "b") }()
	wg.Wait()

	// Assert:
	// - Both calls return nil.
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// - PersistentSession.UpdateCurrentStateSafe called exactly twice (once with "NodeA", once with "NodeB").
	sess.mu.Lock()
	updateCount := sess.updateCurrentStateCalled
	sess.mu.Unlock()
	assert.Equal(t, 2, updateCount)

	// - No data race detected by race detector (-race flag).
}

// safeBuffer is a thread-safe bytes.Buffer for capturing output from
// concurrent calls.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
