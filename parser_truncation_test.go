package rcs

import (
	"strings"
	"testing"
	"time"
)

func TestParseFile_TruncatedYear(t *testing.T) {
	// Input with 2-digit year "99" (1999)
	input := `head	1.1;
access;
symbols;
locks; strict;
comment	@# @;


1.1
date	99.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@Description
@


1.1
log
@Initial revision
@
text
@Content
@
`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check if the date was parsed correctly as 1999
	expectedDate := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.RevisionHeads[0].Date.Equal(expectedDate) {
		t.Errorf("Date parsed incorrectly: got %v, want %v", f.RevisionHeads[0].Date, expectedDate)
	}

	// This is the check that will fail before implementation
	if !f.DateYearPrefixTruncated {
		t.Errorf("DateYearPrefixTruncated should be true for 2-digit year")
	}

	// Check serialization
	output := f.String()
	// We expect the output to also have "99.01.01..." if we support preserving it.
	// Currently it will likely output "1999.01.01..."
	if !strings.Contains(output, "date\t99.01.01.00.00.00;") {
		t.Errorf("Output should contain truncated date '99.01.01.00.00.00;', got:\n%s", output)
	}
}
