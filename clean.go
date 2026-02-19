package rcs

import (
	"fmt"
)

// CleanVerdict is the result of resolving a clean request.
type CleanVerdict struct {
	Revision string
	Clean    bool
	Unlocked bool
}

// Clean compares the working file content with the revision content and optionally unlocks the revision.
//
// Options:
//   - WithRevision("1.2") picks an explicit revision to compare against.
//   - WithClearLock unlocks the revision.
func (file *File) Clean(user string, workingContent []byte, ops ...any) (*CleanVerdict, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file")
	}

	revision := ""
	unlock := false

	for _, op := range ops {
		switch v := op.(type) {
		case WithRevision:
			revision = string(v)
		case WithLock:
			if v == WithClearLock {
				unlock = true
			}
		}
	}

	// If unlock is requested but no revision specified, find the revision locked by user.
	if unlock && revision == "" {
		for _, l := range file.Locks {
			if l.User == user {
				revision = l.Revision
				break
			}
		}
		if revision == "" {
			return nil, fmt.Errorf("no lock found for user %q", user)
		}
	}

	// If revision is still empty, use default revision (Head).
	if revision == "" {
		revision = file.Head
	}

	if revision == "" {
		return nil, fmt.Errorf("revision not specified and no head found")
	}

	verdict := &CleanVerdict{Revision: revision}

	// Compare content
	content, err := file.resolveRevisionContent(revision)
	if err != nil {
		return nil, fmt.Errorf("resolve content for %s: %w", revision, err)
	}

	if string(workingContent) == content {
		verdict.Clean = true
	}

	// Unlock if requested AND clean
	if unlock && verdict.Clean {
		if changed := file.clearLock(user, revision); changed {
			verdict.Unlocked = true
		}
	}

	return verdict, nil
}
