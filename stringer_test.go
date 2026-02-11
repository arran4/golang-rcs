package rcs

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func TestStringTxtarFiles(t *testing.T) {
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

			inputJSON, ok := parts["input.json"]
			if !ok {
				// Skip if no input.json
				return
			}
			expectedRCS, ok := parts["expected,v"]
			if !ok {
				// Skip if no expected,v
				return
			}

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
			}
		})
	}
}
