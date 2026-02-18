package cli

import "strings"

// ParseDelimitedList parses a string into a list of items based on the provided delimiters.
// Multiple adjacent delimiters are treated as a single delimiter.
func ParseDelimitedList(s string, delims string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return strings.ContainsRune(delims, r)
	})
}
