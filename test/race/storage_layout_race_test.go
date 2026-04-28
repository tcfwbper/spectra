package race_test

import (
	"sync"
	"testing"

	"github.com/tcfwbper/spectra/storage"
)

// TestStorageLayout_ConcurrentAccess tests multiple goroutines calling path composition methods concurrently
func TestStorageLayout_ConcurrentAccess(t *testing.T) {
	projectRoot := "/home/user/project"
	numGoroutines := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			sessionUUID := "123e4567-e89b-12d3-a456-42661417" + string(rune('0'+idx))
			_ = storage.GetSessionDir(projectRoot, sessionUUID)
			_ = storage.GetSessionMetadataPath(projectRoot, sessionUUID)
			_ = storage.GetEventHistoryPath(projectRoot, sessionUUID)
		}(i)
	}

	wg.Wait()
}
