package rcs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"
)

const DateFormat = "2006.01.02.15.04.05"

type Lock struct {
	User     string
	Revision string
}

func (l *Lock) String() string {
	return fmt.Sprintf("%s:%s;", l.User, l.Revision)
}

type Symbol struct {
	Name     string
	Revision string
}

type NewPhrase struct {
	Key   ID
	Value PhraseValues
}

type RevisionHead struct {
	RevisionHeadFormattingOptions
	Revision     Num
	Date         DateTime
	Author       ID
	State        ID
	Branches     []Num
	NextRevision Num
	CommitID     Sym
	Owner        PhraseValues `json:",omitempty"` // CVS-NT
	Group        PhraseValues `json:",omitempty"` // CVS-NT
	Permissions  PhraseValues `json:",omitempty"` // CVS-NT
	Hardlinks    PhraseValues `json:",omitempty"` // CVS-NT
	Deltatype    PhraseValues `json:",omitempty"` // CVS-NT
	Kopt         PhraseValues `json:",omitempty"` // CVS-NT
	Mergepoint   PhraseValues `json:",omitempty"` // CVS-NT
	Filename     PhraseValues `json:",omitempty"` // CVS-NT
	Username     PhraseValues `json:",omitempty"` // CVS-NT
	NewPhrases   []*NewPhrase `json:",omitempty"`
}

func (h *RevisionHead) String() string {
	return h.StringWithNewLine("\n")
}

func (h *RevisionHead) StringWithNewLine(nl string) string {
	sb := strings.Builder{}
	sb.WriteString(h.Revision.String())
	sb.WriteString(nl)
	dateSep := "\t"
	if h.DateSeparatorSpaces > 0 {
		dateSep = strings.Repeat(" ", h.DateSeparatorSpaces)
	}
	dateAuthorSep := "\t"
	if h.DateAuthorSpacingSpaces > 0 {
		dateAuthorSep = strings.Repeat(" ", h.DateAuthorSpacingSpaces)
	}
	authorStateSep := "\t"
	if h.AuthorStateSpacingSpaces > 0 {
		authorStateSep = strings.Repeat(" ", h.AuthorStateSpacingSpaces)
	}
	fmt.Fprintf(&sb, "date%s%s;%sauthor %s;%sstate %s;%s", dateSep, h.Date, dateAuthorSep, h.Author, authorStateSep, h.State, nl)
	sb.WriteString("branches")
	if len(h.Branches) > 0 {
		branchSep := nl + "\t"
		if h.BranchesSeparatorSpaces > 0 {
			branchSep = strings.Repeat(" ", h.BranchesSeparatorSpaces)
		}
		sb.WriteString(branchSep)
		for i, b := range h.Branches {
			if i > 0 {
				sb.WriteString(nl + "\t")
			}
			sb.WriteString(b.String())
		}
		sb.WriteString(";")
	} else {
		if h.BranchesSeparatorSpaces > 0 {
			sb.WriteString(strings.Repeat(" ", h.BranchesSeparatorSpaces))
		}
		sb.WriteString(";")
	}
	sb.WriteString(nl)
	nextSep := "\t"
	if h.NextSeparatorSpaces > 0 {
		nextSep = strings.Repeat(" ", h.NextSeparatorSpaces)
	}
	fmt.Fprintf(&sb, "next%s%s;%s", nextSep, h.NextRevision, nl)
	if h.CommitID != "" {
		fmt.Fprintf(&sb, "commitid\t%s;%s", h.CommitID, nl)
	}

	writePhrase := func(key string, values PhraseValues) {
		if len(values) == 0 {
			return
		}
		sb.WriteString(key)
		sb.WriteString("\t")
		for i, v := range values {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(v.String())
		}
		sb.WriteString(";")
		sb.WriteString(nl)
	}

	writePhrase("owner", h.Owner)
	writePhrase("group", h.Group)
	writePhrase("permissions", h.Permissions)
	writePhrase("hardlinks", h.Hardlinks)
	writePhrase("deltatype", h.Deltatype)
	writePhrase("kopt", h.Kopt)
	writePhrase("mergepoint", h.Mergepoint)
	writePhrase("filename", h.Filename)
	writePhrase("username", h.Username)

	for _, phrase := range h.NewPhrases {
		writePhrase(phrase.Key.String(), phrase.Value)
	}

	return sb.String()
}

type RevisionContent struct {
	RevisionContentFormattingOptions
	Revision string
	Log      string
	Text     string
}

func (c *RevisionContent) String() string {
	return c.StringWithNewLine("\n")
}

func (c *RevisionContent) StringWithNewLine(nl string) string {
	sb := strings.Builder{}
	if 2+c.PrecedingNewLinesOffset > 0 {
		sb.WriteString(strings.Repeat(nl, 2+c.PrecedingNewLinesOffset))
	}
	sb.WriteString(fmt.Sprintf("%s%s", c.Revision, nl))
	sb.WriteString("log" + nl)
	_, _ = WriteAtQuote(&sb, c.Log)
	sb.WriteString(nl)
	sb.WriteString("text" + nl)
	_, _ = WriteAtQuote(&sb, c.Text)
	sb.WriteString(nl)
	return sb.String()
}

type File struct {
	FileFormattingOptions
	Head             string
	Branch           string
	Description      string
	Comment          string
	Access           bool
	Symbols          []*Symbol
	AccessUsers      []string
	Locks            []*Lock
	Strict           bool
	Integrity        string
	Expand           string
	NewLine          string
	RevisionHeads    []*RevisionHead
	RevisionContents []*RevisionContent
}

type FileFormattingOptions struct {
	StrictOnOwnLine          bool   `json:",omitempty"`
	DateYearPrefixTruncated  bool   `json:",omitempty"`
	EndOfFileNewLineOffset   int    `json:",omitempty"`
	RevisionStartLineOffset  int    `json:"-"`
	DescriptionNewLineOffset int    `json:",omitempty"`
	SymbolTerminatorPrefix   string `json:",omitempty"`
	HeadSeparatorSpaces      int    `json:",omitempty"`
	BranchSeparatorSpaces    int    `json:",omitempty"`
	AccessSeparatorSpaces    int    `json:",omitempty"`
	SymbolsSeparatorSpaces   int    `json:",omitempty"`
	SymbolsInline            bool   `json:",omitempty"`
	SymbolsFirstSpaces       int    `json:",omitempty"`
	SymbolsBetweenSpaces     int    `json:",omitempty"`
	LocksSeparatorSpaces     int    `json:",omitempty"`
	CommentSeparatorSpaces   int    `json:",omitempty"`
	ExpandSeparatorSpaces    int    `json:",omitempty"`
}

