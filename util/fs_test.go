package util

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestCopyEmbeddedDir(t *testing.T) {
	fsys := fstest.MapFS{
		"root/a.txt":     {Data: []byte("hello")},
		"root/sub/b.txt": {Data: []byte("world")},
		"root/sub/c.txt": {Data: []byte("foo")},
	}

	dest := filepath.Join(t.TempDir(), "out")
	if err := CopyEmbeddedDir(fsys, "root", dest); err != nil {
		t.Fatalf("CopyEmbeddedDir failed: %v", err)
	}

	// Check a.txt
	data, err := os.ReadFile(filepath.Join(dest, "a.txt"))
	if err != nil {
		t.Fatalf("reading a.txt: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("a.txt = %q, want %q", string(data), "hello")
	}

	// Check sub/b.txt
	data, err = os.ReadFile(filepath.Join(dest, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("reading sub/b.txt: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("sub/b.txt = %q, want %q", string(data), "world")
	}

	// Check sub/c.txt
	data, err = os.ReadFile(filepath.Join(dest, "sub", "c.txt"))
	if err != nil {
		t.Fatalf("reading sub/c.txt: %v", err)
	}
	if string(data) != "foo" {
		t.Errorf("sub/c.txt = %q, want %q", string(data), "foo")
	}
}

func TestCopyEmbeddedFile(t *testing.T) {
	fsys := fstest.MapFS{
		"README.md": {Data: []byte("# Title")},
	}

	dest := filepath.Join(t.TempDir(), "out", "README.md")
	if err := CopyEmbeddedFile(fsys, "README.md", dest); err != nil {
		t.Fatalf("CopyEmbeddedFile failed: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Title" {
		t.Errorf("got %q, want %q", string(data), "# Title")
	}
}

func TestEnsureGitignoreLine_NewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gitignore")

	if err := EnsureGitignoreLine(path, ".spectra/vfs"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != ".spectra/vfs\n" {
		t.Errorf("got %q, want %q", string(data), ".spectra/vfs\n")
	}
}

func TestEnsureGitignoreLine_ExistingWithoutLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(path, []byte("node_modules\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureGitignoreLine(path, ".spectra/vfs"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "node_modules\n.spectra/vfs\n"
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}
}

func TestEnsureGitignoreLine_ExistingWithLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(path, []byte(".spectra/vfs\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureGitignoreLine(path, ".spectra/vfs"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != ".spectra/vfs\n" {
		t.Errorf("got %q, want %q", string(data), ".spectra/vfs\n")
	}
}

func TestEnsureGitignoreLine_NoTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(path, []byte("node_modules"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureGitignoreLine(path, ".spectra/vfs"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "node_modules\n.spectra/vfs\n"
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"a\n", []string{"a"}},
		{"a\nb\n", []string{"a", "b"}},
		{"a\nb", []string{"a", "b"}},
		{"single", []string{"single"}},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitLines(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitLines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
