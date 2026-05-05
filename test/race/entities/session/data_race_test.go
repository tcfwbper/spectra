package session_race

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateSessionDataSafe_ConcurrentWrites(t *testing.T) {
	s := newTestSession(t)
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_ = s.UpdateSessionDataSafe("k", fmt.Sprintf("value-%d", idx))
		}(i)
	}

	wg.Wait()

	// Final value should be one of the written values
	val, ok := s.GetSessionDataSafe("k")
	assert.True(t, ok)
	assert.Contains(t, val, "value-")
}

func TestGetSessionDataSafe_ConcurrentReadDuringWrite(t *testing.T) {
	s := newTestSession(t)
	_ = s.UpdateSessionDataSafe("k", "initial")

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // readers + writers

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_ = s.UpdateSessionDataSafe("k", fmt.Sprintf("w-%d", idx))
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			val, ok := s.GetSessionDataSafe("k")
			// Reader should always get a consistent value (not a partially written one)
			if ok {
				assert.NotNil(t, val)
			}
		}()
	}

	wg.Wait()
}