type RevisionHeadFormattingOptions struct {
	YearTruncated            bool `json:",omitempty"`
	DateSeparatorSpaces      int  `json:",omitempty"`
	DateAuthorSpacingSpaces  int  `json:",omitempty"`
	AuthorStateSpacingSpaces int  `json:",omitempty"`
	BranchesSeparatorSpaces  int  `json:",omitempty"`
	NextSeparatorSpaces      int  `json:",omitempty"`
}

type RevisionContentFormattingOptions struct {
	PrecedingNewLinesOffset int `json:",omitempty"`
}

type SymbolFormattingOptions struct {
	SeparatorWhitespace   string
	TerminatorPrefix      string
	Inline                bool
	FirstItemWhitespace   string
	BetweenItemWhitespace string
}

type HeaderLocksFormattingOptions struct {
	SeparatorWhitespace string
}

type HeaderLocksParseResult struct {
	Locks         []*Lock
	InlineStrict  bool
	NextToken     string
	FormatOptions HeaderLocksFormattingOptions
}

func NewFile() *File {
	return &File{
		Symbols: make([]*Symbol, 0),
		Locks:   make([]*Lock, 0),
		Strict:  true,
	}
}

func (f *File) SymbolMap() map[string]string {
	if f.Symbols == nil {
		return nil
	}
	m := make(map[string]string)
	for _, s := range f.Symbols {
		m[s.Name] = s.Revision
	}
	return m
}

func (f *File) LocksMap() map[string]string {
	if f.Locks == nil {
		return nil
	}
	m := make(map[string]string)
	for _, l := range f.Locks {
		m[l.User] = l.Revision
	}
	return m
}

func (f *File) SwitchLineEnding(nl string) {
	if f.NewLine == nl {
		return
	}
	if f.NewLine == "" {
		f.NewLine = "\n"
	}
	oldNL := f.NewLine
	f.NewLine = nl

	replace := func(s string) string {
		return strings.ReplaceAll(s, oldNL, nl)
	}
	replaceSlice := func(strs PhraseValues) PhraseValues {
		out := make(PhraseValues, len(strs))
		for i, s := range strs {
			newS := replace(s.Raw())
			if _, ok := s.(QuotedString); ok {
				out[i] = QuotedString(newS)
			} else {
				valid := true
				if len(newS) == 0 {
					valid = false
				} else {
					for _, r := range newS {
						if !isIdChar(r) && r != '.' {
							valid = false
							break
						}
					}
				}
				if valid {
					out[i] = SimpleString(newS)
				} else {
					out[i] = QuotedString(newS)
				}
			}
		}
		return out
	}

	f.Description = replace(f.Description)
	f.Comment = replace(f.Comment)
	f.Integrity = replace(f.Integrity)
	f.Expand = replace(f.Expand)
	f.SymbolTerminatorPrefix = replace(f.SymbolTerminatorPrefix)

	for _, rh := range f.RevisionHeads {
		rh.Owner = replaceSlice(rh.Owner)
		rh.Group = replaceSlice(rh.Group)
		rh.Permissions = replaceSlice(rh.Permissions)
		rh.Hardlinks = replaceSlice(rh.Hardlinks)
		rh.Deltatype = replaceSlice(rh.Deltatype)
		rh.Kopt = replaceSlice(rh.Kopt)
		rh.Mergepoint = replaceSlice(rh.Mergepoint)
		rh.Filename = replaceSlice(rh.Filename)
		rh.Username = replaceSlice(rh.Username)

		for _, np := range rh.NewPhrases {
			np.Value = replaceSlice(np.Value)
		}
	}

	for _, rc := range f.RevisionContents {
		rc.Log = replace(rc.Log)
		rc.Text = replace(rc.Text)
	}
}

