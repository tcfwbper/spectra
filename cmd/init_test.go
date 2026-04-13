package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Run in a temp directory
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("cleanup chdir failed: %v", err)
		}
	})

	// Execute init command
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify .spectra/roles/ has expected files
	rolesDir := filepath.Join(tmp, ".spectra", "roles")
	expectedRoles := []string{
		"ARCHITECT.md",
		"ARCHITECT_REVIEWER.md",
		"HUMAN.md",
		"QA_ANALYST.md",
		"QA_ENGINEER.md",
		"QA_REVIEWER.md",
		"QA_SPEC_REVIEWER.md",
		"SW_ENGINEER.md",
	}
	for _, name := range expectedRoles {
		if _, err := os.Stat(filepath.Join(rolesDir, name)); err != nil {
			t.Errorf("expected role file %s to exist: %v", name, err)
		}
	}

	// Verify .spectra/skills/ exists
	skillsDir := filepath.Join(tmp, ".spectra", "skills")
	if _, err := os.Stat(skillsDir); err != nil {
		t.Errorf("expected .spectra/skills/ to exist: %v", err)
	}

	// Verify .spectra/README.md exists
	readmePath := filepath.Join(tmp, ".spectra", "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Errorf("expected .spectra/README.md to exist: %v", err)
	}

	// Verify spec/ has template files
	specDir := filepath.Join(tmp, "spec")
	expectedTemplates := []string{
		"ARCHITECTURE.md",
		"CONVENTIONS.md",
		"GLOSSARY.md",
		"logic/README.md",
		"test/README.md",
	}
	for _, name := range expectedTemplates {
		if _, err := os.Stat(filepath.Join(specDir, name)); err != nil {
			t.Errorf("expected template file spec/%s to exist: %v", name, err)
		}
	}

	// Verify .gitignore contains ".spectra/vfs"
	gitignoreContent, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to exist: %v", err)
	}
	if !containsLine(string(gitignoreContent), ".spectra/vfs") {
		t.Errorf(".gitignore does not contain '.spectra/vfs', got: %q", string(gitignoreContent))
	}
}

func TestInitCommandIdempotent(t *testing.T) {
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("cleanup chdir failed: %v", err)
		}
	})

	// Run init twice
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("first init failed: %v", err)
	}
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("second init failed: %v", err)
	}

	// .gitignore should contain ".spectra/vfs" only once
	content, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range splitTestLines(string(content)) {
		if line == ".spectra/vfs" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected .spectra/vfs to appear once, got %d times in: %q", count, string(content))
	}
}

func TestInitCommandExistingGitignore(t *testing.T) {
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("cleanup chdir failed: %v", err)
		}
	})

	// Create existing .gitignore without trailing newline
	if err := os.WriteFile(filepath.Join(tmp, ".gitignore"), []byte("node_modules"), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsLine(string(content), "node_modules") {
		t.Error("existing .gitignore content was lost")
	}
	if !containsLine(string(content), ".spectra/vfs") {
		t.Errorf(".gitignore missing .spectra/vfs, got: %q", string(content))
	}
}

func containsLine(content, line string) bool {
	for _, l := range splitTestLines(content) {
		if l == line {
			return true
		}
	}
	return false
}

func splitTestLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
