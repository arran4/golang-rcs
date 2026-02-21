package rcs

import (
	"reflect"
	"testing"
)

func TestFile_SetFormattingOption(t *testing.T) {
	// Setup
	f := &File{
		RevisionHeads: []*RevisionHead{
			{},
			{},
		},
		RevisionContents: []*RevisionContent{
			{},
		},
	}

	// Test Setting Individual File Options
	f.SetFormattingOption(
		WithFileStrictOnOwnLine(true),
		WithFileDateYearPrefixTruncated(true),
		WithFileEndOfFileNewLineOffset(10),
		WithFileRevisionStartLineOffset(5),
		WithFileDescriptionNewLineOffset(2),
		WithFileSymbolTerminatorPrefix("\n"),
		WithFileHeadSeparatorSpaces(1),
		WithFileAccessSeparatorSpaces(2),
		WithFileSymbolsSeparatorSpaces(3),
		WithFileLocksSeparatorSpaces(4),
		WithFileCommentSeparatorSpaces(5),
	)

	// Verify File Options
	if !f.StrictOnOwnLine {
		t.Errorf("StrictOnOwnLine expected true, got %v", f.StrictOnOwnLine)
	}
	if !f.DateYearPrefixTruncated {
		t.Errorf("DateYearPrefixTruncated expected true, got %v", f.DateYearPrefixTruncated)
	}
	if f.EndOfFileNewLineOffset != 10 {
		t.Errorf("EndOfFileNewLineOffset expected 10, got %d", f.EndOfFileNewLineOffset)
	}
	if f.RevisionStartLineOffset != 5 {
		t.Errorf("RevisionStartLineOffset expected 5, got %d", f.RevisionStartLineOffset)
	}
	if f.DescriptionNewLineOffset != 2 {
		t.Errorf("DescriptionNewLineOffset expected 2, got %d", f.DescriptionNewLineOffset)
	}
	if f.SymbolTerminatorPrefix != "\n" {
		t.Errorf("SymbolTerminatorPrefix expected \"\\n\", got %q", f.SymbolTerminatorPrefix)
	}
	if f.HeadSeparatorSpaces != 1 {
		t.Errorf("HeadSeparatorSpaces expected 1, got %d", f.HeadSeparatorSpaces)
	}
	if f.AccessSeparatorSpaces != 2 {
		t.Errorf("AccessSeparatorSpaces expected 2, got %d", f.AccessSeparatorSpaces)
	}
	if f.SymbolsSeparatorSpaces != 3 {
		t.Errorf("SymbolsSeparatorSpaces expected 3, got %d", f.SymbolsSeparatorSpaces)
	}
	if f.LocksSeparatorSpaces != 4 {
		t.Errorf("LocksSeparatorSpaces expected 4, got %d", f.LocksSeparatorSpaces)
	}
	if f.CommentSeparatorSpaces != 5 {
		t.Errorf("CommentSeparatorSpaces expected 5, got %d", f.CommentSeparatorSpaces)
	}
}

func TestFile_SetFormattingOption_Propagate(t *testing.T) {
	// Setup
	rh1 := &RevisionHead{}
	rh2 := &RevisionHead{}
	rc1 := &RevisionContent{}
	rc2 := &RevisionContent{}
	f := &File{
		RevisionHeads:    []*RevisionHead{rh1, rh2},
		RevisionContents: []*RevisionContent{rc1, rc2},
	}

	// Test Propagating Revision Head Options
	f.SetFormattingOption(WithRevisionHeadYearTruncated(true))

	for i, rh := range f.RevisionHeads {
		if !rh.YearTruncated {
			t.Errorf("RevisionHeads[%d] YearTruncated expected true, got %v", i, rh.YearTruncated)
		}
	}

	// Test Propagating Revision Content Options
	f.SetFormattingOption(WithRevisionContentPrecedingNewLinesOffset(3))

	for i, rc := range f.RevisionContents {
		if rc.PrecedingNewLinesOffset != 3 {
			t.Errorf("RevisionContents[%d] PrecedingNewLinesOffset expected 3, got %d", i, rc.PrecedingNewLinesOffset)
		}
	}
}

