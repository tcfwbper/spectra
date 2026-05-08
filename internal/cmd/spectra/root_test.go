package spectra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — Execute ---

func TestRoot_NoArgs_PrintsUsage(t *testing.T) {
	result := executeRoot(t, nil)

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "spectra [command]")
}

func TestRoot_Help_PrintsUsage(t *testing.T) {
	result := executeRoot(t, []string{"--help"})

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "spectra [command]")
}

func TestRoot_Version_PrintsVersion(t *testing.T) {
	result := executeRoot(t, []string{"--version"})

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "spectra version v0.1.0")
}

// --- Happy Path — Subcommand Registration ---

func TestRoot_RegistersInitSubcommand(t *testing.T) {
	cmd := NewRootCommand()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "init" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'init' subcommand to be registered")
}

func TestRoot_RegistersRunSubcommand(t *testing.T) {
	cmd := NewRootCommand()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "run" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'run' subcommand to be registered")
}

func TestRoot_RegistersClearSubcommand(t *testing.T) {
	cmd := NewRootCommand()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "clear" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'clear' subcommand to be registered")
}

// --- Validation Failures ---

func TestRoot_UnknownSubcommand_ExitsWithError(t *testing.T) {
	result := executeRoot(t, []string{"foo"})

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, `unknown command "foo" for "spectra"`)
}

// --- Happy Path — Exit Code Propagation ---

func TestRoot_PropagatesSubcommandExitCode0(t *testing.T) {
	result := executeRootWithStubSubcommand(t, "stub", 0, []string{"stub"})

	assert.Equal(t, 0, result.exitCode)
}

func TestRoot_PropagatesSubcommandExitCode1(t *testing.T) {
	result := executeRootWithStubSubcommand(t, "stub", 1, []string{"stub"})

	assert.Equal(t, 1, result.exitCode)
}

func TestRoot_PropagatesExitCode130(t *testing.T) {
	result := executeRootWithStubSubcommand(t, "stub", 130, []string{"stub"})

	assert.Equal(t, 130, result.exitCode)
}

func TestRoot_PropagatesExitCode143(t *testing.T) {
	result := executeRootWithStubSubcommand(t, "stub", 143, []string{"stub"})

	assert.Equal(t, 143, result.exitCode)
}
