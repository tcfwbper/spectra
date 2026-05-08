package spectraagent

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// eventWireMessage represents the wire format for event messages.
type eventWireMessage struct {
	Type            string           `json:"type"`
	ClaudeSessionID string           `json:"claudeSessionID"`
	Payload         eventWirePayload `json:"payload"`
}

// eventWirePayload represents the payload of an event wire message.
type eventWirePayload struct {
	EventType string          `json:"eventType"`
	Message   string          `json:"message"`
	Payload   json.RawMessage `json:"payload"`
}

func newEventCmd(ctx *cmdContext, stdoutBuf, stderrBuf *bytes.Buffer) *cobra.Command {
	eventCmd := &cobra.Command{
		Use:           "event",
		Short:         "Event-related commands",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	eventCmd.AddCommand(newEventEmitCmd(ctx, stdoutBuf, stderrBuf))

	return eventCmd
}

func newEventEmitCmd(ctx *cmdContext, stdoutBuf, stderrBuf *bytes.Buffer) *cobra.Command {
	var message string
	var payload string
	var claudeSessionID string

	cmd := &cobra.Command{
		Use:           "emit <type> [flags]",
		Short:         "Emit an event to the runtime",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate positional argument: type.
			if len(args) == 0 || args[0] == "" {
				fmt.Fprintln(stderrBuf, "Error: event type is required")
				return newExitCodeError(1, "event type is required")
			}
			eventType := args[0]

			// Validate --payload: must be a JSON object (not null, not primitives, not arrays).
			payloadBytes, err := validatePayloadFlag(payload)
			if err != nil {
				fmt.Fprintln(stderrBuf, "Error: --payload must be a valid JSON object, e.g., {}")
				return newExitCodeError(1, "--payload must be a valid JSON object, e.g., {}")
			}

			// Construct wire message.
			wireMsg := eventWireMessage{
				Type:            "event",
				ClaudeSessionID: claudeSessionID,
				Payload: eventWirePayload{
					EventType: eventType,
					Message:   message,
					Payload:   payloadBytes,
				},
			}

			// Delegate to SendAndHandle.
			exitCode, stdout, stderr := ctx.sender.SendAndHandle(
				ctx.sessionID,
				ctx.projectRoot,
				wireMsg,
				"Event emitted successfully",
			)

			if stdout != "" {
				fmt.Fprintln(stdoutBuf, stdout)
			}
			if stderr != "" {
				fmt.Fprintln(stderrBuf, stderr)
			}

			if exitCode != 0 {
				return newExitCodeError(exitCode, stderr)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "Message text for the event")
	cmd.Flags().StringVar(&payload, "payload", "{}", "JSON object payload for the event")
	cmd.Flags().StringVar(&claudeSessionID, "claude-session-id", "", "Claude session ID")

	return cmd
}

// validatePayloadFlag validates and returns the raw JSON bytes for the payload flag.
// It accepts only JSON objects. Primitives, arrays, and null are rejected.
func validatePayloadFlag(payload string) (json.RawMessage, error) {
	// Parse to check validity.
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}

	// Determine JSON type: must be an object.
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty JSON")
	}

	if trimmed[0] != '{' {
		return nil, fmt.Errorf("not a JSON object")
	}

	// Verify it's actually an object.
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("not a JSON object")
	}

	return raw, nil
}
