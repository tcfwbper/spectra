package storage

import (
	"fmt"
	"os"
)

// EnsureSessionsDirectory ensures that .spectra/sessions/ exists.
// Returns ErrNotInitialized if .spectra/ does not exist or is not a directory.
func EnsureSessionsDirectory(projectRoot string) error {
	spectraDir := GetSpectraDir(projectRoot)

	info, err := os.Stat(spectraDir)
	if err != nil {
		return ErrNotInitialized
	}
	if !info.IsDir() {
		return ErrNotInitialized
	}

	sessionsDir := GetSessionsDir(projectRoot)

	info, err = os.Stat(sessionsDir)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		// sessions exists but is not a directory - treat as needing creation
	}

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	if mkErr := os.Mkdir(sessionsDir, 0755); mkErr != nil {
		return fmt.Errorf("failed to create sessions directory: %w", mkErr)
	}

	return nil
}

// CreateSessionDirectory creates a new session directory at .spectra/sessions/<uuid>/.
// It calls EnsureSessionsDirectory internally to guarantee the parent exists.
// Returns ErrSessionDirExists if the session directory already exists.
func CreateSessionDirectory(projectRoot, sessionUUID string) error {
	if err := EnsureSessionsDirectory(projectRoot); err != nil {
		return err
	}

	sessionDir := GetSessionDir(projectRoot, sessionUUID)

	_, err := os.Stat(sessionDir)
	if err == nil {
		// Directory already exists.
		return ErrSessionDirExists
	}

	if mkErr := os.Mkdir(sessionDir, 0755); mkErr != nil {
		if os.IsExist(mkErr) {
			return ErrSessionDirExists
		}
		return fmt.Errorf("failed to create session directory: %w", mkErr)
	}

	return nil
}
