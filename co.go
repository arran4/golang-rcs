package rcs

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

type WithDate time.Time

type WithTimeZoneOffset int // Seconds East of UTC

type WithLocation struct {
	*time.Location
}

type WithExpandKeyword KeywordSubstitution

type WithRCSFilename string

// Checkout resolves revision content and applies lock changes to the file.
//
// Options:
//   - WithRevision("1.2") picks an explicit revision (defaults to file head)
//   - WithDate(time.Time) picks the latest revision on the selected branch/trunk <= date
//   - WithTimeZoneOffset(int) sets the timezone offset (in seconds) for parsing revision dates
//   - WithLocation(*time.Location) sets the location for parsing revision dates
//   - WithSetLock / WithClearLock controls lock mutation
//   - WithExpandKeyword(KeywordSubstitution) sets the keyword substitution mode
//   - WithRCSFilename(string) sets the RCS filename used for keyword expansion
func (file *File) Checkout(user string, ops ...any) (*COVerdict, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file")
	}
	revision := file.Head
	lockMode := WithNoLockChange
	var targetDate time.Time
	var targetLocation *time.Location
	expandMode := KV
	rcsFilename := ""

	for _, op := range ops {
		switch v := op.(type) {
		case WithRevision:
			revision = string(v)
		case WithDate:
			targetDate = time.Time(v)
		case WithLock:
			lockMode = v
		case WithTimeZoneOffset:
			targetLocation = time.FixedZone("", int(v))
		case WithLocation:
			targetLocation = v.Location
		case WithExpandKeyword:
			expandMode = KeywordSubstitution(v)
		case WithRCSFilename:
			rcsFilename = string(v)
		default:
			return nil, fmt.Errorf("unsupported checkout option type %T", op)
		}
	}

	if !targetDate.IsZero() {
		defaultZone := time.UTC
		if targetLocation != nil {
			defaultZone = targetLocation
		}
		resolved, err := file.resolveRevisionByDate(revision, targetDate, defaultZone)
		if err != nil {
			return nil, err
		}
		revision = resolved
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

	if expandMode != O && expandMode != B {
		rh, err := file.GetRevisionHead(revision)
		if err != nil {
			return nil, fmt.Errorf("get revision head %s: %w", revision, err)
		}
		rc, err := file.GetRevisionContent(revision)
		if err != nil {
			return nil, fmt.Errorf("get revision content %s: %w", revision, err)
		}

		locker := ""
		// KV: locker inserted only if being locked (ci -l, co -l).
		// KVL: locker inserted if locked.
		isLocked := false
		lockedByUser := ""
		for _, l := range file.Locks {
			if l.Revision == revision {
				isLocked = true
				lockedByUser = l.User
				break
			}
		}

		if expandMode == KVL && isLocked {
			locker = lockedByUser
		} else if expandMode == KV && lockMode == WithSetLock {
			locker = user
		}

		revDate, err := ParseDate(string(rh.Date), time.Time{}, time.UTC)
		if err != nil {
			return nil, fmt.Errorf("parse date for revision %s: %w", revision, err)
		}

		kd := KeywordData{
			Revision: revision,
			Date:     revDate,
			Author:   string(rh.Author),
			State:    string(rh.State),
			Locker:   locker,
			Log:      rc.Log,
			RCSFile:  filepath.Base(rcsFilename),
			Source:   rcsFilename, // Simplified, maybe needs full path?
		}
		v.Content = ExpandKeywords(content, kd, expandMode)
	}

	return v, nil
}

func (file *File) GetRevisionHead(revision string) (*RevisionHead, error) {
	for _, rh := range file.RevisionHeads {
		if rh.Revision.String() == revision {
			return rh, nil
		}
	}
	return nil, fmt.Errorf("revision head %s not found", revision)
}

func (file *File) GetRevisionContent(revision string) (*RevisionContent, error) {
	for _, rc := range file.RevisionContents {
		if rc.Revision == revision {
			return rc, nil
		}
	}
	return nil, fmt.Errorf("revision content %s not found", revision)
}

func (file *File) resolveRevisionByDate(startRev string, targetDate time.Time, defaultZone *time.Location) (string, error) {
	rhByRevision := map[string]*RevisionHead{}
	for _, rh := range file.RevisionHeads {
		rhByRevision[rh.Revision.String()] = rh
	}

	currentRev := startRev
	// Collect all revisions in the chain starting from currentRev
	var chain []*RevisionHead
	visited := map[string]bool{}

	for {
		if visited[currentRev] {
			return "", fmt.Errorf("loop detected while resolving date for %q", startRev)
		}
		visited[currentRev] = true

		rh, ok := rhByRevision[currentRev]
		if !ok {
			return "", fmt.Errorf("revision %q not found", currentRev)
		}
		chain = append(chain, rh)

		currentRev = rh.NextRevision.String()
		if currentRev == "" {
			break
		}
	}

	// Find the revision with highest revision number that is <= targetDate
	var bestRev string

	for _, rh := range chain {
		// Parse date (RCS file dates are UTC usually, or whatever is stored)
		// We use ParseDate which handles formats.
		t, err := ParseDate(string(rh.Date), time.Time{}, defaultZone)
		if err != nil {
			// If date is unparsable, we skip it? Or fail?
			// Fail is safer.
			return "", fmt.Errorf("invalid date in revision %q: %w", rh.Revision, err)
		}

		if !t.After(targetDate) {
			// Candidate
			if bestRev == "" {
				bestRev = rh.Revision.String()
			} else {
				// Compare revisions
				if compareRevisions(rh.Revision.String(), bestRev) > 0 {
					bestRev = rh.Revision.String()
				}
			}
		}
	}

	if bestRev == "" {
		return "", fmt.Errorf("no revision found on or before %v starting from %q", targetDate, startRev)
	}
	return bestRev, nil
}

// compareRevisions compares two revision strings (e.g. "1.2" vs "1.10").
// Returns 1 if a > b, -1 if a < b, 0 if a == b.
func compareRevisions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	lenA := len(partsA)
	lenB := len(partsB)
	minLen := lenA
	if lenB < minLen {
		minLen = lenB
	}

	for i := 0; i < minLen; i++ {
		valA, _ := strconv.Atoi(partsA[i])
		valB, _ := strconv.Atoi(partsB[i])
		if valA > valB {
			return 1
		}
		if valA < valB {
			return -1
		}
	}

	if lenA > lenB {
		return 1
	}
	if lenA < lenB {
		return -1
	}
	return 0
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
