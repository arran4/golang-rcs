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

type RevisionContents struct {
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
	RevisionContents []*RevisionContents
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
	return f, nil
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

func ParseHeaderComment(s *Scanner, pos *Pos, haveKey bool) (string, error) {
	if !haveKey {
		if err := ScanStrings(s, pos, "comment"); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, pos, 0); err != nil {
		return "", err
	}
	sr, s2, err := ParseAtQuotedString(s, pos)
	if err != nil {
		return s2, err
	}
	if err := ParseTerminatorFieldLine(s, pos); err != nil {
		return "", err
	}
	return sr, nil

}

func ParseAtQuotedString(s *Scanner, pos *Pos) (string, string, error) {
	sb := &strings.Builder{}
	if err := ScanStrings(s, pos, "@"); err != nil {
		return "", "", err
	}
	for {
		if err := ScanUntilStrings(s, pos, "@"); err != nil {
			return "", "", err
		}
		sb.WriteString(s.Text())
		if err := ScanStrings(s, pos, "@@", "@"); err != nil {
			return "", "", err
		}
		nt := s.Text()
		switch nt {
		case "@@":
			if _, err := sb.WriteString("@"); err != nil {
				return "", "", err
			}
		case "@":
			return sb.String(), "", nil
		default:
			if _, err := sb.WriteString("@"); err != nil {
				return "", "", err
			}
		}
	}
}

func ParseHeaderLocks(s *Scanner, pos *Pos, haveKey bool) ([]*Lock, error) {
	if !haveKey {
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

func ParseHeaderHead(s *Scanner, pos *Pos, haveHead bool) (string, error) {
	return ParsePropertyLine(s, pos, haveHead, "head")
}

func ParsePropertyLine(s *Scanner, pos *Pos, haveKey bool, propertyName string) (string, error) {
	if !haveKey {
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
	err := ParseTerminatorFieldLine(s, pos)
	if err != nil {
		return "", err
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
	if !s.Scan() {
		return ScanUntilNotFound(name)
	}
	if s.Err() != nil {
		return s.Err()
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
	if !s.Scan() {
		return ScanNotFound(strs)
	}
	if s.Err() != nil {
		return s.Err()
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
	if !s.Scan() {
		return ScanNotFound(strs)
	}
	if s.Err() != nil {
		return s.Err()
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
