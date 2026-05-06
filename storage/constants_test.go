package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — Construction ---

func TestMaxPayloadSize_Value(t *testing.T) {
	assert.Equal(t, 10*1024*1024, MaxPayloadSize)
}

func TestMaxPayloadSize_IsConst(t *testing.T) {
	const expected = 10485760
	var v int = MaxPayloadSize
	assert.Equal(t, expected, v)
}
