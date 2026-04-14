package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var validConfigKeys = map[string]bool{
	"spec":     true,
	"test":     true,
	"src":      true,
	"language": true,
}

// IsValidConfigKey reports whether key is a recognized config key.
func IsValidConfigKey(key string) bool {
	return validConfigKeys[key]
}

// ValidConfigKeys returns the sorted list of recognized keys.
func ValidConfigKeys() []string {
	return []string{"language", "spec", "src", "test"}
}

// DefaultConfig is the default [core] config written by spectra init.
var DefaultConfig = map[string]string{
	"language": "golang",
	"spec":     "./spec",
	"test":     "./src",
	"src":      "./src",
}

// WriteDefaultConfig writes the default config file at path.
// If the file already exists it is left untouched so user customizations
// are preserved across repeated init calls.
func WriteDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists, keep user customizations
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString("[core]\n")
	// Write in a stable order.
	for _, key := range ValidConfigKeys() {
		fmt.Fprintf(&buf, "\t%s = %s\n", key, DefaultConfig[key])
	}
	return os.WriteFile(path, []byte(buf.String()), 0o644)
}

// WriteConfigValue sets a key under [core] in a git-config style file.
// If the key already exists it is updated in place; otherwise it is appended.
// The parent directory is created if needed.
func WriteConfigValue(path, key, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := splitLines(string(content))
	newLine := fmt.Sprintf("\t%s = %s", key, value)

	// Try to find and update existing key in [core] section.
	inCore := false
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[core]" {
			inCore = true
			continue
		}
		if inCore && strings.HasPrefix(trimmed, "[") {
			// Entered a new section, stop looking.
			break
		}
		if inCore && strings.HasPrefix(trimmed, key+" =") || inCore && strings.HasPrefix(trimmed, key+"=") {
			lines[i] = newLine
			found = true
			break
		}
	}

	if found {
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
	}

	// Append to existing [core] section or create one.
	hasCore := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "[core]" {
			hasCore = true
			break
		}
	}

	if hasCore {
		// Insert the new key right after the last key in [core].
		var result []string
		inCore = false
		inserted := false
		for i, line := range lines {
			result = append(result, line)
			trimmed := strings.TrimSpace(line)
			if trimmed == "[core]" {
				inCore = true
				continue
			}
			if inCore && !inserted {
				// Check if next line is outside [core] or we're at end.
				nextIsOutside := i == len(lines)-1
				if !nextIsOutside {
					nextTrimmed := strings.TrimSpace(lines[i+1])
					nextIsOutside = strings.HasPrefix(nextTrimmed, "[") || nextTrimmed == ""
				}
				if nextIsOutside {
					result = append(result, newLine)
					inserted = true
				}
			}
		}
		if !inserted {
			result = append(result, newLine)
		}
		return os.WriteFile(path, []byte(strings.Join(result, "\n")+"\n"), 0o644)
	}

	// No [core] section exists yet; create it.
	var buf strings.Builder
	if len(content) > 0 {
		buf.Write(content)
		if content[len(content)-1] != '\n' {
			buf.WriteByte('\n')
		}
	}
	buf.WriteString("[core]\n")
	buf.WriteString(newLine)
	buf.WriteByte('\n')
	return os.WriteFile(path, []byte(buf.String()), 0o644)
}

// ReadConfigValue reads a key from the [core] section. Returns ("", nil) if
// the key is not set.
func ReadConfigValue(path, key string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	lines := splitLines(string(content))
	inCore := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[core]" {
			inCore = true
			continue
		}
		if inCore && strings.HasPrefix(trimmed, "[") {
			break
		}
		if inCore {
			k, v, ok := parseConfigLine(trimmed)
			if ok && k == key {
				return v, nil
			}
		}
	}
	return "", nil
}

// parseConfigLine parses "key = value" and returns (key, value, true).
func parseConfigLine(line string) (string, string, bool) {
	k, v, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(k), strings.TrimSpace(v), true
}
