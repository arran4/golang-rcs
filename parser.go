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

type Lock struct {
	User     string
	Revision string
	Strict   bool
}

type RevisionHead struct {
	Revision     string
	Date         time.Time
	Author       string
	State        string
	Branches     []string
	NextRevision string
}

type RevisionContent struct {
	Revision string
	Log      string
	Text     string
}

type File struct {
	Head             string
	Description      string
	Comment          string
	Access           bool
	Symbols          bool
	Locks            []*Lock
	RevisionHeads    []*RevisionHead
	RevisionContents []*RevisionContent
}

type Pos struct {
	line   int
	offset int
}

func (p *Pos) String() string {
	return fmt.Sprintf("%d:%d", p.line, p.offset)
}

func ParseFile(r io.Reader) (*File, error) {
	f := new(File)
	s := NewScanner(r)
	pos := &Pos{}
	if err := ParseHeader(s, pos, f); err != nil {
		return nil, err
	}
	if rhs, err := ParseRevisionHeaders(s, pos); err != nil {
		return nil, err
	} else {
		f.RevisionHeads = rhs
	}
	if desc, err := ParseDescription(s, pos); err != nil {
		return nil, err
	} else {
		f.Description = desc
	}
	if rcs, err := ParseRevisionContents(s, pos); err != nil {
		return nil, err
	} else {
		f.RevisionContents = rcs
	}
	return f, nil
}

func ParseDescription(s *Scanner, pos *Pos) (string, error) {
	d, err := ParseMultiLineText(s, pos, false, "desc")
	if err != nil {
		return "", err
	}
	if err := ScanStrings(s, pos, "\n\n", "\r\n\r\n"); err != nil {
		return "", err
	}
	return d, nil
}

func ParseRevisionContentLog(s *Scanner, pos *Pos) (string, error) {
	d, err := ParseMultiLineText(s, pos, false, "log")
	if err != nil {
		return "", err
	}
	if err := ScanStrings(s, pos, "\n", "\r\n"); err != nil {
		return "", err
	}
	return d, nil
}

func ParseRevisionContentText(s *Scanner, pos *Pos) (string, error) {
	d, err := ParseMultiLineText(s, pos, false, "text")
	if err != nil {
		return "", err
	}
	if err := ScanStrings(s, pos, "\n", "\r\n"); err != nil {
		return "", err
	}
	return d, nil
}

func ParseMultiLineText(s *Scanner, pos *Pos, havePropertyName bool, propertyName string) (string, error) {
	p := propertyName
	if !havePropertyName {
		p = ""
	}
	if err := ScanStrings(s, pos, p+"\n", p+"\r\n"); err != nil {
		if IsNotFound(err) {
			return "", err
		}
		return "", err
	}
	d, err := ParseAtQuotedString(s, pos)
	if err != nil {
		return "", err
	}
	if err := ScanNewLine(s, pos); err != nil {
		return "", err
	}
	return d, nil
}

func ParseHeader(s *Scanner, pos *Pos, f *File) error {
	if head, err := ParseHeaderHead(s, pos, false); err != nil {
		return fmt.Errorf("parsing %s: %w", pos, err)
	} else {
		f.Head = head
	}
	for {
		if err := ScanStrings(s, pos, "access", "symbols", "locks", "comment", "\n\n", "\r\n\r\n"); err != nil {
			return err
		}
		nt := s.Text()
		switch nt {
		case "access":
			f.Access = true
			err := ParseTerminatorFieldLine(s, pos)
			if err != nil {
				return err
			}
		case "symbols":
			f.Symbols = true
			if err := ParseTerminatorFieldLine(s, pos); err != nil {
				return err
			}
		case "locks":
			if locks, err := ParseHeaderLocks(s, pos, true); err != nil {
				return fmt.Errorf("parsing %s: %w", pos, err)
			} else {
				f.Locks = locks
			}
		case "comment":
			if comment, err := ParseHeaderComment(s, pos, true); err != nil {
				return fmt.Errorf("parsing %s: %w", pos, err)
			} else {
				f.Comment = comment
			}

		case "\n\n", "\r\n\r\n":
			return nil
		default:
			return fmt.Errorf("unknown token: %s", nt)
		}
	}
}

func ParseRevisionHeaders(s *Scanner, pos *Pos) ([]*RevisionHead, error) {
	var rhs []*RevisionHead
	for {
		if rh, next, err := ParseRevisionHeader(s, pos); err != nil {
			return nil, err
		} else {
			rhs = append(rhs, rh)
			if !next {
				return rhs, nil
			}
		}
	}
}

