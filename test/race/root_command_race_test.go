package race_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	spectra "github.com/tcfwbper/spectra/cmd/spectra"
)

// TestRootCommand_ConcurrentInvocations multiple concurrent invocations execute independently.
func TestRootCommand_ConcurrentInvocations(t *testing.T) {
	numGoroutines := 10

	type result struct {
		stdout   string
		stderr   string
		exitCode int
	}

	results := make([]result, numGoroutines)
	args := [][]string{
		{"--version"},
		{"--help"},
		{"--version"},
		{"--help"},
		{"--version"},
		{"--help"},
		{"--version"},
		{"--help"},
		{"--version"},
		{"--help"},
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			var stdout, stderr bytes.Buffer

			cmd := spectra.NewRootCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(args[idx])

			exitCode := cmd.Execute()

			results[idx] = result{
				stdout:   stdout.String(),
				stderr:   stderr.String(),
				exitCode: exitCode,
			}
		}(i)
	}

	wg.Wait()

	// All invocations should complete without data races and produce correct output
	for i := 0; i < numGoroutines; i++ {
		assert.Equal(t, 0, results[i].exitCode, "goroutine %d should exit with code 0", i)
		assert.NotEmpty(t, results[i].stdout, "goroutine %d should produce output", i)

		if i%2 == 0 {
			assert.Contains(t, results[i].stdout, "spectra version", "goroutine %d should show version", i)
		} else {
			assert.Contains(t, results[i].stdout, "Available Commands", "goroutine %d should show help", i)
		}
	}
}
