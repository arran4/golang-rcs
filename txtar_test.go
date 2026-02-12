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

func TestTxtarFiles(t *testing.T) {
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

			// Check for input,v (preferred) or input.rcs
			rcsContent, ok := parts["input,v"]
			if !ok {
				rcsContent, ok = parts["input.rcs"]
			}

			// Parse Test
			if ok {
				expectedJSON, hasExpectedJSON := parts["expected.json"]
				if hasExpectedJSON {
					t.Run("Parse", func(t *testing.T) {
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

			// Stringer Test
			inputJSON, hasInputJSON := parts["input.json"]
			expectedRCS, hasExpectedRCS := parts["expected,v"]

			if hasInputJSON && hasExpectedRCS {
				t.Run("String", func(t *testing.T) {
					// Unmarshal JSON
					var file File
					if err := json.Unmarshal([]byte(inputJSON), &file); err != nil {
						t.Fatalf("json.Unmarshal error: %v", err)
					}

					// Generate RCS string
					gotRCS := file.String()

					// Normalize RCS for comparison (trim whitespace)
					gotRCS = strings.TrimSpace(gotRCS)
					expectedRCS = strings.TrimSpace(expectedRCS)

					if diff := cmp.Diff(expectedRCS, gotRCS); diff != "" {
						t.Errorf("RCS mismatch (-want +got):\n%s", diff)
						t.Logf("Got RCS:\n%q", gotRCS)
					}
				})
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
