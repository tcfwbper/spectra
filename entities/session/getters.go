package session

// GetStatusSafe returns the current status under the read lock.
func (s *Session) GetStatusSafe() string {
	s.mu.RLock()
	status := s.Status
	s.mu.RUnlock()
	return status
}

// GetCurrentStateSafe returns the current state under the read lock.
func (s *Session) GetCurrentStateSafe() string {
	s.mu.RLock()
	state := s.CurrentState
	s.mu.RUnlock()
	return state
}

// GetErrorSafe returns the current error under the read lock.
func (s *Session) GetErrorSafe() error {
	s.mu.RLock()
	err := s.Error
	s.mu.RUnlock()
	return err
}

// GetMetadataSnapshotSafe returns a detached copy of SessionMetadata under the
// read lock. The returned map is shallow-copied for isolation.
func (s *Session) GetMetadataSnapshotSafe() SessionMetadata {
	s.mu.RLock()
	snapshot := s.SessionMetadata
	// Shallow-copy the SessionData map for isolation.
	dataCopy := make(map[string]any, len(s.SessionData))
	for k, v := range s.SessionData {
		dataCopy[k] = v
	}
	snapshot.SessionData = dataCopy
	s.mu.RUnlock()
	return snapshot
}
