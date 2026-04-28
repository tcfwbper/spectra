package storage

import (
	"fmt"
	"os"
)

// FileAccessor checks if a file exists and invokes a callback to prepare it if not.
// It returns the file path if the file exists or was successfully created by the callback.
func FileAccessor(filePath string, prepareCallback func() error) (string, error) {
	// Check if file exists
	_, err := os.Stat(filePath)
	if err == nil {
		// File exists, return path immediately without invoking callback
		return filePath, nil
	}

	// If error is not "not exist", return the error immediately
	if !os.IsNotExist(err) {
		// Permission denied or other stat error
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied: %s", filePath)
		}
		return "", err
	}

	// File does not exist, invoke callback to prepare it
	if prepareCallback == nil {
		panic("FileAccessor: prepareCallback is nil")
	}

	err = prepareCallback()
	if err != nil {
		// Callback failed, wrap error with context
		return "", fmt.Errorf("failed to prepare file %s: %w", filePath, err)
	}

	// Callback succeeded, verify file now exists
	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Callback returned nil but file was not created
			return "", fmt.Errorf("file preparation succeeded but file was not created: %s", filePath)
		}
		// Other error (e.g., permission denied after callback)
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied: %s", filePath)
		}
		return "", err
	}

	// File exists after callback
	return filePath, nil
}
