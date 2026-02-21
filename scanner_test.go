package rcs

import (
	"bufio"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_scanFound(t *testing.T) {
	type args struct {
		found   []byte
		advance int
		pos     *Pos
	}
	tests := []struct {
		name     string
		args     args
		expected *Pos
	}{
		{
			name: "Advance characters no new Line",
			args: args{
				found:   []byte("testing"),
				advance: len("testing"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   0,
				Offset: len("testing"),
			},
		},
		{
			name: "New Line at end zeros advance and increments new line",
			args: args{
				found:   []byte("testing\n"),
				advance: len("testing\n"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   1,
				Offset: 0,
			},
		},
		{
			name: "Trailing content after new line increments advance correctly",
			args: args{
				found:   []byte("testing\ntesting 123"),
				advance: len("testing\n"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   1,
				Offset: len("testing 123"),
			},
		},
		{
			name: "Multiple new lines don't break anything",
			args: args{
				found:   []byte("testing\ntesting 123\nand some more content\noh and more!"),
				advance: len("testing\ntesting 123\nand some more content\noh and more!"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   3,
				Offset: len("oh and more!"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanFound(tt.args.found, tt.args.advance, tt.args.pos)
			if diff := cmp.Diff(tt.expected, tt.args.pos); diff != "" {
				t.Errorf("Failed: %s", diff)
			}
		})
	}
}

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

func TestPos_String(t *testing.T) {
	type fields struct {
		line   int
		offset int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				line:   3,
				offset: 34,
			},
			want: "3:34",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pos{
				Line:   tt.fields.line,
				Offset: tt.fields.offset,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewScanner(t *testing.T) {
	tests := []struct {
		name string
		want *Scanner
	}{
		{
			name: "Set the pos correctly",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScanner(nil)
			if got.pos.Line != 1 {
				t.Errorf("Wrong line number")
			}
		})
	}
}
