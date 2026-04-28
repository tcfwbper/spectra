package session

import (
	"fmt"
	"time"
)

// UpdateCurrentStateSafe updates the current state with thread safety.
// Returns nil on success (always); errors are logged as warnings.
func (s *Session) UpdateCurrentStateSafe(newState string) error {
	// Validate newState is non-empty before acquiring lock
	if newState == "" {
		s.logger.Warning("UpdateCurrentStateSafe called with empty newState; in-memory state unchanged")
		return nil
	}

	// Acquire write lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update in-memory state
	s.CurrentState = newState
	s.UpdatedAt = time.Now().Unix()

	// Attempt persistence (best-effort)
	if err := s.metadataStore.Write(s.SessionMetadata); err != nil {
		s.logger.Warning(fmt.Sprintf("UpdateCurrentStateSafe persistence failed: %v", err))
	}

	return nil
}
