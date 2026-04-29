package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/tcfwbper/spectra/entities"
)

// SessionMetadataStore manages persistent storage of SessionMetadata for a single session.
type SessionMetadataStore struct {
	projectRoot  string
	sessionUUID  uuid.UUID
	metadataPath string
}

// NewSessionMetadataStore creates a new SessionMetadataStore for the given session.
func NewSessionMetadataStore(projectRoot string, sessionUUID uuid.UUID) *SessionMetadataStore {
	metadataPath := GetSessionMetadataPath(projectRoot, sessionUUID.String())
	return &SessionMetadataStore{
		projectRoot:  projectRoot,
		sessionUUID:  sessionUUID,
		metadataPath: metadataPath,
	}
}

// Write writes session metadata to the session.json file.
func (s *SessionMetadataStore) Write(metadata *entities.SessionMetadata) error {
	// Update the UpdatedAt field to current timestamp
	metadata.UpdatedAt = time.Now().Unix()

	// Serialize metadata to pretty-printed JSON with 2-space indentation
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize session metadata: %w", err)
	}

	// Check size limit (10 MB)
	if len(jsonData) > 10*1024*1024 {
		return fmt.Errorf("session metadata size exceeds 10 MB limit: %d bytes", len(jsonData))
	}

	// Prepare callback: check session directory exists and create file if needed
	prepareCallback := func() error {
		sessionDir := GetSessionDir(s.projectRoot, s.sessionUUID.String())
		if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
			return fmt.Errorf("session directory does not exist: %s", sessionDir)
		} else if err != nil {
			return err
		}

		// Create empty file with proper permissions
		f, err := os.OpenFile(s.metadataPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		return nil
	}

	// Use FileAccessor to ensure file exists
	_, err = FileAccessor(s.metadataPath, prepareCallback)
	if err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}

	// Open file for writing (without truncate yet)
	file, err := os.OpenFile(s.metadataPath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}
	defer file.Close()

	// Acquire exclusive lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Truncate file after acquiring lock
	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}

	// Seek to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}

	// Write metadata to file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}

	// Flush to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to write session metadata: %w", err)
	}

	return nil
}

// Read reads session metadata from the session.json file.
func (s *SessionMetadataStore) Read() (*entities.SessionMetadata, error) {
	// Check if file exists
	if _, err := os.Stat(s.metadataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("session metadata file does not exist: %s", s.metadataPath)
	} else if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied")
		}
		return nil, fmt.Errorf("failed to read session metadata file: %w", err)
	}

	// Open file for reading with shared lock
	file, err := os.Open(s.metadataPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied")
		}
		return nil, fmt.Errorf("failed to read session metadata file: %w", err)
	}
	defer file.Close()

	// Acquire shared lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_SH)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Read file content
	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read session metadata file: %w", err)
	}

	var metadata entities.SessionMetadata
	err = json.Unmarshal(buf, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session metadata: %w", err)
	}

	// Validate required fields
	if metadata.ID == uuid.Nil {
		return nil, fmt.Errorf("failed to parse session metadata: missing required field 'ID'")
	}

	return &metadata, nil
}