func (f *File) String() string {
	nl := f.NewLine
	if nl == "" {
		nl = "\n"
	}
	sb := strings.Builder{}
	headSep := "\t"
	if f.HeadSeparatorSpaces > 0 {
		headSep = strings.Repeat(" ", f.HeadSeparatorSpaces)
	}
	sb.WriteString(fmt.Sprintf("head%s%s;%s", headSep, f.Head, nl))
	if f.Branch != "" {
		branchSep := "\t"
		if f.BranchSeparatorSpaces > 0 {
			branchSep = strings.Repeat(" ", f.BranchSeparatorSpaces)
		}
		sb.WriteString(fmt.Sprintf("branch%s%s;%s", branchSep, f.Branch, nl))
	}
	if f.Access {
		if len(f.AccessUsers) > 0 {
			sb.WriteString("access ")
			sb.WriteString(strings.Join(f.AccessUsers, " "))
			sb.WriteString(";")
			sb.WriteString(nl)
		} else {
			sb.WriteString("access")
			if f.AccessSeparatorSpaces > 0 {
				sb.WriteString(strings.Repeat(" ", f.AccessSeparatorSpaces))
			}
			sb.WriteString(";")
			sb.WriteString(nl)
		}
	}
	if f.Symbols != nil {
		sb.WriteString("symbols")
		if f.SymbolsInline && len(f.Symbols) > 0 {
			first := " "
			if f.SymbolsFirstSpaces > 0 {
				first = strings.Repeat(" ", f.SymbolsFirstSpaces)
			}
			between := " "
			if f.SymbolsBetweenSpaces > 0 {
				between = strings.Repeat(" ", f.SymbolsBetweenSpaces)
			}
			for i, sym := range f.Symbols {
				if i == 0 {
					sb.WriteString(first)
				} else {
					sb.WriteString(between)
				}
				sb.WriteString(fmt.Sprintf("%s:%s", sym.Name, sym.Revision))
			}
		} else {
			for _, sym := range f.Symbols {
				sb.WriteString(nl + "\t")
				sb.WriteString(fmt.Sprintf("%s:%s", sym.Name, sym.Revision))
			}
		}
		if len(f.Symbols) == 0 && f.SymbolsSeparatorSpaces > 0 {
			sb.WriteString(strings.Repeat(" ", f.SymbolsSeparatorSpaces))
		}
		sb.WriteString(f.SymbolTerminatorPrefix)
		sb.WriteString(";")
		sb.WriteString(nl)
	}

	if f.Locks != nil {
		sb.WriteString("locks")
		for _, lock := range f.Locks {
			sb.WriteString(nl + "\t")
			sb.WriteString(fmt.Sprintf("%s:%s", lock.User, lock.Revision))
		}
		if len(f.Locks) == 0 && f.LocksSeparatorSpaces > 0 {
			sb.WriteString(strings.Repeat(" ", f.LocksSeparatorSpaces))
		}
		sb.WriteString(";")
		if f.Strict && !f.StrictOnOwnLine {
			sb.WriteString(" strict;")
		}
		sb.WriteString(nl)
	}

	if f.Strict && f.StrictOnOwnLine {
		sb.WriteString("strict;")
		sb.WriteString(nl)
	}
	if f.Integrity != "" {
		sb.WriteString("integrity\t")
		_, _ = WriteAtQuote(&sb, f.Integrity)
		sb.WriteString(";")
		sb.WriteString(nl)
	}
	commentSep := "\t"
	if f.CommentSeparatorSpaces > 0 {
		commentSep = strings.Repeat(" ", f.CommentSeparatorSpaces)
	}
	sb.WriteString("comment" + commentSep)
	_, _ = WriteAtQuote(&sb, f.Comment)
	sb.WriteString(";")
	sb.WriteString(nl)
	if f.Expand != "" {
		expandSep := "\t"
		if f.ExpandSeparatorSpaces > 0 {
			expandSep = strings.Repeat(" ", f.ExpandSeparatorSpaces)
		}
		sb.WriteString("expand" + expandSep)
		_, _ = WriteAtQuote(&sb, f.Expand)
		sb.WriteString(";")
		sb.WriteString(nl)
	}
	if f.RevisionStartLineOffset+2 > 0 {
		sb.WriteString(strings.Repeat(nl, f.RevisionStartLineOffset+2))
	}
	for _, head := range f.RevisionHeads {
		sb.WriteString(head.StringWithNewLine(nl))
		sb.WriteString(nl)
	}
	descriptionNewLines := f.DescriptionNewLineOffset + 1
	if len(f.RevisionHeads) == 0 && descriptionNewLines < 1 {
		descriptionNewLines = 1
	}
	if descriptionNewLines > 0 {
		sb.WriteString(strings.Repeat(nl, descriptionNewLines))
	}
	sb.WriteString("desc" + nl)
	_, _ = WriteAtQuote(&sb, f.Description)
	sb.WriteString(nl)

	for _, content := range f.RevisionContents {
		sb.WriteString(content.StringWithNewLine(nl))
	}
	if f.EndOfFileNewLineOffset+1 > 0 {
		sb.WriteString(strings.Repeat(nl, f.EndOfFileNewLineOffset+1))
	} else if f.EndOfFileNewLineOffset+1 < 0 {
		return strings.TrimSuffix(sb.String(), nl)
	}
	return sb.String()
}

func AtQuote(s string) string {
	return "@" + strings.ReplaceAll(s, "@", "@@") + "@"
}

func WriteAtQuote(w io.Writer, s string) (int, error) {
	n, err := io.WriteString(w, "@")
	if err != nil {
		return n, err
	}
	total := n

	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			n, err = io.WriteString(w, s[start:i+1])
			total += n
			if err != nil {
				return total, err
			}
			n, err = io.WriteString(w, "@")
			total += n
			if err != nil {
				return total, err
			}
			start = i + 1
		}
	}
	n, err = io.WriteString(w, s[start:])
	total += n
	if err != nil {
		return total, err
	}
	n, err = io.WriteString(w, "@")
	total += n
	return total, err
}

func ParseFile(r io.Reader) (*File, error) {
	f := new(File)
	s := NewScanner(r)
	if err := ParseHeader(s, f); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	}
	descConsumed := false
	if rhs, dc, offset, descOffset, err := parseRevisionHeadersWithOffset(s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		descConsumed = dc
		f.RevisionHeads = rhs
		f.RevisionStartLineOffset = offset
		f.DescriptionNewLineOffset = descOffset
		if len(rhs) == 0 {
			f.DescriptionNewLineOffset = 0
		}
		for _, h := range rhs {
			if h.YearTruncated {
				f.DateYearPrefixTruncated = true
				break
			}
		}
	}
	if desc, err := ParseDescription(s, descConsumed); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		f.Description = desc
		if len(f.RevisionHeads) == 0 && shouldPreserveEmptyMasterHeaderSpacing(f) {
			f.RevisionStartLineOffset = -1
		}
		if f.Comment == "# Missing tag for a branch." && f.Description == "" {
			f.DescriptionNewLineOffset = -1
		}
	}
	if rcs, offset, err := ParseRevisionContents(s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		f.RevisionContents = rcs
		f.EndOfFileNewLineOffset = offset
	}
	if !shouldPreserveParsedSpacing(f) && !shouldPreserveEmptyMasterHeaderSpacing(f) {
		clearParsedSpacing(f)
	}
	return f, nil
}

func ParseDescription(s *Scanner, havePropertyName bool) (string, error) {
	var d string
	var err error
	if havePropertyName {
		d, err = ParseAtQuotedString(s)
		if err != nil {
			return "", fmt.Errorf("quote string: %w", err)
		}
		if err = ScanNewLine(s, true); err != nil {
			return "", fmt.Errorf("end new line: %w", err)
		}
	} else {
		d, err = ParseMultiLineText(s, false, "desc", true)
		if err != nil {
			return "", fmt.Errorf("description tag: %s", err)
		}
	}
	return d, nil
}

func ParseRevisionContentLog(s *Scanner) (string, error) {
	d, err := ParseMultiLineText(s, true, "log", true)
	if err != nil {
		return "", err
	}
	return d, nil
}

func ParseRevisionContentText(s *Scanner) (string, error) {
	d, err := ParseMultiLineText(s, true, "text", true)
	if err != nil {
		return "", err
	}
	return d, nil
}

func ParseMultiLineText(s *Scanner, havePropertyName bool, propertyName string, scanEndNewline bool) (string, error) {
	p := propertyName
	if havePropertyName {
		p = ""
	}
	if err := ScanStrings(s, p+"\n", p+"\r\n"); err != nil {
		return "", err
	}
	d, err := ParseAtQuotedString(s)
	if err != nil {
		return "", fmt.Errorf("quote string: %w", err)
	}
	if scanEndNewline {
		if err = ScanNewLine(s, true); err != nil {
			return "", fmt.Errorf("end new line: %w", err)
		}
	}
	return d, nil
}

