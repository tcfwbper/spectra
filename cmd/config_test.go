package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tcfwbper/spectra/util"
)

func TestConfigCommand_SetLanguage(t *testing.T) {
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

	rootCmd.SetArgs([]string{"config", "language", "golang"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config command failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmp, ".spectra", "config"))
	if err != nil {
		t.Fatalf("expected .spectra/config to exist: %v", err)
	}

	expected := "[core]\n\tlanguage = golang\n"
	if string(content) != expected {
		t.Errorf("got %q, want %q", string(content), expected)
	}
}

func TestConfigCommand_SetAllKeys(t *testing.T) {
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

	settings := []struct{ key, value string }{
		{"language", "golang"},
		{"spec", "./spec"},
		{"test", "./test"},
		{"src", "./src"},
	}

	for _, s := range settings {
		rootCmd.SetArgs([]string{"config", s.key, s.value})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("config %s %s failed: %v", s.key, s.value, err)
		}
	}

	configPath := filepath.Join(tmp, ".spectra", "config")
	for _, s := range settings {
		got, err := util.ReadConfigValue(configPath, s.key)
		if err != nil {
			t.Fatalf("reading %s: %v", s.key, err)
		}
		if got != s.value {
			t.Errorf("config %s = %q, want %q", s.key, got, s.value)
		}
	}
}

func TestConfigCommand_UpdateExisting(t *testing.T) {
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

	rootCmd.SetArgs([]string{"config", "language", "golang"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"config", "language", "python"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(tmp, ".spectra", "config")
	got, err := util.ReadConfigValue(configPath, "language")
	if err != nil {
		t.Fatal(err)
	}
	if got != "python" {
		t.Errorf("language = %q, want %q", got, "python")
	}
}

func TestConfigCommand_OverrideInitDefaults(t *testing.T) {
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

	// Run init to get default config.
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// Override each default with spectra config.
	overrides := []struct{ key, value string }{
		{"language", "python"},
		{"spec", "./docs/spec"},
		{"test", "./tests"},
		{"src", "./lib"},
	}
	for _, o := range overrides {
		rootCmd.SetArgs([]string{"config", o.key, o.value})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("config %s %s failed: %v", o.key, o.value, err)
		}
	}

	configPath := filepath.Join(tmp, ".spectra", "config")
	for _, o := range overrides {
		got, err := util.ReadConfigValue(configPath, o.key)
		if err != nil {
			t.Fatalf("reading %s: %v", o.key, err)
		}
		if got != o.value {
			t.Errorf("config %s = %q, want %q", o.key, got, o.value)
		}
	}
}

func TestConfigCommand_InvalidKey(t *testing.T) {
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

	rootCmd.SetArgs([]string{"config", "invalid", "value"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
}

func TestConfigCommand_MissingArgs(t *testing.T) {
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

	rootCmd.SetArgs([]string{"config"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}

	rootCmd.SetArgs([]string{"config", "language"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing value, got nil")
	}
}
