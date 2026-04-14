package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteConfigValue_NewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".spectra", "config")

	if err := WriteConfigValue(path, "language", "golang"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "[core]\n\tlanguage = golang\n"
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}
}

func TestWriteConfigValue_MultipleKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".spectra", "config")

	for _, kv := range []struct{ k, v string }{
		{"language", "golang"},
		{"spec", "./spec"},
		{"src", "./src"},
	} {
		if err := WriteConfigValue(path, kv.k, kv.v); err != nil {
			t.Fatalf("setting %s: %v", kv.k, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "[core]\n\tlanguage = golang\n\tspec = ./spec\n\tsrc = ./src\n"
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}
}

func TestWriteConfigValue_UpdateExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".spectra", "config")

	if err := WriteConfigValue(path, "language", "golang"); err != nil {
		t.Fatal(err)
	}
	if err := WriteConfigValue(path, "language", "python"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "[core]\n\tlanguage = python\n"
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}
}

func TestWriteConfigValue_UpdatePreservesOthers(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".spectra", "config")

	if err := WriteConfigValue(path, "language", "golang"); err != nil {
		t.Fatal(err)
	}
	if err := WriteConfigValue(path, "spec", "./spec"); err != nil {
		t.Fatal(err)
	}
	// Update language, spec should remain.
	if err := WriteConfigValue(path, "language", "rust"); err != nil {
		t.Fatal(err)
	}

	got, err := ReadConfigValue(path, "language")
	if err != nil {
		t.Fatal(err)
	}
	if got != "rust" {
		t.Errorf("language = %q, want %q", got, "rust")
	}

	got, err = ReadConfigValue(path, "spec")
	if err != nil {
		t.Fatal(err)
	}
	if got != "./spec" {
		t.Errorf("spec = %q, want %q", got, "./spec")
	}
}

func TestReadConfigValue_NonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such-file")

	val, err := ReadConfigValue(path, "language")
	if err != nil {
		t.Fatal(err)
	}
	if val != "" {
		t.Errorf("got %q, want empty string", val)
	}
}

func TestReadConfigValue_KeyNotSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".spectra", "config")

	if err := WriteConfigValue(path, "language", "golang"); err != nil {
		t.Fatal(err)
	}

	val, err := ReadConfigValue(path, "src")
	if err != nil {
		t.Fatal(err)
	}
	if val != "" {
		t.Errorf("got %q, want empty string", val)
	}
}

func TestReadConfigValue_ReadsCorrectKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".spectra", "config")

	for _, kv := range []struct{ k, v string }{
		{"language", "golang"},
		{"spec", "./spec"},
		{"test", "./test"},
		{"src", "./src"},
	} {
		if err := WriteConfigValue(path, kv.k, kv.v); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		key  string
		want string
	}{
		{"language", "golang"},
		{"spec", "./spec"},
		{"test", "./test"},
		{"src", "./src"},
	}
	for _, tt := range tests {
		got, err := ReadConfigValue(path, tt.key)
		if err != nil {
			t.Fatalf("ReadConfigValue(%q): %v", tt.key, err)
		}
		if got != tt.want {
			t.Errorf("ReadConfigValue(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestIsValidConfigKey(t *testing.T) {
	valid := []string{"spec", "test", "src", "language"}
	for _, k := range valid {
		if !IsValidConfigKey(k) {
			t.Errorf("IsValidConfigKey(%q) = false, want true", k)
		}
	}

	invalid := []string{"foo", "bar", "", "LANGUAGE", "Spec"}
	for _, k := range invalid {
		if IsValidConfigKey(k) {
			t.Errorf("IsValidConfigKey(%q) = true, want false", k)
		}
	}
}
