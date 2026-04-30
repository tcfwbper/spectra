package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed builtin/workflows/*.yaml
var builtinWorkflows embed.FS

//go:embed builtin/agents/*.yaml
var builtinAgents embed.FS

//go:embed builtin/spec/ARCHITECTURE.md
//go:embed builtin/spec/CONVENTIONS.md
//go:embed builtin/spec/logic/README.md
//go:embed builtin/spec/test/README.md
var builtinSpecFiles embed.FS

// BuiltinWorkflowsFS returns the embedded workflows filesystem
func BuiltinWorkflowsFS() embed.FS {
	return builtinWorkflows
}

// BuiltinAgentsFS returns the embedded agents filesystem
func BuiltinAgentsFS() embed.FS {
	return builtinAgents
}

// BuiltinSpecFilesFS returns the embedded spec files filesystem
func BuiltinSpecFilesFS() embed.FS {
	return builtinSpecFiles
}

// InitHandler handles the init command
type InitHandler interface {
	Execute() int
}

// initHandler implements InitHandler
type initHandler struct {
	projectRoot string
	copier      BuiltinResourceCopier
	stdout      io.Writer
	stderr      io.Writer
}

// NewInitHandler creates a new InitHandler
func NewInitHandler(projectRoot string, copier BuiltinResourceCopier) InitHandler {
	return NewInitHandlerWithOutput(projectRoot, copier, os.Stdout, os.Stderr)
}

// NewInitHandlerWithOutput creates a new InitHandler with custom output streams
func NewInitHandlerWithOutput(projectRoot string, copier BuiltinResourceCopier, stdout, stderr io.Writer) InitHandler {
	return &initHandler{
		projectRoot: projectRoot,
		copier:      copier,
		stdout:      stdout,
		stderr:      stderr,
	}
}

// Execute runs the init command
func (h *initHandler) Execute() int {
	// Phase 0: Ensure .gitignore contains .spectra
	if err := h.ensureGitignore(); err != nil {
		fmt.Fprintf(h.stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 1: Create .spectra/ directories
	dirs := []string{
		filepath.Join(h.projectRoot, ".spectra"),
		filepath.Join(h.projectRoot, ".spectra", "sessions"),
		filepath.Join(h.projectRoot, ".spectra", "workflows"),
		filepath.Join(h.projectRoot, ".spectra", "agents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(h.stderr, "Error: failed to create directory '%s': %v\n", dir, err)
			return 1
		}
	}

	// Phase 2: Create .spectra/ files
	workflowWarnings, err := h.copier.CopyWorkflows(h.projectRoot)
	if err != nil {
		fmt.Fprintf(h.stderr, "Error: %v\n", err)
		return 1
	}
	for _, warning := range workflowWarnings {
		fmt.Fprintln(h.stdout, warning)
	}

	agentWarnings, err := h.copier.CopyAgents(h.projectRoot)
	if err != nil {
		fmt.Fprintf(h.stderr, "Error: %v\n", err)
		return 1
	}
	for _, warning := range agentWarnings {
		fmt.Fprintln(h.stdout, warning)
	}

	// Phase 3: Create spec/ directories
	specDirs := []string{
		filepath.Join(h.projectRoot, "spec"),
		filepath.Join(h.projectRoot, "spec", "logic"),
		filepath.Join(h.projectRoot, "spec", "test"),
	}

	for _, dir := range specDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(h.stderr, "Error: failed to create directory '%s': %v\n", dir, err)
			return 1
		}
	}

	// Phase 4: Create spec/ files
	specWarnings, err := h.copier.CopySpecFiles(h.projectRoot)
	if err != nil {
		fmt.Fprintf(h.stderr, "Error: %v\n", err)
		return 1
	}
	for _, warning := range specWarnings {
		fmt.Fprintln(h.stdout, warning)
	}

	// Success
	fmt.Fprintln(h.stdout, "Spectra project initialized successfully")
	return 0
}

// ensureGitignore ensures .gitignore contains .spectra entry
func (h *initHandler) ensureGitignore() error {
	gitignorePath := filepath.Join(h.projectRoot, ".gitignore")

	// Check if .gitignore exists (follows symlinks)
	_, err := os.Stat(gitignorePath)
	if err != nil {
		// Check if it's a broken symlink
		if _, lstatErr := os.Lstat(gitignorePath); lstatErr == nil {
			// File exists according to Lstat but Stat failed - broken symlink
			return fmt.Errorf("failed to read '.gitignore': %w", err)
		}
		// File doesn't exist at all
		if os.IsNotExist(err) {
			// Create .gitignore with .spectra entry
			if err := os.WriteFile(gitignorePath, []byte(".spectra\n"), 0644); err != nil {
				return fmt.Errorf("failed to update '.gitignore': %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to read '.gitignore': %w", err)
	}

	// Read existing .gitignore content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("failed to read '.gitignore': %w", err)
	}

	// Check if .spectra entry exists
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	hasSpectra := false

	for scanner.Scan() {
		line := scanner.Text()

		// Trim spaces and tabs from the line
		trimmed := strings.Trim(line, " \t")
		if trimmed == ".spectra" {
			hasSpectra = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read '.gitignore': %w", err)
	}

	// If .spectra is already present, no need to modify
	if hasSpectra {
		return nil
	}

	// Append .spectra entry
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to update '.gitignore': %w", err)
	}
	defer f.Close()

	// Check if file ends with newline
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to update '.gitignore': %w", err)
		}
	}

	if _, err := f.WriteString(".spectra\n"); err != nil {
		return fmt.Errorf("failed to update '.gitignore': %w", err)
	}

	return nil
}
