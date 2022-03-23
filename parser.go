package rcs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
	"unicode"
)

type Lock struct {
	User     string
	Revision string
	Strict   bool
	Comment  string
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
	s := bufio.NewScanner(r)
	pos := &Pos{}
	if head, err := ParseHead(s, pos); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", pos, err)
	} else {
		f.Head = head
	}
	return f, nil
}

func ParseHead(s *bufio.Scanner, pos *Pos) (string, error) {
	if err := ScanStrings(s, pos, "head"); err != nil {
		return "", err
	}
	if err := ScanWhiteSpace(s, pos, 1); err != nil {
		return "", err
	}
	if err := ScanUntilNewLine(s, pos); err != nil {
		return "", err
	}
	head := s.Text()
	if err := ScanNewLine(s, pos); err != nil {
		return "", err
	}
	return head, nil
}

func ScanWhiteSpace(s *bufio.Scanner, pos *Pos, minimum int) error {
	return ScanRunesUntil(s, pos, minimum, func(i []byte) bool {
		return unicode.IsSpace(bytes.Runes(i)[0])
	})
}

func ScanUntilNewLine(s *bufio.Scanner, pos *Pos) error {
	return ScanUntilStrings(s, pos, "\r\n", "\n")
}

func ScanRunesUntil(s *bufio.Scanner, pos *Pos, minimum int, until func([]byte) bool) error {
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
		return 0, nil, errors.New("no match")
	})
	if !s.Scan() {
		return fmt.Errorf("finding 'head'")
	}
	if s.Err() != nil {
		return s.Err()
	}
	return nil
}

func ScanNewLine(s *bufio.Scanner, pos *Pos) error {
	return ScanStrings(s, pos, "\r\n", "\n")
}

func ScanStrings(s *bufio.Scanner, pos *Pos, strs ...string) error {
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
		return 0, nil, errors.New("no match")
	})
	if !s.Scan() {
		return fmt.Errorf("finding 'head'")
	}
	if s.Err() != nil {
		return s.Err()
	}
	return nil
}

func ScanUntilStrings(s *bufio.Scanner, pos *Pos, strs ...string) error {
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
		return 0, nil, errors.New("no match")
	})
	if !s.Scan() {
		return fmt.Errorf("finding 'head'")
	}
	if s.Err() != nil {
		return s.Err()
	}
	return nil
}

func scanFound(found []byte, advance int, pos *Pos) {
	if nlp := bytes.LastIndexByte(found, '\n'); nlp > -1 {
		pos.offset = -nlp
		pos.line += bytes.Count(found, []byte("\n"))
	}
	pos.offset += advance
}
