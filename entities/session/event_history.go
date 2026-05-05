package session

import (
	"fmt"
	"time"

	"github.com/spectra-ai/spectra/entities"
)

// UpdateEventHistorySafe validates the event and appends it to EventHistory
// under the write lock.
func (s *Session) UpdateEventHistorySafe(event entities.Event) error {
	// Required-field validation before lock acquisition.
	if event.ID() == "" {
		return fmt.Errorf("invalid event: ID is required")
	}
	if event.Type() == "" {
		return fmt.Errorf("invalid event: Type is required")
	}
	if event.SessionID() == "" {
		return fmt.Errorf("invalid event: SessionID is required")
	}
	if event.EmittedAt() <= 0 {
		return fmt.Errorf("invalid event: EmittedAt is required")
	}
	if event.EmittedBy() == "" {
		return fmt.Errorf("invalid event: EmittedBy is required")
	}

	s.mu.Lock()
	s.EventHistory = append(s.EventHistory, event)
	s.UpdatedAt = time.Now().Unix()
	s.mu.Unlock()
	return nil
}
