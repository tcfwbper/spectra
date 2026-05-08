package cmdutil

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- SendAndHandle function signature (expected from production) ---
// func SendAndHandle(client socketClientSender, formatter errorFormatterFunc, sessionID, projectRoot string, message any, successText string) (exitCode int, stdout string, stderr string)
//
// The production SendAndHandle does not exist yet. Tests below use a callSendAndHandle
// helper that delegates to the production function once it is implemented.
// Until then, tests are scaffolded with t.Skip to name the missing production surface.

// callSendAndHandle is a test helper that calls the production SendAndHandle.
func callSendAndHandle(t *testing.T, client socketClientSender, formatter errorFormatterFunc, sessionID, projectRoot string, message any, successText string) (int, string, string) {
	t.Helper()
	return SendAndHandle(client, formatter, sessionID, projectRoot, message, successText)
}

// --- Happy Path — SendAndHandle ---

func TestSendAndHandle_Success(t *testing.T) {
	client := &mockSocketClient{
		response: &Response{Status: "success", Message: "done"},
		exitCode: 0,
		err:      nil,
	}
	formatter := func(msg string) string {
		t.Errorf("FormatError should not be called on success, but was called with: %s", msg)
		return ""
	}

	exitCode, stdout, stderr := callSendAndHandle(t, client, formatter, "sess-1", "/tmp/project", validStruct{}, "Event emitted")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "Event emitted", stdout)
	assert.Empty(t, stderr)
}

// --- Error Propagation ---

func TestSendAndHandle_SerializationFailure(t *testing.T) {
	// No mock needed — the message itself is unencodable.
	client := &mockSocketClient{
		response: &Response{Status: "success", Message: "ok"},
		exitCode: 0,
		err:      nil,
	}
	formatter := func(msg string) string {
		return "Error: " + msg
	}

	exitCode, stdout, stderr := callSendAndHandle(t, client, formatter, "sess-1", "/tmp/project", unserializableStruct{Ch: make(chan int)}, "ok")

	assert.Equal(t, ExitInvocationError, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "failed to serialize message")
}

func TestSendAndHandle_TransportError(t *testing.T) {
	client := &mockSocketClient{
		response: nil,
		exitCode: 2,
		err:      errors.New("socket file not found: /path"),
	}
	formatter := func(msg string) string {
		return "Error: socket file not found: /path"
	}

	exitCode, stdout, stderr := callSendAndHandle(t, client, formatter, "sess-1", "/tmp/project", validStruct{}, "ok")

	assert.Equal(t, ExitTransportError, exitCode)
	assert.Empty(t, stdout)
	assert.Equal(t, "Error: socket file not found: /path", stderr)
}

func TestSendAndHandle_RuntimeErrorWithResponse(t *testing.T) {
	client := &mockSocketClient{
		response: &Response{Status: "error", Message: "session not found: abc"},
		exitCode: 3,
		err:      nil,
	}
	formatter := func(msg string) string {
		return "Error: session not found: abc"
	}

	exitCode, stdout, stderr := callSendAndHandle(t, client, formatter, "sess-1", "/tmp/project", validStruct{}, "ok")

	assert.Equal(t, ExitRuntimeError, exitCode)
	assert.Empty(t, stdout)
	assert.Equal(t, "Error: session not found: abc", stderr)
}

func TestSendAndHandle_MalformedResponseNilResponse(t *testing.T) {
	client := &mockSocketClient{
		response: nil,
		exitCode: 3,
		err:      errors.New("malformed response from Runtime: ..."),
	}
	formatter := func(msg string) string {
		return "Error: malformed response from Runtime: ..."
	}

	exitCode, stdout, stderr := callSendAndHandle(t, client, formatter, "sess-1", "/tmp/project", validStruct{}, "ok")

	assert.Equal(t, ExitRuntimeError, exitCode)
	assert.Empty(t, stdout)
	assert.Equal(t, "Error: malformed response from Runtime: ...", stderr)
}

// --- Mock / Dependency Interaction ---

func TestSendAndHandle_SerializesMessageToJSON(t *testing.T) {
	client := &mockSocketClient{
		response: &Response{Status: "success", Message: "ok"},
		exitCode: 0,
		err:      nil,
	}
	formatter := func(msg string) string { return "" }

	msg := testMsg{Type: "event", ClaudeSessionID: "c-1"}
	callSendAndHandle(t, client, formatter, "sess-1", "/tmp/project", msg, "ok")

	calls := client.calls()
	require.Len(t, calls, 1)

	expectedJSON, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Equal(t, expectedJSON, calls[0].message)
}

func TestSendAndHandle_PassesSessionIDAndProjectRoot(t *testing.T) {
	client := &mockSocketClient{
		response: &Response{Status: "success", Message: "ok"},
		exitCode: 0,
		err:      nil,
	}
	formatter := func(msg string) string { return "" }

	callSendAndHandle(t, client, formatter, "my-session", "/home/user/project", validStruct{}, "ok")

	calls := client.calls()
	require.Len(t, calls, 1)
	assert.Equal(t, "my-session", calls[0].sessionID)
	assert.Equal(t, "/home/user/project", calls[0].projectRoot)
}

func TestSendAndHandle_CallsFormatErrorOnRuntimeError(t *testing.T) {
	client := &mockSocketClient{
		response: &Response{Status: "error", Message: "bad request"},
		exitCode: 3,
		err:      nil,
	}
	mock := &mockErrorFormatter{result: "Error: bad request"}

	callSendAndHandle(t, client, mock.FormatError, "sess-1", "/tmp/project", validStruct{}, "ok")

	fmtCalls := mock.calls()
	require.Len(t, fmtCalls, 1)
	assert.Equal(t, "bad request", fmtCalls[0])
}

func TestSendAndHandle_DoesNotCallFormatErrorOnSuccess(t *testing.T) {
	client := &mockSocketClient{
		response: &Response{Status: "success", Message: "done"},
		exitCode: 0,
		err:      nil,
	}
	mock := &mockErrorFormatter{result: "Error: should not be called"}

	callSendAndHandle(t, client, mock.FormatError, "sess-1", "/tmp/project", validStruct{}, "ok")

	fmtCalls := mock.calls()
	assert.Empty(t, fmtCalls)
}
