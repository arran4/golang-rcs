package rcs

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/pseudo_grammar.txt
var expectedGrammar string

func TestPseudoGrammar(t *testing.T) {
	f := &File{}
	grammar := f.PseudoGrammar()

	if diff := cmp.Diff(strings.TrimSpace(expectedGrammar), strings.TrimSpace(grammar)); diff != "" {
		t.Errorf("PseudoGrammar mismatch (-expected +got):\n%s", diff)
	}
}
