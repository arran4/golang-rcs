package rcs

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
	"io/fs"
	"strings"
	"testing"
	"time"
)

var (
	//go:embed "testdata/testinput.go,v"
	testinputv []byte
	//go:embed "testdata/testinput1.go,v"
	testinputv1 []byte
	//go:embed testdata/txtar/*.txtar
	txtarTests embed.FS
	//go:embed testdata/local/*
	localTests embed.FS
	//go:embed "testdata/expand_integrity.go,v"
	expandIntegrityv []byte
	//go:embed "testdata/expand_integrity_unquoted.go,v"
	expandIntegrityUnquotedv []byte
	//go:embed "testdata/access_symbols.go,v"
	accessSymbolsv []byte
)

func TestParseAccessSymbols(t *testing.T) {
	f, err := ParseFile(bytes.NewReader(accessSymbolsv))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if diff := cmp.Diff(f.AccessUsers, []string{"john", "jane"}); diff != "" {
		t.Errorf("AccessUsers: %s", diff)
	}
	expectedMap := map[string]string{"rel": "1.1", "tag": "1.1.0.2"}
	if diff := cmp.Diff(expectedMap, f.SymbolMap); diff != "" {
		t.Errorf("SymbolMap: %s", diff)
	}
	if diff := cmp.Diff(f.Description, "Sample\n"); diff != "" {
		t.Errorf("Description: %s", diff)
	}

	if diff := cmp.Diff(f.String(), string(accessSymbolsv)); diff != "" {
		t.Errorf("String(): %s", diff)
	}
}

