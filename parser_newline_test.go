package rcs

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseFile_NewLines(t *testing.T) {
	rcsContent := `head     1.1;
branch   1.1.1;
access   ;
symbols  rc1:1.1.1.1 ic:1.1.1;
locks    ; strict;
comment  @# Master with an expand field that's data, not a token.@;
expand   @b@;


1.1
date     2004.09.07.08.24.09;  author cisers;  state Exp;
branches 1.1.1.1;
next     ;

1.1.1.1
date     2004.09.07.08.24.09;  author cisers;  state Exp;
branches ;
next     ;


desc
@@



1.1
log
@da28248b4ec75efbe0ba7461142ed60d
@
text
@projectx/doc/jms-1_0_2b-spec.pdf,v content for 1.1
@




1.1.1.1
log
@03a6ecc1dbc74cfaacfeb8b1f09b8998
@
text
@d1 1
a1 1
projectx/doc/jms-1_0_2b-spec.pdf,v content for 1.1.1.1
@
`

	got, err := ParseFile(strings.NewReader(rcsContent))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if got.Head != "1.1" {
		t.Errorf("Head = %q, want 1.1", got.Head)
	}
	if got.Branch != "1.1.1" {
		t.Errorf("Branch = %q, want 1.1.1", got.Branch)
	}
	if len(got.Locks) != 0 {
		t.Errorf("Locks count = %d, want 0", len(got.Locks))
	}
	if !got.Strict {
		t.Errorf("Strict = %v, want true", got.Strict)
	}
	if got.StrictOnOwnLine {
		t.Errorf("StrictOnOwnLine = %v, want false", got.StrictOnOwnLine)
	}
	if got.Comment != "# Master with an expand field that's data, not a token." {
		t.Errorf("Comment = %q", got.Comment)
	}
	if got.Expand != "b" {
		t.Errorf("Expand = %q, want b", got.Expand)
	}
	if got.Description != "" {
		t.Errorf("Description = %q, want empty", got.Description)
	}

	if len(got.RevisionHeads) != 2 {
		t.Errorf("RevisionHeads count = %d, want 2", len(got.RevisionHeads))
	}

	if len(got.RevisionContents) != 2 {
		t.Errorf("RevisionContents count = %d, want 2", len(got.RevisionContents))
	}

	// Check specific revisions
	for _, rc := range got.RevisionContents {
		if rc.Revision == "1.1" {
			expectedText := "projectx/doc/jms-1_0_2b-spec.pdf,v content for 1.1\n"
			if rc.Text != expectedText {
				t.Errorf("Revision 1.1 text = %q, want %q", rc.Text, expectedText)
			}
			// From desc @@ to 1.1 there are 3 empty lines.
			// @@\n (1)
			// \n (2)
			// \n (3)
			// \n (4)
			// 1.1
			// ParseDescription consumes \n\n (1 and 2).
			// Left 3 and 4. So 2.
			// Wait, in the string literal above:
			// desc
			// @@
			//
			//
			//
			// 1.1
			// It looks like 3 empty lines.
			// PrecedingNewLines is 1.
			// Offset = 2 - PrecedingNewLines = 2 - 1 = 1.
			if rc.RevisionDescriptionNewLineOffset != 1 {
				t.Errorf("Revision 1.1 RevisionDescriptionNewLineOffset = %d, want 1", rc.RevisionDescriptionNewLineOffset)
			}
		} else if rc.Revision == "1.1.1.1" {
			expectedText := "d1 1\na1 1\nprojectx/doc/jms-1_0_2b-spec.pdf,v content for 1.1.1.1\n"
			if rc.Text != expectedText {
				t.Errorf("Revision 1.1.1.1 text = %q, want %q", rc.Text, expectedText)
			}
			// Added extra newlines above 1.1.1.1.
			// Text block ends with @\n.
			// Then \n
			// Then \n
			// Then \n
			// Then \n
			// 1.1.1.1
			// ParseRevisionContent(1.1) consumes \n\n (1 and 2).
			// Left 3 and 4. So 2.
			// Offset = 2 - 2 = 0.
			if rc.RevisionDescriptionNewLineOffset != 0 {
				t.Errorf("Revision 1.1.1.1 RevisionDescriptionNewLineOffset = %d, want 0", rc.RevisionDescriptionNewLineOffset)
			}
		} else {
			t.Errorf("Unexpected revision %s", rc.Revision)
		}
	}

	// Check round trip
	// Note: The original input has varying amounts of whitespace which might not be preserved exactly by String()
	// But we can check if it parses back to the same structure.

	output := got.String()
	got2, err := ParseFile(strings.NewReader(output))
	if err != nil {
		t.Fatalf("ParseFile() round trip error = %v", err)
	}

	if diff := cmp.Diff(got, got2); diff != "" {
		t.Errorf("Round trip diff: %s", diff)
	}
}

func TestParseFile_TooManyNewLines(t *testing.T) {
	rcsContent := `head     1.1;
branch   ;
access   ;
symbols  ;
locks    ; strict;
comment  @# @;


1.1
date     2004.09.07.08.24.09;  author cisers;  state Exp;
branches ;
next     ;


desc
@@







1.1
log
@log@
text
@text@
`
	// Added one more newline to trigger > 4.

	_, err := ParseFile(strings.NewReader(rcsContent))
	if err == nil {
		t.Fatal("ParseFile() expected error for too many newlines, got nil")
	}
	expectedErr := "too many new lines: 5"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("ParseFile() error = %v, want error containing %q", err, expectedErr)
	}
}
