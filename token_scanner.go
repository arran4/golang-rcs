package rcs

import (
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

type PhraseValue interface {
	fmt.Stringer
	Raw() string
}

type SimpleString string

func (s SimpleString) String() string {
	return string(s)
}

func (s SimpleString) Raw() string {
	return string(s)
}

type QuotedString string

func (s QuotedString) String() string {
	return "@" + strings.ReplaceAll(string(s), "@", "@@") + "@"
}

func (s QuotedString) Raw() string {
	return string(s)
}

type PhraseValues []PhraseValue

func (p PhraseValues) Format() {
	for i, v := range p {
		raw := v.Raw()
		valid := true
		if len(raw) == 0 {
			valid = false
		} else {
			for _, r := range raw {
				if !isIdChar(r) && r != '.' {
					valid = false
					break
				}
			}
		}

		if valid {
			// If it is valid ID, prefer SimpleString
			if _, ok := v.(SimpleString); !ok {
				p[i] = SimpleString(raw)
			}
		} else {
			// If it is invalid ID, must be QuotedString
			if _, ok := v.(QuotedString); !ok {
				p[i] = QuotedString(raw)
			}
		}
	}
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
