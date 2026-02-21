package rcs

import (
	"fmt"
	"io"
	"strings"
	"time"
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
	fmt.Fprintf(&sb, "%s%s", c.Revision, nl)
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
	fmt.Fprintf(&sb, "head%s%s;%s", headSep, f.Head, nl)
	if f.Branch != "" {
		branchSep := "\t"
		if f.BranchSeparatorSpaces > 0 {
			branchSep = strings.Repeat(" ", f.BranchSeparatorSpaces)
		}
		fmt.Fprintf(&sb, "branch%s%s;%s", branchSep, f.Branch, nl)
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
				fmt.Fprintf(&sb, "%s:%s", sym.Name, sym.Revision)
			}
		} else {
			for _, sym := range f.Symbols {
				sb.WriteString(nl + "\t")
				fmt.Fprintf(&sb, "%s:%s", sym.Name, sym.Revision)
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
			fmt.Fprintf(&sb, "%s:%s", lock.User, lock.Revision)
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

type PhraseValue interface {
	fmt.Stringer
	Raw() string
}

type SimpleString string

func (s SimpleString) String() string {
	return string(s)
}

func (s SimpleString) Raw() string {
	return string(s)
}

type QuotedString string

func (s QuotedString) String() string {
	return "@" + strings.ReplaceAll(string(s), "@", "@@") + "@"
}

func (s QuotedString) Raw() string {
	return string(s)
}

type PhraseValues []PhraseValue

func (p PhraseValues) Format() {
	for i, v := range p {
		raw := v.Raw()
		valid := true
		if len(raw) == 0 {
			valid = false
		} else {
			for _, r := range raw {
				if !isIdChar(r) && r != '.' {
					valid = false
					break
				}
			}
		}

		if valid {
			// If it is valid ID, prefer SimpleString
			if _, ok := v.(SimpleString); !ok {
				p[i] = SimpleString(raw)
			}
		} else {
			// If it is invalid ID, must be QuotedString
			if _, ok := v.(QuotedString); !ok {
				p[i] = QuotedString(raw)
			}
		}
	}
}

type DateTime string

func (dt DateTime) String() string {
	return string(dt)
}

// DateTime returns the date.Time representation of the DateTime string.
// It tries to parse using DateFormat and DateFormatTruncated.
func (dt DateTime) DateTime() (time.Time, error) {
	return ParseDate(string(dt), time.Time{}, nil)
}

type Num string

func (n Num) String() string {
	return string(n)
}

type ID string

func (i ID) String() string {
	return string(i)
}

type Sym string

func (s Sym) String() string {
	return string(s)
}
