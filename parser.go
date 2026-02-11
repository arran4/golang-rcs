package rcs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode"
)

const DateFormat = "2006.01.02.15.04.05"

type Lock struct {
	User     string
	Revision string
}

func (l *Lock) String() string {
	return fmt.Sprintf("%s:%s;", l.User, l.Revision)
}

type RevisionHead struct {
	Revision     string
	Date         time.Time
	Author       string
	State        string
	Branches     []string
	NextRevision string
	CommitID     string
}

func (h *RevisionHead) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s\n", h.Revision))
	sb.WriteString(fmt.Sprintf("date\t%s;\tauthor %s;\tstate %s;\n", h.Date.Format(DateFormat), h.Author, h.State))
	sb.WriteString("branches")
	if len(h.Branches) > 0 {
		sb.WriteString("\n\t")
		sb.WriteString(strings.Join(h.Branches, "\n\t"))
		sb.WriteString(";")
	} else {
		sb.WriteString(";")
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("next\t%s;\n", h.NextRevision))
	if h.CommitID != "" {
		sb.WriteString(fmt.Sprintf("commitid\t%s;\n", h.CommitID))
	}
	return sb.String()
}

type RevisionContent struct {
	Revision string
	Log      string
	Text     string
}

func (c *RevisionContent) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s\n", c.Revision))
	sb.WriteString("log\n")
	sb.WriteString(AtQuote(c.Log))
	sb.WriteString("\n")
	sb.WriteString("text\n")
	sb.WriteString(AtQuote(c.Text))
	sb.WriteString("\n")
	return sb.String()
}

type File struct {
	Head             string
	Branch           string
	Description      string
	Comment          string
	Access           bool
	Symbols          bool
	AccessUsers      []string
	SymbolMap        map[string]string
	Locks            []*Lock
	Strict           bool
	Integrity        string
	Expand           string
	RevisionHeads    []*RevisionHead
	RevisionContents []*RevisionContent
}

func (f *File) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("head\t%s;\n", f.Head))
	if f.Branch != "" {
		sb.WriteString(fmt.Sprintf("branch\t%s;\n", f.Branch))
	}
	if f.Access {
		if len(f.AccessUsers) > 0 {
			sb.WriteString("access ")
			sb.WriteString(strings.Join(f.AccessUsers, " "))
			sb.WriteString(";\n")
		} else {
			sb.WriteString("access;\n")
		}
	}
	if f.Symbols {
		if len(f.SymbolMap) > 0 {
			sb.WriteString("symbols")
			keys := make([]string, 0, len(f.SymbolMap))
			for k := range f.SymbolMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				sb.WriteString("\n\t")
				sb.WriteString(fmt.Sprintf("%s:%s", k, f.SymbolMap[k]))
			}
			sb.WriteString(";\n")
		} else {
			sb.WriteString("symbols;\n")
		}
	}
	sb.WriteString("locks")
	if len(f.Locks) == 0 {
		sb.WriteString(";")
	}
	for _, lock := range f.Locks {
		sb.WriteString("\n\t")
		sb.WriteString(lock.String())
	}
	if f.Strict {
		sb.WriteString(" strict;")
	}
	sb.WriteString("\n")
	if f.Integrity != "" {
		sb.WriteString(fmt.Sprintf("integrity\t%s;\n", AtQuote(f.Integrity)))
	}
	sb.WriteString(fmt.Sprintf("comment\t%s;\n", AtQuote(f.Comment)))
	if f.Expand != "" {
		sb.WriteString(fmt.Sprintf("expand\t%s;\n", f.Expand)) // Assuming expand value doesn't need @ quoting in output if parsed raw
	}
	sb.WriteString("\n")
	sb.WriteString("\n")
	for _, head := range f.RevisionHeads {
		sb.WriteString(head.String())
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString("desc\n")
	sb.WriteString(fmt.Sprintf("%s\n", AtQuote(f.Description)))

	for _, content := range f.RevisionContents {
		sb.WriteString("\n")
		sb.WriteString("\n")
		sb.WriteString(content.String())
	}
	return sb.String()
}

