package rcs

import (
	"strings"
	"testing"
)

func TestNewFile(t *testing.T) {
	f := NewFile()
	if f.Symbols == nil {
		t.Error("NewFile().Symbols should not be nil")
	}
	if len(f.Symbols) != 0 {
		t.Error("NewFile().Symbols should be empty")
	}
	if f.Locks == nil {
		t.Error("NewFile().Locks should not be nil")
	}
	if len(f.Locks) != 0 {
		t.Error("NewFile().Locks should be empty")
	}
}

func TestParseFile_NilFields(t *testing.T) {
	input := `head 1.1;
comment @c@;


1.1
date 2022.01.01.00.00.00; author a; state s;
branches;
next ;


desc
@@


1.1
log
@@
text
@@
`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if f.Symbols != nil {
		t.Error("ParseFile should produce nil Symbols if missing")
	}
	if f.Locks != nil {
		t.Error("ParseFile should produce nil Locks if missing")
	}

	s := f.String()
	if strings.Contains(s, "symbols") {
		t.Error("String() should not output symbols if nil")
	}
	if strings.Contains(s, "locks") {
		t.Error("String() should not output locks if nil")
	}
}

func TestParseFile_EmptyFields(t *testing.T) {
	input := `head 1.1;
symbols;
locks;
comment @c@;


1.1
date 2022.01.01.00.00.00; author a; state s;
branches;
next ;


desc
@@


1.1
log
@@
text
@@
`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if f.Symbols == nil {
		t.Error("ParseFile should produce non-nil Symbols if 'symbols;' present")
	}
	if len(f.Symbols) != 0 {
		t.Errorf("Symbols should be empty, got %d", len(f.Symbols))
	}
	if f.Locks == nil {
		t.Error("ParseFile should produce non-nil Locks if 'locks;' present")
	}
	if len(f.Locks) != 0 {
		t.Errorf("Locks should be empty, got %d", len(f.Locks))
	}

	s := f.String()
	if !strings.Contains(s, "symbols;\n") {
		t.Error("String() should output 'symbols;' if empty")
	}
	if !strings.Contains(s, "locks;") {
		t.Error("String() should output 'locks;' if empty")
	}
}

func TestFile_String_Output(t *testing.T) {
	f := NewFile()
	f.Head = "1.1"
	f.Comment = "c"
	f.Description = "d"
	// Symbols and Locks are empty non-nil

	s := f.String()
	if !strings.Contains(s, "symbols;\n") {
		t.Error("String() should output 'symbols;' for empty NewFile")
	}
	if !strings.Contains(s, "locks;") {
		t.Error("String() should output 'locks;' for empty NewFile")
	}
}
