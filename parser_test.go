package rcs

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
	"time"
)

var (
	//go:embed "testdata/testinput.go,v"
	testinputv []byte
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name    string
		r       []byte
		wantErr bool
	}{
		{
			name:    "Test parse of testinput.go,v",
			r:       testinputv,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(bytes.NewReader(tt.r))
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
			if err := ScanNewLine(tt.args.s); (err != nil) != tt.wantErr {
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
			err:  ScanNotFound([]string{"123", "123"}),
			want: true,
		},
		{
			name: "Nested ScanNotFound error is",
			err:  fmt.Errorf("hi: %w", ScanNotFound([]string{"123", "123"})),
			want: true,
		},
		{
			name: "ScanUntilNotFound error is",
			err:  ScanUntilNotFound("sadf"),
			want: true,
		},
		{
			name: "Nested ScanUntilNotFound error is",
			err:  fmt.Errorf("hi: %w", ScanUntilNotFound("123")),
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
		name    string
		args    args
		want    []*Lock
		wantErr bool
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
					Strict:   true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderLocks(tt.args.s, tt.args.havePropertyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaderLocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ParseHeaderLocks() %s", diff)
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
				Strict:   true,
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
			wantErr: true,
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
			wantNext: false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRH, gotNext, err := ParseRevisionHeader(tt.args.s)
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
	}{
		{
			name: "Basic branches parse",
			args: args{
				s:                 NewScanner(strings.NewReader("branches;\n")),
				rh:                &RevisionHead{},
				propertyNameKnown: false,
			},
			wantErr: false,
		},
		{
			name: "Basic branches parse - known",
			args: args{
				s:                 NewScanner(strings.NewReader(";\n")),
				rh:                &RevisionHead{},
				propertyNameKnown: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseRevisionHeaderBranches(tt.args.s, tt.args.rh, tt.args.propertyNameKnown); (err != nil) != tt.wantErr {
				t.Errorf("ParseRevisionHeaderBranches() error = %v, wantErr %v", err, tt.wantErr)
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
			got, err := ParseRevisionHeaders(tt.args.s)
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
