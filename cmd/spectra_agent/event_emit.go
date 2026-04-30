package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tcfwbper/spectra/entities"
)

// EventEmitCommandHandler handles the event emit subcommand execution.
type EventEmitCommandHandler struct {
	cmd         *cobra.Command
	args        []string
	sessionID   string
	projectRoot string
	config      *rootCommandConfig
}

// NewEventEmitCommandHandler creates a new event emit command handler.
func NewEventEmitCommandHandler(cmd *cobra.Command, args []string, sessionID, projectRoot string, config *rootCommandConfig) *EventEmitCommandHandler {
	return &EventEmitCommandHandler{
		cmd:         cmd,
		args:        args,
		sessionID:   sessionID,
		projectRoot: projectRoot,
		config:      config,
	}
}

// Execute runs the event emit command and returns an exit code.
// Exit codes: 0 = success, 1 = validation error, 2 = transport error, 3 = runtime error
func (h *EventEmitCommandHandler) Execute() int {
	// Extract eventType from args
	var eventType string
	if len(h.args) > 0 {
		eventType = h.args[0]
	}

	// Validate eventType
	if eventType == "" {
		_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: event type is required\n")
		return 1
	}

	// Get flags
	message, _ := h.cmd.Flags().GetString("message")
	payloadStr, _ := h.cmd.Flags().GetString("payload")
	claudeSessionID, _ := h.cmd.Flags().GetString("claude-session-id")

	// Parse payload
	var payload json.RawMessage
	if payloadStr == "" {
		// Default to empty object
		payload = json.RawMessage("{}")
	} else {
		// Validate JSON
		var parsed any
		if err := json.Unmarshal([]byte(payloadStr), &parsed); err != nil {
			_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: --payload must be a valid JSON object, e.g., {}\n")
			return 1
		}

		// Check if it's an object
		switch parsed.(type) {
		case map[string]any:
			// Valid object
			payload = json.RawMessage(payloadStr)
		default:
			_, _ = fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: --payload must be a valid JSON object, e.g., {}\n")
			return 1
		}
	}

	// Build event payload
	eventPayload := entities.EventPayload{
		EventType: eventType,
		Message:   message,
		Payload:   payload,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		fmt.Fprintf(h.cmd.ErrOrStderr(), "Error: failed to marshal event payload: %v\n", err)
		return 1
	}

	// Build runtime message
	runtimeMsg := entities.RuntimeMessage{
		Type:            "event",
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
		fmt.Fprintf(h.cmd.ErrOrStderr(), "%s\n", sendErr.Error())
		return exitCode
	}

	// Success - print success message (ignore Runtime's message)
	fmt.Fprintf(h.cmd.OutOrStdout(), "Event emitted successfully\n")
	return 0
}
