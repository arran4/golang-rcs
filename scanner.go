package rcs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
)

type Scanner struct {
	*bufio.Scanner
	sf                bufio.SplitFunc
	lastScan          bool
	pos               *Pos
	scanStringsFunc   bufio.SplitFunc
	scanStringsTarget []string
	scanStringsError  error
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
	scanner.scanStringsFunc = scanner.scanStringsSplit
	scanner.Scanner.Split(scanner.scannerWrapper)
	scanner.Buffer(nil, math.MaxInt/2)
	for _, opt := range opts {
		opt.ScannerOpt(scanner)
	}
	return scanner
}

func (s *Scanner) scanStringsSplit(data []byte, atEOF bool) (int, []byte, error) {
	s.scanStringsError = nil
	for _, ss := range s.scanStringsTarget {
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
		s.scanStringsError = ScanNotFound{
			LookingFor: s.scanStringsTarget,
			Pos:        *s.pos,
			Found:      string(data),
		}
		return 0, []byte{}, nil
	}
	return 0, nil, nil
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
