package lcs

import (
	"path/filepath"
	"strings"
	"testing"

	rcstesting "github.com/arran4/golang-rcs/internal/testing"
	"golang.org/x/tools/txtar"
)

func TestLCSWithTxtar(t *testing.T) {
	files, err := filepath.Glob("testdata/*.txtar")
	if err != nil {
		t.Fatalf("failed to glob files: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatalf("failed to parse txtar file: %v", err)
			}

			var input1, input2 []string
			for _, f := range a.Files {
				switch strings.TrimSpace(f.Name) {
				case "input1.txt":
					input1 = splitLines(string(f.Data))
				case "input2.txt":
					input2 = splitLines(string(f.Data))
				}
			}

			if input1 == nil || input2 == nil {
				t.Fatalf("missing input1.txt or input2.txt in txtar file")
			}

			diff, err := GenerateEdDiffFromLines(input1, input2)
			if err != nil {
				t.Fatalf("GenerateEdDiffFromLines failed: %v", err)
			}

			// Verify apply
			r := rcstesting.NewStringLineReader(strings.Join(input1, "\n"))
			w := &rcstesting.StringLineWriter{}
			if err := diff.Apply(r, w); err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			got := strings.TrimSpace(w.String())
			want := strings.TrimSpace(strings.Join(input2, "\n"))

			if got != want {
				t.Errorf("Apply result mismatch.\nGot:\n%s\nWant:\n%s", got, want)
			}
		})
	}
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
