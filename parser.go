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
	Revision      Num
	Date          DateTime
	YearTruncated bool `json:",omitempty"`
	Author        ID
	State         ID
	Branches      []Num
	NextRevision  Num
	CommitID      Sym
	Owner         PhraseValues `json:",omitempty"` // CVS-NT
	Group         PhraseValues `json:",omitempty"` // CVS-NT
	Permissions   PhraseValues `json:",omitempty"` // CVS-NT
	Hardlinks     PhraseValues `json:",omitempty"` // CVS-NT
	Deltatype     PhraseValues `json:",omitempty"` // CVS-NT
	Kopt          PhraseValues `json:",omitempty"` // CVS-NT
	Mergepoint    PhraseValues `json:",omitempty"` // CVS-NT
	Filename      PhraseValues `json:",omitempty"` // CVS-NT
	Username      PhraseValues `json:",omitempty"` // CVS-NT
	NewPhrases    []*NewPhrase `json:",omitempty"`
}

func (h *RevisionHead) String() string {
	return h.StringWithNewLine("\n")
}

func (h *RevisionHead) StringWithNewLine(nl string) string {
	sb := strings.Builder{}
	sb.WriteString(h.Revision.String())
	sb.WriteString(nl)
	fmt.Fprintf(&sb, "date\t%s;\tauthor %s;\tstate %s;%s", h.Date, h.Author, h.State, nl)
	sb.WriteString("branches")
	if len(h.Branches) > 0 {
		sb.WriteString(nl + "\t")
		for i, b := range h.Branches {
			if i > 0 {
				sb.WriteString(nl + "\t")
			}
			sb.WriteString(b.String())
		}
		sb.WriteString(";")
	} else {
		sb.WriteString(";")
	}
	sb.WriteString(nl)
	fmt.Fprintf(&sb, "next\t%s;%s", h.NextRevision, nl)
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
	Revision                string
	Log                     string
	Text                    string
	PrecedingNewLinesOffset int `json:",omitempty"`
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
	Head                     string
	Branch                   string
	Description              string
	Comment                  string
	Access                   bool
	Symbols                  []*Symbol
	AccessUsers              []string
	Locks                    []*Lock
	Strict                   bool
	StrictOnOwnLine          bool `json:",omitempty"`
	DateYearPrefixTruncated  bool `json:",omitempty"`
	Integrity                string
	Expand                   string
	NewLine                  string
	EndOfFileNewLineOffset   int `json:",omitempty"`
	DescriptionNewLineOffset int `json:",omitempty"`
	RevisionHeads            []*RevisionHead
	RevisionContents         []*RevisionContent
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
	sb.WriteString(fmt.Sprintf("head\t%s;%s", f.Head, nl))
	if f.Branch != "" {
		sb.WriteString(fmt.Sprintf("branch\t%s;%s", f.Branch, nl))
	}
	if f.Access {
		if len(f.AccessUsers) > 0 {
			sb.WriteString("access ")
			sb.WriteString(strings.Join(f.AccessUsers, " "))
			sb.WriteString(";")
			sb.WriteString(nl)
		} else {
			sb.WriteString("access;")
			sb.WriteString(nl)
		}
	}
	if f.Symbols != nil {
		sb.WriteString("symbols")
		for _, sym := range f.Symbols {
			sb.WriteString(nl + "\t")
			sb.WriteString(fmt.Sprintf("%s:%s", sym.Name, sym.Revision))
		}
		sb.WriteString(";")
		sb.WriteString(nl)
	}

	if f.Locks != nil {
		sb.WriteString("locks")
		for _, lock := range f.Locks {
			sb.WriteString(nl + "\t")
			sb.WriteString(fmt.Sprintf("%s:%s", lock.User, lock.Revision))
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
	sb.WriteString("comment\t")
	_, _ = WriteAtQuote(&sb, f.Comment)
	sb.WriteString(";")
	sb.WriteString(nl)
	if f.Expand != "" {
		sb.WriteString("expand\t")
		_, _ = WriteAtQuote(&sb, f.Expand)
		sb.WriteString(";")
		sb.WriteString(nl)
	}
	sb.WriteString(nl)
	sb.WriteString(nl)
	for _, head := range f.RevisionHeads {
		sb.WriteString(head.StringWithNewLine(nl))
		sb.WriteString(nl)
	}
	descriptionNewLines := f.DescriptionNewLineOffset + 1
	if len(f.RevisionHeads) == 0 {
		if descriptionNewLines < 1 {
			descriptionNewLines = 1
		}
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
	if rhs, dc, dno, err := ParseRevisionHeaders(s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.pos, err)
	} else {
		descConsumed = dc
		f.RevisionHeads = rhs
		f.DescriptionNewLineOffset = dno
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
	if head, err := ParseOptionalToken(s, ScanTokenNum, WithPropertyName("head"), WithLine(true)); err != nil {
		return err
	} else {
		f.Head = head
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
			if branch, err := ParseOptionalToken(s, ScanTokenNum, WithPropertyName("branch"), WithConsumed(true), WithLine(true)); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Branch = branch
			}
		case "access":
			f.Access = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				users, err := ParseHeaderAccess(s, true)
				if err != nil {
					return fmt.Errorf("token %#v: %w", nt, err)
				}
				f.AccessUsers = users
			}
		case "symbols":
			if sym, err := ParseHeaderSymbols(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Symbols = sym
			}
		case "locks":
			var err error
			var locks []*Lock
			var strict bool
			if locks, strict, nextToken, err = ParseHeaderLocks(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Locks = locks
				if strict {
					f.Strict = true
					// StrictOnOwnLine remains false
				}
			}
		case "strict":
			f.Strict = true
			f.StrictOnOwnLine = true
			if err := ParseTerminatorFieldLine(s); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			}
		case "integrity":
			if integrity, err := ParseHeaderComment(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Integrity = integrity
			}
		case "comment":
			if comment, err := ParseHeaderComment(s, true); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Comment = comment
			}
		case "expand":
			if expand, err := ParseOptionalToken(s, ScanTokenWord, WithPropertyName("expand"), WithConsumed(true), WithLine(true)); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				f.Expand = expand
			}

		case "\n\n", "\r\n\r\n":
			return nil
		default:
			return fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
}

