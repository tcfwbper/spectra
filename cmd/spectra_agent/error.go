package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/entities"
)

// ErrorCommandHandler handles the error subcommand execution.
type ErrorCommandHandler struct {
	cmd         *cobra.Command
	args        []string
	sessionID   string
	projectRoot string
	config      *rootCommandConfig
}

// NewErrorCommandHandler creates a new error command handler.
func NewErrorCommandHandler(cmd *cobra.Command, args []string, sessionID, projectRoot string, config *rootCommandConfig) *ErrorCommandHandler {
	return &ErrorCommandHandler{
		cmd:         cmd,
		args:        args,
		sessionID:   sessionID,
		projectRoot: projectRoot,
		config:      config,
	}
}

// Execute runs the error command and returns an exit code.
// Exit codes: 0 = success, 1 = validation error, 2 = transport error, 3 = runtime error
func (h *ErrorCommandHandler) Execute() int {
	// Extract message from args
	var message string
	if len(h.args) > 0 {
		message = h.args[0]
	}

	// Validate message
	if message == "" {
		_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: error message is required\n")
		return 1
	}

	// Get flags
	detailStr, _ := h.cmd.Flags().GetString("detail")
	claudeSessionID, _ := h.cmd.Flags().GetString("claude-session-id")

	// Parse detail
	var detail json.RawMessage
	if detailStr == "" {
		// Default to empty object
		detail = json.RawMessage("{}")
	} else {
		// Validate JSON
		var parsed any
		if err := json.Unmarshal([]byte(detailStr), &parsed); err != nil {
			_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: --detail must be a JSON object or null\n")
			return 1
		}

		// Check if it's an object or null
		switch parsed.(type) {
		case map[string]any:
			// Valid object
			detail = json.RawMessage(detailStr)
		case nil:
			// Valid null
			detail = json.RawMessage("null")
		default:
			_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: --detail must be a JSON object or null\n")
			return 1
		}
	}

	// Build error payload
	errorPayload := entities.ErrorPayload{
		Message: message,
		Detail:  detail,
	}

	payloadBytes, err := json.Marshal(errorPayload)
	if err != nil {
		_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: failed to marshal error payload: %v\n", err)
		return 1
	}

	// Build runtime message
	runtimeMsg := entities.RuntimeMessage{
		Type:            "error",
		ClaudeSessionID: claudeSessionID,
		Payload:         payloadBytes,
	}

	// Get or create socket client and send message
	var exitCode int
	var sendErr error

	if h.config.mockSocketClient != nil {
		// Use mock client (from tests)
		type mockSender interface {
			Send(sessionID, projectRoot string, msg entities.RuntimeMessage) (*RuntimeResponse, int, error)
		}
		mockClient := h.config.mockSocketClient.(mockSender)
		_, exitCode, sendErr = mockClient.Send(h.sessionID, h.projectRoot, runtimeMsg)
	} else {
		// Create a real socket client
		var socketClient *SocketClient
		if h.config.socketClientTimeout > 0 {
			socketClient = NewSocketClientWithTimeout(h.config.socketClientTimeout)
		} else {
			socketClient = NewSocketClient()
		}
		_, exitCode, sendErr = socketClient.Send(h.sessionID, h.projectRoot, runtimeMsg)
	}
	if sendErr != nil {
		_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "%s\n", sendErr.Error())
		return exitCode
	}

	// Success - print success message (ignore Runtime's message)
	_, _ = fmt.Fprintf(h.cmd.OutOrStdout(), "Error reported successfully\n")
	return 0
}
