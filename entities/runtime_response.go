package entities

// RuntimeResponse is the structured response entity returned by the runtime to
// spectra-agent after processing a RuntimeMessage. It is a pure data entity —
// it does not perform serialization, transmission, or connection management.
type RuntimeResponse struct {
	status  string
	message string
}

// SuccessResponse creates an immutable RuntimeResponse with status "success".
func SuccessResponse(message string) *RuntimeResponse {
	return &RuntimeResponse{
		status:  "success",
		message: message,
	}
}

// ErrorResponse creates an immutable RuntimeResponse with status "error".
func ErrorResponse(message string) *RuntimeResponse {
	return &RuntimeResponse{
		status:  "error",
		message: message,
	}
}

// Status returns the response status ("success" or "error").
func (r *RuntimeResponse) Status() string { return r.status }

// Message returns the human-readable result description or error details.
func (r *RuntimeResponse) Message() string { return r.message }
