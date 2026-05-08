package cmdutil

import "fmt"

// FormatError returns a string formatted as "Error: <msg>".
func FormatError(msg string) string {
	return fmt.Sprintf("Error: %s", msg)
}

// FormatWarning returns a string formatted as "Warning: <msg>".
func FormatWarning(msg string) string {
	return fmt.Sprintf("Warning: %s", msg)
}
