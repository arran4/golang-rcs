package rcs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

type Scanner struct {
	*bufio.Scanner
	sf             bufio.SplitFunc
	lastScan       bool
	pos            *Pos
	matchSplitFunc bufio.SplitFunc
	matchTarget    []string
	matchError     error
}

type scannerInterface interface {
	Err() error
	Bytes() []byte
	Text() string
	Scan() bool
	Buffer(buf []byte, max int)
	Split(split bufio.SplitFunc)
}

var _ scannerInterface = (*Scanner)(nil)

func (s *Scanner) Split(split bufio.SplitFunc) {
	s.sf = split
}

// There is probably a very good reason you shouldn't do this.
func (s *Scanner) scannerWrapper(data []byte, eof bool) (advance int, token []byte, err error) {
	a, t, err := s.sf(data, eof)
	scanFound(t, a, s.pos)
	return a, t, err
}

func (s *Scanner) Scan() bool {
	s.lastScan = s.Scanner.Scan()
	return s.lastScan
}

func (s *Scanner) Text() string {
	return s.Scanner.Text()
}

func (s *Scanner) Bytes() []byte {
	return s.Scanner.Bytes()
}

func (s *Scanner) LastScan() bool {
	return s.lastScan
}

type ScannerOpt interface {
	ScannerOpt(scanner *Scanner)
}

type MaxBuffer int

func (mb MaxBuffer) ScannerOpt(scanner *Scanner) {
	scanner.Buffer(nil, int(mb))
}

func NewScanner(r io.Reader, opts ...ScannerOpt) *Scanner {
	scanner := &Scanner{
		Scanner: bufio.NewScanner(r),
		pos: &Pos{
			Line: 1,
		},
		sf: bufio.ScanLines,
	}
	scanner.matchSplitFunc = func(data []byte, atEOF bool) (int, []byte, error) {
		scanner.matchError = nil
		for _, ss := range scanner.matchTarget {
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
			scanner.matchError = ScanNotFound{
				LookingFor: scanner.matchTarget,
				Pos:        *scanner.pos,
				Found:      string(data),
			}
			return 0, []byte{}, nil
		}
		return 0, nil, nil
	}
	scanner.Scanner.Split(scanner.scannerWrapper)
	scanner.Buffer(nil, math.MaxInt/2)
	for _, opt := range opts {
		opt.ScannerOpt(scanner)
	}
	return scanner
}

func (s *Scanner) ScanMatch(strs ...string) error {
	s.matchError = nil
	s.matchTarget = strs
	s.Split(s.matchSplitFunc)
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if s.matchError != nil {
			return s.matchError
		}
		return ScanNotFound{
			LookingFor: strs,
			Pos:        *s.pos,
			Found:      "",
		}
	}
	if s.matchError != nil {
		return s.matchError
	}
	return nil
}

func scanFound(found []byte, advance int, pos *Pos) {
	if nlp := bytes.LastIndexByte(found, '\n'); nlp > -1 {
		pos.Offset = len(found) - nlp - 1
		pos.Line += bytes.Count(found, []byte("\n"))
	} else {
		pos.Offset += advance
	}
}

type Pos struct {
	Line   int
	Offset int
}

func (p *Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Offset)
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
	return fmt.Sprintf("scanning for %q at %s but found %q", se.Until, se.Pos.String(), found)
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
