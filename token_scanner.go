package rcs

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"unicode"
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

func ScanTokenPhrase(s *Scanner) (PhraseValue, error) {
	if err := ScanStrings(s, "@"); err == nil {
		body, err := ParseAtQuotedStringBody(s)
		if err != nil {
			return nil, err
		}
		return QuotedString(body), nil
	}
	if err := ScanStrings(s, ":"); err == nil {
		return SimpleString(":"), nil
	}
	id, err := ScanTokenId(s)
	if err != nil {
		return nil, err
	}
	return SimpleString(id), nil
}

func ScanTokenWord(s *Scanner) (string, error) {
	pv, err := ScanTokenPhrase(s)
	if err != nil {
		return "", err
	}
	return pv.Raw(), nil
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

func ScanTokenAuthor(s *Scanner) (string, error) {
	if err := ScanStrings(s, "@"); err == nil {
		return ParseAtQuotedStringBody(s)
	}
	err := ScanRunesUntil(s, 1, func(b []byte) bool {
		return b[0] == ';'
	}, "author")
	if err != nil {
		return "", err
	}
	return s.Text(), nil
}

func ParseAtQuotedString(s *Scanner) (string, error) {
	if err := ScanStrings(s, "@"); err != nil {
		return "", fmt.Errorf("open quote: %v", err)
	}
	return ParseAtQuotedStringBody(s)
}

func ParseAtQuotedStringBody(s *Scanner) (string, error) {
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
}

func ScanLockIdOrStrings(s *Scanner, strs ...string) (id string, match string, err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		// Check strings first
		for _, ss := range strs {
			if len(ss) == 0 {
				continue
			}
			if len(data) < len(ss) {
				if atEOF {
					if bytes.Equal(data, []byte(ss)) {
						return len(data), data, nil
					}
					// mismatch or partial mismatch at EOF
				} else {
					if bytes.HasPrefix([]byte(ss), data) {
						// potential partial match
						return 0, nil, nil
					}
				}
			} else {
				if bytes.HasPrefix(data, []byte(ss)) {
					return len(ss), data[:len(ss)], nil
				}
			}
		}

		// Check Lock ID
		for i, b := range data {
			if b == ':' {
				if i == 0 {
					return 0, nil, ErrEmptyId
				}
				return i, data[:i], nil
			}
			switch b {
			case ' ', '\t', '\n', '\r', ';':
				return 0, nil, bufio.ErrFinalToken
			}
		}

		if atEOF {
			return 0, nil, nil
		}
		return 0, nil, nil
	})

	if !s.Scan() {
		if s.Err() != nil {
			return "", "", s.Err()
		}
		return "", "", ScanNotFound{
			LookingFor: append(strs, "lock_id"),
			Pos:        *s.pos,
			Found:      "",
		}
	}
	text := s.Text()
	for _, ss := range strs {
		if text == ss {
			return "", text, nil
		}
	}
	return text, "", nil
}
