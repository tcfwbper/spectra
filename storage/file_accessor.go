package storage

import (
	"errors"
	"fmt"
	"os"
)

// FileAccessor checks file existence and invokes the preparation callback if the
// file does not exist. It returns the file path on success or an error.
func FileAccessor(filePath string, prepare func() error) (string, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		// File exists; return immediately without invoking callback.
		return filePath, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		// Stat failed with an error other than ErrNotExist (e.g., permission denied).
		return "", err
	}

	// File does not exist; invoke preparation callback.
	if cbErr := prepare(); cbErr != nil {
		return "", fmt.Errorf("failed to prepare file %s: %w", filePath, cbErr)
	}

	// Post-callback verification.
	_, err = os.Stat(filePath)
	if err == nil {
		return filePath, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("file preparation succeeded but file was not created: %s", filePath)
	}

	// Post-callback stat failed with non-ErrNotExist error (e.g., permission denied).
	return "", err
}