func AtQuote(s string) string {
	return "@" + strings.ReplaceAll(s, "@", "@@") + "@"
}

func ParseFile(r io.Reader) (*File, error) {
	f := new(File)
	s := NewScanner(r)
	if err := ParseHeader(s, f); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	}
	descConsumed := false
	if rhs, dc, err := ParseRevisionHeaders(s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		descConsumed = dc
		f.RevisionHeads = rhs
	}
	if desc, err := ParseDescription(s, descConsumed); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		f.Description = desc
	}
	if rcs, err := ParseRevisionContents(s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		f.RevisionContents = rcs
	}
	return f, nil
}

func ParseDescription(s *Scanner, havePropertyName bool) (string, error) {
	var d string
	var err error
	if havePropertyName {
		d, err = ParseAtQuotedString(s)
		if err != nil {
			return "", fmt.Errorf("quote string: %w", err)
		}
		if err = ScanNewLine(s, true); err != nil {
			return "", fmt.Errorf("end new line: %w", err)
		}
	} else {
		d, err = ParseMultiLineText(s, false, "desc", true)
		if err != nil {
			return "", fmt.Errorf("description tag: %s", err)
		}
	}
	if err = ScanStrings(s, "\n\n", "\r\n\r\n"); err != nil {
		return "", fmt.Errorf("description scan string: %s", err)
	}
	return d, nil
}

func ParseRevisionContentLog(s *Scanner) (string, error) {
	d, err := ParseMultiLineText(s, true, "log", true)
	if err != nil {
		return "", err
	}
	return d, nil
}

func ParseRevisionContentText(s *Scanner) (string, error) {
	d, err := ParseMultiLineText(s, true, "text", true)
	if err != nil {
		return "", err
	}
	return d, nil
}

func ParseMultiLineText(s *Scanner, havePropertyName bool, propertyName string, scanEndNewline bool) (string, error) {
	p := propertyName
	if havePropertyName {
		p = ""
	}
	if err := ScanStrings(s, p+"\n", p+"\r\n"); err != nil {
		return "", err
	}
	d, err := ParseAtQuotedString(s)
	if err != nil {
		return "", fmt.Errorf("quote string: %w", err)
	}
	if scanEndNewline {
		if err = ScanNewLine(s, true); err != nil {
			return "", fmt.Errorf("end new line: %w", err)
		}
	}
	return d, nil
}

func ParseHeader(s *Scanner, f *File) error {
	if head, err := ParseHeaderHead(s, false); err != nil {
		return err
	} else {
		f.Head = head
	}
	var nextToken string
	for {
		var nt string
		if nextToken == "" {
			if err := ScanStrings(s, "branch", "access", "symbols", "locks", "strict", "integrity", "comment", "expand", "\n\n", "\r\n\r\n", " ", "\t", "\n", "\r\n"); err != nil {
				return err
			}
			nt = s.Text()
		} else {
			nt = nextToken
			nextToken = ""
		}

		switch nt {
		case " ", "\t", "\n", "\r\n":
			continue
		case "branch":
			if branch, err := ParseProperty(s, true, "branch", true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Branch = branch
			}
		case "access":
			f.Access = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				users, err := ParseHeaderAccess(s, true)
				if err != nil {
					return fmt.Errorf("token %#v: %w", nt, err)
				}
				f.AccessUsers = users
			}
		case "symbols":
			f.Symbols = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				sym, err := ParseHeaderSymbols(s, true)
				if err != nil {
					return fmt.Errorf("token %#v: %w", nt, err)
				}
				f.SymbolMap = sym
			}
		case "locks":
			var err error
			var locks []*Lock
			if locks, nextToken, err = ParseHeaderLocks(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Locks = locks
			}
		case "strict":
			f.Strict = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			}
		case "integrity":
			// TODO: Integrity parsing might need AtQuote string parsing similar to comment/desc?
			// RCS usually has simple strings? integrity @...@;
			// Let's assume quoted string for integrity.
			if integrity, err := ParseHeaderComment(s, true); err != nil { // Reusing ParseHeaderComment logic (quote + terminator)
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Integrity = integrity
			}
		case "comment":
			if comment, err := ParseHeaderComment(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Comment = comment
			}
		case "expand":
			if expand, err := ParseProperty(s, true, "expand", true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				// ParseProperty reads raw text. If it is @quoted@, we might need to handle it.
				// For now, assuming identifiers like @kv@ are returned as is.
				// The test expects "kv" if input is "@kv@".
				// If ParseProperty returns "@kv@", I need to strip it.
				// But ParseProperty calls ScanUntilFieldTerminator, which reads everything including @.
				if strings.HasPrefix(expand, "@") && strings.HasSuffix(expand, "@") {
					expand = expand[1 : len(expand)-1]
				}
				f.Expand = expand
			}

		case "\n\n", "\r\n\r\n":
			return nil
		default:
			return fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func ParseRevisionHeaders(s *Scanner) ([]*RevisionHead, bool, error) {
	var rhs []*RevisionHead
	for {
		if rh, next, descConsumed, err := ParseRevisionHeader(s); err != nil {
			return nil, false, err
		} else {
			if descConsumed {
				return rhs, true, nil
			}
			if rh == nil {
				return rhs, false, nil
			}
			rhs = append(rhs, rh)
			if !next {
				return rhs, false, nil
			}
		}
	}
}

