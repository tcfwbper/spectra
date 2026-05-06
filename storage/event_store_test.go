package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectra-ai/spectra/entities"
	"github.com/spectra-ai/spectra/logger"
)

// --- Mock Logger for EventStore tests ---

// mockLogger records Warn calls for assertion purposes.
type mockLogger struct {
	logger.NopLogger
	mu       sync.Mutex
	warnMsgs []string
	warnArgs [][]any
}

func newMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnMsgs = append(m.warnMsgs, msg)
	m.warnArgs = append(m.warnArgs, args)
}

func (m *mockLogger) warnCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.warnMsgs)
}

// --- Fixture Builders for EventStore tests ---

// makeValidEvent creates a valid Event entity for testing.
func makeValidEvent(t *testing.T, message string) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskCompleted",
		message,
		json.RawMessage(`{"key":"value"}`),
		"agent-node",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)
	require.NoError(t, err)
	return ev
}

// makeValidEventWithID creates a valid Event entity with a specific ID.
func makeValidEventWithID(t *testing.T, id string) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		id,
		"TaskCompleted",
		"test message",
		json.RawMessage(`{"key":"value"}`),
		"agent-node",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)
	require.NoError(t, err)
	return ev
}

// makeEventWithPayload creates a valid Event with a specific payload.
func makeEventWithPayload(t *testing.T, payload json.RawMessage) *entities.Event {
	t.Helper()
	ev, err := entities.NewEvent(
		"550e8400-e29b-41d4-a716-446655440000",
		"TaskCompleted",
		"test message",
		payload,
		"agent-node",
		1700000000,
		"660e8400-e29b-41d4-a716-446655440000",
	)
	require.NoError(t, err)
	return ev
}

// makeSessionDirFixture creates a temp directory with the session dir structure.
func makeSessionDirFixture(t *testing.T) (projectRoot string, sessionDir string) {
	t.Helper()
	projectRoot = makeTempDirWithSessionDir(t, testSessionUUID)
	sessionDir = filepath.Join(projectRoot, ".spectra", "sessions", testSessionUUID)
	return projectRoot, sessionDir
}

// writeEventsFile writes event JSON lines to an events.jsonl file in the session dir.
func writeEventsFile(t *testing.T, sessionDir string, lines []string) string {
	t.Helper()
	filePath := filepath.Join(sessionDir, EventHistoryFile)
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}

// --- Happy Path — Construction ---

func TestNewEventStore_ValidInputs(t *testing.T) {
	ml := newMockLogger()

	es := NewEventStore("/tmp/project", testSessionUUID, ml)
	require.NotNil(t, es)
}

func TestNewEventStore_NoFileSystemAccess(t *testing.T) {
	ml := newMockLogger()

	// Provide a non-existent projectRoot — constructor must not touch filesystem.
	es := NewEventStore("/nonexistent", testSessionUUID, ml)
	require.NotNil(t, es)
}

// --- Happy Path — Append ---

func TestEventStore_Append_FirstEvent(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "hello")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	err := es.Append(ev)
	require.NoError(t, err)

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, readErr := os.ReadFile(filePath)
	require.NoError(t, readErr)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 1)
	assert.True(t, strings.HasSuffix(string(data), "\n"))
}

func TestEventStore_Append_MultipleEvents(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev1 := makeValidEvent(t, "first")
	ev2 := makeValidEvent(t, "second")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev1))
	require.NoError(t, es.Append(ev2))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, _ := os.ReadFile(filePath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 2)
}

func TestEventStore_Append_CompactJSON(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeEventWithPayload(t, json.RawMessage(`{"a":"1","b":"2"}`))

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, _ := os.ReadFile(filePath)
	line := strings.TrimSpace(string(data))
	assert.NotContains(t, line, "  ")
	assert.NotContains(t, line, "\t")
}

func TestEventStore_Append_MessageFieldLast(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "hello world")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, _ := os.ReadFile(filePath)
	line := strings.TrimSpace(string(data))
	// Verify "Message" is the last key in the JSON object
	lastBrace := strings.LastIndex(line, "}")
	beforeBrace := line[:lastBrace]
	lastKey := strings.LastIndex(beforeBrace, `"Message"`)
	// Ensure no other key appears after "Message"
	segment := beforeBrace[lastKey+len(`"Message"`):]
	// Should not contain another key (no additional `":` pattern after the message value start)
	assert.NotContains(t, segment, `":`+`"`)
}

func TestEventStore_Append_EscapesNewlinesInMessage(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "line1\nline2")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, _ := os.ReadFile(filePath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 1, "event with newline in message should serialize as single line")
	assert.Contains(t, lines[0], `\n`)
}

