package rcs

import (
	"bytes"
	"fmt"
	"io"
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

func ScanTokenWord(s *Scanner) (string, error) {
	// word ::= id | num | string | ":"
	// string starts with @
	// id starts with idchar
	// num starts with digit or .
	// : is :

	// Note: This function assumes whitespace has already been consumed (e.g. by ParseOptionalToken).
	// It relies on rcs.Scanner wrapping bufio.Scanner and allowing dynamic switching of SplitFunc via s.Split.

	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		first := data[0]

		if first == '@' {
			// Quoted string logic
			// Look for matching @ that is not doubled (@@)
			i := 1
			for i < len(data) {
				if data[i] == '@' {
					// Check if next char is also @ (doubled)
					if i+1 < len(data) {
						if data[i+1] == '@' {
							i += 2 // Skip @@
							continue
						}
					} else if !atEOF {
						// Need more data to check if next char is @
						return 0, nil, nil
					}
					// Found closing quote at i (and next char is NOT @ or EOF)
					// Token is everything from start to i (inclusive of quotes)
					// We return it as raw bytes.
					return i + 1, data[:i+1], nil
				}
				i++
			}
			if atEOF {
				return 0, nil, fmt.Errorf("open quote: %w", io.ErrUnexpectedEOF)
			}
			return 0, nil, nil
		}

		if first == ':' {
			return 1, data[:1], nil
		}

		// ID/Num logic: consume until invalid char
		i := 0
		for i < len(data) {
			r, w := utf8.DecodeRune(data[i:])
			if r == utf8.RuneError && w == 1 {
				// Invalid rune? Maybe just consume it as part of ID if isIdChar allows?
				// isIdChar(utf8.RuneError) is likely false.
			}

			if !isIdChar(r) && r != '.' {
				if i == 0 {
					// Found invalid char at start (but we checked @ and :)
					// This should be treated as end of token if we scanned something?
					// But we are at 0.
					return 0, nil, fmt.Errorf("invalid character %q at start of word", r)
				}
				return i, data[:i], nil
			}
			i += w
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	if !s.Scan() {
		// If Scan returns false, Err() returns the error.
		// If Err() is nil, it means EOF.
		if s.Err() == nil {
			return "", io.EOF
		}
		return "", s.Err()
	}

	token := s.Text()
	if len(token) == 0 {
		return "", nil
	}

	if token[0] == '@' {
		// Unquote and unescape
		// Token is @content@
		// Content is token[1:len(token)-1]
		// Replace @@ with @
		if len(token) < 2 {
			return "", fmt.Errorf("invalid quoted string: %q", token)
		}
		content := token[1 : len(token)-1]
		return strings.ReplaceAll(content, "@@", "@"), nil
	}

	return token, nil
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
