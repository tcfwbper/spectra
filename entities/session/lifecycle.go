package session

import (
	"fmt"
	"time"
)

// Run transitions the session from "initializing" to "running".
func (s *Session) Run(terminationNotifier chan<- struct{}) error {
	s.mu.Lock()

	// Validate status precondition
	if s.Status != "initializing" {
		s.mu.Unlock()
		return fmt.Errorf("cannot run session: status is '%s', expected 'initializing'", s.Status)
	}

	// Update in-memory state
	s.Status = "running"
	s.UpdatedAt = time.Now().Unix()

	// Release lock before persistence attempt
	s.mu.Unlock()

	// Attempt persistence (best-effort)
	if err := s.metadataStore.Write(s.SessionMetadata); err != nil {
		s.logger.Warning(fmt.Sprintf("Run persistence failed: %v", err))
	}

	return nil
}

// Done transitions the session from "running" to "completed".
func (s *Session) Done(terminationNotifier chan<- struct{}) error {
	s.mu.Lock()

	// Validate status precondition
	if s.Status != "running" {
		s.mu.Unlock()
		return fmt.Errorf("cannot complete session: status is '%s', expected 'running'", s.Status)
	}

	// Update in-memory state
	s.Status = "completed"
	s.UpdatedAt = time.Now().Unix()

	// Release lock before persistence attempt
	s.mu.Unlock()

	// Attempt persistence (best-effort)
	if err := s.metadataStore.Write(s.SessionMetadata); err != nil {
		s.logger.Warning(fmt.Sprintf("Done persistence failed: %v", err))
	}

	// Send termination notification (non-blocking)
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}

	return nil
}

// Fail transitions the session to "failed" status.
func (s *Session) Fail(err error, terminationNotifier chan<- struct{}) error {
	// Validate error is not nil (including nil typed pointers)
	if err == nil {
		return fmt.Errorf("error cannot be nil")
	}

	// Validate error type before acquiring lock and check for nil typed pointers
	switch e := err.(type) {
	case *AgentError:
		if e == nil {
			return fmt.Errorf("error cannot be nil")
		}
	case *RuntimeError:
		if e == nil {
			return fmt.Errorf("error cannot be nil")
		}
	default:
		return fmt.Errorf("invalid error type: must be *AgentError or *RuntimeError")
	}

	s.mu.Lock()

	// Validate status preconditions
	if s.Status == "failed" {
		s.mu.Unlock()
		return fmt.Errorf("session already failed")
	}
	if s.Status == "completed" {
		s.mu.Unlock()
		return fmt.Errorf("cannot fail session: status is 'completed', workflow already terminated")
	}

	// Update in-memory state atomically
	s.Status = "failed"
	s.Error = err
	s.UpdatedAt = time.Now().Unix()

	// Release lock before persistence attempt
	s.mu.Unlock()

	// Attempt persistence (best-effort)
	if persistErr := s.metadataStore.Write(s.SessionMetadata); persistErr != nil {
		s.logger.Warning(fmt.Sprintf("Fail persistence failed: %v", persistErr))
	}

	// Send termination notification (non-blocking)
	select {
	case terminationNotifier <- struct{}{}:
	default:
	}

	return nil
}
