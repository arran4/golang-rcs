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

func ScanStrings(s *Scanner, strs ...string) (err error) {
	return s.ScanMatch(strs...)
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
