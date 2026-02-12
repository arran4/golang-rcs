package rcs

import (
	"strings"
	"testing"
)

func TestParseIntegrity(t *testing.T) {
	input := `head	1.1;
integrity	@some @@ value@;
comment	@This is a comment@;

desc
@@


`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if f.Integrity != "some @ value" {
		t.Errorf("expected Integrity 'some @ value', got %q", f.Integrity)
	}
}

func TestParseIntegrityUnquoted(t *testing.T) {
	input := `head	1.1;
integrity	simplevalue;
comment	@This is a comment@;

desc
@@


`
	_, err := ParseFile(strings.NewReader(input))
	if err == nil {
		t.Errorf("expected error for unquoted integrity, but got nil")
	} else if !strings.Contains(err.Error(), "looking for \"@\"") {
		// ParseHeaderComment -> ParseAtQuotedString -> ScanStrings("@") -> ScanNotFound -> Error()
		t.Errorf("expected 'looking for \"@\"' error, got %q", err)
	}
}

func TestStringIntegrity(t *testing.T) {
	f := &File{
		Head:      "1.1",
		Integrity: "some @ value",
		Comment:   "This is a comment",
		Description: "",
	}
	s := f.String()
	expected := "integrity\t@some @@ value@;\n"
	if !strings.Contains(s, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, s)
	}
}
