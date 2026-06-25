package runtime

import "io"

// --- Bridge: mockCommandStarter implements CommandStarter and CommandHandle ---

// Command implements the CommandStarter interface for mockCommandStarter.
// It captures the command name and arguments, then returns itself as the handle.
func (m *mockCommandStarter) Command(name string, args ...string) CommandHandle {
	m.path = name
	m.args = args
	return m
}

// SetDir implements the CommandHandle interface for mockCommandStarter.
func (m *mockCommandStarter) SetDir(dir string) {
	m.dir = dir
}

// SetEnv implements the CommandHandle interface for mockCommandStarter.
func (m *mockCommandStarter) SetEnv(env []string) {
	m.env = env
}

// SetStdout implements the CommandHandle interface for mockCommandStarter.
func (m *mockCommandStarter) SetStdout(_ io.Writer) {
	m.stdoutSet = true
}

// SetStderr implements the CommandHandle interface for mockCommandStarter.
func (m *mockCommandStarter) SetStderr(_ io.Writer) {
	m.stderrSet = true
}

// Start implements the CommandHandle interface for mockCommandStarter.
func (m *mockCommandStarter) Start() error {
	m.startCalled++
	return m.startErr
}

// Pid implements the CommandHandle interface for mockCommandStarter.
func (m *mockCommandStarter) Pid() int {
	return m.pid
}
