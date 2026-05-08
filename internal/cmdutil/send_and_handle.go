package cmdutil

import (
	"encoding/json"
	"fmt"
)

// socketClientSender defines the interface for SocketClient's Send method.
type socketClientSender interface {
	Send(sessionID, projectRoot string, message []byte) (*Response, int, error)
}

// errorFormatterFunc defines the interface for error formatting.
type errorFormatterFunc func(msg string) string

// SendAndHandle serializes the message to JSON, sends it via the client, interprets
// the response, and returns the final exit code, stdout, and stderr strings.
func SendAndHandle(client socketClientSender, formatter errorFormatterFunc, sessionID, projectRoot string, message any, successText string) (exitCode int, stdout string, stderr string) {
	// Serialize message to JSON.
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return ExitInvocationError, "", formatter(fmt.Sprintf("failed to serialize message: %s", err))
	}

	// Send via client.
	response, code, sendErr := client.Send(sessionID, projectRoot, jsonBytes)

	// Interpret the result.
	switch {
	case code == ExitTransportError:
		// Transport error: exit code 2.
		return ExitTransportError, "", formatter(sendErr.Error())

	case code == ExitRuntimeError && response == nil:
		// Malformed response or missing fields: exit code 3, nil response.
		return ExitRuntimeError, "", formatter(sendErr.Error())

	case code == ExitRuntimeError && response != nil:
		// Runtime error with response: exit code 3, format the response message.
		return ExitRuntimeError, "", formatter(response.Message)

	default:
		// Success: exit code 0.
		return ExitSuccess, successText, ""
	}
}