func TestFile_SetFormattingOption_Structs(t *testing.T) {
	// Setup
	rh := &RevisionHead{}
	rc := &RevisionContent{}
	f := &File{
		RevisionHeads:    []*RevisionHead{rh},
		RevisionContents: []*RevisionContent{rc},
	}

	// File Formatting Options
	fileOpts := FileFormattingOptions{
		StrictOnOwnLine:          true,
		DateYearPrefixTruncated:  true,
		EndOfFileNewLineOffset:   7,
		RevisionStartLineOffset:  8,
		DescriptionNewLineOffset: 9,
		SymbolTerminatorPrefix:   "\t",
		HeadSeparatorSpaces:      1,
		AccessSeparatorSpaces:    2,
		SymbolsSeparatorSpaces:   3,
		LocksSeparatorSpaces:     4,
		CommentSeparatorSpaces:   5,
	}
	f.SetFormattingOption(WithFileFormattingOptions(fileOpts))

	if !reflect.DeepEqual(f.FileFormattingOptions, fileOpts) {
		t.Errorf("FileFormattingOptions mismatch.\nGot: %+v\nWant: %+v", f.FileFormattingOptions, fileOpts)
	}

	// Revision Head Formatting Options
	headOpts := RevisionHeadFormattingOptions{
		YearTruncated:            true,
		DateSeparatorSpaces:      1,
		DateAuthorSpacingSpaces:  2,
		AuthorStateSpacingSpaces: 3,
		BranchesSeparatorSpaces:  4,
		NextSeparatorSpaces:      5,
	}
	f.SetFormattingOption(WithRevisionHeadFormattingOptions(headOpts))

	if !reflect.DeepEqual(rh.RevisionHeadFormattingOptions, headOpts) {
		t.Errorf("RevisionHeadFormattingOptions mismatch.\nGot: %+v\nWant: %+v", rh.RevisionHeadFormattingOptions, headOpts)
	}

	// Revision Content Formatting Options
	contentOpts := RevisionContentFormattingOptions{
		PrecedingNewLinesOffset: 6,
	}
	f.SetFormattingOption(WithRevisionContentFormattingOptions(contentOpts))

	if !reflect.DeepEqual(rc.RevisionContentFormattingOptions, contentOpts) {
		t.Errorf("RevisionContentFormattingOptions mismatch.\nGot: %+v\nWant: %+v", rc.RevisionContentFormattingOptions, contentOpts)
	}
}

func TestFile_ResetFormatting(t *testing.T) {
	// Setup with formatted options
	rh := &RevisionHead{
		RevisionHeadFormattingOptions: RevisionHeadFormattingOptions{
			YearTruncated: true,
		},
	}
	rc := &RevisionContent{
		RevisionContentFormattingOptions: RevisionContentFormattingOptions{
			PrecedingNewLinesOffset: 5,
		},
	}
	f := &File{
		FileFormattingOptions: FileFormattingOptions{
			StrictOnOwnLine: true,
		},
		RevisionHeads:    []*RevisionHead{rh},
		RevisionContents: []*RevisionContent{rc},
	}

	// Reset
	f.ResetFormatting()

	// Verify Reset on File
	emptyFileOpts := FileFormattingOptions{}
	if !reflect.DeepEqual(f.FileFormattingOptions, emptyFileOpts) {
		t.Errorf("FileFormattingOptions not reset.\nGot: %+v", f.FileFormattingOptions)
	}

	// Verify Reset on Heads
	emptyHeadOpts := RevisionHeadFormattingOptions{}
	if !reflect.DeepEqual(rh.RevisionHeadFormattingOptions, emptyHeadOpts) {
		t.Errorf("RevisionHeadFormattingOptions not reset.\nGot: %+v", rh.RevisionHeadFormattingOptions)
	}

	// Verify Reset on Contents
	emptyContentOpts := RevisionContentFormattingOptions{}
	if !reflect.DeepEqual(rc.RevisionContentFormattingOptions, emptyContentOpts) {
		t.Errorf("RevisionContentFormattingOptions not reset.\nGot: %+v", rc.RevisionContentFormattingOptions)
	}
}
