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

// Checkout resolves revision content and applies lock changes to the file.
//
// Options:
//   - WithRevision("1.2") picks an explicit revision (defaults to file head)
//   - WithSetLock / WithClearLock controls lock mutation
func (file *File) Checkout(user string, ops ...any) (*COVerdict, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file")
	}
	revision := file.Head
	lockMode := WithNoLockChange
	for _, op := range ops {
		switch v := op.(type) {
		case WithRevision:
			revision = string(v)
		case WithLock:
			lockMode = v
		default:
			return nil, fmt.Errorf("unsupported checkout option type %T", op)
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

	v := &COVerdict{Revision: revision, Content: content}
	switch lockMode {
	case WithNoLockChange:
	case WithSetLock:
		changed := file.SetLock(user, revision)
		v.FileModified = changed
		v.LockSet = changed
	case WithClearLock:
		changed := file.ClearLock(user, revision)
		v.FileModified = changed
		v.LockCleared = changed
	default:
		return nil, fmt.Errorf("invalid lock mode %d", lockMode)
	}

	return v, nil
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

// TODO: create own file for File and other major elements

// SetLock sets a lock for the user on the given revision.
// Returns true if the lock was set or updated, false if it already existed identically.
func (file *File) SetLock(user, revision string) bool {
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

// ClearLock removes the lock for the user on the given revision.
// Returns true if a lock was removed.
func (file *File) ClearLock(user, revision string) bool {
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

// AddAccess adds users to the access list.
func (file *File) AddAccess(users []string) {
	if len(users) == 0 {
		return
	}
	file.Access = true
	existing := make(map[string]bool)
	for _, u := range file.AccessUsers {
		existing[u] = true
	}
	for _, u := range users {
		if !existing[u] {
			file.AccessUsers = append(file.AccessUsers, u)
			existing[u] = true
		}
	}
}

// RemoveAccess removes users from the access list.
func (file *File) RemoveAccess(users []string) {
	if len(users) == 0 {
		return
	}
	toRemove := make(map[string]bool)
	for _, u := range users {
		toRemove[u] = true
	}
	out := file.AccessUsers[:0]
	for _, u := range file.AccessUsers {
		if !toRemove[u] {
			out = append(out, u)
		}
	}
	file.AccessUsers = out
}

// RemoveAllAccess removes all users from the access list.
func (file *File) RemoveAllAccess() {
	file.AccessUsers = nil
	file.Access = true
}

// DeleteRevision deletes a revision or range of revisions.
// Currently only supports simple deletion logging or stub.
// Implementing full revision deletion is complex as it involves fixing up the delta tree.
func (file *File) DeleteRevision(revision string) error {
	// TODO: Implement revision deletion logic.
	// For now, this is a placeholder to allow verifying argument parsing.
	return nil
}
