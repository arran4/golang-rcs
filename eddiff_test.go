package rcs

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

type StringLineReader struct {
	lines []string
	pos   int
}

func NewStringLineReader(content string) *StringLineReader {
	lines := strings.Split(content, "\n")
	// strings.Split always returns at least one element. If content is empty, returns [""] -> 1 line.
	// But empty file should have 0 lines?
	if content == "" {
		lines = []string{}
	} else {
		// If ends with newline, Split gives empty string at end.
		// RCS files often end with newline.
		// "foo\nbar\n" -> ["foo", "bar", ""].
		// We want ["foo", "bar"].
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}
	return &StringLineReader{lines: lines}
}

func (r *StringLineReader) ReadLine() (string, error) {
	if r.pos >= len(r.lines) {
		return "", io.EOF
	}
	line := r.lines[r.pos]
	r.pos++
	return line, nil
}

type StringLineWriter struct {
	lines []string
}

func (w *StringLineWriter) WriteLine(line string) error {
	w.lines = append(w.lines, line)
	return nil
}

func (w *StringLineWriter) String() string {
	if len(w.lines) == 0 {
		return ""
	}
	return strings.Join(w.lines, "\n") + "\n"
}

func TestParseEdDiff(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    EdDiff
		wantErr bool
	}{
		{
			name:  "Delete",
			input: "d1 1\n",
			want: EdDiff{
				Delete{1, 1},
			},
		},
		{
			name:  "Add",
			input: "a1 1\nfoo\n",
			want: EdDiff{
				Add{LineStart: 1, Lines: []string{"foo"}},
			},
		},
		{
			name:  "Multiple",
			input: "d1 1\na2 2\nfoo\nbar\n",
			want: EdDiff{
				Delete{1, 1},
				Add{LineStart: 2, Lines: []string{"foo", "bar"}},
			},
		},
		{
			name:    "Invalid command",
			input:   "x1 1\n",
			wantErr: true,
		},
		{
			name:    "Missing add lines",
			input:   "a1 1\n",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			input:   "d1\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEdDiff(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEdDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEdDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEdDiff_Apply(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		original string
		want     string
		wantErr  bool
	}{
		{
			name:     "Simple Delete",
			diff:     "d1 1\n",
			original: "line1\nline2\n",
			want:     "line2\n",
		},
		{
			name:     "Delete multiple",
			diff:     "d1 2\n",
			original: "line1\nline2\nline3\n",
			want:     "line3\n",
		},
		{
			name:     "Simple Add at start",
			diff:     "a0 1\nnew\n",
			original: "line1\n",
			want:     "new\nline1\n",
		},
		{
			name:     "Add after line 1",
			diff:     "a1 1\nnew\n",
			original: "line1\n",
			want:     "line1\nnew\n",
		},
		{
			name:     "Replace (Delete then Add)",
			diff:     "d1 1\na0 1\nnew\n",
			original: "line1\nline2\n",
			want:     "new\nline2\n",
		},
		{
			name:     "Replace middle",
			diff:     "d2 1\na1 1\nnew\n",
			original: "line1\nline2\nline3\n",
			want:     "line1\nnew\nline3\n",
		},
		{
			name:     "Delete at end",
			diff:     "d2 1\n",
			original: "line1\nline2\n",
			want:     "line1\n",
		},
		{
			name:     "Add at end",
			diff:     "a2 1\nnew\n",
			original: "line1\nline2\n",
			want:     "line1\nline2\nnew\n",
		},
		{
			name:     "Multiple disjoint edits",
			diff:     "d1 1\na2 1\nnew\n",
			original: "line1\nline2\nline3\n",
			want:     "line2\nnew\nline3\n",
		},
		{
			name:     "Empty file add",
			diff:     "a0 1\nfoo\n",
			original: "",
			want:     "foo\n",
		},
		{
			name:     "Delete out of bounds",
			diff:     "d2 1\n",
			original: "line1\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := ParseEdDiff(strings.NewReader(tt.diff))
			if err != nil {
				t.Fatalf("ParseEdDiff error: %v", err)
			}

			r := NewStringLineReader(tt.original)
			w := &StringLineWriter{}

			if err := ed.Apply(r, w); (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got := w.String()
				if got != tt.want {
					t.Errorf("Apply() result:\n%q\nwant:\n%q", got, tt.want)
				}
			}
		})
	}
}

func TestEdDiff_RoundTrip(t *testing.T) {
	diff := "d1 1\na2 2\nfoo\nbar\nd5 1\n"
	ed, err := ParseEdDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("ParseEdDiff error: %v", err)
	}

	got := ed.String()
	if strings.TrimSpace(got) != strings.TrimSpace(diff) {
		t.Errorf("RoundTrip mismatch:\nGot:\n%q\nWant:\n%q", got, diff)
	}
}
