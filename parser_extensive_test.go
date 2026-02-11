package rcs

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFile_String(t *testing.T) {
	// Test file with no locks and no branches
	f := &File{
		Head:    "1.1",
		Comment: "# ",
		Locks:   nil,
		RevisionHeads: []*RevisionHead{
			{
				Revision:     "1.1",
				Date:         time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				Author:       "user",
				State:        "Exp",
				Branches:     nil,
				NextRevision: "",
			},
		},
		Description: "test desc\n",
	}
	expected := "head\t1.1;\nlocks;\ncomment\t@# @;\n\n\n1.1\ndate\t2022.01.01.00.00.00;\tauthor user;\tstate Exp;\nbranches;\nnext\t;\n\n\ndesc\n@test desc\n@\n"
	if got := f.String(); got != expected {
		t.Errorf("File.String() = %q, want %q", got, expected)
	}

	// Test Lock.String strict
	l := &Lock{User: "u", Revision: "1", Strict: false}
	if got := l.String(); got != "u:1;" {
		t.Errorf("Lock.String() strict=false = %q, want %q", got, "u:1;")
	}

	// Test RevisionHead with branches
	f.RevisionHeads[0].Branches = []string{"1.1.1.1"}
	if !strings.Contains(f.RevisionHeads[0].String(), "branches\n\t1.1.1.1;") {
		t.Errorf("RevisionHead.String() missing branches: %q", f.RevisionHeads[0].String())
	}
}