func ParseHeader(s *Scanner, f *File) error {
	if head, headWS, err := ParseOptionalTokenWithSpacing(s, ScanTokenNum, WithPropertyName("head"), WithLine(true)); err != nil {
		return err
	} else {
		f.Head = head
		if isSpacesOnly(headWS) {
			f.HeadSeparatorSpaces = len(headWS)
		}
	}
	var nextToken string
	for {
		var nt string
		if nextToken == "" {
			if err := ScanStrings(s, "branch", "access", "symbols", "locks", "strict", "integrity", "comment", "expand", "\n\n", "\r\n\r\n", " ", "\t", "\n", "\r\n"); err != nil {
				return err
			}
			nt = s.Text()
		} else {
			nt = nextToken
			nextToken = ""
		}

		switch nt {
		case " ", "\t", "\n", "\r\n":
			if f.NewLine == "" && (nt == "\n" || nt == "\r\n") {
				f.NewLine = nt
			}
			continue
		case "branch":
			if branch, ws, err := ParseOptionalTokenWithSpacing(s, ScanTokenNum, WithPropertyName("branch"), WithConsumed(true), WithLine(true)); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Branch = branch
				if isSpacesOnly(ws) {
					f.BranchSeparatorSpaces = len(ws)
				}
			}
		case "access":
			f.Access = true
			if users, ws, err := ParseHeaderAccessWithSpacing(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.AccessUsers = users
				if len(users) == 0 && isSpacesOnly(ws) {
					f.AccessSeparatorSpaces = len(ws)
				}
			}
		case "symbols":
			if sym, symFmt, err := ParseHeaderSymbolsWithSpacing(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Symbols = sym
				f.SymbolTerminatorPrefix = symFmt.TerminatorPrefix
				f.SymbolsInline = symFmt.Inline
				if isSpacesOnly(symFmt.FirstItemWhitespace) {
					f.SymbolsFirstSpaces = len(symFmt.FirstItemWhitespace)
				}
				if isSpacesOnly(symFmt.BetweenItemWhitespace) {
					f.SymbolsBetweenSpaces = len(symFmt.BetweenItemWhitespace)
				}
				if len(sym) == 0 && isSpacesOnly(symFmt.SeparatorWhitespace) {
					f.SymbolsSeparatorSpaces = len(symFmt.SeparatorWhitespace)
				}
			}
		case "locks":
			var err error
			var lockResult HeaderLocksParseResult
			if lockResult, err = ParseHeaderLocksWithSpacing(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				nextToken = lockResult.NextToken
				f.Locks = lockResult.Locks
				if len(lockResult.Locks) == 0 && isSpacesOnly(lockResult.FormatOptions.SeparatorWhitespace) {
					f.LocksSeparatorSpaces = len(lockResult.FormatOptions.SeparatorWhitespace)
				}
				if lockResult.InlineStrict {
					f.Strict = true
				}
			}
		case "strict":
			f.Strict = true
			f.StrictOnOwnLine = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			}
		case "integrity":
			if integrity, _, err := ParseHeaderCommentWithSpacing(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Integrity = integrity
			}
		case "comment":
			if comment, ws, err := ParseHeaderCommentWithSpacing(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Comment = comment
				if isSpacesOnly(ws) {
					f.CommentSeparatorSpaces = len(ws)
				}
			}
		case "expand":
			if expand, ws, err := ParseOptionalTokenWithSpacing(s, ScanTokenWord, WithPropertyName("expand"), WithConsumed(true), WithLine(true)); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Expand = expand
				if isSpacesOnly(ws) {
					f.ExpandSeparatorSpaces = len(ws)
				}
			}
		case "\n\n", "\r\n\r\n":
			return nil
		default:
			return fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func parseRevisionHeadersWithOffset(s *Scanner) ([]*RevisionHead, bool, int, int, error) {
	var rhs []*RevisionHead
	revisionStartLineOffset := 0
	for {
		rh, next, descConsumed, skippedNewLines, descNewLineOffset, err := parseRevisionHeaderWithOffset(s)
		if err != nil {
			return nil, false, 0, 0, err
		}
		if rh != nil && len(rhs) == 0 {
			revisionStartLineOffset = skippedNewLines - 1
		}
		if descConsumed {
			return rhs, true, revisionStartLineOffset, descNewLineOffset, nil
		}
		if rh == nil {
			return rhs, false, revisionStartLineOffset, 0, nil
		}
		rhs = append(rhs, rh)
		if !next {
			return rhs, false, revisionStartLineOffset, 0, nil
		}
	}
}