func ParseRevisionHeaders(s *Scanner) ([]*RevisionHead, bool, int, error) {
	var rhs []*RevisionHead
	for {
		if rh, next, descConsumed, descNewLineOffset, err := ParseRevisionHeader(s); err != nil {
			return nil, false, 0, err
		} else {
			if descConsumed {
				return rhs, true, descNewLineOffset, nil
			}
			if rh == nil {
				return rhs, false, 0, nil
			}
			rhs = append(rhs, rh)
			if !next {
				return rhs, false, 0, nil
			}
		}
	}
}

func ParseRevisionHeader(s *Scanner) (*RevisionHead, bool, bool, int, error) {
	rh := &RevisionHead{}
	for {
		if err := ScanUntilStrings(s, "\r\n", "\n"); err != nil {
			if IsNotFound(err) {
				return nil, false, false, 0, nil
			}
			if IsEOFError(err) {
				return nil, false, false, 0, nil
			}
			return nil, false, false, 0, err
		}
		rev := s.Text()
		if err := ScanNewLine(s, false); err != nil {
			return nil, false, false, 0, err
		}
		if rev != "" {
			rh.Revision = Num(rev)
			break
		}
	}
	if rh.Revision == "desc" {
		return nil, false, true, 0, nil
	}
	for {
		if err := ScanStrings(s, "branches", "date", "next", "commitid", "owner", "group", "permissions", "hardlinks", "deltatype", "kopt", "mergepoint", "filename", "username", "\n\n", "\r\n\r\n", "\n", "\r\n"); err != nil {
			if IsNotFound(err) {
				// Try to parse as generic ID (NewPhrase)
				// We expect an ID here.
				id, idErr := ScanTokenId(s)
				if idErr == nil && id != "" {
					// Check if it is "desc"
					if id == "desc" {
						// "desc" marks start of description.
						// We should return.
						// But we consumed "desc".
						// We need to return true (descConsumed) but `ParseRevisionHeader` returns (rh, next, descConsumed, err).
						// Wait, `ParseRevisionHeader` returns `descConsumed` as bool.
						// If we return `descConsumed=true`, caller assumes we consumed "desc".

						// We must also consume the following whitespace/newline to match ParseRevisionHeader logic when it consumes "desc".
						// ParseDescription expects to be at the value (start of @...@) if descConsumed is true.
						if err := ScanWhiteSpace(s, 0); err != nil {
							return nil, false, false, 0, err
						}

						return rh, false, true, 0, nil
					}

					// It's a new phrase key
					if np, err := ParseNewPhraseValue(s); err != nil {
						return nil, false, false, 0, fmt.Errorf("parsing new phrase %q: %w", id, err)
					} else {
						rh.NewPhrases = append(rh.NewPhrases, &NewPhrase{Key: ID(id), Value: np})
						continue
					}
				}

				// If ScanTokenId failed or empty
				return rh, false, false, 0, nil
			}
			return nil, false, false, 0, fmt.Errorf("finding revision header field: %w", err)
		}

		nt := s.Text()
		switch nt {
		case "branches":
			if err := ParseRevisionHeaderBranches(s, rh, true); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "date":
			if err := ParseRevisionHeaderDateLine(s, true, rh); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			}
		case "next":
			if n, err := ParseOptionalToken(s, ScanTokenNum, WithPropertyName("next"), WithConsumed(true), WithLine(true)); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.NextRevision = Num(n)
			}
		case "commitid":
			if c, err := ParseOptionalToken(s, ScanTokenId, WithPropertyName("commitid"), WithConsumed(true), WithLine(true)); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.CommitID = Sym(c)
			}
		case "owner":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Owner = v
			}
		case "group":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Group = v
			}
		case "permissions":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Permissions = v
			}
		case "hardlinks":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Hardlinks = v
			}
		case "deltatype":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Deltatype = v
			}
		case "kopt":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Kopt = v
			}
		case "mergepoint":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Mergepoint = v
			}
		case "filename":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Filename = v
			}
		case "username":
			if v, err := ParseNewPhraseValue(s); err != nil {
				return nil, false, false, 0, fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Username = v
			}
		case "\n\n", "\r\n\r\n":
			return rh, true, false, 0, nil
		case "\n", "\r\n":
			continue
		default:
			return nil, false, false, 0, fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
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
		if err := ScanStrings(s, ";"); err == nil {
			break
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
	if !havePropertyName {
		if err := ScanStrings(s, "comment"); err != nil {
			return "", err
		}
	}
	if err := ScanWhiteSpace(s, 0); err != nil {
		return "", err
	}
	sr, err := ParseAtQuotedString(s)
	if err != nil {
		return "", err
	}
	if err := ParseTerminatorFieldLine(s); err != nil {
		return "", err
	}
	return sr, nil

}

