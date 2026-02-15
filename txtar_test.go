package rcs

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/txtar/*.txtar testdata/txtar/operations/*.txtar
var txtarTests embed.FS

var ErrorTestProvider = map[string]func(b []byte) (error, error){
	"ErrParseProperty": func(b []byte) (error, error) {
		var targetErr ErrParseProperty
		err := json.Unmarshal(b, &targetErr)
		if targetErr.Err == nil {
			targetErr.Err = errors.New("placeholder")
		}
		return targetErr, err
	},
	"ErrTooManyNewLines": func(b []byte) (error, error) {
		return ErrTooManyNewLines, nil
	},
	"ErrEmptyId": func(b []byte) (error, error) {
		return ErrEmptyId, nil
	},
	"ErrRevisionEmpty": func(b []byte) (error, error) {
		return ErrRevisionEmpty, nil
	},
	"ErrDateParse": func(b []byte) (error, error) {
		return ErrDateParse, nil
	},
	"ErrUnknownToken": func(b []byte) (error, error) {
		return ErrUnknownToken, nil
	},
}

func TestTxtarFiles(t *testing.T) {
	err := fs.WalkDir(txtarTests, "testdata/txtar", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".txtar") {
			return nil
		}
		t.Run(d.Name(), func(t *testing.T) {
			runTest(t, txtarTests, path)
		})
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}
}

func runTest(t *testing.T, fsys fs.FS, filename string) {
	content, err := fs.ReadFile(fsys, filename)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	archive := txtar.Parse(content)
	parts := make(map[string]string)
	for _, f := range archive.Files {
		parts[f.Name] = string(f.Data)
	}

	// description.txt check
	if _, ok := parts["description.txt"]; !ok {
		t.Log("description.txt is missing")
	}

	// options.conf
	options := make(map[string]bool)
	if optContent, ok := parts["options.conf"]; ok {
		scanner := bufio.NewScanner(strings.NewReader(optContent))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "* ") {
				options[strings.TrimPrefix(line, "* ")] = true
			}
		}
	}

	// tests.txt or tests.md
	testContent, ok := parts["tests.txt"]
	if !ok {
		testContent, ok = parts["tests.md"]
	}

	// Fallback for migration: if tests.txt is missing, try to guess based on old logic
	if !ok {
		t.Skip("Missing tests.txt or tests.md")
		//if _, ok := parts["input,v"]; ok {
		//	testRCSToJSON(t, parts, options)
		//	testCircular(t, parts, options) // rcs to rcs
		//} else if _, ok := parts["input.rcs"]; ok {
		//	parts["input,v"] = parts["input.rcs"] // map old name
		//	testRCSToJSON(t, parts, options)
		//	testCircular(t, parts, options)
		//} else if _, ok := parts["input.json"]; ok {
		//	testJSONToRCS(t, parts, options)
		//}
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(testContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip markdown list markers
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "- ")

		// Split by comma for multiple tests on one line?
		testLine := strings.SplitN(line, ":", 2)
		testName := testLine[0]

		testName = strings.TrimSpace(testName)
		if testName == "" {
			continue
		}

		switch {
		case testName == "json to rcs":
			testJSONToRCS(t, parts, options)
		case testName == "rcs to json":
			testRCSToJSON(t, parts, options)
		case testName == "rcs to rcs":
			testCircular(t, parts, options)
		case testName == "format rcs":
			testFormatRCS(t, parts, options)
		case testName == "validate rcs":
			testValidateRCS(t, parts, options)
		case testName == "new rcs":
			testNewRCS(t, parts, options)
		case testName == "list heads":
			testListHeads(t, parts, options)
		case testName == "normalize revisions":
			testNormalizeRevisions(t, parts, options)
		case strings.HasPrefix(testName, "parse error:"):
			testParseError(t, testName, parts, options)
		case testName == "ci" || testName == "co" || testName == "rcs" || testName == "parse error":
			// Skip currently unsupported test types from operations
		default:
			t.Errorf("Unknown test type: %q", testName)
		}
	}
}

func testJSONToRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("json to rcs", func(t *testing.T) {
		inputJSON, ok := parts["input.json"]
		if !ok {
			t.Fatal("Missing input.json")
		}
		expectedRCS, ok := parts["expected,v"]
		if !ok {
			t.Fatal("Missing expected,v")
		}

		var file File
		if err := json.Unmarshal([]byte(inputJSON), &file); err != nil {
			t.Fatalf("json.Unmarshal error: %v", err)
		}

		gotRCS := file.String()
		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testRCSToJSON(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("rcs to json", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedJSON, ok := parts["expected.json"]
		if !ok {
			t.Fatal("Missing expected.json")
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		if options["unix line endings"] {
			parsedFile.SwitchLineEnding("\n")
		}

		gotJSONBytes, err := json.MarshalIndent(parsedFile, "", "  ")
		if err != nil {
			t.Fatalf("json.MarshalIndent error: %v", err)
		}
		gotJSON := string(gotJSONBytes)

		if options["force unix line endings"] {
			expectedJSON = strings.ReplaceAll(expectedJSON, "\r\n", "\n")
		}

		if diff := cmp.Diff(strings.TrimSpace(expectedJSON), strings.TrimSpace(gotJSON)); diff != "" {
			t.Errorf("JSON mismatch (-want +got):\n%s", diff)
		}
	})
}

func testCircular(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("rcs to rcs", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		if options["unix line endings"] {
			parsedFile.SwitchLineEnding("\n")
		}

		gotRCS := parsedFile.String()

		expectedRCS := inputRCS
		// If expected,v is present, strictly match that, otherwise match input
		if exp, ok := parts["expected,v"]; ok {
			expectedRCS = exp
		}

		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testFormatRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("format rcs", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedRCS, ok := parts["expected,v"]
		if !ok {
			t.Fatal("Missing expected,v")
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		if options["unix line endings"] {
			parsedFile.SwitchLineEnding("\n")
		}

		gotRCS := parsedFile.String()
		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testValidateRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("validate rcs", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)

		_, err := parseRCS(inputRCS)
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
	})
}

func testNewRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("new rcs", func(t *testing.T) {
		expectedRCS, ok := parts["expected,v"]
		if !ok {
			t.Fatal("Missing expected,v")
		}

		f := NewFile()
		if options["unix line endings"] {
			f.NewLine = "\n"
		}
		gotRCS := f.String()
		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testListHeads(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("list heads", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedOut, ok := parts["expected.out"]
		if !ok {
			t.Fatal("Missing expected.out")
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		var sb strings.Builder
		for _, rev := range parsedFile.RevisionHeads {
			sb.WriteString(rev.Revision.String() + "\n")
		}
		gotOut := sb.String()

		if options["force unix line endings"] {
			expectedOut = strings.ReplaceAll(expectedOut, "\r\n", "\n")
		}

		if diff := cmp.Diff(strings.TrimSpace(expectedOut), strings.TrimSpace(gotOut)); diff != "" {
			t.Errorf("List Heads mismatch (-want +got):\n%s", diff)
		}
	})
}

func testNormalizeRevisions(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("normalize revisions", func(t *testing.T) {
		testFormatRCS(t, parts, options)
	})
}

func testParseError(t *testing.T, testName string, parts map[string]string, options map[string]bool) {
	errorName := strings.TrimPrefix(testName, "parse error: ")
	errorName = strings.TrimSpace(errorName)

	t.Run(testName, func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)

		provider, ok := ErrorTestProvider[errorName]
		if !ok {
			t.Fatalf("Unknown error provider: %s", errorName)
		}

		// Prepare expected error
		var expectedErr error
		if errorJSON, ok := parts["error.json"]; ok {
			var err error
			expectedErr, err = provider([]byte(errorJSON))
			if err != nil {
				t.Fatalf("Failed to prepare expected error: %v", err)
			}
		} else {
			expectedErr, _ = provider(nil)
		}

		_, err := parseRCS(inputRCS)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Special handling for ErrParseProperty
		var pErr ErrParseProperty
		if errors.As(expectedErr, &pErr) {
			var actualPErr ErrParseProperty
			if errors.As(err, &actualPErr) {
				if actualPErr.Property != pErr.Property {
					t.Errorf("Property mismatch: want %q, got %q", pErr.Property, actualPErr.Property)
				}
				return
			}
		}

		if !errors.Is(err, expectedErr) {
			// Also check strict equality if Is fails (for structs)
			if err.Error() != expectedErr.Error() {
				t.Errorf("Error mismatch:\nWant: %v\nGot:  %v", expectedErr, err)
			}
		}
	})
}

// Helpers

func getInputRCS(t *testing.T, parts map[string]string) string {
	if content, ok := parts["input,v"]; ok {
		return content
	}
	if content, ok := parts["input.rcs"]; ok {
		return content
	}
	t.Fatal("Missing input,v")
	return ""
}

func parseRCS(content string) (*File, error) {
	// Retry logic from original test
	parsedFile, err := ParseFile(strings.NewReader(content))
	if err != nil {
		parsedFile, err = ParseFile(strings.NewReader(content + "\n\n\n"))
		if err != nil {
			parsedFile, err = ParseFile(strings.NewReader(content + "\n"))
		}
	}
	return parsedFile, err
}

func checkRCS(t *testing.T, expected, got string, options map[string]bool) {
	ignoreWhitespace := options["ignore white space"]

	normExpected := strings.TrimSpace(expected)
	normGot := strings.TrimSpace(got)

	if options["force unix line endings"] {
		normExpected = strings.ReplaceAll(normExpected, "\r\n", "\n")
	}

	// 'got' comes from parsedFile.String().
	// If 'unix line endings' is ON, parsedFile is normalized to \n, so got has \n.
	// If 'unix line endings' is OFF, parsedFile might have \r\n (from input).
	// If input had \r\n and we compare against normalized expected (\n), we need to normalize got too?
	// The user request suggests 'unix line endings' ensures the object has \n.
	// So got should be correct.

	if ignoreWhitespace {
		normExpected = strings.Join(strings.Fields(normExpected), " ")
		normGot = strings.Join(strings.Fields(normGot), " ")
	}

	if diff := cmp.Diff(normExpected, normGot); diff != "" {
		t.Errorf("RCS mismatch (-want +got):\n%s", diff)
		if !ignoreWhitespace {
			t.Logf("Got RCS:\n%q", got)
		}
	}
}
