package cmdutil

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers ---

// errWriter is a writer that always returns an error on Write.
type errWriter struct {
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

// errReader is a reader that always returns an error on Read.
type errReader struct {
	err error
}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, r.err
}

// --- Happy Path — ConfirmPrompt ---

func TestConfirmPrompt_LowercaseY(t *testing.T) {
	reader := strings.NewReader("y\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, "Confirm? [y/N]: ", writer.String())
}

func TestConfirmPrompt_UppercaseY(t *testing.T) {
	reader := strings.NewReader("Y\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.True(t, result)
}

func TestConfirmPrompt_YWithWhitespace(t *testing.T) {
	reader := strings.NewReader("  y  \n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.True(t, result)
}

func TestConfirmPrompt_PromptWritten(t *testing.T) {
	reader := strings.NewReader("n\n")
	writer := &bytes.Buffer{}

	_, err := ConfirmPrompt(reader, writer, "Delete all? [y/N]: ")

	require.NoError(t, err)
	assert.Equal(t, "Delete all? [y/N]: ", writer.String())
}

// --- Happy Path — Rejection ---

func TestConfirmPrompt_EmptyInput(t *testing.T) {
	reader := strings.NewReader("\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirmPrompt_LowercaseN(t *testing.T) {
	reader := strings.NewReader("n\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirmPrompt_UppercaseN(t *testing.T) {
	reader := strings.NewReader("N\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirmPrompt_Yes(t *testing.T) {
	reader := strings.NewReader("yes\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirmPrompt_No(t *testing.T) {
	reader := strings.NewReader("no\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirmPrompt_ArbitraryText(t *testing.T) {
	reader := strings.NewReader("maybe\n")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

// --- Null / Empty Input ---

func TestConfirmPrompt_EOF(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirmPrompt_ReaderClosed(t *testing.T) {
	reader := &errReader{err: io.ErrClosedPipe}
	writer := &bytes.Buffer{}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.NoError(t, err)
	assert.False(t, result)
}

// --- Error Propagation ---

func TestConfirmPrompt_WriterError(t *testing.T) {
	reader := strings.NewReader("y\n")
	writer := &errWriter{err: errors.New("broken pipe")}

	result, err := ConfirmPrompt(reader, writer, "Confirm? [y/N]: ")

	require.Error(t, err)
	assert.False(t, result)
}
