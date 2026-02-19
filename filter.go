package rcs

import (
	"fmt"
	"strings"
	"unicode"
)

// Filter is an interface for filtering RCS revisions.
type Filter interface {
	Match(r *RevisionHead) bool
}

// StateFilter filters revisions by state.
type StateFilter struct {
	State string
}

func (f *StateFilter) Match(r *RevisionHead) bool {
	return string(r.State) == f.State
}

// OrFilter combines multiple filters with OR logic.
type OrFilter struct {
	Filters []Filter
}

func (f *OrFilter) Match(r *RevisionHead) bool {
	for _, sub := range f.Filters {
		if sub.Match(r) {
			return true
		}
	}
	return false
}

// AndFilter combines multiple filters with AND logic.
type AndFilter struct {
	Filters []Filter
}

func (f *AndFilter) Match(r *RevisionHead) bool {
	for _, sub := range f.Filters {
		if !sub.Match(r) {
			return false
		}
	}
	return true
}

// InFilter filters revisions where a field matches one of the values.
type InFilter struct {
	Field  string
	Values []string
}

func (f *InFilter) Match(r *RevisionHead) bool {
	val := ""
	switch f.Field {
	case "state", "s":
		val = string(r.State)
	}
	for _, v := range f.Values {
		if val == v {
			return true
		}
	}
	return false
}

// ParseFilter parses a filter string into a Filter object.
// Supported syntax:
// - state=<value> or s=<value>
// - state in (<value1> <value2> ...)
// - <expression> OR <expression>
// - <expression> || <expression>
func ParseFilter(input string) (Filter, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	return parseExpression(tokens)
}

type tokenType int

const (
	tokenError tokenType = iota
	tokenIdentifier
	tokenEquals
	tokenLParen
	tokenRParen
	tokenOr
	tokenIn
)

type token struct {
	typ tokenType
	val string
}

func tokenize(input string) ([]token, error) {
	var tokens []token
	runes := []rune(input)
	length := len(runes)
	for i := 0; i < length; {
		ch := runes[i]
		if unicode.IsSpace(ch) {
			i++
			continue
		}
		switch ch {
		case '=':
			tokens = append(tokens, token{tokenEquals, "="})
			i++
		case '(':
			tokens = append(tokens, token{tokenLParen, "("})
			i++
		case ')':
			tokens = append(tokens, token{tokenRParen, ")"})
			i++
		case '|':
			if i+1 < length && runes[i+1] == '|' {
				tokens = append(tokens, token{tokenOr, "||"})
				i += 2
			} else {
				return nil, fmt.Errorf("unexpected character '|' at position %d", i)
			}
		default:
			start := i
			for i < length && !unicode.IsSpace(runes[i]) && runes[i] != '=' && runes[i] != '(' && runes[i] != ')' {
				// Special check for '||' inside identifier? No, identifiers shouldn't contain | unless escaped, but we don't support escaping yet.
				// However, if we hit '|', we should stop if it looks like an operator.
				if runes[i] == '|' && i+1 < length && runes[i+1] == '|' {
					break
				}
				i++
			}
			word := string(runes[start:i])
			if strings.EqualFold(word, "OR") {
				tokens = append(tokens, token{tokenOr, "OR"})
			} else if strings.EqualFold(word, "in") {
				tokens = append(tokens, token{tokenIn, "in"})
			} else {
				tokens = append(tokens, token{tokenIdentifier, word})
			}
		}
	}
	return tokens, nil
}

func parseExpression(tokens []token) (Filter, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty filter expression")
	}

	// Split by OR
	var parts [][]token
	var currentPart []token
	for _, t := range tokens {
		if t.typ == tokenOr {
			if len(currentPart) > 0 {
				parts = append(parts, currentPart)
				currentPart = []token{}
			}
		} else {
			currentPart = append(currentPart, t)
		}
	}
	if len(currentPart) > 0 {
		parts = append(parts, currentPart)
	}

	if len(parts) == 1 {
		return parseSimpleExpression(parts[0])
	}

	var filters []Filter
	for _, part := range parts {
		f, err := parseSimpleExpression(part)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}
	return &OrFilter{Filters: filters}, nil
}

func parseSimpleExpression(tokens []token) (Filter, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty expression")
	}

	// Check for "field = value"
	if len(tokens) == 3 && tokens[1].typ == tokenEquals {
		field := tokens[0].val
		value := tokens[2].val
		switch field {
		case "state", "s":
			return &StateFilter{State: value}, nil
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
	}

	// Check for "field in (values...)"
	if len(tokens) >= 4 && tokens[1].typ == tokenIn && tokens[2].typ == tokenLParen && tokens[len(tokens)-1].typ == tokenRParen {
		field := tokens[0].val
		values := []string{}
		for i := 3; i < len(tokens)-1; i++ {
			if tokens[i].typ != tokenIdentifier {
				return nil, fmt.Errorf("expected identifier in list, got %v", tokens[i])
			}
			values = append(values, tokens[i].val)
		}
		return &InFilter{Field: field, Values: values}, nil
	}

	return nil, fmt.Errorf("invalid expression: %v", tokens)
}
