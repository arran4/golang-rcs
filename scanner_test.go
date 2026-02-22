package rcs

import (
	"bytes"
	"errors"
	"fmt"
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

func TestScanNewLine(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Scans a unix new Line",
			args: args{
				s: NewScanner(strings.NewReader("\n")),
			},
			wantErr: false,
		},
		{
			name: "Scans a windows new Line",
			args: args{
				s: NewScanner(strings.NewReader("\r\n")),
			},
			wantErr: false,
		},
		{
			name: "Fails to scan nothing",
			args: args{
				s: NewScanner(strings.NewReader("")),
			},
			wantErr: true,
		},
		{
			name: "Fails to scan non new Line data",
			args: args{
				s: NewScanner(strings.NewReader("asdfasdfasdf")),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanNewLine(tt.args.s, false); (err != nil) != tt.wantErr {
				t.Errorf("ScanNewLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanStrings(t *testing.T) {
	type args struct {
		s    *Scanner
		strs []string
	}
	tests := []struct {
		name     string
		expected string
		args     args
		wantErr  bool
	}{
		{
			name:     "Scans a word before a space",
			expected: "This",
			args: args{
				s:    NewScanner(strings.NewReader("This is a word")),
				strs: []string{"This"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanStrings(tt.args.s, tt.args.strs...); (err != nil) != tt.wantErr {
				t.Errorf("ScanStrings() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := tt.args.s.Text(); got != tt.expected {
				t.Errorf("ScanRunesUntil() s.Text() = %v, want s.Text() %v", got, tt.expected)
			}
		})
	}
}

func TestScanUntilNewLine(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected string
	}{
		{
			name:     "Scans a word before a space",
			expected: "This is",
			args: args{
				s: NewScanner(strings.NewReader("This is\n a word")),
			},
			wantErr: false,
		},
		{
			name:     "No new Line no result",
			expected: "",
			args: args{
				s: NewScanner(strings.NewReader("This is a word")),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilNewLine(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanUntilNewLine() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := tt.args.s.Text(); got != tt.expected {
				t.Errorf("ScanRunesUntil() s.Text() = %v, want s.Text() %v", got, tt.expected)
			}
		})
	}
}

func TestScanUntilStrings(t *testing.T) {
	type args struct {
		s    *Scanner
		strs []string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected string
	}{
		{
			name:     "Scans until a word",
			expected: "This is a ",
			args: args{
				s:    NewScanner(strings.NewReader("This is a word")),
				strs: []string{"word"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilStrings(tt.args.s, tt.args.strs...); (err != nil) != tt.wantErr {
				t.Errorf("ScanUntilStrings() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := tt.args.s.Text(); got != tt.expected {
				t.Errorf("ScanRunesUntil() s.Text() = %v, want s.Text() %v", got, tt.expected)
			}
		})
	}
}

func TestScanWhiteSpace(t *testing.T) {
	type args struct {
		s       *Scanner
		minimum int
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected string
	}{
		{
			name:     "Scans until a word",
			expected: " ",
			args: args{
				s:       NewScanner(strings.NewReader(" word")),
				minimum: 1,
			},
			wantErr: false,
		},
		{
			name:     "Minimum fails it",
			expected: "",
			args: args{
				s:       NewScanner(strings.NewReader(" word")),
				minimum: 2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanWhiteSpace(tt.args.s, tt.args.minimum); (err != nil) != tt.wantErr {
				t.Errorf("ScanWhiteSpace() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := tt.args.s.Text(); got != tt.expected {
				t.Errorf("ScanRunesUntil() s.Text() = %v, want s.Text() %v", got, tt.expected)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "random error isn't",
			err:  errors.New("hi"),
			want: false,
		},
		{
			name: "ScanNotFound error is",
			err:  ScanNotFound{LookingFor: []string{"123", "123"}},
			want: true,
		},
		{
			name: "Nested ScanNotFound error is",
			err:  fmt.Errorf("hi: %w", ScanNotFound{LookingFor: []string{"123", "123"}}),
			want: true,
		},
		{
			name: "ScanUntilNotFound error is",
			err:  ScanUntilNotFound{Until: "sadf"},
			want: true,
		},
		{
			name: "Nested ScanUntilNotFound error is",
			err:  fmt.Errorf("hi: %w", ScanUntilNotFound{Until: "123"}),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanFieldTerminator(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy path",
			args: args{
				s: NewScanner(strings.NewReader(";")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanFieldTerminator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanFieldTerminator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanRunesUntil(t *testing.T) {
	type args struct {
		s       *Scanner
		minimum int
		until   func([]byte) bool
		name    string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantText string
	}{
		{
			name: "Happy path",
			args: args{
				s:       NewScanner(strings.NewReader("let's scan to ... here: ; but no further")),
				minimum: 1,
				until: func(i []byte) bool {
					return bytes.EqualFold(i, []byte(";"))
				},
				name: ";",
			},
			wantText: "let's scan to ... here: ",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanRunesUntil(tt.args.s, tt.args.minimum, tt.args.until, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ScanRunesUntil() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.args.s.Text(), tt.wantText); diff != "" {
				t.Errorf("ScanRunesUntil() %s", diff)
			}
		})
	}
}

func TestScanUntilFieldTerminator(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantText string
	}{
		{
			name: "Happy path",
			args: args{
				s: NewScanner(strings.NewReader("let's scan to ... here: ; but no further")),
			},
			wantText: "let's scan to ... here: ",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilFieldTerminator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanUntilFieldTerminator() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.args.s.Text(), tt.wantText); diff != "" {
				t.Errorf("ScanUntilFieldTerminator() %s", diff)
			}
		})
	}
}