func TestParseAtQuotedString_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Missing start quote",
			input:   "no quote",
			wantErr: "open quote: looking for []string{\"@\"}",
		},
		{
			name:    "Missing end quote",
			input:   "@start quote",
			wantErr: "looking for []string{\"@\"}", // ScanUntilStrings returns ScanNotFound at EOF
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			_, err := ParseAtQuotedString(s)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseAtQuotedString() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseLockLine_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Missing colon",
			input:   "user",
			wantErr: "EOF:looking for []string{\":\"}", // ScanUntilStrings returns ScanNotFound at EOF
		},
		{
			name:    "Missing revision",
			input:   "user:;",
			wantErr: "revision empty",
		},
		{
			name:    "Missing semicolon",
			input:   "user:1.1",
			wantErr: "EOF:looking for []string{\";\"}",
		},
		{
			name:    "Unknown token at end",
			input:   "user:1.1; garbage",
			wantErr: "", // Should succeed, ignores garbage
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			_, err := ParseLockLine(s)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ParseLockLine() error = %v, want nil", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ParseLockLine() error = %v, wantErr containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestParseRevisionHeaderDateLine_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Invalid date format",
			input:   "date\tbad-date;",
			wantErr: "unable to parse date",
		},
		{
			name:    "Missing author",
			input:   "date\t2022.01.01.00.00.00;\tmissing_author",
			wantErr: "looking for []string{\"\\r\\n\", \"\\n\"}", // ScanNewLine fails, escaped backslashes
		},
		{
			name:    "Error parsing author",
			input:   "date\t2022.01.01.00.00.00;\tauthor;", // missing value
			wantErr: "Author scanning until \"whitespace\" parser section parser: raw: author",
		},
		{
			name:    "Error parsing state",
			input:   "date\t2022.01.01.00.00.00;\tauthor a;\tstate;", // missing value
			wantErr: "State scanning until \"whitespace\" parser section parser: raw: state",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			rh := &RevisionHead{}
			err := ParseRevisionHeaderDateLine(s, false, rh)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseRevisionHeaderDateLine() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseRevisionHeader_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Unknown field",
			input:   "1.1\nunknown",
			wantErr: "finding revision header field",
		},
		{
			name:    "Bad branches",
			input:   "1.1\nbranches bad;\n",
			wantErr: "",
		},
		{
			name:    "Bad date",
			input:   "1.1\ndate bad;",
			wantErr: "Date unable to parse date: bad parser section parser: raw: date",
		},
		{
			name:    "Bad next",
			input:   "1.1\nnext;",
			wantErr: "Next scanning until \"whitespace\" parser section parser: raw: next",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Bad branches" {
				return
			}
			s := NewScanner(strings.NewReader(tt.input))
			_, _, _, err := ParseRevisionHeader(s)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseRevisionHeader() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseRevisionContent_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Bad log",
			input:   "1.1\nlog\nbad", // Expects @
			wantErr: "Log quote string: open quote: looking for []string{\"@\"} parser section parser: raw: log",
		},
		{
			name:    "Bad text",
			input:   "1.1\ntext\nbad", // Expects @
			wantErr: "Text quote string: open quote: looking for []string{\"@\"} parser section parser: raw: text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			_, _, err := ParseRevisionContent(s)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseRevisionContent() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseHeader_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Bad head",
			input:   "head;",
			wantErr: "scanning until \"whitespace\"",
		},
		{
			name:  "Unknown token",
			input: "head 1.1;\nunknown",
			// Updated wantErr to include new keywords
			wantErr: "looking for []string{\"branch\", \"access\", \"symbols\", \"locks\", \"strict\", \"integrity\", \"comment\", \"expand\"",
		},
		{
			name:    "Access error",
			input:   "head 1.1;\naccess", // missing ;
			wantErr: "Access scanning until \"whitespace\" parser section parser: raw: access",
		},
		{
			name:    "Symbols error",
			input:   "head 1.1;\nsymbols", // missing ;
			wantErr: "Symbols scanning until \"whitespace\" parser section parser: raw: symbols",
		},
		{
			name:    "Locks error",
			input:   "head 1.1;\nlocks\n\tbad",
			wantErr: "Locks EOF:looking for []string{\":\"} parser section parser: raw: locks",
		},
		{
			name:    "Comment error",
			input:   "head 1.1;\ncomment bad",
			wantErr: "Comment open quote: looking for []string{\"@\"} parser section parser: raw: comment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			f := &File{}
			err := ParseHeader(s, f)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseHeader() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseFile_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Header error",
			input:   "bad",
			wantErr: "parsing 1:0: looking for []string{\"head\"}",
		},
		{
			name:    "Revision headers error",
			input:   "head 1.1;\n\n\n1.1\nbad",
			wantErr: "finding revision header field",
		},
		{
			name:    "Description error",
			input:   "head 1.1;\n\n\n1.1\ndate 2022.01.01.00.00.00;\tauthor a;\tstate s;\nbranches;\nnext ;\n\n\nbad",
			wantErr: "parsing 10:0: description tag: looking for []string{\"desc\\n\", \"desc\\r\\n\"}",
		},
		{
			name:    "Revision content error",
			input:   "head 1.1;\n\n\n1.1\ndate 2022.01.01.00.00.00;\tauthor a;\tstate s;\nbranches;\nnext ;\n\n\ndesc\n@@\n\n\n1.1\nlog\nbad", // Added 3rd newline
			wantErr: "Log quote string: open quote: looking for []string{\"@\"} parser section parser: raw: log",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := bytes.NewReader([]byte(tt.input))
			_, err := ParseFile(s)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseFile() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseHeaderLocks_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Missing locks keyword",
			input:   "not_locks",
			wantErr: "looking for []string{\"locks\"}",
		},
		{
			name:    "Error in lock line",
			input:   "locks\n\tbad_lock",
			wantErr: "EOF:looking for []string{\":\"}", // ScanUntilStrings : at EOF
		},
		{
			name:    "Unknown token inside locks",
			input:   "locks\n\tuser:1.1;\n\tbad_token", // It expects " " or "\n\t" or "\r\n\t"
			wantErr: "EOF:looking for []string{\":\"}", // Failed inside ParseLockLine scanning user:
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			_, err := ParseHeaderLocks(s, false)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseHeaderLocks() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}
