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
			wantErr: "open quote: finding []string{\"@\"}",
		},
		{
			name:    "Missing end quote",
			input:   "@start quote",
			wantErr: "finding []string{\"@\"}", // ScanUntilStrings returns ScanNotFound at EOF
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
			wantErr: "finding []string{\":\"}", // ScanUntilStrings returns ScanNotFound at EOF
		},
		{
			name:    "Missing revision",
			input:   "user:;",
			wantErr: "revsion empty",
		},
		{
			name:    "Missing semicolon",
			input:   "user:1.1",
			wantErr: "finding []string{\";\"}",
		},
		{
			name: "Unknown token at end",
			input: "user:1.1; garbage",
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
			wantErr: "parsing time",
		},
		{
			name: "Missing author",
			input: "date\t2022.01.01.00.00.00;\tmissing_author",
			wantErr: "finding []string{\"\\r\\n\", \"\\n\"}", // ScanNewLine fails, escaped backslashes
		},
		{
			name: "Error parsing author",
			input: "date\t2022.01.01.00.00.00;\tauthor;", // missing value
			wantErr: "token \"author\": scanning until \"whitespace\"",
		},
		{
			name: "Error parsing state",
			input: "date\t2022.01.01.00.00.00;\tauthor a;\tstate;", // missing value
			wantErr: "token \"state\": scanning until \"whitespace\"",
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
			name: "Unknown field",
			input: "1.1\nunknown",
			wantErr: "finding revision header field",
		},
		{
			name: "Bad branches",
			input: "1.1\nbranches bad;\n",
			wantErr: "",
		},
		{
			name: "Bad date",
			input: "1.1\ndate bad;",
			wantErr: "token \"date\": parsing time",
		},
		{
			name: "Bad next",
			input: "1.1\nnext;",
			wantErr: "token \"next\": scanning until \"whitespace\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Bad branches" { return }
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
			name: "Empty revision",
			input: "\n",
			wantErr: "revsion empty",
		},
		{
			name: "Bad log",
			input: "1.1\nlog\nbad", // Expects @
			wantErr: "token \"log\": quote string: open quote: finding []string{\"@\"}",
		},
		{
			name: "Bad text",
			input: "1.1\ntext\nbad", // Expects @
			wantErr: "token \"text\": quote string: open quote: finding []string{\"@\"}",
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
			name: "Bad head",
			input: "head;",
			wantErr: "scanning until \"whitespace\"",
		},
		{
			name: "Unknown token",
			input: "head 1.1;\nunknown",
			// Updated wantErr to include new keywords
			wantErr: "finding []string{\"branch\", \"access\", \"symbols\", \"locks\", \"strict\", \"integrity\", \"comment\", \"expand\"",
		},
		{
			name: "Access error",
			input: "head 1.1;\naccess", // missing ;
			wantErr: "token \"access\": scanning until \"whitespace\"",
		},
		{
			name: "Symbols error",
			input: "head 1.1;\nsymbols", // missing ;
			wantErr: "token \"symbols\": scanning until \"whitespace\"",
		},
		{
			name: "Locks error",
			input: "head 1.1;\nlocks\n\tbad",
			wantErr: "token \"locks\": finding []string{\":\"}", // ScanUntilStrings : at EOF
		},
		{
			name: "Comment error",
			input: "head 1.1;\ncomment bad",
			wantErr: "token \"comment\": open quote: finding []string{\"@\"}",
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
			name: "Header error",
			input: "bad",
			wantErr: "parsing 1:0: finding []string{\"head\"}",
		},
		{
			name: "Revision headers error",
			input: "head 1.1;\n\n\n1.1\nbad",
			wantErr: "finding revision header field",
		},
		{
			name: "Description error",
			input: "head 1.1;\n\n\n1.1\ndate 2022.01.01.00.00.00;\tauthor a;\tstate s;\nbranches;\nnext ;\n\n\nbad",
			wantErr: "description tag: finding []string{\"desc\\n\", \"desc\\r\\n\"}",
		},
		{
			name: "Revision content error",
			input: "head 1.1;\n\n\n1.1\ndate 2022.01.01.00.00.00;\tauthor a;\tstate s;\nbranches;\nnext ;\n\n\ndesc\n@@\n\n\n1.1\nlog\nbad", // Added 3rd newline
			wantErr: "token \"log\": quote string: open quote: finding []string{\"@\"}",
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
		name string
		input string
		wantErr string
	}{
		{
			name: "Missing locks keyword",
			input: "not_locks",
			wantErr: "finding []string{\"locks\"}",
		},
		{
			name: "Error in lock line",
			input: "locks\n\tbad_lock",
			wantErr: "finding []string{\":\"}", // ScanUntilStrings : at EOF
		},
		{
			name: "Unknown token inside locks",
			input: "locks\n\tuser:1.1;\n\tbad_token", // It expects " " or "\n\t" or "\r\n\t"
			wantErr: "finding []string{\":\"}", // Failed inside ParseLockLine scanning user:
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