func ParseRevisionHeader(s *Scanner) (*RevisionHead, bool, bool, error) {
	rh := &RevisionHead{}
	for {
		if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
			if IsNotFound(err) {
				return nil, false, false, nil
			}
			if IsEOFError(err) {
				return nil, false, false, nil
			}
			return nil, false, false, err
		}
		rev := s.Text()
		if err := ScanNewLine(s, false); err != nil {
			return nil, false, false, err
		}
		if rev != "" {
			rh.Revision = rev
			break
		}
	}
	if rh.Revision == "desc" {
		return nil, false, true, nil
	}
	for {
		if err := ScanStrings(s, "branches", "date", "next", "commitid", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			return nil, false, false, fmt.Errorf("finding revision header field: %w", err)
		}
		nt := s.Text()
		switch nt {
		case "branches":
			if err := ParseRevisionHeaderBranches(s, rh, true); err != nil {
				return nil, false, false, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "date":
			if err := ParseRevisionHeaderDateLine(s, true, rh); err != nil {
				return nil, false, false, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "next":
			if n, err := ParseRevisionHeaderNext(s, true); err != nil {
				return nil, false, false, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.NextRevision = n
			}
		case "commitid":
			if c, err := ParseProperty(s, true, "commitid", true); err != nil {
				return nil, false, false, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.CommitID = c
			}
		case "\n\n", "\r\n\r\n":
			return rh, true, false, nil
		case "\n", "\r\n":
			continue
		default:
			return nil, false, false, fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func ParseRevisionContents(s *Scanner) ([]*RevisionContent, error) {
	var rcs []*RevisionContent
	for {
		rc, next, err := ParseRevisionContent(s)
		if err != nil {
			return nil, err
		}
		if rc != nil {
			rcs = append(rcs, rc)
		}
		if !next {
			return rcs, nil
		}
	}
}

func ParseRevisionContent(s *Scanner) (*RevisionContent, bool, error) {
	rh := &RevisionContent{}
	if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
		if IsEOFError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	rh.Revision = s.Text()
	if rh.Revision == "" {
		return nil, false, ErrRevisionEmpty
	}
	if err := ScanNewLine(s, false); err != nil {
		return nil, false, err
	}
	for {
		if err := ScanStrings(s, "log", "text", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			if s.Bytes() != nil && len(s.Bytes()) == 0 {
				return rh, false, nil
			}
			return nil, false, err
		}
		nt := s.Text()
		switch nt {
		case "log":
			if s, err := ParseRevisionContentLog(s); err != nil {
				return nil, false, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Log = s
			}
		case "text":
			if s, err := ParseRevisionContentText(s); err != nil {
				return nil, false, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Text = s
			}
		case "\n\n", "\r\n\r\n":
			return rh, true, nil
		case "\n", "\r\n":
			return rh, false, nil
		default:
			return nil, false, fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func IsEOFError(err error) bool {
	var eofErr ErrEOF
	isEof := errors.As(err, &eofErr)
	return isEof
}

func ParseRevisionHeaderBranches(s *Scanner, rh *RevisionHead, havePropertyName bool) error {
	if !havePropertyName {
		if err := ScanStrings(s, "branches"); err != nil {
			return err
		}
	}
	rh.Branches = []string{}
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return err
		}
		if err := ScanStrings(s, ";"); err == nil {
			break
		}
		num, err := ScanTokenNum(s)
		if err != nil {
			return fmt.Errorf("expected num in branches: %w", err)
		}
		rh.Branches = append(rh.Branches, num)
	}
	return nil
}

func ParseHeaderComment(s *Scanner, havePropertyName bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "comment"); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 0); err != nil {
		return "", err
	}
	sr, err := ParseAtQuotedString(s)
	if err != nil {
		return "", err
	}
	if err := ParseTerminatorFieldLine(s); err != nil {
		return "", err
	}
	return sr, nil

}

func ParseHeaderAccess(s *Scanner, havePropertyName bool) ([]string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "access"); err != nil {
			return nil, err
		}
	}
	var ids []string
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, err
		}
		if err := ScanStrings(s, ";"); err == nil {
			break
		}
		id, err := ScanTokenId(s)
		if err != nil {
			return nil, fmt.Errorf("expected id in access: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func ParseHeaderSymbols(s *Scanner, havePropertyName bool) (map[string]string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "symbols"); err != nil {
			return nil, err
		}
	}
	m := map[string]string{}
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, err
		}
		if err := ScanStrings(s, ";"); err == nil {
			break
		}

		sym, err := ScanTokenSym(s)
		if err != nil {
			return nil, fmt.Errorf("expected sym in symbols: %w", err)
		}

		if err := ScanStrings(s, ":"); err != nil {
			return nil, fmt.Errorf("expected : after sym %q: %w", sym, err)
		}

		num, err := ScanTokenNum(s)
		if err != nil {
			return nil, fmt.Errorf("expected num for sym %q: %w", sym, err)
		}
		m[sym] = num
	}
	return m, nil
}

