package rcs

import (
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"reflect"
	"strings"
	"testing"
)

var (
	//go:embed "testdata/testinput.go,v"
	testinputv []byte
)

func TestParseFile(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *File
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHead(t *testing.T) {
	type args struct {
		s   *bufio.Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test header of testinput.go,v",
			args: args{
				s:   bufio.NewScanner(bytes.NewReader(testinputv)),
				pos: &Pos{},
			},
			want:    "1.6",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHead(tt.args.s, tt.args.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseHead() got = %v, want %v", got, tt.want)
			}
		})
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
				line:   tt.fields.line,
				offset: tt.fields.offset,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanNewLine(t *testing.T) {
	type args struct {
		s   *bufio.Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Scans a unix new line",
			args: args{
				s:   bufio.NewScanner(strings.NewReader("\n")),
				pos: &Pos{},
			},
			wantErr: false,
		},
		{
			name: "Scans a windows new line",
			args: args{
				s:   bufio.NewScanner(strings.NewReader("\r\n")),
				pos: &Pos{},
			},
			wantErr: false,
		},
		{
			name: "Fails to scan nothing",
			args: args{
				s:   bufio.NewScanner(strings.NewReader("")),
				pos: &Pos{},
			},
			wantErr: true,
		},
		{
			name: "Fails to scan non new line data",
			args: args{
				s:   bufio.NewScanner(strings.NewReader("asdfasdfasdf")),
				pos: &Pos{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanNewLine(tt.args.s, tt.args.pos); (err != nil) != tt.wantErr {
				t.Errorf("ScanNewLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanStrings(t *testing.T) {
	type args struct {
		s    *bufio.Scanner
		pos  *Pos
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
				s:    bufio.NewScanner(strings.NewReader("This is a word")),
				pos:  &Pos{},
				strs: []string{"This"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanStrings(tt.args.s, tt.args.pos, tt.args.strs...); (err != nil) != tt.wantErr {
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
		s   *bufio.Scanner
		pos *Pos
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
				s:   bufio.NewScanner(strings.NewReader("This is\n a word")),
				pos: &Pos{},
			},
			wantErr: false,
		},
		{
			name:     "No new line no result",
			expected: "",
			args: args{
				s:   bufio.NewScanner(strings.NewReader("This is a word")),
				pos: &Pos{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilNewLine(tt.args.s, tt.args.pos); (err != nil) != tt.wantErr {
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
		s    *bufio.Scanner
		pos  *Pos
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
				s:    bufio.NewScanner(strings.NewReader("This is a word")),
				pos:  &Pos{},
				strs: []string{"word"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilStrings(tt.args.s, tt.args.pos, tt.args.strs...); (err != nil) != tt.wantErr {
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
		s       *bufio.Scanner
		pos     *Pos
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
				s:       bufio.NewScanner(strings.NewReader(" word")),
				pos:     &Pos{},
				minimum: 1,
			},
			wantErr: false,
		},
		{
			name:     "Minimum fails it",
			expected: "",
			args: args{
				s:       bufio.NewScanner(strings.NewReader(" word")),
				pos:     &Pos{},
				minimum: 2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanWhiteSpace(tt.args.s, tt.args.pos, tt.args.minimum); (err != nil) != tt.wantErr {
				t.Errorf("ScanWhiteSpace() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := tt.args.s.Text(); got != tt.expected {
				t.Errorf("ScanRunesUntil() s.Text() = %v, want s.Text() %v", got, tt.expected)
			}
		})
	}
}
