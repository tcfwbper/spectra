package session_race

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCurrentStateSafe_ConcurrentUpdates(t *testing.T) {
	s := newTestSession(t)
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_ = s.UpdateCurrentStateSafe(fmt.Sprintf("state-%d", idx))
		}(i)
	}

	wg.Wait()

	// After all goroutines complete, GetCurrentStateSafe must return one of the written values
	result := s.GetCurrentStateSafe()
	assert.Contains(t, result, "state-")
}
