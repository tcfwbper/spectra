package components

import "unicode"

// isPascalCase checks whether s is a valid PascalCase identifier:
// starts with an uppercase letter and contains only alphanumeric characters.
func isPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsUpper(r) {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return false
			}
		}
	}
	return true
}