func ParseRevisionHeader(s *Scanner, pos *Pos) (*RevisionHead, bool, error) {
	rh := &RevisionHead{}
	if err := ScanUntilStrings(s, pos, "\r\n", "\n"); err != nil {
		return nil, false, err
	}
	rh.Revision = s.Text()
	if err := ScanNewLine(s, pos); err != nil {
		return nil, false, err
	}
	for {
		if err := ScanStrings(s, pos, "branches", "date", "next", "\n\n", "\r\n\r\n"); err != nil {
			return nil, false, err
		}
		nt := s.Text()
		switch nt {
		case "branches":
			if err := ParseRevisionHeaderBranches(s, pos, rh); err != nil {
				return nil, false, err
			}
		case "date":
			if err := ParseRevisionHeaderDateLine(s, pos, true, rh); err != nil {
				return nil, false, err
			}
		case "next":
			if n, err := ParseRevisionHeaderNext(s, pos, true); err != nil {
				return nil, false, err
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

func ParseRevisionContents(s *Scanner, pos *Pos) ([]*RevisionContent, error) {
	var rcs []*RevisionContent
	for {
		if rc, next, err := ParseRevisionContent(s, pos); err != nil {
			return nil, err
		} else {
			rcs = append(rcs, rc)
			if !next {
				return rcs, nil
			}
		}
	}
}

func ParseRevisionContent(s *Scanner, pos *Pos) (*RevisionContent, bool, error) {
	rh := &RevisionContent{}
	if err := ScanUntilStrings(s, pos, "\r\n", "\n"); err != nil {
		return nil, false, err
	}
	rh.Revision = s.Text()
	if err := ScanNewLine(s, pos); err != nil {
		return nil, false, err
	}
	for {
		if err := ScanStrings(s, pos, "log", "text", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			return nil, false, err
		}
		if !s.LastScan() {
			return rh, false, nil
		}
		nt := s.Text()
		switch nt {
		case "log":
			if s, err := ParseRevisionContentLog(s, pos); err != nil {
				return nil, false, err
			} else {
				rh.Log = s
			}
		case "text":
			if s, err := ParseRevisionContentText(s, pos); err != nil {
				return nil, false, err
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

func ParseRevisionHeaderBranches(s *Scanner, pos *Pos, rh *RevisionHead) error {
	rh.Branches = []string{}
	err := ParseTerminatorFieldLine(s, pos)
	return err
}

func ParseHeaderComment(s *Scanner, pos *Pos, havePropertyName bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, pos, "comment"); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, pos, 0); err != nil {
		return "", err
	}
	sr, err := ParseAtQuotedString(s, pos)
	if err != nil {
		return "", err
	}
	if err := ParseTerminatorFieldLine(s, pos); err != nil {
		return "", err
	}
	return sr, nil

}

func ParseAtQuotedString(s *Scanner, pos *Pos) (string, error) {
	sb := &strings.Builder{}
	if err := ScanStrings(s, pos, "@"); err != nil {
		return "", err
	}
	for {
		if err := ScanUntilStrings(s, pos, "@"); err != nil {
			return "", err
		}
		sb.WriteString(s.Text())
		if err := ScanStrings(s, pos, "@@", "@"); err != nil {
			return "", err
		}
		nt := s.Text()
		switch nt {
		case "@@":
			if _, err := sb.WriteString("@"); err != nil {
				return "", err
			}
		case "@":
			return sb.String(), nil
		default:
			if _, err := sb.WriteString("@"); err != nil {
				return "", err
			}
		}
	}
}

func ParseHeaderLocks(s *Scanner, pos *Pos, havePropertyName bool) ([]*Lock, error) {
	if !havePropertyName {
		if err := ScanStrings(s, pos, "locks"); err != nil {
			return nil, err
		}
	}
	var locks []*Lock
	for {
		if err := ScanStrings(s, pos, "\n\t", "\r\n\t", " "); err != nil {
			if IsNotFound(err) {
				break
			}
			return nil, err
		}
		nt := s.Text()
		switch nt {
		case "\n\t", "\r\n\t":
			if l, err := ParseLockLine(s, pos); err != nil {
				return nil, err
			} else {
				locks = append(locks, l)
			}
		case " ":
		default:
			return nil, fmt.Errorf("unknown token: %s", nt)
		}
	}
	if err := ScanNewLine(s, pos); err != nil {
		return nil, err
	}
	return locks, nil
}

func ParseLockLine(s *Scanner, pos *Pos) (*Lock, error) {
	l := &Lock{}
	if err := ScanUntilStrings(s, pos, ":"); err != nil {
		return nil, err
	}
	l.User = s.Text()
	if err := ScanStrings(s, pos, ":"); err != nil {
		return nil, err
	}
	if err := ScanUntilFieldTerminator(s, pos); err != nil {
		return nil, err
	}
	l.Revision = s.Text()
	if err := ScanFieldTerminator(s, pos); err != nil {
		return nil, err
	}
	for {
		if err := ScanStrings(s, pos, " ", "strict"); err != nil {
			if IsNotFound(err) {
				return l, nil
			}
			return nil, err
		}
		nt := s.Text()
		switch nt {
		case "strict":
			l.Strict = true
			if err := ScanFieldTerminator(s, pos); err != nil {
				return nil, err
			}
		case " ":
		default:
			return nil, fmt.Errorf("unknown token: %s", nt)
		}
	}
}

func ParseRevisionHeaderDateLine(s *Scanner, pos *Pos, haveHead bool, rh *RevisionHead) error {
	if dateStr, err := ParseProperty(s, pos, haveHead, "date", true); err != nil {
		return err
	} else if date, err := time.Parse("2006.01.02.15.04.05", dateStr); err != nil {
		return err
	} else {
		rh.Date = date
	}
	for {
		if err := ScanStrings(s, pos, "\t", "author", "state"); err != nil {
			if IsNotFound(err) {
				return nil
			}
			return err
		}
		nt := s.Text()
		switch nt {
		case "author":
			if s, err := ParseProperty(s, pos, true, "author", false); err != nil {
				return err
			} else {
				rh.Author = s
			}
		case "state":
			if s, err := ParseProperty(s, pos, true, "state", false); err != nil {
				return err
			} else {
				rh.State = s
			}
		case " ", "\t":
		default:
			return fmt.Errorf("unknown token: %s", nt)
		}
	}

}

func ParseRevisionHeaderNext(s *Scanner, pos *Pos, haveHead bool) (string, error) {
	return ParseProperty(s, pos, haveHead, "next", true)
}

func ParseHeaderHead(s *Scanner, pos *Pos, haveHead bool) (string, error) {
	return ParseProperty(s, pos, haveHead, "head", true)
}

func ParseProperty(s *Scanner, pos *Pos, havePropertyName bool, propertyName string, line bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, pos, propertyName); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, pos, 1); err != nil {
		return "", err
	}
	if err := ScanUntilFieldTerminator(s, pos); err != nil {
		return "", err
	}
	result := s.Text()
	if line {
		err := ParseTerminatorFieldLine(s, pos)
		if err != nil {
			return "", err
		}
	}
	return result, nil
}

func ParseTerminatorFieldLine(s *Scanner, pos *Pos) error {
	if err := ScanFieldTerminator(s, pos); err != nil {
		return err
	}
	if err := ScanNewLine(s, pos); err != nil {
		return err
	}
	return nil
}

func ScanWhiteSpace(s *Scanner, pos *Pos, minimum int) error {
	return ScanRunesUntil(s, pos, minimum, func(i []byte) bool {
		return unicode.IsSpace(bytes.Runes(i)[0])
	}, "whitespace")
}

func ScanUntilNewLine(s *Scanner, pos *Pos) error {
	return ScanUntilStrings(s, pos, "\r\n", "\n")
}

func ScanUntilFieldTerminator(s *Scanner, pos *Pos) error {
	return ScanUntilStrings(s, pos, ";")
}

func ScanRunesUntil(s *Scanner, pos *Pos, minimum int, until func([]byte) bool, name string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		adv := 0
		for {
			a, t, err := bufio.ScanRunes(data[adv:], atEOF)
			if err != nil {
				return 0, nil, err
			}
			if a == 0 && t == nil {
				return 0, nil, nil
			}
			if !until(t) {
				if minimum > 0 && minimum > adv {
					break
				}
				f := data[:adv]
				scanFound(f, adv, pos)
				return adv, f, nil
			}
			adv += a
		}
		err = ScanUntilNotFound(name)
		return 0, []byte{}, nil
	})
	if s.Err() != nil {
		return s.Err()
	}
	if !s.Scan() {
		return ScanUntilNotFound(name)
	}
	return
}

func ScanFieldTerminator(s *Scanner, pos *Pos) error {
	return ScanStrings(s, pos, ";")
}

func ScanNewLine(s *Scanner, pos *Pos) error {
	return ScanStrings(s, pos, "\r\n", "\n")
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
	return errors.Is(err, ScanNotFound(nil)) || errors.Is(err, ScanUntilNotFound(""))
}

func ScanStrings(s *Scanner, pos *Pos, strs ...string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		for _, ss := range strs {
			i := len(ss)
			if i > len(data) && !atEOF && bytes.HasPrefix([]byte(ss), data) {
				return 0, nil, nil
			}
			if bytes.HasPrefix(data, []byte(ss)) {
				rs := data[:i]
				scanFound(rs, i, pos)
				return i, rs, nil
			}
		}
		err = ScanNotFound(strs)
		return 0, []byte{}, nil
	})
	if s.Err() != nil {
		return s.Err()
	}
	if !s.Scan() {
		return ScanNotFound(strs)
	}
	return
}

func ScanUntilStrings(s *Scanner, pos *Pos, strs ...string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		for o := 0; o < len(data); o++ {
			for _, ss := range strs {
				i := len(ss)
				if i > len(data[o:]) && !atEOF && bytes.HasPrefix([]byte(ss), data[o:]) {
					return 0, nil, nil
				}
				if bytes.HasPrefix(data[o:], []byte(ss)) {
					rs := data[:o]
					scanFound(rs, o, pos)
					return o, rs, nil
				}
			}
		}
		err = ScanNotFound(strs)
		return 0, []byte{}, nil
	})
	if s.Err() != nil {
		return s.Err()
	}
	if !s.Scan() {
		return ScanNotFound(strs)
	}
	return err
}

func scanFound(found []byte, advance int, pos *Pos) {
	if nlp := bytes.LastIndexByte(found, '\n'); nlp > -1 {
		pos.offset = -nlp
		pos.line += bytes.Count(found, []byte("\n"))
	}
	pos.offset += advance
}
