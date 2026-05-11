package storage_race

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfwbper/spectra/storage"
)

func TestStorageLayout_ConcurrentAccess(t *testing.T) {

	const goroutines = 20
	projectRoot := "/home/user/project"
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	workflowName := "CodeReview"
	agentRole := "Architect"

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			r1 := storage.GetSpectraDir(projectRoot)
			assert.Equal(t, "/home/user/project/.spectra", r1)

			r2 := storage.GetSessionsDir(projectRoot)
			assert.Equal(t, "/home/user/project/.spectra/sessions", r2)

			r3 := storage.GetWorkflowsDir(projectRoot)
			assert.Equal(t, "/home/user/project/.spectra/workflows", r3)

			r4 := storage.GetAgentsDir(projectRoot)
			assert.Equal(t, "/home/user/project/.spectra/agents", r4)

			r5 := storage.GetSessionDir(projectRoot, uuid)
			assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000", r5)

			r6 := storage.GetSessionMetadataPath(projectRoot, uuid)
			assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/session.json", r6)

			r7 := storage.GetEventHistoryPath(projectRoot, uuid)
			assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/events.jsonl", r7)

			r8 := storage.GetRuntimeSocketPath(projectRoot, uuid)
			assert.Equal(t, "/home/user/project/.spectra/sessions/550e8400-e29b-41d4-a716-446655440000/runtime.sock", r8)

			r9 := storage.GetWorkflowPath(projectRoot, workflowName)
			assert.Equal(t, "/home/user/project/.spectra/workflows/CodeReview.yaml", r9)

			r10 := storage.GetAgentPath(projectRoot, agentRole)
			assert.Equal(t, "/home/user/project/.spectra/agents/Architect.yaml", r10)
		}()
	}

	wg.Wait()
}
