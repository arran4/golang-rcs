package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	rcs "github.com/arran4/golang-rcs"
)

type COVerdict struct {
	File          string
	RCSFile       string
	Revision      string
	FileModified  bool
	LockSet       bool
	LockCleared   bool
	CheckedOutLen int
}

// Co performs checkout operations over one or more working files.
//
// Flags:
//
//	revision: -r revision to check out
//	lock: -l lock checked-out revision
//	unlock: -u unlock checked-out revision
//	user: -w user for lock operations
//	quiet: -q suppress status output
//	date: -d date to check out
//	zone: -z zone for date parsing (e.g. "LT", "UTC", "-0700", "America/New_York")
//	files: ... List of working files to process
func Co(revision string, lock, unlock bool, user string, quiet bool, checkoutDate, checkoutZone string, files ...string) error {
	if lock && unlock {
		return fmt.Errorf("cannot combine -l and -u")
	}
	if user == "" {
		user = currentLoggedInUser()
	}
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		result, err := coFile(revision, lock, unlock, user, quiet, checkoutDate, checkoutZone, file)
		if err != nil {
			return err
		}
		if !quiet {
			fmt.Printf("co %s: rev=%s bytes=%d lock-set=%t lock-cleared=%t rcs-updated=%t\n",
				filepath.Base(result.File),
				result.Revision,
				result.CheckedOutLen,
				result.LockSet,
				result.LockCleared,
				result.FileModified,
			)
		}
	}
	return nil
}

func coFile(revision string, lock, unlock bool, user string, quiet bool, checkoutDate, checkoutZone string, workingFile string) (COVerdict, error) {
	rcsFile := workingFile
	if !strings.HasSuffix(rcsFile, ",v") {
		rcsFile += ",v"
	}
	f, err := os.Open(rcsFile)
	if err != nil {
		return COVerdict{}, fmt.Errorf("open %s: %w", rcsFile, err)
	}
	defer func() { _ = f.Close() }()

	rcsStat, err := os.Stat(rcsFile)
	if err != nil {
		return COVerdict{}, fmt.Errorf("stat %s: %w", rcsFile, err)
	}
	rcsMode := rcsStat.Mode()

	parsed, err := rcs.ParseFile(f)
	if err != nil {
		return COVerdict{}, fmt.Errorf("parse %s: %w", rcsFile, err)
	}

	ops := make([]any, 0, 3)
	if revision != "" {
		ops = append(ops, rcs.WithRevision(revision))
	}
	if checkoutDate != "" {
		zone, err := rcs.ParseZone(checkoutZone)
		if err != nil {
			return COVerdict{}, fmt.Errorf("invalid zone %q: %w", checkoutZone, err)
		}
		t, err := rcs.ParseDate(checkoutDate, time.Now(), zone)
		if err != nil {
			return COVerdict{}, fmt.Errorf("invalid date %q: %w", checkoutDate, err)
		}
		ops = append(ops, rcs.WithDate(t))
	}
	if lock {
		ops = append(ops, rcs.WithSetLock)
	} else if unlock {
		ops = append(ops, rcs.WithClearLock)
	}

	verdict, err := parsed.Checkout(user, ops...)
	if err != nil {
		return COVerdict{}, fmt.Errorf("co %s: %w", rcsFile, err)
	}

	perm := rcsMode.Perm()
	if isLockedBy(parsed, user, verdict.Revision) {
		perm |= 0200
	} else {
		perm &= ^os.FileMode(0222)
	}

	if err := os.WriteFile(workingFile, []byte(verdict.Content), perm); err != nil {
		return COVerdict{}, fmt.Errorf("write %s: %w", workingFile, err)
	}
	if err := os.Chmod(workingFile, perm); err != nil {
		return COVerdict{}, fmt.Errorf("chmod %s: %w", workingFile, err)
	}

	if verdict.FileModified {
		if err := os.WriteFile(rcsFile, []byte(parsed.String()), rcsMode.Perm()); err != nil {
			return COVerdict{}, fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return COVerdict{
		File:          workingFile,
		RCSFile:       rcsFile,
		Revision:      verdict.Revision,
		FileModified:  verdict.FileModified,
		LockSet:       verdict.LockSet,
		LockCleared:   verdict.LockCleared,
		CheckedOutLen: len(verdict.Content),
	}, nil
}

func isLockedBy(f *rcs.File, user, revision string) bool {
	for _, l := range f.Locks {
		if l.User == user && l.Revision == revision {
			return true
		}
	}
	return false
}

func currentLoggedInUser() string {
	if u := strings.TrimSpace(os.Getenv("USER")); u != "" {
		return u
	}
	if u := strings.TrimSpace(os.Getenv("USERNAME")); u != "" {
		return u
	}
	return "unknown"
}
