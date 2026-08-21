package rcs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CIVerdict is the result of a checkin operation.
type CIVerdict struct {
	Revision     string
	FileModified bool
	LockSet      bool
	LockCleared  bool
	Created      bool // true if this was the initial checkin (no prior revisions)
	// AppendedNewline is true when the input text lacked a trailing newline and
	// one was appended during normalization, altering the stored content.
	AppendedNewline bool
}

// WithForce forces a checkin even when text is unchanged from the head.
type WithForce struct{}

// WithState sets the revision state (defaults to "Exp").
type WithState string

// WithInitial indicates this is an initial checkin into an empty File.
type WithInitial struct{}

// Checkin adds a new revision to the file with the given text content.
//
// Options:
//   - WithRevision("1.5") — explicit revision number (default: increment head)
//   - WithDate(time.Time) — override checkin date (default: now)
//   - WithState("Rel") — set revision state (default: "Exp")
//   - WithForce{} — checkin even if text is unchanged
//   - WithSetLock / WithClearLock — lock mutation after checkin
//   - WithInitial{} — initial checkin into an empty file
//
// The stored text is normalized to end with a trailing newline. When the input
// lacks one it is appended and CIVerdict.AppendedNewline is set, so callers can
// warn. RCS itself can store text without a trailing newline, so this is a
// deliberate simplification rather than exact byte preservation.
func (file *File) Checkin(user, logMsg, text string, ops ...any) (*CIVerdict, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file")
	}
	if user == "" {
		return nil, fmt.Errorf("checkin requires user")
	}

	lockMode := WithNoLockChange
	var targetRev string
	var checkinDate time.Time
	state := ID("Exp")
	force := false
	initial := false

	for _, op := range ops {
		switch v := op.(type) {
		case WithRevision:
			targetRev = string(v)
		case WithDate:
			checkinDate = time.Time(v)
		case WithState:
			state = ID(v)
		case WithForce:
			force = true
		case WithInitial:
			initial = true
		case WithLock:
			lockMode = v
		default:
			return nil, fmt.Errorf("unsupported checkin option type %T", op)
		}
	}

	if checkinDate.IsZero() {
		checkinDate = time.Now().UTC()
	}
	dateStr := DateTime(checkinDate.UTC().Format(DateFormat))

	// Branch checkins (an explicit revision with more than two components, e.g.
	// 1.1.1.1) are not supported: this implementation only maintains the trunk.
	// Reject a branch-shaped revision up front rather than silently corrupting
	// the revision graph by grafting it onto the trunk head.
	if targetRev != "" && strings.Count(targetRev, ".") > 1 {
		return nil, fmt.Errorf("branch revision %q not supported: only trunk checkins (N.M) are implemented", targetRev)
	}

	if initial || file.Head == "" {
		if initial && file.Head != "" {
			return nil, fmt.Errorf("WithInitial cannot be used on a file with existing revisions (head %s)", file.Head)
		}
		return file.checkinInitial(user, logMsg, text, targetRev, dateStr, state, lockMode)
	}
	return file.checkinNormal(user, logMsg, text, targetRev, dateStr, state, lockMode, force)
}

// checkinInitial handles the first checkin into a file with no revisions.
func (file *File) checkinInitial(user, logMsg, text, targetRev string, date DateTime, state ID, lockMode WithLock) (*CIVerdict, error) {
	if targetRev == "" {
		targetRev = "1.1"
	}

	// Normalize log message: RCS convention requires trailing newline.
	if logMsg == "" {
		logMsg = "*** empty log message ***\n"
	} else if !strings.HasSuffix(logMsg, "\n") {
		logMsg += "\n"
	}

	// RCS text is normalized to end with a newline. This can alter a working
	// file that lacks a final newline; the verdict records it so the CLI warns.
	appendedNewline := text != "" && !strings.HasSuffix(text, "\n")
	if appendedNewline {
		text += "\n"
	}

	rh := &RevisionHead{
		Revision: Num(targetRev),
		Date:     date,
		Author:   ID(user),
		State:    state,
	}

	rc := &RevisionContent{
		Revision: targetRev,
		Log:      logMsg,
		Text:     text,
	}

	file.Head = targetRev
	file.Access = true
	if file.Comment == "" {
		file.Comment = "# "
	}
	file.RevisionHeads = append(file.RevisionHeads, rh)
	file.RevisionContents = append(file.RevisionContents, rc)

	v := &CIVerdict{Revision: targetRev, FileModified: true, Created: true, AppendedNewline: appendedNewline}

	switch lockMode {
	case WithSetLock:
		file.SetLock(user, targetRev)
		v.LockSet = true
	case WithClearLock:
		file.ClearLock(user, targetRev)
		v.LockCleared = true
	}

	return v, nil
}

