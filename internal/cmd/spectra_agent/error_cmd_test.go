package spectraagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — ErrorCmd ---

func TestErrorCmd_Success(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "something broke"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "Error reported successfully")
}

func TestErrorCmd_WithDetailObject(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "crash", "--detail", `{"stack":"trace","code":500}`}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	payload := parseErrorPayload(t, msg.Payload)
	assertJSONEqual(t, `{"stack":"trace","code":500}`, string(payload.Detail))
}

func TestErrorCmd_WithDetailNull(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "crash", "--detail", "null"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	payload := parseErrorPayload(t, msg.Payload)
	assert.Equal(t, "null", string(payload.Detail))
}

func TestErrorCmd_WithClaudeSessionID(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "oops", "--claude-session-id", "claude-abc"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	assert.Equal(t, "claude-abc", msg.ClaudeSessionID)
}

func TestErrorCmd_WhitespaceOnlyMessageAccepted(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "   "}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	payload := parseErrorPayload(t, msg.Payload)
	assert.Equal(t, "   ", payload.Message)
}

// --- Null / Empty Input ---

func TestErrorCmd_MissingMessage(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "error message is required")
}

func TestErrorCmd_EmptyMessage(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", ""}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "error message is required")
}

// --- Validation Failures — detail ---

func TestErrorCmd_DetailPrimitive(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "msg", "--detail", `"string"`}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--detail must be a JSON object or null")
}

func TestErrorCmd_DetailNumber(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "msg", "--detail", "42"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--detail must be a JSON object or null")
}

func TestErrorCmd_DetailBoolean(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "msg", "--detail", "true"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--detail must be a JSON object or null")
}

func TestErrorCmd_DetailArray(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "msg", "--detail", "[1,2,3]"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--detail must be a JSON object or null")
}

func TestErrorCmd_DetailInvalidJSON(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "error", "msg", "--detail", "{invalid}"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--detail must be a JSON object or null")
}

// --- Mock / Dependency Interaction ---

func TestErrorCmd_ConstructsWireFormat(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/proj"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	executeCommand(t, []string{
		"--session-id", "sess-1",
		"error", "disk full",
		"--detail", `{"disk":"/dev/sda"}`,
		"--claude-session-id", "c-99",
	}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	assert.Equal(t, "error", msg.Type)
	assert.Equal(t, "c-99", msg.ClaudeSessionID)

	payload := parseErrorPayload(t, msg.Payload)
	assert.Equal(t, "disk full", payload.Message)
	assertJSONEqual(t, `{"disk":"/dev/sda"}`, string(payload.Detail))
}

func TestErrorCmd_DefaultDetail(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	executeCommand(t, []string{"--session-id", "abc", "error", "msg"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	payload := parseErrorPayload(t, msg.Payload)
	assertJSONEqual(t, `{}`, string(payload.Detail))
}

func TestErrorCmd_DefaultClaudeSessionID(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	executeCommand(t, []string{"--session-id", "abc", "error", "msg"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	assert.Equal(t, "", msg.ClaudeSessionID)
}

func TestErrorCmd_PassesCorrectSuccessText(t *testing.T) {

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Error reported successfully"}

	executeCommand(t, []string{"--session-id", "abc", "error", "msg"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)
	assert.Equal(t, "Error reported successfully", calls[0].successText)
}
