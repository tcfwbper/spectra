package race_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tcfwbper/spectra/runtime"
	"github.com/tcfwbper/spectra/storage"
)

// --- Mocks for SessionInitializer race tests ---

// raceMockWDLForInit is a thread-safe mock WorkflowDefinitionLoader for race tests.
type raceMockWDLForInit struct {
	mock.Mock
}

func (m *raceMockWDLForInit) Load(workflowName string) (*storage.WorkflowDefinition, error) {
	args := m.Called(workflowName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.WorkflowDefinition), args.Error(1)
}

// raceMockSDMForInit is a thread-safe mock SessionDirectoryManager for race tests.
type raceMockSDMForInit struct {
	mock.Mock
}

func (m *raceMockSDMForInit) CreateSessionDirectory(sessionUUID string) error {
	args := m.Called(sessionUUID)
	return args.Error(0)
}

// =====================================================================
// Concurrent Behaviour — Multiple Initializers
// =====================================================================

func TestInitialize_ConcurrentInitializers(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".spectra", "sessions"), 0775))

	wdl := &raceMockWDLForInit{}
	wdl.On("Load", "TestWorkflow").Return(&storage.WorkflowDefinition{
		Name:      "TestWorkflow",
		EntryNode: "start",
		Nodes:     []storage.Node{{Name: "start", Type: "agent"}},
	}, nil)

	sdm := &raceMockSDMForInit{}
	sdm.On("CreateSessionDirectory", mock.AnythingOfType("string")).Return(nil)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	var mu sync.Mutex
	sessionIDs := make(map[string]bool)
	errCount := 0

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			si, err := runtime.NewSessionInitializer(tmpDir, wdl, sdm)
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
				return
			}

			terminationNotifier := make(chan struct{}, 2)
			sess, err := si.Initialize("TestWorkflow", terminationNotifier)
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			sessionIDs[sess.GetID()] = true
			mu.Unlock()
		}()
	}

	wg.Wait()

	assert.Equal(t, 0, errCount, "all initializations should succeed")
	assert.Equal(t, goroutines, len(sessionIDs), "all session UUIDs should be unique")
}
