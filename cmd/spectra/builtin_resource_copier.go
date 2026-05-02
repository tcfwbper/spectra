package spectra

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcfwbper/spectra/storage"
)

// BuiltinResourceCopier copies embedded built-in files to the project directory
type BuiltinResourceCopier interface {
	CopyWorkflows(projectRoot string) ([]string, error)
	CopyAgents(projectRoot string) ([]string, error)
	CopySpecFiles(projectRoot string) ([]string, error)
}

// builtinResourceCopier implements BuiltinResourceCopier
type builtinResourceCopier struct {
	workflowsFS fs.FS
	agentsFS    fs.FS
	specFilesFS fs.FS
}

// NewBuiltinResourceCopier creates a new BuiltinResourceCopier
func NewBuiltinResourceCopier(workflowsFS, agentsFS, specFilesFS fs.FS) BuiltinResourceCopier {
	return &builtinResourceCopier{
		workflowsFS: workflowsFS,
		agentsFS:    agentsFS,
		specFilesFS: specFilesFS,
	}
}

// CopyWorkflows copies embedded workflow files to .spectra/workflows/
func (c *builtinResourceCopier) CopyWorkflows(projectRoot string) ([]string, error) {
	if c.workflowsFS == nil {
		return nil, nil
	}

	var warnings []string
	const workflowsPrefix = "builtin/workflows"
	entries, err := fs.ReadDir(c.workflowsFS, workflowsPrefix)
	if err != nil {
		// If directory doesn't exist or is empty, return success
		return warnings, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Extract workflow name by removing .yaml extension
		workflowName := strings.TrimSuffix(filename, ".yaml")

		// Compose target path using StorageLayout
		targetPath := storage.GetWorkflowPath(projectRoot, workflowName)

		// Check if file exists
		if _, err := os.Stat(targetPath); err == nil {
			warnings = append(warnings, fmt.Sprintf("Warning: workflow definition '%s' already exists, skipping", filename))
			continue
		}

		// Read embedded file content
		content, err := fs.ReadFile(c.workflowsFS, workflowsPrefix+"/"+filename)
		if err != nil {
			return warnings, fmt.Errorf("failed to read embedded file '%s/%s': %w", workflowsPrefix, filename, err)
		}

		// Write file to target path
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return warnings, fmt.Errorf("failed to write built-in file '%s': %w", targetPath, err)
		}
	}

	return warnings, nil
}

// CopyAgents copies embedded agent files to .spectra/agents/
func (c *builtinResourceCopier) CopyAgents(projectRoot string) ([]string, error) {
	if c.agentsFS == nil {
		return nil, nil
	}

	var warnings []string
	const agentsPrefix = "builtin/agents"
	entries, err := fs.ReadDir(c.agentsFS, agentsPrefix)
	if err != nil {
		// If directory doesn't exist or is empty, return success
		return warnings, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Extract agent role by removing .yaml extension
		agentRole := strings.TrimSuffix(filename, ".yaml")

		// Compose target path using StorageLayout
		targetPath := storage.GetAgentPath(projectRoot, agentRole)

		// Check if file exists
		if _, err := os.Stat(targetPath); err == nil {
			warnings = append(warnings, fmt.Sprintf("Warning: agent definition '%s' already exists, skipping", filename))
			continue
		}

		// Read embedded file content
		content, err := fs.ReadFile(c.agentsFS, agentsPrefix+"/"+filename)
		if err != nil {
			return warnings, fmt.Errorf("failed to read embedded file '%s/%s': %w", agentsPrefix, filename, err)
		}

		// Write file to target path
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return warnings, fmt.Errorf("failed to write built-in file '%s': %w", targetPath, err)
		}
	}

	return warnings, nil
}

// CopySpecFiles copies embedded spec files to spec/
func (c *builtinResourceCopier) CopySpecFiles(projectRoot string) ([]string, error) {
	if c.specFilesFS == nil {
		return nil, nil
	}

	var warnings []string

	// Walk the embedded filesystem
	err := fs.WalkDir(c.specFilesFS, "builtin/spec", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Extract relative path from builtin/spec/
		relPath, err := filepath.Rel("builtin/spec", path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for '%s': %w", path, err)
		}

		// Compose target path
		targetPath := filepath.Join(projectRoot, "spec", relPath)

		// Check if file exists
		if _, err := os.Stat(targetPath); err == nil {
			warnings = append(warnings, fmt.Sprintf("Warning: spec file '%s' already exists, skipping", relPath))
			return nil
		}

		// Read embedded file content
		content, err := fs.ReadFile(c.specFilesFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file '%s': %w", path, err)
		}

		// Write file to target path
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write built-in file '%s': %w", targetPath, err)
		}

		return nil
	})

	if err != nil {
		return warnings, err
	}

	return warnings, nil
}
