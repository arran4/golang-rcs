package rcs

import (
	"strings"
	"testing"
)

func TestParseFile_NewFeatures(t *testing.T) {
	// RCS file with branch, strict, expand
	// Note: locks; followed by newline because parser enforces newline after fields
	rcsData := `head	1.1;
branch	1.1.1;
access;
symbols;
locks;
strict;
comment	@# @;
expand	@kv@;


1.1
date	2022.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@test features
@


1.1
log
@Initial revision
@
text
@content
@
`
	s := strings.NewReader(rcsData)
	f, err := ParseFile(s)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if f.Head != "1.1" {
		t.Errorf("Head = %q, want %q", f.Head, "1.1")
	}
	if f.Branch != "1.1.1" {
		t.Errorf("Branch = %q, want %q", f.Branch, "1.1.1")
	}
	if !f.Strict {
		t.Errorf("Strict = %v, want true", f.Strict)
	}
	if f.Expand != "kv" {
		t.Errorf("Expand = %q, want %q", f.Expand, "kv")
	}

	// Test string representation includes new fields
	sRep := f.String()
	if !strings.Contains(sRep, "branch\t1.1.1;") {
		t.Errorf("String() missing branch: %s", sRep)
	}
	if !strings.Contains(sRep, "strict;") {
		t.Errorf("String() missing strict: %s", sRep)
	}
	if !strings.Contains(sRep, "expand\tkv;") {
		t.Errorf("String() missing expand: %s", sRep)
	}
}
