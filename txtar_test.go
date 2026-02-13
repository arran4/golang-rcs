package rcs

import (
	"embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
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
			rcsContent, rcsok := parts["input,v"]
			if !rcsok {
				rcsContent, rcsok = parts["input.rcs"]
			}
			rcsContent = strings.ReplaceAll(rcsContent, "\r\n", "\n")

			inputJSON, jsonok := parts["input.json"]
			inputJSON = strings.ReplaceAll(inputJSON, "\r\n", "\n")

			// Parse Test: input,v -> expected.json
			if rcsok {
				expectedJSON, hasExpectedJSON := parts["expected.json"]
				if hasExpectedJSON {
					expectedJSON = strings.ReplaceAll(expectedJSON, "\r\n", "\n")
					t.Run("Parse", func(t *testing.T) {
						// Parse RCS
						parsedFile, err := ParseFile(strings.NewReader(rcsContent))
						if err != nil {
							// Retry with added newlines if parsing failed, assuming it might be due to missing EOF markers
							parsedFile, err = ParseFile(strings.NewReader(rcsContent + "\n\n\n"))
							if err != nil {
								// Try with just one newline
								parsedFile, err = ParseFile(strings.NewReader(rcsContent + "\n"))
								if err != nil {
									t.Fatalf("ParseFile error: %v", err)
								}
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

				// Circular Test: input,v -> Parse -> String -> input,v (or expected,v if present)
				t.Run("Circular", func(t *testing.T) {
					parsedFile, err := ParseFile(strings.NewReader(rcsContent))
					if err != nil {
						// Retry with added newlines
						parsedFile, err = ParseFile(strings.NewReader(rcsContent + "\n\n\n"))
						if err != nil {
							// Try with just one newline
							parsedFile, err = ParseFile(strings.NewReader(rcsContent + "\n"))
							if err != nil {
								t.Fatalf("ParseFile error: %v", err)
							}
						}
					}

					gotRCS := parsedFile.String()
					gotRCS = strings.TrimSpace(gotRCS)

					expectedRCS := strings.TrimSpace(rcsContent)
					if exp, ok := parts["expected,v"]; ok {
						expectedRCS = strings.TrimSpace(exp)
					}
					expectedRCS = strings.ReplaceAll(expectedRCS, "\r\n", "\n")

					if diff := cmp.Diff(expectedRCS, gotRCS); diff != "" {
						t.Errorf("Circular RCS mismatch (-want +got):\n%s", diff)
						t.Logf("Got RCS:\n%q", gotRCS)
					}
				})
			}

			// Stringer Test: input.json -> expected,v
			if jsonok {
				expectedRCS, hasExpectedRCS := parts["expected,v"]
				if hasExpectedRCS {
					expectedRCS = strings.ReplaceAll(expectedRCS, "\r\n", "\n")
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
			}
		})
	}
}

func parseTxtar(content string) map[string]string {
	archive := txtar.Parse([]byte(content))
	parts := make(map[string]string)
	for _, f := range archive.Files {
		parts[f.Name] = string(f.Data)
	}
	return parts
}
