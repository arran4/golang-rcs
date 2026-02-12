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
}
