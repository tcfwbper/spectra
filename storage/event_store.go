package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities"
)

// EventStore manages persistent storage of event history for a single session.
type EventStore struct {
	projectRoot string
	sessionUUID uuid.UUID
	eventsPath  string
}

// NewEventStore creates a new EventStore for the given session.
func NewEventStore(projectRoot string, sessionUUID uuid.UUID) *EventStore {
	eventsPath := GetEventHistoryPath(projectRoot, sessionUUID.String())
	return &EventStore{
		projectRoot: projectRoot,
		sessionUUID: sessionUUID,
		eventsPath:  eventsPath,
	}
}

// Append writes a new event to the events.jsonl file.
func (es *EventStore) Append(event *entities.Event) error {
	// Serialize event to compact JSON with Message field last
	jsonData, err := es.serializeEvent(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Check size limit (10 MB)
	if len(jsonData) > 10*1024*1024 {
		return fmt.Errorf("event size exceeds 10 MB limit: %d bytes", len(jsonData))
	}

	// Ensure line ends with newline
	if !bytes.HasSuffix(jsonData, []byte("\n")) {
		jsonData = append(jsonData, '\n')
	}

	// Prepare callback: check session directory exists and create file if needed
	prepareCallback := func() error {
		sessionDir := GetSessionDir(es.projectRoot, es.sessionUUID.String())
		if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
			return fmt.Errorf("session directory does not exist: %s", sessionDir)
		} else if err != nil {
			return err
		}

		// Create empty file with proper permissions
		f, err := os.OpenFile(es.eventsPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		f.Close()
		return nil
	}

	// Use FileAccessor to ensure file exists
	_, err = FileAccessor(es.eventsPath, prepareCallback)
	if err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}

	// Open file for appending with exclusive lock
	file, err := os.OpenFile(es.eventsPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	defer file.Close()

	// Acquire exclusive lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Write event to file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	// Flush to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// Read reads all events from the events.jsonl file.
func (es *EventStore) Read() ([]*entities.Event, error) {
	// Check if file exists
	if _, err := os.Stat(es.eventsPath); os.IsNotExist(err) {
		// File doesn't exist yet, return empty list
		return []*entities.Event{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}

	// Open file for reading with shared lock
	file, err := os.Open(es.eventsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}
	defer file.Close()

	// Acquire shared lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_SH)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Read file line by line
	scanner := bufio.NewScanner(file)
	events := []*entities.Event{}
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip blank lines
		if line == "" {
			log.Printf("Warning: skipping malformed event at line %d: blank line", lineNum)
			continue
		}

		// Parse JSON
		var event entities.Event
		err := json.Unmarshal([]byte(line), &event)
		if err != nil {
			log.Printf("Warning: skipping malformed event at line %d: %v", lineNum, err)
			continue
		}

		// Validate required fields
		if event.ID == uuid.Nil {
			log.Printf("Warning: skipping malformed event at line %d: missing required field 'ID'", lineNum)
			continue
		}

		events = append(events, &event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}

	return events, nil
}

// serializeEvent serializes an event to compact JSON with Message field last.
func (es *EventStore) serializeEvent(event *entities.Event) ([]byte, error) {
	// Custom serialization to ensure Message is last
	// We'll build a map manually and use json.Marshal with custom ordering

	// Create a buffer to write JSON manually
	var buf bytes.Buffer
	buf.WriteString("{")

	// Write fields in order (Message last)
	buf.WriteString(`"ID": "`)
	buf.WriteString(event.ID.String())
	buf.WriteString(`", `)

	buf.WriteString(`"Type": "`)
	buf.WriteString(escapeJSON(event.Type))
	buf.WriteString(`", `)

	// Payload (raw JSON, already marshaled)
	buf.WriteString(`"Payload": `)
	if len(event.Payload) == 0 {
		buf.WriteString("{}")
	} else {
		buf.Write(event.Payload)
	}
	buf.WriteString(`, `)

	buf.WriteString(`"EmittedBy": "`)
	buf.WriteString(escapeJSON(event.EmittedBy))
	buf.WriteString(`", `)

	fmt.Fprintf(&buf, `"EmittedAt": %d, `, event.EmittedAt)

	buf.WriteString(`"SessionID": "`)
	buf.WriteString(event.SessionID.String())
	buf.WriteString(`", `)

	// Message last
	buf.WriteString(`"Message": "`)
	buf.WriteString(escapeJSON(event.Message))
	buf.WriteString(`"`)

	buf.WriteString("}")

	return buf.Bytes(), nil
}

// escapeJSON escapes a string for JSON encoding.
func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	// Remove surrounding quotes
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return string(b[1 : len(b)-1])
	}
	return string(b)
}