func ParseAtQuotedString(s *Scanner) (string, error) {
	sb := &strings.Builder{}
	if err := ScanStrings(s, "@"); err != nil {
		return "", fmt.Errorf("open quote: %v", err)
	}
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

func ParseHeaderLocks(s *Scanner, havePropertyName bool) ([]*Lock, string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "locks"); err != nil {
			return nil, "", err
		}
	}
	var locks []*Lock
	for {
		// Combine checks for separators AND lock ID
		id, match, err := ScanLockIdOrStrings(s,
			"\n\t", "\r\n\t", " ", ";",
			"branch", "access", "symbols", "locks", "strict", "integrity", "comment", "expand", "\n\n", "\r\n\r\n", "\n", "\r\n")

		if err != nil {
			if IsNotFound(err) {
				// No match found.
				return locks, "", nil
			}
			return nil, "", err
		}

		if match != "" {
			switch match {
			case ";":
				// End of locks block
				return locks, "", nil
			case "\n\t", "\r\n\t":
				if l, err := ParseLockLine(s); err != nil {
					return nil, "", err
				} else {
					locks = append(locks, l)
				}
			case " ":
				// continue loop
			default:
				// It is a keyword or newline
				return locks, match, nil
			}
			continue
		}

		if id != "" {
			if err := ScanStrings(s, ":"); err != nil {
				return nil, "", err
			}
			if l, err := ParseLockBody(s, id); err != nil {
				return nil, "", err
			} else {
				locks = append(locks, l)
				continue
			}
		}
	}
}

func ParseLockLine(s *Scanner) (*Lock, error) {
	id, err := ScanTokenId(s)
	if err != nil {
		return nil, fmt.Errorf("expected id in lock: %w", err)
	}

	if err := ScanStrings(s, ":"); err != nil {
		return nil, fmt.Errorf("expected : after lock id %q: %w", id, err)
	}

	return ParseLockBody(s, id)
}

