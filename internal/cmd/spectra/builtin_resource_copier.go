package spectra

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// StorageLayoutInterface defines the interface that BuiltinResourceCopier
// depends on for path composition.
type StorageLayoutInterface interface {
	GetWorkflowPath(projectRoot, name string) string
	GetAgentPath(projectRoot, name string) string
}

// BuiltinResourceCopier copies embedded built-in workflow, agent definition,
// and specification template files to their target locations.
type BuiltinResourceCopier struct {
	workflowsFS fs.FS
	agentsFS    fs.FS
	specFilesFS fs.FS
	layout      StorageLayoutInterface
}

// NewBuiltinResourceCopier returns a new BuiltinResourceCopier.
func NewBuiltinResourceCopier(workflowsFS, agentsFS, specFilesFS fs.FS, layout StorageLayoutInterface) *BuiltinResourceCopier {
	return &BuiltinResourceCopier{
		workflowsFS: workflowsFS,
		agentsFS:    agentsFS,
		specFilesFS: specFilesFS,
		layout:      layout,
	}
}

// CopyWorkflows copies embedded workflow files to the project's .spectra/workflows/ directory.
// Files that already exist are skipped with a warning.
func (b *BuiltinResourceCopier) CopyWorkflows(projectRoot string) ([]string, error) {
	var warnings []string

	entries, err := fs.ReadDir(b.workflowsFS, "workflows")
	if err != nil {
		// If the directory doesn't exist in the FS, treat as empty
		return warnings, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		workflowName := strings.TrimSuffix(name, ".yaml")
		targetPath := b.layout.GetWorkflowPath(projectRoot, workflowName)

		_, statErr := os.Stat(targetPath)
		if statErr == nil {
			// Target exists — skip with warning
			warnings = append(warnings, fmt.Sprintf("Warning: workflow definition '%s' already exists, skipping", name))
			continue
		}

		content, readErr := fs.ReadFile(b.workflowsFS, filepath.Join("workflows", name))
		if readErr != nil {
			return warnings, fmt.Errorf("failed to write built-in file '%s': %w", targetPath, readErr)
		}

		if writeErr := os.WriteFile(targetPath, content, 0644); writeErr != nil {
			return warnings, fmt.Errorf("failed to write built-in file '%s': %w", targetPath, writeErr)
		}
	}

	return warnings, nil
}

// CopyAgents copies embedded agent definition files to the project's .spectra/agents/ directory.
// Files that already exist are skipped with a warning.
func (b *BuiltinResourceCopier) CopyAgents(projectRoot string) ([]string, error) {
	var warnings []string

	entries, err := fs.ReadDir(b.agentsFS, "agents")
	if err != nil {
		// If the directory doesn't exist in the FS, treat as empty
		return warnings, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		agentName := strings.TrimSuffix(name, ".yaml")
		targetPath := b.layout.GetAgentPath(projectRoot, agentName)

		_, statErr := os.Stat(targetPath)
		if statErr == nil {
			// Target exists — skip with warning
			warnings = append(warnings, fmt.Sprintf("Warning: agent definition '%s' already exists, skipping", name))
			continue
		}

		content, readErr := fs.ReadFile(b.agentsFS, filepath.Join("agents", name))
		if readErr != nil {
			return warnings, fmt.Errorf("failed to write built-in file '%s': %w", targetPath, readErr)
		}

		if writeErr := os.WriteFile(targetPath, content, 0644); writeErr != nil {
			return warnings, fmt.Errorf("failed to write built-in file '%s': %w", targetPath, writeErr)
		}
	}

	return warnings, nil
}

// CopySpecFiles copies embedded spec template files to the project's spec/ directory.
// Files that already exist are skipped with a warning.
func (b *BuiltinResourceCopier) CopySpecFiles(projectRoot string) ([]string, error) {
	var warnings []string

	err := fs.WalkDir(b.specFilesFS, "spec", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		// relativePath is the path relative to "spec/" in the embedded FS
		relativePath := strings.TrimPrefix(path, "spec/")
		targetPath := filepath.Join(projectRoot, "spec", relativePath)

		_, statErr := os.Stat(targetPath)
		if statErr == nil {
			// Target exists — skip with warning
			warnings = append(warnings, fmt.Sprintf("Warning: spec file '%s' already exists, skipping", relativePath))
			return nil
		}

		content, readErr := fs.ReadFile(b.specFilesFS, path)
		if readErr != nil {
			return fmt.Errorf("failed to write built-in file '%s': %w", targetPath, readErr)
		}

		if writeErr := os.WriteFile(targetPath, content, 0644); writeErr != nil {
			return fmt.Errorf("failed to write built-in file '%s': %w", targetPath, writeErr)
		}

		return nil
	})

	if err != nil {
		return warnings, err
	}

	return warnings, nil
}