func TestEventStore_Append_EmptyPayload(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeEventWithPayload(t, json.RawMessage(`{}`))

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, _ := os.ReadFile(filePath)
	line := strings.TrimSpace(string(data))
	assert.Contains(t, line, `"Payload":{}`)
}

// --- Happy Path — Read ---

func TestEventStore_Read_MultipleEvents(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	// Pre-populate fixture with 3 valid event JSON lines.
	lines := []string{
		`{"ID":"550e8400-e29b-41d4-a716-446655440000","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"first"}`,
		`{"ID":"550e8400-e29b-41d4-a716-446655440001","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000001,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"second"}`,
		`{"ID":"550e8400-e29b-41d4-a716-446655440002","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000002,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"third"}`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Len(t, events, 3)
}

func TestEventStore_Read_FileNotExists(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)
	ml := newMockLogger()

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Empty(t, events)
	assert.Equal(t, 0, ml.warnCallCount())
}

func TestEventStore_Read_PreservesOrder(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	// Create file with events having distinct IDs in known order.
	lines := []string{
		`{"ID":"550e8400-e29b-41d4-a716-446655440001","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"first"}`,
		`{"ID":"550e8400-e29b-41d4-a716-446655440002","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000001,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"second"}`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", events[0].ID())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", events[1].ID())
}

// --- Error Propagation ---

func TestEventStore_Append_SessionDirNotExists(t *testing.T) {
	// Create temp dir without the session subdirectory.
	projectRoot := makeTempDirWithSessions(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "test")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	err := es.Append(ev)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session directory does not exist:")
}

func TestEventStore_Append_FileAccessorError(t *testing.T) {
	// Use a projectRoot that doesn't have sessions dir at all — FileAccessor callback returns error.
	projectRoot := makeTempDirWithSessions(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "test")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	err := es.Append(ev)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session directory does not exist:")
}

func TestEventStore_Append_ExceedsMaxPayloadSize(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)
	ml := newMockLogger()

	// Construct an Event with a very large message (> 10 MB).
	largeMessage := strings.Repeat("x", MaxPayloadSize+1)
	ev := makeValidEvent(t, largeMessage)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	err := es.Append(ev)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event size exceeds limit:")
	assert.Contains(t, err.Error(), "bytes (max")
}

func TestEventStore_Append_ExceedsMaxPayloadSize_NoWrite(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	// Pre-populate with one event.
	smallEv := makeValidEvent(t, "small")
	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(smallEv))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	dataBefore, _ := os.ReadFile(filePath)

	// Attempt oversized append.
	largeMessage := strings.Repeat("x", MaxPayloadSize+1)
	largeEv := makeValidEvent(t, largeMessage)
	err := es.Append(largeEv)
	require.Error(t, err)

	// File content should remain unchanged (still one line).
	dataAfter, _ := os.ReadFile(filePath)
	assert.Equal(t, string(dataBefore), string(dataAfter))
}

// --- Validation Failures ---

func TestEventStore_Read_MalformedJSON(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	lines := []string{
		`{"ID":"550e8400-e29b-41d4-a716-446655440000","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"ok"}`,
		`{"ID":"550e8400-e29b-41d4-a716-446655440001"`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.GreaterOrEqual(t, ml.warnCallCount(), 1)
}

func TestEventStore_Read_BlankLine(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	lines := []string{
		`{"ID":"550e8400-e29b-41d4-a716-446655440000","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"first"}`,
		"",
		`{"ID":"550e8400-e29b-41d4-a716-446655440001","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000001,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"second"}`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, 1, ml.warnCallCount())
}

func TestEventStore_Read_MissingRequiredFields(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	lines := []string{
		`{"Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"no id"}`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Empty(t, events)
	assert.GreaterOrEqual(t, ml.warnCallCount(), 1)
}

// --- Mock / Dependency Interaction ---

func TestEventStore_Append_CallsFileAccessor(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "test")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev))

	// Verify FileAccessor was called — the file should exist.
	filePath := filepath.Join(sessionDir, EventHistoryFile)
	_, err := os.Stat(filePath)
	assert.NoError(t, err)
}

func TestEventStore_Append_ReadsEventViaGetters(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()
	ev := makeValidEvent(t, "getter check")

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	require.NoError(t, es.Append(ev))

	filePath := filepath.Join(sessionDir, EventHistoryFile)
	data, _ := os.ReadFile(filePath)
	var parsed map[string]any
	json.Unmarshal(data[:len(data)-1], &parsed)
	assert.Equal(t, ev.ID(), parsed["ID"])
	assert.Equal(t, ev.Type(), parsed["Type"])
	assert.Equal(t, ev.Message(), parsed["Message"])
}

