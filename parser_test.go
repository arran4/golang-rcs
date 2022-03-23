package rcs

import (
	"bufio"
	"bytes"
	_ "embed"
	"github.com/google/go-cmp/cmp"
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
	tests := []struct {
		name    string
		r       io.Reader
		wantErr bool
	}{
		{
			name:    "Test parse of testinput.go,v",
			r:       bytes.NewReader(testinputv),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got.Description, "This is a test file.\n"); diff != "" {
				t.Errorf("Description: %s", diff)
			}
			if diff := cmp.Diff(len(got.Locks), 1); diff != "" {
				t.Errorf("Locks: %s", diff)
			}
			if diff := cmp.Diff(len(got.RevisionHeads), 6); diff != "" {
				t.Errorf("RevisionHeads: %s", diff)
			}
			if diff := cmp.Diff(len(got.RevisionContents), 6); diff != "" {
				t.Errorf("RevisionContents: %s", diff)
			}
		})
	}
}

func TestParseHeaderHead(t *testing.T) {
	type args struct {
		s   *Scanner
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
				s:   NewScanner(bytes.NewReader(testinputv)),
				pos: &Pos{},
			},
			want:    "1.6",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderHead(tt.args.s, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderHead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseHeaderHead() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    *File
		wantErr bool
	}{
		{
			name: "Test header of testinput.go,v",
			args: args{
				s:   NewScanner(bytes.NewReader(testinputv)),
				pos: &Pos{},
			},
			want: &File{
				Head:    "1.6",
				Comment: "# ",
				Access:  true,
				Symbols: true,
				Locks: []*Lock{
					&Lock{
						User:     "arran",
						Revision: "1.6",
						Strict:   true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &File{}
			err := ParseHeader(tt.args.s, f)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderHead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, f); diff != "" {
				t.Errorf("ParseHeader() Diff: %s", diff)
			}
		})
	}
}

func TestScanNewLine(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Scans a unix new Line",
			args: args{
				s:   NewScanner(strings.NewReader("\n")),
				pos: &Pos{},
			},
			wantErr: false,
		},
		{
			name: "Scans a windows new Line",
			args: args{
				s:   NewScanner(strings.NewReader("\r\n")),
				pos: &Pos{},
			},
			wantErr: false,
		},
		{
			name: "Fails to scan nothing",
			args: args{
				s:   NewScanner(strings.NewReader("")),
				pos: &Pos{},
			},
			wantErr: true,
		},
		{
			name: "Fails to scan non new Line data",
			args: args{
				s:   NewScanner(strings.NewReader("asdfasdfasdf")),
				pos: &Pos{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanNewLine(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanNewLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanStrings(t *testing.T) {
	type args struct {
		s    *Scanner
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
				s:    NewScanner(strings.NewReader("This is a word")),
				pos:  &Pos{},
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
		s   *Scanner
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
				s:   NewScanner(strings.NewReader("This is\n a word")),
				pos: &Pos{},
			},
			wantErr: false,
		},
		{
			name:     "No new Line no result",
			expected: "",
			args: args{
				s:   NewScanner(strings.NewReader("This is a word")),
				pos: &Pos{},
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
				s:    NewScanner(strings.NewReader("This is a word")),
				pos:  &Pos{},
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
				s:       NewScanner(strings.NewReader(" word")),
				pos:     &Pos{},
				minimum: 1,
			},
			wantErr: false,
		},
		{
			name:     "Minimum fails it",
			expected: "",
			args: args{
				s:       NewScanner(strings.NewReader(" word")),
				pos:     &Pos{},
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
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.args.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewScanner(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want *Scanner
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScanner(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScanner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAtQuotedString(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAtQuotedString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAtQuotedString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAtQuotedString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDescription(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDescription(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDescription() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFile1(t *testing.T) {
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

func TestParseHeader1(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
		f   *File
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseHeader(tt.args.s, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("ParseHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseHeaderComment(t *testing.T) {
	type args struct {
		s                *Scanner
		pos              *Pos
		havePropertyName bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderComment(tt.args.s, tt.args.havePropertyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseHeaderComment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHeaderHead1(t *testing.T) {
	type args struct {
		s        *Scanner
		pos      *Pos
		haveHead bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderHead(tt.args.s, tt.args.haveHead)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderHead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseHeaderHead() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHeaderLocks(t *testing.T) {
	type args struct {
		s                *Scanner
		pos              *Pos
		havePropertyName bool
	}
	tests := []struct {
		name    string
		args    args
		want    []*Lock
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderLocks(tt.args.s, tt.args.havePropertyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderLocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseHeaderLocks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLockLine(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    *Lock
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLockLine(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLockLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLockLine() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMultiLineText(t *testing.T) {
	type args struct {
		s                *Scanner
		pos              *Pos
		havePropertyName bool
		propertyName     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMultiLineText(tt.args.s, tt.args.havePropertyName, tt.args.propertyName, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMultiLineText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseMultiLineText() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseProperty(t *testing.T) {
	type args struct {
		s                *Scanner
		pos              *Pos
		havePropertyName bool
		propertyName     string
		line             bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProperty(tt.args.s, tt.args.havePropertyName, tt.args.propertyName, tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseProperty() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRevisionContent(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    *RevisionContent
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseRevisionContent(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRevisionContent() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseRevisionContent() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParseRevisionContentLog(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRevisionContentLog(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionContentLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRevisionContentLog() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRevisionContentText(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRevisionContentText(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionContentText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRevisionContentText() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRevisionContents(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    []*RevisionContent
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRevisionContents(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRevisionContents() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRevisionHeader(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    *RevisionHead
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseRevisionHeader(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRevisionHeader() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseRevisionHeader() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParseRevisionHeaderBranches(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
		rh  *RevisionHead
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseRevisionHeaderBranches(tt.args.s, tt.args.rh); (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaderBranches() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseRevisionHeaderDateLine(t *testing.T) {
	type args struct {
		s        *Scanner
		pos      *Pos
		haveHead bool
		rh       *RevisionHead
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseRevisionHeaderDateLine(tt.args.s, tt.args.haveHead, tt.args.rh); (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaderDateLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseRevisionHeaderNext(t *testing.T) {
	type args struct {
		s        *Scanner
		pos      *Pos
		haveHead bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRevisionHeaderNext(tt.args.s, tt.args.haveHead)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaderNext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRevisionHeaderNext() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRevisionHeaders(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		want    []*RevisionHead
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRevisionHeaders(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRevisionHeaders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTerminatorFieldLine(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseTerminatorFieldLine(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ParseTerminatorFieldLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPos_String1(t *testing.T) {
	type fields struct {
		line   int
		offset int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
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

func TestScanFieldTerminator(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanFieldTerminator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanFieldTerminator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanNewLine1(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanNewLine(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanNewLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanNotFound_Error(t *testing.T) {
	tests := []struct {
		name string
		se   ScanNotFound
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.se.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanRunesUntil(t *testing.T) {
	type args struct {
		s       *Scanner
		pos     *Pos
		minimum int
		until   func([]byte) bool
		name    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanRunesUntil(tt.args.s, tt.args.minimum, tt.args.until, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ScanRunesUntil() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanStrings1(t *testing.T) {
	type args struct {
		s    *Scanner
		pos  *Pos
		strs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanStrings(tt.args.s, tt.args.strs...); (err != nil) != tt.wantErr {
				t.Errorf("ScanStrings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanUntilFieldTerminator(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilFieldTerminator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanUntilFieldTerminator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanUntilNewLine1(t *testing.T) {
	type args struct {
		s   *Scanner
		pos *Pos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilNewLine(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ScanUntilNewLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanUntilNotFound_Error(t *testing.T) {
	tests := []struct {
		name string
		se   ScanUntilNotFound
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.se.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanUntilStrings1(t *testing.T) {
	type args struct {
		s    *Scanner
		pos  *Pos
		strs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanUntilStrings(tt.args.s, tt.args.strs...); (err != nil) != tt.wantErr {
				t.Errorf("ScanUntilStrings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanWhiteSpace1(t *testing.T) {
	type args struct {
		s       *Scanner
		pos     *Pos
		minimum int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScanWhiteSpace(tt.args.s, tt.args.minimum); (err != nil) != tt.wantErr {
				t.Errorf("ScanWhiteSpace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanner_LastScan(t *testing.T) {
	type fields struct {
		Scanner  *bufio.Scanner
		sf       bufio.SplitFunc
		lastScan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				Scanner:  tt.fields.Scanner,
				sf:       tt.fields.sf,
				lastScan: tt.fields.lastScan,
			}
			if got := s.LastScan(); got != tt.want {
				t.Errorf("LastScan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_Scan(t *testing.T) {
	type fields struct {
		Scanner  *bufio.Scanner
		sf       bufio.SplitFunc
		lastScan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				Scanner:  tt.fields.Scanner,
				sf:       tt.fields.sf,
				lastScan: tt.fields.lastScan,
			}
			if got := s.Scan(); got != tt.want {
				t.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_Split(t *testing.T) {
	type fields struct {
		Scanner  *bufio.Scanner
		sf       bufio.SplitFunc
		lastScan bool
	}
	type args struct {
		split bufio.SplitFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//s := &Scanner{
			//	Scanner:  tt.fields.Scanner,
			//	sf:       tt.fields.sf,
			//	lastScan: tt.fields.lastScan,
			//}
		})
	}
}

func TestScanner_scannerWrapper(t *testing.T) {
	type fields struct {
		Scanner  *bufio.Scanner
		sf       bufio.SplitFunc
		lastScan bool
	}
	type args struct {
		data []byte
		eof  bool
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantAdvance int
		wantToken   []byte
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				Scanner:  tt.fields.Scanner,
				sf:       tt.fields.sf,
				lastScan: tt.fields.lastScan,
			}
			gotAdvance, gotToken, err := s.scannerWrapper(tt.args.data, tt.args.eof)
			if (err != nil) != tt.wantErr {
				t.Errorf("scannerWrapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("scannerWrapper() gotAdvance = %v, want %v", gotAdvance, tt.wantAdvance)
			}
			if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("scannerWrapper() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
