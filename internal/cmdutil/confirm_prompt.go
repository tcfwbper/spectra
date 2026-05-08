package cmdutil

import (
	"bufio"
	"io"
	"strings"
)

// ConfirmPrompt writes the given prompt to writer, reads one line from reader,
// and returns true only if the trimmed input is "y" or "Y". All other input
// (including empty, EOF, or read failure) is treated as rejection. If writing
// the prompt fails, it returns (false, error).
func ConfirmPrompt(reader io.Reader, writer io.Writer, prompt string) (bool, error) {
	if _, err := io.WriteString(writer, prompt); err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		// EOF or read error — treat as rejection, not an error.
		return false, nil
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "y" || input == "Y" {
		return true, nil
	}

	return false, nil
}