func TestEventStore_Read_LogsWarningForEachMalformedLine(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	lines := []string{
		`{"ID":"550e8400-e29b-41d4-a716-446655440000","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"ok"}`,
		`{broken`,
		`also{broken`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events, err := es.Read()
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, 2, ml.warnCallCount())
}

// --- Idempotency ---

func TestEventStore_Read_IdempotentReads(t *testing.T) {
	projectRoot, sessionDir := makeSessionDirFixture(t)
	ml := newMockLogger()

	lines := []string{
		`{"ID":"550e8400-e29b-41d4-a716-446655440000","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000000,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"first"}`,
		`{"ID":"550e8400-e29b-41d4-a716-446655440001","Type":"TaskCompleted","Payload":{},"EmittedBy":"agent","EmittedAt":1700000001,"SessionID":"660e8400-e29b-41d4-a716-446655440000","Message":"second"}`,
	}
	writeEventsFile(t, sessionDir, lines)

	es := NewEventStore(projectRoot, testSessionUUID, ml)
	events1, err1 := es.Read()
	require.NoError(t, err1)
	warnCount1 := ml.warnCallCount()

	events2, err2 := es.Read()
	require.NoError(t, err2)
	warnCount2 := ml.warnCallCount() - warnCount1

	assert.Equal(t, len(events1), len(events2))
	assert.Equal(t, warnCount1, warnCount2)
}

// --- Boundary Values — MaxPayloadSize ---

func TestEventStore_Append_ExactlyAtMaxPayloadSize(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)
	ml := newMockLogger()

	// Create an event, serialize to measure overhead, then build one with exact size.
	baseEv := makeValidEvent(t, "")
	es := NewEventStore(projectRoot, testSessionUUID, ml)

	// Serialize a baseline to compute overhead.
	type orderedEvent struct {
		ID        string          `json:"ID"`
		Type      string          `json:"Type"`
		Payload   json.RawMessage `json:"Payload"`
		EmittedBy string          `json:"EmittedBy"`
		EmittedAt int64           `json:"EmittedAt"`
		SessionID string          `json:"SessionID"`
		Message   string          `json:"Message"`
	}
	oe := orderedEvent{
		ID: baseEv.ID(), Type: baseEv.Type(), Payload: baseEv.Payload(),
		EmittedBy: baseEv.EmittedBy(), EmittedAt: baseEv.EmittedAt(),
		SessionID: baseEv.SessionID(), Message: "",
	}
	overhead, _ := json.Marshal(oe)
	// The message gets JSON-encoded, so a plain alphanumeric string adds 2 bytes for quotes.
	// Overhead includes `"Message":""` with empty message. We need total = MaxPayloadSize.
	// With a message of length N (alphanumeric), serialized JSON size = len(overhead) + N.
	msgLen := MaxPayloadSize - len(overhead)
	if msgLen < 0 {
		t.Skip("overhead alone exceeds MaxPayloadSize")
	}
	exactMsg := strings.Repeat("a", msgLen)
	ev := makeValidEvent(t, exactMsg)

	err := es.Append(ev)
	require.NoError(t, err)
}

func TestEventStore_Append_OneByteOverMaxPayloadSize(t *testing.T) {
	projectRoot, _ := makeSessionDirFixture(t)
	ml := newMockLogger()

	// Create an event, serialize to measure overhead, then build one that's one byte over.
	baseEv := makeValidEvent(t, "")
	es := NewEventStore(projectRoot, testSessionUUID, ml)

	type orderedEvent struct {
		ID        string          `json:"ID"`
		Type      string          `json:"Type"`
		Payload   json.RawMessage `json:"Payload"`
		EmittedBy string          `json:"EmittedBy"`
		EmittedAt int64           `json:"EmittedAt"`
		SessionID string          `json:"SessionID"`
		Message   string          `json:"Message"`
	}
	oe := orderedEvent{
		ID: baseEv.ID(), Type: baseEv.Type(), Payload: baseEv.Payload(),
		EmittedBy: baseEv.EmittedBy(), EmittedAt: baseEv.EmittedAt(),
		SessionID: baseEv.SessionID(), Message: "",
	}
	overhead, _ := json.Marshal(oe)
	msgLen := MaxPayloadSize - len(overhead) + 1
	overMsg := strings.Repeat("a", msgLen)
	ev := makeValidEvent(t, overMsg)

	err := es.Append(ev)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event size exceeds limit:")
}

