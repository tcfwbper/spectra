package spectra

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcfwbper/spectra/internal/cmdutil"
)

// ClearSpectraFinder defines the interface for project root discovery used by the clear command.
type ClearSpectraFinder interface {
	Find() (string, error)
}

// ClearStorageLayout defines the interface for path composition used by the clear command.
type ClearStorageLayout interface {
	GetSessionDir(projectRoot, uuid string) string
	GetSessionsDir(projectRoot string) string
}

// ClearCommandOptions holds injectable dependencies for the clear command.
type ClearCommandOptions struct {
	Finder ClearSpectraFinder
	Layout ClearStorageLayout
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Args   []string
}

// RunClearCommand executes the clear command logic with the given options.
// Returns the exit code.
func RunClearCommand(opts ClearCommandOptions) int {
	// Step 1: Discover project root.
	projectRoot, err := opts.Finder.Find()
	if err != nil {
		_, _ = fmt.Fprintln(opts.Stderr, "Error: .spectra directory not found. Are you in a Spectra project?")
		return 1
	}

	if len(opts.Args) > 0 {
		return clearSpecificSessions(opts, projectRoot)
	}
	return clearAllSessions(opts, projectRoot)
}

// clearSpecificSessions handles Case 1: delete specific sessions by UUID.
func clearSpecificSessions(opts ClearCommandOptions, projectRoot string) int {
	// Build confirmation prompt listing UUIDs.
	var sb strings.Builder
	sb.WriteString("Are you sure you want to delete the following sessions?\n")
	for _, uuid := range opts.Args {
		sb.WriteString("  - ")
		sb.WriteString(uuid)
		sb.WriteString("\n")
	}
	sb.WriteString("[y/N]: ")
	prompt := sb.String()

	confirmed, err := cmdutil.ConfirmPrompt(opts.Stdin, opts.Stdout, prompt)
	if err != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", err)
		return 1
	}
	if !confirmed {
		_, _ = fmt.Fprintln(opts.Stdout, "Operation cancelled")
		return 0
	}

	// Iterate over each UUID and attempt deletion.
	for _, uuid := range opts.Args {
		sessionDir := opts.Layout.GetSessionDir(projectRoot, uuid)

		_, statErr := os.Stat(sessionDir)
		if statErr != nil {
			_, _ = fmt.Fprintf(opts.Stdout, "%s\n", cmdutil.FormatWarning(fmt.Sprintf("session '%s' not found, skipping", uuid)))
			continue
		}

		if removeErr := os.RemoveAll(sessionDir); removeErr != nil {
			_, _ = fmt.Fprintf(opts.Stderr, "%s\n", cmdutil.FormatError(fmt.Sprintf("failed to clear session '%s': %s", uuid, removeErr)))
			continue
		}

		_, _ = fmt.Fprintf(opts.Stdout, "Session '%s' cleared\n", uuid)
	}

	return 0
}

// clearAllSessions handles Case 2: delete all sessions.
func clearAllSessions(opts ClearCommandOptions, projectRoot string) int {
	sessionsDir := opts.Layout.GetSessionsDir(projectRoot)

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(opts.Stdout, "Warning: sessions directory not found, nothing to clear")
			return 0
		}
		_, _ = fmt.Fprintf(opts.Stderr, "%s\n", cmdutil.FormatError(fmt.Sprintf("failed to read sessions directory: %s", err)))
		return 1
	}

	// Filter to directories only.
	var dirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		}
	}

	if len(dirs) == 0 {
		_, _ = fmt.Fprintln(opts.Stdout, "No sessions to clear")
		return 0
	}

	// Prompt for confirmation.
	prompt := "Are you sure you want to delete all sessions? [y/N]: "
	confirmed, promptErr := cmdutil.ConfirmPrompt(opts.Stdin, opts.Stdout, prompt)
	if promptErr != nil {
		_, _ = fmt.Fprintf(opts.Stderr, "Error: %s\n", promptErr)
		return 1
	}
	if !confirmed {
		_, _ = fmt.Fprintln(opts.Stdout, "Operation cancelled")
		return 0
	}

	// Delete each directory.
	hasFailure := false
	for _, dir := range dirs {
		dirPath := filepath.Join(sessionsDir, dir.Name())
		if removeErr := os.RemoveAll(dirPath); removeErr != nil {
			_, _ = fmt.Fprintf(opts.Stderr, "%s\n", cmdutil.FormatError(fmt.Sprintf("failed to clear session '%s': %s", dir.Name(), removeErr)))
			hasFailure = true
			continue
		}
		_, _ = fmt.Fprintf(opts.Stdout, "Session '%s' cleared\n", dir.Name())
	}

	if !hasFailure {
		_, _ = fmt.Fprintln(opts.Stdout, "All sessions cleared successfully")
	}

	return 0
}
