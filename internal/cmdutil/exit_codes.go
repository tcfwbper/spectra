package cmdutil

// ExitSuccess indicates the operation completed successfully.
const ExitSuccess = 0

// ExitInvocationError indicates a missing required argument/flag, invalid flag
// value, invalid JSON, .spectra directory not found, or unknown subcommand.
const ExitInvocationError = 1

// ExitTransportError indicates a socket file not found, connection refused,
// connection timeout, or I/O error during send/receive.
const ExitTransportError = 2

// ExitRuntimeError indicates the runtime responded with error status, malformed
// response JSON, or response missing required fields.
const ExitRuntimeError = 3
