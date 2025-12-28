package rcs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"
)

const DateFormat = "2006.01.02.15.04.05"

type Lock struct {
	User     string
	Revision string
	Strict   bool
}

func (l *Lock) String() string {
	if l.Strict {
		return fmt.Sprintf("%s:%s; strict;", l.User, l.Revision)
	}
	return fmt.Sprintf("%s:%s;", l.User, l.Revision)
}

type RevisionHead struct {
	Revision     string
	Date         time.Time
	Author       string
	State        string
	Branches     []string
	NextRevision string
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
		sb.WriteString("access;\n")
	}
	if f.Symbols {
		sb.WriteString("symbols;\n")
	}
	sb.WriteString("locks")
	if len(f.Locks) == 0 {
		sb.WriteString(";")
	}
	for _, lock := range f.Locks {
		sb.WriteString("\n\t")
		sb.WriteString(lock.String())
	}
	sb.WriteString("\n")
	if f.Strict {
		sb.WriteString("strict;\n")
	}
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
	if rhs, err := ParseRevisionHeaders(s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		f.RevisionHeads = rhs
	}
	if desc, err := ParseDescription(s); err != nil {
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

func ParseDescription(s *Scanner) (string, error) {
	d, err := ParseMultiLineText(s, false, "desc", true)
	if err != nil {
		return "", fmt.Errorf("description tag: %s", err)
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
	for {
		if err := ScanStrings(s, "branch", "access", "symbols", "locks", "strict", "integrity", "comment", "expand", "\n\n", "\r\n\r\n"); err != nil {
			return err
		}
		nt := s.Text()
		switch nt {
		case "branch":
			if branch, err := ParseProperty(s, true, "branch", true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Branch = branch
			}
		case "access":
			f.Access = true
			err := ParseTerminatorFieldLine(s)
			if err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			}
		case "symbols":
			f.Symbols = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			}
		case "locks":
			if locks, err := ParseHeaderLocks(s, true); err != nil {
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
			return fmt.Errorf("unknown token: %s", nt)
		}
	}
}

func ParseRevisionHeaders(s *Scanner) ([]*RevisionHead, error) {
	var rhs []*RevisionHead
	for {
		if rh, next, err := ParseRevisionHeader(s); err != nil {
			return nil, err
		} else {
			rhs = append(rhs, rh)
			if !next {
				return rhs, nil
			}
		}
	}
}

func ParseRevisionHeader(s *Scanner) (*RevisionHead, bool, error) {
	rh := &RevisionHead{}
	if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
		return nil, false, err
	}
	rh.Revision = s.Text()
	if rh.Revision == "" {
		return nil, false, fmt.Errorf("revsion empty")
	}
	if err := ScanNewLine(s, false); err != nil {
		return nil, false, err
	}
	for {
		if err := ScanStrings(s, "branches", "date", "next", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			return nil, false, fmt.Errorf("finding revision header field: %w", err)
		}
		nt := s.Text()
		switch nt {
		case "branches":
			if err := ParseRevisionHeaderBranches(s, rh, true); err != nil {
				return nil, false, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "date":
			if err := ParseRevisionHeaderDateLine(s, true, rh); err != nil {
				return nil, false, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "next":
			if n, err := ParseRevisionHeaderNext(s, true); err != nil {
				return nil, false, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.NextRevision = n
			}
		case "\n\n", "\r\n\r\n":
			return rh, false, nil
		case "\n", "\r\n":
			return rh, true, nil
		default:
			return nil, false, fmt.Errorf("unknown token: %s", nt)
		}
	}
}

func ParseRevisionContents(s *Scanner) ([]*RevisionContent, error) {
	var rcs []*RevisionContent
	for {
		if rc, next, err := ParseRevisionContent(s); err != nil {
			return nil, err
		} else {
			rcs = append(rcs, rc)
			if !next {
				return rcs, nil
			}
		}
	}
}

func ParseRevisionContent(s *Scanner) (*RevisionContent, bool, error) {
	rh := &RevisionContent{}
	if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
		return nil, false, err
	}
	rh.Revision = s.Text()
	if rh.Revision == "" {
		return nil, false, fmt.Errorf("revsion empty")
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
			return nil, false, fmt.Errorf("unknown token: %s", nt)
		}
	}
}

