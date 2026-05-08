package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake Exit & Execute Seams ---

// fakeExitRecorder captures calls to the osExit seam.
type fakeExitRecorder struct {
	mu    sync.Mutex
	codes []int
}

func (f *fakeExitRecorder) exit(code int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes = append(f.codes, code)
}

func (f *fakeExitRecorder) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.codes)
}

func (f *fakeExitRecorder) lastCode() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.codes) == 0 {
		return -1
	}
	return f.codes[len(f.codes)-1]
}

// fakeExecute stubs the execute seam with a configurable return value
// and tracks invocation count.
type fakeExecute struct {
	mu        sync.Mutex
	returnVal int
	callCount int
}

func (f *fakeExecute) fn() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callCount++
	return f.returnVal
}

func (f *fakeExecute) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

// --- Test Setup Helpers ---

// installFakes replaces the package-level osExit and execute seams with fakes,
// returning the fake recorder and a restore function.
// The caller must defer restore() to avoid leaking state across tests.
func installFakes(t *testing.T, exitCode int) (exitRec *fakeExitRecorder, execFake *fakeExecute) {
	t.Helper()

	exitRec = &fakeExitRecorder{}
	execFake = &fakeExecute{returnVal: exitCode}

	origExit := osExit
	origExec := execute

	osExit = exitRec.exit
	execute = execFake.fn

	t.Cleanup(func() {
		osExit = origExit
		execute = origExec
	})

	return exitRec, execFake
}

// --- Happy Path Tests ---

func TestMain_ExitZero(t *testing.T) {
	exitRec, _ := installFakes(t, 0)

	main()

	assert.Equal(t, 1, exitRec.callCount(), "osExit should be called exactly once")
	assert.Equal(t, 0, exitRec.lastCode(), "exit code should be 0")
}

func TestMain_ExitNonZero(t *testing.T) {
	exitRec, _ := installFakes(t, 1)

	main()

	assert.Equal(t, 1, exitRec.callCount(), "osExit should be called exactly once")
	assert.Equal(t, 1, exitRec.lastCode(), "exit code should be 1")
}

func TestMain_ExitCode130(t *testing.T) {
	exitRec, _ := installFakes(t, 130)

	main()

	assert.Equal(t, 1, exitRec.callCount(), "osExit should be called exactly once")
	assert.Equal(t, 130, exitRec.lastCode(), "exit code should be 130")
}

func TestMain_ExitCode143(t *testing.T) {
	exitRec, _ := installFakes(t, 143)

	main()

	assert.Equal(t, 1, exitRec.callCount(), "osExit should be called exactly once")
	assert.Equal(t, 143, exitRec.lastCode(), "exit code should be 143")
}

// --- Mock / Dependency Interaction Tests ---

func TestMain_CallsExecuteExactlyOnce(t *testing.T) {
	_, execFake := installFakes(t, 0)

	main()

	assert.Equal(t, 1, execFake.calls(), "Execute should be called exactly once")
}

func TestMain_NoOtherOsCalls(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", nil, parser.AllErrors)
	require.NoError(t, err, "failed to parse cmd/spectra/main.go")

	var osRefs []string
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Name == "os" {
			osRefs = append(osRefs, "os."+sel.Sel.Name)
		}
		return true
	})

	// The only os.X reference allowed is os.Exit
	for _, ref := range osRefs {
		assert.Equal(t, "os.Exit", ref, "unexpected os reference found: %s; only os.Exit is allowed", ref)
	}
	// Ensure at least one os.Exit reference exists
	require.NotEmpty(t, osRefs, "expected at least one os.Exit reference in main.go")
}
