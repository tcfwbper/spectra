package spectra

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirectoryCreator creates the .spectra/ and spec/ directory structures
// required for a Spectra project.
type DirectoryCreator struct{}

// NewDirectoryCreator returns a new DirectoryCreator instance.
func NewDirectoryCreator() *DirectoryCreator {
	return &DirectoryCreator{}
}

// CreateAll creates all required project directories under projectRoot.
// Directories are created in order with permissions 0755.
// If a directory already exists, it is silently skipped.
// If creation fails, returns an error immediately (fail-fast).
func (d *DirectoryCreator) CreateAll(projectRoot string) error {
	dirs := []string{
		".spectra",
		".spectra/sessions",
		".spectra/workflows",
		".spectra/agents",
		"spec",
		"spec/logic",
		"spec/test",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(projectRoot, dir)
		info, err := os.Stat(fullPath)
		if err == nil {
			// Path exists
			if info.IsDir() {
				continue
			}
			// Exists but is not a directory — Mkdir will fail
		}

		if err := os.Mkdir(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", dir, err)
		}
	}

	return nil
}