// checkinNormal handles adding a new revision on top of an existing head.
func (file *File) checkinNormal(user, logMsg, text, targetRev string, date DateTime, state ID, lockMode WithLock, force bool) (*CIVerdict, error) {
	// Strict locking: the caller must hold the lock on the head revision before
	// checking in on top of it (matches RCS ci(1); not even -f bypasses this).
	if file.Strict {
		lockHolder := ""
		for _, lock := range file.Locks {
			if lock.Revision == file.Head {
				lockHolder = lock.User
				break
			}
		}
		switch {
		case lockHolder == "":
			return nil, fmt.Errorf("checkin aborted: no lock set by %s on %s (strict locking)", user, file.Head)
		case lockHolder != user:
			return nil, fmt.Errorf("checkin aborted: revision %s is locked by %s, not %s", file.Head, lockHolder, user)
		}
	}

	// Normalize log message: RCS convention requires trailing newline.
	if logMsg == "" {
		logMsg = "*** empty log message ***\n"
	} else if !strings.HasSuffix(logMsg, "\n") {
		logMsg += "\n"
	}

	// RCS text is normalized to end with a newline. This can alter a working
	// file that lacks a final newline; the verdict records it so the CLI warns.
	appendedNewline := text != "" && !strings.HasSuffix(text, "\n")
	if appendedNewline {
		text += "\n"
	}

	headText, err := file.resolveRevisionContent(file.Head)
	if err != nil {
		return nil, fmt.Errorf("resolve head %s: %w", file.Head, err)
	}

	if headText == text && !force {
		return nil, fmt.Errorf("checkin aborted: no changes from head %s (use WithForce to override)", file.Head)
	}

	newRev := targetRev
	if newRev == "" {
		newRev, err = incrementNum(Num(file.Head))
		if err != nil {
			return nil, fmt.Errorf("increment revision %s: %w", file.Head, err)
		}
	}

	for _, rh := range file.RevisionHeads {
		if rh.Revision.String() == newRev {
			return nil, fmt.Errorf("revision %s already exists", newRev)
		}
	}

	// Reverse delta: transforms new text → old (head) text.
	newLines := splitLines(text)
	oldLines := splitLines(headText)
	deltaStr := generateDelta(newLines, oldLines)

	for _, rc := range file.RevisionContents {
		if rc.Revision == file.Head {
			rc.Text = deltaStr
			break
		}
	}

	newHead := &RevisionHead{
		Revision:     Num(newRev),
		Date:         date,
		Author:       ID(user),
		State:        state,
		NextRevision: Num(file.Head),
	}

	newContent := &RevisionContent{
		Revision: newRev,
		Log:      logMsg,
		Text:     text,
	}

	file.ClearLock(user, string(Num(file.Head)))

	file.RevisionHeads = append([]*RevisionHead{newHead}, file.RevisionHeads...)
	file.RevisionContents = append([]*RevisionContent{newContent}, file.RevisionContents...)
	file.Head = newRev

	v := &CIVerdict{Revision: newRev, FileModified: true, AppendedNewline: appendedNewline}

	switch lockMode {
	case WithSetLock:
		file.SetLock(user, newRev)
		v.LockSet = true
	case WithClearLock:
		// Already cleared above, nothing more to do.
		v.LockCleared = true
	}

	return v, nil
}

// incrementNum bumps the last component of a revision number.
// "1.5" → "1.6", "1.1.1.3" → "1.1.1.4".
func incrementNum(n Num) (string, error) {
	s := n.String()
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid revision number %q", s)
	}
	last, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return "", fmt.Errorf("non-numeric revision component in %q: %w", s, err)
	}
	parts[len(parts)-1] = strconv.Itoa(last + 1)
	return strings.Join(parts, "."), nil
}

// generateDelta produces an RCS-format reverse delta — an ed-style script of
// "dN M" / "aN M" commands — that transforms newLines back into oldLines. It is
// stored on the previous head so checkout can reconstruct the old text from the
// new head text.
//
// It uses common-prefix/common-suffix matching: it finds the first and last
// lines that differ and rewrites the whole span between them as a single
// delete + add, ordered delete-before-add with ascending line numbers.
//
// Fidelity note: for a single contiguous change this is the minimal edit and is
// byte-identical to what real RCS ci(1) writes. For interleaved changes (several
// separated edit regions) it collapses the span between the first and last
// difference into one hunk. That delta still reconstructs correctly and real
// co(1) applies it, but it is larger than — and not byte-identical to — the
// minimal multi-hunk delta ci(1) would produce. Adopting the pluggable diff
// engine (diff.Generate) would fix minimality but currently regresses byte
// fidelity, because diff.EdDiff.String() emits add-before-delete / a0-anchored
// scripts rather than RCS-canonical form.
func generateDelta(newLines, oldLines []string) string {
	n := len(newLines)
	o := len(oldLines)

	// Common prefix.
	prefix := 0
	for prefix < n && prefix < o && newLines[prefix] == oldLines[prefix] {
		prefix++
	}

	// Common suffix (not overlapping prefix).
	si, sj := n, o
	for si > prefix && sj > prefix && newLines[si-1] == oldLines[sj-1] {
		si--
		sj--
	}

	delCount := si - prefix // lines to delete from new
	addLines := oldLines[prefix:sj]

	if delCount == 0 && len(addLines) == 0 {
		return ""
	}

	var sb strings.Builder
	if delCount > 0 {
		fmt.Fprintf(&sb, "d%d %d\n", prefix+1, delCount)
	}
	if len(addLines) > 0 {
		fmt.Fprintf(&sb, "a%d %d\n", prefix+delCount, len(addLines))
		for _, l := range addLines {
			sb.WriteString(l)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
