package session

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// UpdateSessionDataSafe updates session data with thread safety and validation.
func (s *Session) UpdateSessionDataSafe(key string, value any) error {
	// Validate key is non-empty
	if key == "" {
		return fmt.Errorf("session data key cannot be empty")
	}

	// Validate ClaudeSessionID type if applicable
	if strings.HasSuffix(key, ".ClaudeSessionID") {
		if value == nil {
			return fmt.Errorf("ClaudeSessionID value must be a string, got nil")
		}
		// Check dynamic type using type assertion
		if _, ok := value.(string); !ok {
			typeKind := reflect.TypeOf(value).Kind().String()
			return fmt.Errorf("ClaudeSessionID value must be a string, got %s", typeKind)
		}
	}

	// Acquire write lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update in-memory state
	s.SessionData[key] = value
	s.UpdatedAt = time.Now().Unix()

	// Attempt persistence (best-effort)
	if err := s.metadataStore.Write(s.SessionMetadata); err != nil {
		s.logger.Warning(fmt.Sprintf("UpdateSessionDataSafe persistence failed: %v", err))
	}

	return nil
}

// GetSessionDataSafe retrieves session data with thread safety.
func (s *Session) GetSessionDataSafe(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.SessionData[key]
	return value, ok
}