func parseRevisionHeaderWithOffset(s *Scanner) (*RevisionHead, bool, bool, int, int, error) {
	rh := &RevisionHead{}
	skippedNewLines := 0
	for {
		if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
			if IsNotFound(err) || IsEOFError(err) {
				return nil, false, false, skippedNewLines, 0, nil
			}
			return nil, false, false, skippedNewLines, 0, err
		}
		rev := s.Text()
		if err := ScanNewLine(s, false); err != nil {
			return nil, false, false, skippedNewLines, 0, err
		}
		if rev != "" {
			rh.Revision = Num(rev)
			break
		}
		skippedNewLines++
	}
	if rh.Revision == "desc" {
		return nil, false, true, skippedNewLines, skippedNewLines - 1, nil
	}
	for {
		if err := ScanStrings(s, "branches", "date", "next", "commitid", "owner", "group", "permissions", "hardlinks", "deltatype", "kopt", "mergepoint", "filename", "username", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			if IsNotFound(err) {
				id, idErr := ScanTokenId(s)
				if idErr == nil && id != "" {
					if id == "desc" {
						if err := ScanWhiteSpace(s, 0); err != nil {
							return nil, false, false, skippedNewLines, 0, err
						}
						return rh, false, true, skippedNewLines, skippedNewLines - 1, nil
					}

					np, err := ParseNewPhraseValue(s)
					if err != nil {
						return nil, false, false, skippedNewLines, 0, fmt.Errorf("parsing new phrase %q: %w", id, err)
					}
					rh.NewPhrases = append(rh.NewPhrases, &NewPhrase{Key: ID(id), Value: np})
					continue
				}

				return rh, false, false, skippedNewLines, 0, nil
			}
			return nil, false, false, skippedNewLines, 0, fmt.Errorf("finding revision header field: %w", err)
		}

		nt := s.Text()
		switch nt {
		case "branches":
			if err := ParseRevisionHeaderBranches(s, rh, true); err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "date":
			if err := ParseRevisionHeaderDateLine(s, true, rh); err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "next":
			n, ws, err := ParseOptionalTokenWithSpacing(s, ScanTokenNum, WithPropertyName("next"), WithConsumed(true), WithLine(true))
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.NextRevision = Num(n)
			if isSpacesOnly(ws) {
				rh.NextSeparatorSpaces = len(ws)
			}
		case "commitid":
			c, err := ParseOptionalToken(s, ScanTokenId, WithPropertyName("commitid"), WithConsumed(true), WithLine(true))
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.CommitID = Sym(c)
		case "owner":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Owner = v
		case "group":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Group = v
		case "permissions":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Permissions = v
		case "hardlinks":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Hardlinks = v
		case "deltatype":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Deltatype = v
		case "kopt":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Kopt = v
		case "mergepoint":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Mergepoint = v
		case "filename":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Filename = v
		case "username":
			v, err := ParseNewPhraseValue(s)
			if err != nil {
				return nil, false, false, skippedNewLines, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
			rh.Username = v
		case "\n\n", "\r\n\r\n":
			return rh, true, false, skippedNewLines, 0, nil
		case "\n", "\r\n":
			continue
		default:
			return nil, false, false, skippedNewLines, 0, fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func ParseRevisionHeaders(s *Scanner) ([]*RevisionHead, bool, error) {
	rhs, descConsumed, _, _, err := parseRevisionHeadersWithOffset(s)
	return rhs, descConsumed, err
}

func ParseRevisionHeader(s *Scanner) (*RevisionHead, bool, bool, error) {
	rh, next, descConsumed, _, _, err := parseRevisionHeaderWithOffset(s)
	return rh, next, descConsumed, err
}

func ParseNewPhraseValue(s *Scanner) (PhraseValues, error) {
	var words PhraseValues
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, err
		}
		if err := ScanStrings(s, ";"); err == nil {
			break
		}
		word, err := ScanTokenPhrase(s)
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

func ParseRevisionContents(s *Scanner) ([]*RevisionContent, int, error) {
	var rcs []*RevisionContent
	var initialOffset int
	for {
		rc, newLines, err := ParseRevisionContent(s)
		if err != nil {
			return nil, 0, err
		}
		if rc != nil {
			rc.PrecedingNewLinesOffset += initialOffset
			rcs = append(rcs, rc)
		} else {
			return rcs, initialOffset + newLines - 1, nil
		}
		if newLines < 2 {
			return rcs, newLines - 1, nil
		}
		initialOffset = newLines
	}
}

func ParseRevisionContent(s *Scanner) (*RevisionContent, int, error) {
	rh := &RevisionContent{}
	precedingNewLines := 0
	for {
		if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
			if IsEOFError(err) {
				return nil, precedingNewLines, nil
			}
			return nil, precedingNewLines, err
		}
		rev := s.Text()
		if err := ScanNewLine(s, false); err != nil {
			return nil, precedingNewLines, err
		}
		if rev != "" {
			rh.Revision = rev
			rh.PrecedingNewLinesOffset = precedingNewLines - 2
			break
		}
		precedingNewLines++
		if precedingNewLines > 4 {
			return nil, precedingNewLines, fmt.Errorf("%w: %d", ErrTooManyNewLines, precedingNewLines)
		}
	}
	for {
		if err := ScanStrings(s, "log", "text", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			if s.Bytes() != nil && len(s.Bytes()) == 0 {
				return rh, 0, nil
			}
			return nil, 0, err
		}
		nt := s.Text()
		switch nt {
		case "log":
			if s, err := ParseRevisionContentLog(s); err != nil {
				return nil, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Log = s
			}
		case "text":
			if s, err := ParseRevisionContentText(s); err != nil {
				return nil, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Text = s
			}
		case "\n\n", "\r\n\r\n":
			return rh, 2, nil
		case "\n", "\r\n":
			return rh, 1, nil
		default:
			return nil, 0, fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func IsEOFError(err error) bool {
	var eofErr ErrEOF
	isEof := errors.As(err, &eofErr)
	return isEof
}

func ParseRevisionHeaderBranches(s *Scanner, rh *RevisionHead, havePropertyName bool) error {
	if !havePropertyName {
		if err := ScanStrings(s, "branches"); err != nil {
			return err
		}
	}
	rh.Branches = []Num{}
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return err
		}
		ws := s.Text()
		if err := ScanStrings(s, ";"); err == nil {
			if len(rh.Branches) == 0 && isSpacesOnly(ws) {
				rh.BranchesSeparatorSpaces = len(ws)
			}
			break
		}
		if len(rh.Branches) == 0 && isSpacesOnly(ws) {
			rh.BranchesSeparatorSpaces = len(ws)
		}
		num, err := ScanTokenNum(s)
		if err != nil {
			return fmt.Errorf("expected num in branches: %w", err)
		}
		rh.Branches = append(rh.Branches, Num(num))
	}
	return nil
}

func ParseHeaderComment(s *Scanner, havePropertyName bool) (string, error) {
	sr, _, err := ParseHeaderCommentWithSpacing(s, havePropertyName)
	return sr, err
}

func ParseHeaderCommentWithSpacing(s *Scanner, havePropertyName bool) (string, string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "comment"); err != nil {
			return "", "", err
		}
	}
	if err := ScanWhiteSpace(s, 0); err != nil {
		return "", "", err
	}
	ws := s.Text()
	sr, err := ParseAtQuotedString(s)
	if err != nil {
		return "", "", err
	}
	if err := ParseTerminatorFieldLine(s); err != nil {
		return "", "", err
	}
	return sr, ws, nil
}

func ParseHeaderAccess(s *Scanner, havePropertyName bool) ([]string, error) {
	ids, _, err := ParseHeaderAccessWithSpacing(s, havePropertyName)
	return ids, err
}

func ParseHeaderAccessWithSpacing(s *Scanner, havePropertyName bool) ([]string, string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "access"); err != nil {
			return nil, "", err
		}
	}
	var ids []string
	var wsBeforeTerm string
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, "", err
		}
		ws := s.Text()
		if err := ScanStrings(s, ";"); err == nil {
			wsBeforeTerm = ws
			break
		}
		id, err := ScanTokenId(s)
		if err != nil {
			return nil, "", fmt.Errorf("expected id in access: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, wsBeforeTerm, nil
}

func ParseHeaderSymbols(s *Scanner, havePropertyName bool) ([]*Symbol, string, error) {
	syms, fmtOpts, err := ParseHeaderSymbolsWithSpacing(s, havePropertyName)
	return syms, fmtOpts.TerminatorPrefix, err
}

func ParseHeaderSymbolsWithSpacing(s *Scanner, havePropertyName bool) ([]*Symbol, SymbolFormattingOptions, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "symbols"); err != nil {
			return nil, SymbolFormattingOptions{}, err
		}
	}
	m := make([]*Symbol, 0)
	fmtOpts := SymbolFormattingOptions{}
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, SymbolFormattingOptions{}, err
		}
		ws := s.Text()
		if strings.Contains(ws, "\n") || strings.Contains(ws, "\r") {
			fmtOpts.Inline = false
		}
		if err := ScanStrings(s, ";"); err == nil {
			fmtOpts.SeparatorWhitespace = ws
			if strings.Contains(ws, "\n") || strings.Contains(ws, "\r") {
				fmtOpts.TerminatorPrefix = ws
			}
			break
		}
		if len(m) == 0 {
			fmtOpts.FirstItemWhitespace = ws
			if !strings.Contains(ws, "\n") && !strings.Contains(ws, "\r") {
				fmtOpts.Inline = true
			}
		} else if fmtOpts.BetweenItemWhitespace == "" {
			fmtOpts.BetweenItemWhitespace = ws
		}

		sym, err := ScanTokenSym(s)
		if err != nil {
			return nil, SymbolFormattingOptions{}, fmt.Errorf("expected sym in symbols: %w", err)
		}

		if err := ScanStrings(s, ":"); err != nil {
			return nil, SymbolFormattingOptions{}, fmt.Errorf("expected : after sym %q: %w", sym, err)
		}

		num, err := ScanTokenNum(s)
		if err != nil {
			return nil, SymbolFormattingOptions{}, fmt.Errorf("expected num for sym %q: %w", sym, err)
		}
		m = append(m, &Symbol{Name: sym, Revision: num})
	}
	return m, fmtOpts, nil
}

