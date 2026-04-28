package storage

import (
	"fmt"
	"os"
)

// SessionDirectoryManager manages session directory creation
type SessionDirectoryManager struct {
	projectRoot string
}

// NewSessionDirectoryManager creates a new SessionDirectoryManager
func NewSessionDirectoryManager(projectRoot string) *SessionDirectoryManager {
	return &SessionDirectoryManager{
		projectRoot: projectRoot,
	}
}

// CreateSessionDirectory creates a session directory with the given UUID
func (m *SessionDirectoryManager) CreateSessionDirectory(sessionUUID string) error {
	// Check if .spectra/sessions/ directory exists
	sessionsDir := GetSessionsDir(m.projectRoot)
	info, err := os.Stat(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sessions directory does not exist: %s. Run 'spectra init' to initialize the project", sessionsDir)
		}
		return fmt.Errorf("sessions directory does not exist: %s. Run 'spectra init' to initialize the project", sessionsDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("sessions directory does not exist: %s. Run 'spectra init' to initialize the project", sessionsDir)
	}

	// Get session directory path
	sessionDir := GetSessionDir(m.projectRoot, sessionUUID)

	// Create session directory with permissions 0775
	err = os.Mkdir(sessionDir, 0775)
	if err != nil {
		// Check if error is because directory already exists
		if os.IsExist(err) {
			return fmt.Errorf("session directory already exists: %s. This indicates a UUID collision or a previous session was not cleaned up properly. file exists", sessionDir)
		}
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	return nil
}