func ParseRevisionHeaderBranches(s *Scanner, rh *RevisionHead, havePropertyName bool) error {
	if !havePropertyName {
		if err := ScanStrings(s, "branches"); err != nil {
			return err
		}
	}
	if err := ScanUntilFieldTerminator(s); err != nil {
		return err
	}
	rh.Branches = strings.Fields(s.Text())
	if err := ParseTerminatorFieldLine(s); err != nil {
		return err
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
			if _, err := sb.WriteString("@"); err != nil {
				return "", fmt.Errorf("token %#v: %w", nt, err)
			}
		case "@":
			return sb.String(), nil
		default:
			if _, err := sb.WriteString("@"); err != nil {
				return "", fmt.Errorf("token %#v: %w", nt, err)
			}
		}
	}
}

func ParseHeaderLocks(s *Scanner, havePropertyName bool) ([]*Lock, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "locks"); err != nil {
			return nil, err
		}
	}
	var locks []*Lock
	for {
		if err := ScanStrings(s, "\n\t", "\r\n\t", " "); err != nil {
			if IsNotFound(err) {
				if err := ScanFieldTerminator(s); err == nil {
					break
				}
				break
			}
			return nil, err
		}
		nt := s.Text()
		switch nt {
		case "\n\t", "\r\n\t":
			if l, err := ParseLockLine(s); err != nil {
				return nil, err
			} else {
				locks = append(locks, l)
			}
		case " ":
		default:
			return nil, fmt.Errorf("unknown token: %s", nt)
		}
	}
	if err := ScanNewLine(s, false); err != nil {
		return nil, err
	}
	return locks, nil
}

func ParseLockLine(s *Scanner) (*Lock, error) {
	l := &Lock{}
	if err := ScanUntilStrings(s, ":"); err != nil {
		return nil, err
	}
	l.User = s.Text()
	if err := ScanStrings(s, ":"); err != nil {
		return nil, err
	}
	if err := ScanUntilFieldTerminator(s); err != nil {
		return nil, err
	}
	l.Revision = s.Text()
	if l.Revision == "" {
		return nil, fmt.Errorf("revsion empty")
	}
	if err := ScanFieldTerminator(s); err != nil {
		return nil, err
	}
	for {
		if err := ScanStrings(s, " ", "strict"); err != nil {
			if IsNotFound(err) {
				return l, nil
			}
			return nil, err
		}
		nt := s.Text()
		switch nt {
		case "strict":
			l.Strict = true
			if err := ScanFieldTerminator(s); err != nil {
				return nil, err
			}
		case " ":
		default:
			return nil, fmt.Errorf("unknown token: %s", nt)
		}
	}
}

func ParseRevisionHeaderDateLine(s *Scanner, haveHead bool, rh *RevisionHead) error {
	if dateStr, err := ParseProperty(s, haveHead, "date", false); err != nil {
		return err
	} else if date, err := time.Parse(DateFormat, dateStr); err != nil {
		return err
	} else {
		rh.Date = date
	}
	for {
		if err := ScanStrings(s, "\t", "author", "state"); err != nil {
			if IsNotFound(err) {
				break
			}
			return err
		}
		nt := s.Text()
		switch nt {
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
		case " ", "\t":
		default:
			return fmt.Errorf("unknown token: %s", nt)
		}
	}
	if err := ScanNewLine(s, false); err != nil {
		return err
	}
	return nil
}

func ParseRevisionHeaderNext(s *Scanner, haveHead bool) (string, error) {
	return ParseProperty(s, haveHead, "next", true)
}

func ParseHeaderHead(s *Scanner, haveHead bool) (string, error) {
	return ParseProperty(s, haveHead, "head", true)
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
	if err := ScanNewLine(s, false); err != nil {
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
		err = ScanUntilNotFound(name)
		return 0, []byte{}, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		return ScanUntilNotFound(name)
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

type ScanNotFound []string

func (se ScanNotFound) Error() string {
	return fmt.Sprintf("finding %#v", []string(se))
}

type ScanUntilNotFound string

func (se ScanUntilNotFound) Error() string {
	return fmt.Sprintf("scanning until %#v", string(se))
}

func IsNotFound(err error) bool {
	switch err.(type) {
	case ScanUntilNotFound, ScanNotFound:
		return true
	}
	e1 := ScanNotFound([]string{})
	e2 := ScanUntilNotFound("")
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
			if i >= len(data) && !atEOF && bytes.HasPrefix([]byte(ss), data) {
				return 0, nil, nil
			}
			if bytes.HasPrefix(data, []byte(ss)) {
				rs := data[:i]
				return i, rs, nil
			}
		}
		err = ScanNotFound(strs)
		return 0, []byte{}, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		return ScanNotFound(strs)
	}
	return
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
			err = ScanNotFound(strs)
			return 0, []byte{}, nil
		}
		return 0, nil, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		return ScanNotFound(strs)
	}
	return err
}
