package diff

import (
	"embed"
	"path"
	"strings"
	"testing"

	rcstesting "github.com/arran4/golang-rcs/internal/testing"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/*.txtar
var testData embed.FS

func TestDiffWithTxtar(t *testing.T) {
	files, err := testData.ReadDir("testdata")
	if err != nil {
		t.Fatalf("failed to read testdata dir: %v", err)
	}

	for _, fileEntry := range files {
		if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".txtar") {
			continue
		}
		file := path.Join("testdata", fileEntry.Name())
		t.Run(fileEntry.Name(), func(t *testing.T) {
			content, err := testData.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", file, err)
			}
			if len(content) > 0 {
				content = []byte(strings.ReplaceAll(string(content), "\r\n", "\n"))
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

			// Test the default Generate function (which uses LCS/local implementation)
			diff, err := Generate(input1, input2)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			if expectedDiff != "" {
				gotDiff := strings.TrimSpace(diff.String())
				if gotDiff != expectedDiff {
					t.Errorf("Diff output mismatch.\nGot:\n%s\nWant:\n%s", gotDiff, expectedDiff)
				}
			}

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
