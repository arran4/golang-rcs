package rcs

import (
	"fmt"
	"io"
	"strings"

	"github.com/arran4/golang-rcs/diff"
)

// COVerdict is the result of resolving a checkout request.
type COVerdict struct {
	Revision     string
	Content      string
	FileModified bool
	LockSet      bool
	LockCleared  bool
}

type WithLock int

const (
	WithNoLockChange WithLock = iota
	WithSetLock
	WithClearLock
)

type WithRevision string

// WithExpandKeyword sets the keyword substitution mode (e.g. KV, KVL, K) for the checkout operation.
// This is required to control how keywords like $Id$ and $Log$ are expanded in the working file.
type WithExpandKeyword KeywordSubstitution

// WithRCSFilename sets the filename to be used in keyword expansion (e.g. for $Source$ or $Id$).
// This is required because the file content itself typically doesn't know its own filename.
type WithRCSFilename string

// Checkout resolves revision content and applies lock changes to the file.
//
// Options:
//   - WithRevision("1.2") picks an explicit revision (defaults to file head)
//   - WithSetLock / WithClearLock controls lock mutation
//   - WithExpandKeyword(KV) sets keyword substitution mode
//   - WithRCSFilename("foo.c,v") sets RCS filename for keyword expansion
func (file *File) Checkout(user string, ops ...any) (*COVerdict, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file")
	}
	revision := file.Head
	lockMode := WithNoLockChange
	keywordMode := KV
	rcsFilename := ""

	for _, op := range ops {
		switch v := op.(type) {
		case WithRevision:
			revision = string(v)
		case WithLock:
			lockMode = v
		case WithExpandKeyword:
			keywordMode = KeywordSubstitution(v)
		case WithRCSFilename:
			rcsFilename = string(v)
		default:
			return nil, fmt.Errorf("unsupported checkout option type %T", op)
		}
	}

	// Resolve revision if it is a symbol
	if file.Symbols != nil {
		for _, sym := range file.Symbols {
			if sym.Name == revision {
				revision = sym.Revision
				break
			}
		}
	}

	if revision == "" {
		return nil, fmt.Errorf("missing target revision")
	}
	if lockMode == WithSetLock || lockMode == WithClearLock {
		if user == "" {
			return nil, fmt.Errorf("lock operation requires user")
		}
	}

	content, err := file.resolveRevisionContent(revision)
	if err != nil {
		return nil, err
	}

	// Apply keyword expansion
	if keywordMode != O && keywordMode != B {
		rh := file.GetRevisionHead(revision)
		if rh == nil {
			return nil, fmt.Errorf("revision head %q not found for keyword expansion", revision)
		}
		rc := file.GetRevisionContent(revision)
		if rc == nil {
			return nil, fmt.Errorf("revision content %q not found for keyword expansion", revision)
		}

		ctx := KeywordContext{
			Revision:    revision,
			Date:        string(rh.Date),
			Author:      string(rh.Author),
			State:       string(rh.State),
			Log:         rc.Log,
			RCSFile:     rcsFilename,
			WorkingFile: "", // TODO: Pass working file name if needed
			Strict:      file.Strict,
			Comment:     file.Comment,
		}

		// Determine Locker
		if lockMode == WithSetLock {
			ctx.Locker = user
		} else if keywordMode == KVL {
			// Only for KVL do we check existing locks to populate Locker for expansion
			for _, l := range file.Locks {
				if l.Revision == revision {
					ctx.Locker = l.User
					break
				}
			}
		}

		content = ExpandKeywords(content, keywordMode, ctx)
	}

	v := &COVerdict{Revision: revision, Content: content}
	switch lockMode {
	case WithNoLockChange:
	case WithSetLock:
		changed := file.setLock(user, revision)
		v.FileModified = changed
		v.LockSet = changed
	case WithClearLock:
		changed := file.clearLock(user, revision)
		v.FileModified = changed
		v.LockCleared = changed
	default:
		return nil, fmt.Errorf("invalid lock mode %d", lockMode)
	}

	return v, nil
}

func (file *File) GetRevisionHead(rev string) *RevisionHead {
	for _, rh := range file.RevisionHeads {
		if rh.Revision.String() == rev {
			return rh
		}
	}
	return nil
}

func (file *File) GetRevisionContent(rev string) *RevisionContent {
	for _, rc := range file.RevisionContents {
		if rc.Revision == rev {
			return rc
		}
	}
	return nil
}

func (file *File) resolveRevisionContent(targetRevision string) (string, error) {
	head := file.Head
	if head == "" {
		return "", fmt.Errorf("missing head revision")
	}
	rhByRevision := map[string]*RevisionHead{}
	rcByRevision := map[string]*RevisionContent{}
	for _, rh := range file.RevisionHeads {
		rhByRevision[rh.Revision.String()] = rh
	}
	for _, rc := range file.RevisionContents {
		rcByRevision[rc.Revision] = rc
	}

	headContent, ok := rcByRevision[head]
	if !ok {
		return "", fmt.Errorf("head revision %q content not found", head)
	}
	if targetRevision == head {
		return headContent.Text, nil
	}

	currentRevision := head
	currentContent := headContent.Text
	visited := map[string]bool{head: true}

	for {
		rh, ok := rhByRevision[currentRevision]
		if !ok {
			return "", fmt.Errorf("revision header %q not found", currentRevision)
		}
		nextRevision := rh.NextRevision.String()
		if nextRevision == "" {
			return "", fmt.Errorf("revision %q not reachable from head %q", targetRevision, head)
		}
		if visited[nextRevision] {
			return "", fmt.Errorf("loop detected while resolving revision %q", targetRevision)
		}
		visited[nextRevision] = true

		nextContentDelta, ok := rcByRevision[nextRevision]
		if !ok {
			return "", fmt.Errorf("revision content %q not found", nextRevision)
		}

		nextContent, err := applyDelta(currentContent, nextContentDelta.Text)
		if err != nil {
			return "", fmt.Errorf("apply delta for %q: %w", nextRevision, err)
		}

		if nextRevision == targetRevision {
			return nextContent, nil
		}
		currentRevision = nextRevision
		currentContent = nextContent
	}
}

func applyDelta(from, delta string) (string, error) {
	ed, err := diff.ParseEdDiff(strings.NewReader(delta))
	if err != nil {
		return "", err
	}
	r := &lineReader{lines: splitLines(from)}
	w := &lineWriter{}
	if err := ed.Apply(r, w); err != nil {
		return "", err
	}
	return strings.Join(w.lines, "\n") + trailingNewline(from), nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func trailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return "\n"
	}
	return ""
}

type lineReader struct {
	lines []string
	idx   int
}

func (r *lineReader) ReadLine() (string, error) {
	if r.idx >= len(r.lines) {
		return "", io.EOF
	}
	line := r.lines[r.idx]
	r.idx++
	return line, nil
}

type lineWriter struct {
	lines []string
}

func (w *lineWriter) WriteLine(line string) error {
	w.lines = append(w.lines, line)
	return nil
}

func (file *File) setLock(user, revision string) bool {
	for _, l := range file.Locks {
		if l.User == user {
			if l.Revision == revision {
				return false
			}
			l.Revision = revision
			return true
		}
	}
	file.Locks = append(file.Locks, &Lock{User: user, Revision: revision})
	return true
}

func (file *File) clearLock(user, revision string) bool {
	out := file.Locks[:0]
	changed := false
	for _, l := range file.Locks {
		if l.User == user && l.Revision == revision {
			changed = true
			continue
		}
		out = append(out, l)
	}
	file.Locks = out
	return changed
}
