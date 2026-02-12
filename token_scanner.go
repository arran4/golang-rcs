package rcs

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Grammar Rules:
// num       ::=  { digit | "." }+
// digit     ::=  "0" through "9"
// id        ::=  { idchar | "." }+
// sym       ::=  {idchar}+
// idchar    ::=  any visible graphic character except special
// special   ::=  "$" | "," | "." | ":" | ";" | "@"
// string    ::=  "@" { any character, with @ doubled }* "@"
// intstring ::= "@" {intchar}* {thirdp}* "@"
// thirdp    ::=  "^L" {intchar}*
// intchar   ::=  any character, except @

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isSpecial(r rune) bool {
	switch r {
	case '$', ',', '.', ':', ';', '@':
		return true
	}
	return false
}

func isIdChar(r rune) bool {
	// any visible graphic character except special
	// "visible graphic character" usually excludes control chars and maybe space?
	// RCS idchar excludes space.
	if unicode.IsSpace(r) {
		return false
	}
	if !unicode.IsGraphic(r) {
		return false
	}
	return !isSpecial(r)
}

func ScanTokenNum(s *Scanner) (string, error) {
	// num ::= { digit | "." }+
	// num cannot be empty, but fields like 'next' can have optional num.
	// If this function is called, it expects to find a number.
	// However, if we are integrating, we might call it when we are not sure if there is a number.
	// If minimum 1 is enforced, we must check beforehand.

	err := ScanRunesUntil(s, 1, func(b []byte) bool {
		r := bytes.Runes(b)[0]
		return !isDigit(r) && r != '.'
	}, "num")
	if err != nil {
		return "", err
	}
	return s.Text(), nil
}

func ScanTokenId(s *Scanner) (string, error) {
	// id ::= { idchar | "." }+
	// idchar excludes '.', so id is (idchar OR '.')
	// Basically: visible graphic except (special - '.')
	err := ScanRunesUntil(s, 1, func(b []byte) bool {
		r := bytes.Runes(b)[0]
		return !isIdChar(r) && r != '.'
	}, "id")
	if err != nil {
		return "", err
	}
	return s.Text(), nil
}

func ScanTokenSym(s *Scanner) (string, error) {
	// sym ::= {idchar}+
	// idchar excludes '.'
	err := ScanRunesUntil(s, 1, func(b []byte) bool {
		r := bytes.Runes(b)[0]
		return !isIdChar(r)
	}, "sym")
	if err != nil {
		return "", err
	}
	return s.Text(), nil
}

func ScanTokenString(s *Scanner) (string, error) {
	// string ::= "@" { any character, with @ doubled }* "@"
	return ParseAtQuotedString(s)
}

func ScanTokenStringOrId(s *Scanner) (string, error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if len(data) == 0 {
			if atEOF {
				return 0, nil, nil
			}
			return 0, nil, nil
		}

		if data[0] == '@' {
			// It's a string. Parse as quoted string.
			// String ::= "@" { any character, with @ doubled }* "@"

			// Advance past first @
			i := 1
			for i < len(data) {
				if data[i] == '@' {
					// Check if doubled
					if i+1 < len(data) {
						if data[i+1] == '@' {
							i += 2
							continue
						}
					} else if !atEOF {
						// We found '@' at end of buffer, need more data to check if doubled
						return 0, nil, nil
					}
					// Found single '@', end of string.
					// Return token including quotes.
					return i + 1, data[:i+1], nil
				}
				i++
			}
			if atEOF {
				return 0, nil, fmt.Errorf("unexpected EOF in string")
			}
			return 0, nil, nil // Request more data
		} else {
			// It's an ID.
			// Consume until !IdChar and !'.'
			i := 0
			for i < len(data) {
				r, size := utf8.DecodeRune(data[i:])
				if r == utf8.RuneError {
					// If buffer is too short for a full rune, and not at EOF, wait.
					if size == 0 { // Should not happen with len(data) > i check
						if !atEOF {
							return 0, nil, nil
						}
					} else if size == 1 && !atEOF && len(data)-i < 4 { // Simplified check for potential incomplete rune
						// utf8.DecodeRune returns RuneError for incomplete runes too?
						// It returns RuneError, 1 if invalid.
						// If it's partial, we might need more data.
						// But assuming valid UTF8 or ASCII for now.
					}
				}

				// Check if it is valid ID char OR '.'
				if !isIdChar(r) && r != '.' {
					if i == 0 {
						// Invalid char at start.
						// If we treat this as "Token not found", we should maybe return error?
						// But ScanTokenId behavior is to fail if 0 chars found?
						// ScanRunesUntil with min 1 returns ScanUntilNotFound.
						return 0, nil, ScanNotFound{LookingFor: []string{"id"}, Pos: *s.pos, Found: string(r)}
					}
					return i, data[:i], nil
				}
				i += size
			}
			if atEOF {
				if i > 0 {
					return i, data[:i], nil
				}
				return 0, nil, nil // Or ScanNotFound? If empty at EOF
			}
			return 0, nil, nil // Need more data
		}
	})

	if !s.Scan() {
		if s.Err() != nil {
			return "", s.Err()
		}
		// EOF or empty
		return "", nil
	}

	text := s.Text()
	if len(text) > 0 && text[0] == '@' {
		// Unquote: remove start/end @ and replace @@ with @
		content := text[1 : len(text)-1]
		return strings.ReplaceAll(content, "@@", "@"), nil
	}
	return text, nil
}

func ScanTokenIntString(s *Scanner) (string, error) {
	// intstring ::= "@" {intchar}* {thirdp}* "@"
	if err := ScanStrings(s, "@"); err != nil {
		return "", err
	}
	if err := ScanUntilStrings(s, "@"); err != nil {
		return "", err
	}
	val := s.Text()
	if err := ScanStrings(s, "@"); err != nil {
		return "", err
	}
	return val, nil
}
