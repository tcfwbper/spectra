package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/entities/session"
)

// SessionMetadataStore manages persistent storage of SessionMetadata for a single session.
// It reads and writes the session.json file in pretty-printed JSON format.
type SessionMetadataStore struct {
	filePath string
}

// NewSessionMetadataStore composes the session.json path via StorageLayout and stores it
// internally. No I/O is performed at construction time.
func NewSessionMetadataStore(projectRoot, sessionUUID string) *SessionMetadataStore {
	return &SessionMetadataStore{
		filePath: GetSessionMetadataPath(projectRoot, sessionUUID),
	}
}

// Write serializes the session metadata to pretty-printed JSON and writes it to session.json.
func (s *SessionMetadataStore) Write(meta session.SessionMetadata) error {
	// Call FileAccessor with the preparation callback.
	_, err := FileAccessor(s.filePath, func() error {
		parentDir := filepath.Dir(s.filePath)
		info, statErr := os.Stat(parentDir)
		if statErr != nil || !info.IsDir() {
			return fmt.Errorf("session directory does not exist: %s", parentDir)
		}
		// Create an empty session.json file with permissions 0644.
		f, createErr := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if createErr != nil {
			return createErr
		}
		return f.Close()
	})
	if err != nil {
		return err
	}

	// Serialize the metadata.
	data, err := s.serializeMetadata(meta)
	if err != nil {
		return fmt.Errorf("failed to serialize session metadata: %w", err)
	}

	// Check size limit.
	if len(data) > MaxPayloadSize {
		return fmt.Errorf("session metadata size exceeds limit: %d bytes (max %d bytes)", len(data), MaxPayloadSize)
	}

	// Open file for writing with exclusive lock (truncate).
	f, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire exclusive file-level lock.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	// Write the JSON content.
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}

	// Flush the file buffer.
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}

	return nil
}

// Read reads session metadata from the session.json file.
func (s *SessionMetadataStore) Read() (session.SessionMetadata, error) {
	var zero session.SessionMetadata

	// If file does not exist, return an error.
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return zero, fmt.Errorf("session metadata file does not exist: %s", s.filePath)
	}

	f, err := os.Open(s.filePath)
	if err != nil {
		return zero, fmt.Errorf("failed to read session metadata file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire shared read lock.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return zero, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	// Read file content.
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return zero, fmt.Errorf("failed to read session metadata file: %w", err)
	}

	// Parse into raw map first.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %s", err.Error())
	}

	// Validate required fields.
	requiredFields := []string{"id", "workflowName", "status", "createdAt", "updatedAt", "currentState", "sessionData"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			return zero, fmt.Errorf("failed to parse session metadata: missing required field '%s'", field)
		}
	}

	// Parse basic fields.
	var meta session.SessionMetadata
	if err := json.Unmarshal(raw["id"], &meta.ID); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}
	if err := json.Unmarshal(raw["workflowName"], &meta.WorkflowName); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}
	if err := json.Unmarshal(raw["status"], &meta.Status); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}
	if err := json.Unmarshal(raw["createdAt"], &meta.CreatedAt); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}
	if err := json.Unmarshal(raw["updatedAt"], &meta.UpdatedAt); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}
	if err := json.Unmarshal(raw["currentState"], &meta.CurrentState); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}
	if err := json.Unmarshal(raw["sessionData"], &meta.SessionData); err != nil {
		return zero, fmt.Errorf("failed to parse session metadata: %w", err)
	}

	// Parse error field if present.
	if errRaw, ok := raw["error"]; ok && string(errRaw) != "null" {
		reconstructed, err := s.reconstructError(errRaw)
		if err != nil {
			return zero, err
		}
		meta.Error = reconstructed
	}

	return meta, nil
}

// serializeMetadata serializes SessionMetadata to pretty-printed JSON.
func (s *SessionMetadataStore) serializeMetadata(meta session.SessionMetadata) ([]byte, error) {
	// Build the output map manually to control field ordering and error omitempty.
	output := map[string]any{
		"id":           meta.ID,
		"workflowName": meta.WorkflowName,
		"status":       meta.Status,
		"createdAt":    meta.CreatedAt,
		"updatedAt":    meta.UpdatedAt,
		"currentState": meta.CurrentState,
		"sessionData":  meta.SessionData,
	}

	// Only include error if non-nil.
	if meta.Error != nil {
		errObj, err := s.serializeError(meta.Error)
		if err != nil {
			return nil, err
		}
		output["error"] = errObj
	}

	return json.MarshalIndent(output, "", "  ")
}

