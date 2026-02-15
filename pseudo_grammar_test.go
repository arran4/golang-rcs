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

	expected := strings.ReplaceAll(expectedGrammar, "\r\n", "\n")
	expected = strings.TrimSpace(expected)
	grammar = strings.TrimSpace(grammar)

	if diff := cmp.Diff(expected, grammar); diff != "" {
		t.Errorf("PseudoGrammar mismatch (-expected +got):\n%s", diff)
	}
}
