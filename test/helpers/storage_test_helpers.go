package helpers

// SplitLines splits content by newlines and filters out empty lines.
// Used for parsing JSONL files in tests.
func SplitLines(content string) []string {
	lines := []string{}
	current := ""
	for _, c := range content {
		if c == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
