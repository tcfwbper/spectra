package spectraagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy Path — EventEmitCmd ---

func TestEventEmitCmd_Success(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "ReviewNeeded"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)
	assert.Contains(t, result.stdout, "Event emitted successfully")
}

func TestEventEmitCmd_WithMessage(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "BuildDone", "--message", "build completed"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	payload := parseEventPayload(t, msg.Payload)
	assert.Equal(t, "build completed", payload.Message)
}

func TestEventEmitCmd_WithPayloadObject(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Deploy", "--payload", `{"env":"prod","version":"1.2"}`}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	payload := parseEventPayload(t, msg.Payload)
	assertJSONEqual(t, `{"env":"prod","version":"1.2"}`, string(payload.Payload))
}

func TestEventEmitCmd_WithClaudeSessionID(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Ping", "--claude-session-id", "claude-xyz"}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	assert.Equal(t, "claude-xyz", msg.ClaudeSessionID)
}

func TestEventEmitCmd_EmptyMessageAccepted(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Notify", "--message", ""}, finder, sender)

	assert.Equal(t, 0, result.exitCode)

	calls := sender.calls()
	require.Len(t, calls, 1)
	msg := parseWireMessage(t, calls[0].message)
	payload := parseEventPayload(t, msg.Payload)
	assert.Equal(t, "", payload.Message)
}

// --- Null / Empty Input ---

func TestEventEmitCmd_MissingType(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "event type is required")
}

func TestEventEmitCmd_EmptyType(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", ""}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "event type is required")
}

// --- Validation Failures — payload ---

func TestEventEmitCmd_PayloadPrimitive(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Evt", "--payload", `"hello"`}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--payload must be a valid JSON object, e.g., {}")
}

func TestEventEmitCmd_PayloadArray(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Evt", "--payload", "[1,2,3]"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--payload must be a valid JSON object, e.g., {}")
}

func TestEventEmitCmd_PayloadInvalidJSON(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Evt", "--payload", "{invalid}"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--payload must be a valid JSON object, e.g., {}")
}

func TestEventEmitCmd_PayloadNull(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0}

	result := executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Evt", "--payload", "null"}, finder, sender)

	assert.Equal(t, 1, result.exitCode)
	assert.Contains(t, result.stderr, "--payload must be a valid JSON object, e.g., {}")
}

// --- Mock / Dependency Interaction ---

func TestEventEmitCmd_ConstructsWireFormat(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/proj"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	executeCommand(t, []string{
		"--session-id", "sess-1",
		"event", "emit", "ReviewNeeded",
		"--message", "changes staged",
		"--payload", `{"pr":42}`,
		"--claude-session-id", "c-1",
	}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	assert.Equal(t, "event", msg.Type)
	assert.Equal(t, "c-1", msg.ClaudeSessionID)

	payload := parseEventPayload(t, msg.Payload)
	assert.Equal(t, "ReviewNeeded", payload.EventType)
	assert.Equal(t, "changes staged", payload.Message)
	assertJSONEqual(t, `{"pr":42}`, string(payload.Payload))
}

func TestEventEmitCmd_DefaultPayload(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Ping"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	payload := parseEventPayload(t, msg.Payload)
	assertJSONEqual(t, `{}`, string(payload.Payload))
}

func TestEventEmitCmd_DefaultMessage(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Ping"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	payload := parseEventPayload(t, msg.Payload)
	assert.Equal(t, "", payload.Message)
}

func TestEventEmitCmd_DefaultClaudeSessionID(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Ping"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)

	msg := parseWireMessage(t, calls[0].message)
	assert.Equal(t, "", msg.ClaudeSessionID)
}

func TestEventEmitCmd_PassesCorrectSuccessText(t *testing.T) {
	t.Skip("blocked: production function spectraagent.Execute or NewRootCmd with dependency injection seams does not exist yet")

	finder := &fakeSpectraFinder{projectRoot: "/tmp/project"}
	sender := &fakeSendAndHandle{exitCode: 0, stdout: "Event emitted successfully"}

	executeCommand(t, []string{"--session-id", "abc", "event", "emit", "Ping"}, finder, sender)

	calls := sender.calls()
	require.Len(t, calls, 1)
	assert.Equal(t, "Event emitted successfully", calls[0].successText)
}
