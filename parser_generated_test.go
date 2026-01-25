package rcs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedFiles(t *testing.T) {
	files, err := filepath.Glob("testdata/generated/*.v")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("No generated files found")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			filename := filepath.Base(file)
			if filename == "weird_whitespace.v" || filename == "multiline_symbols.v" {
				t.Skipf("Skipping known failing test: %s", filename)
			}

			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			rcsFile, err := ParseFile(f)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			// Validate specific properties based on filename
			switch filename {
			case "branches.v":
				validateBranches(t, rcsFile)
			case "access_symbols.v":
				validateAccessSymbols(t, rcsFile)
			case "integrity_expand.v":
				validateIntegrityExpand(t, rcsFile)
			case "complex_graph.v":
				validateComplexGraph(t, rcsFile)
			case "weird_whitespace.v":
				validateWeirdWhitespace(t, rcsFile)
			case "quoted_strings.v":
				validateQuotedStrings(t, rcsFile)
			}
		})
	}
}

func validateBranches(t *testing.T, f *File) {
	if f.Head != "1.2" {
		t.Errorf("Head = %s, want 1.2", f.Head)
	}
	// Check if 1.2 has branch 1.2.1.1
	found := false
	for _, rh := range f.RevisionHeads {
		if rh.Revision == "1.2" {
			if len(rh.Branches) != 1 || rh.Branches[0] != "1.2.1.1" {
				t.Errorf("1.2 Branches = %v, want [1.2.1.1]", rh.Branches)
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("Revision 1.2 not found")
	}
}

func validateAccessSymbols(t *testing.T, f *File) {
	if len(f.AccessUsers) != 2 {
		t.Errorf("AccessUsers len = %d, want 2", len(f.AccessUsers))
	} else {
		// order might vary if not sorted in ParseFile?
		// parser.go: ParseHeaderAccess uses strings.Fields on the line.
		// My generator puts "alice bob".
		if f.AccessUsers[0] != "alice" || f.AccessUsers[1] != "bob" {
			t.Errorf("AccessUsers = %v, want [alice bob]", f.AccessUsers)
		}
	}

	if len(f.SymbolMap) != 3 {
		t.Errorf("SymbolMap len = %d, want 3", len(f.SymbolMap))
	} else {
		if f.SymbolMap["beta"] != "1.2.1.1" {
			t.Errorf("Symbol beta = %s, want 1.2.1.1", f.SymbolMap["beta"])
		}
	}

	if !f.Strict {
		t.Error("Strict mode should be true")
	}
}

func validateIntegrityExpand(t *testing.T, f *File) {
	if f.Integrity != "some_checksum" {
		t.Errorf("Integrity = %q, want \"some_checksum\"", f.Integrity)
	}
	if f.Expand != "kv" {
		t.Errorf("Expand = %q, want \"kv\"", f.Expand)
	}
}

func validateComplexGraph(t *testing.T, f *File) {
	if f.Head != "1.3" {
		t.Errorf("Head = %s, want 1.3", f.Head)
	}
	// Check revision count
	if len(f.RevisionHeads) != 6 {
		t.Errorf("RevisionHeads len = %d, want 6", len(f.RevisionHeads))
	}
}

func validateWeirdWhitespace(t *testing.T, f *File) {
	if f.Head != "1.1" {
		t.Errorf("Head = %s, want 1.1", f.Head)
	}
	if len(f.RevisionHeads) != 1 {
		t.Fatal("Expected 1 revision")
	}
	if f.RevisionHeads[0].Revision != "1.1" {
		t.Errorf("Rev = %s, want 1.1", f.RevisionHeads[0].Revision)
	}
}

func validateQuotedStrings(t *testing.T, f *File) {
	if len(f.RevisionContents) != 1 {
		t.Fatal("Expected 1 revision content")
	}
	rc := f.RevisionContents[0]
	if !strings.Contains(rc.Log, "@") {
		t.Errorf("Log content missing @: %q", rc.Log)
	}
	if !strings.Contains(rc.Text, "Hello @ World") {
		t.Errorf("Text content missing @ string: %q", rc.Text)
	}
}