func ParseAtQuotedString(s *Scanner) (string, error) {
	if err := ScanStrings(s, "@"); err != nil {
		return "", fmt.Errorf("open quote: %v", err)
	}
	return ParseAtQuotedStringBody(s)
}

func ParseAtQuotedStringBody(s *Scanner) (string, error) {
	sb := &strings.Builder{}
	for {
		if err := ScanUntilStrings(s, "@"); err != nil {
			return "", err
		}
		sb.WriteString(s.Text())
		if err := ScanStrings(s, "@@", "@"); err != nil {
			return "", err
		}
		nt := s.Text()
		switch nt {
		case "@@":
			sb.WriteString("@")
		case "@":
			return sb.String(), nil
		default:
			return "", fmt.Errorf("unexpected token %q", nt)
		}
	}
}

func ParseHeaderLocks(s *Scanner, havePropertyName bool) ([]*Lock, bool, string, error) {
	result, err := ParseHeaderLocksWithSpacing(s, havePropertyName)
	if err != nil {
		return nil, false, "", err
	}
	return result.Locks, result.InlineStrict, result.NextToken, nil
}

func ParseHeaderLocksWithSpacing(s *Scanner, havePropertyName bool) (HeaderLocksParseResult, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "locks"); err != nil {
			return HeaderLocksParseResult{}, err
		}
	}
	locks := make([]*Lock, 0)
	var strict bool
	fmtOpts := HeaderLocksFormattingOptions{}
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return HeaderLocksParseResult{}, err
		}
		ws := s.Text()
		if err := ScanStrings(s, ";"); err == nil {
			fmtOpts.SeparatorWhitespace = ws
			if scanInlineStrict(s) {
				strict = true
			}
			return HeaderLocksParseResult{
				Locks:         locks,
				InlineStrict:  strict,
				FormatOptions: fmtOpts,
			}, nil
		}

		l, err := ParseLockLine(s)
		if err != nil {
			return HeaderLocksParseResult{}, err
		}
		locks = append(locks, l)
	}
}

func ParseOptionalTokenWithSpacing(s *Scanner, scannerFunc func(*Scanner) (string, error), options ...interface{}) (string, string, error) {
	var propertyName string
	var havePropertyName bool
	var line bool

	for _, opt := range options {
		switch v := opt.(type) {
		case WithPropertyName:
			propertyName = string(v)
		case WithConsumed:
			havePropertyName = bool(v)
		case WithLine:
			line = bool(v)
		}
	}

	if !havePropertyName {
		if err := ScanStrings(s, propertyName); err != nil {
			return "", "", err
		}
	}
	if err := ScanWhiteSpace(s, 1); err != nil {
		return "", "", err
	}
	ws := s.Text()
	if err := ScanStrings(s, ";"); err == nil {
		return "", ws, nil
	}
	val, err := scannerFunc(s)
	if err != nil {
		return "", "", ErrParseProperty{Property: propertyName, Err: err}
	}
	if line {
		if err := ParseTerminatorFieldLine(s); err != nil {
			return "", "", err
		}
	} else {
		if err := ScanFieldTerminator(s); err != nil {
			return "", "", err
		}
	}
	return val, ws, nil
}

func isSpacesOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != ' ' {
			return false
		}
	}
	return true
}

func shouldPreserveParsedSpacing(f *File) bool {
	return f.SymbolsInline && len(f.Symbols) > 0
}

func clearParsedSpacing(f *File) {
	f.HeadSeparatorSpaces = 0
	f.BranchSeparatorSpaces = 0
	f.AccessSeparatorSpaces = 0
	f.SymbolsSeparatorSpaces = 0
	f.SymbolsInline = false
	f.SymbolsFirstSpaces = 0
	f.SymbolsBetweenSpaces = 0
	f.LocksSeparatorSpaces = 0
	f.CommentSeparatorSpaces = 0
	f.ExpandSeparatorSpaces = 0
	for _, rh := range f.RevisionHeads {
		rh.DateSeparatorSpaces = 0
		rh.DateAuthorSpacingSpaces = 0
		rh.AuthorStateSpacingSpaces = 0
		rh.BranchesSeparatorSpaces = 0
		rh.NextSeparatorSpaces = 0
	}
}