func TestParseHeaderExpandIntegrity(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		wantExpand    string
		wantIntegrity string
		wantErr       bool
	}{
		{
			name:          "Expand and Integrity with quotes",
			input:         expandIntegrityv,
			wantExpand:    "kv",
			wantIntegrity: "int123",
			wantErr:       false,
		},
		{
			name:          "Expand without quotes",
			input:         expandIntegrityUnquotedv,
			wantExpand:    "kv",
			wantIntegrity: "",
			wantErr:       true,
		},
		{
			name: "Integrity unquoted should fail",
			input: []byte(`head	1.1;
integrity	unquoted;
comment	@# @;


1.1
date	2022.01.01.00.00.00;	author arran;	state Exp;
branches;
next	;


desc
@@


1.1
log
@@
text
@@
`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseFile(bytes.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if f.Expand != tt.wantExpand {
				t.Errorf("Expand = %q, want %q", f.Expand, tt.wantExpand)
			}
			if f.Integrity != tt.wantIntegrity {
				t.Errorf("Integrity = %q, want %q", f.Integrity, tt.wantIntegrity)
			}

			gotString := f.String()
			f2, err := ParseFile(strings.NewReader(gotString))
			if err != nil {
				t.Errorf("ParseFile(f.String()) error = %v", err)
			} else {
				if f2.Expand != f.Expand {
					t.Errorf("RoundTrip Expand = %q, want %q", f2.Expand, f.Expand)
				}
				if f2.Integrity != f.Integrity {
					t.Errorf("RoundTrip Integrity = %q, want %q", f2.Integrity, f.Integrity)
				}
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name          string
		r             string
		b             []byte
		wantErr       bool
		wantErrString []string
	}{
		{
			name:    "Test parse of testinput.go,v",
			r:       string(testinputv),
			b:       testinputv,
			wantErr: false,
		},
		{
			name:    "Test parse of testinput1.go,v - add a new line for the missing one",
			r:       string(testinputv1) + "\n",
			b:       testinputv1,
			wantErr: false,
		},
		{
			name:    "Invalid header - missing head",
			r:       "invalid",
			b:       []byte("invalid"),
			wantErr: true,
			wantErrString: []string{
				"parsing",
				"looking for \"head\"",
			},
		},
		{
			name:    "Invalid property in header",
			r:       "head invalid",
			b:       []byte("head invalid"),
			wantErr: true,
			wantErrString: []string{
				"parsing",
				"scanning until",
			},
		},
		{
			name:    "Invalid revision header",
			r:       "head 1.1;\n\ninvalid\n",
			b:       []byte("head 1.1;\n\ninvalid\n"),
			wantErr: true,
			wantErrString: []string{
				"parsing",
				"finding revision header field",
			},
		},
		{
			name:    "Invalid description",
			r:       "head 1.1;\n\ndesc\ninvalid",
			b:       []byte("head 1.1;\n\ndesc\ninvalid"),
			wantErr: true,
			wantErrString: []string{
				"parsing",
				"quote string",
			},
		},
		{
			name:    "Invalid revision content",
			r:       "head 1.1;\n\ndesc\n@@\n\ninvalid\ninvalid",
			b:       []byte("head 1.1;\n\ndesc\n@@\n\ninvalid\ninvalid"),
			wantErr: true,
			wantErrString: []string{
				"parsing",
				"looking for",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(bytes.NewReader(tt.b))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrString != nil {
				for _, s := range tt.wantErrString {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("ParseFile() error = %v, want to contain %v", err, s)
					}
				}
				// If we expect an error, we don't need to check the rest
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
			if diff := cmp.Diff(got.String(), string(tt.r)); diff != "" {
				t.Errorf("String(): %s", diff)
			}
		})
	}
}

func TestParseHeaderHead(t *testing.T) {
	type args struct {
		s *Scanner
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
				s: NewScanner(bytes.NewReader(testinputv)),
			},
			want:    "1.6",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOptionalToken(tt.args.s, ScanTokenNum, WithPropertyName("head"), WithLine(true))
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
		s *Scanner
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
				s: NewScanner(bytes.NewReader(testinputv)),
			},
			want: &File{
				Head:    "1.6",
				Comment: "# ",
				Access:  true,
				Symbols: true,
				Locks: []*Lock{
					{
						User:     "arran",
						Revision: "1.6",
					},
				},
				Strict:          true,
				StrictOnOwnLine: false,
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

func TestParseAtQuotedString(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Base @ String",
			args: args{
				s: NewScanner(strings.NewReader("@# @")),
			},
			want:    "# ",
			wantErr: false,
		},
		{
			name: "Double @@ For literal",
			args: args{
				s: NewScanner(strings.NewReader("@ @@ @")),
			},
			want:    " @ ",
			wantErr: false,
		},
		{
			name: "New lines are fine",
			args: args{
				s: NewScanner(strings.NewReader("@Hello\nyou@")),
			},
			want:    "Hello\nyou",
			wantErr: false,
		},
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
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Simple description",
			args:    args{NewScanner(strings.NewReader("desc\n@This is a test file.\n@\n\n\n"))},
			want:    "This is a test file.\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDescription(tt.args.s, false)
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

func TestParseHeaderComment(t *testing.T) {
	type args struct {
		s                *Scanner
		havePropertyName bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Simple comment line need prop name",
			args: args{
				s:                NewScanner(strings.NewReader("comment\t@# @;\n\n")),
				havePropertyName: false,
			},
			want:    "# ",
			wantErr: false,
		},
		{
			name: "Simple comment line already have prop line",
			args: args{
				s:                NewScanner(strings.NewReader("\t@# @;\n\n")),
				havePropertyName: true,
			},
			want:    "# ",
			wantErr: false,
		},
		{
			name: "Simple comment line need prop name",
			args: args{
				s:                NewScanner(strings.NewReader("comment\t@# @;\n\n")),
				havePropertyName: true,
			},
			want:    "# ",
			wantErr: true,
		},
		{
			name: "Simple comment line already have prop line",
			args: args{
				s:                NewScanner(strings.NewReader("\t@# @;\n\n")),
				havePropertyName: false,
			},
			want:    "# ",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderComment(tt.args.s, tt.args.havePropertyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("ParseHeaderComment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHeaderLocks(t *testing.T) {
	type args struct {
		s                *Scanner
		havePropertyName bool
	}
	tests := []struct {
		name       string
		args       args
		want       []*Lock
		wantStrict bool
		wantErr    bool
	}{
		{
			name: "Single lock",
			args: args{
				s:                NewScanner(strings.NewReader("\n\tarran:1.6; strict;\ncomment\t@# @;")),
				havePropertyName: true,
			},
			want: []*Lock{
				{
					User:     "arran",
					Revision: "1.6",
				},
			},
			wantStrict: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, strict, _, err := ParseHeaderLocks(tt.args.s, tt.args.havePropertyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderLocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ParseHeaderLocks() %s", diff)
			}
			if strict != tt.wantStrict {
				t.Errorf("ParseHeaderLocks() strict = %v, want %v", strict, tt.wantStrict)
			}
		})
	}
}

func TestParseLockLine(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		want    *Lock
		wantErr bool
	}{
		{
			name: "Just a lock",
			args: args{
				s: NewScanner(strings.NewReader("arran:1.6; strict;\ncomment\t@# @;")),
			},
			want: &Lock{
				User:     "arran",
				Revision: "1.6",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLockLine(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLockLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ParseLockLine() %s", diff)
			}
		})
	}
}

func TestParseMultiLineText(t *testing.T) {
	type args struct {
		s                *Scanner
		havePropertyName bool
		propertyName     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Desc again - doesn't have prop",
			args: args{
				s:                NewScanner(strings.NewReader("desc\n@This is a test file.\n@\n\n\n")),
				havePropertyName: false,
				propertyName:     "desc",
			},
			want:    "This is a test file.\n",
			wantErr: false,
		},
		{
			name: "Desc again - has prop",
			args: args{
				s:                NewScanner(strings.NewReader("\n@This is a test file.\n@\n\n\n")),
				havePropertyName: true,
				propertyName:     "desc",
			},
			want:    "This is a test file.\n",
			wantErr: false,
		},
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
		{
			name: "A property without a line we know/have the prop already",
			args: args{
				s:                NewScanner(strings.NewReader("\tasdf;")),
				havePropertyName: true,
				propertyName:     "test123",
				line:             false,
			},
			want:    "asdf",
			wantErr: false,
		},
		{
			name: "A property with a line we know/have the prop already",
			args: args{
				s:                NewScanner(strings.NewReader("test123\tasdf;\n")),
				havePropertyName: false,
				propertyName:     "test123",
				line:             true,
			},
			want:    "asdf",
			wantErr: false,
		},
		{
			name: "A property without a line we know/have the prop already but we want a line",
			args: args{
				s:                NewScanner(strings.NewReader("\tasdf;")),
				havePropertyName: true,
				propertyName:     "test123",
				line:             true,
			},
			want:    "asdf",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProperty(tt.args.s, tt.args.havePropertyName, tt.args.propertyName, tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
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
		s *Scanner
	}
	tests := []struct {
		name     string
		args     args
		wantRC   *RevisionContent
		wantMore bool
		wantErr  bool
	}{
		{
			name: "1.1 first and last parse",
			args: args{
				s: NewScanner(strings.NewReader("1.2\nlog\n@New version\n@\ntext\n@a14 10\n\t//Feed in training data\n\tchain.Add(strings.Split(\"I want a cheese burger\", \" \"))\n\tchain.Add(strings.Split(\"I want a chilled sprite\", \" \"))\n\tchain.Add(strings.Split(\"I want to go to the movies\", \" \"))\n\n\t//Get transition probability of a sequence\n\tprob, _ := chain.TransitionProbability(\"a\", []string{\"I\", \"want\"})\n\tfmt.Println(prob)\n\t//Output: 0.6666666666666666\n\n@\n\n\n1.1\nlog\n@Initial revision\n@\ntext\n@d3 7\na9 1\nimport \"fmt\"\nd12 26\na37 1\n\tfmt.Println(\"HI\")\n@\n")),
			},
			wantRC: &RevisionContent{
				Revision: "1.2",
				Log:      "New version\n",
				Text:     "a14 10\n\t//Feed in training data\n\tchain.Add(strings.Split(\"I want a cheese burger\", \" \"))\n\tchain.Add(strings.Split(\"I want a chilled sprite\", \" \"))\n\tchain.Add(strings.Split(\"I want to go to the movies\", \" \"))\n\n\t//Get transition probability of a sequence\n\tprob, _ := chain.TransitionProbability(\"a\", []string{\"I\", \"want\"})\n\tfmt.Println(prob)\n\t//Output: 0.6666666666666666\n\n",
			},
			wantMore: true,
			wantErr:  false,
		},
		{
			name: "1.2 first but not last parse",
			args: args{
				s: NewScanner(strings.NewReader("1.1\nlog\n@Initial revision\n@\ntext\n@d3 7\na9 1\nimport \"fmt\"\nd12 26\na37 1\n\tfmt.Println(\"HI\")\n@\n")),
			},
			wantRC: &RevisionContent{
				Revision: "1.1",
				Log:      "Initial revision\n",
				Text:     "d3 7\na9 1\nimport \"fmt\"\nd12 26\na37 1\n\tfmt.Println(\"HI\")\n",
			},
			wantMore: false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRC, gotMore, err := ParseRevisionContent(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(gotRC, tt.wantRC); diff != "" {
				t.Errorf("ParseRevisionContent() %s", diff)
			}
			if gotMore != tt.wantMore {
				t.Errorf("ParseRevisionContent() gotMore = %v, want %v", gotMore, tt.wantMore)
			}
		})
	}
}

func TestParseRevisionContents(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		wantRcs []*RevisionContent
		wantErr bool
	}{
		{
			name: "1.2 first but not last parse",
			args: args{
				s: NewScanner(strings.NewReader("1.2\nlog\n@New version\n@\ntext\n@a14 10\n\t//Feed in training data\n\tchain.Add(strings.Split(\"I want a cheese burger\", \" \"))\n\tchain.Add(strings.Split(\"I want a chilled sprite\", \" \"))\n\tchain.Add(strings.Split(\"I want to go to the movies\", \" \"))\n\n\t//Get transition probability of a sequence\n\tprob, _ := chain.TransitionProbability(\"a\", []string{\"I\", \"want\"})\n\tfmt.Println(prob)\n\t//Output: 0.6666666666666666\n\n@\n\n\n1.1\nlog\n@Initial revision\n@\ntext\n@d3 7\na9 1\nimport \"fmt\"\nd12 26\na37 1\n\tfmt.Println(\"HI\")\n@\n")),
			},
			wantRcs: []*RevisionContent{
				{
					Revision: "1.2",
					Log:      "New version\n",
					Text:     "a14 10\n\t//Feed in training data\n\tchain.Add(strings.Split(\"I want a cheese burger\", \" \"))\n\tchain.Add(strings.Split(\"I want a chilled sprite\", \" \"))\n\tchain.Add(strings.Split(\"I want to go to the movies\", \" \"))\n\n\t//Get transition probability of a sequence\n\tprob, _ := chain.TransitionProbability(\"a\", []string{\"I\", \"want\"})\n\tfmt.Println(prob)\n\t//Output: 0.6666666666666666\n\n",
				},
				{
					Revision: "1.1",
					Log:      "Initial revision\n",
					Text:     "d3 7\na9 1\nimport \"fmt\"\nd12 26\na37 1\n\tfmt.Println(\"HI\")\n",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRcs, err := ParseRevisionContents(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(gotRcs, tt.wantRcs); diff != "" {
				t.Errorf("ParseRevisionContents() %s", diff)
			}
		})
	}
}

func TestParseRevisionHeader(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name     string
		args     args
		wantRH   *RevisionHead
		wantNext bool
		wantErr  bool
	}{
		{
			name: "Revision string 6",
			args: args{
				s: NewScanner(strings.NewReader("1.6\ndate\t2022.03.23.02.22.51;\tauthor arran;\tstate Exp;\nbranches;\nnext\t1.5;\n\n\n")),
			},
			wantRH: &RevisionHead{
				Revision:     "1.6",
				Date:         time.Date(2022, 3, 23, 2, 22, 51, 0, time.UTC),
				Author:       "arran",
				State:        "Exp",
				Branches:     []string{},
				NextRevision: "1.5",
			},
			wantNext: true,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRH, gotNext, _, err := ParseRevisionHeader(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(gotRH, tt.wantRH); diff != "" {
				t.Errorf("ParseRevisionHeader() %s", diff)
			}
			if gotNext != tt.wantNext {
				t.Errorf("ParseRevisionHeader() gotNext = %v, want %v", gotNext, tt.wantNext)
			}
		})
	}
}

func TestParseRevisionHeaderBranches(t *testing.T) {
	type args struct {
		s                 *Scanner
		rh                *RevisionHead
		propertyNameKnown bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []string
	}{
		{
			name: "Basic branches parse",
			args: args{
				s:                 NewScanner(strings.NewReader("branches;\n")),
				rh:                &RevisionHead{},
				propertyNameKnown: false,
			},
			wantErr: false,
			want:    []string{},
		},
		{
			name: "Basic branches parse - known",
			args: args{
				s:                 NewScanner(strings.NewReader(";\n")),
				rh:                &RevisionHead{},
				propertyNameKnown: true,
			},
			wantErr: false,
			want:    []string{},
		},
		{
			name: "Branches with numbers",
			args: args{
				s:                 NewScanner(strings.NewReader("branches\n\t1.1.1.1\n\t1.1.2.1;\n")),
				rh:                &RevisionHead{},
				propertyNameKnown: false,
			},
			wantErr: false,
			want:    []string{"1.1.1.1", "1.1.2.1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseRevisionHeaderBranches(tt.args.s, tt.args.rh, tt.args.propertyNameKnown); (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaderBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, tt.args.rh.Branches); diff != "" {
				t.Errorf("Branches diff: %s", diff)
			}
		})
	}
}

func TestParseRevisionHeaderDateLine(t *testing.T) {
	type args struct {
		s        *Scanner
		haveHead bool
		rh       *RevisionHead
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRh  *RevisionHead
	}{
		{
			name: "Date Line",
			args: args{
				s:        NewScanner(strings.NewReader("date\t2022.03.23.02.22.51;\tauthor arran;\tstate Exp;\n")),
				haveHead: false,
				rh:       &RevisionHead{},
			},
			wantErr: false,
			wantRh: &RevisionHead{
				Date:   time.Date(2022, 3, 23, 2, 22, 51, 0, time.UTC),
				Author: "arran",
				State:  "Exp",
			},
		},
		{
			name: "Date Line with a head",
			args: args{
				s:        NewScanner(strings.NewReader("\t2022.03.23.02.22.51;\tauthor arran;\tstate Exp;\n")),
				haveHead: true,
				rh:       &RevisionHead{},
			},
			wantErr: false,
			wantRh: &RevisionHead{
				Date:   time.Date(2022, 3, 23, 2, 22, 51, 0, time.UTC),
				Author: "arran",
				State:  "Exp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseRevisionHeaderDateLine(tt.args.s, tt.args.haveHead, tt.args.rh); (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaderDateLine() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.args.rh, tt.wantRh); diff != "" {
				t.Errorf("ParseRevisionHeader() %s", diff)
			}
		})
	}
}

func TestParseRevisionHeaderNext(t *testing.T) {
	type args struct {
		s        *Scanner
		haveHead bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Get Next with a head",
			args: args{
				s:        NewScanner(strings.NewReader("\t1.5;\n")),
				haveHead: true,
			},
			want:    "1.5",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []interface{}{WithPropertyName("next")}
			if tt.args.haveHead {
				opts = append(opts, WithConsumed(true))
			}
			opts = append(opts, WithLine(true))
			got, err := ParseOptionalToken(tt.args.s, ScanTokenNum, opts...)
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
		s *Scanner
	}
	tests := []struct {
		name      string
		args      args
		wantHeads []*RevisionHead
		wantErr   bool
	}{
		{
			name: "General parse",
			args: args{
				s: NewScanner(strings.NewReader("1.2\ndate\t2022.03.23.02.20.39;\tauthor arran;\tstate Exp;\nbranches;\nnext\t1.1;\n\n1.1\ndate\t2022.03.23.02.18.09;\tauthor arran;\tstate Exp;\nbranches;\nnext\t;\n\n\n")),
			},
			wantHeads: []*RevisionHead{
				{
					Revision:     "1.2",
					Date:         time.Date(2022, 3, 23, 2, 20, 39, 0, time.UTC),
					Author:       "arran",
					State:        "Exp",
					Branches:     []string{},
					NextRevision: "1.1",
				},
				{
					Revision:     "1.1",
					Date:         time.Date(2022, 3, 23, 2, 18, 9, 0, time.UTC),
					Author:       "arran",
					State:        "Exp",
					Branches:     []string{},
					NextRevision: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ParseRevisionHeaders(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.wantHeads); diff != "" {
				t.Errorf("ParseRevisionHeader() %s", diff)
			}
		})
	}
}

func TestParseTerminatorFieldLine(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy path - unix",
			args: args{
				s: NewScanner(strings.NewReader(";\n")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseTerminatorFieldLine(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ParseTerminatorFieldLine() error = %v, wantErr %v", err, tt.wantErr)
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

func TestRevisionHeadStringBranches(t *testing.T) {
	h := &RevisionHead{
		Revision:     "1.1",
		Date:         time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Author:       "test",
		State:        "Exp",
		Branches:     []string{"1.1.1.1", "1.1.2.1"},
		NextRevision: "",
	}
	want := "1.1\n" +
		"date\t2022.01.01.00.00.00;\tauthor test;\tstate Exp;\n" +
		"branches\n\t1.1.1.1\n\t1.1.2.1;\n" +
		"next\t;\n"
	if diff := cmp.Diff(h.String(), want); diff != "" {
		t.Errorf("RevisionHead.String() diff: %s", diff)
	}
}

func TestScanLockIdOrStrings(t *testing.T) {
	type args struct {
		s    *Scanner
		strs []string
	}
	tests := []struct {
		name      string
		args      args
		wantId    string
		wantMatch string
		wantErr   bool
	}{
		{
			name: "Match keyword",
			args: args{
				s:    NewScanner(strings.NewReader("keyword")),
				strs: []string{"keyword"},
			},
			wantId:    "",
			wantMatch: "keyword",
			wantErr:   false,
		},
		{
			name: "Match lock ID",
			args: args{
				s:    NewScanner(strings.NewReader("user:1.1")),
				strs: []string{"keyword"},
			},
			wantId:    "user",
			wantMatch: "",
			wantErr:   false,
		},
		{
			name: "Match nothing",
			args: args{
				s:    NewScanner(strings.NewReader("other")),
				strs: []string{"keyword"},
			},
			wantId:    "",
			wantMatch: "",
			wantErr:   true,
		},
		{
			name: "Empty input",
			args: args{
				s:    NewScanner(strings.NewReader("")),
				strs: []string{"keyword"},
			},
			wantId:    "",
			wantMatch: "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, gotMatch, err := ScanLockIdOrStrings(tt.args.s, tt.args.strs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanLockIdOrStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("ScanLockIdOrStrings() gotId = %v, want %v", gotId, tt.wantId)
			}
			if gotMatch != tt.wantMatch {
				t.Errorf("ScanLockIdOrStrings() gotMatch = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

func TestParseLockBody(t *testing.T) {
	type args struct {
		s    *Scanner
		user string
	}
	tests := []struct {
		name    string
		args    args
		want    *Lock
		wantErr bool
	}{
		{
			name: "Simple lock body",
			args: args{
				s:    NewScanner(strings.NewReader("1.1;")),
				user: "user",
			},
			want: &Lock{
				User:     "user",
				Revision: "1.1",
			},
			wantErr: false,
		},
		{
			name: "Strict lock body - strict is ignored here now",
			args: args{
				s:    NewScanner(strings.NewReader("1.1; strict;")),
				user: "user",
			},
			want: &Lock{
				User:     "user",
				Revision: "1.1",
			},
			wantErr: false,
		},
		{
			name: "Empty revision",
			args: args{
				s:    NewScanner(strings.NewReader(";")),
				user: "user",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLockBody(tt.args.s, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLockBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ParseLockBody() %s", diff)
			}
		})
	}
}

func TestParseTxtarFiles(t *testing.T) {
	files, err := txtarTests.ReadDir("testdata/txtar")
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txtar") {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			content, err := txtarTests.ReadFile("testdata/txtar/" + f.Name())
			if err != nil {
				t.Fatalf("ReadFile error: %v", err)
			}
			ar := txtar.Parse(content)

			var rcsContent, expectedJSON string
			for _, f := range ar.Files {
				if f.Name == "input.rcs" {
					rcsContent = strings.ReplaceAll(string(f.Data), "\r\n", "\n")
				}
				if f.Name == "expected.json" {
					expectedJSON = strings.ReplaceAll(string(f.Data), "\r\n", "\n")
				}
			}

			if rcsContent == "" {
				t.Fatalf("input.rcs not found in %s", f.Name())
			}
			if expectedJSON == "" {
				t.Fatalf("expected.json not found in %s", f.Name())
			}

			// Parse RCS
			parsedFile, err := ParseFile(strings.NewReader(rcsContent))
			if err != nil {
				// Retry with added newlines if parsing failed, assuming it might be due to missing EOF markers
				parsedFile, err = ParseFile(strings.NewReader(rcsContent + "\n\n\n"))
				if err != nil {
					t.Fatalf("ParseFile error: %v", err)
				}
			}

			// Marshal to JSON
			gotJSONBytes, err := json.MarshalIndent(parsedFile, "", "  ")
			if err != nil {
				t.Fatalf("json.MarshalIndent error: %v", err)
			}
			gotJSON := string(gotJSONBytes)

			// Normalize JSON for comparison (trim whitespace)
			gotJSON = strings.TrimSpace(gotJSON)
			expectedJSON = strings.TrimSpace(expectedJSON)

			if diff := cmp.Diff(expectedJSON, gotJSON); diff != "" {
				t.Errorf("JSON mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseLocalFiles(t *testing.T) {
	testParseFiles(t, localTests, "testdata/local")
}

func TestParseRepoFiles(t *testing.T) {
	// Placeholder for future repo data tests
	// testParseFiles(t, repoTests, "testdata/repo")
}

func testParseFiles(t *testing.T, fsys fs.FS, root string) {
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ",v") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			b, err := fs.ReadFile(fsys, path)
			if err != nil {
				t.Errorf("ReadFile( %s ) error = %s", path, err)
				return
			}
			_, err = ParseFile(bytes.NewReader(b))
			if err != nil {
				t.Errorf("ParseFile( %s ) error = %s", path, err)
				return
			}
		})
		return nil
	})
	if err != nil {
		t.Logf("WalkDir error: %v", err)
	}
}

func TestParseFile_TruncatedYear(t *testing.T) {
	// Input with 2-digit year "99" (1999)
	input := `head	1.1;
access;
symbols;
locks; strict;
comment	@# @;


1.1
date	99.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@Description
@


1.1
log
@Initial revision
@
text
@Content
@
`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check if the date was parsed correctly as 1999
	expectedDate := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.RevisionHeads[0].Date.Equal(expectedDate) {
		t.Errorf("Date parsed incorrectly: got %v, want %v", f.RevisionHeads[0].Date, expectedDate)
	}

	// This is the check that will fail before implementation
	if !f.DateYearPrefixTruncated {
		t.Errorf("DateYearPrefixTruncated should be true for 2-digit year")
	}

	// Check serialization
	// We expect the output to also have "99.01.01..." if we support preserving it.
	// Currently it will likely output "1999.01.01..."
	if got, want := f.RevisionHeads[0].String(), "1.1\ndate\t99.01.01.00.00.00;\tauthor user;\tstate Exp;\nbranches;\nnext\t;\n"; got != want {
		t.Errorf("Output for revision head should contain truncated date '99.01.01.00.00.00;', got:\n%s\nwant:\n%s", got, want)
	}
}

func TestParseIntegrity(t *testing.T) {
	input := `head	1.1;
integrity	@some @@ value@;
comment	@This is a comment@;

desc
@@


`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if f.Integrity != "some @ value" {
		t.Errorf("expected Integrity 'some @ value', got %q", f.Integrity)
	}
}

func TestParseIntegrityUnquoted(t *testing.T) {
	input := `head	1.1;
integrity	simplevalue;
comment	@This is a comment@;

desc
@@


`
	_, err := ParseFile(strings.NewReader(input))
	if err == nil {
		t.Errorf("expected error for unquoted integrity, but got nil")
	} else if !strings.Contains(err.Error(), "looking for \"@\"") {
		// ParseHeaderComment -> ParseAtQuotedString -> ScanStrings("@") -> ScanNotFound -> Error()
		t.Errorf("expected 'looking for \"@\"' error, got %q", err)
	}
}

func TestStringIntegrity(t *testing.T) {
	f := &File{
		Head:        "1.1",
		Integrity:   "some @ value",
		Comment:     "This is a comment",
		Description: "",
	}
	s := f.String()
	expected := "integrity\t@some @@ value@;\n"
	if !strings.Contains(s, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, s)
	}
}

func TestParseRevisionHeaderWithExtraFields(t *testing.T) {
	input := "1.2\n" +
		"date\t99.01.12.14.05.31;\tauthor lhecking;\tstate dead;\n" +
		"branches;\n" +
		"next\t1.1;\n" +
		"owner\t640;\n" +
		"group\t15;\n" +
		"permissions\t644;\n" +
		"hardlinks\t@stringize.m4@;\n" +
		"\n\n"

	s := NewScanner(strings.NewReader(input))
	rh, _, _, err := ParseRevisionHeader(s)
	if err != nil {
		t.Fatalf("ParseRevisionHeader returned error: %v", err)
	}

	if rh.Revision != "1.2" {
		t.Errorf("Revision = %q, want %q", rh.Revision, "1.2")
	}
	if rh.Owner != "640" {
		t.Errorf("Owner = %q, want %q", rh.Owner, "640")
	}
	if rh.Group != "15" {
		t.Errorf("Group = %q, want %q", rh.Group, "15")
	}
	if rh.Permissions != "644" {
		t.Errorf("Permissions = %q, want %q", rh.Permissions, "644")
	}
	if rh.Hardlinks != "stringize.m4" {
		t.Errorf("Hardlinks = %q, want %q", rh.Hardlinks, "stringize.m4")
	}

	// Verify String() output
	expectedOutput := "1.2\n" +
		"date\t99.01.12.14.05.31;\tauthor lhecking;\tstate dead;\n" +
		"branches;\n" +
		"next\t1.1;\n" +
		"owner\t640;\n" +
		"group\t15;\n" +
		"permissions\t644;\n" +
		"hardlinks\t@stringize.m4@;\n"

	if diff := cmp.Diff(rh.String(), expectedOutput); diff != "" {
		t.Errorf("String() mismatch (-want +got):\n%s", diff)
	}
}
