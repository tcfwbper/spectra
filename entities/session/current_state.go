package session

import (
	"fmt"
	"time"
)

// UpdateCurrentStateSafe sets the CurrentState under the write lock.
func (s *Session) UpdateCurrentStateSafe(newState string) error {
	if newState == "" {
		return fmt.Errorf("current state cannot be empty")
	}

	s.mu.Lock()
	s.CurrentState = newState
	s.UpdatedAt = time.Now().Unix()
	s.mu.Unlock()
	return nil
}
