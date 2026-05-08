package cmdutil

// ExitSignalINT is the exit code for process terminated by SIGINT (128 + 2).
// Standard Unix convention for Ctrl+C.
const ExitSignalINT = 130

// ExitSignalTERM is the exit code for process terminated by SIGTERM (128 + 15).
// Standard Unix convention for kill default signal.
const ExitSignalTERM = 143
