package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/tcfwbper/spectra/entities"
	"github.com/tcfwbper/spectra/logger"
)

// EventStore manages persistent storage of event history for a single session.
// It appends events to a JSONL file in compact format.
type EventStore struct {
	filePath string
	logger   logger.Logger
}

// NewEventStore composes the events.jsonl path via StorageLayout and stores it
// internally. No I/O is performed at construction time.
func NewEventStore(projectRoot, sessionUUID string, log logger.Logger) *EventStore {
	return &EventStore{
		filePath: GetEventHistoryPath(projectRoot, sessionUUID),
		logger:   log,
	}
}

// Append serializes the event to compact JSON and appends it to the events.jsonl file.
func (es *EventStore) Append(event *entities.Event) error {
	// Call FileAccessor with the preparation callback.
	_, err := FileAccessor(es.filePath, func() error {
		parentDir := filepath.Dir(es.filePath)
		info, statErr := os.Stat(parentDir)
		if statErr != nil || !info.IsDir() {
			return fmt.Errorf("session directory does not exist: %s", parentDir)
		}
		// Create an empty events.jsonl file with permissions 0644.
		f, createErr := os.OpenFile(es.filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if createErr != nil {
			return createErr
		}
		return f.Close()
	})
	if err != nil {
		return err
	}

	// Serialize the event to compact JSON with Message last.
	data, err := es.serializeEvent(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Check size limit.
	if len(data) > MaxPayloadSize {
		return fmt.Errorf("event size exceeds limit: %d bytes (max %d bytes)", len(data), MaxPayloadSize)
	}

	// Open file for appending with exclusive lock.
	f, err := os.OpenFile(es.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire exclusive file-level lock.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	// Write the JSON line followed by newline.
	line := append(data, '\n')
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	// Flush the file buffer.
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// Read reads all events from the events.jsonl file, logging warnings for malformed lines.
func (es *EventStore) Read() ([]*entities.Event, error) {
	// If file does not exist, return empty list without error.
	if _, err := os.Stat(es.filePath); os.IsNotExist(err) {
		return []*entities.Event{}, nil
	}

	f, err := os.Open(es.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire shared read lock.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return nil, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	var events []*entities.Event
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip blank lines with a warning.
		if line == "" {
			es.logger.Warn("skipping malformed event line", "line", lineNum, "error", "empty line")
			continue
		}

		// Parse JSON into a raw map.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			es.logger.Warn("skipping malformed event line", "line", lineNum, "error", fmt.Sprintf("invalid JSON: %s", err.Error()))
			continue
		}

		// Extract required fields.
		ev, err := es.parseEventFromRaw(raw)
		if err != nil {
			es.logger.Warn("skipping malformed event line", "line", lineNum, "error", fmt.Sprintf("invalid event: %s", err.Error()))
			continue
		}

		events = append(events, ev)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}

	return events, nil
}

// serializeEvent serializes the event to compact JSON with Message as the last field.
func (es *EventStore) serializeEvent(event *entities.Event) ([]byte, error) {
	// Build ordered map manually to ensure Message is last.
	type orderedEvent struct {
		ID        string          `json:"ID"`
		Type      string          `json:"Type"`
		Payload   json.RawMessage `json:"Payload"`
		EmittedBy string          `json:"EmittedBy"`
		EmittedAt int64           `json:"EmittedAt"`
		SessionID string          `json:"SessionID"`
		Message   string          `json:"Message"`
	}

	oe := orderedEvent{
		ID:        event.ID(),
		Type:      event.Type(),
		Payload:   event.Payload(),
		EmittedBy: event.EmittedBy(),
		EmittedAt: event.EmittedAt(),
		SessionID: event.SessionID(),
		Message:   event.Message(),
	}

	return json.Marshal(oe)
}

// parseEventFromRaw parses event fields from a raw JSON map and reconstructs via NewEvent.
func (es *EventStore) parseEventFromRaw(raw map[string]json.RawMessage) (*entities.Event, error) {
	var id, eventType, message, emittedBy, sessionID string
	var emittedAt int64
	var payload json.RawMessage

	if v, ok := raw["ID"]; ok {
		if err := json.Unmarshal(v, &id); err != nil {
			return nil, fmt.Errorf("invalid ID: %w", err)
		}
	}
	if v, ok := raw["Type"]; ok {
		if err := json.Unmarshal(v, &eventType); err != nil {
			return nil, fmt.Errorf("invalid Type: %w", err)
		}
	}
	if v, ok := raw["Message"]; ok {
		if err := json.Unmarshal(v, &message); err != nil {
			return nil, fmt.Errorf("invalid Message: %w", err)
		}
	}
	if v, ok := raw["Payload"]; ok {
		payload = v
	}
	if v, ok := raw["EmittedBy"]; ok {
		if err := json.Unmarshal(v, &emittedBy); err != nil {
			return nil, fmt.Errorf("invalid EmittedBy: %w", err)
		}
	}
	if v, ok := raw["EmittedAt"]; ok {
		if err := json.Unmarshal(v, &emittedAt); err != nil {
			return nil, fmt.Errorf("invalid EmittedAt: %w", err)
		}
	}
	if v, ok := raw["SessionID"]; ok {
		if err := json.Unmarshal(v, &sessionID); err != nil {
			return nil, fmt.Errorf("invalid SessionID: %w", err)
		}
	}

	return entities.NewEvent(id, eventType, message, payload, emittedBy, emittedAt, sessionID)
}
