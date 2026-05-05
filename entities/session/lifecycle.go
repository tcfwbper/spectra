package session

import (
	"fmt"
	"time"

	"github.com/spectra-ai/spectra/entities"
)

// Run transitions the session from "initializing" to "running".
func (s *Session) Run() error {
	s.mu.Lock()
	if s.Status != "initializing" {
		s.mu.Unlock()
		return fmt.Errorf("cannot run session: status is '%s', expected 'initializing'", s.Status)
	}
	s.Status = "running"
	s.UpdatedAt = time.Now().Unix()
	s.mu.Unlock()
	return nil
}

// Done transitions the session from "running" to "completed" and notifies
// the runtime via terminationNotifier.
func (s *Session) Done(terminationNotifier chan struct{}) error {
	s.mu.Lock()
	if s.Status != "running" {
		status := s.Status
		s.mu.Unlock()
		return fmt.Errorf("cannot complete session: status is '%s', expected 'running'", status)
	}
	s.Status = "completed"
	s.UpdatedAt = time.Now().Unix()
	s.mu.Unlock()

	terminationNotifier <- struct{}{}
	return nil
}

// Fail transitions the session to "failed" from any non-terminal status,
// records the error, and notifies the runtime. The first error wins.
func (s *Session) Fail(err error, terminationNotifier chan struct{}) error {
	if err == nil {
		return fmt.Errorf("error cannot be nil")
	}

	// Validate error type before acquiring the lock.
	switch err.(type) {
	case *entities.AgentError:
	case *entities.RuntimeError:
	default:
		return fmt.Errorf("invalid error type: must be *AgentError or *RuntimeError")
	}

	s.mu.Lock()
	if s.Status == "failed" {
		s.mu.Unlock()
		return fmt.Errorf("session already failed")
	}
	if s.Status == "completed" {
		s.mu.Unlock()
		return fmt.Errorf("cannot fail session: status is 'completed', workflow already terminated")
	}

	s.Status = "failed"
	s.Error = err
	s.UpdatedAt = time.Now().Unix()
	s.mu.Unlock()

	terminationNotifier <- struct{}{}
	return nil
}
