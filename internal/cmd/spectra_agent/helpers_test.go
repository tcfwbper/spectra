package spectraagent

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- Fake: SpectraFinder ---

// fakeSpectraFinder stubs FindProjectRoot for testing the root command's
// project-root discovery behavior without touching the filesystem.
type fakeSpectraFinder struct {
	mu sync.Mutex

	// Return values
	projectRoot string
	err         error

	// Captured arguments
	calledWith []string
}

func (f *fakeSpectraFinder) FindProjectRoot(startDir string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calledWith = append(f.calledWith, startDir)
	return f.projectRoot, f.err
}

func (f *fakeSpectraFinder) calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calledWith...)
}

// --- Fake: SendAndHandle ---

// sendAndHandleCall records a single invocation of the SendAndHandle seam.
type sendAndHandleCall struct {
	sessionID   string
	projectRoot string
	message     json.RawMessage
	successText string
}

// fakeSendAndHandle captures calls to SendAndHandle and returns a configured exit code.
type fakeSendAndHandle struct {
	mu sync.Mutex

	// Return values
	exitCode int
	stdout   string
	stderr   string

	// Captured arguments
	calledWith []sendAndHandleCall
}

func (f *fakeSendAndHandle) SendAndHandle(sessionID, projectRoot string, message any, successText string) (int, string, string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Serialize message so tests can inspect the wire format.
	msgBytes, _ := json.Marshal(message)

	f.calledWith = append(f.calledWith, sendAndHandleCall{
		sessionID:   sessionID,
		projectRoot: projectRoot,
		message:     json.RawMessage(msgBytes),
		successText: successText,
	})
	return f.exitCode, f.stdout, f.stderr
}

func (f *fakeSendAndHandle) calls() []sendAndHandleCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sendAndHandleCall(nil), f.calledWith...)
}

// --- Command execution helper ---

// executeResult holds the result of running the CLI command in a test.
type executeResult struct {
	exitCode int
	stdout   string
	stderr   string
}

// executeCommand runs the spectra-agent CLI with given args and fakes, capturing output.
func executeCommand(t *testing.T, args []string, finder *fakeSpectraFinder, sender *fakeSendAndHandle) executeResult {
	t.Helper()

	exitCode, stdout, stderr := RunForResult(Options{
		Finder: finder,
		Sender: sender,
		Args:   args,
	})

	return executeResult{
		exitCode: exitCode,
		stdout:   stdout,
		stderr:   stderr,
	}
}

// --- Wire format structs for assertion ---
// These mirror the expected RuntimeMessage wire format for deserialization in assertions.

// wireMessage represents the top-level RuntimeMessage wire format.
type wireMessage struct {
	Type            string          `json:"type"`
	ClaudeSessionID string          `json:"claudeSessionID"`
	Payload         json.RawMessage `json:"payload"`
}

// errorPayload represents the payload for error-type messages.
type errorPayload struct {
	Message string          `json:"message"`
	Detail  json.RawMessage `json:"detail"`
}

// eventPayload represents the payload for event-type messages.
type eventPayload struct {
	EventType string          `json:"eventType"`
	Message   string          `json:"message"`
	Payload   json.RawMessage `json:"payload"`
}

// --- Assertion helpers ---

// parseWireMessage deserializes the captured message bytes into a wireMessage.
func parseWireMessage(t *testing.T, raw json.RawMessage) wireMessage {
	t.Helper()
	var msg wireMessage
	require.NoError(t, json.Unmarshal(raw, &msg), "failed to parse wire message")
	return msg
}

// parseErrorPayload deserializes the wire message payload into an errorPayload.
func parseErrorPayload(t *testing.T, raw json.RawMessage) errorPayload {
	t.Helper()
	var p errorPayload
	require.NoError(t, json.Unmarshal(raw, &p), "failed to parse error payload")
	return p
}

// parseEventPayload deserializes the wire message payload into an eventPayload.
func parseEventPayload(t *testing.T, raw json.RawMessage) eventPayload {
	t.Helper()
	var p eventPayload
	require.NoError(t, json.Unmarshal(raw, &p), "failed to parse event payload")
	return p
}

// assertJSONEqual asserts that two JSON values are semantically equal.
func assertJSONEqual(t *testing.T, expected, actual string) {
	t.Helper()
	var expBuf, actBuf bytes.Buffer
	require.NoError(t, json.Compact(&expBuf, []byte(expected)), "invalid expected JSON")
	require.NoError(t, json.Compact(&actBuf, []byte(actual)), "invalid actual JSON")
	require.Equal(t, expBuf.String(), actBuf.String())
}
