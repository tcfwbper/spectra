package spectra

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GitignoreEnsurer ensures that .gitignore in the project root contains
// a .spectra entry.
type GitignoreEnsurer struct{}

// NewGitignoreEnsurer returns a new GitignoreEnsurer instance.
func NewGitignoreEnsurer() *GitignoreEnsurer {
	return &GitignoreEnsurer{}
}

// Ensure checks the .gitignore file in projectRoot for a .spectra entry.
// If .gitignore does not exist, it creates one containing ".spectra\n".
// If .gitignore exists but does not contain .spectra, it appends the entry.
// If .spectra is already present, it does nothing.
func (g *GitignoreEnsurer) Ensure(projectRoot string) error {
	path := filepath.Join(projectRoot, ".gitignore")

	// Check if path exists at all (including as a broken symlink)
	_, lstatErr := os.Lstat(path)
	if lstatErr != nil && os.IsNotExist(lstatErr) {
		// File truly does not exist — create it
		if writeErr := os.WriteFile(path, []byte(".spectra\n"), 0644); writeErr != nil {
			return fmt.Errorf("failed to update '.gitignore': %w", writeErr)
		}
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read '.gitignore': %w", err)
	}

	// Check if .spectra is already present
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := trimSpacesAndTabs(line)
		if trimmed == ".spectra" {
			return nil
		}
	}

	// .spectra not found — append it
	var newContent string
	s := string(content)
	if len(s) == 0 || s[len(s)-1] == '\n' {
		newContent = s + ".spectra\n"
	} else {
		newContent = s + "\n.spectra\n"
	}

	if writeErr := os.WriteFile(path, []byte(newContent), 0644); writeErr != nil {
		return fmt.Errorf("failed to update '.gitignore': %w", writeErr)
	}

	return nil
}

// trimSpacesAndTabs trims only ASCII space and tab characters from both ends of a string.
func trimSpacesAndTabs(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
