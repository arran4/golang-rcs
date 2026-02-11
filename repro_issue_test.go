package rcs

import (
	"strings"
	"testing"
)

func TestParseFileRepro(t *testing.T) {
	input := "head\t1.6;\naccess;\nsymbols;\nlocks; strict;\ncomment\t@# @;\n\n\n1.6\ndate\t2022.01.01.00.00.00;\tauthor user;\tstate Exp;\nbranches;\nnext\t;\n\n\ndesc\n@@\n\n\n1.6\nlog\n@initial@\ntext\n@@\n"
	_, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
}
