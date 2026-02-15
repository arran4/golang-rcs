package rcs

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestPseudoGrammar(t *testing.T) {
	f := &File{}
	grammar := f.PseudoGrammar()

	expected := `File := {
	Head: string;
	Branch: string;
	Description: string;
	Comment: string;
	Access: bool;
	Symbols: {Symbol}*;
	AccessUsers: {string}*;
	Locks: {Lock}*;
	Strict: bool;
	StrictOnOwnLine: bool?;
	DateYearPrefixTruncated: bool?;
	Integrity: string;
	Expand: string;
	NewLine: string;
	EndOfFileNewLineOffset: int?;
	RevisionHeads: {RevisionHead}*;
	RevisionContents: {RevisionContent}*;
};
Lock := {
	User: string;
	Revision: string;
};
NewPhrase := {
	Key: ID;
	Value: {PhraseValue}*;
};
RevisionContent := {
	Revision: string;
	Log: string;
	Text: string;
	PrecedingNewLinesOffset: int?;
};
RevisionHead := {
	Revision: Num;
	Date: DateTime;
	YearTruncated: bool?;
	Author: ID;
	State: ID;
	Branches: {Num}*;
	NextRevision: Num;
	CommitID: Sym;
	Owner: {PhraseValue}*?;
	Group: {PhraseValue}*?;
	Permissions: {PhraseValue}*?;
	Hardlinks: {PhraseValue}*?;
	Deltatype: {PhraseValue}*?;
	Kopt: {PhraseValue}*?;
	Mergepoint: {PhraseValue}*?;
	Filename: {PhraseValue}*?;
	Username: {PhraseValue}*?;
	NewPhrases: {NewPhrase}*?;
};
Symbol := {
	Name: string;
	Revision: string;
};`

	if diff := cmp.Diff(expected, grammar); diff != "" {
		t.Errorf("PseudoGrammar mismatch (-expected +got):\n%s", diff)
	}
}
