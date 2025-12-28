package rcs

import (
	"bufio"
	"strings"
	"testing"
)

func TestScanner_LastScan(t *testing.T) {
	s := NewScanner(strings.NewReader("test"))
	if s.LastScan() {
		t.Errorf("LastScan should be false initially")
	}
	s.Scan()
	if !s.LastScan() {
		t.Errorf("LastScan should be true after successful scan")
	}
	s.Scan() // EOF
	if s.LastScan() {
		t.Errorf("LastScan should be false after EOF")
	}
}

func TestMaxBuffer_ScannerOpt(t *testing.T) {
	// buffer size 1, input "ab". Scan should fail if token is larger than buffer.
	s := NewScanner(strings.NewReader("ab"), MaxBuffer(1))
	s.Split(bufio.ScanBytes) // Split by bytes so it should be fine if buffer is small but we want to test buffer size setting.

	// Actually bufio.Scanner with MaxBuffer will error if token > maxbuffer.
	// Default split is ScanLines. "ab" is one line. Length 2. MaxBuffer 1. Should fail.

	s = NewScanner(strings.NewReader("ab"), MaxBuffer(1))
	if s.Scan() {
		t.Errorf("Should not have scanned 'ab' with buffer size 1")
	}
	if s.Err() != bufio.ErrTooLong {
		t.Errorf("Expected ErrTooLong, got %v", s.Err())
	}
}
