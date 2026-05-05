package session_race

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectra-ai/spectra/entities"
)

func TestUpdateEventHistorySafe_ConcurrentAppends(t *testing.T) {
	s := newTestSession(t)
	const n = 50

	// Create N distinct events
	events := make([]*entities.Event, n)
	for i := 0; i < n; i++ {
		events[i] = newTestEvent(t, generateUUID(i))
	}

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			err := s.UpdateEventHistorySafe(*events[idx])
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// All N events should be present
	require.Len(t, s.EventHistory, n)

	// Verify all events are present (order not guaranteed due to concurrency)
	idSet := make(map[string]bool)
	for _, ev := range s.EventHistory {
		idSet[ev.ID()] = true
	}
	for i := 0; i < n; i++ {
		assert.True(t, idSet[generateUUID(i)], "expected event %d to be present", i)
	}
}

// generateUUID creates a deterministic UUID-like string for test event IDs.
func generateUUID(index int) string {
	return fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", index)
}
