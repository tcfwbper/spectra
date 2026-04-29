package race_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for AgentInvoker race tests ---

// raceMockSessionForInvoker is a thread-safe mock Session for invoker race tests.
type raceMockSessionForInvoker struct {
	mu          sync.RWMutex
	sessionID   string
	sessionData map[string]any
	failCalled  bool
}

func newRaceMockSessionForInvoker() *raceMockSessionForInvoker {
	return &raceMockSessionForInvoker{
		sessionID:   uuid.New().String(),
		sessionData: make(map[string]any),
	}
}

func (m *raceMockSessionForInvoker) GetID() string {
	return m.sessionID
}

func (m *raceMockSessionForInvoker) GetSessionDataSafe(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.sessionData[key]
	return val, ok
}

func (m *raceMockSessionForInvoker) UpdateSessionDataSafe(key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionData[key] = value
	return nil
}

func (m *raceMockSessionForInvoker) Fail(err error, terminationNotifier chan<- struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCalled = true
	return nil
}

// TestAgentInvoker_ConcurrentInvocationsDifferentNodes verifies multiple concurrent
// invocations for different nodes are thread-safe with no data races.
func TestAgentInvoker_ConcurrentInvocationsDifferentNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "agents"), 0755))

	sess := newRaceMockSessionForInvoker()

	agentDef := storage.AgentDefinition{
		Role:         "Worker",
		Model:        "sonnet",
		Effort:       "normal",
		SystemPrompt: "You are a test agent",
		AgentRoot:    "agents",
	}

	invoker, err := runtime.NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make([]error, goroutines)
	nodeNames := []string{"NodeA", "NodeB", "NodeC", "NodeD", "NodeE", "NodeF", "NodeG", "NodeH", "NodeI", "NodeJ"}

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			errors[idx] = invoker.InvokeAgent(nodeNames[idx], fmt.Sprintf("message for %s", nodeNames[idx]), agentDef)
		}(i)
	}
	wg.Wait()

	// All invocations should succeed
	for i, err := range errors {
		assert.NoError(t, err, "invocation for %s should succeed", nodeNames[i])
	}

	// Each node should have a distinct Claude session ID stored
	sess.mu.RLock()
	storedIDs := make(map[string]string)
	for _, name := range nodeNames {
		key := fmt.Sprintf("%s.ClaudeSessionID", name)
		val, ok := sess.sessionData[key]
		assert.True(t, ok, "session data should contain %s", key)
		if ok {
			strVal, isStr := val.(string)
			assert.True(t, isStr, "stored value should be string")
			if isStr {
				_, parseErr := uuid.Parse(strVal)
				assert.NoError(t, parseErr, "stored value should be valid UUID")
				storedIDs[name] = strVal
			}
		}
	}
	sess.mu.RUnlock()

	// All IDs should be distinct
	uniqueIDs := make(map[string]bool)
	for _, id := range storedIDs {
		uniqueIDs[id] = true
	}
	assert.Len(t, uniqueIDs, len(storedIDs), "each node should have a distinct Claude session ID")
}

// TestAgentInvoker_ConcurrentSameNode_RaceCondition verifies concurrent invocations
// for the same node do not panic (behavior is undefined per spec).
func TestAgentInvoker_ConcurrentSameNode_RaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "agents"), 0755))

	sess := newRaceMockSessionForInvoker()

	agentDef := storage.AgentDefinition{
		Role:         "Worker",
		Model:        "sonnet",
		Effort:       "normal",
		SystemPrompt: "You are a test agent",
		AgentRoot:    "agents",
	}

	invoker, err := runtime.NewAgentInvoker(sess, tmpDir)
	require.NoError(t, err)

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Verify no panic occurs
	assert.NotPanics(t, func() {
		for i := 0; i < goroutines; i++ {
			go func(idx int) {
				defer wg.Done()
				// Behavior undefined — may generate multiple UUIDs or race conditions
				_ = invoker.InvokeAgent("SameNode", fmt.Sprintf("message %d", idx), agentDef)
			}(i)
		}
		wg.Wait()
	}, "concurrent invocations for same node should not panic")
}