// serializeError serializes the error entity using getter methods.
func (s *SessionMetadataStore) serializeError(sessionErr error) (map[string]any, error) {
	switch e := sessionErr.(type) {
	case *entities.AgentError:
		obj := map[string]any{
			"agentRole":    e.AgentRole(),
			"message":      e.Message(),
			"occurredAt":   e.OccurredAt(),
			"sessionID":    e.SessionID(),
			"failingState": e.FailingState(),
		}
		if e.Detail() != nil {
			var detail any
			if err := json.Unmarshal(e.Detail(), &detail); err != nil {
				return nil, err
			}
			obj["detail"] = detail
		}
		return obj, nil
	case *entities.RuntimeError:
		obj := map[string]any{
			"issuer":       e.Issuer(),
			"message":      e.Message(),
			"occurredAt":   e.OccurredAt(),
			"sessionID":    e.SessionID(),
			"failingState": e.FailingState(),
		}
		if e.Detail() != nil {
			var detail any
			if err := json.Unmarshal(e.Detail(), &detail); err != nil {
				return nil, err
			}
			obj["detail"] = detail
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("unsupported error type: %T", sessionErr)
	}
}

// reconstructError reconstructs an AgentError or RuntimeError from a raw JSON error object.
func (s *SessionMetadataStore) reconstructError(raw json.RawMessage) (error, error) {
	var errMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &errMap); err != nil {
		return nil, fmt.Errorf("failed to reconstruct error: %w", err)
	}

	_, hasAgentRole := errMap["agentRole"]
	_, hasIssuer := errMap["issuer"]

	if hasAgentRole && hasIssuer {
		return nil, fmt.Errorf("failed to reconstruct error: ambiguous error object contains both 'agentRole' and 'issuer'")
	}
	if !hasAgentRole && !hasIssuer {
		return nil, fmt.Errorf("failed to reconstruct error: cannot determine error type — missing 'agentRole' or 'issuer' field")
	}

	// Extract common fields.
	var message, sessionID, failingState string
	var occurredAt int64
	var detail json.RawMessage

	if v, ok := errMap["message"]; ok {
		if err := json.Unmarshal(v, &message); err != nil {
			return nil, fmt.Errorf("failed to reconstruct error: invalid message: %w", err)
		}
	}
	if v, ok := errMap["occurredAt"]; ok {
		if err := json.Unmarshal(v, &occurredAt); err != nil {
			return nil, fmt.Errorf("failed to reconstruct error: invalid occurredAt: %w", err)
		}
	}
	if v, ok := errMap["sessionID"]; ok {
		if err := json.Unmarshal(v, &sessionID); err != nil {
			return nil, fmt.Errorf("failed to reconstruct error: invalid sessionID: %w", err)
		}
	}
	if v, ok := errMap["failingState"]; ok {
		if err := json.Unmarshal(v, &failingState); err != nil {
			return nil, fmt.Errorf("failed to reconstruct error: invalid failingState: %w", err)
		}
	}
	if v, ok := errMap["detail"]; ok && string(v) != "null" {
		// Re-marshal the detail to json.RawMessage (it's already valid JSON).
		detail = v
	}

	if hasAgentRole {
		var agentRole string
		if err := json.Unmarshal(errMap["agentRole"], &agentRole); err != nil {
			return nil, fmt.Errorf("failed to reconstruct error: invalid agentRole: %w", err)
		}
		ae, err := entities.NewAgentError(agentRole, message, detail, occurredAt, sessionID, failingState)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct error: %w", err)
		}
		return ae, nil
	}

	// hasIssuer
	var issuer string
	if err := json.Unmarshal(errMap["issuer"], &issuer); err != nil {
		return nil, fmt.Errorf("failed to reconstruct error: invalid issuer: %w", err)
	}
	re, err := entities.NewRuntimeError(issuer, message, detail, occurredAt, sessionID, failingState)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct error: %w", err)
	}
	return re, nil
}
