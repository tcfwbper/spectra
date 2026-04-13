package util

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyEmbeddedDir copies all files from an embed.FS (rooted at root) into destDir.
func CopyEmbeddedDir(fsys fs.FS, root string, destDir string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path from root
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		// Skip placeholder files used to keep empty dirs in embed.FS
		if d.Name() == "PLACEHOLDER" {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// CopyEmbeddedFile copies a single file from an embed.FS to destPath.
func CopyEmbeddedFile(fsys fs.FS, name string, destPath string) error {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("reading embedded file %s: %w", name, err)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0o644)
}

// EnsureGitignoreLine ensures the given line exists in the .gitignore file.
// Creates the file if it does not exist.
func EnsureGitignoreLine(path string, line string) error {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := splitLines(string(content))
	for _, l := range lines {
		if l == line {
			return nil // already present
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Add newline before the entry if file is non-empty and doesn't end with newline
	prefix := ""
	if len(content) > 0 && content[len(content)-1] != '\n' {
		prefix = "\n"
	}
	_, err = fmt.Fprintf(f, "%s%s\n", prefix, line)
	return err
}

func splitLines(s string) []string {
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