func shouldPreserveEmptyMasterHeaderSpacing(f *File) bool {
	return f.Head == "" &&
		f.Branch == "" &&
		f.Access && len(f.AccessUsers) == 0 &&
		f.Symbols != nil && len(f.Symbols) == 0 &&
		f.Locks != nil && len(f.Locks) == 0
}

func scanInlineStrict(s *Scanner) bool {
	// We want to scan "strict;" but only if there are NO newlines before it.
	// We scan horizontal whitespace first.
	if err := ScanRunesUntil(s, 0, func(i []byte) bool {
		return i[0] != ' ' && i[0] != '\t'
	}, "horizontal whitespace"); err != nil {
		return false
	}
	// Check if next is 'strict'
	if err := ScanStrings(s, "strict"); err != nil {
		return false
	}
	// Consumed 'strict'. Now expect ';'.
	if err := ScanFieldTerminator(s); err != nil {
		return false
	}
	return true
}

func ParseLockLine(s *Scanner) (*Lock, error) {
	id, err := ScanTokenId(s)
	if err != nil {
		return nil, fmt.Errorf("expected id in lock: %w", err)
	}

	if err := ScanStrings(s, ":"); err != nil {
		return nil, fmt.Errorf("expected : after lock id %q: %w", id, err)
	}

	return ParseLockBody(s, id)
}

func ParseLockBody(s *Scanner, user string) (*Lock, error) {
	l := &Lock{User: user}
	num, err := ScanTokenNum(s)
	if err != nil {
		return nil, fmt.Errorf("expected num in lock: %w", err)
	}
	l.Revision = num
	return l, nil
}

func ParseRevisionHeaderDateLine(s *Scanner, haveHead bool, rh *RevisionHead) error {
	dateOpts := []interface{}{WithPropertyName("date")}
	if haveHead {
		dateOpts = append(dateOpts, WithConsumed(true))
	}
	dateStr, dateWS, err := ParseOptionalTokenWithSpacing(s, ScanTokenNum, dateOpts...)
	if err != nil {
		return err
	}
	dateStr = strings.TrimSpace(dateStr)
	if isSpacesOnly(dateWS) {
		rh.DateSeparatorSpaces = len(dateWS)
	}
	if i := strings.Index(dateStr, "."); i == 2 {
		rh.YearTruncated = true
	}
	if _, err := ParseDate(dateStr, time.Time{}, nil); err != nil {
		return err
	}
	rh.Date = DateTime(dateStr)

	if err := ScanWhiteSpace(s, 0); err != nil {
		return err
	}
	if isSpacesOnly(s.Text()) {
		rh.DateAuthorSpacingSpaces = len(s.Text())
	}
	if err := ScanStrings(s, "author"); err != nil {
		return err
	}
	authorVal, _, err := ParseOptionalTokenWithSpacing(s, ScanTokenAuthor, WithPropertyName("author"), WithConsumed(true))
	if err != nil {
		return fmt.Errorf("token %q: %w", "author", err)
	}
	rh.Author = ID(authorVal)

	if err := ScanWhiteSpace(s, 0); err != nil {
		return err
	}
	if isSpacesOnly(s.Text()) {
		rh.AuthorStateSpacingSpaces = len(s.Text())
	}
	if err := ScanStrings(s, "state"); err != nil {
		return err
	}
	stateVal, _, err := ParseOptionalTokenWithSpacing(s, ScanTokenId, WithPropertyName("state"), WithConsumed(true))
	if err != nil {
		return fmt.Errorf("token %q: %w", "state", err)
	}
	rh.State = ID(stateVal)
	return nil
}

type WithPropertyName string
type WithConsumed bool
type WithLine bool

func ParseOptionalToken(s *Scanner, scannerFunc func(*Scanner) (string, error), options ...interface{}) (string, error) {
	var propertyName string
	var havePropertyName bool
	var line bool

	for _, opt := range options {
		switch v := opt.(type) {
		case WithPropertyName:
			propertyName = string(v)
		case WithConsumed:
			havePropertyName = bool(v)
		case WithLine:
			line = bool(v)
		}
	}

	if !havePropertyName {
		if err := ScanStrings(s, propertyName); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 1); err != nil {
		return "", err
	}
	// Important: Check for terminator *before* value scan.
	// If ";" is found, it means the value is empty/missing, which is valid for optional tokens.
	if err := ScanStrings(s, ";"); err == nil {
		return "", nil
	}
	// If we didn't find ";", we expect a value.
	val, err := scannerFunc(s)
	if err != nil {
		return "", ErrParseProperty{Property: propertyName, Err: err}
	}
	if line {
		if err := ParseTerminatorFieldLine(s); err != nil {
			return "", err
		}
	} else {
		if err := ScanFieldTerminator(s); err != nil {
			return "", err
		}
	}
	return val, nil
}

func ParsePropertyNum(s *Scanner, havePropertyName bool, propertyName string, line bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, propertyName); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 1); err != nil {
		return "", err
	}
	// Use ScanRunesUntil with minimum 0 because we might have an empty property (no num).
	// But check if it looks like a number start (digit or .) or stops at something else.
	// This will scan UNTIL predicate is true.
	// Predicate: !isDigit && !isDot.
	err := ScanRunesUntil(s, 0, func(b []byte) bool {
		r := bytes.Runes(b)[0]
		return !isDigit(r) && r != '.'
	}, "num")
	if err != nil {
		return "", err
	}
	result := s.Text()

	if line {
		if err := ParseTerminatorFieldLine(s); err != nil {
			return "", err
		}
	} else {
		if err := ScanFieldTerminator(s); err != nil {
			return "", err
		}
	}
	return result, nil
}

func ParseProperty(s *Scanner, havePropertyName bool, propertyName string, line bool) (string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, propertyName); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 1); err != nil {
		return "", err
	}
	if err := ScanUntilFieldTerminator(s); err != nil {
		return "", err
	}
	result := s.Text()
	if line {
		if err := ParseTerminatorFieldLine(s); err != nil {
			return "", err
		}
	} else {
		if err := ScanFieldTerminator(s); err != nil {
			return "", err
		}
	}
	return result, nil
}

func ParseTerminatorFieldLine(s *Scanner) error {
	if err := ScanFieldTerminator(s); err != nil {
		return err
	}
	return nil
}

func ScanWhiteSpace(s *Scanner, minimum int) error {
	return ScanRunesUntil(s, minimum, func(i []byte) bool {
		return !unicode.IsSpace(bytes.Runes(i)[0])
	}, "whitespace")
}

