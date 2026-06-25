package session

import (
	"fmt"
	"strings"
	"time"
)

// UpdateSessionDataSafe sets a key-value pair in SessionData under the write lock.
func (s *Session) UpdateSessionDataSafe(key string, value any) error {
	if key == "" {
		return fmt.Errorf("session data key cannot be empty")
	}

	// ClaudeSessionID type validation.
	if strings.HasSuffix(key, ".ClaudeSessionID") {
		if _, ok := value.(string); !ok {
			return fmt.Errorf("ClaudeSessionID value must be a string, got %T", value)
		}
	}

	// PID type validation.
	if strings.HasSuffix(key, ".PID") {
		if _, ok := value.(int); !ok {
			return fmt.Errorf("PID value must be an int, got %T", value)
		}
	}

	s.mu.Lock()
	s.SessionData[key] = value
	s.UpdatedAt = time.Now().Unix()
	s.mu.Unlock()
	return nil
}

// GetSessionDataSafe retrieves a value from SessionData under the read lock.
func (s *Session) GetSessionDataSafe(key string) (any, bool) {
	s.mu.RLock()
	val, ok := s.SessionData[key]
	s.mu.RUnlock()
	return val, ok
}
