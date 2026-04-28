package session

// GetStatusSafe returns the current status with thread safety.
func (s *Session) GetStatusSafe() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// GetCurrentStateSafe returns the current state with thread safety.
func (s *Session) GetCurrentStateSafe() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentState
}

// GetErrorSafe returns the error with thread safety.
func (s *Session) GetErrorSafe() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Error
}
