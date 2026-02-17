package rcs

type Formattable interface {
	ResetFormatting()
	SetFormattingOption(op ...any)
}

type WithFileFormattingOptions FileFormattingOptions
type WithRevisionHeadFormattingOptions RevisionHeadFormattingOptions
type WithRevisionContentFormattingOptions RevisionContentFormattingOptions

type WithFileStrictOnOwnLine bool
type WithFileDateYearPrefixTruncated bool
type WithFileEndOfFileNewLineOffset int
type WithFileRevisionStartLineOffset int
type WithFileDescriptionNewLineOffset int
type WithFileSymbolTerminatorPrefix string
type WithFileHeadSeparatorSpaces int
type WithFileAccessSeparatorSpaces int
type WithFileSymbolsSeparatorSpaces int
type WithFileLocksSeparatorSpaces int
type WithFileCommentSeparatorSpaces int

type WithRevisionHeadYearTruncated bool

type WithRevisionContentPrecedingNewLinesOffset int

func (f *File) ResetFormatting() {
	f.FileFormattingOptions = FileFormattingOptions{}
	for _, rh := range f.RevisionHeads {
		rh.ResetFormatting()
	}
	for _, rc := range f.RevisionContents {
		rc.ResetFormatting()
	}
}

func (f *File) SetFormattingOption(ops ...any) {
	var headOps []any
	var contentOps []any
	for _, op := range ops {
		switch v := op.(type) {
		case WithFileFormattingOptions:
			f.FileFormattingOptions = FileFormattingOptions(v)
		case WithFileStrictOnOwnLine:
			f.StrictOnOwnLine = bool(v)
		case WithFileDateYearPrefixTruncated:
			f.DateYearPrefixTruncated = bool(v)
		case WithFileEndOfFileNewLineOffset:
			f.EndOfFileNewLineOffset = int(v)
		case WithFileRevisionStartLineOffset:
			f.RevisionStartLineOffset = int(v)
		case WithFileDescriptionNewLineOffset:
			f.DescriptionNewLineOffset = int(v)
		case WithFileSymbolTerminatorPrefix:
			f.SymbolTerminatorPrefix = string(v)
		case WithFileHeadSeparatorSpaces:
			f.HeadSeparatorSpaces = int(v)
		case WithFileAccessSeparatorSpaces:
			f.AccessSeparatorSpaces = int(v)
		case WithFileSymbolsSeparatorSpaces:
			f.SymbolsSeparatorSpaces = int(v)
		case WithFileLocksSeparatorSpaces:
			f.LocksSeparatorSpaces = int(v)
		case WithFileCommentSeparatorSpaces:
			f.CommentSeparatorSpaces = int(v)
		case WithRevisionHeadFormattingOptions, WithRevisionHeadYearTruncated:
			headOps = append(headOps, op)
		case WithRevisionContentFormattingOptions, WithRevisionContentPrecedingNewLinesOffset:
			contentOps = append(contentOps, op)
		}
	}
	if len(headOps) > 0 {
		for _, rh := range f.RevisionHeads {
			rh.SetFormattingOption(headOps...)
		}
	}
	if len(contentOps) > 0 {
		for _, rc := range f.RevisionContents {
			rc.SetFormattingOption(contentOps...)
		}
	}
}

func (h *RevisionHead) ResetFormatting() {
	h.RevisionHeadFormattingOptions = RevisionHeadFormattingOptions{}
}

func (h *RevisionHead) SetFormattingOption(ops ...any) {
	for _, op := range ops {
		switch v := op.(type) {
		case WithRevisionHeadFormattingOptions:
			h.RevisionHeadFormattingOptions = RevisionHeadFormattingOptions(v)
		case WithRevisionHeadYearTruncated:
			h.YearTruncated = bool(v)
		}
	}
}

func (c *RevisionContent) ResetFormatting() {
	c.RevisionContentFormattingOptions = RevisionContentFormattingOptions{}
}

func (c *RevisionContent) SetFormattingOption(ops ...any) {
	for _, op := range ops {
		switch v := op.(type) {
		case WithRevisionContentFormattingOptions:
			c.RevisionContentFormattingOptions = RevisionContentFormattingOptions(v)
		case WithRevisionContentPrecedingNewLinesOffset:
			c.PrecedingNewLinesOffset = int(v)
		}
	}
}
