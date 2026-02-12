package rcs

import (
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
	"io/fs"
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
			ar := txtar.Parse(content)

			var inputJSON, expectedRCS string
			for _, f := range ar.Files {
				if f.Name == "input.json" {
					inputJSON = string(f.Data)
				}
				if f.Name == "expected,v" {
					expectedRCS = string(f.Data)
				}
			}

			if inputJSON == "" {
				// Skip if no input.json
				return
			}
			if expectedRCS == "" {
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

func TestStringLocalFiles(t *testing.T) {
	testRoundTrip(t, localTests, "testdata/local")
}

func TestStringRepoFiles(t *testing.T) {
	// Placeholder for future repo data tests
	// testRoundTrip(t, repoTests, "testdata/repo")
}

func testRoundTrip(t *testing.T, fsys fs.FS, root string) {
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ",v") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			b, err := fs.ReadFile(fsys, path)
			if err != nil {
				t.Errorf("ReadFile( %s ) error = %s", path, err)
				return
			}
			got, err := ParseFile(bytes.NewReader(b))
			if err != nil {
				t.Errorf("ParseFile( %s ) error = %s", path, err)
				return
			}
			if diff := cmp.Diff(strings.Split(got.String(), "\n"), strings.Split(string(b), "\n")); diff != "" {
				t.Errorf("String(): %s", diff)
			}
		})
		return nil
	})
	if err != nil {
		t.Logf("WalkDir error: %v", err)
	}
}
