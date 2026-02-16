package hashline

import (
	"embed"
	"path/filepath"
	"strings"
	"testing"

	rcstesting "github.com/arran4/golang-rcs/internal/testing"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/*.txtar
var testData embed.FS

func TestHashLineWithTxtar(t *testing.T) {
	files, err := testData.ReadDir("testdata")
	if err != nil {
		t.Fatalf("failed to read testdata dir: %v", err)
	}

	for _, fileEntry := range files {
		if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".txtar") {
			continue
		}
		file := filepath.Join("testdata", fileEntry.Name())
		t.Run(fileEntry.Name(), func(t *testing.T) {
			content, err := testData.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", file, err)
			}
			a := txtar.Parse(content)

			var input1, input2 []string
			var expectedDiff string
			for _, f := range a.Files {
				switch strings.TrimSpace(f.Name) {
				case "input1.txt":
					input1 = splitLines(string(f.Data))
				case "input2.txt":
					input2 = splitLines(string(f.Data))
				case "expected.diff":
					expectedDiff = strings.TrimSpace(string(f.Data))
				}
			}

			if input1 == nil || input2 == nil {
				t.Fatalf("missing input1.txt or input2.txt in txtar file")
			}

			diff, err := GenerateEdDiffFromLines(input1, input2)
			if err != nil {
				t.Fatalf("GenerateEdDiffFromLines failed: %v", err)
			}

			// Verify generated diff string matches expected.diff if present
			// Note: hashline produces LCS-like output so it should match the expectations
			// derived from LCS if they are minimal/standard.
			if expectedDiff != "" {
				gotDiff := strings.TrimSpace(diff.String())
				if gotDiff != expectedDiff {
					t.Errorf("Diff output mismatch.\nGot:\n%s\nWant:\n%s", gotDiff, expectedDiff)
				}
			}

			// Verify apply (Round Trip)
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
