package rcs

import (
	"bufio"
	"io"
)

type Scanner struct {
	*bufio.Scanner
	sf       bufio.SplitFunc
	lastScan bool
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
	return s.sf(data, eof)
}
func (s *Scanner) Scan() bool {
	s.lastScan = s.Scanner.Scan()
	return s.lastScan
}

func NewScanner(r io.Reader) *Scanner {
	scanner := &Scanner{
		Scanner: bufio.NewScanner(r),
	}
	scanner.Scanner.Split(scanner.scannerWrapper)
	return scanner
}
