package session_race

import (
	"fmt"
	"sync"
	"testing"
)

func TestGetters_ConcurrentReadDuringWrite(t *testing.T) {
	s := newTestSession(t)
	const goroutines = 50

	var wg sync.WaitGroup
	// Writers: Run (1), UpdateSessionDataSafe (goroutines), UpdateCurrentStateSafe (goroutines)
	// Readers: GetStatusSafe, GetCurrentStateSafe, GetErrorSafe, GetMetadataSnapshotSafe (goroutines each)
	totalWriters := 1 + goroutines + goroutines
	totalReaders := goroutines * 4
	wg.Add(totalWriters + totalReaders)

	// One goroutine calls Run()
	go func() {
		defer wg.Done()
		_ = s.Run()
	}()

	// Writer goroutines: UpdateSessionDataSafe
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_ = s.UpdateSessionDataSafe(fmt.Sprintf("key-%d", idx), fmt.Sprintf("val-%d", idx))
		}(i)
	}

	// Writer goroutines: UpdateCurrentStateSafe
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_ = s.UpdateCurrentStateSafe(fmt.Sprintf("state-%d", idx))
		}(i)
	}

	// Reader goroutines: GetStatusSafe
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = s.GetStatusSafe()
		}()
	}

	// Reader goroutines: GetCurrentStateSafe
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = s.GetCurrentStateSafe()
		}()
	}

	// Reader goroutines: GetErrorSafe
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = s.GetErrorSafe()
		}()
	}

	// Reader goroutines: GetMetadataSnapshotSafe
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = s.GetMetadataSnapshotSafe()
		}()
	}

	wg.Wait()
}
