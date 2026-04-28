package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// SpectraFinder searches upward from startDir to find a .spectra directory.
// If startDir is empty, it uses the current working directory.
// Returns the absolute path to the project root containing .spectra.
func SpectraFinder(startDir string) (string, error) {
	// If startDir is empty, use current working directory
	if startDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("spectra not initialized")
		}
		startDir = cwd
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("spectra not initialized")
	}

	// Try to resolve symlinks in the start directory first
	// This will fail for symlink loops or non-existent paths
	realStartPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If it's a "not exist" error, report invalid start directory
		// Otherwise (symlink loop, permission denied, etc), report not initialized
		if os.IsNotExist(err) {
			return "", fmt.Errorf("invalid start directory: %s", startDir)
		}
		return "", fmt.Errorf("spectra not initialized")
	}

	// Validate that the resolved path is a directory
	info, err := os.Stat(realStartPath)
	if err != nil {
		return "", fmt.Errorf("invalid start directory: %s", startDir)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("invalid start directory: %s", startDir)
	}

	// Track visited directories to detect symlink loops
	visited := make(map[string]bool)

	currentDir := realStartPath
	for {
		// Check for symlink loop
		realPath, err := filepath.EvalSymlinks(currentDir)
		if err != nil {
			// Error evaluating symlinks, treat as not found
			return "", fmt.Errorf("spectra not initialized")
		}

		if visited[realPath] {
			// Symlink loop detected
			return "", fmt.Errorf("spectra not initialized")
		}
		visited[realPath] = true

		// Check if .spectra exists in current directory
		spectraPath := filepath.Join(currentDir, SpectraDir)
		info, err := os.Stat(spectraPath)
		if err == nil && info.IsDir() {
			// Found .spectra directory
			return currentDir, nil
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the filesystem root
		if parentDir == currentDir {
			// Reached root without finding .spectra
			return "", fmt.Errorf("spectra not initialized")
		}

		currentDir = parentDir
	}
}
