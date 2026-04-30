package spectra

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/storage"
)

// newClearCommand creates the clear subcommand
func newClearCommand(handler SubcommandHandler) *cobra.Command {
	var sessionID string
	var sessionIDSet bool

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear session data",
		Long:  "Clear session data",
		Example: `  spectra clear
  spectra clear --session-id 12345678-1234-1234-1234-123456789abc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If a handler is provided (for testing), use it
			if handler != nil {
				exitCode := handler.Execute()
				if exitCode != 0 {
					return &exitError{code: exitCode}
				}
				return nil
			}

			// Check if the flag was explicitly set
			sessionIDSet = cmd.Flags().Changed("session-id")

			// Default implementation
			exitCode := executeClear(cmd, sessionID, sessionIDSet)
			if exitCode != 0 {
				return &exitError{code: exitCode}
			}
			return nil
		},
	}

	clearCmd.Flags().StringVar(&sessionID, "session-id", "", "UUID of the session to clear (if not provided, clears all sessions)")

	return clearCmd
}

// executeClear implements the clear command logic
func executeClear(cmd *cobra.Command, sessionID string, sessionIDSet bool) int {
	// Find the project root using SpectraFinder
	projectRoot, err := storage.SpectraFinder("")
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: .spectra directory not found. Are you in a Spectra project?\n")
		return 1
	}

	// Get the sessions directory path
	sessionsDir := storage.GetSessionsDir(projectRoot)

	// Case 1: Delete specific session (flag was explicitly set)
	if sessionIDSet {
		return clearSpecificSession(cmd, projectRoot, sessionID)
	}

	// Case 2: Delete all sessions
	return clearAllSessions(cmd, sessionsDir)
}

// clearSpecificSession deletes a specific session by ID
func clearSpecificSession(cmd *cobra.Command, projectRoot string, sessionID string) int {
	// Special case: empty session ID results in invalid path
	if sessionID == "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: session '%s' not found, skipping\n", sessionID)
		return 0
	}

	sessionDir := storage.GetSessionDir(projectRoot, sessionID)

	// Check if session directory exists
	_, err := os.Stat(sessionDir)
	if os.IsNotExist(err) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: session '%s' not found, skipping\n", sessionID)
		return 0
	}
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to check session '%s': %v\n", sessionID, err)
		return 1
	}

	// Delete the session directory recursively
	err = os.RemoveAll(sessionDir)
	if err != nil {
		// Check for permission denied error
		if os.IsPermission(err) {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to clear session '%s': permission denied\n", sessionID)
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to clear session '%s': %v\n", sessionID, err)
		}
		return 1
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Session '%s' cleared successfully\n", sessionID)
	return 0
}

// clearAllSessions deletes all sessions after user confirmation
func clearAllSessions(cmd *cobra.Command, sessionsDir string) int {
	// Check if sessions directory exists
	entries, err := os.ReadDir(sessionsDir)
	if os.IsNotExist(err) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: sessions directory not found, nothing to clear\n")
		return 0
	}
	if err != nil {
		// Check for permission denied
		if os.IsPermission(err) {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to read sessions directory: permission denied\n")
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to read sessions directory: %v\n", err)
		}
		return 1
	}

	// Count directories (sessions)
	sessionDirs := make([]os.DirEntry, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			sessionDirs = append(sessionDirs, entry)
		}
	}

	// If no sessions exist, print message and exit
	if len(sessionDirs) == 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No sessions to clear\n")
		return 0
	}

	// Prompt for confirmation
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete all sessions? [y/N]: ")

	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		// Treat read error as cancellation
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Operation cancelled\n")
		return 0
	}

	response = strings.TrimSpace(response)
	if response != "y" && response != "Y" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Operation cancelled\n")
		return 0
	}

	// Delete all session directories
	for _, entry := range sessionDirs {
		sessionName := entry.Name()
		sessionPath := filepath.Join(sessionsDir, sessionName)

		err := os.RemoveAll(sessionPath)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to clear session '%s': %v\n", sessionName, err)
			continue
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Session '%s' cleared\n", sessionName)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "All sessions cleared successfully\n")

	// Return 0 even if some deletions failed (as per spec)
	return 0
}
