package rcs

import (
	"strings"
	"testing"
)

func TestSymbolOrder(t *testing.T) {
	input := `head	1.1;
access;
symbols
	B:1.2
	A:1.1;
locks; strict;
comment	@# @;


1.1
date	2022.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@test
@


1.1
log
@Initial revision
@
text
@content
@
`

	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	// Verify order in slice
	if len(f.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(f.Symbols))
	}
	if f.Symbols[0].Name != "B" {
		t.Errorf("Expected first symbol to be B, got %s", f.Symbols[0].Name)
	}
	if f.Symbols[1].Name != "A" {
		t.Errorf("Expected second symbol to be A, got %s", f.Symbols[1].Name)
	}

	// Verify order in String() output
	output := f.String()
	// Check that B comes before A in output
	idxB := strings.Index(output, "B:1.2")
	idxA := strings.Index(output, "A:1.1")
	if idxB == -1 {
		t.Error("Output missing B:1.2")
	}
	if idxA == -1 {
		t.Error("Output missing A:1.1")
	}
	if idxB > idxA {
		t.Errorf("Expected B to appear before A in output, but got B at %d and A at %d", idxB, idxA)
	}

	// Mutation test
	f.Symbols = append(f.Symbols, Symbol{Name: "C", Revision: "1.3"})
	output = f.String()
	idxC := strings.Index(output, "C:1.3")
	if idxC == -1 {
		t.Error("Output missing C:1.3")
	}
	// Verify C is after A
	idxA = strings.Index(output, "A:1.1") // Re-index as string changed
	if idxA > idxC {
		t.Errorf("Expected A to appear before C in output, but got A at %d and C at %d", idxA, idxC)
	}
}
