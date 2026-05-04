package session

import (
	"fmt"
	"time"
)

// UpdateEventHistorySafe appends an event to the history with validation and thread safety.
func (s *Session) UpdateEventHistorySafe(event Event) error {
	// Validate required fields before acquiring lock
	if event.ID == "" {
		return fmt.Errorf("invalid event: ID is required")
	}
	if event.Type == "" {
		return fmt.Errorf("invalid event: Type is required")
	}
	if event.SessionID == "" {
		return fmt.Errorf("invalid event: SessionID is required")
	}
	if event.EmittedAt <= 0 {
		return fmt.Errorf("invalid event: EmittedAt is required")
	}
	if event.EmittedBy == "" {
		return fmt.Errorf("invalid event: EmittedBy is required")
	}

	// Acquire write lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Append event to history
	s.EventHistory = append(s.EventHistory, event)
	s.UpdatedAt = time.Now().Unix()

	// Attempt persistence to EventStore (best-effort)
	if s.eventStore != nil {
		if err := s.eventStore.WriteEvent(event); err != nil {
			if s.logger != nil {
				s.logger.Warning(fmt.Sprintf("UpdateEventHistorySafe: EventStore persistence failed: %v", err))
			}
		}
	}

	return nil
}
