package rcs

import (
	"fmt"
	"strings"
)

// ScanTokenStringOrId scans either an @-quoted string or an identifier.
func ScanTokenStringOrId(s *Scanner) (string, error) {
	// Try scanning for a string starting with '@'
	if err := ScanStrings(s, "@"); err == nil {
		// We found '@', consume the rest of the string
		sb := &strings.Builder{}
		for {
			if err := ScanUntilStrings(s, "@"); err != nil {
				return "", err
			}
			sb.WriteString(s.Text())
			if err := ScanStrings(s, "@@", "@"); err != nil {
				return "", err
			}
			nt := s.Text()
			switch nt {
			case "@@":
				sb.WriteString("@")
			case "@":
				return sb.String(), nil
			default:
				return "", fmt.Errorf("unexpected token %q", nt)
			}
		}
	} else {
		// If not a string (starts with something else), try scanning as an ID
		if IsNotFound(err) {
			return ScanTokenId(s)
		}
		return "", err
	}
}
