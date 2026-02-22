package rcs

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFile_String(t *testing.T) {
	// Test file with no locks and no branches
	f := &File{
		Head:    "1.1",
		Comment: "# ",
		Locks:   nil,
		RevisionHeads: []*RevisionHead{
			{
				Revision:     "1.1",
				Date:         "2022.01.01.00.00.00",
				Author:       "user",
				State:        "Exp",
				Branches:     nil,
				NextRevision: "",
			},
		},
		Description: "test desc\n",
	}
	expected := "head\t1.1;\ncomment\t@# @;\n\n\n1.1\ndate\t2022.01.01.00.00.00;\tauthor user;\tstate Exp;\nbranches;\nnext\t;\n\n\ndesc\n@test desc\n@\n\n"
	if got := f.String(); got != expected {
		t.Errorf("File.String() = %q, want %q", got, expected)
	}

	// Test Lock.String
	l := &Lock{User: "u", Revision: "1"}
	if got := l.String(); got != "u:1;" {
		t.Errorf("Lock.String() = %q, want %q", got, "u:1;")
	}

	// Test RevisionHead with branches
	f.RevisionHeads[0].Branches = []Num{"1.1.1.1"}
	if !strings.Contains(f.RevisionHeads[0].String(), "branches\n\t1.1.1.1;") {
		t.Errorf("RevisionHead.String() missing branches: %q", f.RevisionHeads[0].String())
	}
}

func TestRevisionHeadStringBranches(t *testing.T) {
	h := &RevisionHead{
		Revision:     "1.1",
		Date:         "2022.01.01.00.00.00",
		Author:       "test",
		State:        "Exp",
		Branches:     []Num{"1.1.1.1", "1.1.2.1"},
		NextRevision: "",
	}
	want := "1.1\n" +
		"date\t2022.01.01.00.00.00;\tauthor test;\tstate Exp;\n" +
		"branches\n\t1.1.1.1\n\t1.1.2.1;\n" +
		"next\t;\n"
	if diff := cmp.Diff(h.String(), want); diff != "" {
		t.Errorf("RevisionHead.String() diff: %s", diff)
	}
}

func TestStringIntegrity(t *testing.T) {
	f := &File{
		Head:        "1.1",
		Integrity:   "some @ value",
		Comment:     "This is a comment",
		Description: "",
	}
	s := f.String()
	expected := "integrity\t@some @@ value@;\n"
	if !strings.Contains(s, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, s)
	}
}
