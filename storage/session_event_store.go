package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities/session"
)

// SessionEventStore persists session.Event records to the events.jsonl file.
type SessionEventStore struct {
	projectRoot string
	sessionUUID uuid.UUID
	eventsPath  string
}

// NewSessionEventStore creates a new SessionEventStore for the given session.
func NewSessionEventStore(projectRoot string, sessionUUID uuid.UUID) *SessionEventStore {
	eventsPath := GetEventHistoryPath(projectRoot, sessionUUID.String())
	return &SessionEventStore{
		projectRoot: projectRoot,
		sessionUUID: sessionUUID,
		eventsPath:  eventsPath,
	}
}

// WriteEvent appends a session.Event to the events.jsonl file.
func (s *SessionEventStore) WriteEvent(event session.Event) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}
	jsonData = append(jsonData, '\n')

	// Prepare callback: ensure session directory and file exist
	prepareCallback := func() error {
		sessionDir := GetSessionDir(s.projectRoot, s.sessionUUID.String())
		if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
			return fmt.Errorf("session directory does not exist: %s", sessionDir)
		} else if err != nil {
			return err
		}
		f, err := os.OpenFile(s.eventsPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		return f.Close()
	}

	_, err = FileAccessor(s.eventsPath, prepareCallback)
	if err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}

	file, err := os.OpenFile(s.eventsPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	defer file.Close()

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN) //nolint:errcheck

	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}
