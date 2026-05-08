package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Happy Path — Constants ---

func TestExitSignalINT_Value(t *testing.T) {
	assert.Equal(t, 130, ExitSignalINT)
}

func TestExitSignalTERM_Value(t *testing.T) {
	assert.Equal(t, 143, ExitSignalTERM)
}

// --- Boundary Values — No Overlap With Base Exit Codes ---

func TestSignalExitCodes_NoOverlapWithBaseCodes(t *testing.T) {
	baseCodes := map[int]bool{0: true, 1: true, 2: true, 3: true}
	assert.False(t, baseCodes[ExitSignalINT], "ExitSignalINT must not overlap with base exit codes")
	assert.False(t, baseCodes[ExitSignalTERM], "ExitSignalTERM must not overlap with base exit codes")
}

func TestSignalExitCodes_Unique(t *testing.T) {
	assert.NotEqual(t, ExitSignalINT, ExitSignalTERM)
}