func ParseLockBody(s *Scanner, user string) (*Lock, error) {
	l := &Lock{User: user}
	num, err := ScanTokenNum(s)
	if err != nil {
		return nil, fmt.Errorf("expected num in lock: %w", err)
	}
	l.Revision = num
	return l, nil
}

func ParseRevisionHeaderDateLine(s *Scanner, haveHead bool, rh *RevisionHead) error {
	if dateStr, err := ParseProperty(s, haveHead, "date", false); err != nil {
		return err
	} else if date, err := ParseDate(dateStr, time.Time{}, nil); err != nil {
		return err
	} else {
		rh.Date = date
	}
	for {
		if err := ScanStrings(s, " ", "\t", "author", "state"); err != nil {
			if IsNotFound(err) {
				break
			}
			return err
		}
		nt := s.Text()
		switch nt {
		case " ", "\t":
			continue
		case "author":
			if s, err := ParseProperty(s, true, "author", false); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Author = s
			}
		case "state":
			if s, err := ParseProperty(s, true, "state", false); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.State = s
			}
		default:
			return fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
	if err := ScanNewLine(s, false); err != nil {
		return err
	}
	return nil
}

func ParseRevisionHeaderNext(s *Scanner, haveHead bool) (string, error) {
	return ParsePropertyNum(s, haveHead, "next", true)
}

func ParseHeaderHead(s *Scanner, haveHead bool) (string, error) {
	return ParsePropertyNum(s, haveHead, "head", true)
}

func ParsePropertyNum(s *Scanner, havePropertyName bool, propertyName string, line bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, propertyName); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 1); err != nil {
		return "", err
	}
	// Use ScanRunesUntil with minimum 0 because we might have an empty property (no num).
	// But check if it looks like a number start (digit or .) or stops at something else.
	// This will scan UNTIL predicate is true.
	// Predicate: !isDigit && !isDot.
	err := ScanRunesUntil(s, 0, func(b []byte) bool {
		r := bytes.Runes(b)[0]
		return !isDigit(r) && r != '.'
	}, "num")
	if err != nil {
		return "", err
	}
	result := s.Text()

	if line {
		if err := ParseTerminatorFieldLine(s); err != nil {
			return "", err
		}
	} else {
		if err := ScanFieldTerminator(s); err != nil {
			return "", err
		}
	}
	return result, nil
}

func ParseProperty(s *Scanner, havePropertyName bool, propertyName string, line bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, propertyName); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 1); err != nil {
		return "", err
	}
	if err := ScanUntilFieldTerminator(s); err != nil {
		return "", err
	}
	result := s.Text()
	if line {
		if err := ParseTerminatorFieldLine(s); err != nil {
			return "", err
		}
	} else {
		if err := ScanFieldTerminator(s); err != nil {
			return "", err
		}
	}
	return result, nil
}

func ParseTerminatorFieldLine(s *Scanner) error {
	if err := ScanFieldTerminator(s); err != nil {
		return err
	}
	return nil
}

func ScanWhiteSpace(s *Scanner, minimum int) error {
	return ScanRunesUntil(s, minimum, func(i []byte) bool {
		return !unicode.IsSpace(bytes.Runes(i)[0])
	}, "whitespace")
}

func ScanUntilNewLine(s *Scanner) error {
	return ScanUntilStrings(s, "\r\n", "\n")
}

func ScanUntilFieldTerminator(s *Scanner) error {
	return ScanUntilStrings(s, ";")
}

func ScanRunesUntil(s *Scanner, minimum int, until func([]byte) bool, name string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		err = nil
		adv := 0
		for {
			if !atEOF && adv >= len(data) {
				return 0, nil, nil
			}
			a, t, err := bufio.ScanRunes(data[adv:], atEOF)
			if err != nil {
				return 0, nil, err
			}
			if a == 0 && t == nil {
				if atEOF {
					if minimum > 0 && minimum > adv {
						break
					}
					f := data[:adv]
					return adv, f, nil
				}
				return 0, nil, nil
			}
			if until(t) {
				if minimum > 0 && minimum > adv {
					break
				}
				f := data[:adv]
				return adv, f, nil
			}
			adv += a
		}
		err = ScanUntilNotFound{
			Until: name,
			Pos:   *s.pos,
			Found: string(data),
		}
		return 0, []byte{}, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if err != nil {
			return err
		}
		return ScanUntilNotFound{
			Until: name,
			Pos:   *s.pos,
			Found: "",
		}
	}
	return
}

