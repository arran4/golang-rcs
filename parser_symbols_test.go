package rcs

import (
	"strings"
	"testing"
)

func TestSymbolOrderPreservation(t *testing.T) {
	input := `head	1.2;
access;
symbols
	4-4-6:1.2
	4-4-5:1.2
	4-4-4:1.2
	4-4-3:1.2
	4-4-2:1.2
	4-4-1:1.2
	4-4-0:1.2;
locks; strict;
comment	@# Test old-fashioned format of tags.@;





1.2
date	2002.10.06.03.23.17;	author esr;	state Exp;
branches;
next	1.1;

1.1
date	2002.10.06.03.23.17;	author esr;	state Exp;
branches;
next	;

desc
@@

1.2
log
@@
text
@@

1.1
log
@@
text
@@
`
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	got := f.String()

	// Check that symbols appear in the original order
	expectedSymbolsOrder := []string{
		"4-4-6:1.2",
		"4-4-5:1.2",
		"4-4-4:1.2",
		"4-4-3:1.2",
		"4-4-2:1.2",
		"4-4-1:1.2",
		"4-4-0:1.2",
	}

	lastIndex := -1
	for _, sym := range expectedSymbolsOrder {
		idx := strings.Index(got, sym)
		if idx == -1 {
			t.Errorf("Symbol %q not found in output", sym)
			continue
		}
		if idx < lastIndex {
			t.Errorf("Symbol %q appeared out of order (at %d, previous at %d)", sym, idx, lastIndex)
		}
		lastIndex = idx
	}

	// Verify that f.Symbols is populated correctly
	if len(f.Symbols) != 7 {
		t.Errorf("Expected 7 symbols in Symbols slice, got %d", len(f.Symbols))
	}

	// Verify backward compatibility method
	sm := f.SymbolMap()
	if len(sm) != 7 {
		t.Errorf("Expected 7 symbols in SymbolMap, got %d", len(sm))
	}
	if sm["4-4-6"] != "1.2" {
		t.Errorf("Expected SymbolMap['4-4-6'] to be '1.2', got %q", sm["4-4-6"])
	}
}
