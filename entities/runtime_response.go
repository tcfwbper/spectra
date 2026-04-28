package entities

import (
	"fmt"
)

// RuntimeResponse represents the response returned by RuntimeSocketManager to spectra-agent
// after processing a RuntimeMessage.
type RuntimeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewRuntimeResponse creates a new RuntimeResponse with validation.
// It validates that:
// - Status is either "success" or "error"
// - Message is a valid string (may be empty)
func NewRuntimeResponse(status string, message string) (*RuntimeResponse, error) {
	// Validate status
	if status != "success" && status != "error" {
		return nil, fmt.Errorf("invalid response status: must be 'success' or 'error'")
	}

	return &RuntimeResponse{
		Status:  status,
		Message: message,
	}, nil
}

// Validate validates a RuntimeResponse.
func (rr *RuntimeResponse) Validate() error {
	if rr.Status != "success" && rr.Status != "error" {
		return fmt.Errorf("invalid response status: must be 'success' or 'error'")
	}
	return nil
}
