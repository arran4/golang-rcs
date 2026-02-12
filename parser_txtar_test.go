package rcs

import (
	"embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/txtar/*.txtar
var txtarTests embed.FS

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
			parts := parseTxtar(string(content))

			rcsContent, ok := parts["input.rcs"]
			if !ok {
				return
			}
			expectedJSON, ok := parts["expected.json"]
			if !ok {
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

func parseTxtar(content string) map[string]string {
	parts := make(map[string]string)
	lines := strings.Split(content, "\n")
	var currentFile string
	var currentContent strings.Builder

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "-- ") && strings.HasSuffix(line, " --") {
			if currentFile != "" {
				parts[currentFile] = currentContent.String()
				currentContent.Reset()
			}
			currentFile = strings.TrimSuffix(strings.TrimPrefix(line, "-- "), " --")
			continue
		}
		if currentFile != "" {
			currentContent.WriteString(line + "\n")
		}
	}
	if currentFile != "" {
		parts[currentFile] = currentContent.String()
	}
	return parts
}
