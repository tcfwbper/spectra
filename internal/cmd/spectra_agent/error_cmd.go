package spectraagent

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// errorWireMessage represents the wire format for error messages.
type errorWireMessage struct {
	Type            string           `json:"type"`
	ClaudeSessionID string           `json:"claudeSessionID"`
	Payload         errorWirePayload `json:"payload"`
}

// errorWirePayload represents the payload of an error wire message.
type errorWirePayload struct {
	Message string          `json:"message"`
	Detail  json.RawMessage `json:"detail"`
}

func newErrorCmd(ctx *cmdContext, stdoutBuf, stderrBuf *bytes.Buffer) *cobra.Command {
	var detail string
	var claudeSessionID string

	cmd := &cobra.Command{
		Use:           "error <message> [flags]",
		Short:         "Report an error to the runtime",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate positional argument: message.
			if len(args) == 0 || args[0] == "" {
				fmt.Fprintln(stderrBuf, "Error: error message is required")
				return newExitCodeError(1, "error message is required")
			}
			message := args[0]

			// Validate --detail: must be a JSON object or null.
			detailBytes, err := validateDetailFlag(detail)
			if err != nil {
				fmt.Fprintln(stderrBuf, "Error: --detail must be a JSON object or null")
				return newExitCodeError(1, "--detail must be a JSON object or null")
			}

			// Construct wire message.
			wireMsg := errorWireMessage{
				Type:            "error",
				ClaudeSessionID: claudeSessionID,
				Payload: errorWirePayload{
					Message: message,
					Detail:  detailBytes,
				},
			}

			// Delegate to SendAndHandle.
			exitCode, stdout, stderr := ctx.sender.SendAndHandle(
				ctx.sessionID,
				ctx.projectRoot,
				wireMsg,
				"Error reported successfully",
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

	cmd.Flags().StringVar(&detail, "detail", "{}", "JSON object with error details")
	cmd.Flags().StringVar(&claudeSessionID, "claude-session-id", "", "Claude session ID")

	return cmd
}

// validateDetailFlag validates and returns the raw JSON bytes for the detail flag.
// It accepts JSON objects and null. Primitives and arrays are rejected.
func validateDetailFlag(detail string) (json.RawMessage, error) {
	// Parse to check validity.
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(detail), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}

	// Determine JSON type.
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty JSON")
	}

	switch {
	case string(trimmed) == "null":
		return raw, nil
	case trimmed[0] == '{':
		// Verify it's actually an object.
		var obj map[string]any
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, fmt.Errorf("not a JSON object")
		}
		return raw, nil
	default:
		return nil, fmt.Errorf("not a JSON object or null")
	}
}
