package race_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// TestInit_ConcurrentInvocations handles concurrent invocations in same directory (race condition).
func TestInit_ConcurrentInvocations(t *testing.T) {
	tmpDir := t.TempDir()

	type result struct {
		stdout   string
		stderr   string
		exitCode int
	}

	numGoroutines := 2
	results := make([]result, numGoroutines)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	done := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			var stdout, stderr bytes.Buffer

			cmd := spectra.NewRootCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{"init"})

			origDir, err := os.Getwd()
			if err != nil {
				results[idx] = result{exitCode: -1, stderr: err.Error()}
				return
			}
			os.Chdir(tmpDir)
			defer os.Chdir(origDir)

			exitCode := cmd.Execute()

			results[idx] = result{
				stdout:   stdout.String(),
				stderr:   stderr.String(),
				exitCode: exitCode,
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Completed within timeout
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent invocations deadlocked (exceeded 5 second timeout)")
	}

	// Both commands may succeed or one may fail with file exists error; no data corruption
	for i := 0; i < numGoroutines; i++ {
		assert.True(t, results[i].exitCode == 0 || results[i].exitCode == 1,
			"goroutine %d should exit with code 0 or 1, got %d", i, results[i].exitCode)
	}

	// Filesystem remains consistent — all expected directories/files exist
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "sessions"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "workflows"))
	assert.DirExists(t, filepath.Join(tmpDir, ".spectra", "agents"))

	// .gitignore should exist and contain .spectra
	gitignoreContent, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(gitignoreContent), ".spectra")
}
