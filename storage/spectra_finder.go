package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSpectraRoot locates the .spectra directory by searching upward from startDir.
// If startDir is empty, the current working directory is used.
// Returns the absolute path to the project root (directory containing .spectra) or an error.
func FindSpectraRoot(startDir string) (string, error) {
	var dir string

	if startDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		dir = wd
	} else {
		absDir, err := filepath.Abs(startDir)
		if err != nil {
			return "", fmt.Errorf("invalid start directory: %s: %w", startDir, err)
		}
		dir = absDir
	}

	// Validate that the start directory exists and is a directory.
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsPermission(err) {
			// Permission denied accessing the start directory means we cannot traverse.
			return "", ErrNotInitialized
		}
		return "", fmt.Errorf("invalid start directory: %s: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("invalid start directory: %s: not a directory", dir)
	}

	// Track visited paths to detect symlink loops.
	visited := make(map[string]bool)

	for {
		// Resolve symlinks for loop detection.
		resolved, err := filepath.EvalSymlinks(dir)
		if err != nil {
			// Permission error or broken symlink; stop traversal.
			return "", ErrNotInitialized
		}

		if visited[resolved] {
			// Symlink loop detected.
			return "", ErrNotInitialized
		}
		visited[resolved] = true

		// Check if .spectra exists in this directory and is a directory.
		spectraPath := filepath.Join(dir, SpectraDir)
		info, err := os.Stat(spectraPath)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		// Move to parent directory.
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return "", ErrNotInitialized
		}

		// Check if we can access the parent.
		_, err = os.Stat(parent)
		if err != nil {
			// Permission denied or other access error.
			return "", ErrNotInitialized
		}

		dir = parent
	}
}