func ParseHeaderAccess(s *Scanner, havePropertyName bool) ([]string, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "access"); err != nil {
			return nil, err
		}
	}
	var ids []string
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, err
		}
		if err := ScanStrings(s, ";"); err == nil {
			break
		}
		id, err := ScanTokenId(s)
		if err != nil {
			return nil, fmt.Errorf("expected id in access: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func ParseHeaderSymbols(s *Scanner, havePropertyName bool) ([]*Symbol, error) {
	if !havePropertyName {
		if err := ScanStrings(s, "symbols"); err != nil {
			return nil, err
		}
	}
	m := make([]*Symbol, 0)
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, err
		}
		if err := ScanStrings(s, ";"); err == nil {
			break
		}

		sym, err := ScanTokenSym(s)
		if err != nil {
			return nil, fmt.Errorf("expected sym in symbols: %w", err)
		}

		if err := ScanStrings(s, ":"); err != nil {
			return nil, fmt.Errorf("expected : after sym %q: %w", sym, err)
		}

		num, err := ScanTokenNum(s)
		if err != nil {
			return nil, fmt.Errorf("expected num for sym %q: %w", sym, err)
		}
		m = append(m, &Symbol{Name: sym, Revision: num})
	}
	return m, nil
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
	if !havePropertyName {
		if err := ScanStrings(s, "locks"); err != nil {
			return nil, false, "", err
		}
	}
	locks := make([]*Lock, 0)
	var strict bool
	for {
		if err := ScanWhiteSpace(s, 0); err != nil {
			return nil, false, "", err
		}
		if err := ScanStrings(s, ";"); err == nil {
			if scanInlineStrict(s) {
				strict = true
			}
			return locks, strict, "", nil
		}

		l, err := ParseLockLine(s)
		if err != nil {
			return nil, false, "", err
		}
		locks = append(locks, l)
	}
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
	opts := []interface{}{WithPropertyName("date")}
	if haveHead {
		opts = append(opts, WithConsumed(true))
	}
	if dateStr, err := ParseOptionalToken(s, ScanTokenNum, opts...); err != nil {
		return err
	} else {
		dateStr = strings.TrimSpace(dateStr)
		if i := strings.Index(dateStr, "."); i == 2 {
			rh.YearTruncated = true
		}
		if _, err := ParseDate(dateStr, time.Time{}, nil); err != nil {
			return err
		} else {
			rh.Date = DateTime(dateStr)
		}
	}
	for {
		if err := ScanStrings(s, " ", "\t", "author", "state"); err != nil {
			if IsNotFound(err) {
				break
			}
			return err
		}
		nt := s.Text()
		switch nt {
		case " ", "\t":
			continue
		case "author":
			if s, err := ParseOptionalToken(s, ScanTokenAuthor, WithPropertyName("author"), WithConsumed(true)); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.Author = ID(s)
			}
		case "state":
			if s, err := ParseOptionalToken(s, ScanTokenId, WithPropertyName("state"), WithConsumed(true)); err != nil {
				return fmt.Errorf("token %#v: %w", nt, err)
			} else {
				rh.State = ID(s)
			}
		default:
			return fmt.Errorf("%w: %s", ErrUnknownToken, nt)
		}
	}
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