func ScanFieldTerminator(s *Scanner) error {
	return ScanStrings(s, ";")
}

func ScanNewLine(s *Scanner, orEof bool) error {
	if orEof {
		return ScanStrings(s, "\r\n", "\n", "")
	}
	return ScanStrings(s, "\r\n", "\n")
}

type ScanNotFound struct {
	LookingFor []string
	Pos        Pos
	Found      string
}

func (se ScanNotFound) Error() string {
	strs := make([]string, len(se.LookingFor))
	for i, s := range se.LookingFor {
		strs[i] = fmt.Sprintf("%#v", s)
	}
	lookingFor := strings.Join(strs, ", ")
	found := se.Found
	if len(found) > 20 {
		runes := []rune(found)
		if len(runes) > 20 {
			found = string(runes[:20]) + "..."
		}
	}
	return fmt.Sprintf("looking for %s at %s but found %q", lookingFor, se.Pos.String(), found)
}

type ScanUntilNotFound struct {
	Until string
	Pos   Pos
	Found string
}

func (se ScanUntilNotFound) Error() string {
	found := se.Found
	if len(found) > 20 {
		runes := []rune(found)
		if len(runes) > 20 {
			found = string(runes[:20]) + "..."
		}
	}
	return fmt.Sprintf("scanning until %q at %s but found %q", se.Until, se.Pos.String(), found)
}

func IsNotFound(err error) bool {
	switch err.(type) {
	case ScanUntilNotFound, ScanNotFound:
		return true
	}
	e1 := ScanNotFound{}
	e2 := ScanUntilNotFound{}
	return errors.As(err, &e1) || errors.As(err, &e2)
}

func ScanStrings(s *Scanner, strs ...string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		err = nil
		for _, ss := range strs {
			if len(ss) == 0 && atEOF && len(data) == 0 {
				return 0, []byte{}, nil
			} else if len(ss) == 0 {
				continue
			}
			i := len(ss)
			if i > len(data) && !atEOF && bytes.HasPrefix([]byte(ss), data) {
				return 0, nil, nil
			}
			if bytes.HasPrefix(data, []byte(ss)) {
				rs := data[:i]
				return i, rs, nil
			}
		}
		if atEOF {
			err = ScanNotFound{
				LookingFor: strs,
				Pos:        *s.pos,
				Found:      string(data),
			}
			return 0, []byte{}, nil
		}
		return 0, nil, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if err != nil {
			return err
		}
		return ScanNotFound{
			LookingFor: strs,
			Pos:        *s.pos,
			Found:      "",
		}
	}
	return
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

type ErrEOF struct{ error }

func (e ErrEOF) Error() string {
	return "EOF:" + e.error.Error()
}

func ScanUntilStrings(s *Scanner, strs ...string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		err = nil
		for o := 0; o < len(data); o++ {
			for _, ss := range strs {
				if len(ss) == 0 && atEOF {
					rs := data[:o]
					return o, rs, nil
				} else if len(ss) == 0 {
					continue
				}
				i := len(ss)
				if i >= len(data[o:]) && !atEOF && bytes.HasPrefix([]byte(ss), data[o:]) {
					return 0, nil, nil
				}
				if bytes.HasPrefix(data[o:], []byte(ss)) {
					rs := data[:o]
					return o, rs, nil
				}
			}
		}
		if atEOF {
			err = ErrEOF{ScanNotFound{
				LookingFor: strs,
				Pos:        *s.pos,
				Found:      string(data),
			}}
			return 0, []byte{}, nil
		}
		return 0, nil, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if err != nil {
			return err
		}
		return ScanNotFound{
			LookingFor: strs,
			Pos:        *s.pos,
			Found:      "",
		}
	}
	return err
}