func ScanUntilNewLine(s *Scanner) error {
	return ScanUntilStrings(s, "\r\n", "\n")
}

func ScanUntilFieldTerminator(s *Scanner) error {
	return ScanUntilStrings(s, ";")
}

func ScanRunesUntil(s *Scanner, minimum int, until func([]byte) bool, name string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		err = nil
		adv := 0
		for {
			if !atEOF && adv >= len(data) {
				return 0, nil, nil
			}
			a, t, err := bufio.ScanRunes(data[adv:], atEOF)
			if err != nil {
				return 0, nil, err
			}
			if a == 0 && t == nil {
				if atEOF {
					if minimum > 0 && minimum > adv {
						break
					}
					f := data[:adv]
					return adv, f, nil
				}
				return 0, nil, nil
			}
			if until(t) {
				if minimum > 0 && minimum > adv {
					break
				}
				// STOP and return what we have so far
				f := data[:adv]
				return adv, f, nil
			}
			adv += a
		}
		err = ScanUntilNotFound{
			Until: name,
			Pos:   *s.pos,
			Found: string(data),
		}
		return 0, []byte{}, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if err != nil {
			return err
		}
		return ScanUntilNotFound{
			Until: name,
			Pos:   *s.pos,
			Found: "",
		}
	}
	return
}

func ScanFieldTerminator(s *Scanner) error {
	return ScanStrings(s, ";")
}

func ScanNewLine(s *Scanner, orEof bool) error {
	if orEof {
		return ScanStrings(s, "\r\n", "\n", "")
	}
	return ScanStrings(s, "\r\n", "\n")
}

type ScanNotFound struct {
	LookingFor []string
	Pos        Pos
	Found      string
}

func (se ScanNotFound) Error() string {
	strs := make([]string, len(se.LookingFor))
	for i, s := range se.LookingFor {
		strs[i] = fmt.Sprintf("%#v", s)
	}
	lookingFor := strings.Join(strs, ", ")
	found := se.Found
	if len(found) > 20 {
		runes := []rune(found)
		if len(runes) > 20 {
			found = string(runes[:20]) + "..."
		}
	}
	return fmt.Sprintf("looking for %s at %s but found %q", lookingFor, se.Pos.String(), found)
}

type ScanUntilNotFound struct {
	Until string
	Pos   Pos
	Found string
}

func (se ScanUntilNotFound) Error() string {
	found := se.Found
	if len(found) > 20 {
		runes := []rune(found)
		if len(runes) > 20 {
			found = string(runes[:20]) + "..."
		}
	}
	return fmt.Sprintf("scanning for %q at %s but found %q", se.Until, se.Pos.String(), found)
}

func IsNotFound(err error) bool {
	switch err.(type) {
	case ScanUntilNotFound, ScanNotFound:
		return true
	}
	e1 := ScanNotFound{}
	e2 := ScanUntilNotFound{}
	return errors.As(err, &e1) || errors.As(err, &e2)
}

func ScanStrings(s *Scanner, strs ...string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		err = nil
		for _, ss := range strs {
			if len(ss) == 0 && atEOF && len(data) == 0 {
				return 0, []byte{}, nil
			} else if len(ss) == 0 {
				continue
			}
			i := len(ss)
			if i > len(data) && !atEOF && bytes.HasPrefix([]byte(ss), data) {
				return 0, nil, nil
			}
			if bytes.HasPrefix(data, []byte(ss)) {
				rs := data[:i]
				return i, rs, nil
			}
		}
		if atEOF {
			err = ScanNotFound{
				LookingFor: strs,
				Pos:        *s.pos,
				Found:      string(data),
			}
			return 0, []byte{}, nil
		}
		return 0, nil, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if err != nil {
			return err
		}
		return ScanNotFound{
			LookingFor: strs,
			Pos:        *s.pos,
			Found:      "",
		}
	}
	return
}

type ErrEOF struct{ error }

func (e ErrEOF) Error() string {
	return "EOF:" + e.error.Error()
}

func ScanUntilStrings(s *Scanner, strs ...string) (err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		bestIdx := -1
		for _, ss := range strs {
			if len(ss) == 0 {
				continue
			}
			idx := bytes.Index(data, []byte(ss))
			if idx >= 0 {
				if bestIdx == -1 || idx < bestIdx {
					bestIdx = idx
				}
			}
		}
		if bestIdx >= 0 {
			return bestIdx, data[:bestIdx], nil
		}
		if atEOF {
			err = ErrEOF{ScanNotFound{
				LookingFor: strs,
				Pos:        *s.pos,
				Found:      string(data),
			}}
			return 0, []byte{}, nil
		}
		return 0, nil, nil
	})
	if !s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if err != nil {
			return err
		}
		return ScanNotFound{
			LookingFor: strs,
			Pos:        *s.pos,
			Found:      "",
		}
	}
	return err
}

func ScanLockIdOrStrings(s *Scanner, strs ...string) (id string, match string, err error) {
	s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		// Check strings first
		for _, ss := range strs {
			if len(ss) == 0 {
				continue
			}
			if len(data) < len(ss) {
				if atEOF {
					if bytes.Equal(data, []byte(ss)) {
						return len(data), data, nil
					}
					// mismatch or partial mismatch at EOF
				} else {
					if bytes.HasPrefix([]byte(ss), data) {
						// potential partial match
						return 0, nil, nil
					}
				}
			} else {
				if bytes.HasPrefix(data, []byte(ss)) {
					return len(ss), data[:len(ss)], nil
				}
			}
		}

		// Check Lock ID
		for i, b := range data {
			if b == ':' {
				if i == 0 {
					return 0, nil, ErrEmptyId
				}
				return i, data[:i], nil
			}
			switch b {
			case ' ', '\t', '\n', '\r', ';':
				return 0, nil, bufio.ErrFinalToken
			}
		}

		if atEOF {
			return 0, nil, nil
		}
		return 0, nil, nil
	})

	if !s.Scan() {
		if s.Err() != nil {
			return "", "", s.Err()
		}
		return "", "", ScanNotFound{
			LookingFor: append(strs, "lock_id"),
			Pos:        *s.pos,
			Found:      "",
		}
	}
	text := s.Text()
	for _, ss := range strs {
		if text == ss {
			return "", text, nil
		}
	}
	return text, "", nil
}
