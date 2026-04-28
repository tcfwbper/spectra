package helpers

import (
	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities"
)

// SessionMetadata represents session metadata for testing.
// This is a placeholder until the full Session entity is implemented.
// Once the actual Session entity exists in entities package, tests should
// import and use that instead.
type SessionMetadata struct {
	ID           uuid.UUID              `json:"id"`
	WorkflowName string                 `json:"workflowName"`
	Status       string                 `json:"status"`
	CreatedAt    int64                  `json:"createdAt"`
	UpdatedAt    int64                  `json:"updatedAt"`
	CurrentState string                 `json:"currentState"`
	SessionData  map[string]interface{} `json:"sessionData"`
	Error        *entities.AgentError   `json:"error,omitempty"`
	EventHistory []entities.Event       `json:"-"` // Excluded from serialization
}

// SplitLines splits content by newlines and filters out empty lines.
// Used for parsing JSONL files in tests.
func SplitLines(content string) []string {
	lines := []string{}
	current := ""
	for _, c := range content {
		if c == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
